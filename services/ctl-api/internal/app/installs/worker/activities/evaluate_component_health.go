package activities

import (
	"context"
	"encoding/json"
	"fmt"
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

// ComponentHealthNotification is a component transition worth notifying on.
// Only crossings into and out of the bad band produce one.
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

// InstallHealthNotification is the install-level rollup crossing, set only when
// the composite health entered or left the bad band.
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

	// Verdicts for every component are needed before any can be written,
	// because dependency root-cause analysis reads the whole set.
	evals := make([]componentEval, 0, len(installComponents))
	for i := range installComponents {
		ic := &installComponents[i]
		resp.Evaluated++

		reports := reportsByComponent[ic.ID]
		var latest *componentHealthReport
		if len(reports) > 0 {
			latest = &reports[0]
		}

		evals = append(evals, componentEval{
			ic:              ic,
			prior:           ic.HealthStatus,
			verdict:         a.componentVerdict(ic, reports, now),
			latest:          latest,
			priorAlerted:    componentAlerted(ic),
			clusterEvidence: anyClusterEvidence(reports),
			clusterBlind:    componentClusterBlind(ic, reports, now),
		})
	}

	a.markDownstream(ctx, install, evals)

	transitions := make([]app.InstallComponentHealthTransition, 0)
	priorVerdicts := make([]app.InstallComponentHealthStatus, 0, len(evals))
	newVerdicts := make([]app.InstallComponentHealthStatus, 0, len(evals))

	for i := range evals {
		e := &evals[i]
		priorVerdicts = append(priorVerdicts, e.prior)
		newVerdicts = append(newVerdicts, e.verdict)

		description := componentHealthDescriptionFor(e, now)

		n, fire := componentHealthNotificationFor(e, description)
		priorFlags := healthFlags{alerted: e.priorAlerted, clusterSeen: componentClusterSeen(e.ic)}
		flags := healthFlags{
			alerted:     e.verdict.IsBadHealth() && (e.priorAlerted || fire),
			clusterSeen: priorFlags.clusterSeen || e.clusterEvidence,
		}

		if e.verdict == e.prior && description == e.ic.HealthStatusDescription && flags == priorFlags {
			continue
		}

		if err := a.writeComponentHealth(ctx, e.ic, e.verdict, description, e.downstreamOf, flags, e.latest, now); err != nil {
			return nil, errors.Wrapf(err, "unable to update health for install component %s", e.ic.ID)
		}
		resp.Updated++

		if fire {
			resp.Notifications = append(resp.Notifications, n)
		}

		if e.verdict == e.prior {
			continue
		}

		transitions = append(transitions, newComponentHealthTransition(e.ic, e.verdict, e.latest, now))
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

// componentVerdict wraps the debounce with the component-level short-circuits.
// Disabled or never-deployed components carry no signal at all, so a
// trivially-passing probe can't report an undeployed component as healthy.
func (a *Activities) componentVerdict(ic *app.InstallComponent, reports []componentHealthReport, now time.Time) app.InstallComponentHealthStatus {
	if ic.Status == app.InstallComponentStatusDisabled || !ic.EverDeployed() {
		return app.InstallComponentHealthStatusNotApplicable
	}

	if !clusterWatchedComponent(ic.Component.Type) && len(reports) == 0 {
		return app.InstallComponentHealthStatusNotApplicable
	}

	if componentClusterBlind(ic, reports, now) {
		return app.InstallComponentHealthStatusUnknown
	}

	return nextComponentHealthVerdict(ic.HealthStatus, reports, now)
}

func clusterWatchedComponent(t app.ComponentType) bool {
	return t == app.ComponentTypeHelmChart || t == app.ComponentTypeKubernetesManifest
}

// componentClusterBlind reports a watched component whose cluster observations
// went stale while probes kept reporting: a passing probe must not certify a
// workload nobody can see. Requires cluster_seen, so probe-only charts still work.
func componentClusterBlind(ic *app.InstallComponent, reports []componentHealthReport, now time.Time) bool {
	if !clusterWatchedComponent(ic.Component.Type) || !componentClusterSeen(ic) {
		return false
	}
	for i := range reports {
		if reports[i].ClusterEvidence && now.Sub(reports[i].ObservedAt) <= componentHealthStaleAfter {
			return false
		}
	}
	return true
}

func componentClusterSeen(ic *app.InstallComponent) bool {
	return ic.ClusterHealthSeen()
}

func anyClusterEvidence(reports []componentHealthReport) bool {
	for i := range reports {
		if reports[i].ClusterEvidence {
			return true
		}
	}
	return false
}

// bearsVerdict reports whether a provider's observations assess live state
// rather than being identity-only inventory. Cloud rows are identity-only today,
// so counting them would drag every terraform component to unknown.
func bearsVerdict(provider string) bool {
	switch provider {
	case providerAWS, providerGCP, providerAzure:
		return false
	}
	return true
}

const (
	providerAWS        = "aws"
	providerGCP        = "gcp"
	providerAzure      = "azure"
	providerCustom     = "custom"
	providerKubernetes = "kubernetes"
)

// customCheckObservation is one reported state of a named custom check.
type customCheckObservation struct {
	Name              string
	Health            app.InstallComponentHealthStatus
	Message           string
	ObservedAt        time.Time
	StaleAfterSeconds uint32
}

// staleAfter is how long this report stands before it reads as unknown.
func (o customCheckObservation) staleAfter() time.Duration {
	if o.StaleAfterSeconds == 0 {
		return componentHealthStaleAfter
	}
	return time.Duration(o.StaleAfterSeconds) * time.Second
}

// applyCustomChecks merges pushed checks onto the runner's report clock so one
// can't become the newest observation and discard what the runner saw.
func applyCustomChecks(reports []componentHealthReport, customs []customCheckObservation) []componentHealthReport {
	if len(customs) == 0 {
		return reports
	}

	byName := map[string][]customCheckObservation{}
	for _, c := range customs {
		byName[c.Name] = append(byName[c.Name], c)
	}
	for name := range byName {
		sort.Slice(byName[name], func(i, j int) bool {
			return byName[name][i].ObservedAt.Before(byName[name][j].ObservedAt)
		})
	}

	if len(reports) == 0 {
		seen := map[int64]bool{}
		for _, c := range customs {
			if seen[c.ObservedAt.UnixNano()] {
				continue
			}
			seen[c.ObservedAt.UnixNano()] = true
			reports = append(reports, componentHealthReport{
				ObservedAt:     c.ObservedAt,
				Health:         app.InstallComponentHealthStatusHealthy,
				ResourceCounts: map[string]int{},
			})
		}
		sort.Slice(reports, func(i, j int) bool {
			return reports[i].ObservedAt.After(reports[j].ObservedAt)
		})
	}

	for i := range reports {
		rep := &reports[i]
		for name, obs := range byName {
			state, ok := customStateAt(obs, rep.ObservedAt)
			if !ok {
				continue
			}
			health, message := state.Health, state.Message
			// Past its TTL the check reads as unknown, but is still counted:
			// dropping it would silently remove a configured check.
			if age := rep.ObservedAt.Sub(state.ObservedAt); age > state.staleAfter() {
				health = app.InstallComponentHealthStatusUnknown
				message = fmt.Sprintf("no report in %s", state.staleAfter())
			}
			rep.Resources++
			rep.ResourceCounts[string(health)]++
			// unknown is absence of information, so it must never outrank a
			// check that did report.
			if health == app.InstallComponentHealthStatusUnknown {
				continue
			}
			if componentHealthSeverity[health] > componentHealthSeverity[rep.Health] {
				rep.Health = health
				rep.RootKind = "CustomCheck"
				rep.RootNamespace = ""
				rep.RootName = name
				rep.Message = message
			}
		}

		if rep.Resources > 0 && assessedResourceCount(rep) == 0 {
			rep.Health = app.InstallComponentHealthStatusUnknown
		}
	}

	return reports
}

// assessedResourceCount is how many of a report's resources carry a real verdict.
func assessedResourceCount(rep *componentHealthReport) int {
	assessed := 0
	for health, n := range rep.ResourceCounts {
		if health != string(app.InstallComponentHealthStatusUnknown) {
			assessed += n
		}
	}
	return assessed
}

// customStateAt returns the newest observation at or before t.
func customStateAt(obs []customCheckObservation, t time.Time) (customCheckObservation, bool) {
	var out customCheckObservation
	found := false
	for _, o := range obs {
		if o.ObservedAt.After(t) {
			break
		}
		out = o
		found = true
	}
	return out, found
}

// recentComponentHealthReports reads the observation window from ClickHouse and
// collapses it to one report (worst resource) per component per timestamp.
func (a *Activities) recentComponentHealthReports(ctx context.Context, orgID, installID string, now time.Time) (map[string][]componentHealthReport, error) {
	cols := []string{"install_component_id", "provider", "kind", "namespace", "name", "health", "message", "native_status", "observed_at", "stale_after_seconds"}
	base := func() *gorm.DB {
		return a.chDB.WithContext(ctx).
			Select(cols).
			Where(app.InstallComponentResourceState{
				OrgID:     orgID,
				InstallID: installID,
				Source:    app.InstallComponentResourceSourceComponent,
			})
	}

	var rows []app.InstallComponentResourceState
	if err := base().
		Where("provider != ?", providerCustom).
		Where("observed_at > ?", now.Add(-componentHealthObservationWindow)).
		Find(&rows).Error; err != nil {
		return nil, err
	}

	// Pushed checks declare their own TTL, so they read over the longest TTL we
	// honour — separate query so the runner-report path picks up no extra rows.
	var customRows []app.InstallComponentResourceState
	if err := base().
		Where(app.InstallComponentResourceState{Provider: providerCustom}).
		Where("observed_at > ?", now.Add(-customCheckRetentionWindow)).
		Find(&customRows).Error; err != nil {
		return nil, err
	}

	return collapseComponentHealthRows(append(rows, customRows...)), nil
}

// collapseComponentHealthRows folds raw resource observations into one report
// (the worst resource) per component per timestamp, newest first.
func collapseComponentHealthRows(rows []app.InstallComponentResourceState) map[string][]componentHealthReport {
	type reportKey struct {
		componentID string
		observedAt  int64
	}
	customs := map[string][]customCheckObservation{}
	merged := make(map[reportKey]*componentHealthReport)
	// unknownFallback keeps one unassessable resource per report, used only when
	// nothing in that report could be assessed.
	knownSeen := map[reportKey]bool{}
	unknownFallback := map[reportKey]app.InstallComponentResourceState{}

	for _, r := range rows {
		if !bearsVerdict(r.Provider) {
			continue
		}
		if r.Provider == providerCustom {
			customs[r.InstallComponentID] = append(customs[r.InstallComponentID], customCheckObservation{
				Name:              r.Name,
				Health:            app.InstallComponentHealthStatus(r.Health),
				Message:           r.Message,
				ObservedAt:        r.ObservedAt,
				StaleAfterSeconds: r.StaleAfterSeconds,
			})
			continue
		}

		key := reportKey{componentID: r.InstallComponentID, observedAt: r.ObservedAt.UnixNano()}
		rep, ok := merged[key]
		if !ok {
			rep = &componentHealthReport{
				ObservedAt:     r.ObservedAt,
				ResourceCounts: map[string]int{},
			}
			merged[key] = rep
		}
		if rep.NativeStatus == "" {
			rep.NativeStatus = r.NativeStatus
		}
		if r.Provider == providerKubernetes {
			rep.ClusterEvidence = true
		}
		rep.Resources++
		rep.ResourceCounts[r.Health]++

		// unknown is absence of information, not a severity, so it must never
		// outrank an assessed resource — otherwise one unrunnable probe masks
		// every healthy resource behind it.
		health := app.InstallComponentHealthStatus(r.Health)
		if health == app.InstallComponentHealthStatusUnknown {
			if _, seen := unknownFallback[key]; !seen {
				unknownFallback[key] = r
			}
			continue
		}

		if !knownSeen[key] || componentHealthSeverity[health] > componentHealthSeverity[rep.Health] {
			knownSeen[key] = true
			rep.Health = health
			rep.RootKind = r.Kind
			rep.RootNamespace = r.Namespace
			rep.RootName = r.Name
			rep.Message = r.Message
		}
	}

	// Only a report in which nothing at all could be assessed is unknown.
	for key, rep := range merged {
		if knownSeen[key] {
			continue
		}
		fallback, ok := unknownFallback[key]
		if !ok {
			continue
		}
		rep.Health = app.InstallComponentHealthStatusUnknown
		rep.RootKind = fallback.Kind
		rep.RootNamespace = fallback.Namespace
		rep.RootName = fallback.Name
		rep.Message = fallback.Message
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

	for id, obs := range customs {
		out[id] = applyCustomChecks(out[id], obs)
	}

	return out
}

func (a *Activities) writeComponentHealth(ctx context.Context, ic *app.InstallComponent, verdict app.InstallComponentHealthStatus, description, downstreamOf string, flags healthFlags, latest *componentHealthReport, now time.Time) error {
	metadata := map[string]any{}
	// Structured twin of the "(downstream of X)" description suffix so UIs don't
	// have to parse text.
	if downstreamOf != "" {
		metadata["downstream_of"] = downstreamOf
	}
	if flags.alerted {
		metadata["alerted"] = true
	}
	if flags.clusterSeen {
		metadata["cluster_seen"] = true
	}
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

	// CreatedAtTS marks when the current verdict began, not when it was last
	// written: the verified-deploy gate measures hold time from it.
	startedAt := now.Unix()
	if verdict == ic.HealthStatus && ic.HealthStatusV2.CreatedAtTS > 0 {
		startedAt = ic.HealthStatusV2.CreatedAtTS
	}

	return a.db.WithContext(ctx).
		Model(&app.InstallComponent{ID: ic.ID}).
		Select("health_status", "health_status_v2").
		Updates(app.InstallComponent{
			HealthStatus: verdict,
			HealthStatusV2: app.CompositeStatus{
				CreatedAtTS:            startedAt,
				Status:                 app.Status(verdict),
				StatusHumanDescription: description,
				Metadata:               metadata,
			},
		}).Error
}

// componentEval is one component's pending evaluation, held before any write so
// dependency root-cause analysis can see the whole set at once.
type componentEval struct {
	ic      *app.InstallComponent
	prior   app.InstallComponentHealthStatus
	verdict app.InstallComponentHealthStatus
	latest  *componentHealthReport

	// downstreamOf names the unhealthy dependency this failure is most likely a
	// consequence of. Empty means this component may alert.
	downstreamOf string

	// priorAlerted is whether the current bad spell already alerted, which is
	// what pairs alerts with resolutions across evaluations.
	priorAlerted bool

	// clusterBlind is whether the absence of cluster observations forced the
	// verdict to unknown.
	clusterEvidence bool
	clusterBlind    bool
}

// healthFlags are sticky bits carried in health metadata across evaluations, so
// they must be read back and rewritten rather than derived fresh.
type healthFlags struct {
	alerted     bool
	clusterSeen bool
}

func componentAlerted(ic *app.InstallComponent) bool {
	v, _ := ic.HealthStatusV2.Metadata["alerted"].(bool)
	return v
}

// markDownstream labels each bad component that has a bad dependency so its
// alert is suppressed. One hop only, keeping the label on the nearest thing to
// investigate. Failing to load the graph over-alerts rather than going silent.
func (a *Activities) markDownstream(ctx context.Context, install app.Install, evals []componentEval) {
	badCount := 0
	for i := range evals {
		if evals[i].verdict.IsBadHealth() {
			badCount++
		}
	}
	if badCount < 2 {
		// A single failure is always its own root cause.
		return
	}

	componentIDs := make([]string, 0, len(evals))
	for i := range evals {
		componentIDs = append(componentIDs, evals[i].ic.ComponentID)
	}
	deps, err := a.componentDependencies(ctx, install.AppConfigID, componentIDs)
	if err != nil {
		a.l.Warn("unable to load component dependencies for health root-cause",
			zap.String("install_id", install.ID), zap.Error(err))
		return
	}

	markDownstreamWithDeps(evals, deps)
}

// markDownstreamWithDeps is the pure graph half of markDownstream.
func markDownstreamWithDeps(evals []componentEval, deps map[string][]string) {
	bad := map[string]*componentEval{}
	names := map[string]string{}
	for i := range evals {
		names[evals[i].ic.ComponentID] = evals[i].ic.Component.Name
		if evals[i].verdict.IsBadHealth() {
			bad[evals[i].ic.ComponentID] = &evals[i]
		}
	}
	if len(bad) < 2 {
		return
	}

	for componentID, e := range bad {
		for _, depID := range deps[componentID] {
			if _, depIsBad := bad[depID]; !depIsBad {
				continue
			}
			name := names[depID]
			if name == "" {
				name = depID
			}
			e.downstreamOf = name
			break
		}
	}
}

// componentDependencies returns componentID -> dependency component IDs for
// the install's app config.
func (a *Activities) componentDependencies(ctx context.Context, appConfigID string, componentIDs []string) (map[string][]string, error) {
	out := map[string][]string{}
	seen := map[string]bool{}

	if appConfigID != "" {
		var conns []app.ComponentConfigConnection
		if err := a.db.WithContext(ctx).
			Select("component_id", "component_dependency_ids").
			Where(app.ComponentConfigConnection{AppConfigID: appConfigID}).
			Find(&conns).Error; err != nil {
			return out, err
		}
		for _, c := range conns {
			seen[c.ComponentID] = true
			if len(c.ComponentDependencyIDs) > 0 {
				out[c.ComponentID] = c.ComponentDependencyIDs
			}
		}
	}

	// ccc rows are deltas — an app config version only carries rows for
	// components changed in that sync, so the pin alone can yield an empty graph.
	var missing []string
	for _, id := range componentIDs {
		if !seen[id] {
			missing = append(missing, id)
		}
	}
	if len(missing) > 0 {
		var fallback []app.ComponentConfigConnection
		if err := a.db.WithContext(ctx).
			Scopes(
				scopes.WithDisableViews,
				scopes.WithOverrideTable("component_config_connections_latest_configs_view"),
			).
			Select("component_id", "component_dependency_ids").
			Where("component_id IN ?", missing).
			Find(&fallback).Error; err != nil {
			return out, err
		}
		for _, c := range fallback {
			if len(c.ComponentDependencyIDs) > 0 {
				out[c.ComponentID] = c.ComponentDependencyIDs
			}
		}
	}
	return out, nil
}

// componentHealthDeployWindow is how close a deploy must be to a health
// transition for the two to be considered related.
const componentHealthDeployWindow = 10 * time.Minute

// enrichTransitions attaches the runner's diagnosis and the deploy a transition
// followed. Both best-effort — a transition is worth recording without them.
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

// resourceDiagnoses returns the runner's diagnosis per non-healthy resource,
// keyed by resource identity. Its own query because details is the largest
// column and transitions are rare.
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
// resource's details blob, or "" when there is none.
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
// that started inside componentHealthDeployWindow.
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

// componentHealthDescriptionFor renders an eval's description, naming the
// unhealthy dependency when this failure is downstream of one.
func componentHealthDescriptionFor(e *componentEval, now time.Time) string {
	// The generic unknown wording would point the reader at the wrong thing here,
	// since the remaining checks are reporting and healthy.
	if e.clusterBlind && e.verdict == app.InstallComponentHealthStatusUnknown {
		return "the runner is no longer reporting this component's cluster resources, so its health cannot be assessed from the remaining checks alone"
	}

	description := componentHealthDescription(e.verdict, e.latest, now)
	if e.downstreamOf != "" {
		description = description + " (downstream of " + e.downstreamOf + ")"
	}
	return description
}

// componentHealthNotificationFor decides whether an eval alerts. Downstream
// failures are suppressed so one root cause produces one alert, but the verdict
// itself stays truthful.
func componentHealthNotificationFor(e *componentEval, description string) (ComponentHealthNotification, bool) {
	if e.downstreamOf != "" {
		return ComponentHealthNotification{}, false
	}
	return componentHealthNotification(e.ic, e.prior, e.verdict, e.priorAlerted, description, e.latest)
}

// Alerts and resolutions are strictly paired via the persisted alerted flag, so
// a suppressed failure never sends an orphan "recovered" and a component still
// broken after its root cause recovers fires its own late alert.
func componentHealthNotification(ic *app.InstallComponent, prior, verdict app.InstallComponentHealthStatus, priorAlerted bool, description string, latest *componentHealthReport) (ComponentHealthNotification, bool) {
	alerting := verdict.IsBadHealth() && !priorAlerted
	recovered := verdict == app.InstallComponentHealthStatusHealthy && prior.IsBadHealth() && priorAlerted
	if !alerting && !recovered {
		return ComponentHealthNotification{}, false
	}

	n := ComponentHealthNotification{
		Recovered:          recovered,
		InstallComponentID: ic.ID,
		ComponentID:        ic.ComponentID,
		ComponentName:      ic.Component.Name,
		Health:             string(verdict),
		Message:            description,
	}
	// A late alert has no transition to report, and "previously unhealthy" on an
	// unhealthy alert is noise.
	if prior != verdict {
		n.PreviousHealth = string(prior)
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
