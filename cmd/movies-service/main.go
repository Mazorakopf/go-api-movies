package main

import (
	"log"
	app "movies-service/internal"
)

func main() {
	config, err := app.LoadConfig()
	if err != nil {
		log.Fatalf("[ERROR] Failed to load configuration: %s", err)
	}

	connectionInfo := &app.ConnectionInfo{
		Username: config.DBUser,
		Password: config.DBPass,
		Host:     config.DBHost,
		Port:     config.DBPort,
	}

	app.New(connectionInfo).Run(config.AppPort)
}
