package handlers

import (
	"net/http"

	"github.com/appclacks/go-client"
	"github.com/appclacks/server/pkg/healthcheck"
	"github.com/appclacks/server/pkg/healthcheck/aggregates"
	"github.com/labstack/echo/v4"
)

func (b *Builder) CreateCommandHealthcheck(ec echo.Context) error {
	var payload client.CreateCommandHealthcheckInput
	if err := ec.Bind(&payload); err != nil {
		return err
	}

	check := &aggregates.Healthcheck{
		Name:     payload.Name,
		Labels:   payload.Labels,
		Interval: payload.Interval,
		Timeout:  payload.Timeout,
		Enabled:  payload.Enabled,
		Definition: &aggregates.HealthcheckCommandDefinition{
			Command:   payload.Command,
			Arguments: payload.Arguments,
		},
	}
	if payload.Description != "" {
		check.Description = &payload.Description
	}
	healthcheck.InitCommandHealthcheck(check)
	err := b.healthcheck.CreateHealthcheck(ec.Request().Context(), check)
	if err != nil {
		return err
	}
	healthcheckResult := client.Healthcheck{
		ID:         check.ID,
		Name:       check.Name,
		Type:       check.Type,
		Interval:   check.Interval,
		Timeout:    check.Timeout,
		Labels:     check.Labels,
		CreatedAt:  check.CreatedAt,
		Enabled:    check.Enabled,
		Definition: payload.HealthcheckCommandDefinition,
	}
	if check.Description != nil {
		healthcheckResult.Description = *check.Description
	}
	return ec.JSON(http.StatusOK, &healthcheckResult)
}

func (b *Builder) UpdateCommandHealthcheck(ec echo.Context) error {
	var payload client.UpdateCommandHealthcheckInput
	if err := ec.Bind(&payload); err != nil {
		return err
	}

	check := &aggregates.Healthcheck{
		ID:       payload.ID,
		Name:     payload.Name,
		Labels:   payload.Labels,
		Interval: payload.Interval,
		Enabled:  payload.Enabled,
		Timeout:  payload.Timeout,
		Definition: &aggregates.HealthcheckCommandDefinition{
			Command:   payload.Command,
			Arguments: payload.Arguments,
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
