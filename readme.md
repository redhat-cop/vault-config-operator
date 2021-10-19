# Vault Config Operator

This operator helps set up Vault Configurations. The main intent is to do so such that subsequently pods can consume the secrets made available.
There are two main principles through all of the capabilities of this operator:

1. high-fidelity API. The CRD exposed by this operator reflect field by field the Vault APIs. This is because we don't want to make any assumption on the kinds of configuration workflow that user will set up. That being said the Vault API is very extensive and we are starting with enough API coverage to support, we think, some simple and very common configuration workflows.
2. attention to security (after all we are integrating with a security tool). To prevent credential leaks we give no permissions to the operator itself against Vault. All APIs exposed by this operator contains enough information to authenticate to Vault using a local service account (local to the namespace where the API exist). In other word for a namespace user to be abel to successfully configure Vault, a service account in that namespace must have been previously given the needed Vault permissions.

Currently this operator supports the following CRDs:

1. [Policy](#policy) Configures Vault [Policies](https://www.vaultproject.io/docs/concepts/policies)
2. [KubernetesAuthEngineRole](#KubernetesAuthEngineRole) Configures a Vault [Kubernetes Authentication](https://www.vaultproject.io/docs/auth/kubernetes) Role
3. [SecretEngineMount](#SecretEngineMount) Configures a Mount point for a [SecretEngine](https://www.vaultproject.io/docs/secrets)
4. [DatabaseSecretEngineConfig](#DatabaseSecretEngineConfig) Configures a [Database Secret Engine](https://www.vaultproject.io/docs/secrets/databases) Connection
5. [DatabaseSecretEngineRole](#DatabaseSecretEngineRole) Configures a [Database Secret Engine](https://www.vaultproject.io/docs/secrets/databases) Role
6. [RandomSecret](#RandomSecret) Creates a random secret in a vault [kv Secret Engine](https://www.vaultproject.io/docs/secrets/kv) with one password field generated using a [PasswordPolicy](https://www.vaultproject.io/docs/concepts/password-policies)

## The Authentication Section

As discussed each API has an Authentication Section that specify how to authenticate to Vault. Here is an example:

```yaml
  authentication: 
    path: kubernetes
    role: policy-admin
    namespace: tenant-namespace
    serviceAccount:
      name: vaultsa
```

The `path` field specifies the path at which the Kubernetes authentication role is mounted.

The `role` field specifies which role to request when authenticating

The `namespace` field specifies the Vault namespace (not related to Kubernetes namespace) to use. This is optional.

The `serviceAccount.name` specifies the token of which service account to use during the authentication process.

So the above configuration roughly correspond to the following command:

```shell
vault write [tenant-namespace/]auth/kubernetes/login role=policy-admin jwt=<vaultsa jwt token>
```

## Policy

The `Policy` CRD allows a user to create a [Vault Policy], here is an example:

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: Policy
metadata:
  name: database-creds-reader
spec:
  authentication: 
    path: kubernetes
    role: policy-admin
  policy: |
    # Configure read secrets
    path "/{{identity.entity.aliases.auth_kubernetes_804f1655.metadata.service_account_namespace}}/database/creds/+" {
      capabilities = ["read"]
    }
```

Notice that in this policy we have parametrized the path based on the namespace of the connecting service account.

## KubernetesAuthEngineRole

The `KubernetesAuthEngineRole` creates a Vault Authentication Role for a Kubernetes Authentication mount, here is an example:

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: KubernetesAuthEngineRole
metadata:
  name: database-engine-admin
spec:
  authentication: 
    path: kubernetes
    role: policy-admin
  path: kubernetes  
  policies:
    - database-engine-admin
  targetServiceAccounts: 
  - vaultsa  
  targetNamespaceSelector:
    matchLabels:
      postgresql-enabled: "true"
```

The `path` field specifies the path of the Kubernetes Authentication Mount at which the role will be mounted.

The `policies` field specifies which Vault policies will be associated with this role.

The `targetServiceAccounts` field specifies which service accounts can authenticate. If not specified, it defaults to `default`.

The `targetNamespaceSelector` field specifies from which kubernetes namespaces it is possible to authenticate. Notice as the set of namespaces selected by the selector varies, this configuration will be updated. It is also possible to specify a static set of namespaces.

Many other standard Kubernetes Authentication Role fields are available for fine tuning, see the [Vault Documentation](https://www.vaultproject.io/api-docs/auth/kubernetes#create-role)

This CR is roughly equivalent to this Vault CLI command:

```shell
vault write [namespace/]auth/kubernetes/role/database-engine-admin bound_service_account_names=vaultsa bound_service_account_namespaces=<dynamically generated> policies=database-engine-admin
```

## SecretEngineMount

The `SecretEngineMount` CRD allows a user to create a Secret Engine mount point, here is an example:

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: SecretEngineMount
metadata:
  name: database
spec:
  authentication: 
    path: kubernetes
    role: database-engine-admin
  type: database
  path: postgresql-vault-demo
```

The `type` field specifies the secret engine type.

The `path` field specifies the path at which to mount the secret engine

Many other standard Secret Engine Mount fields are available for fine tuning, see the [Vault Documentation](https://www.vaultproject.io/api-docs/system/mounts#enable-secrets-engine)

This CR is roughly equivalent to this Vault CLI command:

```shell
vault secrets enable -path [namespace/]postgresql-vault-demo/database database
```

## DatabaseSecretEngineConfig

`DatabaseSecretEngineConfig` CRD allows a user to create a Database Secret Engine configuration, also called connection for an existing Database Secret Engine Mount. Here is an example

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: DatabaseSecretEngineConfig
metadata:
  name: my-postgresql-database
spec:
  authentication: 
    path: kubernetes
    role: database-engine-admin
  pluginName: postgresql-database-plugin
  allowedRoles:
    - read-write
    - read-only
  connectionURL: postgresql://{{username}}:{{password}}@my-postgresql-database.postgresql-vault-demo.svc:5432
  username: admin
  rootCredentialsFromSecret:
    name: postgresql-admin-password
  path: postgresql-vault-demo/database
```

The `pluginName` field specifies what type of database this connection is for.

The `allowedRoles` field specifies which role names can be created for this connection.

The `connectionURL` field specifies how to connect to the database.

The `username` field specific the username to be used to connect to the database. This field is optional, if not specified the username will be retrieved from teh credential secret.

The `path` field specifies the path of the secret engine to which this connection will be added.

The password and possibly the username can be retrived a three different ways:

1. From a Kubernetes secret, specifying the `rootCredentialsFromSecret` field. The secret must be of [basic auth type](https://kubernetes.io/docs/concepts/configuration/secret/#basic-authentication-secret). If the secret is updated this connection will also be updated.
2. From a Vault secret, specifying the `rootCredentialsFromVaultSecret` field.
3. From a [RandomSecret](#RandomSecret), specifying the `rootCredentialsFromRandomSecret` field. When the RandomSecret generates a new secret, this connection will also be updated.

Many other standard Database Secret Engine Config fields are available for fine tuning, see the [Vault Documentation](https://www.vaultproject.io/api-docs/secret/databases#configure-connection)

This CR is roughly equivalent to this Vault CLI command:

```shell
vault write [namespace/]postgresql-vault-demo/database/config/my-postgresql-database plugin_name=postgresql-database-plugin allowed_roles="read-write,read-only" connection_url="postgresql://{{username}}:{{password}}@my-postgresql-database.postgresql-vault-demo.svc:5432/" username=<retrieved dynamically> password=<retrieved dynamically>
```

## DatabaseSecretEngineRole

The `DatabaseSecretEngineRole` CRD allows a user to create a Database Secret Engine Role, here is an example:

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: DatabaseSecretEngineRole
metadata:
  name: read-only
spec:
  authentication: 
    path: kubernetes
    role: database-engine-admin
  path: postgresql-vault-demo/database
  dBName: my-postgresql-database
  creationStatements:
    - CREATE ROLE "{{name}}" WITH LOGIN PASSWORD '{{password}}' VALID UNTIL '{{expiration}}'; GRANT SELECT ON ALL TABLES IN SCHEMA public TO "{{name}}";
```

The `path` field specifies the path of the secret engine that will contain this role.

The `dBname` field specifies the name of the connection to be used with this role.

The `creationStatements` field specifies the statements to run to create a new account.

Many other standard Database Secret Engine Role fields are available for fine tuning, see the [Vault Documentation](https://www.vaultproject.io/api-docs/secret/databases#create-role)

This CR is roughly equivalent to this Vault CLI command:

```shell
vault write [namespace/]postgresql-vault-demo/database/roles/read-only db_name=my-postgresql-database creation_statements="CREATE ROLE \"{{name}}\" WITH LOGIN PASSWORD '{{password}}' VALID UNTIL '{{expiration}}'; GRANT SELECT ON ALL TABLES IN SCHEMA public TO \"{{name}}\";"
```

## RandomSecret

The RandomSecret CRD allows a user to generate a random secret (normally a password) and store it in Vault with a given Key. The generated secret will be compliant with a Vault [Password Policy], here is an example:

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: RandomSecret
metadata:
  name: my-postgresql-admin-password
spec:
  authentication: 
    path: kubernetes
    role: database-engine-admin
  path: kv/vault-tenant
  secretKey: password
  secretFormat:
    passwordPolicyName: my-complex-password-format
  refreshPeriod: 1h
```

The `path` field specifies the path at which the secret will be written, it must correspond to a kv Secret Engine mount.

The `secretKey` field is the key of the secret.

The `secretFormat` is a reference to a Vault Password policy, it can also supplied inline.

The `refreshPeriod` specifies the frequency at which this secret will be regenerated. This is an optional field, if not specified the secret will be generated once and then never updated.

With a RandomSecret it is possible to build workflow in which the root password of a resource that we need to protect is never stored anywhere, except in vault. One way to achieve this is to have a random secret seed the root password. Then crete an operator that watches the RandomSecret and retrieves ths generated secret from vault and updates the resource to be protected. Finally configure the Secret Engine object to watch for the RandomSecret updates.

This CR is roughly equivalent to this Vault CLI command:

```shell
vault kv put [namespace/]kv/vault-tenant password=<generated value>
```

## Metrics

Prometheus compatible metrics are exposed by the Operator and can be integrated into OpenShift's default cluster monitoring. To enable OpenShift cluster monitoring, label the namespace the operator is deployed in with the label `openshift.io/cluster-monitoring="true"`.

```shell
oc label namespace <namespace> openshift.io/cluster-monitoring="true"
```

### Testing metrics

```sh
export operatorNamespace=vault-config-operator-local # or vault-config-operator
oc label namespace ${operatorNamespace} openshift.io/cluster-monitoring="true"
oc rsh -n openshift-monitoring -c prometheus prometheus-k8s-0 /bin/bash
export operatorNamespace=vault-config-operator-local # or vault-config-operator
curl -v -s -k -H "Authorization: Bearer $(cat /var/run/secrets/kubernetes.io/serviceaccount/token)" https://vault-config-operator-controller-manager-metrics.${operatorNamespace}.svc.cluster.local:8443/metrics
exit
```

## Deploying the Operator

This is a cluster-level operator that you can deploy in any namespace, `vault-config-operator` is recommended.

It is recommended to deploy this operator via [`OperatorHub`](https://operatorhub.io/), but you can also deploy it using [`Helm`](https://helm.sh/).

### Multiarch Support

| Arch  | Support  |
|:-:|:-:|
| amd64  | ✅ |
| arm64  | ✅  |
| ppc64le  | ✅  |
| s390x  | ❌  |

### Deploying from OperatorHub

> **Note**: This operator supports being installed disconnected environments

If you want to utilize the Operator Lifecycle Manager (OLM) to install this operator, you can do so in two ways: from the UI or the CLI.

#### Deploying from OperatorHub UI

* If you would like to launch this operator from the UI, you'll need to navigate to the OperatorHub tab in the console. Before starting, make sure you've created the namespace that you want to install this operator to with the following:

```shell
oc new-project vault-config-operator
```

* Once there, you can search for this operator by name: `vault config operator`. This will then return an item for our operator and you can select it to get started. Once you've arrived here, you'll be presented with an option to install, which will begin the process.
* After clicking the install button, you can then select the namespace that you would like to install this to as well as the installation strategy you would like to proceed with (`Automatic` or `Manual`).
* Once you've made your selection, you can select `Subscribe` and the installation will begin. After a few moments you can go ahead and check your namespace and you should see the operator running.

![Cert Utils Operator](./media/vault-config-operator.png)

#### Deploying from OperatorHub using CLI

If you'd like to launch this operator from the command line, you can use the manifests contained in this repository by running the following:

oc new-project vault-config-operator

```shell
oc apply -f config/operatorhub -n vault-config-operator
```

This will create the appropriate OperatorGroup and Subscription and will trigger OLM to launch the operator in the specified namespace.

### Deploying with Helm

Here are the instructions to install the latest release with Helm.

```shell
oc new-project vault-config-operator
helm repo add vault-config-operator https://redhat-cop.github.io/vault-config-operator
helm repo update
helm install vault-config-operator vault-config-operator/vault-config-operator
```

This can later be updated with the following commands:

```shell
helm repo update
helm upgrade vault-config-operator vault-config-operator/vault-config-operator
```

## Development

## Running the operator locally

### Deploy a Vault instance

If you don't have a Vault instance available for testing, deploy one with these steps:

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

### Configure an Kubernetes Authentication mount point

All the configuration made by the operator need to authenticate via a Kubernetes Authentication. So you need a root Kubernetes Authentication mount point and role. The you can create more roles via the operator.
If you don't have a root mount point and role, you can create them as follows:

```shell
oc new-project vault-admin
export cluster_base_domain=$(oc get dns cluster -o jsonpath='{.spec.baseDomain}')
export VAULT_ADDR=https://vault-vault.apps.${cluster_base_domain}
export VAULT_TOKEN=$(oc get secret vault-init -n vault -o jsonpath='{.data.root_token}' | base64 -d )
# this policy is intentionally broad to allow to test anything in Vault. In a real life scenario this policy would be scoped down.
vault policy write -tls-skip-verify vault-admin  ./config/local-development/vault-admin-policy.hcl
vault auth enable -tls-skip-verify kubernetes
export sa_secret_name=$(oc get sa default -n vault -o jsonpath='{.secrets[*].name}' | grep -o '\b\w*\-token-\w*\b')
oc get secret ${sa_secret_name} -n vault -o jsonpath='{.data.ca\.crt}' | base64 -d > /tmp/ca.crt
vault write -tls-skip-verify auth/kubernetes/config token_reviewer_jwt="$(oc serviceaccounts get-token vault -n vault)" kubernetes_host=https://kubernetes.default.svc:443 kubernetes_ca_cert=@/tmp/ca.crt
vault write -tls-skip-verify auth/kubernetes/role/policy-admin bound_service_account_names=default bound_service_account_namespaces=vault-admin policies=vault-admin ttl=1h
export accessor=$(vault read -tls-skip-verify -format json sys/auth | jq -r '.data["kubernetes/"].accessor')
```

### Run the operator

```shell
export repo=raffaelespazzoli #replace with yours, this has also to be replaced in the following files: Tiltfile, ./config/local-development/tilt/replace-image.yaml. Further improvements may be able to remove this constraint.
docker login quay.io/$repo
oc new-project vault-config-operator
tilt up
```

### Test Manually

Policy

```shell
envsubst < ./test/database-engine-admin-policy.yaml | oc apply -f - -n vault-admin
```

Vault Role

```shell
oc new-project test-vault-config-operator
oc label namespace test-vault-config-operator database-engine-admin=true
oc apply -f ./test/database-engine-admin-role.yaml -n vault-admin
```

Secret Engine Mount

```shell
oc apply -f ./test/database-secret-engine.yaml -n test-vault-config-operator
```

Database secret engine connection. This will deploy a postgresql database to connect to

```shell
oc create secret generic postgresql-admin-password --from-literal=postgresql-password=changeit -n test-vault-config-operator
export uid=$(oc get project test-vault-config-operator -o jsonpath='{.metadata.annotations.openshift\.io/sa\.scc\.uid-range}'|sed 's/\/.*//')
export guid=$(oc get project test-vault-config-operator -o jsonpath='{.metadata.annotations.openshift\.io/sa\.scc\.supplemental-groups}'|sed 's/\/.*//')
helm upgrade my-postgresql-database bitnami/postgresql -i --create-namespace -n test-vault-config-operator -f ./examples/postgresql/postgresql-values.yaml --set securityContext.fsGroup=${guid} --set containerSecurityContext.runAsUser=${uid} --set volumePermissions.securityContext.runAsUser=${uid} --set metrics.securityContext.runAsUser=${uid}
oc apply -f ./test/database-engine-config.yaml -n test-vault-config-operator
```

Database Secret engine role

```shell
oc apply -f ./test/database-engine-read-only-role.yaml -n test-vault-config-operator
```

RandomSecret

```shell
vault write -tls-skip-verify /sys/policies/password/simple-password-policy policy=@./test/password-policy.hcl
envsubst < ./test/kv-engine-admin-policy.yaml | oc apply -f - -n vault-admin
envsubst < ./test/secret-writer-policy.yaml | oc apply -f - -n vault-admin
oc apply -f ./test/kv-engine-admin-role.yaml -n vault-admin
oc apply -f ./test/secret-writer-role.yaml -n vault-admin
oc apply -f ./test/kv-secret-engine.yaml -n test-vault-config-operator
oc apply -f ./test/random-secret.yaml -n test-vault-config-operator
```

### Test helm chart locally

Define an image and tag. For example...

```shell
export imageRepository="quay.io/redhat-cop/vault-config-operator"
export imageTag="$(git -c 'versionsort.suffix=-' ls-remote --exit-code --refs --sort='version:refname' --tags https://github.com/redhat-cop/vault-config-operator.git '*.*.*' | tail --lines=1 | cut --delimiter='/' --fields=3)"
```

Deploy chart...

```shell
make helmchart IMG=${imageRepository} VERSION=${imageTag}
helm upgrade -i vault-config-operator-local charts/vault-config-operator -n vault-config-operator-local --create-namespace
```

Delete...

```shell
helm delete vault-config-operator-local -n vault-config-operator-local
kubectl delete -f charts/vault-config-operator/crds/crds.yaml
```

## Building/Pushing the operator image

```shell
export repo=raffaelespazzoli #replace with yours
docker login quay.io/$repo
make docker-build IMG=quay.io/$repo/vault-config-operator:latest
make docker-push IMG=quay.io/$repo/vault-config-operator:latest
```

## Deploy to OLM via bundle

```shell
make manifests
make bundle IMG=quay.io/$repo/vault-config-operator:latest
operator-sdk bundle validate ./bundle --select-optional name=operatorhub
make bundle-build BUNDLE_IMG=quay.io/$repo/vault-config-operator-bundle:latest
docker push quay.io/$repo/vault-config-operator-bundle:latest
operator-sdk bundle validate quay.io/$repo/vault-config-operator-bundle:latest --select-optional name=operatorhub
oc new-project vault-config-operator
oc label namespace vault-config-operator openshift.io/cluster-monitoring="true"
operator-sdk cleanup vault-config-operator -n vault-config-operator
operator-sdk run bundle --install-mode AllNamespaces -n vault-config-operator quay.io/$repo/vault-config-operator-bundle:latest
```

## Releasing

```shell
git tag -a "<tagname>" -m "<commit message>"
git push upstream <tagname>
```

If you need to remove a release:

```shell
git tag -d <tagname>
git push upstream --delete <tagname>
```

If you need to "move" a release to the current main

```shell
git tag -f <tagname>
git push upstream -f <tagname>
```

### Cleaning up

```shell
operator-sdk cleanup vault-config-operator -n vault-config-operator
oc delete operatorgroup operator-sdk-og
oc delete catalogsource vault-config-operator-catalog
```
