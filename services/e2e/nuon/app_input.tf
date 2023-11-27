resource "nuon_app_input" "main" {
  app_id = nuon_app.main.id

  input {
    name = "required-value"
    description = "required value description"
    default = ""
    required = true
  }

  input {
    name = "optional-value"
    description = "optional value description"
    default = "default"
    required = false
  }
}
