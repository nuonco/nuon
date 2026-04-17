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
		// Terraform resource-specific plans
		{
			Key:         "terraform-apply-rds",
			Description: "Terraform plan creating RDS Postgres instance with subnet group, parameter group, and security group",
			Category:    "plans",
			JobTypes:    []string{"terraform-deploy", "sandbox-terraform", "sandbox-terraform-plan"},
			Contents: `{
  "format_version": "1.2",
  "terraform_version": "1.7.5",
  "resource_changes": [
    {
      "address": "aws_db_subnet_group.main",
      "mode": "managed",
      "type": "aws_db_subnet_group",
      "name": "main",
      "provider_name": "registry.terraform.io/hashicorp/aws",
      "change": {
        "actions": ["create"],
        "before": null,
        "after": {
          "name": "nuon-app-db-subnet-group",
          "subnet_ids": ["subnet-0abc123", "subnet-0def456", "subnet-0ghi789"],
          "tags": {"Environment": "production", "ManagedBy": "nuon"}
        },
        "after_unknown": {"arn": true, "id": true}
      }
    },
    {
      "address": "aws_db_parameter_group.main",
      "mode": "managed",
      "type": "aws_db_parameter_group",
      "name": "main",
      "provider_name": "registry.terraform.io/hashicorp/aws",
      "change": {
        "actions": ["create"],
        "before": null,
        "after": {
          "name": "nuon-app-db-params",
          "family": "postgres15"
        },
        "after_unknown": {"arn": true, "id": true}
      }
    },
    {
      "address": "aws_security_group.rds",
      "mode": "managed",
      "type": "aws_security_group",
      "name": "rds",
      "provider_name": "registry.terraform.io/hashicorp/aws",
      "change": {
        "actions": ["create"],
        "before": null,
        "after": {
          "name": "nuon-app-rds-sg",
          "vpc_id": "vpc-0abc123def456",
          "ingress": [{"from_port": 5432, "to_port": 5432, "protocol": "tcp", "cidr_blocks": ["10.0.0.0/16"]}],
          "tags": {"Environment": "production", "ManagedBy": "nuon"}
        },
        "after_unknown": {"arn": true, "id": true}
      }
    },
    {
      "address": "aws_db_instance.main",
      "mode": "managed",
      "type": "aws_db_instance",
      "name": "main",
      "provider_name": "registry.terraform.io/hashicorp/aws",
      "change": {
        "actions": ["create"],
        "before": null,
        "after": {
          "identifier": "nuon-app-db-prod",
          "engine": "postgres",
          "engine_version": "15.4",
          "instance_class": "db.t3.medium",
          "allocated_storage": 20,
          "max_allocated_storage": 100,
          "storage_type": "gp3",
          "storage_encrypted": true,
          "db_name": "appdb",
          "username": "appuser",
          "port": 5432,
          "multi_az": false,
          "publicly_accessible": false,
          "backup_retention_period": 7,
          "deletion_protection": true,
          "skip_final_snapshot": false,
          "tags": {"Environment": "production", "ManagedBy": "nuon"}
        },
        "after_unknown": {"arn": true, "id": true, "endpoint": true, "address": true}
      }
    }
  ],
  "output_changes": {
    "db_endpoint": {"actions": ["create"], "before": null, "after_unknown": true},
    "db_name": {"actions": ["create"], "before": null, "after": "appdb"},
    "db_instance_id": {"actions": ["create"], "before": null, "after_unknown": true},
    "security_group_id": {"actions": ["create"], "before": null, "after_unknown": true}
  }
}`,
		},
		{
			Key:         "terraform-apply-ecs",
			Description: "Terraform plan creating ECS Fargate cluster, task definition, service, and target group",
			Category:    "plans",
			JobTypes:    []string{"terraform-deploy", "sandbox-terraform", "sandbox-terraform-plan"},
			Contents: `{
  "format_version": "1.2",
  "terraform_version": "1.7.5",
  "resource_changes": [
    {
      "address": "aws_ecs_cluster.main",
      "mode": "managed",
      "type": "aws_ecs_cluster",
      "name": "main",
      "provider_name": "registry.terraform.io/hashicorp/aws",
      "change": {
        "actions": ["create"],
        "before": null,
        "after": {
          "name": "nuon-app-prod",
          "setting": [{"name": "containerInsights", "value": "enabled"}],
          "tags": {"Environment": "production", "ManagedBy": "nuon"}
        },
        "after_unknown": {"arn": true, "id": true}
      }
    },
    {
      "address": "aws_ecs_task_definition.app",
      "mode": "managed",
      "type": "aws_ecs_task_definition",
      "name": "app",
      "provider_name": "registry.terraform.io/hashicorp/aws",
      "change": {
        "actions": ["create"],
        "before": null,
        "after": {
          "family": "nuon-app",
          "cpu": "256",
          "memory": "512",
          "network_mode": "awsvpc",
          "requires_compatibilities": ["FARGATE"],
          "execution_role_arn": "arn:aws:iam::123456789012:role/ecsTaskExecutionRole"
        },
        "after_unknown": {"arn": true, "id": true, "revision": true}
      }
    },
    {
      "address": "aws_ecs_service.app",
      "mode": "managed",
      "type": "aws_ecs_service",
      "name": "app",
      "provider_name": "registry.terraform.io/hashicorp/aws",
      "change": {
        "actions": ["create"],
        "before": null,
        "after": {
          "name": "nuon-app-svc",
          "desired_count": 2,
          "launch_type": "FARGATE",
          "platform_version": "LATEST"
        },
        "after_unknown": {"id": true, "cluster": true}
      }
    },
    {
      "address": "aws_lb_target_group.app",
      "mode": "managed",
      "type": "aws_lb_target_group",
      "name": "app",
      "provider_name": "registry.terraform.io/hashicorp/aws",
      "change": {
        "actions": ["create"],
        "before": null,
        "after": {
          "name": "nuon-app-tg",
          "port": 8080,
          "protocol": "HTTP",
          "target_type": "ip",
          "vpc_id": "vpc-0abc123def456"
        },
        "after_unknown": {"arn": true, "id": true}
      }
    }
  ],
  "output_changes": {
    "cluster_arn": {"actions": ["create"], "before": null, "after_unknown": true},
    "service_name": {"actions": ["create"], "before": null, "after": "nuon-app-svc"},
    "task_definition_arn": {"actions": ["create"], "before": null, "after_unknown": true},
    "target_group_arn": {"actions": ["create"], "before": null, "after_unknown": true}
  }
}`,
		},
		{
			Key:         "terraform-apply-lambda",
			Description: "Terraform plan creating Lambda function, IAM role, and CloudWatch log group",
			Category:    "plans",
			JobTypes:    []string{"terraform-deploy", "sandbox-terraform", "sandbox-terraform-plan"},
			Contents: `{
  "format_version": "1.2",
  "terraform_version": "1.7.5",
  "resource_changes": [
    {
      "address": "aws_iam_role.lambda_exec",
      "mode": "managed",
      "type": "aws_iam_role",
      "name": "lambda_exec",
      "provider_name": "registry.terraform.io/hashicorp/aws",
      "change": {
        "actions": ["create"],
        "before": null,
        "after": {
          "name": "nuon-app-lambda-exec",
          "force_detach_policies": false
        },
        "after_unknown": {"arn": true, "id": true, "unique_id": true, "create_date": true}
      }
    },
    {
      "address": "aws_cloudwatch_log_group.lambda",
      "mode": "managed",
      "type": "aws_cloudwatch_log_group",
      "name": "lambda",
      "provider_name": "registry.terraform.io/hashicorp/aws",
      "change": {
        "actions": ["create"],
        "before": null,
        "after": {
          "name": "/aws/lambda/nuon-app-handler",
          "retention_in_days": 14
        },
        "after_unknown": {"arn": true, "id": true}
      }
    },
    {
      "address": "aws_lambda_function.main",
      "mode": "managed",
      "type": "aws_lambda_function",
      "name": "main",
      "provider_name": "registry.terraform.io/hashicorp/aws",
      "change": {
        "actions": ["create"],
        "before": null,
        "after": {
          "function_name": "nuon-app-handler",
          "handler": "bootstrap",
          "runtime": "provided.al2023",
          "memory_size": 128,
          "timeout": 30,
          "architectures": ["arm64"],
          "tags": {"Environment": "production", "ManagedBy": "nuon"}
        },
        "after_unknown": {"arn": true, "id": true, "invoke_arn": true, "source_code_hash": true, "last_modified": true}
      }
    }
  ],
  "output_changes": {
    "function_name": {"actions": ["create"], "before": null, "after": "nuon-app-handler"},
    "function_arn": {"actions": ["create"], "before": null, "after_unknown": true},
    "invoke_url": {"actions": ["create"], "before": null, "after_unknown": true},
    "log_group": {"actions": ["create"], "before": null, "after": "/aws/lambda/nuon-app-handler"}
  }
}`,
		},
		{
			Key:         "terraform-apply-vpc",
			Description: "Terraform plan creating VPC with public/private subnets, internet gateway, and NAT gateway",
			Category:    "plans",
			JobTypes:    []string{"terraform-deploy", "sandbox-terraform", "sandbox-terraform-plan"},
			Contents: `{
  "format_version": "1.2",
  "terraform_version": "1.7.5",
  "resource_changes": [
    {
      "address": "aws_vpc.main",
      "mode": "managed",
      "type": "aws_vpc",
      "name": "main",
      "provider_name": "registry.terraform.io/hashicorp/aws",
      "change": {
        "actions": ["create"],
        "before": null,
        "after": {
          "cidr_block": "10.0.0.0/16",
          "enable_dns_hostnames": true,
          "enable_dns_support": true,
          "tags": {"Name": "nuon-app-vpc", "Environment": "production", "ManagedBy": "nuon"}
        },
        "after_unknown": {"arn": true, "id": true, "default_security_group_id": true, "default_route_table_id": true}
      }
    },
    {
      "address": "aws_subnet.public[0]",
      "mode": "managed",
      "type": "aws_subnet",
      "name": "public",
      "index": 0,
      "provider_name": "registry.terraform.io/hashicorp/aws",
      "change": {
        "actions": ["create"],
        "before": null,
        "after": {"cidr_block": "10.0.1.0/24", "availability_zone": "us-west-2a", "map_public_ip_on_launch": true},
        "after_unknown": {"arn": true, "id": true, "vpc_id": true}
      }
    },
    {
      "address": "aws_subnet.private[0]",
      "mode": "managed",
      "type": "aws_subnet",
      "name": "private",
      "index": 0,
      "provider_name": "registry.terraform.io/hashicorp/aws",
      "change": {
        "actions": ["create"],
        "before": null,
        "after": {"cidr_block": "10.0.10.0/24", "availability_zone": "us-west-2a", "map_public_ip_on_launch": false},
        "after_unknown": {"arn": true, "id": true, "vpc_id": true}
      }
    },
    {
      "address": "aws_internet_gateway.main",
      "mode": "managed",
      "type": "aws_internet_gateway",
      "name": "main",
      "provider_name": "registry.terraform.io/hashicorp/aws",
      "change": {
        "actions": ["create"],
        "before": null,
        "after": {"tags": {"Name": "nuon-app-igw"}},
        "after_unknown": {"arn": true, "id": true, "vpc_id": true}
      }
    },
    {
      "address": "aws_eip.nat",
      "mode": "managed",
      "type": "aws_eip",
      "name": "nat",
      "provider_name": "registry.terraform.io/hashicorp/aws",
      "change": {
        "actions": ["create"],
        "before": null,
        "after": {"domain": "vpc"},
        "after_unknown": {"allocation_id": true, "id": true, "public_ip": true}
      }
    },
    {
      "address": "aws_nat_gateway.main",
      "mode": "managed",
      "type": "aws_nat_gateway",
      "name": "main",
      "provider_name": "registry.terraform.io/hashicorp/aws",
      "change": {
        "actions": ["create"],
        "before": null,
        "after": {"connectivity_type": "public", "tags": {"Name": "nuon-app-nat"}},
        "after_unknown": {"allocation_id": true, "id": true, "public_ip": true, "subnet_id": true}
      }
    }
  ],
  "output_changes": {
    "vpc_id": {"actions": ["create"], "before": null, "after_unknown": true},
    "vpc_cidr": {"actions": ["create"], "before": null, "after": "10.0.0.0/16"},
    "public_subnet_ids": {"actions": ["create"], "before": null, "after_unknown": true},
    "private_subnet_ids": {"actions": ["create"], "before": null, "after_unknown": true},
    "nat_gateway_ip": {"actions": ["create"], "before": null, "after_unknown": true}
  }
}`,
		},
		{
			Key:         "terraform-apply-s3-detailed",
			Description: "Terraform plan creating S3 bucket with versioning, encryption, public access block, and lifecycle",
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
          "tags": {"Environment": "production", "ManagedBy": "nuon"}
        },
        "after_unknown": {"arn": true, "id": true, "bucket_domain_name": true, "region": true}
      }
    },
    {
      "address": "aws_s3_bucket_versioning.main",
      "mode": "managed",
      "type": "aws_s3_bucket_versioning",
      "name": "main",
      "provider_name": "registry.terraform.io/hashicorp/aws",
      "change": {
        "actions": ["create"],
        "before": null,
        "after": {"versioning_configuration": [{"status": "Enabled", "mfa_delete": "Disabled"}]},
        "after_unknown": {"bucket": true, "id": true}
      }
    },
    {
      "address": "aws_s3_bucket_server_side_encryption_configuration.main",
      "mode": "managed",
      "type": "aws_s3_bucket_server_side_encryption_configuration",
      "name": "main",
      "provider_name": "registry.terraform.io/hashicorp/aws",
      "change": {
        "actions": ["create"],
        "before": null,
        "after": {"rule": [{"apply_server_side_encryption_by_default": [{"sse_algorithm": "aws:kms"}]}]},
        "after_unknown": {"bucket": true, "id": true}
      }
    },
    {
      "address": "aws_s3_bucket_public_access_block.main",
      "mode": "managed",
      "type": "aws_s3_bucket_public_access_block",
      "name": "main",
      "provider_name": "registry.terraform.io/hashicorp/aws",
      "change": {
        "actions": ["create"],
        "before": null,
        "after": {"block_public_acls": true, "block_public_policy": true, "ignore_public_acls": true, "restrict_public_buckets": true},
        "after_unknown": {"bucket": true, "id": true}
      }
    },
    {
      "address": "aws_s3_bucket_lifecycle_configuration.main",
      "mode": "managed",
      "type": "aws_s3_bucket_lifecycle_configuration",
      "name": "main",
      "provider_name": "registry.terraform.io/hashicorp/aws",
      "change": {
        "actions": ["create"],
        "before": null,
        "after": {"rule": [{"id": "archive-old-versions", "status": "Enabled", "noncurrent_version_transition": [{"noncurrent_days": 30, "storage_class": "GLACIER"}]}]},
        "after_unknown": {"bucket": true, "id": true}
      }
    }
  ],
  "output_changes": {
    "bucket_arn": {"actions": ["create"], "before": null, "after_unknown": true},
    "bucket_name": {"actions": ["create"], "before": null, "after": "nuon-app-data-prod"},
    "bucket_domain_name": {"actions": ["create"], "before": null, "after_unknown": true},
    "versioning_enabled": {"actions": ["create"], "before": null, "after": true},
    "encryption_algorithm": {"actions": ["create"], "before": null, "after": "aws:kms"}
  }
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
