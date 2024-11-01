package main

import (
	app "movies-service/internal"
)

func main() {
	dbConnection := app.NewDbConnection("postgres", "root", "secret", "localhost", "5432", "movies")
	app.New(dbConnection).Run(8000)
}
