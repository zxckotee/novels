package handlers

import (
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"novels-backend/pkg/response"
)

// UploadHandler handles authenticated user uploads (e.g., proposal cover images).
type UploadHandler struct {
	uploadDir string
}

func NewUploadHandler(uploadDir string) *UploadHandler {
	return &UploadHandler{uploadDir: uploadDir}
}

// Upload uploads a single image file and returns a public URL under /uploads/*.
// POST /api/v1/upload (multipart/form-data, field: "file")
func (h *UploadHandler) Upload(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form (10MB max memory; actual body may be larger but we enforce by server/proxy too)
	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10MB
		response.BadRequest(w, "file too large")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		response.BadRequest(w, "file is required")
		return
	}
	defer file.Close()

	ext := filepath.Ext(header.Filename)
	allowedExts := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".webp": true}
	if !allowedExts[ext] {
		response.BadRequest(w, "only jpg, jpeg, png, webp files are allowed")
		return
	}

	uploadsDir := filepath.Join(h.uploadDir, "proposal_covers")
	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		response.InternalError(w)
		return
	}

	filename := uuid.New().String() + ext
	filePath := filepath.Join(uploadsDir, filename)

	dst, err := os.Create(filePath)
	if err != nil {
		response.InternalError(w)
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		response.InternalError(w)
		return
	}

	fileKey := "proposal_covers/" + filename
	response.OK(w, map[string]string{
		"message":  "file uploaded",
		"file_url": "/uploads/" + fileKey,
		"file_key": fileKey,
	})
}

