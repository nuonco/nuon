locals {
  twingate = {
    connector_count = 2
    namespace       = "twingate"
  }
}

resource "twingate_remote_network" "vpc" {
  name = local.workspace_trimmed
}

resource "twingate_resource" "internal_dns" {
  name              = local.dns.zone
  address           = "*.${local.dns.zone}"
  remote_network_id = twingate_remote_network.vpc.id

  access {
    group_ids = [
      data.twingate_groups.engineers.groups[0].id,
      data.twingate_groups.internal_access.groups[0].id,
    ]
    service_account_ids = [
      twingate_service_account.github_actions.id
    ]
  }
}

resource "twingate_resource" "private_subnets" {
  for_each          = toset(local.networks[local.workspace_trimmed].private_subnets)
  name              = each.value
  address           = each.value
  remote_network_id = twingate_remote_network.vpc.id

  access {
    group_ids = [
      data.twingate_groups.engineers.groups[0].id,
    ]
    service_account_ids = [
      twingate_service_account.github_actions.id
    ]
  }
}

resource "twingate_connector" "vpc_connector" {
  count             = local.twingate.connector_count
  remote_network_id = twingate_remote_network.vpc.id
}

resource "twingate_connector_tokens" "vpc_connector_tokens" {
  count        = local.twingate.connector_count
  connector_id = twingate_connector.vpc_connector[count.index].id
}

resource "helm_release" "twingate" {
  count            = local.twingate.connector_count
  namespace        = local.twingate.namespace
  create_namespace = true

  name       = "twingate-${twingate_connector.vpc_connector[count.index].name}"
  repository = "https://twingate.github.io/helm-charts"
  chart      = "connector"
  version    = "0.1.13"

  set {
    name  = "icmpSupport.enabled"
    value = "true"
  }

  set {
    name  = "connector.network"
    value = "nuonco"
  }

  set_sensitive {
    name  = "connector.accessToken"
    value = twingate_connector_tokens.vpc_connector_tokens[count.index].access_token
  }

  set_sensitive {
    name  = "connector.refreshToken"
    value = twingate_connector_tokens.vpc_connector_tokens[count.index].refresh_token
  }

  depends_on = [
    kubectl_manifest.karpenter_provisioner,
  ]
}
