package service

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/devldavydov/myhealth-ng/internal/httpapi"
	"github.com/devldavydov/myhealth-ng/internal/repository"
)

const (
	readHeaderTimeout = 5 * time.Second
	readTimeout       = 10 * time.Second
	writeTimeout      = 15 * time.Second
	idleTimeout       = 60 * time.Second
	shutdownTimeout   = 10 * time.Second
)

type Service struct {
	httpServer *http.Server
}

func New(config Config) *Service {
	if os.Getenv(gin.EnvGinMode) == "" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := httpapi.NewRouter(
		repository.NewInMemoryMeasurementRepository(nil),
		repository.NewInMemoryUserRegistry(),
		httpapi.Options{
			CertificateRequired: config.RequireClientCertificate,
			ClientOrigin:        config.ClientOrigin,
		},
	)

	return &Service{
		httpServer: &http.Server{
			Addr:              net.JoinHostPort(config.Host, strconv.Itoa(config.Port)),
			Handler:           router,
			ReadHeaderTimeout: readHeaderTimeout,
			ReadTimeout:       readTimeout,
			WriteTimeout:      writeTimeout,
			IdleTimeout:       idleTimeout,
		},
	}
}

func (service *Service) Run() error {
	shutdownSignal, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	shutdownDone := make(chan error, 1)
	go func() {
		<-shutdownSignal.Done()
		shutdownContext, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		shutdownDone <- service.httpServer.Shutdown(shutdownContext)
	}()

	log.Printf("MyHealth API: http://%s", service.httpServer.Addr)
	err := service.httpServer.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("запуск MyHealth API: %w", err)
	}

	if shutdownSignal.Err() != nil {
		if err := <-shutdownDone; err != nil {
			return fmt.Errorf("graceful shutdown MyHealth API: %w", err)
		}
	}
	return nil
}
