package database

import (
	"errors"
	"fmt"
	"log/slog"

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

// pgxURL converts a postgres:// URL to the pgx5:// scheme that golang-migrate expects.
func pgxURL(databaseURL string) string {
	if len(databaseURL) > 11 && databaseURL[:11] == "postgres://" {
		return "pgx5://" + databaseURL[11:]
	}
	if len(databaseURL) > 13 && databaseURL[:13] == "postgresql://" {
		return "pgx5://" + databaseURL[13:]
	}
	return databaseURL
}
