variable "env" {}

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
