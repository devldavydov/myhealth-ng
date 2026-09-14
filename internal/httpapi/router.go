package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/devldavydov/myhealth-ng/internal/auth"
	"github.com/devldavydov/myhealth-ng/internal/domain"
	"github.com/devldavydov/myhealth-ng/internal/repository"
)

const (
	identityContextKey = "myhealth.identity"
	maxJSONBodyBytes   = 100 * 1024
)

type Options struct {
	CertificateRequired bool
	ClientOrigin        string
}

type createMeasurementRequest struct {
	Type       *domain.MeasurementType `json:"type"`
	Value      *float64                `json:"value"`
	Unit       *string                 `json:"unit"`
	MeasuredAt *string                 `json:"measuredAt"`
}

func NewRouter(measurements repository.MeasurementRepository, users repository.UserRegistry, options Options) *gin.Engine {
	router := gin.New()
	_ = router.SetTrustedProxies(nil)
	router.Use(gin.Logger(), recovery(), cors(options.ClientOrigin), identity(users, options.CertificateRequired))

	router.GET("/api/health", func(context *gin.Context) {
		context.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	router.GET("/api/me", func(context *gin.Context) {
		context.JSON(http.StatusOK, gin.H{"data": context.MustGet(identityContextKey)})
	})
	router.GET("/api/measurements", func(context *gin.Context) {
		items, err := measurements.FindAll(context.Request.Context())
		if err != nil {
			_ = context.Error(err)
			context.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Внутренняя ошибка сервера"})
			return
		}
		context.JSON(http.StatusOK, gin.H{"data": items})
	})
	router.POST("/api/measurements", func(context *gin.Context) {
		request, details, err := decodeMeasurement(context)
		if err != nil {
			context.JSON(http.StatusBadRequest, gin.H{
				"error":   "Некорректные данные измерения",
				"details": details,
			})
			return
		}

		created, err := measurements.Create(context.Request.Context(), request)
		if err != nil {
			_ = context.Error(err)
			context.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Внутренняя ошибка сервера"})
			return
		}
		context.JSON(http.StatusCreated, gin.H{"data": created})
	})
	router.NoRoute(func(context *gin.Context) {
		context.JSON(http.StatusNotFound, gin.H{"error": "Маршрут не найден"})
	})
	return router
}

func decodeMeasurement(context *gin.Context) (domain.NewMeasurement, map[string][]string, error) {
	context.Request.Body = http.MaxBytesReader(context.Writer, context.Request.Body, maxJSONBodyBytes)
	decoder := json.NewDecoder(context.Request.Body)
	var payload createMeasurementRequest
	if err := decoder.Decode(&payload); err != nil {
		return domain.NewMeasurement{}, map[string][]string{"body": {"Ожидается корректный JSON не более 100 КБ"}}, err
	}
	if err := ensureJSONEnds(decoder); err != nil {
		return domain.NewMeasurement{}, map[string][]string{"body": {"Ожидается один JSON-объект"}}, err
	}

	details := make(map[string][]string)
	if payload.Type == nil || !payload.Type.Valid() {
		details["type"] = []string{"Допустимые значения: weight, pressure, pulse"}
	}
	if payload.Value == nil || math.IsNaN(valueOrZero(payload.Value)) || math.IsInf(valueOrZero(payload.Value), 0) {
		details["value"] = []string{"Ожидается конечное число"}
	}
	unit := ""
	if payload.Unit != nil {
		unit = strings.TrimSpace(*payload.Unit)
	}
	if unit == "" || len([]rune(unit)) > 20 {
		details["unit"] = []string{"Единица измерения должна содержать от 1 до 20 символов"}
	}
	if payload.MeasuredAt == nil {
		details["measuredAt"] = []string{"Ожидается дата и время в формате ISO 8601"}
	} else if _, err := time.Parse(time.RFC3339, *payload.MeasuredAt); err != nil {
		details["measuredAt"] = []string{"Ожидается дата и время в формате ISO 8601"}
	}
	if len(details) > 0 {
		return domain.NewMeasurement{}, details, errors.New("measurement validation failed")
	}

	return domain.NewMeasurement{
		Type:       *payload.Type,
		Value:      *payload.Value,
		Unit:       unit,
		MeasuredAt: *payload.MeasuredAt,
	}, nil, nil
}

func ensureJSONEnds(decoder *json.Decoder) error {
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("multiple JSON values")
		}
		return err
	}
	return nil
}

func valueOrZero(value *float64) float64 {
	if value == nil {
		return 0
	}
	return *value
}

func identity(users repository.UserRegistry, certificateRequired bool) gin.HandlerFunc {
	return func(context *gin.Context) {
		user, err := auth.ReadClientIdentity(context.Request, certificateRequired)
		if err != nil {
			var certificateError *auth.ClientCertificateError
			if errors.As(err, &certificateError) {
				context.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": certificateError.Error()})
				return
			}
			_ = context.Error(err)
			context.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Внутренняя ошибка сервера"})
			return
		}
		if err := users.Remember(context.Request.Context(), user); err != nil {
			_ = context.Error(err)
			context.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Внутренняя ошибка сервера"})
			return
		}
		context.Set(identityContextKey, user)
		context.Next()
	}
}

func cors(origin string) gin.HandlerFunc {
	if origin == "" {
		origin = "http://localhost:5173"
	}
	return func(context *gin.Context) {
		if context.GetHeader("Origin") == origin {
			context.Header("Access-Control-Allow-Origin", origin)
			context.Header("Access-Control-Allow-Headers", "Content-Type")
			context.Header("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			context.Header("Vary", "Origin")
		}
		if context.Request.Method == http.MethodOptions {
			context.AbortWithStatus(http.StatusNoContent)
			return
		}
		context.Next()
	}
}

func recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(context *gin.Context, recovered any) {
		log.Printf("panic while handling request: %v", recovered)
		context.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "Внутренняя ошибка сервера"})
	})
}
