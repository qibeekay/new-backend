package handlers

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"portfolio-api/internal/models"
	"portfolio-api/internal/repository"
	"portfolio-api/internal/response"
)

type ProjectHandler struct{ repo repository.ProjectRepository }

func NewProjectHandler(repo repository.ProjectRepository) *ProjectHandler {
	return &ProjectHandler{repo: repo}
}

// List godoc
// @Summary      List projects
// @Description  Returns all projects ordered by newest first
// @Tags         projects
// @Produce      json
// @Success      200 {array} models.Project
// @Failure      500 {object} response.ErrorEnvelope
// @Router       /projects [get]
func (h *ProjectHandler) List(c *gin.Context) {
	ps, err := h.repo.List(c.Request.Context())
	if err != nil {
		response.ServerError(c, err)
		return
	}
	response.OK(c, ps)
}

// Get godoc
// @Summary      Get a project
// @Description  Returns a single project by ID
// @Tags         projects
// @Produce      json
// @Param        id path string true "Project ID"
// @Success      200 {object} models.Project
// @Failure      404 {object} response.ErrorEnvelope
// @Failure      500 {object} response.ErrorEnvelope
// @Router       /projects/{id} [get]
func (h *ProjectHandler) Get(c *gin.Context) {
	p, err := h.repo.Get(c.Request.Context(), c.Param("id"))
	if h.failed(c, err) {
		return
	}
	response.OK(c, p)
}

// Create godoc
// @Summary      Create a project
// @Description  Creates a new portfolio project (Admin only)
// @Tags         admin-projects
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        input body models.ProjectInput true "Project details"
// @Success      201 {object} models.Project
// @Failure      400 {object} response.ErrorEnvelope
// @Failure      401 {object} response.ErrorEnvelope
// @Failure      422 {object} response.ValidationEnvelope
// @Failure      500 {object} response.ErrorEnvelope
// @Router       /admin/projects [post]
func (h *ProjectHandler) Create(c *gin.Context) {
	p, ok := bindProject(c)
	if !ok {
		return
	}
	p.ID = models.NewProjectID(p.Title)
	if err := h.repo.Create(c.Request.Context(), p); err != nil {
		response.ServerError(c, err)
		return
	}
	response.Created(c, p)
}

// Update godoc
// @Summary      Update a project
// @Description  Updates an existing project by ID (Admin only)
// @Tags         admin-projects
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        id path string true "Project ID"
// @Param        input body models.ProjectInput true "Project details"
// @Success      200 {object} models.Project
// @Failure      400 {object} response.ErrorEnvelope
// @Failure      401 {object} response.ErrorEnvelope
// @Failure      404 {object} response.ErrorEnvelope
// @Failure      422 {object} response.ValidationEnvelope
// @Failure      500 {object} response.ErrorEnvelope
// @Router       /admin/projects/{id} [put]
func (h *ProjectHandler) Update(c *gin.Context) {
	p, ok := bindProject(c)
	if !ok {
		return
	}
	p.ID = c.Param("id")
	if h.failed(c, h.repo.Update(c.Request.Context(), p)) {
		return
	}
	response.OK(c, p)
}

// Delete godoc
// @Summary      Delete a project
// @Description  Deletes a project by ID (Admin only)
// @Tags         admin-projects
// @Security     BearerAuth
// @Produce      json
// @Param        id path string true "Project ID"
// @Success      204 "Project deleted successfully"
// @Failure      401 {object} response.ErrorEnvelope
// @Failure      404 {object} response.ErrorEnvelope
// @Failure      500 {object} response.ErrorEnvelope
// @Router       /admin/projects/{id} [delete]
func (h *ProjectHandler) Delete(c *gin.Context) {
	if h.failed(c, h.repo.Delete(c.Request.Context(), c.Param("id"))) {
		return
	}
	response.NoContent(c)
}

// failed writes the right response and returns true when err is non-nil.
func (h *ProjectHandler) failed(c *gin.Context, err error) bool {
	switch {
	case err == nil:
		return false
	case errors.Is(err, repository.ErrNotFound):
		response.Error(c, http.StatusNotFound, "project not found")
	default:
		response.ServerError(c, err)
	}
	return true
}

func bindProject(c *gin.Context) (models.Project, bool) {
	var p models.Project
	if err := c.ShouldBindJSON(&p); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid JSON body")
		return p, false
	}
	p.Normalize()
	if errs := p.Validate(); len(errs) > 0 {
		response.Validation(c, errs)
		return p, false
	}
	return p, true
}
