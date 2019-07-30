```bash

PROJECT_ID=appcrd-cicenas-20190730
GCP_RELATED_PREFIX="gcp-${PROJECT_ID}"
KSA_RELATED=gke-appcrd-cicenas-20190730-us-west1-b-c-b-bcicen-uc-secrets-vault-13

vault secrets enable --path="${GCP_RELATED_PREFIX}" gcp
vault write ${GCP_RELATED_PREFIX}/config credentials=@sa1-vault.json


```

```bash
gcloud iam service-accounts add-iam-policy-binding \
  --role roles/iam.serviceAccountAdmin \
  --role roles/iam.serviceAccountKeyAdmin \
  --role roles/resourcemanager.projectIamAdmin \
  --role roles/pubsub.admin \
  --member serviceAccount:${SERVICE_ACCOUNT}@${PROJECT_ID}.iam.gserviceaccount.com \
  ${SERVICE_ACCOUNT}@${PROJECT_ID}.iam.gserviceaccount.com

```

```bash
VAULT_SA_SECRET=$(kubectl get -n appconfigmgrv2-system sa vault-auth -o jsonpath="{.secrets[*]['name']}")
VAULT_SA_JWT_TOKEN=$(kubectl get -n appconfigmgrv2-system secret $VAULT_SA_SECRET -o jsonpath="{.data.token}" | base64 --decode; echo)
VAULT_SA_CA_CRT=$(kubectl get -n appconfigmgrv2-system secret $VAULT_SA_SECRET -o jsonpath="{.data['ca\.crt']}" | base64 --decode; echo)



vault auth enable --path="${KSA_RELATED}" kubernetes

vault write auth/${KSA_RELATED}/config \
    token_reviewer_jwt="$VAULT_SA_JWT_TOKEN" \
    kubernetes_host=https://${K8S_HOST} \
    kubernetes_ca_cert="$VAULT_SA_CA_CRT"

```

### CRD Setup Helper
```bash
export VAULT_ADDR=<vault url 8200>
export VAULT_CACERT=ca.pem
kubectl create configmap vault \
  --namespace=appconfigmgrv2-system \
  --from-literal vault-addr=${VAULT_ADDR}

kubectl create secret generic vault-ca \
  --namespace=appconfigmgrv2-system \
  --from-file=${VAULT_CACERT}

```

### Build (setup)
```bash

. /Users/joseret/go112/src/github.com/GoogleCloudPlatform/anthos-appconfig/examples/use-cases/uc-secrets-vault-k8s/vault-roles.sh

export ROLE_NAME="uc-secrets-vault-k8s"

vault write ${GCP_RELATED_PREFIX}/roleset/${ROLE_NAME} \
    project="${PROJECT_NAME}" \
    secret_type="service_account_key"  \
    bindings=@${ROLE_NAME}.hcl
    

```