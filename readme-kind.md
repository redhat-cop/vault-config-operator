# kind integration testing

> [kind](https://kind.sigs.k8s.io/)

## create kind cluster

```sh
go install sigs.k8s.io/kind@v0.11.1

kind create cluster
```

## dashboard (optional)

```sh
helm repo add kubernetes-dashboard https://kubernetes.github.io/dashboard/
helm repo update
helm install dashboard kubernetes-dashboard/kubernetes-dashboard -n kubernetes-dashboard --create-namespace

kubectl proxy

kubectl apply -f integration/serviceaccount-admin-user.yaml -n kubernetes-dashboard

kubectl describe serviceaccount admin-user -n kubernetes-dashboard

kubectl get $(kubectl get secret -n kubernetes-dashboard -o name | grep admin-user-token) -n kubernetes-dashboard -o jsonpath={.data.token} | base64 -d

# copy the token 

echo "open me and paste token http://localhost:8001/api/v1/namespaces/kubernetes-dashboard/services/https:dashboard-kubernetes-dashboard:https/proxy/#/login"

```

## vault

Install

```sh
kubectl apply -f integration/rolebinding-admin.yaml -n vault
helm repo add hashicorp https://helm.releases.hashicorp.com
helm upgrade vault hashicorp/vault -i --create-namespace -n vault --atomic -f ./integration/vault-values.yaml
```

```sh

kubectl port-forward pod/vault-0 8200:8200 -n vault

# kubectl create namespace vault-admin
export VAULT_ADDR=http://localhost:8200
export VAULT_TOKEN=$(kubectl get secret vault-init -n vault -o jsonpath='{.data.root_token}' | base64 -d )
# this policy is intentionally broad to allow to test anything in Vault. In a real life scenario this policy would be scoped down.
vault policy write -tls-skip-verify vault-admin  ./config/local-development/vault-admin-policy.hcl
vault auth enable -tls-skip-verify kubernetes
export sa_secret_name=$(kubectl get sa default -n vault -o jsonpath='{.secrets[*].name}' | grep -o '\b\w*\-token-\w*\b')
oc get secret ${sa_secret_name} -n vault -o jsonpath='{.data.ca\.crt}' | base64 -d > /tmp/ca.crt
vault write -tls-skip-verify auth/kubernetes/config token_reviewer_jwt="$(kubectl get $(kubectl get secret -n vault -o name | grep vault-token) -n vault -o jsonpath={.data.token} | base64 -d)" kubernetes_host=https://kubernetes.default.svc:443 kubernetes_ca_cert=@/tmp/ca.crt
vault write -tls-skip-verify auth/kubernetes/role/policy-admin bound_service_account_names=default bound_service_account_namespaces=vault-admin policies=vault-admin ttl=1h
export accessor=$(vault read -tls-skip-verify -format json sys/auth | jq -r '.data["kubernetes/"].accessor')
```

> TODO

```sh
helm delete vault -n vault --wait
kubectl delete pvc data-vault-0 -n vault
```


