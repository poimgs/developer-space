package database

import (
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// RunMigrations applies all pending up migrations.
func RunMigrations(databaseURL string, migrationsPath string) error {
	m, err := migrate.New("file://"+migrationsPath, pgxURL(databaseURL))
	if err != nil {
		return fmt.Errorf("creating migrator: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("running migrations: %w", err)
	}

	version, dirty, _ := m.Version()
	slog.Info("migrations complete", "version", version, "dirty", dirty)
	return nil
}

// MigrateDown rolls back N migrations.
func MigrateDown(databaseURL string, migrationsPath string, steps int) error {
	m, err := migrate.New("file://"+migrationsPath, pgxURL(databaseURL))
	if err != nil {
		return fmt.Errorf("creating migrator: %w", err)
	}
	defer m.Close()

	if err := m.Steps(-steps); err != nil {
		return fmt.Errorf("rolling back %d migrations: %w", steps, err)
	}

	version, dirty, _ := m.Version()
	slog.Info("migration rollback complete", "version", version, "dirty", dirty)
	return nil
}

// pgxURL converts a database connection string to the pgx5:// URL that golang-migrate expects.
// Accepts both URL format (postgres://...) and keyword/value format (host=... port=...).
func pgxURL(databaseURL string) string {
	if strings.HasPrefix(databaseURL, "postgres://") {
		return "pgx5://" + databaseURL[len("postgres://"):]
	}
	if strings.HasPrefix(databaseURL, "postgresql://") {
		return "pgx5://" + databaseURL[len("postgresql://"):]
	}
	// Already a URL with a scheme (e.g. pgx5://, mysql://) — pass through
	if strings.Contains(databaseURL, "://") {
		return databaseURL
	}

	// Parse keyword/value format (e.g. "host=localhost port=5432 user=x password=y dbname=z sslmode=disable")
	params := map[string]string{}
	for _, part := range strings.Fields(databaseURL) {
		k, v, ok := strings.Cut(part, "=")
		if ok {
			params[k] = v
		}
	}

	host := params["host"]
	if host == "" {
		host = "localhost"
	}
	port := params["port"]
	if port == "" {
		port = "5432"
	}
	dbname := params["dbname"]
	if dbname == "" {
		dbname = params["user"]
	}

	u := &url.URL{
		Scheme: "pgx5",
		Host:   host + ":" + port,
		Path:   dbname,
	}
	if user, ok := params["user"]; ok {
		if pass, ok := params["password"]; ok {
			u.User = url.UserPassword(user, pass)
		} else {
			u.User = url.User(user)
		}
	}
	if sslmode, ok := params["sslmode"]; ok {
		u.RawQuery = "sslmode=" + sslmode
	}

	return u.String()
}
