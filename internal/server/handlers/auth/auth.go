package auth

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"

	"auth-as-a-service/internal/hasher"
	"auth-as-a-service/internal/token"
)

func (h *Handler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var req authRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid request body"})
		return
	}

	if !validEmail(req.Email) {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid email format"})
		return
	}
	if len(req.Password) < 8 {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "password must be at least 8 characters"})
		return
	}

	// Hash password via the Hasher worker pool.
	result := make(chan hasher.HashResult, 1)
	if err := h.hasher.Submit(hasher.HashJob{Password: req.Password, Result: result}); err != nil {
		if errors.Is(err, hasher.ErrQueueFull) {
			writeJSON(w, http.StatusServiceUnavailable, errorResponse{Error: "service busy, try again later"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal error"})
		return
	}

	hr := <-result
	if hr.Err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal error"})
		return
	}

	user, err := h.users.Create(r.Context(), req.Email, hr.Hash)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			writeJSON(w, http.StatusConflict, errorResponse{Error: "email already exists"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal error"})
		return
	}

	writeJSON(w, http.StatusCreated, registerResponse{ID: user.ID, Email: user.Email})
}

func (h *Handler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req authRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid request body"})
		return
	}

	if !validEmail(req.Email) {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid email format"})
		return
	}
	if len(req.Password) < 8 {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "password must be at least 8 characters"})
		return
	}

	user, err := h.users.GetByEmail(r.Context(), req.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "invalid credentials"})
			return
		}

		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal error"})
		return
	}

	// Verify password via the Hasher worker pool.
	result := make(chan hasher.VerifyResult, 1)
	if err := h.hasher.Submit(hasher.VerifyJob{
		Password:   req.Password,
		StoredHash: user.PasswordHash,
		Result:     result,
	}); err != nil {
		if errors.Is(err, hasher.ErrQueueFull) {
			writeJSON(w, http.StatusServiceUnavailable, errorResponse{Error: "service busy, try again later"})
			return
		}
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal error"})
		return
	}

	vr := <-result
	if vr.Err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal error"})
		return
	}
	if !vr.Match {
		writeJSON(w, http.StatusUnauthorized, errorResponse{Error: "invalid credentials"})
		return
	}

	tok, err := token.Generate(user.ID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal error"})
		return
	}

	writeJSON(w, http.StatusOK, loginResponse{Token: tok})
}

func (h *Handler) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	tokenString := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	if err := token.Revoke(r.Context(), tokenString, h.redis); err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: "internal error"})
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
