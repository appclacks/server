package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/appclacks/server/internal/util"
	"github.com/appclacks/server/pkg/pushgateway/aggregates"
	er "github.com/mcorbin/corbierror"
)

type pushgatewayMetric struct {
	ID          string
	Name        string
	Description *string
	Labels      *string
	TTL         *string
	Type        *string
	CreatedAt   time.Time  `db:"created_at"`
	ExpiresAt   *time.Time `db:"expires_at"`
	Value       float32
}

func toPushGatewayMetric(metric *pushgatewayMetric) (*aggregates.PushgatewayMetric, error) {
	labels, err := stringToLabels(metric.Labels)
	if err != nil {
		return nil, err
	}

	result := &aggregates.PushgatewayMetric{
		ID:          metric.ID,
		Name:        metric.Name,
		Description: metric.Description,
		Labels:      labels,
		TTL:         metric.TTL,
		Type:        metric.Type,
		CreatedAt:   metric.CreatedAt.UTC(),
		Value:       metric.Value,
	}
	if metric.ExpiresAt != nil {
		expiresAt := metric.ExpiresAt.UTC()
		result.ExpiresAt = &expiresAt
	}
	return result, nil
}

func (c *Database) CreateOrUpdatePushgatewayMetric(ctx context.Context, metric aggregates.PushgatewayMetric, cumulative bool) (string, error) {
	metricID := ""
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
	_, err := tx.ExecContext(ctx, "SELECT pg_advisory_xact_lock(hashtext($1))", metric.Name)
	if err != nil {
		return "", err
	}

	currentMetric := pushgatewayMetric{}
	labelCondition := "{}"
	if metric.Labels != nil {
		labelString, err := labelsToString(metric.Labels)
		if err != nil {
			return "", err
		}
		labelCondition = *labelString
	}
	err = c.db.GetContext(ctx, &currentMetric, "SELECT id, name, value FROM pushgateway_metric WHERE name=$1 AND labels @> $2 AND $2 <@ labels", metric.Name, labelCondition)
	if err != nil {
		if err != sql.ErrNoRows {
			return "", err
		}
	}
	if currentMetric.ID == "" {
		metricID = util.NewUUID()
		newMetric := pushgatewayMetric{
			ID:          metricID,
			Name:        metric.Name,
			Description: metric.Description,
			Labels:      &labelCondition,
			TTL:         metric.TTL,
			Type:        metric.Type,
			CreatedAt:   metric.CreatedAt,
			ExpiresAt:   metric.ExpiresAt,
			Value:       metric.Value,
		}
		c.Logger.Debug(fmt.Sprintf("creating metric %s", metric.Name))
		result, err := tx.NamedExecContext(ctx, "INSERT INTO pushgateway_metric(id, name, description, ttl, labels, value, type, created_at, expires_at) VALUES (:id, :name, :description, :ttl, :labels, :value, :type, :created_at, :expires_at)", newMetric)
		if err != nil {
			return "", err
		}
		err = checkResult(result, 1)
		if err != nil {
			return "", err
		}
	} else {
		metricID = currentMetric.ID
		metricValue := metric.Value
		if cumulative {
			metricValue += currentMetric.Value
		}
		updatedMetric := pushgatewayMetric{
			ID:          currentMetric.ID,
			Description: metric.Description,
			TTL:         metric.TTL,
			Type:        metric.Type,
			CreatedAt:   metric.CreatedAt,
			ExpiresAt:   metric.ExpiresAt,
			Value:       metricValue,
		}
		c.Logger.Debug(fmt.Sprintf("updating metric %s", metric.Name))
		result, err := c.db.NamedExecContext(ctx, "UPDATE pushgateway_metric SET description=:description, ttl=:ttl, type=:type, created_at=:created_at, expires_at=:expires_at, value=:value where id=:id", updatedMetric)
		if err != nil {
			return "", err
		}
		err = checkResult(result, 1)
		if err != nil {
			return "", err
		}
	}
	err = tx.Commit()
	if err != nil {
		return "", err
	}
	shouldRollback = false
	return metricID, nil
}

func (c *Database) GetMetrics(ctx context.Context) ([]*aggregates.PushgatewayMetric, error) {
	metrics := []pushgatewayMetric{}
	err := c.db.SelectContext(ctx, &metrics, "SELECT id, name, description, ttl, labels, value, type, created_at, expires_at FROM pushgateway_metric")
	if err != nil {
		return nil, err
	}
	result := []*aggregates.PushgatewayMetric{}
	for i := range metrics {
		dbMetric := metrics[i]
		metric, err := toPushGatewayMetric(&dbMetric)
		if err != nil {
			return nil, err
		}
		result = append(result, metric)
	}
	return result, nil
}

func (c *Database) DeleteMetricsByName(ctx context.Context, name string) error {
	result, err := c.db.ExecContext(ctx, "DELETE FROM pushgateway_metric WHERE name=$1", name)
	if err != nil {
		return fmt.Errorf("fail to delete metrics: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("fail to check affected row: %w", err)
	}
	if affected == 0 {
		return er.New("no metric were deleted", er.NotFound, true)
	}
	return nil
}

func (c *Database) DeleteMetricByID(ctx context.Context, id string) error {
	result, err := c.db.ExecContext(ctx, "DELETE FROM pushgateway_metric WHERE id=$1", id)
	if err != nil {
		return fmt.Errorf("fail to delete metric: %w", err)
	}
	return checkResult(result, 1)
}

func (c *Database) CleanPushgatewayMetrics(ctx context.Context) (int64, error) {
	result, err := c.db.ExecContext(ctx, "DELETE FROM pushgateway_metric WHERE expires_at < $1", time.Now().UTC())
	if err != nil {
		return 0, fmt.Errorf("fail to clean metric: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("fail to check affected row: %w", err)
	}
	return affected, nil
}
