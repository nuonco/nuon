package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
)

// ShutdownHintOrgRunnerProcesses sends a shutdown hint (via the full shutdown
// helper) for the most recent runner process per runner + process_type
// combination for the given org. This enqueues the process_shutdown signal
// and writes a red health check in addition to creating the shutdown record.
func (s *service) ShutdownHintOrgRunnerProcesses(c *gin.Context) {
	orgID := c.Param("id")
	ctx := c.Request.Context()

	var processes []app.RunnerProcess
	if res := s.db.WithContext(ctx).
		Where("org_id = ?", orgID).
		Scopes(generics.WhereJSONBStatusIn("composite_status", "active", "offline")).
		Order("runner_id, type, created_at DESC").
		Find(&processes); res.Error != nil {
		s.l.Error("failed to list runner processes", zap.Error(res.Error), zap.String("org_id", orgID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list runner processes"})
		return
	}

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

	shutdowns := 0
	for i := range latest {
		if _, err := s.runnersHelpers.ShutdownProcess(ctx, &latest[i], app.RunnerProcessShutdownTypeGraceful); err != nil {
			s.l.Warn("failed to send shutdown hint",
				zap.String("process_id", latest[i].ID),
				zap.String("runner_id", latest[i].RunnerID),
				zap.String("org_id", orgID),
				zap.Error(err))
			continue
		}
		shutdowns++
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok", "processes_shutdown": shutdowns})
}
