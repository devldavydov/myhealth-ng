package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/devldavydov/myhealth-ng/internal/entity"
	"github.com/devldavydov/myhealth-ng/internal/port"
)

type ActiveCaloriesRepository struct {
	db *sql.DB
}

func NewActiveCaloriesRepository(db *sql.DB) *ActiveCaloriesRepository {
	return &ActiveCaloriesRepository{db: db}
}

func (repository *ActiveCaloriesRepository) Get(ctx context.Context, userID string, dt time.Time) (*entity.ActiveCalories, error) {
	item, err := scanActiveCalories(repository.db.QueryRowContext(ctx, `
SELECT dt, value
FROM act_calories
WHERE user_id = $1 AND dt = $2`, userID, dt))
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get active calories: %w", err)
	}
	return &item, nil
}

func (repository *ActiveCaloriesRepository) Save(ctx context.Context, userID string, data entity.ActiveCalories) (entity.ActiveCalories, error) {
	item, err := scanActiveCalories(repository.db.QueryRowContext(ctx, `
INSERT INTO act_calories (user_id, dt, value)
VALUES ($1, $2, $3)
ON CONFLICT (user_id, dt) DO UPDATE SET value = EXCLUDED.value
RETURNING dt, value`, userID, data.DT, data.Value))
	if err != nil {
		return entity.ActiveCalories{}, fmt.Errorf("save active calories: %w", err)
	}
	return item, nil
}

func (repository *ActiveCaloriesRepository) Delete(ctx context.Context, userID string, dt time.Time) error {
	result, err := repository.db.ExecContext(ctx, `DELETE FROM act_calories WHERE user_id = $1 AND dt = $2`, userID, dt)
	if err != nil {
		return fmt.Errorf("delete active calories: %w", err)
	}
	deleted, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete active calories result: %w", err)
	}
	if deleted == 0 {
		return port.ErrActiveCaloriesNotFound
	}
	return nil
}

func scanActiveCalories(row scanner) (entity.ActiveCalories, error) {
	var item entity.ActiveCalories
	err := row.Scan(&item.DT, &item.Value)
	return item, err
}
