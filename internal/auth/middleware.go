package auth

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/a2sh3r/gophkeeper/internal/logger"
	"github.com/urfave/negroni"
	"go.uber.org/zap"
)

// AuthMiddleware creates authentication middleware
func AuthMiddleware(jwtManager *JWTManager) negroni.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
			return
		}

		token := parts[1]
		claims, err := jwtManager.ValidateToken(token)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		r.Header.Set("X-User-ID", claims.UserID.String())
		r.Header.Set("X-Username", claims.Username)

		next(w, r)
	}
}

// ErrorResponse represents error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// writeError writes error to response
func writeError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(ErrorResponse{Error: message}); err != nil {
		logger.Log.Error("Failed to encode data", zap.Error(err))
	}
}
