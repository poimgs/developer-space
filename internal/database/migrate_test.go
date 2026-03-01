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
			expected: "pgx5://localhost:5432",
		},
		{
			name:     "converts keyword/value format",
			input:    "host=postgres port=5432 user=coworkspace password=secret dbname=coworkspace sslmode=disable",
			expected: "pgx5://coworkspace:secret@postgres:5432/coworkspace?sslmode=disable",
		},
		{
			name:     "keyword/value with special chars in password",
			input:    "host=postgres port=5432 user=coworkspace password=PwZ2+xv7/b5k= dbname=coworkspace sslmode=disable",
			expected: "pgx5://coworkspace:PwZ2+xv7%2Fb5k=@postgres:5432/coworkspace?sslmode=disable",
		},
		{
			name:     "keyword/value with defaults",
			input:    "user=myuser password=mypass dbname=mydb",
			expected: "pgx5://myuser:mypass@localhost:5432/mydb",
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
