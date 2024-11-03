package internal

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

var secretKey = []byte("1234")

var users = []map[string]string{
	{"name": "admin", "password": "$2a$10$miwrWXNyiF7Qiv6ir9YTueSulDDrJfjj1w2r1dLpAmaOj/TglYyKG"},
}

func authenticate() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var payload map[string]string
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid request payload")
			return
		}

		missingFields := checkMissingFields(payload, "username", "password")
		if len(missingFields) > 0 {
			respondWithError(w, http.StatusBadRequest, fmt.Sprintf("Missing field(s): %s", strings.Join(missingFields, ",")))
			return
		}

		matched := false
		for _, user := range users {
			if user["name"] == payload["username"] {
				matched = bcrypt.CompareHashAndPassword([]byte(user["password"]), []byte(payload["password"])) == nil
			}
		}

		if !matched {
			respondWithError(w, http.StatusUnauthorized, "Wrong username or password.")
			return
		}

		t := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"iss":      "movies-service",
			"username": payload["username"],
			"exp":      time.Now().Add(30 * time.Minute).Unix(),
		})

		s, err := t.SignedString(secretKey)
		if err != nil {
			log.Println("Token cannot be signed.", err)
			respondWithError(w, http.StatusInternalServerError, "Internal server error")
			return
		}

		respondWithJSON(w, http.StatusOK, map[string]string{"token": s})
	}
}

func getMovies(connection *connection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		movies, err := connection.findAllMovies()
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Internal server error.")
			return
		}
		respondWithJSON(w, http.StatusOK, movies)
	}
}

func deleteMovie(connection *connection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		missingFields := checkMissingFields(params, "id")
		if len(missingFields) > 0 {
			respondWithError(w, http.StatusBadRequest, fmt.Sprintf("Missing field(s): %s", strings.Join(missingFields, ",")))
			return
		}
		id, err := strconv.Atoi(params["id"])
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "id is not anumber")
			return
		}

		removed, err := connection.removeMovieByID(id)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Internal Server Error")
			return
		}

		if !removed {
			respondWithError(w, http.StatusNotFound, "Not Found")
			return
		}

		respondWithJSON(w, http.StatusNoContent, nil)
	}
}

func getMovieByID(connection *connection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)

		missingFields := checkMissingFields(params, "id")
		if len(missingFields) > 0 {
			respondWithError(w, http.StatusBadRequest, fmt.Sprintf("Missing field(s): %s", strings.Join(missingFields, ",")))
			return
		}

		id, err := strconv.Atoi(params["id"])
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Invalid movie id")
			return
		}

		movie, err := connection.findMovieByID(id)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Internal server error.")
			return
		}
		if movie == nil {
			respondWithError(w, http.StatusNotFound, fmt.Sprintf("Movie is not found by id - '%d'", id))
			return
		}

		respondWithJSON(w, http.StatusOK, movie)
	}
}

func createMovie(connection *connection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var movie movie
		err := json.NewDecoder(r.Body).Decode(&movie)
		if err != nil {
			respondWithError(w, http.StatusBadRequest, "Malformed request body.")
			return
		}

		director, err := connection.findDirectorByID(movie.Director.ID)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Internal server error.")
			return
		}

		if director == nil {
			log.Printf("[INFO] Use existing director by id - '%d' when creating movie.\n", movie.Director.ID)
		}

		id, err := connection.insertMovie(movie, director)
		if err != nil {
			respondWithError(w, http.StatusInternalServerError, "Internal server error.")
			return
		}

		w.Header().Set("Resource-Id", strconv.FormatInt(id, 10))
		respondWithJSON(w, http.StatusCreated, nil)
	}
}
