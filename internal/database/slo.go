package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/appclacks/server/pkg/slo/aggregates"
	er "github.com/mcorbin/corbierror"
)

type sloRecordAggregated struct {
	Name      string
	StartedAt time.Time `db:"started_at"`
	Success   bool
	Value     int64
}

type sloRecordSummed struct {
	Name    string
	Success bool
	Value   int64
}

type dbSLO struct {
	ID          string
	Name        string
	Description *string
	Labels      *string
	Objective   float32
	CreatedAt   time.Time `db:"created_at"`
}

func (c *Database) CreateSLO(ctx context.Context, slo aggregates.SLO) error {
	sloExists := dbSLO{}
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
	lock := fmt.Sprintf("slo-%s", slo.Name)
	_, err := tx.ExecContext(ctx, "SELECT pg_advisory_xact_lock(hashtext($1))", lock)
	if err != nil {
		return err
	}
	err = tx.GetContext(ctx, &sloExists, "SELECT slo.id, slo.name FROM slo WHERE name=$1", slo.Name)
	if err != nil {
		if err != sql.ErrNoRows {
			return fmt.Errorf("fail to get SLO%s: %w", slo.Name, err)
		}
	} else {
		return er.Newf("a SLO named %s already exists", er.Conflict, true, slo.Name)
	}
	labels, err := labelsToString(slo.Labels)
	if err != nil {
		return err
	}
	data := dbSLO{
		ID:          slo.ID,
		Name:        slo.Name,
		Description: slo.Description,
		Labels:      labels,
		Objective:   slo.Objective,
		CreatedAt:   slo.CreatedAt,
	}
	result, err := tx.NamedExecContext(ctx, "INSERT INTO slo (id, name, description, labels, created_at, objective) VALUES (:id, :name, :description, :labels, :created_at, :objective)", data)
	if err != nil {
		return fmt.Errorf("fail to create SLO %s: %w", data.Name, err)
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

func (c *Database) AddRecord(ctx context.Context, record aggregates.Record) error {
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
	lock := fmt.Sprintf("%s-%t", record.Name, record.Success)
	_, err := tx.ExecContext(ctx, "SELECT pg_advisory_xact_lock(hashtext($1))", lock)
	if err != nil {
		return err
	}

	currentAggregation := sloRecordAggregated{}
	err = c.db.GetContext(ctx, &currentAggregation, "SELECT name, started_at, success, value FROM slo_records_aggregated WHERE name=$1 AND success=$2 ORDER BY started_at DESC limit 1", record.Name, record.Success)
	if err != nil {
		if err != sql.ErrNoRows {
			return err
		}
	}
	if currentAggregation.Name == "" || currentAggregation.StartedAt.Before(time.Now().UTC().Add(-1*time.Hour)) {
		newSLO := sloRecordAggregated{
			Name:      record.Name,
			StartedAt: time.Now().UTC(),
			Value:     record.Value,
			Success:   record.Success,
		}
		c.Logger.Debug(fmt.Sprintf("creating slo %s - success %t", record.Name, record.Success))
		result, err := tx.NamedExecContext(ctx, "INSERT INTO slo_records_aggregated(name, started_at, success, value) VALUES (:name, :started_at, :success, :value)", newSLO)
		if err != nil {
			return err
		}
		err = checkResult(result, 1)
		if err != nil {
			return err
		}
	} else {
		updatedSLO := sloRecordAggregated{
			Name:      currentAggregation.Name,
			StartedAt: currentAggregation.StartedAt,
			Success:   currentAggregation.Success,
			Value:     currentAggregation.Value + record.Value,
		}
		c.Logger.Debug(fmt.Sprintf("updating SLO %s - success %t", currentAggregation.Name, currentAggregation.Success))
		result, err := c.db.NamedExecContext(ctx, "UPDATE slo_records_aggregated SET value=:value where name=:name AND started_at=:started_at AND success=:success", updatedSLO)
		if err != nil {
			return err
		}
		err = checkResult(result, 1)
		if err != nil {
			return err
		}
	}
	err = tx.Commit()
	if err != nil {
		return err
	}
	shouldRollback = false
	return nil
}

func (c *Database) ListAggregatedRecords(ctx context.Context, threshold time.Time) ([]*aggregates.SLOSum, error) {
	recordsSummed := []sloRecordSummed{}
	err := c.db.SelectContext(ctx, &recordsSummed, "SELECT name, success, sum(value) as value FROM slo_records_aggregated WHERE started_at > $1 GROUP BY (name, success)", threshold)
	if err != nil {
		return nil, err
	}
	sumMap := make(map[string]*aggregates.SLOSum)
	for i := range recordsSummed {
		dbSum := recordsSummed[i]
		_, ok := sumMap[dbSum.Name]
		if !ok {
			sumMap[dbSum.Name] = &aggregates.SLOSum{
				Name:      dbSum.Name,
				StartDate: threshold,
			}
		}
		if dbSum.Success {
			sumMap[dbSum.Name].Success = dbSum.Value
		} else {
			sumMap[dbSum.Name].Failure = dbSum.Value
		}

	}

	sumResult := []*aggregates.SLOSum{}
	for i := range sumMap {
		sloSum := sumMap[i]
		sumResult = append(sumResult, sloSum)
	}
	return sumResult, nil
}
