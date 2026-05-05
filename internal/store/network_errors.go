package store

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

type NetworkErrorCode string

const (
	NetworkErrorCodeNetworkRequired        NetworkErrorCode = "network_required"
	NetworkErrorCodeNameRequired           NetworkErrorCode = "name_required"
	NetworkErrorCodeTokenRequired          NetworkErrorCode = "token_required"
	NetworkErrorCodeDDNSTypeRequired       NetworkErrorCode = "ddns_type_required"
	NetworkErrorCodeDDNSTypeUnsupported    NetworkErrorCode = "ddns_type_unsupported"
	NetworkErrorCodeDDNSConfigInvalidJSON  NetworkErrorCode = "ddns_config_invalid_json"
	NetworkErrorCodeDDNSConfigNotObject    NetworkErrorCode = "ddns_config_not_object"
	NetworkErrorCodeDDNSConfigFieldMissing NetworkErrorCode = "ddns_config_required_field"
	NetworkErrorCodeNameConflict           NetworkErrorCode = "network_name_conflict"
	NetworkErrorCodeTokenConflict          NetworkErrorCode = "network_token_conflict"
)

type NetworkError struct {
	kind error
	Code NetworkErrorCode
	Meta map[string]string
}

func (e *NetworkError) Error() string {
	if e == nil {
		return ""
	}

	if len(e.Meta) == 0 {
		return fmt.Sprintf("%v: %s", e.kind, e.Code)
	}

	pairs := make([]string, 0, len(e.Meta))
	for key, value := range e.Meta {
		pairs = append(pairs, key+"="+value)
	}
	sort.Strings(pairs)

	return fmt.Sprintf("%v: %s (%s)", e.kind, e.Code, strings.Join(pairs, ", "))
}

func (e *NetworkError) Unwrap() error {
	return e.kind
}

func AsNetworkError(err error) (*NetworkError, bool) {
	var target *NetworkError
	if !errors.As(err, &target) {
		return nil, false
	}

	return target, true
}

func newInvalidNetworkError(code NetworkErrorCode, meta map[string]string) error {
	return &NetworkError{
		kind: ErrInvalidNetwork,
		Code: code,
		Meta: cloneMeta(meta),
	}
}

func newNetworkConflictError(kind error, code NetworkErrorCode, meta map[string]string) error {
	return &NetworkError{
		kind: kind,
		Code: code,
		Meta: cloneMeta(meta),
	}
}

func cloneMeta(meta map[string]string) map[string]string {
	if len(meta) == 0 {
		return nil
	}

	cloned := make(map[string]string, len(meta))
	for key, value := range meta {
		cloned[key] = value
	}

	return cloned
}
