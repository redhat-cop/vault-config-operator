# {{EngineName}} {{Auth|Secret}} Engine

<!-- TEMPLATE INSTRUCTIONS:
     This template defines the required structure for all engine documentation files.
     Replace all {{placeholder}} values with actual content for the specific engine.
     Remove this instructions block and all HTML comments when creating the actual doc.
     
     Naming convention for files:
       Auth engines:  docs/auth-engines/{{engine-name-lowercase}}.md
       Secret engines: docs/secret-engines/{{engine-name-lowercase}}.md
-->

[{{EngineName}} engine documentation]({{vault_documentation_url}})

## Overview

<!-- Write 2-3 sentences describing what this engine does, what use cases it serves,
     and why a user would configure it via the operator. -->

{{EngineName}} provides {{brief description of engine functionality}}. It enables {{primary use case}} and integrates with {{external system or Vault capability}}.

The vault-config-operator supports the following CRDs for the {{EngineName}} engine:

- [{{EngineName}}{{Auth|Secret}}EngineConfig](#{{enginename}}{{auth|secret}}engineconfig)
- [{{EngineName}}{{Auth|Secret}}Engine{{Role|Group}}](#{{enginename}}{{auth|secret}}engine{{role|group}})

## {{EngineName}}{{Auth|Secret}}EngineConfig

<!-- Config CRD section: This documents the configuration/connection CRD for the engine.
     Every engine has a Config CRD that sets up the engine's connection or base settings. -->

The `{{EngineName}}{{Auth|Secret}}EngineConfig` CRD allows you to configure a [{{EngineName}} {{auth|secret}} engine]({{vault_api_docs_url_for_config}}).

<!-- EXAMPLE (filled in) for LDAPAuthEngineConfig:
     The `LDAPAuthEngineConfig` CRD allows you to configure an authentication engine mount of
     [type LDAP](https://developer.hashicorp.com/vault/docs/auth/ldap).
-->

### Example

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: {{EngineName}}{{Auth|Secret}}EngineConfig
metadata:
  name: {{enginename}}-config-sample
spec:
  authentication:
    path: kubernetes
    role: policy-admin
    # Optional: override service account for authentication
    # serviceAccount:
    #   name: admin-sa
  # Optional: override Vault connection settings
  # connection:
  #   address: https://vault.example.com:8200
  #   tls_config:
  #     ca_bundle: ...
  path: {{engine-mount-path}}
  {{config_specific_fields}}
```

<!-- EXAMPLE (filled in) for KubernetesAuthEngineConfig:
```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: KubernetesAuthEngineConfig
metadata:
  name: kubernetes-config
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  path: kubernetes
  tokenReviewerServiceAccount:
    name: token-review-sa
  kubernetesHost: https://kubernetes.default.svc:443
```
-->

### Vault CLI Equivalent

<!-- Show the Vault CLI command that achieves the same result as this CR.
     Use [namespace/] prefix to indicate optional namespace. -->

```shell
vault write [namespace/]{{auth_or_secrets_path}}/{{engine-mount-path}}/config {{key=value pairs matching the spec fields}}
```

<!-- EXAMPLE (filled in) for DatabaseSecretEngineConfig:
```shell
vault write [namespace/]postgresql-vault-demo/database/config/my-postgresql-database \
    plugin_name=postgresql-database-plugin \
    allowed_roles="read-write,read-only" \
    connection_url="postgresql://{{username}}:{{password}}@my-db.svc:5432" \
    username=<retrieved dynamically> \
    password=<retrieved dynamically>
```
-->

### Field Descriptions

<!-- Use a markdown table with camelCase field names (matching the CRD spec fields).
     Do NOT use snake_case Vault API names.
     Include all commonly used fields. Mark truly optional fields clearly.
     For authentication and connection fields, link to the shared docs rather than duplicating. -->

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| path | string | Yes | Path where the engine is mounted. Full path: `[namespace/]{{auth/|secrets/}}{path}/config` |
| authentication | object | Yes | Kubernetes auth configuration. See [Authentication](auth-section.md) |
| connection | object | No | Override Vault connection settings. See [Vault Connection](contributing-vault-apis.md) |
| {{field1}} | {{type}} | {{Yes/No}} | {{Description of what this field does}} |
| {{field2}} | {{type}} | {{Yes/No}} | {{Description of what this field does}} |

<!-- EXAMPLE (filled in) for KubernetesAuthEngineConfig:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| path | string | Yes | Path of the Kubernetes auth mount. Full path: `[namespace/]auth/{path}/config` (auth engines) |
| authentication | object | Yes | Kubernetes auth configuration. See [Authentication](auth-section.md) |
| connection | object | No | Override Vault connection settings. See [Vault Connection](contributing-vault-apis.md) |
| tokenReviewerServiceAccount | object | No | Service account for token review. Defaults to `default` |
| kubernetesHost | string | No | API server URL. Defaults to `https://kubernetes.default.svc:443` |
| kubernetesCACert | string | No | Base64-encoded CA certificate for API server validation |
-->

## {{EngineName}}{{Auth|Secret}}Engine{{Role|Group}}

<!-- Role/Group CRD section: This documents the role (or group) CRD for the engine.
     Roles define specific permissions, access policies, or credential templates.
     Some engines use "Group" instead of "Role" (e.g., LDAPAuthEngineGroup).
     Replace {{Role|Group}} with the appropriate variant for this engine. -->

The `{{EngineName}}{{Auth|Secret}}Engine{{Role|Group}}` CRD allows you to create a [{{EngineName}} {{role|group}}]({{vault_api_docs_url_for_role}}).

<!-- EXAMPLE (filled in) for KubernetesAuthEngineRole:
     The `KubernetesAuthEngineRole` CRD allows you to create a
     [Kubernetes Authentication Role](https://developer.hashicorp.com/vault/docs/auth/kubernetes#configuration).
-->

### Example

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: {{EngineName}}{{Auth|Secret}}Engine{{Role|Group}}
metadata:
  name: {{enginename}}-{{role|group}}-sample
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  # Optional: override Vault connection settings
  # connection:
  #   address: https://vault.example.com:8200
  #   tls_config:
  #     ca_bundle: ...
  path: {{engine-mount-path}}
  {{role_specific_fields}}
```

<!-- EXAMPLE (filled in) for KubernetesAuthEngineRole:
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
-->

### Vault CLI Equivalent

```shell
vault write [namespace/]{{auth_or_secrets_path}}/{{engine-mount-path}}/{{roles|groups}}/{{role-or-group-name}} {{key=value pairs matching the spec fields}}
```

<!-- EXAMPLE (filled in) for KubernetesAuthEngineRole:
```shell
vault write [namespace/]auth/kubernetes/role/database-engine-admin \
    bound_service_account_names=vaultsa \
    bound_service_account_namespaces=<dynamically generated> \
    policies=database-engine-admin
```
EXAMPLE (filled in) for LDAPAuthEngineGroup:
```shell
vault write [namespace/]auth/ldap/test/groups/test-3 \
    policies="admin, audit, users"
```
-->

### Field Descriptions

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| path | string | Yes | Path of the engine mount where the role will be created |
| authentication | object | Yes | Kubernetes auth configuration. See [Authentication](auth-section.md) |
| connection | object | No | Override Vault connection settings. See [Vault Connection](contributing-vault-apis.md) |
| {{field1}} | {{type}} | {{Yes/No}} | {{Description of what this field does}} |
| {{field2}} | {{type}} | {{Yes/No}} | {{Description of what this field does}} |

<!-- EXAMPLE (filled in) for KubernetesAuthEngineRole:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| path | string | Yes | Path of the Kubernetes auth mount where the role is created |
| authentication | object | Yes | Kubernetes auth configuration. See [Authentication](auth-section.md) |
| connection | object | No | Override Vault connection settings. See [Vault Connection](contributing-vault-apis.md) |
| policies | []string | No | Vault policies to associate with this role |
| targetServiceAccounts | []string | No | Service accounts that can authenticate. Defaults to `default` |
| targetNamespaceSelector | object | No | Label selector for namespaces allowed to authenticate |
| targetNamespaces | []string | No | Static list of namespaces allowed to authenticate |
| tokenTTL | string | No | TTL for generated tokens |
| tokenMaxTTL | string | No | Max TTL for generated tokens |
-->

## Credential Resolution

<!-- INCLUDE THIS SECTION ONLY for engines that require external credentials
     (e.g., database passwords, LDAP bind credentials, OIDC client secrets).
     Auth engines that only use Kubernetes auth tokens do NOT need this section.
     
     The credential field name varies by engine. Two patterns exist:
     
     PATTERN A — Flat prefix fields (most engines):
       - DatabaseSecretEngineConfig: rootCredentialsFrom{Secret,VaultSecret,RandomSecret}
       - LDAPAuthEngineConfig: bindCredentialsFrom{Secret,VaultSecret,RandomSecret}
       Replace {{credentialFieldPrefix}} with the engine-specific prefix.
     
     PATTERN B — Nested credential object (JWT/OIDC, Azure):
       - JWTOIDCAuthEngineConfig: OIDCCredentials.{secret,vaultSecret,randomSecret}
       - AzureSecretEngineConfig: azureCredentials.{secret,vaultSecret,randomSecret}
       Replace {{credentialObject}} with the parent field name (e.g., OIDCCredentials).
       Use the nested YAML examples shown below instead of Pattern A.
     
     Choose Pattern A or Pattern B based on how the engine's CRD defines its credential fields.
-->

The {{credential description, e.g., "password and possibly the username"}} can be retrieved in three different ways:

### Using a Kubernetes Secret

Specify the `{{credentialFieldPrefix}}FromSecret` field. The secret must be of [basic auth type](https://kubernetes.io/docs/concepts/configuration/secret/#basic-authentication-secret). If the secret is updated, this configuration will also be updated.

```yaml
spec:
  {{credentialFieldPrefix}}FromSecret:
    name: {{secret-name}}
```

<!-- EXAMPLE (filled in) for DatabaseSecretEngineConfig:
```yaml
spec:
  rootCredentialsFromSecret:
    name: postgresql-admin-password
```
-->

### Using a Vault Secret

Specify the `{{credentialFieldPrefix}}FromVaultSecret` field to retrieve credentials from another Vault path.

```yaml
spec:
  {{credentialFieldPrefix}}FromVaultSecret:
    path: {{vault-secret-path}}
    usernameKey: {{username-field-in-vault-secret}}
    passwordKey: {{password-field-in-vault-secret}}
```

<!-- EXAMPLE (filled in) for DatabaseSecretEngineConfig:
```yaml
spec:
  rootCredentialsFromVaultSecret:
    path: secret/data/db-credentials
    usernameKey: username
    passwordKey: password
```
-->

### Using a RandomSecret

Specify the `{{credentialFieldPrefix}}FromRandomSecret` field. When the [RandomSecret](secret-management.md#RandomSecret) generates a new secret, this configuration will also be updated.

```yaml
spec:
  {{credentialFieldPrefix}}FromRandomSecret:
    name: {{randomsecret-name}}
```

<!-- EXAMPLE (filled in) for DatabaseSecretEngineConfig:
```yaml
spec:
  rootCredentialsFromRandomSecret:
    name: db-random-password
```
-->

<!-- ====================================================================
     PATTERN B — Nested credential object (use INSTEAD OF Pattern A above
     when the engine uses a nested object like OIDCCredentials or azureCredentials).
     Delete the Pattern A sections and use these instead.
     ==================================================================== -->

<!-- PATTERN B: Using a Kubernetes Secret
```yaml
spec:
  {{credentialObject}}:
    secret:
      name: {{secret-name}}
    usernameKey: {{username-field-key}}
    passwordKey: {{password-field-key}}
```
EXAMPLE (filled in) for JWTOIDCAuthEngineConfig:
```yaml
spec:
  OIDCCredentials:
    secret:
      name: oidccredentials
    usernameKey: client_id
    passwordKey: client_secret
```
-->

<!-- PATTERN B: Using a Vault Secret
```yaml
spec:
  {{credentialObject}}:
    vaultSecret:
      path: {{vault-secret-path}}
    usernameKey: {{username-field-key}}
    passwordKey: {{password-field-key}}
```
-->

<!-- PATTERN B: Using a RandomSecret
```yaml
spec:
  {{credentialObject}}:
    randomSecret:
      name: {{randomsecret-name}}
    usernameKey: {{username-field-key}}
    passwordKey: {{password-field-key}}
```
-->

## See Also

- [Authentication](auth-section.md) — Common authentication section configuration
- [Contributing a New Vault API](contributing-vault-apis.md) — Developer guide for adding new CRD types
- {{Additional relevant links for this specific engine}}
