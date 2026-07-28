package activities

import (
	"context"
	"encoding/json"
	"sort"
	"strings"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins/views"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
)

type EvaluateComponentHealthRequest struct {
	InstallID string `validate:"required"`
}

// ComponentHealthNotification describes a debounced component transition that
// should fan out to webhook / Slack subscribers. Only crossings into and out
// of a bad verdict produce one — transitions involving progressing, unknown,
// or not-applicable are silent, and a worsening inside the bad band
// (degraded to unhealthy) does not re-notify.
type ComponentHealthNotification struct {
	Recovered             bool   `json:"recovered"`
	InstallComponentID    string `json:"install_component_id"`
	ComponentID           string `json:"component_id"`
	ComponentName         string `json:"component_name"`
	Health                string `json:"health"`
	PreviousHealth        string `json:"previous_health"`
	Message               string `json:"message"`
	RootResourceKind      string `json:"root_resource_kind"`
	RootResourceNamespace string `json:"root_resource_namespace"`
	RootResourceName      string `json:"root_resource_name"`
}

// InstallHealthNotification is the install-level rollup crossing: set only
// when the composite health entered or left the bad band in this evaluation.
type InstallHealthNotification struct {
	Health                  string `json:"health"`
	PreviousHealth          string `json:"previous_health"`
	Message                 string `json:"message"`
	UnhealthyComponentCount int    `json:"unhealthy_component_count"`
	DegradedComponentCount  int    `json:"degraded_component_count"`
}

type EvaluateComponentHealthResponse struct {
	Skipped     bool `json:"skipped"`
	Evaluated   int  `json:"evaluated"`
	Updated     int  `json:"updated"`
	Transitions int  `json:"transitions"`

	InstallName         string                        `json:"install_name"`
	Notifications       []ComponentHealthNotification `json:"notifications"`
	InstallNotification *InstallHealthNotification    `json:"install_notification"`
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

	resp.InstallName = install.Name

	transitions := make([]app.InstallComponentHealthTransition, 0)
	priorVerdicts := make([]app.InstallComponentHealthStatus, 0, len(installComponents))
	newVerdicts := make([]app.InstallComponentHealthStatus, 0, len(installComponents))

	for i := range installComponents {
		ic := &installComponents[i]
		resp.Evaluated++

		reports := reportsByComponent[ic.ID]
		verdict := a.componentVerdict(ic, reports, now)
		prior := ic.HealthStatus

		priorVerdicts = append(priorVerdicts, prior)
		newVerdicts = append(newVerdicts, verdict)

		var latest *componentHealthReport
		if len(reports) > 0 {
			latest = &reports[0]
		}
		description := componentHealthDescription(verdict, latest)

		if verdict == prior && description == ic.HealthStatusDescription {
			continue
		}

		if err := a.writeComponentHealth(ctx, ic, verdict, description, latest, now); err != nil {
			return nil, errors.Wrapf(err, "unable to update health for install component %s", ic.ID)
		}
		resp.Updated++

		if verdict == prior {
			continue
		}

		transitions = append(transitions, newComponentHealthTransition(ic, verdict, latest, now))

		if n, ok := componentHealthNotification(ic, prior, verdict, description, latest); ok {
			resp.Notifications = append(resp.Notifications, n)
		}
	}

	resp.InstallNotification = installHealthNotification(priorVerdicts, newVerdicts)

	if len(transitions) > 0 {
		resp.Transitions = len(transitions)
		a.enrichTransitions(ctx, install.OrgID, install.ID, transitions, now)
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

// componentHealthDeployWindow is how close a deploy must be to a health
// transition for the two to be considered related. Wide enough to catch a
// workload that only fails once traffic reaches it, narrow enough that an
// unrelated later failure isn't blamed on the deploy.
const componentHealthDeployWindow = 10 * time.Minute

// enrichTransitions adds the story around each transition: the diagnosis the
// runner captured for the resource that caused it, and the deploy it followed
// if one landed recently. Both are best-effort — a transition is still worth
// recording without them, so failures are logged and skipped.
func (a *Activities) enrichTransitions(ctx context.Context, orgID, installID string, transitions []app.InstallComponentHealthTransition, now time.Time) {
	componentIDs := make([]string, 0, len(transitions))
	for i := range transitions {
		componentIDs = append(componentIDs, transitions[i].InstallComponentID)
	}

	diagnoses, err := a.resourceDiagnoses(ctx, orgID, installID, componentIDs)
	if err != nil {
		a.l.Warn("unable to load component health diagnoses",
			zap.String("install_id", installID), zap.Error(err))
	}

	deploys, err := a.recentDeploysByComponent(ctx, componentIDs, now)
	if err != nil {
		a.l.Warn("unable to correlate component health transitions with deploys",
			zap.String("install_id", installID), zap.Error(err))
	}

	for i := range transitions {
		t := &transitions[i]
		if d, ok := diagnoses[transitionResourceKey(t)]; ok {
			t.Diagnosis = d
		}
		if id, ok := deploys[t.InstallComponentID]; ok {
			t.CorrelatedDeployID = id
		}
	}
}

func transitionResourceKey(t *app.InstallComponentHealthTransition) string {
	return strings.Join([]string{
		t.InstallComponentID,
		t.RootResourceKind,
		t.RootResourceNamespace,
		t.RootResourceName,
	}, "\x00")
}

// resourceDiagnoses returns the diagnosis blob the runner attached to each
// currently non-healthy resource of the given components, keyed by resource
// identity. Fetched as its own narrow query rather than widening the main
// observation read, because details is by far the largest column and
// transitions are rare.
func (a *Activities) resourceDiagnoses(ctx context.Context, orgID, installID string, installComponentIDs []string) (map[string]string, error) {
	out := map[string]string{}
	if len(installComponentIDs) == 0 {
		return out, nil
	}

	var rows []app.InstallComponentResourceState
	if err := a.chDB.WithContext(ctx).
		Scopes(scopes.WithOverrideTable(views.CurrentViewName(a.chDB, &app.InstallComponentResourceState{}))).
		Select("install_component_id", "kind", "namespace", "name", "details").
		Where(app.InstallComponentResourceState{OrgID: orgID, InstallID: installID}).
		Where("install_component_id IN ?", installComponentIDs).
		Where("health != ?", string(app.InstallComponentHealthStatusHealthy)).
		Find(&rows).Error; err != nil {
		return out, err
	}

	for _, r := range rows {
		diagnosis := diagnosisFromDetails(r.Details)
		if diagnosis == "" {
			continue
		}
		out[strings.Join([]string{r.InstallComponentID, r.Kind, r.Namespace, r.Name}, "\x00")] = diagnosis
	}
	return out, nil
}

// diagnosisFromDetails extracts the runner's diagnosis object out of a
// resource's details blob, re-serialized on its own. Returns "" when the blob
// has no diagnosis or isn't parseable.
func diagnosisFromDetails(details string) string {
	if details == "" {
		return ""
	}
	var parsed struct {
		Diagnosis json.RawMessage `json:"diagnosis"`
	}
	if err := json.Unmarshal([]byte(details), &parsed); err != nil {
		return ""
	}
	if len(parsed.Diagnosis) == 0 {
		return ""
	}
	return string(parsed.Diagnosis)
}

// recentDeploysByComponent returns the most recent deploy per install component
// that started inside componentHealthDeployWindow, so a transition can be
// labelled with the deploy it followed.
func (a *Activities) recentDeploysByComponent(ctx context.Context, installComponentIDs []string, now time.Time) (map[string]string, error) {
	out := map[string]string{}
	if len(installComponentIDs) == 0 {
		return out, nil
	}

	var rows []struct {
		InstallComponentID string
		ID                 string
	}
	if err := a.db.WithContext(ctx).
		Table("install_deploys").
		Select("install_component_id, id").
		Where("install_component_id IN ?", installComponentIDs).
		Where("created_at > ?", now.Add(-componentHealthDeployWindow)).
		Order("created_at DESC").
		Scan(&rows).Error; err != nil {
		return out, err
	}

	for _, r := range rows {
		if _, seen := out[r.InstallComponentID]; !seen {
			out[r.InstallComponentID] = r.ID
		}
	}
	return out, nil
}

// componentHealthNotification decides whether a component transition is worth
// telling a human about. Only crossings into and out of the bad band notify:
// entering bad is the alert, leaving it is the resolution. A worsening inside
// the band (degraded to unhealthy) is deliberately silent — the subscriber
// already knows the component is broken, and re-alerting on every escalation
// is the flapping we promised not to ship.
func componentHealthNotification(ic *app.InstallComponent, prior, verdict app.InstallComponentHealthStatus, description string, latest *componentHealthReport) (ComponentHealthNotification, bool) {
	enteredBad := verdict.IsBadHealth() && !prior.IsBadHealth()
	recovered := verdict == app.InstallComponentHealthStatusHealthy && prior.IsBadHealth()
	if !enteredBad && !recovered {
		return ComponentHealthNotification{}, false
	}

	n := ComponentHealthNotification{
		Recovered:          recovered,
		InstallComponentID: ic.ID,
		ComponentID:        ic.ComponentID,
		ComponentName:      ic.Component.Name,
		Health:             string(verdict),
		PreviousHealth:     string(prior),
		Message:            description,
	}
	if latest != nil {
		n.RootResourceKind = latest.RootKind
		n.RootResourceNamespace = latest.RootNamespace
		n.RootResourceName = latest.RootName
	}
	return n, true
}

// installHealthNotification returns the install-level rollup crossing, or nil
// when the composite verdict stayed on the same side of the bad band.
func installHealthNotification(prior, current []app.InstallComponentHealthStatus) *InstallHealthNotification {
	priorComposite, _ := app.CompositeComponentHealthStatus(prior)
	composite, description := app.CompositeComponentHealthStatus(current)

	enteredBad := composite.IsBadHealth() && !priorComposite.IsBadHealth()
	recovered := composite == app.InstallComponentHealthStatusHealthy && priorComposite.IsBadHealth()
	if !enteredBad && !recovered {
		return nil
	}

	n := &InstallHealthNotification{
		Health:         string(composite),
		PreviousHealth: string(priorComposite),
		Message:        description,
	}
	for _, s := range current {
		switch s {
		case app.InstallComponentHealthStatusUnhealthy:
			n.UnhealthyComponentCount++
		case app.InstallComponentHealthStatusDegraded:
			n.DegradedComponentCount++
		}
	}
	return n
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
