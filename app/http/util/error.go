package util

import (
	"fmt"
	"strings"
)

// errorBody is the JSON shape written for all error responses.
type errorBody struct {
	Message string              `json:"message"`
	Errors  map[string][]string `json:"errors,omitempty"`
}

// Error is a user-facing HTTP error with a status code and message.
type Error struct {
	Code    int
	Message string
}

func (e Error) Error() string { return e.Message }

// ClientErr constructs a user-facing Error with the given HTTP status and message.
func ClientErr(code int, msg string) error {
	return Error{Code: code, Message: msg}
}

// ValidationError carries per-field validation failures.
// Handle renders it as a 400 with the field map in the "errors" key.
type ValidationError struct {
	Fields map[string][]string
}

func (e ValidationError) Error() string {
	msgs := make([]string, 0, len(e.Fields))
	for field, errs := range e.Fields {
		msgs = append(msgs, fmt.Sprintf("%s: %s", field, strings.Join(errs, ", ")))
	}
	return strings.Join(msgs, "; ")
}
