package port

import (
	"context"
	"errors"

	"github.com/devldavydov/myhealth-ng/internal/entity"
)

var (
	ErrFoodNotFound = errors.New("food not found")
	ErrFoodConflict = errors.New("food key conflict")
	ErrFoodInUse    = errors.New("food is used in bundle")
)

type FoodRepository interface {
	List(context.Context, string) ([]entity.Food, error)
	Get(context.Context, string) (entity.Food, error)
	Create(context.Context, entity.Food) (entity.Food, error)
	Update(context.Context, string, entity.FoodData) (entity.Food, error)
	Delete(context.Context, string) error
}

type FoodUseCases interface {
	List(context.Context, string) ([]entity.Food, error)
	Get(context.Context, string) (entity.Food, error)
	Create(context.Context, entity.FoodData) (entity.Food, error)
	Update(context.Context, string, entity.FoodData) (entity.Food, error)
	Delete(context.Context, string) error
}
