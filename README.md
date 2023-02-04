# Tfv

Tfv generates Terraform variable declarations from `var.foo` in *.tf files that are not already declared.

Only works on tf files in current directory, and generate empty description field for now.

# installation

```
go get github.com/teru01/tfv
```

# example

Supporse you define resources like below in *.tf file

```
resource "aws_s3_bucket" "example" {
  bucket        = var.my_bucket_name
  ...
}
```

then you can generate `variable` block.

```
$ tfv
variable "my_bucket_name" {
  description = ""
}

```

currently references to a variable inside `%{}` expression is not supported.


