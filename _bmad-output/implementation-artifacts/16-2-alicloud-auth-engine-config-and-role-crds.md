# Story 16.2: AliCloud Auth Engine — Config and Role CRDs

Status: ready-for-dev

## Story

As an operator developer,
I want a CRD for AliCloudAuthEngineRole,
So that Vault's AliCloud auth method can be managed declaratively.

## Scope Note — No Config CRD

The AliCloud auth method has **no config endpoint**. The Vault plugin (`vault-plugin-auth-alicloud`) registers only `login`, `role/:role`, and `roles` paths — no `config` path exists in `backend.go`. The official Vault API docs (v1.21.x and v2.x) confirm this: only role CRUD, list, and login are documented.

This is identical to AppRole (Story 14.1), which also has no config endpoint. The mount itself (via `AuthEngineMount`) is the configuration. This story produces a **single CRD**: `AliCloudAuthEngineRole`. No `AliCloudAuthEngineConfig` CRD is created.

**Evidence:**
- Plugin source `backend.go` `Paths` array: `pathLogin(b)`, `pathListRole(b)`, `pathListRoles(b)`, `pathRole(b)` — no `pathConfig`
- Official API docs: https://developer.hashicorp.com/vault/api-docs/auth/alicloud — no config section
- AliCloud auth works via signed `GetCallerIdentity` requests validated against AliCloud STS — no operator-configured credentials needed

## Acceptance Criteria

1. **Given** an AliCloudAuthEngineRole CR is created with an ARN and token policies **When** the reconciler processes it **Then** the role exists in Vault at `auth/{path}/role/{name}` and ReconcileSuccessful=True

2. **Given** an AliCloudAuthEngineRole CR spec is updated **When** the reconciler processes the update **Then** the Vault role reflects the updated values (verified via `IsEquivalentToDesiredState` comparison)

3. **Given** the CR is deleted **When** the reconciler processes deletion **Then** the role is removed from Vault via `DELETE auth/{path}/role/{name}`

4. **Given** an AliCloudAuthEngineRole CR is created or updated **When** the webhook validates it **Then** `spec.path` immutability is enforced on updates, `spec.name` immutability is enforced on updates

5. **Given** the CRD types are implemented **When** the story is marked done **Then** a documentation file exists at `docs/auth-engines/alicloud.md` following `docs/engine-doc-template.md` (DNFR5)

## Tasks / Subtasks

- [ ] Task 1: Create `AliCloudAuthEngineRole` type (AC: 1, 2, 4)
  - [ ] 1.1: Create `api/v1alpha1/alicloudauthenginerole_types.go` — Spec with `Connection`, `Authentication`, `Path`, inline `AliCloudAuthRole` struct, `Name`
  - [ ] 1.2: Implement `VaultObject` interface — `GetPath()` returns `auth/{path}/role/{name}`, `GetPayload()`, `IsEquivalentToDesiredState()`, `IsDeletable()=true`
  - [ ] 1.3: Implement `ConditionsAware` interface — Status with `Conditions []metav1.Condition`
  - [ ] 1.4: Implement `toMap()` on `AliCloudAuthRole` — emit `arn` plus standard token fields; use `durationToSeconds()` for duration fields, `json.Number` for integer fields
  - [ ] 1.5: Implement `IsEquivalentToDesiredState()` — `normalizeVaultReadAliases` + `removeUnsetFields` + `filterPayloadToDesiredKeys` + set sorting + `reflect.DeepEqual`

- [ ] Task 2: Create webhook (AC: 4)
  - [ ] 2.1: Create `api/v1alpha1/alicloudauthenginerole_webhook.go` — `admission.Defaulter[*AliCloudAuthEngineRole]`, `admission.Validator[*AliCloudAuthEngineRole]`, immutable `spec.path`/`spec.name`

- [ ] Task 3: Create controller (AC: 1, 2, 3)
  - [ ] 3.1: Create `internal/controller/alicloudauthenginerole_controller.go` — embed `ReconcilerBase`, standard `VaultResource` reconcile pattern, `For()` with `NewDefaultPeriodicReconcilePredicate()`

- [ ] Task 4: Register in main.go (AC: 1)
  - [ ] 4.1: Add controller registration for the reconciler
  - [ ] 4.2: Add webhook registration inside `ENABLE_WEBHOOKS` guard

- [ ] Task 5: Unit tests (AC: 1, 2, 4)
  - [ ] 5.1: Create `api/v1alpha1/alicloudauthenginerole_test.go` — test `toMap()` output, `IsEquivalentToDesiredState()` match/mismatch/extra-fields, `GetPath()`, `IsDeletable()`, `GetConditions()`/`SetConditions()`

- [ ] Task 6: Test fixtures (AC: all)
  - [ ] 6.1: Create test YAML fixtures in `test/alicloudauthengine/` — role CR
  - [ ] 6.2: Integration tests — SKIP (AliCloud is a cloud provider, no test double in Kind)

- [ ] Task 7: CRD registration and code generation (AC: all)
  - [ ] 7.1: Run `make manifests generate fmt vet test`
  - [ ] 7.2: Add 1 new CRD YAML file to `config/crd/kustomization.yaml` (CRD registration checklist)
  - [ ] 7.3: Verify all existing tests still pass

- [ ] Task 8: Documentation (AC: 5)
  - [ ] 8.1: Create `docs/auth-engines/alicloud.md` following `docs/engine-doc-template.md`
  - [ ] 8.2: Update `docs/auth-engines/index.md` — add AliCloud row to Supported Auth Engines table

## Dev Notes

### Integration Test Classification: SKIP

Per the project's Integration Test Infrastructure Philosophy:
> **Skip it** — For cloud providers and services that cannot be installed in Kind and are hard to mock (e.g., AWS, Azure, GCP, GitHub App, Quay), the integration test must not be run. These types rely on unit test coverage only.

AliCloud is a cloud provider that cannot be installed in Kind. There is no AliCloud test double available. No integration tests. Comprehensive unit tests for `toMap()`, `IsEquivalentToDesiredState()`, and webhook validation are the primary quality gate.

### Story Intelligence Chain — Previous Story Context

**Story 14.1 (AppRole Auth Engine)** is the **direct structural analog** for this story:
- AppRole also has no config endpoint → single CRD (role only)
- Established: `toMap()` with `durationToSeconds()` and `json.Number`, `removeUnsetFields` + `filterPayloadToDesiredKeys` pipeline, `normalizeVaultReadAliases` for Vault read-response key mapping
- No credential resolution, no watches, simplest controller pattern
- Webhook: immutable `spec.path` and `spec.name`, plus AppRole-specific `localSecretIDs` immutability (AliCloud has no equivalent)

**Story 16.1 (RADIUS Auth Engine)** is the predecessor story in this epic:
- If available, check for any patterns or learnings from that story

**Story 15.3 (Okta Auth Engine)** — latest completed auth engine story:
- TTL fields must use `durationToSeconds()` — corrected during Epic 15 retrospective
- `json.Number` for all Vault-facing integer fields

**Epic 15 retrospective action items (all completed):**
- `durationToSeconds()` propagated to all auth engine types
- Strict webhook validation philosophy documented
- Credential key defaulting documented (not applicable here — no credentials)

### Vault API Paths

| Operation | Method | Path |
|-----------|--------|------|
| Create/update role | POST | `auth/{path}/role/{name}` |
| Read role | GET | `auth/{path}/role/{name}` |
| Delete role | DELETE | `auth/{path}/role/{name}` |
| List roles | LIST | `auth/{path}/roles` |

Only the first three operations are managed by this CRD.

### AliCloudAuthEngineRole — Vault API Field Reference

**Write (`POST auth/{path}/role/:role`) fields:**

| Vault API Field | Type | Default | Description |
|-----------------|------|---------|-------------|
| `arn` | string | required | The role's AliCloud RAM ARN |
| `token_ttl` | duration | `0` | Incremental lifetime for tokens |
| `token_max_ttl` | duration | `0` | Maximum lifetime for tokens |
| `token_policies` | array | `[]` | Token policies |
| `token_bound_cidrs` | array | `[]` | IP blocks for token use |
| `token_explicit_max_ttl` | duration | `0` | Hard cap max TTL |
| `token_no_default_policy` | bool | `false` | Exclude default policy |
| `token_num_uses` | integer | `0` | Max token uses (0=unlimited) |
| `token_period` | duration | `0` | Period for periodic tokens |
| `token_type` | string | `""` | Token type: service, batch, default |

**Deprecated fields (DO NOT include in CRD):**
- `policies` — use `token_policies` instead
- `ttl` — use `token_ttl` instead
- `max_ttl` — use `token_max_ttl` instead
- `period` — use `token_period` instead
- `bound_cidrs` — use `token_bound_cidrs` instead

**Read (`GET auth/{path}/role/:role`) sample response:**
```json
{
  "data": {
    "arn": "acs:ram::5138828231865461:role/dev-role",
    "policies": ["default", "dev", "prod"],
    "ttl": 1800000,
    "max_ttl": 1800000,
    "period": 0,
    "token_ttl": 1800000,
    "token_max_ttl": 1800000,
    "token_policies": ["default", "dev", "prod"],
    "token_bound_cidrs": [],
    "token_explicit_max_ttl": 0,
    "token_no_default_policy": false,
    "token_num_uses": 0,
    "token_period": 0,
    "token_type": ""
  }
}
```

**Critical observations for `IsEquivalentToDesiredState`:**
- Vault returns duration fields (`token_ttl`, `token_max_ttl`, `token_explicit_max_ttl`, `token_period`) as **integer seconds** (`json.Number`) — use `durationToSeconds()` in `toMap()`
- Vault returns integer fields (`token_num_uses`) as `json.Number` — use `json.Number(strconv.Itoa())` in `toMap()`
- Vault may return deprecated aliases (`ttl`, `max_ttl`, `period`) alongside the `token_*` versions — use `normalizeVaultReadAliases` to map them to canonical write keys
- Vault may return extra fields not in the write payload (deprecated fields, etc.) — `filterPayloadToDesiredKeys` handles this

### Vault Read-Response Alias Mapping

Follow the AppRole pattern using `normalizeVaultReadAliases`:

```go
var aliCloudVaultReadAliases = map[string]string{
    "ttl":     "token_ttl",
    "max_ttl": "token_max_ttl",
    "period":  "token_period",
}
```

These aliases ensure that if Vault returns the deprecated field names (without the `token_` prefix), the values are mapped to the canonical keys before comparison.

### CRD Field Spec — AliCloudAuthEngineRole

```go
type AliCloudAuthEngineRoleSpec struct {
    // Connection represents the information needed to connect to Vault. This operator uses the standard Vault environment variables to connect to Vault. If you need to override those settings and for example connect to a different Vault instance, you can do with this section of the CR.
    // +kubebuilder:validation:Optional
    Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

    // Authentication is the kube auth configuration to be used to execute this request
    // +kubebuilder:validation:Required
    Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

    // Path at which to make the configuration.
    // The final path in Vault will be {[spec.authentication.namespace]}/auth/{spec.path}/role/{metadata.name}.
    // The authentication role must have the following capabilities = [ "create", "read", "update", "delete"] on that path.
    // +kubebuilder:validation:Required
    Path vaultutils.Path `json:"path,omitempty"`

    AliCloudAuthRole `json:",inline"`

    // The name of the object created in Vault. If this is specified it takes precedence over {metadata.name}
    // +kubebuilder:validation:Optional
    // +kubebuilder:validation:Pattern:=`[a-z0-9]([-a-z0-9]*[a-z0-9])?`
    Name string `json:"name,omitempty"`
}

type AliCloudAuthRole struct {
    // ARN is the AliCloud RAM role ARN. Must correspond with the name of the role reflected in the arn.
    // +kubebuilder:validation:Required
    ARN string `json:"arn"`

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
- `ARN` is `+kubebuilder:validation:Required` (no default, no `omitempty`) — the AliCloud role ARN is mandatory
- All duration fields are `string` (user provides `"1h"`, `"10m"`); `toMap()` converts via `durationToSeconds()`
- Integer field `TokenNumUses` uses Go `int` in CRD; `toMap()` emits `json.Number`
- Array fields use `+listType=set`
- `TokenType` uses `+kubebuilder:validation:Enum`
- `TokenNoDefaultPolicy` has zero-value default (`false`) → uses `omitempty`
- No deprecated `policies`, `ttl`, `max_ttl`, `period`, `bound_cidrs` fields in the CRD

### `toMap()` Implementation Notes

```go
func (d *AliCloudAuthRole) toMap() map[string]any {
    payload := map[string]any{
        "arn":                     d.ARN,
        "token_ttl":              durationToSeconds(d.TokenTTL),
        "token_max_ttl":          durationToSeconds(d.TokenMaxTTL),
        "token_policies":         toInterfaceArray(d.TokenPolicies),
        "token_bound_cidrs":      toInterfaceArray(d.TokenBoundCIDRs),
        "token_explicit_max_ttl": durationToSeconds(d.TokenExplicitMaxTTL),
        "token_no_default_policy": d.TokenNoDefaultPolicy,
        "token_num_uses":         json.Number(strconv.Itoa(d.TokenNumUses)),
        "token_period":           durationToSeconds(d.TokenPeriod),
        "token_type":             d.TokenType,
    }
    return payload
}
```

**Key:** `arn` is always included (required field). All fields are emitted unconditionally (matching the AppRole pattern where `toMap()` emits all fields and `removeUnsetFields` + `filterPayloadToDesiredKeys` handle zero-value pruning during drift comparison). Duration fields use `durationToSeconds()` returning `json.Number`. Integer fields use `json.Number(strconv.Itoa())`. Array fields use `toInterfaceArray()`.

### `IsEquivalentToDesiredState` Notes

Follow the AppRole pattern:

```go
var aliCloudVaultReadAliases = map[string]string{
    "ttl":     "token_ttl",
    "max_ttl": "token_max_ttl",
    "period":  "token_period",
}

func (d *AliCloudAuthEngineRole) IsEquivalentToDesiredState(payload map[string]any) bool {
    desiredState := d.Spec.AliCloudAuthRole.toMap()
    normalizeVaultReadAliases(payload, aliCloudVaultReadAliases)
    removeUnsetFields(desiredState, payload)

    for _, boolKey := range []string{"token_no_default_policy"} {
        if boolVal, ok := desiredState[boolKey].(bool); ok && !boolVal {
            if _, inPayload := payload[boolKey]; !inPayload {
                delete(desiredState, boolKey)
            }
        }
    }

    filteredPayload := filterPayloadToDesiredKeys(desiredState, payload)
    setFields := []string{"token_policies", "token_bound_cidrs"}
    for _, key := range setFields {
        sortAnyStringSlice(desiredState, key)
        sortAnyStringSlice(filteredPayload, key)
    }
    return reflect.DeepEqual(desiredState, filteredPayload)
}
```

No write-only fields to strip (no credentials). No field remapping beyond the deprecated aliases. No create-only fields. Standard pipeline with set sorting for array fields.

### `GetPath()` Implementation

```go
func (d *AliCloudAuthEngineRole) GetPath() string {
    if d.Spec.Name != "" {
        return vaultutils.CleansePath("auth/" + string(d.Spec.Path) + "/role/" + d.Spec.Name)
    }
    return vaultutils.CleansePath("auth/" + string(d.Spec.Path) + "/role/" + d.Name)
}
```

### Webhook Rules

- `Default()` — no-op (no defaulting needed; `arn` is required, no credential key remapping)
- `ValidateCreate()` — call `isValid()` (basic validation)
- `ValidateUpdate()` — block `spec.path` changes, block `spec.name` changes, call `isValid()`
- `ValidateDelete()` — no-op

No credential validation needed (AliCloud auth has no operator-configured secrets). No complex cross-field validation (no multi-mode types).

### Controller Pattern

Standard `VaultResource` reconcile — simplest controller pattern (identical to AppRole):
- Embed `ReconcilerBase`
- Fetch instance → `prepareContext` → `NewVaultResource` → `Reconcile`
- `SetupWithManager`: `For()` with `NewDefaultPeriodicReconcilePredicate()`, no extra watches (no credential resolution)

### RBAC Markers

```go
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=alicloudauthengineroles,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=alicloudauthengineroles/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=alicloudauthengineroles/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=serviceaccounts/token,verbs=create
//+kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;patch
```

### Kubebuilder Markers Checklist

- All exported Spec fields: `+kubebuilder:validation:Required` or `+kubebuilder:validation:Optional`
- `ARN`: `+kubebuilder:validation:Required` (no default, no `omitempty`)
- `TokenType`: `+kubebuilder:validation:Enum:={"service","batch","default","default-service","default-batch",""}`
- `TokenNumUses`: `+kubebuilder:validation:Minimum=0`
- Root type: `+kubebuilder:object:root=true`, `+kubebuilder:subresource:status`
- Conditions field: `+patchMergeKey=type`, `+patchStrategy=merge`, `+listType=map`, `+listMapKey=type`
- Webhook markers: mutating + validating paths for `alicloudauthenginerole`

### Unit Test Requirements

**Tests (`alicloudauthenginerole_test.go`):**

1. `TestAliCloudAuthEngineRoleGetPath` — with `spec.name` override and without (falls back to `metadata.name`)
2. `TestAliCloudAuthRoleToMap` — verify all field keys match Vault API snake_case names, verify `arn` always present, verify `json.Number` for `token_num_uses`, verify `durationToSeconds` output for TTL fields
3. `TestAliCloudAuthRoleToMap_MinimalFields` — only `arn` set, verify minimal output includes `arn` and zero-valued token fields
4. `TestAliCloudAuthEngineRoleIsEquivalentToDesiredState_Match` — construct Vault-read-shaped payload with `json.Number` durations and integers, verify returns `true`
5. `TestAliCloudAuthEngineRoleIsEquivalentToDesiredState_Mismatch` — change `arn` value, verify returns `false`
6. `TestAliCloudAuthEngineRoleIsEquivalentToDesiredState_ExtraVaultFields` — add extra Vault fields (e.g., deprecated `policies`, `ttl`), verify still returns `true` (filtered out)
7. `TestAliCloudAuthEngineRoleIsEquivalentToDesiredState_DeprecatedAliases` — Vault returns `ttl`/`max_ttl`/`period` instead of `token_*` versions, verify `normalizeVaultReadAliases` maps them correctly and returns `true`
8. `TestAliCloudAuthEngineRoleIsDeletable` — returns `true`
9. `TestAliCloudAuthEngineRoleConditions` — Get/SetConditions round-trip

**Critical:** All Vault payload fixtures must use `json.Number` for numeric values, not Go `int` or `float64`.

### File Structure

| File | Action | Description |
|------|--------|-------------|
| `api/v1alpha1/alicloudauthenginerole_types.go` | NEW | Role CRD type, VaultObject, ConditionsAware, toMap, IsEquivalentToDesiredState |
| `api/v1alpha1/alicloudauthenginerole_webhook.go` | NEW | Role webhook — defaulter, validator, immutable path/name |
| `api/v1alpha1/alicloudauthenginerole_test.go` | NEW | Unit tests for toMap, IsEquivalentToDesiredState, GetPath |
| `internal/controller/alicloudauthenginerole_controller.go` | NEW | Role reconciler — standard VaultResource pattern |
| `cmd/main.go` | UPDATE | Register 1 controller + 1 webhook |
| `config/crd/kustomization.yaml` | UPDATE | Add 1 new CRD YAML file to resources list |
| `test/alicloudauthengine/` | NEW | Test YAML fixtures for role CR |
| `docs/auth-engines/alicloud.md` | NEW | Engine documentation per DNFR5 |
| `docs/auth-engines/index.md` | UPDATE | Add AliCloud row to Supported Auth Engines table |

### Files Being Modified — Current State

**`cmd/main.go`**: Currently registers ~48+ controllers and webhooks (including Epics 14 and 15 additions). New registrations follow the exact same pattern:
- Controller: `(&controller.AliCloudAuthEngineRoleReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "AliCloudAuthEngineRole")}).SetupWithManager(mgr)`
- Webhook (inside ENABLE_WEBHOOKS): `(&redhatcopv1alpha1.AliCloudAuthEngineRole{}).SetupWebhookWithManager(mgr)`

No existing behavior is changed — purely additive.

**`config/crd/kustomization.yaml`**: Add 1 new CRD YAML file to the `resources` list. Required for Helm chart build.

**`docs/auth-engines/index.md`**: Add AliCloud row to the Supported Auth Engines table:
```
| AliCloud | — | AliCloudAuthEngineRole | [alicloud.md](alicloud.md) |
```
Note: Config CRD column shows "—" because AliCloud has no config endpoint.

### Existing Patterns to Follow (Reference Files)

| Pattern | Reference File |
|---------|---------------|
| Auth engine role type (exact structural analog) | `api/v1alpha1/approleauthenginerole_types.go` |
| Auth role with standard token fields | `api/v1alpha1/approleauthenginerole_types.go` |
| Auth role webhook (immutable path/name) | `api/v1alpha1/approleauthenginerole_webhook.go` |
| Auth role controller (standard VaultResource) | `internal/controller/approleauthenginerole_controller.go` |
| normalizeVaultReadAliases usage | `api/v1alpha1/approleauthenginerole_types.go` |
| filterPayloadToDesiredKeys + removeUnsetFields | `api/v1alpha1/payload_filter.go` |
| durationToSeconds helper | `api/v1alpha1/utils/vaultutils.go` |
| toInterfaceArray helper | `api/v1alpha1/utils/vaultutils.go` |
| sortAnyStringSlice helper | `api/v1alpha1/approleauthenginerole_types.go` |
| Documentation template | `docs/engine-doc-template.md` |

### Anti-Patterns / DO NOT

- **DO NOT** create an `AliCloudAuthEngineConfig` CRD — AliCloud auth has no config endpoint; the mount itself is the config (confirmed via plugin source and API docs)
- **DO NOT** include deprecated `policies`, `ttl`, `max_ttl`, `period`, `bound_cidrs` fields in the CRD — use `token_policies`, `token_ttl`, etc. only
- **DO NOT** modify shared framework behavior (`reconcile_skeleton.go`, `vaultresourcereconciler.go`, etc.) — only add new files
- **DO NOT** derive expected payloads from `toMap()` in unit tests — construct independent Vault-read-shaped fixtures
- **DO NOT** use Go `int` or `float64` in unit test Vault payloads — use `json.Number` to match real Vault client behavior
- **DO NOT** emit TTL/duration fields as duration strings in `toMap()` — use `durationToSeconds()` to emit `json.Number` seconds matching Vault's read format
- **DO NOT** forget to add new CRD YAML file to `config/crd/kustomization.yaml` after `make manifests`
- **DO NOT** add credential resolution or controller watches — AliCloud auth has no operator-configured credentials
- **DO NOT** create integration tests — AliCloud is a cloud provider, cannot be installed in Kind (per "Skip it" rule). Document Skip explicitly.
- **DO NOT** confuse `AliCloudAuthEngineRole` (auth engine role, this story) with AliCloud secrets engine types — different Vault API paths and purposes

### Novelty Risk: LOW

This is a direct pattern copy of AppRole (14.1). The AliCloud auth role has identical structure: a required identifying field (`arn` instead of AppRole's various secret-id fields) plus standard token parameters. No credential resolution, no multi-mode validation, no config endpoint, no write-only fields to strip. The only AliCloud-specific element is the `arn` field, which is a simple required string.

- Duration/TTL fields in `toMap()` must use `durationToSeconds()` or `json.Number` (Vault-read format). Never emit raw duration strings.

### Project Structure Notes

- All new files follow existing naming conventions: `alicloudauthenginerole` lowercase for file names
- New files go in established directories (`api/v1alpha1/`, `internal/controller/`, `test/`, `docs/auth-engines/`)
- Test fixture directory `test/alicloudauthengine/` follows existing pattern (`test/approleauthengine/`, `test/kubernetesauthengine/`)
- No conflicts with existing code — purely additive

### References

- [Source: _bmad-output/planning-artifacts/epics.md#Epic-16, Story 16.2 — requirements and acceptance criteria]
- [Source: _bmad-output/project-context.md — all project rules, patterns, and conventions]
- [Source: _bmad-output/implementation-artifacts/14-1-approle-auth-engine-config-and-role-crds.md — exact structural analog (no config CRD pattern)]
- [Source: api/v1alpha1/approleauthenginerole_types.go — AppRole auth role type (primary pattern)]
- [Source: api/v1alpha1/approleauthenginerole_webhook.go — AppRole auth role webhook]
- [Source: internal/controller/approleauthenginerole_controller.go — standard VaultResource controller]
- [Source: api/v1alpha1/payload_filter.go — filterPayloadToDesiredKeys and removeUnsetFields]
- [Source: api/v1alpha1/utils/vaultutils.go — durationToSeconds, toInterfaceArray helpers]
- [Source: docs/engine-doc-template.md — documentation template]
- [Source: Vault AliCloud Auth Method API — https://developer.hashicorp.com/vault/api-docs/auth/alicloud]
- [Source: vault-plugin-auth-alicloud backend.go — confirms no config path registered]
- [Source: _bmad-output/implementation-artifacts/epic-15-retro-2026-08-18.md — TTL and webhook validation mandates]

## Dev Agent Record

### Agent Model Used

{{agent_model_name_version}}

### Debug Log References

### Completion Notes List

### File List

## Code Review Record

### Review Model Used

{{review_model_name_version}}

### Review Findings

### Decisions Needed / Decisions Taken

- No AliCloudAuthEngineConfig CRD — AliCloud auth has no config endpoint (confirmed via plugin source `backend.go` and official API docs). Decision taken pre-story (same as AppRole 14.1).

### Fixes Applied
