package aggregates

import "time"

type SLO struct {
	ID          string
	Name        string
	Description *string
	Labels      map[string]string
	CreatedAt   time.Time
	Objective   float32
}

type Record struct {
	Name    string
	Success bool
	Value   int64
}

type SLOSum struct {
	StartDate time.Time
	Name      string
	Success   int64
	Failure   int64
}
