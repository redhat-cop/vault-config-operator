# Story 14.1: AppRole Auth Engine — Config and Role CRDs

Status: in-progress

## Story

As an operator developer,
I want a CRD for AppRoleAuthEngineRole,
So that Vault's AppRole auth method (the #1 machine-to-machine auth) role definitions can be managed declaratively.

## Acceptance Criteria

1. **Given** an AppRoleAuthEngineRole CR is created with policies and token settings **When** the reconciler processes it **Then** the role exists in Vault at `auth/{path}/role/{name}` and ReconcileSuccessful=True

2. **Given** the role spec includes secret_id configuration (bound_cidr_list, secret_id_ttl, secret_id_num_uses) **When** the reconciler processes it **Then** the secret-id constraints are configured on the role

3. **Given** an AppRoleAuthEngineRole CR spec is updated **When** the reconciler processes the update **Then** the Vault role reflects the updated values (verified via `IsEquivalentToDesiredState` comparison)

4. **Given** the CR is deleted **When** the reconciler processes deletion **Then** the role is removed from Vault via `DELETE auth/{path}/role/{name}`

5. **Given** an AppRoleAuthEngineRole CR is created or updated **When** the webhook validates it **Then** `spec.path` immutability is enforced on updates, `spec.name` immutability is enforced on updates

6. **Given** the CRD types are implemented **When** the story is marked done **Then** a documentation file exists at `docs/auth-engines/approle.md` following `docs/engine-doc-template.md` (DNFR5)

## Scope Note — No Config CRD

AppRole has **no separate config endpoint** — the mount itself (via `AuthEngineMount`) is the configuration. Roles are the only API resource. This story produces a single CRD: `AppRoleAuthEngineRole`. No `AppRoleAuthEngineConfig` CRD is created.

Secret-ID management (generate, list, destroy) is operational — the CRD manages the **role definition** only, not individual secret-IDs. No SecretID CRD is needed.

## Tasks / Subtasks

- [x] Task 1: Create `AppRoleAuthEngineRole` type (AC: 1, 2, 3, 5)
  - [x] 1.1: Create `api/v1alpha1/approleauthenginerole_types.go` — Spec with `Connection`, `Authentication`, `Path`, inline `AppRoleRole` struct, `Name`
  - [x] 1.2: Implement `VaultObject` interface — `GetPath()` returns `auth/{path}/role/{name}`, `GetPayload()`, `IsEquivalentToDesiredState()`, `IsDeletable()=true`
  - [x] 1.3: Implement `ConditionsAware` interface — Status with `Conditions []metav1.Condition`
  - [x] 1.4: Implement `toMap()` on `AppRoleRole` — emit all fields in Vault API snake_case; use `durationToSeconds()` for duration fields, `json.Number` for integer fields
  - [x] 1.5: Implement `IsEquivalentToDesiredState()` — `removeUnsetFields` + `filterPayloadToDesiredKeys` + `reflect.DeepEqual`

- [x] Task 2: Create webhook (AC: 5)
  - [x] 2.1: Create `api/v1alpha1/approleauthenginerole_webhook.go` — `admission.Defaulter[*AppRoleAuthEngineRole]`, `admission.Validator[*AppRoleAuthEngineRole]`, immutable `spec.path`/`spec.name`, `local_secret_ids` immutable on update

- [x] Task 3: Create controller (AC: 1, 2, 3, 4)
  - [x] 3.1: Create `internal/controller/approleauthenginerole_controller.go` — embed `ReconcilerBase`, standard `VaultResource` reconcile pattern, `For()` with `NewDefaultPeriodicReconcilePredicate()`

- [x] Task 4: Register in main.go (AC: 1)
  - [x] 4.1: Add controller registration for the reconciler
  - [x] 4.2: Add webhook registration inside `ENABLE_WEBHOOKS` guard

- [x] Task 5: Unit tests (AC: 1, 2, 3, 5)
  - [x] 5.1: Create `api/v1alpha1/approleauthenginerole_test.go` — test `toMap()` output, `IsEquivalentToDesiredState()` match/mismatch/extra-fields, `GetPath()`, `IsDeletable()`, `GetConditions()`/`SetConditions()`

- [x] Task 6: Integration tests (AC: 1, 2, 3, 4)
  - [x] 6.1: Create test YAML fixtures in `test/approleauthengine/` — mount CR + role CR + updated role CR
  - [x] 6.2: Create `internal/controller/approleauthenginerole_controller_test.go` with `//go:build integration` — create, verify ReconcileSuccessful, update, verify update, delete, verify Vault cleanup

- [x] Task 7: CRD registration and code generation (AC: all)
  - [x] 7.1: Run `make manifests generate fmt vet test`
  - [x] 7.2: Add new CRD YAML file to `config/crd/kustomization.yaml` (CRD registration checklist)
  - [x] 7.3: Verify all existing tests still pass

- [x] Task 8: Documentation (AC: 6)
  - [x] 8.1: Create `docs/auth-engines/approle.md` following `docs/engine-doc-template.md`
  - [x] 8.2: Update `docs/auth-engines/index.md` — add AppRole row to Supported Auth Engines table

## Dev Notes

### Integration Test Classification: INSTALL IN KIND (Full Integration Tests)

Per the project's Integration Test Infrastructure Philosophy:
> **Install in Kind** — If the service can be installed in the Kind cluster and configured to work with Vault, the test must deploy it as a real service.

AppRole is a Vault-native auth method with zero external dependencies. Only Vault itself is needed, and it's already running in the Kind cluster integration test infrastructure. Full integration tests are required.

### Story Intelligence Chain — Previous Story Context

**Epic 13 stories (13.1–13.4)** are the direct predecessors — all are secret engine CRD stories, but the patterns are identical for auth engine role types:
- **Established pattern pipeline:** `toMap()` → `removeUnsetFields` → `filterPayloadToDesiredKeys` → `reflect.DeepEqual` for `IsEquivalentToDesiredState`
- **`durationToSeconds` helper** converts duration strings to `json.Number` seconds
- **`json.Number` for all numeric fields** in `toMap()` — mandatory per project context
- **CRD registration checklist** — add to `config/crd/kustomization.yaml` after `make manifests`

**Epic 11 patterns** (AWS, Transit, SSH) — established the new-engine-type workflow for Phase 2:
- Type + webhook + controller + main.go registration + unit tests + integration tests + docs
- `spec.name` immutability rule in webhook `ValidateUpdate`

**Existing auth engine role types** are the closest structural analogs:
- **`KubernetesAuthEngineRole`** — standard auth role with `GetPath() = "auth/{path}/role/{name}"`, `IsDeletable()=true`, `VaultResource` reconcile
- **`JWTOIDCAuthEngineRole`** — auth role with many fields, uses `filterPayloadToDesiredKeys` for drift detection
- **`LDAPAuthEngineGroup`** — simple auth sub-resource with standard lifecycle

**Epic 13 retrospective action items** (relevant to this story):
- Validation matrix requirement — roles with boolean and conditional fields should include validation coverage
- Novelty risk flag — this story is LOW novelty (standard auth role pattern with no credential resolution)

### Vault API Paths

| Operation | Method | Path |
|-----------|--------|------|
| Create/update role | POST | `auth/{path}/role/{name}` |
| Read role | GET | `auth/{path}/role/{name}` |
| Delete role | DELETE | `auth/{path}/role/{name}` |
| List roles | LIST | `auth/{path}/role` |
| Read role ID | GET | `auth/{path}/role/{name}/role-id` |
| Generate secret ID | POST | `auth/{path}/role/{name}/secret-id` |

Only the first three operations are managed by this CRD. Role ID and secret ID operations are operational, not declarative.

### AppRoleAuthEngineRole — Vault API Field Reference

**Write (`POST auth/{path}/role/{name}`) fields:**

| Vault API Field | Type | Default | Description |
|-----------------|------|---------|-------------|
| `bind_secret_id` | bool | `true` | Require secret_id for login |
| `secret_id_bound_cidrs` | array | `[]` | CIDR blocks for login operations |
| `secret_id_num_uses` | integer | `0` | Times a SecretID can be used (0=unlimited) |
| `secret_id_ttl` | duration | `""` | Duration after which SecretID expires |
| `local_secret_ids` | bool | `false` | Cluster-local SecretIDs (**immutable after creation**) |
| `token_ttl` | duration | `0` | Incremental lifetime for tokens |
| `token_max_ttl` | duration | `0` | Maximum lifetime for tokens |
| `token_policies` | array | `[]` | Token policies |
| `token_bound_cidrs` | array | `[]` | IP blocks for token use |
| `token_explicit_max_ttl` | duration | `0` | Hard cap max TTL |
| `token_no_default_policy` | bool | `false` | Exclude default policy |
| `token_num_uses` | integer | `0` | Max token uses (0=unlimited) |
| `token_period` | duration | `0` | Period for periodic tokens |
| `token_type` | string | `""` | Token type: service, batch, default |

**Read (`GET auth/{path}/role/{name}`) sample response:**
```json
{
  "data": {
    "bind_secret_id": true,
    "secret_id_bound_cidrs": [],
    "secret_id_num_uses": 40,
    "secret_id_ttl": 600,
    "local_secret_ids": false,
    "token_ttl": 1200,
    "token_max_ttl": 1800,
    "token_policies": ["default"],
    "token_bound_cidrs": [],
    "token_explicit_max_ttl": 0,
    "token_no_default_policy": false,
    "token_num_uses": 0,
    "period": 0,
    "token_type": ""
  }
}
```

**Critical observations for `IsEquivalentToDesiredState`:**
- Vault returns duration fields (`token_ttl`, `token_max_ttl`, `secret_id_ttl`, `token_explicit_max_ttl`, `token_period`) as **integer seconds** (not strings) — use `durationToSeconds()` in `toMap()`
- Vault returns integer fields (`secret_id_num_uses`, `token_num_uses`) as `json.Number` — use `json.Number(strconv.Itoa())` in `toMap()`
- Vault returns `period` (not `token_period`) in read response — the field name mapping may differ between write and read. Verify actual read response field names during implementation and adjust `toMap()` or `filterPayloadToDesiredKeys` accordingly
- Vault may return extra fields not in the write payload — `filterPayloadToDesiredKeys` handles this automatically

### CRD Field Spec — AppRoleAuthEngineRole

```go
type AppRoleAuthEngineRoleSpec struct {
    // Connection is the Vault connection override
    // +kubebuilder:validation:Optional
    Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

    // Authentication is the kube auth configuration
    // +kubebuilder:validation:Required
    Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

    // Path is the path of the approle auth engine mount
    // +kubebuilder:validation:Required
    Path vaultutils.Path `json:"path,omitempty"`

    AppRoleRole `json:",inline"`

    // Name is an optional override for the Vault role name (defaults to metadata.name)
    // +kubebuilder:validation:Optional
    // +kubebuilder:validation:Pattern:="[a-z0-9]([-a-z0-9]*[a-z0-9])?"
    Name string `json:"name,omitempty"`
}

type AppRoleRole struct {
    // BindSecretID requires the secret_id to be presented during login
    // +kubebuilder:validation:Optional
    // +kubebuilder:default=true
    BindSecretID bool `json:"bindSecretID"`

    // SecretIDBoundCIDRs is the list of CIDR blocks for login operations
    // +kubebuilder:validation:Optional
    // +listType=set
    SecretIDBoundCIDRs []string `json:"secretIDBoundCIDRs,omitempty"`

    // SecretIDNumUses is the number of times a SecretID can be used (0=unlimited)
    // +kubebuilder:validation:Optional
    // +kubebuilder:validation:Minimum=0
    SecretIDNumUses int `json:"secretIDNumUses,omitempty"`

    // SecretIDTTL is the duration after which a SecretID expires (e.g. "10m", "1h")
    // +kubebuilder:validation:Optional
    SecretIDTTL string `json:"secretIDTTL,omitempty"`

    // LocalSecretIDs makes SecretIDs cluster-local. Immutable after creation.
    // +kubebuilder:validation:Optional
    LocalSecretIDs bool `json:"localSecretIDs,omitempty"`

    // TokenTTL is the incremental lifetime for generated tokens (e.g. "1h")
    // +kubebuilder:validation:Optional
    TokenTTL string `json:"tokenTTL,omitempty"`

    // TokenMaxTTL is the maximum lifetime for generated tokens (e.g. "24h")
    // +kubebuilder:validation:Optional
    TokenMaxTTL string `json:"tokenMaxTTL,omitempty"`

    // TokenPolicies is the list of token policies
    // +kubebuilder:validation:Optional
    // +listType=set
    TokenPolicies []string `json:"tokenPolicies,omitempty"`

    // TokenBoundCIDRs is the list of CIDR blocks for token authentication
    // +kubebuilder:validation:Optional
    // +listType=set
    TokenBoundCIDRs []string `json:"tokenBoundCIDRs,omitempty"`

    // TokenExplicitMaxTTL is the hard cap max TTL (e.g. "24h")
    // +kubebuilder:validation:Optional
    TokenExplicitMaxTTL string `json:"tokenExplicitMaxTTL,omitempty"`

    // TokenNoDefaultPolicy excludes the default policy from generated tokens
    // +kubebuilder:validation:Optional
    TokenNoDefaultPolicy bool `json:"tokenNoDefaultPolicy,omitempty"`

    // TokenNumUses is the max number of token uses (0=unlimited)
    // +kubebuilder:validation:Optional
    // +kubebuilder:validation:Minimum=0
    TokenNumUses int `json:"tokenNumUses,omitempty"`

    // TokenPeriod is the renewal period for periodic tokens (e.g. "24h")
    // +kubebuilder:validation:Optional
    TokenPeriod string `json:"tokenPeriod,omitempty"`

    // TokenType is the type of token to generate: service, batch, or default
    // +kubebuilder:validation:Optional
    // +kubebuilder:validation:Enum:={"service","batch","default","default-service","default-batch",""}
    TokenType string `json:"tokenType,omitempty"`
}
```

**Key field rules:**
- `BindSecretID` has `kubebuilder:default=true` (non-zero default) → **no** `omitempty` on json tag
- `LocalSecretIDs` has zero-value default (`false`) → uses `omitempty`
- `TokenNoDefaultPolicy` has zero-value default (`false`) → uses `omitempty`
- All duration fields are `string` (user provides `"1h"`, `"10m"`); `toMap()` converts via `durationToSeconds()`
- Integer fields `SecretIDNumUses`, `TokenNumUses` use Go `int` in CRD; `toMap()` emits `json.Number`
- Array fields use `+listType=set` for uniqueness
- `TokenType` uses `+kubebuilder:validation:Enum` for admission-time validation

### `toMap()` Implementation Notes

```go
func (d *AppRoleRole) toMap() map[string]any {
    payload := map[string]any{
        "bind_secret_id": d.BindSecretID,
    }
    if len(d.SecretIDBoundCIDRs) > 0 {
        payload["secret_id_bound_cidrs"] = toInterfaceArray(d.SecretIDBoundCIDRs)
    }
    if d.SecretIDNumUses != 0 {
        payload["secret_id_num_uses"] = json.Number(strconv.Itoa(d.SecretIDNumUses))
    }
    if d.SecretIDTTL != "" {
        payload["secret_id_ttl"] = durationToSeconds(d.SecretIDTTL)
    }
    if d.LocalSecretIDs {
        payload["local_secret_ids"] = d.LocalSecretIDs
    }
    if d.TokenTTL != "" {
        payload["token_ttl"] = durationToSeconds(d.TokenTTL)
    }
    if d.TokenMaxTTL != "" {
        payload["token_max_ttl"] = durationToSeconds(d.TokenMaxTTL)
    }
    if len(d.TokenPolicies) > 0 {
        payload["token_policies"] = toInterfaceArray(d.TokenPolicies)
    }
    if len(d.TokenBoundCIDRs) > 0 {
        payload["token_bound_cidrs"] = toInterfaceArray(d.TokenBoundCIDRs)
    }
    if d.TokenExplicitMaxTTL != "" {
        payload["token_explicit_max_ttl"] = durationToSeconds(d.TokenExplicitMaxTTL)
    }
    if d.TokenNoDefaultPolicy {
        payload["token_no_default_policy"] = d.TokenNoDefaultPolicy
    }
    if d.TokenNumUses != 0 {
        payload["token_num_uses"] = json.Number(strconv.Itoa(d.TokenNumUses))
    }
    if d.TokenPeriod != "" {
        payload["token_period"] = durationToSeconds(d.TokenPeriod)
    }
    if d.TokenType != "" {
        payload["token_type"] = d.TokenType
    }
    return payload
}
```

**Key:** `bind_secret_id` is always included (it has a non-zero default). All other fields are conditional — only emitted when non-zero. Duration fields use `durationToSeconds()`, integer fields use `json.Number`, array fields use `toInterfaceArray()`.

### `IsEquivalentToDesiredState` Notes

Standard pipeline — no custom logic needed:
1. Build `desiredState` from `toMap()`
2. `removeUnsetFields(desiredState, payload)` — removes zero-value keys not in Vault response
3. `filterPayloadToDesiredKeys(desiredState, payload)` — filter Vault read to managed keys only
4. `reflect.DeepEqual(desiredState, filteredPayload)`

No field stripping needed (no write-only secrets). No field remapping needed (Vault read keys match write keys for AppRole roles). Verify the `period` vs `token_period` field name mapping during implementation — if Vault returns `period` on read but accepts `token_period` on write, the `toMap()` key must match the read response key.

### `GetPath()` Implementation

```go
func (d *AppRoleAuthEngineRole) GetPath() string {
    name := d.Name
    if d.Spec.Name != "" {
        name = d.Spec.Name
    }
    return vaultutils.CleansePath("auth/" + string(d.Spec.Path) + "/role/" + name)
}
```

### `local_secret_ids` Immutability

Per Vault docs: "This can only be set during role creation and once set, it can't be reset later." The webhook `ValidateUpdate` must reject changes to `spec.localSecretIDs`:
```go
if newObj.Spec.LocalSecretIDs != oldObj.Spec.LocalSecretIDs {
    return nil, errors.New("spec.localSecretIDs cannot be updated")
}
```

### Webhook Rules

- `Default()` — no-op (no defaulting needed beyond `kubebuilder:default=true` for `bindSecretID`)
- `ValidateCreate()` — call `isValid()` (basic validation, no complex cross-field rules)
- `ValidateUpdate()` — block `spec.path` changes, block `spec.name` changes, block `spec.localSecretIDs` changes, call `isValid()`
- `ValidateDelete()` — no-op

### Controller Pattern

Standard `VaultResource` reconcile — simplest controller pattern:
- Embed `ReconcilerBase`
- Fetch instance → `prepareContext` → `NewVaultResource` → `Reconcile`
- `SetupWithManager`: `For()` with `NewDefaultPeriodicReconcilePredicate()`, no extra watches (no credential resolution)

### RBAC Markers

```go
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=approleauthengineroles,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=approleauthengineroles/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=approleauthengineroles/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=serviceaccounts/token,verbs=create
//+kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;patch
```

### Kubebuilder Markers Checklist

- All exported Spec fields: `+kubebuilder:validation:Required` or `+kubebuilder:validation:Optional`
- `BindSecretID`: `+kubebuilder:default=true` (non-zero default → no `omitempty`)
- `TokenType`: `+kubebuilder:validation:Enum:={"service","batch","default","default-service","default-batch",""}`
- `SecretIDNumUses`/`TokenNumUses`: `+kubebuilder:validation:Minimum=0`
- Root type: `+kubebuilder:object:root=true`, `+kubebuilder:subresource:status`
- Conditions field: `+patchMergeKey=type`, `+patchStrategy=merge`, `+listType=map`, `+listMapKey=type`
- Webhook markers: mutating + validating paths for `approleauthenginerole`

### Unit Test Requirements

**Tests (`approleauthenginerole_test.go`):**

1. `TestAppRoleAuthEngineRoleGetPath` — with `spec.name` override and without (falls back to `metadata.name`)
2. `TestAppRoleRoleToMap` — verify all field keys match Vault API snake_case names, verify `bind_secret_id` always present, verify `json.Number` for `secret_id_num_uses`/`token_num_uses`, verify `durationToSeconds` output for TTL fields
3. `TestAppRoleRoleToMap_MinimalFields` — only `bind_secret_id` set, verify minimal output
4. `TestAppRoleAuthEngineRoleIsEquivalentToDesiredState_Match` — construct Vault-read-shaped payload with `json.Number` durations and integers, verify returns `true`
5. `TestAppRoleAuthEngineRoleIsEquivalentToDesiredState_Mismatch` — change `token_policies`, verify returns `false`
6. `TestAppRoleAuthEngineRoleIsEquivalentToDesiredState_ExtraVaultFields` — add extra Vault fields (e.g., `role_id`), verify still returns `true` (filtered out)
7. `TestAppRoleAuthEngineRoleIsDeletable` — returns `true`
8. `TestAppRoleAuthEngineRoleConditions` — Get/SetConditions round-trip

**Critical:** All Vault payload fixtures must use `json.Number` for numeric values, not Go `int` or `float64`.

### Integration Test Plan

**Fixture files in `test/approleauthengine/`:**
- `test-approle-auth-mount.yaml` — `AuthEngineMount` with `type: approle`, unique path for test isolation
- `test-approle-auth-role.yaml` — `AppRoleAuthEngineRole` with `bindSecretID: true`, `tokenPolicies`, `secretIDTTL`, `tokenTTL`, `tokenMaxTTL`
- `test-approle-auth-role-updated.yaml` — updated version with changed `tokenPolicies` and `tokenTTL` for update verification

**Test file (`internal/controller/approleauthenginerole_controller_test.go`):**

Test structure: `Describe("AppRoleAuthEngineRole controller", Ordered, func() {...})` with cross-Context state sharing:

1. **Context "When creating an AppRole auth mount"** — create `AuthEngineMount` of `type: approle`, wait for ReconcileSuccessful
2. **Context "When creating an AppRoleAuthEngineRole"** — create role CR, wait for ReconcileSuccessful, verify role exists in Vault via direct API read
3. **Context "When updating an AppRoleAuthEngineRole"** — update spec fields (tokenPolicies, tokenTTL), wait for re-reconcile, verify Vault state matches updated values
4. **Context "When deleting"** — delete role CR, wait for K8s NotFound, verify role is gone from Vault (404). Then delete the auth mount.

### File Structure

| File | Action | Description |
|------|--------|-------------|
| `api/v1alpha1/approleauthenginerole_types.go` | NEW | Role CRD type, VaultObject, ConditionsAware, toMap, IsEquivalentToDesiredState |
| `api/v1alpha1/approleauthenginerole_webhook.go` | NEW | Role webhook — defaulter, validator, immutable path/name/localSecretIDs |
| `api/v1alpha1/approleauthenginerole_test.go` | NEW | Unit tests for toMap, IsEquivalentToDesiredState, GetPath |
| `internal/controller/approleauthenginerole_controller.go` | NEW | Role reconciler — standard VaultResource pattern |
| `internal/controller/approleauthenginerole_controller_test.go` | NEW | Integration tests (create, update, delete lifecycle) |
| `cmd/main.go` | UPDATE | Register 1 controller + 1 webhook |
| `config/crd/kustomization.yaml` | UPDATE | Add 1 new CRD YAML file to resources list |
| `test/approleauthengine/` | NEW | Test YAML fixtures (mount, role, updated-role) |
| `docs/auth-engines/approle.md` | NEW | Engine documentation per DNFR5 |
| `docs/auth-engines/index.md` | UPDATE | Add AppRole row to Supported Auth Engines table |

### Files Being Modified — Current State

**`cmd/main.go`**: Currently registers ~45+ controllers and webhooks. New registrations follow the exact same pattern:
- Controller: `(&controller.AppRoleAuthEngineRoleReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "AppRoleAuthEngineRole")}).SetupWithManager(mgr)`
- Webhook (inside ENABLE_WEBHOOKS): `(&redhatcopv1alpha1.AppRoleAuthEngineRole{}).SetupWebhookWithManager(mgr)`

No existing behavior is changed — purely additive.

**`config/crd/kustomization.yaml`**: Add 1 new CRD YAML file to the `resources` list. Required for Helm chart build.

**`docs/auth-engines/index.md`**: Add AppRole row to the Supported Auth Engines table:
```
| AppRole | — | AppRoleAuthEngineRole | [approle.md](approle.md) |
```
Note: Config CRD column shows "—" because AppRole has no config CRD (mount-level config only).

### Existing Patterns to Follow (Reference Files)

| Pattern | Reference File |
|---------|---------------|
| Auth engine role type (closest analog) | `api/v1alpha1/kubernetesauthenginerole_types.go` |
| Auth role with many fields + filterPayloadToDesiredKeys | `api/v1alpha1/jwtoidcauthenginerole_types.go` |
| Auth role webhook (immutable path/name) | `api/v1alpha1/kubernetesauthenginerole_webhook.go` |
| Auth role controller (standard VaultResource) | `internal/controller/kubernetesauthenginerole_controller.go` |
| Auth role integration test (Ordered lifecycle) | `internal/controller/kubernetesauthenginerole_controller_test.go` |
| Auth role unit test patterns | `api/v1alpha1/kubernetesauthenginerole_test.go` |
| filterPayloadToDesiredKeys + removeUnsetFields | `api/v1alpha1/payload_filter.go` |
| durationToSeconds helper | `api/v1alpha1/utils/vaultutils.go` |
| toInterfaceArray helper | `api/v1alpha1/utils/vaultutils.go` |
| Auth engine mount fixture | `test/kubernetesauthengine/test-kube-auth-mount.yaml` |
| Auth role fixture | `test/kubernetesauthengine/test-kube-auth-role.yaml` |
| Documentation template | `docs/engine-doc-template.md` |

### Anti-Patterns / DO NOT

- **DO NOT** create an `AppRoleAuthEngineConfig` CRD — AppRole has no separate config endpoint; the mount itself is the config
- **DO NOT** create a `SecretID` CRD — secret-ID management is operational (generate/list/destroy), not declarative
- **DO NOT** include deprecated `policies` field in the CRD — use `tokenPolicies` only (Vault's `policies` is deprecated in favor of `token_policies`)
- **DO NOT** modify shared framework behavior (`reconcile_skeleton.go`, `vaultresourcereconciler.go`, etc.) — only add new files
- **DO NOT** derive expected payloads from `toMap()` in unit tests — construct independent Vault-read-shaped fixtures
- **DO NOT** use Go `int` or `float64` in unit test Vault payloads — use `json.Number` to match real Vault client behavior
- **DO NOT** emit TTL/duration fields as duration strings in `toMap()` — use `durationToSeconds()` to emit `json.Number` seconds matching Vault's read format
- **DO NOT** forget to add new CRD YAML file to `config/crd/kustomization.yaml` after `make manifests`
- **DO NOT** forget to add `local_secret_ids` immutability check in `ValidateUpdate`
- **DO NOT** add namespace watches to the controller — AppRole roles have no namespace binding (unlike KubernetesAuthEngineRole)

### Project Structure Notes

- All new files follow existing naming conventions: `approleauthenginerole` lowercase for file names
- New files go in established directories (`api/v1alpha1/`, `internal/controller/`, `test/`, `docs/auth-engines/`)
- Test fixture directory `test/approleauthengine/` follows existing pattern (`test/kubernetesauthengine/`)
- No conflicts with existing code — purely additive

### References

- [Source: _bmad-output/planning-artifacts/epics.md#Epic-14, Story 14.1 — requirements and acceptance criteria]
- [Source: _bmad-output/project-context.md — all project rules, patterns, and conventions]
- [Source: api/v1alpha1/kubernetesauthenginerole_types.go — closest auth role type analog]
- [Source: api/v1alpha1/jwtoidcauthenginerole_types.go — auth role with many fields]
- [Source: api/v1alpha1/kubernetesauthenginerole_webhook.go — auth role webhook pattern]
- [Source: internal/controller/kubernetesauthenginerole_controller.go — standard VaultResource controller]
- [Source: api/v1alpha1/payload_filter.go — filterPayloadToDesiredKeys and removeUnsetFields]
- [Source: Vault AppRole API — https://developer.hashicorp.com/vault/api-docs/auth/approle]
- [Source: _bmad-output/implementation-artifacts/13-4-terraform-cloud-secret-engine-config-and-role-crds.md — latest predecessor story]

## Code Review Record

### Review Model Used

GPT-5.4 (via Cursor)

### Review Findings

- [ ] [Review][Patch] AppRole updates still cannot clear previously-set optional fields back to default or empty values, so stale Vault role settings remain after spec updates [api/v1alpha1/approleauthenginerole_types.go:181]
- [ ] [Review][Patch] AppRole regression coverage still misses the clear-to-default update path, so the stale-state bug passes both unit and integration tests unnoticed [api/v1alpha1/approleauthenginerole_test.go:116]

### Decisions Needed / Decisions Taken

- No SecretID CRD — scope to role definitions only (decided pre-story)
- No AppRoleAuthEngineConfig CRD — AppRole has no config endpoint (decided pre-story)
- `local_secret_ids` immutability enforced in webhook ValidateUpdate (Vault enforces this too, but admission-time rejection is better UX)

### Fixes Applied

None during review; findings recorded for follow-up.

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6 (via Cursor)

### Debug Log References

No debug issues encountered.

### Completion Notes List

- Implemented AppRoleAuthEngineRole CRD type with full VaultObject/ConditionsAware interface support
- `toMap()` emits all fields in Vault API snake_case with proper `durationToSeconds()` for duration fields and `json.Number` for integer fields
- `IsEquivalentToDesiredState()` uses standard pipeline: `removeUnsetFields` + `filterPayloadToDesiredKeys` + `reflect.DeepEqual`
- Webhook enforces immutability on `spec.path`, `spec.name`, and `spec.localSecretIDs`
- Controller follows simplest VaultResource pattern (no namespace watches, no credential resolution)
- Unit tests cover: `GetPath()` (with/without name override), `toMap()` (full/minimal), `IsEquivalentToDesiredState()` (match/mismatch/extra-fields), `IsDeletable()`, `Get/SetConditions()`
- Integration tests cover full lifecycle: create mount → create role → update role → delete role → delete mount
- All existing tests pass with no regressions

### Change Log

- 2026-08-15: Implemented AppRoleAuthEngineRole CRD — types, webhook, controller, unit tests, integration tests, documentation

### File List

- api/v1alpha1/approleauthenginerole_types.go (NEW)
- api/v1alpha1/approleauthenginerole_webhook.go (NEW)
- api/v1alpha1/approleauthenginerole_test.go (NEW)
- internal/controller/approleauthenginerole_controller.go (NEW)
- internal/controller/approleauthenginerole_controller_test.go (NEW)
- cmd/main.go (MODIFIED)
- config/crd/kustomization.yaml (MODIFIED)
- config/crd/bases/redhatcop.redhat.io_approleauthengineroles.yaml (GENERATED)
- config/rbac/role.yaml (GENERATED)
- config/webhook/manifests.yaml (GENERATED)
- api/v1alpha1/zz_generated.deepcopy.go (GENERATED)
- test/approleauthengine/test-approle-auth-mount.yaml (NEW)
- test/approleauthengine/test-approle-auth-role.yaml (NEW)
- test/approleauthengine/test-approle-auth-role-updated.yaml (NEW)
- docs/auth-engines/approle.md (NEW)
- docs/auth-engines/index.md (MODIFIED)
