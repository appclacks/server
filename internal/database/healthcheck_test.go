package database_test

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/appclacks/server/pkg/healthcheck/aggregates"
	"github.com/baidubce/bce-sdk-go/util"
	"github.com/stretchr/testify/assert"
)

func TestHealthcheckCRUD(t *testing.T) {
	def1 := &aggregates.HealthcheckDNSDefinition{
		Domain: "mcorbin.fr",
	}

	labels := map[string]string{"aze": "a", "b": "c"}
	healthcheck := aggregates.Healthcheck{
		ID:         util.NewUUID(),
		CreatedAt:  time.Now(),
		Name:       "test",
		Type:       "dns",
		Timeout:    "3s",
		Labels:     labels,
		Definition: def1,
		Interval:   "60s",
	}
	err := TestComponent.CreateHealthcheck(context.Background(), &healthcheck)
	assert.NoError(t, err)
	count, err := TestComponent.CountHealthchecks(context.Background())
	assert.NoError(t, err)
	assert.True(t, count > 0)
	err = TestComponent.CreateHealthcheck(context.Background(), &healthcheck)
	assert.ErrorContains(t, err, "already exists")

	checkGet, err := TestComponent.GetHealthcheck(context.Background(), healthcheck.ID)
	assert.NoError(t, err)
	if checkGet.ID == "" || checkGet.Name != healthcheck.Name || checkGet.Description != healthcheck.Description || !reflect.DeepEqual(checkGet.Labels, healthcheck.Labels) || checkGet.CreatedAt.IsZero() {
		t.Fatalf("Invalid healcheck returned by CreateHealthcheck\n%+v", checkGet)
	}

	checkGetByName, err := TestComponent.GetHealthcheckByName(context.Background(), healthcheck.Name)
	assert.NoError(t, err)
	if checkGetByName.ID == "" || checkGetByName.Name != healthcheck.Name || checkGetByName.Description != healthcheck.Description || !reflect.DeepEqual(checkGetByName.Labels, healthcheck.Labels) || checkGetByName.CreatedAt.IsZero() {
		t.Fatalf("Invalid healcheck returned by CreateHealthcheck\n%+v", checkGetByName)
	}

	listChecks, err := TestComponent.ListHealthchecks(context.Background())
	assert.NoError(t, err)
	firstCheck := listChecks[0]
	newLabels := map[string]string{"update": "yes"}

	newDesc := "updated description"
	newDef := &aggregates.HealthcheckDNSDefinition{
		Domain: "update.mcorbin.fr",
	}
	healthcheckUpdate := aggregates.Healthcheck{
		ID:          firstCheck.ID,
		Name:        "newname",
		Labels:      newLabels,
		Description: &newDesc,
		Interval:    "300s",
		Timeout:     "3s",
		Enabled:     true,
		Definition:  newDef,
	}
	err = TestComponent.UpdateHealthcheck(context.Background(), &healthcheckUpdate)
	assert.NoError(t, err)
	checkGet, err = TestComponent.GetHealthcheck(context.Background(), healthcheck.ID)
	assert.NoError(t, err)
	if checkGet.ID == "" || checkGet.Name != healthcheckUpdate.Name || *checkGet.Description != *healthcheckUpdate.Description || !reflect.DeepEqual(checkGet.Labels, healthcheckUpdate.Labels) || checkGet.CreatedAt.IsZero() || checkGet.Enabled != healthcheckUpdate.Enabled || checkGet.Interval != healthcheckUpdate.Interval {
		t.Fatalf("Invalid healcheck returned by GetHealthcheck after update\n%+v", checkGet)
	}

	err = TestComponent.UpdateHealthcheck(context.Background(), &healthcheckUpdate)
	assert.NoError(t, err)

	err = TestComponent.DeleteHealthcheck(context.Background(), healthcheck.ID)
	assert.NoError(t, err)

	err = TestComponent.DeleteHealthcheck(context.Background(), healthcheck.ID)
	assert.ErrorContains(t, err, "not found")
}
