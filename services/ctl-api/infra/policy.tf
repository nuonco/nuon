data "aws_iam_policy_document" "additional_permissions" {
  statement {
    effect = "Allow"
    actions = [
      "s3:ListBucket",
      "s3:*Object",
      "sts:AssumeRole",
    ]
    resources = ["*", ]
  }
  statement {
    effect = "Allow"
    actions = [
      "rds-db:connect",
    ]
    resources = [
      format("arn:aws:rds-db:%s:%s:dbuser:%s/%s",
        local.vars.region,
        local.accounts[var.env].id,
        module.primary.db_instance_resource_id,
        local.name
      ),
    ]
  }
}

resource "aws_iam_policy" "additional_permissions" {
  name   = "eks-policy-${local.name}-additional"
  policy = data.aws_iam_policy_document.additional_permissions.json
}
