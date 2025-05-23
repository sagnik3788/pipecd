terraform {
  required_version = ">= 1.0.0"
}

resource "null_resource" "test" {}
 
variable "environment" {
  type = string
} 