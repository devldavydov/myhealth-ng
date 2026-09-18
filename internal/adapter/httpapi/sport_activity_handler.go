package httpapi

import (
	"errors"
	"net/http"

	"github.com/devldavydov/myhealth-ng/internal/entity"
	"github.com/devldavydov/myhealth-ng/internal/port"
	"github.com/gin-gonic/gin"
)

type sportActivityHandler struct{ activity port.SportActivityUseCases }
type sportActivityRequest struct {
	DT       *string    `json:"dt"`
	SportKey *string    `json:"sportKey"`
	Sets     *[]float64 `json:"sets"`
}
type sportActivityResponse struct {
	DT    string        `json:"dt"`
	Sport sportResponse `json:"sport"`
	Sets  []float64     `json:"sets"`
}

func newSportActivityHandler(activity port.SportActivityUseCases) *sportActivityHandler {
	return &sportActivityHandler{activity: activity}
}
func (h *sportActivityHandler) list(ctx *gin.Context) {
	period, err := decodeSportActivityRange(ctx)
	if err != nil {
		respondSportActivityError(ctx, err)
		return
	}
	items, err := h.activity.List(ctx.Request.Context(), currentUserID(ctx), period)
	if err != nil {
		respondSportActivityError(ctx, err)
		return
	}
	response := make([]sportActivityResponse, 0, len(items))
	for _, item := range items {
		response = append(response, toSportActivityResponse(item))
	}
	ctx.JSON(http.StatusOK, gin.H{"data": response})
}
func (h *sportActivityHandler) save(ctx *gin.Context) {
	data, err := decodeSportActivity(ctx)
	if err != nil {
		respondSportActivityError(ctx, err)
		return
	}
	item, err := h.activity.Save(ctx.Request.Context(), currentUserID(ctx), data)
	if err != nil {
		respondSportActivityError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": toSportActivityResponse(item)})
}
func (h *sportActivityHandler) delete(ctx *gin.Context) {
	dt, err := parseWeightDate(ctx.Param("dt"), "dt")
	if err != nil {
		respondSportActivityError(ctx, err)
		return
	}
	if err = h.activity.Delete(ctx.Request.Context(), currentUserID(ctx), entity.SportActivityData{DT: dt, SportKey: ctx.Param("sportKey")}); err != nil {
		respondSportActivityError(ctx, err)
		return
	}
	ctx.Status(http.StatusNoContent)
}
func decodeSportActivity(ctx *gin.Context) (entity.SportActivityData, error) {
	var payload sportActivityRequest
	if err := decodeJSON(ctx, &payload); err != nil {
		return entity.SportActivityData{}, err
	}
	details := map[string][]string{}
	if payload.DT == nil {
		details["dt"] = []string{"Поле обязательно"}
	}
	if payload.SportKey == nil {
		details["sportKey"] = []string{"Поле обязательно"}
	}
	if payload.Sets == nil {
		details["sets"] = []string{"Поле обязательно"}
	}
	if len(details) > 0 {
		return entity.SportActivityData{}, &entity.ValidationError{Fields: details}
	}
	dt, err := parseWeightDate(*payload.DT, "dt")
	if err != nil {
		return entity.SportActivityData{}, err
	}
	return entity.SportActivityData{DT: dt, SportKey: *payload.SportKey, Sets: *payload.Sets}, nil
}
func decodeSportActivityRange(ctx *gin.Context) (entity.SportActivityRange, error) {
	from, err := parseOptionalWeightDate(ctx.Query("from"), "from")
	if err != nil {
		return entity.SportActivityRange{}, err
	}
	to, err := parseOptionalWeightDate(ctx.Query("to"), "to")
	if err != nil {
		return entity.SportActivityRange{}, err
	}
	return entity.SportActivityRange{From: from, To: to}, nil
}
func toSportActivityResponse(item entity.SportActivity) sportActivityResponse {
	return sportActivityResponse{DT: item.DT.Format(dateLayout), Sport: toSportResponse(item.Sport), Sets: item.Sets}
}
func respondSportActivityError(ctx *gin.Context, err error) {
	var validation *entity.ValidationError
	switch {
	case errors.As(err, &validation):
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Некорректные данные спортивной активности", "details": validation.Fields})
	case errors.Is(err, port.ErrSportNotFound):
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Вид спорта не найден"})
	case errors.Is(err, port.ErrSportActivityNotFound):
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Запись активности не найдена"})
	default:
		_ = ctx.Error(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Внутренняя ошибка сервера"})
	}
}
