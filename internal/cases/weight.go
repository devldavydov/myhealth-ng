package cases

import (
	"context"
	"math"
	"strings"
	"time"

	"github.com/devldavydov/myhealth-ng/internal/entity"
	"github.com/devldavydov/myhealth-ng/internal/port"
)

type Weight struct {
	repository port.WeightRepository
}

func NewWeight(repository port.WeightRepository) *Weight {
	return &Weight{repository: repository}
}

func (weight *Weight) List(ctx context.Context, userID string, period entity.WeightRange) ([]entity.Weight, error) {
	userID, period, err := validateWeightRange(userID, period)
	if err != nil {
		return nil, err
	}
	return weight.repository.List(ctx, userID, period)
}

func (weight *Weight) Save(ctx context.Context, userID string, data entity.Weight) (entity.Weight, error) {
	userID = strings.TrimSpace(userID)
	details := make(map[string][]string)
	validateUserID(details, userID)
	if data.DT.IsZero() {
		details["dt"] = []string{"Дата обязательна"}
	}
	if math.IsNaN(data.Value) || math.IsInf(data.Value, 0) || data.Value <= 0 {
		details["value"] = []string{"Ожидается конечное положительное число"}
	}
	if len(details) > 0 {
		return entity.Weight{}, &entity.ValidationError{Fields: details}
	}
	data.DT = dateOnly(data.DT)
	return weight.repository.Save(ctx, userID, data)
}

func (weight *Weight) Delete(ctx context.Context, userID string, data entity.Weight) error {
	userID = strings.TrimSpace(userID)
	details := make(map[string][]string)
	validateUserID(details, userID)
	if data.DT.IsZero() {
		details["dt"] = []string{"Дата обязательна"}
	}
	if len(details) > 0 {
		return &entity.ValidationError{Fields: details}
	}
	data.DT = dateOnly(data.DT)
	return weight.repository.Delete(ctx, userID, data)
}

func validateWeightRange(userID string, period entity.WeightRange) (string, entity.WeightRange, error) {
	userID = strings.TrimSpace(userID)
	details := make(map[string][]string)
	validateUserID(details, userID)
	if period.From != nil {
		value := dateOnly(*period.From)
		period.From = &value
	}
	if period.To != nil {
		value := dateOnly(*period.To)
		period.To = &value
	}
	if period.From != nil && period.To != nil && period.From.After(*period.To) {
		details["from"] = []string{"Начальная дата должна быть не позже конечной"}
	}
	if len(details) > 0 {
		return "", entity.WeightRange{}, &entity.ValidationError{Fields: details}
	}
	return userID, period, nil
}

func validateUserID(details map[string][]string, userID string) {
	if userID == "" {
		details["userId"] = []string{"Пользователь не определён"}
	}
}

func dateOnly(value time.Time) time.Time {
	return time.Date(value.Year(), value.Month(), value.Day(), 0, 0, 0, 0, time.UTC)
}
