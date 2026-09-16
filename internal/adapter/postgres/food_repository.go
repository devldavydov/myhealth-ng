package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/devldavydov/myhealth-ng/internal/entity"
	"github.com/devldavydov/myhealth-ng/internal/port"
)

const countFoodQuery = `SELECT COUNT(*) FROM food`

const countSearchFoodQuery = `
SELECT COUNT(*)
FROM food
WHERE name ILIKE $1 ESCAPE E'\\'
   OR brand ILIKE $1 ESCAPE E'\\'`

const listFoodQuery = `
SELECT key, name, brand, cal100, prot100, fat100, carb100, comment
FROM food
ORDER BY lower(name), lower(brand), key
LIMIT $1 OFFSET $2`

const searchFoodQuery = `
SELECT key, name, brand, cal100, prot100, fat100, carb100, comment
FROM food
WHERE name ILIKE $1 ESCAPE E'\\'
   OR brand ILIKE $1 ESCAPE E'\\'
ORDER BY lower(name), lower(brand), key
LIMIT $2 OFFSET $3`

type FoodRepository struct {
	db *sql.DB
}

func NewFoodRepository(db *sql.DB) *FoodRepository {
	return &FoodRepository{db: db}
}

func (repository *FoodRepository) List(ctx context.Context, request entity.PageRequest) (entity.Page[entity.Food], error) {
	var total int
	var err error
	pattern := "%" + escapeLike(request.Query) + "%"
	if request.Query == "" {
		err = repository.db.QueryRowContext(ctx, countFoodQuery).Scan(&total)
	} else {
		err = repository.db.QueryRowContext(ctx, countSearchFoodQuery, pattern).Scan(&total)
	}
	if err != nil {
		return entity.Page[entity.Food]{}, fmt.Errorf("count food: %w", err)
	}

	offset := int64(request.Page-1) * int64(request.PageSize)
	var rows *sql.Rows
	if request.Query == "" {
		rows, err = repository.db.QueryContext(ctx, listFoodQuery, request.PageSize, offset)
	} else {
		rows, err = repository.db.QueryContext(ctx, searchFoodQuery, pattern, request.PageSize, offset)
	}
	if err != nil {
		return entity.Page[entity.Food]{}, fmt.Errorf("list food: %w", err)
	}
	defer rows.Close()

	items := make([]entity.Food, 0)
	for rows.Next() {
		item, err := scanFood(rows)
		if err != nil {
			return entity.Page[entity.Food]{}, fmt.Errorf("scan food list: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return entity.Page[entity.Food]{}, fmt.Errorf("iterate food list: %w", err)
	}
	return entity.Page[entity.Food]{
		Items: items, Page: request.Page, PageSize: request.PageSize, Total: total,
		TotalPages: calculateTotalPages(total, request.PageSize),
	}, nil
}

func (repository *FoodRepository) Get(ctx context.Context, key string) (entity.Food, error) {
	item, err := scanFood(repository.db.QueryRowContext(ctx, `
SELECT key, name, brand, cal100, prot100, fat100, carb100, comment
FROM food
WHERE key = $1`, key))
	return item, mapQueryError("get food", err)
}

func (repository *FoodRepository) Create(ctx context.Context, food entity.Food) (entity.Food, error) {
	item, err := scanFood(repository.db.QueryRowContext(ctx, `
INSERT INTO food (key, name, brand, cal100, prot100, fat100, carb100, comment)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING key, name, brand, cal100, prot100, fat100, carb100, comment`,
		food.Key, food.Name, food.Brand, food.Cal100, food.Prot100, food.Fat100, food.Carb100, food.Comment))
	return item, mapQueryError("create food", err)
}

func (repository *FoodRepository) Update(ctx context.Context, key string, data entity.FoodData) (entity.Food, error) {
	item, err := scanFood(repository.db.QueryRowContext(ctx, `
UPDATE food
SET name = $2, brand = $3, cal100 = $4, prot100 = $5, fat100 = $6, carb100 = $7, comment = $8
WHERE key = $1
RETURNING key, name, brand, cal100, prot100, fat100, carb100, comment`,
		key, data.Name, data.Brand, data.Cal100, data.Prot100, data.Fat100, data.Carb100, data.Comment))
	return item, mapQueryError("update food", err)
}

func (repository *FoodRepository) Delete(ctx context.Context, key string) error {
	result, err := repository.db.ExecContext(ctx, `DELETE FROM food WHERE key = $1`, key)
	if err != nil {
		var postgresError *pgconn.PgError
		if errors.As(err, &postgresError) && (postgresError.Code == "23001" || postgresError.Code == "23503") {
			return port.ErrFoodInUse
		}
		return fmt.Errorf("delete food: %w", err)
	}
	deleted, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete food result: %w", err)
	}
	if deleted == 0 {
		return port.ErrFoodNotFound
	}
	return nil
}

type scanner interface {
	Scan(...any) error
}

func scanFood(row scanner) (entity.Food, error) {
	var food entity.Food
	err := row.Scan(&food.Key, &food.Name, &food.Brand, &food.Cal100, &food.Prot100, &food.Fat100, &food.Carb100, &food.Comment)
	return food, err
}

func mapQueryError(operation string, err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return port.ErrFoodNotFound
	}
	var postgresError *pgconn.PgError
	if errors.As(err, &postgresError) && postgresError.Code == "23505" {
		return port.ErrFoodConflict
	}
	return fmt.Errorf("%s: %w", operation, err)
}

func escapeLike(value string) string {
	replacer := strings.NewReplacer(`\`, `\\`, "%", `\%`, "_", `\_`)
	return replacer.Replace(value)
}
