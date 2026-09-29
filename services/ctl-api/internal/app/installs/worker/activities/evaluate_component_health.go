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
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
)

type EvaluateComponentHealthRequest struct {
	InstallID string `validate:"required"`
}

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
// transition row on every verdict change. No-ops when the install is gone.
//
// @temporal-gen-v2 activity
// @start-to-close-timeout 60s
// @by-field InstallID
func (a *Activities) EvaluateComponentHealth(ctx context.Context, req *EvaluateComponentHealthRequest) (result *EvaluateComponentHealthResponse, err error) {
	started := time.Now()
	reason := "load_install"
	defer func() { a.healthMetrics.recordForInstall(ctx, started, req.InstallID, reason, result, err) }()
	resp := &EvaluateComponentHealthResponse{}

	var install app.Install
	if err := a.db.WithContext(ctx).Where(app.Install{ID: req.InstallID}).First(&install).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			reason = "install_missing"
			resp.Skipped = true
			return resp, nil
		}
		return nil, errors.Wrap(err, "unable to get install")
	}

	reason = "load_components"
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

	reason = "load_observations"
	now := time.Now()
	reportsByComponent, err := a.recentComponentHealthReports(ctx, install.OrgID, install.ID, now)
	if err != nil {
		return nil, errors.Wrap(err, "unable to load resource observations")
	}

	resp.InstallName = install.Name

	// why: Verdicts for every component are needed before any can be written,
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

	reason = "persist_verdicts"
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
			a.l.Warn("unable to record component health transitions",
				zap.String("install_id", install.ID),
				zap.Error(err),
			)
		}
	}

	return resp, nil
}

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

	verdict := nextComponentHealthVerdict(ic.HealthStatus, reports, now)
	return escalateStuckProgressing(verdict, ic, now)
}

// why: escalateStuckProgressing turns a progressing verdict that has not moved in a
// long time into degraded.
//
// Progressing means "on its way", and the resource libraries have no clock, so a
// workload that never becomes ready reports progressing forever. Since
// progressing never alerts, the most durable failure state was also the
// quietest: a live install sat progressing for 15h because its ingress had no
// class and nothing ever told anyone.
func escalateStuckProgressing(verdict app.InstallComponentHealthStatus, ic *app.InstallComponent, now time.Time) app.InstallComponentHealthStatus {
	if verdict != app.InstallComponentHealthStatusProgressing {
		return verdict
	}
	if ic.HealthStatus != app.InstallComponentHealthStatusProgressing || ic.HealthStatusV2.CreatedAtTS <= 0 {
		return verdict
	}
	if now.Sub(time.Unix(ic.HealthStatusV2.CreatedAtTS, 0)) < componentHealthProgressingLimit {
		return verdict
	}
	return app.InstallComponentHealthStatusDegraded
}

func clusterWatchedComponent(t app.ComponentType) bool {
	return t == app.ComponentTypeHelmChart || t == app.ComponentTypeKubernetesManifest
}

// why: componentClusterBlind reports a watched component whose cluster observations
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

type customCheckObservation struct {
	Name              string
	Health            app.InstallComponentHealthStatus
	Message           string
	ObservedAt        time.Time
	StaleAfterSeconds uint32
}

func (o customCheckObservation) staleAfter() time.Duration {
	if o.StaleAfterSeconds == 0 {
		return componentHealthStaleAfter
	}
	return time.Duration(o.StaleAfterSeconds) * time.Second
}

// why: applyCustomChecks merges pushed checks onto the runner's report clock so one
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
				ValidFor:       c.staleAfter(),
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
			// why: Past its TTL the check reads as unknown, but is still counted:
			// dropping it would silently remove a configured check.
			if age := rep.ObservedAt.Sub(state.ObservedAt); age > state.staleAfter() {
				health = app.InstallComponentHealthStatusUnknown
				message = fmt.Sprintf("no report in %s", state.staleAfter())
			}
			rep.Resources++
			rep.ResourceCounts[string(health)]++
			// why: unknown is absence of information, so it must never outrank a
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

func assessedResourceCount(rep *componentHealthReport) int {
	assessed := 0
	for health, n := range rep.ResourceCounts {
		if health != string(app.InstallComponentHealthStatusUnknown) {
			assessed += n
		}
	}
	return assessed
}

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

	var customRows []app.InstallComponentResourceState
	if err := base().
		Where(app.InstallComponentResourceState{Provider: providerCustom}).
		Where("observed_at > ?", now.Add(-customCheckRetentionWindow)).
		Find(&customRows).Error; err != nil {
		return nil, err
	}

	return collapseComponentHealthRows(append(rows, customRows...)), nil
}

func collapseComponentHealthRows(rows []app.InstallComponentResourceState) map[string][]componentHealthReport {
	type reportKey struct {
		componentID string
		observedAt  int64
	}
	customs := map[string][]customCheckObservation{}
	merged := make(map[reportKey]*componentHealthReport)
	knownSeen := map[reportKey]bool{}
	unknownFallback := map[reportKey]app.InstallComponentResourceState{}
	naFallback := map[reportKey]app.InstallComponentResourceState{}

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

		// why: unknown is absence of information, not a severity, so it must never
		// outrank an assessed resource — otherwise one unrunnable probe masks
		// every healthy resource behind it.
		health := app.InstallComponentHealthStatus(r.Health)
		if health == app.InstallComponentHealthStatusUnknown {
			if _, seen := unknownFallback[key]; !seen {
				unknownFallback[key] = r
			}
			continue
		}
		// why: not-applicable is not a severity either: it says this resource has no
		// signal, which must never outrank one that does. It shares unknown's
		// zero severity, so whichever row the store returned first won and the
		// same cluster state reported healthy or not-applicable at random.
		if health == app.InstallComponentHealthStatusNotApplicable {
			if _, seen := naFallback[key]; !seen {
				naFallback[key] = r
			}
			continue
		}

		if !knownSeen[key] || betterRoot(rep, health, r.Kind, r.Namespace, r.Name) {
			knownSeen[key] = true
			rep.Health = health
			rep.RootKind = r.Kind
			rep.RootNamespace = r.Namespace
			rep.RootName = r.Name
			rep.Message = r.Message
		}
	}

	for key, rep := range merged {
		if knownSeen[key] {
			continue
		}
		fallback, ok := unknownFallback[key]
		health := app.InstallComponentHealthStatusUnknown
		if !ok {
			fallback, ok = naFallback[key]
			health = app.InstallComponentHealthStatusNotApplicable
		}
		if !ok {
			continue
		}
		rep.Health = health
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

type componentEval struct {
	ic      *app.InstallComponent
	prior   app.InstallComponentHealthStatus
	verdict app.InstallComponentHealthStatus
	latest  *componentHealthReport

	downstreamOf string

	priorAlerted bool

	clusterEvidence bool
	clusterBlind    bool
}

type healthFlags struct {
	alerted     bool
	clusterSeen bool
}

func componentAlerted(ic *app.InstallComponent) bool {
	v, _ := ic.HealthStatusV2.Metadata["alerted"].(bool)
	return v
}

func (a *Activities) markDownstream(ctx context.Context, install app.Install, evals []componentEval) {
	badCount := 0
	for i := range evals {
		if evals[i].verdict.IsBadHealth() {
			badCount++
		}
	}
	if badCount < 2 {
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
				scopes.WithOverrideTable(app.LatestComponentConfigConnectionsViewName),
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

const componentHealthDeployWindow = 10 * time.Minute

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

func (a *Activities) resourceDiagnoses(ctx context.Context, orgID, installID string, installComponentIDs []string) (map[string]string, error) {
	out := map[string]string{}
	if len(installComponentIDs) == 0 {
		return out, nil
	}

	var rows []app.InstallComponentResourceState
	if err := a.chDB.WithContext(ctx).
		Scopes(scopes.WithOverrideTable(app.InstallComponentResourceStatesLatestView)).
		Select("install_component_id", "kind", "namespace", "name", "details").
		Where(app.InstallComponentResourceState{OrgID: orgID, InstallID: installID}).
		Where(app.LatestReportOnlySQL(), app.LatestReportOnlyArgs(orgID, installID)...).
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

func componentHealthDescriptionFor(e *componentEval, now time.Time) string {
	if e.clusterBlind && e.verdict == app.InstallComponentHealthStatusUnknown {
		return "the runner is no longer reporting this component's cluster resources, so its health cannot be assessed from the remaining checks alone"
	}

	description := componentHealthDescription(e.verdict, e.latest, now)
	if e.downstreamOf != "" {
		description = description + " (downstream of " + e.downstreamOf + ")"
	}
	return description
}

func componentHealthNotificationFor(e *componentEval, description string) (ComponentHealthNotification, bool) {
	if e.downstreamOf != "" {
		return ComponentHealthNotification{}, false
	}
	return componentHealthNotification(e.ic, e.prior, e.verdict, e.priorAlerted, description, e.latest)
}

// why: Alerts and resolutions are strictly paired via the persisted alerted flag, so
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
