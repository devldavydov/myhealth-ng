package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/devldavydov/myhealth-ng/internal/entity"
	"github.com/devldavydov/myhealth-ng/internal/port"
)

const dateLayout = "2006-01-02"

type weightHandler struct {
	weight port.WeightUseCases
}

type weightRequest struct {
	DT    *string  `json:"dt"`
	Value *float64 `json:"value"`
}

type weightResponse struct {
	DT    string  `json:"dt"`
	Value float64 `json:"value"`
}

func newWeightHandler(weight port.WeightUseCases) *weightHandler {
	return &weightHandler{weight: weight}
}

func (handler *weightHandler) list(ctx *gin.Context) {
	period, err := decodeWeightRange(ctx)
	if err != nil {
		respondWeightError(ctx, err)
		return
	}
	items, err := handler.weight.List(ctx.Request.Context(), currentUserID(ctx), period)
	if err != nil {
		respondWeightError(ctx, err)
		return
	}
	response := make([]weightResponse, 0, len(items))
	for _, item := range items {
		response = append(response, toWeightResponse(item))
	}
	ctx.JSON(http.StatusOK, gin.H{"data": response})
}

func (handler *weightHandler) save(ctx *gin.Context) {
	data, err := decodeWeight(ctx)
	if err != nil {
		respondWeightError(ctx, err)
		return
	}
	item, err := handler.weight.Save(ctx.Request.Context(), currentUserID(ctx), data)
	if err != nil {
		respondWeightError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": toWeightResponse(item)})
}

func (handler *weightHandler) delete(ctx *gin.Context) {
	dt, err := parseWeightDate(ctx.Param("dt"), "dt")
	if err != nil {
		respondWeightError(ctx, err)
		return
	}
	if err := handler.weight.Delete(ctx.Request.Context(), currentUserID(ctx), entity.Weight{DT: dt}); err != nil {
		respondWeightError(ctx, err)
		return
	}
	ctx.Status(http.StatusNoContent)
}

func decodeWeight(ctx *gin.Context) (entity.Weight, error) {
	ctx.Request.Body = http.MaxBytesReader(ctx.Writer, ctx.Request.Body, maxJSONBodyBytes)
	decoder := json.NewDecoder(ctx.Request.Body)
	decoder.DisallowUnknownFields()
	var payload weightRequest
	if err := decoder.Decode(&payload); err != nil {
		return entity.Weight{}, bodyValidationError("Ожидается корректный JSON не более 100 КБ")
	}
	if err := ensureJSONEnds(decoder); err != nil {
		return entity.Weight{}, bodyValidationError("Ожидается один JSON-объект")
	}

	details := make(map[string][]string)
	if payload.DT == nil {
		details["dt"] = []string{"Поле обязательно"}
	}
	if payload.Value == nil {
		details["value"] = []string{"Поле обязательно"}
	}
	if len(details) > 0 {
		return entity.Weight{}, &entity.ValidationError{Fields: details}
	}
	dt, err := parseWeightDate(*payload.DT, "dt")
	if err != nil {
		return entity.Weight{}, err
	}
	return entity.Weight{DT: dt, Value: *payload.Value}, nil
}

func decodeWeightRange(ctx *gin.Context) (entity.WeightRange, error) {
	from, err := parseOptionalWeightDate(ctx.Query("from"), "from")
	if err != nil {
		return entity.WeightRange{}, err
	}
	to, err := parseOptionalWeightDate(ctx.Query("to"), "to")
	if err != nil {
		return entity.WeightRange{}, err
	}
	return entity.WeightRange{From: from, To: to}, nil
}

func parseOptionalWeightDate(value, field string) (*time.Time, error) {
	if value == "" {
		return nil, nil
	}
	parsed, err := parseWeightDate(value, field)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func parseWeightDate(value, field string) (time.Time, error) {
	parsed, err := time.Parse(dateLayout, value)
	if err != nil {
		return time.Time{}, &entity.ValidationError{Fields: map[string][]string{
			field: {"Ожидается дата в формате ГГГГ-ММ-ДД"},
		}}
	}
	return parsed, nil
}

func currentUserID(ctx *gin.Context) string {
	return ctx.MustGet(identityContextKey).(entity.UserIdentity).GUID
}

func respondWeightError(ctx *gin.Context, err error) {
	var validationError *entity.ValidationError
	switch {
	case errors.As(err, &validationError):
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Некорректные данные веса", "details": validationError.Fields,
		})
	case errors.Is(err, port.ErrWeightNotFound):
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Запись веса не найдена"})
	default:
		_ = ctx.Error(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Внутренняя ошибка сервера"})
	}
}

func toWeightResponse(weight entity.Weight) weightResponse {
	return weightResponse{DT: weight.DT.Format(dateLayout), Value: weight.Value}
}
