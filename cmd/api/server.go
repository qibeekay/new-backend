package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"portfolio-api/internal/config"
	"portfolio-api/internal/repository"
	"portfolio-api/internal/routes"
)

type Server struct{ http *http.Server }

func NewServer(cfg *config.Config, repo *repository.Repository) *Server {
	r := gin.New()
	corsCfg := cors.Config{
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS", "HEAD"},
		AllowHeaders:     []string{"Authorization", "Content-Type", "Accept", "Origin", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
		AllowOriginFunc: func(origin string) bool {
			originNorm := strings.TrimRight(strings.ToLower(origin), "/")
			for _, o := range cfg.AllowedOrigins {
				oNorm := strings.TrimRight(strings.ToLower(o), "/")
				if oNorm == "*" || oNorm == originNorm {
					return true
				}
			}
			return strings.HasSuffix(originNorm, ".vercel.app")
		},
	}
	r.Use(gin.Logger(), gin.Recovery(), cors.New(corsCfg))
	routes.Setup(r, cfg, repo)

	return &Server{http: &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
	}}
}

// Run blocks until ctx is cancelled, then shuts down gracefully.
func (s *Server) Run(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		log.Printf("listening on %s", s.http.Addr)
		if err := s.http.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return s.http.Shutdown(shutdownCtx)
	}
}
