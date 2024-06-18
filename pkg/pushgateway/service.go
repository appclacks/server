package pushgateway

import (
	"context"

	"github.com/appclacks/server/pkg/pushgateway/aggregates"
)

type Store interface {
	CreateOrUpdatePushgatewayMetric(ctx context.Context, metric aggregates.PushgatewayMetric, cumulative bool) (string, error)
	GetMetrics(ctx context.Context) ([]*aggregates.PushgatewayMetric, error)
	DeleteMetricsByName(ctx context.Context, name string) error
	DeleteMetricByID(ctx context.Context, id string) error
}

type Service struct {
	//logger *slog.Logger
	store Store
}

func New(store Store) *Service {
	return &Service{
		store: store,
	}
}
