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
				{JobType: "terraform-deploy", LogTemplate: "terraform-noop-logs", PlanTemplate: "terraform-noop", PlanDisplayTemplate: "terraform-display-noop", StateTemplate: "terraform-state-empty", OutputTemplate: "terraform-outputs", DurationMs: 2000, Enabled: true},
				{JobType: "sandbox-terraform", LogTemplate: "terraform-noop-logs", PlanTemplate: "terraform-noop", PlanDisplayTemplate: "terraform-display-noop", StateTemplate: "terraform-state-empty", OutputTemplate: "terraform-outputs", DurationMs: 2000, Enabled: true},
				{JobType: "sandbox-terraform-plan", LogTemplate: "terraform-noop-logs", PlanTemplate: "terraform-noop", PlanDisplayTemplate: "terraform-display-noop", DurationMs: 1000, Enabled: true},
			},
		},
		{
			Key:         "terraform-apply",
			Name:        "Terraform apply",
			Description: "Terraform jobs with realistic resource creation plans and outputs",
			Configs: []FlowConfig{
				{JobType: "terraform-deploy", LogTemplate: "terraform-apply", PlanTemplate: "terraform-apply", PlanDisplayTemplate: "terraform-display-apply", StateTemplate: "terraform-state-s3-iam", OutputTemplate: "terraform-outputs", DurationMs: 5000, Enabled: true},
				{JobType: "sandbox-terraform", LogTemplate: "terraform-apply", PlanTemplate: "terraform-apply", PlanDisplayTemplate: "terraform-display-apply", StateTemplate: "terraform-state-s3-iam", OutputTemplate: "terraform-outputs", DurationMs: 5000, Enabled: true},
				{JobType: "sandbox-terraform-plan", LogTemplate: "terraform-plan", PlanTemplate: "terraform-apply", PlanDisplayTemplate: "terraform-display-apply", DurationMs: 2000, Enabled: true},
			},
		},
		{
			Key:         "terraform-destroy",
			Name:        "Terraform destroy",
			Description: "Terraform jobs with resource destruction plans",
			Configs: []FlowConfig{
				{JobType: "terraform-deploy", LogTemplate: "terraform-apply", PlanTemplate: "terraform-destroy", PlanDisplayTemplate: "terraform-display-destroy", DurationMs: 5000, Enabled: true},
				{JobType: "sandbox-terraform", LogTemplate: "terraform-apply", PlanTemplate: "terraform-destroy", PlanDisplayTemplate: "terraform-display-destroy", DurationMs: 5000, Enabled: true},
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

		// Resource-specific terraform flows
		{
			Key:         "terraform-rds",
			Name:        "Terraform RDS Postgres",
			Description: "Terraform jobs creating an RDS Postgres instance with supporting resources",
			Configs: []FlowConfig{
				{JobType: "terraform-deploy", LogTemplate: "terraform-apply-rds", PlanTemplate: "terraform-apply-rds", PlanDisplayTemplate: "terraform-display-apply-rds", StateTemplate: "terraform-state-rds", OutputTemplate: "terraform-outputs-rds", DurationMs: 12000, Enabled: true},
				{JobType: "sandbox-terraform", LogTemplate: "terraform-apply-rds", PlanTemplate: "terraform-apply-rds", PlanDisplayTemplate: "terraform-display-apply-rds", StateTemplate: "terraform-state-rds", OutputTemplate: "terraform-outputs-rds", DurationMs: 12000, Enabled: true},
				{JobType: "sandbox-terraform-plan", LogTemplate: "terraform-plan", PlanTemplate: "terraform-apply-rds", PlanDisplayTemplate: "terraform-display-apply-rds", DurationMs: 3000, Enabled: true},
			},
		},
		{
			Key:         "terraform-ecs",
			Name:        "Terraform ECS Fargate",
			Description: "Terraform jobs creating an ECS Fargate cluster with service and target group",
			Configs: []FlowConfig{
				{JobType: "terraform-deploy", LogTemplate: "terraform-apply-ecs", PlanTemplate: "terraform-apply-ecs", PlanDisplayTemplate: "terraform-display-apply-ecs", StateTemplate: "terraform-state-ecs", OutputTemplate: "terraform-outputs-ecs", DurationMs: 8000, Enabled: true},
				{JobType: "sandbox-terraform", LogTemplate: "terraform-apply-ecs", PlanTemplate: "terraform-apply-ecs", PlanDisplayTemplate: "terraform-display-apply-ecs", StateTemplate: "terraform-state-ecs", OutputTemplate: "terraform-outputs-ecs", DurationMs: 8000, Enabled: true},
				{JobType: "sandbox-terraform-plan", LogTemplate: "terraform-plan", PlanTemplate: "terraform-apply-ecs", PlanDisplayTemplate: "terraform-display-apply-ecs", DurationMs: 2000, Enabled: true},
			},
		},
		{
			Key:         "terraform-lambda",
			Name:        "Terraform Lambda",
			Description: "Terraform jobs creating a Lambda function with IAM role and CloudWatch logs",
			Configs: []FlowConfig{
				{JobType: "terraform-deploy", LogTemplate: "terraform-apply-lambda", PlanTemplate: "terraform-apply-lambda", PlanDisplayTemplate: "terraform-display-apply-lambda", StateTemplate: "terraform-state-lambda", OutputTemplate: "terraform-outputs-lambda", DurationMs: 5000, Enabled: true},
				{JobType: "sandbox-terraform", LogTemplate: "terraform-apply-lambda", PlanTemplate: "terraform-apply-lambda", PlanDisplayTemplate: "terraform-display-apply-lambda", StateTemplate: "terraform-state-lambda", OutputTemplate: "terraform-outputs-lambda", DurationMs: 5000, Enabled: true},
				{JobType: "sandbox-terraform-plan", LogTemplate: "terraform-plan", PlanTemplate: "terraform-apply-lambda", PlanDisplayTemplate: "terraform-display-apply-lambda", DurationMs: 2000, Enabled: true},
			},
		},
		{
			Key:         "terraform-vpc",
			Name:        "Terraform VPC",
			Description: "Terraform jobs creating a VPC with subnets, internet gateway, and NAT gateway",
			Configs: []FlowConfig{
				{JobType: "terraform-deploy", LogTemplate: "terraform-apply-vpc", PlanTemplate: "terraform-apply-vpc", PlanDisplayTemplate: "terraform-display-apply-vpc", StateTemplate: "terraform-state-vpc", OutputTemplate: "terraform-outputs-vpc", DurationMs: 10000, Enabled: true},
				{JobType: "sandbox-terraform", LogTemplate: "terraform-apply-vpc", PlanTemplate: "terraform-apply-vpc", PlanDisplayTemplate: "terraform-display-apply-vpc", StateTemplate: "terraform-state-vpc", OutputTemplate: "terraform-outputs-vpc", DurationMs: 10000, Enabled: true},
				{JobType: "sandbox-terraform-plan", LogTemplate: "terraform-plan", PlanTemplate: "terraform-apply-vpc", PlanDisplayTemplate: "terraform-display-apply-vpc", DurationMs: 2000, Enabled: true},
			},
		},
		{
			Key:         "terraform-s3-detailed",
			Name:        "Terraform S3 (detailed)",
			Description: "Terraform jobs creating an S3 bucket with versioning, encryption, and lifecycle",
			Configs: []FlowConfig{
				{JobType: "terraform-deploy", LogTemplate: "terraform-apply-s3-detailed", PlanTemplate: "terraform-apply-s3-detailed", PlanDisplayTemplate: "terraform-display-apply-s3-detailed", StateTemplate: "terraform-state-s3-detailed", OutputTemplate: "terraform-outputs-s3-detailed", DurationMs: 5000, Enabled: true},
				{JobType: "sandbox-terraform", LogTemplate: "terraform-apply-s3-detailed", PlanTemplate: "terraform-apply-s3-detailed", PlanDisplayTemplate: "terraform-display-apply-s3-detailed", StateTemplate: "terraform-state-s3-detailed", OutputTemplate: "terraform-outputs-s3-detailed", DurationMs: 5000, Enabled: true},
				{JobType: "sandbox-terraform-plan", LogTemplate: "terraform-plan", PlanTemplate: "terraform-apply-s3-detailed", PlanDisplayTemplate: "terraform-display-apply-s3-detailed", DurationMs: 2000, Enabled: true},
			},
		},
		{
			Key:         "full-infra-stack",
			Name:        "Full infrastructure stack",
			Description: "VPC + RDS + ECS combined for a complete infrastructure deployment",
			Configs: []FlowConfig{
				{JobType: "terraform-deploy", LogTemplate: "terraform-apply-vpc", PlanTemplate: "terraform-apply-vpc", PlanDisplayTemplate: "terraform-display-apply-vpc", StateTemplate: "terraform-state-vpc", OutputTemplate: "terraform-outputs-vpc", DurationMs: 10000, Enabled: true},
				{JobType: "sandbox-terraform", LogTemplate: "terraform-apply-rds", PlanTemplate: "terraform-apply-rds", PlanDisplayTemplate: "terraform-display-apply-rds", StateTemplate: "terraform-state-rds", OutputTemplate: "terraform-outputs-rds", DurationMs: 12000, Enabled: true},
				{JobType: "sandbox-terraform-plan", LogTemplate: "terraform-plan", PlanTemplate: "terraform-apply-ecs", PlanDisplayTemplate: "terraform-display-apply-ecs", DurationMs: 3000, Enabled: true},
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
