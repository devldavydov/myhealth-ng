package cases

import (
	"context"
	"errors"
	"math"
	"testing"
	"time"

	"github.com/devldavydov/myhealth-ng/internal/entity"
	"github.com/devldavydov/myhealth-ng/internal/port"
)

const testUserID = "3f67c05f-7c9e-4cb5-b26a-f9ce5b065865"

type weightRepositoryStub struct {
	userID  string
	period  entity.WeightRange
	saved   entity.Weight
	deleted entity.Weight
	items   []entity.Weight
	err     error
}

func (stub *weightRepositoryStub) List(_ context.Context, userID string, period entity.WeightRange) ([]entity.Weight, error) {
	stub.userID, stub.period = userID, period
	return stub.items, stub.err
}

func (stub *weightRepositoryStub) Save(_ context.Context, userID string, data entity.Weight) (entity.Weight, error) {
	stub.userID, stub.saved = userID, data
	return data, stub.err
}

func (stub *weightRepositoryStub) Delete(_ context.Context, userID string, data entity.Weight) error {
	stub.userID, stub.deleted = userID, data
	return stub.err
}

func TestWeightListValidatesAndNormalizesRange(t *testing.T) {
	repository := &weightRepositoryStub{}
	useCases := NewWeight(repository)
	from := time.Date(2026, time.January, 2, 17, 30, 0, 0, time.FixedZone("test", 3*60*60))
	to := time.Date(2026, time.February, 3, 9, 0, 0, 0, time.Local)

	if _, err := useCases.List(context.Background(), "  "+testUserID+"  ", entity.WeightRange{From: &from, To: &to}); err != nil {
		t.Fatal(err)
	}
	if repository.userID != testUserID {
		t.Fatalf("userID = %q", repository.userID)
	}
	if repository.period.From.Format("2006-01-02T15:04:05Z07:00") != "2026-01-02T00:00:00Z" {
		t.Fatalf("from = %v", repository.period.From)
	}
}

func TestWeightRejectsInvalidRange(t *testing.T) {
	useCases := NewWeight(&weightRepositoryStub{})
	from := time.Date(2026, time.March, 2, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, time.March, 1, 0, 0, 0, 0, time.UTC)
	_, err := useCases.List(context.Background(), testUserID, entity.WeightRange{From: &from, To: &to})
	assertWeightValidationField(t, err, "from")
}

func TestWeightSaveValidatesAndDelegates(t *testing.T) {
	repository := &weightRepositoryStub{}
	useCases := NewWeight(repository)
	dt := time.Date(2026, time.March, 4, 19, 0, 0, 0, time.Local)

	saved, err := useCases.Save(context.Background(), testUserID, entity.Weight{DT: dt, Value: 82.4})
	if err != nil {
		t.Fatal(err)
	}
	if saved.Value != 82.4 || repository.saved.DT.Hour() != 0 || repository.userID != testUserID {
		t.Fatalf("saved=%+v repository=%+v user=%q", saved, repository.saved, repository.userID)
	}

	for name, data := range map[string]entity.Weight{
		"missing date": {Value: 82},
		"zero":         {DT: dt, Value: 0},
		"negative":     {DT: dt, Value: -1},
		"infinite":     {DT: dt, Value: math.Inf(1)},
	} {
		t.Run(name, func(t *testing.T) {
			_, err := useCases.Save(context.Background(), testUserID, data)
			if data.DT.IsZero() {
				assertWeightValidationField(t, err, "dt")
			} else {
				assertWeightValidationField(t, err, "value")
			}
		})
	}
}

func TestWeightDeleteAndErrors(t *testing.T) {
	dt := time.Date(2026, time.April, 5, 0, 0, 0, 0, time.UTC)
	repository := &weightRepositoryStub{}
	useCases := NewWeight(repository)
	if err := useCases.Delete(context.Background(), testUserID, entity.Weight{DT: dt}); err != nil {
		t.Fatal(err)
	}
	if repository.deleted.DT != dt || repository.userID != testUserID {
		t.Fatalf("deleted=%+v user=%q", repository.deleted, repository.userID)
	}

	repository.err = port.ErrWeightNotFound
	if err := useCases.Delete(context.Background(), testUserID, entity.Weight{DT: dt}); !errors.Is(err, port.ErrWeightNotFound) {
		t.Fatalf("error = %v", err)
	}
	assertWeightValidationField(t, useCases.Delete(context.Background(), "", entity.Weight{DT: dt}), "userId")
}

func assertWeightValidationField(t *testing.T, err error, field string) {
	t.Helper()
	var validationError *entity.ValidationError
	if !errors.As(err, &validationError) || len(validationError.Fields[field]) == 0 {
		t.Fatalf("error = %v, want validation for %s", err, field)
	}
}
