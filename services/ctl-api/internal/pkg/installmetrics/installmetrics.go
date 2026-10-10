package installmetrics

import (
	"context"
	"database/sql"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

const queryTimeout = 20 * time.Second

var deployWorkflowTypes = []app.WorkflowType{
	app.WorkflowTypeProvision,
	app.WorkflowTypeReprovision,
	app.WorkflowTypeReprovisionSandbox,
	app.WorkflowTypeReprovisionStack,
	app.WorkflowTypeManualDeploy,
	app.WorkflowTypeDeployComponents,
	app.WorkflowTypeComponentEnabled,
	app.WorkflowTypeAppBranchConfigUpdate,
	app.WorkflowTypeAppBranchesRun,
	app.WorkflowTypeAppBranchesConfigRepoUpdate,
	app.WorkflowTypeAppBranchesComponentRepoUpdate,
}

type Params struct {
	fx.In

	LC            fx.Lifecycle
	Cfg           *internal.Config
	DB            *gorm.DB `name:"psql"`
	L             *zap.Logger
	MeterProvider metric.MeterProvider
}

type instruments struct {
	info             metric.Int64ObservableGauge
	health           metric.Int64ObservableGauge
	components       metric.Int64ObservableGauge
	healthLastReport metric.Float64ObservableGauge
	runner           metric.Int64ObservableGauge
	sandbox          metric.Int64ObservableGauge
	configVersion    metric.Int64ObservableGauge
	branchVersion    metric.Int64ObservableGauge
	branchCreated    metric.Float64ObservableGauge
	deployStatus     metric.Int64ObservableGauge
	deployCreated    metric.Float64ObservableGauge
	deployFinished   metric.Float64ObservableGauge
}

type collector struct {
	db     *gorm.DB
	l      *zap.Logger
	labels []string
	inst   instruments
}

func Register(params Params) error {
	if !params.Cfg.InstallStateMetricsEnabled {
		return nil
	}

	meter := params.MeterProvider.Meter("github.com/nuonco/nuon/services/ctl-api/installmetrics")
	c := &collector{db: params.DB, l: params.L, labels: params.Cfg.InstallStateMetricsLabels}
	var err error
	gauges := []struct {
		dst  *metric.Int64ObservableGauge
		name string
		desc string
		unit string
	}{
		{&c.inst.info, "nuon.install.info", "Always 1; carries the install's name, app branch and allowlisted labels.", ""},
		{&c.inst.health, "nuon.install.health.status", "1 for the install's current composite health status.", ""},
		{&c.inst.components, "nuon.install.components", "Install components per health status.", "{component}"},
		{&c.inst.runner, "nuon.install.runner.status", "1 for the current status of the install's runner.", ""},
		{&c.inst.sandbox, "nuon.install.sandbox.status", "1 for the current status of the install's sandbox.", ""},
		{&c.inst.configVersion, "nuon.install.app_config.version", "Version of the app config the install runs.", ""},
		{&c.inst.branchVersion, "nuon.app_branch.latest_config.version", "Version of the newest app config on the app branch.", ""},
		{&c.inst.deployStatus, "nuon.install.deploy.latest.status", "1 for the type and status of the install's newest deploy workflow.", ""},
	}
	for _, g := range gauges {
		opts := []metric.Int64ObservableGaugeOption{metric.WithDescription(g.desc)}
		if g.unit != "" {
			opts = append(opts, metric.WithUnit(g.unit))
		}
		if *g.dst, err = meter.Int64ObservableGauge(g.name, opts...); err != nil {
			return err
		}
	}
	timestamps := []struct {
		dst  *metric.Float64ObservableGauge
		name string
		desc string
	}{
		{&c.inst.healthLastReport, "nuon.install.health.last_report", "Unix timestamp of the runner's last component health report."},
		{&c.inst.branchCreated, "nuon.app_branch.latest_config.created", "Unix timestamp of the newest app config on the app branch."},
		{&c.inst.deployCreated, "nuon.install.deploy.latest.created", "Unix timestamp when the install's newest deploy workflow was created."},
		{&c.inst.deployFinished, "nuon.install.deploy.latest.finished", "Unix timestamp when the install's newest deploy workflow finished; absent while it runs."},
	}
	for _, g := range timestamps {
		if *g.dst, err = meter.Float64ObservableGauge(g.name, metric.WithUnit("s"), metric.WithDescription(g.desc)); err != nil {
			return err
		}
	}

	reg, err := meter.RegisterCallback(c.observe,
		c.inst.info, c.inst.health, c.inst.components, c.inst.healthLastReport, c.inst.runner, c.inst.sandbox,
		c.inst.configVersion, c.inst.branchVersion, c.inst.branchCreated,
		c.inst.deployStatus, c.inst.deployCreated, c.inst.deployFinished,
	)
	if err != nil {
		return err
	}
	params.LC.Append(fx.Hook{OnStop: func(context.Context) error { return reg.Unregister() }})
	return nil
}

func (c *collector) observe(ctx context.Context, o metric.Observer) error {
	ctx, cancel := context.WithTimeout(ctx, queryTimeout)
	defer cancel()

	installs, err := c.installs(ctx)
	if err != nil {
		c.l.Warn("unable to collect install state metrics", zap.Error(err))
		return nil
	}

	appIDs := make([]string, 0, len(installs))
	installIDs := make([]string, 0, len(installs))
	configIDs := make([]string, 0, len(installs))
	seenApps := map[string]struct{}{}
	for i := range installs {
		install := &installs[i]
		installIDs = append(installIDs, install.ID)
		if install.AppConfigID != "" {
			configIDs = append(configIDs, install.AppConfigID)
		}
		if _, ok := seenApps[install.AppID]; !ok {
			seenApps[install.AppID] = struct{}{}
			appIDs = append(appIDs, install.AppID)
		}
		c.observeInstall(o, install)
	}
	if len(installs) == 0 {
		return nil
	}

	if versions, err := c.configVersions(ctx, appIDs, configIDs); err != nil {
		c.l.Warn("unable to collect install app config versions", zap.Error(err))
	} else {
		for i := range installs {
			if version, ok := versions[installs[i].AppConfigID]; ok {
				o.ObserveInt64(c.inst.configVersion, version, metric.WithAttributes(installAttrs(&installs[i])...))
			}
		}
	}

	if heads, err := c.branchHeads(ctx, appIDs); err != nil {
		c.l.Warn("unable to collect app branch config versions", zap.Error(err))
	} else {
		for _, head := range heads {
			attrs := metric.WithAttributes(
				attribute.String("nuon.org.id", head.OrgID),
				attribute.String("nuon.app.id", head.AppID),
				attribute.String("nuon.app_branch.name", head.Name),
			)
			o.ObserveInt64(c.inst.branchVersion, head.Version, attrs)
			o.ObserveFloat64(c.inst.branchCreated, unixSeconds(head.CreatedAt), attrs)
		}
	}

	if deploys, err := c.latestDeploys(ctx, installIDs); err != nil {
		c.l.Warn("unable to collect install deploy workflows", zap.Error(err))
	} else {
		for i := range installs {
			deploy, ok := deploys[installs[i].ID]
			if !ok {
				continue
			}
			base := installAttrs(&installs[i])
			o.ObserveInt64(c.inst.deployStatus, 1, metric.WithAttributes(append(base,
				attribute.String("nuon.workflow.type", deploy.Type),
				attribute.String("nuon.workflow.status", valueOr(deploy.Status, "unknown")),
			)...))
			o.ObserveFloat64(c.inst.deployCreated, unixSeconds(deploy.CreatedAt), metric.WithAttributes(base...))
			if deploy.FinishedAt.Valid {
				o.ObserveFloat64(c.inst.deployFinished, unixSeconds(deploy.FinishedAt.Time), metric.WithAttributes(base...))
			}
		}
	}
	return nil
}

func (c *collector) observeInstall(o metric.Observer, install *app.Install) {
	base := installAttrs(install)

	info := append(base, attribute.String("nuon.install.name", install.Name))
	if install.AppBranch != nil && install.AppBranch.Name != "" {
		info = append(info, attribute.String("nuon.app_branch.name", install.AppBranch.Name))
	}
	for _, key := range c.labels {
		if value, ok := install.Labels[key]; ok {
			info = append(info, attribute.String("nuon.install.label."+key, value))
		}
	}
	o.ObserveInt64(c.inst.info, 1, metric.WithAttributes(info...))

	o.ObserveInt64(c.inst.health, 1, metric.WithAttributes(append(base,
		attribute.String("nuon.health.status", healthStatus(string(install.CompositeHealthStatus))))...))

	counts := map[string]int64{}
	for _, status := range install.ComponentHealthStatuses {
		if status != nil {
			counts[healthStatus(*status)]++
		}
	}
	for status, n := range counts {
		o.ObserveInt64(c.inst.components, n, metric.WithAttributes(append(base,
			attribute.String("nuon.health.status", status))...))
	}

	if install.LastHealthReportAt != nil {
		o.ObserveFloat64(c.inst.healthLastReport, unixSeconds(*install.LastHealthReportAt), metric.WithAttributes(base...))
	}

	o.ObserveInt64(c.inst.runner, 1, metric.WithAttributes(append(base,
		attribute.String("nuon.runner.status", valueOr(string(install.RunnerStatus), "unknown")))...))

	if install.SandboxStatus != "" {
		o.ObserveInt64(c.inst.sandbox, 1, metric.WithAttributes(append(base,
			attribute.String("nuon.sandbox.status", string(install.SandboxStatus)))...))
	}
}

func (c *collector) installs(ctx context.Context) ([]app.Install, error) {
	var installs []app.Install
	err := c.db.WithContext(ctx).
		Preload("Org").
		Preload("RunnerGroup.Runners").
		Preload("AppBranchConnections", func(db *gorm.DB) *gorm.DB {
			return db.Where(app.InstallAppBranchConnection{Active: true}).Order("created_at DESC, id DESC")
		}).
		Preload("AppBranchConnections.AppBranch").
		Find(&installs).Error
	return installs, err
}

func (c *collector) configVersions(ctx context.Context, appIDs, configIDs []string) (map[string]int64, error) {
	out := map[string]int64{}
	if len(configIDs) == 0 {
		return out, nil
	}
	var rows []struct {
		ID      string
		Version int64
	}
	err := c.db.WithContext(ctx).Raw(`
SELECT id, version FROM (
	SELECT id, row_number() OVER (PARTITION BY app_id ORDER BY created_at) AS version
	FROM app_configs
	WHERE app_id IN ?
) v
WHERE id IN ?`, appIDs, configIDs).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		out[row.ID] = row.Version
	}
	return out, nil
}

type branchHead struct {
	OrgID     string
	AppID     string
	Name      string
	Version   int64
	CreatedAt time.Time
}

func (c *collector) branchHeads(ctx context.Context, appIDs []string) ([]branchHead, error) {
	var heads []branchHead
	err := c.db.WithContext(ctx).Raw(`
SELECT DISTINCT ON (v.app_id, v.app_branch_id)
	b.org_id, v.app_id, b.name, v.version, v.created_at
FROM (
	SELECT app_id, app_branch_id, created_at,
		row_number() OVER (PARTITION BY app_id ORDER BY created_at) AS version
	FROM app_configs
	WHERE app_id IN ?
) v
JOIN app_branches b ON b.id = v.app_branch_id AND b.deleted_at = 0
ORDER BY v.app_id, v.app_branch_id, v.created_at DESC`, appIDs).Scan(&heads).Error
	return heads, err
}

type deployRow struct {
	OwnerID    string
	Type       string
	Status     string
	CreatedAt  time.Time
	FinishedAt sql.NullTime
}

func (c *collector) latestDeploys(ctx context.Context, installIDs []string) (map[string]deployRow, error) {
	var rows []deployRow
	err := c.db.WithContext(ctx).Raw(`
SELECT DISTINCT ON (owner_id)
	owner_id, type, status->>'status' AS status, created_at, finished_at
FROM install_workflows
WHERE owner_type = 'installs' AND owner_id IN ? AND type IN ? AND deleted_at = 0
ORDER BY owner_id, created_at DESC`, installIDs, deployWorkflowTypes).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	out := make(map[string]deployRow, len(rows))
	for _, row := range rows {
		out[row.OwnerID] = row
	}
	return out, nil
}

func installAttrs(install *app.Install) []attribute.KeyValue {
	return []attribute.KeyValue{
		attribute.String("nuon.org.id", install.OrgID),
		attribute.String("nuon.app.id", install.AppID),
		attribute.String("nuon.install.id", install.ID),
	}
}

func healthStatus(status string) string {
	return valueOr(status, "unset")
}

func valueOr(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func unixSeconds(t time.Time) float64 {
	return float64(t.UnixNano()) / 1e9
}
