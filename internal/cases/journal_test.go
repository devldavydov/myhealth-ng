package cases

import (
	"context"
	"errors"
	"math"
	"testing"
	"time"

	"github.com/devldavydov/myhealth-ng/internal/entity"
	"github.com/devldavydov/myhealth-ng/internal/port"
)

type journalRepositoryStub struct {
	userID  string
	dt      time.Time
	meal    entity.MealType
	items   []entity.JournalItemData
	foodKey string
	listed  []entity.JournalItem
	err     error
}

func (stub *journalRepositoryStub) List(_ context.Context, userID string, dt time.Time) ([]entity.JournalItem, error) {
	stub.userID, stub.dt = userID, dt
	return stub.listed, stub.err
}
func (stub *journalRepositoryStub) Save(_ context.Context, userID string, dt time.Time, meal entity.MealType, items []entity.JournalItemData) error {
	stub.userID, stub.dt, stub.meal, stub.items = userID, dt, meal, items
	return stub.err
}
func (stub *journalRepositoryStub) DeleteItem(_ context.Context, userID string, dt time.Time, meal entity.MealType, foodKey string) error {
	stub.userID, stub.dt, stub.meal, stub.foodKey = userID, dt, meal, foodKey
	return stub.err
}
func (stub *journalRepositoryStub) ClearMeal(_ context.Context, userID string, dt time.Time, meal entity.MealType) error {
	stub.userID, stub.dt, stub.meal = userID, dt, meal
	return stub.err
}

func TestJournalBuildsOrderedZonesTotalsAndMassPercent(t *testing.T) {
	dt := time.Date(2026, time.September, 16, 15, 0, 0, 0, time.UTC)
	repository := &journalRepositoryStub{listed: []entity.JournalItem{{
		Meal:   entity.MealLunch,
		Food:   entity.Food{Key: "3f67c05f-7c9e-4cb5-b26a-f9ce5b065865", Cal100: 200, Prot100: 10, Fat100: 5, Carb100: 25},
		Weight: 100,
	}}}
	day, err := NewJournal(repository).Get(context.Background(), " user-1 ", dt)
	if err != nil {
		t.Fatal(err)
	}
	if len(day.Zones) != 6 || day.Zones[0].Meal != entity.MealBreakfast || day.Zones[2].Meal != entity.MealLunch || len(day.Zones[2].Items) != 1 {
		t.Fatalf("unexpected zones: %+v", day.Zones)
	}
	if day.DT.Hour() != 0 || repository.userID != "user-1" || repository.dt.Hour() != 0 || day.Totals.Cal != 200 {
		t.Fatalf("day=%+v repository=%+v", day, repository)
	}
	if math.Abs(day.MacroPercent.Protein-25) > .001 || math.Abs(day.MacroPercent.Fat-12.5) > .001 || math.Abs(day.MacroPercent.Carb-62.5) > .001 {
		t.Fatalf("unexpected percentages: %+v", day.MacroPercent)
	}
}

func TestJournalSaveReplacesDuplicatePayloadAndValidates(t *testing.T) {
	key := "3f67c05f-7c9e-4cb5-b26a-f9ce5b065865"
	repository := &journalRepositoryStub{}
	_, err := NewJournal(repository).Save(context.Background(), "user-1", time.Now(), entity.MealBreakfast, []entity.JournalItemData{
		{FoodKey: key, Weight: 40},
		{FoodKey: key, Weight: 60},
	})
	if err != nil || len(repository.items) != 1 || repository.items[0].Weight != 60 {
		t.Fatalf("items=%+v err=%v", repository.items, err)
	}

	for _, test := range []struct {
		name   string
		userID string
		dt     time.Time
		meal   entity.MealType
		items  []entity.JournalItemData
	}{
		{"user", "", time.Now(), entity.MealBreakfast, []entity.JournalItemData{{FoodKey: key, Weight: 1}}},
		{"date", "user", time.Time{}, entity.MealBreakfast, []entity.JournalItemData{{FoodKey: key, Weight: 1}}},
		{"meal", "user", time.Now(), "ночь", []entity.JournalItemData{{FoodKey: key, Weight: 1}}},
		{"empty", "user", time.Now(), entity.MealBreakfast, nil},
		{"weight", "user", time.Now(), entity.MealBreakfast, []entity.JournalItemData{{FoodKey: key, Weight: 0}}},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewJournal(&journalRepositoryStub{}).Save(context.Background(), test.userID, test.dt, test.meal, test.items)
			var validationError *entity.ValidationError
			if !errors.As(err, &validationError) {
				t.Fatalf("expected validation error, got %v", err)
			}
		})
	}
}

func TestJournalMapsMissingFood(t *testing.T) {
	repository := &journalRepositoryStub{err: port.ErrJournalFoodNotFound}
	_, err := NewJournal(repository).Save(context.Background(), "user", time.Now(), entity.MealDinner, []entity.JournalItemData{{
		FoodKey: "3f67c05f-7c9e-4cb5-b26a-f9ce5b065865", Weight: 100,
	}})
	var validationError *entity.ValidationError
	if !errors.As(err, &validationError) || len(validationError.Fields["items"]) == 0 {
		t.Fatalf("unexpected error: %v", err)
	}
}
