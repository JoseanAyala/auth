package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

// Response is what every handler returns on success.
type Response struct {
	Status int
	Body   any
}

// Func is the handler signature every endpoint uses.
// Return (*Response, nil) on success or (nil, err) on failure.
type Func func(r *http.Request) (*Response, error)

// Handle adapts a Func into a standard http.HandlerFunc.
// It is the single place that writes HTTP responses.
func Handle(fn Func) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		resp, err := fn(r)
		if err != nil {
			var apiErr Error
			if errors.As(err, &apiErr) {
				writeJSON(w, apiErr.Code, errorBody{Errors: apiErr.Messages})
			} else {
				writeJSON(w, http.StatusInternalServerError, errorBody{Errors: []string{"internal error"}})
			}
			return
		}
		if resp.Body != nil {
			writeJSON(w, resp.Status, resp.Body)
		} else {
			w.WriteHeader(resp.Status)
		}
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
