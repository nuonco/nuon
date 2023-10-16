data "utils_deep_merge_yaml" "sandbox" {
  input = [
    file("vars/sandbox.yml"),
  ]
}

locals {
  sandbox = yamldecode(data.utils_deep_merge_yaml.sandbox.output)
  sandbox_installs = flatten([
    for app in local.sandbox.apps : [
      for install in app.installs : {
        install = install
        app = app
      }
    ]
  ])
}

resource "nuon_app" "sandbox" {
  for_each = { for app in local.sandbox.apps : app.name => app }

  provider = nuon.sandbox
  name = each.value.name
}

resource "nuon_install" "sandbox" {
  for_each = {
    for obj in local.sandbox_installs : "${obj.app.name}.${obj.install}" => obj
  }

  provider = nuon.sandbox
  name = each.value.install
  app_id = nuon_app.sandbox[each.value.app.name].id
  region = "us-east-1"
  iam_role_arn = "iam-role-arn"
}

resource "nuon_app_installer" "sandbox" {
  for_each = { for app in local.sandbox.apps : app.name => app }

  provider = nuon.sandbox
  app_id = nuon_app.sandbox[each.value.name].id
  name = each.value.installer.name
  description = each.value.installer.description
  slug = each.value.installer.slug

  documentation_url = each.value.urls.documentation
  community_url = each.value.urls.community
  logo_url = each.value.urls.logo
  github_url = each.value.urls.github
  homepage_url = each.value.urls.homepage
  demo_url = each.value.urls.demo
}
