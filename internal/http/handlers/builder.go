package handlers

import (
	"context"
	"regexp"

	"github.com/appclacks/server/pkg/healthcheck/aggregates"
)

type HealthcheckService interface {
	UpdateHealthcheck(ctx context.Context, healthcheck *aggregates.Healthcheck) error
	CreateHealthcheck(ctx context.Context, healthcheck *aggregates.Healthcheck) error
	GetHealthcheck(ctx context.Context, id string) (*aggregates.Healthcheck, error)
	GetHealthcheckByName(ctx context.Context, name string) (*aggregates.Healthcheck, error)
	DeleteHealthcheck(ctx context.Context, id string) error
	ListHealthchecks(ctx context.Context, regex *regexp.Regexp) ([]*aggregates.Healthcheck, error)
	CountHealthchecks(ctx context.Context) (int, error)
}

type Builder struct {
	healthcheck HealthcheckService
}

func NewBuilder(healthcheck HealthcheckService) *Builder {
	return &Builder{
		healthcheck: healthcheck,
	}
}
