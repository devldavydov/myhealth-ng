package port

import (
	"context"

	"github.com/devldavydov/myhealth-ng/internal/entity"
)

type SettingsRepository interface {
	Get(context.Context, string) (*entity.UserSettings, error)
	Save(context.Context, string, entity.UserSettings) (entity.UserSettings, error)
}

type SettingsUseCases interface {
	Get(context.Context, string) (*entity.UserSettings, error)
	Save(context.Context, string, entity.UserSettings) (entity.UserSettings, error)
}
