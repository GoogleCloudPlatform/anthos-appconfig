locals {
    build_roles = [
      "roles/compute.admin",
      "roles/editor",
      "roles/container.clusterAdmin",
      "roles/iam.roleAdmin",
      "roles/resourcemanager.projectIamAdmin"

    ] ,

}


module "project-factory-build" {
  source = "git@github.com:joseret/terraform-google-project-factory.git?ref=fork-v1"
  random_project_id = "false"
  name = "appconfig-crd-env-${var.suffix}"
  folder_id = "${var.folder_id}"
  org_id            = "${var.org_id}"
  billing_account = "${var.billing_id}"
  disable_services_on_destroy = false
  disable_dependent_services = false
  bucket_name = "appconfig-crd-env-${var.suffix}"
  bucket_project = "appconfig-crd-env-${var.suffix}"
  //  shared_vpc = "${data.terraform_remote_state.net.host_project_id}"
  activate_apis = [
    "cloudbuild.googleapis.com",
    "compute.googleapis.com",
    "cloudresourcemanager.googleapis.com",
    "cloudbilling.googleapis.com",
//    "storage-components.googleapis.com",
    "container.googleapis.com",
    "sourcerepo.googleapis.com",
    "cloudkms.googleapis.com",
    "serviceusage.googleapis.com",
  ],
  //  shared_vpc_subnets = "${local.subnet_self_links_clean_join}"
}

resource "google_project_iam_member" "project" {
  count = "${length(local.build_roles)}"
  project = "${module.project-factory-build.project_id}"
  role               = "${element(local.build_roles, count.index)}"
  member             = "serviceAccount:${module.project-factory-build.project_number}@cloudbuild.gserviceaccount.com"
}

