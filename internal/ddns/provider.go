package ddns

import "fmt"

type UpdateResult struct {
	Updated    bool
	OldIP      string
	NewIP      string
	Skipped    bool
	SkipReason string
}

type Provider interface {
	Update(config string, ip string) (*UpdateResult, error)
}

func GetProvider(providerType string) (Provider, error) {
	switch providerType {
	case "dnspod":
		return &DNSPodProvider{}, nil
	default:
		return nil, fmt.Errorf("unsupported provider type: %s", providerType)
	}
}
