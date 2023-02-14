resource "my_resource" "myvar" {
  url        = "http://semvar.co.jp/${var.appname}/${var.env}"
  title      = var.moge
  expiration = var.expiration_days
}

locals {
  foo = <<EOF
    python
    ${var.limit} + ${var.requests}
    rust
  EOF
}

resource "my_server" "orange" {
  # title      = var.moge
  locations = var.locations

  /*
  apple {
    teste = "sweet"
    created_at = var.created_at
  }
  */

  mogura {
    meta = var.metadata
  }

  lang = var.lang
}
