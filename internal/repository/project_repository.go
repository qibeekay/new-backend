package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"portfolio-api/internal/models"
)

type ProjectRepository interface {
	List(ctx context.Context) ([]models.Project, error)
	Get(ctx context.Context, id string) (models.Project, error)
	Create(ctx context.Context, p models.Project) error
	Update(ctx context.Context, p models.Project) error
	Delete(ctx context.Context, id string) error
}

type projectRepository struct{ db *pgxpool.Pool }

func NewProjectRepository(db *pgxpool.Pool) ProjectRepository {
	return &projectRepository{db: db}
}

const projectCols = `id, title, tagline, year, category, status, featured, image, image_alt,
	description, role, stack, metrics, live_url, code_url`

func scanProject(row pgx.Row) (models.Project, error) {
	var p models.Project
	err := row.Scan(&p.ID, &p.Title, &p.Tagline, &p.Year, &p.Category, &p.Status, &p.Featured,
		&p.Image, &p.ImageAlt, &p.Description, &p.Role, &p.Stack, &p.Metrics, &p.LiveURL, &p.CodeURL)
	if p.Stack == nil {
		p.Stack = []string{}
	}
	if p.Metrics == nil {
		p.Metrics = []models.Metric{}
	}
	return p, err
}

func (r *projectRepository) List(ctx context.Context) ([]models.Project, error) {
	rows, err := r.db.Query(ctx, `SELECT `+projectCols+` FROM projects ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []models.Project{}
	for rows.Next() {
		p, err := scanProject(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *projectRepository) Get(ctx context.Context, id string) (models.Project, error) {
	p, err := scanProject(r.db.QueryRow(ctx, `SELECT `+projectCols+` FROM projects WHERE id = $1`, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return models.Project{}, ErrNotFound
	}
	return p, err
}

func (r *projectRepository) Create(ctx context.Context, p models.Project) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO projects (`+projectCols+`)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)`,
		p.ID, p.Title, p.Tagline, p.Year, p.Category, p.Status, p.Featured, p.Image, p.ImageAlt,
		p.Description, p.Role, p.Stack, p.Metrics, p.LiveURL, p.CodeURL)
	return err
}

func (r *projectRepository) Update(ctx context.Context, p models.Project) error {
	tag, err := r.db.Exec(ctx, `
		UPDATE projects SET title=$2, tagline=$3, year=$4, category=$5, status=$6, featured=$7,
			image=$8, image_alt=$9, description=$10, role=$11, stack=$12, metrics=$13,
			live_url=$14, code_url=$15, updated_at=now()
		WHERE id=$1`,
		p.ID, p.Title, p.Tagline, p.Year, p.Category, p.Status, p.Featured, p.Image, p.ImageAlt,
		p.Description, p.Role, p.Stack, p.Metrics, p.LiveURL, p.CodeURL)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *projectRepository) Delete(ctx context.Context, id string) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM projects WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
