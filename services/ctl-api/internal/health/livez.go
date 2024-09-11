package health

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.temporal.io/sdk/client"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/middlewares/stderr"
)

func (s *Service) GetLivezHandler(ctx *gin.Context) {
	// ping psql
	sqlDB, err := s.db.DB()
	if err != nil {
		ctx.Error(stderr.ErrSystem{
			Err:         err,
			Description: "unable to get psql connection",
		})
		return
	}
	if err := sqlDB.PingContext(ctx); err != nil {
		ctx.Error(stderr.ErrSystem{
			Err:         err,
			Description: "unable to ping psql db",
		})
		return
	}

	// ping ch
	chDB, err := s.chDB.DB()
	if err != nil {
		ctx.Error(stderr.ErrSystem{
			Err:         err,
			Description: "unable to get clickhouse connection",
		})
		return
	}
	if err := chDB.PingContext(ctx); err != nil {
		ctx.Error(stderr.ErrSystem{
			Err:         err,
			Description: "unable to ping clickhouse db",
		})
		return
	}

	// ping temporal
	_, err = s.tclient.CheckHealth(ctx, &client.CheckHealthRequest{})
	if err != nil {
		ctx.Error(stderr.ErrSystem{
			Err:         err,
			Description: "unable to check temporal health",
		})
		return
	}

	ctx.JSON(http.StatusOK, map[string]interface{}{
		"status": "ok",
	})
}
