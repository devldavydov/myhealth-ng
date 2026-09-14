package auth_test

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/devldavydov/myhealth-ng/internal/auth"
)

func TestReadClientIdentity(t *testing.T) {
	const guid = "3f67c05f-7c9e-4cb5-b26a-f9ce5b065865"
	certificatePEM := makeCertificate(t, "Иван Иванов", guid)
	request := httptest.NewRequest("GET", "/api/me", nil)
	request.Header.Set("X-Client-Verify", "SUCCESS")
	request.Header.Set("X-Client-Cert", url.PathEscape(certificatePEM))

	identity, err := auth.ReadClientIdentity(request, true)
	if err != nil {
		t.Fatal(err)
	}
	if identity.GUID != guid || identity.Name != "Иван Иванов" {
		t.Fatalf("unexpected identity: %+v", identity)
	}
	if identity.CertificateFingerprint == "" || identity.LastSeenAt == "" {
		t.Fatalf("identity metadata is missing: %+v", identity)
	}
}

func TestRejectRequestWithoutCertificate(t *testing.T) {
	request := httptest.NewRequest("GET", "/api/me", nil)
	_, err := auth.ReadClientIdentity(request, true)
	if err == nil || err.Error() != "Требуется проверенный клиентский сертификат" {
		t.Fatalf("error = %v", err)
	}
}

func makeCertificate(t *testing.T, name, guid string) string {
	t.Helper()
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	identityURI, err := url.Parse("urn:myhealth:user:" + guid)
	if err != nil {
		t.Fatal(err)
	}
	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: name},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(time.Hour),
		URIs:         []*url.URL{identityURI},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &privateKey.PublicKey, privateKey)
	if err != nil {
		t.Fatal(err)
	}
	return string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}))
}
