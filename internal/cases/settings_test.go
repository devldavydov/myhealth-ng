package cases

import (
	"context"
	"errors"
	"testing"

	"github.com/devldavydov/myhealth-ng/internal/entity"
)

type settingsRepositoryStub struct {
	userID string
	item   *entity.UserSettings
	saved  entity.UserSettings
}

func (stub *settingsRepositoryStub) Get(_ context.Context, userID string) (*entity.UserSettings, error) {
	stub.userID = userID
	return stub.item, nil
}

func (stub *settingsRepositoryStub) Save(_ context.Context, userID string, data entity.UserSettings) (entity.UserSettings, error) {
	stub.userID, stub.saved = userID, data
	return data, nil
}

func TestSettingsGetAllowsMissingRow(t *testing.T) {
	repository := &settingsRepositoryStub{}
	settings, err := NewSettings(repository).Get(context.Background(), " user-1 ")
	if err != nil || settings != nil || repository.userID != "user-1" {
		t.Fatalf("settings=%+v user=%q err=%v", settings, repository.userID, err)
	}
}

func TestSettingsSaveValidatesAndPersistsLimits(t *testing.T) {
	for _, test := range []struct {
		name  string
		value int
	}{
		{"minimum", MinDailyCalorieLimit},
		{"maximum", MaxDailyCalorieLimit},
	} {
		t.Run(test.name, func(t *testing.T) {
			repository := &settingsRepositoryStub{}
			saved, err := NewSettings(repository).Save(context.Background(), "user-1", entity.UserSettings{DefaultDailyCalorieLimit: test.value})
			if err != nil || saved.DefaultDailyCalorieLimit != test.value || repository.saved.DefaultDailyCalorieLimit != test.value {
				t.Fatalf("saved=%+v repository=%+v err=%v", saved, repository.saved, err)
			}
		})
	}

	for _, test := range []struct {
		name   string
		userID string
		value  int
	}{
		{"missing user", "", 2000},
		{"zero", "user-1", 0},
		{"too high", "user-1", 10001},
	} {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewSettings(&settingsRepositoryStub{}).Save(context.Background(), test.userID, entity.UserSettings{DefaultDailyCalorieLimit: test.value})
			var validationError *entity.ValidationError
			if !errors.As(err, &validationError) {
				t.Fatalf("expected validation error, got %v", err)
			}
		})
	}
}

func TestSettingsGetRejectsMissingUser(t *testing.T) {
	_, err := NewSettings(&settingsRepositoryStub{}).Get(context.Background(), "  ")
	var validationError *entity.ValidationError
	if !errors.As(err, &validationError) {
		t.Fatalf("expected validation error, got %v", err)
	}
}
