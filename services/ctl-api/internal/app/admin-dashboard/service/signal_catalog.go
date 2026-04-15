package service

import (
	"sort"

	"github.com/a-h/templ"
	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/admin-dashboard/service/views"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/catalog"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

func (s *service) SignalCatalog(c *gin.Context) {
	infos := catalog.InspectAll()
	sort.Slice(infos, func(i, j int) bool {
		return string(infos[i].Type) < string(infos[j].Type)
	})

	component := views.SignalCatalogView(infos)
	templ.Handler(component).ServeHTTP(c.Writer, c.Request)
}

func signalAttributesForType(signalType signal.SignalType) *catalog.SignalTypeInfo {
	info, err := catalog.InspectType(signalType)
	if err != nil {
		return nil
	}
	return &info
}
