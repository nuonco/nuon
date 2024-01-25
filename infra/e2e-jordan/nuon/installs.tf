resource "nuon_install" "east_1" {
  count  = var.east_1_count
  app_id = nuon_app.main.id

  name         = "east-1-${count.index}"
  region       = "us-east-1"
  iam_role_arn = var.install_role_arn

  dynamic "input" {
    for_each = var.install_inputs
    content {
      name  = input.value.name
      value = input.value.value
    }
  }

  depends_on = [
    nuon_app_sandbox.main,
    nuon_app_runner.main,
  ]
}

resource "nuon_install" "east_2" {
  count  = var.east_2_count
  app_id = nuon_app.main.id

  name         = "east-2-${count.index}"
  region       = "us-east-2"
  iam_role_arn = var.install_role_arn

  dynamic "input" {
    for_each = var.install_inputs
    content {
      name  = input.value.name
      value = input.value.value
    }
  }

  depends_on = [
    nuon_app_sandbox.main,
    nuon_app_runner.main,
  ]
}

resource "nuon_install" "west_2" {
  count  = var.west_2_count
  app_id = nuon_app.main.id

  name         = "west-2-${count.index}"
  region       = "us-west-2"
  iam_role_arn = var.install_role_arn

  dynamic "input" {
    for_each = var.install_inputs
    content {
      name  = input.value.name
      value = input.value.value
    }
  }

  depends_on = [
    nuon_app_sandbox.main,
    nuon_app_runner.main,
  ]
}
