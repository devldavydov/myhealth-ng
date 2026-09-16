package cases

import (
	"context"
	"strings"

	"github.com/devldavydov/myhealth-ng/internal/entity"
	"github.com/devldavydov/myhealth-ng/internal/port"
)

const (
	MinDailyCalorieLimit = 1
	MaxDailyCalorieLimit = 10000
)

type Settings struct {
	repository port.SettingsRepository
}

func NewSettings(repository port.SettingsRepository) *Settings {
	return &Settings{repository: repository}
}

func (settings *Settings) Get(ctx context.Context, userID string) (*entity.UserSettings, error) {
	userID = strings.TrimSpace(userID)
	if err := validateSettingsUserID(userID); err != nil {
		return nil, err
	}
	return settings.repository.Get(ctx, userID)
}

func (settings *Settings) Save(ctx context.Context, userID string, data entity.UserSettings) (entity.UserSettings, error) {
	userID = strings.TrimSpace(userID)
	details := make(map[string][]string)
	if userID == "" {
		details["userId"] = []string{"Пользователь не определён"}
	}
	if data.DefaultDailyCalorieLimit < MinDailyCalorieLimit || data.DefaultDailyCalorieLimit > MaxDailyCalorieLimit {
		details["defaultDailyCalorieLimit"] = []string{"Введите целое число от 1 до 10000"}
	}
	if len(details) > 0 {
		return entity.UserSettings{}, &entity.ValidationError{Fields: details}
	}
	return settings.repository.Save(ctx, userID, data)
}

func validateSettingsUserID(userID string) error {
	if userID != "" {
		return nil
	}
	return &entity.ValidationError{Fields: map[string][]string{
		"userId": {"Пользователь не определён"},
	}}
}
