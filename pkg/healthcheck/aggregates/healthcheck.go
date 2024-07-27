package aggregates

import (
	"regexp"
	"time"
)

type Healthcheck struct {
	ID          string
	RandomID    int
	Name        string
	Description *string
	Labels      map[string]string
	Type        string
	Interval    string
	Timeout     string
	Enabled     bool
	CreatedAt   time.Time
	Definition  HealthcheckDefinition
}

type Query struct {
	Enabled *bool
	Regex   *regexp.Regexp
	Prober  uint
}
