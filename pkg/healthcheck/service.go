package healthcheck

import (
	"context"
	"log/slog"

	"github.com/appclacks/server/pkg/healthcheck/aggregates"
)

type Store interface {
	UpdateHealthcheck(ctx context.Context, healthcheck *aggregates.Healthcheck) error
	CreateHealthcheck(ctx context.Context, healthcheck *aggregates.Healthcheck) error
	GetHealthcheck(ctx context.Context, id string) (*aggregates.Healthcheck, error)
	GetHealthcheckByName(ctx context.Context, name string) (*aggregates.Healthcheck, error)
	DeleteHealthcheck(ctx context.Context, id string) error
	ListHealthchecks(ctx context.Context) ([]*aggregates.Healthcheck, error)
	ListHealthchecksForProber(ctx context.Context, prober int) ([]*aggregates.Healthcheck, error)
	CountHealthchecks(ctx context.Context) (int, error)
}

type Service struct {
	logger *slog.Logger
	store  Store
}

func New(logger *slog.Logger, store Store) *Service {
	return &Service{
		logger: logger,
		store:  store,
	}
}
