package httpapi_test

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/devldavydov/myhealth-ng/internal/adapter/httpapi"
	"github.com/devldavydov/myhealth-ng/internal/entity"
	"github.com/devldavydov/myhealth-ng/internal/port"
)

type activeCaloriesUseCasesStub struct {
	userID  string
	dt      time.Time
	item    *entity.ActiveCalories
	saved   entity.ActiveCalories
	deleted time.Time
	err     error
}

func (stub *activeCaloriesUseCasesStub) Get(_ context.Context, userID string, dt time.Time) (*entity.ActiveCalories, error) {
	stub.userID, stub.dt = userID, dt
	return stub.item, stub.err
}

func (stub *activeCaloriesUseCasesStub) Save(_ context.Context, userID string, data entity.ActiveCalories) (entity.ActiveCalories, error) {
	stub.userID, stub.saved = userID, data
	return data, stub.err
}

func (stub *activeCaloriesUseCasesStub) Delete(_ context.Context, userID string, dt time.Time) error {
	stub.userID, stub.deleted = userID, dt
	return stub.err
}

func activeCaloriesRouter(stub port.ActiveCaloriesUseCases) http.Handler {
	return httpapi.NewRouter(&foodUseCasesStub{}, &weightUseCasesStub{}, &bundleUseCasesStub{}, &settingsUseCasesStub{}, &journalUseCasesStub{}, stub, httpapi.Options{})
}

func TestActiveCaloriesRoutesUseCurrentUser(t *testing.T) {
	dt := time.Date(2026, time.September, 16, 0, 0, 0, 0, time.UTC)
	stub := &activeCaloriesUseCasesStub{item: &entity.ActiveCalories{DT: dt, Value: 2600.5}}
	router := activeCaloriesRouter(stub)

	loaded := request(router, http.MethodGet, "/api/active-calories?dt=2026-09-16", "")
	if loaded.Code != http.StatusOK || stub.userID != "00000000-0000-4000-8000-000000000000" || !strings.Contains(loaded.Body.String(), `"value":2600.5`) {
		t.Fatalf("get status=%d user=%q body=%s", loaded.Code, stub.userID, loaded.Body)
	}
	saved := request(router, http.MethodPost, "/api/active-calories", `{"dt":"2026-09-16","value":2750.5}`)
	if saved.Code != http.StatusOK || stub.saved.Value != 2750.5 || stub.saved.DT.Format(dateLayoutForTest) != "2026-09-16" {
		t.Fatalf("save status=%d data=%+v body=%s", saved.Code, stub.saved, saved.Body)
	}
	deleted := request(router, http.MethodDelete, "/api/active-calories/2026-09-16", "")
	if deleted.Code != http.StatusNoContent || stub.deleted.Format(dateLayoutForTest) != "2026-09-16" {
		t.Fatalf("delete status=%d date=%v body=%s", deleted.Code, stub.deleted, deleted.Body)
	}
}

func TestActiveCaloriesEmptyValidationAndNotFound(t *testing.T) {
	empty := request(activeCaloriesRouter(&activeCaloriesUseCasesStub{}), http.MethodGet, "/api/active-calories?dt=2026-09-16", "")
	if empty.Code != http.StatusOK || !strings.Contains(empty.Body.String(), `"data":null`) {
		t.Fatalf("empty status=%d body=%s", empty.Code, empty.Body)
	}
	for _, test := range []struct {
		method string
		path   string
		body   string
	}{
		{http.MethodGet, "/api/active-calories?dt=wrong", ""},
		{http.MethodPost, "/api/active-calories", `{}`},
		{http.MethodPost, "/api/active-calories", `{"dt":"2026-09-16","value":null}`},
		{http.MethodDelete, "/api/active-calories/wrong", ""},
	} {
		response := request(activeCaloriesRouter(&activeCaloriesUseCasesStub{}), test.method, test.path, test.body)
		if response.Code != http.StatusBadRequest {
			t.Fatalf("%s %s status=%d body=%s", test.method, test.path, response.Code, response.Body)
		}
	}
	notFound := request(activeCaloriesRouter(&activeCaloriesUseCasesStub{err: port.ErrActiveCaloriesNotFound}), http.MethodDelete, "/api/active-calories/2026-09-16", "")
	if notFound.Code != http.StatusNotFound {
		t.Fatalf("not found status=%d body=%s", notFound.Code, notFound.Body)
	}
}

const dateLayoutForTest = "2006-01-02"
