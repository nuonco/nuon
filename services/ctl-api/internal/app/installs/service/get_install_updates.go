package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/blobstore"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type InstallUpdateType string

const (
	InstallUpdateTypeAppConfig     InstallUpdateType = "app_config"
	InstallUpdateTypeInputs        InstallUpdateType = "inputs"
	InstallUpdateTypeStack         InstallUpdateType = "stack"
	InstallUpdateTypeInstallConfig InstallUpdateType = "install_config"
)

type InstallAppConfigUpdate struct {
	Version *app.InstallAppConfigVersion `json:"version"`
	Diff    *app.InstallConfigDiff       `json:"diff,omitempty"`
}

type InstallInputsUpdate struct {
	InputConfigID string   `json:"input_config_id,omitempty"`
	Keys          []string `json:"keys"`
}

type InstallStackUpdate struct {
	VersionID string                        `json:"version_id"`
	Status    app.CompositeStatus           `json:"status"`
	RoleDiff  *app.StackVersionRunRoleDiff  `json:"role_diff,omitempty"`
	InputDiff *app.StackVersionRunInputDiff `json:"input_diff,omitempty"`
	RunType   app.StackVersionRunType       `json:"run_type,omitempty"`
}

type InstallConfigUpdate struct {
	Version *app.InstallConfigVersion `json:"version"`
}

type InstallUpdate struct {
	ID          string            `json:"id"`
	Type        InstallUpdateType `json:"type"`
	CreatedAt   time.Time         `json:"created_at"`
	CreatedByID string            `json:"created_by_id,omitempty"`
	WorkflowID  *string           `json:"workflow_id,omitempty"`

	AppConfig     *InstallAppConfigUpdate `json:"app_config,omitempty"`
	Inputs        *InstallInputsUpdate    `json:"inputs,omitempty"`
	Stack         *InstallStackUpdate     `json:"stack,omitempty"`
	InstallConfig *InstallConfigUpdate    `json:"install_config,omitempty"`
}

type InstallUpdatesResponse struct {
	Updates             []InstallUpdate   `json:"updates"`
	CurrentAppBranchRun *app.AppBranchRun `json:"current_app_branch_run,omitempty"`
	Page                int               `json:"page"`
	Limit               int               `json:"limit"`
	HasMore             bool              `json:"has_more"`
}

// @ID                    GetInstallUpdates
// @Summary               get typed updates for an install
// @Description           Returns app config, input, stack, and install config updates in reverse chronological order.
// @Param                 install_id path string true "install ID"
// @Param                 page query int false "page number" Default(0)
// @Param                 offset query int false "offset of results to return" Default(0)
// @Param                 limit query int false "page size" Default(20)
// @Tags                  installs
// @Produce               json
// @Security              APIKey
// @Security              OrgID
// @Failure               400 {object} stderr.ErrResponse
// @Failure               401 {object} stderr.ErrResponse
// @Failure               403 {object} stderr.ErrResponse
// @Failure               500 {object} stderr.ErrResponse
// @Success               200 {object} InstallUpdatesResponse
// @Router                /v1/installs/{install_id}/updates [GET]
func (s *service) GetInstallUpdates(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	limit := queryInt(ctx, "limit", 20, 1, 100)
	page := queryInt(ctx, "page", 0, 0, 10_000)
	offset := queryInt(ctx, "offset", 0, 0, 1_000_000)
	if ctx.Query("page") != "" {
		offset = page * limit
	} else {
		page = offset / limit
	}
	response, err := s.getInstallUpdates(ctx, org.ID, ctx.Param("install_id"), page, offset, limit)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install updates: %w", err))
		return
	}
	ctx.JSON(http.StatusOK, response)
}

func queryInt(ctx *gin.Context, key string, fallback, minValue, maxValue int) int {
	value, err := strconv.Atoi(ctx.Query(key))
	if err != nil || value < minValue {
		return fallback
	}
	if value > maxValue {
		return maxValue
	}
	return value
}

func (s *service) getInstallUpdates(ctx *gin.Context, orgID, installID string, page, offset, limit int) (*InstallUpdatesResponse, error) {
	fetchLimit := offset + limit + 1
	updates := make([]InstallUpdate, 0, fetchLimit*4)

	var appVersions []app.InstallAppConfigVersion
	if err := s.db.WithContext(ctx).
		Preload("Workflow").
		Preload("AppBranchRun").
		Preload("AppBranchRun.Preview").
		Preload("AppBranchRun.AppBranch").
		Preload("AppBranchRun.VCSConnectionCommit").
		Where(app.InstallAppConfigVersion{OrgID: orgID, InstallID: installID}).
		Order("created_at DESC").
		Limit(fetchLimit).
		Find(&appVersions).Error; err != nil {
		return nil, err
	}
	blobCtx := blobstore.WithBlobService(ctx.Request.Context(), s.blobSvc)
	for i := range appVersions {
		version := &appVersions[i]
		diff, err := loadInstallConfigDiff(blobCtx, version)
		if err != nil {
			return nil, err
		}
		updates = append(updates, InstallUpdate{
			ID:          version.ID,
			Type:        InstallUpdateTypeAppConfig,
			CreatedAt:   version.CreatedAt,
			CreatedByID: version.CreatedByID,
			WorkflowID:  version.WorkflowID,
			AppConfig:   &InstallAppConfigUpdate{Version: version, Diff: diff},
		})
	}

	var inputVersions []app.InstallInputs
	if err := s.db.WithContext(ctx).
		Where(app.InstallInputs{OrgID: orgID, InstallID: installID}).
		Order("created_at DESC").
		Limit(fetchLimit).
		Find(&inputVersions).Error; err != nil {
		return nil, err
	}
	for i := range inputVersions {
		inputs := &inputVersions[i]
		keys := make([]string, 0, len(inputs.ValuesRedacted))
		for key := range inputs.ValuesRedacted {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		updates = append(updates, InstallUpdate{
			ID:          inputs.ID,
			Type:        InstallUpdateTypeInputs,
			CreatedAt:   inputs.CreatedAt,
			CreatedByID: inputs.CreatedByID,
			Inputs:      &InstallInputsUpdate{InputConfigID: inputs.AppInputConfigID, Keys: keys},
		})
	}

	var stackVersions []app.InstallStackVersion
	if err := s.db.WithContext(ctx).
		Where(app.InstallStackVersion{OrgID: orgID, InstallID: installID}).
		Order("created_at DESC").
		Limit(fetchLimit).
		Find(&stackVersions).Error; err != nil {
		return nil, err
	}
	if err := s.attachStackVersionRuns(ctx, stackVersions, 1); err != nil {
		return nil, err
	}
	for i := range stackVersions {
		version := &stackVersions[i]
		payload := &InstallStackUpdate{VersionID: version.ID, Status: version.Status}
		if len(version.Runs) > 0 {
			payload.RoleDiff = version.Runs[0].RoleDiff
			payload.InputDiff = version.Runs[0].InputDiff
			payload.RunType = version.Runs[0].RunType
		}
		updates = append(updates, InstallUpdate{
			ID:          version.ID,
			Type:        InstallUpdateTypeStack,
			CreatedAt:   version.CreatedAt,
			CreatedByID: version.CreatedByID,
			Stack:       payload,
		})
	}

	var installVersions []app.InstallConfigVersion
	if err := s.db.WithContext(ctx).
		Preload("InstallConfigSync").
		Preload("InstallConfigSync.VCSConnectionCommit").
		Where(app.InstallConfigVersion{OrgID: orgID, InstallID: installID}).
		Order("created_at DESC").
		Limit(fetchLimit).
		Find(&installVersions).Error; err != nil {
		return nil, err
	}
	for i := range installVersions {
		version := &installVersions[i]
		updates = append(updates, InstallUpdate{
			ID:          version.ID,
			Type:        InstallUpdateTypeInstallConfig,
			CreatedAt:   version.CreatedAt,
			CreatedByID: version.CreatedByID,
			InstallConfig: &InstallConfigUpdate{
				Version: version,
			},
		})
	}

	sort.SliceStable(updates, func(i, j int) bool {
		return updates[i].CreatedAt.After(updates[j].CreatedAt)
	})
	start := offset
	if start > len(updates) {
		start = len(updates)
	}
	end := start + limit
	if end > len(updates) {
		end = len(updates)
	}

	currentRun, err := s.currentAppliedAppBranchRun(ctx, orgID, installID)
	if err != nil {
		return nil, err
	}
	return &InstallUpdatesResponse{
		Updates:             updates[start:end],
		CurrentAppBranchRun: currentRun,
		Page:                page,
		Limit:               limit,
		HasMore:             end < len(updates),
	}, nil
}

func loadInstallConfigDiff(ctx context.Context, version *app.InstallAppConfigVersion) (*app.InstallConfigDiff, error) {
	if version == nil || version.Diff == nil || !version.Diff.IsSet() {
		return nil, nil
	}
	raw, err := version.Diff.Get(ctx)
	if err != nil {
		return nil, err
	}
	if raw == "" {
		return nil, nil
	}
	var diff app.InstallConfigDiff
	if err := json.Unmarshal([]byte(raw), &diff); err != nil {
		return nil, err
	}
	return &diff, nil
}

func (s *service) currentAppliedAppBranchRun(ctx *gin.Context, orgID, installID string) (*app.AppBranchRun, error) {
	var version app.InstallAppConfigVersion
	err := s.db.WithContext(ctx).
		Preload("AppBranchRun").
		Preload("AppBranchRun.Preview").
		Preload("AppBranchRun.AppBranch").
		Preload("AppBranchRun.VCSConnectionCommit").
		Where(app.InstallAppConfigVersion{OrgID: orgID, InstallID: installID}).
		Where("status ->> 'status' = ?", app.StatusSuccess).
		Where("app_branch_run_id IS NOT NULL").
		Order("created_at DESC").
		First(&version).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &version.AppBranchRun, nil
}
