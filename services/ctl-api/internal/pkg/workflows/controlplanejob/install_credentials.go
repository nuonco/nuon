package controlplanejob

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	assumerole "github.com/nuonco/nuon/pkg/aws/assume-role"
	awscredentials "github.com/nuonco/nuon/pkg/aws/credentials"
	"github.com/nuonco/nuon/pkg/kube"
	plantypes "github.com/nuonco/nuon/pkg/plans/types"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// maxAWSSessionNameLen is the AWS-enforced upper bound on sts session names.
const maxAWSSessionNameLen = 64

// injectInstallCredentials mints short-lived cloud connection credentials and
// sets them as the two-step source credentials on every AWS auth config in a
// control-plane install job's plan, so the executor can assume the install's
// stack roles without a runner process in the customer's account.
//
// It is a no-op unless the job's runner belongs to an install's runner group
// and that runner's status is app.RunnerStatusNotUsed (jobs for build/
// org-runner owners, or installs with a live runner, are left untouched).
// Credentials are only ever set on the in-memory plan returned to the
// executor -- never persisted.
func (a *Activities) injectInstallCredentials(ctx context.Context, jobID string, plan *plantypes.CompositePlan) error {
	authConfigs := collectAWSAuthConfigs(plan)
	if len(authConfigs) == 0 {
		return nil
	}

	job, err := a.getJob(ctx, jobID)
	if err != nil {
		return err
	}
	if job.RunnerID == "" {
		return nil
	}

	var runner app.Runner
	if err := a.db.WithContext(ctx).
		Preload("RunnerGroup").
		Where(&app.Runner{ID: job.RunnerID}).
		First(&runner).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return fmt.Errorf("unable to get runner for install credential injection: %w", err)
	}
	if runner.RunnerGroup.OwnerType != "installs" || runner.Status != app.RunnerStatusNotUsed {
		return nil
	}

	var install app.Install
	if err := a.db.WithContext(ctx).
		Where(&app.Install{ID: runner.RunnerGroup.OwnerID}).
		First(&install).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return fmt.Errorf("unable to get install for credential injection: %w", err)
	}
	if install.CloudConnectionID == nil || *install.CloudConnectionID == "" {
		return nil
	}

	var connection app.CloudConnection
	if err := a.db.WithContext(ctx).
		Where(&app.CloudConnection{ID: *install.CloudConnectionID}).
		First(&connection).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return fmt.Errorf("unable to get cloud connection for credential injection: %w", err)
	}
	if connection.Platform != app.CloudPlatformAWS {
		return nil
	}

	sessionName := truncateSessionName(fmt.Sprintf("nuon-install-job-%s", jobID))
	return mintAndInjectCredentials(ctx, a.cloudConnections, &connection, sessionName, authConfigs)
}

// InjectAWSAuthCredentials mints short-lived cloud connection credentials for
// install and sets them as the two-step source credentials on every AWS auth
// config in authConfigs, the same substitution injectInstallCredentials makes
// for a control-plane job's plan. It is the shared path for anything running
// AWS auth on behalf of a runnerless install without a runner process in the
// customer's account -- today that is control-plane install jobs
// (injectInstallCredentials, above) and the install component health
// workflow's report activity (installcomponenthealth.Activities).
//
// A no-op when the install has no cloud connection, or the connection is not
// AWS: callers should treat that as nothing-to-inject rather than an error.
func InjectAWSAuthCredentials(
	ctx context.Context,
	db *gorm.DB,
	cc cloudConnectionsCredentialsHelper,
	installID string,
	authConfigs []*awscredentials.Config,
) error {
	if len(authConfigs) == 0 {
		return nil
	}

	var install app.Install
	if err := db.WithContext(ctx).
		Where(&app.Install{ID: installID}).
		First(&install).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return fmt.Errorf("unable to get install for credential injection: %w", err)
	}
	if install.CloudConnectionID == nil || *install.CloudConnectionID == "" {
		return nil
	}

	var connection app.CloudConnection
	if err := db.WithContext(ctx).
		Where(&app.CloudConnection{ID: *install.CloudConnectionID}).
		First(&connection).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return fmt.Errorf("unable to get cloud connection for credential injection: %w", err)
	}
	if connection.Platform != app.CloudPlatformAWS {
		return nil
	}

	sessionName := truncateSessionName(fmt.Sprintf("nuon-install-%s", installID))
	return mintAndInjectCredentials(ctx, cc, &connection, sessionName, authConfigs)
}

// cloudConnectionsCredentialsHelper is the one method InjectAWSAuthCredentials
// needs, so callers outside this package (installcomponenthealth) don't have
// to depend on the full cloudconnectionshelpers.Helpers type.
type cloudConnectionsCredentialsHelper interface {
	Credentials(ctx context.Context, connection *app.CloudConnection, sessionName string) (*awscredentials.Config, error)
}

func mintAndInjectCredentials(
	ctx context.Context,
	cc cloudConnectionsCredentialsHelper,
	connection *app.CloudConnection,
	sessionName string,
	authConfigs []*awscredentials.Config,
) error {
	minted, err := cc.Credentials(ctx, connection, sessionName)
	if err != nil {
		return fmt.Errorf("unable to mint install credentials: %w", err)
	}
	if minted.Static == nil {
		return fmt.Errorf("cloud connection credentials for install did not include static credentials")
	}

	src := assumerole.StaticCredentials{
		AccessKeyID:     minted.Static.AccessKeyID,
		SecretAccessKey: minted.Static.SecretAccessKey,
		SessionToken:    minted.Static.SessionToken,
	}

	for _, cfg := range authConfigs {
		if cfg.AssumeRole == nil || cfg.AssumeRole.RoleARN == "" {
			continue
		}
		// Replace any runner-side jump role: the cloud connection session
		// assumes RoleARN directly.
		cfg.AssumeRole.TwoStepConfig = &assumerole.TwoStepConfig{SrcStaticCredentials: src}
	}

	return nil
}

func truncateSessionName(s string) string {
	if len(s) <= maxAWSSessionNameLen {
		return s
	}
	return s[:maxAWSSessionNameLen]
}

// collectAWSAuthConfigs walks every AWS auth config in a composite plan: the
// sandbox run plan, deploy plans of every type, sync secrets, the action
// workflow run plan, and any nested kube.ClusterInfo.AWSAuth.
func collectAWSAuthConfigs(plan *plantypes.CompositePlan) []*awscredentials.Config {
	if plan == nil {
		return nil
	}

	var configs []*awscredentials.Config
	add := func(c *awscredentials.Config) {
		if c != nil {
			configs = append(configs, c)
		}
	}
	addCluster := func(ci *kube.ClusterInfo) {
		if ci != nil {
			add(ci.AWSAuth)
		}
	}

	if dp := plan.DeployPlan; dp != nil {
		if dp.HelmDeployPlan != nil {
			add(dp.HelmDeployPlan.AWSAuth)
			addCluster(dp.HelmDeployPlan.ClusterInfo)
		}
		if dp.TerraformDeployPlan != nil {
			add(dp.TerraformDeployPlan.AWSAuth)
			addCluster(dp.TerraformDeployPlan.ClusterInfo)
		}
		if dp.KubernetesManifestDeployPlan != nil {
			add(dp.KubernetesManifestDeployPlan.AWSAuth)
			addCluster(dp.KubernetesManifestDeployPlan.ClusterInfo)
		}
		if dp.PulumiDeployPlan != nil {
			add(dp.PulumiDeployPlan.AWSAuth)
			addCluster(dp.PulumiDeployPlan.ClusterInfo)
		}
	}
	if plan.SandboxRunPlan != nil {
		add(plan.SandboxRunPlan.AWSAuth)
	}
	if plan.SyncSecretsPlan != nil {
		add(plan.SyncSecretsPlan.AWSAuth)
		addCluster(plan.SyncSecretsPlan.ClusterInfo)
	}
	if plan.ActionWorkflowRunPlan != nil {
		add(plan.ActionWorkflowRunPlan.AWSAuth)
		addCluster(plan.ActionWorkflowRunPlan.ClusterInfo)
	}

	return configs
}
