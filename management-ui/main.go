package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/localvsp/management-ui/internal/buildlog"
	"github.com/localvsp/management-ui/internal/docker"
	"github.com/localvsp/management-ui/internal/handlers"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Mark any builds that were "running" when the server last stopped as failed.
	buildlog.RecoverStale()

	// Start background goroutine that prunes finished jobs from memory after 1h.
	handlers.StartBuildJobCleanup()

	// Reconcile any stale project/site Traefik labels after updates.
	go func() {
		if err := docker.ReconcileExistingProjectRouting(); err != nil {
			log.Printf("routing reconciliation failed: %v", err)
		}
	}()

	mux := handlers.NewMux()

	srv := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Run server in a goroutine
	go func() {
		fmt.Printf("Management UI starting on port %s...\n", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	fmt.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	fmt.Println("Server exiting")
}
