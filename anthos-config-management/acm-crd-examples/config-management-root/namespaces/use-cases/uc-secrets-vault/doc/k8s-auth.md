# Kubernetes Auth Method

Below is a minimal setup guide for configuring the Kubernetes Auth Method for Vault. Additional configuration options are available in the [HashiCorp Documentation](https://www.vaultproject.io/docs/auth/kubernetes.html)

## Configure

Create a Kubernetes service account to validate service account tokens to Vault:
```bash
kubectl create serviceaccount vault-auth

cat > vault-auth-rbac.yaml << EOF
  apiVersion: rbac.authorization.k8s.io/v1beta1
  kind: ClusterRoleBinding
  metadata:
    name: role-tokenreview-binding
    namespace: default
  roleRef:
    apiGroup: rbac.authorization.k8s.io
    kind: ClusterRole
    name: system:auth-delegator
  subjects:
  - kind: ServiceAccount
    name: vault-auth
    namespace: default
EOF

kubectl apply -f vault-auth-rbac.yaml
```

Retrieve the service account JWT token and Kubernetes certificate
```bash
VAULT_SA_SECRET=$(kubectl get -n uc-secrets-vault-server sa vault-auth -o jsonpath="{.secrets[*]['name']}")
VAULT_SA_JWT_TOKEN=$(kubectl get -n uc-secrets-vault-server secret $VAULT_SA_SECRET -o jsonpath="{.data.token}" | base64 --decode; echo)
VAULT_SA_CA_CRT=$(kubectl get -n uc-secrets-vault-server secret $VAULT_SA_SECRET -o jsonpath="{.data['ca\.crt']}" | base64 --decode; echo)
```

Lastly, configure the Vault Kubernetes Auth method to use this service account for authentication
```bash
vault auth enable kubernetes

vault write auth/kubernetes/config \
    token_reviewer_jwt="$VAULT_SA_JWT_TOKEN" \
    kubernetes_host=https://${K8S_HOST} \
    kubernetes_ca_cert="$VAULT_SA_CA_CRT"
```
