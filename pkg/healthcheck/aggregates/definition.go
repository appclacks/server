package aggregates

import (
	"encoding/json"
	"fmt"
)

type HealthcheckDefinition interface {
	String() (string, error)
	Summary() string
}

func ToHealthcheckDefinition(templateType string, definition string) (HealthcheckDefinition, error) {

	switch templateType {
	case "dns":
		var def HealthcheckDNSDefinition
		if err := json.Unmarshal([]byte(definition), &def); err != nil {
			return nil, fmt.Errorf("fail to deserialize healthcheck template definition: %w", err)
		}
		return &def, nil
	case "tcp":
		var def HealthcheckTCPDefinition
		if err := json.Unmarshal([]byte(definition), &def); err != nil {
			return nil, fmt.Errorf("fail to deserialize healthcheck template definition: %w", err)
		}
		return &def, nil
	case "tls":
		var def HealthcheckTLSDefinition
		if err := json.Unmarshal([]byte(definition), &def); err != nil {
			return nil, fmt.Errorf("fail to deserialize healthcheck template definition: %w", err)
		}
		return &def, nil
	case "http":
		var def HealthcheckHTTPDefinition
		if err := json.Unmarshal([]byte(definition), &def); err != nil {
			return nil, fmt.Errorf("fail to deserialize healthcheck template definition: %w", err)
		}
		return &def, nil
	case "command":
		var def HealthcheckCommandDefinition
		if err := json.Unmarshal([]byte(definition), &def); err != nil {
			return nil, fmt.Errorf("fail to deserialize healthcheck template definition: %w", err)
		}
		return &def, nil
	}

	return nil, fmt.Errorf("invalid template type %s", templateType)

}
