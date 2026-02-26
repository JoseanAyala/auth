package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

// ParamSetter allows a request struct to receive URL path parameters.
// Implement this on the pointer receiver of your request type.
type ParamSetter interface {
	SetParam(field, value string) error
}

// DecodeBody decodes a JSON request body into T and validates struct tags.
// Use this for endpoints that only read from the body (no path params).
func DecodeBody[T any](r *http.Request) (T, error) {
	var v T
	if err := json.NewDecoder(r.Body).Decode(&v); err != nil {
		return v, ClientErr(http.StatusBadRequest, "invalid request body")
	}
	if err := validateStruct(v); err != nil {
		return v, err
	}
	return v, nil
}

// DecodeRequest parses an HTTP request into T by optionally decoding the JSON
// body, setting chi path parameters via the ParamSetter interface, and
// validating the combined result. T must be a pointer type that implements
// ParamSetter (e.g. *MyRequest).
func DecodeRequest[T ParamSetter](r *http.Request, pathParams ...string) (T, error) {
	var req T

	// Allocate if T is a nil pointer.
	rv := reflect.ValueOf(&req).Elem()
	if rv.Kind() == reflect.Ptr && rv.IsNil() {
		rv.Set(reflect.New(rv.Type().Elem()))
	}

	if r.Body != nil && r.Body != http.NoBody {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			return req, ClientErr(http.StatusBadRequest, "invalid request body")
		}
	}

	for _, key := range pathParams {
		value := chi.URLParam(r, key)
		if value == "" {
			return req, ClientErr(http.StatusBadRequest, "missing path parameter: "+key)
		}
		if err := req.SetParam(key, value); err != nil {
			return req, ClientErr(http.StatusBadRequest, "invalid path parameter: "+key)
		}
	}

	if err := validateStruct(req); err != nil {
		return req, err
	}
	return req, nil
}

// validateStruct runs validator tags and converts failures into a ValidationError.
func validateStruct(v any) error {
	if err := validate.Struct(v); err != nil {
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			return newValidationErr(ve)
		}
		return ClientErr(http.StatusBadRequest, "invalid request")
	}
	return nil
}

func newValidationErr(ve validator.ValidationErrors) ValidationError {
	fields := make(map[string][]string, len(ve))
	for _, fe := range ve {
		field := strings.ToLower(fe.Field())
		switch fe.Tag() {
		case "required":
			fields[field] = append(fields[field], "is required")
		case "email":
			fields[field] = append(fields[field], "must be a valid email")
		case "min":
			fields[field] = append(fields[field], fmt.Sprintf("must be at least %s characters", fe.Param()))
		case "max":
			fields[field] = append(fields[field], fmt.Sprintf("must be at most %s characters", fe.Param()))
		default:
			fields[field] = append(fields[field], "is invalid")
		}
	}
	return ValidationError{Fields: fields}
}
