resource "grafana_cloud_stack" "main" {
  provider = grafana.org

  name        = "nuon-main"
  slug        = "nuon"
  region_slug = "us" # Example “us”,”eu” etc
}

# Creating an API key in Grafana instance to be used for creating resources in Grafana instance
resource "grafana_api_key" "main" {
  provider = grafana.org

  cloud_stack_slug = grafana_cloud_stack.main.slug
  name             = "nuon-main-terraform"
  role             = "Admin"
}

