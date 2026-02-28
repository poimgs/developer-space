package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestJSON(t *testing.T) {
	w := httptest.NewRecorder()
	data := map[string]string{"status": "ok"}

	JSON(w, http.StatusOK, data)

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type = %q, want %q", ct, "application/json")
	}

	var result map[string]map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if result["data"]["status"] != "ok" {
		t.Errorf("data.status = %q, want %q", result["data"]["status"], "ok")
	}
}

func TestJSON_List(t *testing.T) {
	w := httptest.NewRecorder()
	data := []string{"a", "b"}

	JSON(w, http.StatusOK, data)

	var result map[string][]string
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if len(result["data"]) != 2 {
		t.Errorf("data length = %d, want 2", len(result["data"]))
	}
}

func TestJSON_Created(t *testing.T) {
	w := httptest.NewRecorder()
	JSON(w, http.StatusCreated, map[string]string{"id": "123"})

	if w.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d", w.Code, http.StatusCreated)
	}
}

func TestError(t *testing.T) {
	w := httptest.NewRecorder()
	Error(w, http.StatusNotFound, "Not found")

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
	}

	var result map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if result["error"] != "Not found" {
		t.Errorf("error = %q, want %q", result["error"], "Not found")
	}
}

func TestValidationError(t *testing.T) {
	w := httptest.NewRecorder()
	details := map[string]string{
		"email": "is required",
		"name":  "cannot be empty",
	}

	ValidationError(w, details)

	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, want %d", w.Code, http.StatusUnprocessableEntity)
	}

	var result struct {
		Error   string            `json:"error"`
		Details map[string]string `json:"details"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &result); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if result.Error != "Validation failed" {
		t.Errorf("error = %q, want %q", result.Error, "Validation failed")
	}
	if result.Details["email"] != "is required" {
		t.Errorf("details.email = %q, want %q", result.Details["email"], "is required")
	}
	if result.Details["name"] != "cannot be empty" {
		t.Errorf("details.name = %q, want %q", result.Details["name"], "cannot be empty")
	}
}
