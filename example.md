# Postgresql Example

This is an example of how to use this operator to protect postgresql databases

## Install vault-config-operator

TODO

## Install namespace-config-operator

TODO

## Deploy Vault

```shell
helm repo add hashicorp https://helm.releases.hashicorp.com
helm upgrade vault hashicorp/vault -i --create-namespace -n vault --atomic -f ./examples/postgresql/vault-values.yaml

HA_INIT_RESPONSE=$(oc exec vault-0 -n vault -- vault operator init -address https://vault-internal.vault.svc:8200 -ca-path /var/run/secrets/kubernetes.io/serviceaccount/service-ca.crt -format=json -key-shares 1 -key-threshold 1)

HA_UNSEAL_KEY=$(echo "$HA_INIT_RESPONSE" | jq -r .unseal_keys_b64[0])
HA_VAULT_TOKEN=$(echo "$HA_INIT_RESPONSE" | jq -r .root_token)

echo "$HA_UNSEAL_KEY"
echo "$HA_VAULT_TOKEN"

#here we are saving these variable in a secret, this is probably not what you should do in a production environment
oc delete secret vault-init -n vault
oc create secret generic vault-init -n vault --from-literal=unseal_key=${HA_UNSEAL_KEY} --from-literal=root_token=${HA_VAULT_TOKEN}
export HA_UNSEAL_KEY=$(oc get secret vault-init -n vault -o jsonpath='{.data.unseal_key}' | base64 -d )
export HA_VAULT_TOKEN=$(oc get secret vault-init -n vault -o jsonpath='{.data.root_token}' | base64 -d )
oc exec vault-0 -n vault -- vault operator unseal -address https://vault-internal.vault.svc:8200 -ca-path /var/run/secrets/kubernetes.io/serviceaccount/service-ca.crt $HA_UNSEAL_KEY
```

## Manual Vault scripting

```shell
export VAULT_ADDR=https://vault-vault.apps.control-cluster-raffa.demo.red-chesterfield.com/
export VAULT_TOKEN=$(oc get secret vault-init -n vault -o jsonpath='{.data.root_token}' | base64 -d )
oc new-project vault-admin
vault policy write -tls-skip-verify policy-admin  ./examples/postgresql/policy-admin-policy.hcl
vault policy write -tls-skip-verify kubernetes-role-admin  ./examples/postgresql/kubernetes-role-admin-policy.hcl
vault auth enable -tls-skip-verify kubernetes
export sa_secret_name=$(oc get sa default -n vault -o jsonpath='{.secrets[*].name}' | grep -o '\b\w*\-token-\w*\b')
oc get secret ${sa_secret_name} -n vault -o jsonpath='{.data.ca\.crt}' | base64 -d > /tmp/ca.crt
vault write -tls-skip-verify auth/kubernetes/config token_reviewer_jwt="$(oc serviceaccounts get-token vault -n vault)" kubernetes_host=https://kubernetes.default.svc:443 kubernetes_ca_cert=@/tmp/ca.crt
vault write -tls-skip-verify auth/kubernetes/role/policy-admin bound_service_account_names=default bound_service_account_namespaces=vault-admin policies=policy-admin,kubernetes-role-admin ttl=1h

## At this point the seed is completed. We are now going to login as the vault-admin/default service account and continue the configuration. Here starts the configuration that the operator would do.
VAULT_TOKEN=$(vault write -tls-skip-verify auth/kubernetes/login role=policy-admin jwt="$(oc serviceaccounts get-token default -n vault-admin)" -format=json | jq -r .auth.client_token)
#create policy
vault policy write -tls-skip-verify database-engine-admin  ./examples/postgresql/database-engine-admin-policy.hcl
# the operator will execute this so to include all the needed namespaces, here we have one
vault write -tls-skip-verify auth/kubernetes/role/database-engine-admin bound_service_account_names=default bound_service_account_namespaces=postgresql-vault-demo policies=database-engine-admin ttl=1h
# now we create the policy and role to read the secrets
vault policy write -tls-skip-verify database-creds-reader  ./examples/postgresql/database-creds-reader-policy.hcl
# and relative vault role (again the operator would propagate this to every needing namespace)
vault write -tls-skip-verify auth/kubernetes/role/database-creds-reader bound_service_account_names=default bound_service_account_namespaces=postgresql-vault-demo policies=database-creds-reader ttl=1h



## At this point we login as the namespace tenant to configure the postgresql engine
oc new-project postgresql-vault-demo
VAULT_TOKEN=$(vault write -tls-skip-verify auth/kubernetes/login role=database-engine-admin jwt="$(oc serviceaccounts get-token default -n postgresql-vault-demo)" -format=json | jq -r .auth.client_token)
vault secrets enable -tls-skip-verify -path postgresql-vault-demo/database database
vault write -tls-skip-verify postgresql-vault-demo/database/config/my-postgresql-database plugin_name=postgresql-database-plugin allowed_roles="read-write,read-only" connection_url="postgresql://{{username}}:{{password}}@my-postgresql-database.postgresql-vault-demo.svc:5432/" username="postgres" password="changeit" verify_connection="true"
vault write -tls-skip-verify postgresql-vault-demo/database/roles/read-only db_name=my-postgresql-database creation_statements="CREATE ROLE \"{{name}}\" WITH LOGIN PASSWORD '{{password}}' VALID UNTIL '{{expiration}}'; GRANT SELECT ON ALL TABLES IN SCHEMA public TO \"{{name}}\";" default_ttl="1h" max_ttl="24h"
vault write -tls-skip-verify postgresql-vault-demo/database/roles/read-write db_name=my-postgresql-database creation_statements="CREATE ROLE \"{{name}}\" WITH LOGIN PASSWORD '{{password}}' VALID UNTIL '{{expiration}}'; GRANT SELECT, UPDATE, INSERT, DELETE ON ALL TABLES IN SCHEMA public TO \"{{name}}\";" default_ttl="1h" max_ttl="24h"


## now we login as an application that wants to read the credentials
VAULT_TOKEN=$(vault write -tls-skip-verify auth/kubernetes/login role=database-creds-reader jwt="$(oc serviceaccounts get-token default -n postgresql-vault-demo)" -format=json | jq -r .auth.client_token)
vault read -tls-skip-verify postgresql-vault-demo/database/creds/read-write
```

## Create seed configuration in Vault and Kube

```shell
oc new-project vault-admin
vault policy write policy-admin ./examples/postgresql/policy-admin-policy.hcl
vault auth enable kubernetes
vault write auth/kubernetes/config token_reviewer_jwt=${jwt} kubernetes_host=https://kubernetes.default.svc:443 kubernetes_ca_cert=@ca.crt
vault write auth/kubernetes/role/demo bound_service_account_names=default bound_service_account_namespaces=vault-admin policies=policy-admin ttl=1h
oc apply -f ./example/postgresql/postgresql-enabled-namespace-config.yaml -n vault-admin
oc apply -f ./example/postgresql/postgresql-vault-role.yaml -n vault-admin
```

## Deploy postgresql

```shell
oc new-project postgresql-vault-demo
oc label namespace postgresql-vault-demo postgresql-enabled=true
oc create secret generic postgresql-admin-password --from-literal=postgresql-password=changeit -n postgresql-vault-demo
export uid=$(oc get project postgresql-vault-demo -o jsonpath='{.metadata.annotations.openshift\.io/sa\.scc\.uid-range}'|sed 's/\/.*//')
export guid=$(oc get project postgresql-vault-demo -o jsonpath='{.metadata.annotations.openshift\.io/sa\.scc\.supplemental-groups}'|sed 's/\/.*//')
helm upgrade my-postgresql-database bitnami/postgresql -i --create-namespace -n postgresql-vault-demo -f ./examples/postgresql/postgresql-values.yaml --set securityContext.fsGroup=${guid} --set containerSecurityContext.runAsUser=${uid} --set volumePermissions.securityContext.runAsUser=${uid} --set metrics.securityContext.runAsUser=${uid}
oc apply -f ./example/postgresql/postgresql-secret-engine.yaml
// create vault role for local read-write
// create vault role for exteranl read-only
```

## Deploy local client


## Deploy external client


## Clean up

### Delete Vault

```shell
helm uninstall vault -n vault
oc delete pvc --all -n vault
```
