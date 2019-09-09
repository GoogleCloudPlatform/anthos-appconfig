output "service_account_email" {
  value       = "${module.project-factory-build.service_account_email}"
  description = "The email-build of the default service account"
}

output "service_account_name" {
  value       = "${module.project-factory-build.service_account_name}"
  description = "The fully-qualified name-build of the default service account"
}


output "project_bucket_self_link" {
  value       = "${module.project-factory-build.project_bucket_self_link}"
  description = "Project's bucket-build selfLink"
}

output "project_bucket_url" {
  value       = "${module.project-factory-build.project_bucket_url}"
  description = "Project's bucket-build url"
}

output "build_project_id" {
  value       = "${module.project-factory-build.project_id}"
  description = "Build - Project id"
}



