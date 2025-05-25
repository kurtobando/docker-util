package app

import (
	"context"
	"embed"
	"log"
	"os"
	"os/signal"
	"syscall"

	"bw_util/internal/config"
	"bw_util/internal/database"
	"bw_util/internal/docker"
	"bw_util/internal/web"
)

// App represents the main application
type App struct {
	config          *config.Config
	db              *database.Database
	repository      *database.Repository
	dockerClient    *docker.Client
	collector       *docker.Collector
	templateManager *web.TemplateManager
	handler         *web.Handler
	server          *web.Server
	templatesFS     embed.FS
}

// New creates a new application instance
func New(templatesFS embed.FS) *App {
	return &App{
		templatesFS: templatesFS,
	}
}

// Initialize sets up all application components
// This preserves the exact same initialization logic from the original main function
func (a *App) Initialize() error {
	// Load configuration
	a.config = config.Load()

	// Initialize database
	a.db = database.NewDatabase(a.config.DBPath)
	if err := a.db.Connect(); err != nil {
		return err
	}

	// Create repository
	a.repository = database.NewRepository(a.db)

	// Initialize template manager
	a.templateManager = web.NewTemplateManager(a.templatesFS)
	if err := a.templateManager.LoadTemplates(); err != nil {
		return err
	}

	// Create web components
	a.handler = web.NewHandler(a.repository, a.templateManager, a.config.DBPath)
	a.server = web.NewServer(a.config.ServerPort, a.handler)

	// Initialize Docker client
	var err error
	a.dockerClient, err = docker.NewClient()
	if err != nil {
		log.Printf("Error creating Docker client (log collection might not work): %v", err)
		// Decide if this is fatal. For now, web UI might still work with existing DB data.
		return nil // Continue without Docker client
	}

	// Create log collector
	a.collector = docker.NewCollector(a.dockerClient, a.repository)

	return nil
}

// Run starts the application and handles the main execution loop
// This preserves the exact same execution logic from the original main function
func (a *App) Run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutdown signal received, initiating graceful shutdown...")
		cancel()
	}()

	// Start Docker log collection if Docker client is available
	if a.dockerClient != nil {
		containers, err := a.dockerClient.ListRunningContainers(ctx)
		if err != nil {
			if ctx.Err() == context.Canceled {
				log.Println("Context cancelled during container listing.")
			} else {
				log.Printf("Error listing containers (log collection might be impacted): %v", err)
			}
		} else {
			if err := a.collector.StartCollection(ctx, containers); err != nil {
				log.Printf("Error starting log collection: %v", err)
			}
		}
	}

	// Start web server
	if err := a.server.Start(ctx); err != nil {
		return err
	}

	// Wait for shutdown signal
	<-ctx.Done()

	// Shutdown web server
	if err := a.server.Shutdown(ctx); err != nil {
		log.Printf("Error shutting down web server: %v", err)
	}

	// Wait for log streaming goroutines to finish
	if a.collector != nil {
		log.Println("Waiting for log streaming goroutines to finish...")
		a.collector.Wait()
	}

	// Close resources
	log.Println("All goroutines finished. Closing resources.")
	if a.db != nil {
		if err := a.db.Close(); err != nil {
			log.Printf("Error closing database: %v\n", err)
		}
	}
	if a.dockerClient != nil {
		if err := a.dockerClient.Close(); err != nil {
			log.Printf("Error closing Docker client: %v\n", err)
		}
	}

	log.Println("Application shut down gracefully.")
	return nil
}
