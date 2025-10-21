package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/casapps/casci/internal/config"
	"github.com/casapps/casci/pkg/database"
	"github.com/casapps/casci/pkg/server"
)

var (
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Print banner
	printBanner()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize database
	db, err := database.New(ctx, cfg.Database)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Run migrations
	if err := db.Migrate(ctx); err != nil {
		log.Fatalf("Failed to run database migrations: %v", err)
	}

	// Create server
	srv, err := server.New(cfg, db)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// Start server
	go func() {
		if err := srv.Start(); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Print access information
	fmt.Printf("\n🚀 CASCI is running!\n")
	fmt.Printf("   Web UI: http://localhost:%d\n", cfg.Server.Port)
	fmt.Printf("   API:    http://localhost:%d/api/v1\n", cfg.Server.Port)
	fmt.Printf("\n")
	fmt.Printf("First user to register becomes administrator.\n")
	fmt.Printf("\n")

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\nShutting down gracefully...")
	srv.Shutdown(ctx)
}

func printBanner() {
	banner := `
   ____    _    ____   ____ ___
  / ___|  / \  / ___| / ___|_ _|
 | |     / _ \ \___ \| |    | |
 | |___ / ___ \ ___) | |___ | |
  \____/_/   \_\____/ \____|___|

  CI/CD Application Server
  Version: %s
  Build:   %s
  Commit:  %s
`
	fmt.Printf(banner, Version, BuildTime, GitCommit)
}
