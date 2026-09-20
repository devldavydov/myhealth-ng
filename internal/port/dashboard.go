package port

import (
	"context"

	"github.com/devldavydov/myhealth-ng/internal/entity"
)

type DashboardRepository interface {
	Load(context.Context, string, entity.DashboardRange) (entity.DashboardSource, error)
}

type DashboardUseCases interface {
	Get(context.Context, string, entity.DashboardRange) (entity.Dashboard, error)
}
