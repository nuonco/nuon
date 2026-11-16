package credentialcleanup

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/account"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/queuenames"
	queuesignal "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const SignalType queuesignal.SignalType = "credential-cleanup"

type Request struct {
	ServiceAccountSubject string
	AccountID             string
	OrgID                 string
	// latestTokenID separates a later close, which minted a new token, from a
	// retry of the same close. It is not part of the signal payload.
	latestTokenID string
}

func (r Request) Signal() queuesignal.Signal {
	data := map[string]any{}
	if r.ServiceAccountSubject != "" {
		data["service_account_subject"] = r.ServiceAccountSubject
	}
	if r.AccountID != "" {
		data["account_id"] = r.AccountID
	}
	if r.OrgID != "" {
		data["org_id"] = r.OrgID
	}
	return queuesignal.NewRaw(SignalType, data)
}

func (r Request) IdempotencyKey() string {
	if r.ServiceAccountSubject != "" {
		key := "credential-cleanup:" + r.ServiceAccountSubject
		if r.latestTokenID != "" {
			key += ":" + r.latestTokenID
		}
		return key
	}
	return "credential-cleanup:" + r.OrgID + ":" + r.AccountID
}

type Queue struct {
	ID        string
	OwnerID   string
	OwnerType string
}

func EnqueueOnQueue(ctx context.Context, qc *queueclient.Client, q Queue, req Request) error {
	if qc == nil {
		return fmt.Errorf("queue client is unavailable")
	}
	if _, err := qc.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{
		QueueID:        q.ID,
		OwnerID:        q.OwnerID,
		OwnerType:      q.OwnerType,
		Signal:         req.Signal(),
		IdempotencyKey: req.IdempotencyKey(),
	}); err != nil {
		return fmt.Errorf("enqueue credential cleanup: %w", err)
	}
	return nil
}

func EnqueueForLogStream(ctx context.Context, db *gorm.DB, qc *queueclient.Client, streamID string) error {
	if qc == nil {
		return fmt.Errorf("queue client is unavailable")
	}
	ownerID, ownerType, queueName, err := logStreamQueue(ctx, db, streamID)
	if err != nil {
		return err
	}
	q, err := qc.GetQueueByOwnerAndName(ctx, ownerID, ownerType, queueName)
	if err != nil {
		return fmt.Errorf("find credential cleanup queue: %w", err)
	}
	tokenID, err := latestLiveTokenID(ctx, db, streamID)
	if err != nil {
		return err
	}
	if tokenID == "" {
		return nil
	}
	return EnqueueOnQueue(ctx, qc, Queue{ID: q.ID, OwnerID: ownerID, OwnerType: ownerType}, Request{
		ServiceAccountSubject: streamID,
		latestTokenID:         tokenID,
	})
}

func latestLiveTokenID(ctx context.Context, db *gorm.DB, subject string) (string, error) {
	var acct app.Account
	err := db.WithContext(ctx).
		Select("id").
		Where(app.Account{Email: strings.ToLower(account.ServiceAccountEmail(subject))}).
		Take(&acct).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("load service account: %w", err)
	}
	var token app.Token
	err = db.WithContext(ctx).
		Select("id").
		Where(app.Token{AccountID: acct.ID}).
		Order("created_at DESC, id DESC").
		Take(&token).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("load service account token: %w", err)
	}
	return token.ID, nil
}

func logStreamQueue(ctx context.Context, db *gorm.DB, streamID string) (ownerID, ownerType, queueName string, err error) {
	var stream app.LogStream
	if err := db.WithContext(ctx).Unscoped().
		Select("id", "org_id", "owner_id", "owner_type").
		Where(app.LogStream{ID: streamID}).
		Take(&stream).Error; err != nil {
		return "", "", "", fmt.Errorf("load log stream: %w", err)
	}

	installID, runnerID := resolveLogStreamOwner(ctx, db, stream)
	switch {
	case installID != "":
		return installID, queuenames.OwnerInstalls, queuenames.InstallSignalsQueueName, nil
	case runnerID != "":
		return runnerID, queuenames.OwnerRunners, queuenames.RunnerSignalsQueueName, nil
	case stream.OrgID != "":
		return stream.OrgID, queuenames.OwnerOrgs, queuenames.OrgSignalsQueueName, nil
	default:
		return "", "", "", fmt.Errorf("log stream %s has no credential cleanup queue", streamID)
	}
}

func resolveLogStreamOwner(ctx context.Context, db *gorm.DB, stream app.LogStream) (installID, runnerID string) {
	switch stream.OwnerType {
	case plugins.TableName(db, app.Install{}):
		return stream.OwnerID, ""
	case plugins.TableName(db, app.InstallSandboxRun{}), plugins.TableName(db, app.InstallActionWorkflowRun{}):
		return lookupInstallID(ctx, db, stream.OwnerType, stream.OwnerID), ""
	case plugins.TableName(db, app.InstallDeploy{}):
		return installIDForDeploy(ctx, db, stream.OwnerID), ""
	// TableName has a pointer receiver. A value falls through to GORM's
	// workflow_steps name and misses the stored install_workflow_steps owner.
	case plugins.TableNameOf[app.WorkflowStep]():
		return installIDForWorkflowStep(ctx, db, stream.OwnerID), ""
	case plugins.TableName(db, app.Runner{}):
		return "", stream.OwnerID
	case plugins.TableName(db, app.RunnerProcess{}):
		return "", lookupColumn(ctx, db, &app.RunnerProcess{}, stream.OwnerID, "runner_id")
	case plugins.TableName(db, app.RunnerJob{}):
		return "", lookupColumn(ctx, db, &app.RunnerJob{}, stream.OwnerID, "runner_id")
	default:
		return "", ""
	}
}

func lookupInstallID(ctx context.Context, db *gorm.DB, ownerType, ownerID string) string {
	var installID string
	err := db.WithContext(ctx).Unscoped().
		Table(ownerType).
		Select("install_id").
		Where("id = ?", ownerID).
		Scan(&installID).Error
	if err != nil {
		return ""
	}
	return installID
}

func installIDForDeploy(ctx context.Context, db *gorm.DB, deployID string) string {
	var deploy app.InstallDeploy
	if err := db.WithContext(ctx).Unscoped().
		Select("install_component_id").
		Where(app.InstallDeploy{ID: deployID}).
		Take(&deploy).Error; err != nil {
		return ""
	}
	var component app.InstallComponent
	if err := db.WithContext(ctx).Unscoped().
		Select("install_id").
		Where(app.InstallComponent{ID: deploy.InstallComponentID}).
		Take(&component).Error; err != nil {
		return ""
	}
	return component.InstallID
}

func installIDForWorkflowStep(ctx context.Context, db *gorm.DB, stepID string) string {
	var step app.WorkflowStep
	if err := db.WithContext(ctx).Unscoped().
		Select("install_workflow_id").
		Where(app.WorkflowStep{ID: stepID}).
		Take(&step).Error; err != nil || step.InstallWorkflowID == "" {
		return ""
	}
	var wf app.Workflow
	if err := db.WithContext(ctx).Unscoped().
		Select("owner_id", "owner_type").
		Where(app.Workflow{ID: step.InstallWorkflowID}).
		Take(&wf).Error; err != nil {
		return ""
	}
	if wf.OwnerType != plugins.TableName(db, app.Install{}) {
		return ""
	}
	return wf.OwnerID
}

func lookupColumn(ctx context.Context, db *gorm.DB, model any, id, column string) string {
	var value string
	err := db.WithContext(ctx).Unscoped().
		Model(model).
		Select(column).
		Where("id = ?", id).
		Scan(&value).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return ""
	}
	return value
}
