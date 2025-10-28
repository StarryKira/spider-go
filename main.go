package main

import (
	"log"
	"spider-go/internal/app"
)

func main() {
	if err := app.LoadConfig(); err != nil {
		log.Fatalf("config error: %v \n", err)
	}
	initDB()
}
