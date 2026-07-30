package triggers

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/config/sync"
	configvalidate "github.com/nuonco/nuon/pkg/config/validate"
	"github.com/nuonco/nuon/pkg/eventfilter"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func Sync(ctx context.Context, db *gorm.DB, cfg *config.AppConfig, orgID, appID, appConfigID string) error {
	if err := configvalidate.ValidateTriggers(cfg); err != nil {
		return err
	}
	if cfg.Triggers == nil || len(cfg.Triggers.Rules) == 0 {
		return nil
	}
	var org app.Org
	if err := db.WithContext(ctx).Select("id", "features").Where(app.Org{ID: orgID}).First(&org).Error; err != nil {
		return sync.SyncInternalErr{Description: "unable to check triggers feature", Err: err}
	}
	if !org.Features[string(app.OrgFeatureTriggers)] {
		return sync.SyncErr{Resource: "triggers", Description: "the triggers feature is not enabled for this organization"}
	}

	validFrom := time.Now().UTC()
	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		triggerNames := referencedTriggerNames(cfg.Triggers.Rules)
		lockedTriggers := make(map[string]*app.Trigger, len(triggerNames))
		for _, triggerName := range triggerNames {
			trigger, err := resolveTrigger(ctx, tx.Clauses(clause.Locking{Strength: "UPDATE"}), orgID, triggerName)
			if err != nil {
				return err
			}
			lockedTriggers[triggerName] = trigger
		}

		for _, ruleCfg := range cfg.Triggers.Rules {
			trigger := lockedTriggers[ruleCfg.Trigger]
			var branchID, runbookID *string
			if ruleCfg.Target.Type == "app_branch_run" {
				branch, err := resolveAppBranch(ctx, tx, appID, ruleCfg.Target.AppBranch)
				if err != nil {
					return err
				}
				branchID = &branch.ID
			} else {
				runbook, err := resolveRunbook(ctx, tx, appID, ruleCfg.Target.Runbook)
				if err != nil {
					return err
				}
				runbookID = &runbook.ID
			}
			rule, err := buildRule(ruleCfg, orgID, appID, appConfigID, trigger.ID, branchID, runbookID, validFrom)
			if err != nil {
				return sync.SyncInternalErr{Description: fmt.Sprintf("unable to build trigger rule %q", ruleCfg.Name), Err: err}
			}
			res := tx.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "app_config_id"}, {Name: "name"}, {Name: "deleted_at"}},
				DoNothing: true,
			}).Create(rule)
			if res.Error != nil {
				return sync.SyncInternalErr{Description: fmt.Sprintf("unable to create trigger rule %q", ruleCfg.Name), Err: res.Error}
			}
			if res.RowsAffected == 0 {
				if err := verifyExistingRule(ctx, tx, rule); err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func referencedTriggerNames(rules []*config.TriggerRuleConfig) []string {
	triggerNames := make([]string, 0, len(rules))
	seenTriggerNames := make(map[string]struct{}, len(rules))
	for _, rule := range rules {
		if _, ok := seenTriggerNames[rule.Trigger]; ok {
			continue
		}
		seenTriggerNames[rule.Trigger] = struct{}{}
		triggerNames = append(triggerNames, rule.Trigger)
	}
	sort.Strings(triggerNames)
	return triggerNames
}

func verifyExistingRule(ctx context.Context, db *gorm.DB, desired *app.TriggerRule) error {
	var existing app.TriggerRule
	if err := db.WithContext(ctx).Where(app.TriggerRule{
		AppConfigID: desired.AppConfigID,
		Name:        desired.Name,
	}).First(&existing).Error; err != nil {
		return sync.SyncInternalErr{Description: fmt.Sprintf("unable to verify trigger rule %q", desired.Name), Err: err}
	}
	if existing.OrgID != desired.OrgID || existing.AppID != desired.AppID ||
		existing.TriggerID != desired.TriggerID || !reflect.DeepEqual(existing.AppBranchID, desired.AppBranchID) ||
		!reflect.DeepEqual(existing.RunbookID, desired.RunbookID) || existing.InstallName != desired.InstallName ||
		existing.TargetType != desired.TargetType || existing.ConfigHash != desired.ConfigHash {
		return sync.SyncInternalErr{
			Description: fmt.Sprintf("trigger rule %q conflicts with an existing immutable revision", desired.Name),
			Err:         errors.New("trigger rule revision mismatch"),
		}
	}
	return nil
}

func resolveRunbook(ctx context.Context, db *gorm.DB, appID, name string) (*app.Runbook, error) {
	var runbook app.Runbook
	err := db.WithContext(ctx).Where(app.Runbook{AppID: appID, Name: name}).First(&runbook).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, sync.SyncErr{Resource: "triggers", Description: fmt.Sprintf("trigger references unknown runbook %q", name)}
	}
	if err != nil {
		return nil, sync.SyncInternalErr{Description: fmt.Sprintf("unable to resolve runbook %q", name), Err: err}
	}
	return &runbook, nil
}

func resolveTrigger(ctx context.Context, db *gorm.DB, orgID, name string) (*app.Trigger, error) {
	var trigger app.Trigger
	err := db.WithContext(ctx).Where(app.Trigger{OrgID: orgID, Name: name, Status: app.TriggerStatusActive}).First(&trigger).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, sync.SyncErr{Resource: "triggers", Description: fmt.Sprintf("trigger references unknown or inactive trigger %q", name)}
	}
	if err != nil {
		return nil, sync.SyncInternalErr{Description: fmt.Sprintf("unable to resolve trigger %q", name), Err: err}
	}
	return &trigger, nil
}

func resolveAppBranch(ctx context.Context, db *gorm.DB, appID, name string) (*app.AppBranch, error) {
	var branch app.AppBranch
	err := db.WithContext(ctx).Where(app.AppBranch{AppID: appID, Name: name}).First(&branch).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, sync.SyncErr{Resource: "triggers", Description: fmt.Sprintf("trigger references unknown app branch %q", name)}
	}
	if err != nil {
		return nil, sync.SyncInternalErr{Description: fmt.Sprintf("unable to resolve app branch %q", name), Err: err}
	}
	return &branch, nil
}

func buildRule(ruleCfg *config.TriggerRuleConfig, orgID, appID, appConfigID, triggerID string, branchID, runbookID *string, validFrom time.Time) (*app.TriggerRule, error) {
	hash, err := configHash(ruleCfg)
	if err != nil {
		return nil, err
	}
	filters := make([]app.TriggerFilter, len(ruleCfg.Filters))
	for i, filter := range ruleCfg.Filters {
		if _, err := eventfilter.Compile(eventfilter.Filter{From: eventfilter.Source(filter.From), Path: filter.Path, Op: eventfilter.Operator(filter.Op), Value: filter.Value}); err != nil {
			return nil, fmt.Errorf("compile filter %d: %w", i, err)
		}
		filters[i] = app.TriggerFilter{From: filter.From, Op: app.TriggerFilterType(filter.Op), Path: filter.Path, Value: filter.Value}
	}
	return &app.TriggerRule{
		OrgID: orgID, AppID: appID, AppConfigID: appConfigID, TriggerID: triggerID,
		Name: ruleCfg.Name, Enabled: true, ValidFrom: validFrom,
		EventTypes: pq.StringArray(ruleCfg.EventTypes), Filters: filters,
		TargetType: app.TriggerTargetType(ruleCfg.Target.Type), AppBranchID: branchID,
		RunbookID: runbookID, InstallName: ruleCfg.Target.Install, InputMappings: ruleCfg.Target.Inputs,
		Force: true, PlanOnly: false, ConfigHash: hash,
	}, nil
}

func configHash(ruleCfg *config.TriggerRuleConfig) (string, error) {
	normalized := *ruleCfg
	normalized.EventTypes = append([]string(nil), ruleCfg.EventTypes...)
	sort.Strings(normalized.EventTypes)
	normalized.Filters = append([]config.TriggerFilterConfig(nil), ruleCfg.Filters...)
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
