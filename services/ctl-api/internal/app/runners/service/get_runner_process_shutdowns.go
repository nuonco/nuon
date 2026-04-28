package service

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// forceRestartMetadataKey is the runner_group_settings.metadata key checked by
// this stub. When set to "true", the next /shutdowns poll for any runner in
// that group returns one synthetic graceful shutdown so the runner exits and
// is respawned by systemd. The complete endpoint clears the key so the signal
// is one-shot. Operators set it via SQL on the affected group:
//
//	UPDATE runner_group_settings rgs
//	SET metadata = metadata || hstore('force_restart', 'true')
//	FROM runners r
//	WHERE r.runner_group_id = rgs.runner_group_id
//	  AND r.id = '<runner_id>';
const forceRestartMetadataKey = "force_restart"

// runnerProcessShutdownStub mirrors the JSON shape the runner SDK
// (v0.19.873+) deserializes for AppRunnerProcessShutdown. We only fill the
// fields the client's shutdown_poller actually reads (id, status, type).
type runnerProcessShutdownStub struct {
	ID              string         `json:"id"`
	RunnerProcessID string         `json:"runner_process_id"`
	Type            string         `json:"type"`
	Status          string         `json:"status"`
	CompositeStatus map[string]any `json:"composite_status,omitempty"`
}

// @ID			GetRunnerProcessShutdowns
// @Summary		hotfix stub: returns one synthesized requested shutdown when the runner's group has metadata.force_restart=true; otherwise empty
// @Param		runner_id	path	string	true	"runner ID"
// @Param		process_id	path	string	true	"process ID"
// @Tags		runners/runner
// @Accept		json
// @Produce	json
// @Success	200	{array}	runnerProcessShutdownStub
// @Router		/v1/runners/{runner_id}/processes/{process_id}/shutdowns [GET]
func (s *service) GetRunnerProcessShutdowns(ctx *gin.Context) {
	runnerID := ctx.Param("runner_id")
	processID := ctx.Param("process_id")

	var rgs app.RunnerGroupSettings
	res := s.db.WithContext(ctx).
		Joins("JOIN runner_groups rg ON rg.id = runner_group_settings.runner_group_id").
		Joins("JOIN runners r ON r.runner_group_id = rg.id").
		Where("r.id = ?", runnerID).
		First(&rgs)
	if res.Error != nil {
		ctx.JSON(http.StatusOK, []runnerProcessShutdownStub{})
		return
	}

	val, ok := rgs.Metadata[forceRestartMetadataKey]
	if !ok || val == nil || *val != "true" {
		ctx.JSON(http.StatusOK, []runnerProcessShutdownStub{})
		return
	}

	ctx.JSON(http.StatusOK, []runnerProcessShutdownStub{{
		ID:              "rps" + processID,
		RunnerProcessID: processID,
		Type:            "graceful",
		Status:          "requested",
		CompositeStatus: map[string]any{"status": "requested"},
	}})
}
