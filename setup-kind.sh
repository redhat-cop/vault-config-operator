#!/bin/bash

set -ex

go install sigs.k8s.io/kind@v0.11.1

kind delete cluster

kind create cluster

kubectl create namespace vault
kubectl apply -f integration/rolebinding-admin.yaml -n vault
helm repo add hashicorp https://helm.releases.hashicorp.com
helm upgrade vault hashicorp/vault -i --create-namespace -n vault --atomic -f ./integration/vault-values.yaml

kubectl wait --for=condition=ready pod/vault-0 -n vault --timeout=5m

kubectl port-forward pod/vault-0 8200:8200 -n vault > /dev/null 2>&1 & 

jobs -l

# kubectl create namespace vault-admin
export VAULT_ADDR=http://localhost:8200
export VAULT_TOKEN=$(kubectl get secret vault-init -n vault -o jsonpath='{.data.root_token}' | base64 -d )
# this policy is intentionally broad to allow to test anything in Vault. In a real life scenario this policy would be scoped down.
vault policy write -tls-skip-verify vault-admin  ./config/local-development/vault-admin-policy.hcl
vault auth enable -tls-skip-verify kubernetes
export sa_secret_name=$(kubectl get sa default -n vault -o jsonpath='{.secrets[*].name}' | grep -o '\b\w*\-token-\w*\b')
kubectl get secret ${sa_secret_name} -n vault -o jsonpath='{.data.ca\.crt}' | base64 -d > /tmp/ca.crt
vault write -tls-skip-verify auth/kubernetes/config token_reviewer_jwt="$(kubectl get $(kubectl get secret -n vault -o name | grep vault-token) -n vault -o jsonpath={.data.token} | base64 -d)" kubernetes_host=https://kubernetes.default.svc:443 kubernetes_ca_cert=@/tmp/ca.crt
vault write -tls-skip-verify auth/kubernetes/role/policy-admin bound_service_account_names=default bound_service_account_namespaces=vault-admin policies=vault-admin ttl=1h
export accessor=$(vault read -tls-skip-verify -format json sys/auth | jq -r '.data["kubernetes/"].accessor')

make integration ACCESSOR=${accessor} 
