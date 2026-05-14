package service

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// RemoveOldRunnerProcesses deletes all but the most recent runner process per
// runner + process_type combination for the given org.
func (s *service) RemoveOldRunnerProcesses(c *gin.Context) {
	orgID := c.Param("id")

	// Get all runner processes for the org, ordered so most recent comes first.
	var processes []app.RunnerProcess
	if res := s.db.WithContext(c.Request.Context()).
		Where("org_id = ?", orgID).
		Order("runner_id, type, created_at DESC").
		Find(&processes); res.Error != nil {
		s.l.Error("failed to list runner processes", zap.Error(res.Error), zap.String("org_id", orgID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list runner processes"})
		return
	}

	// For each runner+type combo, keep the most recent and collect the rest for deletion.
	type key struct {
		RunnerID string
		Type     app.RunnerProcessType
	}
	seen := make(map[key]bool)
	var toDelete []string

	for _, p := range processes {
		k := key{RunnerID: p.RunnerID, Type: p.Type}
		if seen[k] {
			toDelete = append(toDelete, p.ID)
		} else {
			seen[k] = true
		}
	}

	deleted := 0
	for _, id := range toDelete {
		if res := s.db.WithContext(c.Request.Context()).
			Delete(&app.RunnerProcess{}, "id = ?", id); res.Error != nil {
			s.l.Warn("failed to delete old runner process",
				zap.String("process_id", id),
				zap.String("org_id", orgID),
				zap.Error(res.Error))
			continue
		}
		deleted++
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok", "processes_deleted": deleted})
}
