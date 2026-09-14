package auth

import (
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/devldavydov/myhealth-ng/internal/domain"
)

const developmentGUID = "00000000-0000-4000-8000-000000000000"

var userGUIDPattern = regexp.MustCompile(`(?i)^urn:myhealth:user:([0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12})$`)

type ClientCertificateError struct {
	Message string
}

func (err *ClientCertificateError) Error() string {
	return err.Message
}

func ReadClientIdentity(request *http.Request, certificateRequired bool) (domain.UserIdentity, error) {
	if request.Header.Get("X-Client-Verify") != "SUCCESS" || request.Header.Get("X-Client-Cert") == "" {
		if certificateRequired {
			return domain.UserIdentity{}, &ClientCertificateError{Message: "Требуется проверенный клиентский сертификат"}
		}
		return domain.UserIdentity{
			GUID:                   developmentGUID,
			Name:                   "Локальный пользователь",
			CertificateFingerprint: "development",
			LastSeenAt:             nowISO(),
		}, nil
	}

	escapedCertificate := request.Header.Get("X-Client-Cert")
	certificatePEM, err := url.PathUnescape(escapedCertificate)
	if err != nil {
		return domain.UserIdentity{}, invalidCertificate(err)
	}

	block, _ := pem.Decode([]byte(certificatePEM))
	if block == nil || block.Type != "CERTIFICATE" {
		return domain.UserIdentity{}, invalidCertificate(fmt.Errorf("не найден PEM-блок сертификата"))
	}
	certificate, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return domain.UserIdentity{}, invalidCertificate(err)
	}

	guid := ""
	for _, uri := range certificate.URIs {
		matches := userGUIDPattern.FindStringSubmatch(uri.String())
		if len(matches) == 2 {
			guid = strings.ToLower(matches[1])
			break
		}
	}
	name := certificate.Subject.CommonName
	if guid == "" || name == "" {
		return domain.UserIdentity{}, invalidCertificate(fmt.Errorf("в сертификате отсутствуют CN или MyHealth GUID"))
	}

	fingerprint := sha256.Sum256(certificate.Raw)
	return domain.UserIdentity{
		GUID:                   guid,
		Name:                   name,
		CertificateFingerprint: formatFingerprint(fingerprint[:]),
		LastSeenAt:             nowISO(),
	}, nil
}

func invalidCertificate(err error) error {
	return &ClientCertificateError{Message: "Некорректный клиентский сертификат: " + err.Error()}
}

func formatFingerprint(value []byte) string {
	encoded := strings.ToUpper(hex.EncodeToString(value))
	parts := make([]string, 0, len(encoded)/2)
	for offset := 0; offset < len(encoded); offset += 2 {
		parts = append(parts, encoded[offset:offset+2])
	}
	return strings.Join(parts, ":")
}

func nowISO() string {
	return time.Now().UTC().Truncate(time.Millisecond).Format("2006-01-02T15:04:05.000Z")
}
