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
	"github.com/pkg/errors"
)

type Database struct {
	DB      *sqlx.DB
	Logger  *slog.Logger
	probers uint
}

func (d *Database) Exec(query string) (sql.Result, error) {
	return d.DB.Exec(query)
}

func New(logger *slog.Logger, config Configuration, probers uint) (*Database, error) {
	err := validator.Validator.Struct(config)
	if err != nil {
		return nil, err
	}
	connectionString := fmt.Sprintf("host=%s port=%d user=%s dbname=%s password=%s sslmode=%s", config.Host, config.Port, config.Username, config.Database, config.Password, config.SSLMode)
	db, err := sqlx.Connect("postgres", connectionString)
	if err != nil {
		return nil, errors.Wrap(err, "Fail to connect to the database")
	}
	db.SetConnMaxLifetime(time.Duration(60) * time.Second)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)
	driver, err := postgres.WithInstance(db.DB, &postgres.Config{})
	if err != nil {
		return nil, errors.Wrap(err, "Fail to create postgres migration driver")
	}
	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", config.Migrations),
		"postgres",
		driver)
	if err != nil {
		return nil, errors.Wrap(err, "Fail to instantiate migrations")
	}
	logger.Info("Applying databases migrations")
	err = m.Up()
	if err != nil && err != migrate.ErrNoChange {
		return nil, errors.Wrap(err, "Fail to apply migrations")
	}
	logger.Info("Migrations applied")
	return &Database{
		DB:      db,
		Logger:  logger,
		probers: probers,
	}, nil
}

func checkResult(result sql.Result, expected int64) error {
	affected, err := result.RowsAffected()
	if err != nil {
		return errors.Wrap(err, "fail to check affected rows")
	}
	if affected != expected {
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
		return nil, errors.Wrapf(err, "fail to deserialize labels: %s", *labels)
	}
	return result, nil
}

func labelsToString(labels map[string]string) (*string, error) {
	if labels == nil {
		return nil, nil
	}
	b, err := json.Marshal(labels)
	if err != nil {
		return nil, errors.Wrap(err, "fail to genrate labels string from labels map")
	}
	result := string(b)
	return &result, nil
}
