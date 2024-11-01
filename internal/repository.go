package internal

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

// [START] Database Entities
type movie struct {
	ID       int      `json:"id"`
	Isbn     string   `json:"isbn"`
	Title    string   `json:"title"`
	Director director `json:"director"`
}

type director struct {
	ID        int    `json:"id"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

// [END] Database Entities

type connection struct {
	DB *sql.DB
}

func NewDbConnection(driver, username, password, host, port, database string) *connection {
	soueceUrl := fmt.Sprintf("%s://%s:%s@%s:%s/%s?sslmode=disable", driver, username, password, host, port, database)
	db, err := sql.Open(driver, soueceUrl)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	return &connection{DB: db}
}

func (c *connection) findAllMovies() ([]movie, error) {
	rows, err := c.DB.Query("SELECT m.id, m.isbn, m.title, d.id, d.first_name, d.last_name FROM movies m INNER JOIN directors d ON d.id = m.director_id")
	if err != nil {
		log.Printf("[ERROR] Failed to execute 'SELECT all movies' query: %v", err)
		return nil, fmt.Errorf("query execution failed: %w", err)
	}

	defer rows.Close()

	var movies []movie
	for rows.Next() {
		var (
			movie    movie
			director director
		)

		if err := rows.Scan(&movie.ID, &movie.Isbn, &movie.Title, &director.ID, &director.FirstName, &director.LastName); err != nil {
			log.Println("[ERROR] Scan movies from select all statement is failed.", err)
			return nil, fmt.Errorf("row scan failed: %w", err)
		}
		movie.Director = director
		movies = append(movies, movie)
	}

	return movies, nil
}

func (c *connection) findMovieById(id int) (*movie, error) {
	row := c.DB.QueryRow("SELECT m.id, m.isbn, m.title, d.id, d.first_name, d.last_name FROM movies m INNER JOIN directors d ON d.id = m.director_id WHERE m.id = $1", id)

	var (
		movie    movie
		director director
	)

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

func (c *connection) removeMovieById(id int) (bool, error) {
	result, err := c.DB.Exec("DELETE FROM movies WHERE id = $1", id)
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

func (c *connection) insertMovie(movie movie, director *director) (int64, error) {
	trx, err := c.DB.Begin()
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
		err = c.DB.QueryRow("INSERT INTO directors (first_name, last_name) VALUES ($1, $2) RETURNING id", movie.Director.FirstName, movie.Director.LastName).Scan(&directorId)
		if err != nil {
			log.Printf("[ERROR] Failed to insert director for movie with id - '%d': %v", movie.ID, err)
			return -1, fmt.Errorf("insert director failed: %w", err)
		}
	} else {
		directorId = int64(director.ID)
	}

	var movieId int64
	err = c.DB.QueryRow("INSERT INTO movies (isbn, title, director_id) VALUES ($1, $2, $3) RETURNING id", movie.Isbn, movie.Title, directorId).Scan(&movieId)
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

func (c *connection) findDirectorById(id int) (*director, error) {
	row := c.DB.QueryRow("SELECT * FROM directors d WHERE d.id = $1", id)

	var director director
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
