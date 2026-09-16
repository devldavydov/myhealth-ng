package cases

import (
	"context"
	"crypto/rand"
	"fmt"
	"math"
	"strings"

	"github.com/devldavydov/myhealth-ng/internal/entity"
	"github.com/devldavydov/myhealth-ng/internal/port"
)

type Food struct {
	repository  port.FoodRepository
	generateKey func() (string, error)
}

func NewFood(repository port.FoodRepository) *Food {
	return &Food{repository: repository, generateKey: randomUUID}
}

func (food *Food) List(ctx context.Context, request entity.PageRequest) (entity.Page[entity.Food], error) {
	request, err := normalizePageRequest(request)
	if err != nil {
		return entity.Page[entity.Food]{}, err
	}
	return food.repository.List(ctx, request)
}

func (food *Food) Get(ctx context.Context, key string) (entity.Food, error) {
	if err := validateKey(key); err != nil {
		return entity.Food{}, err
	}
	return food.repository.Get(ctx, strings.ToLower(strings.TrimSpace(key)))
}

func (food *Food) Create(ctx context.Context, data entity.FoodData) (entity.Food, error) {
	normalized, err := validateFood(data)
	if err != nil {
		return entity.Food{}, err
	}
	key, err := food.generateKey()
	if err != nil {
		return entity.Food{}, fmt.Errorf("generate food key: %w", err)
	}
	return food.repository.Create(ctx, entity.Food{
		Key: key, Name: normalized.Name, Brand: normalized.Brand,
		Cal100: normalized.Cal100, Prot100: normalized.Prot100,
		Fat100: normalized.Fat100, Carb100: normalized.Carb100,
		Comment: normalized.Comment,
	})
}

func (food *Food) Update(ctx context.Context, key string, data entity.FoodData) (entity.Food, error) {
	if err := validateKey(key); err != nil {
		return entity.Food{}, err
	}
	normalized, err := validateFood(data)
	if err != nil {
		return entity.Food{}, err
	}
	return food.repository.Update(ctx, strings.ToLower(strings.TrimSpace(key)), normalized)
}

func (food *Food) Delete(ctx context.Context, key string) error {
	if err := validateKey(key); err != nil {
		return err
	}
	return food.repository.Delete(ctx, strings.ToLower(strings.TrimSpace(key)))
}

func validateFood(data entity.FoodData) (entity.FoodData, error) {
	data.Name = strings.TrimSpace(data.Name)
	data.Brand = strings.TrimSpace(data.Brand)
	data.Comment = strings.TrimSpace(data.Comment)
	details := make(map[string][]string)
	if data.Name == "" {
		details["name"] = []string{"Название обязательно"}
	}
	validateNumber(details, "cal100", data.Cal100)
	validateNumber(details, "prot100", data.Prot100)
	validateNumber(details, "fat100", data.Fat100)
	validateNumber(details, "carb100", data.Carb100)
	if len(details) > 0 {
		return entity.FoodData{}, &entity.ValidationError{Fields: details}
	}
	return data, nil
}

func validateNumber(details map[string][]string, field string, value float64) {
	if math.IsNaN(value) || math.IsInf(value, 0) || value < 0 {
		details[field] = []string{"Ожидается конечное неотрицательное число"}
	}
}

func validateKey(key string) error {
	key = strings.ToLower(strings.TrimSpace(key))
	if len(key) != 36 || key[8] != '-' || key[13] != '-' || key[18] != '-' || key[23] != '-' {
		return invalidKey()
	}
	for index, character := range key {
		if index == 8 || index == 13 || index == 18 || index == 23 {
			continue
		}
		if !strings.ContainsRune("0123456789abcdef", character) {
			return invalidKey()
		}
	}
	return nil
}

func invalidKey() error {
	return &entity.ValidationError{Fields: map[string][]string{"key": {"Ожидается UUID"}}}
}

func randomUUID() (string, error) {
	value := make([]byte, 16)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}
	value[6] = (value[6] & 0x0f) | 0x40
	value[8] = (value[8] & 0x3f) | 0x80
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		value[0:4], value[4:6], value[6:8], value[8:10], value[10:16]), nil
}
