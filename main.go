package main

import (
	"embed"
	"log"

	"bw_util/internal/app"
)

//go:embed templates
var templatesFS embed.FS

func main() {
	// Create and initialize the application
	application := app.New(templatesFS)

	if err := application.Initialize(); err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}

	// Run the application
	if err := application.Run(); err != nil {
		log.Fatalf("Application error: %v", err)
	}
}
