package handlers

import (
	"net/http"

	"github.com/appclacks/go-client"
	"github.com/appclacks/server/internal/validator"
	"github.com/appclacks/server/pkg/healthcheck"
	"github.com/appclacks/server/pkg/healthcheck/aggregates"
	"github.com/labstack/echo/v4"
)

func (b *Builder) CreateTCPHealthcheck(ec echo.Context) error {
	var payload client.CreateTCPHealthcheckInput
	if err := ec.Bind(&payload); err != nil {
		return err
	}
	if err := validator.Validator.Struct(payload); err != nil {
		return err
	}

	check := &aggregates.Healthcheck{
		Name:     payload.Name,
		Labels:   payload.Labels,
		Interval: payload.Interval,
		Timeout:  payload.Timeout,
		Enabled:  payload.Enabled,
		Definition: &aggregates.HealthcheckTCPDefinition{
			Target:     payload.Target,
			Port:       payload.Port,
			ShouldFail: payload.ShouldFail,
		},
	}
	if payload.Description != "" {
		check.Description = &payload.Description
	}
	healthcheck.InitTCPHealthcheck(check)
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
		Definition: payload.HealthcheckTCPDefinition,
	}
	if check.Description != nil {
		healthcheckResult.Description = *check.Description
	}
	return ec.JSON(http.StatusOK, &healthcheckResult)
}

func (b *Builder) UpdateTCPHealthcheck(ec echo.Context) error {
	var payload client.UpdateTCPHealthcheckInput
	if err := ec.Bind(&payload); err != nil {
		return err
	}
	if err := validator.Validator.Struct(payload); err != nil {
		return err
	}

	check := &aggregates.Healthcheck{
		ID:     payload.ID,
		Name:   payload.Name,
		Labels: payload.Labels,

		Interval: payload.Interval,
		Enabled:  payload.Enabled,
		Timeout:  payload.Timeout,
		Definition: &aggregates.HealthcheckTCPDefinition{
			Target:     payload.Target,
			Port:       payload.Port,
			ShouldFail: payload.ShouldFail,
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
