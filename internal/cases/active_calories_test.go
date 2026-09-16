package cases

import (
	"context"
	"errors"
	"math"
	"testing"
	"time"

	"github.com/devldavydov/myhealth-ng/internal/entity"
)

type activeCaloriesRepositoryStub struct {
	userID string
	dt     time.Time
	item   *entity.ActiveCalories
	saved  entity.ActiveCalories
	err    error
}

func (stub *activeCaloriesRepositoryStub) Get(_ context.Context, userID string, dt time.Time) (*entity.ActiveCalories, error) {
	stub.userID, stub.dt = userID, dt
	return stub.item, stub.err
}

func (stub *activeCaloriesRepositoryStub) Save(_ context.Context, userID string, data entity.ActiveCalories) (entity.ActiveCalories, error) {
	stub.userID, stub.dt, stub.saved = userID, data.DT, data
	return data, stub.err
}

func (stub *activeCaloriesRepositoryStub) Delete(_ context.Context, userID string, dt time.Time) error {
	stub.userID, stub.dt = userID, dt
	return stub.err
}

func TestActiveCaloriesGetSaveAndDeleteNormalizeScope(t *testing.T) {
	dt := time.Date(2026, time.September, 16, 18, 30, 0, 0, time.FixedZone("test", 3*60*60))
	item := &entity.ActiveCalories{DT: dt, Value: 2450.5}
	repository := &activeCaloriesRepositoryStub{item: item}
	useCases := NewActiveCalories(repository)

	loaded, err := useCases.Get(context.Background(), " user-1 ", dt)
	if err != nil || loaded != item || repository.userID != "user-1" || repository.dt.Hour() != 0 || repository.dt.Location() != time.UTC {
		t.Fatalf("loaded=%+v repository=%+v err=%v", loaded, repository, err)
	}
	saved, err := useCases.Save(context.Background(), " user-1 ", entity.ActiveCalories{DT: dt, Value: 2600.5})
	if err != nil || saved.Value != 2600.5 || repository.saved.DT.Hour() != 0 {
		t.Fatalf("saved=%+v repository=%+v err=%v", saved, repository, err)
	}
	if err := useCases.Delete(context.Background(), " user-1 ", dt); err != nil || repository.dt.Hour() != 0 {
		t.Fatalf("repository=%+v err=%v", repository, err)
	}
}

func TestActiveCaloriesValidatesInput(t *testing.T) {
	validDate := time.Date(2026, time.September, 16, 0, 0, 0, 0, time.UTC)
	for _, test := range []struct {
		name   string
		userID string
		dt     time.Time
		value  float64
	}{
		{"missing user", "", validDate, 2000},
		{"missing date", "user-1", time.Time{}, 2000},
		{"zero", "user-1", validDate, 0},
		{"negative", "user-1", validDate, -1},
		{"nan", "user-1", validDate, math.NaN()},
		{"infinite", "user-1", validDate, math.Inf(1)},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewActiveCalories(&activeCaloriesRepositoryStub{}).Save(context.Background(), test.userID, entity.ActiveCalories{DT: test.dt, Value: test.value})
			var validationError *entity.ValidationError
			if !errors.As(err, &validationError) {
				t.Fatalf("expected validation error, got %v", err)
			}
		})
	}
}

func TestActiveCaloriesGetAndDeleteValidateScope(t *testing.T) {
	useCases := NewActiveCalories(&activeCaloriesRepositoryStub{})
	if _, err := useCases.Get(context.Background(), "", time.Now()); err == nil {
		t.Fatal("expected get validation error")
	}
	if err := useCases.Delete(context.Background(), "user-1", time.Time{}); err == nil {
		t.Fatal("expected delete validation error")
	}
}
