package templates

func outputTemplates() []Template {
	return []Template{
		{
			Key:         "terraform-outputs",
			Description: "Terraform outputs with S3 bucket and IAM role ARNs",
			Category:    "outputs",
			JobTypes:    []string{"terraform-deploy", "sandbox-terraform"},
			Contents:    `{"bucket_arn":"arn:aws:s3:::nuon-app-data-prod","role_arn":"arn:aws:iam::123456789012:role/nuon-app-role-prod","bucket_name":"nuon-app-data-prod","region":"us-west-2"}`,
		},
		{
			Key:         "terraform-outputs-rds",
			Description: "Terraform outputs for RDS Postgres instance",
			Category:    "outputs",
			JobTypes:    []string{"terraform-deploy", "sandbox-terraform"},
			Contents:    `{"db_endpoint":"nuon-app-db-prod.abc123xyz.us-west-2.rds.amazonaws.com:5432","db_name":"appdb","db_instance_id":"nuon-app-db-prod","security_group_id":"sg-0rds456def789abc","db_arn":"arn:aws:rds:us-west-2:123456789012:db:nuon-app-db-prod"}`,
		},
		{
			Key:         "terraform-outputs-ecs",
			Description: "Terraform outputs for ECS Fargate cluster and service",
			Category:    "outputs",
			JobTypes:    []string{"terraform-deploy", "sandbox-terraform"},
			Contents:    `{"cluster_arn":"arn:aws:ecs:us-west-2:123456789012:cluster/nuon-app-prod","service_name":"nuon-app-svc","task_definition_arn":"arn:aws:ecs:us-west-2:123456789012:task-definition/nuon-app:1","target_group_arn":"arn:aws:elasticloadbalancing:us-west-2:123456789012:targetgroup/nuon-app-tg/abc123def456"}`,
		},
		{
			Key:         "terraform-outputs-lambda",
			Description: "Terraform outputs for Lambda function",
			Category:    "outputs",
			JobTypes:    []string{"terraform-deploy", "sandbox-terraform"},
			Contents:    `{"function_name":"nuon-app-handler","function_arn":"arn:aws:lambda:us-west-2:123456789012:function:nuon-app-handler","invoke_url":"https://abc123xyz.execute-api.us-west-2.amazonaws.com/prod","log_group":"/aws/lambda/nuon-app-handler"}`,
		},
		{
			Key:         "terraform-outputs-vpc",
			Description: "Terraform outputs for VPC with subnets and NAT gateway",
			Category:    "outputs",
			JobTypes:    []string{"terraform-deploy", "sandbox-terraform"},
			Contents:    `{"vpc_id":"vpc-0abc123def456","vpc_cidr":"10.0.0.0/16","public_subnet_ids":["subnet-0pub1aaa","subnet-0pub2bbb","subnet-0pub3ccc"],"private_subnet_ids":["subnet-0prv1ddd","subnet-0prv2eee","subnet-0prv3fff"],"nat_gateway_ip":"54.123.45.67"}`,
		},
		{
			Key:         "terraform-outputs-s3-detailed",
			Description: "Terraform outputs for S3 bucket with versioning and encryption",
			Category:    "outputs",
			JobTypes:    []string{"terraform-deploy", "sandbox-terraform"},
			Contents:    `{"bucket_arn":"arn:aws:s3:::nuon-app-data-prod","bucket_name":"nuon-app-data-prod","bucket_domain_name":"nuon-app-data-prod.s3.amazonaws.com","region":"us-west-2","versioning_enabled":true,"encryption_algorithm":"aws:kms"}`,
		},
		{
			Key:         "helm-outputs",
			Description: "Helm release outputs with release metadata",
			Category:    "outputs",
			JobTypes:    []string{"helm-chart-deploy"},
			Contents:    `{"release_name":"app-release","namespace":"default","revision":"1","status":"deployed","chart":"app-0.1.0"}`,
		},
		{
			Key:         "kube-outputs",
			Description: "Kubernetes manifest apply outputs",
			Category:    "outputs",
			JobTypes:    []string{"kubernetes-manifest-deploy"},
			Contents:    `{"resources_applied":5,"namespace":"default","deployment":"app-deployment","service":"app-service","status":"applied"}`,
		},
		{
			Key:         "pulumi-outputs",
			Description: "Pulumi stack outputs",
			Category:    "outputs",
			JobTypes:    []string{"pulumi-deploy"},
			Contents:    `{"bucketName":"app-data-a1b2c3d","cdnDomain":"d1234567.cloudfront.net","roleArn":"arn:aws:iam::123456789012:role/app-role-prod","stackName":"app-prod"}`,
		},
	}
}
