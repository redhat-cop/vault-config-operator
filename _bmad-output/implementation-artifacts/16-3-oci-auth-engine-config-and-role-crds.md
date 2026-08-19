# Story 16.3: OCI Auth Engine — Config and Role CRDs

Status: ready-for-dev

## Story

As an operator developer,
I want CRDs for OCIAuthEngineConfig and OCIAuthEngineRole,
So that Vault's OCI auth method can be managed declaratively.

## Acceptance Criteria

1. **Given** an OCIAuthEngineConfig CR is created with a `homeTenancyID` **When** the reconciler processes it **Then** the config is written to Vault at `auth/{path}/config` and ReconcileSuccessful=True

2. **Given** an OCIAuthEngineRole CR is created with `ocidList` and token parameters **When** the reconciler processes it **Then** the role exists in Vault at `auth/{path}/role/{name}` and ReconcileSuccessful=True

3. **Given** the OCIAuthEngineConfig CR is deleted **When** the reconciler processes deletion **Then** the K8s object is removed but Vault config is NOT deleted (`IsDeletable=false` — no DELETE endpoint for `auth/{path}/config`)

4. **Given** the OCIAuthEngineRole CR is deleted **When** the reconciler processes deletion **Then** the role is removed from Vault via `DELETE auth/{path}/role/{name}` (`IsDeletable=true`)

5. **Given** any OCI auth engine CR spec is updated **When** the reconciler processes the update **Then** the Vault resource reflects the updated values

6. **Given** any OCI auth engine CR is created or updated **When** the webhook validates it **Then** `spec.path` immutability is enforced on updates, and `spec.name` immutability is enforced on updates for OCIAuthEngineRole

7. **Given** the CRD types are implemented **When** the story is marked done **Then** a documentation file exists at `docs/auth-engines/oci.md` following `docs/engine-doc-template.md` (DNFR5)

## Tasks / Subtasks

- [ ] Task 1: Create `OCIAuthEngineConfig` type (AC: 1, 3, 5, 6)
  - [ ] 1.1: Create `api/v1alpha1/ociauthengineconfig_types.go` — Spec with `Connection`, `Authentication`, `Path`, inline `OCIAuthConfig` struct (single field: `HomeTenancyID`)
  - [ ] 1.2: Implement `VaultObject` interface — `GetPath()` returns `auth/{path}/config`, `GetPayload()`, `IsEquivalentToDesiredState()`, `IsDeletable()=false`
  - [ ] 1.3: Implement `ConditionsAware` interface — Status with `Conditions []metav1.Condition`
  - [ ] 1.4: Implement `toMap()` on `OCIAuthConfig` — emit `home_tenancy_id`
  - [ ] 1.5: Implement `IsEquivalentToDesiredState()` — standard `filterPayloadToDesiredKeys` (no write-only fields to strip)

- [ ] Task 2: Create `OCIAuthEngineRole` type (AC: 2, 4, 5, 6)
  - [ ] 2.1: Create `api/v1alpha1/ociauthenginerole_types.go` — Spec with `Connection`, `Authentication`, `Path`, `Name` (role name override), inline `OCIAuthRole` struct
  - [ ] 2.2: Implement `VaultObject` interface — `GetPath()` returns `auth/{path}/role/{name}`, `IsDeletable()=true`
  - [ ] 2.3: Implement `ConditionsAware` interface
  - [ ] 2.4: Implement `toMap()` on `OCIAuthRole` — emit `ocid_list` via `toInterfaceArray`, token fields via `durationToSeconds`/`json.Number`
  - [ ] 2.5: Implement `IsEquivalentToDesiredState()` — standard `filterPayloadToDesiredKeys`

- [ ] Task 3: Create webhooks (AC: 6)
  - [ ] 3.1: Create `api/v1alpha1/ociauthengineconfig_webhook.go` — `admission.Defaulter[*OCIAuthEngineConfig]`, `admission.Validator[*OCIAuthEngineConfig]`, immutable `spec.path`
  - [ ] 3.2: Create `api/v1alpha1/ociauthenginerole_webhook.go` — `admission.Defaulter[*OCIAuthEngineRole]`, `admission.Validator[*OCIAuthEngineRole]`, immutable `spec.path`/`spec.name`

- [ ] Task 4: Create controllers (AC: 1, 2, 3, 4, 5)
  - [ ] 4.1: Create `internal/controller/ociauthengineconfig_controller.go` — embed `ReconcilerBase`, simple `For()` with default periodic reconcile predicate (no credentials → no watches)
  - [ ] 4.2: Create `internal/controller/ociauthenginerole_controller.go` — simple `For()` with default periodic reconcile predicate (no watches)

- [ ] Task 5: Register in main.go (AC: 1, 2)
  - [ ] 5.1: Add controller registrations for both reconcilers
  - [ ] 5.2: Add webhook registrations inside `ENABLE_WEBHOOKS` guard for both types

- [ ] Task 6: Unit tests (AC: 1, 2, 5, 6)
  - [ ] 6.1: Create `api/v1alpha1/ociauthengineconfig_test.go` — test `toMap()`, test `IsEquivalentToDesiredState()` with independently constructed Vault-read-shaped fixtures, negative tests
  - [ ] 6.2: Create `api/v1alpha1/ociauthenginerole_test.go` — test `toMap()`, test `IsEquivalentToDesiredState()` with Vault-read-shaped fixtures (integer seconds as `json.Number`, `ocid_list` as `[]any`), negative tests

- [ ] Task 7: Test fixtures (AC: all)
  - [ ] 7.1: Create test YAML fixtures in `test/ociauthengine/` — config and role CRs
  - [ ] 7.2: Integration tests — SKIP (OCI is a cloud provider, cannot be installed in Kind)

- [ ] Task 8: CRD registration and code generation (AC: all)
  - [ ] 8.1: Run `make manifests generate fmt vet test`
  - [ ] 8.2: Add 2 new CRD YAML files to `config/crd/kustomization.yaml` (CRD registration checklist)
  - [ ] 8.3: Verify all existing tests still pass

- [ ] Task 9: Documentation (AC: 7)
  - [ ] 9.1: Create `docs/auth-engines/oci.md` following `docs/engine-doc-template.md`
  - [ ] 9.2: Update `docs/auth-engines/index.md` with link to new doc

## Dev Notes

### Integration Test Classification: SKIP

Per the project's Integration Test Infrastructure Philosophy:
> **Skip it** — For cloud providers and services that cannot be installed in Kind and are hard to mock (e.g., AWS, Azure, GCP, GitHub App, Quay), the integration test must not be run. These types rely on unit test coverage only.

OCI is a cloud provider that requires instance principal credentials from an OCI compute instance. It cannot be installed in Kind. No integration tests. Comprehensive unit tests for `toMap()`, `IsEquivalentToDesiredState()`, and webhook validation are the primary quality gate.

### Story Intelligence Chain — Previous Story Context

**Story 15.3 (Okta Auth Engine)** is the most recently completed auth engine CRD story:
- Established config + group pattern with `RootCredentialConfig` for API token, `api_token` stripping in `IsEquivalentToDesiredState`
- Credential defaulter sets `PasswordKey` only when empty (lesson from 14.2 review)
- Duration fields fixed to use `durationToSeconds()` during Epic 15 retro

**Story 14.2 (AWS Auth Engine)** is the closest **config + role** analog (two CRDs for config and role):
- AWSAuthEngineClientConfig has credential resolution — OCI config does NOT need credentials (no write-only fields)
- AWSAuthEngineRole has complex auth_type-specific validation — OCI role is much simpler (single `ocid_list` + standard token params)

**GCPAuthEngineConfig + GCPAuthEngineRole** are the **closest structural match** for OCI:
- GCPAuthEngineConfig: `IsDeletable()=false`, `GetPath()` returns `auth/{path}/config` — same as OCI
- GCPAuthEngineRole: `Name` override, `GetPath()` returns `auth/{path}/role/{name}`, `IsDeletable()=true`, standard token params — same as OCI
- GCPAuthEngineRole.toMap() uses `durationToSeconds()` for TTL fields, `json.Number` for integer fields — follow exactly

**Key difference from GCP/AWS/Okta configs:** OCI config has NO credential resolution. The `home_tenancy_id` is a plain string, not a secret. No `RootCredentialConfig`, no `PrepareInternalValues`, no Secret/RandomSecret watches. This makes `OCIAuthEngineConfig` the simplest auth config in the project.

**Epic 15 Retrospective action items (all completed):**
- TTL normalization: `durationToSeconds()` mandatory — apply to all OCI role duration fields
- Credential key defaulting: not applicable (OCI has no credentials)
- Strict webhook validation: apply — reject empty `homeTenancyID` on config, validate `ocidList` is not empty on role if desired

### Vault API Paths

| Operation | Method | Path |
|-----------|--------|------|
| Write config | POST | `auth/{path}/config` |
| Read config | GET | `auth/{path}/config` |
| Create/update role | POST | `auth/{path}/role/{name}` |
| Read role | GET | `auth/{path}/role/{name}` |
| Delete role | DELETE | `auth/{path}/role/{name}` |
| List roles | LIST | `auth/{path}/role` |

**Note:** There is no `DELETE auth/{path}/config` endpoint — config is tied to the mount lifecycle. `IsDeletable()=false`.

### OCIAuthEngineConfig — Vault API Field Reference

**Write (`POST auth/{path}/config`) fields:**
- `home_tenancy_id` (string: required) — The Tenancy OCID of the OCI account

**Read (`GET auth/{path}/config`) response:**
```json
{
  "data": {
    "home_tenancy_id": "ocid1.tenancy.oc1..aaaaaaaah7zkvaffv26pzyauoe2zbnionqvhvsexamplee557wakiofi4ysgqq"
  }
}
```

**No write-only fields.** Standard `filterPayloadToDesiredKeys` is sufficient for `IsEquivalentToDesiredState`. No secret stripping needed.

### OCIAuthEngineConfig — No Credential Resolution

Unlike AWS, GCP, LDAP, and Okta auth engine configs, OCI config does NOT require external credentials. The `home_tenancy_id` is a plain OCID string, not a secret. The OCI auth method itself uses instance principal credentials from the OCI compute instance where Vault runs — these are not managed by the operator.

This means:
- No `RootCredentialConfig` field in the spec
- `PrepareInternalValues` returns `nil` (no-op)
- No `setInternalCredentials` method needed
- Controller does NOT need Secret or RandomSecret watches
- Webhook does NOT need `ValidateCredentialSource()`

### OCIAuthEngineConfig — Controller Pattern

Simple `For()` with default periodic reconcile predicate. No watches needed. Follow `AWSAuthEngineIdentityConfigReconciler` or `GCPAuthEngineRoleReconciler` pattern (simplest controller variant).

### OCIAuthEngineRole — Vault API Field Reference

**Write (`POST auth/{path}/role/{name}`) fields:**
- `name` (string: required) — Name of the role (from URL path)
- `ocid_list` (string: required) — Comma-separated list of Group or Dynamic Group OCIDs
- `token_ttl` (integer: 0 or string: "") — Incremental lifetime for generated tokens
- `token_max_ttl` (integer: 0 or string: "") — Maximum lifetime for generated tokens
- `token_policies` (array: [] or comma-delimited string: "") — Token policies
- `policies` (array: [], deprecated) — Deprecated, use `token_policies`
- `token_bound_cidrs` (array: [] or comma-delimited string: "") — CIDR blocks restricting auth
- `token_explicit_max_ttl` (integer: 0 or string: "") — Hard cap max TTL
- `token_no_default_policy` (bool: false) — Exclude default policy
- `token_num_uses` (integer: 0) — Max token uses (0 = unlimited)
- `token_period` (integer: 0 or string: "") — Max allowed period for periodic tokens
- `token_type` (string: "") — Token type: service, batch, default, default-service, default-batch

**Read (`GET auth/{path}/role/{name}`) response:**
```json
{
  "data": {
    "ocid_list": [
      "ocid1.group.oc1..aaaaaaaaiqnblimpvmegkqh3bxilrdvjobr7qd223g275idcqhexamplefq",
      "ocid1.dynamicgroup.oc1..aaaaaaaa5hmfyrdaxvmt52ekju5n7ffamn2pdvxaq6esb2vzzoduexamplea"
    ],
    "token_ttl": 1800,
    "token_policies": ["dev", "prod"]
  }
}
```

**Critical observations:**
- Vault returns `ocid_list` as `[]any` (array), even though write accepts comma-separated string — model as `[]string` in CRD, emit as `toInterfaceArray()` in `toMap()`. Vault's write API accepts both string and array JSON formats.
- Vault returns TTL fields as integer seconds — use `durationToSeconds()` in `toMap()`
- `token_num_uses` is an integer — use `json.Number(strconv.FormatInt(...))`
- `token_period` is an integer — use `json.Number(strconv.FormatInt(...))`
- No write-only fields — standard `filterPayloadToDesiredKeys` is sufficient

### CRD Field Spec — OCIAuthConfig

```go
type OCIAuthConfig struct {
    // HomeTenancyID is the Tenancy OCID of the OCI account.
    // +kubebuilder:validation:Required
    HomeTenancyID string `json:"homeTenancyID"`
}
```

### CRD Field Spec — OCIAuthEngineConfigSpec (wrapper)

```go
type OCIAuthEngineConfigSpec struct {
    // +kubebuilder:validation:Optional
    Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

    // +kubebuilder:validation:Required
    Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

    // Path at which the OCI auth engine is mounted.
    // +kubebuilder:validation:Required
    Path vaultutils.Path `json:"path,omitempty"`

    OCIAuthConfig `json:",inline"`
}
```

### CRD Field Spec — OCIAuthRole

```go
type OCIAuthRole struct {
    // Name of the role.
    // +kubebuilder:validation:Required
    Name string `json:"name"`

    // OCIDList is a list of Group or Dynamic Group OCIDs that can take this role.
    // +kubebuilder:validation:Required
    // +listType=set
    OCIDList []string `json:"ocidList"`

    // TokenTTL is the incremental lifetime for generated tokens.
    // +kubebuilder:validation:Optional
    TokenTTL string `json:"tokenTTL,omitempty"`

    // TokenMaxTTL is the maximum lifetime for generated tokens.
    // +kubebuilder:validation:Optional
    TokenMaxTTL string `json:"tokenMaxTTL,omitempty"`

    // TokenPolicies are policies to encode onto generated tokens.
    // +kubebuilder:validation:Optional
    // +listType=set
    TokenPolicies []string `json:"tokenPolicies,omitempty"`

    // Policies is deprecated — use tokenPolicies instead.
    // +kubebuilder:validation:Optional
    // +listType=set
    Policies []string `json:"policies,omitempty"`

    // TokenBoundCIDRs are CIDR blocks restricting authentication and tying the token.
    // +kubebuilder:validation:Optional
    // +listType=set
    TokenBoundCIDRs []string `json:"tokenBoundCIDRs,omitempty"`

    // TokenExplicitMaxTTL is the hard cap max TTL for tokens.
    // +kubebuilder:validation:Optional
    TokenExplicitMaxTTL string `json:"tokenExplicitMaxTTL,omitempty"`

    // TokenNoDefaultPolicy if true, omits the default policy from generated tokens.
    // +kubebuilder:validation:Optional
    TokenNoDefaultPolicy bool `json:"tokenNoDefaultPolicy,omitempty"`

    // TokenNumUses is the max number of times a generated token may be used (0 = unlimited).
    // +kubebuilder:validation:Optional
    // +kubebuilder:validation:Minimum=0
    TokenNumUses int64 `json:"tokenNumUses,omitempty"`

    // TokenPeriod is the maximum allowed period for periodic tokens.
    // +kubebuilder:validation:Optional
    // +kubebuilder:validation:Minimum=0
    TokenPeriod int64 `json:"tokenPeriod,omitempty"`

    // TokenType is the type of token to generate.
    // +kubebuilder:validation:Optional
    // +kubebuilder:validation:Enum={"service","batch","default","default-service","default-batch"}
    TokenType string `json:"tokenType,omitempty"`
}
```

### CRD Field Spec — OCIAuthEngineRoleSpec (wrapper)

```go
type OCIAuthEngineRoleSpec struct {
    // +kubebuilder:validation:Optional
    Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

    // +kubebuilder:validation:Required
    Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

    // Path at which the OCI auth engine is mounted.
    // +kubebuilder:validation:Required
    Path vaultutils.Path `json:"path,omitempty"`

    OCIAuthRole `json:",inline"`
}
```

### `toMap()` Implementation Notes

**OCIAuthConfig.toMap():**
```go
func (c *OCIAuthConfig) toMap() map[string]any {
    payload := map[string]any{}
    payload["home_tenancy_id"] = c.HomeTenancyID
    return payload
}
```

**OCIAuthRole.toMap():**
```go
func (r *OCIAuthRole) toMap() map[string]any {
    payload := map[string]any{}
    payload["ocid_list"] = toInterfaceArray(r.OCIDList)
    payload["token_ttl"] = durationToSeconds(r.TokenTTL)
    payload["token_max_ttl"] = durationToSeconds(r.TokenMaxTTL)
    payload["token_policies"] = toInterfaceArray(r.TokenPolicies)
    payload["policies"] = toInterfaceArray(r.Policies)
    payload["token_bound_cidrs"] = toInterfaceArray(r.TokenBoundCIDRs)
    payload["token_explicit_max_ttl"] = durationToSeconds(r.TokenExplicitMaxTTL)
    payload["token_no_default_policy"] = r.TokenNoDefaultPolicy
    payload["token_num_uses"] = json.Number(strconv.FormatInt(r.TokenNumUses, 10))
    payload["token_period"] = json.Number(strconv.FormatInt(r.TokenPeriod, 10))
    payload["token_type"] = r.TokenType
    return payload
}
```

**Critical notes for `toMap()`:**
- `ocid_list` must use `toInterfaceArray()` — Vault reads back as `[]any`, even though the write API accepts comma-separated strings
- `token_ttl`, `token_max_ttl`, `token_explicit_max_ttl` are string duration fields — MUST use `durationToSeconds()` to convert to `json.Number` matching Vault read format (Epic 15 retro mandate)
- `token_num_uses` and `token_period` are `int64` — MUST emit as `json.Number(strconv.FormatInt(...))` (Vault returns `json.Number` via `UseNumber()`)
- `token_policies`, `policies`, `token_bound_cidrs` are list fields — use `toInterfaceArray()` to emit `[]any`
- Do NOT include `name` in `toMap()` — Vault's role API uses the URL path for the role name, not a body field. The `Name` field is only for `GetPath()`

### Kubebuilder Markers Checklist

- All exported Spec fields: `+kubebuilder:validation:Required` or `+kubebuilder:validation:Optional`
- Config `HomeTenancyID`: `+kubebuilder:validation:Required` (no default, no omitempty — required field)
- Role `Name`: `+kubebuilder:validation:Required` (no default, no omitempty — required field)
- Role `OCIDList`: `+kubebuilder:validation:Required`, `+listType=set` (no omitempty — required field)
- Role `TokenType`: `+kubebuilder:validation:Enum={"service","batch","default","default-service","default-batch"}`
- Role `TokenNumUses`, `TokenPeriod`: `+kubebuilder:validation:Minimum=0`
- Role `TokenNoDefaultPolicy`: zero-value default → use `omitempty`, no `kubebuilder:default`
- Role list fields (TokenPolicies, Policies, TokenBoundCIDRs): `+listType=set`
- Root types: `+kubebuilder:object:root=true`, `+kubebuilder:subresource:status`
- Conditions field: `+patchMergeKey=type`, `+patchStrategy=merge`, `+listType=map`, `+listMapKey=type`

### RBAC Markers

Config controller:
```go
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=ociauthengineconfigs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=ociauthengineconfigs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=ociauthengineconfigs/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=serviceaccounts/token,verbs=create
//+kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;patch
```

Role controller:
```go
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=ociauthengineroles,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=ociauthengineroles/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=ociauthengineroles/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=serviceaccounts/token,verbs=create
//+kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;patch
```

### File Structure

| File | Action | Description |
|------|--------|-------------|
| `api/v1alpha1/ociauthengineconfig_types.go` | NEW | Config CRD type, VaultObject, ConditionsAware, toMap |
| `api/v1alpha1/ociauthenginerole_types.go` | NEW | Role CRD type, VaultObject, ConditionsAware, toMap |
| `api/v1alpha1/ociauthengineconfig_webhook.go` | NEW | Config webhook — defaulter, validator, immutable path |
| `api/v1alpha1/ociauthenginerole_webhook.go` | NEW | Role webhook — defaulter, validator, immutable path/name |
| `api/v1alpha1/ociauthengineconfig_test.go` | NEW | Unit tests for config toMap, IsEquivalentToDesiredState |
| `api/v1alpha1/ociauthenginerole_test.go` | NEW | Unit tests for role toMap, IsEquivalentToDesiredState |
| `internal/controller/ociauthengineconfig_controller.go` | NEW | Config reconciler — simple VaultResource, no watches |
| `internal/controller/ociauthenginerole_controller.go` | NEW | Role reconciler — simple VaultResource, no watches |
| `cmd/main.go` | UPDATE | Register 2 controllers + 2 webhooks |
| `config/crd/kustomization.yaml` | UPDATE | Add 2 new CRD YAML files to resources list |
| `test/ociauthengine/` | NEW | Test YAML fixtures for both types |
| `docs/auth-engines/oci.md` | NEW | Engine documentation per DNFR5 |
| `docs/auth-engines/index.md` | UPDATE | Add link to oci.md |

### Files Being Modified — Current State

**`cmd/main.go`**: Currently registers ~46+ controllers and ~46+ webhooks (including Epic 15 additions). New registrations follow the exact same pattern:
- Controller: `(&controller.OCIAuthEngineConfigReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "OCIAuthEngineConfig")}).SetupWithManager(mgr)`
- Webhook (inside ENABLE_WEBHOOKS): `(&redhatcopv1alpha1.OCIAuthEngineConfig{}).SetupWebhookWithManager(mgr)`

No existing behavior is changed — this is purely additive.

**`config/crd/kustomization.yaml`**: Add the 2 new CRD YAML files to the `resources` list. Required for Helm chart build.

**`docs/auth-engines/index.md`**: Add a row to the Supported Auth Engines table:
```
| OCI | OCIAuthEngineConfig | OCIAuthEngineRole | [oci.md](oci.md) |
```

### Existing Patterns to Follow (Reference Files)

| Pattern | Reference File |
|---------|---------------|
| Auth engine config IsDeletable=false, GetPath `auth/{path}/config` | `api/v1alpha1/gcpauthengineconfig_types.go` |
| Auth engine config without credentials (simplest) | `api/v1alpha1/awsauthengineidentityconfig_types.go` |
| Auth engine role with Name override, GetPath `auth/{path}/role/{name}` | `api/v1alpha1/gcpauthenginerole_types.go` |
| Role toMap() with durationToSeconds, json.Number, toInterfaceArray | `api/v1alpha1/gcpauthenginerole_types.go` |
| Simple controller (no watches, no credentials) | `internal/controller/gcpauthenginerole_controller.go` |
| Auth config webhook (immutable path only) | `api/v1alpha1/gcpauthengineconfig_webhook.go` |
| Auth role webhook (immutable path+name) | `api/v1alpha1/gcpauthenginerole_webhook.go` |
| filterPayloadToDesiredKeys | `api/v1alpha1/payload_filter.go` |
| toInterfaceArray helper | `api/v1alpha1/utils/vaultutils.go` |
| durationToSeconds helper | `api/v1alpha1/sshsecretenginerole_types.go` |
| Documentation template | `docs/engine-doc-template.md` |

### Unit Test Requirements

**Config tests (`ociauthengineconfig_test.go`):**
1. `TestOCIAuthEngineConfig_toMap` — verify `home_tenancy_id` maps correctly from `HomeTenancyID`
2. `TestOCIAuthEngineConfig_IsEquivalentToDesiredState_Match` — construct Vault-read-shaped payload `{"home_tenancy_id": "ocid1.tenancy.oc1..example"}`, verify returns `true`
3. `TestOCIAuthEngineConfig_IsEquivalentToDesiredState_Mismatch` — change `home_tenancy_id` value, verify returns `false`
4. `TestOCIAuthEngineConfig_IsEquivalentToDesiredState_ExtraVaultFields` — add extra Vault-returned field, verify still returns `true` after filtering

**Role tests (`ociauthenginerole_test.go`):**
1. `TestOCIAuthEngineRole_toMap` — verify all fields: `ocid_list` as `[]any`, `token_ttl`/`token_max_ttl`/`token_explicit_max_ttl` as `json.Number` from `durationToSeconds`, `token_num_uses`/`token_period` as `json.Number`, list fields as `[]any`
2. `TestOCIAuthEngineRole_IsEquivalentToDesiredState_Match` — construct Vault-read-shaped payload with `ocid_list` as `[]any{"ocid1...", "ocid1..."}`, integer TTLs as `json.Number("1800")`, `token_policies` as `[]any{"dev", "prod"}`, verify returns `true`
3. `TestOCIAuthEngineRole_IsEquivalentToDesiredState_Mismatch` — change `ocid_list`, verify returns `false`
4. `TestOCIAuthEngineRole_IsEquivalentToDesiredState_ExtraVaultFields` — extra fields from Vault, verify filtering

### Anti-Patterns / DO NOT

- **DO NOT** create integration tests for these types — OCI is a cloud provider that requires instance principal credentials from an OCI compute instance, cannot be installed in Kind (per "Skip it" rule)
- **DO NOT** add credential resolution (`RootCredentialConfig`, `setInternalCredentials`) to the config type — OCI config has no credentials; `home_tenancy_id` is a plain OCID string, not a secret
- **DO NOT** add Secret or RandomSecret watches to either controller — no credential rotation needed (neither type has write-only credentials)
- **DO NOT** modify shared framework behavior (reconcile_skeleton.go, vaultresourcereconciler.go, etc.) — only add new files
- **DO NOT** derive expected payloads from `toMap()` in unit tests — construct independent Vault-read-shaped fixtures
- **DO NOT** forget to add new CRD YAML files to `config/crd/kustomization.yaml` after `make manifests`
- **DO NOT** use Go `int` or `float64` in unit test Vault payloads — use `json.Number` to match real Vault client behavior
- **DO NOT** add `Name` to the role `toMap()` output — Vault's role API uses the URL path for the role name, not a body field. `Name` is only for `GetPath()` override
- **DO NOT** emit raw duration strings in `toMap()` — all duration/TTL fields MUST use `durationToSeconds()` (Epic 15 retro mandate, not optional)
- **DO NOT** emit integer fields (`token_num_uses`, `token_period`) as bare Go `int64` — MUST use `json.Number(strconv.FormatInt(...))` to match Vault's `UseNumber()` decoder

### Novelty Risk: LOW

Both CRD types follow well-established patterns from existing auth engine implementations. OCIAuthEngineConfig is the **simplest** auth config in the project — a single required string field, no credentials, no write-only fields. OCIAuthEngineRole follows the standard GCP/AWS role pattern with token parameters. No novel architectural patterns required.

### Project Structure Notes

- All new files follow existing naming conventions: lowercase type name for files
- New files go in established directories (`api/v1alpha1/`, `internal/controller/`, `test/`, `docs/auth-engines/`)
- Test fixture directory `test/ociauthengine/` follows the existing pattern (`test/gcpauthengine/`, `test/oktaauthengine/`)
- No conflicts with existing code — purely additive

### References

- [Source: _bmad-output/planning-artifacts/epics.md#Epic-16, Story 16.3 — requirements and acceptance criteria]
- [Source: _bmad-output/project-context.md — all project rules, patterns, and conventions]
- [Source: _bmad-output/implementation-artifacts/epic-15-retro-2026-08-18.md — TTL mandate, credential defaulting, webhook philosophy]
- [Source: api/v1alpha1/gcpauthengineconfig_types.go — GCP auth config with IsDeletable=false, GetPath auth/{path}/config]
- [Source: api/v1alpha1/gcpauthenginerole_types.go — GCP auth role with Name override, durationToSeconds, json.Number]
- [Source: api/v1alpha1/awsauthengineidentityconfig_types.go — simplest auth config (no credentials, simple controller)]
- [Source: internal/controller/gcpauthenginerole_controller.go — simple auth role controller (no watches)]
- [Source: api/v1alpha1/gcpauthengineconfig_webhook.go — auth config webhook pattern]
- [Source: api/v1alpha1/gcpauthenginerole_webhook.go — auth role webhook pattern (immutable path+name)]
- [Source: api/v1alpha1/payload_filter.go — filterPayloadToDesiredKeys]
- [Source: docs/engine-doc-template.md — documentation template]
- [Source: Vault OCI Auth Method API — https://docs.hashicorp.com/vault/api-docs/auth/oci]
- [Source: _bmad-output/implementation-artifacts/14-2-aws-auth-engine-config-and-role-crds.md — config+role predecessor story]
- [Source: _bmad-output/implementation-artifacts/15-3-okta-auth-engine-config-and-group-crds.md — most recent predecessor story]

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

- Design decision: `OCIAuthEngineConfig.IsDeletable() = false` — consistent with all other auth engine configs (GCP, LDAP, JWT/OIDC, Azure, Kubernetes, Okta). No DELETE endpoint for `auth/{path}/config`.
- Design decision: No credential resolution for config — `home_tenancy_id` is a plain OCID string, not a secret. OCI auth uses instance principal credentials from the compute instance (not managed by the operator).
- Design decision: `ocid_list` modeled as `[]string` in CRD, emitted as `toInterfaceArray()` — Vault reads back an array even though write accepts comma-separated string. Vault's API accepts both formats. Array matches the read response shape for accurate drift detection.

### Fixes Applied
