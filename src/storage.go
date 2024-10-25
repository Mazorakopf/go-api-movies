package main

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

// [START] Database Entities
type Movie struct {
	ID       int      `json:"id"`
	Isbn     string   `json:"isbn"`
	Title    string   `json:"title"`
	Director Director `json:"director"`
}

type Director struct {
	ID        int    `json:"id"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

// [END] Database Entities

type ConnectionInfo struct {
	driver   string
	username string
	password string
	host     string
	port     int
	database string
}

type Storage struct {
	DB             *sql.DB
	ConnectionInfo ConnectionInfo
}

func (s *Storage) Connect() *Storage {
	ci := s.ConnectionInfo
	soueceUrl := fmt.Sprintf("%s://%s:%s@%s:%d/%s?sslmode=disable", ci.driver, ci.username, ci.password, ci.host, ci.port, ci.database)
	db, err := sql.Open(ci.driver, soueceUrl)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	return &Storage{DB: db}
}

func (s *Storage) findAllMovies() ([]Movie, error) {
	rows, err := s.DB.Query("SELECT m.id, m.isbn, m.title, d.id, d.first_name, d.last_name FROM movies m INNER JOIN directors d ON d.id = m.director_id")
	if err != nil {
		log.Printf("[ERROR] Failed to execute 'SELECT all movies' query: %v", err)
		return nil, fmt.Errorf("query execution failed: %w", err)
	}

	defer rows.Close()

	var movies []Movie
	for rows.Next() {
		var movie Movie
		var director Director

		if err := rows.Scan(&movie.ID, &movie.Isbn, &movie.Title, &director.ID, &director.FirstName, &director.LastName); err != nil {
			log.Println("[ERROR] Scan movies from select all statement is failed.", err)
			return nil, fmt.Errorf("row scan failed: %w", err)
		}
		movie.Director = director
		movies = append(movies, movie)
	}

	return movies, nil
}

func (s *Storage) findMovieById(id int) (*Movie, error) {
	row := s.DB.QueryRow("SELECT m.id, m.isbn, m.title, d.id, d.first_name, d.last_name FROM movies m INNER JOIN directors d ON d.id = m.director_id WHERE m.id = $1", id)

	var movie Movie
	var director Director

	if err := row.Scan(&movie.ID, &movie.Isbn, &movie.Title, &director.ID, &director.FirstName, &director.LastName); err != nil {
		if err == sql.ErrNoRows {
			log.Printf("[WARN] Movie with id '%d' not found", id)
			return nil, nil
		}
		log.Printf("[ERROR] Failed to scan movie data for id '%d': %v", id, err)
		return nil, fmt.Errorf("failed to scan movie data: %w", err)
	}

	movie.Director = director

	return &movie, nil
}

func (s *Storage) removeMovieById(id int) (bool, error) {
	result, err := s.DB.Exec("DELETE FROM movies WHERE id = $1", id)
	if err != nil {
		log.Printf("[ERROR] Failed to execute delete statement for movie with id '%d': %v", id, err)
		return false, fmt.Errorf("delete statement execution failed: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		log.Printf("[ERROR] Failed to retrieve affected rows after deleting movie with id '%d': %v", id, err)
		return false, fmt.Errorf("retrieving affected rows failed: %w", err)
	}

	return affected >= 1, nil
}

func (s *Storage) insertMovie(movie Movie, director *Director) (int64, error) {
	trx, err := s.DB.Begin()
	if err != nil {
		log.Printf("[ERROR] Failed to begin transaction for movie with id - '%d'. %v", movie.ID, err)
		return -1, fmt.Errorf("transaction begin failed: %w", err)
	}

	defer func() {
		if err != nil {
			trx.Rollback()
		}
	}()

	var directorId int64
	if director == nil {
		err = s.DB.QueryRow("INSERT INTO directors (first_name, last_name) VALUES ($1, $2) RETURNING id", movie.Director.FirstName, movie.Director.LastName).Scan(&directorId)
		if err != nil {
			log.Printf("[ERROR] Failed to insert director for movie with id - '%d': %v", movie.ID, err)
			return -1, fmt.Errorf("insert director failed: %w", err)
		}
	} else {
		directorId = int64(director.ID)
	}

	var movieId int64
	err = s.DB.QueryRow("INSERT INTO movies (isbn, title, director_id) VALUES ($1, $2, $3) RETURNING id", movie.Isbn, movie.Title, directorId).Scan(&movieId)
	if err != nil {
		log.Printf("[ERROR] Failed to insert movie with title '%s': %v", movie.Title, err)
		return -1, fmt.Errorf("insert movie failed: %w", err)
	}

	if err = trx.Commit(); err != nil {
		log.Printf("[ERROR] Failed to commit transaction for movie with title '%s': %v", movie.Title, err)
		return -1, fmt.Errorf("transaction commit failed: %w", err)
	}

	return movieId, nil
}

func (s *Storage) findDirectorById(id int) (*Director, error) {
	row := s.DB.QueryRow("SELECT * FROM directors d WHERE d.id = $1", id)

	var director Director
	err := row.Scan(&director.ID, &director.FirstName, &director.LastName)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Printf("[WARN] Director with id '%d' not found", id)
			return nil, nil
		}
		log.Printf("[ERROR] Failed to scan director data for id '%d': %v", id, err)
		return nil, fmt.Errorf("scan failed for director id %d: %w", id, err)
	}

	return &director, nil
}
