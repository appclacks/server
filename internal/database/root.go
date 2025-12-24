package database

import (
	"database/sql"
	"embed"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	otelsql "github.com/XSAM/otelsql"
	"github.com/appclacks/server/internal/validator"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	er "github.com/mcorbin/corbierror"
	"go.opentelemetry.io/otel"
	semconv "go.opentelemetry.io/otel/semconv/v1.37.0"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

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

	dbConn, err := otelsql.Open("postgres", connectionString,
		otelsql.WithTracerProvider(otel.GetTracerProvider()),
		otelsql.WithAttributes(
			semconv.DBSystemNamePostgreSQL,
		),
		otelsql.WithSpanOptions(otelsql.SpanOptions{
			Ping:                 true,
			RowsNext:             true,
			DisableErrSkip:       true,
			OmitConnResetSession: true,
			OmitConnectorConnect: false,
			OmitConnPrepare:      false,
			OmitRows:             false,
		}),
	)
	if err != nil {
		return nil, fmt.Errorf("fail to open database connection: %w", err)
	}

	db := sqlx.NewDb(dbConn, "postgres")

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("fail to connect to the database: %w", err)
	}

	db.SetConnMaxLifetime(time.Duration(60) * time.Second)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)
	driver, err := postgres.WithInstance(db.DB, &postgres.Config{})
	if err != nil {
		return nil, fmt.Errorf("fail to create postgres migration driver: %w", err)
	}
	_, err = otelsql.RegisterDBStatsMetrics(db.DB)
	if err != nil {
		return nil, err
	}
	source, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return nil, fmt.Errorf("fail to create source  migration driver: %w", err)
	}
	m, err := migrate.NewWithInstance(
		"iofs",
		source,
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
