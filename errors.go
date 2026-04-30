package sws

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// APIError is the base type for every error the SDK returns. Specific
// subtypes (AuthenticationError, NotFoundError, ...) embed *APIError so
// callers can either type-assert with errors.As for the subtype or match
// the sentinels below with errors.Is.
//
// Renamed from Error to avoid colliding with the Error() method name
// when embedded in subtypes.
type APIError struct {
	StatusCode int
	Message    string
	Body       json.RawMessage
}

func (e *APIError) Error() string {
	return fmt.Sprintf("sws: HTTP %d: %s", e.StatusCode, e.Message)
}

// Sentinel errors for use with errors.Is. Each subtype's Is method
// matches its sentinel.
var (
	ErrAuthentication = errors.New("sws: authentication failed")
	ErrNotFound       = errors.New("sws: resource not found")
	ErrValidation     = errors.New("sws: validation failed")
	ErrQuota          = errors.New("sws: quota exceeded")
)

// AuthenticationError indicates a missing, invalid, or expired API key (401/403).
type AuthenticationError struct{ *APIError }

func (*AuthenticationError) Is(target error) bool { return target == ErrAuthentication }

// NotFoundError indicates the requested resource does not exist (404).
type NotFoundError struct{ *APIError }

func (*NotFoundError) Is(target error) bool { return target == ErrNotFound }

// ValidationError indicates the server rejected the request payload (400/422).
type ValidationError struct{ *APIError }

func (*ValidationError) Is(target error) bool { return target == ErrValidation }

// QuotaExceededError indicates a 403 whose body mentions a quota limit.
// Distinct from AuthenticationError because callers usually want to retry
// with a smaller request, not re-authenticate.
type QuotaExceededError struct{ *APIError }

func (*QuotaExceededError) Is(target error) bool { return target == ErrQuota }

// parseError maps an HTTP error response to one of the typed errors above.
func parseError(status int, body []byte) error {
	base := &APIError{StatusCode: status, Body: body, Message: extractMessage(body, status)}

	switch status {
	case 401:
		return &AuthenticationError{base}
	case 403:
		if strings.Contains(strings.ToLower(base.Message), "quota") {
			return &QuotaExceededError{base}
		}
		return &AuthenticationError{base}
	case 404:
		return &NotFoundError{base}
	case 400, 422:
		return &ValidationError{base}
	default:
		return base
	}
}

// extractMessage pulls a human-readable message out of common error body
// shapes ({"detail": ...}, {"error": ...}, {"message": ...}). Falls back
// to the raw body or the HTTP status text.
func extractMessage(body []byte, status int) string {
	if len(body) == 0 {
		return fmt.Sprintf("HTTP %d", status)
	}
	var m map[string]any
	if err := json.Unmarshal(body, &m); err == nil {
		for _, k := range []string{"detail", "error", "message"} {
			if v, ok := m[k].(string); ok && v != "" {
				return v
			}
		}
	}
	if len(body) > 200 {
		return string(body[:200]) + "..."
	}
	return string(body)
}
