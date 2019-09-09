terraform {
  backend "gcs" {
    bucket  = "anthos-appconfig_build"
    prefix    = "env-r2/build/projects"
  }
}

//provider "google" {
//  version     = "~> 1.19"
//}

provider "google-beta" {
  //  version     = "~> 1.19"
}


