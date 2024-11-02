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
			writeJSON(w, http.StatusBadRequest, errorResponse("Invalid request payload"))
			return
		}

		missingFields := checkMissingFields(payload, "username", "password")
		if len(missingFields) > 0 {
			writeJSON(w, http.StatusBadRequest, errorResponse(fmt.Sprintf("Missing field(s): %s", strings.Join(missingFields, ","))))
			return
		}

		matched := false
		for _, user := range users {
			if user["name"] == payload["username"] {
				matched = bcrypt.CompareHashAndPassword([]byte(user["password"]), []byte(payload["password"])) == nil
			}
		}

		if !matched {
			writeJSON(w, http.StatusUnauthorized, errorResponse("Wrong username or password."))
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
			writeJSON(w, http.StatusInternalServerError, errorResponse("Internal server error"))
			return
		}

		writeJSON(w, http.StatusOK, map[string]string{"token": s})
	}
}

func getMovies(connection *connection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		movies, err := connection.findAllMovies()
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResponse("Internal server error."))
			return
		}
		writeJSON(w, http.StatusOK, movies)
	}
}

func deleteMovie(connection *connection) http.HandlerFunc {
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

		removed, err := connection.removeMovieByID(id)
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

func getMovieByID(connection *connection) http.HandlerFunc {
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

		movie, err := connection.findMovieByID(id)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"message": "Internal server error."})
			return
		}
		if movie == nil {
			writeJSON(w, http.StatusNotFound, errorResponse(fmt.Sprintf("Movie is not found by id - '%d'", id)))
			return
		}

		writeJSON(w, http.StatusOK, movie)
	}
}

func createMovie(connection *connection) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var movie movie
		err := json.NewDecoder(r.Body).Decode(&movie)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, errorResponse("Malformed request body."))
			return
		}

		director, err := connection.findDirectorByID(movie.Director.ID)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResponse("Internal server error."))
			return
		}

		if director == nil {
			log.Printf("[INFO] Use existing director by id - '%d' when creating movie.\n", movie.Director.ID)
		}

		id, err := connection.insertMovie(movie, director)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, errorResponse("Internal server error."))
			return
		}

		w.Header().Set("Resource-Id", strconv.FormatInt(id, 10))
		writeJSON(w, http.StatusCreated, nil)
	}
}
