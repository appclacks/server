package aggregates

import "time"

type Heartbeat struct {
	ID          string
	Name        string
	Description *string
	Labels      map[string]string
	TTL         *string
	CreatedAt   time.Time  `db:"created_at"`
	RefreshedAt *time.Time `db:"refreshed_at"`
}
