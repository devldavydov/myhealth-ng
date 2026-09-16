package cases

import (
	"context"
	"math"
	"strings"
	"time"

	"github.com/devldavydov/myhealth-ng/internal/entity"
	"github.com/devldavydov/myhealth-ng/internal/port"
)

type ActiveCalories struct {
	repository port.ActiveCaloriesRepository
}

func NewActiveCalories(repository port.ActiveCaloriesRepository) *ActiveCalories {
	return &ActiveCalories{repository: repository}
}

func (activeCalories *ActiveCalories) Get(ctx context.Context, userID string, dt time.Time) (*entity.ActiveCalories, error) {
	userID, dt, err := validateActiveCaloriesScope(userID, dt)
	if err != nil {
		return nil, err
	}
	return activeCalories.repository.Get(ctx, userID, dt)
}

func (activeCalories *ActiveCalories) Save(ctx context.Context, userID string, data entity.ActiveCalories) (entity.ActiveCalories, error) {
	userID, dt, err := validateActiveCaloriesScope(userID, data.DT)
	if err != nil {
		return entity.ActiveCalories{}, err
	}
	if math.IsNaN(data.Value) || math.IsInf(data.Value, 0) || data.Value <= 0 {
		return entity.ActiveCalories{}, &entity.ValidationError{Fields: map[string][]string{
			"value": {"Ожидается конечное положительное число"},
		}}
	}
	data.DT = dt
	return activeCalories.repository.Save(ctx, userID, data)
}

func (activeCalories *ActiveCalories) Delete(ctx context.Context, userID string, dt time.Time) error {
	userID, dt, err := validateActiveCaloriesScope(userID, dt)
	if err != nil {
		return err
	}
	return activeCalories.repository.Delete(ctx, userID, dt)
}

func validateActiveCaloriesScope(userID string, dt time.Time) (string, time.Time, error) {
	userID = strings.TrimSpace(userID)
	details := make(map[string][]string)
	validateUserID(details, userID)
	if dt.IsZero() {
		details["dt"] = []string{"Дата обязательна"}
	}
	if len(details) > 0 {
		return "", time.Time{}, &entity.ValidationError{Fields: details}
	}
	return userID, dateOnly(dt), nil
}
