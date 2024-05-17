resource "nuon_installer" "main" {
  app_ids      = [
    nuon_app.main.id
  ]
  name        = var.app_name
  description = "${var.app_name} installer"

  documentation_url = "https://docs.nuon.co"
  favicon_url = "data:image/x-icon;base64,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"
  community_url     = "https://join.slack.com/t/nuoncommunity/shared_invite/zt-1q323vw9z-C8ztRP~HfWjZx6AXi50VRA"
  homepage_url      = "https://nuon.co"
  github_url        = "https://github.com/nuonco"
  logo_url          = "https://fakeimg.pl/250x100/"
  demo_url          = "https://www.loom.com/share/aec62b468f9747c59ed5c30c79d473c4"

  footer_markdown = <<EOT
  Footer information
  EOT

  copyright_markdown = <<EOT
  CC Nuon Inc. 2024
  EOT
  post_install_markdown = <<EOT
# Install Post

Your install with id {{.nuon.install.id}} and your install {{.nuon.app.Name}} ({{.nuon.app.id}}) is
{{if eq .nuon.install.Status "active" }}
Please check our docs at {{.nuon.installer.metadata.documentation_url}}.
{{else}}
{{.nuon.install.Status}}.
{{end}}

  EOT
}
