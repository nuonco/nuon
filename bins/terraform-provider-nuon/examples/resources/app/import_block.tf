# Apps can be imported via import block.

import {
  to = nuon_app.my_app
  id = "app123"
}

resource "nuon_app" "my_app" {
  name = "My App"
}
