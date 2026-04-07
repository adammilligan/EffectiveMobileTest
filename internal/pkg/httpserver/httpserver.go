package httpserver

import (
	"net/http"
	"time"
)

// Config contains parameters for creating an HTTP server.
type Config struct {
	Addr              string
	Handler           http.Handler
	ReadHeaderTimeout time.Duration
}

// New creates an http.Server with the given config.
func New(cfg Config) *http.Server {
	return &http.Server{
		Addr:              cfg.Addr,
		Handler:           cfg.Handler,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
	}
}
