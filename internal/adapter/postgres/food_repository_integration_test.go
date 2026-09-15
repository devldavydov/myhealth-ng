package postgres

import (
	"context"
	"database/sql"
	"errors"
	"math"
	"os"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/devldavydov/myhealth-ng/internal/entity"
	"github.com/devldavydov/myhealth-ng/internal/port"
)

func TestFoodRepositoryWithPostgres(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("set TEST_DATABASE_URL to run PostgreSQL integration tests")
	}
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "TRUNCATE bundle_item, bundle, food")
		_ = db.Close()
	})
	ctx := context.Background()
	if err := db.PingContext(ctx); err != nil {
		t.Fatal(err)
	}
	if err := Migrate(ctx, db); err != nil {
		t.Fatal(err)
	}
	if err := Migrate(ctx, db); err != nil {
		t.Fatalf("second migration run must be idempotent: %v", err)
	}
	if _, err := db.ExecContext(ctx, "TRUNCATE bundle_item, bundle, food"); err != nil {
		t.Fatal(err)
	}
	repository := NewFoodRepository(db)
	banana := entity.Food{
		Key: "3f67c05f-7c9e-4cb5-b26a-f9ce5b065865", Name: "Банан", Brand: "Фрукты 100%",
		Cal100: 89, Prot100: 1.1, Fat100: .3, Carb100: 22.8, Comment: "Спелый",
	}
	apple := entity.Food{
		Key: "c39dbf56-73ce-4630-b86c-a08c313eab75", Name: "яблоко", Brand: "Сад_1",
		Cal100: 52, Prot100: .3, Fat100: .2, Carb100: 14,
	}
	for _, item := range []entity.Food{banana, apple} {
		if _, err := repository.Create(ctx, item); err != nil {
			t.Fatal(err)
		}
	}

	items, err := repository.List(ctx, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 || items[0].Name != "Банан" || items[1].Name != "яблоко" {
		t.Fatalf("unexpected order: %+v", items)
	}
	percentMatches, err := repository.List(ctx, "100%")
	if err != nil || len(percentMatches) != 1 || percentMatches[0].Key != banana.Key {
		t.Fatalf("literal percent search = %+v, err=%v", percentMatches, err)
	}
	underscoreMatches, err := repository.List(ctx, "_1")
	if err != nil || len(underscoreMatches) != 1 || underscoreMatches[0].Key != apple.Key {
		t.Fatalf("literal underscore search = %+v, err=%v", underscoreMatches, err)
	}
	keyMatches, err := repository.List(ctx, "3f67c05f")
	if err != nil || len(keyMatches) != 0 {
		t.Fatalf("key search = %+v, err=%v", keyMatches, err)
	}
	injectionMatches, err := repository.List(ctx, "%' OR true --")
	if err != nil || len(injectionMatches) != 0 {
		t.Fatalf("injection-like search = %+v, err=%v", injectionMatches, err)
	}

	updated, err := repository.Update(ctx, banana.Key, entity.FoodData{
		Name: "Банан мини", Brand: banana.Brand, Cal100: 90,
		Prot100: 1.2, Fat100: .4, Carb100: 23, Comment: banana.Comment,
	})
	if err != nil || updated.Name != "Банан мини" || math.Abs(updated.Prot100-1.2) > .0001 {
		t.Fatalf("updated=%+v err=%v", updated, err)
	}
	loaded, err := repository.Get(ctx, banana.Key)
	if err != nil || loaded.Name != updated.Name {
		t.Fatalf("loaded=%+v err=%v", loaded, err)
	}
	if err := repository.Delete(ctx, banana.Key); err != nil {
		t.Fatal(err)
	}
	if _, err := repository.Get(ctx, banana.Key); !errors.Is(err, port.ErrFoodNotFound) {
		t.Fatalf("get deleted error=%v", err)
	}
	if err := repository.Delete(ctx, banana.Key); !errors.Is(err, port.ErrFoodNotFound) {
		t.Fatalf("delete missing error=%v", err)
	}
	if _, err := repository.Create(ctx, apple); !errors.Is(err, port.ErrFoodConflict) {
		t.Fatalf("duplicate error=%v", err)
	}
}
