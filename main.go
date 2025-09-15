package tests
package main

import (
	"log"
	"os"

	adapter "github.com/DrWeltschmerz/users-adapter-gin/ginadapter"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	r := adapter.NewRouter()
	log.Printf("Starting users service on :%s", port)
	r.Run(":" + port)
}
