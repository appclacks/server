package handlers

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	er "github.com/mcorbin/corbierror"

	client "github.com/appclacks/go-client"
	"github.com/appclacks/server/pkg/healthcheck"
	"github.com/appclacks/server/pkg/healthcheck/aggregates"
)

func (b *Builder) CreateTLSHealthcheck(ec echo.Context) error {
	var payload client.CreateTLSHealthcheckInput
	if err := ec.Bind(&payload); err != nil {
		return err
	}
	if err := ec.Validate(payload); err != nil {
		return err
	}

	if payload.ExpirationDelay != "" {
		_, err := time.ParseDuration(payload.ExpirationDelay)
		if err != nil {
			return er.New("Invalid healthcheck expiration delay", er.BadRequest, true)
		}
	}

	if payload.ExpirationDelay != "" {
		_, err := time.ParseDuration(payload.ExpirationDelay)
		if err != nil {
			return er.New("Invalid healthcheck expiration delay", er.BadRequest, true)
		}
	}

	check := &aggregates.Healthcheck{
		Name:     payload.Name,
		Labels:   payload.Labels,
		Interval: payload.Interval,
		Timeout:  payload.Timeout,
		Enabled:  payload.Enabled,
		Definition: &aggregates.HealthcheckTLSDefinition{
			Target: payload.Target,
			Port:   payload.Port,

			ServerName:      payload.ServerName,
			Insecure:        payload.Insecure,
			ExpirationDelay: payload.ExpirationDelay,
			Key:             payload.Key,
			Cacert:          payload.Cacert,
			Cert:            payload.Cert,
		},
	}
	if payload.Description != "" {
		check.Description = &payload.Description
	}
	healthcheck.InitTLSHealthcheck(check)
	err := b.healthcheck.CreateHealthcheck(ec.Request().Context(), check)
	if err != nil {
		return err
	}
	healthcheckResult := client.Healthcheck{
		ID:         check.ID,
		Name:       check.Name,
		Type:       check.Type,
		Labels:     check.Labels,
		Timeout:    check.Timeout,
		CreatedAt:  check.CreatedAt,
		Interval:   check.Interval,
		Enabled:    check.Enabled,
		Definition: payload.HealthcheckTLSDefinition,
	}
	if check.Description != nil {
		healthcheckResult.Description = *check.Description
	}
	return ec.JSON(http.StatusOK, &healthcheckResult)
}

func (b *Builder) UpdateTLSHealthcheck(ec echo.Context) error {
	var payload client.UpdateTLSHealthcheckInput
	if err := ec.Bind(&payload); err != nil {
		return err
	}
	if err := ec.Validate(payload); err != nil {
		return err
	}

	if payload.ExpirationDelay != "" {
		_, err := time.ParseDuration(payload.ExpirationDelay)
		if err != nil {
			return er.New("Invalid healthcheck expiration delay", er.BadRequest, true)
		}
	}

	check := &aggregates.Healthcheck{
		ID:       payload.ID,
		Name:     payload.Name,
		Labels:   payload.Labels,
		Timeout:  payload.Timeout,
		Interval: payload.Interval,
		Enabled:  payload.Enabled,
		Definition: &aggregates.HealthcheckTLSDefinition{
			Target: payload.Target,
			Port:   payload.Port,

			ServerName:      payload.ServerName,
			Insecure:        payload.Insecure,
			ExpirationDelay: payload.ExpirationDelay,
			Key:             payload.Key,
			Cacert:          payload.Cacert,
			Cert:            payload.Cert,
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
