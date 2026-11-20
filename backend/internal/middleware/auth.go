package middleware

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const userIDKey contextKey = "userID"
const userRoleKey contextKey = "role"

type Claims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

func UserIDFromContext(ctx context.Context) (int, bool) {
	id, ok := ctx.Value(userIDKey).(int)
	return id, ok && id > 0
}

func RoleFromContext(ctx context.Context) (string, bool) {
	role, ok := ctx.Value(userRoleKey).(string)
	return role, ok && role != ""
}

func isPublicRoute(method, path string) bool {
	if method == http.MethodPost {
		if path == "/auth/register" || path == "/auth/login" {
			return true
		}
		return false
	}

	if method != http.MethodGet {
		return false
	}

	// Публичные ручки
	if path == "/equipment" ||
		path == "/equipment/categories" ||
		path == "/equipment/sets" ||
		path == "/equipment/availability" ||
		path == "/instructors" ||
		path == "/services" {
		return true
	}

	parts := strings.Split(strings.Trim(path, "/"), "/")

	// /equipment/{id}
	// /instructors/{id}
	// /services/{id}
	if len(parts) == 2 {
		if parts[0] == "equipment" ||
			parts[0] == "instructors" ||
			parts[0] == "services" {

			id, err := strconv.Atoi(parts[1])
			return err == nil && id > 0
		}
	}

	// /equipment/sets/{id}
	if len(parts) == 3 &&
		parts[0] == "equipment" &&
		parts[1] == "sets" {

		id, err := strconv.Atoi(parts[2])
		return err == nil && id > 0
	}

	// GET /instructors/{id}/schedule
	// GET /instructors/{id}/available-slots
	// GET /instructors/{id}/services
	if len(parts) == 3 &&
		parts[0] == "instructors" &&
		(parts[2] == "schedule" ||
			parts[2] == "available-slots" ||
			parts[2] == "services") {
		return true
	}

	return false
}

func AuthMiddleware(jwtSecret []byte, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if isPublicRoute(r.Method, r.URL.Path) {
			next.ServeHTTP(w, r)
			return

		}
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		before, after, found := strings.Cut(authHeader, " ")
		after = strings.TrimSpace(after)
		if before != "Bearer" || after == "" || !found {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		claims := &Claims{}

		token, err := jwt.ParseWithClaims(after, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		},
			jwt.WithValidMethods([]string{"HS256"}),
		)
		if err != nil || !token.Valid {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		id, err := strconv.Atoi(claims.Subject)
		if err != nil || id <= 0 {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		if claims.Role != "client" &&
			claims.Role != "instructor" &&
			claims.Role != "admin" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), userIDKey, id)
		ctx = context.WithValue(ctx, userRoleKey, claims.Role)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)

	})
}
