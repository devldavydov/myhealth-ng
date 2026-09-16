package port

import (
	"context"
	"errors"
	"time"

	"github.com/devldavydov/myhealth-ng/internal/entity"
)

var ErrActiveCaloriesNotFound = errors.New("active calories not found")

type ActiveCaloriesRepository interface {
	Get(context.Context, string, time.Time) (*entity.ActiveCalories, error)
	Save(context.Context, string, entity.ActiveCalories) (entity.ActiveCalories, error)
	Delete(context.Context, string, time.Time) error
}

type ActiveCaloriesUseCases interface {
	Get(context.Context, string, time.Time) (*entity.ActiveCalories, error)
	Save(context.Context, string, entity.ActiveCalories) (entity.ActiveCalories, error)
	Delete(context.Context, string, time.Time) error
}
