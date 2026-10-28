package statetelemetry

import (
	"time"

	"go.opentelemetry.io/otel/metric"
)

var orgStates = []string{"active", "error", "provisioning", "deleting", "deprovisioning", "deprovisioned", "unset", "other"}
var appStates = []string{"active", "error", "provisioning", "deprovisioning", "delete_queued", "unset", "other"}
var installStates = []string{"provisioning", "provisioned", "deprovisioning", "deprovisioned", "reprovisioning", "unset", "other"}
var componentStates = []string{"active", "error", "deprovisioning", "delete_queued", "unset", "other"}
var targetStates = []string{"deleted", "delete_failed", "queued", "active", "inactive", "error", "noop", "planning", "syncing", "executing", "unknown", "pending", "disabled", "unset", "other"}
var runnerStates = []string{"error", "active", "pending", "provisioning", "deprovisioning", "deprovisioned", "reprovisioning", "offline", "awaiting-install-stack-run", "awaiting-heartbeat", "disabled", "unknown", "unset", "other", "missing"}

const retainedResourcesSQL = `
WITH live_orgs AS (
    SELECT id, sandbox_mode, COALESCE(NULLIF(status_v2->>'status', ''), status, '') AS state
    FROM orgs WHERE deleted_at = 0
), live_apps AS (
    SELECT a.id, a.org_id, o.sandbox_mode, COALESCE(NULLIF(a.status_v2->>'status', ''), a.status, '') AS state
    FROM apps a JOIN live_orgs o ON o.id = a.org_id WHERE a.deleted_at = 0
), live_installs AS (
    SELECT i.id, i.org_id, i.app_id, i.last_health_report_at,
        COALESCE(i.sandbox_mode, a.sandbox_mode, false) AS sandbox_mode,
        COALESCE(i.lifecycle_phase->>'phase', '') AS state,
        COALESCE(i.sandbox_health_status, '') AS sandbox_health
    FROM installs i JOIN live_apps a ON a.id = i.app_id AND a.org_id = i.org_id WHERE i.deleted_at = 0
), live_components AS (
    SELECT c.id, c.org_id, c.app_id, COALESCE(NULLIF(c.status_v2->>'status', ''), c.status, '') AS state
    FROM components c JOIN live_apps a ON a.id = c.app_id AND a.org_id = c.org_id WHERE c.deleted_at = 0
), live_targets AS (
    SELECT ic.id, ic.org_id, i.app_id, ic.install_id,
        COALESCE(NULLIF(ic.status_v2->>'status', ''), ic.status, '') AS state,
        COALESCE(NULLIF(ic.health_status_v2->>'status', ''), ic.health_status, '') AS health
    FROM install_components ic
    JOIN live_installs i ON i.id = ic.install_id AND i.org_id = ic.org_id
    JOIN live_components c ON c.id = ic.component_id AND c.org_id = ic.org_id AND c.app_id = i.app_id
    WHERE ic.deleted_at = 0
)
`

func newStateFamilies(meter metric.Meter, level string, labelKeys []string) ([]*family, error) {
	definitions := []struct {
		name          string
		interval, ttl time.Duration
		metrics       []populationMetric
	}{
		{"definitions", 5 * time.Minute, 15 * time.Minute, []populationMetric{
			{name: "nuon.org.count", unit: "{org}", group: "org", dimension: "resource.state", states: orgStates, empty: "unset",
				description: "Retained organizations by effective lifecycle state.",
				source:      "SELECT id AS org_id, '' AS app_id, '' AS install_id, state FROM live_orgs"},
			{name: "nuon.app.count", unit: "{app}", group: "app", dimension: "resource.state", states: appStates, empty: "unset",
				description: "Retained applications with matching retained organizations, by effective lifecycle state.",
				source:      "SELECT org_id, id AS app_id, '' AS install_id, state FROM live_apps"},
			{name: "nuon.component.count", unit: "{component}", group: "app", dimension: "resource.state", states: componentStates, empty: "unset",
				description: "Retained app-defined components by effective lifecycle state; not installed targets.",
				source:      "SELECT org_id, app_id, '' AS install_id, state FROM live_components"},
		}},
		{"installs", time.Minute, 3 * time.Minute, []populationMetric{
			{name: "nuon.install.count", unit: "{install}", group: "install", dimension: "resource.state", states: installStates, empty: "unset",
				description: "Retained installs with tenant-consistent parents, by lifecycle phase.",
				source:      "SELECT org_id, app_id, id AS install_id, state FROM live_installs"},
			{name: "nuon.install.health.count", unit: "{install}", group: "install", dimension: "nuon.health.status", states: healthStates, empty: "unset",
				secondDimension: "nuon.install.lifecycle", secondStates: installStates, secondEmpty: "unset",
				description: "Installs by recorded component-health rollup and lifecycle phase; not live availability or evaluation freshness.", source: installHealthSQL},
			{name: "nuon.install.health.report_age.count", unit: "{install}", group: "install", dimension: "nuon.health.report_age", states: reportAgeStates, empty: "never_reported",
				description: "Installs by recorded health-report arrival age at the transaction clock; arrival does not prove successful processing.", source: reportAgeSQL},
			installInfoMetric(labelKeys),
			{name: "nuon.install.health.last_report_at", unit: "s", detail: true, timestamp: true,
				description: "Unix seconds of the recorded health-report arrival; absent when missing, nonpositive or future; not evaluation freshness.",
				source: `SELECT org_id, app_id, id, '', EXTRACT(EPOCH FROM last_health_report_at)::double precision
FROM live_installs WHERE last_health_report_at > 'epoch'::timestamptz AND last_health_report_at <= transaction_timestamp()`},
		}},
		{"components", time.Minute, 3 * time.Minute, []populationMetric{
			{name: "nuon.install.component.count", unit: "{component}", group: "install", dimension: "resource.state", states: targetStates, empty: "unset",
				description: "Retained installed targets with tenant-consistent install and component parents, by effective lifecycle state.",
				source:      "SELECT org_id, app_id, install_id, state FROM live_targets"},
			{name: "nuon.install.component.health.count", unit: "{component}", group: "install", dimension: "nuon.health.status", states: healthStates, empty: "unset",
				description: "Retained installed targets by effective recorded health verdict, including unevaluated targets.",
				source:      "SELECT org_id, app_id, install_id, health AS state FROM live_targets"},
		}},
		{"runners", time.Minute, 3 * time.Minute, []populationMetric{
			{name: "nuon.install.runner.count", unit: "{install}", group: "install", dimension: "nuon.runner.status", states: runnerStates, empty: "missing",
				description: "Installs by their runner's recorded status, missing when absent, or unknown when no unique runner can be identified; not process liveness.", source: installRunnerSQL},
		}},
		{"sandboxes", time.Minute, 3 * time.Minute, []populationMetric{
			{name: "nuon.install.sandbox.lifecycle.count", unit: "{install}", group: "install", dimension: "nuon.sandbox.status", states: sandboxStates, empty: "missing",
				description: "Installs by retained sandbox lifecycle, including missing sandboxes; not runtime health.", source: sandboxLifecycleSQL},
			{name: "nuon.install.sandbox.health.count", unit: "{install}", group: "install", dimension: "nuon.health.status", states: healthStates, empty: "unset",
				description: "Installs by recorded sandbox runtime-health field; unset does not mean healthy.",
				source:      "SELECT org_id, app_id, id AS install_id, sandbox_health AS state FROM live_installs"},
		}},
	}
	families := make([]*family, 0, len(definitions))
	for _, definition := range definitions {
		f, err := newPopulationFamily(meter, level, definition.name, definition.interval, definition.ttl, definition.metrics...)
		if err != nil {
			return nil, err
		}
		families = append(families, f)
	}
	workflows, err := newWorkflowFamily(meter, level)
	if err != nil {
		return nil, err
	}
	deployments, err := newDeploymentFamily(meter, level)
	if err != nil {
		return nil, err
	}
	families = append(families, workflows, deployments)
	if level != "fleet" {
		configurations, err := newConfigurationFamily(meter, level)
		if err != nil {
			return nil, err
		}
		families = append(families, configurations)
	}
	return families, nil
}

const installRunnerSQL = `
SELECT i.org_id, i.app_id, i.id AS install_id,
    CASE
        WHEN COUNT(DISTINCT rg.id) > 1 OR COUNT(r.id) > 1 THEN 'unknown'
        WHEN COUNT(r.id) = 0 THEN 'missing'
        ELSE MAX(COALESCE(NULLIF(r.status_v2->>'status', ''), r.status, ''))
    END AS state
FROM live_installs i
LEFT JOIN runner_groups rg ON rg.owner_id = i.id AND rg.owner_type = 'installs' AND rg.type = 'install' AND rg.org_id = i.org_id AND rg.deleted_at = 0
LEFT JOIN runners r ON r.runner_group_id = rg.id AND r.org_id = i.org_id AND r.deleted_at = 0
GROUP BY i.org_id, i.app_id, i.id
`
