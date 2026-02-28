package database

import (
	"testing"
)

func TestPgxURL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "converts postgres:// to pgx5://",
			input:    "postgres://user:pass@localhost:5432/db?sslmode=disable",
			expected: "pgx5://user:pass@localhost:5432/db?sslmode=disable",
		},
		{
			name:     "converts postgresql:// to pgx5://",
			input:    "postgresql://user:pass@localhost:5432/db",
			expected: "pgx5://user:pass@localhost:5432/db",
		},
		{
			name:     "passes through already pgx5:// URL",
			input:    "pgx5://user:pass@localhost:5432/db",
			expected: "pgx5://user:pass@localhost:5432/db",
		},
		{
			name:     "passes through unknown scheme",
			input:    "mysql://user:pass@localhost:3306/db",
			expected: "mysql://user:pass@localhost:3306/db",
		},
		{
			name:     "handles empty string",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := pgxURL(tt.input)
			if got != tt.expected {
				t.Errorf("pgxURL(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}
