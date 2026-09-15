package httpapi_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/devldavydov/myhealth-ng/internal/adapter/httpapi"
	"github.com/devldavydov/myhealth-ng/internal/entity"
	"github.com/devldavydov/myhealth-ng/internal/port"
)

const foodKey = "3f67c05f-7c9e-4cb5-b26a-f9ce5b065865"
const bundleKey = "8c2cf7aa-44f4-49a4-aef0-e9087088c980"

var sampleFood = entity.Food{
	Key: foodKey, Name: "Творог", Brand: "Ферма", Cal100: 120,
	Prot100: 18, Fat100: 5, Carb100: 3, Comment: "5%",
}

var sampleBundle = entity.Bundle{
	Key: bundleKey, Name: "Завтрак",
	Items:  []entity.BundleItem{{Food: sampleFood, Weight: 200}},
	Totals: entity.BundleTotals{Weight: 200, Cal: 240, Protein: 36, Fat: 10, Carb: 6},
}

type foodUseCasesStub struct {
	items       []entity.Food
	query       string
	createdData entity.FoodData
	updatedKey  string
	updatedData entity.FoodData
	deletedKey  string
	err         error
}

type weightUseCasesStub struct {
	userID  string
	period  entity.WeightRange
	items   []entity.Weight
	saved   entity.Weight
	deleted entity.Weight
	err     error
}

type bundleUseCasesStub struct {
	query       string
	items       []entity.BundleSummary
	createdData entity.BundleData
	updatedKey  string
	updatedData entity.BundleData
	deletedKey  string
	err         error
}

func (stub *bundleUseCasesStub) List(_ context.Context, query string) ([]entity.BundleSummary, error) {
	stub.query = query
	return stub.items, stub.err
}
func (stub *bundleUseCasesStub) Get(_ context.Context, _ string) (entity.Bundle, error) {
	return sampleBundle, stub.err
}
func (stub *bundleUseCasesStub) Create(_ context.Context, data entity.BundleData) (entity.Bundle, error) {
	stub.createdData = data
	return sampleBundle, stub.err
}
func (stub *bundleUseCasesStub) Update(_ context.Context, key string, data entity.BundleData) (entity.Bundle, error) {
	stub.updatedKey, stub.updatedData = key, data
	return sampleBundle, stub.err
}
func (stub *bundleUseCasesStub) Delete(_ context.Context, key string) error {
	stub.deletedKey = key
	return stub.err
}

func (stub *weightUseCasesStub) List(_ context.Context, userID string, period entity.WeightRange) ([]entity.Weight, error) {
	stub.userID, stub.period = userID, period
	return stub.items, stub.err
}
func (stub *weightUseCasesStub) Save(_ context.Context, userID string, data entity.Weight) (entity.Weight, error) {
	stub.userID, stub.saved = userID, data
	return data, stub.err
}
func (stub *weightUseCasesStub) Delete(_ context.Context, userID string, data entity.Weight) error {
	stub.userID, stub.deleted = userID, data
	return stub.err
}

func (stub *foodUseCasesStub) List(_ context.Context, query string) ([]entity.Food, error) {
	stub.query = query
	return stub.items, stub.err
}
func (stub *foodUseCasesStub) Get(_ context.Context, _ string) (entity.Food, error) {
	return sampleFood, stub.err
}
func (stub *foodUseCasesStub) Create(_ context.Context, data entity.FoodData) (entity.Food, error) {
	stub.createdData = data
	return sampleFood, stub.err
}
func (stub *foodUseCasesStub) Update(_ context.Context, key string, data entity.FoodData) (entity.Food, error) {
	stub.updatedKey, stub.updatedData = key, data
	return sampleFood, stub.err
}
func (stub *foodUseCasesStub) Delete(_ context.Context, key string) error {
	stub.deletedKey = key
	return stub.err
}

func TestFoodRoutes(t *testing.T) {
	stub := &foodUseCasesStub{items: []entity.Food{sampleFood}}
	router := newRouter(stub)

	list := request(router, http.MethodGet, "/api/food?q=%D1%82%D0%B2%D0%BE%D1%80%D0%BE%D0%B3", "")
	if list.Code != http.StatusOK || stub.query != "творог" {
		t.Fatalf("list status=%d query=%q body=%s", list.Code, stub.query, list.Body)
	}
	var listed struct {
		Data []map[string]any `json:"data"`
	}
	if err := json.Unmarshal(list.Body.Bytes(), &listed); err != nil || len(listed.Data) != 1 || listed.Data[0]["key"] != foodKey {
		t.Fatalf("unexpected list: %s, err=%v", list.Body, err)
	}

	body := `{"name":"Творог","brand":"Ферма","cal100":120,"prot100":18,"fat100":5,"carb100":3,"comment":"5%"}`
	created := request(router, http.MethodPost, "/api/food", body)
	if created.Code != http.StatusCreated || stub.createdData.Name != "Творог" {
		t.Fatalf("create status=%d data=%+v body=%s", created.Code, stub.createdData, created.Body)
	}
	updated := request(router, http.MethodPut, "/api/food/"+foodKey, body)
	if updated.Code != http.StatusOK || stub.updatedKey != foodKey {
		t.Fatalf("update status=%d key=%q body=%s", updated.Code, stub.updatedKey, updated.Body)
	}
	deleted := request(router, http.MethodDelete, "/api/food/"+foodKey, "")
	if deleted.Code != http.StatusNoContent || stub.deletedKey != foodKey {
		t.Fatalf("delete status=%d key=%q", deleted.Code, stub.deletedKey)
	}
}

func TestFoodValidationAndErrors(t *testing.T) {
	invalid := request(newRouter(&foodUseCasesStub{}), http.MethodPost, "/api/food", `{"name":"x"}`)
	if invalid.Code != http.StatusBadRequest {
		t.Fatalf("invalid status=%d body=%s", invalid.Code, invalid.Body)
	}

	for _, test := range []struct {
		name   string
		err    error
		status int
	}{
		{"not found", port.ErrFoodNotFound, http.StatusNotFound},
		{"conflict", port.ErrFoodConflict, http.StatusConflict},
		{"internal", errors.New("database unavailable"), http.StatusInternalServerError},
	} {
		t.Run(test.name, func(t *testing.T) {
			response := request(newRouter(&foodUseCasesStub{err: test.err}), http.MethodGet, "/api/food/"+foodKey, "")
			if response.Code != test.status {
				t.Fatalf("status=%d want=%d body=%s", response.Code, test.status, response.Body)
			}
		})
	}
}

func TestMeAndRemovedRoutes(t *testing.T) {
	router := newRouter(&foodUseCasesStub{})
	me := request(router, http.MethodGet, "/api/me", "")
	if me.Code != http.StatusOK || !strings.Contains(me.Body.String(), "Локальный пользователь") {
		t.Fatalf("me status=%d body=%s", me.Code, me.Body)
	}
	for _, path := range []string{"/api/health", "/api/measurements"} {
		if response := request(router, http.MethodGet, path, ""); response.Code != http.StatusNotFound {
			t.Fatalf("%s status=%d", path, response.Code)
		}
	}
}

func TestWeightRoutesUseCurrentUserAndDateRange(t *testing.T) {
	stub := &weightUseCasesStub{items: []entity.Weight{{
		DT: time.Date(2026, time.September, 15, 0, 0, 0, 0, time.UTC), Value: 82.4,
	}}}
	router := httpapi.NewRouter(&foodUseCasesStub{}, stub, &bundleUseCasesStub{}, httpapi.Options{})

	listed := request(router, http.MethodGet, "/api/weight?from=2026-03-15&to=2026-09-15", "")
	if listed.Code != http.StatusOK || stub.userID != "00000000-0000-4000-8000-000000000000" {
		t.Fatalf("list status=%d user=%q body=%s", listed.Code, stub.userID, listed.Body)
	}
	if stub.period.From == nil || stub.period.To == nil || stub.period.From.Format("2006-01-02") != "2026-03-15" {
		t.Fatalf("period=%+v", stub.period)
	}
	if !strings.Contains(listed.Body.String(), `"dt":"2026-09-15"`) {
		t.Fatalf("body=%s", listed.Body)
	}

	saved := request(router, http.MethodPost, "/api/weight", `{"dt":"2026-09-15","value":81.7}`)
	if saved.Code != http.StatusOK || stub.saved.Value != 81.7 || stub.saved.DT.Format("2006-01-02") != "2026-09-15" {
		t.Fatalf("save status=%d data=%+v body=%s", saved.Code, stub.saved, saved.Body)
	}

	deleted := request(router, http.MethodDelete, "/api/weight/2026-09-15", "")
	if deleted.Code != http.StatusNoContent || stub.deleted.DT.Format("2006-01-02") != "2026-09-15" {
		t.Fatalf("delete status=%d data=%+v", deleted.Code, stub.deleted)
	}
}

func TestWeightValidationAndErrors(t *testing.T) {
	router := httpapi.NewRouter(&foodUseCasesStub{}, &weightUseCasesStub{}, &bundleUseCasesStub{}, httpapi.Options{})
	for _, requestData := range []struct {
		method string
		path   string
		body   string
	}{
		{http.MethodGet, "/api/weight?from=wrong", ""},
		{http.MethodPost, "/api/weight", `{"dt":"2026-09-15"}`},
		{http.MethodDelete, "/api/weight/wrong", ""},
	} {
		response := request(router, requestData.method, requestData.path, requestData.body)
		if response.Code != http.StatusBadRequest {
			t.Fatalf("%s %s status=%d body=%s", requestData.method, requestData.path, response.Code, response.Body)
		}
	}

	notFound := httpapi.NewRouter(&foodUseCasesStub{}, &weightUseCasesStub{err: port.ErrWeightNotFound}, &bundleUseCasesStub{}, httpapi.Options{})
	if response := request(notFound, http.MethodDelete, "/api/weight/2026-09-15", ""); response.Code != http.StatusNotFound {
		t.Fatalf("not found status=%d body=%s", response.Code, response.Body)
	}
}

func TestBundleRoutes(t *testing.T) {
	stub := &bundleUseCasesStub{items: []entity.BundleSummary{{
		Key: bundleKey, Name: "Завтрак", ItemCount: 1, Totals: sampleBundle.Totals,
	}}}
	router := httpapi.NewRouter(&foodUseCasesStub{}, &weightUseCasesStub{}, stub, httpapi.Options{})

	listed := request(router, http.MethodGet, "/api/bundle?q=%D0%B7%D0%B0%D0%B2%D1%82%D1%80%D0%B0%D0%BA", "")
	if listed.Code != http.StatusOK || stub.query != "завтрак" || !strings.Contains(listed.Body.String(), "\"itemCount\":1") {
		t.Fatalf("list status=%d query=%q body=%s", listed.Code, stub.query, listed.Body)
	}
	loaded := request(router, http.MethodGet, "/api/bundle/"+bundleKey, "")
	if loaded.Code != http.StatusOK || !strings.Contains(loaded.Body.String(), "\"food\":{\"key\":\""+foodKey) {
		t.Fatalf("get status=%d body=%s", loaded.Code, loaded.Body)
	}
	body := "{\"name\":\"Завтрак\",\"items\":[{\"foodKey\":\"" + foodKey + "\",\"weight\":200}]}"
	created := request(router, http.MethodPost, "/api/bundle", body)
	if created.Code != http.StatusCreated || stub.createdData.Name != "Завтрак" || stub.createdData.Items[0].Weight != 200 {
		t.Fatalf("create status=%d data=%+v body=%s", created.Code, stub.createdData, created.Body)
	}
	updated := request(router, http.MethodPut, "/api/bundle/"+bundleKey, body)
	if updated.Code != http.StatusOK || stub.updatedKey != bundleKey {
		t.Fatalf("update status=%d key=%q body=%s", updated.Code, stub.updatedKey, updated.Body)
	}
	deleted := request(router, http.MethodDelete, "/api/bundle/"+bundleKey, "")
	if deleted.Code != http.StatusNoContent || stub.deletedKey != bundleKey {
		t.Fatalf("delete status=%d key=%q", deleted.Code, stub.deletedKey)
	}
}

func TestBundleValidationAndErrors(t *testing.T) {
	router := httpapi.NewRouter(&foodUseCasesStub{}, &weightUseCasesStub{}, &bundleUseCasesStub{}, httpapi.Options{})
	if response := request(router, http.MethodPost, "/api/bundle", "{\"name\":\"Тест\"}"); response.Code != http.StatusBadRequest {
		t.Fatalf("validation status=%d body=%s", response.Code, response.Body)
	}
	for _, test := range []struct {
		err    error
		status int
	}{
		{port.ErrBundleNotFound, http.StatusNotFound},
		{port.ErrBundleConflict, http.StatusConflict},
		{errors.New("database unavailable"), http.StatusInternalServerError},
	} {
		errorRouter := httpapi.NewRouter(&foodUseCasesStub{}, &weightUseCasesStub{}, &bundleUseCasesStub{err: test.err}, httpapi.Options{})
		if response := request(errorRouter, http.MethodGet, "/api/bundle/"+bundleKey, ""); response.Code != test.status {
			t.Fatalf("error=%v status=%d want=%d body=%s", test.err, response.Code, test.status, response.Body)
		}
	}
}

func TestDeleteUsedFoodReturnsConflict(t *testing.T) {
	router := newRouter(&foodUseCasesStub{err: port.ErrFoodInUse})
	response := request(router, http.MethodDelete, "/api/food/"+foodKey, "")
	if response.Code != http.StatusConflict || !strings.Contains(response.Body.String(), "используется в бандле") {
		t.Fatalf("status=%d body=%s", response.Code, response.Body)
	}
}

func TestRequireCertificate(t *testing.T) {
	router := httpapi.NewRouter(&foodUseCasesStub{}, &weightUseCasesStub{}, &bundleUseCasesStub{}, httpapi.Options{CertificateRequired: true})
	if response := request(router, http.MethodGet, "/api/food", ""); response.Code != http.StatusForbidden {
		t.Fatalf("status=%d body=%s", response.Code, response.Body)
	}
}

func newRouter(stub port.FoodUseCases) http.Handler {
	gin.SetMode(gin.TestMode)
	return httpapi.NewRouter(stub, &weightUseCasesStub{}, &bundleUseCasesStub{}, httpapi.Options{})
}

func request(handler http.Handler, method, path, body string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, req)
	return response
}
