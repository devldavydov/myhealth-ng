package httpapi

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/devldavydov/myhealth-ng/internal/auth"
)

const identityContextKey = "myhealth.identity"

func identity(certificateRequired bool) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		user, err := auth.ReadClientIdentity(ctx.Request, certificateRequired)
		if err != nil {
			var certificateError *auth.ClientCertificateError
			if errors.As(err, &certificateError) {
				ctx.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": certificateError.Error()})
				return
			}
			_ = ctx.Error(err)
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Внутренняя ошибка сервера"})
			return
		}
		ctx.Set(identityContextKey, user)
		ctx.Next()
	}
}

func recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(ctx *gin.Context, recovered any) {
		log.Printf("panic while handling request: %v", recovered)
		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Внутренняя ошибка сервера"})
	})
}

func notFound(ctx *gin.Context) {
	ctx.JSON(http.StatusNotFound, gin.H{"error": "Маршрут не найден"})
}
