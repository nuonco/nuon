data "nuon_app" "cool_app" {
  id = "app48yah8clgay8ad442njf01h"
}

resource "nuon_app" "main" {
  name = "managed-by-terraform"
}
