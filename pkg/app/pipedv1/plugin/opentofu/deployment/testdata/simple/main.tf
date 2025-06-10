terraform {
  required_version = ">= 1.0.0"
  required_providers {
    random = {
      source  = "hashicorp/random"
      version = "3.5.1"
    }
  }
}

provider "random" {}

resource "random_string" "test" {
  length  = 8
  special = false
}

resource "null_resource" "test" {}
 
variable "environment" {
  type = string
} 