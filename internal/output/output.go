package output

import (
	"encoding/json"
	"fmt"
	"os"
)

type Envelope struct {
	Data  any    `json:"data,omitempty"`
	Error *Error `json:"error,omitempty"`
	Meta  *Meta  `json:"meta,omitempty"`
}

type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

func PrintJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func Print(format string, env Envelope) error {
	switch format {
	case "table":
		if ok := PrintTable(env); ok {
			return nil
		}
		// fallback
		return PrintJSON(env)
	default:
		return PrintJSON(env)
	}
}

func Fail(code, message string, details any) {
	_ = PrintJSON(Envelope{Error: &Error{Code: code, Message: message, Details: details}})
	fmt.Fprintln(os.Stderr, message)
	os.Exit(1)
}

// FailErr is a convenience wrapper for structured errors
func FailErr(err error) {
	// Try to extract structured error info
	type codeError interface {
		Error() string
	}
	type detailError interface {
		Details() map[string]any
	}

	code := "error"
	message := err.Error()
	var details any

	// Check for our CLIError type (duck typing to avoid import cycle)
	if ce, ok := err.(interface{ Code() string }); ok {
		code = ce.Code()
	}
	if de, ok := err.(detailError); ok {
		details = de.Details()
	}

	Fail(code, message, details)
}
