```bash

PROJECT_ID=appcrd-cicenas-20190730
GCP_RELATED_PREFIX="gcp-${PROJECT_ID}"
KSA_RELATED=gke-appcrd-cicenas-20190730-us-west1-b-c-b-bcicen-uc-secrets-vault-13

vault secrets enable --path="${GCP_RELATED_PREFIX}" gcp
vault write ${GCP_RELATED_PREFIX}/config credentials=@sa1-vault.json
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

KSA_RELATED=$(kubectl config view --minify -o jsonpath="{.clusters[0].cluster.server}" | sed 's#https://##g')

vault auth enable --path="${KSA_RELATED}" kubernetes

vault write auth/${KSA_RELATED}/config \
    token_reviewer_jwt="$VAULT_SA_JWT_TOKEN" \
    kubernetes_host=https://${K8S_HOST} \
    kubernetes_ca_cert="$VAULT_SA_CA_CRT"

```