package cases

import (
	"context"
	"errors"
	"math"
	"strings"

	"github.com/devldavydov/myhealth-ng/internal/entity"
	"github.com/devldavydov/myhealth-ng/internal/port"
)

type Bundle struct {
	repository  port.BundleRepository
	generateKey func() (string, error)
}

func NewBundle(repository port.BundleRepository) *Bundle {
	return &Bundle{repository: repository, generateKey: randomUUID}
}

func (bundle *Bundle) List(ctx context.Context, query string) ([]entity.BundleSummary, error) {
	return bundle.repository.List(ctx, strings.TrimSpace(query))
}

func (bundle *Bundle) Get(ctx context.Context, key string) (entity.Bundle, error) {
	if err := validateKey(key); err != nil {
		return entity.Bundle{}, err
	}
	return bundle.repository.Get(ctx, strings.ToLower(strings.TrimSpace(key)))
}

func (bundle *Bundle) Create(ctx context.Context, data entity.BundleData) (entity.Bundle, error) {
	normalized, err := validateBundle(data)
	if err != nil {
		return entity.Bundle{}, err
	}
	key, err := bundle.generateKey()
	if err != nil {
		return entity.Bundle{}, err
	}
	created, err := bundle.repository.Create(ctx, entity.Bundle{Key: key, Name: normalized.Name, Items: bundleItems(normalized.Items)})
	return created, mapBundleSaveError(err)
}

func (bundle *Bundle) Update(ctx context.Context, key string, data entity.BundleData) (entity.Bundle, error) {
	if err := validateKey(key); err != nil {
		return entity.Bundle{}, err
	}
	normalized, err := validateBundle(data)
	if err != nil {
		return entity.Bundle{}, err
	}
	updated, err := bundle.repository.Update(ctx, strings.ToLower(strings.TrimSpace(key)), normalized)
	return updated, mapBundleSaveError(err)
}

func (bundle *Bundle) Delete(ctx context.Context, key string) error {
	if err := validateKey(key); err != nil {
		return err
	}
	return bundle.repository.Delete(ctx, strings.ToLower(strings.TrimSpace(key)))
}

func validateBundle(data entity.BundleData) (entity.BundleData, error) {
	data.Name = strings.TrimSpace(data.Name)
	details := make(map[string][]string)
	if data.Name == "" {
		details["name"] = []string{"Название обязательно"}
	}
	if len(data.Items) == 0 {
		details["items"] = []string{"Добавьте хотя бы один продукт"}
	}

	merged := make([]entity.BundleItemData, 0, len(data.Items))
	indexes := make(map[string]int)
	for _, item := range data.Items {
		item.FoodKey = strings.ToLower(strings.TrimSpace(item.FoodKey))
		field := "items"
		if err := validateKey(item.FoodKey); err != nil {
			details[field] = append(details[field], "Некорректный ключ продукта")
			continue
		}
		if math.IsNaN(item.Weight) || math.IsInf(item.Weight, 0) || item.Weight <= 0 {
			details[field] = append(details[field], "Вес продукта должен быть положительным числом")
			continue
		}
		if existing, ok := indexes[item.FoodKey]; ok {
			merged[existing].Weight += item.Weight
			if math.IsInf(merged[existing].Weight, 0) {
				details[field] = append(details[field], "Суммарный вес продукта слишком велик")
			}
			continue
		}
		indexes[item.FoodKey] = len(merged)
		merged = append(merged, item)
	}
	if len(details) > 0 {
		return entity.BundleData{}, &entity.ValidationError{Fields: details}
	}
	data.Items = merged
	return data, nil
}

func mapBundleSaveError(err error) error {
	if errors.Is(err, port.ErrBundleFoodNotFound) {
		return &entity.ValidationError{Fields: map[string][]string{"items": {"Один из продуктов не найден"}}}
	}
	return err
}

func bundleItems(items []entity.BundleItemData) []entity.BundleItem {
	result := make([]entity.BundleItem, 0, len(items))
	for _, item := range items {
		result = append(result, entity.BundleItem{Food: entity.Food{Key: item.FoodKey}, Weight: item.Weight})
	}
	return result
}
