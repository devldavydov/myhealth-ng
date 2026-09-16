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

func TestJournalRepositoryWithPostgres(t *testing.T) {
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
		_, _ = db.ExecContext(context.Background(), "TRUNCATE journal, bundle_item, bundle, food")
		_ = db.Close()
	})
	if _, err := db.ExecContext(ctx, "TRUNCATE journal, bundle_item, bundle, food"); err != nil {
		t.Fatal(err)
	}
	foodRepository := NewFoodRepository(db)
	journalRepository := NewJournalRepository(db)
	foodA := entity.Food{Key: "3f67c05f-7c9e-4cb5-b26a-f9ce5b065865", Name: "Творог", Cal100: 100, Prot100: 10, Fat100: 5, Carb100: 20}
	foodB := entity.Food{Key: "c39dbf56-73ce-4630-b86c-a08c313eab75", Name: "Хлеб", Cal100: 250, Prot100: 8, Fat100: 3, Carb100: 45}
	for _, food := range []entity.Food{foodA, foodB} {
		if _, err := foodRepository.Create(ctx, food); err != nil {
			t.Fatal(err)
		}
	}
	dt := time.Date(2026, time.September, 16, 0, 0, 0, 0, time.UTC)
	userA, userB := "user-a", "user-b"
	if err := journalRepository.Save(ctx, userA, dt, entity.MealBreakfast, []entity.JournalItemData{{FoodKey: foodA.Key, Weight: 100}, {FoodKey: foodB.Key, Weight: 40}}); err != nil {
		t.Fatal(err)
	}
	if err := journalRepository.Save(ctx, userA, dt, entity.MealBreakfast, []entity.JournalItemData{{FoodKey: foodA.Key, Weight: 150}}); err != nil {
		t.Fatal(err)
	}
	if err := journalRepository.Save(ctx, userB, dt, entity.MealDinner, []entity.JournalItemData{{FoodKey: foodA.Key, Weight: 70}}); err != nil {
		t.Fatal(err)
	}
	itemsA, err := journalRepository.List(ctx, userA, dt)
	if err != nil || len(itemsA) != 2 || itemsA[0].Weight != 150 {
		t.Fatalf("user A items=%+v err=%v", itemsA, err)
	}
	itemsB, err := journalRepository.List(ctx, userB, dt)
	if err != nil || len(itemsB) != 1 || itemsB[0].Meal != entity.MealDinner {
		t.Fatalf("user B items=%+v err=%v", itemsB, err)
	}
	if err := foodRepository.Delete(ctx, foodA.Key); !errors.Is(err, port.ErrFoodInUse) {
		t.Fatalf("delete used food error=%v", err)
	}
	if err := journalRepository.DeleteItem(ctx, userA, dt, entity.MealBreakfast, foodB.Key); err != nil {
		t.Fatal(err)
	}
	if err := journalRepository.DeleteItem(ctx, userA, dt, entity.MealBreakfast, foodB.Key); !errors.Is(err, port.ErrJournalItemNotFound) {
		t.Fatalf("delete missing error=%v", err)
	}
	if err := journalRepository.ClearMeal(ctx, userA, dt, entity.MealBreakfast); err != nil {
		t.Fatal(err)
	}
	itemsA, err = journalRepository.List(ctx, userA, dt)
	if err != nil || len(itemsA) != 0 {
		t.Fatalf("cleared user A items=%+v err=%v", itemsA, err)
	}
	itemsB, err = journalRepository.List(ctx, userB, dt)
	if err != nil || len(itemsB) != 1 {
		t.Fatalf("clear affected user B: %+v err=%v", itemsB, err)
	}
}
