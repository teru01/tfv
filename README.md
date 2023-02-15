# tfv

tfv generates and/or removes Terraform variable declarations from `var.foo` in *.tf files that are not already declared.

tfv generates Terraform's variables.tf file from the variables used like `var.foo` in the *.tf file

# Installation

```
$ curl -LO https://github.com/teru01/tfv/releases/download/v0.1.0/tfv_0.1.0_[OS]_[ARCH].tar.gz
$ tar -xvf tfv_0.1.0_[OS]_[ARCH].tar.gz
$ mv tfv /usr/local/bin
```

or you can build on your own.

```
go install github.com/teru01/tfv@latest
```

# Usage

```
$ ./tfv -h
NAME:
   tfv - Terraform variables generator

USAGE:
   tfv [global options] command [command options] [arguments...]

COMMANDS:
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --dir DIR              tfv collects variables from DIR (default: ".")
   --sync                 execute in sync mode (tfv generates variables without unused variables) (default: false)
   --variables-file FILE  load variables from FILE (default: "variables.tf")
   --tfvars-file FILE     load tfvars from FILE
   --suffix value         suffix of generated tfvars files (default: ".generated")
   --help, -h             show help (default: false)
```


# Example

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

if you specify `-sync`, tfv remove unused variables, and generates `variables` block only for those used in *.tf files

# Limitations

For strings enclosed in double quotes, only those in the form ${var.foo} are considered variables.

Therefore, `foo` is not considered a variable in use if it is used in a syntax that uses `%{}`, such as `"%{ if $var.foo == "abc"}"`.
