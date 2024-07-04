package pushgateway

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/appclacks/server/pkg/pushgateway/aggregates"
	"github.com/prometheus/client_golang/prometheus"
)

func InitPushgatewayMetric(metric *aggregates.PushgatewayMetric) {
	metric.CreatedAt = time.Now().UTC()
}

func (s *Service) CreateOrUpdatePushgatewayMetric(ctx context.Context, metric aggregates.PushgatewayMetric, cumulative bool) (string, error) {
	s.logger.Debug(fmt.Sprintf("creating or updating metric %s", metric.Name))
	return s.store.CreateOrUpdatePushgatewayMetric(ctx, metric, cumulative)
}

func (s *Service) GetMetrics(ctx context.Context) ([]*aggregates.PushgatewayMetric, error) {
	return s.store.GetMetrics(ctx)
}

func (s *Service) PrometheusMetrics(ctx context.Context) (string, error) {
	result := ""
	metrics, err := s.store.GetMetrics(ctx)
	if err != nil {
		return "", err
	}
	for _, metric := range metrics {
		name := metric.Name
		if metric.Description != nil {
			result += fmt.Sprintf("\nHELP %s %s\n", name, *metric.Description)
		}
		if metric.Type != nil {
			result += fmt.Sprintf("TYPE %s %s\n", name, *metric.Type)
		}
		labels := ""
		if len(metric.Labels) == 0 {
			labels = "{}"
		} else {
			labelsList := []string{}
			for k, v := range metric.Labels {
				labelsList = append(labelsList, fmt.Sprintf("%s=\"%s\"", k, v))
			}
			labels = fmt.Sprintf("{%s}", strings.Join(labelsList, ","))

		}
		result += fmt.Sprintf("%s%s %f\n", metric.Name, labels, metric.Value)
	}
	return result, nil
}

func (s *Service) DeleteMetricsByName(ctx context.Context, name string) error {
	s.logger.Debug(fmt.Sprintf("deleting push gateway metric %s", name))
	return s.store.DeleteMetricsByName(ctx, name)
}

func (s *Service) DeleteMetricByID(ctx context.Context, id string) error {
	s.logger.Debug(fmt.Sprintf("deleting push gateway metric %s", id))
	return s.store.DeleteMetricByID(ctx, id)
}

func (s *Service) DeleteAllPushgatewayMetrics(ctx context.Context) error {
	s.logger.Debug("deleting all push gateway metrics")
	return s.store.DeleteAllPushgatewayMetrics(ctx)
}

func (s *Service) CleanPushgatewayMetrics(ctx context.Context) {
	numberMetricsDeleted, err := s.store.CleanPushgatewayMetrics(ctx)
	if err != nil {
		s.logger.Error(fmt.Sprintf("fail to clean push gateway metrics: %s", err.Error()))
		s.pushgatewayExecutionsCounter.With(prometheus.Labels{"status": "failure"}).Inc()
	} else {
		s.logger.Debug(fmt.Sprintf("%d push gateway metrics cleaned", numberMetricsDeleted))
		s.pushgatewayExecutionsCounter.With(prometheus.Labels{"status": "success"}).Inc()
	}
}
