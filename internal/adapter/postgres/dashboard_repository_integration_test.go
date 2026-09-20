package postgres

import (
	"context"
	"database/sql"
	"math"
	"os"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/devldavydov/myhealth-ng/internal/entity"
)

func TestDashboardRepositoryWithPostgres(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("set TEST_DATABASE_URL to run PostgreSQL integration tests")
	}
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		t.Fatal(err)
	}
	if err := Migrate(ctx, db); err != nil {
		_ = db.Close()
		t.Fatal(err)
	}
	truncateDashboardTables := func(ctx context.Context) {
		_, _ = db.ExecContext(ctx, "TRUNCATE sport_activity, sport, act_calories, journal, user_settings, weight, bundle_item, bundle, food")
	}
	truncateDashboardTables(ctx)
	t.Cleanup(func() {
		truncateDashboardTables(context.Background())
		_ = db.Close()
	})

	statements := []string{
		`INSERT INTO user_settings (user_id, default_daily_calorie_limit) VALUES ('user-a', 2000)`,
		`INSERT INTO food (key, name, brand, cal100, prot100, fat100, carb100, comment) VALUES ('food-a', 'Тест', '', 100, 1, 1, 1, '')`,
		`INSERT INTO journal (user_id, dt, meal, food_key, food_weight) VALUES
            ('user-a', '2026-09-01', 'завтрак', 'food-a', 500),
            ('user-a', '2026-09-30', 'ужин', 'food-a', 1000),
            ('user-b', '2026-09-15', 'обед', 'food-a', 9000)`,
		`INSERT INTO act_calories (user_id, dt, value) VALUES ('user-a', '2026-09-30', 1200)`,
		`INSERT INTO weight (user_id, dt, value) VALUES
            ('user-a', '2026-08-31', 90), ('user-a', '2026-09-01', 82),
            ('user-a', '2026-09-15', 81), ('user-a', '2026-09-30', 80.5),
            ('user-b', '2026-09-10', 120)`,
		`INSERT INTO sport (key, name, unit, comment) VALUES
            ('walk', 'Ходьба', 'км', ''), ('bar', 'Турник', 'шт', '')`,
		`INSERT INTO sport_activity (user_id, dt, sport_key, sets) VALUES
            ('user-a', '2026-09-01', 'walk', ARRAY[5.0, 5.0]::real[]),
            ('user-a', '2026-09-02', 'walk', ARRAY[10.0]::real[]),
            ('user-a', '2026-09-03', 'bar', ARRAY[5.0, 5.0, 5.0]::real[]),
            ('user-b', '2026-09-04', 'walk', ARRAY[100.0]::real[])`,
	}
	for _, statement := range statements {
		if _, err := db.ExecContext(ctx, statement); err != nil {
			t.Fatal(err)
		}
	}

	period := entity.DashboardRange{
		From: time.Date(2026, time.September, 1, 0, 0, 0, 0, time.UTC),
		To:   time.Date(2026, time.September, 30, 0, 0, 0, 0, time.UTC),
	}
	repository := NewDashboardRepository(db)
	result, err := repository.Load(ctx, "user-a", period)
	if err != nil {
		t.Fatal(err)
	}
	if result.DefaultDailyCalorieLimit == nil || *result.DefaultDailyCalorieLimit != 2000 {
		t.Fatalf("default limit=%v", result.DefaultDailyCalorieLimit)
	}
	if len(result.CalorieDays) != 2 || result.CalorieDays[0].Consumed != 500 || result.CalorieDays[1].Consumed != 1000 {
		t.Fatalf("calorie days=%+v", result.CalorieDays)
	}
	if result.CalorieDays[1].ActiveLimit == nil || *result.CalorieDays[1].ActiveLimit != 1200 {
		t.Fatalf("active limit=%v", result.CalorieDays[1].ActiveLimit)
	}
	if len(result.Weights) != 2 || result.Weights[0].Value != 82 || result.Weights[1].Value != 80.5 {
		t.Fatalf("weights=%+v", result.Weights)
	}
	if len(result.Activities) != 2 || result.Activities[0].SportKey != "walk" || result.Activities[0].Count != 2 || math.Abs(result.Activities[0].Total-20) > .0001 {
		t.Fatalf("activities=%+v", result.Activities)
	}

	if _, err := db.ExecContext(ctx, `UPDATE food SET cal100 = 200 WHERE key = 'food-a'`); err != nil {
		t.Fatal(err)
	}
	updated, err := repository.Load(ctx, "user-a", period)
	if err != nil || len(updated.CalorieDays) != 2 || updated.CalorieDays[0].Consumed != 1000 {
		t.Fatalf("updated calorie days=%+v err=%v", updated.CalorieDays, err)
	}

	other, err := repository.Load(ctx, "user-b", period)
	if err != nil || other.DefaultDailyCalorieLimit != nil || len(other.CalorieDays) != 0 || len(other.Weights) != 1 || len(other.Activities) != 1 {
		t.Fatalf("other user=%+v err=%v", other, err)
	}
}
