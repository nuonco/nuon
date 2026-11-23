package adminops

import (
	"context"
	"fmt"
	"sort"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/config/validate"
	"github.com/nuonco/nuon/pkg/urlpolicy"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

const (
	AllowedHostsStatusUnrestricted = "unrestricted"
	AllowedHostsStatusAllowed      = "allowed"
	AllowedHostsStatusViolation    = "violation"
)

type AllowedHostsScanOptions struct {
	AllConfigs bool
}

type AllowedHostsScanRow struct {
	OrgID    string   `json:"org_id"`
	Category string   `json:"category"`
	Value    string   `json:"value"`
	Status   string   `json:"status"`
	Reason   string   `json:"reason,omitempty"`
	Fields   []string `json:"fields"`
	AppIDs   []string `json:"app_ids"`
}

type AllowedHostsScanResult struct {
	Mode       string                `json:"mode"`
	AllConfigs bool                  `json:"all_configs"`
	Total      int                   `json:"total"`
	Violations int                   `json:"violations"`
	Rows       []AllowedHostsScanRow `json:"rows"`
}

type allowedHostsScanKey struct {
	orgID    string
	category urlpolicy.Category
	value    string
}

type allowedHostsScan struct {
	policy *urlpolicy.Policy
	rows   map[allowedHostsScanKey]*allowedHostsScanRow
}

type allowedHostsScanRow struct {
	status string
	reason string
	fields map[string]struct{}
	appIDs map[string]struct{}
}

func (s *allowedHostsScan) add(orgID, appID string, category urlpolicy.Category, field, value string, violation *urlpolicy.Violation) {
	key := allowedHostsScanKey{orgID: orgID, category: category, value: value}
	row, ok := s.rows[key]
	if !ok {
		row = &allowedHostsScanRow{
			status: AllowedHostsStatusUnrestricted,
			fields: map[string]struct{}{},
			appIDs: map[string]struct{}{},
		}
		if s.policy.Restricted(category) {
			row.status = AllowedHostsStatusAllowed
		}
		s.rows[key] = row
	}
	if violation != nil {
		row.status = AllowedHostsStatusViolation
		row.reason = violation.Reason
	}
	row.fields[field] = struct{}{}
	row.appIDs[appID] = struct{}{}
}

func (s *allowedHostsScan) addRefs(orgID, appID string, refs []validate.URLRef) {
	for _, ref := range refs {
		s.add(orgID, appID, ref.Category, ref.Field, ref.Value, s.policy.Check(ref.Category, ref.Field, ref.Value))
	}
}

func (o *Ops) ScanAllowedHosts(ctx context.Context, policy *urlpolicy.Policy, opts AllowedHostsScanOptions) (*AllowedHostsScanResult, error) {
	scan := &allowedHostsScan{policy: policy, rows: map[allowedHostsScanKey]*allowedHostsScanRow{}}

	if err := o.scanRunnerConfigs(ctx, scan, opts); err != nil {
		return nil, err
	}
	stackTypes, err := o.scanStackConfigs(ctx, scan, opts)
	if err != nil {
		return nil, err
	}
	if err := o.scanInstallConfigs(ctx, scan, stackTypes, opts); err != nil {
		return nil, err
	}

	res := &AllowedHostsScanResult{
		Mode:       string(policy.Mode()),
		AllConfigs: opts.AllConfigs,
		Rows:       make([]AllowedHostsScanRow, 0, len(scan.rows)),
	}
	for key, row := range scan.rows {
		if row.status == AllowedHostsStatusViolation {
			res.Violations++
		}
		res.Rows = append(res.Rows, AllowedHostsScanRow{
			OrgID:    key.orgID,
			Category: string(key.category),
			Value:    key.value,
			Status:   row.status,
			Reason:   row.reason,
			Fields:   sortedKeys(row.fields),
			AppIDs:   sortedKeys(row.appIDs),
		})
	}
	res.Total = len(res.Rows)
	sort.Slice(res.Rows, func(i, j int) bool {
		a, b := res.Rows[i], res.Rows[j]
		if a.OrgID != b.OrgID {
			return a.OrgID < b.OrgID
		}
		if a.Category != b.Category {
			return a.Category < b.Category
		}
		return a.Value < b.Value
	})

	return res, nil
}

func (o *Ops) scanRunnerConfigs(ctx context.Context, scan *allowedHostsScan, opts AllowedHostsScanOptions) error {
	columns := "id, org_id, app_id, env_vars, init_script_url, phone_home_script_url, runner_api_url, public_api_url"
	return eachConfig(o.latestPer(ctx, &app.AppRunnerConfig{}, "app_id", columns, opts), o.db, func(cfg *app.AppRunnerConfig) {
		scan.addRefs(cfg.OrgID, cfg.AppID, validate.RunnerURLRefs(validate.RunnerURLs{
			RunnerAPIURL:       cfg.RunnerAPIURL,
			PublicAPIURL:       cfg.PublicAPIURL,
			InitScriptURL:      cfg.InitScriptURL,
			PhoneHomeScriptURL: cfg.PhoneHomeScriptURL,
		}))
		for name := range cfg.EnvVars {
			if !urlpolicy.ReservedRunnerEnvVar(name) {
				continue
			}
			field := "runner.env_vars." + name
			scan.add(cfg.OrgID, cfg.AppID, urlpolicy.CategoryRunnerAPI, field, name, scan.policy.CheckRunnerEnvVar(field, name))
		}
	})
}

func (o *Ops) scanStackConfigs(ctx context.Context, scan *allowedHostsScan, opts AllowedHostsScanOptions) (map[string]string, error) {
	stackTypes := map[string]string{}
	columns := "id, org_id, app_id, type, runner_nested_template_url, vpc_nested_template_url, uidefinition, custom_nested_stacks"
	query := o.latestPer(ctx, &app.AppStackConfig{}, "app_id", columns, opts)
	if opts.AllConfigs {
		query = query.Order("created_at ASC")
	}
	err := eachConfig(query, o.db, func(cfg *app.AppStackConfig) {
		stackTypes[cfg.AppID] = string(cfg.Type)
		urls := validate.StackURLs{
			Type:                    string(cfg.Type),
			VPCNestedTemplateURL:    cfg.VPCNestedTemplateURL,
			RunnerNestedTemplateURL: cfg.RunnerNestedTemplateURL,
			CustomNestedStacks:      cfg.CustomNestedStacks,
		}
		if cfg.UIDefinition != nil {
			urls.UIDefinition = cfg.UIDefinition.Path
		}
		scan.addRefs(cfg.OrgID, cfg.AppID, validate.StackURLRefs("stack", urls))
	})
	return stackTypes, err
}

func (o *Ops) scanInstallConfigs(ctx context.Context, scan *allowedHostsScan, stackTypes map[string]string, opts AllowedHostsScanOptions) error {
	appIDs := map[string]string{}
	installs := o.db.WithContext(ctx).Model(&app.Install{}).Select("id, app_id")
	if err := eachConfig(installs, o.db, func(install *app.Install) {
		appIDs[install.ID] = install.AppID
	}); err != nil {
		return err
	}

	columns := "id, org_id, install_id, vpc_nested_template_url, runner_nested_template_url, uidefinition, custom_nested_stacks"
	return eachConfig(o.latestPer(ctx, &app.InstallConfig{}, "install_id", columns, opts), o.db, func(cfg *app.InstallConfig) {
		appID, ok := appIDs[cfg.InstallID]
		if !ok {
			return
		}
		urls := validate.StackURLs{
			Type:               stackTypes[appID],
			CustomNestedStacks: cfg.CustomNestedStacks,
		}
		if cfg.VPCNestedTemplateURL != nil {
			urls.VPCNestedTemplateURL = *cfg.VPCNestedTemplateURL
		}
		if cfg.RunnerNestedTemplateURL != nil {
			urls.RunnerNestedTemplateURL = *cfg.RunnerNestedTemplateURL
		}
		if cfg.UIDefinition != nil {
			urls.UIDefinition = cfg.UIDefinition.Path
		}
		scan.addRefs(cfg.OrgID, appID, validate.StackURLRefs("installs[].stack_overrides", urls))
	})
}

func (o *Ops) latestPer(ctx context.Context, model any, groupColumn, columns string, opts AllowedHostsScanOptions) *gorm.DB {
	query := o.db.WithContext(ctx).Model(model)
	if opts.AllConfigs {
		return query.Select(columns)
	}
	return query.
		Select(fmt.Sprintf("DISTINCT ON (%s) %s", groupColumn, columns)).
		Order(groupColumn).
		Order("created_at DESC")
}

func eachConfig[T any](query, db *gorm.DB, fn func(*T)) error {
	rows, err := query.Rows()
	if err != nil {
		return fmt.Errorf("unable to scan configs: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var row T
		if err := db.ScanRows(rows, &row); err != nil {
			return fmt.Errorf("unable to read config: %w", err)
		}
		fn(&row)
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("unable to scan configs: %w", err)
	}
	return nil
}

func sortedKeys(set map[string]struct{}) []string {
	out := make([]string, 0, len(set))
	for key := range set {
		out = append(out, key)
	}
	sort.Strings(out)
	return out
}
