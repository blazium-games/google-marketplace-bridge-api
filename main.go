package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google-marketplace-bridge/api/internal/config"
	"google-marketplace-bridge/api/internal/handlers"
	"google-marketplace-bridge/api/internal/models"
	"google-marketplace-bridge/api/internal/webhooks"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	db, err := gorm.Open(postgres.Open(cfg.PostgresDSN), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}
	if err := models.AutoMigrate(db); err != nil {
		log.Fatal(err)
	}

	h := handlers.New(cfg, db)
	mux := http.NewServeMux()
	// Exact root and /health only (Go 1.22+ patterns); do not use "/" alone — it matches all paths.
	mux.HandleFunc("GET /{$}", handlers.Health)
	mux.HandleFunc("HEAD /{$}", handlers.Health)
	mux.HandleFunc("GET /health", handlers.Health)
	mux.HandleFunc("HEAD /health", handlers.Health)
	mux.HandleFunc("/instantiate", h.Instantiate)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go webhooks.RunWorker(ctx, db, cfg.WebhookInterval)

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	go func() {
		log.Printf("listening on %s", cfg.HTTPAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	<-sig
	cancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown: %v", err)
	}
}
