package subscriptionsapi

import (
	"net/http"

	"github.com/adammilligan/EffectiveMobileTest/internal/pkg/config"
	"github.com/adammilligan/EffectiveMobileTest/internal/pkg/httplog"
	"github.com/adammilligan/EffectiveMobileTest/internal/pkg/ratelimit"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// NewRouter builds subscriptions API HTTP router.
func NewRouter(cfg config.Config, h *Handlers) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(httplog.Middleware)

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	r.Route("/subscriptions", func(r chi.Router) {
		r.Use(ratelimit.Middleware(cfg.RateLimit))
		r.Post("/", h.Create)
		r.Get("/", h.List)
		r.Get("/total", h.Total)

		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", h.Get)
			r.Patch("/", h.Patch)
			r.Delete("/", h.Delete)
		})
	})

	return r
}

