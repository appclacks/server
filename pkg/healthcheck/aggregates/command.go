package aggregates

import (
	"encoding/json"
)

type HealthcheckCommandDefinition struct {
	Command   string   `json:"command"`
	Arguments []string `json:"arguments"`
}

func (h *HealthcheckCommandDefinition) String() (string, error) {
	result, err := json.Marshal(h)
	if err != nil {
		return "", err
	}
	return string(result), nil
}

func (h *HealthcheckCommandDefinition) Summary() string {
	return h.Command
}
