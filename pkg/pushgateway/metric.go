package pushgateway

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/appclacks/server/pkg/pushgateway/aggregates"
)

func InitPushgatewayMetric(metric *aggregates.PushgatewayMetric) {
	metric.CreatedAt = time.Now().UTC()
}

func (s *Service) CreateOrUpdatePushgatewayMetric(ctx context.Context, metric aggregates.PushgatewayMetric, cumulative bool) (string, error) {
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
	return s.store.DeleteMetricsByName(ctx, name)
}

func (s *Service) DeleteMetricByID(ctx context.Context, id string) error {
	return s.store.DeleteMetricByID(ctx, id)
}
