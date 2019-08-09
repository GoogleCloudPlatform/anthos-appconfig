cat >  ${ROLE_NAME}-gcp.hcl <<EOF
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

cat > ${ROLE_NAME}-policy.hcl <<EOF
path "${GCP_VAULT_PREFIX}/key/${ROLE_NAME}" {
  capabilities = ["read"]
}
EOF