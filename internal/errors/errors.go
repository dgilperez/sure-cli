package errors

import (
	"errors"
	"fmt"
	"net"
	"strings"
)

// Error codes for agent-friendly error handling
const (
	CodeAuthRequired  = "auth_required"
	CodeAuthInvalid   = "auth_invalid"
	CodeAuthExpired   = "auth_expired"
	CodeNotFound      = "not_found"
	CodeValidation    = "validation_failed"
	CodeNetwork       = "network_error"
	CodeTimeout       = "timeout"
	CodeRateLimit     = "rate_limited"
	CodeServerError   = "server_error"
	CodeConfigMissing = "config_missing"
	CodeConfigInvalid = "config_invalid"
	CodeExportFailed  = "export_failed"
	CodeUnknown       = "unknown_error"
)

// CLIError is a structured error with code and details for agent consumption.
type CLIError struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details,omitempty"`
	Cause   error          `json:"-"`
}

func (e *CLIError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s (%v)", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *CLIError) Unwrap() error {
	return e.Cause
}

// New creates a new CLIError
func New(code, message string) *CLIError {
	return &CLIError{Code: code, Message: message}
}

// Wrap wraps an existing error with a CLIError
func Wrap(code, message string, cause error) *CLIError {
	return &CLIError{Code: code, Message: message, Cause: cause}
}

// WithDetails adds details to the error
func (e *CLIError) WithDetails(details map[string]any) *CLIError {
	e.Details = details
	return e
}

// ClassifyHTTPError maps HTTP status codes to error codes
func ClassifyHTTPError(status int, body string) *CLIError {
	switch {
	case status == 401:
		return New(CodeAuthRequired, "Authentication required")
	case status == 403:
		return New(CodeAuthInvalid, "Access denied")
	case status == 404:
		return New(CodeNotFound, "Resource not found")
	case status == 422:
		return New(CodeValidation, "Validation failed").WithDetails(map[string]any{"body": truncate(body, 500)})
	case status == 429:
		return New(CodeRateLimit, "Rate limit exceeded")
	case status >= 500:
		return New(CodeServerError, fmt.Sprintf("Server error (HTTP %d)", status))
	default:
		return New(CodeUnknown, fmt.Sprintf("HTTP %d", status))
	}
}

// ClassifyNetworkError classifies network errors
func ClassifyNetworkError(err error) *CLIError {
	if err == nil {
		return nil
	}

	// Check for timeout
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return Wrap(CodeTimeout, "Request timed out", err)
	}

	// Check for connection errors
	if strings.Contains(err.Error(), "connection refused") {
		return Wrap(CodeNetwork, "Connection refused - is the server running?", err)
	}
	if strings.Contains(err.Error(), "no such host") {
		return Wrap(CodeNetwork, "Server not found - check api_url config", err)
	}

	return Wrap(CodeNetwork, "Network error", err)
}

// IsRetryable returns true if the error is likely temporary and retryable
func IsRetryable(err error) bool {
	var cliErr *CLIError
	if errors.As(err, &cliErr) {
		switch cliErr.Code {
		case CodeNetwork, CodeTimeout, CodeRateLimit, CodeServerError:
			return true
		}
	}
	return false
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
