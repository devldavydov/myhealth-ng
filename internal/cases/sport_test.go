package cases

import (
	"context"
	"errors"
	"math"
	"testing"
	"time"

	"github.com/devldavydov/myhealth-ng/internal/entity"
)

type sportActivityRepositoryStub struct {
	userID         string
	period         entity.SportActivityRange
	saved, deleted entity.SportActivityData
}

func (stub *sportActivityRepositoryStub) List(_ context.Context, userID string, period entity.SportActivityRange) ([]entity.SportActivity, error) {
	stub.userID = userID
	stub.period = period
	return nil, nil
}
func (stub *sportActivityRepositoryStub) Save(_ context.Context, userID string, data entity.SportActivityData) (entity.SportActivity, error) {
	stub.userID = userID
	stub.saved = data
	return entity.SportActivity{DT: data.DT, Sets: data.Sets}, nil
}
func (stub *sportActivityRepositoryStub) Delete(_ context.Context, userID string, data entity.SportActivityData) error {
	stub.userID = userID
	stub.deleted = data
	return nil
}

func TestSportActivityValidatesAndNormalizes(t *testing.T) {
	repository := &sportActivityRepositoryStub{}
	useCases := NewSportActivity(repository)
	dt := time.Date(2026, time.September, 18, 19, 30, 0, 0, time.Local)
	_, err := useCases.Save(context.Background(), " user ", entity.SportActivityData{DT: dt, SportKey: " турник ", Sets: []float64{5, 3}})
	if err != nil {
		t.Fatal(err)
	}
	if repository.userID != "user" || repository.saved.SportKey != "турник" || repository.saved.DT.Hour() != 0 {
		t.Fatalf("user=%q saved=%+v", repository.userID, repository.saved)
	}
	for _, sets := range [][]float64{nil, {}, {0}, {-1}, {math.Inf(1)}} {
		_, err := useCases.Save(context.Background(), "user", entity.SportActivityData{DT: dt, SportKey: "турник", Sets: sets})
		var validation *entity.ValidationError
		if !errors.As(err, &validation) || len(validation.Fields["sets"]) == 0 {
			t.Fatalf("sets=%v err=%v", sets, err)
		}
	}
}

func TestSportActivityRangeRejectsReverseDates(t *testing.T) {
	from := time.Date(2026, time.September, 19, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, time.September, 18, 0, 0, 0, 0, time.UTC)
	_, err := NewSportActivity(&sportActivityRepositoryStub{}).List(context.Background(), "user", entity.SportActivityRange{From: &from, To: &to})
	var validation *entity.ValidationError
	if !errors.As(err, &validation) || len(validation.Fields["from"]) == 0 {
		t.Fatalf("err=%v", err)
	}
}
