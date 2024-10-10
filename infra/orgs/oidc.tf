data "aws_iam_openid_connect_provider" "runners_k8s" {
  provider = aws.runners

  arn = data.tfe_outputs.infra-eks-runners.values.oidc_provider_arn
}

resource "aws_iam_openid_connect_provider" "runners" {
  provider = aws.orgs

  url = data.tfe_outputs.infra-eks-runners.values.cluster_oidc_issuer_url

  client_id_list = [
    "sts.amazonaws.com"
  ]

  thumbprint_list = data.aws_iam_openid_connect_provider.runners_k8s.thumbprint_list
}
