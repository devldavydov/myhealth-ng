package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/devldavydov/myhealth-ng/internal/entity"
	"github.com/devldavydov/myhealth-ng/internal/port"
)

type settingsHandler struct {
	settings port.SettingsUseCases
}

type settingsRequest struct {
	DefaultDailyCalorieLimit *int `json:"defaultDailyCalorieLimit"`
}

type settingsResponse struct {
	DefaultDailyCalorieLimit *int `json:"defaultDailyCalorieLimit"`
}

func newSettingsHandler(settings port.SettingsUseCases) *settingsHandler {
	return &settingsHandler{settings: settings}
}

func (handler *settingsHandler) get(ctx *gin.Context) {
	settings, err := handler.settings.Get(ctx.Request.Context(), currentUserID(ctx))
	if err != nil {
		respondSettingsError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": toSettingsResponse(settings)})
}

func (handler *settingsHandler) save(ctx *gin.Context) {
	data, err := decodeSettings(ctx)
	if err != nil {
		respondSettingsError(ctx, err)
		return
	}
	saved, err := handler.settings.Save(ctx.Request.Context(), currentUserID(ctx), data)
	if err != nil {
		respondSettingsError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": toSettingsResponse(&saved)})
}

func decodeSettings(ctx *gin.Context) (entity.UserSettings, error) {
	ctx.Request.Body = http.MaxBytesReader(ctx.Writer, ctx.Request.Body, maxJSONBodyBytes)
	decoder := json.NewDecoder(ctx.Request.Body)
	decoder.DisallowUnknownFields()
	var payload settingsRequest
	if err := decoder.Decode(&payload); err != nil {
		return entity.UserSettings{}, bodyValidationError("Ожидается корректный JSON не более 100 КБ")
	}
	if err := ensureJSONEnds(decoder); err != nil {
		return entity.UserSettings{}, bodyValidationError("Ожидается один JSON-объект")
	}
	if payload.DefaultDailyCalorieLimit == nil {
		return entity.UserSettings{}, &entity.ValidationError{Fields: map[string][]string{
			"defaultDailyCalorieLimit": {"Поле обязательно"},
		}}
	}
	return entity.UserSettings{DefaultDailyCalorieLimit: *payload.DefaultDailyCalorieLimit}, nil
}

func toSettingsResponse(settings *entity.UserSettings) settingsResponse {
	if settings == nil {
		return settingsResponse{}
	}
	value := settings.DefaultDailyCalorieLimit
	return settingsResponse{DefaultDailyCalorieLimit: &value}
}

func respondSettingsError(ctx *gin.Context, err error) {
	var validationError *entity.ValidationError
	if errors.As(err, &validationError) {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Некорректные настройки", "details": validationError.Fields,
		})
		return
	}
	_ = ctx.Error(err)
	ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Внутренняя ошибка сервера"})
}
