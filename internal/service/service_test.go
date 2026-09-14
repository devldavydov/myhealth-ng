package service

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestNewConfiguresServerAndRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	application := New(Config{
		Host:         "::1",
		Port:         8080,
		ClientOrigin: "http://localhost:5173",
	})

	if application.httpServer.Addr != "[::1]:8080" {
		t.Fatalf("address = %q", application.httpServer.Addr)
	}
	if application.httpServer.ReadHeaderTimeout != readHeaderTimeout ||
		application.httpServer.ReadTimeout != readTimeout ||
		application.httpServer.WriteTimeout != writeTimeout ||
		application.httpServer.IdleTimeout != idleTimeout {
		t.Fatal("HTTP server timeouts are not configured")
	}

	request := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	response := httptest.NewRecorder()
	application.httpServer.Handler.ServeHTTP(response, request)
	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
}
