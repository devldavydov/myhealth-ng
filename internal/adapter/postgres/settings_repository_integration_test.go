package postgres

import (
	"context"
	"database/sql"
	"os"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/devldavydov/myhealth-ng/internal/entity"
)

func TestSettingsRepositoryWithPostgres(t *testing.T) {
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
		_, _ = db.ExecContext(context.Background(), "TRUNCATE user_settings")
		_ = db.Close()
	})
	if _, err := db.ExecContext(ctx, "TRUNCATE user_settings"); err != nil {
		t.Fatal(err)
	}

	repository := NewSettingsRepository(db)
	userA := "3f67c05f-7c9e-4cb5-b26a-f9ce5b065865"
	userB := "c39dbf56-73ce-4630-b86c-a08c313eab75"
	missing, err := repository.Get(ctx, userA)
	if err != nil || missing != nil {
		t.Fatalf("missing=%+v err=%v", missing, err)
	}

	if _, err := repository.Save(ctx, userA, entity.UserSettings{DefaultDailyCalorieLimit: 2000}); err != nil {
		t.Fatal(err)
	}
	if _, err := repository.Save(ctx, userB, entity.UserSettings{DefaultDailyCalorieLimit: 2600}); err != nil {
		t.Fatal(err)
	}
	updated, err := repository.Save(ctx, userA, entity.UserSettings{DefaultDailyCalorieLimit: 2200})
	if err != nil || updated.DefaultDailyCalorieLimit != 2200 {
		t.Fatalf("updated=%+v err=%v", updated, err)
	}
	loadedA, err := repository.Get(ctx, userA)
	if err != nil || loadedA == nil || loadedA.DefaultDailyCalorieLimit != 2200 {
		t.Fatalf("user A=%+v err=%v", loadedA, err)
	}
	loadedB, err := repository.Get(ctx, userB)
	if err != nil || loadedB == nil || loadedB.DefaultDailyCalorieLimit != 2600 {
		t.Fatalf("user B=%+v err=%v", loadedB, err)
	}
}
