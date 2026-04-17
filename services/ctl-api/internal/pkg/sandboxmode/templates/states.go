package templates

// stateTemplates returns all terraform state JSON templates.
// These match the format of `terraform show -json` output (format_version "1.0").
func stateTemplates() []Template {
	return []Template{
		{
			Key:         "terraform-state-empty",
			Description: "Empty terraform state with no resources",
			Category:    "states",
			JobTypes:    []string{"terraform-deploy", "sandbox-terraform"},
			IsNoop:      true,
			Contents:    `{"format_version":"1.0","terraform_version":"1.7.5","values":{"outputs":{},"root_module":{"resources":[]}}}`,
		},
		{
			Key:         "terraform-state-s3-iam",
			Description: "Terraform state with S3 bucket and IAM role",
			Category:    "states",
			JobTypes:    []string{"terraform-deploy", "sandbox-terraform"},
			Contents: `{
  "format_version": "1.0",
  "terraform_version": "1.7.5",
  "values": {
    "outputs": {
      "bucket_arn": {"sensitive": false, "value": "arn:aws:s3:::nuon-app-data-prod", "type": "string"},
      "role_arn": {"sensitive": false, "value": "arn:aws:iam::123456789012:role/nuon-app-role-prod", "type": "string"}
    },
    "root_module": {
      "resources": [
        {
          "address": "aws_s3_bucket.main",
          "mode": "managed",
          "type": "aws_s3_bucket",
          "name": "main",
          "provider_name": "registry.terraform.io/hashicorp/aws",
          "schema_version": 0,
          "values": {
            "id": "nuon-app-data-prod",
            "arn": "arn:aws:s3:::nuon-app-data-prod",
            "bucket": "nuon-app-data-prod",
            "bucket_domain_name": "nuon-app-data-prod.s3.amazonaws.com",
            "bucket_regional_domain_name": "nuon-app-data-prod.s3.us-west-2.amazonaws.com",
            "region": "us-west-2",
            "force_destroy": false,
            "tags": {"Environment": "production", "ManagedBy": "nuon"}
          }
        },
        {
          "address": "aws_iam_role.app_role",
          "mode": "managed",
          "type": "aws_iam_role",
          "name": "app_role",
          "provider_name": "registry.terraform.io/hashicorp/aws",
          "schema_version": 0,
          "values": {
            "id": "nuon-app-role-prod",
            "arn": "arn:aws:iam::123456789012:role/nuon-app-role-prod",
            "name": "nuon-app-role-prod",
            "unique_id": "AROAEXAMPLEID12345",
            "create_date": "2024-03-15T12:00:00Z",
            "force_detach_policies": false,
            "tags": {"Environment": "production", "ManagedBy": "nuon"}
          }
        }
      ]
    }
  }
}`,
		},
		{
			Key:         "terraform-state-rds",
			Description: "Terraform state with RDS Postgres instance and supporting resources",
			Category:    "states",
			JobTypes:    []string{"terraform-deploy", "sandbox-terraform"},
			Contents: `{
  "format_version": "1.0",
  "terraform_version": "1.7.5",
  "values": {
    "outputs": {
      "db_endpoint": {"sensitive": false, "value": "nuon-app-db-prod.abc123xyz.us-west-2.rds.amazonaws.com:5432", "type": "string"},
      "db_name": {"sensitive": false, "value": "appdb", "type": "string"},
      "db_instance_id": {"sensitive": false, "value": "nuon-app-db-prod", "type": "string"},
      "security_group_id": {"sensitive": false, "value": "sg-0rds456def789abc", "type": "string"}
    },
    "root_module": {
      "resources": [
        {
          "address": "aws_db_instance.main",
          "mode": "managed",
          "type": "aws_db_instance",
          "name": "main",
          "provider_name": "registry.terraform.io/hashicorp/aws",
          "schema_version": 2,
          "values": {
            "id": "nuon-app-db-prod",
            "identifier": "nuon-app-db-prod",
            "arn": "arn:aws:rds:us-west-2:123456789012:db:nuon-app-db-prod",
            "endpoint": "nuon-app-db-prod.abc123xyz.us-west-2.rds.amazonaws.com:5432",
            "address": "nuon-app-db-prod.abc123xyz.us-west-2.rds.amazonaws.com",
            "port": 5432,
            "engine": "postgres",
            "engine_version": "15.4",
            "instance_class": "db.t3.medium",
            "allocated_storage": 20,
            "max_allocated_storage": 100,
            "storage_type": "gp3",
            "storage_encrypted": true,
            "db_name": "appdb",
            "username": "appuser",
            "multi_az": false,
            "publicly_accessible": false,
            "backup_retention_period": 7,
            "deletion_protection": true,
            "skip_final_snapshot": false,
            "availability_zone": "us-west-2a",
            "status": "available",
            "tags": {"Environment": "production", "ManagedBy": "nuon"}
          }
        },
        {
          "address": "aws_db_subnet_group.main",
          "mode": "managed",
          "type": "aws_db_subnet_group",
          "name": "main",
          "provider_name": "registry.terraform.io/hashicorp/aws",
          "schema_version": 0,
          "values": {
            "id": "nuon-app-db-subnet-group",
            "arn": "arn:aws:rds:us-west-2:123456789012:subgrp:nuon-app-db-subnet-group",
            "name": "nuon-app-db-subnet-group",
            "subnet_ids": ["subnet-0abc123", "subnet-0def456", "subnet-0ghi789"]
          }
        },
        {
          "address": "aws_db_parameter_group.main",
          "mode": "managed",
          "type": "aws_db_parameter_group",
          "name": "main",
          "provider_name": "registry.terraform.io/hashicorp/aws",
          "schema_version": 0,
          "values": {
            "id": "nuon-app-db-params",
            "arn": "arn:aws:rds:us-west-2:123456789012:pg:nuon-app-db-params",
            "name": "nuon-app-db-params",
            "family": "postgres15"
          }
        },
        {
          "address": "aws_security_group.rds",
          "mode": "managed",
          "type": "aws_security_group",
          "name": "rds",
          "provider_name": "registry.terraform.io/hashicorp/aws",
          "schema_version": 1,
          "values": {
            "id": "sg-0rds456def789abc",
            "arn": "arn:aws:ec2:us-west-2:123456789012:security-group/sg-0rds456def789abc",
            "name": "nuon-app-rds-sg",
            "vpc_id": "vpc-0abc123def456"
          }
        }
      ]
    }
  }
}`,
		},
		{
			Key:         "terraform-state-ecs",
			Description: "Terraform state with ECS Fargate cluster and service",
			Category:    "states",
			JobTypes:    []string{"terraform-deploy", "sandbox-terraform"},
			Contents: `{
  "format_version": "1.0",
  "terraform_version": "1.7.5",
  "values": {
    "outputs": {
      "cluster_arn": {"sensitive": false, "value": "arn:aws:ecs:us-west-2:123456789012:cluster/nuon-app-prod", "type": "string"},
      "service_name": {"sensitive": false, "value": "nuon-app-svc", "type": "string"},
      "task_definition_arn": {"sensitive": false, "value": "arn:aws:ecs:us-west-2:123456789012:task-definition/nuon-app:1", "type": "string"},
      "target_group_arn": {"sensitive": false, "value": "arn:aws:elasticloadbalancing:us-west-2:123456789012:targetgroup/nuon-app-tg/abc123def456", "type": "string"}
    },
    "root_module": {
      "resources": [
        {
          "address": "aws_ecs_cluster.main",
          "mode": "managed",
          "type": "aws_ecs_cluster",
          "name": "main",
          "provider_name": "registry.terraform.io/hashicorp/aws",
          "schema_version": 0,
          "values": {
            "id": "arn:aws:ecs:us-west-2:123456789012:cluster/nuon-app-prod",
            "arn": "arn:aws:ecs:us-west-2:123456789012:cluster/nuon-app-prod",
            "name": "nuon-app-prod",
            "setting": [{"name": "containerInsights", "value": "enabled"}],
            "tags": {"Environment": "production", "ManagedBy": "nuon"}
          }
        },
        {
          "address": "aws_ecs_task_definition.app",
          "mode": "managed",
          "type": "aws_ecs_task_definition",
          "name": "app",
          "provider_name": "registry.terraform.io/hashicorp/aws",
          "schema_version": 1,
          "values": {
            "id": "nuon-app",
            "arn": "arn:aws:ecs:us-west-2:123456789012:task-definition/nuon-app:1",
            "family": "nuon-app",
            "revision": 1,
            "cpu": "256",
            "memory": "512",
            "network_mode": "awsvpc",
            "requires_compatibilities": ["FARGATE"],
            "execution_role_arn": "arn:aws:iam::123456789012:role/ecsTaskExecutionRole"
          }
        },
        {
          "address": "aws_ecs_service.app",
          "mode": "managed",
          "type": "aws_ecs_service",
          "name": "app",
          "provider_name": "registry.terraform.io/hashicorp/aws",
          "schema_version": 0,
          "values": {
            "id": "arn:aws:ecs:us-west-2:123456789012:service/nuon-app-prod/nuon-app-svc",
            "name": "nuon-app-svc",
            "cluster": "arn:aws:ecs:us-west-2:123456789012:cluster/nuon-app-prod",
            "desired_count": 2,
            "launch_type": "FARGATE",
            "platform_version": "LATEST"
          }
        },
        {
          "address": "aws_lb_target_group.app",
          "mode": "managed",
          "type": "aws_lb_target_group",
          "name": "app",
          "provider_name": "registry.terraform.io/hashicorp/aws",
          "schema_version": 0,
          "values": {
            "id": "arn:aws:elasticloadbalancing:us-west-2:123456789012:targetgroup/nuon-app-tg/abc123def456",
            "arn": "arn:aws:elasticloadbalancing:us-west-2:123456789012:targetgroup/nuon-app-tg/abc123def456",
            "name": "nuon-app-tg",
            "port": 8080,
            "protocol": "HTTP",
            "target_type": "ip",
            "vpc_id": "vpc-0abc123def456"
          }
        }
      ]
    }
  }
}`,
		},
		{
			Key:         "terraform-state-lambda",
			Description: "Terraform state with Lambda function, IAM role, and CloudWatch log group",
			Category:    "states",
			JobTypes:    []string{"terraform-deploy", "sandbox-terraform"},
			Contents: `{
  "format_version": "1.0",
  "terraform_version": "1.7.5",
  "values": {
    "outputs": {
      "function_name": {"sensitive": false, "value": "nuon-app-handler", "type": "string"},
      "function_arn": {"sensitive": false, "value": "arn:aws:lambda:us-west-2:123456789012:function:nuon-app-handler", "type": "string"},
      "invoke_url": {"sensitive": false, "value": "https://abc123xyz.execute-api.us-west-2.amazonaws.com/prod", "type": "string"},
      "log_group": {"sensitive": false, "value": "/aws/lambda/nuon-app-handler", "type": "string"}
    },
    "root_module": {
      "resources": [
        {
          "address": "aws_lambda_function.main",
          "mode": "managed",
          "type": "aws_lambda_function",
          "name": "main",
          "provider_name": "registry.terraform.io/hashicorp/aws",
          "schema_version": 0,
          "values": {
            "id": "nuon-app-handler",
            "arn": "arn:aws:lambda:us-west-2:123456789012:function:nuon-app-handler",
            "function_name": "nuon-app-handler",
            "handler": "bootstrap",
            "runtime": "provided.al2023",
            "memory_size": 128,
            "timeout": 30,
            "architectures": ["arm64"],
            "source_code_hash": "abc123def456ghi789",
            "last_modified": "2024-03-15T12:00:00.000+0000",
            "tags": {"Environment": "production", "ManagedBy": "nuon"}
          }
        },
        {
          "address": "aws_iam_role.lambda_exec",
          "mode": "managed",
          "type": "aws_iam_role",
          "name": "lambda_exec",
          "provider_name": "registry.terraform.io/hashicorp/aws",
          "schema_version": 0,
          "values": {
            "id": "nuon-app-lambda-exec",
            "arn": "arn:aws:iam::123456789012:role/nuon-app-lambda-exec",
            "name": "nuon-app-lambda-exec",
            "create_date": "2024-03-15T12:00:00Z"
          }
        },
        {
          "address": "aws_cloudwatch_log_group.lambda",
          "mode": "managed",
          "type": "aws_cloudwatch_log_group",
          "name": "lambda",
          "provider_name": "registry.terraform.io/hashicorp/aws",
          "schema_version": 0,
          "values": {
            "id": "/aws/lambda/nuon-app-handler",
            "arn": "arn:aws:logs:us-west-2:123456789012:log-group:/aws/lambda/nuon-app-handler:*",
            "name": "/aws/lambda/nuon-app-handler",
            "retention_in_days": 14
          }
        }
      ]
    }
  }
}`,
		},
		{
			Key:         "terraform-state-vpc",
			Description: "Terraform state with VPC, subnets, internet gateway, and NAT gateway",
			Category:    "states",
			JobTypes:    []string{"terraform-deploy", "sandbox-terraform"},
			Contents: `{
  "format_version": "1.0",
  "terraform_version": "1.7.5",
  "values": {
    "outputs": {
      "vpc_id": {"sensitive": false, "value": "vpc-0abc123def456", "type": "string"},
      "vpc_cidr": {"sensitive": false, "value": "10.0.0.0/16", "type": "string"},
      "public_subnet_ids": {"sensitive": false, "value": ["subnet-0pub1aaa", "subnet-0pub2bbb", "subnet-0pub3ccc"], "type": ["list", "string"]},
      "private_subnet_ids": {"sensitive": false, "value": ["subnet-0prv1ddd", "subnet-0prv2eee", "subnet-0prv3fff"], "type": ["list", "string"]},
      "nat_gateway_ip": {"sensitive": false, "value": "54.123.45.67", "type": "string"}
    },
    "root_module": {
      "resources": [
        {
          "address": "aws_vpc.main",
          "mode": "managed",
          "type": "aws_vpc",
          "name": "main",
          "provider_name": "registry.terraform.io/hashicorp/aws",
          "schema_version": 1,
          "values": {
            "id": "vpc-0abc123def456",
            "arn": "arn:aws:ec2:us-west-2:123456789012:vpc/vpc-0abc123def456",
            "cidr_block": "10.0.0.0/16",
            "enable_dns_hostnames": true,
            "enable_dns_support": true,
            "tags": {"Name": "nuon-app-vpc", "Environment": "production", "ManagedBy": "nuon"}
          }
        },
        {
          "address": "aws_subnet.public[0]",
          "mode": "managed",
          "type": "aws_subnet",
          "name": "public",
          "index": 0,
          "provider_name": "registry.terraform.io/hashicorp/aws",
          "schema_version": 1,
          "values": {
            "id": "subnet-0pub1aaa",
            "arn": "arn:aws:ec2:us-west-2:123456789012:subnet/subnet-0pub1aaa",
            "cidr_block": "10.0.1.0/24",
            "availability_zone": "us-west-2a",
            "map_public_ip_on_launch": true,
            "vpc_id": "vpc-0abc123def456"
          }
        },
        {
          "address": "aws_subnet.private[0]",
          "mode": "managed",
          "type": "aws_subnet",
          "name": "private",
          "index": 0,
          "provider_name": "registry.terraform.io/hashicorp/aws",
          "schema_version": 1,
          "values": {
            "id": "subnet-0prv1ddd",
            "arn": "arn:aws:ec2:us-west-2:123456789012:subnet/subnet-0prv1ddd",
            "cidr_block": "10.0.10.0/24",
            "availability_zone": "us-west-2a",
            "map_public_ip_on_launch": false,
            "vpc_id": "vpc-0abc123def456"
          }
        },
        {
          "address": "aws_internet_gateway.main",
          "mode": "managed",
          "type": "aws_internet_gateway",
          "name": "main",
          "provider_name": "registry.terraform.io/hashicorp/aws",
          "schema_version": 0,
          "values": {
            "id": "igw-0abc123def456",
            "arn": "arn:aws:ec2:us-west-2:123456789012:internet-gateway/igw-0abc123def456",
            "vpc_id": "vpc-0abc123def456",
            "tags": {"Name": "nuon-app-igw"}
          }
        },
        {
          "address": "aws_nat_gateway.main",
          "mode": "managed",
          "type": "aws_nat_gateway",
          "name": "main",
          "provider_name": "registry.terraform.io/hashicorp/aws",
          "schema_version": 0,
          "values": {
            "id": "nat-0abc123def456",
            "allocation_id": "eipalloc-0abc123",
            "subnet_id": "subnet-0pub1aaa",
            "public_ip": "54.123.45.67",
            "connectivity_type": "public",
            "tags": {"Name": "nuon-app-nat"}
          }
        }
      ]
    }
  }
}`,
		},
		{
			Key:         "terraform-state-s3-detailed",
			Description: "Terraform state with S3 bucket, versioning, encryption, and lifecycle",
			Category:    "states",
			JobTypes:    []string{"terraform-deploy", "sandbox-terraform"},
			Contents: `{
  "format_version": "1.0",
  "terraform_version": "1.7.5",
  "values": {
    "outputs": {
      "bucket_arn": {"sensitive": false, "value": "arn:aws:s3:::nuon-app-data-prod", "type": "string"},
      "bucket_name": {"sensitive": false, "value": "nuon-app-data-prod", "type": "string"},
      "bucket_domain_name": {"sensitive": false, "value": "nuon-app-data-prod.s3.amazonaws.com", "type": "string"},
      "versioning_enabled": {"sensitive": false, "value": true, "type": "bool"},
      "encryption_algorithm": {"sensitive": false, "value": "aws:kms", "type": "string"}
    },
    "root_module": {
      "resources": [
        {
          "address": "aws_s3_bucket.main",
          "mode": "managed",
          "type": "aws_s3_bucket",
          "name": "main",
          "provider_name": "registry.terraform.io/hashicorp/aws",
          "schema_version": 0,
          "values": {
            "id": "nuon-app-data-prod",
            "arn": "arn:aws:s3:::nuon-app-data-prod",
            "bucket": "nuon-app-data-prod",
            "bucket_domain_name": "nuon-app-data-prod.s3.amazonaws.com",
            "region": "us-west-2",
            "tags": {"Environment": "production", "ManagedBy": "nuon"}
          }
        },
        {
          "address": "aws_s3_bucket_versioning.main",
          "mode": "managed",
          "type": "aws_s3_bucket_versioning",
          "name": "main",
          "provider_name": "registry.terraform.io/hashicorp/aws",
          "schema_version": 0,
          "values": {
            "id": "nuon-app-data-prod",
            "bucket": "nuon-app-data-prod",
            "versioning_configuration": [{"status": "Enabled", "mfa_delete": "Disabled"}]
          }
        },
        {
          "address": "aws_s3_bucket_server_side_encryption_configuration.main",
          "mode": "managed",
          "type": "aws_s3_bucket_server_side_encryption_configuration",
          "name": "main",
          "provider_name": "registry.terraform.io/hashicorp/aws",
          "schema_version": 0,
          "values": {
            "id": "nuon-app-data-prod",
            "bucket": "nuon-app-data-prod",
            "rule": [{"apply_server_side_encryption_by_default": [{"sse_algorithm": "aws:kms"}]}]
          }
        },
        {
          "address": "aws_s3_bucket_public_access_block.main",
          "mode": "managed",
          "type": "aws_s3_bucket_public_access_block",
          "name": "main",
          "provider_name": "registry.terraform.io/hashicorp/aws",
          "schema_version": 0,
          "values": {
            "id": "nuon-app-data-prod",
            "bucket": "nuon-app-data-prod",
            "block_public_acls": true,
            "block_public_policy": true,
            "ignore_public_acls": true,
            "restrict_public_buckets": true
          }
        },
        {
          "address": "aws_s3_bucket_lifecycle_configuration.main",
          "mode": "managed",
          "type": "aws_s3_bucket_lifecycle_configuration",
          "name": "main",
          "provider_name": "registry.terraform.io/hashicorp/aws",
          "schema_version": 0,
          "values": {
            "id": "nuon-app-data-prod",
            "bucket": "nuon-app-data-prod",
            "rule": [{"id": "archive-old-versions", "status": "Enabled", "noncurrent_version_transition": [{"noncurrent_days": 30, "storage_class": "GLACIER"}]}]
          }
        }
      ]
    }
  }
}`,
		},
	}
}
