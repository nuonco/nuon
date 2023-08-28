resource "nuon_install" "east_1" {
  count = var.east_1_count
  app_id = nuon_app.main.id

  name = "east-1-${count.index}"
  region = "us-east-1"
  iam_role_arn = var.install_role_arn

  depends_on = [
    nuon_docker_build_component.e2e,
    nuon_helm_chart_component.e2e,
    nuon_terraform_module_component.e2e,
    nuon_container_image_component.e2e
  ]
}

resource "nuon_install" "east_2" {
  count = var.east_2_count
  app_id = nuon_app.main.id

  name = "east-2-${count.index}"
  region = "us-east-2"
  iam_role_arn = var.install_role_arn

  depends_on = [
    nuon_docker_build_component.e2e,
    nuon_helm_chart_component.e2e,
    nuon_terraform_module_component.e2e,
    nuon_container_image_component.e2e
  ]
}

resource "nuon_install" "west_2" {
  count = var.west_2_count
  app_id = nuon_app.main.id

  name = "west-2-${count.index}"
  region = "us-west-2"
  iam_role_arn = var.install_role_arn

  depends_on = [
    nuon_docker_build_component.e2e,
    nuon_helm_chart_component.e2e,
    nuon_terraform_module_component.e2e,
    nuon_container_image_component.e2e
  ]
}
