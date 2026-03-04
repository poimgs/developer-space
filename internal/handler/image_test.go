package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/developer-space/api/internal/middleware"
	"github.com/developer-space/api/internal/model"
	"github.com/developer-space/api/internal/service"
)

// --- Mock session repo for image handler tests ---

type mockSessionRepoForImage struct {
	sessions map[uuid.UUID]*model.SpaceSession
}

func newMockSessionRepoForImage() *mockSessionRepoForImage {
	return &mockSessionRepoForImage{
		sessions: make(map[uuid.UUID]*model.SpaceSession),
	}
}

func (m *mockSessionRepoForImage) Create(ctx context.Context, req model.CreateSessionRequest, createdBy uuid.UUID) (*model.SpaceSession, error) {
	return nil, nil
}

func (m *mockSessionRepoForImage) CreateBatch(ctx context.Context, sessions []model.CreateSessionRequest, createdBy uuid.UUID) ([]model.SpaceSession, error) {
	return nil, nil
}

func (m *mockSessionRepoForImage) List(ctx context.Context, from, to, status string, memberID *uuid.UUID) ([]model.SpaceSession, error) {
	return []model.SpaceSession{}, nil
}

func (m *mockSessionRepoForImage) GetByID(ctx context.Context, id uuid.UUID, memberID *uuid.UUID) (*model.SpaceSession, error) {
	s, ok := m.sessions[id]
	if !ok {
		return nil, nil
	}
	return s, nil
}

func (m *mockSessionRepoForImage) Update(ctx context.Context, id uuid.UUID, req model.UpdateSessionRequest, newStatus *string) (*model.SpaceSession, error) {
	return nil, nil
}

func (m *mockSessionRepoForImage) Cancel(ctx context.Context, id uuid.UUID) (*model.SpaceSession, error) {
	return nil, nil
}

func (m *mockSessionRepoForImage) GetRSVPCount(ctx context.Context, sessionID uuid.UUID) (int, error) {
	return 0, nil
}

func (m *mockSessionRepoForImage) UpdateImageURL(ctx context.Context, id uuid.UUID, imageURL *string) (*model.SpaceSession, error) {
	s, ok := m.sessions[id]
	if !ok {
		return nil, nil
	}
	s.ImageURL = imageURL
	s.UpdatedAt = time.Now()
	return s, nil
}

func (m *mockSessionRepoForImage) ListFutureBySeriesID(ctx context.Context, seriesID uuid.UUID) ([]model.SpaceSession, error) {
	return []model.SpaceSession{}, nil
}
func (m *mockSessionRepoForImage) UpdateBulkBySeriesID(ctx context.Context, seriesID uuid.UUID, req model.UpdateSessionRequest, imageURL *string) (int64, error) {
	return 0, nil
}
func (m *mockSessionRepoForImage) CancelFutureBySeriesID(ctx context.Context, seriesID uuid.UUID) ([]model.SpaceSession, error) {
	return []model.SpaceSession{}, nil
}

func (m *mockSessionRepoForImage) addSession(id uuid.UUID) *model.SpaceSession {
	s := &model.SpaceSession{
		ID:        id,
		Title:     "Test Session",
		Date:      "2026-04-01",
		StartTime: "14:00",
		EndTime:   "18:00",
		Status:    "scheduled",
		CreatedBy: uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	m.sessions[id] = s
	return s
}

// --- Test helpers ---

func setupImageHandler(t *testing.T) (*ImageHandler, *mockSessionRepoForImage, string) {
	t.Helper()
	tmpDir := t.TempDir()
	repo := newMockSessionRepoForImage()
	notifier := &mockNotifierForImage{}
	svc := service.NewSessionService(repo, notifier)
	h := NewImageHandler(svc, tmpDir)
	return h, repo, tmpDir
}

func setupImageRouter(h *ImageHandler) *chi.Mux {
	r := chi.NewRouter()
	r.Post("/api/sessions/{id}/image", h.Upload)
	r.Delete("/api/sessions/{id}/image", h.Delete)
	return r
}

func setupImageRouterWithAuth(h *ImageHandler, authSvc *service.AuthService, memberRepo *mockMemberRepo) *chi.Mux {
	r := chi.NewRouter()
	r.Group(func(r chi.Router) {
		r.Use(middleware.Auth(authSvc, memberRepo))
		r.Use(middleware.Admin)
		r.Post("/api/sessions/{id}/image", h.Upload)
		r.Delete("/api/sessions/{id}/image", h.Delete)
	})
	return r
}

type mockNotifierForImage struct{}

func (n *mockNotifierForImage) SessionCreated(session *model.SpaceSession)                          {}
func (n *mockNotifierForImage) SessionsCreatedRecurring(sessions []model.SpaceSession)              {}
func (n *mockNotifierForImage) SessionShifted(session *model.SpaceSession)                          {}
func (n *mockNotifierForImage) SessionCanceled(session *model.SpaceSession)                         {}
func (n *mockNotifierForImage) MemberRSVPed(session *model.SpaceSession, member *model.Member)      {}
func (n *mockNotifierForImage) MemberCanceledRSVP(session *model.SpaceSession, member *model.Member) {}
func (n *mockNotifierForImage) SeriesUpdated(series *model.SessionSeries, affected []model.SpaceSession) {}
func (n *mockNotifierForImage) SeriesCanceled(series *model.SessionSeries, canceled []model.SpaceSession) {}

// createJPEGMultipart creates a multipart form body with a valid JPEG image.
func createJPEGMultipart(t *testing.T) (*bytes.Buffer, string) {
	t.Helper()
	var imgBuf bytes.Buffer
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	if err := jpeg.Encode(&imgBuf, img, nil); err != nil {
		t.Fatalf("failed to encode JPEG: %v", err)
	}
	return createMultipartFromBytes(t, imgBuf.Bytes(), "test.jpg")
}

// createPNGMultipart creates a multipart form body with a valid PNG image.
func createPNGMultipart(t *testing.T) (*bytes.Buffer, string) {
	t.Helper()
	var imgBuf bytes.Buffer
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	if err := png.Encode(&imgBuf, img); err != nil {
		t.Fatalf("failed to encode PNG: %v", err)
	}
	return createMultipartFromBytes(t, imgBuf.Bytes(), "test.png")
}

// createMultipartFromBytes builds a multipart form with the given bytes as the "image" field.
func createMultipartFromBytes(t *testing.T, data []byte, filename string) (*bytes.Buffer, string) {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("image", filename)
	if err != nil {
		t.Fatalf("failed to create form file: %v", err)
	}
	if _, err := io.Copy(part, bytes.NewReader(data)); err != nil {
		t.Fatalf("failed to write form file: %v", err)
	}
	writer.Close()
	return &body, writer.FormDataContentType()
}

// --- Upload Tests ---

func TestImageHandler_Upload_JPEG_200(t *testing.T) {
	h, repo, tmpDir := setupImageHandler(t)
	r := setupImageRouter(h)

	sessionID := uuid.New()
	repo.addSession(sessionID)

	body, contentType := createJPEGMultipart(t)
	req := httptest.NewRequest("POST", "/api/sessions/"+sessionID.String()+"/image", body)
	req.Header.Set("Content-Type", contentType)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Data model.SpaceSession `json:"data"`
	}
	json.Unmarshal(rec.Body.Bytes(), &resp)

	if resp.Data.ImageURL == nil {
		t.Fatal("expected image_url to be set")
	}
	if resp.Data.ID != sessionID {
		t.Errorf("expected session ID %s, got %s", sessionID, resp.Data.ID)
	}

	// Verify file was written to disk
	files, _ := filepath.Glob(filepath.Join(tmpDir, sessionID.String()+"*"))
	if len(files) != 1 {
		t.Errorf("expected 1 file in uploads dir, got %d", len(files))
	}
}

func TestImageHandler_Upload_PNG_200(t *testing.T) {
	h, repo, tmpDir := setupImageHandler(t)
	r := setupImageRouter(h)

	sessionID := uuid.New()
	repo.addSession(sessionID)

	body, contentType := createPNGMultipart(t)
	req := httptest.NewRequest("POST", "/api/sessions/"+sessionID.String()+"/image", body)
	req.Header.Set("Content-Type", contentType)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	// Verify file has .png extension
	files, _ := filepath.Glob(filepath.Join(tmpDir, sessionID.String()+"*.png"))
	if len(files) != 1 {
		t.Errorf("expected 1 PNG file, got %d", len(files))
	}
}

func TestImageHandler_Upload_InvalidType_422(t *testing.T) {
	h, repo, _ := setupImageHandler(t)
	r := setupImageRouter(h)

	sessionID := uuid.New()
	repo.addSession(sessionID)

	// Create multipart with a text file
	body, contentType := createMultipartFromBytes(t, []byte("this is not an image"), "test.txt")
	req := httptest.NewRequest("POST", "/api/sessions/"+sessionID.String()+"/image", body)
	req.Header.Set("Content-Type", contentType)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Error string `json:"error"`
	}
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if resp.Error != "Invalid file type (must be JPEG, PNG, or WebP)" {
		t.Errorf("unexpected error message: %s", resp.Error)
	}
}

func TestImageHandler_Upload_SessionNotFound_404(t *testing.T) {
	h, _, _ := setupImageHandler(t)
	r := setupImageRouter(h)

	body, contentType := createJPEGMultipart(t)
	req := httptest.NewRequest("POST", "/api/sessions/"+uuid.New().String()+"/image", body)
	req.Header.Set("Content-Type", contentType)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestImageHandler_Upload_InvalidSessionID_400(t *testing.T) {
	h, _, _ := setupImageHandler(t)
	r := setupImageRouter(h)

	body, contentType := createJPEGMultipart(t)
	req := httptest.NewRequest("POST", "/api/sessions/not-a-uuid/image", body)
	req.Header.Set("Content-Type", contentType)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestImageHandler_Upload_MissingFile_400(t *testing.T) {
	h, repo, _ := setupImageHandler(t)
	r := setupImageRouter(h)

	sessionID := uuid.New()
	repo.addSession(sessionID)

	// Send empty multipart form without the "image" field
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	writer.Close()

	req := httptest.NewRequest("POST", "/api/sessions/"+sessionID.String()+"/image", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestImageHandler_Upload_ReplacesExistingImage(t *testing.T) {
	h, repo, tmpDir := setupImageHandler(t)
	r := setupImageRouter(h)

	sessionID := uuid.New()
	s := repo.addSession(sessionID)

	// First upload
	body1, contentType1 := createJPEGMultipart(t)
	req1 := httptest.NewRequest("POST", "/api/sessions/"+sessionID.String()+"/image", body1)
	req1.Header.Set("Content-Type", contentType1)
	rec1 := httptest.NewRecorder()
	r.ServeHTTP(rec1, req1)

	if rec1.Code != http.StatusOK {
		t.Fatalf("first upload: expected 200, got %d: %s", rec1.Code, rec1.Body.String())
	}

	// Get the old image URL
	oldImageURL := *s.ImageURL

	// Wait 1 second so timestamp differs
	time.Sleep(1100 * time.Millisecond)

	// Second upload (replace)
	body2, contentType2 := createPNGMultipart(t)
	req2 := httptest.NewRequest("POST", "/api/sessions/"+sessionID.String()+"/image", body2)
	req2.Header.Set("Content-Type", contentType2)
	rec2 := httptest.NewRecorder()
	r.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusOK {
		t.Fatalf("second upload: expected 200, got %d: %s", rec2.Code, rec2.Body.String())
	}

	// Verify new image URL is different from old
	newImageURL := *s.ImageURL
	if newImageURL == oldImageURL {
		t.Error("expected new image URL to differ from old")
	}

	// Verify old file was deleted — extract filename from URL
	oldFilename := filepath.Base(oldImageURL)
	if _, err := os.Stat(filepath.Join(tmpDir, oldFilename)); !os.IsNotExist(err) {
		t.Error("expected old image file to be deleted")
	}

	// Verify new file exists
	newFilename := filepath.Base(newImageURL)
	if _, err := os.Stat(filepath.Join(tmpDir, newFilename)); err != nil {
		t.Errorf("expected new image file to exist: %v", err)
	}
}

func TestImageHandler_Upload_FileTooLarge_413(t *testing.T) {
	h, repo, _ := setupImageHandler(t)
	r := setupImageRouter(h)

	sessionID := uuid.New()
	repo.addSession(sessionID)

	// Create a file larger than 5MB
	largeData := make([]byte, 6*1024*1024)
	// Write JPEG magic bytes so it passes type check but fails size check
	copy(largeData, []byte{0xFF, 0xD8, 0xFF, 0xE0})

	body, contentType := createMultipartFromBytes(t, largeData, "large.jpg")
	req := httptest.NewRequest("POST", "/api/sessions/"+sessionID.String()+"/image", body)
	req.Header.Set("Content-Type", contentType)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	// Should get 413 from MaxBytesReader or 400 from ParseMultipartForm
	if rec.Code != http.StatusRequestEntityTooLarge && rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 413 or 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

// --- Delete Tests ---

func TestImageHandler_Delete_200(t *testing.T) {
	h, repo, tmpDir := setupImageHandler(t)
	r := setupImageRouter(h)

	sessionID := uuid.New()
	s := repo.addSession(sessionID)

	// Create a fake image file
	imageURL := fmt.Sprintf("/uploads/sessions/%s-12345.jpg", sessionID.String())
	s.ImageURL = &imageURL
	filename := fmt.Sprintf("%s-12345.jpg", sessionID.String())
	os.WriteFile(filepath.Join(tmpDir, filename), []byte("fake image"), 0644)

	req := httptest.NewRequest("DELETE", "/api/sessions/"+sessionID.String()+"/image", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Data model.SpaceSession `json:"data"`
	}
	json.Unmarshal(rec.Body.Bytes(), &resp)

	if resp.Data.ImageURL != nil {
		t.Errorf("expected image_url to be nil after delete, got %s", *resp.Data.ImageURL)
	}

	// Verify file was deleted from disk
	if _, err := os.Stat(filepath.Join(tmpDir, filename)); !os.IsNotExist(err) {
		t.Error("expected image file to be deleted from disk")
	}
}

func TestImageHandler_Delete_Idempotent_200(t *testing.T) {
	h, repo, _ := setupImageHandler(t)
	r := setupImageRouter(h)

	sessionID := uuid.New()
	repo.addSession(sessionID) // No image set

	req := httptest.NewRequest("DELETE", "/api/sessions/"+sessionID.String()+"/image", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 for idempotent delete, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestImageHandler_Delete_SessionNotFound_404(t *testing.T) {
	h, _, _ := setupImageHandler(t)
	r := setupImageRouter(h)

	req := httptest.NewRequest("DELETE", "/api/sessions/"+uuid.New().String()+"/image", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestImageHandler_Delete_InvalidSessionID_400(t *testing.T) {
	h, _, _ := setupImageHandler(t)
	r := setupImageRouter(h)

	req := httptest.NewRequest("DELETE", "/api/sessions/not-a-uuid/image", nil)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

// --- Admin middleware integration test ---

func TestImageHandler_Upload_NonAdmin_403(t *testing.T) {
	tmpDir := t.TempDir()
	tokenRepo := &mockTokenRepoForHandler{}
	memberRepo := newMockRepo()
	emailSender := &mockMagicLinkSenderForHandler{}
	authSvc := service.NewAuthService(tokenRepo, memberRepo, emailSender, "test-secret", "http://localhost:5173", false)

	sessionRepo := newMockSessionRepoForImage()
	notifier := &mockNotifierForImage{}
	sessionSvc := service.NewSessionService(sessionRepo, notifier)
	imageHandler := NewImageHandler(sessionSvc, tmpDir)

	r := setupImageRouterWithAuth(imageHandler, authSvc, memberRepo)

	// Create a non-admin member
	memberID := uuid.New()
	memberRepo.members[memberID] = &model.Member{
		ID:       memberID,
		Email:    "user@example.com",
		Name:     "Regular User",
		IsAdmin:  false,
		IsActive: true,
	}

	sessionID := uuid.New()
	sessionRepo.addSession(sessionID)

	cookie, _ := authSvc.CreateSessionCookie(memberID)

	body, contentType := createJPEGMultipart(t)
	req := httptest.NewRequest("POST", "/api/sessions/"+sessionID.String()+"/image", body)
	req.Header.Set("Content-Type", contentType)
	req.AddCookie(cookie)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestImageHandler_Upload_Unauthenticated_401(t *testing.T) {
	tmpDir := t.TempDir()
	tokenRepo := &mockTokenRepoForHandler{}
	memberRepo := newMockRepo()
	emailSender := &mockMagicLinkSenderForHandler{}
	authSvc := service.NewAuthService(tokenRepo, memberRepo, emailSender, "test-secret", "http://localhost:5173", false)

	sessionRepo := newMockSessionRepoForImage()
	notifier := &mockNotifierForImage{}
	sessionSvc := service.NewSessionService(sessionRepo, notifier)
	imageHandler := NewImageHandler(sessionSvc, tmpDir)

	r := setupImageRouterWithAuth(imageHandler, authSvc, memberRepo)

	sessionID := uuid.New()
	sessionRepo.addSession(sessionID)

	body, contentType := createJPEGMultipart(t)
	req := httptest.NewRequest("POST", "/api/sessions/"+sessionID.String()+"/image", body)
	req.Header.Set("Content-Type", contentType)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestImageHandler_Upload_ImageURLFormat(t *testing.T) {
	h, repo, _ := setupImageHandler(t)
	r := setupImageRouter(h)

	sessionID := uuid.New()
	repo.addSession(sessionID)

	body, contentType := createJPEGMultipart(t)
	req := httptest.NewRequest("POST", "/api/sessions/"+sessionID.String()+"/image", body)
	req.Header.Set("Content-Type", contentType)
	rec := httptest.NewRecorder()

	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp struct {
		Data model.SpaceSession `json:"data"`
	}
	json.Unmarshal(rec.Body.Bytes(), &resp)

	imageURL := *resp.Data.ImageURL
	// Should start with /uploads/sessions/ and contain the session ID
	expectedPrefix := "/uploads/sessions/" + sessionID.String()
	if len(imageURL) <= len(expectedPrefix) || imageURL[:len(expectedPrefix)] != expectedPrefix {
		t.Errorf("expected image URL to start with %q, got %q", expectedPrefix, imageURL)
	}
}
