package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

// ShutdownOrgRunnerProcesses creates a shutdown record for the most recent
// runner process per runner + process_type combination for the given org.
// The runner's shutdown poller will pick up the record and execute the shutdown.
func (s *service) ShutdownOrgRunnerProcesses(c *gin.Context) {
	orgID := c.Param("id")
	ctx := c.Request.Context()
	ctx = cctx.SetOrgIDContext(ctx, orgID)

	// Get all active/offline runner processes for the org, most recent first.
	var processes []app.RunnerProcess
	if res := s.db.WithContext(ctx).
		Where("org_id = ?", orgID).
		Where("composite_status::jsonb ->> 'status' IN ('active', 'offline')").
		Order("runner_id, type, created_at DESC").
		Find(&processes); res.Error != nil {
		s.l.Error("failed to list runner processes", zap.Error(res.Error), zap.String("org_id", orgID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list runner processes"})
		return
	}

	// Keep only the most recent process per runner+type.
	type key struct {
		RunnerID string
		Type     app.RunnerProcessType
	}
	seen := make(map[key]bool)
	var latest []app.RunnerProcess

	for _, p := range processes {
		k := key{RunnerID: p.RunnerID, Type: p.Type}
		if !seen[k] {
			seen[k] = true
			latest = append(latest, p)
		}
	}

	// Resolve a real account ID for the created_by_id FK. The admin dashboard
	// middleware sets "admin-dashboard" which isn't a real account. Fall back
	// to the process's own created_by_id so we have a valid FK reference,
	// and set it on the context so BeforeCreate hooks pick it up.
	createdByID, _ := cctx.AccountIDFromContext(ctx)
	if createdByID == "" || createdByID == "admin-dashboard" {
		if len(latest) > 0 {
			createdByID = latest[0].CreatedByID
		}
	}
	ctx = cctx.SetAccountIDContext(ctx, createdByID)

	shutdowns := 0
	var createErrors []string
	for i := range latest {
		shutdown := app.RunnerProcessShutdown{
			RunnerProcessID: latest[i].ID,
			OrgID:           orgID,
			CreatedByID:     createdByID,
			Type:            app.RunnerProcessShutdownTypeGraceful,
			CompositeStatus: app.NewCompositeStatus(ctx, app.Status(app.RunnerProcessShutdownStatusRequested)),
		}
		if res := s.db.WithContext(ctx).Create(&shutdown); res.Error != nil {
			s.l.Warn("failed to create shutdown record",
				zap.String("process_id", latest[i].ID),
				zap.String("org_id", orgID),
				zap.Error(res.Error))
			createErrors = append(createErrors, res.Error.Error())
			continue
		}
		shutdowns++
	}

	c.JSON(http.StatusOK, gin.H{
		"status":             "ok",
		"processes_shutdown": shutdowns,
		"process_count":      len(processes),
		"latest_count":       len(latest),
		"create_errors":      createErrors,
	})
}
