package activities

import (
	"context"
	"sort"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type EvaluateComponentHealthRequest struct {
	InstallID string `validate:"required"`
}

type EvaluateComponentHealthResponse struct {
	Skipped     bool `json:"skipped"`
	Evaluated   int  `json:"evaluated"`
	Updated     int  `json:"updated"`
	Transitions int  `json:"transitions"`
}

// EvaluateComponentHealth derives each install component's debounced health
// verdict from the runner's recent resource observations in ClickHouse and
// persists it on the component (health_status / health_status_v2), recording a
// transition row on every verdict change. No-ops when the org doesn't have the
// component-health feature or the install is gone.
//
// @temporal-gen-v2 activity
// @start-to-close-timeout 60s
// @by-field InstallID
func (a *Activities) EvaluateComponentHealth(ctx context.Context, req *EvaluateComponentHealthRequest) (*EvaluateComponentHealthResponse, error) {
	resp := &EvaluateComponentHealthResponse{}

	var install app.Install
	if err := a.db.WithContext(ctx).Where(app.Install{ID: req.InstallID}).First(&install).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			resp.Skipped = true
			return resp, nil
		}
		return nil, errors.Wrap(err, "unable to get install")
	}

	enabled, err := a.features.OrgHasFeature(ctx, install.OrgID, app.OrgFeatureComponentHealth)
	if err != nil {
		return nil, errors.Wrap(err, "unable to check component-health feature")
	}
	if !enabled {
		resp.Skipped = true
		return resp, nil
	}

	var installComponents []app.InstallComponent
	if err := a.db.WithContext(ctx).
		Preload("Component").
		Where(app.InstallComponent{InstallID: install.ID}).
		Find(&installComponents).Error; err != nil {
		return nil, errors.Wrap(err, "unable to list install components")
	}
	if len(installComponents) == 0 {
		return resp, nil
	}

	now := time.Now()
	reportsByComponent, err := a.recentComponentHealthReports(ctx, install.OrgID, install.ID, now)
	if err != nil {
		return nil, errors.Wrap(err, "unable to load resource observations")
	}

	transitions := make([]app.InstallComponentHealthTransition, 0)
	for i := range installComponents {
		ic := &installComponents[i]
		resp.Evaluated++

		reports := reportsByComponent[ic.ID]
		verdict := a.componentVerdict(ic, reports, now)

		var latest *componentHealthReport
		if len(reports) > 0 {
			latest = &reports[0]
		}
		description := componentHealthDescription(verdict, latest)

		if verdict == ic.HealthStatus && description == ic.HealthStatusDescription {
			continue
		}

		if err := a.writeComponentHealth(ctx, ic, verdict, description, latest, now); err != nil {
			return nil, errors.Wrapf(err, "unable to update health for install component %s", ic.ID)
		}
		resp.Updated++

		if verdict != ic.HealthStatus {
			transitions = append(transitions, newComponentHealthTransition(ic, verdict, latest, now))
		}
	}

	if len(transitions) > 0 {
		resp.Transitions = len(transitions)
		if err := a.chDB.WithContext(ctx).CreateInBatches(&transitions, 100).Error; err != nil {
			// Best-effort: verdicts in Postgres are the source of truth, the
			// transition log only powers the (future) timeline.
			a.l.Warn("unable to record component health transitions",
				zap.String("install_id", install.ID),
				zap.Error(err),
			)
		}
	}

	return resp, nil
}

// componentVerdict wraps the debounce with the component-level short-circuits:
// component types the runner can't observe, and disabled components, carry no
// health signal.
func (a *Activities) componentVerdict(ic *app.InstallComponent, reports []componentHealthReport, now time.Time) app.InstallComponentHealthStatus {
	watchable := ic.Component.Type == app.ComponentTypeHelmChart ||
		ic.Component.Type == app.ComponentTypeKubernetesManifest
	if !watchable || ic.Status == app.InstallComponentStatusDisabled {
		return app.InstallComponentHealthStatusNotApplicable
	}

	return nextComponentHealthVerdict(ic.HealthStatus, reports, now)
}

// recentComponentHealthReports reads the observation window from ClickHouse
// and collapses it to one report (worst resource) per component per report
// timestamp, newest first.
func (a *Activities) recentComponentHealthReports(ctx context.Context, orgID, installID string, now time.Time) (map[string][]componentHealthReport, error) {
	var rows []app.InstallComponentResourceState
	if err := a.chDB.WithContext(ctx).
		Select("install_component_id", "kind", "namespace", "name", "health", "message", "native_status", "observed_at").
		Where(app.InstallComponentResourceState{
			OrgID:     orgID,
			InstallID: installID,
			Source:    app.InstallComponentResourceSourceComponent,
		}).
		Where("observed_at > ?", now.Add(-componentHealthObservationWindow)).
		Find(&rows).Error; err != nil {
		return nil, err
	}

	type reportKey struct {
		componentID string
		observedAt  int64
	}
	merged := make(map[reportKey]*componentHealthReport)
	for _, r := range rows {
		key := reportKey{componentID: r.InstallComponentID, observedAt: r.ObservedAt.UnixNano()}
		rep, ok := merged[key]
		if !ok {
			rep = &componentHealthReport{
				ObservedAt:     r.ObservedAt,
				Health:         app.InstallComponentHealthStatus(r.Health),
				RootKind:       r.Kind,
				RootNamespace:  r.Namespace,
				RootName:       r.Name,
				Message:        r.Message,
				ResourceCounts: map[string]int{},
			}
			merged[key] = rep
		}

		health := app.InstallComponentHealthStatus(r.Health)
		if componentHealthSeverity[health] > componentHealthSeverity[rep.Health] {
			rep.Health = health
			rep.RootKind = r.Kind
			rep.RootNamespace = r.Namespace
			rep.RootName = r.Name
			rep.Message = r.Message
		}
		if rep.NativeStatus == "" {
			rep.NativeStatus = r.NativeStatus
		}
		rep.Resources++
		rep.ResourceCounts[r.Health]++
	}

	out := make(map[string][]componentHealthReport)
	for key, rep := range merged {
		out[key.componentID] = append(out[key.componentID], *rep)
	}
	for id := range out {
		sort.Slice(out[id], func(i, j int) bool {
			return out[id][i].ObservedAt.After(out[id][j].ObservedAt)
		})
	}

	return out, nil
}

func (a *Activities) writeComponentHealth(ctx context.Context, ic *app.InstallComponent, verdict app.InstallComponentHealthStatus, description string, latest *componentHealthReport, now time.Time) error {
	metadata := map[string]any{}
	if latest != nil {
		metadata["observed_at"] = latest.ObservedAt.UTC().Format(time.RFC3339)
		metadata["resources"] = latest.Resources
		metadata["resource_counts"] = latest.ResourceCounts
		if latest.RootKind != "" {
			metadata["root_resource_kind"] = latest.RootKind
			metadata["root_resource_namespace"] = latest.RootNamespace
			metadata["root_resource_name"] = latest.RootName
			metadata["message"] = latest.Message
		}
		if latest.NativeStatus != "" {
			metadata["native_status"] = latest.NativeStatus
		}
	}

	return a.db.WithContext(ctx).
		Model(&app.InstallComponent{ID: ic.ID}).
		Select("health_status", "health_status_v2").
		Updates(app.InstallComponent{
			HealthStatus: verdict,
			HealthStatusV2: app.CompositeStatus{
				CreatedAtTS:            now.Unix(),
				Status:                 app.Status(verdict),
				StatusHumanDescription: description,
				Metadata:               metadata,
			},
		}).Error
}

func newComponentHealthTransition(ic *app.InstallComponent, verdict app.InstallComponentHealthStatus, latest *componentHealthReport, now time.Time) app.InstallComponentHealthTransition {
	t := app.InstallComponentHealthTransition{
		OrgID:              ic.OrgID,
		InstallID:          ic.InstallID,
		InstallComponentID: ic.ID,
		ComponentID:        ic.ComponentID,
		FromHealth:         string(ic.HealthStatus),
		ToHealth:           string(verdict),
		ObservedAt:         now,
	}
	if latest != nil {
		t.RootResourceKind = latest.RootKind
		t.RootResourceNamespace = latest.RootNamespace
		t.RootResourceName = latest.RootName
		t.Message = latest.Message
	}
	return t
}
