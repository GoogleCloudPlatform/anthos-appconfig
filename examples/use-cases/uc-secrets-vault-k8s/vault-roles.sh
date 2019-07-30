cat > uc-secrets-vault-k8s.hcl <<EOF
resource "projects/${PROJECT_NAME}/topics/appconfigcrd-demo-topic1" {
  roles = [
    "roles/pubsub.publisher",
  ]
}

resource "projects/${PROJECT_NAME}/topics/appconfigcrd-demo-topic2" {
  roles = [
    "roles/pubsub.publisher",
  ]
}
EOF