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

const listSportQuery = `SELECT key, name, unit, comment FROM sport ORDER BY lower(name), key LIMIT $1 OFFSET $2`
const searchSportQuery = `SELECT key, name, unit, comment FROM sport WHERE name ILIKE $1 ESCAPE E'\\' ORDER BY lower(name), key LIMIT $2 OFFSET $3`

type SportRepository struct{ db *sql.DB }

func NewSportRepository(db *sql.DB) *SportRepository { return &SportRepository{db: db} }
func (repository *SportRepository) List(ctx context.Context, request entity.PageRequest) (entity.Page[entity.Sport], error) {
	var total int
	var err error
	pattern := "%" + escapeLike(request.Query) + "%"
	if request.Query == "" {
		err = repository.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM sport`).Scan(&total)
	} else {
		err = repository.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM sport WHERE name ILIKE $1 ESCAPE E'\\'`, pattern).Scan(&total)
	}
	if err != nil {
		return entity.Page[entity.Sport]{}, fmt.Errorf("count sport: %w", err)
	}
	offset := int64(request.Page-1) * int64(request.PageSize)
	var rows *sql.Rows
	if request.Query == "" {
		rows, err = repository.db.QueryContext(ctx, listSportQuery, request.PageSize, offset)
	} else {
		rows, err = repository.db.QueryContext(ctx, searchSportQuery, pattern, request.PageSize, offset)
	}
	if err != nil {
		return entity.Page[entity.Sport]{}, fmt.Errorf("list sport: %w", err)
	}
	defer rows.Close()
	items := make([]entity.Sport, 0)
	for rows.Next() {
		item, err := scanSport(rows)
		if err != nil {
			return entity.Page[entity.Sport]{}, fmt.Errorf("scan sport list: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return entity.Page[entity.Sport]{}, fmt.Errorf("iterate sport list: %w", err)
	}
	return entity.Page[entity.Sport]{Items: items, Page: request.Page, PageSize: request.PageSize, Total: total, TotalPages: calculateTotalPages(total, request.PageSize)}, nil
}
func (repository *SportRepository) Get(ctx context.Context, key string) (entity.Sport, error) {
	item, err := scanSport(repository.db.QueryRowContext(ctx, `SELECT key, name, unit, comment FROM sport WHERE key = $1`, key))
	return item, mapSportQueryError("get sport", err)
}
func (repository *SportRepository) Create(ctx context.Context, sport entity.Sport) (entity.Sport, error) {
	item, err := scanSport(repository.db.QueryRowContext(ctx, `INSERT INTO sport (key,name,unit,comment) VALUES ($1,$2,$3,$4) RETURNING key,name,unit,comment`, sport.Key, sport.Name, sport.Unit, sport.Comment))
	return item, mapSportQueryError("create sport", err)
}
func (repository *SportRepository) Update(ctx context.Context, key string, data entity.SportData) (entity.Sport, error) {
	item, err := scanSport(repository.db.QueryRowContext(ctx, `UPDATE sport SET name=$2,unit=$3,comment=$4 WHERE key=$1 RETURNING key,name,unit,comment`, key, data.Name, data.Unit, data.Comment))
	return item, mapSportQueryError("update sport", err)
}
func (repository *SportRepository) Delete(ctx context.Context, key string) error {
	result, err := repository.db.ExecContext(ctx, `DELETE FROM sport WHERE key=$1`, key)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return port.ErrSportInUse
		}
		return fmt.Errorf("delete sport: %w", err)
	}
	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete sport result: %w", err)
	}
	if count == 0 {
		return port.ErrSportNotFound
	}
	return nil
}
func scanSport(row scanner) (entity.Sport, error) {
	var item entity.Sport
	err := row.Scan(&item.Key, &item.Name, &item.Unit, &item.Comment)
	return item, err
}
func mapSportQueryError(operation string, err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, sql.ErrNoRows) {
		return port.ErrSportNotFound
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return port.ErrSportConflict
	}
	return fmt.Errorf("%s: %w", operation, err)
}
