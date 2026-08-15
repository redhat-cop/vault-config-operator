---
baseline_commit: 2ccad3a67685ec8bdd0c585247ee148143415e6c
---

# Story 13.3: MongoDB Atlas Secret Engine — Config and Role CRDs

Status: done

## Story

As an operator developer,
I want CRDs for MongoDBAtlasSecretEngineConfig and MongoDBAtlasSecretEngineRole,
So that Vault's MongoDB Atlas secret engine can be managed declaratively.

## Acceptance Criteria

1. **Given** a MongoDBAtlasSecretEngineConfig CR is created with MongoDB Atlas API credentials (public_key + private_key via K8s Secret reference)
   **When** the reconciler processes it
   **Then** the Atlas config is written to Vault at `{path}/config` and ReconcileSuccessful=True

2. **Given** a MongoDBAtlasSecretEngineRole CR is created with organization_id or project_id plus roles
   **When** the reconciler processes it
   **Then** the role exists in Vault at `{path}/roles/{name}` and can generate dynamic Atlas Programmatic API keys

3. **Given** the MongoDBAtlasSecretEngineConfig CR is deleted
   **When** the reconciler processes deletion
   **Then** the K8s object is removed but Vault config is **NOT** deleted (`IsDeletable=false` — Vault has no `DELETE /mongodbatlas/config` endpoint)

4. **Given** the MongoDBAtlasSecretEngineRole CR is deleted
   **When** the reconciler processes deletion
   **Then** the role is removed from Vault via `DELETE /roles/{name}` and the CR is deleted from K8s

5. **Given** the MongoDBAtlasSecretEngineRole CR spec is updated (e.g., `roles` changed)
   **When** the reconciler processes the update
   **Then** the Vault role reflects the updated value

6. **Given** any MongoDB Atlas CR is created or updated
   **When** the webhook validates it
   **Then** `spec.path` and `spec.name` immutability is enforced on updates, and credential source validation passes for config

## Tasks / Subtasks

- [x] Task 1: Create `MongoDBAtlasSecretEngineConfig` type (AC: 1, 3, 6)
  - [x] 1.1: Create `api/v1alpha1/mongodbatlassecretengineconfig_types.go` — Spec with `Connection`, `Authentication`, `Path`, inline `MongoDBAtlasSEConfig` struct, `RootCredentials` (RootCredentialConfig)
  - [x] 1.2: Implement `VaultObject` interface — `GetPath()` returns `{path}/config`, `GetPayload()`, `IsEquivalentToDesiredState()`, `PrepareInternalValues()`, `IsDeletable()=false`
  - [x] 1.3: Implement `ConditionsAware` interface — Status with `Conditions []metav1.Condition`
  - [x] 1.4: Implement `setInternalCredentials()` — resolve public_key (UsernameKey) + private_key (PasswordKey) from K8s Secret, VaultSecret, or RandomSecret (follow AWS dual-credential pattern)
  - [x] 1.5: Implement `toMap()` on `MongoDBAtlasSEConfig` — convert to Vault API snake_case fields (`public_key`, `private_key`)
  - [x] 1.6: Implement `IsEquivalentToDesiredState()` — must delete `private_key` from desired state (Vault never returns it on read), then `removeUnsetFields` + `filterPayloadToDesiredKeys`

- [x] Task 2: Create `MongoDBAtlasSecretEngineRole` type (AC: 2, 4, 5)
  - [x] 2.1: Create `api/v1alpha1/mongodbatlassecretenginerole_types.go` — Spec with `Connection`, `Authentication`, `Path`, inline `MongoDBAtlasSERole` struct, `Name`
  - [x] 2.2: Implement `VaultObject` interface — `GetPath()` returns `{path}/roles/{name}`, `IsDeletable()=true`
  - [x] 2.3: Implement `ConditionsAware` interface
  - [x] 2.4: Implement `toMap()` on `MongoDBAtlasSERole` — include all role fields with correct snake_case mapping, use `toInterfaceArray` for list fields
  - [x] 2.5: Implement `IsEquivalentToDesiredState()` — use `removeUnsetFields` + `filterPayloadToDesiredKeys`, sort set fields (`roles`, `ip_addresses`, `cidr_blocks`, `project_roles`)

- [x] Task 3: Create webhooks (AC: 6)
  - [x] 3.1: Create `api/v1alpha1/mongodbatlassecretengineconfig_webhook.go` — `admission.Defaulter[*MongoDBAtlasSecretEngineConfig]`, `admission.Validator[*MongoDBAtlasSecretEngineConfig]`, immutable `spec.path`, credential validation via `ValidateCredentialSource()`
  - [x] 3.2: Create `api/v1alpha1/mongodbatlassecretenginerole_webhook.go` — `admission.Defaulter[*MongoDBAtlasSecretEngineRole]`, `admission.Validator[*MongoDBAtlasSecretEngineRole]`, immutable `spec.path` + `spec.name`

- [x] Task 4: Create controllers (AC: 1, 2, 3, 4, 5)
  - [x] 4.1: Create `internal/controller/mongodbatlassecretengineconfig_controller.go` — embed `ReconcilerBase`, always-write reconcile logic (private_key is write-only, same as AWS secret_key), watches on `corev1.Secret` and `RandomSecret`
  - [x] 4.2: Create `internal/controller/mongodbatlassecretenginerole_controller.go` — embed `ReconcilerBase`, standard VaultResource reconcile flow

- [x] Task 5: Register in main.go (AC: 1, 2)
  - [x] 5.1: Add controller registrations for `MongoDBAtlasSecretEngineConfigReconciler` and `MongoDBAtlasSecretEngineRoleReconciler`
  - [x] 5.2: Add webhook registrations inside `ENABLE_WEBHOOKS` guard for both types

- [x] Task 6: Unit tests (AC: 1, 2, 5, 6)
  - [x] 6.1: Create `api/v1alpha1/mongodbatlassecretengineconfig_test.go` — test `toMap()`, test `IsEquivalentToDesiredState()` with independently constructed Vault-read-shaped fixtures (private_key stripping), negative test proving managed field mismatch returns `false`
  - [x] 6.2: Create `api/v1alpha1/mongodbatlassecretenginerole_test.go` — test `toMap()` for roles, ip_addresses, cidr_blocks, project_roles variants; test `IsEquivalentToDesiredState()`; negative tests

- [x] Task 7: Test fixtures and integration tests (AC: 1, 2, 4, 5)
  - [x] 7.1: Create test YAML fixtures in `test/mongodbatlassecretengine/` — config and role CRs
  - [x] 7.2: Integration tests — **SKIP** (see Dev Notes: MongoDB Atlas is a cloud service that cannot be installed in Kind; falls under "Skip it")

- [x] Task 8: Documentation (AC: all)
  - [x] 8.1: Create `docs/secret-engines/mongodb-atlas.md` following `docs/engine-doc-template.md` (DNFR5)
  - [x] 8.2: Update `docs/secret-engines/index.md` with link to new doc file

- [x] Task 9: Code generation and validation (AC: all)
  - [x] 9.1: Run `make manifests generate fmt vet test`
  - [x] 9.2: Verify all existing tests still pass
  - [x] 9.3: Add new CRDs to `config/crd/kustomization.yaml` under `resources` list

## Dev Notes

### Story Intelligence Chain — Predecessor Context

**Epic 12 (completed)** established the refined pattern for adding new secret engine CRDs (Consul 12.1, GCP 12.2, LDAP/AD 12.3). Story 13.3 follows the same pattern.

**Epics 11-12 key learnings:**
- **Always-write pattern for config types with write-only credentials**: AWS config controller uses a `manageReconcileLogic` that always writes (skipping `IsEquivalentToDesiredState`) because the secret_key is write-only — Vault never returns it on read, so drift detection cannot observe credential rotations. The MongoDB Atlas `config` endpoint has the same characteristic (private_key is write-only). Follow the AWS config controller pattern exactly.
- **Dual-credential resolution (AWS pattern)**: The MongoDB Atlas config requires TWO credentials — `public_key` (username-like) + `private_key` (password-like). This is the same pattern as AWS (`access_key` + `secret_key`). Follow `AWSSecretEngineConfig.setInternalCredentials()` exactly, using `UsernameKey` → public_key, `PasswordKey` → private_key. The `publicKey` field in the spec allows setting the public key directly (like AWS's `accessKey`), falling back to credential resolution.
- **`removeUnsetFields` for clean drift detection**: AWS role uses `removeUnsetFields(desiredState, payload)` before `filterPayloadToDesiredKeys` to prevent false drift from zero-value fields that Vault omits. Apply the same pattern for MongoDBAtlasSecretEngineRole.
- **`spec.name` immutability**: Every type with a Name override field must include `spec.name` in `ValidateUpdate` (Epic 11 action item).
- **CRD registration checklist**: After `make manifests`, add new CRD YAML files to `config/crd/kustomization.yaml`. Missing this breaks Helm chart builds.
- **Shared framework protection**: Do NOT modify behavior in shared framework files (`reconcile_skeleton.go`, `vaultresourcereconciler.go`, etc.). Only add new exported functions/types.
- **`toMap()` normalization rule**: Emit Vault-read format for correct `IsEquivalentToDesiredState` comparison (Epic 12 action item).
- **`sortAnyStringSlice` already exists** in `awssecretenginerole_types.go` — do NOT redefine.
- **`toInterfaceArray` already exists** — used throughout for `[]string` → `[]any` conversion.

### Integration Test Classification: SKIP

Per the project's Integration Test Infrastructure Philosophy:
> **Skip it** — For cloud providers and services that cannot be installed in Kind and are hard to mock (e.g., AWS, Azure, GCP, GitHub App, Quay), the integration test must not be run.

MongoDB Atlas is a fully managed cloud database service. It cannot be installed in Kind and requires a real Atlas account with Programmatic API Key credentials. This falls squarely in the SKIP category. Comprehensive unit tests for `toMap()`, `IsEquivalentToDesiredState()`, and webhook validation are the primary quality gate.

### Vault API Reference

**Config endpoint:** `POST {path}/config`

| Field | Type | Description |
|-------|------|-------------|
| `public_key` | string | The Public Programmatic API Key used to authenticate with the MongoDB Atlas API |
| `private_key` | string | The Private Programmatic API Key used to connect with MongoDB Atlas API |

**No GET endpoint is explicitly documented** for `/mongodbatlas/config`. If a GET exists, `private_key` is write-only (never returned).
**No DELETE endpoint exists** for `/mongodbatlas/config` → `MongoDBAtlasSecretEngineConfig.IsDeletable()` must return `false`.

**Roles endpoint:** `POST/GET/DELETE {path}/roles/{name}`

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `organization_id` | string | "" | Unique org identifier. Required if `project_id` is not set |
| `project_id` | string | "" | Unique project identifier. Required if `organization_id` is not set |
| `roles` | list [string] | required | Atlas roles for the API Key (e.g., `ORG_READ_ONLY`, `GROUP_CLUSTER_MANAGER`) |
| `ip_addresses` | list [string] | [] | IP whitelist entries. Mutually exclusive with `cidr_blocks` |
| `cidr_blocks` | list [string] | [] | CIDR whitelist entries. Mutually exclusive with `ip_addresses` |
| `project_roles` | list [string] | [] | Roles assigned when an org API key is assigned to a project |
| `ttl` | string | "0" | TTL for generated credentials (duration format) |
| `max_ttl` | string | "0" | Max TTL for generated credentials (duration format) |

**Valid organization roles:** `ORG_OWNER`, `ORG_MEMBER`, `ORG_GROUP_CREATOR`, `ORG_BILLING_ADMIN`, `ORG_READ_ONLY`

**Valid project roles:** `GROUP_CHARTS_ADMIN`, `GROUP_CLUSTER_MANAGER`, `GROUP_DATA_ACCESS_ADMIN`, `GROUP_DATA_ACCESS_READ_ONLY`, `GROUP_DATA_ACCESS_READ_WRITE`, `GROUP_OWNER`, `GROUP_READ_ONLY`

**GET response for roles:**
```json
{
  "project_id": "5cf5a45a9ccf6400e60981b6",
  "roles": ["GROUP_CLUSTER_MANAGER"],
  "cidr_blocks": ["192.168.1.3/32"],
  "ip_addresses": ["192.168.1.3", "192.168.1.4"],
  "organization_id": "7cf5a45a9ccf6400e60981b7",
  "ttl": "30m",
  "max_ttl": "1h"
}
```

### Critical: `IsEquivalentToDesiredState` for Config — Private Key Stripping

Vault never returns `private_key` on read from `config`. The `IsEquivalentToDesiredState` implementation must:
1. Build `desiredState` from `toMap()`
2. `delete(desiredState, "private_key")` — remove before comparison
3. Use `removeUnsetFields(desiredState, payload)` then `filterPayloadToDesiredKeys(desiredState, payload)` then `reflect.DeepEqual`

Follow the pattern from `AWSSecretEngineConfig` (deletes `secret_key`), `ConsulSecretEngineConfig` (deletes `token`, `ca_cert`, `client_cert`, `client_key`).

### Critical: Config Controller Uses Always-Write Pattern

Since `private_key` is write-only (never returned on read), drift detection cannot observe credential rotations. The config controller must use the **always-write pattern** (same as `AWSSecretEngineConfigReconciler.manageReconcileLogic`):
- Call `PrepareInternalValues` to resolve the credentials
- Always call `vaultEndpoint.Create(ctx)` (skip the standard `CreateOrUpdate` which reads first)
- This ensures credential rotations propagate to Vault

### MongoDBAtlasSecretEngineConfig — Dual-Credential Resolution via RootCredentialConfig

The `public_key` and `private_key` must be resolved from one of three sources via `RootCredentialConfig`:
- **K8s Secret**: `UsernameKey` (default: `username`) → public_key, `PasswordKey` (default: `password`) → private_key
- **VaultSecret**: same key mapping from a Vault KV path
- **RandomSecret**: retrieve private_key from RandomSecret's Vault path; `spec.publicKey` must be set (randomSecret only provides the private key)

Pattern: follow `AWSSecretEngineConfig.setInternalCredentials()` exactly. Store resolved values in unexported fields (`retrievedPublicKey` and `retrievedPrivateKey` with `json:"-"`) and include them in `toMap()` output.

**Key similarity to AWS**: Dual credential — `publicKey` acts like AWS's `accessKey` (can be set in spec directly), `private_key` acts like AWS's `secret_key` (always resolved from credentials).

**RandomSecret validation**: When using `RandomSecret`, `spec.publicKey` must be set in the spec since RandomSecret only provides the private key (same rule as AWS: `spec.accessKey` must be set when using randomSecret).

### MongoDBAtlasSecretEngineConfig — `GetPath()` Is Fixed

Like Consul, the MongoDB Atlas config endpoint is always `{path}/config` (no per-name suffix). `GetPath()` must return `CleansePath(string(d.Spec.Path) + "/config")`.

The `Name` field is NOT needed for MongoDBAtlasSecretEngineConfig since the path is fixed. Omit it from the spec to keep the type clean (same decision as Consul config).

### CRD Field Spec — MongoDBAtlasSecretEngineConfig

```go
type MongoDBAtlasSEConfig struct {
    // PublicKey is the MongoDB Atlas Public Programmatic API Key.
    // Can be set directly here or resolved from credentials via RootCredentials.
    // When using RandomSecret, this field must be set (RandomSecret only provides the private key).
    // +kubebuilder:validation:Optional
    PublicKey string `json:"publicKey,omitempty"`

    retrievedPublicKey  string `json:"-"`
    retrievedPrivateKey string `json:"-"`
}
```

### MongoDBAtlasSecretEngineConfig Spec Design

```go
type MongoDBAtlasSecretEngineConfigSpec struct {
    // Connection represents the information needed to connect to Vault.
    // +kubebuilder:validation:Optional
    Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

    // Authentication is the kube auth configuration to be used to execute this request
    // +kubebuilder:validation:Required
    Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

    // Path at which to make the configuration.
    // The final path in Vault will be {[spec.authentication.namespace]}/{spec.path}/config.
    // +kubebuilder:validation:Required
    Path vaultutils.Path `json:"path,omitempty"`

    MongoDBAtlasSEConfig `json:",inline"`

    // RootCredentials specifies how to retrieve the MongoDB Atlas Programmatic API Key pair.
    // UsernameKey maps to public_key, PasswordKey maps to private_key.
    // +kubebuilder:validation:Required
    RootCredentials vaultutils.RootCredentialConfig `json:"rootCredentials,omitempty"`
}
```

### CRD Field Spec — MongoDBAtlasSecretEngineRole

```go
type MongoDBAtlasSERole struct {
    // OrganizationID is the unique identifier for the Atlas organization.
    // Required if projectID is not set.
    // +kubebuilder:validation:Optional
    OrganizationID string `json:"organizationID,omitempty"`

    // ProjectID is the unique identifier for the Atlas project.
    // Required if organizationID is not set.
    // +kubebuilder:validation:Optional
    ProjectID string `json:"projectID,omitempty"`

    // Roles is the list of Atlas roles that the API Key needs to have.
    // At least one role is required. Org roles: ORG_OWNER, ORG_MEMBER, ORG_GROUP_CREATOR,
    // ORG_BILLING_ADMIN, ORG_READ_ONLY.
    // Project roles: GROUP_CHARTS_ADMIN, GROUP_CLUSTER_MANAGER, GROUP_DATA_ACCESS_ADMIN,
    // GROUP_DATA_ACCESS_READ_ONLY, GROUP_DATA_ACCESS_READ_WRITE, GROUP_OWNER, GROUP_READ_ONLY.
    // +kubebuilder:validation:Required
    // +kubebuilder:validation:MinItems=1
    // +listType=set
    Roles []string `json:"roles"`

    // IPAddresses is the list of IP addresses to add to the API key whitelist.
    // Mutually exclusive with cidrBlocks.
    // +kubebuilder:validation:Optional
    // +listType=set
    IPAddresses []string `json:"ipAddresses,omitempty"`

    // CIDRBlocks is the list of CIDR notation entries to add to the API key whitelist.
    // Mutually exclusive with ipAddresses.
    // +kubebuilder:validation:Optional
    // +listType=set
    CIDRBlocks []string `json:"cidrBlocks,omitempty"`

    // ProjectRoles is the list of roles assigned when an Organization API key is
    // assigned to a Project API key.
    // +kubebuilder:validation:Optional
    // +listType=set
    ProjectRoles []string `json:"projectRoles,omitempty"`

    // TTL specifies the duration after which the issued credential should expire.
    // Uses duration format strings (e.g., "2h", "30m").
    // +kubebuilder:validation:Optional
    TTL string `json:"ttl,omitempty"`

    // MaxTTL specifies the maximum allowed lifetime of credentials issued using this role.
    // +kubebuilder:validation:Optional
    MaxTTL string `json:"maxTTL,omitempty"`
}
```

### MongoDBAtlasSecretEngineRole Spec Design

```go
type MongoDBAtlasSecretEngineRoleSpec struct {
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

    MongoDBAtlasSERole `json:",inline"`

    // The name of the object created in Vault. If this is specified it takes precedence over {metadata.name}
    // +kubebuilder:validation:Optional
    // +kubebuilder:validation:Pattern:=`[a-z0-9]([-a-z0-9]*[a-z0-9])?`
    Name string `json:"name,omitempty"`
}
```

### toMap() for MongoDBAtlasSEConfig

```go
func (i *MongoDBAtlasSEConfig) toMap() map[string]any {
    payload := map[string]any{}
    if i.retrievedPublicKey != "" {
        payload["public_key"] = i.retrievedPublicKey
    } else if i.PublicKey != "" {
        payload["public_key"] = i.PublicKey
    }
    if i.retrievedPrivateKey != "" {
        payload["private_key"] = i.retrievedPrivateKey
    }
    return payload
}
```

### toMap() for MongoDBAtlasSERole

```go
func (i *MongoDBAtlasSERole) toMap() map[string]any {
    payload := map[string]any{}
    payload["organization_id"] = i.OrganizationID
    payload["project_id"] = i.ProjectID
    payload["roles"] = toInterfaceArray(i.Roles)
    payload["ip_addresses"] = toInterfaceArray(i.IPAddresses)
    payload["cidr_blocks"] = toInterfaceArray(i.CIDRBlocks)
    payload["project_roles"] = toInterfaceArray(i.ProjectRoles)
    payload["ttl"] = i.TTL
    payload["max_ttl"] = i.MaxTTL
    return payload
}
```

### IsEquivalentToDesiredState for MongoDBAtlasSecretEngineConfig

```go
func (d *MongoDBAtlasSecretEngineConfig) IsEquivalentToDesiredState(payload map[string]any) bool {
    desiredState := d.Spec.MongoDBAtlasSEConfig.toMap()
    delete(desiredState, "private_key")
    removeUnsetFields(desiredState, payload)
    return reflect.DeepEqual(desiredState, filterPayloadToDesiredKeys(desiredState, payload))
}
```

### IsEquivalentToDesiredState for MongoDBAtlasSecretEngineRole

```go
func (d *MongoDBAtlasSecretEngineRole) IsEquivalentToDesiredState(payload map[string]any) bool {
    desiredState := d.Spec.MongoDBAtlasSERole.toMap()
    removeUnsetFields(desiredState, payload)
    filteredPayload := filterPayloadToDesiredKeys(desiredState, payload)
    setFields := []string{"roles", "ip_addresses", "cidr_blocks", "project_roles"}
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
api/v1alpha1/mongodbatlassecretengineconfig_types.go      (NEW)
api/v1alpha1/mongodbatlassecretengineconfig_webhook.go     (NEW)
api/v1alpha1/mongodbatlassecretengineconfig_test.go        (NEW)
api/v1alpha1/mongodbatlassecretenginerole_types.go         (NEW)
api/v1alpha1/mongodbatlassecretenginerole_webhook.go       (NEW)
api/v1alpha1/mongodbatlassecretenginerole_test.go          (NEW)
internal/controller/mongodbatlassecretengineconfig_controller.go       (NEW)
internal/controller/mongodbatlassecretenginerole_controller.go         (NEW)
test/mongodbatlassecretengine/mongodb-atlas-secret-engine-config.yaml  (NEW)
test/mongodbatlassecretengine/mongodb-atlas-secret-engine-role.yaml    (NEW)
docs/secret-engines/mongodb-atlas.md                                   (NEW)
```

### Files to Update

```
cmd/main.go                                       (UPDATE - register controllers + webhooks)
config/crd/kustomization.yaml                     (UPDATE - add new CRD resources)
docs/secret-engines/index.md                      (UPDATE - add link to mongodb-atlas.md)
api/v1alpha1/zz_generated.deepcopy.go             (UPDATE - auto-generated by make generate)
config/crd/bases/                                 (UPDATE - auto-generated CRD YAMLs)
config/rbac/role.yaml                             (UPDATE - auto-generated RBAC)
config/webhook/manifests.yaml                     (UPDATE - auto-generated webhook config)
```

### Files Being Modified — Current State

**`cmd/main.go`**: Currently registers ~40+ controllers and ~40+ webhooks. New registrations follow the exact same pattern:
- Controller: `(&controller.MongoDBAtlasSecretEngineConfigReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "MongoDBAtlasSecretEngineConfig")}).SetupWithManager(mgr)`
- Webhook (inside ENABLE_WEBHOOKS): `(&redhatcopv1alpha1.MongoDBAtlasSecretEngineConfig{}).SetupWebhookWithManager(mgr)`

No existing behavior is changed — this is purely additive.

### Existing Patterns to Follow (Reference Files)

| Pattern | Reference File |
|---------|---------------|
| Config type with dual-credential RootCredentialConfig | `api/v1alpha1/awssecretengineconfig_types.go` |
| Config dual-credential resolution (public_key + private_key) | `AWSSecretEngineConfig.setInternalCredentials()` (access_key + secret_key) |
| Secret-stripping in IsEquivalentToDesiredState | `api/v1alpha1/awssecretengineconfig_types.go` (deletes `secret_key`) |
| Config always-write controller | `internal/controller/awssecretengineconfig_controller.go` (manageReconcileLogic) |
| Config with fixed path (no Name field) | `api/v1alpha1/consulsecretengineconfig_types.go` (config/access) |
| Role type (simple, no credentials) | `api/v1alpha1/consulsecretenginerole_types.go` |
| Role with set field sorting | `AWSSecretEngineRole.IsEquivalentToDesiredState()` (sortAnyStringSlice) |
| Webhook pattern | `api/v1alpha1/awssecretengineconfig_webhook.go` |
| Controller with credential watches | `internal/controller/awssecretengineconfig_controller.go` |
| Controller (simple role) | `internal/controller/consulsecretenginerole_controller.go` |
| filterPayloadToDesiredKeys helper | `api/v1alpha1/payload_filter.go` |
| removeUnsetFields helper | `api/v1alpha1/payload_filter.go` |
| sortAnyStringSlice helper | `api/v1alpha1/awssecretenginerole_types.go` (reuse, do NOT redefine) |

### RBAC Markers

Config controller:
```go
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=mongodbatlassecretengineconfigs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=mongodbatlassecretengineconfigs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=mongodbatlassecretengineconfigs/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=serviceaccounts/token,verbs=create
//+kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;patch
```

Role controller:
```go
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=mongodbatlassecretengineroles,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=mongodbatlassecretengineroles/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=mongodbatlassecretengineroles/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=serviceaccounts/token,verbs=create
//+kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;patch
```

### Kubebuilder Markers Checklist

- All exported Spec fields: `+kubebuilder:validation:Required` or `+kubebuilder:validation:Optional`
- `Roles` (on role type): `+kubebuilder:validation:Required`, `+kubebuilder:validation:MinItems=1`, `+listType=set`, no `omitempty` (required field)
- List fields (ipAddresses, cidrBlocks, projectRoles): `+listType=set`
- Root type: `+kubebuilder:object:root=true`, `+kubebuilder:subresource:status`
- Conditions field: `+patchMergeKey=type`, `+patchStrategy=merge`, `+listType=map`, `+listMapKey=type`
- Webhook paths: `/mutate-redhatcop-redhat-io-v1alpha1-mongodbatlassecretengineconfig`, `/validate-redhatcop-redhat-io-v1alpha1-mongodbatlassecretengineconfig` (all lowercase)

### Unit Test Requirements

**Config tests (`mongodbatlassecretengineconfig_test.go`):**
1. `TestMongoDBAtlasSecretEngineConfig_toMap` — verify snake_case keys (`public_key`, `private_key`), verify resolved credentials appear
2. `TestMongoDBAtlasSecretEngineConfig_toMap_PublicKeyFromSpec` — verify spec.publicKey is used when no retrieved value
3. `TestMongoDBAtlasSecretEngineConfig_IsEquivalentToDesiredState_Match` — construct Vault-read-shaped payload (only `public_key`, no `private_key`), verify returns `true`
4. `TestMongoDBAtlasSecretEngineConfig_IsEquivalentToDesiredState_Mismatch` — change `public_key`, verify returns `false`
5. `TestMongoDBAtlasSecretEngineConfig_IsEquivalentToDesiredState_ExtraVaultFields` — add extra Vault fields, verify still returns `true`

**Role tests (`mongodbatlassecretenginerole_test.go`):**
1. `TestMongoDBAtlasSecretEngineRole_toMap_OrgLevel` — verify organization_id, roles, list fields use `toInterfaceArray`, verify snake_case keys
2. `TestMongoDBAtlasSecretEngineRole_toMap_ProjectLevel` — verify project_id, roles, cidr_blocks format
3. `TestMongoDBAtlasSecretEngineRole_IsEquivalentToDesiredState_Match` — construct Vault-read fixture, verify `true`
4. `TestMongoDBAtlasSecretEngineRole_IsEquivalentToDesiredState_Mismatch` — change a field, verify `false`
5. `TestMongoDBAtlasSecretEngineRole_IsEquivalentToDesiredState_UnsetFields` — verify `removeUnsetFields` prevents false drift from zero-value fields Vault omits
6. `TestMongoDBAtlasSecretEngineRole_IsEquivalentToDesiredState_UnsortedRoles` — verify `sortAnyStringSlice` handles unsorted set fields correctly

### Critical Gotchas

1. **`private_key` is never returned by Vault on read** — `IsEquivalentToDesiredState` MUST delete `private_key` from desiredState.

2. **Always-write controller pattern** — Since the private_key is write-only, the config controller must use the always-write pattern (direct `vaultEndpoint.Create` without read-compare). Copy the `manageReconcileLogic` method from `AWSSecretEngineConfigReconciler`.

3. **No `Name` field on Config** — The config path is fixed (`config`). Do NOT add a `Name` field to `MongoDBAtlasSecretEngineConfigSpec`. The `spec.name` immutability check in `ValidateUpdate` is NOT needed for config since there's no Name field — only check `spec.path`.

4. **`sortAnyStringSlice` already exists** — It's defined in `awssecretenginerole_types.go` in the same package. Do NOT redefine it. Just call it directly.

5. **`toInterfaceArray` already exists** — Used throughout the codebase for `[]string` to `[]any` conversion. Use it for all list fields in the role's `toMap()`.

6. **CRD registration in kustomization.yaml** — After `make manifests`, add `redhatcop.redhat.io_mongodbatlassecretengineconfigs.yaml` and `redhatcop.redhat.io_mongodbatlassecretengineroles.yaml` to `config/crd/kustomization.yaml` resources list.

7. **Webhook registration uses lowercase type name** — Paths: `/mutate-redhatcop-redhat-io-v1alpha1-mongodbatlassecretengineconfig` and `/validate-redhatcop-redhat-io-v1alpha1-mongodbatlassecretengineconfig` (no hyphens in type name).

8. **No json.Number needed** — MongoDB Atlas config has no integer fields (all strings). Role TTL/max_ttl are strings. No numeric conversion needed.

9. **Dual-credential validation** — When using `RandomSecret`, `spec.publicKey` must be set in the spec (same rule as AWS: `spec.accessKey` must be set when using randomSecret). Add this validation in `isValid()`.

10. **`ip_addresses` and `cidr_blocks` mutual exclusivity** — The Vault API states these are mutually exclusive. Do NOT add webhook validation for this — Vault itself enforces this rule and returns a clear error. Adding duplicate validation creates coupling to Vault's evolving rules (same precedent as Consul where we didn't add role-requirement validation).

### Anti-Patterns / DO NOT

- **DO NOT** add a `Name` field to MongoDBAtlasSecretEngineConfig — the path is fixed at `config`
- **DO NOT** redefine `sortAnyStringSlice` or `toInterfaceArray` — they exist in the same package
- **DO NOT** modify shared framework files (`reconcile_skeleton.go`, `vaultresourcereconciler.go`, etc.)
- **DO NOT** write integration tests — MongoDB Atlas is classified as SKIP (cloud service)
- **DO NOT** use `CreateOrUpdate` for the config controller — use the always-write pattern (private_key is write-only)
- **DO NOT** emit `json.Number` for any fields — MongoDB Atlas has no integer fields that need it (TTL/max_ttl are strings)
- **DO NOT** add webhook validation for ip_addresses/cidr_blocks mutual exclusivity — Vault enforces this
- **DO NOT** add webhook validation requiring at least one of organization_id/project_id — Vault enforces this
- **DO NOT** add webhook validation for role enum values — Vault enforces valid role names

### Project Structure Notes

- All files follow existing naming conventions: lowercase type name for files
- New files go in established directories (`api/v1alpha1/`, `internal/controller/`, `test/`, `docs/secret-engines/`)
- Test fixture directory `test/mongodbatlassecretengine/` follows the existing pattern (`test/awssecretengine/`, `test/consulsecretengine/`)
- Controllers in `internal/controller/` (go/v4 layout since Epic 10)
- No conflicts with existing code — this is entirely additive

### References

- [Source: _bmad-output/planning-artifacts/epics.md#Epic-13, Story 13.3 — requirements and acceptance criteria]
- [Source: _bmad-output/project-context.md — all project rules, patterns, and conventions]
- [Source: api/v1alpha1/awssecretengineconfig_types.go — config type with dual-credential RootCredentialConfig and always-write pattern (primary reference)]
- [Source: api/v1alpha1/awssecretenginerole_types.go — role type with sortAnyStringSlice, removeUnsetFields]
- [Source: api/v1alpha1/consulsecretengineconfig_types.go — config with fixed path (no Name), single-credential]
- [Source: api/v1alpha1/consulsecretenginerole_types.go — role type pattern]
- [Source: api/v1alpha1/awssecretengineconfig_webhook.go — webhook pattern with credential validation]
- [Source: internal/controller/awssecretengineconfig_controller.go — always-write config controller with Secret/RandomSecret watches]
- [Source: internal/controller/consulsecretenginerole_controller.go — simple role controller]
- [Source: api/v1alpha1/payload_filter.go — filterPayloadToDesiredKeys + removeUnsetFields helpers]
- [Vault API: https://developer.hashicorp.com/vault/api-docs/secret/mongodbatlas — MongoDB Atlas secrets engine API reference]
- [Vault Docs: https://developer.hashicorp.com/vault/docs/secrets/mongodbatlas — MongoDB Atlas secrets engine overview]
- [Source: _bmad-output/implementation-artifacts/12-1-consul-secret-engine-config-and-role-crds.md — predecessor story with patterns]
- [Source: _bmad-output/implementation-artifacts/11-1-aws-secret-engine-config-and-role-crds.md — AWS story with dual-credential pattern]

## Code Review Record

### Review Model Used

GPT-5.4

### Review Findings

Approved with 0 patches on first review (iteration 1). No changes requested.

### Decisions Needed / Decisions Taken

None.

### Fixes Applied

None required — approved on first review.

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6

### Debug Log References

No blocking issues encountered.

### Completion Notes List

- Implemented MongoDBAtlasSecretEngineConfig with dual-credential resolution (public_key + private_key) following the AWS config pattern exactly
- Implemented MongoDBAtlasSecretEngineRole with all Vault API fields (organization_id, project_id, roles, ip_addresses, cidr_blocks, project_roles, ttl, max_ttl)
- Config uses always-write controller pattern (private_key is write-only, never returned on read)
- Config IsEquivalentToDesiredState correctly strips private_key before comparison
- Role IsEquivalentToDesiredState sorts set fields (roles, ip_addresses, cidr_blocks, project_roles) for order-independent comparison
- Config IsDeletable=false (Vault has no DELETE /mongodbatlas/config endpoint)
- Role IsDeletable=true (Vault supports DELETE /roles/{name})
- Webhooks enforce spec.path immutability on config, spec.path + spec.name immutability on role
- Config webhook validates credential source and requires spec.publicKey when using RandomSecret
- 36 unit tests all passing covering toMap(), IsEquivalentToDesiredState(), PrepareInternalValues(), validation, and webhook logic
- Integration tests skipped per project policy (MongoDB Atlas is a cloud service that cannot be installed in Kind)
- All existing tests pass with no regressions (make test succeeds)
- CRDs generated and registered in kustomization.yaml
- Controllers and webhooks registered in main.go

### Change Log

- 2026-08-12: Implemented story 13.3 — MongoDB Atlas Secret Engine Config and Role CRDs (all 9 tasks complete)

### File List

- api/v1alpha1/mongodbatlassecretengineconfig_types.go (NEW)
- api/v1alpha1/mongodbatlassecretengineconfig_webhook.go (NEW)
- api/v1alpha1/mongodbatlassecretengineconfig_test.go (NEW)
- api/v1alpha1/mongodbatlassecretenginerole_types.go (NEW)
- api/v1alpha1/mongodbatlassecretenginerole_webhook.go (NEW)
- api/v1alpha1/mongodbatlassecretenginerole_test.go (NEW)
- api/v1alpha1/zz_generated.deepcopy.go (MODIFIED)
- internal/controller/mongodbatlassecretengineconfig_controller.go (NEW)
- internal/controller/mongodbatlassecretenginerole_controller.go (NEW)
- cmd/main.go (MODIFIED)
- config/crd/kustomization.yaml (MODIFIED)
- config/crd/bases/redhatcop.redhat.io_mongodbatlassecretengineconfigs.yaml (NEW - generated)
- config/crd/bases/redhatcop.redhat.io_mongodbatlassecretengineroles.yaml (NEW - generated)
- config/rbac/role.yaml (MODIFIED - generated)
- config/webhook/manifests.yaml (MODIFIED - generated)
- test/mongodbatlassecretengine/mongodb-atlas-secret-engine-config.yaml (NEW)
- test/mongodbatlassecretengine/mongodb-atlas-secret-engine-role.yaml (NEW)
- docs/secret-engines/mongodb-atlas.md (NEW)
- docs/secret-engines/index.md (MODIFIED)
