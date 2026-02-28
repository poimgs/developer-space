package config

import (
	"os"
	"testing"
)

func setRequiredEnv(t *testing.T) {
	t.Helper()
	t.Setenv("DATABASE_URL", "postgres://user:pass@localhost:5432/testdb")
	t.Setenv("SESSION_SECRET", "test-secret")
	t.Setenv("RESEND_API_KEY", "re_test")
	t.Setenv("RESEND_FROM_EMAIL", "test@example.com")
}

func TestLoad_Defaults(t *testing.T) {
	setRequiredEnv(t)

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Port != "8080" {
		t.Errorf("Port = %q, want %q", cfg.Port, "8080")
	}
	if cfg.FrontendURL != "http://localhost:5173" {
		t.Errorf("FrontendURL = %q, want %q", cfg.FrontendURL, "http://localhost:5173")
	}
	if cfg.LogLevel != "info" {
		t.Errorf("LogLevel = %q, want %q", cfg.LogLevel, "info")
	}
}

func TestLoad_CustomValues(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("PORT", "9090")
	t.Setenv("FRONTEND_URL", "https://app.example.com")
	t.Setenv("LOG_LEVEL", "DEBUG")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Port != "9090" {
		t.Errorf("Port = %q, want %q", cfg.Port, "9090")
	}
	if cfg.FrontendURL != "https://app.example.com" {
		t.Errorf("FrontendURL = %q, want %q", cfg.FrontendURL, "https://app.example.com")
	}
	if cfg.LogLevel != "debug" {
		t.Errorf("LogLevel = %q, want %q", cfg.LogLevel, "debug")
	}
}

func TestLoad_MissingDatabaseURL(t *testing.T) {
	t.Setenv("SESSION_SECRET", "test-secret")
	t.Setenv("RESEND_API_KEY", "re_test")
	t.Setenv("RESEND_FROM_EMAIL", "test@example.com")
	os.Unsetenv("DATABASE_URL")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for missing DATABASE_URL")
	}
}

func TestLoad_MissingSessionSecret(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://localhost/testdb")
	t.Setenv("RESEND_API_KEY", "re_test")
	t.Setenv("RESEND_FROM_EMAIL", "test@example.com")
	os.Unsetenv("SESSION_SECRET")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for missing SESSION_SECRET")
	}
}

func TestLoad_MissingResendAPIKey(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://localhost/testdb")
	t.Setenv("SESSION_SECRET", "test-secret")
	t.Setenv("RESEND_FROM_EMAIL", "test@example.com")
	os.Unsetenv("RESEND_API_KEY")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for missing RESEND_API_KEY")
	}
}

func TestLoad_MissingResendFromEmail(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://localhost/testdb")
	t.Setenv("SESSION_SECRET", "test-secret")
	t.Setenv("RESEND_API_KEY", "re_test")
	os.Unsetenv("RESEND_FROM_EMAIL")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for missing RESEND_FROM_EMAIL")
	}
}

func TestIsSecure(t *testing.T) {
	tests := []struct {
		frontendURL string
		want        bool
	}{
		{"https://app.example.com", true},
		{"http://localhost:5173", false},
		{"http://app.example.com", false},
	}

	for _, tt := range tests {
		cfg := &Config{FrontendURL: tt.frontendURL}
		if got := cfg.IsSecure(); got != tt.want {
			t.Errorf("IsSecure() for %q = %v, want %v", tt.frontendURL, got, tt.want)
		}
	}
}
