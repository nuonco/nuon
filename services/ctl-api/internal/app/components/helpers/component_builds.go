package helpers

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/views"
	"gorm.io/gorm"
)

func (s *Helpers) GetComponentLatestBuilds(ctx *gin.Context, cmpIDs ...string) ([]app.ComponentBuild, error) {
	components := make([]app.Component, 0, len(cmpIDs))
	res := s.db.WithContext(ctx).
		Preload("ComponentConfigs", func(db *gorm.DB) *gorm.DB {
			return db.Order(views.TableOrViewName(s.db, &app.ComponentConfigConnection{}, ".created_at DESC"))
		}).
		Preload("ComponentConfigs.ComponentBuilds", func(db *gorm.DB) *gorm.DB {
			return db.Order("component_builds.created_at DESC")
		}).
		Preload("ComponentConfigs.ComponentBuilds.ComponentConfigConnection").
		Preload("ComponentConfigs.ComponentBuilds.VCSConnectionCommit").
		Find(&components, "id IN ?", cmpIDs)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get components: %w", res.Error)
	}

	builds := make([]app.ComponentBuild, 0, len(components))
	for _, cmp := range components {
		for _, cfg := range cmp.ComponentConfigs {
			builds = append(builds, cfg.ComponentBuilds...)
		}
	}
	return builds, nil
}
