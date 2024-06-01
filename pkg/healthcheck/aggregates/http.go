package aggregates

import (
	"encoding/json"
	"fmt"
)

type HealthcheckHTTPDefinition struct {
	// can be an IP or a domain
	ValidStatus []uint            `json:"valid-status"`
	Target      string            `json:"target"`
	Method      string            `json:"method"`
	Port        uint              `json:"port"`
	Redirect    bool              `json:"redirect"`
	Body        string            `json:"body,omitempty"`
	Query       map[string]string `json:"query,omitempty"`
	Headers     map[string]string `json:"headers,omitempty"`
	Protocol    string            `json:"protocol"`
	Path        string            `json:"path,omitempty"`
	BodyRegexp  []string          `json:"body-regexp,omitempty"`
	Key         string            `json:"key,omitempty"`
	Cert        string            `json:"cert,omitempty"`
	Cacert      string            `json:"cacert,omitempty"`
	Host        string            `json:"host,omitempty"`
	Insecure    bool              `json:"insecure"`
	ServerName  string            `json:"server-name"`
}

func (h *HealthcheckHTTPDefinition) String() (string, error) {
	result, err := json.Marshal(h)
	if err != nil {
		return "", err
	}
	return string(result), nil
}

func (h *HealthcheckHTTPDefinition) Summary() string {
	return fmt.Sprintf("%s %s:%d", h.Method, h.Target, h.Port)
}
