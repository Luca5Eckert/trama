package database

import (
	"context"
	"embed"
	"fmt"
	"io/fs"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/tern/v2/migrate"
)

//go:embed migrations/*.sql
var migrationFiles embed.FS

func Migrate(ctx context.Context, pool *pgxpool.Pool) error {
	migrationFS, err := fs.Sub(migrationFiles, "migrations")
	if err != nil {
		return fmt.Errorf("open migration filesystem: %w", err)
	}

	connection, err := pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("acquire migration connection: %w", err)
	}
	defer connection.Release()

	migrator, err := migrate.NewMigrator(ctx, connection.Conn(), "public.trama_schema_version")
	if err != nil {
		return fmt.Errorf("initialize migrator: %w", err)
	}
	if err := migrator.LoadMigrations(migrationFS); err != nil {
		return fmt.Errorf("load migrations: %w", err)
	}
	if err := migrator.Migrate(ctx); err != nil {
		return fmt.Errorf("apply migrations: %w", err)
	}
	return nil
}
