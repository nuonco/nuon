package templates

func flowTemplates() []FlowTemplate {
	return []FlowTemplate{
		// Terraform flows
		{
			Key:         "terraform-noop",
			Name:        "Terraform noop",
			Description: "All terraform jobs return no-op plans with no changes",
			IsNoop:      true,
			Configs: []FlowConfig{
				{JobType: "terraform-deploy", LogTemplate: "terraform-noop-logs", PlanTemplate: "terraform-noop", OutputTemplate: "terraform-outputs", DurationMs: 2000, Enabled: true},
				{JobType: "sandbox-terraform", LogTemplate: "terraform-noop-logs", PlanTemplate: "terraform-noop", OutputTemplate: "terraform-outputs", DurationMs: 2000, Enabled: true},
				{JobType: "sandbox-terraform-plan", LogTemplate: "terraform-noop-logs", PlanTemplate: "terraform-noop", DurationMs: 1000, Enabled: true},
			},
		},
		{
			Key:         "terraform-apply",
			Name:        "Terraform apply",
			Description: "Terraform jobs with realistic resource creation plans and outputs",
			Configs: []FlowConfig{
				{JobType: "terraform-deploy", LogTemplate: "terraform-apply", PlanTemplate: "terraform-apply", OutputTemplate: "terraform-outputs", DurationMs: 5000, Enabled: true},
				{JobType: "sandbox-terraform", LogTemplate: "terraform-apply", PlanTemplate: "terraform-apply", OutputTemplate: "terraform-outputs", DurationMs: 5000, Enabled: true},
				{JobType: "sandbox-terraform-plan", LogTemplate: "terraform-plan", PlanTemplate: "terraform-apply", DurationMs: 2000, Enabled: true},
			},
		},
		{
			Key:         "terraform-destroy",
			Name:        "Terraform destroy",
			Description: "Terraform jobs with resource destruction plans",
			Configs: []FlowConfig{
				{JobType: "terraform-deploy", LogTemplate: "terraform-apply", PlanTemplate: "terraform-destroy", DurationMs: 5000, Enabled: true},
				{JobType: "sandbox-terraform", LogTemplate: "terraform-apply", PlanTemplate: "terraform-destroy", DurationMs: 5000, Enabled: true},
			},
		},

		// Helm flows
		{
			Key:         "helm-noop",
			Name:        "Helm noop",
			Description: "Helm deploy returns no diff, no changes applied",
			IsNoop:      true,
			Configs: []FlowConfig{
				{JobType: "helm-chart-deploy", LogTemplate: "helm-noop-logs", PlanTemplate: "helm-noop", OutputTemplate: "helm-outputs", DurationMs: 2000, Enabled: true},
			},
		},
		{
			Key:         "helm-install",
			Name:        "Helm install",
			Description: "Helm chart install with new resources",
			Configs: []FlowConfig{
				{JobType: "helm-chart-deploy", LogTemplate: "helm", PlanTemplate: "helm-install", OutputTemplate: "helm-outputs", DurationMs: 5000, Enabled: true},
			},
		},

		// Kubernetes flows
		{
			Key:         "kube-noop",
			Name:        "Kubernetes noop",
			Description: "Kubernetes manifest deploy with no diff",
			IsNoop:      true,
			Configs: []FlowConfig{
				{JobType: "kubernetes-manifest-deploy", LogTemplate: "kube-noop-logs", PlanTemplate: "kube-manifest-noop", OutputTemplate: "kube-outputs", DurationMs: 2000, Enabled: true},
			},
		},
		{
			Key:         "kube-apply",
			Name:        "Kubernetes apply",
			Description: "Kubernetes manifest deploy with resource changes",
			Configs: []FlowConfig{
				{JobType: "kubernetes-manifest-deploy", LogTemplate: "kubernetes-manifest", PlanTemplate: "kube-manifest-apply", OutputTemplate: "kube-outputs", DurationMs: 5000, Enabled: true},
			},
		},

		// Pulumi flows
		{
			Key:         "pulumi-noop",
			Name:        "Pulumi noop",
			Description: "Pulumi deploy with no changes in preview",
			IsNoop:      true,
			Configs: []FlowConfig{
				{JobType: "pulumi-deploy", LogTemplate: "pulumi-noop-logs", PlanTemplate: "pulumi-noop", OutputTemplate: "pulumi-outputs", DurationMs: 2000, Enabled: true},
			},
		},
		{
			Key:         "pulumi-up",
			Name:        "Pulumi up",
			Description: "Pulumi deploy creating new resources",
			Configs: []FlowConfig{
				{JobType: "pulumi-deploy", LogTemplate: "pulumi", PlanTemplate: "pulumi-up", OutputTemplate: "pulumi-outputs", DurationMs: 5000, Enabled: true},
			},
		},

		// Cross-type flows
		{
			Key:         "full-success-fast",
			Name:        "All success (fast)",
			Description: "All job types succeed quickly with realistic output",
			Configs: []FlowConfig{
				{JobType: "terraform-deploy", LogTemplate: "terraform-apply", PlanTemplate: "terraform-apply", OutputTemplate: "terraform-outputs", DurationMs: 1000, Enabled: true},
				{JobType: "helm-chart-deploy", LogTemplate: "helm", PlanTemplate: "helm-install", OutputTemplate: "helm-outputs", DurationMs: 1000, Enabled: true},
				{JobType: "kubernetes-manifest-deploy", LogTemplate: "kubernetes-manifest", PlanTemplate: "kube-manifest-apply", OutputTemplate: "kube-outputs", DurationMs: 1000, Enabled: true},
				{JobType: "sandbox-terraform", LogTemplate: "terraform-apply", PlanTemplate: "terraform-apply", OutputTemplate: "terraform-outputs", DurationMs: 1000, Enabled: true},
				{JobType: "sandbox-terraform-plan", LogTemplate: "terraform-plan", PlanTemplate: "terraform-apply", DurationMs: 500, Enabled: true},
				{JobType: "docker-build", LogTemplate: "docker-build", DurationMs: 2000, Enabled: true},
				{JobType: "oci-sync", LogTemplate: "oci-sync", DurationMs: 1000, Enabled: true},
			},
		},
		{
			Key:         "full-noop",
			Name:        "All noop",
			Description: "All job types return no-op / no changes",
			IsNoop:      true,
			Configs: []FlowConfig{
				{JobType: "terraform-deploy", LogTemplate: "terraform-noop-logs", PlanTemplate: "terraform-noop", OutputTemplate: "terraform-outputs", DurationMs: 1000, Enabled: true},
				{JobType: "helm-chart-deploy", LogTemplate: "helm-noop-logs", PlanTemplate: "helm-noop", OutputTemplate: "helm-outputs", DurationMs: 1000, Enabled: true},
				{JobType: "kubernetes-manifest-deploy", LogTemplate: "kube-noop-logs", PlanTemplate: "kube-manifest-noop", OutputTemplate: "kube-outputs", DurationMs: 1000, Enabled: true},
				{JobType: "sandbox-terraform", LogTemplate: "terraform-noop-logs", PlanTemplate: "terraform-noop", OutputTemplate: "terraform-outputs", DurationMs: 1000, Enabled: true},
				{JobType: "sandbox-terraform-plan", LogTemplate: "terraform-noop-logs", PlanTemplate: "terraform-noop", DurationMs: 500, Enabled: true},
			},
		},
	}
}
