package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/devldavydov/myhealth-ng/internal/entity"
	"github.com/devldavydov/myhealth-ng/internal/port"
)

type activeCaloriesHandler struct {
	activeCalories port.ActiveCaloriesUseCases
}

type activeCaloriesRequest struct {
	DT    *string  `json:"dt"`
	Value *float64 `json:"value"`
}

type activeCaloriesResponse struct {
	DT    string  `json:"dt"`
	Value float64 `json:"value"`
}

func newActiveCaloriesHandler(activeCalories port.ActiveCaloriesUseCases) *activeCaloriesHandler {
	return &activeCaloriesHandler{activeCalories: activeCalories}
}

func (handler *activeCaloriesHandler) get(ctx *gin.Context) {
	dt, err := parseWeightDate(ctx.Query("dt"), "dt")
	if err != nil {
		respondActiveCaloriesError(ctx, err)
		return
	}
	item, err := handler.activeCalories.Get(ctx.Request.Context(), currentUserID(ctx), dt)
	if err != nil {
		respondActiveCaloriesError(ctx, err)
		return
	}
	if item == nil {
		ctx.JSON(http.StatusOK, gin.H{"data": nil})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": toActiveCaloriesResponse(*item)})
}

func (handler *activeCaloriesHandler) save(ctx *gin.Context) {
	data, err := decodeActiveCalories(ctx)
	if err != nil {
		respondActiveCaloriesError(ctx, err)
		return
	}
	item, err := handler.activeCalories.Save(ctx.Request.Context(), currentUserID(ctx), data)
	if err != nil {
		respondActiveCaloriesError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": toActiveCaloriesResponse(item)})
}

func (handler *activeCaloriesHandler) delete(ctx *gin.Context) {
	dt, err := parseWeightDate(ctx.Param("dt"), "dt")
	if err != nil {
		respondActiveCaloriesError(ctx, err)
		return
	}
	if err := handler.activeCalories.Delete(ctx.Request.Context(), currentUserID(ctx), dt); err != nil {
		respondActiveCaloriesError(ctx, err)
		return
	}
	ctx.Status(http.StatusNoContent)
}

func decodeActiveCalories(ctx *gin.Context) (entity.ActiveCalories, error) {
	ctx.Request.Body = http.MaxBytesReader(ctx.Writer, ctx.Request.Body, maxJSONBodyBytes)
	decoder := json.NewDecoder(ctx.Request.Body)
	decoder.DisallowUnknownFields()
	var payload activeCaloriesRequest
	if err := decoder.Decode(&payload); err != nil {
		return entity.ActiveCalories{}, bodyValidationError("Ожидается корректный JSON не более 100 КБ")
	}
	if err := ensureJSONEnds(decoder); err != nil {
		return entity.ActiveCalories{}, bodyValidationError("Ожидается один JSON-объект")
	}
	details := make(map[string][]string)
	if payload.DT == nil {
		details["dt"] = []string{"Поле обязательно"}
	}
	if payload.Value == nil {
		details["value"] = []string{"Поле обязательно"}
	}
	if len(details) > 0 {
		return entity.ActiveCalories{}, &entity.ValidationError{Fields: details}
	}
	dt, err := parseWeightDate(*payload.DT, "dt")
	if err != nil {
		return entity.ActiveCalories{}, err
	}
	return entity.ActiveCalories{DT: dt, Value: *payload.Value}, nil
}

func respondActiveCaloriesError(ctx *gin.Context, err error) {
	var validationError *entity.ValidationError
	switch {
	case errors.As(err, &validationError):
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Некорректные активные калории", "details": validationError.Fields})
	case errors.Is(err, port.ErrActiveCaloriesNotFound):
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Активные калории за дату не найдены"})
	default:
		_ = ctx.Error(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Внутренняя ошибка сервера"})
	}
}

func toActiveCaloriesResponse(item entity.ActiveCalories) activeCaloriesResponse {
	return activeCaloriesResponse{DT: item.DT.Format(dateLayout), Value: item.Value}
}
