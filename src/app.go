package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

var SECRET_KEY = []byte("1234")

var users = []map[string]string{
	{"name": "admin", "password": "$2a$10$miwrWXNyiF7Qiv6ir9YTueSulDDrJfjj1w2r1dLpAmaOj/TglYyKG"},
}

type App struct {
	router *mux.Router
}

func NewApp(storage *Storage) *App {
	r := mux.NewRouter()
	r.Use(applicationJsonContentTypeHeaderMiddleware)

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Test-Header", "wow")
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "Let's start your journey into movies!")
	}).Methods(http.MethodGet)

	r.HandleFunc("/api/auth", authenticate)

	mr := SetupMovieRoutes(r, storage)
	mr.Use(verifyAuthorizationHeaderMiddleware)

	return &App{r}
}

func (a *App) Run(port int) {
	fmt.Printf("Starting server at port %d\n", port)
	r := handlers.LoggingHandler(os.Stdout, a.router)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), r))
}

func applicationJsonContentTypeHeaderMiddleware(next http.Handler) http.Handler {
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
			return SECRET_KEY, nil
		})

		if err != nil || !token.Valid {
			log.Println("[DEBUG] Jwt token cannot be verified.", err)
			writeJSON(w, http.StatusUnauthorized, errorResponse("Jwt token is invalid."))
			return
		}

		next.ServeHTTP(w, r)
	})
}

func authenticate(w http.ResponseWriter, r *http.Request) {
	var payload map[string]string
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse("Invalid request payload"))
		return
	}

	missingFields := checkMissingFields(payload, "username", "password")
	if len(missingFields) > 0 {
		writeJSON(w, http.StatusBadRequest, errorResponse(fmt.Sprintf("Missing field(s): %s", strings.Join(missingFields, ","))))
		return
	}

	matched := false
	for _, user := range users {
		if user["name"] == payload["username"] {
			matched = bcrypt.CompareHashAndPassword([]byte(payload["password"]), []byte(user["password"])) == nil
		}
	}

	if !matched {
		writeJSON(w, http.StatusUnauthorized, errorResponse("Wrong username or password."))
		return
	}

	t := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iss":      "movies-service",
		"username": payload["username"],
		"exp":      time.Now().Add(30 * time.Minute).Unix(),
	})

	s, err := t.SignedString(SECRET_KEY)
	if err != nil {
		log.Println("Token cannot be signed.", err)
		writeJSON(w, http.StatusInternalServerError, errorResponse("Internal server error"))
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"token": s})
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
