package main

import (
	"acl-challenge/internal/api/router"
	"log"
	"os"
)

func main() {
	r := router.NewRouter(router.Dependencies{})
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	if err := r.Run("0.0.0.0:" + port); err != nil {
		log.Fatal(err)
	}
}
