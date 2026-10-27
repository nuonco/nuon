package alignment

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	installhelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/buildselection"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/driftresolution"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/request"
)

type ResolveRequest struct {
	ResourceID     string `json:"resource_id" validate:"required"`
	RequestID      string `json:"request_id" validate:"required,max=255"`
	ApprovedPlanID string `json:"approved_plan_id,omitempty"`
}

type ResolveResult struct {
	WorkflowID       string
	Created          bool
	Resource         InstallAlignmentResource
	Deploy           *app.InstallDeploy
	Approval         *app.WorkflowStepApproval
	ApprovalResponse *app.WorkflowStepApprovalResponse
}

type openDeploymentExistsError struct {
	workflowID string
}

func (e *openDeploymentExistsError) Error() string {
	return fmt.Sprintf("resource already has an open deployment %s", e.workflowID)
}

type resolveDriftTarget struct {
	resource InstallAlignmentResource
	ic       *app.InstallComponent
}

func (s *Helpers) loadResolveDriftTarget(ctx context.Context, db *gorm.DB, install *app.Install, resourceID string) (*installAlignmentState, *resolveDriftTarget, error) {
	st, err := s.loadInstallAlignmentState(ctx, db, install)
	if err != nil {
		return nil, nil, err
	}
	target, err := st.resolveDriftTarget(resourceID)
	if err != nil {
		return nil, nil, err
	}
	evidence, err := s.loadAlignmentEvidence(ctx, db, st)
	if err != nil {
		return nil, nil, err
	}
	target.resource.applyDriftEvidence(evidence)
	return st, target, nil
}

func (s *Helpers) Resolve(ctx context.Context, orgID, installID string, req *ResolveRequest) (*ResolveResult, error) {
	var install app.Install
	if err := s.db.WithContext(ctx).Where(app.Install{ID: installID, OrgID: orgID}).First(&install).Error; err != nil {
		if isNotFound(err) {
			return nil, stderr.ErrNotFound{Err: err, Description: "install not found"}
		}
		return nil, fmt.Errorf("unable to get install: %w", err)
	}
	hash, err := request.Hash(struct {
		ResourceID     string `json:"resource_id"`
		ApprovedPlanID string `json:"approved_plan_id"`
	}{ResourceID: req.ResourceID, ApprovedPlanID: req.ApprovedPlanID})
	if err != nil {
		return nil, err
	}
	var replay app.Workflow
	err = s.db.WithContext(ctx).
		Where(app.Workflow{OrgID: install.OrgID, OwnerID: install.ID, OwnerType: "installs"}).
		Where("request->>'request_id' = ?", req.RequestID).
		Where("metadata -> ? = ?", app.WorkflowMetadataKeyResolveDrift, "true").First(&replay).Error
	if err == nil {
		if err := request.Check(replay.Request, hash, replay.Request.PinnedAppConfigID); err != nil {
			return nil, err
		}
		return &ResolveResult{WorkflowID: replay.ID}, nil
	}
	if !isNotFound(err) {
		return nil, fmt.Errorf("unable to look up resolve drift request: %w", err)
	}
	st, target, err := s.loadResolveDriftTarget(ctx, s.db, &install, req.ResourceID)
	if err != nil {
		return nil, err
	}
	res := &target.resource
	open, err := s.openDeployments(ctx, s.db, st)
	if err != nil {
		return nil, err
	}
	if existing := open[res.ID]; existing != nil {
		return &ResolveResult{WorkflowID: existing.ID}, nil
	}
	if err := validateResolvable(res); err != nil {
		return nil, err
	}
	var reuse *alignmentApproval
	if req.ApprovedPlanID != "" {
		reuse, err = s.validateResolvePlan(ctx, s.db, st, res, req.ApprovedPlanID, true)
	} else {
		reuse, err = s.findApprovedResolvePlan(ctx, st, res)
	}
	if err != nil {
		return nil, err
	}
	metadata := map[string]string{app.WorkflowMetadataKeyResolveDrift: "true"}
	if reuse != nil {
		metadata["reuse_approval_id"] = reuse.approval.ID
	}
	resolution := &app.DriftResolutionTarget{
		ResourceID: res.ID, Kind: string(res.Kind), AppConfigID: install.AppConfigID,
		BuildID: res.ExpectedBuildID,
	}
	var workflowType app.WorkflowType
	switch res.Kind {
	case InstallAlignmentResourceKindComponent, InstallAlignmentResourceKindImage:
		workflowType = app.WorkflowTypeManualDeploy
		metadata[app.WorkflowMetadataKeyWorkflowNameSuffix] = res.Name
		metadata["deploy_dependents"] = strconv.FormatBool(false)
		metadata["deploy_dependencies"] = strconv.FormatBool(false)
		selected := st.targets[target.ic.ComponentID]
		resolution.ComponentID = target.ic.ComponentID
		resolution.BuildAppConfigID = selected.Build.ComponentConfigConnection.AppConfigID
		if res.Kind == InstallAlignmentResourceKindComponent {
			resolution.InputNames = installhelpers.ComponentInputNames(selected.Config, st.desiredInputs(resolution.BuildAppConfigID))
			if len(resolution.InputNames) > 0 && res.Category == InstallAlignmentCategoryChangesNotApplied {
				workflowType = app.WorkflowTypeInputUpdate
				metadata["inputs"] = strings.Join(resolution.InputNames, ",")
				metadata["resolve_drift_component_id"] = target.ic.ComponentID
			}
		}
	case InstallAlignmentResourceKindSandbox:
		workflowType = app.WorkflowTypeReprovisionSandbox
		metadata["skip_components"] = "true"
		resolution.InputNames = installhelpers.InputNames(st.expectedSandbox.Refs)
	case InstallAlignmentResourceKindStack:
		workflowType = app.WorkflowTypeReprovisionStack
		for _, input := range st.appConfigs[install.AppConfigID].InputConfig.AppInputs {
			resolution.InputNames = append(resolution.InputNames, input.Name)
		}
	}
	resolution.InputsDigest = installhelpers.InputsDigest(resolution.InputNames, st.desiredInputs(resolution.BuildAppConfigID))
	var createdDeploy *app.InstallDeploy
	var createdResponse *app.WorkflowStepApprovalResponse
	wf, created, err := s.installs.RunIdempotentInstallWorkflow(ctx, installhelpers.IdempotentInstallWorkflowRequest{
		InstallID: install.ID, WorkflowType: workflowType, Metadata: metadata,
		RequestID: req.RequestID, RequestHash: hash, Operation: "resolve-drift",
		QueueName: installhelpers.InstallWorkflowsQueueName, SkipAppConfigConflict: true,
	}, &installhelpers.IdempotentInstallWorkflowHooks{
		Prepare: func(tx *gorm.DB) (map[string]string, error) {
			createdDeploy = nil
			createdResponse = nil
			locked, err := driftresolution.LockInstall(ctx, tx, install.ID)
			if err != nil {
				return nil, err
			}
			if locked.AppConfigID != install.AppConfigID {
				return nil, stderr.ErrConflict{Err: fmt.Errorf("resolution config changed"), Description: "the install's selected app config changed; reload Resources and try again"}
			}
			digest, _, err := driftresolution.InputsDigest(ctx, tx, install.ID, install.AppConfigID, resolution)
			if err != nil {
				return nil, err
			}
			if digest != resolution.InputsDigest {
				return nil, stderr.ErrConflict{Err: fmt.Errorf("resolution inputs changed"), Description: "input values changed; reload Resources and try again"}
			}
			fresh, freshTarget, err := s.loadResolveDriftTarget(ctx, tx, &install, req.ResourceID)
			if err != nil {
				return nil, err
			}
			if err := validateResolvable(&freshTarget.resource); err != nil {
				return nil, err
			}
			if freshTarget.resource.ExpectedBuildID != res.ExpectedBuildID {
				return nil, stderr.ErrConflict{Err: fmt.Errorf("resolution build changed"), Description: "the selected build changed; reload Resources and try again"}
			}
			open, err := s.openDeployments(ctx, tx, fresh)
			if err != nil {
				return nil, err
			}
			if existing := open[res.ID]; existing != nil {
				return nil, &openDeploymentExistsError{workflowID: existing.ID}
			}
			ref, err := driftresolution.AppliedRef(ctx, tx, resolution)
			if err != nil {
				return nil, err
			}
			resolution.AppliedByID = ref.AppliedConfigByID
			resolution.AppliedAt = ref.AppliedConfigAt
			if reuse != nil {
				if _, err := s.validateResolvePlan(ctx, tx, fresh, &freshTarget.resource, reuse.approval.ID, req.ApprovedPlanID != ""); err != nil {
					return nil, err
				}
				if req.ApprovedPlanID != "" {
					createdResponse, err = s.ensureResolvePlanApproved(ctx, tx, reuse.approval.ID)
					if err != nil {
						return nil, err
					}
				}
			}
			if target.ic == nil || workflowType == app.WorkflowTypeInputUpdate {
				return nil, nil
			}
			if err := buildselection.ValidateComponent(ctx, tx, install.AppConfigID, target.ic.ComponentID, res.ExpectedBuildID); err != nil {
				return nil, err
			}
			deployType := app.InstallDeployTypeApply
			if target.ic.Component.Type.IsImage() {
				deployType = app.InstallDeployTypeSync
			}
			createdDeploy = &app.InstallDeploy{
				OrgID: install.OrgID, InstallComponentID: target.ic.ID, ComponentBuildID: res.ExpectedBuildID,
				Type: deployType, Status: app.InstallDeployStatusQueued,
				StatusDescription: "waiting to be deployed to install",
			}
			if err := tx.WithContext(ctx).Create(createdDeploy).Error; err != nil {
				return nil, fmt.Errorf("unable to create drift resolution deploy: %w", err)
			}
			return map[string]string{"install_deploy_id": createdDeploy.ID}, nil
		},
		AfterCreate: func(tx *gorm.DB, wf *app.Workflow) error {
			wf.Request.DriftResolution = resolution
			if err := tx.WithContext(ctx).Model(wf).Select("request").Updates(wf).Error; err != nil {
				return err
			}
			if createdDeploy == nil {
				return nil
			}
			return tx.WithContext(ctx).Model(&app.InstallDeploy{}).Where(app.InstallDeploy{ID: createdDeploy.ID}).Update("install_workflow_id", wf.ID).Error
		},
	})
	if err != nil {
		var openErr *openDeploymentExistsError
		if errors.As(err, &openErr) {
			return &ResolveResult{WorkflowID: openErr.workflowID}, nil
		}
		return nil, err
	}
	result := &ResolveResult{
		WorkflowID: wf.ID, Created: created, Resource: *res,
		Deploy: createdDeploy, ApprovalResponse: createdResponse,
	}
	if createdResponse != nil {
		result.Approval = reuse.approval
	}
	return result, nil
}

func (st *installAlignmentState) resolveDriftTarget(resourceID string) (*resolveDriftTarget, error) {
	if st.stack != nil && st.stack.ID == resourceID {
		return &resolveDriftTarget{resource: *st.stackResource()}, nil
	}
	if st.sandbox != nil && st.sandbox.ID == resourceID {
		return &resolveDriftTarget{resource: *st.sandboxResource()}, nil
	}
	for i := range st.components {
		if st.components[i].ID == resourceID {
			return &resolveDriftTarget{resource: st.componentResource(&st.components[i]), ic: &st.components[i]}, nil
		}
	}
	return nil, stderr.ErrNotFound{Err: fmt.Errorf("resource %s not found on install %s", resourceID, st.install.ID), Description: "resource not found on this install"}
}

func validateResolvable(res *InstallAlignmentResource) error {
	switch res.Category {
	case InstallAlignmentCategoryChangesNotApplied, InstallAlignmentCategoryStateDriftDetected:
	case InstallAlignmentCategoryTornDown:
		return stderr.ErrUser{Err: fmt.Errorf("resource %s is torn down", res.ID), Description: "this component was explicitly torn down and is not considered drifted"}
	case InstallAlignmentCategoryDisabled:
		return stderr.ErrUser{Err: fmt.Errorf("resource %s is disabled", res.ID), Description: "this resource is disabled on the install"}
	case InstallAlignmentCategoryInSync:
		return stderr.ErrUser{Err: fmt.Errorf("resource %s is in sync", res.ID), Description: "this resource is already in sync with the install's selected config"}
	default:
		description := "this resource's alignment is unknown"
		if res.Reason != "" {
			description = res.Reason
		}
		return stderr.ErrUser{Err: fmt.Errorf("resource %s alignment is %s", res.ID, res.Category), Description: description}
	}
	if (res.Kind == InstallAlignmentResourceKindComponent || res.Kind == InstallAlignmentResourceKindImage) && res.ExpectedBuildID == "" {
		return stderr.ErrUser{Err: fmt.Errorf("resource %s has no deployable build", res.ID), Description: "no active build exists for the selected component config"}
	}
	if res.ExpectedConfigID == "" {
		return stderr.ErrUser{Err: fmt.Errorf("resource %s has no selected target", res.ID), Description: "the install's selected app config has no target for this resource"}
	}
	return nil
}
