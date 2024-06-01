package config

import (
	"github.com/appclacks/server/internal/database"
	"github.com/appclacks/server/internal/http"
)

type Healthchecks struct {
	Probers uint
}

type Configuration struct {
	HTTP         http.Configuration
	Database     database.Configuration
	Healthchecks Healthchecks
}
