package activities

import (
	"context"
	"fmt"
	"time"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type GetGateHealthReportsRequest struct {
	// InstallID is required for more than filtering: it is the first column of
	// the table's sort key, so omitting it makes every poll full-scan.
	InstallID          string    `validate:"required"`
	InstallComponentID string    `validate:"required"`
	Since              time.Time `validate:"required"`
}

// GateHealthReport is one collapsed runner report (worst resource at that
// timestamp), read raw by the gate so it can conclude exactly when the window ends.
type GateHealthReport struct {
	ObservedAtTS int64  `json:"observed_at_ts" temporaljson:"observed_at_ts"`
	Health       string `json:"health" temporaljson:"health"`
	RootKind     string `json:"root_kind,omitempty" temporaljson:"root_kind,omitempty"`
	RootName     string `json:"root_name,omitempty" temporaljson:"root_name,omitempty"`
	Message      string `json:"message,omitempty" temporaljson:"message,omitempty"`
	Resources    int    `json:"resources" temporaljson:"resources"`
	// ClusterEvidence marks a report that actually saw the component's cluster
	// resources, as opposed to probes and pushed checks alone.
	ClusterEvidence bool `json:"cluster_evidence" temporaljson:"cluster_evidence"`
}

// GetGateHealthReports returns the component's collapsed health reports
// observed since the given time, newest first.
//
// @temporal-gen-v2 activity
// @start-to-close-timeout 30s
func (a *Activities) GetGateHealthReports(ctx context.Context, req *GetGateHealthReportsRequest) ([]GateHealthReport, error) {
	var rows []app.InstallComponentResourceState
	if err := a.chDB.WithContext(ctx).
		Select("install_component_id", "provider", "kind", "namespace", "name", "health", "message", "native_status", "observed_at").
		Where(app.InstallComponentResourceState{
			InstallID:          req.InstallID,
			InstallComponentID: req.InstallComponentID,
			Source:             app.InstallComponentResourceSourceComponent,
		}).
		Where("observed_at > ?", req.Since).
		Find(&rows).Error; err != nil {
		return nil, fmt.Errorf("unable to query gate health observations: %w", err)
	}

	collapsed := collapseComponentHealthRows(rows)
	reports := collapsed[req.InstallComponentID]

	out := make([]GateHealthReport, 0, len(reports))
	for _, r := range reports {
		out = append(out, GateHealthReport{
			ObservedAtTS:    r.ObservedAt.Unix(),
			Health:          string(r.Health),
			RootKind:        r.RootKind,
			RootName:        r.RootName,
			Message:         r.Message,
			Resources:       r.Resources,
			ClusterEvidence: r.ClusterEvidence,
		})
	}
	return out, nil
}
