resource "nuon_app_input" "meroxa" {
  provider = nuon.sandbox
  app_id   = nuon_app.sandbox["meroxa"].id

  // NOTE: this can also just be a secret input, if needed - where a customer provides the secret directly, if desired.
  input {
    name        = "secret_role_arn"
    required    = true
    description = "ARN for the secret that is placed into AWS SM"
    default     = ""
  }

  input {
    name        = "storage"
    default     = "20Gi"
    required    = false
    description = "description"
  }
}

resource "nuon_helm_chart_component" "meroxa-platform" {
  provider = nuon.sandbox

  name       = "meroxa"
  app_id     = nuon_app.sandbox["meroxa"].id
  chart_name = "meroxa"

  connected_repo = {
    // TODO: change this to use meroxa/mpdx + consider making a directory for `byoc`
    directory = "platform/helm/production"
    repo      = "powertoolsdev/mono"
    branch    = "main"
  }

  value {
    name  = "host"
    value = "mpdx.{{.nuon.install.internal_domain}}"
  }

  value {
    name  = "storage"
    value = "{{.nuon.install.inputs.storage}}"
  }

  value {
    name  = "secret_role_arn"
    value = "{{.nuon.install.inputs.secret_role_arn}}"
  }

  value {
    name  = "region"
    value = "{{.nuon.install.sandbox.outputs.aws_region}}"
  }
}

resource "nuon_install" "meroxa_test_install" {
  provider = nuon.sandbox

  app_id = nuon_app.sandbox["meroxa"].id

  name         = "meroxa"
  region       = "us-west-2"
  iam_role_arn = "arn:aws:iam::949309607565:role/nuon-demo-install-access"

  input {
    name = "secret_role_arn"
    value = "value"
  }
}
