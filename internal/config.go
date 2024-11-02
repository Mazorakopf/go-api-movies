package internal

import (
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type config struct {
	AppPort int
	DBPort  int
	DBUser  string
	DBPass  string
	DBHost  string
}

func LoadConfig() (*config, error) {
	if err := godotenv.Load(); err != nil {
		log.Printf("[WARN] .env file not found: %s", err)
	}

	appPort, err := getEnvAsInt("APP_PORT")
	if err != nil {
		return nil, err
	}

	dbPort, err := getEnvAsInt("DB_PORT")
	if err != nil {
		return nil, err
	}

	return &config{
		AppPort: appPort,
		DBPort:  dbPort,
		DBUser:  os.Getenv("DB_USER"),
		DBPass:  os.Getenv("DB_PASSWORD"),
		DBHost:  os.Getenv("DB_HOST"),
	}, nil
}

func getEnvAsInt(key string) (int, error) {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return 0, fmt.Errorf("missing environment variable: %s", key)
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return 0, fmt.Errorf("environment variable %s is not a valid integer: %v", key, err)
	}
	return value, nil
}
