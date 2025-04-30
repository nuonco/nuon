package jobloop

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/nuonco/nuon-runner-go/models"

	"github.com/powertoolsdev/mono/bins/runner/internal/jobs"
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
			"tag":           "v1.2.3",
			"repository":    "nuon/app-service",
			"media_type":    "application/vnd.docker.distribution.manifest.v2+json",
			"digest":        "sha256:a123b456c789d012e345f678g901h234i567j890k123l456m789n012o345p",
			"size":          28437192,
			"urls":          []string{"registry.example.com/nuon/app-service:v1.2.3"},
			"annotations":   map[string]string{"org.opencontainers.image.created": "2024-04-29T10:15:30Z"},
			"artifact_type": "application/vnd.docker.container.image.v1+json",
			"platform": map[string]any{
				"architecture": "arm64",
				"os":           "linux",
				"os_version":   "10.0",
				"variant":      "v8",
				"os_features":  []string{"sse4", "aes"},
			},
		},

		// copied from the aws-eks output
		"account": map[string]any{
			"id":     "123456789012",
			"region": "us-west-2",
		},
		"cluster": map[string]any{
			"arn":                        "arn:aws:eks:us-west-2:123456789012:cluster/nuon-cluster",
			"certificate_authority_data": "LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUN5RENDQWJDZ0F3SUJBZ0lCQURBTkJna3Foa2lHOXcwQkFRc0ZBREF2TVMwd0t3WURWUVFERXlRME4yVTEKWkRNeE5DMDROelk...",
			"endpoint":                   "https://A1B2C3D4E5F6.gr7.us-west-2.eks.amazonaws.com",
			"name":                       "nuon-cluster",
			"platform_version":           "eks.9",
			"status":                     "ACTIVE",
			"oidc_issuer_url":            "https://oidc.eks.us-west-2.amazonaws.com/id/A1B2C3D4E5F6",
			"oidc_provider_arn":          "arn:aws:iam::123456789012:oidc-provider/oidc.eks.us-west-2.amazonaws.com/id/A1B2C3D4E5F6",
			"cluster_security_group_id":  "sg-0abc123def456",
			"node_security_group_id":     "sg-0xyz789uvw456",
		},
		"vpc": map[string]any{
			"id":                         "vpc-0abc123def456",
			"arn":                        "arn:aws:ec2:us-west-2:123456789012:vpc/vpc-0abc123def456",
			"cidr":                       "10.0.0.0/16",
			"azs":                        []string{"us-west-2a", "us-west-2b", "us-west-2c"},
			"private_subnet_cidr_blocks": []string{"10.0.1.0/24", "10.0.2.0/24", "10.0.3.0/24"},
			"private_subnet_ids":         []string{"subnet-0abc123def456", "subnet-0ghi789jkl012", "subnet-0mno345pqr678"},
			"public_subnet_cidr_blocks":  []string{"10.0.4.0/24", "10.0.5.0/24", "10.0.6.0/24"},
			"public_subnet_ids":          []string{"subnet-0stu901vwx234", "subnet-0yza567bcd890", "subnet-0efg123hij456"},
			"runner_subnet_id":           "subnet-0klm789nop012",
			"runner_subnet_cidr":         "10.0.7.0/24",
			"default_security_group_id":  "sg-0qrs345tuv678",
		},
		"ecr": map[string]any{
			"repository_url":  "123456789012.dkr.ecr.us-west-2.amazonaws.com/nuon-app",
			"repository_arn":  "arn:aws:ecr:us-west-2:123456789012:repository/nuon-app",
			"repository_name": "nuon-app",
			"registry_id":     "123456789012",
			"registry_url":    "123456789012.dkr.ecr.us-west-2.amazonaws.com",
		},
		"nuon_dns": map[string]any{
			"enabled": true,
			"public_domain": map[string]any{
				"zone_id":     "Z1A2B3C4D5E6F7",
				"name":        "example.com",
				"nameservers": []string{"ns-1234.awsdns-12.org", "ns-567.awsdns-34.com", "ns-890.awsdns-56.net", "ns-1234.awsdns-78.co.uk"},
			},
			"internal_domain": map[string]any{
				"zone_id":     "Z8G9H0I1J2K3L4",
				"name":        "internal.example.com",
				"nameservers": []string{"ns-5678.awsdns-90.org", "ns-123.awsdns-12.com", "ns-456.awsdns-34.net", "ns-789.awsdns-56.co.uk"},
			},
			"alb_ingress_controller": map[string]any{
				"enabled":  true,
				"id":       "alb-ingress-controller",
				"chart":    "aws-load-balancer-controller",
				"revision": "1.4.7",
			},
			"external_dns": map[string]any{
				"enabled":  true,
				"id":       "external-dns",
				"chart":    "external-dns",
				"revision": "1.12.1",
			},
			"cert_manager": map[string]any{
				"enabled":  true,
				"id":       "cert-manager",
				"chart":    "cert-manager",
				"revision": "1.11.0",
			},
			"ingress_nginx": map[string]any{
				"enabled":  true,
				"id":       "ingress-nginx",
				"chart":    "ingress-nginx",
				"revision": "4.7.1",
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
