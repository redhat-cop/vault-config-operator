# Secret Engines

- [Secret Engines](#secret-engines)
  - [SecretEngineMount](#secretenginemount)
  - [DatabaseSecretEngineConfig](#databasesecretengineconfig)
  - [DatabaseSecretEngineRole](#databasesecretenginerole)
  - [DatabaseSecretEngineStaticRole](#databasesecretenginestaticrole)
  - [GitHubSecretEngineConfig](#githubsecretengineconfig)
  - [GitHubSecretEngineRole](#githubsecretenginerole)
  - [QuaySecretEngineConfig](#quaysecretengineconfig)
  - [QuaySecretEngineRole](#quaysecretenginerole)
  - [QuaySecretEngineStaticRole](#quaysecretenginestaticrole)
  - [RabbitMQSecretEngineConfig](#rabbitmqsecretengineconfig)
  - [RabbitMQSecretEngineRole](#rabbitmqsecretenginerole)
  - [PKISecretEngineConfig](#pkisecretengineconfig)
  - [PKISecretEngineRole](#pkisecretenginerole)
  - [KubernetesSecretEngineConfig](#kubernetessecretengineconfig)
  - [KubernetesSecretEngineRole](#kubernetessecretenginerole)
  - [AzureSecretEngineConfig] (#azuresecretengineconfig) 
  - [AzureSecretEngineRole] (#azuresecretenginerole)
  - [AWSSecretEngineConfig] (#awssecretengineconfig) 
  - [AWSSecretEngineRole] (#awssecretenginerole)


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
  rootPasswordRotation:
    enable: true
    rotationPeriod: 2m  
```

The `pluginName` field specifies what type of database this connection is for.

The `allowedRoles` field specifies which role names can be created for this connection.

The `connectionURL` field specifies how to connect to the database.

The `username` field specific the username to be used to connect to the database. This field is optional, if not specified the username will be retrieved from teh credential secret.

The `path` field specifies the path of the secret engine to which this connection will be added.

The `rootPasswordRotation.enable` field activates the root password rotation. The root password wil be rotated immediately. It is recommended to use this feature with care as there is no way to recover the root password. 

The `rootPasswordRotation.rotationPeriod` field tells the operator to periodically rotate the root password. If only enable is specified the password will be rotated only once.

The password and possibly the username can be retrived a three different ways:

1. From a Kubernetes secret, specifying the `rootCredentialsFromSecret` field. The secret must be of [basic auth type](https://kubernetes.io/docs/concepts/configuration/secret/#basic-authentication-secret). If the secret is updated this connection will also be updated.
2. From a Vault secret, specifying the `rootCredentialsFromVaultSecret` field.
3. From a [RandomSecret](#RandomSecret), specifying the `rootCredentialsFromRandomSecret` field. When the RandomSecret generates a new secret, this connection will also be updated.

The `passwordAuthentication` field, set to `scram-sha-256`, tells Vault to hash the password before sending it to PostgreSQL. This field is optional; if not specified, the default value is "password". When set to "password", passwords are sent to PostgreSQL in clear text and may appear as such in PostgreSQL logs.

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

## DatabaseSecretEngineStaticRole

The `DatabaseSecretEngineStaticRole` CRD allows a user to create a Database Secret Engine Static Role, here is an example:

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: DatabaseSecretEngineStaticRole
metadata:
  name: read-only-static
spec:
  authentication: 
    path: kubernetes
    role: database-engine-admin
  path: test-vault-config-operator/database
  dBName: my-postgresql-database
  username: helloworld
  rotationPeriod: 3600
  rotationStatements:
    - ALTER USER "{{name}}" WITH PASSWORD '{{password}}'; 
  credentialType: password
  passwordCredentialConfig: {}
```  

The `path` field specifies the path of the secret engine that will contain this role.

The `dBname` field specifies the name of the connection to be used with this role.

The `username` field specifies a username/role pre-existing in the database. 

The `rotationStatements` field specifies the statements to run to rotate the user credentials.

Other standard Database Secret Static Engine Role fields are available for fine tuning, see the [Vault Documentation](https://developer.hashicorp.com/vault/api-docs/secret/databases#create-static-role)

This CR is roughly equivalent to this Vault CLI command:

```shell
vault write [namespace/]test-vault-config-operator/database/static-roles/read-only-static db_name=my-postgresql-database username=helloworld rotation_statements="ALTER USER \"{{name}}\" WITH PASSWORD \"{{password}}\";" rotationPeriod=3600 credentialType=password
```

## GitHubSecretEngineConfig

The `GitHubSecretEngineConfig` CRD allows a user to create a GitHub Secret engine configuration. 
Note: this secret engine requires the [vault-plugin-secrets-github](https://github.com/martinbaillie/vault-plugin-secrets-github) `v2.0.0` to be installed. 
Only one configuration can exists per GitHub secret engine mount point, here is an example:

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
```

The `path` field specifies the path of the secret engine that will contain this configuration.

The `sSHKeyReference` field specifies how to retrieve the ssh key to the GitHub application.

The `applicationID` field specifies application id of the GitHub application.

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
  organizationName: raf-backstage-demo
```

The `path` field specifies the path of the secret engine that will contain this role.

The `repositories` field specifies on which repositories the generated credential can act.

The `organizationName` field specifies organization in which the application is installed.

More parameters exists for their explanation and for how to install the vault-plugin-secret-github engine see [here](https://github.com/martinbaillie/vault-plugin-secrets-github#permission-sets)

Available permissions are listed [here](https://docs.github.com/en/rest/apps/apps?apiVersion=2022-11-28#create-an-installation-access-token-for-an-app)

This CR is roughly equivalent to this Vault CLI command:

```shell
vault write [namespace/]github/raf-backstage-demo/permissionset/one-repo-only repositories=hello-world
```

to read a new credential from this role, execute the following:

```shell
vault read -tls-skip-verify github/raf-backstage-demo/token/one-repo-only
```

## QuaySecretEngineConfig

The `QuaySecretEngineConfig` CRD allows a user to create a Quay Secret engine configuration. Only one configuration can exists per Quay secret engine mount point, here is an example:

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: QuaySecretEngineConfig
metadata:
  name: quay-org
spec:
  authentication: 
    path: kubernetes
    role: policy-admin
  rootCredentials:
    secret:
      name: quay-token
  path: quay/demo
  url: https://quay.io
```

The `path` field specifies the path of the secret engine that will contain this configuration.

The `url` field specifies the endpoint for the Quay server

The token can retrieved in three different ways:

1. From a Kubernetes secret, specifying the `rootCredentialsFromSecret` field. The secret must be of [basic auth type](https://kubernetes.io/docs/concepts/configuration/secret/#basic-authentication-secret). If the secret is updated this connection will also be updated.
2. From a Vault secret, specifying the `rootCredentialsFromVaultSecret` field.
3. From a [RandomSecret](#RandomSecret), specifying the `rootCredentialsFromRandomSecret` field. When the RandomSecret generates a new secret, this connection will also be updated.

More parameters exists for their explanation and for how to install the vault-plugin-secrets-quay engine see [here](https://github.com/redhat-cop/vault-plugin-secrets-quay)

This CR is roughly equivalent to this Vault CLI command:

```shell
vault write [namespace/]quay/demo/config url=https://<QUAY_HOST> token=<token>
```

## QuaySecretEngineRole

The `QuaySecretEngineRole` CRD allows a user to create a Quay secret engine role. A role allows for the creation of a narrowly scoped Quay Robot account within an organization.

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: QuaySecretEngineRole
metadata:
  name: repo-manager
spec:
  authentication: 
    path: kubernetes
    role: policy-admin
  path: quay/demo
  namespaceName: myorg
  createRepositories: true
  defaultPermission: write
  TTL: "8760h"
  
```

The `path` field specifies the path of the secret engine that will contain this role.

The `namespaceName` field specifies the type of name of the namespace the robot account should be created within.

The `createRepositories` field specifies whether the robot account should be granted permission to create new repositories.

The `defaultPermission` field specifies the default permission that should apply to newly created repositories.

The `TTL` specifies the requested Time To Live (after which the certificate will be expired). This cannot be larger than the engine's max (or, if not set, the system max).

More parameters exists for their explanation and for how to install the vault-plugin-secrets-quay engine see [here](https://github.com/redhat-cop/vault-plugin-secrets-quay)

This CR is roughly equivalent to this Vault CLI command:

```shell
vault write [namespace/]quay/demo/roles/repo-manager namespaceName=myorg createRepositories=true defaultPermission=write TTL=8760h
```

to read a new credential from this role, execute the following:

```shell
vault read quay/demo/creds/repo-manager
```

## QuaySecretEngineStaticRole

The `QuaySecretEngineStaticRole` CRD allows a user to create a Quay secret engine role. A role allows for the creation of a narrowly scoped Quay Robot account within an organization where a fixed robot account is set.

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: QuaySecretEngineStaticRole
metadata:
  name: repo-manager
spec:
  authentication: 
    path: kubernetes
    role: policy-admin
  path: quay/demo
  namespaceName: myorg
  createRepositories: true
  defaultPermission: write
```

The `path` field specifies the path of the secret engine that will contain this role.

The `namespaceName` field specifies the type of name of the namespace the robot account should be created within.

The `createRepositories` field specifies whether the robot account should be granted permission to create new repositories.

The `defaultPermission` field specifies the default permission that should apply to newly created repositories.

More parameters exists for their explanation and for how to install the vault-plugin-secrets-quay engine see [here](https://github.com/redhat-cop/vault-plugin-secrets-quay)

This CR is roughly equivalent to this Vault CLI command:

```shell
vault write [namespace/]quay/demo/static-roles/repo-manager namespaceName=myorg createRepositories=true defaultPermission=write TTL=8760h
```

to read a new credential from this role, execute the following:

```shell
vault read quay/demo/static-creds/repo-manager
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

## KubernetesSecretEngineConfig

`KubernetesSecretEngineConfig` CRD allows a user to create a [Kubernetes Secret Engine configuration](https://www.vaultproject.io/api-docs/secret/kubernetes#write-configuration). Here is an example:

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: KubernetesSecretEngineConfig
metadata:
  name: kubese-test
spec:
  authentication: 
    path: kubernetes
    role: policy-admin
  path: kubese-test 
  kubernetesHost: https://kubernetes.default.svc:443
  jwtReference: 
    secret:
      name: default-token-lbnfc
```

The `kubernetesHost` specifies URL of the API server to connect to.

The `path` field specifies the path of the secret engine to which this connection will be added.

The `jwtReference` specifies a reference to service account token to be used as credentials when connecting to the API server.

This CR is roughly equivalent to this Vault CLI command:

```shell
vault write -f kube-setest/config \
    kubernetes_host=https://kubernetes.default.svc:443 \
    service_account_jwt=xxxx
```

## KubernetesSecretEngineRole

The `KubernetesSecretEngineRole` CRD allows a user to create a [Kubernetes Secret Engine Role](https://www.vaultproject.io/api-docs/secret/kubernetes#create-role), here is an example:

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: KubernetesSecretEngineRole
metadata:
  name: kubese-default-edit
spec:
  authentication: 
    path: kubernetes
    role: policy-admin
  path: kubese-test
  allowedKubernetesNamespaces:
  - default
  kubernetesRoleName: "edit"
  kubernetesRoleType: "ClusterRole"
  nameTemplate: vault-sa-{{random 10 | lowercase}}
```

The `allowedKubernetesNamespaces` field specifies on which namespaces it is possible to request this role.

The `kubernetesRoleName` field specifies which role should the service account receive.

The `kubernetesRoleType` field specifies whether the role is a namespaced role or a cluster role.

The `nameTemplate` field specifies the name template to be sued to create the service account. There are other ways of handling the service accounts, see all the options in the API documentation.

This CR is roughly equivalent to this Vault CLI command:

```shell
vault write kubese-test/roles/kubese-default-edit \
    allowed_kubernetes_namespaces="default" \
    kubernetes_role_name="edit" \
    kubernetes_role_name="ClusterRole" \
    nameTemplate="vault-sa-{{random 10 | lowercase}}" \
```

## AzureSecretEngineConfig

The `AzureSecretEngineConfig` CRD allows a user to create an Azure Secret Engine configuration. 

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: AzureSecretEngineConfig
metadata:
  labels:
    app.kubernetes.io/name: azuresecretengineconfig
    app.kubernetes.io/instance: azuresecretengineconfig-sample
    app.kubernetes.io/part-of: vault-config-operator
    app.kubernetes.io/managed-by: kustomize
    app.kubernetes.io/created-by: vault-config-operator
  name: azuresecretengineconfig-sample
spec:
  authentication:
    path: vault-admin
    role: vault-admin
    serviceAccount:
      name: vault
  connection:
    address: 'https://vault.example.com'
  path: azure
  tenantID: "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxx"
  clientID: "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxx"
  subscriptionID: "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxx"
  environment: "AzurePublicCloud"
  passwordPolicy: ""
  rootPasswordTTL: "128d"
  azureCredentials:
    secret: 
      name: aad-credentials
    usernameKey: clientid
    passwordKey: clientsecret
```

The `subscriptionID` field - The subscription id for the Azure Active Directory. This value can also be provided with the AZURE_SUBSCRIPTION_ID environment variable.

The `tenantID` field - The tenant id for the Azure Active Directory. This value can also be provided with the AZURE_TENANT_ID environment variable.

The `clientID` field - The OAuth2 client id to connect to Azure. This value can also be provided with the AZURE_CLIENT_ID environment variable. See authentication for more details.

The `environment` field - The Azure environment. This value can also be provided with the AZURE_ENVIRONMENT environment variable. If not specified, Vault will use Azure Public Cloud.

The `passwordPolicy` field - Specifies a password policy to use when creating dynamic credentials. Defaults to generating an alphanumeric password if not set.

The `rootPasswordTTL` field - Specifies how long the root password is valid for in Azure when rotate-root generates a new client secret. Uses duration format strings.

 The `azureCredentials` field - The OAuth Client Secret from the provider for OIDC roles.
 The OIDCClientSecret and possibly the OIDCClientID can be retrived a three different ways :

1. From a Kubernetes secret, specifying the `azureCredentials` field as follows:
```yaml
  azureCredentials:
    secret: 
      name: aad-credentials
    usernameKey: clientid
    passwordKey: clientsecret
```
The secret must be of [basic auth type](https://kubernetes.io/docs/concepts/configuration/secret/#basic-authentication-secret). 

Example Secret : 
```bash
kubectl create secret generic aad-credentials --from-literal=clientid="123456-1234-1234-1234-123456789" --from-literal=clientsecret="saffsfsdfsfsdgsdgsdgsdgghdfhdhdgsjgjgjfj" -n vault-admin
```
If the secret is updated this connection will also be updated.

2. From a [Vault secret](https://developer.hashicorp.com/vault/docs/secrets/kv), specifying the `azureCredentials` field as follows :
```yaml
  azureCredentials:
    vaultSecret: 
      path: secret/foo
    usernameKey: clientid
    passwordKey: clientsecret
```
3. From a [RandomSecret](secret-management.md#RandomSecret), specifying the `azureCredentials` field as follows : 
```yaml
  azureCredentials:
    randomSecret: 
      name: aad-credentials
    usernameKey: clientid
    passwordKey: clientsecret
```
When the RandomSecret generates a new secret, this connection will also be updated.

## AzureSecretEngineRole
The `AzureSecretEngineRole` CRD allows a user to create a [Azure Secret Engine Role](https://developer.hashicorp.com/vault/api-docs/secret/azure#create-update-role)

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: AzureSecretEngineRole
metadata:
  labels:
    app.kubernetes.io/name: azuresecretenginerole
    app.kubernetes.io/instance: azuresecretenginerole-sample
    app.kubernetes.io/part-of: vault-config-operator
    app.kubernetes.io/managed-by: kustomize
    app.kubernetes.io/created-by: vault-config-operator
  name: azuresecretenginerole-sample
spec:
  authentication:
    path: vault-admin
    role: vault-admin
    serviceAccount:
      name: vault
  connection:
    address: 'https://vault.example.com'
  path: azure
  name: "azure-role"
  azureRoles: ""
  azureGroups: ""
  applicationObjectID: ""
  persistApp: ""
  TTL: ""
  maxTTL: ""
  permanentlyDelete: ""
  signInAudience: ""
  tags: ""
```
 The `azureRoles` field - List of Azure roles to be assigned to the generated service principal. The array must be in JSON format, properly escaped as a string. See roles docs for details on role definition.
 
 The `azureGroups` field - List of Azure groups that the generated service principal will be assigned to. The array must be in JSON format, properly escaped as a string. See groups docs for more details.

 The `applicationObjectID` field - Application Object ID for an existing service principal that will be used instead of creating dynamic service principals. If present, azure_roles will be ignored. See roles docs for details on role definition.

 The `persistApp` field - If set to true, persists the created service principal and application for the lifetime of the role. Useful for when the Service Principal needs to maintain ownership of objects it creates.

 The `TTL` field - Specifies the default TTL for service principals generated using this role. Accepts time suffixed strings ("1h") or an integer number of seconds. Defaults to the system/engine default TTL time.

 The `maxTTL` field - Specifies the maximum TTL for service principals generated using this role. Accepts time suffixed strings ("1h") or an integer number of seconds. Defaults to the system/engine max TTL time.

 The `permanentlyDelete` field - Specifies whether to permanently delete Applications and Service Principals that are dynamically created by Vault. If application_object_id is present, permanently_delete must be false.

 The `signInAudience` field - Specifies the security principal types that are allowed to sign in to the application. Valid values are: AzureADMyOrg, AzureADMultipleOrgs, AzureADandPersonalMicrosoftAccount, PersonalMicrosoftAccount.

 The `tags` field - A comma-separated string of Azure tags to attach to an application.

## AWSSecretEngineConfig

The `AWSSecretEngineConfig` CRD allows a user to create an [AWS Secret Engine configuration](https://developer.hashicorp.com/vault/api-docs/secret/aws#configure-root-credentials). 

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: AWSSecretEngineConfig
metadata:
  labels:
    app.kubernetes.io/name: awssecretengineconfig
    app.kubernetes.io/instance: awssecretengineconfig-sample
    app.kubernetes.io/part-of: vault-config-operator
    app.kubernetes.io/managed-by: kustomize
    app.kubernetes.io/created-by: vault-config-operator
  name: awssecretengineconfig-sample
spec:
  authentication:
    path: vault-admin
    role: vault-admin
  connection:
    address: 'https://vault.example.com'
  path: aws
  awsCredentials:
    secret: 
      name: aws-credentials
    usernameKey: accessKey
    passwordKey: secretKey
```

The previous example is equivalent to 

```sh
vault write aws/config/root \
    access_key=xxx \
    secret_key=xxx \
    region=us-east-1
```

The `maxRetries` field - (Optional) Number of max retries the client should use for recoverable errors. The default (-1) falls back to the AWS SDK's default behavior.

The `roleArn` field - (Optional) Role ARN to assume for plugin workload identity federation (Enterprise). Required with identity_token_audience.

The `identityTokenAudience` field - (Optional) The audience claim value for plugin identity tokens (Enterprise). Must match an allowed audience configured for the target IAM OIDC identity provider.

The `IdentityTokenTTL` field - (Optional) The TTL of generated tokens (Enterprise). Defaults to 1 hour. Uses duration format strings.

The `region` field - (Optional) Specifies the AWS region. If not set it will use the AWS_REGION env var, AWS_DEFAULT_REGION env var, or us-east-1 in that order.

The `iamEndpoint` field - (Optional) Specifies a custom HTTP IAM endpoint to use.

The `stsEndpoint` field - (Optional) Specifies a custom HTTP STS endpoint to use.

The `stsRegion` field - (Optional) Specifies a custom STS region to use (should match sts_endpoint).

The `STSFallbackEndpoints` field - (Optional) Specifies an ordered list of fallback STS endpoints to use.

The `stsFallbackRegions` field - (Optional) Specifies an ordered list of fallback STS regions to use (should match fallback endpoints).

The `usernameTemplate` field - (Optional) Template describing how dynamic usernames are generated.

The `rotationPeriod` field - (Optional) The amount of time, in seconds, Vault should wait before rotating the root credential (Enterprise). A zero value tells Vault not to rotate the root credential.

The `rotationSchedule` field - (Optional) The schedule, in cron-style time format, defining the schedule on which Vault should rotate the root token (Enterprise).

The `rotationWindow` field - (Optional) The maximum amount of time, in seconds, allowed to complete a rotation when a scheduled token rotation occurs (Enterprise).

The `disableAutomatedRotation` field - (Optional) Cancels all upcoming rotations of the root credential until unset (Enterprise).

The `awsCredentials` field - The OAuth Client Secret from the provider for OIDC roles.
The access key and secret can be retrived a three different ways:

1. From a Kubernetes secret, specifying the `azureCredentials` field as follows:
```yaml
  awsCredentials:
    secret: 
      name: aws-credentials
    usernameKey: accessKey
    passwordKey: secretKey
```
The secret must be of [basic auth type](https://kubernetes.io/docs/concepts/configuration/secret/#basic-authentication-secret). 

Example Secret : 
```bash
kubectl create secret generic aad-credentials --from-literal=clientid="123456-1234-1234-1234-123456789" --from-literal=clientsecret="saffsfsdfsfsdgsdgsdgsdgghdfhdhdgsjgjgjfj" -n vault-admin
```
If the secret is updated this connection will also be updated.

2. From a [Vault secret](https://developer.hashicorp.com/vault/docs/secrets/kv), specifying the `azureCredentials` field as follows:
```yaml
  awsCredentials:
    vaultSecret: 
      path: secret/foo
    usernameKey: accessKey
    passwordKey: secretKey
```
3. From a [RandomSecret](secret-management.md#RandomSecret), specifying the `awsCredentials` field as follows: 
```yaml
  awsCredentials:
    randomSecret: 
      name: aws-credentials
    usernameKey: accessKey
    passwordKey: secretKey
```
When the RandomSecret generates a new secret, this connection will also be updated.

Alternatively, the `identityTokenAudience` and `roleArn` can be set to enable [Plugin Workload Identity Federation](https://developer.hashicorp.com/vault/docs/secrets/aws#plugin-workload-identity-federation-wif)

## AWSSecretEngineRole

The `AWSSecretEngineRole` CRD allows a user to create a [AWS Secret Engine Role](https://developer.hashicorp.com/vault/api-docs/secret/aws#create-update-role)

The `credentialType` field - Specifies the type of credential to be used when retrieving credentials from the role. Must be one of iam_user, assumed_role, federation_token, or session_token.

The `roleArns` field - Specifies the ARNs of the AWS roles this Vault role is allowed to assume. Required when credential_type is `assumed_role` and prohibited otherwise.

The `policyArns` field - Specifies a list of AWS managed policy ARNs. The behavior depends on the credential type. With `iam_user`, the policies will be attached to IAM users when they are requested. With `assumed_role` and `federation_token`, the policy ARNs will act as a filter on what the credentials can do.

The `policyDocument` field - The IAM policy document for the role. The behavior depends on the credential type. With `iam_user`, the policy document will be attached to the IAM user generated.

The `iamGroups` field - A list of IAM group names. IAM users generated against this vault role will be added to these IAM Groups. For a credential type of `assumed_role` or `federation_token`, the policies sent to the corresponding AWS call will be the policies from each group in `iam_groups` combined with the `policy_document` and `policy_arns` parameters.

The `iamTags` field - A list of strings representing a key/value pair to be used as a tag for any `iam_user` user that is created by this role. Format is a key and value separated by an = (e.g. test_key=value).
	
The `defaultStsTtl` field - The default TTL for STS credentials. When a TTL is not specified when STS credentials are requested, and a default TTL is specified on the role, then this default TTL will be used. Valid only when `credentialType` is one of `assumed_role` or `federation_token`.
	
The `maxStsTtl` field - The max allowed TTL for STS credentials (credentials TTL are capped to max_sts_ttl). Valid only when `credentialType` is one of `assumed_role` or `federation_token`.
	
The `sessionTags` field - The set of key-value pairs to be included as tags for the STS session. Format is key=value. Valid only when `credentialType` is set to `assumed_role`.

The `externalId` field - The external ID to use when assuming the role. Valid only when `credentialType` is set to `assumed_role`.

The `userPath` field - The path for the user name. Valid only when `credentialType` is `iam_user`. Default is `/`

The `permissionsBoundaryArn` field - The ARN of the AWS Permissions Boundary to attach to IAM users created in the role. Valid only when `credentialType` is `iam_user`. If not specified, then no permissions boundary policy will be attached.
	
The `mfaSerialNumber` field - The ARN or hardware device number of the device configured to the IAM user for multi-factor authentication. Only required if the IAM user has an MFA device set up in AWS.

### IAM User

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: AWSSecretEngineRole
metadata:
  labels:
    app.kubernetes.io/name: awssecretenginerole
    app.kubernetes.io/instance: awssecretenginerole-sample
    app.kubernetes.io/part-of: vault-config-operator
    app.kubernetes.io/managed-by: kustomize
    app.kubernetes.io/created-by: vault-config-operator
  name: awssecretenginerole-sample-iam-user
spec:
  authentication:
    path: vault-admin
    role: vault-admin
  connection:
    address: 'https://vault.example.com'
  path: aws
  name: "iam-user-role"
  credentialType: iam_user
  policyArns:
    - arn:aws:iam::aws:policy/AmazonEC2ReadOnlyAccess
    - arn:aws:iam::aws:policy/IAMReadOnlyAccess
  iamGroups:
    - group1
    - group2
  policyDocument: |
    {
      "Version": "2012-10-17",
      "Statement": [
        {
          "Effect": "Allow",
          "Action": "ec2:*",
          "Resource": "*"
        }
      ]
    }
```

### Federation Token

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: AWSSecretEngineRole
metadata:
  labels:
    app.kubernetes.io/name: awssecretenginerole
    app.kubernetes.io/instance: awssecretenginerole-sample
    app.kubernetes.io/part-of: vault-config-operator
    app.kubernetes.io/managed-by: kustomize
    app.kubernetes.io/created-by: vault-config-operator
  name: awssecretenginerole-sample-federation-token
spec:
  authentication:
    path: vault-admin
    role: vault-admin
  connection:
    address: 'https://vault.example.com'
  path: aws
  name: "federation-token-role"
  credentialType: federation_token
  policyDocument: |
    {
      "Version": "2012-10-17",
      "Statement": {
        "Effect": "Allow",
        "Action": [
          "ec2:*",
          "sts:GetFederationToken"
        ],
        "Resource": "*"
      }
    }
```

### Session Token

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: AWSSecretEngineRole
metadata:
  labels:
    app.kubernetes.io/name: awssecretenginerole
    app.kubernetes.io/instance: awssecretenginerole-sample
    app.kubernetes.io/part-of: vault-config-operator
    app.kubernetes.io/managed-by: kustomize
    app.kubernetes.io/created-by: vault-config-operator
  name: awssecretenginerole-sample-session-token
spec:
  authentication:
    path: vault-admin
    role: vault-admin
  connection:
    address: 'https://vault.example.com'
  path: aws
  name: "session-token-role"
  credentialType: session_token
```

### Assume Role

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: AWSSecretEngineRole
metadata:
  labels:
    app.kubernetes.io/name: awssecretenginerole
    app.kubernetes.io/instance: awssecretenginerole-sample
    app.kubernetes.io/part-of: vault-config-operator
    app.kubernetes.io/managed-by: kustomize
    app.kubernetes.io/created-by: vault-config-operator
  name: awssecretenginerole-sample-assume-role
spec:
  authentication:
    path: vault-admin
    role: vault-admin
  connection:
    address: 'https://vault.example.com'
  path: aws
  name: "assume-role"
  credentialType: assumed_role
  roleArns:
    - arn:aws:iam::ACCOUNT-ID-WITHOUT-HYPHENS:role/RoleNameToAssume
```
