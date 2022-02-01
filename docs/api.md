# Supported Vault API

This section of the documentation provides high-level documentation on the supported Vault API. For a complete list of all the supported, see [here](https://doc.crds.dev/github.com/redhat-cop/vault-config-operator).

- [Supported Vault API](#supported-vault-api)
  - [The Authentication Section](#the-authentication-section)
  - [Policy](#policy)
  - [PasswordPolicy](#passwordpolicy)
  - [AuthEngineMount](#authenginemount)
  - [KubernetesAuthEngineConfig](#kubernetesauthengineconfig)
  - [KubernetesAuthEngineRole](#kubernetesauthenginerole)
  - [SecretEngineMount](#secretenginemount)
  - [DatabaseSecretEngineConfig](#databasesecretengineconfig)
  - [DatabaseSecretEngineRole](#databasesecretenginerole)
  - [RandomSecret](#randomsecret)
  - [GitHubSecretEngineConfig](#githubsecretengineconfig)
  - [GitHubSecretEngineRole](#githubsecretenginerole)
  - [VaultSecret](#vaultsecret)
  - [RabbitMQSecretEngineConfig](#rabbitmqsecretengineconfig)
  - [RabbitMQSecretEngineRole](#rabbitmqsecretenginerole)
  - [PKISecretEngineConfig](#pkisecretengineconfig)
  - [PKISecretEngineRole](#pkisecretenginerole)

## The Authentication Section

Each API has an Authentication section that specifies how to authenticate to Vault. Here is an example:

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

The `Policy` CRD allows a user to create a [Vault Policy](https://www.vaultproject.io/docs/concepts/policies), here is an example:

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

## PasswordPolicy

The `PasswordPolicy` CRD allows a user to create a [Vault Password Policy](https://www.vaultproject.io/docs/concepts/password-policies), here is an example:

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: PasswordPolicy
metadata:
  name: simple-password-policy
spec:
  authentication: 
    path: kubernetes
    role: policy-admin  
  passwordPolicy: |
    length = 20
    rule "charset" {
      charset = "abcdefghijklmnopqrstuvwxyz"
    }
```

## AuthEngineMount

The `AuthEngineMount` CRD allows a user to define an [authentication engine endpoint](https://www.vaultproject.io/docs/auth)

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: AuthEngineMount
metadata:
  name: authenginemount-sample
spec:
  # Add fields here
  authentication: 
    path: kubernetes
    role: policy-admin
  type: kubernetes
  path: kube-authengine-mount-sample
```

The `type` field specifies the type of the authentication engine.

The `path` field specifies the path at which the auth engine is mounted. The complete path will be: `[namespace/]auth/{.spec.path}/{metadata.name}`

## KubernetesAuthEngineConfig

The `KubernetesAuthEngineConfig` CRD allows a user to configure an authentication engine mount of [type Kubernetes](https://www.vaultproject.io/docs/auth/kubernetes).

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: KubernetesAuthEngineConfig
metadata:
  name: authenginemount-sample
spec:
  authentication: 
    path: kubernetes
    role: policy-admin
  path: kube-authengine-mount-sample
  tokenReviewerServiceAccount:
    name: token-review-sa
  kubernetesHost:   
  kubernetesCACert: ...  
```

The `path` field specifies the path to configure. the complete path of the configuration will be: `[namespace/]auth/{.spec.path}/{metadata.name}/config`

The `tokenReviewerServiceAccount.name` field specifies the service account to be used to perform the token review. This account must exists and must be granted the TokenReviews create permission. If not specified it will default to `default`.

The `kubernetesCACert` field is the base64 encoded CA certificate that can be used to validate the connection to the master API. It will default to the content of the file `/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"`. This default should work for most cases.

The `kubernetesHost` field defines the master api endpoint. It defaults to `https://kubernetes.default.svc:443` and it should work most cases.

## KubernetesAuthEngineRole

The `KubernetesAuthEngineRole` creates a [Vault Authentication Role](https://www.vaultproject.io/docs/auth/kubernetes#configuration) for a Kubernetes Authentication mount, here is an example:

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

## GitHubSecretEngineConfig

The `GitHubSecretEngineConfig` CRD allows a user to create a GitHub Secret engine configuration. Only one configuration can exists per GitHub secret engine mount point, here is an example:

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: GitHubSecretEngineConfig
metadata:
  name: raf-backstage-demo-org
spec:
  authentication: 
    path: kubernetes
    role: policy-admin
  sSHKeyReference:
    secret:
      name: vault-github-app-key
  path: github/raf-backstage-demo
  applicationID: 123456
  organizationName: raf-backstage-demo
```

The `path` field specifies the path of the secret engine that will contain this configuration.

The `sSHKeyReference` field specifies how to retrieve the ssh key to the GitHub application.

The `applicationID` field specifies application id of the GitHub application.

The `organizationName` field specifies organization in which the application is installed.

More parameters exists for their explanation and for how to install the vault-plugin-secret-github engine see [here](https://github.com/martinbaillie/vault-plugin-secrets-github#config)

This CR is roughly equivalent to this Vault CLI command:

```shell
vault write [namespace/]github/raf-backstage-demo/config app_id=123456 prv_key=@key.pem org_name=raf-backstage-demo
```

## GitHubSecretEngineRole

The `GitHubSecretEngineRole` CRD allows a user to create a GitHub Secret engine role. A role allows to create narrowly scoped github tokens limiting the permission or the repositories on which they can be used, here is an example:

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: GitHubSecretEngineRole
metadata:
  name: one-repo-only
spec:
  authentication: 
    path: kubernetes
    role: policy-admin
  path: github/raf-backstage-demo
  repositories:
  - hello-world
```

The `path` field specifies the path of the secret engine that will contain this role.

The `repositories` field specifies on which repositories the generated credential can act.

More parameters exists for their explanation and for how to install the vault-plugin-secret-github engine see [here](https://github.com/martinbaillie/vault-plugin-secrets-github#permission-sets)

This CR is roughly equivalent to this Vault CLI command:

```shell
vault write [namespace/]github/raf-backstage-demo/permissionset/one-repo-only repositories=hello-world
```

to read a new credential from this role, execute the following:

```shell
vault read -tls-skip-verify github/raf-backstage-demo/token/one-repo-only
```

## VaultSecret

The VaultSecret CRD allows a user to create a K8s Secret from one or more Vault Secrets. It uses go templating to allow formatting of the K8s Secret in the `output.stringData` section of the spec.

> Note: if reading a dynamic secret you typically care to set the `refreshThreshold` only (not the `refreshPeriod`). For just Key/Value Vault secrets, set the `refreshPeriod`.
> 
> See https://www.vaultproject.io/docs/concepts/lease to understand lease durations.

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

## RabbitMQSecretEngineConfig

`RabbitMQSecretEngineConfig` CRD allows a user to create a RabbitMQ Secret Engine configuration, also called connection for an existing RabbitMQ Secret Engine Mount. Here is an example

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: RabbitMQSecretEngineConfig
metadata:
  name: example-rabbitmq-secret-engine-config
spec:
  authentication: 
    path: kubernetes
    role: rabbitmq-engine-admin
  connectionURI: https://my-rabbitMQ.com
  rootCredentials:
    secret:
      name: rabbitmq-admin-password
    passwordKey: rabbitmq-password
  path: vault-config-operator/rabbitmq
  username: rabbitmq
  leaseTTL: 86400 # 24 hours
  leaseMaxTTL: 86400 # 24 hours
```

The `connectionURI` field specifies how to connect to the rabbitMQ cluster. Supports http and https protocols.

The `username` field specific the username to be used to connect to the rabbitMQ cluster. This field is optional, if not specified the username will be retrieved from the credential secret.

The `path` field specifies the path of the secret engine to which this connection will be added.

The password and possibly the username can be retrieved in three different ways:

1. From a Kubernetes secret, specifying the `rootCredentialsFromSecret` field. The secret must be of [basic auth type](https://kubernetes.io/docs/concepts/configuration/secret/#basic-authentication-secret). If the secret is updated this connection will also be updated.
2. From a Vault secret, specifying the `rootCredentialsFromVaultSecret` field.
3. From a [RandomSecret](#RandomSecret), specifying the `rootCredentialsFromRandomSecret` field. When the RandomSecret generates a new secret, this connection will also be updated.

Additional options supported from [Vault Documentation](https://www.vaultproject.io/api-docs/secret/rabbitmq#configure-connection)

## RabbitMQSecretEngineRole

The `RabbitMQSecretEngineRole` CRD allows a user to create a Database Secret Engine Role, here is an example:

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: RabbitMQSecretEngineRole
metadata:
  name: rabbitmqsecretenginerole-sample
spec:
  authentication: 
    path: kubernetes
    role: rabbitmq-engine-admin
  path: vault-config-operator/rabbitmq
  tags: 'administrator'
  vhosts:
  - vhostName: '/'
    permissions:
      read: '.*'
      write: '.*'
      configure: '.*'
  - vhostName: 'my-vhost'
    permissions:
      read: 'my-queue'
      write: 'my-exchange'
  vhostTopics:
  - vhostName: '/'
    topics:
    - topicName: 'my-topic'
      permissions:
        read: '.*'
        write: '.*'
        configure: '.*'
    - topicName: 'my-read-topic'
      permissions:
        read: '.*'
```

The `tags` field specifies RabbitMQ permissions tags to associate with the user. This determines the level of access to the RabbitMQ management UI granted to the user.

The `vhostName` field specifies the name of vhost where permissions will be provided.

Permissions `read/write/configure` provides ability to `read/write/configure` specified queues and/or exchanges.

[Vault Documentation](https://www.vaultproject.io/api-docs/secret/rabbitmq#create-role)

## PKISecretEngineConfig

`PKISecretEngineConfig` CRD allows a user to create a PKI Secret Engine configuration. Here is an example

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: PKISecretEngineConfig
metadata:
  name: my-pki
spec:
  authentication: 
    path: kubernetes
    role: pki-engine-admin
  path: pki-vault-demo/pki
  commonName: pki-vault-demo.internal.io
  TTL: "8760h"
```

The `commonName` specifies the requested CN for the certificate.

The `path` field specifies the path of the secret engine to which this connection will be added.

The `TTL` specifies the requested Time To Live (after which the certificate will be expired). This cannot be larger than the engine's max (or, if not set, the system max).

This CR is roughly equivalent to this Vault CLI command:

```shell
vault write pki-vault-demo/pki/root/generate/internal \
    common_name=pki-vault-demo.internal.io \
    ttl=8760h
```

## PKISecretEngineRole

The `PKISecretEngineRole` CRD allows a user to create a PKI Secret Engine Role, here is an example:

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: PKISecretEngineRole
metadata:
  name: my-role
spec:
  authentication: 
    path: kubernetes
    role: database-engine-admin
  path: pki-vault-demo/pki
  allowedDomains: 
   - internal.io
   - pki-vault-demo.svc
  maxTTL: "8760h"
```

The `allowedDomains` specifies the domains of the role. This is used with the allow_bare_domains and allow_subdomains options.

The `maxTTL` specifies the maximum Time To Live provided as a string duration with time suffix. Hour is the largest suffix. If not set, defaults to the system maximum lease TTL.

The `path` field specifies the path of the secret engine to which this connection will be added.

This CR is roughly equivalent to this Vault CLI command:

```shell
vault write pki-vault-demo/pki/roles/my-role \
    allowed_domains=internal.io,pki-vault-demo.svc \
    max_ttl="8760h"
```