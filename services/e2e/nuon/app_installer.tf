resource "nuon_app_installer" "main" {
  app_id = nuon_app.main.id
  name = var.app_name
  description = "${var.app_name} installer"
  slug = nuon_app.main.id

  documentation_url = "https://docs.nuon.co"
  community_url = "https://join.slack.com/t/nuoncommunity/shared_invite/zt-1q323vw9z-C8ztRP~HfWjZx6AXi50VRA"
  homepage_url = "https://nuon.co"
  github_url = "https://github.com/nuonco"
  logo_url = "https://fakeimg.pl/250x100/"
  demo_url = "https://www.loom.com/share/aec62b468f9747c59ed5c30c79d473c4"

  post_install_markdown = <<EOT
  # Install Post

  Your install with id {{.install.id}}.

  EOT
}
