package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/devldavydov/myhealth-ng/internal/adapter/httpapi"
	postgresadapter "github.com/devldavydov/myhealth-ng/internal/adapter/postgres"
	"github.com/devldavydov/myhealth-ng/internal/cases"
)

const (
	readHeaderTimeout = 5 * time.Second
	readTimeout       = 10 * time.Second
	writeTimeout      = 15 * time.Second
	idleTimeout       = 60 * time.Second
	shutdownTimeout   = 10 * time.Second
	databaseTimeout   = 10 * time.Second
	migrationTimeout  = time.Minute
)

type Service struct {
	httpServer *http.Server
	database   *sql.DB
}

func New(config Config) (*Service, error) {
	gin.SetMode(gin.ReleaseMode)

	database, err := sql.Open("pgx", config.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("open PostgreSQL: %w", err)
	}
	connectionContext, cancelConnection := context.WithTimeout(context.Background(), databaseTimeout)
	defer cancelConnection()
	if err := database.PingContext(connectionContext); err != nil {
		database.Close()
		return nil, fmt.Errorf("connect PostgreSQL: %w", err)
	}
	migrationContext, cancelMigration := context.WithTimeout(context.Background(), migrationTimeout)
	defer cancelMigration()
	if err := postgresadapter.Migrate(migrationContext, database); err != nil {
		database.Close()
		return nil, err
	}

	foodRepository := postgresadapter.NewFoodRepository(database)
	weightRepository := postgresadapter.NewWeightRepository(database)
	foodCases := cases.NewFood(foodRepository)
	weightCases := cases.NewWeight(weightRepository)
	router := httpapi.NewRouter(foodCases, weightCases, httpapi.Options{
		CertificateRequired: config.RequireClientCertificate,
	})

	return &Service{
		httpServer: &http.Server{
			Addr:              net.JoinHostPort(config.Host, strconv.Itoa(config.Port)),
			Handler:           router,
			ReadHeaderTimeout: readHeaderTimeout,
			ReadTimeout:       readTimeout,
			WriteTimeout:      writeTimeout,
			IdleTimeout:       idleTimeout,
		},
		database: database,
	}, nil
}

func (service *Service) Run() error {
	defer service.database.Close()
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
