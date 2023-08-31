// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package platform

import (
	"context"
	"fmt"

	"github.com/hashicorp/go-hclog"
	"github.com/hashicorp/waypoint-plugin-sdk/component"
	"github.com/hashicorp/waypoint-plugin-sdk/framework/resource"
	sdk "github.com/hashicorp/waypoint-plugin-sdk/proto/gen"
	"github.com/hashicorp/waypoint-plugin-sdk/terminal"
	"github.com/powertoolsdev/mono/bins/waypoint-plugin-exp/internal/registry"
	expv1 "github.com/powertoolsdev/mono/pkg/types/plugins/exp/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

//nolint:all
type DeployConfig struct {
	Region string "hcl:directory,optional"
}

type Platform struct {
	config DeployConfig
}

// Implement Configurable
//
//nolint:all
func (p *Platform) Config() (interface{}, error) {
	return &p.config, nil
}

// Implement ConfigurableNotify
//
//nolint:all
func (p *Platform) ConfigSet(config interface{}) error {
	c, ok := config.(*DeployConfig)
	if !ok {
		// The Waypoint SDK should ensure this never gets hit
		return fmt.Errorf("Expected *DeployConfig as parameter")
	}

	// validate the config
	if c.Region == "" {
		return fmt.Errorf("Region must be set to a valid directory")
	}

	return nil
}

// This function can be implemented to return various connection info required
// to connect to your given platform for Resource Manager. It could return
// a struct with client information, what namespace to connect to, a config,
// and so on.
func (p *Platform) getConnectContext() (interface{}, error) {
	return nil, nil
}

// Resource manager will tell the Waypoint Plugin SDK how to create and delete
// certain resources for your deployments.
//
// For example, your deployment might need to create a "container" or "load balancer".
// Your plugin could implement two resources through ResourceManager and the Waypoint
// Plugin SDK will automatically create or delete these resources as well as
// obtain the defined status for them.
//
// ResourceManager can also be implemented for Release as well.
func (p *Platform) resourceManager(log hclog.Logger, dcr *component.DeclaredResourcesResp) *resource.Manager {
	return resource.NewManager(
		resource.WithLogger(log.Named("resource_manager")),
		resource.WithValueProvider(p.getConnectContext),
		resource.WithDeclaredResourcesResp(dcr),
		resource.WithResource(resource.NewResource(
			resource.WithName("template_example"),
			resource.WithState(&expv1.Resource_Deployment{}),
			resource.WithCreate(p.resourceDeploymentCreate),
			resource.WithDestroy(p.resourceDeploymentDestroy),
			resource.WithStatus(p.resourceDeploymentStatus),
			resource.WithPlatform("template_platform"),                                         // Update this to match your plugins platform, like Kubernetes
			resource.WithCategoryDisplayHint(sdk.ResourceCategoryDisplayHint_INSTANCE_MANAGER), // This is meant for the UI to determine what kind of icon to show
		)),
		// NOTE: Add more resource funcs here if your plugin has more than 1 resource
	)
}

// Implement Builder
func (p *Platform) DeployFunc() interface{} {
	// return a function which will be called by Waypoint
	return p.deploy
}

func (p *Platform) StatusFunc() interface{} {
	return p.status
}

// A BuildFunc does not have a strict signature, you can define the parameters
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

// In addition to default input parameters the registry.Artifact from the Build step
// can also be injected.
//
// The output parameters for BuildFunc must be a Struct which can
// be serialzied to Protocol Buffers binary format and an error.
// This Output Value will be made available for other functions
// as an input parameter.
// If an error is returned, Waypoint stops the execution flow and
// returns an error to the user.
func (b *Platform) deploy(
	ctx context.Context,
	ui terminal.UI,
	log hclog.Logger,
	dcr *component.DeclaredResourcesResp,
	artifact *registry.Artifact,
) (*expv1.Deployment, error) {
	u := ui.Status()
	defer u.Close()
	u.Update("Deploy application")

	var result *expv1.Deployment

	// Create our resource manager and create deployment resources
	rm := b.resourceManager(log, dcr)

	// These params must match exactly to your resource manager functions. Otherwise
	// they will not be invoked during CreateAll()
	if err := rm.CreateAll(
		ctx, log, u, ui,
		artifact, &result,
	); err != nil {
		return nil, err
	}

	// Store our resource state
	//result.ResourceState = rm.State()

	u.Update("Application deployed")

	return &expv1.Deployment{}, nil
}

// This function is the top level status command that gets invoked when Waypoint
// attempts to determine the health of a dpeloyment. It will also invoke the
// status for each resource involed for the given deployment if any.
//
//nolint:all
func (d *Platform) status(
	ctx context.Context,
	ji *component.JobInfo,
	ui terminal.UI,
	log hclog.Logger,
	deployment *expv1.Deployment,
) (*sdk.StatusReport, error) {
	sg := ui.StepGroup()
	s := sg.Add("Checking the status of the deployment...")

	rm := d.resourceManager(log, nil)

	// If we don't have resource state, this state is from an older version
	// and we need to manually recreate it.
	//if deployment.ResourceState == nil {
	//rm.Resource("deployment").SetState(&expv1.Resource_Deployment{
	//Name: deployment.Id,
	//})
	//} else {
	//// Load our set state
	//if err := rm.LoadState(deployment.ResourceState); err != nil {
	//return nil, err
	//}
	//}

	// This will call the StatusReport func on every defined resource in ResourceManager
	report, err := rm.StatusReport(ctx, log, sg, ui)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "resource manager failed to generate resource statuses: %s", err)
	}

	report.Health = sdk.StatusReport_UNKNOWN
	s.Update("Deployment is currently not implemented!")
	s.Done()

	return report, nil
}

func (b *Platform) resourceDeploymentCreate(
	ctx context.Context,
	log hclog.Logger,
	st terminal.Status,
	ui terminal.UI,
	artifact *registry.Artifact,
	result *expv1.Deployment,
) error {
	// Create your deployment resource here!

	return nil
}

func (b *Platform) resourceDeploymentStatus(
	ctx context.Context,
	ui terminal.UI,
	sg terminal.StepGroup,
	artifact *registry.Artifact,
) error {
	// Determine health status of "this" resource.
	return nil
}
