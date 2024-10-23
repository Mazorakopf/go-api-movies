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
	_ = json.NewDecoder(r.Body).Decode(&payload)

	encoder := json.NewEncoder(w)

	un, okU := payload["username"]
	p, okP := payload["password"]

	if !okU || !okP {
		w.WriteHeader(http.StatusBadRequest)

		misingFieslds := []string{}
		if !okU {
			misingFieslds = append(misingFieslds, "username")
		}
		if !okP {
			misingFieslds = append(misingFieslds, "password")
		}

		encoder.Encode(map[string]string{
			"message": fmt.Sprintf("Missing field(s): %s", strings.Join(misingFieslds, ",")),
		})
		return
	}

	matched := false
	for _, user := range users {
		if user["name"] == un && comparePassword(p, user["password"]) {
			matched = true
		}
	}

	if !matched {
		w.WriteHeader(http.StatusUnauthorized)
		encoder.Encode(map[string]string{"message": "Wrong username or password."})
		return
	}

	t := jwt.NewWithClaims(jwt.SigningMethodHS256,
		jwt.MapClaims{
			"iss":      "movies-service",
			"username": un,
		})

	s, e := t.SignedString([]byte("1234"))

	if e != nil {
		log.Panic("Token can not be signed.")
		w.WriteHeader(http.StatusInternalServerError)
	}

	encoder.Encode(map[string]interface{}{"token": s})
}

func comparePassword(password string, original string) bool {
	hashedPass, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return false
	}

	return bcrypt.CompareHashAndPassword(hashedPass, []byte(original)) == nil
}
