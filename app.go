package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

type App struct {
	router *mux.Router
}

func NewApp() *App {
	r := mux.NewRouter()

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Test-Header", "wow")
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "Let's start your journey into movies!")
	}).Methods(http.MethodGet)

	r.HandleFunc("/api/auth", authenticate)

	SetupMovieRoutes(r)

	r.Use(applicationJsonResponseContent)

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
	})

	s, e := t.SignedString([]byte("1234"))
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
