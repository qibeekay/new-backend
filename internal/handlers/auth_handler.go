package handlers

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"portfolio-api/internal/response"
	"portfolio-api/internal/service"
)

type AuthHandler struct{ svc *service.AuthService }

func NewAuthHandler(svc *service.AuthService) *AuthHandler { return &AuthHandler{svc: svc} }

type LoginRequest struct {
	Email    string `json:"email" binding:"required" example:"admin@example.com"`
	Password string `json:"password" binding:"required" example:"admin123"`
}

type LoginResponse struct {
	Token     string    `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	ExpiresAt time.Time `json:"expiresAt"`
}

// Login godoc
// @Summary      Admin login
// @Description  Authenticates the administrator using email and password, returning a JWT token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        input body LoginRequest true "Admin login credentials"
// @Success      200 {object} LoginResponse
// @Failure      400 {object} response.ErrorEnvelope
// @Failure      401 {object} response.ErrorEnvelope
// @Failure      500 {object} response.ErrorEnvelope
// @Router       /admin/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var in LoginRequest
	if err := c.ShouldBindJSON(&in); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request body")
		return
	}

	token, exp, err := h.svc.Login(in.Email, in.Password)
	if errors.Is(err, service.ErrInvalidCredentials) {
		response.Error(c, http.StatusUnauthorized, "invalid credentials")
		return
	}
	if err != nil {
		response.ServerError(c, err)
		return
	}
	response.OK(c, LoginResponse{Token: token, ExpiresAt: exp})
}
