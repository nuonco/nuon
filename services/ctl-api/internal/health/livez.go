package health

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.temporal.io/sdk/client"

	"github.com/powertoolsdev/mono/pkg/metrics"
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
		s.mw.Incr("healthcheck.check", metrics.ToTags(map[string]string{
			"system": "psql",
			"status": "unable_to_ping",
		}))
		ctx.Error(stderr.ErrSystem{
			Err:         err,
			Description: "unable to ping psql db",
		})
		return
	}
	s.mw.Incr("healthcheck.check", metrics.ToTags(map[string]string{
		"system": "psql",
		"status": "ok",
	}))

	degraded := make([]string, 0)

	// ping ch
	chDB, err := s.chDB.DB()
	if err != nil {
		degraded = append(degraded, "ch")
		s.mw.Incr("healthcheck.check", metrics.ToTags(map[string]string{
			"system": "ch",
			"status": "unable_to_connect",
		}))
	} else {
		// attempt to ping clickhouse, if we get a connection
		if err := chDB.PingContext(ctx); err != nil {
			degraded = append(degraded, "ch")
			s.mw.Incr("healthcheck.check", metrics.ToTags(map[string]string{
				"system": "ch",
				"status": "unable_to_ping",
			}))
		}

		s.mw.Incr("healthcheck.check", metrics.ToTags(map[string]string{
			"system": "ch",
			"status": "ok",
		}))
	}

	// ping temporal
	_, err = s.tclient.CheckHealth(ctx, &client.CheckHealthRequest{})
	if err != nil {
		degraded = append(degraded, "temporal")
		s.mw.Incr("healthcheck.check", metrics.ToTags(map[string]string{
			"system": "temporal",
			"status": "unable_to_ping",
		}))
	}
	s.mw.Incr("healthcheck.check", metrics.ToTags(map[string]string{
		"system": "temporal",
		"status": "ok",
	}))

	statusCode := http.StatusOK
	status := "ok"
	if len(degraded) > 0 {
		status = "degraded"
		statusCode = http.StatusMultiStatus
	}

	ctx.JSON(statusCode, map[string]interface{}{
		"status":   status,
		"degraded": degraded,
	})
}
