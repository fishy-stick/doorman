package store

import (
	"encoding/json"
)

const supportedDDNSProviderDNSPod = "dnspod"

func validateDDNSSettings(enabled bool, providerType string, configStr string) error {
	config, err := decodeDDNSConfigObject(configStr)
	if err != nil {
		return err
	}

	if enabled && providerType == "" {
		return newInvalidNetworkError(NetworkErrorCodeDDNSTypeRequired, nil)
	}

	if providerType == "" {
		return nil
	}

	if !isSupportedDDNSProvider(providerType) {
		return newInvalidNetworkError(NetworkErrorCodeDDNSTypeUnsupported, map[string]string{
			"provider": providerType,
		})
	}

	switch providerType {
	case supportedDDNSProviderDNSPod:
		return validateDNSPodConfig(config)
	default:
		return newInvalidNetworkError(NetworkErrorCodeDDNSTypeUnsupported, map[string]string{
			"provider": providerType,
		})
	}
}

func decodeDDNSConfigObject(configStr string) (map[string]any, error) {
	var value any
	if err := json.Unmarshal([]byte(configStr), &value); err != nil {
		return nil, newInvalidNetworkError(NetworkErrorCodeDDNSConfigInvalidJSON, nil)
	}

	config, ok := value.(map[string]any)
	if !ok {
		return nil, newInvalidNetworkError(NetworkErrorCodeDDNSConfigNotObject, nil)
	}

	return config, nil
}

func validateDNSPodConfig(config map[string]any) error {
	for _, field := range []string{"domain", "record", "id", "token"} {
		if !hasNonEmptyStringField(config, field) {
			return newInvalidNetworkError(NetworkErrorCodeDDNSConfigFieldMissing, map[string]string{
				"field": field,
			})
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
