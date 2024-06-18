package database

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/appclacks/server/internal/validator"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	er "github.com/mcorbin/corbierror"
)

type Database struct {
	db      *sqlx.DB
	Logger  *slog.Logger
	probers uint
}

var CleanupQueries = []string{
	"TRUNCATE healthcheck CASCADE",
	"TRUNCATE pushgateway_metric CASCADE",
	"TRUNCATE schema_migrations CASCADE",
}

func (d *Database) Exec(query string) (sql.Result, error) {
	return d.db.Exec(query)
}

func New(logger *slog.Logger, config Configuration, probers uint) (*Database, error) {
	err := validator.Validator.Struct(config)
	if err != nil {
		return nil, err
	}
	connectionString := fmt.Sprintf("host=%s port=%d user=%s dbname=%s password=%s sslmode=%s", config.Host, config.Port, config.Username, config.Database, config.Password, config.SSLMode)
	db, err := sqlx.Connect("postgres", connectionString)
	if err != nil {
		return nil, fmt.Errorf("fail to connect to the database: %w", err)
	}
	db.SetConnMaxLifetime(time.Duration(60) * time.Second)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)
	driver, err := postgres.WithInstance(db.DB, &postgres.Config{})
	if err != nil {
		return nil, fmt.Errorf("fail to create postgres migration driver: %w", err)
	}
	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", config.Migrations),
		"postgres",
		driver)
	if err != nil {
		return nil, fmt.Errorf("fail to instantiate migrations: %w", err)
	}
	logger.Info("Applying databases migrations")
	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return nil, fmt.Errorf("fail to apply migrations: %w", err)
	}
	logger.Info("Migrations applied")
	return &Database{
		db:      db,
		Logger:  logger,
		probers: probers,
	}, nil
}

func checkResult(result sql.Result, expected int64) error {
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("fail to check affected row: %w", err)
	}
	if affected != expected {
		if affected == 0 {
			return er.New("resource not found", er.NotFound, true)
		}
		return fmt.Errorf("expected %d rows changed, got %d", expected, affected)
	}
	return nil
}

func stringToLabels(labels *string) (map[string]string, error) {
	if labels == nil {
		return nil, nil
	}
	var result map[string]string
	if err := json.Unmarshal([]byte(*labels), &result); err != nil {
		return nil, fmt.Errorf("fail to deserialize labels %s: %w", *labels, err)
	}
	return result, nil
}

func labelsToString(labels map[string]string) (*string, error) {
	if labels == nil {
		return nil, nil
	}
	b, err := json.Marshal(labels)
	if err != nil {
		return nil, fmt.Errorf("fail to genrate labels string from labels map: %w", err)
	}
	result := string(b)
	return &result, nil
}
