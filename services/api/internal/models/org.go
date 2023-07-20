package models

import (
	"fmt"
	"time"

	orgsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/orgs/v1"
	"github.com/powertoolsdev/mono/services/api/internal/jobs"
	"gorm.io/gorm"
)

type Org struct {
	Model

	CreatedByID     string
	Name            string `gorm:"uniqueIndex"`
	Apps            []App  `faker:"-"`
	IsNew           bool   `gorm:"-:all"`
	GithubInstallID string
}

func (o *Org) AfterCreate(tx *gorm.DB) (err error) {
	ctx := tx.Statement.Context
	mgr, err := jobs.FromContext(ctx)
	if err != nil {
		return fmt.Errorf("unable to get job manager: %w", err)
	}

	if err := mgr.CreateOrg(ctx, o.ID); err != nil {
		return fmt.Errorf("unable to create org: %w", err)
	}

	return nil
}

func (o *Org) AfterDelete(tx *gorm.DB) (err error) {
	ctx := tx.Statement.Context
	mgr, err := jobs.FromContext(ctx)
	if err != nil {
		return fmt.Errorf("unable to get job manager: %w", err)
	}

	if err := mgr.DeleteOrg(ctx, o.ID); err != nil {
		return fmt.Errorf("unable to delete org: %w", err)
	}

	return nil
}

func (Org) IsNode() {}

func (o Org) GetID() string {
	return o.Model.ID
}

func (o Org) GetCreatedAt() time.Time {
	return o.Model.CreatedAt
}

func (o Org) GetUpdatedAt() time.Time {
	return o.Model.UpdatedAt
}

func (o Org) ToProvisionRequest() *orgsv1.ProvisionRequest {
	return &orgsv1.ProvisionRequest{
		OrgId:  o.ID,
		Region: "us-west-2",
	}
}

func (o Org) ToDeprovisionRequest() *orgsv1.DeprovisionRequest {
	return &orgsv1.DeprovisionRequest{
		OrgId:  o.ID,
		Region: "us-west-2",
	}
}
