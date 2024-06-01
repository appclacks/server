package aggregates

import (
	"encoding/json"
)

type HealthcheckDNSDefinition struct {
	Domain      string   `json:"domain,omitempty"`
	ExpectedIPs []string `json:"expected-ips,omitempty"`
}

func (h *HealthcheckDNSDefinition) String() (string, error) {
	result, err := json.Marshal(h)
	if err != nil {
		return "", err
	}
	return string(result), nil
}

func (h *HealthcheckDNSDefinition) Summary() string {
	return h.Domain
}
