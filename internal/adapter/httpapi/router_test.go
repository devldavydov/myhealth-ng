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
	request     entity.PageRequest
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
	request     entity.PageRequest
	items       []entity.BundleSummary
	createdData entity.BundleData
	updatedKey  string
	updatedData entity.BundleData
	deletedKey  string
	err         error
}

type settingsUseCasesStub struct {
	userID string
	item   *entity.UserSettings
	saved  entity.UserSettings
	err    error
}

type journalUseCasesStub struct {
	userID  string
	dt      time.Time
	meal    entity.MealType
	items   []entity.JournalItemData
	foodKey string
	day     entity.JournalDay
	err     error
}

func (stub *journalUseCasesStub) Get(_ context.Context, userID string, dt time.Time) (entity.JournalDay, error) {
	stub.userID, stub.dt = userID, dt
	return stub.day, stub.err
}

func (stub *journalUseCasesStub) Save(_ context.Context, userID string, dt time.Time, meal entity.MealType, items []entity.JournalItemData) (entity.JournalDay, error) {
	stub.userID, stub.dt, stub.meal, stub.items = userID, dt, meal, items
	return stub.day, stub.err
}

func (stub *journalUseCasesStub) DeleteItem(_ context.Context, userID string, dt time.Time, meal entity.MealType, foodKey string) (entity.JournalDay, error) {
	stub.userID, stub.dt, stub.meal, stub.foodKey = userID, dt, meal, foodKey
	return stub.day, stub.err
}

func (stub *journalUseCasesStub) ClearMeal(_ context.Context, userID string, dt time.Time, meal entity.MealType) (entity.JournalDay, error) {
	stub.userID, stub.dt, stub.meal = userID, dt, meal
	return stub.day, stub.err
}

func (stub *settingsUseCasesStub) Get(_ context.Context, userID string) (*entity.UserSettings, error) {
	stub.userID = userID
	return stub.item, stub.err
}

func (stub *settingsUseCasesStub) Save(_ context.Context, userID string, data entity.UserSettings) (entity.UserSettings, error) {
	stub.userID, stub.saved = userID, data
	return data, stub.err
}

func (stub *bundleUseCasesStub) List(_ context.Context, request entity.PageRequest) (entity.Page[entity.BundleSummary], error) {
	if request.Page == 0 {
		request.Page = entity.DefaultPage
	}
	if request.PageSize == 0 {
		request.PageSize = entity.DefaultPageSize
	}
	stub.request = request
	return entity.Page[entity.BundleSummary]{Items: stub.items, Page: request.Page, PageSize: request.PageSize, Total: len(stub.items), TotalPages: 1}, stub.err
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

func (stub *foodUseCasesStub) List(_ context.Context, request entity.PageRequest) (entity.Page[entity.Food], error) {
	if request.Page == 0 {
		request.Page = entity.DefaultPage
	}
	if request.PageSize == 0 {
		request.PageSize = entity.DefaultPageSize
	}
	stub.request = request
	return entity.Page[entity.Food]{Items: stub.items, Page: request.Page, PageSize: request.PageSize, Total: len(stub.items), TotalPages: 1}, stub.err
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

	list := request(router, http.MethodGet, "/api/food?q=%D1%82%D0%B2%D0%BE%D1%80%D0%BE%D0%B3&page=2&pageSize=50", "")
	if list.Code != http.StatusOK || stub.request.Query != "творог" || stub.request.Page != 2 || stub.request.PageSize != 50 {
		t.Fatalf("list status=%d request=%+v body=%s", list.Code, stub.request, list.Body)
	}
	var listed struct {
		Data       []map[string]any `json:"data"`
		Pagination struct {
			Page       int `json:"page"`
			PageSize   int `json:"pageSize"`
			Total      int `json:"total"`
			TotalPages int `json:"totalPages"`
		} `json:"pagination"`
	}
	if err := json.Unmarshal(list.Body.Bytes(), &listed); err != nil || len(listed.Data) != 1 || listed.Data[0]["key"] != foodKey || listed.Pagination.Page != 2 || listed.Pagination.PageSize != 50 || listed.Pagination.Total != 1 || listed.Pagination.TotalPages != 1 {
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
	for _, path := range []string{"/api/food?page=wrong", "/api/food?pageSize=10.5"} {
		response := request(newRouter(&foodUseCasesStub{}), http.MethodGet, path, "")
		if response.Code != http.StatusBadRequest {
			t.Fatalf("path=%s status=%d body=%s", path, response.Code, response.Body)
		}
	}

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
	router := httpapi.NewRouter(&foodUseCasesStub{}, stub, &bundleUseCasesStub{}, &settingsUseCasesStub{}, &journalUseCasesStub{}, &activeCaloriesUseCasesStub{}, httpapi.Options{})

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
	router := httpapi.NewRouter(&foodUseCasesStub{}, &weightUseCasesStub{}, &bundleUseCasesStub{}, &settingsUseCasesStub{}, &journalUseCasesStub{}, &activeCaloriesUseCasesStub{}, httpapi.Options{})
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

	notFound := httpapi.NewRouter(&foodUseCasesStub{}, &weightUseCasesStub{err: port.ErrWeightNotFound}, &bundleUseCasesStub{}, &settingsUseCasesStub{}, &journalUseCasesStub{}, &activeCaloriesUseCasesStub{}, httpapi.Options{})
	if response := request(notFound, http.MethodDelete, "/api/weight/2026-09-15", ""); response.Code != http.StatusNotFound {
		t.Fatalf("not found status=%d body=%s", response.Code, response.Body)
	}
}

func TestBundleRoutes(t *testing.T) {
	stub := &bundleUseCasesStub{items: []entity.BundleSummary{{
		Key: bundleKey, Name: "Завтрак", ItemCount: 1, Totals: sampleBundle.Totals,
	}}}
	router := httpapi.NewRouter(&foodUseCasesStub{}, &weightUseCasesStub{}, stub, &settingsUseCasesStub{}, &journalUseCasesStub{}, &activeCaloriesUseCasesStub{}, httpapi.Options{})

	listed := request(router, http.MethodGet, "/api/bundle?q=%D0%B7%D0%B0%D0%B2%D1%82%D1%80%D0%B0%D0%BA", "")
	if listed.Code != http.StatusOK || stub.request.Query != "завтрак" || stub.request.Page != 1 || stub.request.PageSize != 20 || !strings.Contains(listed.Body.String(), "\"itemCount\":1") {
		t.Fatalf("list status=%d request=%+v body=%s", listed.Code, stub.request, listed.Body)
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
	router := httpapi.NewRouter(&foodUseCasesStub{}, &weightUseCasesStub{}, &bundleUseCasesStub{}, &settingsUseCasesStub{}, &journalUseCasesStub{}, &activeCaloriesUseCasesStub{}, httpapi.Options{})
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
		errorRouter := httpapi.NewRouter(&foodUseCasesStub{}, &weightUseCasesStub{}, &bundleUseCasesStub{err: test.err}, &settingsUseCasesStub{}, &journalUseCasesStub{}, &activeCaloriesUseCasesStub{}, httpapi.Options{})
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

func TestSettingsRoutesUseCurrentUser(t *testing.T) {
	limit := 2300
	stub := &settingsUseCasesStub{item: &entity.UserSettings{DefaultDailyCalorieLimit: limit}}
	router := httpapi.NewRouter(&foodUseCasesStub{}, &weightUseCasesStub{}, &bundleUseCasesStub{}, stub, &journalUseCasesStub{}, &activeCaloriesUseCasesStub{}, httpapi.Options{})

	loaded := request(router, http.MethodGet, "/api/settings", "")
	if loaded.Code != http.StatusOK || stub.userID != "00000000-0000-4000-8000-000000000000" || !strings.Contains(loaded.Body.String(), `"defaultDailyCalorieLimit":2300`) {
		t.Fatalf("get status=%d user=%q body=%s", loaded.Code, stub.userID, loaded.Body)
	}

	saved := request(router, http.MethodPut, "/api/settings", `{"defaultDailyCalorieLimit":2450}`)
	if saved.Code != http.StatusOK || stub.saved.DefaultDailyCalorieLimit != 2450 {
		t.Fatalf("save status=%d data=%+v body=%s", saved.Code, stub.saved, saved.Body)
	}
}

func TestSettingsEmptyAndValidation(t *testing.T) {
	router := httpapi.NewRouter(&foodUseCasesStub{}, &weightUseCasesStub{}, &bundleUseCasesStub{}, &settingsUseCasesStub{}, &journalUseCasesStub{}, &activeCaloriesUseCasesStub{}, httpapi.Options{})
	empty := request(router, http.MethodGet, "/api/settings", "")
	if empty.Code != http.StatusOK || !strings.Contains(empty.Body.String(), `"defaultDailyCalorieLimit":null`) {
		t.Fatalf("empty status=%d body=%s", empty.Code, empty.Body)
	}

	for _, body := range []string{
		`{}`,
		`{"defaultDailyCalorieLimit":null}`,
		`{"defaultDailyCalorieLimit":2000.5}`,
		`{"defaultDailyCalorieLimit":2000,"extra":true}`,
	} {
		response := request(router, http.MethodPut, "/api/settings", body)
		if response.Code != http.StatusBadRequest {
			t.Fatalf("body=%s status=%d response=%s", body, response.Code, response.Body)
		}
	}

	validation := &entity.ValidationError{Fields: map[string][]string{
		"defaultDailyCalorieLimit": {"Введите целое число от 1 до 10000"},
	}}
	invalidRouter := httpapi.NewRouter(&foodUseCasesStub{}, &weightUseCasesStub{}, &bundleUseCasesStub{}, &settingsUseCasesStub{err: validation}, &journalUseCasesStub{}, &activeCaloriesUseCasesStub{}, httpapi.Options{})
	response := request(invalidRouter, http.MethodPut, "/api/settings", `{"defaultDailyCalorieLimit":0}`)
	if response.Code != http.StatusBadRequest || !strings.Contains(response.Body.String(), "defaultDailyCalorieLimit") {
		t.Fatalf("validation status=%d body=%s", response.Code, response.Body)
	}
}

func TestJournalRoutesUseCurrentUser(t *testing.T) {
	dt := time.Date(2026, time.September, 16, 0, 0, 0, 0, time.UTC)
	day := entity.JournalDay{
		DT: dt,
		Zones: []entity.JournalZone{{
			Meal:   entity.MealBreakfast,
			Items:  []entity.JournalItem{{Meal: entity.MealBreakfast, Food: sampleFood, Weight: 200}},
			Totals: entity.JournalTotals{Weight: 200, Cal: 240, Protein: 36, Fat: 10, Carb: 6},
		}},
		Totals:       entity.JournalTotals{Weight: 200, Cal: 240, Protein: 36, Fat: 10, Carb: 6},
		MacroPercent: entity.MacroPercent{Protein: 69.23, Fat: 19.23, Carb: 11.54},
	}
	stub := &journalUseCasesStub{day: day}
	router := httpapi.NewRouter(&foodUseCasesStub{}, &weightUseCasesStub{}, &bundleUseCasesStub{}, &settingsUseCasesStub{}, stub, &activeCaloriesUseCasesStub{}, httpapi.Options{})

	loaded := request(router, http.MethodGet, "/api/journal?dt=2026-09-16", "")
	if loaded.Code != http.StatusOK || stub.userID != "00000000-0000-4000-8000-000000000000" || stub.dt.Format("2006-01-02") != "2026-09-16" || !strings.Contains(loaded.Body.String(), `"macroPercent"`) {
		t.Fatalf("get status=%d stub=%+v body=%s", loaded.Code, stub, loaded.Body)
	}
	body := `{"dt":"2026-09-16","meal":"завтрак","items":[{"foodKey":"` + foodKey + `","weight":150}]}`
	saved := request(router, http.MethodPost, "/api/journal", body)
	if saved.Code != http.StatusOK || stub.meal != entity.MealBreakfast || len(stub.items) != 1 || stub.items[0].Weight != 150 {
		t.Fatalf("save status=%d stub=%+v body=%s", saved.Code, stub, saved.Body)
	}
	mealPath := "%D0%B7%D0%B0%D0%B2%D1%82%D1%80%D0%B0%D0%BA"
	deleted := request(router, http.MethodDelete, "/api/journal/2026-09-16/"+mealPath+"/"+foodKey, "")
	if deleted.Code != http.StatusOK || stub.foodKey != foodKey {
		t.Fatalf("delete status=%d stub=%+v body=%s", deleted.Code, stub, deleted.Body)
	}
	cleared := request(router, http.MethodDelete, "/api/journal/2026-09-16/"+mealPath, "")
	if cleared.Code != http.StatusOK || stub.meal != entity.MealBreakfast {
		t.Fatalf("clear status=%d stub=%+v body=%s", cleared.Code, stub, cleared.Body)
	}
}

func TestJournalValidationAndErrors(t *testing.T) {
	router := httpapi.NewRouter(&foodUseCasesStub{}, &weightUseCasesStub{}, &bundleUseCasesStub{}, &settingsUseCasesStub{}, &journalUseCasesStub{}, &activeCaloriesUseCasesStub{}, httpapi.Options{})
	for _, test := range []struct {
		method string
		path   string
		body   string
	}{
		{http.MethodGet, "/api/journal?dt=wrong", ""},
		{http.MethodPost, "/api/journal", `{}`},
		{http.MethodPost, "/api/journal", `{"dt":"2026-09-16","meal":"завтрак","items":[{}]}`},
		{http.MethodDelete, "/api/journal/wrong/%D0%B7%D0%B0%D0%B2%D1%82%D1%80%D0%B0%D0%BA/" + foodKey, ""},
	} {
		response := request(router, test.method, test.path, test.body)
		if response.Code != http.StatusBadRequest {
			t.Fatalf("%s %s status=%d body=%s", test.method, test.path, response.Code, response.Body)
		}
	}
	notFound := httpapi.NewRouter(&foodUseCasesStub{}, &weightUseCasesStub{}, &bundleUseCasesStub{}, &settingsUseCasesStub{}, &journalUseCasesStub{err: port.ErrJournalItemNotFound}, &activeCaloriesUseCasesStub{}, httpapi.Options{})
	response := request(notFound, http.MethodDelete, "/api/journal/2026-09-16/%D0%B7%D0%B0%D0%B2%D1%82%D1%80%D0%B0%D0%BA/"+foodKey, "")
	if response.Code != http.StatusNotFound {
		t.Fatalf("not found status=%d body=%s", response.Code, response.Body)
	}
}

func TestRequireCertificate(t *testing.T) {
	router := httpapi.NewRouter(&foodUseCasesStub{}, &weightUseCasesStub{}, &bundleUseCasesStub{}, &settingsUseCasesStub{}, &journalUseCasesStub{}, &activeCaloriesUseCasesStub{}, httpapi.Options{CertificateRequired: true})
	if response := request(router, http.MethodGet, "/api/food", ""); response.Code != http.StatusForbidden {
		t.Fatalf("status=%d body=%s", response.Code, response.Body)
	}
}

func newRouter(stub port.FoodUseCases) http.Handler {
	gin.SetMode(gin.TestMode)
	return httpapi.NewRouter(stub, &weightUseCasesStub{}, &bundleUseCasesStub{}, &settingsUseCasesStub{}, &journalUseCasesStub{}, &activeCaloriesUseCasesStub{}, httpapi.Options{})
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
