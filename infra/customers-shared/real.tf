data "utils_deep_merge_yaml" "real" {
  input = [
    file("vars/real.yml"),
  ]
}

locals {
  real = yamldecode(data.utils_deep_merge_yaml.real.output)

  real_installs_array = flatten([
    for app in local.real.apps : [
      for install in app.real_installs : {
        app = app
        install = install
      }
    ]
  ])

  # create a map, where all installs have a key, that we can use to look them up + id them.
  real_installs = {
    for si in local.real_installs_array : "${si.app.name}.${si.install.name}" => si
  }
}

resource "nuon_app" "real" {
  provider = nuon.real
  for_each = { for app in local.real.apps : app.name => app }

  name       = each.value.name
  depends_on = [nuon_app.sandbox]
}

resource "nuon_app_runner" "real" {
  for_each = { for app in local.real.apps : app.name => app }

  provider = nuon.sandbox
  app_id   = nuon_app.real[each.value.name].id

  runner_type = "aws-eks"
  env_var {
    name = "NUON_RUNNER_TYPE"
    value = "aws-eks"
  }
}

resource "nuon_app_sandbox" "real" {
  for_each = { for app in local.real.apps : app.name => app }

  provider = nuon.real
  app_id   = nuon_app.real[each.value.name].id

  terraform_version = "v1.6.3"
  public_repo = {
    repo      = "nuonco/sandboxes"
    branch    = "main"
    directory = "aws-eks"
  }
}

resource "nuon_app_installer" "real" {
  provider = nuon.real
  for_each = { for app in local.real.apps : app.name => app }

  app_id      = nuon_app.real[each.value.name].id
  name        = each.value.installer.name
  description = each.value.installer.description
  slug        = each.value.installer.slug

  documentation_url = each.value.urls.documentation
  community_url     = each.value.urls.community
  logo_url          = each.value.urls.logo
  github_url        = each.value.urls.github
  homepage_url      = each.value.urls.homepage
  demo_url          = each.value.urls.demo
  post_install_markdown = ""
}

resource "nuon_install" "real" {
  #for_each = var.disable_installs ? {} : local.real_installs
  for_each = local.real_installs

  provider     = nuon.real
  name         = each.value.install.name
  app_id       = nuon_app.real[each.value.app.name].id
  region       = each.value.install.region
  iam_role_arn = each.value.install.iam_role_arn

  depends_on = [
    nuon_app_sandbox.real
  ]
}
