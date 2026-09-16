package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/devldavydov/myhealth-ng/internal/entity"
	"github.com/devldavydov/myhealth-ng/internal/port"
)

type bundleHandler struct {
	bundles port.BundleUseCases
}

type bundleItemRequest struct {
	FoodKey *string  `json:"foodKey"`
	Weight  *float64 `json:"weight"`
}

type bundleRequest struct {
	Name  *string              `json:"name"`
	Items *[]bundleItemRequest `json:"items"`
}

type bundleTotalsResponse struct {
	Weight  float64 `json:"weight"`
	Cal     float64 `json:"cal"`
	Protein float64 `json:"protein"`
	Fat     float64 `json:"fat"`
	Carb    float64 `json:"carb"`
}

type bundleSummaryResponse struct {
	Key       string               `json:"key"`
	Name      string               `json:"name"`
	ItemCount int                  `json:"itemCount"`
	Totals    bundleTotalsResponse `json:"totals"`
}

type bundleItemResponse struct {
	Food   foodResponse `json:"food"`
	Weight float64      `json:"weight"`
}

type bundleResponse struct {
	Key    string               `json:"key"`
	Name   string               `json:"name"`
	Items  []bundleItemResponse `json:"items"`
	Totals bundleTotalsResponse `json:"totals"`
}

func newBundleHandler(bundles port.BundleUseCases) *bundleHandler {
	return &bundleHandler{bundles: bundles}
}

func (handler *bundleHandler) list(ctx *gin.Context) {
	request, err := decodePageRequest(ctx)
	if err != nil {
		respondBundleError(ctx, err)
		return
	}
	page, err := handler.bundles.List(ctx.Request.Context(), request)
	if err != nil {
		respondBundleError(ctx, err)
		return
	}
	response := make([]bundleSummaryResponse, 0, len(page.Items))
	for _, item := range page.Items {
		response = append(response, toBundleSummaryResponse(item))
	}
	ctx.JSON(http.StatusOK, gin.H{"data": response, "pagination": toPaginationResponse(page)})
}

func (handler *bundleHandler) get(ctx *gin.Context) {
	item, err := handler.bundles.Get(ctx.Request.Context(), ctx.Param("key"))
	if err != nil {
		respondBundleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": toBundleResponse(item)})
}

func (handler *bundleHandler) create(ctx *gin.Context) {
	data, err := decodeBundle(ctx)
	if err != nil {
		respondBundleError(ctx, err)
		return
	}
	item, err := handler.bundles.Create(ctx.Request.Context(), data)
	if err != nil {
		respondBundleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusCreated, gin.H{"data": toBundleResponse(item)})
}

func (handler *bundleHandler) update(ctx *gin.Context) {
	data, err := decodeBundle(ctx)
	if err != nil {
		respondBundleError(ctx, err)
		return
	}
	item, err := handler.bundles.Update(ctx.Request.Context(), ctx.Param("key"), data)
	if err != nil {
		respondBundleError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": toBundleResponse(item)})
}

func (handler *bundleHandler) delete(ctx *gin.Context) {
	if err := handler.bundles.Delete(ctx.Request.Context(), ctx.Param("key")); err != nil {
		respondBundleError(ctx, err)
		return
	}
	ctx.Status(http.StatusNoContent)
}

func decodeBundle(ctx *gin.Context) (entity.BundleData, error) {
	ctx.Request.Body = http.MaxBytesReader(ctx.Writer, ctx.Request.Body, maxJSONBodyBytes)
	decoder := json.NewDecoder(ctx.Request.Body)
	decoder.DisallowUnknownFields()
	var payload bundleRequest
	if err := decoder.Decode(&payload); err != nil {
		return entity.BundleData{}, bodyValidationError("Ожидается корректный JSON не более 100 КБ")
	}
	if err := ensureJSONEnds(decoder); err != nil {
		return entity.BundleData{}, bodyValidationError("Ожидается один JSON-объект")
	}

	details := make(map[string][]string)
	if payload.Name == nil {
		details["name"] = []string{"Поле обязательно"}
	}
	if payload.Items == nil {
		details["items"] = []string{"Поле обязательно"}
	}
	if len(details) > 0 {
		return entity.BundleData{}, &entity.ValidationError{Fields: details}
	}

	items := make([]entity.BundleItemData, 0, len(*payload.Items))
	for index, item := range *payload.Items {
		if item.FoodKey == nil || item.Weight == nil {
			details["items"] = append(details["items"], fmt.Sprintf("Строка %d должна содержать foodKey и weight", index+1))
			continue
		}
		items = append(items, entity.BundleItemData{FoodKey: *item.FoodKey, Weight: *item.Weight})
	}
	if len(details) > 0 {
		return entity.BundleData{}, &entity.ValidationError{Fields: details}
	}
	return entity.BundleData{Name: *payload.Name, Items: items}, nil
}

func respondBundleError(ctx *gin.Context, err error) {
	var validationError *entity.ValidationError
	switch {
	case errors.As(err, &validationError):
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Некорректные данные бандла", "details": validationError.Fields})
	case errors.Is(err, port.ErrBundleNotFound):
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Бандл не найден"})
	case errors.Is(err, port.ErrBundleConflict):
		ctx.JSON(http.StatusConflict, gin.H{"error": "Бандл с таким ключом уже существует"})
	default:
		_ = ctx.Error(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Внутренняя ошибка сервера"})
	}
}

func toBundleSummaryResponse(bundle entity.BundleSummary) bundleSummaryResponse {
	return bundleSummaryResponse{Key: bundle.Key, Name: bundle.Name, ItemCount: bundle.ItemCount, Totals: toBundleTotalsResponse(bundle.Totals)}
}

func toBundleResponse(bundle entity.Bundle) bundleResponse {
	items := make([]bundleItemResponse, 0, len(bundle.Items))
	for _, item := range bundle.Items {
		items = append(items, bundleItemResponse{Food: toFoodResponse(item.Food), Weight: item.Weight})
	}
	return bundleResponse{Key: bundle.Key, Name: bundle.Name, Items: items, Totals: toBundleTotalsResponse(bundle.Totals)}
}

func toBundleTotalsResponse(totals entity.BundleTotals) bundleTotalsResponse {
	return bundleTotalsResponse{Weight: totals.Weight, Cal: totals.Cal, Protein: totals.Protein, Fat: totals.Fat, Carb: totals.Carb}
}
