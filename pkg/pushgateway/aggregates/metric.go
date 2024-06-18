package aggregates

import "time"

type PushgatewayMetric struct {
	ID          string
	Name        string
	Description *string
	Labels      map[string]string
	TTL         *string
	Type        *string
	CreatedAt   time.Time
	ExpiresAt   *time.Time
	Value       float32
}
