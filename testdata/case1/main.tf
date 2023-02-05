resource "my_resource" "myvar" {
  url        = "http://semvar.co.jp/${var.appname}/${var.env}"
  title      = var.moge
  expiration = var.expiration_days
}

locals {
  foo = "Hello, %{if var.name != "%{if var.name != "${var.foo}"}${var.name}%{else}unnamed%{endif}"}${var.name}%{else}unnamed%{endif}!"
}
