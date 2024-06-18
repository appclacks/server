package handlers

import (
	"net/http"

	"github.com/appclacks/go-client"
	"github.com/appclacks/server/pkg/healthcheck"
	"github.com/appclacks/server/pkg/healthcheck/aggregates"
	"github.com/labstack/echo/v4"
)

func (b *Builder) CreateDNSHealthcheck(ec echo.Context) error {
	var payload client.CreateDNSHealthcheckInput
	if err := ec.Bind(&payload); err != nil {
		return err
	}
	if err := ec.Validate(payload); err != nil {
		return err
	}
	check := &aggregates.Healthcheck{
		Name:     payload.Name,
		Labels:   payload.Labels,
		Interval: payload.Interval,
		Timeout:  payload.Timeout,
		Enabled:  payload.Enabled,
		Definition: &aggregates.HealthcheckDNSDefinition{
			Domain:      payload.Domain,
			ExpectedIPs: payload.ExpectedIPs,
		},
	}
	if payload.Description != "" {
		check.Description = &payload.Description
	}
	healthcheck.InitDNSHealthcheck(check)
	err := b.healthcheck.CreateHealthcheck(ec.Request().Context(), check)
	if err != nil {
		return err
	}
	healthcheckResult := client.Healthcheck{
		ID:         check.ID,
		Name:       check.Name,
		Type:       check.Type,
		Interval:   check.Interval,
		Labels:     check.Labels,
		Timeout:    check.Timeout,
		Enabled:    check.Enabled,
		CreatedAt:  check.CreatedAt,
		Definition: payload.HealthcheckDNSDefinition,
	}
	if check.Description != nil {
		healthcheckResult.Description = *check.Description
	}
	return ec.JSON(http.StatusOK, &healthcheckResult)
}

func (b *Builder) UpdateDNSHealthcheck(ec echo.Context) error {
	var payload client.UpdateDNSHealthcheckInput
	if err := ec.Bind(&payload); err != nil {
		return err
	}
	if err := ec.Validate(payload); err != nil {
		return err
	}

	check := &aggregates.Healthcheck{
		ID:       payload.ID,
		Name:     payload.Name,
		Labels:   payload.Labels,
		Interval: payload.Interval,
		Timeout:  payload.Timeout,
		Enabled:  payload.Enabled,
		Definition: &aggregates.HealthcheckDNSDefinition{
			Domain:      payload.Domain,
			ExpectedIPs: payload.ExpectedIPs,
		},
	}
	if payload.Description != "" {
		check.Description = &payload.Description
	}
	err := b.healthcheck.UpdateHealthcheck(ec.Request().Context(), check)
	if err != nil {
		return err
	}
	healthcheckResult, err := b.healthcheck.GetHealthcheck(ec.Request().Context(), payload.ID)
	if err != nil {
		return err
	}
	result := toHealthcheck(*healthcheckResult)
	return ec.JSON(http.StatusOK, &result)
}
