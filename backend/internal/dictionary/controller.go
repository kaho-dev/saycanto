package dictionary

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/kaho-dev/saycanto/backend/internal/shared/httputil"
)

// Controller handles HTTP requests for the dictionary feature.
type Controller struct {
	service Service
}

// NewController constructs a dictionary Controller with an injected Service.
func NewController(service Service) *Controller {
	return &Controller{service: service}
}

// RegisterRoutes mounts dictionary routes onto the router.
func (c *Controller) RegisterRoutes(r chi.Router) {
	r.Get("/api/search", c.Search)
	r.Get("/api/words", c.ListWords)
	r.Get("/api/words/{id}", c.GetByID)
}

// Search handles GET /api/search?q=...&limit=...&register=...&type=...
func (c *Controller) Search(w http.ResponseWriter, r *http.Request) {
	q := SearchQuery{
		Query:    r.URL.Query().Get("q"),
		Register: r.URL.Query().Get("register"),
		WordType: r.URL.Query().Get("type"),
		Limit:    httputil.ParseIntQuery(r, "limit", defaultSearchLimit),
	}

	result, err := c.service.Search(r.Context(), q)
	if err != nil {
		if errors.Is(err, ErrInvalidInput) {
			httputil.WriteError(w, http.StatusBadRequest, err.Error())
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "failed to search dictionary")
		return
	}

	httputil.WriteJSON(w, http.StatusOK, result)
}

// GetByID handles GET /api/words/{id}
func (c *Controller) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	entry, err := c.service.GetByID(r.Context(), id)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidInput):
			httputil.WriteError(w, http.StatusBadRequest, err.Error())
		case errors.Is(err, ErrNotFound):
			httputil.WriteError(w, http.StatusNotFound, err.Error())
		default:
			httputil.WriteError(w, http.StatusInternalServerError, "failed to load word")
		}
		return
	}

	httputil.WriteJSON(w, http.StatusOK, entry)
}

// ListWords handles GET /api/words?limit=...&offset=...&register=...&type=...
func (c *Controller) ListWords(w http.ResponseWriter, r *http.Request) {
	q := ListQuery{
		Register: r.URL.Query().Get("register"),
		WordType: r.URL.Query().Get("type"),
		Limit:    httputil.ParseIntQuery(r, "limit", defaultSearchLimit),
		Offset:   httputil.ParseIntQuery(r, "offset", 0),
	}

	result, err := c.service.ListWords(r.Context(), q)
	if err != nil {
		if errors.Is(err, ErrInvalidInput) {
			httputil.WriteError(w, http.StatusBadRequest, err.Error())
			return
		}
		httputil.WriteError(w, http.StatusInternalServerError, "failed to list words")
		return
	}

	httputil.WriteJSON(w, http.StatusOK, result)
}
