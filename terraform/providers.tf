terraform {
  backend "s3" {
    bucket = "app-blog"
    key    = "blog.tfstate"
    region = "fr-par"

    skip_credentials_validation = true
    skip_region_validation      = true
    # Need terraform>=1.6.1
    skip_requesting_account_id = true

    endpoints = {
      s3 = "https://s3.fr-par.scw.cloud"
    }
  }

  required_providers {
    scaleway = {
      source = "scaleway/scaleway"
    }
  }
  required_version = ">= 0.13"
}

provider "scaleway" {
  alias   = "p2"
  profile = "myProfile"
  zone    = "fr-par-1"
  region  = "fr-par"
}

