package database_test

import (
	"log/slog"
	"os"
	"testing"

	"github.com/appclacks/server/internal/database"
	"github.com/pkg/errors"
)

var TestComponent *database.Database

func InitTestDB(logger *slog.Logger) *database.Database {

	config := database.Configuration{
		Migrations: "../../dev/migrations",
		Username:   "appclacks",
		Password:   "appclacks",
		Database:   "appclacks",
		Host:       "127.0.0.1",
		Port:       5432,
		SSLMode:    "disable",
	}
	c, err := database.New(logger, config, 1)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	err = cleanup(c)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	logger.Info("db cleanup done")
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	return c

}

func cleanup(c *database.Database) error {
	queries := []string{
		"TRUNCATE healthcheck CASCADE",
		"TRUNCATE schema_migrations CASCADE",
	}
	for _, query := range queries {
		_, err := c.DB.Exec(query)
		if err != nil {
			return errors.Wrapf(err, "fail to clean DB on query %s", query)
		}
	}
	return nil
}

func TestMain(m *testing.M) {
	logger := slog.Default()
	TestComponent = InitTestDB(logger)
	exitVal := m.Run()
	os.Exit(exitVal)

}
