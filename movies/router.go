package movies

import (
	"net/http"

	"github.com/gorilla/mux"
)

func contentTypeApplicationJsonMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func SetupRouting(router *mux.Router) {
	mr := router.PathPrefix("/api/movies").Subrouter()
	mr.Use(contentTypeApplicationJsonMiddleware)

	mr.HandleFunc("", getMovies).Methods("GET")
	mr.HandleFunc("/{id}", getMovie).Methods("GET")
	mr.HandleFunc("", createMovie).Methods("POST")
	mr.HandleFunc("/{id}", updateMovie).Methods("PUT")
	mr.HandleFunc("/{id}", deleteMovie).Methods("DELETE")
}
