package postgres

import (
	"context"
	"embed"
	"fmt"
	"net/http"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/httpfs"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"

	"github.com/go-api-template/app/config"
)

//go:embed migrations
var MigrationsFS embed.FS

type Client struct {
	Pool *pgxpool.Pool
}

func (c *Client) Close() {
	c.Pool.Close()
}

// New connects to the Postgres database and performs migrations.
func New(ctx context.Context, config config.Postgres) (*Client, error) {
	const operation = "Postgres.New"

	connString := fmt.Sprintf("user=%s password=%s host=%s port=%s dbname=%s", config.User, config.Password, config.Host, config.Port, config.DatabaseName)

	pgxConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, pgxConfig)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	driver, err := postgres.WithInstance(stdlib.OpenDB(*pgxConfig.ConnConfig), &postgres.Config{
		DatabaseName: pgxConfig.ConnConfig.Database,
	})
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	source, err := httpfs.New(http.FS(MigrationsFS), "migrations")
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	migration, err := migrate.NewWithInstance("httpfs", source, pgxConfig.ConnConfig.Database, driver)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	err = migration.Up()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	srcErr, dbErr := migration.Close()
	if srcErr != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	if dbErr != nil {
		return nil, fmt.Errorf("%s: %w", operation, err)
	}

	return &Client{pool}, nil
}
