package store

import (
	"encoding/json"
	"fmt"
)

const supportedDDNSProviderDNSPod = "dnspod"

func validateDDNSSettings(enabled bool, providerType string, configStr string) error {
	config, err := decodeDDNSConfigObject(configStr)
	if err != nil {
		return err
	}

	if enabled && providerType == "" {
		return fmt.Errorf("%w: ddns_type is required when ddns is enabled", ErrInvalidNetwork)
	}

	if providerType == "" {
		return nil
	}

	if !isSupportedDDNSProvider(providerType) {
		return fmt.Errorf("%w: unsupported ddns_type %q; supported providers: %s", ErrInvalidNetwork, providerType, supportedDDNSProviderDNSPod)
	}

	switch providerType {
	case supportedDDNSProviderDNSPod:
		return validateDNSPodConfig(config)
	default:
		return fmt.Errorf("%w: unsupported ddns_type %q; supported providers: %s", ErrInvalidNetwork, providerType, supportedDDNSProviderDNSPod)
	}
}

func decodeDDNSConfigObject(configStr string) (map[string]any, error) {
	var value any
	if err := json.Unmarshal([]byte(configStr), &value); err != nil {
		return nil, fmt.Errorf("%w: ddns_config must be valid JSON", ErrInvalidNetwork)
	}

	config, ok := value.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%w: ddns_config must be a JSON object", ErrInvalidNetwork)
	}

	return config, nil
}

func validateDNSPodConfig(config map[string]any) error {
	for _, field := range []string{"domain", "record", "id", "token"} {
		if !hasNonEmptyStringField(config, field) {
			return fmt.Errorf("%w: ddns_config for dnspod requires non-empty string field %q", ErrInvalidNetwork, field)
		}
	}

	return nil
}

func hasNonEmptyStringField(config map[string]any, field string) bool {
	value, ok := config[field]
	if !ok {
		return false
	}

	str, ok := value.(string)
	return ok && str != ""
}

func isSupportedDDNSProvider(providerType string) bool {
	return providerType == supportedDDNSProviderDNSPod
}
