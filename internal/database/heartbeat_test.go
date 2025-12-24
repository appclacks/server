package database_test

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/appclacks/server/pkg/heartbeat/aggregates"
	"github.com/baidubce/bce-sdk-go/util"
	"github.com/stretchr/testify/assert"
)

func TestHeartbeatCRUD(t *testing.T) {
	labels := map[string]string{"env": "test", "app": "myapp"}
	description := "test heartbeat"
	ttl := "30s"

	heartbeat := aggregates.Heartbeat{
		ID:          util.NewUUID(),
		CreatedAt:   time.Now().UTC(),
		Name:        "test-heartbeat",
		Description: &description,
		Labels:      labels,
		TTL:         &ttl,
		RefreshedAt: nil,
	}

	err := TestComponent.CreateHeartbeat(context.Background(), &heartbeat)
	assert.NoError(t, err)

	count, err := TestComponent.CountHeartbeats(context.Background())
	assert.NoError(t, err)
	assert.True(t, count > 0)

	err = TestComponent.CreateHeartbeat(context.Background(), &heartbeat)
	assert.ErrorContains(t, err, "already exists")

	checkGet, err := TestComponent.GetHeartbeat(context.Background(), heartbeat.ID)
	assert.NoError(t, err)
	if checkGet.ID == "" || checkGet.Name != heartbeat.Name || !reflect.DeepEqual(checkGet.Description, heartbeat.Description) || !reflect.DeepEqual(checkGet.Labels, heartbeat.Labels) || checkGet.CreatedAt.IsZero() {
		t.Fatalf("Invalid heartbeat returned by GetHeartbeat\n%+v", checkGet)
	}

	checkGetByName, err := TestComponent.GetHeartbeatByName(context.Background(), heartbeat.Name)
	assert.NoError(t, err)
	if checkGetByName.ID == "" || checkGetByName.Name != heartbeat.Name || !reflect.DeepEqual(checkGetByName.Description, heartbeat.Description) || !reflect.DeepEqual(checkGetByName.Labels, heartbeat.Labels) || checkGetByName.CreatedAt.IsZero() {
		t.Fatalf("Invalid heartbeat returned by GetHeartbeatByName\n%+v", checkGetByName)
	}

	listHeartbeats, err := TestComponent.ListHeartbeats(context.Background())
	assert.NoError(t, err)
	assert.True(t, len(listHeartbeats) > 0)

	firstHeartbeat := listHeartbeats[0]
	newLabels := map[string]string{"update": "yes"}
	newDesc := "updated description"
	newTTL := "60s"

	heartbeatUpdate := aggregates.Heartbeat{
		ID:          firstHeartbeat.ID,
		Name:        "newname",
		Labels:      newLabels,
		Description: &newDesc,
		TTL:         &newTTL,
	}

	err = TestComponent.UpdateHeartbeat(context.Background(), &heartbeatUpdate)
	assert.NoError(t, err)

	checkGet, err = TestComponent.GetHeartbeat(context.Background(), heartbeat.ID)
	assert.NoError(t, err)
	if checkGet.ID == "" || checkGet.Name != heartbeatUpdate.Name || !reflect.DeepEqual(checkGet.Description, heartbeatUpdate.Description) || !reflect.DeepEqual(checkGet.Labels, heartbeatUpdate.Labels) || checkGet.CreatedAt.IsZero() || !reflect.DeepEqual(checkGet.TTL, heartbeatUpdate.TTL) {
		t.Fatalf("Invalid heartbeat returned by GetHeartbeat after update\n%+v", checkGet)
	}

	err = TestComponent.UpdateHeartbeat(context.Background(), &heartbeatUpdate)
	assert.NoError(t, err)

	err = TestComponent.RefreshHeartbeat(context.Background(), heartbeat.ID)
	assert.NoError(t, err)

	checkGet, err = TestComponent.GetHeartbeat(context.Background(), heartbeat.ID)
	assert.NoError(t, err)
	assert.NotNil(t, checkGet.RefreshedAt)

	err = TestComponent.DeleteHeartbeat(context.Background(), heartbeat.ID)
	assert.NoError(t, err)

	err = TestComponent.DeleteHeartbeat(context.Background(), heartbeat.ID)
	assert.ErrorContains(t, err, "not found")
}

func TestHeartbeatCreateConflict(t *testing.T) {
	labels := map[string]string{"env": "test"}
	description := "test heartbeat"
	ttl := "30s"

	heartbeat1 := aggregates.Heartbeat{
		ID:          util.NewUUID(),
		CreatedAt:   time.Now().UTC(),
		Name:        "conflict-test-heartbeat",
		Description: &description,
		Labels:      labels,
		TTL:         &ttl,
	}

	heartbeat2 := aggregates.Heartbeat{
		ID:          util.NewUUID(),
		CreatedAt:   time.Now().UTC(),
		Name:        "conflict-test-heartbeat",
		Description: &description,
		Labels:      labels,
		TTL:         &ttl,
	}

	err := TestComponent.CreateHeartbeat(context.Background(), &heartbeat1)
	assert.NoError(t, err)

	err = TestComponent.CreateHeartbeat(context.Background(), &heartbeat2)
	assert.ErrorContains(t, err, "already exists")

	err = TestComponent.DeleteHeartbeat(context.Background(), heartbeat1.ID)
	assert.NoError(t, err)
}

func TestHeartbeatUpdateConflict(t *testing.T) {
	labels := map[string]string{"env": "test"}
	description := "test heartbeat"
	ttl := "30s"

	heartbeat1 := aggregates.Heartbeat{
		ID:          util.NewUUID(),
		CreatedAt:   time.Now().UTC(),
		Name:        "update-conflict-test-1",
		Description: &description,
		Labels:      labels,
		TTL:         &ttl,
	}

	heartbeat2 := aggregates.Heartbeat{
		ID:          util.NewUUID(),
		CreatedAt:   time.Now().UTC(),
		Name:        "update-conflict-test-2",
		Description: &description,
		Labels:      labels,
		TTL:         &ttl,
	}

	err := TestComponent.CreateHeartbeat(context.Background(), &heartbeat1)
	assert.NoError(t, err)

	err = TestComponent.CreateHeartbeat(context.Background(), &heartbeat2)
	assert.NoError(t, err)

	heartbeat1.Name = "update-conflict-test-2"
	err = TestComponent.UpdateHeartbeat(context.Background(), &heartbeat1)
	assert.ErrorContains(t, err, "already exists")

	err = TestComponent.DeleteHeartbeat(context.Background(), heartbeat1.ID)
	assert.NoError(t, err)

	err = TestComponent.DeleteHeartbeat(context.Background(), heartbeat2.ID)
	assert.NoError(t, err)
}

func TestHeartbeatGetNotFound(t *testing.T) {
	nonExistentID := util.NewUUID()

	_, err := TestComponent.GetHeartbeat(context.Background(), nonExistentID)
	assert.ErrorContains(t, err, "not found")

	_, err = TestComponent.GetHeartbeatByName(context.Background(), "non-existent-heartbeat")
	assert.ErrorContains(t, err, "not found")
}

func TestHeartbeatRefreshNonExistent(t *testing.T) {
	nonExistentID := util.NewUUID()

	err := TestComponent.RefreshHeartbeat(context.Background(), nonExistentID)
	assert.Error(t, err)
}
