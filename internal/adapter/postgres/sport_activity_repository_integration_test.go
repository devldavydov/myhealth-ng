package postgres

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/devldavydov/myhealth-ng/internal/entity"
)

func TestSportActivityRepositoryWithPostgres(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("set TEST_DATABASE_URL to run PostgreSQL integration tests")
	}
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ctx := context.Background()
	if err := db.PingContext(ctx); err != nil {
		t.Fatal(err)
	}
	if err := Migrate(ctx, db); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, "TRUNCATE sport_activity, sport"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _, _ = db.ExecContext(context.Background(), "TRUNCATE sport_activity, sport") })
	sportRepository := NewSportRepository(db)
	sport, err := sportRepository.Create(ctx, entity.Sport{Key: "турник", Name: "Турник", Unit: "шт", Comment: "Подтягивания"})
	if err != nil {
		t.Fatal(err)
	}
	repository := NewSportActivityRepository(db)
	dt := time.Date(2026, time.September, 18, 0, 0, 0, 0, time.UTC)
	saved, err := repository.Save(ctx, "user-a", entity.SportActivityData{DT: dt, SportKey: sport.Key, Sets: []float64{5, 3, 3.5}})
	if err != nil {
		t.Fatal(err)
	}
	if len(saved.Sets) != 3 || saved.Sets[2] != 3.5 {
		t.Fatalf("saved=%+v", saved)
	}
	items, err := repository.List(ctx, "user-a", entity.SportActivityRange{})
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 || len(items[0].Sets) != 3 || items[0].Sets[0] != 5 {
		t.Fatalf("items=%+v", items)
	}
	other, err := repository.List(ctx, "user-b", entity.SportActivityRange{})
	if err != nil || len(other) != 0 {
		t.Fatalf("other=%+v err=%v", other, err)
	}
}
