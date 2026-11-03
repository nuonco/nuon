package authz

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	strictjson "sigs.k8s.io/json"

	"github.com/nuonco/nuon/pkg/shortid/domains"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/authz/permissions"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/request"
)

const WorkflowAuthorizationDenied = "workflow-authorization-denied"

type acceptedWorkflow struct {
	Version       int                           `json:"version"`
	AuthorizerID  string                        `json:"authorizer_id"`
	Profile       permissions.CredentialProfile `json:"profile"`
	SourceTokenID string                        `json:"source_token_id"`
	Permission    permissions.PermissionID      `json:"permission"`
	Target        permissions.ResolvedTarget    `json:"target"`
	RequestHash   string                        `json:"request_hash"`
	AcceptedAt    time.Time                     `json:"accepted_at"`
	ExpiresAt     time.Time                     `json:"expires_at"`
}

func (c *Client) AcceptWorkflow(ctx context.Context, req MutationRequest, wf *app.Workflow, expiresAt time.Time) (Decision, error) {
	if wf == nil || wf.ID != "" || len(wf.Authorization) != 0 || wf.AuthorizationFailure != "" ||
		!expiresAt.After(time.Now()) || req.SubjectAccountID != "" {
		return Decision{Reason: ReasonInvalidRequest}, nil
	}
	return c.MutateCustomer(ctx, req, func(tx *gorm.DB, target permissions.ResolvedTarget) error {
		if !workflowTargetMatches(wf, target) {
			return fmt.Errorf("workflow owner must match the authorized target")
		}
		credential, _ := cctx.VerifiedCredentialFromContext(ctx)
		actor, err := cctx.AccountIDFromContext(ctx)
		if err != nil || actor != credential.AccountID {
			return fmt.Errorf("workflow actor must match the verified authorizer")
		}
		orgID, err := cctx.OrgIDFromContext(ctx)
		if err != nil || orgID != target.OrgID {
			return fmt.Errorf("workflow tenant must match the authorized target")
		}
		wf.ID = domains.NewWorkflowID()
		wf.OrgID = target.OrgID
		wf.CreatedByID = actor
		if err := tx.Create(wf).Error; err != nil {
			return err
		}
		hash, err := workflowRequestHash(tx, wf)
		if err != nil {
			return err
		}
		wf.Authorization, err = json.Marshal(acceptedWorkflow{
			Version: 1, AuthorizerID: actor, Profile: credential.Profile, SourceTokenID: credential.TokenID,
			Permission: req.Permission, Target: target, RequestHash: hash, AcceptedAt: time.Now(), ExpiresAt: expiresAt,
		})
		if err != nil {
			return err
		}
		return tx.Table(wf.TableName()).Where(app.Workflow{ID: wf.ID, OrgID: wf.OrgID}).Update("authorization", wf.Authorization).Error
	})
}

func (c *Client) AdmitWorkflow(ctx context.Context, workflowID string, admit func(*gorm.DB, *app.Workflow) error) (Decision, error) {
	if workflowID == "" || admit == nil {
		return Decision{Reason: ReasonInvalidRequest}, nil
	}
	var initial app.Workflow
	if err := c.db.WithContext(ctx).Where(app.Workflow{ID: workflowID}).Take(&initial).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return Decision{Reason: ReasonTargetUnavailable}, nil
		}
		return Decision{}, err
	}
	var decision Decision
	err := c.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := requireRoleMutationIsolation(tx); err != nil {
			return err
		}
		orgErr := lockRoleOrg(tx, initial.OrgID)
		if orgErr != nil && !errors.Is(orgErr, gorm.ErrRecordNotFound) {
			return orgErr
		}
		accepted, valid := decodeAcceptedWorkflow(&initial)
		var account app.Account
		if valid && orgErr == nil {
			var err error
			account, err = lockRoleSubject(tx, accepted.AuthorizerID)
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
		}
		var wf app.Workflow
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where(app.Workflow{ID: workflowID, OrgID: initial.OrgID}).Take(&wf).Error; err != nil {
			return err
		}
		if wf.AuthorizationFailure != "" {
			decision.Reason = ReasonTargetUnavailable
			return nil
		}
		_, cancelRequested := wf.Status.Metadata["cancel_requested_at"]
		if wf.Status.Status == app.StatusCancelled || cancelRequested {
			decision.Reason = ReasonTargetUnavailable
			return tx.Table(wf.TableName()).Where(app.Workflow{ID: wf.ID, OrgID: wf.OrgID}).Update("authorization_failure", "cancelled").Error
		}
		if !wf.FinishedAt.IsZero() {
			decision.Reason = ReasonTargetUnavailable
			return nil
		}
		current, currentValid := decodeAcceptedWorkflow(&wf)
		if orgErr != nil {
			decision.Reason = ReasonTargetUnavailable
		} else if !valid || !currentValid || current != accepted {
			decision.Reason = ReasonInvalidRequest
		} else if account.ID == "" {
			decision.Reason = ReasonCredentialUnavailable
		} else {
			var err error
			decision, err = authorizeAcceptedWorkflow(tx, &wf, accepted, account, time.Now())
			if err != nil {
				return err
			}
		}
		if decision.Allowed {
			return admit(tx, &wf)
		}
		wf.AuthorizationFailure = string(decision.Reason)
		wf.FinishedAt = time.Now()
		wf.Status = app.CompositeStatus{
			Status: app.StatusError, StatusHumanDescription: "workflow authorization failed; submit a new request",
			Metadata: map[string]any{"authorization_failure": wf.AuthorizationFailure},
		}
		if err := tx.Table(wf.TableName()).Where(app.Workflow{ID: wf.ID, OrgID: wf.OrgID}).Update("authorization_failure", wf.AuthorizationFailure).Error; err != nil {
			return err
		}
		return tx.Model(&wf).Select("finished_at", "status", "name").Updates(&wf).Error
	}, &sql.TxOptions{Isolation: sql.LevelReadCommitted})
	if err != nil {
		return Decision{}, err
	}
	return decision, nil
}

func decodeAcceptedWorkflow(wf *app.Workflow) (acceptedWorkflow, bool) {
	var accepted acceptedWorkflow
	strictErrors, err := strictjson.UnmarshalStrict(wf.Authorization, &accepted)
	if err != nil || len(strictErrors) != 0 || accepted.Version != 1 || accepted.AuthorizerID == "" ||
		accepted.SourceTokenID == "" || accepted.RequestHash == "" || accepted.AcceptedAt.IsZero() ||
		!accepted.ExpiresAt.After(accepted.AcceptedAt) || accepted.Target.ValidateShape() != nil ||
		accepted.Target.OrgID != wf.OrgID || accepted.AuthorizerID != wf.CreatedByID ||
		!workflowTargetMatches(wf, accepted.Target) {
		return acceptedWorkflow{}, false
	}
	if accepted.Profile != permissions.ProfileHuman && accepted.Profile != permissions.ProfileServiceAccount {
		return acceptedWorkflow{}, false
	}
	switch accepted.Permission {
	case permissions.AppWrite, permissions.AppDelete, permissions.InstallWrite, permissions.InstallDelete:
		return accepted, accepted.Permission.AppliesTo(accepted.Target.Kind)
	default:
		return acceptedWorkflow{}, false
	}
}

func authorizeAcceptedWorkflow(tx *gorm.DB, wf *app.Workflow, accepted acceptedWorkflow, account app.Account, now time.Time) (Decision, error) {
	if now.Before(accepted.AcceptedAt) || !now.Before(accepted.ExpiresAt) {
		return Decision{Reason: ReasonCredentialExpired}, nil
	}
	if (accepted.Profile == permissions.ProfileHuman && account.AccountType != app.AccountTypeAuth && account.AccountType != app.AccountTypeAuth0) ||
		(accepted.Profile == permissions.ProfileServiceAccount && account.AccountType != app.AccountTypeService) {
		return Decision{Reason: ReasonCredentialNotAllowed}, nil
	}
	var managed int64
	if err := tx.Unscoped().Model(&app.ManagedServiceAccount{}).Where(app.ManagedServiceAccount{AccountID: account.ID}).Count(&managed).Error; err != nil {
		return Decision{}, err
	}
	if managed != 0 {
		return Decision{Reason: ReasonCredentialNotAllowed}, nil
	}
	hash, err := workflowRequestHash(tx, wf)
	if err != nil {
		return Decision{}, err
	}
	if hash != accepted.RequestHash {
		return Decision{Reason: ReasonCeilingExceeded}, nil
	}
	target, err := resolveTarget(tx.Clauses(clause.Locking{Strength: "UPDATE"}).Session(&gorm.Session{}), accepted.Target.Scope)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Decision{Reason: ReasonTargetUnavailable}, nil
	}
	if err != nil {
		return Decision{}, err
	}
	if target != accepted.Target {
		return Decision{Reason: ReasonTargetUnavailable}, nil
	}
	grants, err := loadScopedGrants(tx, accepted.AuthorizerID, target.OrgID)
	if errors.Is(err, errNoScopedGrants) {
		return Decision{Reason: ReasonNoGrant}, nil
	}
	if err != nil {
		return Decision{}, err
	}
	if grants.API != permissions.APICtl {
		return Decision{Reason: ReasonCredentialNotAllowed}, nil
	}
	if !permissions.ContainsGrant(grants.Grants, target, permissions.PermissionSet{accepted.Permission: {}}) {
		return Decision{Reason: ReasonNoGrant}, nil
	}
	return Decision{Allowed: true, Reason: ReasonAllowed, Target: target}, nil
}

func workflowTargetMatches(wf *app.Workflow, target permissions.ResolvedTarget) bool {
	return wf.OwnerID == target.ID && ((wf.OwnerType == "apps" && target.Kind == permissions.KindApp) ||
		(wf.OwnerType == "installs" && target.Kind == permissions.KindInstall))
}

func workflowRequestHash(db *gorm.DB, wf *app.Workflow) (string, error) {
	var row struct{ GenerateStepsSignal json.RawMessage }
	if err := db.Model(&app.Workflow{}).Select("generate_steps_signal").Where(app.Workflow{ID: wf.ID, OrgID: wf.OrgID}).Take(&row).Error; err != nil {
		return "", err
	}
	return request.Hash(struct {
		ID             string
		OrgID          string
		OwnerID        string
		OwnerType      string
		Type           app.WorkflowType
		Metadata       any
		Request        *app.WorkflowRequest
		Role           string
		PlanOnly       bool
		ApprovalOption app.InstallApprovalOption
		Generator      json.RawMessage
	}{wf.ID, wf.OrgID, wf.OwnerID, wf.OwnerType, wf.Type, wf.Metadata, wf.Request, wf.Role, wf.PlanOnly, wf.ApprovalOption, row.GenerateStepsSignal})
}
