# End to end  example

In this scenario, we explain how to use this operator to address the following requirements:

1. Each team will have a dedicated kubernetes auth endpoint. Each team may use multiple namespaces for their environments. Each namespace is labeled with a the `team` name and the namespace `environment` (dev, qa, prod...).
2. The path for such created auth endpoint should be `auth/cluster1/{team}-kubernetes`. We assume here we are working in `cluster1` and this Vault instance is serving multiple clusters.
3. Teams are allowed to create secret engines of type `database` and `kv`, the ones currently supported.
4. the path for these engines should be `{team}/{engine_name}`

## Install vault-config-operator

see [here](./../readme.md##deploying-the-operator)

## Install namespace-config-operator

see [here](https://github.com/redhat-cop/namespace-configuration-operator#deploying-the-operator)

## Deploy Vault

```shell
helm repo add hashicorp https://helm.releases.hashicorp.com
export cluster_base_domain=$(oc get dns cluster -o jsonpath='{.spec.baseDomain}')
envsubst < ./config/local-development/vault-values.yaml > /tmp/values
helm upgrade vault hashicorp/vault -i --create-namespace -n vault --atomic -f /tmp/values

INIT_RESPONSE=$(oc exec vault-0 -n vault -- vault operator init -address https://vault.vault.svc:8200 -ca-path /var/run/secrets/kubernetes.io/serviceaccount/service-ca.crt -format=json -key-shares 1 -key-threshold 1)

UNSEAL_KEY=$(echo "$INIT_RESPONSE" | jq -r .unseal_keys_b64[0])
ROOT_TOKEN=$(echo "$INIT_RESPONSE" | jq -r .root_token)

echo "$UNSEAL_KEY"
echo "$ROOT_TOKEN"

#here we are saving these variable in a secret, this is probably not what you should do in a production environment
oc delete secret vault-init -n vault
oc create secret generic vault-init -n vault --from-literal=unseal_key=${UNSEAL_KEY} --from-literal=root_token=${ROOT_TOKEN}
export UNSEAL_KEY=$(oc get secret vault-init -n vault -o jsonpath='{.data.unseal_key}' | base64 -d )
export ROOT_TOKEN=$(oc get secret vault-init -n vault -o jsonpath='{.data.root_token}' | base64 -d )
oc exec vault-0 -n vault -- vault operator unseal -address https://vault.vault.svc:8200 -ca-path /var/run/secrets/kubernetes.io/serviceaccount/service-ca.crt $UNSEAL_KEY
```

## Prepare Vault for Kubernetes

In this phase a vault admin prepares Vault for Kubernetes. To keep this simple we are going to give very wide permissions to a single account in the vault-admin namespace.

```shell
oc new-project vault-admin
oc adm policy add-cluster-role-to-user system:auth-delegator -z default -n vault-admin
export cluster_base_domain=$(oc get dns cluster -o jsonpath='{.spec.baseDomain}')
export VAULT_ADDR=https://vault-vault.apps.${cluster_base_domain}
export VAULT_TOKEN=$(oc get secret vault-init -n vault -o jsonpath='{.data.root_token}' | base64 -d )
# this policy is intentionally broad to allow to test anything in Vault. In a real life scenario this policy would be scoped down.
vault policy write -tls-skip-verify vault-admin  ./config/local-development/vault-admin-policy.hcl
vault auth enable -tls-skip-verify -path cluster1-admin kubernetes 
export sa_secret_name=$(oc get sa default -n vault -o jsonpath='{.secrets[*].name}' | grep -o '\b\w*\-token-\w*\b')
oc get secret ${sa_secret_name} -n vault -o jsonpath='{.data.ca\.crt}' | base64 -d > /tmp/ca.crt
vault write -tls-skip-verify auth/cluster1-admin/config token_reviewer_jwt="$(oc serviceaccounts get-token vault -n vault)" kubernetes_host=https://kubernetes.default.svc:443 kubernetes_ca_cert=@/tmp/ca.crt
vault write -tls-skip-verify auth/cluster1-admin/role/vault-admin bound_service_account_names=default bound_service_account_namespaces=vault-admin policies=vault-admin ttl=1h
export accessor=$(vault read -tls-skip-verify -format json sys/auth | jq -r '.data["cluster1-admin/"].accessor')
```

## Vault-admin configuration

In this phase the vault-admin tenant deploys a set of configuration to set up the kubernetes auth for each namespace.
Using the namespace-config-operator, we ensure that for each namespace with label `environment: dev` we create the following objects:

1. a kubernetes authentication endpoint at the path `auth/cluster1/{team}-kubernetes`
2. a secret engine admin policy customized for that specific team, in which secret engines can only be created under the path: `/cluster1/{{ team }}`
3. a secret reader policy customized for that specific team, in which secrets can only be read under the path: `/cluster1/{{ team }}`
4. a vault role associated with the above authentication endpoint that provides the secret engine admin policy and will be allowed to the `default` service account of all the namespaces belonging to the `team`
5. a vault role associated with the above authentication endpoint that provides the secret reader policy and will be allowed to the `default` service account of all the namespaces belonging to the `team`

```shell
oc apply -f ./docs/examples/postgresql/namespace-config.yaml
```

## Tenant configuration

In this phase, as an example, the tenant will deploy a postgresql database and create a corresponding authentication engine with two roles: a read/write and a read only role.

Let's create the Vault database first

```shell
oc new-project team-a-dev
oc label namespace team-a-dev environment=dev team=team-a
oc create secret generic postgresql-admin-password --from-literal=postgresql-password=changeit -n team-a-dev
export uid=$(oc get project team-a-dev -o jsonpath='{.metadata.annotations.openshift\.io/sa\.scc\.uid-range}'|sed 's/\/.*//')
export guid=$(oc get project team-a-dev -o jsonpath='{.metadata.annotations.openshift\.io/sa\.scc\.supplemental-groups}'|sed 's/\/.*//')
helm upgrade my-postgresql-database bitnami/postgresql -i --create-namespace -n team-a-dev -f ./docs/examples/postgresql/postgresql-values.yaml --set securityContext.fsGroup=${guid} --set containerSecurityContext.runAsUser=${uid} --set volumePermissions.securityContext.runAsUser=${uid} --set metrics.securityContext.runAsUser=${uid}
```

The next command mounts the database secret engine at: `cluster1/team-a/postregsql`, then creates a configuration to point to the database that we just installed, finally it creates the two roles.

```shell
oc apply -f ./docs/examples/postgresql/postgresql-secret-engine.yaml -n team-a-dev
```

At this point anyone with access to the role's path can create database credentials. As a matter of proof, you can run the following commands:

```shell
VAULT_TOKEN=$(vault write -tls-skip-verify auth/cluster1/team-a-kubernetes/login role=team-a-secret-reader jwt="$(oc serviceaccounts get-token default -n team-a-dev)" -format=json | jq -r .auth.client_token)
vault read -tls-skip-verify cluster1/team-a/postgresql/creds/read-write
```

you should see an output that looks like this:

```shell
Key                Value
---                -----
lease_id           cluster1/team-a/postgresql/creds/read-write/G7Aqsr2ZGGgsPunNMMBrpXNC
lease_duration     1h
lease_renewable    true
password           3CW-1J39hBlRghssjGpD
username           v-cluster1-read-wri-xmTaI6G7c2BxecIFenL5-1635017821
```