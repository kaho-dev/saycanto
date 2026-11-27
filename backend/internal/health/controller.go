package health

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/kaho-dev/saycanto/backend/internal/shared/httputil"
)

// Controller handles health check endpoints.
type Controller struct{}

// NewController creates a new health Controller.
func NewController() *Controller {
	return &Controller{}
}

// RegisterRoutes mounts health routes onto the router.
func (c *Controller) RegisterRoutes(r chi.Router) {
	r.Get("/api/health", c.Check)
}

// Check returns HTTP 200 with service status.
func (c *Controller) Check(w http.ResponseWriter, _ *http.Request) {
	httputil.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
