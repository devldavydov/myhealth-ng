package httpapi_test

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/devldavydov/myhealth-ng/internal/adapter/httpapi"
	"github.com/devldavydov/myhealth-ng/internal/entity"
)

type dashboardUseCasesStub struct {
	userID string
	period entity.DashboardRange
	result entity.Dashboard
	err    error
}

func (stub *dashboardUseCasesStub) Get(_ context.Context, userID string, period entity.DashboardRange) (entity.Dashboard, error) {
	stub.userID, stub.period = userID, period
	return stub.result, stub.err
}

func TestDashboardRoute(t *testing.T) {
	average, change := 50.0, -1.5
	stub := &dashboardUseCasesStub{result: entity.Dashboard{
		Calories: &entity.DashboardCalories{
			Days:    []entity.DashboardCalorieDay{{DT: time.Date(2026, time.September, 10, 0, 0, 0, 0, time.UTC), Balance: 200}},
			Average: &average,
		},
		WeightChange: &change,
		Activities:   []entity.DashboardActivity{{SportKey: "walk", Name: "Ходьба", Count: 3, Total: 20, Unit: "км"}},
	}}
	router := dashboardRouter(stub)
	response := request(router, http.MethodGet, "/api/dashboard?from=2026-09-01&to=2026-09-30", "")
	if response.Code != http.StatusOK || stub.userID != "00000000-0000-4000-8000-000000000000" {
		t.Fatalf("status=%d user=%q body=%s", response.Code, stub.userID, response.Body)
	}
	if stub.period.From.Day() != 1 || stub.period.To.Day() != 30 {
		t.Fatalf("period=%+v", stub.period)
	}
	body := response.Body.String()
	for _, fragment := range []string{`"balance":200`, `"average":50`, `"weightChange":-1.5`, `"name":"Ходьба"`, `"activities":[`} {
		if !strings.Contains(body, fragment) {
			t.Fatalf("missing %s in %s", fragment, body)
		}
	}
}

func TestDashboardRouteValidationAndErrors(t *testing.T) {
	for _, path := range []string{
		"/api/dashboard",
		"/api/dashboard?from=wrong&to=2026-09-30",
		"/api/dashboard?from=2026-10-01&to=2026-09-30",
	} {
		response := request(dashboardRouter(&dashboardUseCasesStub{}), http.MethodGet, path, "")
		if response.Code != http.StatusBadRequest {
			t.Fatalf("path=%s status=%d body=%s", path, response.Code, response.Body)
		}
	}
	response := request(dashboardRouter(&dashboardUseCasesStub{err: errors.New("database unavailable")}), http.MethodGet, "/api/dashboard?from=2026-09-01&to=2026-09-30", "")
	if response.Code != http.StatusInternalServerError {
		t.Fatalf("status=%d body=%s", response.Code, response.Body)
	}
}

func dashboardRouter(dashboard *dashboardUseCasesStub) http.Handler {
	return httpapi.NewRouterWithSportAndDashboard(
		&foodUseCasesStub{}, &weightUseCasesStub{}, &bundleUseCasesStub{}, &settingsUseCasesStub{},
		&journalUseCasesStub{}, &activeCaloriesUseCasesStub{}, nil, nil, dashboard, httpapi.Options{},
	)
}
