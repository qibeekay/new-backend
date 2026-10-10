package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                string
	DatabaseURL         string
	JWTSecret           string
	AdminEmail          string
	AdminPasswordHash   string
	UploadDir           string
	PublicBaseURL       string
	AllowedOrigins      []string
	MaxUploadBytes      int64
	CloudinaryCloudName string
	CloudinaryAPIKey    string
	CloudinaryAPISecret string
	CloudinaryURL       string
	CloudinaryFolder    string
	KeepAliveURL        string
}

func (c *Config) HasCloudinary() bool {
	return (c.CloudinaryCloudName != "" && c.CloudinaryAPIKey != "" && c.CloudinaryAPISecret != "") || c.CloudinaryURL != ""
}

func Load() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		_ = godotenv.Load("../../.env", "../.env")
	}

	mb, err := strconv.ParseInt(getEnv("MAX_UPLOAD_MB", "5"), 10, 64)
	if err != nil || mb <= 0 {
		return nil, fmt.Errorf("MAX_UPLOAD_MB must be a positive integer")
	}

	publicBase := os.Getenv("PUBLIC_BASE_URL")
	if publicBase == "" && os.Getenv("RENDER_EXTERNAL_URL") != "" {
		publicBase = os.Getenv("RENDER_EXTERNAL_URL")
	}
	if publicBase == "" {
		publicBase = "http://localhost:8080"
	}

	keepAlive := os.Getenv("KEEP_ALIVE_URL")
	if keepAlive == "" {
		keepAlive = os.Getenv("RENDER_EXTERNAL_URL")
	}

	cfg := &Config{
		Port:                getEnv("PORT", "8080"),
		DatabaseURL:         os.Getenv("DATABASE_URL"),
		JWTSecret:           os.Getenv("JWT_SECRET"),
		AdminEmail:          os.Getenv("ADMIN_EMAIL"),
		AdminPasswordHash:   os.Getenv("ADMIN_PASSWORD_HASH"),
		UploadDir:           getEnv("UPLOAD_DIR", "./uploads"),
		PublicBaseURL:       strings.TrimRight(publicBase, "/"),
		AllowedOrigins:      strings.Split(getEnv("ALLOWED_ORIGINS", "http://localhost:5173"), ","),
		MaxUploadBytes:      mb << 20,
		CloudinaryCloudName: os.Getenv("CLOUDINARY_CLOUD_NAME"),
		CloudinaryAPIKey:    os.Getenv("CLOUDINARY_API_KEY"),
		CloudinaryAPISecret: os.Getenv("CLOUDINARY_API_SECRET"),
		CloudinaryURL:       os.Getenv("CLOUDINARY_URL"),
		CloudinaryFolder:    getEnv("CLOUDINARY_FOLDER", "portfolio"),
		KeepAliveURL:        keepAlive,
	}

	required := map[string]string{
		"DATABASE_URL": cfg.DatabaseURL, "JWT_SECRET": cfg.JWTSecret,
		"ADMIN_EMAIL": cfg.AdminEmail, "ADMIN_PASSWORD_HASH": cfg.AdminPasswordHash,
	}
	for name, v := range required {
		if v == "" {
			return nil, fmt.Errorf("%s is required", name)
		}
	}
	if len(cfg.JWTSecret) < 32 {
		return nil, fmt.Errorf("JWT_SECRET must be at least 32 characters")
	}
	return cfg, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
