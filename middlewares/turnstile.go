// src/github.com/lulexhostt/authapp/middlewares/turnstile.go
package middlewares

import (
	"encoding/json"
	"net/http"
	"net/url"
	"os"
)

// TurnstileResponse represents the response structure from Cloudflare's Turnstile verification.
type TurnstileResponse struct {
	Success bool     `json:"success"`
	Error   []string `json:"error-codes,omitempty"` // Change to slice to capture multiple errors
}

// TurnstileVerify handles Turnstile token verification.
func TurnstileVerify(token string) (bool, error) {
	secret := os.Getenv("TURNSTILE_SECRET_KEY")
	resp, err := http.PostForm("https://challenges.cloudflare.com/turnstile/v0/siteverify",
		url.Values{"secret": {secret}, "response": {token}})
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	var turnstileResp TurnstileResponse
	if err := json.NewDecoder(resp.Body).Decode(&turnstileResp); err != nil {
		return false, err
	}

	return turnstileResp.Success, nil
}

// TurnstilePreloadMiddleware runs Turnstile validation before page access.
func TurnstilePreloadMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check for token in form data (not URL query)
		token := r.FormValue("cf-turnstile-response")
		if token == "" {
			http.Error(w, "Turnstile token missing", http.StatusForbidden)
			return
		}

		// Verify Turnstile
		success, err := TurnstileVerify(token)
		if err != nil || !success {
			http.Error(w, "Turnstile verification failed: "+err.Error(), http.StatusForbidden)
			return
		}

		// Serve the next handler if Turnstile verification is successful
		next.ServeHTTP(w, r)
	})
}
