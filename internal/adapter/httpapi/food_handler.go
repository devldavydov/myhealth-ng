package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/devldavydov/myhealth-ng/internal/entity"
	"github.com/devldavydov/myhealth-ng/internal/port"
)

const maxJSONBodyBytes = 100 * 1024

type foodHandler struct {
	food port.FoodUseCases
}

type foodRequest struct {
	Name    *string  `json:"name"`
	Brand   *string  `json:"brand"`
	Cal100  *float64 `json:"cal100"`
	Prot100 *float64 `json:"prot100"`
	Fat100  *float64 `json:"fat100"`
	Carb100 *float64 `json:"carb100"`
	Comment *string  `json:"comment"`
}

type foodResponse struct {
	Key     string  `json:"key"`
	Name    string  `json:"name"`
	Brand   string  `json:"brand"`
	Cal100  float64 `json:"cal100"`
	Prot100 float64 `json:"prot100"`
	Fat100  float64 `json:"fat100"`
	Carb100 float64 `json:"carb100"`
	Comment string  `json:"comment"`
}

func newFoodHandler(food port.FoodUseCases) *foodHandler {
	return &foodHandler{food: food}
}

func (handler *foodHandler) list(ctx *gin.Context) {
	items, err := handler.food.List(ctx.Request.Context(), ctx.Query("q"))
	if err != nil {
		respondFoodError(ctx, err)
		return
	}
	response := make([]foodResponse, 0, len(items))
	for _, item := range items {
		response = append(response, toFoodResponse(item))
	}
	ctx.JSON(http.StatusOK, gin.H{"data": response})
}

func (handler *foodHandler) get(ctx *gin.Context) {
	item, err := handler.food.Get(ctx.Request.Context(), ctx.Param("key"))
	if err != nil {
		respondFoodError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": toFoodResponse(item)})
}

func (handler *foodHandler) create(ctx *gin.Context) {
	data, err := decodeFood(ctx)
	if err != nil {
		respondFoodError(ctx, err)
		return
	}
	item, err := handler.food.Create(ctx.Request.Context(), data)
	if err != nil {
		respondFoodError(ctx, err)
		return
	}
	ctx.JSON(http.StatusCreated, gin.H{"data": toFoodResponse(item)})
}

func (handler *foodHandler) update(ctx *gin.Context) {
	data, err := decodeFood(ctx)
	if err != nil {
		respondFoodError(ctx, err)
		return
	}
	item, err := handler.food.Update(ctx.Request.Context(), ctx.Param("key"), data)
	if err != nil {
		respondFoodError(ctx, err)
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": toFoodResponse(item)})
}

func (handler *foodHandler) delete(ctx *gin.Context) {
	if err := handler.food.Delete(ctx.Request.Context(), ctx.Param("key")); err != nil {
		respondFoodError(ctx, err)
		return
	}
	ctx.Status(http.StatusNoContent)
}

func decodeFood(ctx *gin.Context) (entity.FoodData, error) {
	ctx.Request.Body = http.MaxBytesReader(ctx.Writer, ctx.Request.Body, maxJSONBodyBytes)
	decoder := json.NewDecoder(ctx.Request.Body)
	decoder.DisallowUnknownFields()
	var payload foodRequest
	if err := decoder.Decode(&payload); err != nil {
		return entity.FoodData{}, bodyValidationError("Ожидается корректный JSON не более 100 КБ")
	}
	if err := ensureJSONEnds(decoder); err != nil {
		return entity.FoodData{}, bodyValidationError("Ожидается один JSON-объект")
	}

	details := make(map[string][]string)
	if payload.Name == nil {
		details["name"] = []string{"Поле обязательно"}
	}
	if payload.Brand == nil {
		details["brand"] = []string{"Поле обязательно"}
	}
	if payload.Cal100 == nil {
		details["cal100"] = []string{"Поле обязательно"}
	}
	if payload.Prot100 == nil {
		details["prot100"] = []string{"Поле обязательно"}
	}
	if payload.Fat100 == nil {
		details["fat100"] = []string{"Поле обязательно"}
	}
	if payload.Carb100 == nil {
		details["carb100"] = []string{"Поле обязательно"}
	}
	if payload.Comment == nil {
		details["comment"] = []string{"Поле обязательно"}
	}
	if len(details) > 0 {
		return entity.FoodData{}, &entity.ValidationError{Fields: details}
	}
	return entity.FoodData{
		Name: *payload.Name, Brand: *payload.Brand,
		Cal100: *payload.Cal100, Prot100: *payload.Prot100,
		Fat100: *payload.Fat100, Carb100: *payload.Carb100,
		Comment: *payload.Comment,
	}, nil
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

func bodyValidationError(message string) error {
	return &entity.ValidationError{Fields: map[string][]string{"body": {message}}}
}

func respondFoodError(ctx *gin.Context, err error) {
	var validationError *entity.ValidationError
	switch {
	case errors.As(err, &validationError):
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Некорректные данные продукта", "details": validationError.Fields,
		})
	case errors.Is(err, port.ErrFoodNotFound):
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Продукт не найден"})
	case errors.Is(err, port.ErrFoodConflict):
		ctx.JSON(http.StatusConflict, gin.H{"error": "Продукт с таким ключом уже существует"})
	case errors.Is(err, port.ErrFoodInUse):
		ctx.JSON(http.StatusConflict, gin.H{"error": "Продукт используется в бандле. Сначала удалите его из всех бандлов"})
	default:
		_ = ctx.Error(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Внутренняя ошибка сервера"})
	}
}

func toFoodResponse(food entity.Food) foodResponse {
	return foodResponse{
		Key: food.Key, Name: food.Name, Brand: food.Brand,
		Cal100: food.Cal100, Prot100: food.Prot100,
		Fat100: food.Fat100, Carb100: food.Carb100,
		Comment: food.Comment,
	}
}
