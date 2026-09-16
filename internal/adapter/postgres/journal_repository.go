package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/devldavydov/myhealth-ng/internal/entity"
	"github.com/devldavydov/myhealth-ng/internal/port"
)

type JournalRepository struct {
	db *sql.DB
}

func NewJournalRepository(db *sql.DB) *JournalRepository {
	return &JournalRepository{db: db}
}

func (repository *JournalRepository) List(ctx context.Context, userID string, dt time.Time) ([]entity.JournalItem, error) {
	rows, err := repository.db.QueryContext(ctx, `
SELECT j.meal, f.key, f.name, f.brand, f.cal100, f.prot100, f.fat100, f.carb100, f.comment, j.food_weight
FROM journal j
JOIN food f ON f.key = j.food_key
WHERE j.user_id = $1 AND j.dt = $2
ORDER BY j.meal, lower(f.name), lower(f.brand), f.key`, userID, dt)
	if err != nil {
		return nil, fmt.Errorf("list journal: %w", err)
	}
	defer rows.Close()

	items := make([]entity.JournalItem, 0)
	for rows.Next() {
		var item entity.JournalItem
		if err := rows.Scan(&item.Meal, &item.Food.Key, &item.Food.Name, &item.Food.Brand, &item.Food.Cal100,
			&item.Food.Prot100, &item.Food.Fat100, &item.Food.Carb100, &item.Food.Comment, &item.Weight); err != nil {
			return nil, fmt.Errorf("scan journal: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate journal: %w", err)
	}
	return items, nil
}

func (repository *JournalRepository) Save(ctx context.Context, userID string, dt time.Time, meal entity.MealType, items []entity.JournalItemData) error {
	transaction, err := repository.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin journal save: %w", err)
	}
	defer transaction.Rollback()
	statement, err := transaction.PrepareContext(ctx, `
INSERT INTO journal (user_id, dt, meal, food_key, food_weight)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (user_id, dt, meal, food_key) DO UPDATE
SET food_weight = EXCLUDED.food_weight`)
	if err != nil {
		return fmt.Errorf("prepare journal save: %w", err)
	}
	defer statement.Close()
	for _, item := range items {
		if _, err := statement.ExecContext(ctx, userID, dt, meal, item.FoodKey, item.Weight); err != nil {
			var postgresError *pgconn.PgError
			if errors.As(err, &postgresError) && postgresError.Code == "23503" {
				return port.ErrJournalFoodNotFound
			}
			return fmt.Errorf("save journal item: %w", err)
		}
	}
	if err := transaction.Commit(); err != nil {
		return fmt.Errorf("commit journal save: %w", err)
	}
	return nil
}

func (repository *JournalRepository) DeleteItem(ctx context.Context, userID string, dt time.Time, meal entity.MealType, foodKey string) error {
	result, err := repository.db.ExecContext(ctx, `
DELETE FROM journal
WHERE user_id = $1 AND dt = $2 AND meal = $3 AND food_key = $4`, userID, dt, meal, foodKey)
	if err != nil {
		return fmt.Errorf("delete journal item: %w", err)
	}
	deleted, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete journal item result: %w", err)
	}
	if deleted == 0 {
		return port.ErrJournalItemNotFound
	}
	return nil
}

func (repository *JournalRepository) ClearMeal(ctx context.Context, userID string, dt time.Time, meal entity.MealType) error {
	if _, err := repository.db.ExecContext(ctx, `
DELETE FROM journal
WHERE user_id = $1 AND dt = $2 AND meal = $3`, userID, dt, meal); err != nil {
		return fmt.Errorf("clear journal meal: %w", err)
	}
	return nil
}
