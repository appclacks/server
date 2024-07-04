package handlers

import (
	"context"

	"github.com/appclacks/server/pkg/healthcheck/aggregates"
	pgaggregates "github.com/appclacks/server/pkg/pushgateway/aggregates"
)

type HealthcheckService interface {
	UpdateHealthcheck(ctx context.Context, healthcheck *aggregates.Healthcheck) error
	CreateHealthcheck(ctx context.Context, healthcheck *aggregates.Healthcheck) error
	GetHealthcheck(ctx context.Context, id string) (*aggregates.Healthcheck, error)
	GetHealthcheckByName(ctx context.Context, name string) (*aggregates.Healthcheck, error)
	DeleteHealthcheck(ctx context.Context, id string) error
	ListHealthchecks(ctx context.Context, query aggregates.Query) ([]*aggregates.Healthcheck, error)
	CountHealthchecks(ctx context.Context) (int, error)
}

type PushgatewayService interface {
	CreateOrUpdatePushgatewayMetric(ctx context.Context, metric pgaggregates.PushgatewayMetric, cumulative bool) (string, error)
	GetMetrics(ctx context.Context) ([]*pgaggregates.PushgatewayMetric, error)
	DeleteMetricsByName(ctx context.Context, name string) error
	DeleteMetricByID(ctx context.Context, id string) error
	PrometheusMetrics(ctx context.Context) (string, error)
	DeleteAllPushgatewayMetrics(ctx context.Context) error
}

type Builder struct {
	healthcheck HealthcheckService
	pushgateway PushgatewayService
}

func NewBuilder(healthcheck HealthcheckService, pushgateway PushgatewayService) *Builder {
	return &Builder{
		healthcheck: healthcheck,
		pushgateway: pushgateway,
	}
}
