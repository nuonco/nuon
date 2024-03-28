resource "nuon_container_image_component" "e2e-ecr" {
  count = var.create_components ? 1 : 0

  name   = "${var.component_prefix}e2e_ecr_external_image"
  app_id = nuon_app.main.id

  dependencies = [
    nuon_terraform_module_component.e2e[0].id
  ]

  aws_ecr = {
    image_url    = "ecr-image-repository"
    tag          = "latest"
    region       = "us-east-1"
    iam_role_arn = "ecr-access-iam-role-arn"
  }
}
