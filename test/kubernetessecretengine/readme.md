# Kubernetes Secret Engine Test

Assuming you have done the initial setup

Create the kubernetes secret engine mount

```shell
oc adm policy add-cluster-role-to-user cluster-admin -z default -n vault-admin
oc apply -f ./test/kubernetessecretengine/kubese-mount.yaml -n vault-admin
```

Create the Kubernetes secret engine config

```shell
oc apply -f ./test/kubernetessecretengine/kubese-config.yaml -n vault-admin
```

Create a kubernetes secret engine role. In this example the role will create a service account, role, and token for a service account in the `vault-admin` namespace to have edit privileges on the `default` namespace.

```shell
oc apply -f ./test/kubernetessecretengine/kubese-role.yaml -n vault-admin
```

Retrieve a token

```shell
export cluster_base_domain=$(oc get dns cluster -o jsonpath='{.spec.baseDomain}')
export VAULT_ADDR=https://vault-vault.apps.${cluster_base_domain}
export VAULT_TOKEN=$(oc get secret vault-init -n vault -o jsonpath='{.data.root_token}' | base64 -d )
vault write -tls-skip-verify kubese-test/creds/kubese-default-edit kubernetes_namespace=default
```

Retrieve a token via VaultSecret

```shell
oc apply -f ./test/kubernetessecretengine/vaultsecret.yaml -n vault-admin
```

check that the `vault-admin/kubese-test secret` is created
