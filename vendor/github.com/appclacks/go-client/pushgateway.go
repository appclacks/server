package client

import "time"

type PushgatewayMetric struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	TTL         string            `json:"ttl"`
	Type        string            `json:"type"`
	CreatedAt   time.Time         `json:"created_at"`
	ExpiresAt   time.Time         `json:"expires_at"`
	Value       float32           `json:"value"`
}

type CreateOrUpdatePushgatewayMetricInput struct {
	Name        string            `json:"name" validate:"required,max=255,min=1"`
	Description string            `json:"description,omitempty"`
	Labels      map[string]string `json:"labels" description:"Healthcheck labels" validate:"dive,keys,max=255,min=1,endkeys,max=255,min=1"`
	TTL         string            `json:"ttl"`
	Type        string            `json:"type" validate:"oneof=counter gauge histogram summary"`
	Value       float32           `json:"value" validate:"required"`
}

type DeletePushgatewayMetricInput struct {
	Identifier string `json:"identifier" validate:"required"`
}

type ListPushgatewayMetricsOutput struct {
	Result []PushgatewayMetric `json:"result"`
}
