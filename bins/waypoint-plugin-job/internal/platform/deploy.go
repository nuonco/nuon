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
	p.logger.Debug("initialized logger")

	// init k8s client
	p.logger.Debug("getting k8s client set")
	clientset, err := p.getClientset()
	if err != nil {
		p.logger.Error(err.Error())
		return nil, err
	}
	p.logger.Debug("got k8s client set")

	// start k8s job
	p.logger.Debug("starting k8s job")
	job, err := p.startJob(ctx, clientset)
	if err != nil {
		p.logger.Error(err.Error())
		return nil, err
	}
	p.logger.Debug("started k8s job")

	// monitor job
	p.logger.Debug("monitoring k8s job")
	err = p.monitorJob(ctx, clientset, job)
	if err != nil {
		p.logger.Error(err.Error())
		return nil, err
	}
	p.logger.Debug("done monitoring k8s job")

	return &jobv1.Deployment{}, nil
}
