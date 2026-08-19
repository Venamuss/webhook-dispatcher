package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"time"

	"github.com/Venamuss/webhook-dispatcher/internal/platform/database"
)

func main() {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		log.Fatalf("DB_DSN environment variable is not set")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	db, err := database.New(ctx, dsn)
	if err != nil {
		log.Fatalf("Failed to create database instance: %v", err)
	}
	defer db.Close()

	slog.Info("running database migrations...")
	if err := db.RunMigrations(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
	slog.Info("database migrations completed successfully")

	db.Close()
}
