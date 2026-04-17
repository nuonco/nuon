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
