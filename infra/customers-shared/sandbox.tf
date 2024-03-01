data "utils_deep_merge_yaml" "sandbox" {
  input = [
    file("vars/sandbox.yml"),
  ]
}

locals {
  sandbox = yamldecode(data.utils_deep_merge_yaml.sandbox.output)

  # this creates a flattened list of all installs for all apps, dynamically
  sandbox_installs_array = flatten([
    for app in local.sandbox.apps : [
      for idx in range(app.sandbox_install_count): {
          install = idx
          inputs = app.sandbox_install_inputs
          app     = app
      }
    ]
  ])

  # create a map, where all installs have a key, that we can use to look them up + id them.
  sandbox_installs = {
    for si in local.sandbox_installs_array : "${si.app.name}.${si.install}" => si
  }
}

resource "nuon_app" "sandbox" {
  for_each = { for app in local.sandbox.apps : app.name => app }

  provider = nuon.sandbox
  name     = each.value.name
}

resource "nuon_app_input" "sandbox" {
  for_each = { for app in local.sandbox.apps : app.name => app }
  app_id   = nuon_app.sandbox[each.value.name].id

  provider = nuon.sandbox

  dynamic "input" {
    for_each = each.value.install_inputs
    content {
      name        = input.value.name
      description = input.value.description
      default     = input.value.default
      required    = input.value.required
      display_name = input.value.name
    }
  }
}

resource "nuon_app_sandbox" "sandbox" {
  for_each = { for app in local.sandbox.apps : app.name => app }

  provider = nuon.sandbox
  app_id   = nuon_app.sandbox[each.value.name].id

  terraform_version = "v1.6.3"
  public_repo = {
    repo      = "nuonco/sandboxes"
    branch    = "main"
    directory = "aws-eks"
  }
}

resource "nuon_app_runner" "sandbox" {
  for_each = { for app in local.sandbox.apps : app.name => app }

  provider = nuon.sandbox
  app_id   = nuon_app.sandbox[each.value.name].id

  runner_type = "aws-eks"
  env_var {
    name = "NUON_RUNNER_TYPE"
    value = "aws-eks"
  }
}

resource "random_pet" "sandbox" {
  for_each = local.sandbox_installs
  keepers = {
    install_id = each.key
  }
}

resource "nuon_install" "sandbox" {
  for_each = local.sandbox_installs

  provider     = nuon.sandbox
  name         = random_pet.sandbox[each.key].id
  app_id       = nuon_app.sandbox[each.value.app.name].id
  region       = "us-east-1"
  iam_role_arn = "iam-role-arn"

  dynamic "input" {
    for_each = each.value.inputs
    iterator = ev
    content {
      name = ev.key
      value = ev.value
    }
  }

  depends_on = [
    nuon_app_sandbox.sandbox
  ]
}

resource "nuon_app_installer" "sandbox" {
  for_each = { for app in local.sandbox.apps : app.name => app }

  provider    = nuon.sandbox
  app_id      = nuon_app.sandbox[each.value.name].id
  name        = each.value.installer.name
  description = each.value.installer.description
  slug        = each.value.installer.slug

  documentation_url = each.value.urls.documentation
  community_url     = each.value.urls.community
  logo_url          = each.value.urls.logo
  github_url        = each.value.urls.github
  homepage_url      = each.value.urls.homepage
  demo_url          = each.value.urls.demo
  post_install_markdown = "Post Install"
}
