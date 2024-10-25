package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

func SetupMovieRoutes(router *mux.Router, storage *Storage) *mux.Router {
	sr := router.PathPrefix("/api/movies").Subrouter()

	sr.HandleFunc("", getMovies(storage)).Methods("GET")
	sr.HandleFunc("/{id}", getMovieById(storage)).Methods("GET")
	sr.HandleFunc("", createMovie(storage)).Methods("POST")
	sr.HandleFunc("/{id}", deleteMovie(storage)).Methods("DELETE")

	return sr
}

func getMovies(storage *Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		movies, err := storage.findAllMovies()
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResponse("Internal server error."))
			return
		}
		writeJSON(w, http.StatusOK, movies)
	}
}

func deleteMovie(storage *Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		missingFields := checkMissingFields(params, "id")
		if len(missingFields) > 0 {
			writeJSON(w, http.StatusBadRequest, errorResponse(fmt.Sprintf("Missing field(s): %s", strings.Join(missingFields, ","))))
			return
		}
		id, err := strconv.Atoi(params["id"])
		if err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse("id is not anumber"))
			return
		}

		removed, err := storage.removeMovieById(id)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResponse("Internal Server Error"))
			return
		}

		if !removed {
			writeJSON(w, http.StatusNotFound, errorResponse("Not Found"))
			return
		}

		writeJSON(w, http.StatusNoContent, nil)
	}
}

func getMovieById(storage *Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)

		missingFields := checkMissingFields(params, "id")
		if len(missingFields) > 0 {
			writeJSON(w, http.StatusBadRequest, errorResponse(fmt.Sprintf("Missing field(s): %s", strings.Join(missingFields, ","))))
			return
		}

		id, err := strconv.Atoi(params["id"])
		if err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse("Invalid movie id"))
			return
		}

		movie, err := storage.findMovieById(id)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "Internal server error."})
			return
		}
		if movie == nil {
			writeJSON(w, http.StatusNotFound, errorResponse(fmt.Sprintf("Movie is not found by id - '%d'", movie.ID)))
			return
		}

		writeJSON(w, http.StatusOK, movie)
	}
}

func createMovie(storage *Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var movie Movie
		err := json.NewDecoder(r.Body).Decode(&movie)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse("Malformed request body."))
			return
		}

		director, err := storage.findDirectorById(movie.Director.ID)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResponse("Internal server error."))
			return
		}

		if director == nil {
			log.Printf("[INFO] Use existing director by id - '%d' when creating movie.\n", movie.Director.ID)
		}

		id, err := storage.insertMovie(movie, director)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResponse("Internal server error."))
			return
		}

		w.Header().Set("Resource-Id", strconv.FormatInt(id, 10))
		writeJSON(w, http.StatusCreated, nil)
	}
}
