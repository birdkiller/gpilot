package postgres

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gpilot/internal/infra/logger"

	"github.com/jackc/pgx/v5/pgxpool"
)

func RunMigrations(pool *pgxpool.Pool, migrationsDir string) error {
	// Create migrations tracking table
	_, err := pool.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`)
	if err != nil {
		return fmt.Errorf("create migrations table: %w", err)
	}

	// Read migration files
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}

	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			files = append(files, e.Name())
		}
	}
	sort.Strings(files)

	for _, f := range files {
		// Check if already applied
		var count int
		err := pool.QueryRow(context.Background(),
			"SELECT COUNT(*) FROM schema_migrations WHERE version = $1", f).Scan(&count)
		if err != nil {
			return fmt.Errorf("check migration %s: %w", f, err)
		}
		if count > 0 {
			logger.L.Infow("migration already applied", "file", f)
			continue
		}

		// Read and execute
		content, err := os.ReadFile(filepath.Join(migrationsDir, f))
		if err != nil {
			return fmt.Errorf("read migration %s: %w", f, err)
		}

		_, err = pool.Exec(context.Background(), string(content))
		if err != nil {
			return fmt.Errorf("execute migration %s: %w", f, err)
		}

		// Record migration
		_, err = pool.Exec(context.Background(),
			"INSERT INTO schema_migrations (version) VALUES ($1)", f)
		if err != nil {
			return fmt.Errorf("record migration %s: %w", f, err)
		}

		logger.L.Infow("migration applied", "file", f)
	}

	return nil
}
