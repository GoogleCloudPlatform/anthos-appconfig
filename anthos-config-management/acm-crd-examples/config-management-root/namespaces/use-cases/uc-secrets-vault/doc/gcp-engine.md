# Google Cloud Secrets Engine

Below is a minimal setup guide for configuring GCP Secrets Engine for Vault. Additional configuration options are available in the [HashiCorp Documentation](https://www.vaultproject.io/docs/secrets/gcp/index.html)

## Configure

Set some environment variables for your cluster and project:
```bash
export PROJECT_ID="<GCP project name>"
export CLUSTER_ID="<GKE cluster name>"
export SERVICE_ACCOUNT="acm-${CLUSTER_ID}-vault"
```

Create a Google Cloud service account, assign IAM roles necessary for dynamic credentials, and generate an access key:
```bash
gcloud iam service-accounts create ${SERVICE_ACCOUNT} \
  --display-name "acm ${CLUSTER_ID} vault service account" \
  --project ${PROJECT_ID}

gcloud iam service-accounts add-iam-policy-binding \
  --role roles/iam.serviceAccountAdmin \
  --member serviceAccount:${SERVICE_ACCOUNT}@${PROJECT_ID}.iam.gserviceaccount.com \
  ${SERVICE_ACCOUNT}@${PROJECT_ID}.iam.gserviceaccount.com

gcloud iam service-accounts keys create \
  --iam-account=${SERVICE_ACCOUNT}@${PROJECT_ID}.iam.gserviceaccount.com \
  ${SERVICE_ACCOUNT}-creds.json
```

Enable the Google Cloud secrets engine and provide the managing service account key configuration:
```bash
vault secrets enable gcp
vault write gcp/config credentials=@${SERVICE_ACCOUNT}-creds.json
```
