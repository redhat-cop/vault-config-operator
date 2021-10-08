# Vault Config Operator

This operator helps set up Vault Configurations. The main intent is to do so such that subsequently pods can consume the secrets made available.
There are two main principles through all of the capabilities of this operator:

1. high-fidelity API. The CRD exposed by this operator reflect field by field the Vault APIs. This is because we don't want to make any assumption on the kinds of configuration workflow that user will set up. That being said the Vault API is very extensive and we are starting with enough API coverage to support, we think, some simple and very common configuration workflows.
2. attention to security (after all we are integrating with a security tool). To prevent credential leaks we give no permissions to the operator itself against Vault. All APIs exposed by this operator contains enough information to authenticate to Vault using a local service account (local to the namespace where the API exist). In other word for a namespace user to be abel to successfully configure Vault, a service account in that namespace must have been previously given the needed Vault permissions.

Currently this operator supports the following CRDs:

1. [Policy](#policy) Configures Vault [Policies](https://www.vaultproject.io/docs/concepts/policies)
2. [VaultRole](#VaultRole) Configures a Vault [Kubernetes Authentication](https://www.vaultproject.io/docs/auth/kubernetes) Role
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

## VaultRole

The `VaultRole` creates a Vault Authentication Role for a Kubernetes Authentication mount, here is an example:

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: VaultRole
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
