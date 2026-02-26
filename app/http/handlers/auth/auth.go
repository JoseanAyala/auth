package auth

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"

	"auth-as-a-service/app/hasher"
	"auth-as-a-service/app/http/util"
	"auth-as-a-service/app/token"
)

func (h *Handler) register(r *http.Request) (*util.Response, error) {
	req, err := util.DecodeBody[*authRequest](r)
	if err != nil {
		return nil, err
	}

	result := make(chan hasher.HashResult, 1)
	if err := h.hasher.Submit(hasher.HashJob{Password: req.Password, Result: result}); err != nil {
		if errors.Is(err, hasher.ErrQueueFull) {
			return nil, util.ClientErr(http.StatusServiceUnavailable, "service busy, try again later")
		}
		return nil, err
	}

	hr := <-result
	if hr.Err != nil {
		return nil, hr.Err
	}

	user, err := h.users.Create(r.Context(), req.Email, hr.Hash)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, util.ClientErr(http.StatusConflict, "email already exists")
		}
		return nil, err
	}

	return &util.Response{
		Status: http.StatusCreated,
		Body:   registerResponse{ID: user.ID, Email: user.Email},
	}, nil
}

func (h *Handler) login(r *http.Request) (*util.Response, error) {
	req, err := util.DecodeBody[*authRequest](r)
	if err != nil {
		return nil, err
	}

	user, err := h.users.GetByEmail(r.Context(), req.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, util.ClientErr(http.StatusUnauthorized, "invalid credentials")
		}
		return nil, err
	}

	result := make(chan hasher.VerifyResult, 1)
	if err := h.hasher.Submit(hasher.VerifyJob{
		Password:   req.Password,
		StoredHash: user.PasswordHash,
		Result:     result,
	}); err != nil {
		if errors.Is(err, hasher.ErrQueueFull) {
			return nil, util.ClientErr(http.StatusServiceUnavailable, "service busy, try again later")
		}
		return nil, err
	}

	vr := <-result
	if vr.Err != nil {
		return nil, vr.Err
	}
	if !vr.Match {
		return nil, util.ClientErr(http.StatusUnauthorized, "invalid credentials")
	}

	accessTok, err := token.Generate(user.ID)
	if err != nil {
		return nil, err
	}

	refreshTok, err := token.GenerateRefresh(user.ID)
	if err != nil {
		return nil, err
	}

	return &util.Response{
		Status: http.StatusOK,
		Body:   loginResponse{AccessToken: accessTok, RefreshToken: refreshTok},
	}, nil
}

func (h *Handler) logout(r *http.Request) (*util.Response, error) {
	accessToken := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	if err := token.Revoke(r.Context(), accessToken, h.redis); err != nil {
		return nil, err
	}

	var body logoutRequest
	json.NewDecoder(r.Body).Decode(&body)
	if body.RefreshToken != "" {
		token.Revoke(r.Context(), body.RefreshToken, h.redis)
	}

	return &util.Response{Status: http.StatusNoContent}, nil
}

func (h *Handler) refresh(r *http.Request) (*util.Response, error) {
	tokenString := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
	if tokenString == "" {
		return nil, util.ClientErr(http.StatusUnauthorized, "missing token")
	}

	userID, err := token.ValidateRefresh(r.Context(), tokenString, h.redis)
	if err != nil {
		return nil, util.ClientErr(http.StatusUnauthorized, "invalid or expired refresh token")
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

	return &util.Response{
		Status: http.StatusOK,
		Body:   refreshResponse{AccessToken: accessTok, RefreshToken: refreshTok},
	}, nil
}
