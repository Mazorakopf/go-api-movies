package internal

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

type app struct {
	router *mux.Router
}

func New(cInfo *ConnectionInfo) *app {
	connection := newDbConnection(cInfo)

	r := mux.NewRouter()
	r.Use(applicationJsonContentTypeHeaderMiddleware)

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Test-Header", "wow")
		w.Header().Set("Content-Type", "text/plain")
		fmt.Fprintf(w, "Let's start your journey into movies!")
	}).Methods(http.MethodGet)

	r.HandleFunc("/api/auth", authenticate())

	mr := r.PathPrefix("/api/movies").Subrouter()
	mr.Use(verifyAuthorizationHeaderMiddleware)

	mr.HandleFunc("", getMovies(connection)).Methods(http.MethodGet)
	mr.HandleFunc("/{id}", getMovieById(connection)).Methods(http.MethodGet)
	mr.HandleFunc("", createMovie(connection)).Methods(http.MethodPost)
	mr.HandleFunc("/{id}", deleteMovie(connection)).Methods(http.MethodDelete)

	return &app{r}
}

func (a *app) Run(port int) {
	fmt.Printf("Starting server at port %d\n", port)
	r := handlers.LoggingHandler(os.Stdout, a.router)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), r))
}
