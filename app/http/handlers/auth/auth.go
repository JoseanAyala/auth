package auth

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"

	"auth-as-a-service/app/http/httpkit"
	"auth-as-a-service/sdk/crypto"
	"auth-as-a-service/sdk/token"
)

func (h *Handler) register(r *http.Request) (*httpkit.Response, error) {
	req, err := httpkit.DecodeBody[*authRequest](r)
	if err != nil {
		return nil, err
	}

	if err := checkBreachedPassword(req.Password); err != nil {
		return nil, err
	}

	hashPW, err := crypto.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	user, err := h.users.Create(r.Context(), req.Email, hashPW)
	if err != nil {
		// TODO: Better SQL default error handling
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, httpkit.FieldError{
				Code: http.StatusConflict,
				Fields: map[string][]string{
					"email": {"email already exists"},
				},
			}
		}
		return nil, err
	}

	return &httpkit.Response{
		Status: http.StatusCreated,
		Body:   registerResponse{ID: user.ID, Email: user.Email},
	}, nil
}

func (h *Handler) login(r *http.Request) (*httpkit.Response, error) {
	req, err := httpkit.DecodeBody[*authRequest](r)
	if err != nil {
		return nil, err
	}

	user, err := h.users.GetByEmail(r.Context(), req.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, httpkit.ClientErr(http.StatusUnauthorized, "Invalid credentials")
		}
		return nil, err
	}

	doesMatch, err := crypto.VerifyPassword(req.Password, user.PasswordHash)
	if err != nil {
		return nil, err
	}

	if !doesMatch {
		return nil, httpkit.ClientErr(http.StatusUnauthorized, "Invalid credentials")
	}

	accessTok, err := token.Generate(user.ID)
	if err != nil {
		return nil, err
	}

	refreshTok, err := token.GenerateRefresh(user.ID)
	if err != nil {
		return nil, err
	}

	return &httpkit.Response{
		Status: http.StatusOK,
		Body:   loginResponse{AccessToken: accessTok, RefreshToken: refreshTok},
	}, nil
}

func (h *Handler) logout(r *http.Request) (*httpkit.Response, error) {
	accessToken := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	if err := token.Revoke(r.Context(), accessToken, h.redis); err != nil {
		return nil, err
	}

	req, err := httpkit.DecodeBody[*logoutRequest](r)
	if err != nil {
		return nil, err
	}

	if err := token.Revoke(r.Context(), req.RefreshToken, h.redis); err != nil {
		return nil, err
	}

	return &httpkit.Response{Status: http.StatusNoContent}, nil
}

func (h *Handler) refresh(r *http.Request) (*httpkit.Response, error) {
	tokenString := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	if tokenString == "" {
		return nil, httpkit.ClientErr(http.StatusUnauthorized, "missing token")
	}

	userID, err := token.ValidateRefresh(r.Context(), tokenString, h.redis)
	if err != nil {
		return nil, httpkit.ClientErr(http.StatusUnauthorized, "invalid or expired refresh token")
	}

	accessTok, err := token.Generate(userID)
	if err != nil {
		return nil, err
	}

	refreshTok, err := token.GenerateRefresh(userID)
	if err != nil {
		return nil, err
	}

	if err := token.Revoke(r.Context(), tokenString, h.redis); err != nil {
		return nil, err
	}

	return &httpkit.Response{
		Status: http.StatusOK,
		Body:   refreshResponse{AccessToken: accessTok, RefreshToken: refreshTok},
	}, nil
}
