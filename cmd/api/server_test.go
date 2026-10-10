package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"portfolio-api/internal/config"
	"portfolio-api/internal/repository"
)

func TestCORS(t *testing.T) {
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("config: %v", err)
	}

	t.Logf("AllowedOrigins: %#v", cfg.AllowedOrigins)

	srv := NewServer(cfg, repository.New(nil))

	req, _ := http.NewRequest(http.MethodOptions, "/api/projects", nil)
	req.Header.Set("Origin", "https://qibeekay.vercel.app")
	req.Header.Set("Access-Control-Request-Method", "GET")

	w := httptest.NewRecorder()
	srv.http.Handler.ServeHTTP(w, req)

	t.Logf("Response status: %d, headers: %#v", w.Code, w.Header())
	if w.Code != http.StatusNoContent && w.Code != http.StatusOK {
		t.Errorf("expected 200 or 204 for CORS preflight, got %d", w.Code)
	}
	if w.Header().Get("Access-Control-Allow-Origin") != "https://qibeekay.vercel.app" {
		t.Errorf("expected Access-Control-Allow-Origin: https://qibeekay.vercel.app, got %q", w.Header().Get("Access-Control-Allow-Origin"))
	}
}
