package service

import (
	"github.com/a-h/templ"
	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/admin-dashboard/service/views"
)

func (s *service) Index(c *gin.Context) {
	// Render the templ component
	component := views.Index("Nuon Admin Dashboard", "Welcome to the Nuon admin dashboard service")

	// Use templ's Handler to render to the response
	templ.Handler(component).ServeHTTP(c.Writer, c.Request)
}
