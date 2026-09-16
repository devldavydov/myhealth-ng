package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/devldavydov/myhealth-ng/internal/entity"
)

type SettingsRepository struct {
	db *sql.DB
}

func NewSettingsRepository(db *sql.DB) *SettingsRepository {
	return &SettingsRepository{db: db}
}

func (repository *SettingsRepository) Get(ctx context.Context, userID string) (*entity.UserSettings, error) {
	var settings entity.UserSettings
	err := repository.db.QueryRowContext(ctx, `
SELECT default_daily_calorie_limit
FROM user_settings
WHERE user_id = $1`, userID).Scan(&settings.DefaultDailyCalorieLimit)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get user settings: %w", err)
	}
	return &settings, nil
}

func (repository *SettingsRepository) Save(ctx context.Context, userID string, settings entity.UserSettings) (entity.UserSettings, error) {
	var saved entity.UserSettings
	err := repository.db.QueryRowContext(ctx, `
INSERT INTO user_settings (user_id, default_daily_calorie_limit)
VALUES ($1, $2)
ON CONFLICT (user_id) DO UPDATE
SET default_daily_calorie_limit = EXCLUDED.default_daily_calorie_limit
RETURNING default_daily_calorie_limit`, userID, settings.DefaultDailyCalorieLimit).Scan(&saved.DefaultDailyCalorieLimit)
	if err != nil {
		return entity.UserSettings{}, fmt.Errorf("save user settings: %w", err)
	}
	return saved, nil
}
