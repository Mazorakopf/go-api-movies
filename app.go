package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
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
