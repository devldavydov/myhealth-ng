package httpapi

import (
	"errors"
	"net/http"

	"github.com/devldavydov/myhealth-ng/internal/entity"
	"github.com/devldavydov/myhealth-ng/internal/port"
	"github.com/gin-gonic/gin"
)

type sportHandler struct{ sport port.SportUseCases }
type sportRequest struct {
	Name    *string `json:"name"`
	Unit    *string `json:"unit"`
	Comment *string `json:"comment"`
}
type sportResponse struct {
	Key     string `json:"key"`
	Name    string `json:"name"`
	Unit    string `json:"unit"`
	Comment string `json:"comment"`
}

func newSportHandler(sport port.SportUseCases) *sportHandler { return &sportHandler{sport: sport} }
func (h *sportHandler) list(ctx *gin.Context) {
	request, err := decodePageRequest(ctx)
	if err != nil {
		respondSportError(ctx, err)
		return
	}
	page, err := h.sport.List(ctx.Request.Context(), request)
	if err != nil {
		respondSportError(ctx, err)
		return
	}
	items := make([]sportResponse, 0, len(page.Items))
	for _, item := range page.Items {
		items = append(items, toSportResponse(item))
	}
	ctx.JSON(http.StatusOK, gin.H{"data": items, "pagination": toPaginationResponse(page)})
}
func (h *sportHandler) get(ctx *gin.Context) {
	item, err := h.sport.Get(ctx.Request.Context(), ctx.Param("key"))
	if err != nil {
		respondSportError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": toSportResponse(item)})
}
func (h *sportHandler) create(ctx *gin.Context) {
	data, err := decodeSport(ctx)
	if err != nil {
		respondSportError(ctx, err)
		return
	}
	item, err := h.sport.Create(ctx.Request.Context(), data)
	if err != nil {
		respondSportError(ctx, err)
		return
	}
	ctx.JSON(http.StatusCreated, gin.H{"data": toSportResponse(item)})
}
func (h *sportHandler) update(ctx *gin.Context) {
	data, err := decodeSport(ctx)
	if err != nil {
		respondSportError(ctx, err)
		return
	}
	item, err := h.sport.Update(ctx.Request.Context(), ctx.Param("key"), data)
	if err != nil {
		respondSportError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": toSportResponse(item)})
}
func (h *sportHandler) delete(ctx *gin.Context) {
	if err := h.sport.Delete(ctx.Request.Context(), ctx.Param("key")); err != nil {
		respondSportError(ctx, err)
		return
	}
	ctx.Status(http.StatusNoContent)
}
func decodeSport(ctx *gin.Context) (entity.SportData, error) {
	var payload sportRequest
	if err := decodeJSON(ctx, &payload); err != nil {
		return entity.SportData{}, err
	}
	details := map[string][]string{}
	if payload.Name == nil {
		details["name"] = []string{"Поле обязательно"}
	}
	if payload.Unit == nil {
		details["unit"] = []string{"Поле обязательно"}
	}
	if payload.Comment == nil {
		details["comment"] = []string{"Поле обязательно"}
	}
	if len(details) > 0 {
		return entity.SportData{}, &entity.ValidationError{Fields: details}
	}
	return entity.SportData{Name: *payload.Name, Unit: *payload.Unit, Comment: *payload.Comment}, nil
}
func toSportResponse(item entity.Sport) sportResponse {
	return sportResponse{Key: item.Key, Name: item.Name, Unit: item.Unit, Comment: item.Comment}
}
func respondSportError(ctx *gin.Context, err error) {
	var validation *entity.ValidationError
	switch {
	case errors.As(err, &validation):
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Некорректные данные спорта", "details": validation.Fields})
	case errors.Is(err, port.ErrSportNotFound):
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Вид спорта не найден"})
	case errors.Is(err, port.ErrSportConflict):
		ctx.JSON(http.StatusConflict, gin.H{"error": "Вид спорта с таким ключом уже существует"})
	case errors.Is(err, port.ErrSportInUse):
		ctx.JSON(http.StatusConflict, gin.H{"error": "Вид спорта используется в активности. Сначала удалите связанные записи"})
	default:
		_ = ctx.Error(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Внутренняя ошибка сервера"})
	}
}
