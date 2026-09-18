package port

import (
	"context"
	"errors"

	"github.com/devldavydov/myhealth-ng/internal/entity"
)

var (
	ErrSportNotFound         = errors.New("sport not found")
	ErrSportConflict         = errors.New("sport key conflict")
	ErrSportInUse            = errors.New("sport is used in activity")
	ErrSportActivityNotFound = errors.New("sport activity not found")
)

type SportRepository interface {
	List(context.Context, entity.PageRequest) (entity.Page[entity.Sport], error)
	Get(context.Context, string) (entity.Sport, error)
	Create(context.Context, entity.Sport) (entity.Sport, error)
	Update(context.Context, string, entity.SportData) (entity.Sport, error)
	Delete(context.Context, string) error
}

type SportUseCases interface {
	List(context.Context, entity.PageRequest) (entity.Page[entity.Sport], error)
	Get(context.Context, string) (entity.Sport, error)
	Create(context.Context, entity.SportData) (entity.Sport, error)
	Update(context.Context, string, entity.SportData) (entity.Sport, error)
	Delete(context.Context, string) error
}

type SportActivityRepository interface {
	List(context.Context, string, entity.SportActivityRange) ([]entity.SportActivity, error)
	Save(context.Context, string, entity.SportActivityData) (entity.SportActivity, error)
	Delete(context.Context, string, entity.SportActivityData) error
}

type SportActivityUseCases interface {
	List(context.Context, string, entity.SportActivityRange) ([]entity.SportActivity, error)
	Save(context.Context, string, entity.SportActivityData) (entity.SportActivity, error)
	Delete(context.Context, string, entity.SportActivityData) error
}
