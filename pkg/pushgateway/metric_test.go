package pushgateway_test

import (
	"context"
	"log/slog"
	"testing"

	mocks "github.com/appclacks/server/mocks/github.com/appclacks/server/pkg/pushgateway"
	"github.com/appclacks/server/pkg/pushgateway"
	"github.com/appclacks/server/pkg/pushgateway/aggregates"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestPrometheusMetrics(t *testing.T) {
	store := new(mocks.MockStore)
	reg := prometheus.NewRegistry()
	logger := slog.Default()

	service, err := pushgateway.New(logger, store, reg)
	assert.NoError(t, err)

	desc1 := "my description"
	type1 := "gauge"

	desc2 := "my super description"
	type2 := "counter"

	cases := []struct {
		result  string
		metrics []*aggregates.PushgatewayMetric
	}{
		{
			metrics: []*aggregates.PushgatewayMetric{},
			result:  "",
		},
		{
			metrics: []*aggregates.PushgatewayMetric{
				{
					Name:  "metric1",
					Value: "123.4",
				},
			},
			result: "metric1{} 123.4\n",
		},
		{
			metrics: []*aggregates.PushgatewayMetric{
				{
					Name:        "metric2",
					Description: &desc1,
					Type:        &type1,
					Value:       "121",
				},
			},
			result: `# HELP metric2 my description
# TYPE metric2 gauge
metric2{} 121
`,
		},
		{
			metrics: []*aggregates.PushgatewayMetric{
				{
					Name:        "metric2",
					Description: &desc1,
					Type:        &type1,
					Labels: map[string]string{
						"env": "test",
					},
					Value: "121",
				},
			},
			result: `# HELP metric2 my description
# TYPE metric2 gauge
metric2{env="test"} 121
`,
		},
		{
			metrics: []*aggregates.PushgatewayMetric{
				{
					Name:        "metric1",
					Description: &desc1,
					Type:        &type1,
					Value:       "121",
				},
				{
					Name:  "metric2",
					Value: "10.1",
					Labels: map[string]string{
						"env":  "prod",
						"team": "sre",
					},
				},
				{
					Name:        "metric2",
					Description: &desc2,
					Type:        &type2,
					Labels: map[string]string{
						"env":  "staging",
						"team": "data",
					},
					Value: "121",
				},
			},
			result: `# HELP metric1 my description
# TYPE metric1 gauge
metric1{} 121
# HELP metric2 my super description
# TYPE metric2 counter
metric2{env="prod", team="sre"} 10.1
metric2{env="staging", team="data"} 121
`,
		},
		{
			metrics: []*aggregates.PushgatewayMetric{
				{
					Name:  "metric3",
					Value: "123.4567",
					Labels: map[string]string{
						"team": "backend",
					},
				},
				{
					Name:  "metric4",
					Value: "123.4567",
					Labels: map[string]string{
						"team": "front",
					},
				},
				{
					Name:        "metric1",
					Description: &desc1,
					Type:        &type1,
					Value:       "121",
				},
				{
					Name:  "metric2",
					Value: "10.1",
					Labels: map[string]string{
						"env":  "prod",
						"team": "sre",
					},
				},
				{
					Name:        "metric2",
					Description: &desc2,
					Type:        &type2,
					Labels: map[string]string{
						"env":  "staging",
						"team": "data",
					},
					Value: "121",
				},
			},
			result: `# HELP metric1 my description
# TYPE metric1 gauge
metric1{} 121
# HELP metric2 my super description
# TYPE metric2 counter
metric2{env="prod", team="sre"} 10.1
metric2{env="staging", team="data"} 121
metric3{team="backend"} 123.4567
metric4{team="front"} 123.4567
`,
		},
	}

	for _, c := range cases {
		call := store.On("GetMetrics", mock.Anything).Return(c.metrics, nil)
		result, err := service.PrometheusMetrics(context.Background())
		assert.NoError(t, err)
		assert.Equal(t, c.result, result)
		call.Unset()
	}

}
