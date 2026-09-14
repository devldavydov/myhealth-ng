package repository

import (
	"context"

	"github.com/devldavydov/myhealth-ng/internal/domain"
)

type UserRegistry interface {
	Remember(context.Context, domain.UserIdentity) error
}
