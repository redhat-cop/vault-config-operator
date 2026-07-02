# Kubernetes Auth Engine

[Kubernetes engine documentation](https://developer.hashicorp.com/vault/docs/auth/kubernetes)

## Overview

The Kubernetes auth method allows Kubernetes service accounts to authenticate with Vault by verifying their JWT tokens against the Kubernetes API. It is the most common auth method when running Vault alongside Kubernetes workloads, enabling pods to obtain Vault tokens without managing static credentials.

The vault-config-operator supports the following CRDs for the Kubernetes engine:

- [KubernetesAuthEngineConfig](#kubernetesauthengineconfig)
- [KubernetesAuthEngineRole](#kubernetesauthenginerole)

## KubernetesAuthEngineConfig

The `KubernetesAuthEngineConfig` CRD allows you to configure a [Kubernetes auth engine](https://developer.hashicorp.com/vault/api-docs/auth/kubernetes#configure-kubernetes-method).

### Example

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

> **`spec.authentication.path` vs `spec.path`:** `spec.authentication.path` is the auth mount the operator itself uses to authenticate with Vault. `spec.path` is the mount path of the Kubernetes auth engine being configured. They may point to different mounts.

### Vault CLI Equivalent

```shell
vault write [namespace/]auth/<path>/<name>/config \
    kubernetes_host="https://kubernetes.default.svc:443" \
    kubernetes_ca_cert=@ca.pem \
    token_reviewer_jwt="<jwt-token>"
```

> **Note:** The operator resolves some fields automatically. `tokenReviewerServiceAccount` in the CR causes the operator to create a short-lived JWT and send it as `token_reviewer_jwt`. Similarly, the CA cert may be injected based on `useOperatorPodCA` — see the [CA resolution table](#kubernetescacert-behavior).

### Field Descriptions

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| path | string | Yes | Mount path for the Kubernetes auth engine. Full path: `[namespace/]auth/{path}/{name}/config` |
| authentication | object | Yes | Kubernetes auth configuration. See [Authentication](../auth-section.md) |
| connection | object | No | Override Vault connection settings. See [Vault Connection](../contributing-vault-apis.md) |
| name | string | No | Override the Vault object name. Defaults to `metadata.name` |
| tokenReviewerServiceAccount | object | No | Service account for token review. If not set, the JWT submitted in the login payload is used to access the Kubernetes TokenReview API |
| kubernetesHost | string | Yes | Kubernetes API server URL. Defaults to `https://kubernetes.default.svc:443` |
| kubernetesCACert | string | No | PEM-encoded CA cert for the Kubernetes API TLS client. See the [CA resolution table](#kubernetescacert-behavior) below |
| PEMKeys | []string | No | PEM-formatted public keys or certificates to verify Kubernetes service account JWT signatures |
| issuer | string | No | JWT issuer. Defaults to `kubernetes/serviceaccount` |
| disableISSValidation | bool | No | Disable JWT issuer validation. Defaults to `false` |
| disableLocalCAJWT | bool | No | Disable defaulting to the local CA cert and service account JWT when running in a Kubernetes pod. Defaults to `false` |
| useOperatorPodCA | bool | No | When `kubernetesCACert` is unset and `disableLocalCAJWT` is `true`, inject the operator pod's CA cert. Defaults to `true`. This field is not sent to Vault |
| useAnnotationsAsAliasMetadata | bool | No | Use annotations from the client token's service account as alias metadata for the Vault entity. Only `vault.hashicorp.com/alias-metadata-*` annotations are captured (512 character limit). Defaults to `false` |

### kubernetesCACert Behavior

The CA certificate used by Vault to validate Kubernetes API requests depends on the combination of three fields:

| `kubernetesCACert` | `disableLocalCAJWT` | `useOperatorPodCA` | Behaviour |
|--------------------|---------------------|--------------------|-----------|
| set | ignored | ignored | The provided CA cert is used |
| unset | false | ignored | Vault pod's `/var/run/secrets/kubernetes.io/serviceaccount/ca.crt` is used |
| unset | true | false | The default OS CA where Vault runs is used |
| unset | true | true | The operator pod's CA cert is injected and used |

## KubernetesAuthEngineRole

The `KubernetesAuthEngineRole` CRD allows you to create a [Kubernetes Authentication Role](https://developer.hashicorp.com/vault/api-docs/auth/kubernetes#create-role).

### Example

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
  targetNamespaces:
    targetNamespaceSelector:
      matchLabels:
        postgresql-enabled: "true"
```

### Vault CLI Equivalent

```shell
vault write [namespace/]auth/<path>/role/<role-name> \
    bound_service_account_names=vaultsa \
    bound_service_account_namespaces=<dynamically resolved> \
    token_policies=database-engine-admin \
    alias_name_source=serviceaccount_uid \
    token_type=default
```

### Field Descriptions

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| path | string | Yes | Mount path of the Kubernetes auth engine where the role is created. Full path: `[namespace/]auth/{path}/role/{name}` |
| authentication | object | Yes | Kubernetes auth configuration. See [Authentication](../auth-section.md) |
| connection | object | No | Override Vault connection settings. See [Vault Connection](../contributing-vault-apis.md) |
| name | string | No | Override the Vault object name. Defaults to `metadata.name` |
| targetNamespaces | object | Yes | Namespace binding configuration. See [Target Namespaces](#target-namespaces) below |
| targetServiceAccounts | []string | Yes | Service accounts that can authenticate via this role. Minimum 1 entry. Defaults to `["default"]` |
| policies | []string | Yes | Vault policies to associate with this role. Minimum 1 entry |
| audience | *string | No | Audience claim to verify in the JWT |
| aliasNameSource | string | No | Identity alias generation strategy: `serviceaccount_uid` or `serviceaccount_name`. Defaults to `serviceaccount_uid` |
| tokenTTL | int | No | Incremental lifetime for generated tokens (seconds). Defaults to `0` |
| tokenMaxTTL | int | No | Maximum lifetime for generated tokens (seconds). Defaults to `0` |
| tokenBoundCIDRs | []string | No | CIDR blocks restricting token usage |
| tokenExplicitMaxTTL | int | No | Hard cap TTL overriding `tokenTTL` and `tokenMaxTTL` (seconds). Defaults to `0` |
| tokenNoDefaultPolicy | bool | No | Exclude the `default` policy from generated tokens. Defaults to `false` |
| tokenNumUses | int | No | Maximum token usage count. `0` means unlimited. Defaults to `0` |
| tokenPeriod | int | No | Maximum allowed period for periodic token requests (seconds). Defaults to `0` |
| tokenType | string | No | Token type: `service`, `batch`, `default`, `default-service`, or `default-batch`. Defaults to `default` |

### Target Namespaces

The `targetNamespaces` field determines which Kubernetes namespaces are bound to the Vault role. It contains exactly **one** of the following sub-fields (the webhook validates mutual exclusivity):

- **`targetNamespaceSelector`** — A Kubernetes [label selector](https://kubernetes.io/docs/concepts/overview/working-with-objects/labels/#label-selectors). Namespaces are resolved dynamically at each reconcile. An empty selector (`{}`) matches **all** namespaces in the cluster. If the selector matches zero namespaces, the role is written with `bound_service_account_namespaces=["__no_namespace__"]` to avoid Vault API errors.

  ```yaml
  targetNamespaces:
    targetNamespaceSelector:
      matchLabels:
        environment: production
  ```

- **`targetNamespaces`** — A static list of namespace names. Must contain at least one entry.

  ```yaml
  targetNamespaces:
    targetNamespaces:
      - namespace-a
      - namespace-b
  ```

## See Also

- [Authentication](../auth-section.md) — Common authentication section configuration
- [Contributing a New Vault API](../contributing-vault-apis.md) — Developer guide for adding new CRD types
- [Vault Kubernetes Auth Method](https://developer.hashicorp.com/vault/docs/auth/kubernetes) — Vault documentation
- [Vault Kubernetes Auth API](https://developer.hashicorp.com/vault/api-docs/auth/kubernetes) — Vault API reference
