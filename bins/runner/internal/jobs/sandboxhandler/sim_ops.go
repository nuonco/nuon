package sandboxhandler

import (
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

// simOp describes a fake op span the sandbox handler should emit under the
// per-step span so the dashboard span-tree UI renders sandbox-mode jobs with
// the same shape as real deploy jobs. Each op gets a proportional slice of
// the total execute-step duration and a list of canned log lines spread
// across that slice.
//
// op names intentionally mirror the real handler op.Tool callsites wrapped
// in Phase 3 (terraform.plan, helm.upgrade, kubernetes_manifest.apply, …) so
// dashboards filtering by nuon.tool / nuon.op see consistent values across
// real and sandbox executions.
type simOp struct {
	op       string
	fraction float64
	logs     []string
}

// simulatedOps returns the simulated op span layout for a given sandbox job
// type + operation. Returning nil signals "fall back to a single <tool>.exec
// span" (preserves prior behavior for job types we have not enumerated yet).
func simulatedOps(jobType models.AppRunnerJobType, jobOp models.AppRunnerJobOperationType) []simOp {
	switch jobType {
	case models.AppRunnerJobTypeTerraformDashDeploy,
		models.AppRunnerJobTypeRunnerDashTerraform,
		models.AppRunnerJobTypeSandboxDashTerraform,
		models.AppRunnerJobTypeSandboxDashTerraformDashPlan:
		switch jobOp {
		case models.AppRunnerJobOperationTypeCreateDashApplyDashPlan:
			return []simOp{
				{op: "plan", fraction: 1.0, logs: []string{
					"(simulated) initializing terraform",
					"(simulated) refreshing state",
					"(simulated) plan: 3 to add, 1 to change, 0 to destroy",
					"(simulated) writing tfplan",
				}},
			}
		case models.AppRunnerJobOperationTypeCreateDashTeardownDashPlan:
			return []simOp{
				{op: "destroy_plan", fraction: 1.0, logs: []string{
					"(simulated) refreshing state",
					"(simulated) plan: 0 to add, 0 to change, 5 to destroy",
				}},
			}
		case models.AppRunnerJobOperationTypeApplyDashPlan:
			return []simOp{
				{op: "apply_plan", fraction: 1.0, logs: []string{
					"(simulated) applying plan",
					"(simulated) creating resources",
					"(simulated) apply complete: 3 added, 1 changed, 0 destroyed",
				}},
			}
		}
		return []simOp{{op: "plan", fraction: 1.0, logs: []string{"(simulated) running terraform"}}}

	case models.AppRunnerJobTypeHelmDashChartDashDeploy,
		models.AppRunnerJobTypeRunnerDashHelm:
		return []simOp{
			{op: "upgrade_diff", fraction: 0.25, logs: []string{
				"(simulated) computing helm release diff",
				"(simulated) 4 resources will change",
			}},
			{op: "upgrade", fraction: 0.75, logs: []string{
				"(simulated) installing/upgrading helm release",
				"(simulated) waiting for resources to be ready",
				"(simulated) helm release upgraded",
			}},
		}

	case models.AppRunnerJobTypeKubernetesDashManifestDashDeploy:
		return []simOp{
			{op: "apply_dry_run", fraction: 0.25, logs: []string{
				"(simulated) running kubectl apply --dry-run",
				"(simulated) 7 resources will be applied",
			}},
			{op: "apply", fraction: 0.75, logs: []string{
				"(simulated) applying manifests",
				"(simulated) waiting for resources to converge",
				"(simulated) apply complete",
			}},
		}

	case models.AppRunnerJobTypePulumiDashDeploy,
		models.AppRunnerJobTypeSandboxDashPulumi:
		return []simOp{
			{op: "preview", fraction: 0.3, logs: []string{
				"(simulated) computing pulumi preview",
			}},
			{op: "up", fraction: 0.7, logs: []string{
				"(simulated) applying pulumi stack updates",
				"(simulated) stack update complete",
			}},
		}

	case models.AppRunnerJobTypeSandboxDashSyncDashSecrets:
		return []simOp{
			{op: "sync", fraction: 1.0, logs: []string{
				"(simulated) reading secret from cloud provider",
				"(simulated) writing secret to cluster",
			}},
		}

	case models.AppRunnerJobTypeDockerDashBuild,
		models.AppRunnerJobTypeContainerDashImageDashBuild:
		return []simOp{
			{op: "build", fraction: 0.7, logs: []string{
				"(simulated) building container image",
				"(simulated) layers cached",
			}},
			{op: "push", fraction: 0.3, logs: []string{
				"(simulated) pushing image to registry",
			}},
		}

	case models.AppRunnerJobTypeHelmDashChartDashBuild:
		return []simOp{
			{op: "package_chart", fraction: 1.0, logs: []string{
				"(simulated) packaging helm chart",
			}},
		}

	case models.AppRunnerJobTypeTerraformDashModuleDashBuild:
		return []simOp{
			{op: "provider_mirror", fraction: 1.0, logs: []string{
				"(simulated) generating terraform provider mirror",
			}},
		}

	case models.AppRunnerJobTypeKubernetesDashManifestDashBuild:
		return []simOp{
			{op: "build", fraction: 1.0, logs: []string{
				"(simulated) running kustomize build",
			}},
		}

	case models.AppRunnerJobTypeOciDashSync:
		return []simOp{
			{op: "copy", fraction: 1.0, logs: []string{
				"(simulated) copying oci artifact between registries",
			}},
		}
	}

	return nil
}
