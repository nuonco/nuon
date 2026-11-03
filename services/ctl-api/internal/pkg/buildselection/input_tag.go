package buildselection

import (
	"context"
	"encoding/json"
	"strings"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
)

type changedInputValue struct {
	Old string `json:"old"`
	New string `json:"new"`
}

func ValidateComponentForWorkflow(ctx context.Context, db *gorm.DB, appConfigID, componentID, buildID, workflowID string) error {
	err := ValidateComponent(ctx, db, appConfigID, componentID, buildID)
	if err == nil {
		return nil
	}
	ok, allowErr := allowChangedInputImageTag(ctx, db, workflowID, buildID)
	if allowErr != nil {
		return allowErr
	}
	if ok {
		return nil
	}
	return err
}

func allowChangedInputImageTag(ctx context.Context, db *gorm.DB, workflowID, buildID string) (bool, error) {
	if workflowID == "" || buildID == "" {
		return false, nil
	}
	var flow app.Workflow
	if err := db.WithContext(ctx).Scopes(scopes.WithDisableViews).Where(app.Workflow{ID: workflowID}).First(&flow).Error; err != nil {
		return false, err
	}
	if flow.Type != app.WorkflowTypeInputUpdate {
		return false, nil
	}
	raw := generics.FromPtrStr(flow.Metadata[app.WorkflowMetadataKeyChangedInputValues])
	if raw == "" {
		return false, nil
	}
	var changed map[string]changedInputValue
	if err := json.Unmarshal([]byte(raw), &changed); err != nil {
		return false, nil
	}
	var build app.ComponentBuild
	if err := db.WithContext(ctx).Scopes(scopes.WithDisableViews).Where(app.ComponentBuild{ID: buildID}).First(&build).Error; err != nil {
		return false, err
	}
	if !changedInputMatchesTag(build.ResolvedTag, build.SourceRef, changed) {
		return false, nil
	}
	var ccc app.ComponentConfigConnection
	if err := db.WithContext(ctx).Scopes(scopes.WithDisableViews).Where(app.ComponentConfigConnection{ID: build.ComponentConfigConnectionID}).First(&ccc).Error; err != nil {
		return false, err
	}
	if err := ValidateComponent(ctx, db, ccc.AppConfigID, ccc.ComponentID, build.ID); err != nil {
		return false, nil
	}
	return true, nil
}

func changedInputMatchesTag(resolvedTag, sourceRef string, changed map[string]changedInputValue) bool {
	for _, value := range changed {
		if value.New == "" || value.New == value.Old || value.New == "***" {
			continue
		}
		if resolvedTag == value.New || strings.HasSuffix(sourceRef, ":"+value.New) {
			return true
		}
	}
	return false
}
