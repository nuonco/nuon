#
# Vantage Kubernetes Agent
#
locals {
  bucket_name   = "nuon-vantage-k8s-agent-${local.workspace_trimmed}"
  account_id    = data.aws_caller_identity.current.account_id
  oidc_provider = try(data.tfe_outputs.infra-eks-nuon.values.oidc_provider, module.eks.oidc_provider)
}

# bucket for deployment state

#
# KMS
#
data "aws_iam_policy_document" "vantage_k8s_agent_bucket_key_policy" {
  # enable IAM User Permissions
  statement {
    effect    = "Allow"
    actions   = ["kms:*", ]
    resources = ["*", ]
    principals {
      type        = "AWS"
      identifiers = [local.account_id, ]
    }
  }

  # allow all principals in this account that are authorized for s3
  statement {
    effect = "Allow"
    actions = [
      "kms:Encrypt",
      "kms:Decrypt",
      "kms:ReEncrypt*",
      "kms:GenerateDataKey*",
      "kms:DescribeKey"
    ]
    resources = ["*", ]
    principals {
      type        = "AWS"
      identifiers = ["*", ]
    }
    condition {
      test     = "StringEquals"
      variable = "kms:ViaService"
      values   = ["s3.us-west-2.amazonaws.com", ]
    }
    condition {
      test     = "StringEquals"
      variable = "kms:CallerAccount"
      values   = [local.account_id, ]
    }
  }

  # allow all principals in the nuon org that are authorized for s3
  statement {
    effect = "Allow"
    actions = [
      "kms:Encrypt",
      "kms:Decrypt",
      "kms:ReEncrypt*",
      "kms:GenerateDataKey*",
      "kms:DescribeKey"
    ]
    resources = ["*", ]
    principals {
      type        = "AWS"
      identifiers = ["*", ]
    }
    condition {
      test     = "StringEquals"
      variable = "kms:ViaService"
      values   = ["s3.us-west-2.amazonaws.com", ]
    }
    condition {
      test     = "StringEquals"
      variable = "aws:PrincipalOrgID"
      values   = [data.aws_organizations_organization.orgs.id]
    }
  }
}

resource "aws_kms_key" "vantage_k8s_agent_bucket" {
  description = "KMS key for ${local.bucket_name}"
  policy      = data.aws_iam_policy_document.vantage_k8s_agent_bucket_key_policy.json
}

resource "aws_kms_alias" "vantage_k8s_agent_bucket" {
  name          = "alias/bucket-key-${local.bucket_name}"
  target_key_id = aws_kms_key.vantage_k8s_agent_bucket.key_id
}

#
# IAM
#

# policy to allow access to this bucket: will be assigned to the default ServiceAccount
# for the vantage k8s agent

data "aws_iam_policy_document" "vantage_k8s_agent_bucket_access_policy" {
  # allow list bucket on this bucket
  statement {
    effect    = "Allow"
    actions   = ["s3:ListBucket"]
    resources = ["arn:aws:s3:::${local.bucket_name}"]
  }

  # allow all object actions on all objects in this bucket
  statement {
    effect    = "Allow"
    actions   = ["s3:*Object"]
    resources = ["arn:aws:s3:::${local.bucket_name}/*"]
  }
}

# so we can attach this to a role with which we tag the ServiceAccount
data "aws_iam_policy_document" "vantage_k8s_agent_trust_policy" {
  statement {
    effect  = "Allow"
    actions = ["sts:AssumeRoleWithWebIdentity"]
    principals {
      type        = "Federated"
      identifiers = ["arn:aws:iam::${local.account_id}:oidc-provider/${local.oidc_provider}"]
    }
    condition {
      test     = "StringEquals"
      variable = "${local.oidc_provider}:aud"
      values   = ["sts.amazonaws.com"]
    }
    condition {
      test     = "StringEquals"
      variable = "${local.oidc_provider}:sub"
      // NOTE(fd): named after the fact because we didn't know what the generator would produce
      values = ["system:serviceaccount:vantage:vantage-k8s-agent-service-account"]
    }
  }
}

# role that can be assumed by the service account and has access to the bucket
resource "aws_iam_role" "vantage_k8s_agent_role" {
  name               = "nuon-vantage_k8s_agent-role-${local.workspace_trimmed}"
  assume_role_policy = data.aws_iam_policy_document.vantage_k8s_agent_trust_policy.json

  # bucket access policy
  inline_policy {
    name   = "nuon-vantage_k8s_agent-role-inline-bucket-access-policy-${local.workspace_trimmed}"
    policy = data.aws_iam_policy_document.vantage_k8s_agent_bucket_access_policy.json
  }

  tags = local.tags
}

#
# Bucket
#

module "bucket" {
  source  = "terraform-aws-modules/s3-bucket/aws"
  version = "v4.11.0"

  bucket = local.bucket_name
  versioning = {
    enabled = false
  }

  attach_deny_insecure_transport_policy = true
  attach_require_latest_tls_policy      = true

  attach_public_policy = false

  control_object_ownership = true
  object_ownership         = "BucketOwnerEnforced"

  # the bucket access policy is inlined with the role
  # this bucket has no bucket policy to dictate access. access is exclusively managed through the role.
  # attach_policy = true
  # policy        = data.aws_iam_policy_document.s3_bucket_policy.json

  server_side_encryption_configuration = {
    rule : [
      {
        apply_server_side_encryption_by_default : {
          kms_master_key_id = aws_kms_key.vantage_k8s_agent_bucket.arn
          sse_algorithm : "aws:kms",
        },
        bucket_key_enabled : true,
      },
    ],
  }
}

# helm release
resource "helm_release" "vantage-k8s-agent" {
  namespace        = "vantage"
  create_namespace = true

  name       = "vantage-kubernetes-agent"
  chart      = "vantage-kubernetes-agent"
  repository = "https://vantage-sh.github.io/helm-charts"
  version    = "1.0.37"

  values = [
    # https://github.com/vantage-sh/helm-charts/blob/main/charts/vantage-kubernetes-agent/values.yaml
    yamlencode({
      "image" = {
        "repository" = "quay.io/vantage-sh/kubernetes-agent"
        "pullPolicy" = "IfNotPresent"
      }

      "agent" = {
        "useDeployment"          = true
        "debug"                  = false
        "logLevel"               = "0"
        "clusterID"              = local.workspace_trimmed
        "token"                  = var.vantage_api_token
        "collectNamespaceLabels" = "true"
        "gpu" = {
          "usageMetrics" = false
        }
      }

      "persistS3" = {
        "bucket" = local.bucket_name
      }

      "service" = {
        "name" = "report"
        "type" = "ClusterIP"
        "port" = 9010
      }

      "resources" = {
        "limits" = {
          "cpu"    = "500m"
          "memory" = "1000Mi"
        }
        "requests" = {
          "cpu"    = "100m"
          "memory" = "100Mi"
        }
      }
      "serviceAccount" = {
        "annotations" = {
          "eks.amazonaws.com/role-arn" = aws_iam_role.vantage_k8s_agent_role.arn
        }
        "name" : "vantage-k8s-agent-service-account"
      }
    }),
  ]
  depends_on = [
    module.bucket
  ]
}
