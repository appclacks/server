package handlers

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/appclacks/go-client"
	"github.com/appclacks/server/internal/validator"
	"github.com/appclacks/server/pkg/healthcheck"
	"github.com/appclacks/server/pkg/healthcheck/aggregates"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	er "github.com/mcorbin/corbierror"
)

func (b *Builder) DeleteHealthcheck(ec echo.Context) error {
	var payload client.DeleteHealthcheckInput
	if err := ec.Bind(&payload); err != nil {
		return err
	}
	if err := validator.Validator.Struct(payload); err != nil {
		return err
	}

	err := b.healthcheck.DeleteHealthcheck(ec.Request().Context(), payload.ID)
	if err != nil {
		return err
	}
	return ec.JSON(http.StatusOK, NewResponse("Healthcheck deleted"))
}

func toHealthcheck(healthcheck aggregates.Healthcheck) client.Healthcheck {
	result := client.Healthcheck{
		ID:         healthcheck.ID,
		Name:       healthcheck.Name,
		Type:       healthcheck.Type,
		Labels:     healthcheck.Labels,
		Timeout:    healthcheck.Timeout,
		Interval:   healthcheck.Interval,
		CreatedAt:  healthcheck.CreatedAt,
		Enabled:    healthcheck.Enabled,
		Definition: healthcheck.Definition,
	}
	if healthcheck.Description != nil {
		result.Description = *healthcheck.Description
	}
	return result
}

func toHealthchecks(healthchecks []*aggregates.Healthcheck) []client.Healthcheck {
	result := []client.Healthcheck{}
	for i := range healthchecks {
		check := *healthchecks[i]
		result = append(result, toHealthcheck(check))
	}
	return result
}

func (b *Builder) GetHealthcheck(ec echo.Context) error {
	var payload client.GetHealthcheckInput
	if err := ec.Bind(&payload); err != nil {
		return err
	}
	if err := validator.Validator.Struct(payload); err != nil {
		return err
	}

	isUUID := false
	_, err := uuid.Parse(payload.Identifier)
	if err == nil {
		isUUID = true
	}
	var healthcheck *aggregates.Healthcheck
	if isUUID {
		healthcheck, err = b.healthcheck.GetHealthcheck(ec.Request().Context(), payload.Identifier)
		if err != nil {
			return err
		}
	} else {
		healthcheck, err = b.healthcheck.GetHealthcheckByName(ec.Request().Context(), payload.Identifier)
		if err != nil {
			return err
		}
	}
	result := toHealthcheck(*healthcheck)

	return ec.JSON(http.StatusOK, &result)
}

func (b *Builder) ListHealthchecks(ec echo.Context) error {
	var payload client.ListHealthchecksInput
	if err := ec.Bind(&payload); err != nil {
		return err
	}
	if err := validator.Validator.Struct(payload); err != nil {
		return err
	}

	var nameRegex *regexp.Regexp
	var err error
	if payload.NamePattern != "" {
		nameRegex, err = regexp.Compile(payload.NamePattern)
		if err != nil {
			return er.New("Invalid regex for the name-pattern parameter", er.BadRequest, true)
		}
	}
	query := aggregates.Query{
		Regex: nameRegex,
	}
	healthchecks, err := b.healthcheck.ListHealthchecks(ec.Request().Context(), query)
	if err != nil {
		return err
	}
	result := client.ListHealthchecksOutput{
		Result: toHealthchecks(healthchecks),
	}

	return ec.JSON(http.StatusOK, result)
}

func (b *Builder) CabourotteDiscovery(ec echo.Context) error {
	var payload client.CabourotteDiscoveryInput
	if err := ec.Bind(&payload); err != nil {
		return err
	}
	if err := validator.Validator.Struct(payload); err != nil {
		return err
	}

	// foo=bar,a=b
	// TODO move this in service
	unformattedTags := payload.Labels
	labels := make(map[string]string)
	if unformattedTags != "" {
		splittedTags := strings.Split(unformattedTags, ",")
		for _, tag := range splittedTags {
			kvSplitted := strings.Split(tag, "=")
			if len(kvSplitted) != 2 {
				return er.New("Invalid labels parameter", er.BadRequest, true)
			}
			labels[kvSplitted[0]] = kvSplitted[1]
		}
	}
	t := true
	query := aggregates.Query{Enabled: &t}
	healthchecks, err := b.healthcheck.ListHealthchecks(ec.Request().Context(), query)
	if err != nil {
		return err
	}
	result := client.CabourotteDiscoveryOutput{}
	for i := range healthchecks {
		hc := healthchecks[i]
		if healthcheck.MatchLabels(hc, labels) {
			switch hc.Type {
			case "http":
				result.HTTPChecks = append(result.HTTPChecks, toHealthcheck(*hc))
			case "tcp":
				result.TCPChecks = append(result.TCPChecks, toHealthcheck(*hc))
			case "dns":
				result.DNSChecks = append(result.DNSChecks, toHealthcheck(*hc))
			case "tls":
				result.TLSChecks = append(result.TLSChecks, toHealthcheck(*hc))
			case "command":
				result.CommandChecks = append(result.CommandChecks, toHealthcheck(*hc))
			default:
				return fmt.Errorf("healthcheck type %s unknown for healthcheck %s", hc.Type, hc.ID)
			}

		}
	}

	return ec.JSON(http.StatusOK, result)
}
