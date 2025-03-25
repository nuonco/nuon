package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/db/plugins/migrations"
)

//	@ID						GetMigrations
//	@Summary				get all migrations
//	@Description.markdown	get_migrations.md
//	@Tags					general/admin
//	@Accept					json
//	@Produce				json
//	@Success				200	{array}	migrations.MigrationModel
//	@Router					/v1/general/migrations [get]
func (s *service) GetMigrations(ctx *gin.Context) {
	migrations, err := s.getMigrations(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, migrations)
}

func (s *service) getMigrations(ctx context.Context) ([]*migrations.MigrationModel, error) {
	var migrations []*migrations.MigrationModel

	res := s.db.WithContext(ctx).
		Order("created_at DESC").
		Find(&migrations)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get migrations: %w", res.Error)
	}

	return migrations, nil
}
