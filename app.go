package main

import (
	"fmt"
	"log"
	"movies-service/movies"
	"net/http"

	"github.com/gorilla/mux"
)

type App struct {
	router *mux.Router
}

func NewApp() *App {
	app := App{router: mux.NewRouter()}
	movies.SetupRouting(app.router)
	return &app
}

func (a *App) Run(port int) {
	fmt.Printf("Starting server at port %d\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), a.router))
}
