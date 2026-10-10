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

	tests := []struct {
		origin  string
		allowed bool
	}{
		{"https://qibeekay.vercel.app", true},
		{"https://preview-deploy.vercel.app", true},
		{"http://localhost:5173", true},
		{"https://evil.com", false},
	}

	for _, tt := range tests {
		req, _ := http.NewRequest(http.MethodOptions, "/api/projects", nil)
		req.Header.Set("Origin", tt.origin)
		req.Header.Set("Access-Control-Request-Method", "GET")

		w := httptest.NewRecorder()
		srv.http.Handler.ServeHTTP(w, req)

		if tt.allowed {
			if w.Code != http.StatusNoContent && w.Code != http.StatusOK {
				t.Errorf("origin %s: expected 200 or 204, got %d", tt.origin, w.Code)
			}
			if w.Header().Get("Access-Control-Allow-Origin") != tt.origin {
				t.Errorf("origin %s: expected header %q, got %q", tt.origin, tt.origin, w.Header().Get("Access-Control-Allow-Origin"))
			}
		} else {
			if w.Header().Get("Access-Control-Allow-Origin") != "" {
				t.Errorf("origin %s: expected no CORS header, got %q", tt.origin, w.Header().Get("Access-Control-Allow-Origin"))
			}
		}
	}
}
