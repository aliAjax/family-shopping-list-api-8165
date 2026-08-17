package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"family-shopping-list-api/pkg/response"
	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const (
	userIDKey   contextKey = "user_id"
	usernameKey contextKey = "username"
)

func UserIDFromContext(ctx context.Context) (uint64, bool) {
	value, ok := ctx.Value(userIDKey).(uint64)
	return value, ok
}

func UsernameFromContext(ctx context.Context) (string, bool) {
	value, ok := ctx.Value(usernameKey).(string)
	return value, ok
}

func Auth(secret string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if isPublicPath(r.URL.Path) {
			next.ServeHTTP(w, r)
			return
		}

		header := r.Header.Get("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			response.Error(w, http.StatusUnauthorized, "unauthorized", "缺少 Bearer Token")
			return
		}
		tokenString := strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
		claims := jwt.MapClaims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("无效的签名方法")
			}
			return []byte(secret), nil
		})
		if err != nil || !token.Valid {
			response.Error(w, http.StatusUnauthorized, "unauthorized", "Token 无效或已过期")
			return
		}

		userID, ok := claims["user_id"].(float64)
		if !ok {
			response.Error(w, http.StatusUnauthorized, "unauthorized", "Token 缺少用户信息")
			return
		}
		username, _ := claims["username"].(string)
		ctx := context.WithValue(r.Context(), userIDKey, uint64(userID))
		ctx = context.WithValue(ctx, usernameKey, username)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func isPublicPath(path string) bool {
	return path == "/healthz" || strings.HasPrefix(path, "/api/v1/auth/")
}
