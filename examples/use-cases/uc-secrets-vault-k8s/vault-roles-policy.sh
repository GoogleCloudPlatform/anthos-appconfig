# Copyright 2019 Google LLC
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#      http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#
# Copyright 2019 Google LLC. This software is provided as-is,
# without warranty or representation for any use or purpose.#
#

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