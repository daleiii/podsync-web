package middleware

import (
	"crypto/subtle"
	"net/http"

	log "github.com/sirupsen/logrus"
)

// BasicAuth middleware to protect endpoints with HTTP basic authentication
func BasicAuth(username, password string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// If username or password is empty, skip authentication
			if username == "" || password == "" {
				next.ServeHTTP(w, r)
				return
			}

			// Get credentials from request
			user, pass, ok := r.BasicAuth()

			// Use constant time comparison to prevent timing attacks
			validUsername := subtle.ConstantTimeCompare([]byte(user), []byte(username)) == 1
			validPassword := subtle.ConstantTimeCompare([]byte(pass), []byte(password)) == 1

			if !ok || !validUsername || !validPassword {
				log.Debugf("unauthorized access attempt from %s", r.RemoteAddr)
				w.Header().Set("WWW-Authenticate", `Basic realm="Podsync"`)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
