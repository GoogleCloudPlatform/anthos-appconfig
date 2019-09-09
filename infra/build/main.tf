//resource "google_cloudbuild_trigger" "controller-branch" {
//  trigger_template {
//    branch_name = "master"
//    repo_name   = "my-repo"
//  }
//
//  substitutions = {
//    _FOO = "bar"
//    _BAZ = "qux"
//  }
//
//  filename = "cloudbuild.yaml"
//}

// https://cloud.google.com/sdk/gcloud/reference/alpha/builds/triggers/create/github

data "terraform_remote_state" "remote-project-info" {
  backend = "gcs"

  config = {
    bucket  = "anthos-appconfig_build"
    prefix    = "env-r2/build/projects"

  }

}

resource "google_sourcerepo_repository" "github-mirror" {
  name = "test-x1"
  project = "${data.terraform_remote_state.remote-project-info.build_project_id}"

}