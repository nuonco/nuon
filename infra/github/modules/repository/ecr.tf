data "aws_caller_identity" "current" {}

# Optionally, create an ECR repo for the GH repo
# This should be enabled for any services deployed to EKS
module "ecr" {
  source  = "terraform-aws-modules/ecr/aws"
  version = ">= 2.4.0"

  create                                    = var.enable_ecr == true && var.archived == false
  create_repository                         = true
  create_repository_policy                  = false
  attach_repository_policy                  = true
  create_lifecycle_policy                   = true
  create_registry_replication_configuration = true

  // NOTE(jm): we enable mutable images to help when debugging our CI systems. The only time we would _ever_ run into an
  // issue where the image is updated in place is when our CI is broken for some reason.
  repository_image_tag_mutability = "MUTABLE"

  repository_name               = var.name
  repository_encryption_type    = "KMS"
  repository_image_scan_on_push = true
  repository_policy             = data.aws_iam_policy_document.ecr_policy.json
  repository_lifecycle_policy = jsonencode({
    rules = [
      {
        "rulePriority": 1,
        "description": "Remove untagged images older than 7 days",
        "selection": {
          "tagStatus": "untagged",
          "countType": "sinceImagePushed",
          "countUnit": "days",
          "countNumber": 7
        },
        "action": {
          "type": "expire"
        }
      },
    ]
  })
  registry_replication_rules = [
    {
      destinations = [{
        region      = "us-east-2"
        registry_id = data.aws_caller_identity.current.account_id
      }]
    }
  ]

  tags = { service = var.name }
}


# If extra_ecr_repos is set, create additional repositories that include the name of the github repository as a prefix
module "extra-ecr-repos" {
  source  = "terraform-aws-modules/ecr/aws"
  version = ">= 1.3.2"

  count                                     = length(var.extra_ecr_repos)
  create_repository                         = true
  create_repository_policy                  = false
  attach_repository_policy                  = true
  create_lifecycle_policy                   = true
  create_registry_replication_configuration = true

  repository_name                 = "${var.name}/${element(var.extra_ecr_repos, count.index)}"
  repository_image_tag_mutability = "MUTABLE"
  repository_encryption_type      = "KMS"
  repository_image_scan_on_push   = true
  repository_policy               = data.aws_iam_policy_document.ecr_policy.json
  repository_lifecycle_policy = jsonencode({
    rules = [
      {
        "rulePriority": 1,
        "description": "Remove untagged images older than 7 days",
        "selection": {
          "tagStatus": "untagged",
          "countType": "sinceImagePushed",
          "countUnit": "days",
          "countNumber": 7
        },
        "action": {
          "type": "expire"
        }
      },
    ]
  })
  registry_replication_rules = [
    {
      destinations = [{
        region      = "us-east-2"
        registry_id = data.aws_caller_identity.current.account_id
      }]
    }
  ]

  tags = { service = var.name }
}

data "aws_iam_policy_document" "ecr_policy" {
  statement {
    effect = "Allow"
    actions = [
      "ecr:BatchCheckLayerAvailability",
      "ecr:BatchDeleteImage",
      "ecr:BatchGetImage",
      "ecr:CompleteLayerUpload",
      "ecr:DescribeImageScanFindings",
      "ecr:DescribeImages",
      "ecr:DescribeRepositories",
      "ecr:GetDownloadUrlForLayer",
      "ecr:GetLifecyclePolicy",
      "ecr:GetLifecyclePolicyPreview",
      "ecr:GetRepositoryPolicy",
      "ecr:InitiateLayerUpload",
      "ecr:ListImages",
      "ecr:ListTagsForResource",
      "ecr:PutImage",
      "ecr:UploadLayerPart"
    ]

    principals {
      type        = "AWS"
      identifiers = ["*", ]
    }
    condition {
      test     = "StringEquals"
      variable = "aws:PrincipalOrgID"
      # NOTE(jdt): this sucks but it's better than passing in the same value for every module invocation
      # TODO(jdt): should this be restricted to just accounts in the workloads ou?
      values = ["o-thxealue7f", ]
    }
  }
}
