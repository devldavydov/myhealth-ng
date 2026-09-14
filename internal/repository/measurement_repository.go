package repository

import (
	"context"

	"github.com/devldavydov/myhealth-ng/internal/domain"
)

type MeasurementRepository interface {
	FindAll(context.Context) ([]domain.Measurement, error)
	Create(context.Context, domain.NewMeasurement) (domain.Measurement, error)
}
