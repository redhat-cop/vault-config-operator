# Secret Engines

- [Secret Engines](#secret-engines)
  - [SecretEngineMount](#secretenginemount)
  - [DatabaseSecretEngineConfig](#databasesecretengineconfig)
  - [DatabaseSecretEngineRole](#databasesecretenginerole)
  - [GitHubSecretEngineConfig](#githubsecretengineconfig)
  - [GitHubSecretEngineRole](#githubsecretenginerole)
  - [RabbitMQSecretEngineConfig](#rabbitmqsecretengineconfig)
  - [RabbitMQSecretEngineRole](#rabbitmqsecretenginerole)
  - [PKISecretEngineConfig](#pkisecretengineconfig)
  - [PKISecretEngineRole](#pkisecretenginerole)

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
