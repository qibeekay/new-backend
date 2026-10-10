package repository_test

import (
	"context"
	"os"
	"testing"
	"time"

	"portfolio-api/internal/config"
	"portfolio-api/internal/database"
	"portfolio-api/internal/models"
	"portfolio-api/internal/repository"
)

func TestProjectRepositoryCRUD(t *testing.T) {
	_ = os.Setenv("PORT", "8080")
	cfg, err := config.Load()
	if err != nil {
		t.Skipf("skipping db test, config load failed: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		t.Fatalf("database connect error: %v", err)
	}
	defer pool.Close()

	repo := repository.NewProjectRepository(pool)

	testID := "test-" + models.NewProjectID("Test Project")
	p := models.Project{
		ID:          testID,
		Title:       "Test Project",
		Tagline:     "Test Tagline",
		Year:        "2026",
		Category:    "systems",
		Status:      "completed",
		Featured:    true,
		Image:       "https://example.com/test.png",
		ImageAlt:    "Test Alt",
		Description: "This is a test description that meets minimum length requirements.",
		Role:        "Tester",
		Stack:       []string{"Go", "PostgreSQL"},
		Metrics: []models.Metric{
			{Value: "100%", Label: "Test Pass"},
		},
		LiveURL: "https://example.com",
		CodeURL: "https://github.com/example/test",
	}

	// Clean up after test
	defer func() {
		_ = repo.Delete(context.Background(), testID)
	}()

	// 1. Create
	if err := repo.Create(ctx, p); err != nil {
		t.Fatalf("repo.Create failed: %v", err)
	}

	// 2. Get
	got, err := repo.Get(ctx, testID)
	if err != nil {
		t.Fatalf("repo.Get failed: %v", err)
	}
	if got.Title != p.Title {
		t.Errorf("expected title %s, got %s", p.Title, got.Title)
	}
	if len(got.Metrics) != 1 || got.Metrics[0].Value != "100%" {
		t.Errorf("metrics not scanned properly: %+v", got.Metrics)
	}

	// 3. List
	list, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("repo.List failed: %v", err)
	}
	found := false
	for _, item := range list {
		if item.ID == testID {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("created project not found in list")
	}

	// 4. Update
	p.Title = "Updated Test Project"
	if err := repo.Update(ctx, p); err != nil {
		t.Fatalf("repo.Update failed: %v", err)
	}
	gotUpdated, err := repo.Get(ctx, testID)
	if err != nil {
		t.Fatalf("repo.Get after update failed: %v", err)
	}
	if gotUpdated.Title != "Updated Test Project" {
		t.Errorf("expected updated title, got %s", gotUpdated.Title)
	}

	// 5. Delete
	if err := repo.Delete(ctx, testID); err != nil {
		t.Fatalf("repo.Delete failed: %v", err)
	}
	_, err = repo.Get(ctx, testID)
	if err != repository.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}
