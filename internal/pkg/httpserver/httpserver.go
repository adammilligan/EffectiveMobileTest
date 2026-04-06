package httpserver

import (
	"net/http"
	"time"
)

type Config struct {
	Addr              string
	Handler           http.Handler
	ReadHeaderTimeout time.Duration
}

func New(cfg Config) *http.Server {
	return &http.Server{
		Addr:              cfg.Addr,
		Handler:           cfg.Handler,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout,
	}
}
