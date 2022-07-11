![build status](https://github.com/redhat-cop/vault-config-operator/workflows/push/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/redhat-cop/vault-config-operator)](https://goreportcard.com/report/github.com/redhat-cop/vault-config-operator)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/redhat-cop/vault-config-operator)
[![CRD Docs](https://img.shields.io/badge/CRD-Docs-brightgreen)](https://doc.crds.dev/github.com/redhat-cop/vault-config-operator)

- [Vault Config Operator](#vault-config-operator)
  - [Authentication Engines](#authentication-engines)
  - [Policy management](#policy-management)
  - [Secret Engines](#secret-engines)
  - [Secret Management](#secret-management)
  - [The common authentication section](#the-common-authentication-section)
  - [End to end example](#end-to-end-example)
  - [Contributing a new Vault type](#contributing-a-new-vault-type)
  - [Initializing the connection to Vault](#initializing-the-connection-to-vault)
  - [Deploying the Operator](#deploying-the-operator)
    - [Multiarch Support](#multiarch-support)
    - [Deploying from OperatorHub](#deploying-from-operatorhub)
      - [Deploying from OperatorHub UI](#deploying-from-operatorhub-ui)
      - [Deploying from OperatorHub using CLI](#deploying-from-operatorhub-using-cli)
    - [Deploying with Helm](#deploying-with-helm)
  - [Metrics](#metrics)
    - [Testing metrics](#testing-metrics)
  - [Development](#development)
    - [Setup](#setup)
    - [Running the operator locally](#running-the-operator-locally)
      - [Deploy a Vault instance](#deploy-a-vault-instance)
      - [Configure an Kubernetes Authentication mount point](#configure-an-kubernetes-authentication-mount-point)
      - [Run the operator](#run-the-operator)
    - [Test Manually](#test-manually)
    - [Test helm chart locally](#test-helm-chart-locally)
      - [Run the automated helmchart test](#run-the-automated-helmchart-test)
      - [Manually test the helmchart](#manually-test-the-helmchart)
  - [Building/Pushing the operator image](#buildingpushing-the-operator-image)
  - [Deploy to OLM via bundle](#deploy-to-olm-via-bundle)
  - [Integration Test](#integration-test)
  - [Releasing](#releasing)
    - [Cleaning up](#cleaning-up)

# Vault Config Operator

This operator helps set up Vault Configurations. The main intent is to do so such that subsequently pods can consume the secrets made available.
There are two main principles through all of the capabilities of this operator:

1. high-fidelity API. The CRD exposed by this operator reflect field by field the Vault APIs. This is because we don't want to make any assumption on the kinds of configuration workflow that user will set up. That being said the Vault API is very extensive and we are starting with enough API coverage to support, we think, some simple and very common configuration workflows.
2. attention to security (after all we are integrating with a security tool). To prevent credential leaks we give no permissions to the operator itself against Vault. All APIs exposed by this operator contains enough information to authenticate to Vault using a local service account (local to the namespace where the API exist). In other word for a namespace user to be abel to successfully configure Vault, a service account in that namespace must have been previously given the needed Vault permissions.

Currently this operator covers the following Vault APIs:

## Authentication Engines

1. [AuthEngineMount](./docs/auth-engines.md#authenginemount) Sets up a [Vault Authentication Endpoint](https://www.vaultproject.io/docs/auth)
2. [KubernetesAuthEngineConfig](./docs/auth-engines.md#kubernetesauthengineconfig) Configures a [Vault Kubernetes Authentication Endpoint](https://www.vaultproject.io/docs/auth/kubernetes).
3. [KubernetesAuthEngineRole](./docs/auth-engines.md#KubernetesAuthEngineRole) Configures a Vault [Kubernetes Authentication](https://www.vaultproject.io/docs/auth/kubernetes) Role
4. [LDAPAuthEngineConfig](./docs/auth-engines.md#ldapauthengineconfig) Configures a [Vault LDAP Authentication Endpoint](https://www.vaultproject.io/docs/auth/ldap).
   - [LDAPAuthEngineGroup](./docs/auth-engines.md#ldapauthenginegroup) Creates or updates [Vault LDAP Authentication Engine Group](https://www.vaultproject.io/api-docs/auth/ldap#create-update-ldap-group) policies.

## Policy management

1. [Policy](./docs/policy-management.md#policy) Configures Vault [Policies](https://www.vaultproject.io/docs/concepts/policies)
2. [PasswordPolicy](./docs/policy-management.md#passwordpolicy) Configures Vault [Password Policies](https://www.vaultproject.io/docs/concepts/password-policies)

## Secret Engines

1. [SecretEngineMount](./docs/secret-engines.md#SecretEngineMount) Configures a Mount point for a [SecretEngine](https://www.vaultproject.io/docs/secrets)
2. [DatabaseSecretEngineConfig](./docs/secret-engines.md#DatabaseSecretEngineConfig) Configures a [Database Secret Engine](https://www.vaultproject.io/docs/secrets/databases) Connection
3. [DatabaseSecretEngineRole](./docs/secret-engines.md#DatabaseSecretEngineRole) Configures a [Database Secret Engine](https://www.vaultproject.io/docs/secrets/databases) Role
4. [GitHubSecretEngineConfig](./docs/secret-engines.md#GitHubSecretEngineConfig) Configures a Github Application to produce tokens, see the also the [vault-plugin-secrets-github](https://github.com/martinbaillie/vault-plugin-secrets-github)
5. [GitHubSecretEngineRole](./docs/secret-engines.md#GitHubSecretEngineRole) Configures a Github Application to produce scoped tokens, see the also the [vault-plugin-secrets-github](https://github.com/martinbaillie/vault-plugin-secrets-github)
6. [PKISecretEngineConfig](./docs/secret-engines.md#pkisecretengineconfig)  Configures a [PKI Secret Engine](https://www.vaultproject.io/docs/secrets/pki)
7. [PKISecretEngineRole](./docs/secret-engines.md#pkisecretenginerole)  Configures a [PKI Secret Engine](https://www.vaultproject.io/docs/secrets/pki) Role
8. [QuaySecretEngineConfig](./docs/secret-engines.md#QuaySecretEngineConfig) Configures a Quay server to produce Robot accounts, see the also the [vault-plugin-secrets-quay](https://github.com/redhat-cop/vault-plugin-secrets-quay)
9. [QuaySecretEngineRole](./docs/secret-engines.md#QuaySecretEngineRole) Configures a Quay server to produce credentials for a Robot account, see the also the [vault-plugin-secrets-quay](https://github.com/redhat-cop/vault-plugin-secrets-quay)
10. [QuaySecretEngineStaticRole](./docs/secret-engines.md#QuaySecretEngineStaticRole) Configures a Quay server to produce credentials for a Robot account using a fixed username and generated credentials, see the also the [vault-plugin-secrets-quay](https://github.com/redhat-cop/vault-plugin-secrets-quay)
11. [RabbitMQSecretEngineConfig](./docs/secret-engines.md#rabbitmqsecretengineconfig) Configures a [RabbitMQ Secret Engine](https://www.vaultproject.io/docs/secrets/rabbitmq#rabbitmq-secrets-engine)
12. [RabbitMQSecretEngineRole](./docs/secret-engines.md#rabbitmqsecretenginerole) Configures a [RabbitMQ Secret Engine Role](https://www.vaultproject.io/docs/secrets/rabbitmq#rabbitmq-secrets-engine)

## Secret Management

1. [RandomSecret](./docs/secret-management.md#RandomSecret) Creates a random secret in a vault [kv Secret Engine](https://www.vaultproject.io/docs/secrets/kv) with one password field generated using a [PasswordPolicy](https://www.vaultproject.io/docs/concepts/password-policies)
2. [VaultSecret](./docs/secret-management.md#VaultSecret) Creates a K8s Secret from one or more Vault Secrets

## The common authentication section

All APIs share a common authentication section, details can be found [here](./docs/auth-section.md)

## End to end example

See [this section](./docs/end-to-end-example.md) for an example scenario in which this operator could be used.

## Contributing a new Vault type

See [this section](./docs/contributing-vault-apis.md) of the documentation for a details on how to contribute a new type.

## Initializing the connection to Vault

At the moment the connection to Vault can be initialized with [Vault's standard environment variables](https://www.vaultproject.io/docs/commands#environment-variables).
See the [OLM documentation](https://github.com/operator-framework/operator-lifecycle-manager/blob/master/doc/design/subscription-config.md#env) on how to pass environment variables via a Subscription.
The variable that are read at client initialization are listed [here](https://github.com/hashicorp/vault/blob/14101f866414d2ed7850648b465c746ac8fda621/api/client.go#L35).

For certificates, the recommended approach is to mount the secret or configmap containing the certificate as described [here](https://github.com/operator-framework/operator-lifecycle-manager/blob/master/doc/design/subscription-config.md#volumes), and the configure the corresponding variables to point at the files location in the mounted path.

Here is an example:

This is a configmap that will be injected with the OpenShift service-ca, you will have different ways of injecting your CA:

```yaml
apiVersion: v1

kind: ConfigMap
metadata:
  annotations:
    service.beta.openshift.io/inject-cabundle: "true"
  name: ocp-service-ca
data: []
```

Here is the subscription using that configmap:

```yaml
apiVersion: operators.coreos.com/v1alpha1
kind: Subscription
metadata:
  name: vault-config-operator 
spec:
  channel: alpha
  installPlanApproval: Automatic
  name: vault-config-operator
  source: community-operators
  sourceNamespace: openshift-marketplace
  config:
    env:
    - name: VAULT_CACERT
      value: /vault-ca/ca.crt
    volumes:
    - name: vault-ca
      configMap:
        name: ocp-service-ca
    volumeMounts:
    - mountPath: /vault-ca
      name: vault-ca  
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
| s390x  | ✅  |

### Deploying from OperatorHub

> **Note**: This operator supports being installed disconnected environments

If you want to utilize the Operator Lifecycle Manager (OLM) to install this operator, you can do so in two ways: from the UI or the CLI.

#### Deploying from OperatorHub UI

- If you would like to launch this operator from the UI, you'll need to navigate to the OperatorHub tab in the console. Before starting, make sure you've created the namespace that you want to install this operator to with the following:

```shell
oc new-project vault-config-operator
```

- Once there, you can search for this operator by name: `vault config operator`. This will then return an item for our operator and you can select it to get started. Once you've arrived here, you'll be presented with an option to install, which will begin the process.
- After clicking the install button, you can then select the namespace that you would like to install this to as well as the installation strategy you would like to proceed with (`Automatic` or `Manual`).
- Once you've made your selection, you can select `Subscribe` and the installation will begin. After a few moments you can go ahead and check your namespace and you should see the operator running.

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

## Metrics

Prometheus compatible metrics are exposed by the Operator and can be integrated into OpenShift's default cluster monitoring. To enable OpenShift cluster monitoring, label the namespace the operator is deployed in with the label `openshift.io/cluster-monitoring="true"`.

```shell
oc label namespace <namespace> openshift.io/cluster-monitoring="true"
```

### Testing metrics

Openshift monitoring...

```sh
export operatorNamespace=vault-config-operator
oc label namespace ${operatorNamespace} openshift.io/cluster-monitoring="true"
oc rsh -n openshift-monitoring -c prometheus prometheus-k8s-0 /bin/bash
export operatorNamespace=vault-config-operator
curl -v --data-urlencode "query=controller_runtime_active_workers{namespace=\"${operatorNamespace}\"}" localhost:9090/api/v1/query
exit
```

in Kubernetes...

See the [Test helm chart locally](#test-helm-chart-locally) section to run the helmchart test which will also test that metrics work against a k8s kind cluster.

## Development

### Setup

If using vscode, you may need a `./.vscode/settings.json` file in this directory to make gopls happy since build tags are used.

See <https://github.com/golang/vscode-go/blob/master/docs/settings.md#buildbuildflags>

```json
{
    "gopls": {
        "ui.completion.usePlaceholders": true,
        "build.buildFlags": [
            "--tags=integration"
        ] 
    },
}
```

### Running the operator locally

#### Deploy a Vault instance

If you don't have a Vault instance available for testing, deploy one with these steps:

```shell
helm repo add hashicorp https://helm.releases.hashicorp.com
oc new-project vault
oc adm policy add-role-to-user admin -z vault -n vault
export cluster_base_domain=$(oc get dns cluster -o jsonpath='{.spec.baseDomain}')
helm upgrade vault hashicorp/vault -i --create-namespace -n vault --atomic -f ./config/local-development/vault-values.yaml --set server.route.host=vault-vault.apps.${cluster_base_domain}
```

> Note: you may need to manually remove the `service.beta.openshift.io/serving-cert-secret-name: vault-server-tls` annotation from the Service `vault-internal` and force the recreation of the service serving certificate since the annotation gets applied to both `vault` and `vault-internal` Services by the helm chart. The implementation expects the vault.vault.svc certificate. See [server-service.yaml](https://github.com/hashicorp/vault-helm/blob/main/templates/server-service.yaml) and [server-headless-service.yaml](https://github.com/hashicorp/vault-helm/blob/main/templates/server-headless-service.yaml).

#### Configure an Kubernetes Authentication mount point

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

#### Run the operator

> Note: this operator build process is tested with [podman](https://podman.io/), but some of the build files (Makefile specifically) use docker because they are generated automatically by operator-sdk. It is recommended [remap the docker command to the podman command](https://developers.redhat.com/blog/2020/11/19/transitioning-from-docker-to-podman#transition_to_the_podman_cli).

```shell
export repo=raffaelespazzoli
docker login quay.io/$repo
oc new-project vault-config-operator
oc project vault-config-operator
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
helm repo add bitnami https://charts.bitnami.com/bitnami
helm upgrade my-postgresql-database bitnami/postgresql -i --create-namespace -n test-vault-config-operator -f ./docs/examples/postgresql/postgresql-values.yaml --set securityContext.fsGroup=${guid} --set containerSecurityContext.runAsUser=${uid} --set volumePermissions.securityContext.runAsUser=${uid} --set metrics.securityContext.runAsUser=${uid}
oc apply -f ./test/database-engine-config.yaml -n test-vault-config-operator
```

Database Secret engine role

```shell
oc apply -f ./test/database-engine-read-only-role.yaml -n test-vault-config-operator
```

RandomSecret

```shell
oc apply -f ./test/password-policy.yaml -n vault-admin
envsubst < ./test/kv-engine-admin-policy.yaml | oc apply -f - -n vault-admin
envsubst < ./test/secret-writer-policy.yaml | oc apply -f - -n vault-admin
oc apply -f ./test/kv-engine-admin-role.yaml -n vault-admin
oc apply -f ./test/secret-writer-role.yaml -n vault-admin
oc apply -f ./test/kv-secret-engine.yaml -n test-vault-config-operator
oc apply -f ./test/random-secret.yaml -n test-vault-config-operator
```

VaultSecret

> Note: you must run the previous tests

```shell
oc apply -f ./test/vaultsecret/kubernetesauthenginerole-secret-reader.yaml -n vault-admin
envsubst < ./test/vaultsecret/policy-secret-reader.yaml | oc apply -f - -n vault-admin
oc apply -f ./test/vaultsecret/randomsecret-another-password.yaml -n test-vault-config-operator
oc apply -f ./test/vaultsecret/vaultsecret-randomsecret.yaml -n test-vault-config-operator
oc apply -f ./test/vaultsecret/vaultsecret-dynamicsecret.yaml -n test-vault-config-operator
# test a pullsecret... Create the secret in test-vault-config-operator/kv/pullsecret using the Vault UI first
oc apply -f ./test/vaultsecret/vaultsecret-pullsecret.yaml -n test-vault-config-operator
```

VaultSecret & RandomSecret using KV Secret Engine V2

> Note: you must run the previous tests

```sh
#Random Secret using v2 steps
envsubst < ./test/randomsecret/v2/01-policy-kv-engine-admin-v2.yaml | oc apply -f - -n vault-admin 
oc apply -f ./test/randomsecret/v2/02-kubernetesauthenginerole-kv-engine-admin-v2.yaml -n vault-admin 
oc apply -f test/randomsecret/v2/03-secretenginemount-kv-v2.yaml -n test-vault-config-operator
envsubst < ./test/randomsecret/v2/04-policy-secret-writer-v2.yaml | oc apply -f - -n vault-admin
oc apply -f ./test/randomsecret/v2/05-kubernetesauthenginerole-secret-writer-v2.yaml -n vault-admin
oc apply -f ./test/randomsecret/v2/06-randomsecret-randomsecret-password-v2.yaml -n test-vault-config-operator

#VaultSecret reading RandomSecret v2 steps
oc apply -f ./test/vaultsecret/v2/07-vaultsecret-randomsecret-v2.yaml -n test-vault-config-operator
```

Kube auth engine mount and config and role

```shell
oc apply -f ./test/kube-auth-engine-mount.yaml -n vault-admin
oc apply -f ./test/kube-auth-engine-config.yaml -n vault-admin
oc apply -f ./test/kube-auth-engine-role.yaml -n vault-admin
```

Github secret engine

create a github application following the instructions [here](https://github.com/martinbaillie/vault-plugin-secrets-github#setup-github).
save the ssh key in a file called: ./test/vault-secret-engine.private-key.pem and the application id

```shell
export application_id=163698 #replace with your own
export ssh_key=$(cat ./test/vault-secret-engine.private-key.pem | base64 -w 0)
envsubst < ./test/github-secret-engine-config-secret-template.yaml | oc apply -f - -n vault-admin
oc apply -f ./test/github-secret-engine-mount.yaml -n vault-admin
envsubst < ./test/github-secret-engine-config.yaml | oc apply -f - -n vault-admin
vault read -tls-skip-verify github/raf-backstage-demo/token
oc apply -f ./test/github-secret-engine-role.yaml -n vault-admin
vault read -tls-skip-verify github/raf-backstage-demo/token/one-repo-only
```

### Test helm chart locally

#### Run the automated helmchart test

```sh
make helmchart-test
```

Since this will run a kind instance, you may then test creating CRs manually...

```sh
export VAULT_TOKEN=$(kubectl get secret vault-init -n vault -o jsonpath='{.data.root_token}' | base64 -d)
export VAULT_ADDR="http://localhost"
export accessor=$(vault read -tls-skip-verify -format json sys/auth | jq -r '.data["kubernetes/"].accessor')
kubectl create namespace vault-admin
kubectl create namespace test-vault-config-operator
```

Next, run through the [Test Manually](#test-manually) steps.

#### Manually test the helmchart

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

## Integration Test

```sh
make integration
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
