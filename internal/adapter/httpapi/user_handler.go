package httpapi

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/devldavydov/myhealth-ng/internal/entity"
)

type userHandler struct{}

type userResponse struct {
	GUID                   string `json:"guid"`
	Name                   string `json:"name"`
	CertificateFingerprint string `json:"certificateFingerprint"`
	LastSeenAt             string `json:"lastSeenAt"`
}

func newUserHandler() *userHandler {
	return &userHandler{}
}

func (*userHandler) getCurrent(ctx *gin.Context) {
	user := ctx.MustGet(identityContextKey).(entity.UserIdentity)
	ctx.JSON(http.StatusOK, gin.H{"data": userResponse{
		GUID: user.GUID, Name: user.Name,
		CertificateFingerprint: user.CertificateFingerprint,
		LastSeenAt:             user.LastSeenAt,
	}})
}
