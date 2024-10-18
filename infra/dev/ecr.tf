module "dev_ecr" {
  source  = "terraform-aws-modules/ecr/aws"
  version = ">= 1.3.2"

  create                                    = true
  create_repository                         = true
  create_repository_policy                  = false
  attach_repository_policy                  = true
  create_lifecycle_policy                   = false
  create_registry_replication_configuration = false

  // NOTE(jm): we enable mutable images to help when debugging our CI systems. The only time we would _ever_ run into an
  // issue where the image is updated in place is when our CI is broken for some reason.
  repository_image_tag_mutability = "MUTABLE"

  repository_name               = "dev"
  repository_image_scan_on_push = true
  repository_policy             = data.aws_iam_policy_document.ecr_policy.json

  tags = {
    "dev": "dev"
  }

  providers = {
    aws = aws.demo
  }
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
      identifiers = ["*"]
    }
  }

  statement {
    effect = "Allow"
    actions = [
      "ecr:GetDownloadUrlForLayer",
      "ecr:GetLifecyclePolicy",
      "ecr:GetLifecyclePolicyPreview",
      "ecr:GetRepositoryPolicy",
      "ecr:InitiateLayerUpload",
      "ecr:ListImages",
      "ecr:ListTagsForResource",
      "ecr:BatchCheckLayerAvailability",
      "ecr:BatchGetImage",
    ]

    principals {
      type        = "*"
      identifiers = ["*", ]
    }
  }
}
