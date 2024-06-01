package handlers

import (
	"net/http"
	"regexp"

	"github.com/appclacks/go-client"
	"github.com/appclacks/server/internal/validator"
	"github.com/appclacks/server/pkg/healthcheck"
	"github.com/appclacks/server/pkg/healthcheck/aggregates"
	"github.com/labstack/echo/v4"
	er "github.com/mcorbin/corbierror"
)

func (b *Builder) CreateHTTPHealthcheck(ec echo.Context) error {
	var payload client.CreateHTTPHealthcheckInput
	if err := ec.Bind(&payload); err != nil {
		return err
	}
	if err := validator.Validator.Struct(payload); err != nil {
		return err
	}

	for _, r := range payload.BodyRegexp {
		_, err := regexp.Compile(r)
		if err != nil {
			return er.Newf("Invalid Regex %s in body-regexp", er.BadRequest, true, r)
		}
	}

	check := &aggregates.Healthcheck{
		Name:     payload.Name,
		Labels:   payload.Labels,
		Timeout:  payload.Timeout,
		Interval: payload.Interval,
		Enabled:  payload.Enabled,
		Definition: &aggregates.HealthcheckHTTPDefinition{
			ValidStatus: payload.ValidStatus,
			Target:      payload.Target,
			Method:      payload.Method,
			Port:        payload.Port,
			Query:       payload.Query,
			Redirect:    payload.Redirect,
			Body:        payload.Body,
			Headers:     payload.Headers,
			Protocol:    payload.Protocol,
			Path:        payload.Path,
			BodyRegexp:  payload.BodyRegexp,
			Key:         payload.Key,
			Cacert:      payload.Cacert,
			Cert:        payload.Cert,
			Host:        payload.Host,
			Insecure:    payload.Insecure,
			ServerName:  payload.ServerName,
		},
	}

	if payload.Description != "" {
		check.Description = &payload.Description
	}
	healthcheck.InitHTTPHealthcheck(check)
	err := b.healthcheck.CreateHealthcheck(ec.Request().Context(), check)
	if err != nil {
		return err
	}
	healthcheckResult := client.Healthcheck{
		ID:         check.ID,
		Name:       check.Name,
		Type:       check.Type,
		Labels:     check.Labels,
		CreatedAt:  check.CreatedAt,
		Interval:   check.Interval,
		Enabled:    check.Enabled,
		Timeout:    check.Timeout,
		Definition: payload.HealthcheckHTTPDefinition,
	}
	if check.Description != nil {
		healthcheckResult.Description = *check.Description
	}
	return ec.JSON(http.StatusOK, &healthcheckResult)
}

func (b *Builder) UpdateHTTPHealthcheck(ec echo.Context) error {
	var payload client.UpdateHTTPHealthcheckInput
	if err := ec.Bind(&payload); err != nil {
		return err
	}
	if err := validator.Validator.Struct(payload); err != nil {
		return err
	}
	for _, r := range payload.BodyRegexp {
		_, err := regexp.Compile(r)
		if err != nil {
			return er.Newf("Invalid Regex %s in body-regexp", er.BadRequest, true, r)
		}
	}

	check := &aggregates.Healthcheck{
		ID:       payload.ID,
		Name:     payload.Name,
		Labels:   payload.Labels,
		Timeout:  payload.Timeout,
		Interval: payload.Interval,
		Enabled:  payload.Enabled,
		Definition: &aggregates.HealthcheckHTTPDefinition{
			ValidStatus: payload.ValidStatus,
			Target:      payload.Target,
			Method:      payload.Method,
			Port:        payload.Port,
			Redirect:    payload.Redirect,
			Body:        payload.Body,
			Headers:     payload.Headers,
			Protocol:    payload.Protocol,
			Path:        payload.Path,
			Query:       payload.Query,
			BodyRegexp:  payload.BodyRegexp,
			Key:         payload.Key,
			Cacert:      payload.Cacert,
			Cert:        payload.Cert,
			Host:        payload.Host,
			Insecure:    payload.Insecure,
			ServerName:  payload.ServerName,
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
