package database_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/appclacks/server/pkg/pushgateway/aggregates"
	"github.com/stretchr/testify/assert"
)

func TestPushgatewayCRUD(t *testing.T) {
	// create
	desc1 := "description"
	ttl1 := "60s"
	now1 := time.Now().UTC().Round(1 * time.Second)
	expired1 := now1.Add(100 * time.Second).Round(1 * time.Second)
	def1 := aggregates.PushgatewayMetric{
		Name:        "test1",
		Description: &desc1,
		Labels: map[string]string{
			"a":   "b",
			"foo": "bar",
		},
		TTL:       &ttl1,
		CreatedAt: now1,
		ExpiresAt: &expired1,
		Value:     1000,
	}
	id1, err := TestComponent.CreateOrUpdatePushgatewayMetric(context.Background(), def1, false)
	assert.NoError(t, err)
	metric, err := TestComponent.GetMetrics(context.Background())
	assert.NoError(t, err)
	assert.Len(t, metric, 1)

	m1 := metric[0]
	assert.NotEmpty(t, m1.ID)
	assert.Equal(t, id1, m1.ID)
	assert.Equal(t, m1.Name, def1.Name)
	assert.Equal(t, *m1.Description, *def1.Description)
	assert.Equal(t, m1.Labels, def1.Labels)
	assert.Equal(t, *m1.TTL, *def1.TTL)
	assert.Equal(t, m1.Value, def1.Value)
	assert.Equal(t, def1.ExpiresAt, m1.ExpiresAt)
	assert.Equal(t, def1.CreatedAt, m1.CreatedAt)

	// update
	desc2 := "desc2"
	ttl2 := "120ss"
	now2 := time.Now().UTC().Round(2 * time.Second)
	expired2 := now2.Add(200 * time.Second).Round(1 * time.Second)

	update1 := aggregates.PushgatewayMetric{
		Name:        "test1",
		Description: &desc2,
		Labels: map[string]string{
			"a":   "b",
			"foo": "bar",
		},
		TTL:       &ttl2,
		CreatedAt: now2,
		ExpiresAt: &expired2,
		Value:     4000,
	}

	id2, err := TestComponent.CreateOrUpdatePushgatewayMetric(context.Background(), update1, false)
	assert.NoError(t, err)
	metric, err = TestComponent.GetMetrics(context.Background())
	assert.NoError(t, err)
	assert.Len(t, metric, 1)

	m2 := metric[0]
	assert.NotEmpty(t, m2.ID)
	assert.Equal(t, id2, m1.ID)
	assert.Equal(t, id2, m2.ID)
	assert.Equal(t, m2.Name, update1.Name)
	assert.Equal(t, *m2.Description, *update1.Description)
	assert.Equal(t, m2.Labels, update1.Labels)
	assert.Equal(t, *m2.TTL, *update1.TTL)
	assert.Equal(t, m2.Value, update1.Value)
	assert.Equal(t, update1.ExpiresAt, m2.ExpiresAt)
	assert.Equal(t, update1.CreatedAt, m2.CreatedAt)

	// cumulative test
	id3, err := TestComponent.CreateOrUpdatePushgatewayMetric(context.Background(), update1, true)
	assert.NoError(t, err)
	assert.Equal(t, id3, m1.ID)
	metric, err = TestComponent.GetMetrics(context.Background())
	assert.NoError(t, err)
	assert.Len(t, metric, 1)
	assert.Equal(t, metric[0].Value, float64(8000))

	// new metric by name
	desc3 := "description"
	ttl3 := "60s"
	now3 := time.Now().UTC().Round(1 * time.Second)
	expired3 := now3.Add(100 * time.Second).Round(1 * time.Second)
	def3 := aggregates.PushgatewayMetric{
		Name:        "test3",
		Description: &desc3,
		Labels: map[string]string{
			"a":   "b",
			"foo": "bar",
		},
		TTL:       &ttl3,
		CreatedAt: now3,
		ExpiresAt: &expired3,
		Value:     1000,
	}

	_, err = TestComponent.CreateOrUpdatePushgatewayMetric(context.Background(), def3, false)
	assert.NoError(t, err)
	metric, err = TestComponent.GetMetrics(context.Background())
	assert.NoError(t, err)
	assert.Len(t, metric, 2)

	// new metric by label
	desc4 := "description"
	ttl4 := "60s"
	now4 := time.Now().UTC().Round(1 * time.Second)
	expired4 := now4.Add(100 * time.Second).Round(1 * time.Second)
	def4 := aggregates.PushgatewayMetric{
		Name:        "test4",
		Description: &desc4,
		Labels: map[string]string{
			"new": "label",
			"foo": "bar",
		},
		TTL:       &ttl4,
		CreatedAt: now4,
		ExpiresAt: &expired4,
		Value:     1000,
	}

	_, err = TestComponent.CreateOrUpdatePushgatewayMetric(context.Background(), def4, false)
	assert.NoError(t, err)
	metric, err = TestComponent.GetMetrics(context.Background())
	assert.NoError(t, err)
	assert.Len(t, metric, 3)

	// delete by ID

	err = TestComponent.DeleteMetricByID(context.Background(), m1.ID)
	assert.NoError(t, err)
	metric, err = TestComponent.GetMetrics(context.Background())
	assert.NoError(t, err)
	assert.Len(t, metric, 2)

	err = TestComponent.DeleteMetricByID(context.Background(), m1.ID)
	assert.ErrorContains(t, err, "not found")

	err = TestComponent.DeleteMetricsByName(context.Background(), "test4")
	assert.NoError(t, err)
	metric, err = TestComponent.GetMetrics(context.Background())
	assert.NoError(t, err)
	assert.Len(t, metric, 1)
	err = TestComponent.DeleteMetricsByName(context.Background(), "test4")
	assert.ErrorContains(t, err, "no metric were deleted")

	// cleanup

	now := time.Now().UTC()
	for i := 0; i < 10; i++ {
		expiresAt := now
		if i > 4 {
			expiresAt = now.Add(300 * time.Second)
		}
		def1 := aggregates.PushgatewayMetric{
			Name:      fmt.Sprintf("test%d", i),
			TTL:       &ttl1,
			CreatedAt: now1,
			ExpiresAt: &expiresAt,
			Value:     1000,
		}
		_, err := TestComponent.CreateOrUpdatePushgatewayMetric(context.Background(), def1, false)
		assert.NoError(t, err)
	}
	deleteCount, err := TestComponent.CleanPushgatewayMetrics(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, int64(5), deleteCount)

	metric, err = TestComponent.GetMetrics(context.Background())
	assert.NoError(t, err)
	assert.Len(t, metric, 5)

	err = TestComponent.DeleteAllPushgatewayMetrics(context.Background())
	assert.NoError(t, err)

	metric, err = TestComponent.GetMetrics(context.Background())
	assert.NoError(t, err)
	assert.Len(t, metric, 0)
}
