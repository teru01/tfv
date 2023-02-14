variable "env" {}

variable "unused" {
  description = "my variable"
}

variable "appname" {
  description = ""
}

// mogemoge
variable "moge" {
  type = object({
    foo  = number
    hoge = string
  })

  default = {
    foo  = 1
    hoge = "value"
  }
  description = ""
}

variable "unused_last" {
  type = object({
    foo  = number
    hoge = string
  })

  default = {
    foo  = 1
    hoge = "value"
  }
}

# 果物が取れる場所
variable "locations" {
  type        = list(string)
  description = ""
}

variable "expiration_days" {
  description = ""
}

variable "lang" {
  description = ""
}

variable "limit" {
  description = ""
}

variable "metadata" {
  description = ""
}

variable "requests" {
  description = ""
}
