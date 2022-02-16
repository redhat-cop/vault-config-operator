# Secret Management

- [Secret Management](#secret-management)
  - [RandomSecret](#randomsecret)
  - [VaultSecret](#vaultsecret)

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
  isKVSecretsEngineV2: false
  path: kv/vault-tenant
  secretKey: password
  secretFormat:
    passwordPolicyName: my-complex-password-format
  refreshPeriod: 1h
```

The `path` field specifies the path at which the secret will be written, it must correspond to a kv Secret Engine mount.

The `isKVSecretsEngineV2` field is used to indicate if the KV Secrets Engine (defined in the `path`) expects a v1 or v2 data payload. Defaults to false to indicate that a v1 payload will be sent.

The `secretKey` field is the key of the secret.

The `secretFormat` is a reference to a Vault Password policy, it can also supplied inline.

The `refreshPeriod` specifies the frequency at which this secret will be regenerated. This is an optional field, if not specified the secret will be generated once and then never updated.

With a RandomSecret it is possible to build workflow in which the root password of a resource that we need to protect is never stored anywhere, except in vault. One way to achieve this is to have a random secret seed the root password. Then crete an operator that watches the RandomSecret and retrieves ths generated secret from vault and updates the resource to be protected. Finally configure the Secret Engine object to watch for the RandomSecret updates.

This CR is roughly equivalent to this Vault CLI command:

```shell
vault kv put [namespace/]kv/vault-tenant password=<generated value>
```

## VaultSecret

The VaultSecret CRD allows a user to create a K8s Secret from one or more Vault Secrets. It uses go templating to allow formatting of the K8s Secret in the `output.stringData` section of the spec.

Any manual data change or deletion of the K8s Secret owned by a VaultSecret CR will result in a re-reconciliation by the controller. A hash annotation `vaultsecret.redhatcop.redhat.io/secret-hash`, computed when the K8s Secret is created/updated, is used to verify the integrity of the K8s Secret data.

> Note: if reading a dynamic secret you typically care to set the `refreshThreshold` only (not the `refreshPeriod`). For just Key/Value Vault secrets, set the `refreshPeriod`.
> See <https://www.vaultproject.io/docs/concepts/lease> to understand lease durations.

- `refreshPeriod` the pull interval for syncing Vault secrets with the K8s Secret. This settings takes precedence over any lease duration returned by vault, effectively controlling when exactly all vault secrets defined in the vaultSecretDefinitions should re-sync.
- `refreshThreshold` this is will instruct the operator to refresh the K8s Secret when a percentage of the lease duration has elapsed, if no `refreshPeriod` is specified. This is particularly useful for controlling when dynamic secrets should be refreshed before the lease duration is exceeded. The default is 90, meaning the secret would refresh after 90% of the time has passed from the vault secret's lease duration. When multiple vaultSecretDefinitions are defined, the smallest lease duration will be used when performing the scheduling for the next refresh.
- `vaultSecretDefinitions` is an array of Vault Secret References. Every `vaultSecretDefinition` has...
  - [authentication](#the-authentication-section) section.
  - `name` a unique name for the Vault secret to reference when templating, since many Vault secrets may have the same name.
  - `path` field specifies the path at which the secret will be read from.
- `output` is the K8s Secret to output to after go template processing.
  - `name` the final K8s Secret Name to output to.
  - `stringData` stringData allows specifying non-binary secret data in string form. It is provided as a write-only input field for convenience. All keys and values are merged into the data field on write, overwriting any existing values. The stringData field is never output when reading from the API. You specify variables from `vaultSecretDefinitions` in the form of *'{{ .name.key }}'* using go templating where name is the arbitrary name in the vaultSecretDefinition and key matches the Vault secret key. The go text and most [sprig](http://masterminds.github.io/sprig/) library functions are also available when templating.
  - `type` is the K8s Secret type used to facilitate programmatic handling of secret data.
  - `labels` are any k8s Secret [labels](http://kubernetes.io/docs/user-guide/labels) to include.
  - `annotations` are any k8s Secret [annotations](http://kubernetes.io/docs/user-guide/annotations) to include.

Example CR for Key/Value Vault Secrets with `refreshPeriod`...

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: VaultSecret
metadata:
  name: randomsecret
spec:
  refreshPeriod: 1m0s
  vaultSecretDefinitions:
    - authentication:
        path: kubernetes
        role: secret-reader
        serviceAccount:
          name: default
      name: randomsecret
      path: test-vault-config-operator/kv/randomsecret-password
    - authentication:
        path: kubernetes
        role: secret-reader
        serviceAccount:
          name: default
      name: anotherrandomsecret
      path: test-vault-config-operator/kv/another-password
  output:
    name: randomsecret
    stringData:
      password: '{{ .randomsecret.password }}'
      anotherpassword: '{{ .anotherrandomsecret.password }}'
    type: Opaque
    labels:
      app: test-vault-config-operator
    annotations:
      refresh: every-minute
```

Example CR for a dynamic Vault Secret without `refreshPeriod` defined (we rely on the lease duration returned to us by Vault to calculate when to refresh next)...

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: VaultSecret
metadata:
  name: dynamicsecret
spec:
  refreshThreshold: 85 # after 85% of the lease_duration of the dynamic secret has elapsed, refresh the secret
  vaultSecretDefinitions:
    - authentication:
        path: kubernetes
        role: secret-reader
        serviceAccount:
          name: default
      name: dynamicsecret
      path: test-vault-config-operator/database/creds/read-only
  output:
    name: dynamicsecret
    stringData:
      username: '{{ .dynamicsecret.username }}'
      password: '{{ .dynamicsecret.password }}'
    type: Opaque
    labels:
      app: test-label
    annotations:
      refresh: test-annotation
```
