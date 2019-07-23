# uc-secrets-vault

This demo requires an existing [Vault installation](vault-server.md) with [Google Cloud Secrets Engine](gcp-engine.md) and [Kubernetes Auth Method](k8s-auth.md) enabled and configured.

## Configure ACM

Set environment variables for your Vault configuration:
```bash
export VAULT_ADDR=https://vault.default.svc.cluster-domain.example:8200
export VAULT_CACERT=/path/to/vault/ca.crt
```

Create a service account to be used for ACM Operator Vault access:
```bash
kubectl create serviceaccount -n appconfigmgrv2-system acm-vault
```

Provide Vault location, CA cert, and service account token details to ACM:
```bash
ACM_VAULT_SA_SECRET=$(kubectl get -n appconfigmgrv2-system sa acm-vault -o jsonpath="{.secrets[*]['name']}")
ACM_VAULT_SA_JWT_TOKEN=$(kubectl get -n appconfigmgrv2-system secret $ACM_VAULT_SA_SECRET -o jsonpath="{.data.token}" | base64 --decode; echo)

kubectl create configmap vault \
  --namespace=appconfigmgrv2-system \
  --from-literal vault-addr=${VAULT_ADDR} \
  --from-literal serviceaccount-jwt="${ACM_VAULT_SA_JWT_TOKEN}"

kubectl create secret generic vault-ca \
  --namespace=appconfigmgrv2-system \
  --from-file=ca.pem
```

Configure a Vault access policy for the ACM service account:
```bash
cat > acm-controller-policy.hcl <<EOF
path "sys/policy/acm-*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}
path "auth/kubernetes/role/acm-*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}
EOF

vault policy write acm-controller ./acm-controller-policy.hcl
```

Lastly, create a Vault Kubernetes Auth role and attach policy:
```bash
vault write auth/kubernetes/role/acm-vault \
    bound_service_account_names=acm-vault \
    bound_service_account_namespaces=appconfigmgrv2-system \
    policies=acm-controller \
    ttl=1h
```
