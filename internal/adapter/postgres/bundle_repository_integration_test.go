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

func TestBundleRepositoryWithPostgres(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("set TEST_DATABASE_URL to run PostgreSQL integration tests")
	}
	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	t.Cleanup(func() {
		_, _ = db.ExecContext(context.Background(), "TRUNCATE bundle_item, bundle, food")
		_ = db.Close()
	})
	if err := db.PingContext(ctx); err != nil {
		t.Fatal(err)
	}
	if err := Migrate(ctx, db); err != nil {
		t.Fatal(err)
	}
	if _, err := db.ExecContext(ctx, "TRUNCATE bundle_item, bundle, food"); err != nil {
		t.Fatal(err)
	}

	foodRepository := NewFoodRepository(db)
	cheese := entity.Food{Key: "3f67c05f-7c9e-4cb5-b26a-f9ce5b065865", Name: "Сыр", Brand: "Ферма", Cal100: 300, Prot100: 20, Fat100: 25, Carb100: 1}
	bread := entity.Food{Key: "c39dbf56-73ce-4630-b86c-a08c313eab75", Name: "Хлеб", Brand: "", Cal100: 250, Prot100: 8, Fat100: 2, Carb100: 50}
	for _, food := range []entity.Food{cheese, bread} {
		if _, err := foodRepository.Create(ctx, food); err != nil {
			t.Fatal(err)
		}
	}

	repository := NewBundleRepository(db)
	bundleKey := "8c2cf7aa-44f4-49a4-aef0-e9087088c980"
	created, err := repository.Create(ctx, entity.Bundle{
		Key: bundleKey, Name: "Бутерброд_1",
		Items: []entity.BundleItem{{Food: cheese, Weight: 20}, {Food: bread, Weight: 40}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(created.Items) != 2 || math.Abs(created.Totals.Weight-60) > .0001 || math.Abs(created.Totals.Cal-160) > .0001 {
		t.Fatalf("unexpected created bundle: %+v", created)
	}
	listed, err := repository.List(ctx, "_1")
	if err != nil || len(listed) != 1 || listed[0].ItemCount != 2 || listed[0].Key != bundleKey {
		t.Fatalf("listed=%+v err=%v", listed, err)
	}
	updated, err := repository.Update(ctx, bundleKey, entity.BundleData{Name: "Бутерброд", Items: []entity.BundleItemData{{FoodKey: bread.Key, Weight: 80}}})
	if err != nil || updated.Name != "Бутерброд" || len(updated.Items) != 1 || updated.Items[0].Weight != 80 {
		t.Fatalf("updated=%+v err=%v", updated, err)
	}
	if err := foodRepository.Delete(ctx, bread.Key); !errors.Is(err, port.ErrFoodInUse) {
		t.Fatalf("delete used food error=%v", err)
	}
	if err := repository.Delete(ctx, bundleKey); err != nil {
		t.Fatal(err)
	}
	if _, err := repository.Get(ctx, bundleKey); !errors.Is(err, port.ErrBundleNotFound) {
		t.Fatalf("get deleted bundle error=%v", err)
	}
	if err := foodRepository.Delete(ctx, bread.Key); err != nil {
		t.Fatalf("delete released food: %v", err)
	}
}
