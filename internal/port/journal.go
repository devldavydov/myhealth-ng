package port

import (
	"context"
	"errors"
	"time"

	"github.com/devldavydov/myhealth-ng/internal/entity"
)

var (
	ErrJournalFoodNotFound = errors.New("journal food not found")
	ErrJournalItemNotFound = errors.New("journal item not found")
)

type JournalRepository interface {
	List(context.Context, string, time.Time) ([]entity.JournalItem, error)
	Save(context.Context, string, time.Time, entity.MealType, []entity.JournalItemData) error
	DeleteItem(context.Context, string, time.Time, entity.MealType, string) error
	ClearMeal(context.Context, string, time.Time, entity.MealType) error
}

type JournalUseCases interface {
	Get(context.Context, string, time.Time) (entity.JournalDay, error)
	Save(context.Context, string, time.Time, entity.MealType, []entity.JournalItemData) (entity.JournalDay, error)
	DeleteItem(context.Context, string, time.Time, entity.MealType, string) (entity.JournalDay, error)
	ClearMeal(context.Context, string, time.Time, entity.MealType) (entity.JournalDay, error)
}
