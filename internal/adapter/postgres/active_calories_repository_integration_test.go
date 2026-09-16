package postgres

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/devldavydov/myhealth-ng/internal/entity"
	"github.com/devldavydov/myhealth-ng/internal/port"
)

func TestActiveCaloriesRepositoryWithPostgres(t *testing.T) {
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
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "TRUNCATE act_calories")
		_ = db.Close()
	})
	if _, err := db.ExecContext(ctx, "TRUNCATE act_calories"); err != nil {
		t.Fatal(err)
	}

	repository := NewActiveCaloriesRepository(db)
	dt := time.Date(2026, time.September, 16, 0, 0, 0, 0, time.UTC)
	if missing, err := repository.Get(ctx, "user-a", dt); err != nil || missing != nil {
		t.Fatalf("missing=%+v err=%v", missing, err)
	}
	if _, err := repository.Save(ctx, "user-a", entity.ActiveCalories{DT: dt, Value: 2500}); err != nil {
		t.Fatal(err)
	}
	if _, err := repository.Save(ctx, "user-a", entity.ActiveCalories{DT: dt, Value: 2750.5}); err != nil {
		t.Fatal(err)
	}
	if _, err := repository.Save(ctx, "user-b", entity.ActiveCalories{DT: dt, Value: 3100}); err != nil {
		t.Fatal(err)
	}
	userA, err := repository.Get(ctx, "user-a", dt)
	if err != nil || userA == nil || userA.Value != 2750.5 {
		t.Fatalf("user A=%+v err=%v", userA, err)
	}
	userB, err := repository.Get(ctx, "user-b", dt)
	if err != nil || userB == nil || userB.Value != 3100 {
		t.Fatalf("user B=%+v err=%v", userB, err)
	}
	if err := repository.Delete(ctx, "user-a", dt); err != nil {
		t.Fatal(err)
	}
	if err := repository.Delete(ctx, "user-a", dt); !errors.Is(err, port.ErrActiveCaloriesNotFound) {
		t.Fatalf("delete missing error=%v", err)
	}
	userB, err = repository.Get(ctx, "user-b", dt)
	if err != nil || userB == nil {
		t.Fatalf("deleting user A affected user B: item=%+v err=%v", userB, err)
	}
}
