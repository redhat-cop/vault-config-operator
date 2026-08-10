# Story 12.1: Consul Secret Engine — Config and Role CRDs

Status: ready-for-dev

## Story

As an operator developer,
I want CRDs for ConsulSecretEngineConfig and ConsulSecretEngineRole,
So that Vault's Consul secret engine can be managed declaratively.

## Acceptance Criteria

1. **Given** a ConsulSecretEngineConfig CR is created with Consul address and token (via K8s Secret reference)
   **When** the reconciler processes it
   **Then** the Consul config is written to Vault at `{path}/config/access` and ReconcileSuccessful=True

2. **Given** a ConsulSecretEngineRole CR is created with policies/service identities
   **When** the reconciler processes it
   **Then** the role exists in Vault at `{path}/roles/{name}` and can generate dynamic Consul tokens

3. **Given** the ConsulSecretEngineConfig CR is deleted
   **When** the reconciler processes deletion
   **Then** the K8s object is removed but Vault config is **NOT** deleted (`IsDeletable=false` — Vault has no `DELETE /consul/config/access` endpoint)

4. **Given** the ConsulSecretEngineRole CR is deleted
   **When** the reconciler processes deletion
   **Then** the role is removed from Vault via `DELETE /consul/roles/{name}` and the CR is deleted from K8s

5. **Given** the ConsulSecretEngineRole CR spec is updated (e.g., `consulPolicies` changed)
   **When** the reconciler processes the update
   **Then** the Vault role reflects the updated value

6. **Given** any Consul CR is created or updated
   **When** the webhook validates it
   **Then** `spec.path` and `spec.name` immutability is enforced on updates, and credential source validation passes for config

## Tasks / Subtasks

- [ ] Task 1: Create `ConsulSecretEngineConfig` type (AC: 1, 3, 6)
  - [ ] 1.1: Create `api/v1alpha1/consulsecretengineconfig_types.go` — Spec with `Connection`, `Authentication`, `Path`, inline `ConsulSEConfig` struct, `RootCredentials` (RootCredentialConfig)
  - [ ] 1.2: Implement `VaultObject` interface — `GetPath()` returns `{path}/config/access`, `GetPayload()`, `IsEquivalentToDesiredState()`, `PrepareInternalValues()`, `IsDeletable()=false`
  - [ ] 1.3: Implement `ConditionsAware` interface — Status with `Conditions []metav1.Condition`
  - [ ] 1.4: Implement `setInternalCredentials()` — resolve Consul token from K8s Secret, VaultSecret, or RandomSecret
  - [ ] 1.5: Implement `toMap()` on `ConsulSEConfig` — convert to Vault API snake_case fields
  - [ ] 1.6: Implement `IsEquivalentToDesiredState()` — must delete `token` from desired state (Vault never returns it on read), then `filterPayloadToDesiredKeys`

- [ ] Task 2: Create `ConsulSecretEngineRole` type (AC: 2, 4, 5)
  - [ ] 2.1: Create `api/v1alpha1/consulsecretenginerole_types.go` — Spec with `Connection`, `Authentication`, `Path`, inline `ConsulSERole` struct, `Name`
  - [ ] 2.2: Implement `VaultObject` interface — `GetPath()` returns `{path}/roles/{name}`, `IsDeletable()=true`
  - [ ] 2.3: Implement `ConditionsAware` interface
  - [ ] 2.4: Implement `toMap()` on `ConsulSERole` — include all role fields with correct snake_case mapping
  - [ ] 2.5: Implement `IsEquivalentToDesiredState()` — use `removeUnsetFields` + `filterPayloadToDesiredKeys`

- [ ] Task 3: Create webhooks (AC: 6)
  - [ ] 3.1: Create `api/v1alpha1/consulsecretengineconfig_webhook.go` — `admission.Defaulter[*ConsulSecretEngineConfig]`, `admission.Validator[*ConsulSecretEngineConfig]`, immutable `spec.path`, credential validation via `ValidateCredentialSource()`
  - [ ] 3.2: Create `api/v1alpha1/consulsecretenginerole_webhook.go` — `admission.Defaulter[*ConsulSecretEngineRole]`, `admission.Validator[*ConsulSecretEngineRole]`, immutable `spec.path` + `spec.name`

- [ ] Task 4: Create controllers (AC: 1, 2, 3, 4, 5)
  - [ ] 4.1: Create `internal/controller/consulsecretengineconfig_controller.go` — embed `ReconcilerBase`, config reconcile flow (always-write pattern for token, same as AWS config), watches on `corev1.Secret` and `RandomSecret`
  - [ ] 4.2: Create `internal/controller/consulsecretenginerole_controller.go` — embed `ReconcilerBase`, standard VaultResource reconcile flow

- [ ] Task 5: Register in main.go (AC: 1, 2)
  - [ ] 5.1: Add controller registrations for `ConsulSecretEngineConfigReconciler` and `ConsulSecretEngineRoleReconciler`
  - [ ] 5.2: Add webhook registrations inside `ENABLE_WEBHOOKS` guard for both types

- [ ] Task 6: Unit tests (AC: 1, 2, 5, 6)
  - [ ] 6.1: Create `api/v1alpha1/consulsecretengineconfig_test.go` — test `toMap()`, test `IsEquivalentToDesiredState()` with independently constructed Vault-read-shaped fixtures (token stripping), negative test proving managed field mismatch returns `false`
  - [ ] 6.2: Create `api/v1alpha1/consulsecretenginerole_test.go` — test `toMap()` for consul_policies, consul_roles, service_identities, node_identities variants; test `IsEquivalentToDesiredState()`; negative tests

- [ ] Task 7: Test fixtures and integration tests (AC: 1, 2, 4, 5)
  - [ ] 7.1: Create test YAML fixtures in `test/consulsecretengine/` — config and role CRs
  - [ ] 7.2: Integration tests — **SKIP** (see Dev Notes: Consul is an external service that cannot be trivially installed in Kind for this secret engine's purposes; falls under "Skip it")

- [ ] Task 8: Code generation and validation (AC: all)
  - [ ] 8.1: Run `make manifests generate fmt vet test`
  - [ ] 8.2: Verify all existing tests still pass
  - [ ] 8.3: Add new CRDs to `config/crd/kustomization.yaml` under `resources` list

## Dev Notes

### Story Intelligence Chain — Predecessor Context

**Epic 11 (completed)** established the pattern for adding new secret engine CRDs (AWS 11.1, Transit 11.2, SSH 11.3). Story 12.1 follows the same pattern exactly.

Key learnings from Epic 11:
- **Always-write pattern for config types with write-only credentials**: AWS config controller uses a `manageReconcileLogic` that always writes (skipping `IsEquivalentToDesiredState`) because the token/secret_key is write-only — Vault never returns it on read, so drift detection cannot observe credential rotations. The Consul `config/access` endpoint has the same characteristic (token is write-only). Follow the AWS config controller pattern.
- **`removeUnsetFields` for clean drift detection**: AWS role uses `removeUnsetFields(desiredState, payload)` before `filterPayloadToDesiredKeys` to prevent false drift from zero-value fields that Vault omits. Apply the same pattern for ConsulSecretEngineRole.
- **`spec.name` immutability**: Every type with a Name override field must include `spec.name` in `ValidateUpdate` (Epic 11 action item, confirmed applied).
- **json.Number rule**: All numeric fields in `toMap()` must emit `json.Number`. However, none of the Consul fields are numeric (TTL/max_ttl are duration strings, `local` is bool), so this rule is not applicable here.
- **CRD registration checklist**: After `make manifests`, add new CRD YAML files to `config/crd/kustomization.yaml`. Missing this breaks Helm chart builds.
- **Shared framework protection**: Do NOT modify behavior in shared framework files (`reconcile_skeleton.go`, `vaultresourcereconciler.go`, etc.). Only add new exported functions/types.

### Integration Test Classification: SKIP

Per the project's Integration Test Infrastructure Philosophy:
> **Skip it** — For cloud providers and services that cannot be installed in Kind and are hard to mock (e.g., AWS, Azure, GCP, GitHub App, Quay), the integration test must not be run.

The Consul secret engine requires a running Consul cluster with ACL enabled for Vault to generate tokens against. While Consul *can* be installed in Kind, configuring ACL bootstrapping and management tokens is non-trivial infrastructure for a secret engine test. This is similar in complexity to cloud providers. Additionally, the AWS, Transit, and SSH stories in Epic 11 set the precedent: AWS was SKIP (cloud), Transit was SKIP (encryption-only, no external service), SSH used internal CA (no external service needed). Consul requires an external ACL-enabled Consul cluster, placing it in the SKIP category. Comprehensive unit tests for `toMap()`, `IsEquivalentToDesiredState()`, and webhook validation are the primary quality gate.

### Vault API Reference

**Config Access endpoint:** `POST {path}/config/access`

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `address` | string | required | Consul instance address as "host:port" (e.g., "127.0.0.1:8500") |
| `scheme` | string | "http" | URL scheme to use (http or https) |
| `token` | string | "" | Consul ACL management token. If not provided, Vault tries to bootstrap ACL automatically |
| `ca_cert` | string | "" | CA certificate for verifying Consul server certificate (x509 PEM) |
| `client_cert` | string | "" | Client certificate for Consul TLS (x509 PEM, requires client_key) |
| `client_key` | string | "" | Client key for Consul TLS (x509 PEM, requires client_cert) |

**No DELETE endpoint exists for `/consul/config/access`** → `ConsulSecretEngineConfig.IsDeletable()` must return `false`.

**GET response for config/access:** Vault returns the config but **NEVER** returns the `token` field. The `ca_cert`, `client_cert`, and `client_key` fields are also NOT returned on read. Only `address` and `scheme` are returned.

**Roles endpoint:** `POST/GET/DELETE {path}/roles/{name}`

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `consul_policies` | list | [] | Consul ACL policies to assign (Consul 1.4+) |
| `consul_roles` | list | [] | Consul roles to attach to token (Consul 1.5+) |
| `service_identities` | list | [] | Service identities to assign (Consul 1.5+) |
| `node_identities` | list | [] | Node identities to assign (Consul 1.8+) |
| `consul_namespace` | string | "" | Consul Enterprise namespace (Consul 1.7+) |
| `partition` | string | "" | Consul admin partition (Consul 1.11+) |
| `local` | bool | false | If true, token is not replicated globally (Consul 1.4+) |
| `ttl` | string | "" | TTL for generated tokens (duration format) |
| `max_ttl` | string | "" | Max TTL for generated tokens (duration format) |
| `token_type` | string | "client" | DEPRECATED (Consul 1.11) — "client" or "management" |
| `policy` | string | "" | DEPRECATED (Consul 1.11) — base64-encoded ACL policy |
| `policies` | list | [] | DEPRECATED — alias for `consul_policies` |

**At least one of `consul_policies`, `consul_roles`, `service_identities`, or `node_identities` is required.**

**GET response for roles:** Returns all set fields. Example:
```json
{
  "data": {
    "consul_policies": ["readonly"],
    "consul_roles": [],
    "service_identities": [],
    "node_identities": [],
    "consul_namespace": "",
    "partition": "",
    "local": false,
    "ttl": "1h",
    "max_ttl": "24h",
    "token_type": "client"
  }
}
```

### Critical: `IsEquivalentToDesiredState` for Config — Token Stripping

Vault never returns `token` on read from `config/access`. The `IsEquivalentToDesiredState` implementation must:
1. Build `desiredState` from `toMap()`
2. `delete(desiredState, "token")` — remove before comparison
3. Also delete `ca_cert`, `client_cert`, `client_key` — these are also not returned on read
4. Use `removeUnsetFields(desiredState, payload)` then `filterPayloadToDesiredKeys(desiredState, payload)` then `reflect.DeepEqual`

Follow the pattern from `AWSSecretEngineConfig` (deletes `secret_key`), `GitHubSecretEngineConfig` (deletes `prv_key`), `KubernetesSecretEngineConfig` (deletes `service_account_jwt`).

### Critical: Config Controller Uses Always-Write Pattern

Since `token` is write-only (never returned on read), drift detection cannot observe credential rotations. The config controller must use the **always-write pattern** (same as `AWSSecretEngineConfigReconciler.manageReconcileLogic`):
- Call `PrepareInternalValues` to resolve the token
- Always call `vaultEndpoint.Create(ctx)` (skip the standard `CreateOrUpdate` which reads first)
- This ensures credential rotations propagate to Vault

### ConsulSecretEngineConfig — Credential Resolution via RootCredentialConfig

The `token` must be resolved from one of three sources via `RootCredentialConfig`:
- **K8s Secret**: key `password` (or custom via `PasswordKey`) maps to the Consul management token
- **VaultSecret**: same key mapping from a Vault KV path
- **RandomSecret**: retrieve token from RandomSecret's Vault path

Pattern: follow `AWSSecretEngineConfig.setInternalCredentials()` exactly. Store resolved value in unexported field (`retrievedToken` with `json:"-"`) and include it in `toMap()` output.

**Key difference from AWS**: The Consul config only has ONE credential field (`token`), not two (`access_key` + `secret_key`). So the credential resolution is simpler — only need `PasswordKey` from the K8s Secret. No `UsernameKey` needed.

### ConsulSecretEngineConfig — `GetPath()` Is Fixed

Like AWS, the Consul config endpoint is always `{path}/config/access` (no per-name suffix). `GetPath()` must return `CleansePath(string(d.Spec.Path) + "/config/access")`.

The `Name` field is NOT needed for ConsulSecretEngineConfig since the path is fixed. Omit it from the spec to keep the type clean (unlike AWS which included it for consistency but never used it).

### CRD Field Spec — ConsulSecretEngineConfig

```go
type ConsulSEConfig struct {
    // Address specifies the Consul instance address as "host:port".
    // +kubebuilder:validation:Required
    Address string `json:"address"`

    // Scheme specifies the URL scheme (http or https).
    // +kubebuilder:validation:Optional
    // +kubebuilder:default="http"
    // +kubebuilder:validation:Enum:={"http","https"}
    Scheme string `json:"scheme"`

    // CACert is the CA certificate for verifying the Consul server certificate (x509 PEM).
    // +kubebuilder:validation:Optional
    CACert string `json:"caCert,omitempty"`

    // ClientCert is the client certificate for Consul TLS communication (x509 PEM).
    // If set, clientKey must also be set.
    // +kubebuilder:validation:Optional
    ClientCert string `json:"clientCert,omitempty"`

    // ClientKey is the client key for Consul TLS communication (x509 PEM).
    // If set, clientCert must also be set.
    // +kubebuilder:validation:Optional
    ClientKey string `json:"clientKey,omitempty"`

    retrievedToken string `json:"-"`
}
```

### ConsulSecretEngineConfig Spec Design

```go
type ConsulSecretEngineConfigSpec struct {
    // Connection represents the information needed to connect to Vault.
    // +kubebuilder:validation:Optional
    Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

    // Authentication is the kube auth configuration to be used to execute this request
    // +kubebuilder:validation:Required
    Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

    // Path at which to make the configuration.
    // The final path in Vault will be {[spec.authentication.namespace]}/{spec.path}/config/access.
    // +kubebuilder:validation:Required
    Path vaultutils.Path `json:"path,omitempty"`

    ConsulSEConfig `json:",inline"`

    // RootCredentials specifies how to retrieve the Consul ACL management token.
    // +kubebuilder:validation:Required
    RootCredentials vaultutils.RootCredentialConfig `json:"rootCredentials,omitempty"`
}
```

### CRD Field Spec — ConsulSecretEngineRole

```go
type ConsulSERole struct {
    // ConsulPolicies is the list of Consul ACL policies to assign to the generated token.
    // +kubebuilder:validation:Optional
    // +listType=set
    ConsulPolicies []string `json:"consulPolicies,omitempty"`

    // ConsulRoles is the list of Consul roles to attach to the generated token (Consul 1.5+).
    // +kubebuilder:validation:Optional
    // +listType=set
    ConsulRoles []string `json:"consulRoles,omitempty"`

    // ServiceIdentities is the list of Consul service identities to assign (Consul 1.5+).
    // +kubebuilder:validation:Optional
    // +listType=set
    ServiceIdentities []string `json:"serviceIdentities,omitempty"`

    // NodeIdentities is the list of Consul node identities to assign (Consul 1.8+).
    // +kubebuilder:validation:Optional
    // +listType=set
    NodeIdentities []string `json:"nodeIdentities,omitempty"`

    // ConsulNamespace specifies the Consul Enterprise namespace for the token (Consul 1.7+).
    // +kubebuilder:validation:Optional
    ConsulNamespace string `json:"consulNamespace,omitempty"`

    // Partition specifies the Consul admin partition for the token (Consul 1.11+).
    // +kubebuilder:validation:Optional
    Partition string `json:"partition,omitempty"`

    // Local if true creates a token that is not replicated globally (Consul 1.4+).
    // +kubebuilder:validation:Optional
    Local bool `json:"local,omitempty"`

    // TTL specifies the TTL for the generated Consul token. Uses duration format strings.
    // +kubebuilder:validation:Optional
    TTL string `json:"ttl,omitempty"`

    // MaxTTL specifies the max TTL for the generated Consul token. Uses duration format strings.
    // +kubebuilder:validation:Optional
    MaxTTL string `json:"maxTTL,omitempty"`
}
```

**Design decisions:**
- Deprecated fields (`token_type`, `policy`, `policies`) are **omitted** from the CRD. Modern Consul (1.4+) uses `consul_policies`, `consul_roles`, `service_identities`, and `node_identities`. There is no value in exposing deprecated API fields in a new CRD.
- `serviceIdentities` and `nodeIdentities` use `[]string` — each string is formatted as `"servicename:dc1,dc2"` or `"nodename:dc1"` per the Vault API.

### ConsulSecretEngineRole Spec Design

```go
type ConsulSecretEngineRoleSpec struct {
    // Connection represents the information needed to connect to Vault.
    // +kubebuilder:validation:Optional
    Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

    // Authentication is the kube auth configuration to be used to execute this request
    // +kubebuilder:validation:Required
    Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

    // Path at which to create the role.
    // The final path in Vault will be {[spec.authentication.namespace]}/{spec.path}/roles/{metadata.name}.
    // +kubebuilder:validation:Required
    Path vaultutils.Path `json:"path,omitempty"`

    ConsulSERole `json:",inline"`

    // The name of the object created in Vault. If this is specified it takes precedence over {metadata.name}
    // +kubebuilder:validation:Optional
    // +kubebuilder:validation:Pattern:=`[a-z0-9]([-a-z0-9]*[a-z0-9])?`
    Name string `json:"name,omitempty"`
}
```

### toMap() for ConsulSEConfig

```go
func (i *ConsulSEConfig) toMap() map[string]any {
    payload := map[string]any{}
    payload["address"] = i.Address
    payload["scheme"] = i.Scheme
    if i.retrievedToken != "" {
        payload["token"] = i.retrievedToken
    }
    payload["ca_cert"] = i.CACert
    payload["client_cert"] = i.ClientCert
    payload["client_key"] = i.ClientKey
    return payload
}
```

### toMap() for ConsulSERole

```go
func (i *ConsulSERole) toMap() map[string]any {
    payload := map[string]any{}
    payload["consul_policies"] = toInterfaceArray(i.ConsulPolicies)
    payload["consul_roles"] = toInterfaceArray(i.ConsulRoles)
    payload["service_identities"] = toInterfaceArray(i.ServiceIdentities)
    payload["node_identities"] = toInterfaceArray(i.NodeIdentities)
    payload["consul_namespace"] = i.ConsulNamespace
    payload["partition"] = i.Partition
    payload["local"] = i.Local
    payload["ttl"] = i.TTL
    payload["max_ttl"] = i.MaxTTL
    return payload
}
```

### IsEquivalentToDesiredState for ConsulSERole

```go
func (d *ConsulSecretEngineRole) IsEquivalentToDesiredState(payload map[string]any) bool {
    desiredState := d.Spec.ConsulSERole.toMap()
    removeUnsetFields(desiredState, payload)
    filteredPayload := filterPayloadToDesiredKeys(desiredState, payload)
    setFields := []string{"consul_policies", "consul_roles", "service_identities", "node_identities"}
    for _, key := range setFields {
        sortAnyStringSlice(desiredState, key)
        sortAnyStringSlice(filteredPayload, key)
    }
    return reflect.DeepEqual(desiredState, filteredPayload)
}
```

**Note:** `sortAnyStringSlice` is already defined in `awssecretenginerole_types.go`. Since it's in the same package (`v1alpha1`), it can be called directly. Do NOT redefine it.

### File Structure (New Files)

```
api/v1alpha1/consulsecretengineconfig_types.go      (NEW)
api/v1alpha1/consulsecretengineconfig_webhook.go     (NEW)
api/v1alpha1/consulsecretengineconfig_test.go        (NEW)
api/v1alpha1/consulsecretenginerole_types.go         (NEW)
api/v1alpha1/consulsecretenginerole_webhook.go       (NEW)
api/v1alpha1/consulsecretenginerole_test.go          (NEW)
internal/controller/consulsecretengineconfig_controller.go       (NEW)
internal/controller/consulsecretenginerole_controller.go         (NEW)
test/consulsecretengine/consul-secret-engine-config.yaml         (NEW)
test/consulsecretengine/consul-secret-engine-role.yaml           (NEW)
```

### Files to Update

```
cmd/main.go                                       (UPDATE - register controllers + webhooks)
config/crd/kustomization.yaml                     (UPDATE - add new CRD resources)
api/v1alpha1/zz_generated.deepcopy.go             (UPDATE - auto-generated by make generate)
config/crd/bases/                                 (UPDATE - auto-generated CRD YAMLs)
config/rbac/role.yaml                             (UPDATE - auto-generated RBAC)
config/webhook/manifests.yaml                     (UPDATE - auto-generated webhook config)
```

### Files Being Modified — Current State

**`cmd/main.go`**: Currently registers ~35+ controllers and ~35+ webhooks. New registrations follow the exact same pattern:
- Controller: `(&controller.ConsulSecretEngineConfigReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "ConsulSecretEngineConfig")}).SetupWithManager(mgr)`
- Webhook (inside ENABLE_WEBHOOKS): `(&redhatcopv1alpha1.ConsulSecretEngineConfig{}).SetupWebhookWithManager(mgr)`

No existing behavior is changed — this is purely additive.

### Existing Patterns to Follow (Reference Files)

| Pattern | Reference File |
|---------|---------------|
| Config type with RootCredentialConfig (most recent) | `api/v1alpha1/awssecretengineconfig_types.go` |
| Config credential resolution | `AWSSecretEngineConfig.setInternalCredentials()` |
| Secret-stripping in IsEquivalentToDesiredState | `api/v1alpha1/awssecretengineconfig_types.go` (deletes `secret_key`) |
| Config always-write controller | `internal/controller/awssecretengineconfig_controller.go` (manageReconcileLogic) |
| Role type (simple, no credentials) | `api/v1alpha1/awssecretenginerole_types.go` |
| Role with set field sorting | `AWSSecretEngineRole.IsEquivalentToDesiredState()` (sortAnyStringSlice) |
| Webhook pattern | `api/v1alpha1/awssecretengineconfig_webhook.go` |
| Controller with credential watches | `internal/controller/awssecretengineconfig_controller.go` |
| Controller (simple role) | `internal/controller/awssecretenginerole_controller.go` |
| filterPayloadToDesiredKeys helper | `api/v1alpha1/payload_filter.go` |
| removeUnsetFields helper | `api/v1alpha1/payload_filter.go` |
| sortAnyStringSlice helper | `api/v1alpha1/awssecretenginerole_types.go` (reuse, do NOT redefine) |

### RBAC Markers

Config controller:
```go
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=consulsecretengineconfigs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=consulsecretengineconfigs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=consulsecretengineconfigs/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=serviceaccounts/token,verbs=create
//+kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;patch
```

Role controller:
```go
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=consulsecretengineroles,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=consulsecretengineroles/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=consulsecretengineroles/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=serviceaccounts/token,verbs=create
//+kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;patch
```

### Kubebuilder Markers Checklist

- All exported Spec fields: `+kubebuilder:validation:Required` or `+kubebuilder:validation:Optional`
- `Scheme`: `+kubebuilder:default="http"`, `+kubebuilder:validation:Enum:={"http","https"}`, no `omitempty` (non-zero default)
- `Address`: `+kubebuilder:validation:Required`
- List fields (consulPolicies, consulRoles, etc.): `+listType=set`
- Root type: `+kubebuilder:object:root=true`, `+kubebuilder:subresource:status`
- Conditions field: `+patchMergeKey=type`, `+patchStrategy=merge`, `+listType=map`, `+listMapKey=type`
- Webhook paths: `/mutate-redhatcop-redhat-io-v1alpha1-consulsecretengineconfig`, `/validate-redhatcop-redhat-io-v1alpha1-consulsecretengineconfig` (all lowercase)

### Unit Test Requirements

**Config tests (`consulsecretengineconfig_test.go`):**
1. `TestConsulSecretEngineConfig_toMap` — verify snake_case keys (`address`, `scheme`, `token`, `ca_cert`, `client_cert`, `client_key`), verify resolved token appears
2. `TestConsulSecretEngineConfig_IsEquivalentToDesiredState_Match` — construct Vault-read-shaped payload (only `address` + `scheme`, no `token`/`ca_cert`/`client_cert`/`client_key`), verify returns `true`
3. `TestConsulSecretEngineConfig_IsEquivalentToDesiredState_Mismatch` — change `address`, verify returns `false`
4. `TestConsulSecretEngineConfig_IsEquivalentToDesiredState_ExtraVaultFields` — add extra Vault fields, verify still returns `true`
5. `TestConsulSecretEngineConfig_IsEquivalentToDesiredState_TokenInPayload` — ensure `token` in payload doesn't cause false drift

**Role tests (`consulsecretenginerole_test.go`):**
1. `TestConsulSecretEngineRole_toMap_Policies` — verify list fields use `toInterfaceArray`, verify snake_case keys
2. `TestConsulSecretEngineRole_toMap_ServiceIdentities` — verify service_identities format
3. `TestConsulSecretEngineRole_IsEquivalentToDesiredState_Match` — construct Vault-read fixture, verify `true`
4. `TestConsulSecretEngineRole_IsEquivalentToDesiredState_Mismatch` — change a field, verify `false`
5. `TestConsulSecretEngineRole_IsEquivalentToDesiredState_UnsetFields` — verify `removeUnsetFields` prevents false drift from zero-value fields Vault omits

### Critical Gotchas

1. **`token` is never returned by Vault on read** — `IsEquivalentToDesiredState` MUST delete `token` from desiredState. Also delete `ca_cert`, `client_cert`, `client_key` (not returned on read).

2. **Always-write controller pattern** — Since the token is write-only, the config controller must use the always-write pattern (direct `vaultEndpoint.Create` without read-compare). Copy the `manageReconcileLogic` method from `AWSSecretEngineConfigReconciler`.

3. **No `Name` field on Config** — The config path is fixed (`config/access`). Do NOT add a `Name` field to `ConsulSecretEngineConfigSpec`. The `spec.name` immutability check in `ValidateUpdate` is NOT needed for config since there's no Name field — only check `spec.path`.

4. **`sortAnyStringSlice` already exists** — It's defined in `awssecretenginerole_types.go` in the same package. Do NOT redefine it. Just call it directly.

5. **`toInterfaceArray` already exists** — Used throughout the codebase for `[]string` to `[]any` conversion. Use it for all list fields in the role's `toMap()`.

6. **CRD registration in kustomization.yaml** — After `make manifests`, add `redhatcop.redhat.io_consulsecretengineconfigs.yaml` and `redhatcop.redhat.io_consulsecretengineroles.yaml` to `config/crd/kustomization.yaml` resources list.

7. **Webhook registration uses lowercase type name** — Paths: `/mutate-redhatcop-redhat-io-v1alpha1-consulsecretengineconfig` and `/validate-redhatcop-redhat-io-v1alpha1-consulsecretengineconfig` (no hyphens in type name).

8. **`local` bool field** — Uses `omitempty` per CRD field rules (zero-value default `false`). Vault returns `local: false` in read response, so `removeUnsetFields` will keep it if Vault returns it, or remove it if Vault omits it and we have `false`.

9. **Consul credential resolution only needs one field** — Unlike AWS (access_key + secret_key), Consul only needs `token`. The `RootCredentialConfig.PasswordKey` maps to the token. No `UsernameKey` needed. When using RandomSecret, the user does NOT need a separate `spec.token` field for username override (unlike AWS which needs `spec.accessKey`).

### Anti-Patterns / DO NOT

- **DO NOT** add deprecated Vault API fields (`token_type`, `policy`, `policies`) to the CRD — these are from Consul < 1.4 and removed in Consul 1.11
- **DO NOT** add a `Name` field to ConsulSecretEngineConfig — the path is fixed at `config/access`
- **DO NOT** redefine `sortAnyStringSlice` or `toInterfaceArray` — they exist in the same package
- **DO NOT** modify shared framework files (`reconcile_skeleton.go`, `vaultresourcereconciler.go`, etc.)
- **DO NOT** write integration tests — Consul is classified as SKIP
- **DO NOT** use `CreateOrUpdate` for the config controller — use the always-write pattern (token is write-only)
- **DO NOT** emit `json.Number` for any fields — Consul has no integer fields that need it (TTL/max_ttl are strings, `local` is bool)
- **DO NOT** add `isValid()` validation requiring at least one of consulPolicies/consulRoles/serviceIdentities/nodeIdentities — Vault itself enforces this requirement and returns a clear error. Adding duplicate validation in the webhook creates coupling to Vault's evolving rules.

### Project Structure Notes

- All files follow existing naming conventions: lowercase type name for files
- New files go in established directories (`api/v1alpha1/`, `internal/controller/`, `test/`)
- Test fixture directory `test/consulsecretengine/` follows the existing pattern (`test/awssecretengine/`, `test/ssh/`)
- Controllers in `internal/controller/` (go/v4 layout since Epic 10)
- No conflicts with existing code — this is entirely additive

### References

- [Source: _bmad-output/planning-artifacts/epics.md#Epic-12, Story 12.1 — requirements and acceptance criteria]
- [Source: _bmad-output/project-context.md — all project rules, patterns, and conventions]
- [Source: api/v1alpha1/awssecretengineconfig_types.go — config type with RootCredentialConfig and always-write pattern (most recent reference)]
- [Source: api/v1alpha1/awssecretenginerole_types.go — role type with sortAnyStringSlice, removeUnsetFields]
- [Source: api/v1alpha1/awssecretengineconfig_webhook.go — webhook pattern with credential validation]
- [Source: internal/controller/awssecretengineconfig_controller.go — always-write config controller with Secret/RandomSecret watches]
- [Source: internal/controller/awssecretenginerole_controller.go — simple role controller]
- [Source: api/v1alpha1/payload_filter.go — filterPayloadToDesiredKeys + removeUnsetFields helpers]
- [Vault API: https://developer.hashicorp.com/vault/api-docs/secret/consul — Consul secrets engine API reference]
- [Source: _bmad-output/implementation-artifacts/11-1-aws-secret-engine-config-and-role-crds.md — predecessor story with patterns]
- [Source: _bmad-output/implementation-artifacts/11-3-ssh-secret-engine-config-and-role-crds.md — latest predecessor story]

## Code Review Record

### Review Model Used

(To be filled during review — must differ from dev model)

### Review Findings

### Decisions Needed / Decisions Taken

### Fixes Applied

## Dev Agent Record

### Agent Model Used

{{agent_model_name_version}}

### Debug Log References

### Completion Notes List

### File List
