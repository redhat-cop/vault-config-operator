# Story 15.3: Okta Auth Engine — Config and Group CRDs

Status: ready-for-dev

## Story

As an operator developer,
I want CRDs for OktaAuthEngineConfig and OktaAuthEngineGroup,
So that Vault's Okta auth method can be managed declaratively.

## Acceptance Criteria

1. **Given** an OktaAuthEngineConfig CR is created with Okta org name, API token (from K8s Secret), and base URL **When** the reconciler processes it **Then** the config is written to Vault at `auth/{path}/config` and ReconcileSuccessful=True

2. **Given** an OktaAuthEngineGroup CR is created with group name and policies **When** the reconciler processes it **Then** the group mapping exists in Vault at `auth/{path}/groups/{name}` and ReconcileSuccessful=True

3. **Given** the OktaAuthEngineConfig CR is deleted **When** the reconciler processes deletion **Then** the K8s object is removed but Vault config is NOT deleted (`IsDeletable=false` — auth engine configs are not deletable; the mount owns the config lifecycle)

4. **Given** the OktaAuthEngineGroup CR is deleted **When** the reconciler processes deletion **Then** the group mapping is removed from Vault via `DELETE auth/{path}/groups/{name}` (`IsDeletable=true`)

5. **Given** any Okta auth engine CR spec is updated **When** the reconciler processes the update **Then** the Vault resource reflects the updated values

6. **Given** any Okta auth engine CR is created or updated **When** the webhook validates it **Then** `spec.path` immutability is enforced on updates, and `spec.name` immutability is enforced on updates for OktaAuthEngineGroup

7. **Given** the OktaAuthEngineConfig CR has an `OktaCredentials` field referencing a K8s Secret **When** the Secret is updated **Then** the reconciler re-reads the API token and writes the updated config to Vault

8. **Given** the CRD types are implemented **When** the story is marked done **Then** a documentation file exists at `docs/auth-engines/okta.md` following `docs/engine-doc-template.md` (DNFR5)

## Tasks / Subtasks

- [ ] Task 1: Create `OktaAuthEngineConfig` type (AC: 1, 3, 5, 6, 7)
  - [ ] 1.1: Create `api/v1alpha1/oktaauthengineconfig_types.go` — Spec with `Connection`, `Authentication`, `Path`, inline `OktaAuthConfig` struct, `OktaCredentials` (`RootCredentialConfig` with passwordKey="api_token")
  - [ ] 1.2: Implement `VaultObject` interface — `GetPath()` returns `auth/{path}/config`, `GetPayload()`, `IsEquivalentToDesiredState()`, `PrepareInternalValues()`, `IsDeletable()=false`
  - [ ] 1.3: Implement `ConditionsAware` interface — Status with `Conditions []metav1.Condition`
  - [ ] 1.4: Implement `setInternalCredentials()` — resolve api_token from K8s Secret, VaultSecret, or RandomSecret
  - [ ] 1.5: Implement `toMap()` on `OktaAuthConfig` — convert to Vault API snake_case fields, include resolved `api_token` from internal field
  - [ ] 1.6: Implement `IsEquivalentToDesiredState()` — must delete `api_token` from desired state (Vault never returns it on read), then `filterPayloadToDesiredKeys`

- [ ] Task 2: Create `OktaAuthEngineGroup` type (AC: 2, 4, 5, 6)
  - [ ] 2.1: Create `api/v1alpha1/oktaauthenginegroup_types.go` — Spec with `Connection`, `Authentication`, `Path`, `Name` (group name override), `Policies` (comma-separated string)
  - [ ] 2.2: Implement `VaultObject` interface — `GetPath()` returns `auth/{path}/groups/{name}`, `IsDeletable()=true`, no `PrepareInternalValues` needed
  - [ ] 2.3: Implement `ConditionsAware` interface
  - [ ] 2.4: Implement `toMap()` — emit `policies` field
  - [ ] 2.5: Implement `IsEquivalentToDesiredState()` — standard `filterPayloadToDesiredKeys`

- [ ] Task 3: Create webhooks (AC: 6)
  - [ ] 3.1: Create `api/v1alpha1/oktaauthengineconfig_webhook.go` — `admission.Defaulter[*OktaAuthEngineConfig]`, `admission.Validator[*OktaAuthEngineConfig]`, immutable `spec.path`, credential validation via `ValidateCredentialSource()`
  - [ ] 3.2: Create `api/v1alpha1/oktaauthenginegroup_webhook.go` — `admission.Defaulter[*OktaAuthEngineGroup]`, `admission.Validator[*OktaAuthEngineGroup]`, immutable `spec.path`/`spec.name`

- [ ] Task 4: Create controllers (AC: 1, 2, 3, 4, 5, 7)
  - [ ] 4.1: Create `internal/controller/oktaauthengineconfig_controller.go` — embed `ReconcilerBase`, standard VaultResource reconcile, watches on `corev1.Secret` and `RandomSecret` for API token rotation
  - [ ] 4.2: Create `internal/controller/oktaauthenginegroup_controller.go` — simple `For()` with default periodic reconcile predicate (no watches needed)

- [ ] Task 5: Register in main.go (AC: 1, 2)
  - [ ] 5.1: Add controller registrations for both reconcilers
  - [ ] 5.2: Add webhook registrations inside `ENABLE_WEBHOOKS` guard for both types

- [ ] Task 6: Unit tests (AC: 1, 2, 5, 6)
  - [ ] 6.1: Create `api/v1alpha1/oktaauthengineconfig_test.go` — test `toMap()`, test `IsEquivalentToDesiredState()` with independently constructed Vault-read-shaped fixtures (api_token stripped), negative tests
  - [ ] 6.2: Create `api/v1alpha1/oktaauthenginegroup_test.go` — test `toMap()`, test `IsEquivalentToDesiredState()` with Vault-read fixture, negative tests

- [ ] Task 7: Test fixtures (AC: all)
  - [ ] 7.1: Create test YAML fixtures in `test/oktaauthengine/` — config and group CRs
  - [ ] 7.2: Integration tests — SKIP (Okta is a cloud auth provider, falls under "Skip it" per project integration test philosophy)

- [ ] Task 8: CRD registration and code generation (AC: all)
  - [ ] 8.1: Run `make manifests generate fmt vet test`
  - [ ] 8.2: Add 2 new CRD YAML files to `config/crd/kustomization.yaml` (CRD registration checklist)
  - [ ] 8.3: Verify all existing tests still pass

- [ ] Task 9: Documentation (AC: 8)
  - [ ] 9.1: Create `docs/auth-engines/okta.md` following `docs/engine-doc-template.md`
  - [ ] 9.2: Update `docs/auth-engines/index.md` with link to new doc

## Dev Notes

### Integration Test Classification: SKIP

Per the project's Integration Test Infrastructure Philosophy:
> **Skip it** — For cloud providers and services that cannot be installed in Kind and are hard to mock (e.g., AWS, Azure, GCP, GitHub App, Quay), the integration test must not be run. These types rely on unit test coverage only.

Okta is a cloud identity provider that cannot be installed in Kind. No integration tests. Comprehensive unit tests for `toMap()`, `IsEquivalentToDesiredState()`, and webhook validation are the primary quality gate.

### Story Intelligence Chain — Previous Story Context

**Story 14.2 (AWS Auth Engine)** is the most recently completed auth engine CRD story:
- Established three-CRD pattern (two configs + role) for a complex auth engine
- Key patterns: `RootCredentialConfig` for credential resolution, `secret_key` stripping in `IsEquivalentToDesiredState`, controller with Secret/RandomSecret watches for credential rotation
- Review findings identified issues with credential defaults (`usernameKey`/`passwordKey`) — the webhook defaulter must NOT unconditionally overwrite explicit key mappings; instead, only set defaults when the fields are empty
- Review findings also identified `max_retries` should use `json.Number` — apply `json.Number` to all Vault-facing numeric fields in `toMap()`
- All 14.2 review findings are recorded but unfixed — avoid repeating the same patterns

**Story 14.1 (AppRole Auth Engine)** completed:
- AppRole had no config endpoint (mount is the config) — different pattern from Okta which HAS a config endpoint
- Established the Epic 14 auth engine pattern

**LDAP Auth Engine (LDAPAuthEngineConfig + LDAPAuthEngineGroup)** is the **closest structural analog** for this story:
- LDAPAuthEngineConfig: auth engine config with `IsDeletable()=false`, credential resolution via `BindCredentials` (`RootCredentialConfig`), `GetPath()` returns `auth/{path}/config`, controller with Secret/RandomSecret watches
- LDAPAuthEngineGroup: group mapping CRD with `Name` override, `Policies` string, `GetPath()` returns `auth/{path}/groups/{name}`, `IsDeletable()=true`, simple controller (no watches)
- The LDAPAuthEngineGroup pattern is the **exact template** for OktaAuthEngineGroup

**Epic 13 Retrospective action items (still pending):**
- Validation matrix requirement for multi-mode types — NOT applicable to Okta (no mode-specific fields)
- Novelty risk flag — include in this story

### Vault API Paths

| Operation | Method | Path |
|-----------|--------|------|
| Write config | POST | `auth/{path}/config` |
| Read config | GET | `auth/{path}/config` |
| Register group | POST | `auth/{path}/groups/{name}` |
| Read group | GET | `auth/{path}/groups/{name}` |
| Delete group | DELETE | `auth/{path}/groups/{name}` |
| List groups | LIST | `auth/{path}/groups` |
| Register user | POST | `auth/{path}/users/{username}` |
| Read user | GET | `auth/{path}/users/{username}` |
| Delete user | DELETE | `auth/{path}/users/{username}` |
| List users | LIST | `auth/{path}/users` |

**Note:** User endpoints are documented here for completeness but are **NOT in scope** for this story — only config and group CRDs are implemented.

### OktaAuthEngineConfig — Vault API Field Reference

**Write (`POST auth/{path}/config`) fields:**
- `org_name` (string: required) — Name of the organization to be used in the Okta API
- `api_token` (string: "") — Okta API token, required for Okta group membership queries. Write-only.
- `base_url` (string: "") — Base domain for API requests. Defaults to "okta.com". Other values: "oktapreview.com", "okta-emea.com"
- `bypass_okta_mfa` (bool: false) — Bypass Okta MFA request
- `token_ttl` (integer: 0 or string: "") — Incremental lifetime for generated tokens
- `token_max_ttl` (integer: 0 or string: "") — Maximum lifetime for generated tokens
- `token_policies` (array: [] or string: "") — Token policies to encode onto generated tokens
- `policies` (array: [] or string: "", DEPRECATED) — Use `token_policies` instead
- `token_bound_cidrs` (array: [] or string: "") — CIDR blocks restricting authentication
- `token_explicit_max_ttl` (integer: 0 or string: "") — Hard cap max TTL
- `token_no_default_policy` (bool: false) — Exclude default policy from tokens
- `token_num_uses` (integer: 0) — Max token uses (0 = unlimited)
- `token_period` (integer: 0 or string: "") — Max allowed period for periodic tokens
- `token_type` (string: "") — Token type: service, batch, default, default-service, default-batch

**Read (`GET auth/{path}/config`) response — `api_token` is NEVER returned:**
```json
{
  "data": {
    "base_url": "okta.com",
    "bypass_okta_mfa": false,
    "org_name": "example",
    "token_bound_cidrs": [],
    "token_explicit_max_ttl": 0,
    "token_max_ttl": 0,
    "token_no_default_policy": false,
    "token_num_uses": 0,
    "token_period": 0,
    "token_policies": [],
    "token_ttl": 0,
    "token_type": "default"
  }
}
```

**Critical:** `api_token` is write-only. Must be deleted from `desiredState` before `IsEquivalentToDesiredState` comparison. Follow `AWSAuthEngineClientConfig` pattern (deletes `secret_key`).

### Critical: `IsEquivalentToDesiredState` for Config — API Token Stripping

Vault never returns `api_token` on read. The implementation must:
1. Build `desiredState` from `OktaAuthConfig.toMap()`
2. `delete(desiredState, "api_token")` — remove before comparison
3. Use `filterPayloadToDesiredKeys(desiredState, payload)` → `reflect.DeepEqual`

Follow the established pattern from `AWSAuthEngineClientConfig.IsEquivalentToDesiredState()` and `LDAPAuthEngineConfig` (which deletes `bindpass`).

### OktaAuthEngineConfig — Credential Resolution (API Token)

The Okta API token must be resolved from one of three sources using `RootCredentialConfig`:
- **K8s Secret**: passwordKey defaults to "api_token"
- **VaultSecret**: same key mapping from a Vault KV path
- **RandomSecret**: retrieve from RandomSecret's Vault path

Pattern: Use `RootCredentialConfig` with custom default key (`passwordKey: "api_token"`). The usernameKey is not used for Okta (there is no separate username credential for the config endpoint — `org_name` is specified directly in the spec). Store the resolved API token in an unexported field (`retrievedAPIToken` with `json:"-"`). Follow `LDAPAuthEngineConfig.setInternalCredentials()` for the credential resolution pattern, adapted for a single credential (API token only, no username).

**Webhook Defaulter:** Set `OktaCredentials.PasswordKey` to `"api_token"` when empty. Do NOT unconditionally overwrite — only set when the field is empty string (lesson from 14.2 review findings).

### OktaAuthEngineConfig — Controller Pattern

The config controller uses the **standard VaultResource reconcile** (NOT always-write) because the read endpoint returns enough fields for meaningful drift detection (everything except `api_token`). The `IsEquivalentToDesiredState` strips `api_token` to avoid false drift.

The controller MUST include Secret and RandomSecret watches for API token rotation detection, following `LDAPAuthEngineConfigReconciler.SetupWithManager()` (simplified — no TLS watch needed for Okta).

### OktaAuthEngineGroup — Vault API Field Reference

**Write (`POST auth/{path}/groups/{name}`) fields:**
- `name` (string: required) — Group name (from URL path)
- `policies` (array: [] or string: "") — Policies associated with the group

**Read (`GET auth/{path}/groups/{name}`) response:**
```json
{
  "data": {
    "policies": ["default", "admin"]
  }
}
```

No write-only fields. Standard `filterPayloadToDesiredKeys` is sufficient.

### OktaAuthEngineGroup — Pattern

Follow `LDAPAuthEngineGroup` exactly:
- Spec: `Connection`, `Authentication`, `Path`, `Name` (group name override), `Policies` (string — comma-separated list)
- `GetPath()`: `auth/{path}/groups/{name}` using `vaultutils.CleansePath`
- `IsDeletable()`: `true`
- `toMap()`: emit `policies` field (as string, NOT array — matches LDAP group pattern)
- `IsEquivalentToDesiredState()`: standard `filterPayloadToDesiredKeys`
- Simple controller: `For()` with default periodic reconcile predicate, no watches

**Important:** The Vault API accepts `policies` as either an array or comma-separated string. The LDAP group pattern uses a single string. Follow the same pattern for consistency.

### CRD Field Spec — OktaAuthConfig

```go
type OktaAuthConfig struct {
    // OrgName is the name of the organization to be used in the Okta API.
    // +kubebuilder:validation:Required
    OrgName string `json:"orgName"`

    // BaseURL is the base domain for Okta API requests.
    // If unset, "okta.com" is used. Other valid values: "oktapreview.com", "okta-emea.com".
    // +kubebuilder:validation:Optional
    BaseURL string `json:"baseURL,omitempty"`

    // BypassOktaMFA bypasses an Okta MFA request. Useful when using Vault's built-in MFA.
    // +kubebuilder:validation:Optional
    BypassOktaMFA bool `json:"bypassOktaMFA,omitempty"`

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
    TokenPeriod string `json:"tokenPeriod,omitempty"`

    // TokenType is the type of token to generate.
    // +kubebuilder:validation:Optional
    // +kubebuilder:validation:Enum={"service","batch","default","default-service","default-batch"}
    TokenType string `json:"tokenType,omitempty"`

    retrievedAPIToken string `json:"-"`
}
```

### CRD Field Spec — OktaAuthEngineConfigSpec (wrapper)

```go
type OktaAuthEngineConfigSpec struct {
    // +kubebuilder:validation:Optional
    Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

    Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

    // Path at which the Okta auth engine is mounted.
    // +kubebuilder:validation:Required
    Path vaultutils.Path `json:"path,omitempty"`

    OktaAuthConfig `json:",inline"`

    // OktaCredentials is used to provide the Okta API token.
    // The API token can be sourced from a Kubernetes Secret, VaultSecret, or RandomSecret.
    // +kubebuilder:validation:Optional
    OktaCredentials vaultutils.RootCredentialConfig `json:"oktaCredentials,omitempty"`
}
```

### CRD Field Spec — OktaAuthEngineGroup

Follow LDAPAuthEngineGroup pattern exactly:

```go
type OktaAuthEngineGroupSpec struct {
    // +kubebuilder:validation:Optional
    Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

    // +kubebuilder:validation:Required
    Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

    // Path at which the Okta auth engine is mounted.
    // +kubebuilder:validation:Required
    Path vaultutils.Path `json:"path,omitempty"`

    // Name of the Okta group.
    // +kubebuilder:validation:Required
    Name string `json:"name,omitempty"`

    // Policies is a comma-separated list of policies associated with the group.
    // +kubebuilder:validation:Optional
    Policies string `json:"policies,omitempty"`
}
```

### `toMap()` Implementation Notes

**OktaAuthConfig.toMap():**
```go
func (c *OktaAuthConfig) toMap() map[string]any {
    payload := map[string]any{}
    payload["org_name"] = c.OrgName
    payload["api_token"] = c.retrievedAPIToken
    payload["base_url"] = c.BaseURL
    payload["bypass_okta_mfa"] = c.BypassOktaMFA
    payload["token_ttl"] = c.TokenTTL
    payload["token_max_ttl"] = c.TokenMaxTTL
    payload["token_policies"] = toInterfaceArray(c.TokenPolicies)
    payload["token_bound_cidrs"] = toInterfaceArray(c.TokenBoundCIDRs)
    payload["token_explicit_max_ttl"] = c.TokenExplicitMaxTTL
    payload["token_no_default_policy"] = c.TokenNoDefaultPolicy
    payload["token_num_uses"] = json.Number(strconv.FormatInt(c.TokenNumUses, 10))
    payload["token_period"] = c.TokenPeriod
    payload["token_type"] = c.TokenType
    return payload
}
```

**Critical notes for `toMap()`:**
- `api_token` must come from `retrievedAPIToken` (internal field set by PrepareInternalValues), NOT from the spec directly
- `token_num_uses` must be emitted as `json.Number` (Vault returns integer seconds as `json.Number`)
- `token_ttl`, `token_max_ttl`, `token_explicit_max_ttl`, `token_period` are string duration fields — Vault accepts both string and integer format. Vault READ returns integer seconds (0). For `IsEquivalentToDesiredState`, the comparison after `filterPayloadToDesiredKeys` must handle this. If the user specifies "0" or empty string, the zero value should match Vault's `0`. For non-zero durations, use `durationToSeconds()` to normalize to Vault-read format.
- `toInterfaceArray()` for list fields to emit `[]any` matching Vault's read format

**OktaAuthEngineGroup.toMap():**
```go
func (g *OktaAuthEngineGroup) toMap() map[string]any {
    payload := map[string]any{}
    payload["policies"] = g.Spec.Policies
    return payload
}
```

Follow LDAPAuthEngineGroup.toMap() exactly — emit `policies` as a string.

### Kubebuilder Markers Checklist

- All exported Spec fields: `+kubebuilder:validation:Required` or `+kubebuilder:validation:Optional`
- Config `OrgName`: `+kubebuilder:validation:Required` (no default, no omitempty — required field)
- Config `TokenType`: `+kubebuilder:validation:Enum={"service","batch","default","default-service","default-batch"}`
- Config `TokenNumUses`: `+kubebuilder:validation:Minimum=0`
- Config `BypassOktaMFA`, `TokenNoDefaultPolicy`: zero-value defaults → use `omitempty`, no `kubebuilder:default`
- Config list fields (TokenPolicies, TokenBoundCIDRs): `+listType=set`
- Root types: `+kubebuilder:object:root=true`, `+kubebuilder:subresource:status`
- Conditions field: `+patchMergeKey=type`, `+patchStrategy=merge`, `+listType=map`, `+listMapKey=type`
- Group `Name`: `+kubebuilder:validation:Required`

### RBAC Markers

Config controller:
```go
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=oktaauthengineconfigs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=oktaauthengineconfigs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=oktaauthengineconfigs/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=serviceaccounts/token,verbs=create
//+kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;patch
```

Group controller:
```go
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=oktaauthenginegroups,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=oktaauthenginegroups/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=oktaauthenginegroups/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=serviceaccounts/token,verbs=create
//+kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;patch
```

### File Structure

| File | Action | Description |
|------|--------|-------------|
| `api/v1alpha1/oktaauthengineconfig_types.go` | NEW | Config CRD type, VaultObject, ConditionsAware, toMap, credential resolution |
| `api/v1alpha1/oktaauthenginegroup_types.go` | NEW | Group CRD type, VaultObject, ConditionsAware, toMap |
| `api/v1alpha1/oktaauthengineconfig_webhook.go` | NEW | Config webhook — defaulter, validator, immutable path, credential validation |
| `api/v1alpha1/oktaauthenginegroup_webhook.go` | NEW | Group webhook — defaulter, validator, immutable path/name |
| `api/v1alpha1/oktaauthengineconfig_test.go` | NEW | Unit tests for config toMap, IsEquivalentToDesiredState |
| `api/v1alpha1/oktaauthenginegroup_test.go` | NEW | Unit tests for group toMap, IsEquivalentToDesiredState |
| `internal/controller/oktaauthengineconfig_controller.go` | NEW | Config reconciler with Secret/RandomSecret watches |
| `internal/controller/oktaauthenginegroup_controller.go` | NEW | Group reconciler — simple VaultResource |
| `cmd/main.go` | UPDATE | Register 2 controllers + 2 webhooks |
| `config/crd/kustomization.yaml` | UPDATE | Add 2 new CRD YAML files to resources list |
| `test/oktaauthengine/` | NEW | Test YAML fixtures for both types |
| `docs/auth-engines/okta.md` | NEW | Engine documentation per DNFR5 |
| `docs/auth-engines/index.md` | UPDATE | Add link to okta.md |

### Files Being Modified — Current State

**`cmd/main.go`**: Currently registers ~44+ controllers and ~44+ webhooks (including Epic 14 additions). New registrations follow the exact same pattern:
- Controller: `(&controller.OktaAuthEngineConfigReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "OktaAuthEngineConfig")}).SetupWithManager(mgr)`
- Webhook (inside ENABLE_WEBHOOKS): `(&redhatcopv1alpha1.OktaAuthEngineConfig{}).SetupWebhookWithManager(mgr)`

No existing behavior is changed — this is purely additive.

**`config/crd/kustomization.yaml`**: Add the 2 new CRD YAML files to the `resources` list. Required for Helm chart build.

**`docs/auth-engines/index.md`**: Add a row to the Supported Auth Engines table:
```
| Okta | OktaAuthEngineConfig | OktaAuthEngineGroup | [okta.md](okta.md) |
```

### Existing Patterns to Follow (Reference Files)

| Pattern | Reference File |
|---------|---------------|
| Auth engine config with credential resolution | `api/v1alpha1/ldapauthengineconfig_types.go` |
| Auth engine config IsDeletable=false | `api/v1alpha1/gcpauthengineconfig_types.go` |
| Auth engine config GetPath (auth/{path}/config) | `api/v1alpha1/gcpauthengineconfig_types.go` |
| Auth engine group with Name override | `api/v1alpha1/ldapauthenginegroup_types.go` |
| Auth engine group GetPath (auth/{path}/groups/{name}) | `api/v1alpha1/ldapauthenginegroup_types.go` |
| API-token stripping in IsEquivalentToDesiredState | `api/v1alpha1/awsauthengineconfig_types.go` (deletes `secret_key`) |
| Config controller with Secret+RandomSecret watches | `internal/controller/ldapauthengineconfig_controller.go` |
| Simple group controller (no watches) | `internal/controller/ldapauthenginegroup_controller.go` |
| Auth config webhook (immutable path only) | `api/v1alpha1/gcpauthengineconfig_webhook.go` |
| Auth group webhook (immutable path+name) | `api/v1alpha1/ldapauthenginegroup_webhook.go` |
| RootCredentialConfig usage for API credentials | `api/v1alpha1/awsauthengineconfig_types.go` |
| filterPayloadToDesiredKeys | `api/v1alpha1/payload_filter.go` |
| toInterfaceArray helper | `api/v1alpha1/utils/vaultutils.go` |
| durationToSeconds helper | `api/v1alpha1/utils/vaultutils.go` |
| Documentation template | `docs/engine-doc-template.md` |

### Unit Test Requirements

**Config tests (`oktaauthengineconfig_test.go`):**
1. `TestOktaAuthEngineConfig_toMap` — verify all fields in snake_case, verify `api_token` from resolved internal field, verify `token_num_uses` is `json.Number`, verify list fields are `[]any`
2. `TestOktaAuthEngineConfig_IsEquivalentToDesiredState_Match` — construct Vault-read-shaped payload (without `api_token`, with `org_name`, `base_url`, `bypass_okta_mfa`, integer 0 values as `json.Number`, `token_policies` as `[]any`, `token_type` as "default"), verify returns `true`
3. `TestOktaAuthEngineConfig_IsEquivalentToDesiredState_Mismatch` — change `org_name` value, verify returns `false`
4. `TestOktaAuthEngineConfig_IsEquivalentToDesiredState_ExtraVaultFields` — add extra Vault-returned field, verify still returns `true` after filtering
5. `TestOktaAuthEngineConfig_IsEquivalentToDesiredState_APITokenStripping` — verify that a Vault payload without `api_token` still matches when desired state would have it (proves stripping works)

**Group tests (`oktaauthenginegroup_test.go`):**
1. `TestOktaAuthEngineGroup_toMap` — verify `policies` field matches spec
2. `TestOktaAuthEngineGroup_IsEquivalentToDesiredState_Match` — Vault-read fixture with `policies` as string, verify returns `true`
3. `TestOktaAuthEngineGroup_IsEquivalentToDesiredState_Mismatch` — change `policies`, verify returns `false`

### Anti-Patterns / DO NOT

- **DO NOT** create integration tests for these types — Okta is a cloud identity provider that cannot be installed in Kind (per "Skip it" rule)
- **DO NOT** create an OktaAuthEngineUser CRD in this story — user mappings are out of scope (only config and group CRDs are required by Epic 15 Story 15.3)
- **DO NOT** modify shared framework behavior (reconcile_skeleton.go, vaultresourcereconciler.go, etc.) — only add new files
- **DO NOT** derive expected payloads from `toMap()` in unit tests — construct independent Vault-read-shaped fixtures
- **DO NOT** forget to add new CRD YAML files to `config/crd/kustomization.yaml` after `make manifests`
- **DO NOT** use Go `int` or `float64` in unit test Vault payloads — use `json.Number` to match real Vault client behavior
- **DO NOT** include Enterprise-only fields in the CRD
- **DO NOT** include `api_token` in the Vault-read fixture for `IsEquivalentToDesiredState` tests — Vault never returns it
- **DO NOT** unconditionally overwrite `OktaCredentials.PasswordKey` in the webhook defaulter — only set when empty (lesson from 14.2 review)
- **DO NOT** confuse `OktaAuthEngineGroup` (group-to-policy mapping) with `LDAPAuthEngineGroup` (same pattern, different auth engine)
- **DO NOT** include `name` in the `toMap()` output for the group type if the Vault API uses the URL path for the name parameter — check whether Vault's group read response includes `name` in the payload. Looking at the LDAP group pattern: `LDAPAuthEngineGroup.toMap()` DOES include `name` in the payload. Follow the same pattern for Okta groups, but note: the Okta group read response only returns `{"policies": [...]}` — the group name is part of the URL path, not the body. So do **NOT** include `name` in `toMap()` for OktaAuthEngineGroup. This differs from LDAPAuthEngineGroup which includes it.

### Novelty Risk: LOW

Both CRD types follow well-established patterns from existing auth engine implementations. OktaAuthEngineConfig follows GCP/LDAP auth config patterns with credential resolution, and OktaAuthEngineGroup follows the LDAPAuthEngineGroup pattern exactly. No novel architectural patterns required.

### Project Structure Notes

- All new files follow existing naming conventions: lowercase type name for files
- New files go in established directories (`api/v1alpha1/`, `internal/controller/`, `test/`, `docs/auth-engines/`)
- Test fixture directory `test/oktaauthengine/` follows the existing pattern (`test/gcpauthengine/`, `test/ldapauthengine/`)
- No conflicts with existing code — purely additive

### References

- [Source: _bmad-output/planning-artifacts/epics.md#Epic-15, Story 15.3 — requirements and acceptance criteria]
- [Source: _bmad-output/project-context.md — all project rules, patterns, and conventions]
- [Source: api/v1alpha1/ldapauthengineconfig_types.go — LDAP auth config with credential resolution, IsDeletable=false]
- [Source: api/v1alpha1/ldapauthenginegroup_types.go — LDAP auth group with Name override, policies string]
- [Source: api/v1alpha1/gcpauthengineconfig_types.go — GCP auth config with RootCredentialConfig, GetPath auth/{path}/config]
- [Source: api/v1alpha1/awsauthengineconfig_types.go — AWS auth config with secret_key stripping in IsEquivalentToDesiredState]
- [Source: internal/controller/ldapauthengineconfig_controller.go — auth config controller with Secret/RandomSecret watches]
- [Source: internal/controller/ldapauthenginegroup_controller.go — simple auth group controller]
- [Source: api/v1alpha1/ldapauthenginegroup_webhook.go — auth group webhook pattern]
- [Source: docs/engine-doc-template.md — documentation template]
- [Source: Vault Okta Auth Method API — https://developer.hashicorp.com/vault/api-docs/auth/okta]
- [Source: _bmad-output/implementation-artifacts/14-2-aws-auth-engine-config-and-role-crds.md — most recent predecessor story]

## Code Review Record

### Review Model Used

(To be filled during code review — must differ from dev model)

### Review Findings

(To be filled during code review)

### Decisions Needed / Decisions Taken

- Design decision: OktaAuthEngineGroup `toMap()` should NOT include `name` — Vault's group read response only returns `{"policies": [...]}`, unlike LDAP group which includes `name` in the response. This must be verified during implementation by checking what Vault actually returns on `GET auth/{path}/groups/{name}`.
- Design decision: Use `RootCredentialConfig` for API token credential resolution, with `passwordKey` defaulting to `"api_token"` — this provides flexibility to source the token from Secret, VaultSecret, or RandomSecret while following existing patterns.
- Design decision: `OktaAuthEngineConfig.IsDeletable() = false` — consistent with all other auth engine configs (GCP, LDAP, JWT/OIDC, Azure, Kubernetes). No DELETE endpoint for `auth/{path}/config`.

### Fixes Applied

(To be filled during code review)

## Dev Agent Record

### Agent Model Used

(To be filled during implementation)

### Debug Log References

### Completion Notes List

### Change Log

### File List
