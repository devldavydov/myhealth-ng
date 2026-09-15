package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/devldavydov/myhealth-ng/internal/entity"
	"github.com/devldavydov/myhealth-ng/internal/port"
)

const listBundlesQuery = `
SELECT b.key, b.name, COUNT(*)::integer,
       COALESCE(SUM(bi.weight), 0)::double precision,
       COALESCE(SUM(f.cal100 * bi.weight / 100), 0)::double precision,
       COALESCE(SUM(f.prot100 * bi.weight / 100), 0)::double precision,
       COALESCE(SUM(f.fat100 * bi.weight / 100), 0)::double precision,
       COALESCE(SUM(f.carb100 * bi.weight / 100), 0)::double precision
FROM bundle b
JOIN bundle_item bi ON bi.bundle_key = b.key
JOIN food f ON f.key = bi.food_key
GROUP BY b.key, b.name
ORDER BY lower(b.name), b.key`

const searchBundlesQuery = `
SELECT b.key, b.name, COUNT(*)::integer,
       COALESCE(SUM(bi.weight), 0)::double precision,
       COALESCE(SUM(f.cal100 * bi.weight / 100), 0)::double precision,
       COALESCE(SUM(f.prot100 * bi.weight / 100), 0)::double precision,
       COALESCE(SUM(f.fat100 * bi.weight / 100), 0)::double precision,
       COALESCE(SUM(f.carb100 * bi.weight / 100), 0)::double precision
FROM bundle b
JOIN bundle_item bi ON bi.bundle_key = b.key
JOIN food f ON f.key = bi.food_key
WHERE b.name ILIKE $1 ESCAPE E'\\'
GROUP BY b.key, b.name
ORDER BY lower(b.name), b.key`

type BundleRepository struct {
	db *sql.DB
}

func NewBundleRepository(db *sql.DB) *BundleRepository {
	return &BundleRepository{db: db}
}

func (repository *BundleRepository) List(ctx context.Context, query string) ([]entity.BundleSummary, error) {
	var (
		rows *sql.Rows
		err  error
	)
	if query == "" {
		rows, err = repository.db.QueryContext(ctx, listBundlesQuery)
	} else {
		rows, err = repository.db.QueryContext(ctx, searchBundlesQuery, "%"+escapeLike(query)+"%")
	}
	if err != nil {
		return nil, fmt.Errorf("list bundles: %w", err)
	}
	defer rows.Close()

	items := make([]entity.BundleSummary, 0)
	for rows.Next() {
		var item entity.BundleSummary
		err := rows.Scan(&item.Key, &item.Name, &item.ItemCount, &item.Totals.Weight, &item.Totals.Cal, &item.Totals.Protein, &item.Totals.Fat, &item.Totals.Carb)
		if err != nil {
			return nil, fmt.Errorf("scan bundle list: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate bundle list: %w", err)
	}
	return items, nil
}

func (repository *BundleRepository) Get(ctx context.Context, key string) (entity.Bundle, error) {
	rows, err := repository.db.QueryContext(ctx, `
SELECT b.key, b.name,
       f.key, f.name, f.brand, f.cal100, f.prot100, f.fat100, f.carb100, f.comment,
       bi.weight
FROM bundle b
JOIN bundle_item bi ON bi.bundle_key = b.key
JOIN food f ON f.key = bi.food_key
WHERE b.key = $1
ORDER BY lower(f.name), lower(f.brand), f.key`, key)
	if err != nil {
		return entity.Bundle{}, fmt.Errorf("get bundle: %w", err)
	}
	defer rows.Close()

	var bundle entity.Bundle
	for rows.Next() {
		var item entity.BundleItem
		if err := rows.Scan(&bundle.Key, &bundle.Name, &item.Food.Key, &item.Food.Name, &item.Food.Brand, &item.Food.Cal100, &item.Food.Prot100, &item.Food.Fat100, &item.Food.Carb100, &item.Food.Comment, &item.Weight); err != nil {
			return entity.Bundle{}, fmt.Errorf("scan bundle: %w", err)
		}
		bundle.Items = append(bundle.Items, item)
		addBundleTotals(&bundle.Totals, item)
	}
	if err := rows.Err(); err != nil {
		return entity.Bundle{}, fmt.Errorf("iterate bundle: %w", err)
	}
	if bundle.Key == "" {
		return entity.Bundle{}, port.ErrBundleNotFound
	}
	return bundle, nil
}

func (repository *BundleRepository) Create(ctx context.Context, bundle entity.Bundle) (entity.Bundle, error) {
	transaction, err := repository.db.BeginTx(ctx, nil)
	if err != nil {
		return entity.Bundle{}, fmt.Errorf("begin create bundle: %w", err)
	}
	defer transaction.Rollback()
	if _, err := transaction.ExecContext(ctx, `INSERT INTO bundle (key, name) VALUES ($1, $2)`, bundle.Key, bundle.Name); err != nil {
		return entity.Bundle{}, mapBundleQueryError("create bundle", err)
	}
	if err := insertBundleItems(ctx, transaction, bundle.Key, bundle.Items); err != nil {
		return entity.Bundle{}, err
	}
	if err := transaction.Commit(); err != nil {
		return entity.Bundle{}, fmt.Errorf("commit create bundle: %w", err)
	}
	return repository.Get(ctx, bundle.Key)
}

func (repository *BundleRepository) Update(ctx context.Context, key string, data entity.BundleData) (entity.Bundle, error) {
	transaction, err := repository.db.BeginTx(ctx, nil)
	if err != nil {
		return entity.Bundle{}, fmt.Errorf("begin update bundle: %w", err)
	}
	defer transaction.Rollback()
	result, err := transaction.ExecContext(ctx, `UPDATE bundle SET name = $2 WHERE key = $1`, key, data.Name)
	if err != nil {
		return entity.Bundle{}, fmt.Errorf("update bundle: %w", err)
	}
	updated, err := result.RowsAffected()
	if err != nil {
		return entity.Bundle{}, fmt.Errorf("update bundle result: %w", err)
	}
	if updated == 0 {
		return entity.Bundle{}, port.ErrBundleNotFound
	}
	if _, err := transaction.ExecContext(ctx, `DELETE FROM bundle_item WHERE bundle_key = $1`, key); err != nil {
		return entity.Bundle{}, fmt.Errorf("clear bundle items: %w", err)
	}
	if err := insertBundleItemData(ctx, transaction, key, data.Items); err != nil {
		return entity.Bundle{}, err
	}
	if err := transaction.Commit(); err != nil {
		return entity.Bundle{}, fmt.Errorf("commit update bundle: %w", err)
	}
	return repository.Get(ctx, key)
}

func (repository *BundleRepository) Delete(ctx context.Context, key string) error {
	result, err := repository.db.ExecContext(ctx, `DELETE FROM bundle WHERE key = $1`, key)
	if err != nil {
		return fmt.Errorf("delete bundle: %w", err)
	}
	deleted, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete bundle result: %w", err)
	}
	if deleted == 0 {
		return port.ErrBundleNotFound
	}
	return nil
}

func insertBundleItems(ctx context.Context, transaction *sql.Tx, bundleKey string, items []entity.BundleItem) error {
	data := make([]entity.BundleItemData, 0, len(items))
	for _, item := range items {
		data = append(data, entity.BundleItemData{FoodKey: item.Food.Key, Weight: item.Weight})
	}
	return insertBundleItemData(ctx, transaction, bundleKey, data)
}

func insertBundleItemData(ctx context.Context, transaction *sql.Tx, bundleKey string, items []entity.BundleItemData) error {
	statement, err := transaction.PrepareContext(ctx, `INSERT INTO bundle_item (bundle_key, food_key, weight) VALUES ($1, $2, $3)`)
	if err != nil {
		return fmt.Errorf("prepare bundle items: %w", err)
	}
	defer statement.Close()
	for _, item := range items {
		if _, err := statement.ExecContext(ctx, bundleKey, item.FoodKey, item.Weight); err != nil {
			return mapBundleQueryError("insert bundle item", err)
		}
	}
	return nil
}

func mapBundleQueryError(operation string, err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return port.ErrBundleNotFound
	}
	var postgresError *pgconn.PgError
	if errors.As(err, &postgresError) {
		switch postgresError.Code {
		case "23503":
			return port.ErrBundleFoodNotFound
		case "23505":
			return port.ErrBundleConflict
		}
	}
	return fmt.Errorf("%s: %w", operation, err)
}

func addBundleTotals(totals *entity.BundleTotals, item entity.BundleItem) {
	multiplier := item.Weight / 100
	totals.Weight += item.Weight
	totals.Cal += item.Food.Cal100 * multiplier
	totals.Protein += item.Food.Prot100 * multiplier
	totals.Fat += item.Food.Fat100 * multiplier
	totals.Carb += item.Food.Carb100 * multiplier
}
