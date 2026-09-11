package health

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.temporal.io/sdk/client"

	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

// GetReadyzHandler reports whether the process is ready to serve
// traffic by checking external dependencies (psql, clickhouse,
// temporal). Returns 207 Multi-Status with a degraded list when any
// dependency is unhappy.
func (s *Service) GetReadyzHandler(ctx *gin.Context) {
	var checks [dependencyCount]dependencyCheck
	defer func() { s.metrics.record(ctx.Request.Context(), checks) }()
	checks[postgresDependency].started = time.Now()
	// ping psql
	sqlDB, err := s.db.DB()
	if err != nil {
		checks[postgresDependency].finish("connection")
		ctx.Error(stderr.ErrSystem{
			Err:         err,
			Description: "unable to get psql connection",
		})
		return
	}
	if err := sqlDB.PingContext(ctx); err != nil {
		checks[postgresDependency].finish("ping")
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
	checks[postgresDependency].finish("")
	s.mw.Incr("healthcheck.check", metrics.ToTags(map[string]string{
		"system": "psql",
		"status": "ok",
	}))

	degraded := make([]string, 0)

	// ping ch
	checks[clickhouseDependency].started = time.Now()
	chFailure := ""
	chDB, err := s.chDB.DB()
	if err != nil {
		chFailure = "connection"
		degraded = append(degraded, "ch")
		s.mw.Incr("healthcheck.check", metrics.ToTags(map[string]string{
			"system": "ch",
			"status": "unable_to_connect",
		}))
	} else {
		// attempt to ping clickhouse, if we get a connection
		if err := chDB.PingContext(ctx); err != nil {
			chFailure = "ping"
			degraded = append(degraded, "ch")
			s.mw.Incr("healthcheck.check", metrics.ToTags(map[string]string{
				"system": "ch",
				"status": "unable_to_ping",
			}))
		} else {
			// Only increment OK metric if ping succeeded
			s.mw.Incr("healthcheck.check", metrics.ToTags(map[string]string{
				"system": "ch",
				"status": "ok",
			}))
		}

		// Check for read-only replicas (only if connection was successful)
		rows, err := chDB.Query("SELECT table FROM system.replicas WHERE database = 'ctl_api' AND is_readonly = 1")
		if err != nil {
			chFailure = "query"
			degraded = append(degraded, "ch")
			s.mw.Incr("healthcheck.check", metrics.ToTags(map[string]string{
				"system": "ch",
				"status": "unable_to_connect",
			}))
		} else {
			defer rows.Close()

			var tables []string
			var tableName string // Variable to scan each row into

			for rows.Next() {
				err := rows.Scan(&tableName)
				if err != nil {
					chFailure = "scan"
					// Handle scan error
					degraded = append(degraded, "ch")
					s.mw.Incr("healthcheck.check", metrics.ToTags(map[string]string{
						"system": "ch",
						"status": "scan_error",
					}))
					break
				}
				tables = append(tables, tableName)
			}

			// NOTE(fd): we check for iteration errors (but why?)
			if err = rows.Err(); err != nil {
				chFailure = "iteration"
				degraded = append(degraded, "ch")
				s.mw.Incr("healthcheck.check", metrics.ToTags(map[string]string{
					"system": "ch",
					"status": "iteration_error",
				}))
			}

			rowCount := len(tables)
			if rowCount > 0 {
				chFailure = "readonly_replicas"
				degraded = append(degraded, "ch")
				s.mw.Incr("healthcheck.check", metrics.ToTags(map[string]string{
					"system": "ch",
					"status": "readonly_replicas_found",
				}))
				for i, table := range tables {
					ctx.Header(fmt.Sprintf("x-ch-table-in-read-only-%d", i), table)
				}
			}
		}
	}
	checks[clickhouseDependency].finish(chFailure)

	// ping temporal
	checks[temporalDependency].started = time.Now()
	_, err = s.tclient.CheckHealth(ctx, &client.CheckHealthRequest{})
	temporalFailure := ""
	if err != nil {
		temporalFailure = "ping"
		degraded = append(degraded, "temporal")
		s.mw.Incr("healthcheck.check", metrics.ToTags(map[string]string{
			"system": "temporal",
			"status": "unable_to_ping",
		}))
	}
	checks[temporalDependency].finish(temporalFailure)
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

	ctx.JSON(statusCode, map[string]any{
		"status":   status,
		"degraded": degraded,
	})
}
