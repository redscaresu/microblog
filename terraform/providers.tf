terraform {
    backend "s3" {
      bucket = "terraform-state-blog"
      key    = "terraform.tfstate"
      region = "uk"
      endpoints = {
        s3 = "https://s3.uk.io.cloud.ovh.net/"
      }
      skip_credentials_validation = true
      skip_region_validation      = true
      skip_requesting_account_id  = true
      skip_s3_checksum            = true
    }
}