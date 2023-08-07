// nuon allows you to connect any helm chart in a connected or public repo to install in your application
resource "nuon_helm_chart_component" "e2e" {
  name = "e2e-helm"
  app_id = nuon_app.main.id

  chart_name = "e2e-helm"
  connected_repo = {
    directory = "services/e2e/chart"
    repo = data.nuon_connected_repo.mono.name
    branch = data.nuon_connected_repo.mono.default_branch
  }

  // dynamically set env vars from another source
  dynamic "value" {
    for_each = local.helm_values
    iterator = ev
    content {
      name = ev.key
      value = ev.value
    }
  }
}

locals {
  helm_values = {
    "env.DEFAULT_VALUE" = "set-by-terraform-provider-as-default"

    // nuon built ins
    "env.NUON_APP_ID" = "{{.nuon.app.id}}"
    "env.NUON_ORG_ID" = "{{.nuon.org.id}}"
    "env.NUON_INSTALL_ID" = "{{.nuon.install.id}}"

    // image component outputs
    "env.EXTERNAL_IMAGE_TAG" = "{{.nuon.components.e2e_external_image.image.tag}}"
    "env.EXTERNAL_IMAGE_REPOSITORY_ARN" = "{{.nuon.components.e2e_external_image.image.repository.arn}}"
    "env.EXTERNAL_IMAGE_REPOSITORY_NAME" = "{{.nuon.components.e2e_external_image.image.repository.name}}"
    "env.EXTERNAL_IMAGE_REPOSITORY_URI" = "{{.nuon.components.e2e_external_image.image.repository.uri}}"
    "env.EXTERNAL_IMAGE_REGISTRY_ID" = "{{.nuon.components.e2e_external_image.image.registry.id}}"

    // docker build component outputs
    "env.DOCKER_BUILD_TAG" = "{{.nuon.components.e2e_docker_build.image.tag}}"
    "env.DOCKER_BUILD_REPOSITORY_ARN" = "{{.nuon.components.e2e_docker_build.image.repository.arn}}"
    "env.DOCKER_BUILD_REPOSITORY_NAME" = "{{.nuon.components.e2e_docker_build.image.repository.name}}"
    "env.DOCKER_BUILD_REPOSITORY_URI" = "{{.nuon.components.e2e_docker_build.image.repository.uri}}"
    "env.DOCKER_BUILD_REGISTRY_ID" = "{{.nuon.components.e2e_docker_build.image.registry.id}}"

    // terraform component outputs
    "env.TERRAFORM_REPO_NAME" = "{{.nuon.components.e2e-infra.outputs.repo_name}}"
    "env.TERRAFORM_BUCKET_NAME" = "{{.nuon.components.e2e-infra.outputs.bucket_name}}"

    // sandbox outputs
    "env.SANDBOX_TYPE" = "{{.nuon.install.sandbox.type}}"
    "env.SANDBOX_VERSION" = "sandbox-{{.nuon.install.sandbox.version}}"
    "env.SANDBOX_PUBLIC_DOMAIN" = "{{.nuon.install.public_domain}}"
    "env.SANDBOX_INTERNAL_DOMAIN" = "{{.nuon.install.internal_domain}}"
    // sandbox runner outputs
    "env.SANDBOX_OUTPUT_RUNNER_DEFAULT_IAM_ROLE_ARN" = "{{.nuon.install.sandbox.outputs.runner.default_iam_role_arn}}"
    // sandbox cluster outputs
    "env.SANDBOX_OUTPUT_CLUSTER_ARN" = "{{.nuon.install.sandbox.outputs.cluster.arn}}"
    "env.SANDBOX_OUTPUT_CLUSTER_CERTIFICATE_AUTHORITY_DATA" = "{{.nuon.install.sandbox.outputs.cluster.certificate_authority_data}}"
    "env.SANDBOX_OUTPUT_CLUSTER_ENDPOINT" = "{{.nuon.install.sandbox.outputs.cluster.endpoint}}"
    "env.SANDBOX_OUTPUT_CLUSTER_NAME" = "{{.nuon.install.sandbox.outputs.cluster.name}}"
    "env.SANDBOX_OUTPUT_CLUSTERPLATFORM_VERSION" = "{{.nuon.install.sandbox.outputs.cluster.platform_version}}"
    "env.SANDBOX_OUTPUT_CLUSTER_STATUS" = "{{.nuon.install.sandbox.outputs.cluster.status}}"
    // sandbox vpc outputs
    "env.SANDBOX_OUTPUT_VPC_NAME" = "{{.nuon.install.sandbox.outputs.vpc.name}}"
    "env.SANDBOX_OUTPUT_VPC_ID" = "{{.nuon.install.sandbox.outputs.vpc.id}}"
    "env.SANDBOX_OUTPUT_VPC_CIDR" = "{{.nuon.install.sandbox.outputs.vpc.cidr}}"
    "env.SANDBOX_OUTPUT_VPC_AZS" = "{{.nuon.install.sandbox.outputs.vpc.azs}}"
    "env.SANDBOX_OUTPUT_VPC_PRIVATE_SUBNET_CIDR_BLOCKS" = "{{.nuon.install.sandbox.outputs.vpc.private_subnet_cidr_blocks}}"
    "env.SANDBOX_OUTPUT_VPC_PRIVATE_SUBNET_IDS" = "{{.nuon.install.sandbox.outputs.vpc.private_subnet_ids}}"
    "env.SANDBOX_OUTPUT_VPC_PUBLIC_SUBNET_CIDR_BLOCKS" = "{{.nuon.install.sandbox.outputs.vpc.public_subnet_cidr_blocks}}"
    "env.SANDBOX_OUTPUT_VPC_PUBLIC_SUBNET_IDS" = "{{.nuon.install.sandbox.outputs.vpc.public_subnet_ids}}"
    // sandbox account outputs
    "env.SANDBOX_OUTPUT_ACCOUNT_ID" = "{{.nuon.install.sandbox.outputs.account.id}}"
    "env.SANDBOX_OUTPUT_ACCOUNT_REGION" = "{{.nuon.install.sandbox.outputs.account.region}}"
    // sandbox ecr outputs
    "env.SANDBOX_OUTPUT_ECR_REPOSITORY_URL" = "{{.nuon.install.sandbox.outputs.ecr.repository_url}}"
    "env.SANDBOX_OUTPUT_ECR_REPOSITORY_ARN" = "{{.nuon.install.sandbox.outputs.ecr.repository_arn}}"
    "env.SANDBOX_OUTPUT_ECR_REPOSITORY_NAME" = "{{.nuon.install.sandbox.outputs.ecr.repository_name}}"
    "env.SANDBOX_OUTPUT_ECR_REGISTRY_ID" = "{{.nuon.install.sandbox.outputs.ecr.registry_id}}"
    "env.SANDBOX_OUTPUT_ECR_REGISTRY_URL" = "{{.nuon.install.sandbox.outputs.ecr.registry_url}}"
    // sandbox public domain outputs
    "env.SANDBOX_OUTPUT_PUBLIC_DOMAIN_NAMESERVERS" = "{{.nuon.install.sandbox.outputs.public_domain.nameservers}}"
    "env.SANDBOX_OUTPUT_PUBLIC_DOMAIN_NAME" = "{{.nuon.install.sandbox.outputs.public_domain.name}}"
    "env.SANDBOX_OUTPUT_PUBLIC_DOMAIN_ZONE_ID" = "{{.nuon.install.sandbox.outputs.public_domain.zone_id}}"
    // sandbox internal domain outputs
    "env.SANDBOX_OUTPUT_INTERNAL_DOMAIN_NAMESERVERS" = "{{.nuon.install.sandbox.outputs.internal_domain.nameservers}}"
    "env.SANDBOX_OUTPUT_INTERNAL_DOMAIN_NAME" = "{{.nuon.install.sandbox.outputs.internal_domain.name}}"
    "env.SANDBOX_OUTPUT_INTERNAL_DOMAIN_ZONE_ID" = "{{.nuon.install.sandbox.outputs.internal_domain.zone_id}}"
  }
}
