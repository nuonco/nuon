// deploy.go
package models

import (
	"fmt"

	"github.com/powertoolsdev/mono/pkg/common/shortid/domains"
	deployv1 "github.com/powertoolsdev/mono/pkg/types/api/deploy/v1"
	"github.com/powertoolsdev/mono/services/api/internal/jobs"
	"gorm.io/gorm"
)

type Deploy struct {
	Model

	BuildID string
	Build   Build

	InstallID string
	Install   Install

	ComponentID string
	Component   Component
}

func (d *Deploy) NewID() error {
	if d.ID == "" {
		d.ID = domains.NewDeploymentID()
	}
	return nil
}

func (d *Deploy) ToProto() *deployv1.Deploy {
	return &deployv1.Deploy{
		Id:        d.GetID(),
		InstallId: d.InstallID,
		BuildId:   d.BuildID,
		UpdatedAt: TimeToDatetime(d.UpdatedAt),
		CreatedAt: TimeToDatetime(d.CreatedAt),
	}
}

func (d *Deploy) AfterCreate(tx *gorm.DB) (err error) {
	ctx := tx.Statement.Context
	mgr, err := jobs.FromContext(ctx)
	if err != nil {
		return fmt.Errorf("unable to get job manager: %w", err)
	}

	_, err = mgr.StartDeploy(ctx, d.ID)
	if err != nil {
		return fmt.Errorf("unable to create deploy: %w", err)
	}

	return nil
}
