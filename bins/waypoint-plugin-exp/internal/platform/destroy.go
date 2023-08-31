// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package platform

import (
	"context"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/waypoint-plugin-sdk/terminal"
	expv1 "github.com/powertoolsdev/mono/pkg/types/plugins/exp/v1"
)

// Implement the Destroyer interface
func (p *Platform) DestroyFunc() interface{} {
	return p.destroy
}

// A DestroyFunc does not have a strict signature, you can define the parameters
// you need based on the Available parameters that the Waypoint SDK provides.
// Waypoint will automatically inject parameters as specified
// in the signature at run time.
//
// Available input parameters:
// - context.Context
// - *component.Source
// - *component.JobInfo
// - *component.DeploymentConfig
// - hclog.Logger
// - terminal.UI
// - *component.LabelSet
//
// In addition to default input parameters the Deployment from the DeployFunc step
// can also be injected.
//
// The output parameters for PushFunc must be a Struct which can
// be serialzied to Protocol Buffers binary format and an error.
// This Output Value will be made available for other functions
// as an input parameter.
//
// If an error is returned, Waypoint stops the execution flow and
// returns an error to the user.
//
//nolint:all
func (p *Platform) destroy(
	ctx context.Context,
	ui terminal.UI,
	log hclog.Logger,
	deployment *expv1.Deployment,
) error {
	sg := ui.StepGroup()
	defer sg.Wait()

	rm := p.resourceManager(log, nil)

	// If we don't have resource state, this state is from an older version
	// and we need to manually recreate it.
	//if deployment.ResourceState == nil {
	//rm.Resource("deployment").SetState(&expv1.Resource_Deployment{
	//Name: deployment.Name,
	//})
	//} else {
	//// Load our set state
	//if err := rm.LoadState(deployment.ResourceState); err != nil {
	//return err
	//}
	//}

	// Destroy
	return rm.DestroyAll(ctx, log, sg, ui)
}

func (b *Platform) resourceDeploymentDestroy(
	ctx context.Context,
	log hclog.Logger,
	sg terminal.StepGroup,
	ui terminal.UI,
) error {
	// Destroy your deployment resource
	return nil
}
