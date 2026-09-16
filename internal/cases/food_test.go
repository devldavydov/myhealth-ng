package cases

import (
	"context"
	"errors"
	"math"
	"testing"

	"github.com/devldavydov/myhealth-ng/internal/entity"
	"github.com/devldavydov/myhealth-ng/internal/port"
)

const testKey = "3f67c05f-7c9e-4cb5-b26a-f9ce5b065865"

type foodRepositoryStub struct {
	items      []entity.Food
	request    entity.PageRequest
	created    entity.Food
	updatedKey string
	updated    entity.FoodData
	deletedKey string
	err        error
}

func (stub *foodRepositoryStub) List(_ context.Context, request entity.PageRequest) (entity.Page[entity.Food], error) {
	stub.request = request
	return entity.Page[entity.Food]{Items: stub.items, Page: request.Page, PageSize: request.PageSize, Total: len(stub.items)}, stub.err
}
func (stub *foodRepositoryStub) Get(_ context.Context, _ string) (entity.Food, error) {
	if len(stub.items) == 0 {
		return entity.Food{}, stub.err
	}
	return stub.items[0], stub.err
}
func (stub *foodRepositoryStub) Create(_ context.Context, food entity.Food) (entity.Food, error) {
	stub.created = food
	return food, stub.err
}
func (stub *foodRepositoryStub) Update(_ context.Context, key string, data entity.FoodData) (entity.Food, error) {
	stub.updatedKey, stub.updated = key, data
	return entity.Food{Key: key, Name: data.Name}, stub.err
}
func (stub *foodRepositoryStub) Delete(_ context.Context, key string) error {
	stub.deletedKey = key
	return stub.err
}

func TestFoodCreateNormalizesAndGeneratesKey(t *testing.T) {
	repository := &foodRepositoryStub{}
	useCases := NewFood(repository)
	useCases.generateKey = func() (string, error) { return testKey, nil }

	created, err := useCases.Create(context.Background(), entity.FoodData{
		Name: "  Творог  ", Brand: " Ферма ", Cal100: 120,
		Prot100: 18, Fat100: 5, Carb100: 3, Comment: " Хороший ",
	})
	if err != nil {
		t.Fatal(err)
	}
	if created.Key != testKey || created.Name != "Творог" || created.Brand != "Ферма" || created.Comment != "Хороший" {
		t.Fatalf("unexpected food: %+v", created)
	}
}

func TestFoodValidation(t *testing.T) {
	useCases := NewFood(&foodRepositoryStub{})
	_, err := useCases.Create(context.Background(), entity.FoodData{Name: " ", Cal100: -1, Prot100: math.Inf(1)})
	var validationError *entity.ValidationError
	if !errors.As(err, &validationError) {
		t.Fatalf("error = %v, want ValidationError", err)
	}
	for _, field := range []string{"name", "cal100", "prot100"} {
		if len(validationError.Fields[field]) == 0 {
			t.Errorf("missing validation for %s", field)
		}
	}
}

func TestFoodListTrimsQuery(t *testing.T) {
	repository := &foodRepositoryStub{}
	useCases := NewFood(repository)
	if _, err := useCases.List(context.Background(), entity.PageRequest{Query: "  молоко  "}); err != nil {
		t.Fatal(err)
	}
	if repository.request.Query != "молоко" || repository.request.Page != 1 || repository.request.PageSize != 20 {
		t.Fatalf("request = %+v", repository.request)
	}
}

func TestFoodUpdateAndDeleteValidateKey(t *testing.T) {
	repository := &foodRepositoryStub{}
	useCases := NewFood(repository)
	data := entity.FoodData{Name: "Яблоко"}
	if _, err := useCases.Update(context.Background(), testKey, data); err != nil {
		t.Fatal(err)
	}
	if repository.updatedKey != testKey || repository.updated.Name != "Яблоко" {
		t.Fatalf("unexpected update: %q %+v", repository.updatedKey, repository.updated)
	}
	if err := useCases.Delete(context.Background(), testKey); err != nil {
		t.Fatal(err)
	}
	if repository.deletedKey != testKey {
		t.Fatalf("deleted key = %q", repository.deletedKey)
	}
	if err := useCases.Delete(context.Background(), "not-a-uuid"); err == nil {
		t.Fatal("invalid key must be rejected")
	}
}

func TestFoodPropagatesRepositoryError(t *testing.T) {
	repository := &foodRepositoryStub{err: port.ErrFoodNotFound}
	useCases := NewFood(repository)
	if _, err := useCases.Get(context.Background(), testKey); !errors.Is(err, port.ErrFoodNotFound) {
		t.Fatalf("error = %v", err)
	}
}
