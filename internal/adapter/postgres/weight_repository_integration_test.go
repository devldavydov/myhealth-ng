package postgres

import (
	"context"
	"database/sql"
	"errors"
	"math"
	"os"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/devldavydov/myhealth-ng/internal/entity"
	"github.com/devldavydov/myhealth-ng/internal/port"
)

func TestWeightRepositoryWithPostgres(t *testing.T) {
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
		_, _ = db.ExecContext(context.Background(), "TRUNCATE weight")
		_ = db.Close()
	})
	if _, err := db.ExecContext(ctx, "TRUNCATE weight"); err != nil {
		t.Fatal(err)
	}

	repository := NewWeightRepository(db)
	userA := "3f67c05f-7c9e-4cb5-b26a-f9ce5b065865"
	userB := "c39dbf56-73ce-4630-b86c-a08c313eab75"
	dates := []time.Time{
		time.Date(2026, time.March, 15, 0, 0, 0, 0, time.UTC),
		time.Date(2026, time.May, 15, 0, 0, 0, 0, time.UTC),
		time.Date(2026, time.September, 15, 0, 0, 0, 0, time.UTC),
	}
	for index, dt := range dates {
		if _, err := repository.Save(ctx, userA, entity.Weight{DT: dt, Value: 83 - float64(index)}); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := repository.Save(ctx, userB, entity.Weight{DT: dates[2], Value: 99}); err != nil {
		t.Fatal(err)
	}

	from := time.Date(2026, time.April, 1, 0, 0, 0, 0, time.UTC)
	to := dates[2]
	items, err := repository.List(ctx, userA, entity.WeightRange{From: &from, To: &to})
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 || !items[0].DT.Equal(dates[2]) || !items[1].DT.Equal(dates[1]) {
		t.Fatalf("unexpected filtered order: %+v", items)
	}

	updated, err := repository.Save(ctx, userA, entity.Weight{DT: dates[2], Value: 79.4})
	if err != nil || math.Abs(updated.Value-79.4) > .0001 {
		t.Fatalf("updated=%+v err=%v", updated, err)
	}
	allUserA, err := repository.List(ctx, userA, entity.WeightRange{})
	if err != nil || len(allUserA) != 3 || math.Abs(allUserA[0].Value-79.4) > .0001 {
		t.Fatalf("user A items=%+v err=%v", allUserA, err)
	}
	allUserB, err := repository.List(ctx, userB, entity.WeightRange{})
	if err != nil || len(allUserB) != 1 || allUserB[0].Value != 99 {
		t.Fatalf("user B items=%+v err=%v", allUserB, err)
	}

	if err := repository.Delete(ctx, userA, entity.Weight{DT: dates[2]}); err != nil {
		t.Fatal(err)
	}
	if err := repository.Delete(ctx, userA, entity.Weight{DT: dates[2]}); !errors.Is(err, port.ErrWeightNotFound) {
		t.Fatalf("delete missing error=%v", err)
	}
	allUserB, err = repository.List(ctx, userB, entity.WeightRange{})
	if err != nil || len(allUserB) != 1 {
		t.Fatalf("deleting user A affected user B: items=%+v err=%v", allUserB, err)
	}
}
