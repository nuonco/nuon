locals {
  controller = {
    value_file = "values/controller.yaml"
    override_file = "values/controller-${var.env}.yaml"
  }
  
  # Base configuration for runner scale sets
  scale_set_base = {
    value_file = "values/scale-set.yaml"
    override_file = "values/scale-set-${var.env}.yaml"
  }
}

# Install the controller first (required before scale sets)
resource "helm_release" "gha_runner_controller" {
  namespace        = local.vars.controller_namespace
  name             = "gha-runner-scale-set-controller"
  create_namespace = true

  repository = "./charts"
  chart      = "gha-runner-scale-set-controller"
  version    = "0.12.1"

  values = [
    file(local.controller.value_file),
    fileexists(local.controller.override_file) ? file(local.controller.override_file) : "",
  ]
}

# Create GitHub token secret in runner namespace
resource "kubernetes_secret" "gha_runner_github_token" {
  metadata {
    name      = local.vars.github_secret_name
    namespace = local.vars.runner_namespace
  }

  data = {
    github_token = var.github_token
  }

  type = "Opaque"

  depends_on = [helm_release.gha_runner_controller]
}

# Deploy multiple runner scale sets based on configuration
# Only deploy if scale_sets are defined (environment-specific)
resource "helm_release" "gha_runner_scale_sets" {
  for_each = lookup(local.vars, "scale_sets", {})

  namespace = local.vars.runner_namespace
  name      = each.key
  create_namespace = true

  repository = "./charts"
  chart      = "gha-runner-scale-set"
  version    = "0.12.1"

  values = [
    file(local.scale_set_base.value_file),
    fileexists(local.scale_set_base.override_file) ? file(local.scale_set_base.override_file) : "",
    yamlencode({
      runnerScaleSetName = each.key
      githubConfigUrl    = each.value.github_config_url
      githubConfigSecret = local.vars.github_secret_name
      maxRunners         = each.value.max_runners
      minRunners         = each.value.min_runners
      containerMode      = each.value.container_mode
      template           = each.value.template
      controllerServiceAccount = local.vars.controller_service_account
      nodeSelector       = {
        "karpenter.sh/nodepool" = local.vars.node_pool_name
      }
      tolerations = [{
        key      = "pool.nuon.co"
        operator = "Equal"
        value    = local.vars.node_pool_name
        effect   = "NoSchedule"
      }]
    })
  ]

  depends_on = [helm_release.gha_runner_controller, kubernetes_secret.gha_runner_github_token]
}