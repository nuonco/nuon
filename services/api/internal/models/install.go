// install.go
package models

import (
	"fmt"
	"time"

	installsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/installs/v1"
	"github.com/powertoolsdev/mono/services/api/internal/jobs"
	"gorm.io/gorm"
)

const (
	NUON_AWS_ACCOUNT_ID = "548377525120"
)

type Install struct {
	Model
	CreatedByID string

	Name  string
	AppID string
	App   App

	Domain   Domain          // all the domain stuff
	Settings InstallSettings `gorm:"-" faker:"-"`

	AWSSettings *AWSSettings `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" faker:"-"`
	GCPSettings *GCPSettings `faker:"-"`
}

func (i Install) AfterCreate(tx *gorm.DB) error {
	ctx := tx.Statement.Context
	mgr, err := jobs.FromContext(ctx)
	if err != nil {
		return fmt.Errorf("unable to get job manager: %w", err)
	}

	if err := mgr.CreateInstall(ctx, i.ID); err != nil {
		return fmt.Errorf("unable to create org: %w", err)
	}

	return nil
}

func (i Install) BeforeDelete(tx *gorm.DB) error {
	ctx := tx.Statement.Context
	mgr, err := jobs.FromContext(ctx)
	if err != nil {
		return fmt.Errorf("unable to get job manager: %w", err)
	}

	if err := mgr.DeleteInstall(ctx, i.ID); err != nil {
		return fmt.Errorf("unable to delete install: %w", err)
	}

	return nil
}
func (Install) IsNode() {}

func (i Install) GetID() string {
	return i.Model.ID
}

func (i Install) GetCreatedAt() time.Time {
	return i.Model.CreatedAt
}

func (i Install) GetUpdatedAt() time.Time {
	return i.Model.UpdatedAt
}

func (i Install) ToProvisionRequest(orgID string, sandboxVersion *SandboxVersion) *installsv1.ProvisionRequest {
	return &installsv1.ProvisionRequest{
		OrgId:     orgID,
		AppId:     i.AppID,
		InstallId: i.ID,
		AccountSettings: &installsv1.AccountSettings{
			Region:       i.AWSSettings.Region.ToRegion(),
			AwsAccountId: NUON_AWS_ACCOUNT_ID,
			AwsRoleArn:   i.AWSSettings.IamRoleArn,
		},
		SandboxSettings: &installsv1.SandboxSettings{
			Name:    sandboxVersion.SandboxName,
			Version: sandboxVersion.SandboxVersion,
		},
	}
}

func (i Install) ToDeprovisionRequest(orgID string, sandboxVersion *SandboxVersion) *installsv1.DeprovisionRequest {
	return &installsv1.DeprovisionRequest{
		OrgId:     orgID,
		AppId:     i.AppID,
		InstallId: i.ID,
		AccountSettings: &installsv1.AccountSettings{
			Region:       i.AWSSettings.Region.ToRegion(),
			AwsAccountId: NUON_AWS_ACCOUNT_ID,
			AwsRoleArn:   i.AWSSettings.IamRoleArn,
		},
		SandboxSettings: &installsv1.SandboxSettings{
			Name:    sandboxVersion.SandboxName,
			Version: sandboxVersion.SandboxVersion,
		},
	}
}
