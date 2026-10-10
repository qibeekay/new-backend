package routes_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"portfolio-api/internal/config"
	"portfolio-api/internal/database"
	"portfolio-api/internal/models"
	"portfolio-api/internal/repository"
	"portfolio-api/internal/routes"
)

func setupTestApp(t *testing.T) (*gin.Engine, *config.Config, *repository.Repository, func()) {
	gin.SetMode(gin.TestMode)
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("config load failed: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		t.Fatalf("database connection failed: %v", err)
	}

	repo := repository.New(pool)
	r := gin.New()
	routes.Setup(r, cfg, repo)

	cleanup := func() {
		pool.Close()
	}

	return r, cfg, repo, cleanup
}

func TestHealthAndSwaggerEndpoints(t *testing.T) {
	r, _, _, cleanup := setupTestApp(t)
	defer cleanup()

	// 1. Health check
	req, _ := http.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 on /healthz, got %d", w.Code)
	}

	for _, ri := range r.Routes() {
		t.Logf("Route: %s %s", ri.Method, ri.Path)
	}

	// 2. Swagger doc json
	req, _ = http.NewRequest(http.MethodGet, "/swagger/doc.json", nil)
	req.RequestURI = "/swagger/doc.json"
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 on /swagger/doc.json, got %d", w.Code)
	}

	// 3. Swagger UI html
	req, _ = http.NewRequest(http.MethodGet, "/swagger/index.html", nil)
	req.RequestURI = "/swagger/index.html"
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200 on /swagger/index.html, got %d", w.Code)
	}

	// 4. Swagger redirect from /swagger
	req, _ = http.NewRequest(http.MethodGet, "/swagger", nil)
	req.RequestURI = "/swagger"
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusMovedPermanently {
		t.Errorf("expected 301 on /swagger, got %d", w.Code)
	}
}

func TestAuthAndProjectAPIFlow(t *testing.T) {
	r, _, repo, cleanup := setupTestApp(t)
	defer cleanup()

	// 1. Attempt login with bad password -> 401
	badLoginBody, _ := json.Marshal(map[string]string{
		"email":    "admin@example.com",
		"password": "wrongpassword",
	})
	req, _ := http.NewRequest(http.MethodPost, "/api/admin/login", bytes.NewBuffer(badLoginBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 on bad login, got %d", w.Code)
	}

	// 2. Attempt login with correct password -> 200 + token
	goodLoginBody, _ := json.Marshal(map[string]string{
		"email":    "admin@example.com",
		"password": "admin123",
	})
	req, _ = http.NewRequest(http.MethodPost, "/api/admin/login", bytes.NewBuffer(goodLoginBody))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 on successful login, got %d: %s", w.Code, w.Body.String())
	}

	var authResp struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &authResp); err != nil || authResp.Token == "" {
		t.Fatalf("invalid auth response: %s", w.Body.String())
	}

	// 3. Unauthorized attempt to create project -> 401
	projectPayload := map[string]any{
		"title":       "E2E Integration Project",
		"tagline":     "Integration testing with Gin & Swagger",
		"year":        "2026",
		"category":    "systems",
		"status":      "completed",
		"description": "Comprehensive integration test ensuring end to end workflow validity.",
		"role":        "Core Architect",
		"stack":       []string{"Go", "PostgreSQL", "Swagger"},
		"metrics": []map[string]string{
			{"value": "100%", "label": "Test coverage"},
		},
	}
	pBytes, _ := json.Marshal(projectPayload)

	req, _ = http.NewRequest(http.MethodPost, "/api/admin/projects", bytes.NewBuffer(pBytes))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 when creating project without token, got %d", w.Code)
	}

	// 4. Authorized create project -> 201
	req, _ = http.NewRequest(http.MethodPost, "/api/admin/projects", bytes.NewBuffer(pBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+authResp.Token)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201 on authorized project creation, got %d: %s", w.Code, w.Body.String())
	}

	var createdProject models.Project
	if err := json.Unmarshal(w.Body.Bytes(), &createdProject); err != nil {
		t.Fatalf("failed to parse created project: %v", err)
	}
	defer func() {
		_ = repo.Projects.Delete(context.Background(), createdProject.ID)
	}()

	// 5. Public GET /api/projects -> 200
	req, _ = http.NewRequest(http.MethodGet, "/api/projects", nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 on /api/projects, got %d", w.Code)
	}

	// 6. Public GET /api/projects/:id -> 200
	req, _ = http.NewRequest(http.MethodGet, "/api/projects/"+createdProject.ID, nil)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 on /api/projects/:id, got %d", w.Code)
	}

	// 7. Authorized DELETE /api/admin/projects/:id -> 204
	req, _ = http.NewRequest(http.MethodDelete, "/api/admin/projects/"+createdProject.ID, nil)
	req.Header.Set("Authorization", "Bearer "+authResp.Token)
	w = httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204 on delete project, got %d", w.Code)
	}
}
