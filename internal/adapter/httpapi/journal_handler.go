package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/devldavydov/myhealth-ng/internal/entity"
	"github.com/devldavydov/myhealth-ng/internal/port"
)

type journalHandler struct {
	journal port.JournalUseCases
}

type journalItemRequest struct {
	FoodKey *string  `json:"foodKey"`
	Weight  *float64 `json:"weight"`
}

type journalSaveRequest struct {
	DT    *string               `json:"dt"`
	Meal  *string               `json:"meal"`
	Items *[]journalItemRequest `json:"items"`
}

type journalPercentResponse struct {
	Protein float64 `json:"protein"`
	Fat     float64 `json:"fat"`
	Carb    float64 `json:"carb"`
}

type journalTotalsResponse struct {
	Weight  float64 `json:"weight"`
	Cal     float64 `json:"cal"`
	Protein float64 `json:"protein"`
	Fat     float64 `json:"fat"`
	Carb    float64 `json:"carb"`
}

type journalItemResponse struct {
	Food   foodResponse `json:"food"`
	Weight float64      `json:"weight"`
}

type journalZoneResponse struct {
	Meal   string                `json:"meal"`
	Items  []journalItemResponse `json:"items"`
	Totals journalTotalsResponse `json:"totals"`
}

type journalDayResponse struct {
	DT           string                 `json:"dt"`
	Zones        []journalZoneResponse  `json:"zones"`
	Totals       journalTotalsResponse  `json:"totals"`
	MacroPercent journalPercentResponse `json:"macroPercent"`
}

func newJournalHandler(journal port.JournalUseCases) *journalHandler {
	return &journalHandler{journal: journal}
}

func (handler *journalHandler) get(ctx *gin.Context) {
	dt, err := parseWeightDate(ctx.Query("dt"), "dt")
	if err != nil {
		respondJournalError(ctx, err)
		return
	}
	day, err := handler.journal.Get(ctx.Request.Context(), currentUserID(ctx), dt)
	if err != nil {
		respondJournalError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": toJournalDayResponse(day)})
}

func (handler *journalHandler) save(ctx *gin.Context) {
	dt, meal, items, err := decodeJournalSave(ctx)
	if err != nil {
		respondJournalError(ctx, err)
		return
	}
	day, err := handler.journal.Save(ctx.Request.Context(), currentUserID(ctx), dt, meal, items)
	if err != nil {
		respondJournalError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": toJournalDayResponse(day)})
}

func (handler *journalHandler) deleteItem(ctx *gin.Context) {
	dt, err := parseWeightDate(ctx.Param("dt"), "dt")
	if err != nil {
		respondJournalError(ctx, err)
		return
	}
	day, err := handler.journal.DeleteItem(ctx.Request.Context(), currentUserID(ctx), dt, entity.MealType(ctx.Param("meal")), ctx.Param("foodKey"))
	if err != nil {
		respondJournalError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": toJournalDayResponse(day)})
}

func (handler *journalHandler) clearMeal(ctx *gin.Context) {
	dt, err := parseWeightDate(ctx.Param("dt"), "dt")
	if err != nil {
		respondJournalError(ctx, err)
		return
	}
	day, err := handler.journal.ClearMeal(ctx.Request.Context(), currentUserID(ctx), dt, entity.MealType(ctx.Param("meal")))
	if err != nil {
		respondJournalError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": toJournalDayResponse(day)})
}

func decodeJournalSave(ctx *gin.Context) (time.Time, entity.MealType, []entity.JournalItemData, error) {
	ctx.Request.Body = http.MaxBytesReader(ctx.Writer, ctx.Request.Body, maxJSONBodyBytes)
	decoder := json.NewDecoder(ctx.Request.Body)
	decoder.DisallowUnknownFields()
	var payload journalSaveRequest
	if err := decoder.Decode(&payload); err != nil {
		return time.Time{}, "", nil, bodyValidationError("Ожидается корректный JSON не более 100 КБ")
	}
	if err := ensureJSONEnds(decoder); err != nil {
		return time.Time{}, "", nil, bodyValidationError("Ожидается один JSON-объект")
	}
	details := make(map[string][]string)
	if payload.DT == nil {
		details["dt"] = []string{"Поле обязательно"}
	}
	if payload.Meal == nil {
		details["meal"] = []string{"Поле обязательно"}
	}
	if payload.Items == nil {
		details["items"] = []string{"Поле обязательно"}
	}
	if len(details) > 0 {
		return time.Time{}, "", nil, &entity.ValidationError{Fields: details}
	}
	dt, err := parseWeightDate(*payload.DT, "dt")
	if err != nil {
		return time.Time{}, "", nil, err
	}
	items := make([]entity.JournalItemData, 0, len(*payload.Items))
	for index, item := range *payload.Items {
		if item.FoodKey == nil || item.Weight == nil {
			details["items"] = append(details["items"], fmt.Sprintf("Строка %d должна содержать foodKey и weight", index+1))
			continue
		}
		items = append(items, entity.JournalItemData{FoodKey: *item.FoodKey, Weight: *item.Weight})
	}
	if len(details) > 0 {
		return time.Time{}, "", nil, &entity.ValidationError{Fields: details}
	}
	return dt, entity.MealType(*payload.Meal), items, nil
}

func respondJournalError(ctx *gin.Context, err error) {
	var validationError *entity.ValidationError
	switch {
	case errors.As(err, &validationError):
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Некорректные данные журнала", "details": validationError.Fields})
	case errors.Is(err, port.ErrJournalItemNotFound):
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Запись журнала не найдена"})
	default:
		_ = ctx.Error(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Внутренняя ошибка сервера"})
	}
}

func toJournalDayResponse(day entity.JournalDay) journalDayResponse {
	zones := make([]journalZoneResponse, 0, len(day.Zones))
	for _, zone := range day.Zones {
		items := make([]journalItemResponse, 0, len(zone.Items))
		for _, item := range zone.Items {
			items = append(items, journalItemResponse{Food: toFoodResponse(item.Food), Weight: item.Weight})
		}
		zones = append(zones, journalZoneResponse{Meal: string(zone.Meal), Items: items, Totals: toJournalTotalsResponse(zone.Totals)})
	}
	return journalDayResponse{
		DT: day.DT.Format(dateLayout), Zones: zones, Totals: toJournalTotalsResponse(day.Totals),
		MacroPercent: journalPercentResponse{Protein: day.MacroPercent.Protein, Fat: day.MacroPercent.Fat, Carb: day.MacroPercent.Carb},
	}
}

func toJournalTotalsResponse(totals entity.JournalTotals) journalTotalsResponse {
	return journalTotalsResponse{Weight: totals.Weight, Cal: totals.Cal, Protein: totals.Protein, Fat: totals.Fat, Carb: totals.Carb}
}
