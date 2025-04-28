package jobloop

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/nuonco/nuon-runner-go/models"

	"github.com/powertoolsdev/mono/bins/runner/internal/jobs"
	"github.com/powertoolsdev/mono/pkg/generics"
)

func (j *jobLoop) sandboxOutputs() map[string]interface{} {
	return map[string]interface{}{
		"sandbox-outputs": map[string]interface{}{
			"sandbox-mode": true,
			"map": map[string]interface{}{
				"k": "v",
			},
		},
		"image": map[string]interface{}{
			"tag": "local",
		},

		// copied from the aws-eks output
		"account": map[string]any{
			"id":     generics.GetFakeObj[string](),
			"region": generics.GetFakeObj[string](),
		},
		"cluster": map[string]any{
			"arn":                        generics.GetFakeObj[string](),
			"certificate_authority_data": generics.GetFakeObj[string](),
			"endpoint":                   generics.GetFakeObj[string](),
			"name":                       generics.GetFakeObj[string](),
			"platform_version":           generics.GetFakeObj[string](),
			"status":                     generics.GetFakeObj[string](),
			"oidc_issuer_url":            generics.GetFakeObj[string](),
			"oidc_provider_arn":          generics.GetFakeObj[string](),
			"cluster_security_group_id":  generics.GetFakeObj[string](),
			"node_security_group_id":     generics.GetFakeObj[string](),
		},
		"vpc": map[string]any{
			"id":                         generics.GetFakeObj[string](),
			"arn":                        generics.GetFakeObj[string](),
			"cidr":                       generics.GetFakeObj[string](),
			"azs":                        generics.GetFakeObj[[]string](),
			"private_subnet_cidr_blocks": generics.GetFakeObj[[]string](),
			"private_subnet_ids":         generics.GetFakeObj[[]string](),
			"public_subnet_cidr_blocks":  generics.GetFakeObj[[]string](),
			"public_subnet_ids":          generics.GetFakeObj[[]string](),
			"runner_subnet_id":           generics.GetFakeObj[string](),
			"runner_subnet_cidr":         generics.GetFakeObj[string](),
			"default_security_group_id":  generics.GetFakeObj[string](),
		},
		"ecr": map[string]any{
			"repository_url":  generics.GetFakeObj[string](),
			"repository_arn":  generics.GetFakeObj[string](),
			"repository_name": generics.GetFakeObj[string](),
			"registry_id":     generics.GetFakeObj[string](),
			"registry_url":    generics.GetFakeObj[string](),
		},
		"nuon_dns": map[string]any{
			"enabled": generics.GetFakeObj[bool](),
			"public_domain": map[string]any{
				"zone_id":     generics.GetFakeObj[string](),
				"name":        generics.GetFakeObj[string](),
				"nameservers": generics.GetFakeObj[[]string](),
			},
			"internal_domain": map[string]any{
				"zone_id":     generics.GetFakeObj[string](),
				"name":        generics.GetFakeObj[string](),
				"nameservers": generics.GetFakeObj[[]string](),
			},
			"alb_ingress_controller": map[string]any{
				"enabled":  generics.GetFakeObj[bool](),
				"id":       generics.GetFakeObj[string](),
				"chart":    generics.GetFakeObj[string](),
				"revision": generics.GetFakeObj[string](),
			},
			"external_dns": map[string]any{
				"enabled":  generics.GetFakeObj[bool](),
				"id":       generics.GetFakeObj[string](),
				"chart":    generics.GetFakeObj[string](),
				"revision": generics.GetFakeObj[string](),
			},
			"cert_manager": map[string]any{
				"enabled":  generics.GetFakeObj[bool](),
				"id":       generics.GetFakeObj[string](),
				"chart":    generics.GetFakeObj[string](),
				"revision": generics.GetFakeObj[string](),
			},
			"ingress_nginx": map[string]any{
				"enabled":  generics.GetFakeObj[bool](),
				"id":       generics.GetFakeObj[string](),
				"chart":    generics.GetFakeObj[string](),
				"revision": generics.GetFakeObj[string](),
			},
		},
	}
}

func (j *jobLoop) executeOutputsJobStep(ctx context.Context, handler jobs.JobHandler, job *models.AppRunnerJob, jobExecution *models.AppRunnerJobExecution) error {
	var (
		outputs map[string]interface{}
		err     error
	)

	if j.isSandbox(job) {
		outputs = j.sandboxOutputs()
	} else {
		outputs, err = handler.Outputs(ctx)
		if err != nil {
			return errors.Wrap(err, "unable to get outputs")
		}
	}

	_, err = j.apiClient.CreateJobExecutionOutputs(ctx, job.ID, jobExecution.ID, &models.ServiceCreateRunnerJobExecutionOutputsRequest{
		Outputs: outputs,
	})
	if err != nil {
		return errors.Wrap(err, "unable to write outputs to api")
	}

	return nil
}
