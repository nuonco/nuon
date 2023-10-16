data "utils_deep_merge_yaml" "real" {
  input = [
    file("vars/real.yml"),
  ]
}

locals {
  real = yamldecode(data.utils_deep_merge_yaml.real.output)
  real_installs = flatten([
    for app in local.real.apps : [
      for install in app.installs : {
        install = install
        app = app
      }
    ]
  ])
}

resource "nuon_app" "real" {
  for_each = { for app in local.real.apps : app.name => app }

  name = each.value.name
  depends_on = [nuon_app.sandbox]
}

resource "nuon_app_installer" "real" {
  for_each = { for app in local.real.apps : app.name => app }

  app_id = nuon_app.real[each.value.name].id
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
