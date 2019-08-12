#!/usr/bin/env bash
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


rm KC1;touch KC1;export KUBECONFIG=KC1; \
gcloud container clusters get-credentials c-b-bcicen-uc-secrets-vault-13 --zone us-west1-b --project appcrd-cicenas-20190806

export PROJECT_NAME=appcrd-cicenas-20190806
export APPCONFIG_CRD_PREFIX=app-crd-vault-test
export APPCONFIG_CRD_CLUSTER=${APPCONFIG_CRD_PREFIX}-appcrd-cicenas-20190806-c-b-bcicen-uc-secrets-vault-13

export VAULT_ADDR=https://35.245.198.47:8200
export VAULT_CACERT=/Users/joseret/dev/m1/test1/ca.pem
export VAULT_GCP_RELATED_PREFIX=gcp-${APPCONFIG_CRD_CLUSTER}
export VAULT_KSA_RELATED=k8s-${APPCONFIG_CRD_CLUSTER}

export ROLE_NAME=${APPCONFIG_CRD_PREFIX}-role

export VAULT_NS=${APPCONFIG_CRD_PREFIX}-ns
export VAULT_KSA=${APPCONFIG_CRD_PREFIX}-ksa
export VAULT_ROLE_NAME=${APPCONFIG_CRD_PREFIX}-role
export VAULT_SA_KEY_PATH=./vault-sa1.json

#rm *.hcl
#gsutil cp gs://anthos-appconfig_build/tests/gsa_keys/appcrd-cicenas-20190806/* .

set -xe

VAULT_SA_EMAIL=${APPCONFIG_CRD_PREFIX}-vault-sa1@${PROJECT_NAME}.iam.gserviceaccount.com
CHECK_VAULT_SA1=$(gcloud iam service-accounts describe ${VAULT_SA_EMAIL}  \
  --project ${PROJECT_NAME}  --format="value(name)"   || echo "")
[ ! -z  $CHECK_VAULT_SA1 ] || gcloud iam service-accounts create ${APPCONFIG_CRD_PREFIX}-vault-sa1 \
  --display-name=${APPCONFIG_CRD_PREFIX}-vault-sa1 --project ${PROJECT_NAME}


if [ ! -f ${VAULT_SA_KEY_PATH} ] ; then
  gcloud iam service-accounts keys create "${VAULT_SA_KEY_PATH}" --project ${PROJECT_NAME} \
    --iam-account=${VAULT_SA_EMAIL}
fi

#CHECK_GCP_1=$(vault read "${VAULT_GCP_RELATED_PREFIX}")
#[ ! -z  $CHECK_GCP_1 ] || vault secrets enable --path="${VAULT_GCP_RELATED_PREFIX}" gcp

#vault auth enable --path="${VAULT_KSA_RELATED}" kubernetes
CHECK_GCP_2=$(vault read "${VAULT_GCP_RELATED_PREFIX}/config")
[[ ! -z  "$CHECK_GCP_2" ]] || echo "vault write ${VAULT_GCP_RELATED_PREFIX}/config"
#vault write ${VAULT_GCP_RELATED_PREFIX}/config \
#  project=${PROJECT_NAME} \
#  ttl=180 \
#  max_ttl=300 \
#  credentials=@${VAULT_SA_KEY_PATH}





gcloud projects add-iam-policy-binding ${PROJECT_NAME} \
  --member=serviceAccount:${VAULT_SA_EMAIL} \
  --role roles/pubsub.admin

gcloud projects add-iam-policy-binding  ${PROJECT_NAME} \
  --member=serviceAccount:${VAULT_SA_EMAIL} \
  --role roles/iam.serviceAccountAdmin

gcloud projects add-iam-policy-binding  ${PROJECT_NAME} \
  --member=serviceAccount:${VAULT_SA_EMAIL} \
  --role roles/iam.serviceAccountKeyAdmin



export ROLE_NAME=${VAULT_ROLE_NAME}
export GCP_RELATED_PREFIX=${VAULT_GCP_RELATED_PREFIX}

. ../../../../examples/use-cases/uc-secrets-vault-k8s/vault-roles-policy.sh

cat ${ROLE_NAME}-policy.hcl
cat ${ROLE_NAME}-gcp.hcl


vault policy write ${VAULT_ROLE_NAME} ./${VAULT_ROLE_NAME}-policy.hcl

vault write ${VAULT_GCP_RELATED_PREFIX}/roleset/${VAULT_ROLE_NAME} \
  project="${PROJECT_NAME}" \
  secret_type="service_account_key"  \
  bindings=@${VAULT_ROLE_NAME}-gcp.hcl \


vault write auth/${VAULT_KSA_RELATED}/role/${VAULT_ROLE_NAME} \
  bound_service_account_names="${VAULT_KSA}" \
  bound_service_account_namespaces="${VAULT_NS}" \
  policies=${VAULT_ROLE_NAME} \
  ttl=5m


[ ! -z  "$(kubectl get namespace ${VAULT_NS} --output 'jsonpath={.metadata.name}')" ] || kubectl create ns ${VAULT_NS}
[ ! -z  "$(kubectl get sa -n ${VAULT_NS} vault-auth --output 'jsonpath={.metadata.name}')" ] || kubectl create -n ${VAULT_NS} sa vault-auth
[ ! -z  "$(kubectl get secret vault-ca -n  ${VAULT_NS}  --output 'jsonpath={.metadata.name}')" ] || kubectl create secret generic vault-ca \
  --namespace=${VAULT_NS} \
  --from-file=${VAULT_CACERT}


VAULT_SA_SECRET=$(kubectl get -n appconfigmgrv2-system sa vault-auth -o jsonpath="{.secrets[*]['name']}")
VAULT_SA_JWT_TOKEN=$(kubectl get -n appconfigmgrv2-system secret $VAULT_SA_SECRET -o jsonpath="{.data.token}" | base64 --decode; echo)
VAULT_SA_CA_CRT=$(kubectl get -n appconfigmgrv2-system secret $VAULT_SA_SECRET -o jsonpath="{.data['ca\.crt']}" | base64 --decode; echo)
VAULT_REVIEWER_CLUSTER=$(kubectl config current-context)
VAULT_REVIEWER_CLIENT_API_SERVER=$(kubectl config view -o jsonpath="{.clusters[?(@.name==\"${VAULT_REVIEWER_CLUSTER}\")].cluster.server}")

VAULT_USER_SECRET=$(kubectl get -n ${VAULT_NS} sa ${VAULT_KSA} -o jsonpath="{.secrets[*]['name']}")
VAULT_USER_JWT_TOKEN=$(kubectl get -n ${VAULT_NS} secret $VAULT_USER_SECRET -o jsonpath="{.data.token}" | base64 --decode; echo)
VAULT_USER_CA_CRT=$(kubectl get -n a ${VAULT_NS} secret $VAULT_USER_SECRET -o jsonpath="{.data['ca\.crt']}" | base64 --decode; echo)


vault write auth/${VAULT_KSA_RELATED}/config \
  token_reviewer_jwt="${VAULT_SA_JWT_TOKEN}" \
  kubernetes_host="${VAULT_REVIEWER_CLIENT_API_SERVER}" \
  kubernetes_ca_cert="${VAULT_SA_CA_CRT}"

#vaupath “secret/demo/*” {
# capabilities = [“create”, “read”, “update”, “delete”, “list”]
#}

cat > test-vault-auth.yaml << EOF
apiVersion: rbac.authorization.k8s.io/v1beta1
kind: ClusterRoleBinding
metadata:
  name: role-tokenreview-binding
  namespace: ${VAULT_NS}
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: system:auth-delegator
subjects:
- kind: ServiceAccount
  name: vault-auth
  namespace: ${VAULT_NS}
---
apiVersion: apps/v1beta1
kind: Deployment
metadata:
  name: ${VAULT_NS}-dep
  namespace: ${VAULT_NS}
  labels:
    app: ${VAULT_NS}-app
    version: v1.0.0
spec:
  replicas: 1
  template:
    metadata:
      labels:
        app: ${VAULT_NS}-app
        version: v1.0.0
    spec:
      serviceAccountName:  ${VAULT_KSA}
      containers:
        - name: ${VAULT_NS}-app-c1
          image: gcr.io/anthos-appconfig/vault-api-helper:latest
          imagePullPolicy: Always
          command: ["/bin/sh"]
          args: ["-c", "while [ true ] ; do echo 'sleeping..'; sleep 10; done"]
          tty: true
          env:
           - name: KSA_JWT
             value: ${VAULT_USER_JWT_TOKEN}
           - name: INIT_GCP_KEYPATH
             value:  ${VAULT_GCP_RELATED_PREFIX}/key/${VAULT_ROLE_NAME}
           - name: INIT_K8S_KEYPATH
             value: auth/${VAULT_KSA_RELATED}/role/${VAULT_ROLE_NAME}
           - name: VAULT_ADDR
             value: ${VAULT_ADDR}
           - name: VAULT_CAPATH
             value: /stuff/vault-info/ca.pem
           - name: GOOGLE_APPLICATION_CREDENTIALS
             value: /stuff/google-info/key.json
          volumeMounts:
            - mountPath: /stuff/google-info
              name: google-auth-token
            - mountPath: /stuff/vault-info
              name: cert
              readOnly: true
            - mountPath: /var/run/secrets/tokens
              name: vault-token
              readOnly: true
      volumes:
        - emptyDir:
            medium: Memory
          name: google-auth-token
        - name: cert
          secret:
            defaultMode: 420
            secretName: vault-ca
        - name: vault-token
          projected:
            sources:
            - serviceAccountToken:
                path: vault-token
                expirationSeconds: 7200
                audience: vault
EOF


kubectl apply -f test-vault-auth.yaml



echo "kubectl exec -it -n ${VAULT_NS} \$(kubectl get pod  -n \${VAULT_NS} --selector=app=\${VAULT_NS}-app -o=jsonpath='{.items[0].metadata.name'}) /bin/sh"

kubectl exec -n app-crd-vault-test-ns $(kubectl get pod  -n app-crd-vault-test-ns--selector=app=app-crd-vault-test-ns-app -o=jsonpath='{.items[0].metadata.name}') /bin/sh