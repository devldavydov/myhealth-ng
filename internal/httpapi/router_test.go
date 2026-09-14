package httpapi_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/devldavydov/myhealth-ng/internal/domain"
	"github.com/devldavydov/myhealth-ng/internal/httpapi"
	"github.com/devldavydov/myhealth-ng/internal/repository"
)

func TestHealth(t *testing.T) {
	response := performRequest(t, http.MethodGet, "/api/health", "")
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	assertJSON(t, response.Body.String(), map[string]any{"status": "ok"})
}

func TestEmptyMeasurementListIsArray(t *testing.T) {
	response := performRequest(t, http.MethodGet, "/api/measurements", "")
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	assertJSON(t, response.Body.String(), map[string]any{"data": []any{}})
}

func TestCreateAndReturnMeasurement(t *testing.T) {
	body := `{"type":"weight","value":71.8,"unit":" кг ","measuredAt":"2026-09-13T10:00:00.000Z"}`
	response := performRequest(t, http.MethodPost, "/api/measurements", body)
	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body = %s", response.Code, http.StatusCreated, response.Body)
	}

	var result struct {
		Data domain.Measurement `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.Data.ID == "" || result.Data.Type != domain.MeasurementTypeWeight || result.Data.Value != 71.8 || result.Data.Unit != "кг" {
		t.Fatalf("unexpected measurement: %+v", result.Data)
	}
}

func TestRejectInvalidMeasurement(t *testing.T) {
	response := performRequest(t, http.MethodPost, "/api/measurements", `{"type":"unknown","value":"много"}`)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}

	var result struct {
		Error string `json:"error"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if result.Error != "Некорректные данные измерения" {
		t.Fatalf("error = %q", result.Error)
	}
}

func TestRequireCertificate(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := httpapi.NewRouter(
		repository.NewInMemoryMeasurementRepository([]domain.Measurement{}),
		repository.NewInMemoryUserRegistry(),
		httpapi.Options{CertificateRequired: true},
	)
	request := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusForbidden)
	}
}

func performRequest(t *testing.T, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)
	router := httpapi.NewRouter(
		repository.NewInMemoryMeasurementRepository([]domain.Measurement{}),
		repository.NewInMemoryUserRegistry(),
		httpapi.Options{},
	)
	request := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}

func assertJSON(t *testing.T, actual string, expected any) {
	t.Helper()
	var decoded any
	if err := json.Unmarshal([]byte(actual), &decoded); err != nil {
		t.Fatal(err)
	}
	expectedJSON, err := json.Marshal(expected)
	if err != nil {
		t.Fatal(err)
	}
	var decodedExpected any
	if err := json.Unmarshal(expectedJSON, &decodedExpected); err != nil {
		t.Fatal(err)
	}
	if !jsonEqual(decoded, decodedExpected) {
		t.Fatalf("JSON = %s, want %s", actual, expectedJSON)
	}
}

func jsonEqual(left, right any) bool {
	leftJSON, _ := json.Marshal(left)
	rightJSON, _ := json.Marshal(right)
	return string(leftJSON) == string(rightJSON)
}
