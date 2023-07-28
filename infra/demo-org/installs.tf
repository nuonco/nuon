resource "nuon_install" "demo-stage" {
  app_id = nuon_app.main.id

  name = "managed-by-terraform"
  region = "us-east-1"
  iam_role_arn = "arn:aws:iam::949309607565:role/nuon-demo-install-access"
}
