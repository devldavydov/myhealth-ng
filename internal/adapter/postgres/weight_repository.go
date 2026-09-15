package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/devldavydov/myhealth-ng/internal/entity"
	"github.com/devldavydov/myhealth-ng/internal/port"
)

const listWeightQuery = `
SELECT dt, value
FROM weight
WHERE user_id = $1
  AND ($2::date IS NULL OR dt >= $2::date)
  AND ($3::date IS NULL OR dt <= $3::date)
ORDER BY dt DESC`

type WeightRepository struct {
	db *sql.DB
}

func NewWeightRepository(db *sql.DB) *WeightRepository {
	return &WeightRepository{db: db}
}

func (repository *WeightRepository) List(ctx context.Context, userID string, period entity.WeightRange) ([]entity.Weight, error) {
	rows, err := repository.db.QueryContext(ctx, listWeightQuery, userID, nullableDate(period.From), nullableDate(period.To))
	if err != nil {
		return nil, fmt.Errorf("list weight: %w", err)
	}
	defer rows.Close()

	items := make([]entity.Weight, 0)
	for rows.Next() {
		item, err := scanWeight(rows)
		if err != nil {
			return nil, fmt.Errorf("scan weight list: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate weight list: %w", err)
	}
	return items, nil
}

func (repository *WeightRepository) Save(ctx context.Context, userID string, weight entity.Weight) (entity.Weight, error) {
	item, err := scanWeight(repository.db.QueryRowContext(ctx, `
INSERT INTO weight (user_id, dt, value)
VALUES ($1, $2, $3)
ON CONFLICT (user_id, dt) DO UPDATE SET value = EXCLUDED.value
RETURNING dt, value`, userID, weight.DT, weight.Value))
	if err != nil {
		return entity.Weight{}, fmt.Errorf("save weight: %w", err)
	}
	return item, nil
}

func (repository *WeightRepository) Delete(ctx context.Context, userID string, weight entity.Weight) error {
	result, err := repository.db.ExecContext(ctx, `DELETE FROM weight WHERE user_id = $1 AND dt = $2`, userID, weight.DT)
	if err != nil {
		return fmt.Errorf("delete weight: %w", err)
	}
	deleted, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete weight result: %w", err)
	}
	if deleted == 0 {
		return port.ErrWeightNotFound
	}
	return nil
}

func scanWeight(row scanner) (entity.Weight, error) {
	var weight entity.Weight
	err := row.Scan(&weight.DT, &weight.Value)
	return weight, err
}

func nullableDate(value *time.Time) any {
	if value == nil {
		return nil
	}
	return *value
}
