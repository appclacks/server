package pushgateway

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/appclacks/server/pkg/pushgateway/aggregates"
	"github.com/prometheus/client_golang/prometheus"
)

type Store interface {
	CreateOrUpdatePushgatewayMetric(ctx context.Context, metric aggregates.PushgatewayMetric, cumulative bool) (string, error)
	GetMetrics(ctx context.Context) ([]*aggregates.PushgatewayMetric, error)
	DeleteMetricsByName(ctx context.Context, name string) error
	DeleteMetricByID(ctx context.Context, id string) error
	CleanPushgatewayMetrics(ctx context.Context) (int64, error)
	DeleteAllPushgatewayMetrics(ctx context.Context) error
}

type Service struct {
	logger                       *slog.Logger
	store                        Store
	pushgatewayExecutionsCounter *prometheus.CounterVec
	wg                           sync.WaitGroup
	stop                         chan bool
	ticker                       *time.Ticker
}

func New(logger *slog.Logger, store Store, registry *prometheus.Registry) (*Service, error) {
	pushgatewayExecutionsCounter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "pushgateway_cleanup_executions_total",
			Help: "Count the number of executions of the job cleaning pushgateway metrics",
		},
		[]string{"status"})
	err := registry.Register(pushgatewayExecutionsCounter)
	if err != nil {
		return nil, err
	}
	return &Service{
		stop:                         make(chan bool),
		store:                        store,
		logger:                       logger,
		pushgatewayExecutionsCounter: pushgatewayExecutionsCounter,
		ticker:                       time.NewTicker(60 * time.Second),
	}, nil
}

func (s *Service) Start() {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		for {
			select {
			case <-s.stop:
				return
			case <-s.ticker.C:
				s.logger.Debug("cleaning push gateway expired metrics")
				ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
				s.CleanPushgatewayMetrics(ctx)
				cancel()
			}
		}
	}()
}

func (s *Service) Stop() {
	s.ticker.Stop()
	s.stop <- true
	s.wg.Wait()
}
