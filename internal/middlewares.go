package internal

import (
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

func applicationJSONContentTypeHeaderMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func verifyAuthorizationHeaderMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString, err := extractToken(r.Header.Get("Authorization"))
		if err != nil {
			log.Println("[DEBUG] Failed to extract token from Authorization header.", err)
			writeJSON(w, http.StatusBadRequest, errorResponse(err.Error()))
			return
		}

		token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
			return secretKey, nil
		})

		if err != nil || !token.Valid {
			log.Println("[DEBUG] Jwt token cannot be verified.", err)
			writeJSON(w, http.StatusUnauthorized, errorResponse("Jwt token is invalid."))
			return
		}

		next.ServeHTTP(w, r)
	})
}

func extractToken(authHeader string) (string, error) {
	const bearerPrefix = "Bearer "

	if authHeader == "" {
		return "", errors.New("authorization header is empty")
	}

	if !strings.HasPrefix(authHeader, bearerPrefix) {
		return "", errors.New("authorization header does not contain Bearer prefix")
	}

	token := strings.TrimPrefix(authHeader, bearerPrefix)
	token = strings.TrimSpace(token)

	if token == "" {
		return "", errors.New("authorization header does not contain token")
	}

	return token, nil
}
