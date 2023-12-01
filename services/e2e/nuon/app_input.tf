resource "nuon_app_input" "main" {
  app_id = nuon_app.main.id

  input {
    name = "required_input"
    description = "required value description"
    default = ""
    required = true
  }

  input {
    name = "optional_input"
    description = "optional value description"
    default = "default"
    required = false
  }
}
