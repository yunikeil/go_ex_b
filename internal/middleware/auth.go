package middleware

import (
	"context"
	"net/http"
	"strings"

	"bankapi/internal/services"
)

type contextKey string

const userIDKey contextKey = "user_id"

func Auth(auth *services.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenText := bearerToken(r.Header.Get("Authorization"))
			if tokenText == "" {
				http.Error(w, `{"error":"missing bearer token"}`, http.StatusUnauthorized)
				return
			}
			claims, err := auth.ParseToken(tokenText)
			if err != nil {
				http.Error(w, `{"error":"invalid token"}`, http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), userIDKey, claims.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func bearerToken(header string) string {
	header = strings.TrimSpace(header)
	if header == "" {
		return ""
	}
	parts := strings.Fields(header)
	if len(parts) == 1 {
		return strings.Trim(parts[0], `"`)
	}
	if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
		return strings.Trim(parts[1], `"`)
	}
	if len(parts) == 3 && strings.EqualFold(parts[0], "Bearer") && strings.EqualFold(parts[1], "Bearer") {
		return strings.Trim(parts[2], `"`)
	}
	return ""
}

func UserID(ctx context.Context) int64 {
	value, _ := ctx.Value(userIDKey).(int64)
	return value
}
