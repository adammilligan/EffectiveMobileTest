package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/adammilligan/EffectiveMobileTest/internal/pkg/config"
	"github.com/adammilligan/EffectiveMobileTest/internal/pkg/httpserver"
	"github.com/adammilligan/EffectiveMobileTest/internal/pkg/logger"
	subscriptionsapi "github.com/adammilligan/EffectiveMobileTest/internal/app/subscriptions-api"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("config load failed", "error", err)
		os.Exit(1)
	}

	log := logger.New(cfg.Log.Level)
	slog.SetDefault(log)

	app, err := subscriptionsapi.New(context.Background(), cfg)
	if err != nil {
		slog.Error("app init failed", "error", err)
		os.Exit(1)
	}
	defer app.Pool.Close()

	srv := httpserver.New(httpserver.Config{
		Addr:              cfg.HTTPAddr(),
		Handler:           app.Router,
		ReadHeaderTimeout: 5 * time.Second,
	})

	go func() {
		slog.Info("http server started", "addr", cfg.HTTPAddr())

		runErr := srv.ListenAndServe()
		if runErr != nil && !errors.Is(runErr, http.ErrServerClosed) {
			slog.Error("http server failed", "error", runErr)
		}
	}()

	stopCh := make(chan os.Signal, 1)
	signal.Notify(stopCh, os.Interrupt, syscall.SIGTERM)
	<-stopCh

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	slog.Info("shutting down")

	if shutdownErr := srv.Shutdown(ctx); shutdownErr != nil {
		slog.Error("http shutdown failed", "error", shutdownErr)
	}
}

