package activities

import (
	"context"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/generics"
)

// InstallUseRunnerInputs carries the values the AWS CloudFormation renderer
// needs to decide whether the stack includes a runner and which principal, if
// any, control-plane install workers should be trusted as.
type InstallUseRunnerInputs struct {
	// UseRunner is the install's desired runner mode (InstallConfig.UseRunner,
	// defaulting to true).
	UseRunner bool
	// CloudConnectionPrincipalARN is the install's cloud connection role, empty
	// unless the install has an AWS cloud connection.
	CloudConnectionPrincipalARN string
}

type GetInstallUseRunnerInputsRequest struct {
	InstallID string `json:"install_id" validate:"required"`
}

// @temporal-gen-v2 activity
// @by-field InstallID
func (a *Activities) GetInstallUseRunnerInputs(ctx context.Context, req GetInstallUseRunnerInputsRequest) (*InstallUseRunnerInputs, error) {
	var install app.Install

	res := a.db.WithContext(ctx).
		Preload("InstallConfig").
		Preload("CloudConnection").
		First(&install, "id = ?", req.InstallID)
	if res.Error != nil {
		return nil, generics.TemporalGormError(res.Error, "unable to get install")
	}

	out := &InstallUseRunnerInputs{
		UseRunner: install.InstallConfig.UsesRunner(),
	}
	if install.CloudConnection != nil && install.CloudConnection.Platform == app.CloudPlatformAWS {
		out.CloudConnectionPrincipalARN = install.CloudConnection.Principal
	}

	return out, nil
}
