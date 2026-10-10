package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ErrorEnvelope standard error payload
type ErrorEnvelope struct {
	Error string `json:"error" example:"invalid request body"`
}

// ValidationEnvelope standard validation error payload
type ValidationEnvelope struct {
	Errors map[string]string `json:"errors"`
}

func OK(c *gin.Context, data any)      { c.JSON(http.StatusOK, data) }
func Created(c *gin.Context, data any) { c.JSON(http.StatusCreated, data) }
func NoContent(c *gin.Context)         { c.Status(http.StatusNoContent) }

func Error(c *gin.Context, status int, msg string) {
	c.JSON(status, gin.H{"error": msg})
}

func Validation(c *gin.Context, errs map[string]string) {
	c.JSON(http.StatusUnprocessableEntity, gin.H{"errors": errs})
}

func ServerError(c *gin.Context, err error) {
	_ = c.Error(err) // surfaces in gin's logger
	Error(c, http.StatusInternalServerError, "internal error")
}
