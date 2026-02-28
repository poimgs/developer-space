package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequestID_SetsHeader(t *testing.T) {
	handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	id := w.Header().Get("X-Request-Id")
	if id == "" {
		t.Fatal("X-Request-Id header is empty")
	}
	if len(id) != 36 { // UUID v4 format
		t.Errorf("X-Request-Id length = %d, want 36", len(id))
	}
}

func TestRequestID_UniquePerRequest(t *testing.T) {
	handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	w1 := httptest.NewRecorder()
	w2 := httptest.NewRecorder()

	handler.ServeHTTP(w1, httptest.NewRequest("GET", "/test", nil))
	handler.ServeHTTP(w2, httptest.NewRequest("GET", "/test", nil))

	id1 := w1.Header().Get("X-Request-Id")
	id2 := w2.Header().Get("X-Request-Id")

	if id1 == id2 {
		t.Error("expected unique request IDs, got identical ones")
	}
}

func TestRequestID_AvailableInContext(t *testing.T) {
	var ctxID string
	handler := RequestID(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctxID = GetRequestID(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	w := httptest.NewRecorder()
	handler.ServeHTTP(w, httptest.NewRequest("GET", "/test", nil))

	headerID := w.Header().Get("X-Request-Id")
	if ctxID != headerID {
		t.Errorf("context ID = %q, header ID = %q, want equal", ctxID, headerID)
	}
}

func TestGetRequestID_EmptyContext(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	id := GetRequestID(req.Context())
	if id != "" {
		t.Errorf("GetRequestID from empty context = %q, want empty", id)
	}
}
