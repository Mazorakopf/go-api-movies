package main

import (
	app "movies-service/internal"
)

func main() {
	app.New(
		&app.ConnectionInfo{
			Username: "root",
			Password: "secret",
			Host:     "localhost",
			Port:     5432,
		},
	).Run(8000)
}
