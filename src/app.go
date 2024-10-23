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

type App struct {
	router *mux.Router
}

func NewApp() *App {
	r := mux.NewRouter()
	r.Use(applicationJsonResponseContent)

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Test-Header", "wow")
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "Let's start your journey into movies!")
	}).Methods(http.MethodGet)

	r.HandleFunc("/api/auth", authenticate)

	mr := SetupMovieRoutes(r)
	mr.Use(vaerifyJwtMiddleware)

	return &App{r}
}

func (a *App) Run(port int) {
	fmt.Printf("Starting server at port %d\n", port)
	r := handlers.LoggingHandler(os.Stdout, a.router)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), r))
}

func applicationJsonResponseContent(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func vaerifyJwtMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenString, errExtractionToken := extractToken(r.Header.Get("Authorization"))
		if errExtractionToken != nil {
			log.Println(errExtractionToken)
			writeJSON(w, http.StatusBadRequest, map[string]string{"message": "Authorization header is invalid."})
			return
		}

		token, errVerifyingToken := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
			return SECRET_KEY, nil
		})

		if errVerifyingToken != nil || !token.Valid {
			writeJSON(w, http.StatusUnauthorized, map[string]string{"message": "Jwt token is invalid."})
			return
		}

		next.ServeHTTP(w, r)
	})
}

func authenticate(w http.ResponseWriter, r *http.Request) {
	var payload map[string]string
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"message": "Invalid request payload"})
		return
	}

	missingFields := checkMissingFields(payload, "username", "password")
	if len(missingFields) > 0 {
		writeJSON(w, http.StatusBadRequest, map[string]string{
			"message": fmt.Sprintf("Missing field(s): %s", strings.Join(missingFields, ",")),
		})
		return
	}

	matched := false
	for _, user := range users {
		if user["name"] == payload["username"] && comparePassword(payload["password"], user["password"]) {
			matched = true
		}
	}

	if !matched {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"message": "Wrong username or password."})
		return
	}

	t := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"iss":      "movies-service",
		"username": payload["username"],
		"exp":      time.Now().Add(5 * time.Minute).Unix(),
	})

	s, e := t.SignedString(SECRET_KEY)
	if e != nil {
		log.Println("Token cannot be signed:", e)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "Internal server error"})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{"token": s})
}

func checkMissingFields(payload map[string]string, fields ...string) []string {
	var missing []string
	for _, field := range fields {
		if _, ok := payload[field]; !ok {
			missing = append(missing, field)
		}
	}
	return missing
}

func comparePassword(password string, hashed string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(password)) == nil
}

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
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
