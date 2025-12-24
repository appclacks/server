package handlers

import (
	"fmt"
	"net/http"

	"github.com/appclacks/go-client"
	"github.com/appclacks/server/pkg/heartbeat"
	"github.com/appclacks/server/pkg/heartbeat/aggregates"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

func (b *Builder) DeleteHeartbeat(ec echo.Context) error {
	var payload client.DeleteHeartbeatInput
	if err := ec.Bind(&payload); err != nil {
		return err
	}
	if err := ec.Validate(payload); err != nil {
		return err
	}

	err := b.heartbeat.DeleteHeartbeat(ec.Request().Context(), payload.ID)
	if err != nil {
		return err
	}
	return ec.JSON(http.StatusOK, NewResponse("Heartbeat deleted"))
}

func toHeartbeat(heartbeat aggregates.Heartbeat) client.Heartbeat {
	result := client.Heartbeat{
		ID:          heartbeat.ID,
		Name:        heartbeat.Name,
		Labels:      heartbeat.Labels,
		TTL:         heartbeat.TTL,
		CreatedAt:   heartbeat.CreatedAt,
		RefreshedAt: heartbeat.RefreshedAt,
	}
	if heartbeat.Description != nil {
		result.Description = *heartbeat.Description
	}
	return result
}

func toHeartbeats(heartbeats []*aggregates.Heartbeat) []client.Heartbeat {
	result := []client.Heartbeat{}
	for i := range heartbeats {
		hb := *heartbeats[i]
		result = append(result, toHeartbeat(hb))
	}
	return result
}

func (b *Builder) GetHeartbeat(ec echo.Context) error {
	var payload client.GetHeartbeatInput
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
	var heartbeat *aggregates.Heartbeat
	if isUUID {
		heartbeat, err = b.heartbeat.GetHeartbeat(ec.Request().Context(), payload.Identifier)
		if err != nil {
			return err
		}
	} else {
		heartbeat, err = b.heartbeat.GetHeartbeatByName(ec.Request().Context(), payload.Identifier)
		if err != nil {
			return err
		}
	}
	result := toHeartbeat(*heartbeat)

	return ec.JSON(http.StatusOK, &result)
}

func (b *Builder) ListHeartbeats(ec echo.Context) error {
	heartbeats, err := b.heartbeat.ListHeartbeats(ec.Request().Context())
	if err != nil {
		return err
	}
	result := client.ListHeartbeatsOutput{
		Result: toHeartbeats(heartbeats),
	}
	return ec.JSON(http.StatusOK, &result)
}

func (b *Builder) CreateHeartbeat(ec echo.Context) error {
	var payload client.CreateHeartbeatInput
	if err := ec.Bind(&payload); err != nil {
		return err
	}
	if err := ec.Validate(payload); err != nil {
		return err
	}

	labels := payload.Labels
	if labels == nil {
		labels = make(map[string]string)
	}

	hb := aggregates.Heartbeat{
		Name:   payload.Name,
		Labels: labels,
		TTL:    nil,
	}
	if payload.Description != "" {
		hb.Description = &payload.Description
	}
	if payload.TTL != "" {
		hb.TTL = &payload.TTL
	}

	heartbeat.InitHeartbeat(&hb)
	err := b.heartbeat.CreateHeartbeat(ec.Request().Context(), &hb)
	if err != nil {
		return err
	}
	return ec.JSON(http.StatusCreated, NewResponse(fmt.Sprintf("Heartbeat %s created", hb.Name)))
}

func (b *Builder) UpdateHeartbeat(ec echo.Context) error {
	var payload client.UpdateHeartbeatInput
	if err := ec.Bind(&payload); err != nil {
		return err
	}
	if err := ec.Validate(payload); err != nil {
		return err
	}

	labels := payload.Labels
	if labels == nil {
		labels = make(map[string]string)
	}

	hb := aggregates.Heartbeat{
		ID:     payload.ID,
		Name:   payload.Name,
		Labels: labels,
		TTL:    nil,
	}
	if payload.Description != "" {
		hb.Description = &payload.Description
	}
	if payload.TTL != "" {
		hb.TTL = &payload.TTL
	}

	err := b.heartbeat.UpdateHeartbeat(ec.Request().Context(), &hb)
	if err != nil {
		return err
	}
	return ec.JSON(http.StatusOK, NewResponse(fmt.Sprintf("Heartbeat %s updated", hb.Name)))
}

func (b *Builder) RefreshHeartbeat(ec echo.Context) error {
	var payload client.RefreshHeartbeatInput
	if err := ec.Bind(&payload); err != nil {
		return err
	}
	if err := ec.Validate(payload); err != nil {
		return err
	}

	err := b.heartbeat.RefreshHeartbeat(ec.Request().Context(), payload.ID)
	if err != nil {
		return err
	}
	return ec.JSON(http.StatusOK, NewResponse("Heartbeat refreshed"))
}