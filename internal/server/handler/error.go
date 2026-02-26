package handler

import "strings"

type errorBody struct {
	Errors []string `json:"errors"`
}

// Error is a user-facing HTTP error with a status code and messages.
// Return one from a handler to control exactly what the client sees.
// Any other error type results in a generic 500.
type Error struct {
	Code     int
	Messages []string
}

func (e Error) Error() string { return strings.Join(e.Messages, "; ") }

// ClientErr constructs a user-facing Error with the given HTTP status and message(s).
func ClientErr(code int, msgs ...string) error {
	return Error{Code: code, Messages: msgs}
}
