// deployment.go
package models

import (
	"fmt"
	"time"

	componentv1 "github.com/powertoolsdev/mono/pkg/types/components/component/v1"
	deploymentsv1 "github.com/powertoolsdev/mono/pkg/types/workflows/deployments/v1"
	"github.com/powertoolsdev/mono/services/api/internal/jobs"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/types/known/durationpb"
	"gorm.io/gorm"
)

type Deployment struct {
	Model

	ComponentID string
	Component   Component `fake:"skip"`
	CreatedByID string

	CommitHash   string `json:"commit_hash"`
	CommitAuthor string `json:"commit_author"`
}

const (
	defaultDeployTimeout time.Duration = time.Second * 10
)

func (d Deployment) AfterCreate(tx *gorm.DB) error {
	ctx := tx.Statement.Context
	mgr, err := jobs.FromContext(ctx)
	if err != nil {
		return fmt.Errorf("unable to get job manager: %w", err)
	}

	if err := mgr.CreateDeployment(ctx, d.ID); err != nil {
		return fmt.Errorf("unable to create deployment: %w", err)
	}

	return nil
}

func (Deployment) IsNode() {}

func (d Deployment) GetID() string {
	return d.Model.ID
}

func (d Deployment) GetCreatedAt() time.Time {
	return d.Model.CreatedAt
}

func (d Deployment) GetUpdatedAt() time.Time {
	return d.Model.UpdatedAt
}

func (d Deployment) ToStartRequest() (*deploymentsv1.StartRequest, error) {
	app := d.Component.App
	orgID, appID, componentID, deploymentID := app.OrgID, app.ID, d.Component.ID, d.ID
	var compConf componentv1.Component
	if d.Component.Config != nil {
		if err := protojson.Unmarshal([]byte(d.Component.Config.String()), &compConf); err != nil {
			return nil, fmt.Errorf("failed to unmarshal DB JSON: %w", err)
		}
	}
	compConf.Id = componentID
	compConf.DeployCfg.Timeout = durationpb.New(defaultDeployTimeout)

	req := &deploymentsv1.StartRequest{
		OrgId:        orgID,
		AppId:        appID,
		DeploymentId: deploymentID,
		InstallIds:   make([]string, len(app.Installs)),
		Component:    &compConf,
	}
	for idx, install := range app.Installs {
		req.InstallIds[idx] = install.ID
	}
	return req, nil
}
