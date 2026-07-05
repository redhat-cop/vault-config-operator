# Kubernetes Secret Engine

[Kubernetes engine documentation](https://developer.hashicorp.com/vault/docs/secrets/kubernetes)

## Overview

The Kubernetes secret engine generates Kubernetes service account tokens, service accounts, and role bindings with scoped RBAC permissions. It enables applications to obtain short-lived Kubernetes credentials for cross-cluster or cross-namespace access without managing static service account tokens.

The vault-config-operator supports the following CRDs for the Kubernetes engine:

- [KubernetesSecretEngineConfig](#kubernetessecretengineconfig)
- [KubernetesSecretEngineRole](#kubernetessecretenginerole)

## KubernetesSecretEngineConfig

The `KubernetesSecretEngineConfig` CRD allows you to configure a [Kubernetes secret engine](https://developer.hashicorp.com/vault/api-docs/secret/kubernetes#write-configuration) connection.

### Example

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
  disableLocalCAJWT: false
  jwtReference:
    secret:
      name: sa-token-secret
```

> **`spec.authentication.path` vs `spec.path`:** `spec.authentication.path` is the auth mount the operator itself uses to authenticate with Vault. `spec.path` is the mount path of the Kubernetes secret engine being configured. They may point to different mounts.

### Vault CLI Equivalent

```shell
vault write [namespace/]<path>/config \
    kubernetes_host="https://kubernetes.default.svc:443" \
    service_account_jwt=<retrieved from jwtReference> \
    disable_local_ca_jwt=false
```

### Field Descriptions

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| path | string | Yes | — | Mount path of the Kubernetes secret engine. Full path: `[namespace/]{path}/config` |
| authentication | object | Yes | — | Kubernetes auth configuration. See [Authentication](../auth-section.md) |
| connection | object | No | — | Override Vault connection settings. See [Vault Connection](../contributing-vault-apis.md) |
| kubernetesHost | string | Yes | — | Kubernetes API URL to connect to |
| kubernetesCACert | string | No | — | PEM-encoded CA certificate to verify the Kubernetes API server certificate |
| disableLocalCAJWT | bool | No | `false` | Disable defaulting to the local CA certificate and service account JWT when running in a Kubernetes pod |
| jwtReference | object | Yes | — | Service account JWT for the Kubernetes engine connection. See [JWT Reference](#jwt-reference) |

## KubernetesSecretEngineRole

The `KubernetesSecretEngineRole` CRD allows you to create a [Kubernetes secret engine role](https://developer.hashicorp.com/vault/api-docs/secret/kubernetes#create-role) for generating dynamic Kubernetes credentials.

### Example

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: KubernetesSecretEngineRole
metadata:
  name: kubese-role
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  path: kubese-test
  allowedKubernetesNamespaces:
    - test-namespace
  kubernetesRoleName: my-cluster-role
  kubernetesRoleType: ClusterRole
  defaultTTL: 1h
  maxTTL: 24h
  nameTemplate: "v-{{ .RoleName }}-{{ unix_time }}"
  targetNamespaces:
    targetNamespaceSelector:
      matchLabels:
        app: my-app
```

### Vault CLI Equivalent

```shell
vault write [namespace/]<path>/roles/<role-name> \
    allowed_kubernetes_namespaces="test-namespace" \
    kubernetes_role_name=my-cluster-role \
    kubernetes_role_type=ClusterRole \
    token_default_ttl=1h \
    token_max_ttl=24h \
    name_template="v-{{ .RoleName }}-{{ unix_time }}"
```

### Field Descriptions

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| path | string | Yes | — | Mount path of the Kubernetes secret engine. Full Vault path: `[namespace/]{path}/roles/{name}` |
| authentication | object | Yes | — | Kubernetes auth configuration. See [Authentication](../auth-section.md) |
| connection | object | No | — | Override Vault connection settings. See [Vault Connection](../contributing-vault-apis.md) |
| name | string | No | — | Override the Vault object name. Defaults to `metadata.name` |
| targetNamespaces | object | Yes | — | Namespace binding configuration (same pattern as KubernetesAuthEngineRole) |
| allowedKubernetesNamespaces | []string | No | — | Kubernetes namespaces this role can generate credentials for. Use `["*"]` for all namespaces |
| allowedKubernetesNamespaceSelector | string | No | — | A label selector for Kubernetes namespaces in which credentials can be generated. Accepts JSON or YAML. If set with `allowedKubernetesNamespaces`, the conditions are ORed |
| serviceAccountName | string | No | — | Pre-existing service account to generate tokens for. Mutually exclusive with `kubernetesRoleName` and `generateRoleRules` |
| kubernetesRoleName | string | No | — | Pre-existing Role or ClusterRole to bind a generated service account to. Mutually exclusive with `serviceAccountName` and `generateRoleRules` |
| kubernetesRoleType | string | No | `"Role"` | Whether the Kubernetes role is a `Role` or `ClusterRole` |
| generateRoleRules | string | No | — | Role or ClusterRole rules (JSON or YAML) to use when generating a role. Mutually exclusive with `serviceAccountName` and `kubernetesRoleName` |
| defaultTTL | duration | No | — | Default TTL for the leases associated with this role |
| maxTTL | duration | No | — | Maximum TTL for the leases associated with this role |
| defaultAudiences | string | No | — | Default intended audiences for generated tokens, comma-separated |
| nameTemplate | string | No | — | Name template for generating service accounts, roles, and role bindings |
| extraAnnotations | map[string]string | No | — | Additional annotations to apply to all generated Kubernetes objects |
| extraLabels | map[string]string | No | — | Additional labels to apply to all generated Kubernetes objects |

### Credential Generation Modes

The role supports three mutually exclusive credential generation modes:

1. **`serviceAccountName`** — Use a pre-existing service account (only a token is generated)
2. **`kubernetesRoleName`** — Bind to a pre-existing Role or ClusterRole (service account, binding, and token are created)
3. **`generateRoleRules`** — Generate the Role/ClusterRole from inline rules (entire RBAC chain is created)

## JWT Reference

The service account JWT (used to authenticate the Kubernetes secret engine connection) can be provided in two ways via the `jwtReference` field. Exactly one of `secret` or `vaultSecret` must be specified — the webhook rejects manifests that set none, both, or `randomSecret`.

### Using a Kubernetes Secret

Specify the `secret` field. The secret must be of [service account token type](https://kubernetes.io/docs/concepts/configuration/secret/#service-account-token-secrets) (`kubernetes.io/service-account-token`). The JWT is read from the `token` data field.

```yaml
spec:
  jwtReference:
    secret:
      name: sa-token-secret
```

### Using a Vault Secret

Specify the `vaultSecret` field to retrieve the JWT from another Vault path. The JWT is read from the `key` field of the Vault secret.

```yaml
spec:
  jwtReference:
    vaultSecret:
      path: secret/data/kube-sa-jwt
```

## See Also

- [Authentication](../auth-section.md) — Common authentication section configuration
- [Contributing a New Vault API](../contributing-vault-apis.md) — Developer guide for adding new CRD types
- [Vault Kubernetes Secret Engine](https://developer.hashicorp.com/vault/docs/secrets/kubernetes) — Vault documentation
- [Vault Kubernetes Secret Engine API](https://developer.hashicorp.com/vault/api-docs/secret/kubernetes) — Vault API reference
