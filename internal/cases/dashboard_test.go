package cases

import (
	"context"
	"errors"
	"math"
	"testing"
	"time"

	"github.com/devldavydov/myhealth-ng/internal/entity"
)

type dashboardRepositoryStub struct {
	userID string
	period entity.DashboardRange
	source entity.DashboardSource
	err    error
}

func (stub *dashboardRepositoryStub) Load(_ context.Context, userID string, period entity.DashboardRange) (entity.DashboardSource, error) {
	stub.userID, stub.period = userID, period
	return stub.source, stub.err
}

func TestDashboardCalculatesSummary(t *testing.T) {
	defaultLimit := 2000
	activeLimit := 2100.0
	stub := &dashboardRepositoryStub{source: entity.DashboardSource{
		DefaultDailyCalorieLimit: &defaultLimit,
		CalorieDays: []entity.DashboardCalorieSourceDay{
			{DT: dashboardDate(10), Consumed: 1800},
			{DT: dashboardDate(11), Consumed: 2200, ActiveLimit: &activeLimit},
		},
		Weights: []entity.Weight{
			{DT: dashboardDate(10), Value: 82},
			{DT: dashboardDate(20), Value: 80.5},
		},
		Activities: []entity.DashboardActivity{{SportKey: "walk", Name: "Ходьба", Count: 3, Total: 20, Unit: "км"}},
	}}
	result, err := NewDashboard(stub).Get(context.Background(), " user-a ", entity.DashboardRange{
		From: time.Date(2026, time.September, 1, 14, 0, 0, 0, time.Local),
		To:   time.Date(2026, time.September, 30, 22, 0, 0, 0, time.Local),
	})
	if err != nil {
		t.Fatal(err)
	}
	if stub.userID != "user-a" || stub.period.From.Hour() != 0 || stub.period.To.Hour() != 0 {
		t.Fatalf("scope was not normalized: user=%q period=%+v", stub.userID, stub.period)
	}
	if result.Calories == nil || len(result.Calories.Days) != 2 {
		t.Fatalf("calories=%+v", result.Calories)
	}
	if result.Calories.Days[0].Balance != 200 || result.Calories.Days[1].Balance != -100 {
		t.Fatalf("days=%+v", result.Calories.Days)
	}
	if result.Calories.Average == nil || math.Abs(*result.Calories.Average-50) > .0001 {
		t.Fatalf("average=%v", result.Calories.Average)
	}
	if result.WeightChange == nil || math.Abs(*result.WeightChange+1.5) > .0001 {
		t.Fatalf("weight change=%v", result.WeightChange)
	}
	if len(result.Activities) != 1 || result.Activities[0].Count != 3 {
		t.Fatalf("activities=%+v", result.Activities)
	}
}

func TestDashboardRequiresDefaultForCaloriesAndTwoWeights(t *testing.T) {
	stub := &dashboardRepositoryStub{source: entity.DashboardSource{
		CalorieDays: []entity.DashboardCalorieSourceDay{{DT: dashboardDate(10), Consumed: 1800}},
		Weights:     []entity.Weight{{DT: dashboardDate(10), Value: 82}},
	}}
	result, err := NewDashboard(stub).Get(context.Background(), "user-a", entity.DashboardRange{From: dashboardDate(1), To: dashboardDate(30)})
	if err != nil {
		t.Fatal(err)
	}
	if result.Calories != nil || result.WeightChange != nil || result.Activities == nil {
		t.Fatalf("unexpected empty result: %+v", result)
	}
}

func TestDashboardValidationAndRepositoryError(t *testing.T) {
	for _, test := range []struct {
		name   string
		userID string
		period entity.DashboardRange
	}{
		{name: "user", period: entity.DashboardRange{From: dashboardDate(1), To: dashboardDate(2)}},
		{name: "from", userID: "user", period: entity.DashboardRange{To: dashboardDate(2)}},
		{name: "to", userID: "user", period: entity.DashboardRange{From: dashboardDate(1)}},
		{name: "order", userID: "user", period: entity.DashboardRange{From: dashboardDate(2), To: dashboardDate(1)}},
	} {
		t.Run(test.name, func(t *testing.T) {
			if _, err := NewDashboard(&dashboardRepositoryStub{}).Get(context.Background(), test.userID, test.period); err == nil {
				t.Fatal("expected validation error")
			}
		})
	}
	want := errors.New("database unavailable")
	_, err := NewDashboard(&dashboardRepositoryStub{err: want}).Get(context.Background(), "user", entity.DashboardRange{From: dashboardDate(1), To: dashboardDate(2)})
	if !errors.Is(err, want) {
		t.Fatalf("error=%v", err)
	}
}

func dashboardDate(day int) time.Time {
	return time.Date(2026, time.September, day, 0, 0, 0, 0, time.UTC)
}
