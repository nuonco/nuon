package eventautomations

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/config/sync"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func Sync(ctx context.Context, db *gorm.DB, cfg *config.AppConfig, orgID, appID, appConfigID string) error {
	if cfg.Events == nil || len(cfg.Events.Rules) == 0 {
		return nil
	}

	validFrom := time.Now().UTC()
	for _, ruleCfg := range cfg.Events.Rules {
		source, err := resolveEventSource(ctx, db, orgID, ruleCfg.Source)
		if err != nil {
			return err
		}
		branch, err := resolveAppBranch(ctx, db, appID, ruleCfg.Target.AppBranch)
		if err != nil {
			return err
		}
		rule, err := buildRule(ruleCfg, orgID, appID, appConfigID, source.ID, branch.ID, validFrom)
		if err != nil {
			return sync.SyncInternalErr{Description: fmt.Sprintf("unable to build event automation rule %q", ruleCfg.Name), Err: err}
		}
		res := db.WithContext(ctx).Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "app_config_id"}, {Name: "name"}, {Name: "deleted_at"}},
			DoNothing: true,
		}).Create(rule)
		if res.Error != nil {
			return sync.SyncInternalErr{Description: fmt.Sprintf("unable to create event automation rule %q", ruleCfg.Name), Err: res.Error}
		}
		if res.RowsAffected == 0 {
			if err := verifyExistingRule(ctx, db, rule); err != nil {
				return err
			}
		}
	}
	return nil
}

func verifyExistingRule(ctx context.Context, db *gorm.DB, desired *app.EventAutomationRule) error {
	var existing app.EventAutomationRule
	if err := db.WithContext(ctx).Where(app.EventAutomationRule{
		AppConfigID: desired.AppConfigID,
		Name:        desired.Name,
	}).First(&existing).Error; err != nil {
		return sync.SyncInternalErr{Description: fmt.Sprintf("unable to verify event automation rule %q", desired.Name), Err: err}
	}
	if existing.OrgID != desired.OrgID || existing.AppID != desired.AppID ||
		existing.EventSourceID != desired.EventSourceID || existing.AppBranchID != desired.AppBranchID ||
		existing.TargetType != desired.TargetType || existing.ConfigHash != desired.ConfigHash {
		return sync.SyncInternalErr{
			Description: fmt.Sprintf("event automation rule %q conflicts with an existing immutable revision", desired.Name),
			Err:         errors.New("event automation rule revision mismatch"),
		}
	}
	return nil
}

func resolveEventSource(ctx context.Context, db *gorm.DB, orgID, name string) (*app.EventSource, error) {
	var source app.EventSource
	err := db.WithContext(ctx).Where(app.EventSource{OrgID: orgID, Name: name, Status: app.EventSourceStatusActive}).First(&source).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, sync.SyncErr{Resource: "event-automations", Description: fmt.Sprintf("event automation references unknown or inactive event source %q", name)}
	}
	if err != nil {
		return nil, sync.SyncInternalErr{Description: fmt.Sprintf("unable to resolve event source %q", name), Err: err}
	}
	return &source, nil
}

func resolveAppBranch(ctx context.Context, db *gorm.DB, appID, name string) (*app.AppBranch, error) {
	var branch app.AppBranch
	err := db.WithContext(ctx).Where(app.AppBranch{AppID: appID, Name: name}).First(&branch).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, sync.SyncErr{Resource: "event-automations", Description: fmt.Sprintf("event automation references unknown app branch %q", name)}
	}
	if err != nil {
		return nil, sync.SyncInternalErr{Description: fmt.Sprintf("unable to resolve app branch %q", name), Err: err}
	}
	return &branch, nil
}

func buildRule(ruleCfg *config.EventRuleConfig, orgID, appID, appConfigID, sourceID, branchID string, validFrom time.Time) (*app.EventAutomationRule, error) {
	hash, err := configHash(ruleCfg)
	if err != nil {
		return nil, err
	}
	filters := make([]app.EventAutomationFilter, len(ruleCfg.Filters))
	for i, filter := range ruleCfg.Filters {
		filters[i] = app.EventAutomationFilter{From: filter.From, Op: app.EventAutomationFilterType(filter.Op), Path: filter.Path, Value: filter.Value}
	}
	return &app.EventAutomationRule{
		OrgID: orgID, AppID: appID, AppConfigID: appConfigID, EventSourceID: sourceID,
		Name: ruleCfg.Name, Enabled: true, ValidFrom: validFrom,
		EventTypes: pq.StringArray(ruleCfg.EventTypes), Filters: filters,
		TargetType: app.EventAutomationTargetTypeAppBranchRun, AppBranchID: branchID,
		Force: true, PlanOnly: false, ConfigHash: hash,
	}, nil
}

func configHash(ruleCfg *config.EventRuleConfig) (string, error) {
	normalized := *ruleCfg
	normalized.EventTypes = append([]string(nil), ruleCfg.EventTypes...)
	sort.Strings(normalized.EventTypes)
	normalized.Filters = append([]config.EventFilterConfig(nil), ruleCfg.Filters...)
	sort.Slice(normalized.Filters, func(i, j int) bool {
		left, _ := json.Marshal(normalized.Filters[i])
		right, _ := json.Marshal(normalized.Filters[j])
		return string(left) < string(right)
	})
	contents, err := json.Marshal(normalized)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(contents)
	return hex.EncodeToString(sum[:]), nil
}
