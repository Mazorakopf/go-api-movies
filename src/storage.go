package main

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

type Movie struct {
	ID       int       `json:"id"`
	Isbn     string    `json:"isbn"`
	Title    string    `json:"title"`
	Director *Director `json:"director"`
}

type Director struct {
	ID        int    `json:"id"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

var DB *sql.DB

func connectDB() {
	db, err := sql.Open("postgres", "postgres://root:secret@localhost:5432/movies?sslmode=disable")
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	DB = db
}

func findAllMovies() (*[]Movie, error) {
	rows, err := DB.Query("SELECT m.id, m.isbn, m.title, d.id, d.first_name, d.last_name FROM movies m INNER JOIN directors d ON d.id = m.id")
	if err != nil {
		log.Println("SQL statement failed", err)
		return nil, err
	}

	defer rows.Close()

	var movies []Movie

	for rows.Next() {
		var movie Movie
		var director Director

		if err := rows.Scan(&movie.ID, &movie.Isbn, &movie.Title, &director.ID, &director.FirstName, &director.LastName); err != nil {
			log.Println("Scan clumns failed", err)
			return nil, err
		}
		movie.Director = &director
		movies = append(movies, movie)
	}

	return &movies, nil
}

func findMovieById(id int) (*Movie, error) {
	row := DB.QueryRow("SELECT m.id, m.isbn, m.title, d.id, d.first_name, d.last_name FROM movies m INNER JOIN directors d ON d.id = m.id WHERE m.id = $1", id)

	var movie Movie
	var director Director

	if err := row.Scan(&movie.ID, &movie.Isbn, &movie.Title, &director.ID, &director.FirstName, &director.LastName); err != nil {
		log.Println("Scan clumns failed", err)
		return nil, err
	}
	movie.Director = &director

	return &movie, nil
}

var users = []map[string]string{
	{"name": "admin", "password": "$2a$10$miwrWXNyiF7Qiv6ir9YTueSulDDrJfjj1w2r1dLpAmaOj/TglYyKG"},
}
