package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"go-app/internal/config"
	"go-app/internal/db"
	"go-app/internal/server"
	"go-app/internal/todo"
	"go.uber.org/zap"
)

func main() {
	ctx := context.Background()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	database, err := db.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer database.Close()

	logger, err := newLogger(cfg.LogFilePath)
	if err != nil {
		log.Fatalf("failed to init logger: %v", err)
	}
	defer func() {
		_ = logger.Sync()
	}()

	repo := todo.NewRepository(database)
	srv := server.New(repo, logger)

	httpServer := &http.Server{
		Addr:              ":" + fmt.Sprint(cfg.Port),
		Handler:           srv.Router(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	log.Printf("Go Todo API listening on :%d", cfg.Port)
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("server error: %v", err)
	}
}

func newLogger(logFilePath string) (*zap.Logger, error) {
	cfg := zap.NewProductionConfig()

	outputs := []string{"stdout"}
	if logFilePath != "" {
		if err := os.MkdirAll(filepath.Dir(logFilePath), 0o755); err != nil {
			// Fall back to stdout-only rather than failing the service.
			log.Printf("failed to create log dir: %v", err)
		} else {
			outputs = append(outputs, logFilePath)
		}
	}

	cfg.OutputPaths = outputs
	cfg.ErrorOutputPaths = outputs
	return cfg.Build()
}
