// main.go
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

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"document-embeddings/internal/api"
	"document-embeddings/internal/config"
	"document-embeddings/internal/repository"
	"document-embeddings/internal/services"
	"document-embeddings/pkg/database"
	"document-embeddings/pkg/logger"
	"document-embeddings/pkg/minio"
	"document-embeddings/pkg/openai"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Printf("Warning: .env file not found: %v", err)
	}

	// Initialize configuration
	cfg := config.Load()

	// Initialize logger
	logger := logger.New(cfg.LogLevel)

	// Initialize database
	db, err := database.New(cfg.Database)
	if err != nil {
		logger.Fatal("Failed to initialize database", "error", err)
	}
	defer db.Close()

	// Initialize MinIO client
	minioClient, err := minio.New(cfg.MinIO)
	if err != nil {
		logger.Fatal("Failed to initialize MinIO client", "error", err)
	}

	// Initialize OpenAI client
	openaiClient := openai.New(cfg.OpenAI)

	// Initialize repositories
	repo := repository.New(db, logger)

	// Initialize services
	svc := services.New(repo, minioClient, openaiClient, cfg, logger)

	// Initialize API handlers
	handler := api.New(svc, logger)

	// Setup Gin router
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(api.CORSMiddleware())
	r.Use(api.LoggingMiddleware(logger))

	// Register routes
	api.RegisterRoutes(r, handler)

	// Create HTTP server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.Port),
		Handler: r,
	}

	// Start server in goroutine
	go func() {
		logger.Info("Starting server", "port", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", "error", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", "error", err)
	}

	logger.Info("Server exited")
}

