package handlers

import (
	"net/http"
	"time"

	"github.com/appclacks/go-client"
	"github.com/appclacks/server/pkg/pushgateway"
	"github.com/appclacks/server/pkg/pushgateway/aggregates"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	er "github.com/mcorbin/corbierror"
)

func (b *Builder) CreateOrUpdatePushgatewayMetric(ec echo.Context) error {
	var payload client.CreateOrUpdatePushgatewayMetricInput
	if err := ec.Bind(&payload); err != nil {
		return err
	}
	if err := ec.Validate(payload); err != nil {
		return err
	}

	metric := &aggregates.PushgatewayMetric{
		Name:   payload.Name,
		Labels: payload.Labels,
		Value:  payload.Value,
	}
	pushgateway.InitPushgatewayMetric(metric)
	if payload.Description != "" {
		metric.Description = &payload.Description
	}
	if payload.TTL != "" {
		metric.TTL = &payload.TTL
		ttl, err := time.ParseDuration(*metric.TTL)
		if err != nil {
			return er.Newf("Invalid TTL: %s", er.BadRequest, true, err.Error())
		}
		expiresAt := metric.CreatedAt.Add(ttl)
		metric.ExpiresAt = &expiresAt
	}
	if payload.Type != "" {
		metric.Type = &payload.Type
	}
	id, err := b.pushgateway.CreateOrUpdatePushgatewayMetric(ec.Request().Context(), *metric, false)
	if err != nil {
		return err
	}
	// todo return something else ?
	return ec.JSON(http.StatusOK, NewResponse("metric created", id))
}

func (b *Builder) PushgatewayMetrics(ec echo.Context) error {
	metricString, err := b.pushgateway.PrometheusMetrics(ec.Request().Context())
	if err != nil {
		return err
	}
	return ec.String(200, metricString)
}

func (b *Builder) DeleteMetric(ec echo.Context) error {
	var payload client.DeletePushgatewayMetricInput
	if err := ec.Bind(&payload); err != nil {
		return err
	}
	if err := ec.Validate(payload); err != nil {
		return err
	}
	isUUID := false
	_, err := uuid.Parse(payload.Identifier)
	if err == nil {
		isUUID = true
	}
	if isUUID {
		err := b.pushgateway.DeleteMetricByID(ec.Request().Context(), payload.Identifier)
		if err != nil {
			return err
		}
	} else {
		err := b.pushgateway.DeleteMetricsByName(ec.Request().Context(), payload.Identifier)
		if err != nil {
			return err
		}
	}
	return ec.JSON(http.StatusOK, NewResponse("metrics deleted"))
}

func (b *Builder) ListPushgatewayMetrics(ec echo.Context) error {
	metrics, err := b.pushgateway.GetMetrics(ec.Request().Context())
	if err != nil {
		return err
	}
	result := []client.PushgatewayMetric{}
	for _, metric := range metrics {
		m := client.PushgatewayMetric{
			ID:        metric.ID,
			Name:      metric.Name,
			Labels:    metric.Labels,
			CreatedAt: metric.CreatedAt,
			Value:     metric.Value,
		}
		if metric.Description != nil {
			m.Description = *metric.Description
		}
		if metric.TTL != nil {
			m.TTL = *metric.TTL
		}
		if metric.Type != nil {
			m.Type = *metric.Type
		}
		if metric.ExpiresAt != nil {
			m.ExpiresAt = metric.ExpiresAt
		}
		result = append(result, m)
	}
	return ec.JSON(http.StatusOK, client.ListPushgatewayMetricsOutput{
		Result: result,
	})
}
