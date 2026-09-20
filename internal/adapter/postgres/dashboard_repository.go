package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/devldavydov/myhealth-ng/internal/entity"
)

type DashboardRepository struct {
	db *sql.DB
}

func NewDashboardRepository(db *sql.DB) *DashboardRepository {
	return &DashboardRepository{db: db}
}

func (repository *DashboardRepository) Load(ctx context.Context, userID string, period entity.DashboardRange) (entity.DashboardSource, error) {
	result := entity.DashboardSource{
		CalorieDays: make([]entity.DashboardCalorieSourceDay, 0),
		Weights:     make([]entity.Weight, 0),
		Activities:  make([]entity.DashboardActivity, 0),
	}
	var defaultLimit int
	err := repository.db.QueryRowContext(ctx, `
SELECT default_daily_calorie_limit
FROM user_settings
WHERE user_id = $1`, userID).Scan(&defaultLimit)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return entity.DashboardSource{}, fmt.Errorf("load dashboard settings: %w", err)
	}
	if err == nil {
		result.DefaultDailyCalorieLimit = &defaultLimit
		if err := repository.loadCalorieDays(ctx, userID, period, &result); err != nil {
			return entity.DashboardSource{}, err
		}
	}
	if err := repository.loadWeights(ctx, userID, period, &result); err != nil {
		return entity.DashboardSource{}, err
	}
	if err := repository.loadActivities(ctx, userID, period, &result); err != nil {
		return entity.DashboardSource{}, err
	}
	return result, nil
}

func (repository *DashboardRepository) loadCalorieDays(ctx context.Context, userID string, period entity.DashboardRange, result *entity.DashboardSource) error {
	rows, err := repository.db.QueryContext(ctx, `
SELECT j.dt, SUM(f.cal100 * j.food_weight / 100.0), a.value
FROM journal j
JOIN food f ON f.key = j.food_key
LEFT JOIN act_calories a ON a.user_id = j.user_id AND a.dt = j.dt
WHERE j.user_id = $1 AND j.dt >= $2 AND j.dt <= $3
GROUP BY j.dt, a.value
ORDER BY j.dt`, userID, period.From, period.To)
	if err != nil {
		return fmt.Errorf("load dashboard calorie days: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var day entity.DashboardCalorieSourceDay
		var activeLimit sql.NullFloat64
		if err := rows.Scan(&day.DT, &day.Consumed, &activeLimit); err != nil {
			return fmt.Errorf("scan dashboard calorie day: %w", err)
		}
		if activeLimit.Valid {
			value := activeLimit.Float64
			day.ActiveLimit = &value
		}
		result.CalorieDays = append(result.CalorieDays, day)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate dashboard calorie days: %w", err)
	}
	return nil
}

func (repository *DashboardRepository) loadWeights(ctx context.Context, userID string, period entity.DashboardRange, result *entity.DashboardSource) error {
	rows, err := repository.db.QueryContext(ctx, `
WITH ranged AS (
    SELECT dt, value
    FROM weight
    WHERE user_id = $1 AND dt >= $2 AND dt <= $3
), bounds AS (
    SELECT MIN(dt) AS first_dt, MAX(dt) AS last_dt FROM ranged
)
SELECT ranged.dt, ranged.value
FROM ranged
CROSS JOIN bounds
WHERE ranged.dt = bounds.first_dt OR ranged.dt = bounds.last_dt
ORDER BY ranged.dt`, userID, period.From, period.To)
	if err != nil {
		return fmt.Errorf("load dashboard weights: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var item entity.Weight
		if err := rows.Scan(&item.DT, &item.Value); err != nil {
			return fmt.Errorf("scan dashboard weight: %w", err)
		}
		result.Weights = append(result.Weights, item)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate dashboard weights: %w", err)
	}
	return nil
}

func (repository *DashboardRepository) loadActivities(ctx context.Context, userID string, period entity.DashboardRange, result *entity.DashboardSource) error {
	rows, err := repository.db.QueryContext(ctx, `
SELECT s.key, s.name, COUNT(*)::int, SUM(totals.total), s.unit
FROM sport_activity a
JOIN sport s ON s.key = a.sport_key
JOIN LATERAL (
    SELECT COALESCE(SUM(value), 0) AS total
    FROM unnest(a.sets) AS value
) totals ON true
WHERE a.user_id = $1 AND a.dt >= $2 AND a.dt <= $3
GROUP BY s.key, s.name, s.unit
ORDER BY COUNT(*) DESC, lower(s.name), s.name, s.key
LIMIT 10`, userID, period.From, period.To)
	if err != nil {
		return fmt.Errorf("load dashboard activities: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var item entity.DashboardActivity
		if err := rows.Scan(&item.SportKey, &item.Name, &item.Count, &item.Total, &item.Unit); err != nil {
			return fmt.Errorf("scan dashboard activity: %w", err)
		}
		result.Activities = append(result.Activities, item)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate dashboard activities: %w", err)
	}
	return nil
}
