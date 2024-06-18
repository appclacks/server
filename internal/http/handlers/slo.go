package handlers

import (
	"net/http"

	"github.com/appclacks/server/pkg/slo/aggregates"
	"github.com/labstack/echo/v4"
)

type Record struct {
	Name    string
	Success bool
	Value   int64
}

func (b *Builder) AddSLORecord(ec echo.Context) error {
	var payload Record
	if err := ec.Bind(&payload); err != nil {
		return err
	}
	if err := ec.Validate(payload); err != nil {
		return err
	}

	record := aggregates.Record{
		Name:    payload.Name,
		Success: payload.Success,
		Value:   payload.Value,
	}

	err := b.slo.AddRecord(ec.Request().Context(), record)
	if err != nil {
		return err
	}
	return ec.JSON(http.StatusOK, NewResponse("record created"))
}

func (b *Builder) SLOMetrics(ec echo.Context) error {

	err := b.slo.Metrics(ec.Request().Context())
	if err != nil {
		return err
	}
	return ec.JSON(http.StatusOK, NewResponse("metrics"))
}
