package aggregates

import (
	"encoding/json"
	"fmt"
)

type HealthcheckTCPDefinition struct {
	// can be an IP or a domain
	Target     string `json:"target"`
	Port       uint   `json:"port"`
	ShouldFail bool   `json:"should-fail"`
}

func (h *HealthcheckTCPDefinition) String() (string, error) {
	result, err := json.Marshal(h)
	if err != nil {
		return "", err
	}
	return string(result), nil
}

func (h *HealthcheckTCPDefinition) Summary() string {
	return fmt.Sprintf("%s:%d", h.Target, h.Port)
}
