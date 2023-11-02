package platform

import (
	"context"

	"github.com/hashicorp/waypoint-plugin-sdk/component"
	"github.com/hashicorp/waypoint-plugin-sdk/terminal"
	jobv1 "github.com/powertoolsdev/mono/pkg/types/plugins/job/v1"
)

func (p *Platform) DeployFunc() interface{} {
	return p.deploy
}

func (p *Platform) deploy(
	ctx context.Context,
	ji *component.JobInfo,
	ui terminal.UI,
) (*jobv1.Deployment, error) {

	// TODO: uncomment this when the logging bug is fixed.
	// Will use ui.Output for now.
	// stdout, _, err := ui.OutputWriters()
	// if err != nil {
	// 	return nil, fmt.Errorf("unable to get output writers")
	// }
	// logger := hclog.New(&hclog.LoggerOptions{
	// 	Name:   "job",
	// 	Output: stdout,
	// })
	// logger.Error("LOGGER TEST")

	// init k8s client
	clientset, err := p.getClientset()
	if err != nil {
		return nil, err
	}

	// start k8s job
	_, err = p.startJob(ctx, clientset, ji)
	if err != nil {
		return nil, err
	}

	return &jobv1.Deployment{}, nil
}
