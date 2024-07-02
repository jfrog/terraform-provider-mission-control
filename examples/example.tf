terraform {
  required_providers {
    missioncontrol = {
      source  = "jfrog/mission-control"
      version = "1.0.0"
    }
  }
}

variable "jfrog_url" {
  type = string
  default = "http://localhost:8081"
}

provider "missioncontrol" {
  url = "${var.jfrog_url}"
  // supply JFROG_ACCESS_TOKEN as env var
}

resource "missioncontrol_license_bucket" "my-license-bucket" {
  name = "my-license-bucket"
  url  = "https://buckets.jfrog.io/download/...63aeb8c664"
  key  = "my-license-bucket-key"
}