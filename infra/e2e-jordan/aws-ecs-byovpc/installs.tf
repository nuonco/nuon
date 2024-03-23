resource "nuon_install" "us_west_2" {
  count  = 1
  app_id = nuon_app.my_byovpc_app.id

  name         = nuon_app.my_byovpc_app.name
  region       = "us-west-2"
  iam_role_arn = module.access.iam_role_arn

  input {
    name  = "vpc_id"
    value = var.vpc_id
  }

  depends_on = [
    nuon_app_sandbox.main,
    nuon_app_runner.main,
  ]
}
