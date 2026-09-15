package port

import (
	"context"
	"errors"

	"github.com/devldavydov/myhealth-ng/internal/entity"
)

var (
	ErrBundleNotFound     = errors.New("bundle not found")
	ErrBundleConflict     = errors.New("bundle key conflict")
	ErrBundleFoodNotFound = errors.New("bundle food not found")
)

type BundleRepository interface {
	List(context.Context, string) ([]entity.BundleSummary, error)
	Get(context.Context, string) (entity.Bundle, error)
	Create(context.Context, entity.Bundle) (entity.Bundle, error)
	Update(context.Context, string, entity.BundleData) (entity.Bundle, error)
	Delete(context.Context, string) error
}

type BundleUseCases interface {
	List(context.Context, string) ([]entity.BundleSummary, error)
	Get(context.Context, string) (entity.Bundle, error)
	Create(context.Context, entity.BundleData) (entity.Bundle, error)
	Update(context.Context, string, entity.BundleData) (entity.Bundle, error)
	Delete(context.Context, string) error
}
