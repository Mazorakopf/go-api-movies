package main

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func SetupMovieRoutes(router *mux.Router) {
	sr := router.PathPrefix("/api/movies").Subrouter()

	sr.HandleFunc("", getMovies).Methods("GET")
	sr.HandleFunc("/{id}", getMovie).Methods("GET")
	sr.HandleFunc("", createMovie).Methods("POST")
	sr.HandleFunc("/{id}", updateMovie).Methods("PUT")
	sr.HandleFunc("/{id}", deleteMovie).Methods("DELETE")
}

func getMovies(w http.ResponseWriter, r *http.Request) {
	json.NewEncoder(w).Encode(movies)
	w.WriteHeader(http.StatusOK)
}

func deleteMovie(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	for idx, item := range movies {
		if item.ID == params["id"] {
			movies = append(movies[:idx], movies[idx+1:]...)
			w.WriteHeader(http.StatusNoContent)
			return
		}
	}
	writeMessageNotFoundResponse(w)
}

func getMovie(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	for _, item := range movies {
		if item.ID == params["id"] {
			json.NewEncoder(w).Encode(item)
			return
		}
	}
	writeMessageNotFoundResponse(w)
}

func createMovie(w http.ResponseWriter, r *http.Request) {
	var movie Movie
	_ = json.NewDecoder(r.Body).Decode(&movie)
	movie.ID = uuid.New().String()
	movies = append(movies, movie)
	json.NewEncoder(w).Encode(movie)
	w.WriteHeader(http.StatusCreated)
}

func updateMovie(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	for idx, item := range movies {
		if item.ID == params["id"] {
			movies = append(movies[:idx], movies[idx+1:]...)
			var movie Movie
			_ = json.NewDecoder(r.Body).Decode(&movie)
			movie.ID = params["id"]
			movies = append(movies, movie)
			json.NewEncoder(w).Encode(movie)
			return
		}
	}
	writeMessageNotFoundResponse(w)
}

func writeMessageNotFoundResponse(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNotFound)
	json.NewEncoder(w).Encode(map[string]interface{}{"message": "Not Found"})
}
