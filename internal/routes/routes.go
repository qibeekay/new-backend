package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "portfolio-api/docs"
	"portfolio-api/internal/config"
	"portfolio-api/internal/handlers"
	"portfolio-api/internal/middleware"
	"portfolio-api/internal/repository"
	"portfolio-api/internal/service"
)

func Setup(r *gin.Engine, cfg *config.Config, repo *repository.Repository) {
	authHandler := handlers.NewAuthHandler(service.NewAuthService(cfg))
	projectHandler := handlers.NewProjectHandler(repo.Projects)
	uploadHandler := handlers.NewUploadHandler(cfg)

	r.GET("/healthz", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"ok": true}) })
	r.Static("/uploads", cfg.UploadDir)

	// Swagger documentation
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	r.GET("/swagger", func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/swagger/index.html")
	})

	api := r.Group("/api")

	// Public: what the portfolio site reads.
	api.GET("/projects", projectHandler.List)
	api.GET("/projects/:id", projectHandler.Get)

	// Admin.
	admin := api.Group("/admin")
	admin.POST("/login", authHandler.Login)

	protected := admin.Group("", middleware.Auth(cfg.JWTSecret))
	protected.POST("/projects", projectHandler.Create)
	protected.PUT("/projects/:id", projectHandler.Update)
	protected.DELETE("/projects/:id", projectHandler.Delete)
	protected.POST("/uploads", uploadHandler.Upload)
}
