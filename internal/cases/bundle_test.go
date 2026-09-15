package cases

import (
	"context"
	"errors"
	"testing"

	"github.com/devldavydov/myhealth-ng/internal/entity"
	"github.com/devldavydov/myhealth-ng/internal/port"
)

const (
	testBundleKey = "8c2cf7aa-44f4-49a4-aef0-e9087088c980"
	testFoodKey   = "3f67c05f-7c9e-4cb5-b26a-f9ce5b065865"
)

type bundleRepositoryStub struct {
	created entity.Bundle
	updated entity.BundleData
	err     error
}

func (stub *bundleRepositoryStub) List(context.Context, string) ([]entity.BundleSummary, error) {
	return nil, stub.err
}
func (stub *bundleRepositoryStub) Get(context.Context, string) (entity.Bundle, error) {
	return entity.Bundle{}, stub.err
}
func (stub *bundleRepositoryStub) Create(_ context.Context, bundle entity.Bundle) (entity.Bundle, error) {
	stub.created = bundle
	return bundle, stub.err
}
func (stub *bundleRepositoryStub) Update(_ context.Context, _ string, data entity.BundleData) (entity.Bundle, error) {
	stub.updated = data
	return entity.Bundle{Name: data.Name}, stub.err
}
func (stub *bundleRepositoryStub) Delete(context.Context, string) error { return stub.err }

func TestBundleCreateNormalizesAndMergesItems(t *testing.T) {
	repository := &bundleRepositoryStub{}
	useCases := NewBundle(repository)
	useCases.generateKey = func() (string, error) { return testBundleKey, nil }

	created, err := useCases.Create(context.Background(), entity.BundleData{
		Name: "  Завтрак  ",
		Items: []entity.BundleItemData{
			{FoodKey: "  " + testFoodKey + "  ", Weight: 40},
			{FoodKey: testFoodKey, Weight: 8},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if created.Key != testBundleKey || repository.created.Name != "Завтрак" || len(repository.created.Items) != 1 || repository.created.Items[0].Weight != 48 {
		t.Fatalf("unexpected bundle: %+v", repository.created)
	}
}

func TestBundleValidation(t *testing.T) {
	useCases := NewBundle(&bundleRepositoryStub{})
	for _, data := range []entity.BundleData{
		{},
		{Name: "Тест", Items: []entity.BundleItemData{{FoodKey: "wrong", Weight: 10}}},
		{Name: "Тест", Items: []entity.BundleItemData{{FoodKey: testFoodKey, Weight: 0}}},
	} {
		if _, err := useCases.Create(context.Background(), data); err == nil {
			t.Fatalf("expected validation error for %+v", data)
		} else {
			var validationError *entity.ValidationError
			if !errors.As(err, &validationError) {
				t.Fatalf("expected validation error, got %v", err)
			}
		}
	}
}

func TestBundleMapsMissingFoodToValidation(t *testing.T) {
	useCases := NewBundle(&bundleRepositoryStub{err: port.ErrBundleFoodNotFound})
	useCases.generateKey = func() (string, error) { return testBundleKey, nil }
	_, err := useCases.Create(context.Background(), entity.BundleData{
		Name: "Тест", Items: []entity.BundleItemData{{FoodKey: testFoodKey, Weight: 10}},
	})
	var validationError *entity.ValidationError
	if !errors.As(err, &validationError) || len(validationError.Fields["items"]) == 0 {
		t.Fatalf("unexpected error: %v", err)
	}
}
