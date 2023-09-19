resource "nuon_app" "main" {
  name = "managed-by-terraform"
}

data "nuon_app" "main_data" {
  id = nuon_app.main.id
}
