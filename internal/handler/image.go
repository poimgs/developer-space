package handler

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/developer-space/api/internal/response"
	"github.com/developer-space/api/internal/service"
)

const (
	maxImageSize = 5 << 20 // 5MB
	uploadDir    = "uploads/sessions"
)

var allowedContentTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
}

type ImageHandler struct {
	svc       *service.SessionService
	uploadDir string
}

func NewImageHandler(svc *service.SessionService, uploadDir string) *ImageHandler {
	return &ImageHandler{svc: svc, uploadDir: uploadDir}
}

func (h *ImageHandler) Upload(w http.ResponseWriter, r *http.Request) {
	sessionID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid session ID")
		return
	}

	// Check session exists before processing file
	oldImageURL, err := h.svc.GetImageURL(r.Context(), sessionID)
	if err != nil {
		if isNotFound(err) {
			response.Error(w, http.StatusNotFound, "Session not found")
			return
		}
		slog.Error("failed to get session", "error", err)
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	// Limit request body to maxImageSize + 1KB overhead for multipart headers
	r.Body = http.MaxBytesReader(w, r.Body, maxImageSize+1024)

	if err := r.ParseMultipartForm(maxImageSize); err != nil {
		if err.Error() == "http: request body too large" {
			response.Error(w, http.StatusRequestEntityTooLarge, "File too large (max 5MB)")
			return
		}
		response.Error(w, http.StatusBadRequest, "Invalid multipart form")
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Missing image file")
		return
	}
	defer file.Close()

	if header.Size > maxImageSize {
		response.Error(w, http.StatusRequestEntityTooLarge, "File too large (max 5MB)")
		return
	}

	// Read first 512 bytes for magic byte detection
	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		slog.Error("failed to read file header", "error", err)
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	detectedType := http.DetectContentType(buf[:n])
	ext, ok := allowedContentTypes[detectedType]
	if !ok {
		response.Error(w, http.StatusUnprocessableEntity, "Invalid file type (must be JPEG, PNG, or WebP)")
		return
	}

	// Seek back to start
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		slog.Error("failed to seek file", "error", err)
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	// Generate server-side filename
	filename := fmt.Sprintf("%s-%d%s", sessionID.String(), time.Now().Unix(), ext)
	filePath := filepath.Join(h.uploadDir, filename)

	// Save file
	dst, err := os.Create(filePath)
	if err != nil {
		slog.Error("failed to create file", "error", err, "path", filePath)
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		slog.Error("failed to write file", "error", err)
		os.Remove(filePath)
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	// Update database
	imageURL := "/" + uploadDir + "/" + filename
	session, err := h.svc.UpdateImageURL(r.Context(), sessionID, imageURL)
	if err != nil {
		os.Remove(filePath)
		slog.Error("failed to update image url", "error", err)
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	// Delete old image file if one existed
	if oldImageURL != "" {
		oldPath := h.urlToPath(oldImageURL)
		if oldPath != "" {
			os.Remove(oldPath)
		}
	}

	response.JSON(w, http.StatusOK, session)
}

func (h *ImageHandler) Delete(w http.ResponseWriter, r *http.Request) {
	sessionID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid session ID")
		return
	}

	// Get current image URL before clearing
	oldImageURL, err := h.svc.GetImageURL(r.Context(), sessionID)
	if err != nil {
		if isNotFound(err) {
			response.Error(w, http.StatusNotFound, "Session not found")
			return
		}
		slog.Error("failed to get session", "error", err)
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	// If no image, return success (idempotent)
	if oldImageURL == "" {
		session, err := h.svc.GetByID(r.Context(), sessionID, nil)
		if err != nil {
			slog.Error("failed to get session", "error", err)
			response.Error(w, http.StatusInternalServerError, "Internal server error")
			return
		}
		response.JSON(w, http.StatusOK, session)
		return
	}

	// Clear in database
	session, err := h.svc.ClearImageURL(r.Context(), sessionID)
	if err != nil {
		slog.Error("failed to clear image url", "error", err)
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	// Delete file from disk
	oldPath := h.urlToPath(oldImageURL)
	if oldPath != "" {
		if err := os.Remove(oldPath); err != nil && !os.IsNotExist(err) {
			slog.Error("failed to delete image file", "error", err, "path", oldPath)
		}
	}

	response.JSON(w, http.StatusOK, session)
}

// urlToPath converts an image URL like "/uploads/sessions/abc.jpg" to a filesystem path.
func (h *ImageHandler) urlToPath(imageURL string) string {
	// URL format: /uploads/sessions/filename.ext
	prefix := "/" + uploadDir + "/"
	if !strings.HasPrefix(imageURL, prefix) {
		return ""
	}
	filename := strings.TrimPrefix(imageURL, prefix)
	// Prevent path traversal
	if strings.Contains(filename, "/") || strings.Contains(filename, "..") {
		return ""
	}
	return filepath.Join(h.uploadDir, filename)
}

func (h *ImageHandler) UploadSeriesImage(w http.ResponseWriter, r *http.Request) {
	seriesID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid series ID")
		return
	}

	oldImageURL, err := h.svc.GetSeriesImageURL(r.Context(), seriesID)
	if err != nil {
		if errors.Is(err, service.ErrSeriesNotFound) {
			response.Error(w, http.StatusNotFound, "Series not found")
			return
		}
		slog.Error("failed to get series", "error", err)
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxImageSize+1024)

	if err := r.ParseMultipartForm(maxImageSize); err != nil {
		if err.Error() == "http: request body too large" {
			response.Error(w, http.StatusRequestEntityTooLarge, "File too large (max 5MB)")
			return
		}
		response.Error(w, http.StatusBadRequest, "Invalid multipart form")
		return
	}

	file, header, err := r.FormFile("image")
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Missing image file")
		return
	}
	defer file.Close()

	if header.Size > maxImageSize {
		response.Error(w, http.StatusRequestEntityTooLarge, "File too large (max 5MB)")
		return
	}

	buf := make([]byte, 512)
	n, err := file.Read(buf)
	if err != nil && err != io.EOF {
		slog.Error("failed to read file header", "error", err)
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	detectedType := http.DetectContentType(buf[:n])
	ext, ok := allowedContentTypes[detectedType]
	if !ok {
		response.Error(w, http.StatusUnprocessableEntity, "Invalid file type (must be JPEG, PNG, or WebP)")
		return
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		slog.Error("failed to seek file", "error", err)
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	filename := fmt.Sprintf("series-%s-%d%s", seriesID.String(), time.Now().Unix(), ext)
	filePath := filepath.Join(h.uploadDir, filename)

	dst, err := os.Create(filePath)
	if err != nil {
		slog.Error("failed to create file", "error", err, "path", filePath)
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		slog.Error("failed to write file", "error", err)
		os.Remove(filePath)
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	imageURL := "/" + uploadDir + "/" + filename
	series, err := h.svc.UpdateSeriesImageURL(r.Context(), seriesID, imageURL)
	if err != nil {
		os.Remove(filePath)
		slog.Error("failed to update series image url", "error", err)
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	if oldImageURL != "" {
		oldPath := h.urlToPath(oldImageURL)
		if oldPath != "" {
			os.Remove(oldPath)
		}
	}

	response.JSON(w, http.StatusOK, series)
}

func (h *ImageHandler) DeleteSeriesImage(w http.ResponseWriter, r *http.Request) {
	seriesID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid series ID")
		return
	}

	oldImageURL, err := h.svc.GetSeriesImageURL(r.Context(), seriesID)
	if err != nil {
		if errors.Is(err, service.ErrSeriesNotFound) {
			response.Error(w, http.StatusNotFound, "Series not found")
			return
		}
		slog.Error("failed to get series", "error", err)
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	if oldImageURL == "" {
		series, err := h.svc.GetSeriesByID(r.Context(), seriesID)
		if err != nil {
			slog.Error("failed to get series", "error", err)
			response.Error(w, http.StatusInternalServerError, "Internal server error")
			return
		}
		response.JSON(w, http.StatusOK, series)
		return
	}

	series, err := h.svc.ClearSeriesImageURL(r.Context(), seriesID)
	if err != nil {
		slog.Error("failed to clear series image url", "error", err)
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}

	oldPath := h.urlToPath(oldImageURL)
	if oldPath != "" {
		if err := os.Remove(oldPath); err != nil && !os.IsNotExist(err) {
			slog.Error("failed to delete image file", "error", err, "path", oldPath)
		}
	}

	response.JSON(w, http.StatusOK, series)
}

func isNotFound(err error) bool {
	return errors.Is(err, service.ErrSessionNotFound)
}
