package main

import (
	"database/sql"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

func SetupMovieRoutes(router *mux.Router) *mux.Router {
	sr := router.PathPrefix("/api/movies").Subrouter()

	sr.HandleFunc("", getMovies).Methods("GET")
	sr.HandleFunc("/{id}", getMovie).Methods("GET")
	// sr.HandleFunc("", createMovie).Methods("POST")
	// sr.HandleFunc("/{id}", updateMovie).Methods("PUT")
	// sr.HandleFunc("/{id}", deleteMovie).Methods("DELETE")

	return sr
}

func getMovies(w http.ResponseWriter, r *http.Request) {
	movies, err := findAllMovies()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "Internal server error."})
		return
	}
	writeJSON(w, http.StatusOK, movies)
}

// func deleteMovie(w http.ResponseWriter, r *http.Request) {
// 	params := mux.Vars(r)
// 	for idx, item := range movies {
// 		if item.ID == params["id"] {
// 			movies = append(movies[:idx], movies[idx+1:]...)
// 			w.WriteHeader(http.StatusNoContent)
// 			return
// 		}
// 	}
// 	writeMessageNotFoundResponse(w)
// }

func getMovie(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)

	id, err := strconv.Atoi(params["id"])
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]interface{}{"message": "Invalid movie id"})
	}

	movie, err := findMovieById(id)
	if err == sql.ErrNoRows {
		writeJSON(w, http.StatusNotFound, map[string]interface{}{"message": "Not Found"})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "Internal server error."})
		return
	}

	writeJSON(w, http.StatusOK, movie)
}

// func createMovie(w http.ResponseWriter, r *http.Request) {
// 	var movie Movie
// 	_ = json.NewDecoder(r.Body).Decode(&movie)
// 	movie.ID = uuid.New().String()
// 	movies = append(movies, movie)
// 	json.NewEncoder(w).Encode(movie)
// 	w.WriteHeader(http.StatusCreated)
// }

// func updateMovie(w http.ResponseWriter, r *http.Request) {
// 	params := mux.Vars(r)

// 	for idx, item := range movies {
// 		if item.ID == params["id"] {
// 			movies = append(movies[:idx], movies[idx+1:]...)
// 			var movie Movie
// 			_ = json.NewDecoder(r.Body).Decode(&movie)
// 			movie.ID = params["id"]
// 			movies = append(movies, movie)
// 			json.NewEncoder(w).Encode(movie)
// 			return
// 		}
// 	}
// 	writeMessageNotFoundResponse(w)
// }
