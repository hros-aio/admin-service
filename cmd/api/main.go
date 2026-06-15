// Package main is the entry point for the API service.
package main

import (
	"log"

	"github.com/hros/admin-service/internal/app"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}
	app.New().Run()
}
