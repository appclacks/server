package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/appclacks/server/pkg/heartbeat/aggregates"
	er "github.com/mcorbin/corbierror"
)

type heartbeat struct {
	ID          string
	Name        string
	Description *string
	Labels      *string
	TTL         *string
	CreatedAt   time.Time  `db:"created_at"`
	RefreshedAt *time.Time `db:"refreshed_at"`
}

func toHeartbeat(heartbeat *heartbeat) (*aggregates.Heartbeat, error) {
	labels, err := stringToLabels(heartbeat.Labels)
	if err != nil {
		return nil, err
	}
	return &aggregates.Heartbeat{
		ID:          heartbeat.ID,
		Name:        heartbeat.Name,
		Description: heartbeat.Description,
		Labels:      labels,
		TTL:         heartbeat.TTL,
		CreatedAt:   heartbeat.CreatedAt.UTC(),
		RefreshedAt: heartbeat.RefreshedAt,
	}, nil
}

func (c *Database) CreateHeartbeat(ctx context.Context, hb *aggregates.Heartbeat) error {
	var checkExists heartbeat
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
	_, err := tx.ExecContext(ctx, "SELECT pg_advisory_xact_lock(hashtext($1))", hb.Name)
	if err != nil {
		return err
	}
	err = tx.GetContext(ctx, &checkExists, "SELECT heartbeat.id, heartbeat.name FROM heartbeat WHERE name=$1", hb.Name)
	if err != nil {
		if err != sql.ErrNoRows {
			return fmt.Errorf("fail to get heartbeat %s: %w", hb.Name, err)
		}
	} else {
		return er.Newf("a heartbeat named %s already exists", er.Conflict, true, hb.Name)
	}
	labels, err := labelsToString(hb.Labels)
	if err != nil {
		return err
	}
	dbHeartbeat := heartbeat{
		ID:          hb.ID,
		Name:        hb.Name,
		Labels:      labels,
		Description: hb.Description,
		TTL:         hb.TTL,
		CreatedAt:   hb.CreatedAt,
		RefreshedAt: hb.RefreshedAt,
	}
	result, err := tx.NamedExecContext(ctx, "INSERT INTO heartbeat (id, name, description, labels, ttl, created_at, refreshed_at) VALUES (:id, :name, :description, :labels, :ttl, :created_at, :refreshed_at)", dbHeartbeat)
	if err != nil {
		return fmt.Errorf("fail to create heartbeat %s: %w", hb.Name, err)
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

func (c *Database) GetHeartbeat(ctx context.Context, id string) (*aggregates.Heartbeat, error) {
	heartbeat := heartbeat{}
	err := c.db.GetContext(ctx, &heartbeat, "SELECT heartbeat.id, heartbeat.name, heartbeat.description, heartbeat.labels, heartbeat.ttl, heartbeat.created_at, heartbeat.refreshed_at FROM heartbeat WHERE id=$1", id)
	if err != nil {
		if err != sql.ErrNoRows {
			return nil, fmt.Errorf("fail to get heartbeat %s: %w", id, err)
		} else {
			return nil, er.New("heartbeat not found", er.NotFound, true)
		}
	}
	hb, err := toHeartbeat(&heartbeat)
	if err != nil {
		return nil, err
	}
	return hb, nil
}

func (c *Database) GetHeartbeatByName(ctx context.Context, name string) (*aggregates.Heartbeat, error) {
	heartbeat := heartbeat{}
	err := c.db.GetContext(ctx, &heartbeat, "SELECT heartbeat.id, heartbeat.name, heartbeat.description, heartbeat.labels, heartbeat.ttl, heartbeat.created_at, heartbeat.refreshed_at FROM heartbeat WHERE name=$1", name)
	if err != nil {
		if err != sql.ErrNoRows {
			return nil, fmt.Errorf("fail to get heartbeat %s: %w", name, err)
		} else {
			return nil, er.New("heartbeat not found", er.NotFound, true)
		}
	}
	return toHeartbeat(&heartbeat)
}

func (c *Database) DeleteHeartbeat(ctx context.Context, id string) error {
	_, err := c.GetHeartbeat(ctx, id)
	if err != nil {
		return err
	}
	result, err := c.db.ExecContext(ctx, "DELETE FROM heartbeat WHERE id=$1", id)
	if err != nil {
		return fmt.Errorf("fail to delete heartbeat: %w", err)
	}
	err = checkResult(result, 1)
	if err != nil {
		return err
	}
	return nil
}

func (c *Database) ListHeartbeats(ctx context.Context) ([]*aggregates.Heartbeat, error) {
	heartbeats := []heartbeat{}
	err := c.db.SelectContext(ctx, &heartbeats, "SELECT heartbeat.id, heartbeat.name, heartbeat.description, heartbeat.labels, heartbeat.ttl, heartbeat.created_at, heartbeat.refreshed_at FROM heartbeat")
	if err != nil {
		return nil, fmt.Errorf("fail to list heartbeats: %w", err)
	}
	result := []*aggregates.Heartbeat{}
	for i := range heartbeats {
		heartbeat := heartbeats[i]
		hb, err := toHeartbeat(&heartbeat)
		if err != nil {
			return nil, err
		}
		result = append(result, hb)
	}
	return result, nil
}

func (c *Database) UpdateHeartbeat(ctx context.Context, update *aggregates.Heartbeat) error {
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
	var checkExists heartbeat
	err = tx.GetContext(ctx, &checkExists, "SELECT heartbeat.id, heartbeat.name FROM heartbeat WHERE name=$1", update.Name)
	if err != nil {
		if err != sql.ErrNoRows {
			return fmt.Errorf("fail to get heartbeat %s: %w", update.Name, err)
		}
	} else {
		if checkExists.ID != update.ID {
			return er.Newf("A heartbeat named %s already exists", er.Conflict, true, checkExists.Name)
		}
	}
	labels, err := labelsToString(update.Labels)
	if err != nil {
		return err
	}
	dbHeartbeat := heartbeat{
		ID:          update.ID,
		Name:        update.Name,
		Labels:      labels,
		Description: update.Description,
		TTL:         update.TTL,
	}
	result, err := tx.NamedExecContext(ctx, "UPDATE heartbeat SET name=:name, description=:description, labels=:labels, ttl=:ttl WHERE id=:id", dbHeartbeat)
	if err != nil {
		return fmt.Errorf("fail to update heartbeat %s: %w", update.ID, err)
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

func (c *Database) RefreshHeartbeat(ctx context.Context, id string) error {
	now := time.Now().UTC()
	result, err := c.db.ExecContext(ctx, "UPDATE heartbeat SET refreshed_at=$1 WHERE id=$2", now, id)
	if err != nil {
		return fmt.Errorf("fail to refresh heartbeat: %w", err)
	}
	err = checkResult(result, 1)
	if err != nil {
		return err
	}
	return nil
}

func (c *Database) CountHeartbeats(ctx context.Context) (int, error) {
	var count int
	row := c.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM heartbeat")
	err := row.Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}
