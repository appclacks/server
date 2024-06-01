package aggregates

import (
	"encoding/json"
	"fmt"
)

type HealthcheckTLSDefinition struct {
	Target          string `json:"target"`
	Port            uint   `json:"port"`
	ServerName      string `json:"server-name,omitempty"`
	ExpirationDelay string `json:"expiration-delay,omitempty"`
	Insecure        bool   `json:"insecure"`
	Key             string `json:"key,omitempty"`
	Cert            string `json:"cert,omitempty"`
	Cacert          string `json:"cacert,omitempty"`
}

func (h *HealthcheckTLSDefinition) String() (string, error) {
	result, err := json.Marshal(h)
	if err != nil {
		return "", err
	}
	return string(result), nil
}

func (h *HealthcheckTLSDefinition) Summary() string {
	return fmt.Sprintf("%s:%d", h.Target, h.Port)
}
