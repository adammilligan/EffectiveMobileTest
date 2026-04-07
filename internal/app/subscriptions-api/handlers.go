package subscriptionsapi

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/adammilligan/EffectiveMobileTest/internal/pkg/httpjson"
	"github.com/adammilligan/EffectiveMobileTest/internal/pkg/optional"
	"github.com/adammilligan/EffectiveMobileTest/internal/pkg/subscriptions"
	"github.com/adammilligan/EffectiveMobileTest/internal/pkg/validate"
	"github.com/go-chi/chi/v5"
)

type Handlers struct {
	repo repo
}

type repo interface {
	Create(ctx context.Context, p subscriptions.CreateParams) (subscriptions.Subscription, error)
	Get(ctx context.Context, id string) (subscriptions.Subscription, bool, error)
	Update(ctx context.Context, id string, p subscriptions.UpdateParams) (subscriptions.Subscription, bool, error)
	Delete(ctx context.Context, id string) (bool, error)
	List(ctx context.Context, p subscriptions.ListParams) ([]subscriptions.Subscription, error)
	FindOverlappingForTotal(ctx context.Context, p subscriptions.TotalQueryParams) ([]subscriptions.Subscription, error)
}

func NewHandlers(repo repo) *Handlers {
	return &Handlers{repo: repo}
}

type subscriptionResponse struct {
	ID          string  `json:"id"`
	ServiceName string  `json:"service_name"`
	Price       int     `json:"price"`
	UserID      string  `json:"user_id"`
	StartDate   string  `json:"start_date"`
	EndDate     *string `json:"end_date,omitempty"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

func toResponse(s subscriptions.Subscription) subscriptionResponse {
	var end *string

	if s.EndMonth != nil {
		v := s.EndMonth.String()
		end = &v
	}

	return subscriptionResponse{
		ID:          s.ID,
		ServiceName: s.ServiceName,
		Price:       s.PriceRub,
		UserID:      s.UserID,
		StartDate:   s.StartMonth.String(),
		EndDate:     end,
		CreatedAt:   s.CreatedAt.UTC().Format(timeRFC3339),
		UpdatedAt:   s.UpdatedAt.UTC().Format(timeRFC3339),
	}
}

const timeRFC3339 = "2006-01-02T15:04:05Z07:00"

type createSubscriptionRequest struct {
	ServiceName string  `json:"service_name"`
	Price       int     `json:"price"`
	UserID      string  `json:"user_id"`
	StartDate   string  `json:"start_date"`
	EndDate     *string `json:"end_date,omitempty"`
}

func (h *Handlers) Create(w http.ResponseWriter, r *http.Request) {
	var req createSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpjson.WriteError(w, http.StatusBadRequest, "invalid json")
		return
	}

	if err := validate.Required(req.ServiceName, "service_name"); err != nil {
		httpjson.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := validate.NonNegativeInt(req.Price, "price"); err != nil {
		httpjson.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	if err := validate.UUID(req.UserID, "user_id"); err != nil {
		httpjson.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	start, err := subscriptions.ParseMonth(req.StartDate)
	if err != nil {
		httpjson.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	var end *subscriptions.Month

	if req.EndDate != nil {
		m, err := subscriptions.ParseMonth(*req.EndDate)
		if err != nil {
			httpjson.WriteError(w, http.StatusBadRequest, err.Error())
			return
		}

		if m.IsBefore(start) {
			httpjson.WriteError(w, http.StatusBadRequest, "end_date must be >= start_date")
			return
		}

		end = &m
	}

	created, err := h.repo.Create(r.Context(), subscriptions.CreateParams{
		ServiceName: req.ServiceName,
		PriceRub:    req.Price,
		UserID:      req.UserID,
		StartMonth:  start,
		EndMonth:    end,
	})
	if err != nil {
		httpjson.WriteError(w, http.StatusInternalServerError, "db error")
		return
	}

	httpjson.WriteJSON(w, http.StatusCreated, toResponse(created))
}

func (h *Handlers) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := validate.UUID(id, "id"); err != nil {
		httpjson.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	s, isFound, err := h.repo.Get(r.Context(), id)
	if err != nil {
		httpjson.WriteError(w, http.StatusInternalServerError, "db error")
		return
	}

	if !isFound {
		httpjson.WriteError(w, http.StatusNotFound, "not found")
		return
	}

	httpjson.WriteJSON(w, http.StatusOK, toResponse(s))
}

type patchSubscriptionRequest struct {
	ServiceName optional.NullableString `json:"service_name"`
	Price       optional.NullableInt    `json:"price"`
	StartDate   optional.NullableMonth  `json:"start_date"`
	EndDate     optional.NullableMonth  `json:"end_date"`
}

func (h *Handlers) Patch(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := validate.UUID(id, "id"); err != nil {
		httpjson.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	var req patchSubscriptionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httpjson.WriteError(w, http.StatusBadRequest, "invalid json")
		return
	}

	var p subscriptions.UpdateParams

	if req.ServiceName.IsSet {
		if req.ServiceName.Value != nil {
			if err := validate.Required(*req.ServiceName.Value, "service_name"); err != nil {
				httpjson.WriteError(w, http.StatusBadRequest, err.Error())
				return
			}
		}

		p.ServiceName = req.ServiceName.Value
	}

	if req.Price.IsSet {
		if req.Price.Value != nil {
			if err := validate.NonNegativeInt(*req.Price.Value, "price"); err != nil {
				httpjson.WriteError(w, http.StatusBadRequest, err.Error())
				return
			}
		}

		p.PriceRub = req.Price.Value
	}

	if req.StartDate.IsSet {
		if req.StartDate.Value == nil {
			httpjson.WriteError(w, http.StatusBadRequest, "start_date cannot be null")
			return
		}

		m, err := subscriptions.ParseMonth(*req.StartDate.Value)
		if err != nil {
			httpjson.WriteError(w, http.StatusBadRequest, err.Error())
			return
		}

		p.StartMonth = &m
	}

	if req.EndDate.IsSet {
		if req.EndDate.Value == nil {
			var nilMonth *subscriptions.Month

			p.EndMonth = &nilMonth
		} else {
			m, err := subscriptions.ParseMonth(*req.EndDate.Value)
			if err != nil {
				httpjson.WriteError(w, http.StatusBadRequest, err.Error())
				return
			}

			end := m
			endPtr := &end
			p.EndMonth = &endPtr
		}
	}

	if p.StartMonth != nil && p.EndMonth != nil && *p.EndMonth != nil && (**p.EndMonth).IsBefore(*p.StartMonth) {
		httpjson.WriteError(w, http.StatusBadRequest, "end_date must be >= start_date")
		return
	}

	updated, isFound, err := h.repo.Update(r.Context(), id, p)
	if err != nil {
		httpjson.WriteError(w, http.StatusInternalServerError, "db error")
		return
	}

	if !isFound {
		httpjson.WriteError(w, http.StatusNotFound, "not found")
		return
	}

	httpjson.WriteJSON(w, http.StatusOK, toResponse(updated))
}

func (h *Handlers) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := validate.UUID(id, "id"); err != nil {
		httpjson.WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	isDeleted, err := h.repo.Delete(r.Context(), id)
	if err != nil {
		httpjson.WriteError(w, http.StatusInternalServerError, "db error")
		return
	}

	if !isDeleted {
		httpjson.WriteError(w, http.StatusNotFound, "not found")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handlers) List(w http.ResponseWriter, r *http.Request) {
	var userID *string

	if v := r.URL.Query().Get("user_id"); v != "" {
		if err := validate.UUID(v, "user_id"); err != nil {
			httpjson.WriteError(w, http.StatusBadRequest, err.Error())
			return
		}

		userID = &v
	}

	var serviceName *string
	if v := r.URL.Query().Get("service_name"); v != "" {
		serviceName = &v
	}

	limit := parseIntQuery(r, "limit", 50)
	offset := parseIntQuery(r, "offset", 0)

	list, err := h.repo.List(r.Context(), subscriptions.ListParams{
		UserID:      userID,
		ServiceName: serviceName,
		Limit:       limit,
		Offset:      offset,
	})
	if err != nil {
		httpjson.WriteError(w, http.StatusInternalServerError, "db error")
		return
	}

	out := make([]subscriptionResponse, 0, len(list))
	for _, s := range list {
		out = append(out, toResponse(s))
	}

	httpjson.WriteJSON(w, http.StatusOK, out)
}

type totalResponse struct {
	Total int `json:"total"`
}

func (h *Handlers) Total(w http.ResponseWriter, r *http.Request) {
	var userID *string

	if v := r.URL.Query().Get("user_id"); v != "" {
		if err := validate.UUID(v, "user_id"); err != nil {
			httpjson.WriteError(w, http.StatusBadRequest, err.Error())
			return
		}

		userID = &v
	}

	var serviceName *string
	if v := r.URL.Query().Get("service_name"); v != "" {
		serviceName = &v
	}

	fromRaw := r.URL.Query().Get("from")
	toRaw := r.URL.Query().Get("to")

	from, err := subscriptions.ParseMonth(fromRaw)
	if err != nil {
		httpjson.WriteError(w, http.StatusBadRequest, "invalid from")
		return
	}

	to, err := subscriptions.ParseMonth(toRaw)
	if err != nil {
		httpjson.WriteError(w, http.StatusBadRequest, "invalid to")
		return
	}

	if from.Time().After(to.Time()) {
		httpjson.WriteError(w, http.StatusBadRequest, "from must be <= to")
		return
	}

	items, err := h.repo.FindOverlappingForTotal(r.Context(), subscriptions.TotalQueryParams{
		UserID:      userID,
		ServiceName: serviceName,
		From:        from,
		To:          to,
	})
	if err != nil {
		httpjson.WriteError(w, http.StatusInternalServerError, "db error")
		return
	}

	total := 0

	for _, s := range items {
		months := subscriptions.OverlapMonthsInclusive(s.StartMonth, s.EndMonth, from, to)
		total += months * s.PriceRub
	}

	httpjson.WriteJSON(w, http.StatusOK, totalResponse{Total: total})
}

func parseIntQuery(r *http.Request, key string, defaultValue int) int {
	raw := r.URL.Query().Get(key)
	if raw == "" {
		return defaultValue
	}

	v, err := strconv.Atoi(raw)
	if err != nil {
		return defaultValue
	}

	return v
}

