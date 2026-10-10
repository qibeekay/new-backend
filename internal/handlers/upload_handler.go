package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"github.com/gin-gonic/gin"

	"portfolio-api/internal/config"
	"portfolio-api/internal/response"
)

type UploadHandler struct {
	cfg *config.Config
	cld *cloudinary.Cloudinary
}

func NewUploadHandler(cfg *config.Config) *UploadHandler {
	var cld *cloudinary.Cloudinary
	if cfg.HasCloudinary() {
		var err error
		if cfg.CloudinaryURL != "" {
			cld, err = cloudinary.NewFromURL(cfg.CloudinaryURL)
		} else {
			cld, err = cloudinary.NewFromParams(cfg.CloudinaryCloudName, cfg.CloudinaryAPIKey, cfg.CloudinaryAPISecret)
		}
		if err != nil {
			log.Printf("warning: cloudinary init error: %v, falling back to local file storage", err)
		}
	}
	return &UploadHandler{cfg: cfg, cld: cld}
}

var allowedImageTypes = map[string]string{
	"image/jpeg": ".jpg", "image/png": ".png", "image/webp": ".webp", "image/gif": ".gif",
}

type UploadResponse struct {
	URL string `json:"url" example:"https://res.cloudinary.com/demo/image/upload/sample.jpg"`
}

// Upload godoc
// @Summary      Upload an image asset
// @Description  Uploads an image file (JPEG, PNG, WebP, GIF) to Cloudinary (or local storage fallback) and returns the public asset URL (Admin only)
// @Tags         admin-uploads
// @Security     BearerAuth
// @Accept       multipart/form-data
// @Produce      json
// @Param        file formData file true "Image file to upload (JPEG, PNG, WebP, GIF, up to 5MB)"
// @Success      201 {object} UploadResponse
// @Failure      400 {object} response.ErrorEnvelope
// @Failure      401 {object} response.ErrorEnvelope
// @Failure      413 {object} response.ErrorEnvelope
// @Failure      415 {object} response.ErrorEnvelope
// @Failure      500 {object} response.ErrorEnvelope
// @Router       /admin/uploads [post]
func (h *UploadHandler) Upload(c *gin.Context) {
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, h.cfg.MaxUploadBytes+(1<<20))
	fh, err := c.FormFile("file")
	if err != nil {
		response.Error(c, http.StatusBadRequest, `attach an image as form field "file"`)
		return
	}
	if fh.Size > h.cfg.MaxUploadBytes {
		response.Error(c, http.StatusRequestEntityTooLarge, "file is too large")
		return
	}
	f, err := fh.Open()
	if err != nil {
		response.ServerError(c, err)
		return
	}
	defer f.Close()

	// Trust the bytes, not the filename or client Content-Type. SVG is excluded on purpose (script risk).
	head := make([]byte, 512)
	n, _ := io.ReadFull(f, head)
	ext, ok := allowedImageTypes[http.DetectContentType(head[:n])]
	if !ok {
		response.Error(c, http.StatusUnsupportedMediaType, "only JPEG, PNG, WebP or GIF images are allowed")
		return
	}
	if _, err := f.Seek(0, io.SeekStart); err != nil {
		response.ServerError(c, err)
		return
	}

	// 1. Upload to Cloudinary if configured
	if h.cld != nil {
		uploadRes, err := h.cld.Upload.Upload(c.Request.Context(), f, uploader.UploadParams{
			Folder: h.cfg.CloudinaryFolder,
		})
		if err != nil {
			response.ServerError(c, err)
			return
		}
		response.Created(c, UploadResponse{URL: uploadRes.SecureURL})
		return
	}

	// 2. Fall back to local file storage
	name := randomName() + ext
	dst, err := os.Create(filepath.Join(h.cfg.UploadDir, name))
	if err != nil {
		response.ServerError(c, err)
		return
	}
	defer dst.Close()
	if _, err := io.Copy(dst, f); err != nil {
		response.ServerError(c, err)
		return
	}
	response.Created(c, UploadResponse{URL: h.cfg.PublicBaseURL + "/uploads/" + name})
}

func randomName() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
