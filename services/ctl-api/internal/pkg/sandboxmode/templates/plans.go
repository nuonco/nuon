package templates

// planTemplates returns all plan templates. Plan contents are machine-readable
// JSON that the noop checkers in pkg/plans/types/approval_plan/ can parse.
func planTemplates() []Template {
	return []Template{
		// Terraform plans
		{
			Key:         "terraform-apply",
			Description: "Terraform plan with resource changes (create S3 bucket and IAM role)",
			Category:    "plans",
			JobTypes:    []string{"terraform-deploy", "sandbox-terraform", "sandbox-terraform-plan"},
			Contents: `{
  "format_version": "1.2",
  "terraform_version": "1.7.5",
  "resource_changes": [
    {
      "address": "aws_s3_bucket.main",
      "mode": "managed",
      "type": "aws_s3_bucket",
      "name": "main",
      "provider_name": "registry.terraform.io/hashicorp/aws",
      "change": {
        "actions": ["create"],
        "before": null,
        "after": {
          "bucket": "nuon-app-data-prod",
          "force_destroy": false,
          "tags": {
            "Environment": "production",
            "ManagedBy": "nuon"
          }
        },
        "after_unknown": {
          "arn": true,
          "id": true,
          "bucket_domain_name": true,
          "region": true
        }
      }
    },
    {
      "address": "aws_iam_role.app_role",
      "mode": "managed",
      "type": "aws_iam_role",
      "name": "app_role",
      "provider_name": "registry.terraform.io/hashicorp/aws",
      "change": {
        "actions": ["create"],
        "before": null,
        "after": {
          "name": "nuon-app-role-prod",
          "force_detach_policies": false,
          "tags": {
            "Environment": "production",
            "ManagedBy": "nuon"
          }
        },
        "after_unknown": {
          "arn": true,
          "id": true,
          "unique_id": true
        }
      }
    }
  ],
  "output_changes": {
    "bucket_arn": {
      "actions": ["create"],
      "before": null,
      "after_unknown": true
    },
    "role_arn": {
      "actions": ["create"],
      "before": null,
      "after_unknown": true
    }
  }
}`,
		},
		{
			Key:         "terraform-noop",
			Description: "Terraform plan with no changes (passes IsNoop() == true)",
			Category:    "plans",
			JobTypes:    []string{"terraform-deploy", "sandbox-terraform", "sandbox-terraform-plan"},
			IsNoop:      true,
			Contents:    `{"format_version":"1.2","terraform_version":"1.7.5","resource_changes":[],"output_changes":{}}`,
		},
		{
			Key:         "terraform-destroy",
			Description: "Terraform plan with destroy actions",
			Category:    "plans",
			JobTypes:    []string{"terraform-deploy", "sandbox-terraform", "sandbox-terraform-plan"},
			Contents: `{
  "format_version": "1.2",
  "terraform_version": "1.7.5",
  "resource_changes": [
    {
      "address": "aws_s3_bucket.data",
      "mode": "managed",
      "type": "aws_s3_bucket",
      "name": "data",
      "provider_name": "registry.terraform.io/hashicorp/aws",
      "change": {
        "actions": ["delete"],
        "before": {
          "bucket": "nuon-app-data-prod",
          "id": "nuon-app-data-prod",
          "region": "us-west-2"
        },
        "after": null
      }
    },
    {
      "address": "aws_iam_role.app_role",
      "mode": "managed",
      "type": "aws_iam_role",
      "name": "app_role",
      "provider_name": "registry.terraform.io/hashicorp/aws",
      "change": {
        "actions": ["delete"],
        "before": {
          "name": "nuon-app-role",
          "id": "nuon-app-role"
        },
        "after": null
      }
    }
  ],
  "output_changes": {}
}`,
		},
		// Helm plans
		{
			Key:         "helm-install",
			Description: "Helm plan for new chart install with resource diffs",
			Category:    "plans",
			JobTypes:    []string{"helm-chart-deploy"},
			Contents: `{
  "plan": "install app-release",
  "op": "install",
  "helm_content_diff": [
    {
      "_version": "v1",
      "name": "app-deployment",
      "namespace": "default",
      "kind": "Deployment",
      "type": 2,
      "entries": [
        {
          "type": 2,
          "applied": {
            "apiVersion": "apps/v1",
            "kind": "Deployment",
            "metadata": {"name": "app-deployment", "namespace": "default"},
            "spec": {"replicas": 3}
          }
        }
      ]
    },
    {
      "_version": "v1",
      "name": "app-service",
      "namespace": "default",
      "kind": "Service",
      "type": 2,
      "entries": [
        {
          "type": 2,
          "applied": {
            "apiVersion": "v1",
            "kind": "Service",
            "metadata": {"name": "app-service", "namespace": "default"},
            "spec": {"type": "ClusterIP", "ports": [{"port": 80, "targetPort": 8080}]}
          }
        }
      ]
    }
  ]
}`,
		},
		{
			Key:         "helm-noop",
			Description: "Helm plan with no changes (passes IsNoop() == true)",
			Category:    "plans",
			JobTypes:    []string{"helm-chart-deploy"},
			IsNoop:      true,
			Contents:    `{"plan":"no changes","op":"upgrade","helm_content_diff":[]}`,
		},
		{
			Key:         "helm-upgrade",
			Description: "Helm plan for chart upgrade with changes",
			Category:    "plans",
			JobTypes:    []string{"helm-chart-deploy"},
			Contents: `{
  "plan": "upgrade app-release",
  "op": "upgrade",
  "helm_content_diff": [
    {
      "_version": "v1",
      "name": "app-deployment",
      "namespace": "default",
      "kind": "Deployment",
      "type": 3,
      "entries": [
        {
          "path": "spec.template.spec.containers.0.image",
          "type": 3,
          "original": "nuon/app:v1.4.0",
          "applied": "nuon/app:v1.5.0"
        },
        {
          "path": "spec.template.spec.containers.0.resources.limits.memory",
          "type": 3,
          "original": "256Mi",
          "applied": "512Mi"
        }
      ]
    },
    {
      "_version": "v1",
      "name": "app-config",
      "namespace": "default",
      "kind": "ConfigMap",
      "type": 3,
      "entries": [
        {
          "path": "data.LOG_LEVEL",
          "type": 3,
          "original": "info",
          "applied": "debug"
        },
        {
          "path": "data.FEATURE_FLAG_V2",
          "type": 2,
          "applied": "true"
        }
      ]
    }
  ]
}`,
		},
		// Kubernetes manifest plans
		{
			Key:         "kube-manifest-apply",
			Description: "Kubernetes manifest plan with new resources",
			Category:    "plans",
			JobTypes:    []string{"kubernetes-manifest-deploy"},
			Contents: `{
  "plan": "apply 3 resources",
  "op": "apply",
  "k8s_content_diff": [
    {
      "_version": "v1",
      "name": "app-deployment",
      "namespace": "default",
      "kind": "Deployment",
      "type": 2,
      "entries": [
        {
          "type": 2,
          "applied": {
            "apiVersion": "apps/v1",
            "kind": "Deployment",
            "metadata": {"name": "app-deployment", "namespace": "default"},
            "spec": {"replicas": 3}
          }
        }
      ]
    },
    {
      "_version": "v1",
      "name": "app-service",
      "namespace": "default",
      "kind": "Service",
      "type": 2,
      "entries": [
        {
          "type": 2,
          "applied": {
            "apiVersion": "v1",
            "kind": "Service",
            "metadata": {"name": "app-service", "namespace": "default"},
            "spec": {"type": "ClusterIP", "ports": [{"port": 80, "targetPort": 8080}]}
          }
        }
      ]
    }
  ]
}`,
		},
		{
			Key:         "kube-kustomize-apply",
			Description: "Kustomize-based Kubernetes manifest plan with resources",
			Category:    "plans",
			JobTypes:    []string{"kubernetes-manifest-deploy"},
			Contents: `{
  "plan": "apply kustomize overlay",
  "op": "apply",
  "k8s_content_diff": [
    {
      "_version": "v1",
      "name": "app-deployment",
      "namespace": "production",
      "kind": "Deployment",
      "type": 2,
      "entries": [
        {
          "type": 2,
          "applied": {
            "apiVersion": "apps/v1",
            "kind": "Deployment",
            "metadata": {"name": "app-deployment", "namespace": "production", "labels": {"app.kubernetes.io/managed-by": "kustomize"}},
            "spec": {"replicas": 5}
          }
        }
      ]
    },
    {
      "_version": "v1",
      "name": "app-config",
      "namespace": "production",
      "kind": "ConfigMap",
      "type": 2,
      "entries": [
        {
          "type": 2,
          "applied": {
            "apiVersion": "v1",
            "kind": "ConfigMap",
            "metadata": {"name": "app-config", "namespace": "production"},
            "data": {"LOG_LEVEL": "info", "API_PORT": "8080"}
          }
        }
      ]
    },
    {
      "_version": "v1",
      "name": "app-hpa",
      "namespace": "production",
      "kind": "HorizontalPodAutoscaler",
      "type": 2,
      "entries": [
        {
          "type": 2,
          "applied": {
            "apiVersion": "autoscaling/v2",
            "kind": "HorizontalPodAutoscaler",
            "metadata": {"name": "app-hpa", "namespace": "production"},
            "spec": {"minReplicas": 3, "maxReplicas": 10}
          }
        }
      ]
    }
  ]
}`,
		},
		{
			Key:         "kube-manifest-noop",
			Description: "Kubernetes manifest plan with no changes (passes IsNoop() == true)",
			Category:    "plans",
			JobTypes:    []string{"kubernetes-manifest-deploy"},
			IsNoop:      true,
			Contents:    `{"plan":"no changes","op":"apply","k8s_content_diff":[]}`,
		},
		{
			Key:         "kube-manifest-delete",
			Description: "Kubernetes manifest plan with resource deletions",
			Category:    "plans",
			JobTypes:    []string{"kubernetes-manifest-deploy"},
			Contents: `{
  "plan": "delete 3 resources",
  "op": "delete",
  "k8s_content_diff": [
    {
      "_version": "v1",
      "name": "app-deployment",
      "namespace": "default",
      "kind": "Deployment",
      "type": 1,
      "entries": [
        {
          "type": 1,
          "original": {
            "apiVersion": "apps/v1",
            "kind": "Deployment",
            "metadata": {"name": "app-deployment", "namespace": "default"}
          }
        }
      ]
    },
    {
      "_version": "v1",
      "name": "app-service",
      "namespace": "default",
      "kind": "Service",
      "type": 1,
      "entries": [
        {
          "type": 1,
          "original": {
            "apiVersion": "v1",
            "kind": "Service",
            "metadata": {"name": "app-service", "namespace": "default"}
          }
        }
      ]
    }
  ]
}`,
		},
		// Pulumi plans
		{
			Key:         "pulumi-up",
			Description: "Pulumi plan creating resources",
			Category:    "plans",
			JobTypes:    []string{"pulumi-deploy"},
			Contents:    `{"stdout":"Updating (prod)\n\n     Type                          Name              Status\n +   pulumi:pulumi:Stack            app-prod          created\n +   ├─ aws:s3:Bucket               app-data          created\n +   ├─ aws:iam:Role                app-role          created\n +   └─ aws:cloudfront:Distribution app-cdn           created\n\nResources:\n    + 4 created\n\nDuration: 45s","change_summary":{"create":4}}`,
		},
		{
			Key:         "pulumi-noop",
			Description: "Pulumi plan with no changes (passes IsNoop() == true)",
			Category:    "plans",
			JobTypes:    []string{"pulumi-deploy"},
			IsNoop:      true,
			Contents:    `{"stdout":"Updating (prod)\n\n     Type                 Name          Status\n     pulumi:pulumi:Stack  app-prod\n\nResources:\n    5 unchanged\n\nDuration: 3s","change_summary":{"same":5}}`,
		},
		{
			Key:         "pulumi-destroy",
			Description: "Pulumi plan destroying resources",
			Category:    "plans",
			JobTypes:    []string{"pulumi-deploy"},
			Contents:    `{"stdout":"Destroying (prod)\n\n     Type                          Name              Status\n -   pulumi:pulumi:Stack            app-prod          deleted\n -   ├─ aws:s3:Bucket               app-data          deleted\n -   ├─ aws:iam:Role                app-role          deleted\n -   └─ aws:cloudfront:Distribution app-cdn           deleted\n\nResources:\n    - 4 deleted\n\nDuration: 30s","change_summary":{"delete":4}}`,
		},
	}
}
