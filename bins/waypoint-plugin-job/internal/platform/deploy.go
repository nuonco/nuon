package platform

import (
	"context"

	"github.com/hashicorp/waypoint-plugin-sdk/terminal"
	jobv1 "github.com/powertoolsdev/mono/pkg/types/plugins/job/v1"
)

func (p *Platform) DeployFunc() interface{} {
	return p.deploy
}

func (p *Platform) deploy(
	ctx context.Context,
	ui terminal.UI,
) (*jobv1.Deployment, error) {

	// init logger
	err := p.initLogger(ui)
	if err != nil {
		return nil, err
	}

	clientset, err := p.getClientset()
	if err != nil {
		return nil, err
	}

	// start k8s job
	ui.Output("starting job")
	job, err := p.startJob(ctx, clientset)
	if err != nil {
		ui.Output("error starting job: %v", err)
		return nil, err
	}

	// monitor job
	ui.Output("polling job")
	err = p.pollJob(ctx, clientset, job)
	if err != nil {
		ui.Output("error polling job: %v", err)
		return nil, err
	}

	return &jobv1.Deployment{}, nil
}
