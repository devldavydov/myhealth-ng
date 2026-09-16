package cases

import (
	"context"
	"errors"
	"math"
	"strings"
	"time"

	"github.com/devldavydov/myhealth-ng/internal/entity"
	"github.com/devldavydov/myhealth-ng/internal/port"
)

type Journal struct {
	repository port.JournalRepository
}

func NewJournal(repository port.JournalRepository) *Journal {
	return &Journal{repository: repository}
}

func (journal *Journal) Get(ctx context.Context, userID string, dt time.Time) (entity.JournalDay, error) {
	userID, dt, err := validateJournalScope(userID, dt, "dt")
	if err != nil {
		return entity.JournalDay{}, err
	}
	return journal.loadDay(ctx, userID, dt)
}

func (journal *Journal) Save(ctx context.Context, userID string, dt time.Time, meal entity.MealType, items []entity.JournalItemData) (entity.JournalDay, error) {
	userID, dt, err := validateJournalScope(userID, dt, "dt")
	if err != nil {
		return entity.JournalDay{}, err
	}
	meal, err = validateMeal(meal)
	if err != nil {
		return entity.JournalDay{}, err
	}
	items, err = validateJournalItems(items)
	if err != nil {
		return entity.JournalDay{}, err
	}
	if err := journal.repository.Save(ctx, userID, dt, meal, items); err != nil {
		if errors.Is(err, port.ErrJournalFoodNotFound) {
			return entity.JournalDay{}, &entity.ValidationError{Fields: map[string][]string{"items": {"Один из продуктов не найден"}}}
		}
		return entity.JournalDay{}, err
	}
	return journal.loadDay(ctx, userID, dt)
}

func (journal *Journal) DeleteItem(ctx context.Context, userID string, dt time.Time, meal entity.MealType, foodKey string) (entity.JournalDay, error) {
	userID, dt, err := validateJournalScope(userID, dt, "dt")
	if err != nil {
		return entity.JournalDay{}, err
	}
	meal, err = validateMeal(meal)
	if err != nil {
		return entity.JournalDay{}, err
	}
	foodKey = strings.ToLower(strings.TrimSpace(foodKey))
	if err := validateKey(foodKey); err != nil {
		return entity.JournalDay{}, err
	}
	if err := journal.repository.DeleteItem(ctx, userID, dt, meal, foodKey); err != nil {
		return entity.JournalDay{}, err
	}
	return journal.loadDay(ctx, userID, dt)
}

func (journal *Journal) ClearMeal(ctx context.Context, userID string, dt time.Time, meal entity.MealType) (entity.JournalDay, error) {
	userID, dt, err := validateJournalScope(userID, dt, "dt")
	if err != nil {
		return entity.JournalDay{}, err
	}
	meal, err = validateMeal(meal)
	if err != nil {
		return entity.JournalDay{}, err
	}
	if err := journal.repository.ClearMeal(ctx, userID, dt, meal); err != nil {
		return entity.JournalDay{}, err
	}
	return journal.loadDay(ctx, userID, dt)
}

func (journal *Journal) loadDay(ctx context.Context, userID string, dt time.Time) (entity.JournalDay, error) {
	items, err := journal.repository.List(ctx, userID, dt)
	if err != nil {
		return entity.JournalDay{}, err
	}
	return assembleJournalDay(dt, items), nil
}

func validateJournalScope(userID string, dt time.Time, dateField string) (string, time.Time, error) {
	userID = strings.TrimSpace(userID)
	details := make(map[string][]string)
	validateUserID(details, userID)
	if dt.IsZero() {
		details[dateField] = []string{"Дата обязательна"}
	}
	if len(details) > 0 {
		return "", time.Time{}, &entity.ValidationError{Fields: details}
	}
	return userID, dateOnly(dt), nil
}

func validateMeal(meal entity.MealType) (entity.MealType, error) {
	meal = entity.MealType(strings.TrimSpace(string(meal)))
	for _, allowed := range entity.MealTypes {
		if meal == allowed {
			return meal, nil
		}
	}
	return "", &entity.ValidationError{Fields: map[string][]string{"meal": {"Неизвестный тип приёма пищи"}}}
}

func validateJournalItems(items []entity.JournalItemData) ([]entity.JournalItemData, error) {
	details := make(map[string][]string)
	if len(items) == 0 {
		details["items"] = []string{"Добавьте хотя бы один продукт"}
	}
	normalized := make([]entity.JournalItemData, 0, len(items))
	indexes := make(map[string]int)
	for _, item := range items {
		item.FoodKey = strings.ToLower(strings.TrimSpace(item.FoodKey))
		if err := validateKey(item.FoodKey); err != nil {
			details["items"] = append(details["items"], "Некорректный ключ продукта")
			continue
		}
		if math.IsNaN(item.Weight) || math.IsInf(item.Weight, 0) || item.Weight <= 0 {
			details["items"] = append(details["items"], "Вес продукта должен быть положительным числом")
			continue
		}
		if index, exists := indexes[item.FoodKey]; exists {
			normalized[index] = item
			continue
		}
		indexes[item.FoodKey] = len(normalized)
		normalized = append(normalized, item)
	}
	if len(details) > 0 {
		return nil, &entity.ValidationError{Fields: details}
	}
	return normalized, nil
}

func assembleJournalDay(dt time.Time, items []entity.JournalItem) entity.JournalDay {
	day := entity.JournalDay{DT: dateOnly(dt), Zones: make([]entity.JournalZone, len(entity.MealTypes))}
	zoneIndexes := make(map[entity.MealType]int, len(entity.MealTypes))
	for index, meal := range entity.MealTypes {
		day.Zones[index] = entity.JournalZone{Meal: meal, Items: make([]entity.JournalItem, 0)}
		zoneIndexes[meal] = index
	}
	for _, item := range items {
		index, exists := zoneIndexes[item.Meal]
		if !exists {
			continue
		}
		day.Zones[index].Items = append(day.Zones[index].Items, item)
		addJournalTotals(&day.Zones[index].Totals, item)
		addJournalTotals(&day.Totals, item)
	}
	totalMacros := day.Totals.Protein + day.Totals.Fat + day.Totals.Carb
	if totalMacros > 0 {
		day.MacroPercent = entity.MacroPercent{
			Protein: day.Totals.Protein / totalMacros * 100,
			Fat:     day.Totals.Fat / totalMacros * 100,
			Carb:    day.Totals.Carb / totalMacros * 100,
		}
	}
	return day
}

func addJournalTotals(totals *entity.JournalTotals, item entity.JournalItem) {
	multiplier := item.Weight / 100
	totals.Weight += item.Weight
	totals.Cal += item.Food.Cal100 * multiplier
	totals.Protein += item.Food.Prot100 * multiplier
	totals.Fat += item.Food.Fat100 * multiplier
	totals.Carb += item.Food.Carb100 * multiplier
}
