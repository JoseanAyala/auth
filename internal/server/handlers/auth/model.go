package auth

import (
	"encoding/json"
	"net/http"
	"net/mail"
)

type authRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type registerResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

type loginResponse struct {
	Token string `json:"token"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

func validEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}
