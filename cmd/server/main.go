package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"family-shopping-list-api/api"
	"family-shopping-list-api/internal/config"
	"family-shopping-list-api/internal/database"
)

func main() {
	cfg := config.Load()
	db, err := database.Open(cfg)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	if err := database.WaitForDB(ctx, db, 80*time.Second); err != nil {
		log.Fatalf("wait for database: %v", err)
	}

	migrationsDir := os.Getenv("MIGRATIONS_DIR")
	if migrationsDir == "" {
		migrationsDir = "./migrations"
	}
	migrateCtx, migrateCancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer migrateCancel()
	if err := database.Migrate(migrateCtx, db, migrationsDir); err != nil {
		log.Fatalf("migrate database: %v", err)
	}
	if err := database.SeedDemoData(migrateCtx, db); err != nil {
		log.Fatalf("seed demo data: %v", err)
	}

	server := &http.Server{
		Addr:              ":" + cfg.ServerPort,
		Handler:           api.NewRouter(db, cfg),
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		log.Printf("HTTP server listening on :%s", cfg.ServerPort)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("http server: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	}
	log.Println("server stopped")
}
