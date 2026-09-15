package port

import (
	"context"
	"errors"

	"github.com/devldavydov/myhealth-ng/internal/entity"
)

var ErrWeightNotFound = errors.New("weight not found")

type WeightRepository interface {
	List(context.Context, string, entity.WeightRange) ([]entity.Weight, error)
	Save(context.Context, string, entity.Weight) (entity.Weight, error)
	Delete(context.Context, string, entity.Weight) error
}

type WeightUseCases interface {
	List(context.Context, string, entity.WeightRange) ([]entity.Weight, error)
	Save(context.Context, string, entity.Weight) (entity.Weight, error)
	Delete(context.Context, string, entity.Weight) error
}
