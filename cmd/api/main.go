package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"portfolio-api/internal/config"
	"portfolio-api/internal/database"
	"portfolio-api/internal/repository"
)

// @title           Portfolio API
// @version         1.0
// @description     Backend REST API for Portfolio platform: projects showcase, admin authentication, and media uploads.
// @termsOfService  http://swagger.io/terms/

// @contact.name    API Support
// @contact.url     http://localhost:8080

// @license.name    MIT
// @license.url     https://opensource.org/licenses/MIT

// @host            localhost:8080
// @BasePath        /api

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and your JWT token.
func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}
	if err := os.MkdirAll(cfg.UploadDir, 0o755); err != nil {
		log.Fatalf("upload dir: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	db, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer db.Close()

	// Start background self-pinger to prevent free-tier inactivity spin-down (e.g. Render 15-min sleep)
	if cfg.KeepAliveURL != "" {
		startKeepAlive(ctx, cfg.KeepAliveURL)
	}

	if err := NewServer(cfg, repository.New(db)).Run(ctx); err != nil {
		log.Fatalf("server: %v", err)
	}
}

// startKeepAlive periodically sends a GET request to the public /healthz endpoint
// to prevent free hosting platforms (like Render) from idling after 15 minutes.
func startKeepAlive(ctx context.Context, targetURL string) {
	pingURL := targetURL
	if !strings.HasSuffix(pingURL, "/healthz") {
		pingURL = strings.TrimRight(pingURL, "/") + "/healthz"
	}

	log.Printf("keep-alive: self-pinger enabled for %s (interval: 10m)", pingURL)
	go func() {
		// Ping every 10 minutes to stay well within Render's 15-minute sleep threshold
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()

		client := &http.Client{Timeout: 15 * time.Second}
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				resp, err := client.Get(pingURL)
				if err != nil {
					log.Printf("keep-alive: warning pinging %s: %v", pingURL, err)
					continue
				}
				_ = resp.Body.Close()
				log.Printf("keep-alive: ping successful (status %d)", resp.StatusCode)
			}
		}
	}()
}
