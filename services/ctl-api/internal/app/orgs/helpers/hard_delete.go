package helpers

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/customer_managed/transport"
)

func (h *Helpers) HardDelete(ctx context.Context, orgID string) error {
	if err := h.deleteCustomerManagedSupportSnapshotObjects(ctx, orgID); err != nil {
		return err
	}
	if err := h.deleteReleasePackageObjects(ctx, orgID); err != nil {
		return err
	}
	childObjs := []interface{}{
		&app.EventDispatch{},
		&app.TriggerRule{},
		&app.EventRunbookWaiter{},
		&app.TriggerEvent{},
		&app.TriggerSecret{},
		&app.Trigger{},
		&app.InstallReleaseDeployment{},
		&app.InstallRegistration{},
		&app.InstallManagementPolicyVersion{},
		&app.ReleasePackageReplica{},
		&app.ReleasePackageMember{},
		&app.ReleasePackage{},
		&app.AppReleaseMember{},
		&app.AppRelease{},
		&app.CustomerManagedBundleTransportReplica{},
		&app.CustomerManagedBundleArtifact{},
		&app.CustomerManagedBundle{},
		&app.RunnerJobExecutionResult{},
		&app.RunnerJobExecutionOutputs{},
		&app.RunnerJobExecution{},
		&app.RunnerJobPlan{},
		&app.InstallIntermediateData{},
		&app.RunnerJob{},
		&app.Runner{},
		&app.RunnerGroupSettings{},
		&app.RunnerGroup{},
		&app.LogStream{},
		&app.InstallComponent{},
		&app.InstallDeploy{},
		&app.ComponentReleaseStep{},
		&app.ComponentRelease{},
		&app.ComponentBuild{},
		&app.AWSECRImageConfig{},
		&app.GCPGARImageConfig{},
		&app.AzureACRImageConfig{},
		&app.PublicGitVCSConfig{},
		&app.ConnectedGithubVCSConfig{},
		&app.ActionWorkflowTriggerConfig{},
		&app.ActionWorkflowStepConfig{},
		&app.InstallActionWorkflow{},
		&app.InstallActionWorkflowRun{},
		&app.ActionWorkflowConfig{},
		&app.ExternalImageComponentConfig{},
		&app.JobComponentConfig{},
		&app.KubernetesManifestComponentConfig{},
		&app.DockerBuildComponentConfig{},
		&app.TerraformModuleComponentConfig{},
		&app.HelmComponentConfig{},
		&app.InstallActionWorkflowRunStep{},
		&app.InstallActionWorkflowRun{},
		&app.InstallActionWorkflow{},
		&app.ComponentConfigConnection{},
		&app.ComponentDependency{},
		&app.Component{},
		&app.InstallSandboxRun{},
		&app.InstallSandbox{},
		&app.InstallInputs{},
		&app.InstallEvent{},
		&app.InstallSupportSnapshot{},
		&app.Install{},
		&app.AzureAccount{},
		&app.AWSAccount{},
		&app.AppSecret{},
		&app.AppInputConfig{},
		&app.AppInputGroup{},
		&app.AppInput{},
		&app.AppRunnerConfig{},
		&app.AppSandboxConfig{},
		&app.AppConfig{},
		&app.App{},
		&app.VCSConnectionCommit{},
		&app.VCSConnection{},
		&app.InstallerMetadata{},
		&app.Installer{},
		&app.OrgInvite{},
		&app.NotificationsConfig{},
		&app.Policy{},
		&app.AccountRole{},
		&app.Role{},
		&app.QueueSignal{},
		&app.QueueEmitter{},
		&app.Queue{},
	}
	for _, obj := range childObjs {
		res := h.db.WithContext(ctx).Unscoped().
			Where("org_id = ?", orgID).
			Delete(obj)
		if res.Error != nil {
			return fmt.Errorf("unable to delete %T for org: %w", obj, res.Error)
		}
	}

	// delete org
	res := h.db.WithContext(ctx).Unscoped().Delete(&app.Org{
		ID: orgID,
	})
	if res.Error != nil {
		return fmt.Errorf("unable to delete org: %w", res.Error)
	}
	if res.RowsAffected != 1 {
		return fmt.Errorf("org not found %w", gorm.ErrRecordNotFound)
	}

	return nil
}

func (h *Helpers) deleteReleasePackageObjects(ctx context.Context, orgID string) error {
	var replicas []app.ReleasePackageReplica
	if err := h.db.WithContext(ctx).Where(app.ReleasePackageReplica{OrgID: orgID}).Find(&replicas).Error; err != nil {
		return fmt.Errorf("load release package replicas for org: %w", err)
	}
	for _, stored := range replicas {
		replica := transport.Replica{
			Provider: stored.Provider, Region: stored.Region, StorageRef: stored.StorageRef,
			StorageVersion: stored.StorageVersion, TransportChecksum: stored.ArchiveChecksum, Size: stored.Size,
		}
		if stored.VerifiedAt != nil {
			replica.VerifiedAt = *stored.VerifiedAt
		}
		if err := h.customerManagedStore.Delete(ctx, replica); err != nil {
			return fmt.Errorf("delete release package %s archive: %w", stored.PackageID, err)
		}
	}
	return nil
}

func (h *Helpers) deleteCustomerManagedSupportSnapshotObjects(ctx context.Context, orgID string) error {
	var snapshots []app.InstallSupportSnapshot
	if err := h.db.WithContext(ctx).Where(&app.InstallSupportSnapshot{OrgID: orgID}).Find(&snapshots).Error; err != nil {
		return fmt.Errorf("load customer-managed support snapshots for org: %w", err)
	}
	for _, snapshot := range snapshots {
		replica := transport.Replica{
			Provider:       snapshot.StorageProvider,
			Region:         snapshot.StorageRegion,
			StorageRef:     snapshot.StorageRef,
			StorageVersion: snapshot.StorageVersion,
			Size:           snapshot.ArchiveSize,
		}
		var errs []error
		if err := h.customerManagedStore.Delete(ctx, replica); err != nil {
			errs = append(errs, fmt.Errorf("delete support snapshot %s archive: %w", snapshot.ID, err))
		}
		if snapshot.SnapshotBlob != nil && snapshot.SnapshotBlob.Metadata().S3Key != "" {
			if err := h.blobStore.Delete(ctx, snapshot.SnapshotBlob.Metadata().S3Key); err != nil {
				errs = append(errs, fmt.Errorf("delete support snapshot %s data: %w", snapshot.ID, err))
			}
		}
		if err := errors.Join(errs...); err != nil {
			return err
		}
	}
	return nil
}
