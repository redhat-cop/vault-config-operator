# GitHub Secret Engine

[vault-plugin-secrets-github](https://github.com/martinbaillie/vault-plugin-secrets-github)

## Overview

The GitHub secret engine is a third-party plugin (vault-plugin-secrets-github v2.0.0) that generates ephemeral GitHub App installation access tokens with scoped permissions. It allows Kubernetes workloads to obtain short-lived GitHub tokens without managing static credentials, enabling fine-grained access to repositories and organizations.

The vault-config-operator supports the following CRDs for the GitHub engine:

- [GitHubSecretEngineConfig](#githubsecretengineconfig)
- [GitHubSecretEngineRole](#githubsecretenginerole)

## GitHubSecretEngineConfig

The `GitHubSecretEngineConfig` CRD allows you to configure a [GitHub secret engine](https://github.com/martinbaillie/vault-plugin-secrets-github#config).

### Example

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: GitHubSecretEngineConfig
metadata:
  name: github-config
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  path: github
  applicationID: 123456
  gitHubAPIBaseURL: "https://api.github.com"
  sSHKeyReference:
    secret:
      name: github-app-private-key
```

> **`spec.authentication.path` vs `spec.path`:** `spec.authentication.path` is the auth mount the operator itself uses to authenticate with Vault. `spec.path` is the mount path of the GitHub secret engine being configured. They may point to different mounts.

### Vault CLI Equivalent

```shell
vault write [namespace/]<path>/config \
    app_id=123456 \
    prv_key=@private-key.pem \
    base_url="https://api.github.com"
```

### Field Descriptions

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| path | string | Yes | — | Mount path of the GitHub secret engine. Full path: `[namespace/]{path}/config` |
| authentication | object | Yes | — | Kubernetes auth configuration. See [Authentication](../auth-section.md) |
| connection | object | No | — | Override Vault connection settings. See [Vault Connection](../contributing-vault-apis.md) |
| applicationID | int64 | Yes | — | The Application ID of the GitHub App. Minimum value: `1` |
| gitHubAPIBaseURL | string | No | `"https://api.github.com"` | The base URL for GitHub API requests |
| sSHKeyReference | object | Yes | — | SSH private key for the GitHub App. See [SSH Key Resolution](#ssh-key-resolution) |

> **Note:** Deleting the `GitHubSecretEngineConfig` CR does **not** remove the config from Vault. The configuration can only be removed by deleting the entire engine mount.

## GitHubSecretEngineRole

The `GitHubSecretEngineRole` CRD allows you to create a [GitHub secret engine permission set](https://github.com/martinbaillie/vault-plugin-secrets-github#permission-sets) for generating scoped installation access tokens.

### Example

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: GitHubSecretEngineRole
metadata:
  name: my-github-permissionset
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  path: github
  organizationName: my-org
  repositories:
    - my-repo
    - my-other-repo
  permissions:
    contents: read
    pull_requests: write
```

### Vault CLI Equivalent

```shell
vault write [namespace/]<path>/permissionset/<name> \
    org_name=my-org \
    repositories="my-repo,my-other-repo" \
    permissions='{"contents":"read","pull_requests":"write"}'
```

> **Note:** The GitHub plugin uses `permissionset` (not `roles`) in the Vault path. To generate a token, read from `{path}/token/{name}`.

### Field Descriptions

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| path | string | Yes | — | Mount path of the GitHub secret engine. Full Vault path: `[namespace/]{path}/permissionset/{name}` |
| authentication | object | Yes | — | Kubernetes auth configuration. See [Authentication](../auth-section.md) |
| connection | object | No | — | Override Vault connection settings. See [Vault Connection](../contributing-vault-apis.md) |
| name | string | No | — | Override the Vault object name. Defaults to `metadata.name` |
| installationID | int64 | No | — | The ID of the GitHub App installation. Only one of `installationID` or `organizationName` is required. If both are provided, `installationID` takes precedence |
| organizationName | string | No | — | The name of the organization with the GitHub App installation. Only one of `installationID` or `organizationName` is required |
| repositories | []string | No | — | Names of the repositories within the organization that the installation token can access |
| repositoriesIDs | []string | No | — | IDs of the repositories that the installation token can access. Repository IDs are immutable and preferred for security-sensitive configurations |
| permissions | map[string]string | No | — | Permission names mapped to their access type (`read` or `write`). Omitting results in a token with all permissions the GitHub App has. See [GitHub permissions documentation](https://developer.github.com/v3/apps/permissions) |

## SSH Key Resolution

The GitHub App's private SSH key (used to authenticate with the GitHub API) can be provided in two ways via the `sSHKeyReference` field. Exactly one of `secret` or `vaultSecret` must be specified — the webhook rejects manifests that set none or both.

### Using a Kubernetes Secret

Specify the `secret` field. The secret must be of [SSH auth type](https://kubernetes.io/docs/concepts/configuration/secret/#ssh-authentication-secrets) (`kubernetes.io/ssh-auth`). The private key is read from the `ssh-privatekey` data field. If the secret is updated, this configuration will also be updated.

```yaml
spec:
  sSHKeyReference:
    secret:
      name: github-app-private-key
```

### Using a Vault Secret

Specify the `vaultSecret` field to retrieve the private key from another Vault path. The key is read from the `key` field of the Vault secret.

```yaml
spec:
  sSHKeyReference:
    vaultSecret:
      path: secret/data/github-app-key
```

## See Also

- [Authentication](../auth-section.md) — Common authentication section configuration
- [Contributing a New Vault API](../contributing-vault-apis.md) — Developer guide for adding new CRD types
- [vault-plugin-secrets-github](https://github.com/martinbaillie/vault-plugin-secrets-github) — Plugin documentation and source
