package auth

import (
	"crypto/sha1"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"auth-as-a-service/app/http/httpkit"
)

type authRequest struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=64"`
}

func (r *authRequest) SetBody() error { return nil }

func checkBreachedPassword(password string) error {
	hash := fmt.Sprintf("%X", sha1.Sum([]byte(password)))
	prefix, suffix := hash[:5], hash[5:]

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get("https://api.pwnedpasswords.com/range/" + prefix)
	if err != nil {
		return nil // fail open if API is unreachable
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil
	}

	for _, line := range strings.Split(string(body), "\r\n") {
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 && parts[0] == suffix {
			return httpkit.FieldError{
				Code: http.StatusBadRequest,
				Fields: map[string][]string{
					"password": {"has been found in a data breach"},
				},
			}
		}
	}

	return nil
}

type registerResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

type loginResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type refreshResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

type logoutRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

func (r *logoutRequest) SetBody() error { return nil }
