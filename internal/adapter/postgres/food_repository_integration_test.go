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
		_, _ = db.ExecContext(context.Background(), "TRUNCATE journal, bundle_item, bundle, food")
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
	if _, err := db.ExecContext(ctx, "TRUNCATE journal, bundle_item, bundle, food"); err != nil {
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

	items, err := repository.List(ctx, entity.PageRequest{Page: 1, PageSize: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(items.Items) != 2 || items.Items[0].Name != "Банан" || items.Items[1].Name != "яблоко" || items.Total != 2 || items.TotalPages != 1 {
		t.Fatalf("unexpected order: %+v", items)
	}
	outOfRange, err := repository.List(ctx, entity.PageRequest{Page: 2, PageSize: 10})
	if err != nil || len(outOfRange.Items) != 0 || outOfRange.Page != 2 || outOfRange.Total != 2 || outOfRange.TotalPages != 1 {
		t.Fatalf("out-of-range page = %+v, err=%v", outOfRange, err)
	}
	percentMatches, err := repository.List(ctx, entity.PageRequest{Query: "100%", Page: 1, PageSize: 10})
	if err != nil || len(percentMatches.Items) != 1 || percentMatches.Items[0].Key != banana.Key || percentMatches.Total != 1 {
		t.Fatalf("literal percent search = %+v, err=%v", percentMatches, err)
	}
	underscoreMatches, err := repository.List(ctx, entity.PageRequest{Query: "_1", Page: 1, PageSize: 10})
	if err != nil || len(underscoreMatches.Items) != 1 || underscoreMatches.Items[0].Key != apple.Key {
		t.Fatalf("literal underscore search = %+v, err=%v", underscoreMatches, err)
	}
	keyMatches, err := repository.List(ctx, entity.PageRequest{Query: "3f67c05f", Page: 1, PageSize: 10})
	if err != nil || len(keyMatches.Items) != 0 {
		t.Fatalf("key search = %+v, err=%v", keyMatches, err)
	}
	injectionMatches, err := repository.List(ctx, entity.PageRequest{Query: "%' OR true --", Page: 1, PageSize: 10})
	if err != nil || len(injectionMatches.Items) != 0 {
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
