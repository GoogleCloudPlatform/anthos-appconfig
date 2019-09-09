terraform {
  backend "gcs" {
    bucket  = "appconfig-crd-env-bld"
    prefix    = "env-r2/build/build"
  }
}

//provider "google" {
//  version     = "~> 1.19"
//}

provider "google-beta" {
  //  version     = "~> 1.19"
}


