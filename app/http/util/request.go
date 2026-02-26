package util

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

// BodySetter allows a request struct to perform post-decode processing
// on JSON body fields (e.g. UUID parsing, date normalization).
// Implement this on the pointer receiver of your request type.
type BodySetter interface {
	SetBody() error
}

// DecodeBody decodes a JSON request body into T, calls SetBody for
// post-decode processing, and validates struct tags.
// T must be a pointer type that implements BodySetter (e.g. *MyRequest).
func DecodeBody[T BodySetter](r *http.Request) (T, error) {
	var req T

	// Allocate if T is a nil pointer.
	rv := reflect.ValueOf(&req).Elem()
	if rv.Kind() == reflect.Ptr && rv.IsNil() {
		rv.Set(reflect.New(rv.Type().Elem()))
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return req, ClientErr(http.StatusBadRequest, "invalid request body")
	}
	if err := req.SetBody(); err != nil {
		return req, ClientErr(http.StatusBadRequest, "invalid request body")
	}
	if err := validateStruct(req); err != nil {
		return req, err
	}
	return req, nil
}

// ParamSetter allows a request struct to receive URL path parameters.
// Implement this on the pointer receiver of your request type.
type ParamSetter interface {
	SetParam(field, value string) error
}

// DecodeRequest extracts chi path parameters into T and validates the result.
// T must be a pointer type that implements ParamSetter (e.g. *MyRequest).
func DecodeRequest[T ParamSetter](r *http.Request, pathParams ...string) (T, error) {
	var req T

	// Allocate if T is a nil pointer.
	rv := reflect.ValueOf(&req).Elem()
	if rv.Kind() == reflect.Ptr && rv.IsNil() {
		rv.Set(reflect.New(rv.Type().Elem()))
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

// QuerySetter allows a request struct to receive URL query parameters.
// Implement this on the pointer receiver of your request type.
type QuerySetter interface {
	SetQuery(field, value string) error
}

// DecodeQuery extracts URL query parameters into T and validates the result.
// T must be a pointer type that implements QuerySetter (e.g. *MyRequest).
func DecodeQuery[T QuerySetter](r *http.Request, queryParams ...string) (T, error) {
	var req T

	// Allocate if T is a nil pointer.
	rv := reflect.ValueOf(&req).Elem()
	if rv.Kind() == reflect.Ptr && rv.IsNil() {
		rv.Set(reflect.New(rv.Type().Elem()))
	}

	q := r.URL.Query()
	for _, key := range queryParams {
		value := q.Get(key)
		if value == "" {
			return req, ClientErr(http.StatusBadRequest, "missing query parameter: "+key)
		}
		if err := req.SetQuery(key, value); err != nil {
			return req, ClientErr(http.StatusBadRequest, "invalid query parameter: "+key)
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
