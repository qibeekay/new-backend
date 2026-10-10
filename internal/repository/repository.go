package repository

import (
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("not found")

// Repository groups every repository so it can be passed around as one value.
type Repository struct {
	Projects ProjectRepository
}

func New(db *pgxpool.Pool) *Repository {
	return &Repository{Projects: NewProjectRepository(db)}
}
