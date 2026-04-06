package subscriptionsapi

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"path/filepath"

	"github.com/adammilligan/EffectiveMobileTest/internal/pkg/config"
	"github.com/adammilligan/EffectiveMobileTest/internal/pkg/migrator"
	"github.com/adammilligan/EffectiveMobileTest/internal/pkg/postgres"
	"github.com/adammilligan/EffectiveMobileTest/internal/pkg/subscriptions"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
)

type App struct {
	Router http.Handler
	Pool   *pgxpool.Pool
}

func New(ctx context.Context, cfg config.Config) (*App, error) {
	pool, err := postgres.NewPool(ctx, cfg.PostgresDSN())
	if err != nil {
		return nil, err
	}

	sqlDB := stdlib.OpenDBFromPool(pool)
	if err := runMigrations(sqlDB); err != nil {
		pool.Close()
		return nil, err
	}

	repo := subscriptions.NewRepo(pool)
	handlers := NewHandlers(repo)

	return &App{
		Router: NewRouter(handlers),
		Pool:   pool,
	}, nil
}

func runMigrations(db *sql.DB) error {
	abs, err := filepath.Abs("migrations")
	if err != nil {
		return fmt.Errorf("resolve migrations path: %w", err)
	}

	slog.Info("running migrations", "path", abs)
	if err := migrator.Up(db, abs); err != nil {
		return err
	}
	slog.Info("migrations applied")
	return nil
}

