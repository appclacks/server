package pushgateway

import (
	"context"
	"fmt"
	"sort"
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
	sort.Slice(metrics, func(i, j int) bool {
		return strings.Compare(metrics[i].Name, metrics[j].Name) == -1
	})
	metricsDescriptions := make(map[string]string)
	// double iteration to build metrics descriptions once
	for _, metric := range metrics {
		name := metric.Name
		description := ""
		if metric.Description != nil {
			description += fmt.Sprintf("# HELP %s %s\n", name, *metric.Description)
		}
		if metric.Type != nil {
			description += fmt.Sprintf("# TYPE %s %s\n", name, *metric.Type)
		}
		if description != "" {
			metricsDescriptions[name] = description
		}
	}

	for _, metric := range metrics {
		name := metric.Name
		description, ok := metricsDescriptions[name]
		if ok {
			result += description
			// don't fetch this description again
			// the metrics being sorted, we only want to print
			// the description once at the first occurence
			delete(metricsDescriptions, name)
		}
		labels := ""
		if len(metric.Labels) == 0 {
			labels = "{}"
		} else {
			labelsList := []string{}
			for k, v := range metric.Labels {
				labelsList = append(labelsList, fmt.Sprintf("%s=\"%s\"", k, v))
			}
			sort.Strings(labelsList)
			labels = fmt.Sprintf("{%s}", strings.Join(labelsList, ", "))

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
