package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"

	"github.com/devldavydov/myhealth-ng/internal/entity"
	"github.com/devldavydov/myhealth-ng/internal/port"
)

type SportActivityRepository struct{ db *sql.DB }

func NewSportActivityRepository(db *sql.DB) *SportActivityRepository {
	return &SportActivityRepository{db: db}
}
func (repository *SportActivityRepository) List(ctx context.Context, userID string, period entity.SportActivityRange) ([]entity.SportActivity, error) {
	rows, err := repository.db.QueryContext(ctx, `SELECT a.dt,s.key,s.name,s.unit,s.comment,array_to_json(a.sets)::text FROM sport_activity a JOIN sport s ON s.key=a.sport_key WHERE a.user_id=$1 AND ($2::date IS NULL OR a.dt >= $2::date) AND ($3::date IS NULL OR a.dt <= $3::date) ORDER BY a.dt DESC, lower(s.name), s.key`, userID, nullableDate(period.From), nullableDate(period.To))
	if err != nil {
		return nil, fmt.Errorf("list sport activity: %w", err)
	}
	defer rows.Close()
	items := make([]entity.SportActivity, 0)
	for rows.Next() {
		item, err := scanSportActivity(rows)
		if err != nil {
			return nil, fmt.Errorf("scan sport activity: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate sport activity: %w", err)
	}
	return items, nil
}
func (repository *SportActivityRepository) Save(ctx context.Context, userID string, data entity.SportActivityData) (entity.SportActivity, error) {
	item, err := scanSportActivity(repository.db.QueryRowContext(ctx, `INSERT INTO sport_activity (user_id,dt,sport_key,sets) VALUES ($1,$2,$3,$4) ON CONFLICT (user_id,dt,sport_key) DO UPDATE SET sets=EXCLUDED.sets RETURNING dt,(SELECT key FROM sport WHERE key=sport_key),(SELECT name FROM sport WHERE key=sport_key),(SELECT unit FROM sport WHERE key=sport_key),(SELECT comment FROM sport WHERE key=sport_key),array_to_json(sets)::text`, userID, data.DT, data.SportKey, data.Sets))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.SportActivity{}, port.ErrSportNotFound
		}
		var postgresError *pgconn.PgError
		if errors.As(err, &postgresError) && postgresError.Code == "23503" {
			return entity.SportActivity{}, port.ErrSportNotFound
		}
		return entity.SportActivity{}, fmt.Errorf("save sport activity: %w", err)
	}
	return item, nil
}
func (repository *SportActivityRepository) Delete(ctx context.Context, userID string, data entity.SportActivityData) error {
	result, err := repository.db.ExecContext(ctx, `DELETE FROM sport_activity WHERE user_id=$1 AND dt=$2 AND sport_key=$3`, userID, data.DT, data.SportKey)
	if err != nil {
		return fmt.Errorf("delete sport activity: %w", err)
	}
	count, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("delete sport activity result: %w", err)
	}
	if count == 0 {
		return port.ErrSportActivityNotFound
	}
	return nil
}
func scanSportActivity(row scanner) (entity.SportActivity, error) {
	var item entity.SportActivity
	var setsJSON string
	if err := row.Scan(&item.DT, &item.Sport.Key, &item.Sport.Name, &item.Sport.Unit, &item.Sport.Comment, &setsJSON); err != nil {
		return item, err
	}
	if err := json.Unmarshal([]byte(setsJSON), &item.Sets); err != nil {
		return item, fmt.Errorf("decode sport activity sets: %w", err)
	}
	return item, nil
}
