resource "nuon_app" "main" {
  name = var.app_name
  org_id = data.nuon_org.org.id
}
