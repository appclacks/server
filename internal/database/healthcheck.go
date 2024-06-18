package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/appclacks/server/pkg/healthcheck/aggregates"

	er "github.com/mcorbin/corbierror"
)

type dbHealthcheck struct {
	ID          string
	Name        string
	Description *string
	Labels      *string
	RandomID    int       `db:"random_id"`
	CreatedAt   time.Time `db:"created_at"`
	Type        string
	Interval    string
	Timeout     string
	Enabled     bool
	Definition  string
}

func toHealthcheck(healthcheck *dbHealthcheck) (*aggregates.Healthcheck, error) {
	labels, err := stringToLabels(healthcheck.Labels)
	if err != nil {
		return nil, err
	}
	def, err := aggregates.ToHealthcheckDefinition(healthcheck.Type, healthcheck.Definition)
	if err != nil {
		return nil, err
	}
	return &aggregates.Healthcheck{
		ID:          healthcheck.ID,
		Name:        healthcheck.Name,
		Description: healthcheck.Description,
		Labels:      labels,
		CreatedAt:   healthcheck.CreatedAt.UTC(),
		Interval:    healthcheck.Interval,
		Timeout:     healthcheck.Timeout,
		Definition:  def,
		Enabled:     healthcheck.Enabled,
		Type:        healthcheck.Type,
		RandomID:    healthcheck.RandomID,
	}, nil
}

func (c *Database) CreateHealthcheck(ctx context.Context, healthcheck *aggregates.Healthcheck) error {
	checkExists := dbHealthcheck{}
	tx := c.db.MustBegin()
	shouldRollback := true
	defer func() {
		if shouldRollback {
			err := tx.Rollback()
			if err != nil {
				c.Logger.Error(err.Error())
			}
		}
	}()
	_, err := tx.ExecContext(ctx, "SELECT pg_advisory_xact_lock(hashtext($1))", healthcheck.Name)
	if err != nil {
		return err
	}
	err = tx.GetContext(ctx, &checkExists, "SELECT healthcheck.id, healthcheck.name FROM healthcheck WHERE name=$1", healthcheck.Name)
	if err != nil {
		if err != sql.ErrNoRows {
			return fmt.Errorf("fail to get healthcheck %s: %w", healthcheck.Name, err)
		}
	} else {
		return er.Newf("a healthcheck named %s already exists", er.Conflict, true, healthcheck.Name)
	}
	labels, err := labelsToString(healthcheck.Labels)
	if err != nil {
		return err
	}
	def, err := healthcheck.Definition.String()
	if err != nil {
		return err
	}
	dbHealthcheck := dbHealthcheck{
		ID:          healthcheck.ID,
		Name:        healthcheck.Name,
		Labels:      labels,
		Description: healthcheck.Description,
		Type:        healthcheck.Type,
		CreatedAt:   healthcheck.CreatedAt,
		Interval:    healthcheck.Interval,
		Timeout:     healthcheck.Timeout,
		Enabled:     healthcheck.Enabled,
		RandomID:    healthcheck.RandomID,
		Definition:  def,
	}
	result, err := tx.NamedExecContext(ctx, "INSERT INTO healthcheck (id, name, description, labels, created_at, definition, type, interval, random_id, enabled, timeout) VALUES (:id, :name, :description, :labels, :created_at, :definition, :type, :interval, :random_id, :enabled, :timeout)", dbHealthcheck)
	if err != nil {
		return fmt.Errorf("fail to create healthcheck %s: %w", healthcheck.Name, err)
	}
	err = checkResult(result, 1)
	if err != nil {
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	shouldRollback = false
	return nil
}

func (c *Database) GetHealthcheck(ctx context.Context, id string) (*aggregates.Healthcheck, error) {
	healthcheck := dbHealthcheck{}
	err := c.db.GetContext(ctx, &healthcheck, "SELECT healthcheck.id, healthcheck.name, healthcheck.description, healthcheck.labels, healthcheck.created_at, healthcheck.definition, healthcheck.type, healthcheck.interval, healthcheck.random_id, healthcheck.enabled, healthcheck.timeout FROM healthcheck WHERE id=$1", id)
	if err != nil {
		if err != sql.ErrNoRows {
			return nil, fmt.Errorf("fail to get healthcheck %s: %w", id, err)
		} else {
			return nil, er.New("healthcheck not found", er.NotFound, true)
		}
	}
	check, err := toHealthcheck(&healthcheck)
	if err != nil {
		return nil, err
	}
	return check, nil
}

func (c *Database) GetHealthcheckByName(ctx context.Context, name string) (*aggregates.Healthcheck, error) {
	healthcheck := dbHealthcheck{}
	err := c.db.GetContext(ctx, &healthcheck, "SELECT healthcheck.id, healthcheck.name, healthcheck.description, healthcheck.labels, healthcheck.created_at, healthcheck.definition, healthcheck.type, healthcheck.interval, healthcheck.random_id, healthcheck.enabled, healthcheck.timeout FROM healthcheck WHERE name=$1", name)
	if err != nil {
		if err != sql.ErrNoRows {
			return nil, fmt.Errorf("fail to get healthcheck %s: %w", name, err)
		} else {
			return nil, er.New("healthcheck not found", er.NotFound, true)
		}
	}
	return toHealthcheck(&healthcheck)
}

func (c *Database) DeleteHealthcheck(ctx context.Context, id string) error {
	_, err := c.GetHealthcheck(ctx, id)
	if err != nil {
		return err
	}
	result, err := c.db.ExecContext(ctx, "DELETE FROM healthcheck WHERE id=$1", id)
	if err != nil {
		return fmt.Errorf("fail to delete healthcheck: %w", err)
	}
	err = checkResult(result, 1)
	if err != nil {
		return err
	}
	return nil
}

func (c *Database) ListHealthchecks(ctx context.Context, enabled *bool) ([]*aggregates.Healthcheck, error) {
	healthchecks := []dbHealthcheck{}
	baseQuery := "SELECT healthcheck.id, healthcheck.name, healthcheck.description, healthcheck.labels, healthcheck.created_at, healthcheck.definition, healthcheck.type, healthcheck.interval, healthcheck.random_id, healthcheck.enabled, healthcheck.timeout FROM healthcheck"
	if enabled != nil {
		baseQuery = fmt.Sprintf("%s WHERE enabled is %t", baseQuery, *enabled)
	}
	err := c.db.SelectContext(ctx, &healthchecks, baseQuery)
	if err != nil {
		return nil, fmt.Errorf("fail to list healthchecks: %w", err)
	}
	result := []*aggregates.Healthcheck{}
	for i := range healthchecks {
		healthcheck := healthchecks[i]
		hc, err := toHealthcheck(&healthcheck)
		if err != nil {
			return nil, err
		}
		result = append(result, hc)
	}

	return result, nil
}

func (c *Database) ListHealthchecksForProber(ctx context.Context, prober int) ([]*aggregates.Healthcheck, error) {
	healthchecks := []dbHealthcheck{}
	err := c.db.SelectContext(ctx, &healthchecks, "SELECT healthcheck.id, healthcheck.name, healthcheck.description, healthcheck.labels, healthcheck.created_at, healthcheck.definition, healthcheck.type, healthcheck.interval, healthcheck.random_id, healthcheck.enabled, healthcheck.timeout FROM healthcheck WHERE healthcheck.random_id%$1=$2 AND healthcheck.enabled=true", c.probers, prober)
	if err != nil {
		return nil, fmt.Errorf("fail to list healthchecks: %w", err)
	}
	result := []*aggregates.Healthcheck{}
	for i := range healthchecks {
		healthcheck := healthchecks[i]
		hc, err := toHealthcheck(&healthcheck)
		if err != nil {
			return nil, err
		}
		result = append(result, hc)
	}

	return result, nil
}

func (c *Database) UpdateHealthcheck(ctx context.Context, update *aggregates.Healthcheck) error {
	tx := c.db.MustBegin()
	shouldRollback := true
	defer func() {
		if shouldRollback {
			err := tx.Rollback()
			if err != nil {
				c.Logger.Error(err.Error())
			}
		}
	}()
	_, err := tx.ExecContext(ctx, "SELECT pg_advisory_xact_lock(hashtext($1))", update.Name)
	if err != nil {
		return err
	}
	checkExists := dbHealthcheck{}
	err = tx.GetContext(ctx, &checkExists, "SELECT healthcheck.id, healthcheck.name FROM healthcheck WHERE name=$1", update.Name)
	if err != nil {
		if err != sql.ErrNoRows {
			return fmt.Errorf("fail to get healthcheck %s: %w", update.Name, err)
		}
	} else {
		if checkExists.ID != update.ID {
			return er.Newf("A healthcheck named %s already exists", er.Conflict, true, checkExists.Name)
		}
	}
	labels, err := labelsToString(update.Labels)
	if err != nil {
		return err
	}
	def, err := update.Definition.String()
	if err != nil {
		return err
	}
	dbHealthcheck := dbHealthcheck{
		ID:          update.ID,
		Name:        update.Name,
		Labels:      labels,
		Description: update.Description,
		Interval:    update.Interval,
		Timeout:     update.Timeout,
		Enabled:     update.Enabled,
		Definition:  def,
	}
	result, err := c.db.NamedExecContext(ctx, "update healthcheck set name=:name, description=:description, labels=:labels, definition=:definition, interval=:interval, enabled=:enabled, timeout=:timeout where id=:id", dbHealthcheck)
	if err != nil {
		return fmt.Errorf("fail to update healthcheck %s: %w", update.ID, err)
	}
	err = checkResult(result, 1)
	if err != nil {
		return err
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	shouldRollback = false
	return nil
}

func (c *Database) CountHealthchecks(ctx context.Context) (int, error) {
	var count int
	row := c.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM healthcheck")
	err := row.Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}
