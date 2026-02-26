package handler

import (
	"encoding/json"
	"errors"
	"log"
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
		// Handler error
		if err != nil {
			var ve ValidationError
			var ce Error
			switch {
			case errors.As(err, &ve):
				writeJSON(w, http.StatusBadRequest, errorBody{
					Message: "validation failed",
					Errors:  ve.Fields,
				})
			case errors.As(err, &ce):
				writeJSON(w, ce.Code, errorBody{Message: ce.Message})
			default:
				writeJSON(w, http.StatusInternalServerError, errorBody{Message: "internal error"})
			}
			return
		}

		// Handle response
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

	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("failed to encode response: %v", err)
	}
}
