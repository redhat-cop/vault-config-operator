---
baseline_commit: ed5ac55b98d341d193b18cfa897f1321655805b7
---

# Story 15.2: GitHub Auth Engine — Config and Team/Org Mapping CRDs

Status: review

## Story

As an operator developer,
I want CRDs for GitHubAuthEngineConfig and GitHubAuthEngineTeamMap/UserMap,
So that Vault's GitHub auth method can be managed declaratively.

## Acceptance Criteria

1. **Given** a GitHubAuthEngineConfig CR is created with organization **When** the reconciler processes it **Then** the config is written to Vault at `auth/{path}/config` and ReconcileSuccessful=True

2. **Given** a GitHubAuthEngineTeamMap CR is created with team name and policies **When** the reconciler processes it **Then** the team mapping exists in Vault at `auth/{path}/map/teams/{name}` and ReconcileSuccessful=True

3. **Given** a GitHubAuthEngineUserMap CR is created with user name and policies **When** the reconciler processes it **Then** the user mapping exists in Vault at `auth/{path}/map/users/{name}` and ReconcileSuccessful=True

4. **Given** the GitHubAuthEngineConfig CR is deleted **When** the reconciler processes deletion **Then** the K8s object is removed but Vault config is NOT deleted (`IsDeletable=false` — config is tied to the mount lifecycle; no standalone DELETE endpoint)

5. **Given** a GitHubAuthEngineTeamMap CR is deleted **When** the reconciler processes deletion **Then** the team mapping is removed from Vault via `DELETE auth/{path}/map/teams/{name}` (`IsDeletable=true`)

6. **Given** a GitHubAuthEngineUserMap CR is deleted **When** the reconciler processes deletion **Then** the user mapping is removed from Vault via `DELETE auth/{path}/map/users/{name}` (`IsDeletable=true`)

7. **Given** any GitHub auth engine CR spec is updated **When** the reconciler processes the update **Then** the Vault resource reflects the updated values

8. **Given** any GitHub auth engine CR is created or updated **When** the webhook validates it **Then** `spec.path` immutability is enforced on updates, and `spec.name` immutability is enforced on updates (for team/user map types with Name field)

9. **Given** the CRD types are implemented **When** the story is marked done **Then** a documentation file exists at `docs/auth-engines/github.md` following `docs/engine-doc-template.md` (DNFR5)

## Tasks / Subtasks

- [x] Task 1: Create `GitHubAuthEngineConfig` type (AC: 1, 4, 7, 8)
  - [x] 1.1: Create `api/v1alpha1/githubauthengineconfig_types.go` — Spec with `Connection`, `Authentication`, `Path`, inline `GitHubAuthConfig` struct (organization, organizationID, baseURL, token params)
  - [x] 1.2: Implement `VaultObject` interface — `GetPath()` returns `auth/{path}/config`, `GetPayload()`, `IsEquivalentToDesiredState()`, `IsDeletable()=false`
  - [x] 1.3: Implement `ConditionsAware` interface — Status with `Conditions []metav1.Condition`
  - [x] 1.4: Implement `toMap()` on `GitHubAuthConfig` — convert to Vault API snake_case fields; conditionally include `organization_id` only when non-zero
  - [x] 1.5: Implement `IsEquivalentToDesiredState()` — standard `filterPayloadToDesiredKeys` (no write-only secrets to strip)

- [x] Task 2: Create `GitHubAuthEngineTeamMap` type (AC: 2, 5, 7, 8)
  - [x] 2.1: Create `api/v1alpha1/githubauthengineteammap_types.go` — Spec with `Connection`, `Authentication`, `Path`, `Name` (team name slug), `Policies` (comma-separated string)
  - [x] 2.2: Implement `VaultObject` interface — `GetPath()` returns `auth/{path}/map/teams/{name}`, `IsDeletable()=true`
  - [x] 2.3: Implement `ConditionsAware` interface
  - [x] 2.4: Implement `toMap()` — emit `{"value": policies}` matching Vault's write payload

- [x] Task 3: Create `GitHubAuthEngineUserMap` type (AC: 3, 6, 7, 8)
  - [x] 3.1: Create `api/v1alpha1/githubauthenginerusermap_types.go` — Spec with `Connection`, `Authentication`, `Path`, `Name` (GitHub username), `Policies` (comma-separated string)
  - [x] 3.2: Implement `VaultObject` interface — `GetPath()` returns `auth/{path}/map/users/{name}`, `IsDeletable()=true`
  - [x] 3.3: Implement `ConditionsAware` interface
  - [x] 3.4: Implement `toMap()` — emit `{"value": policies}` matching Vault's write payload

- [x] Task 4: Create webhooks (AC: 8)
  - [x] 4.1: Create `api/v1alpha1/githubauthengineconfig_webhook.go` — `admission.Defaulter[*GitHubAuthEngineConfig]`, `admission.Validator[*GitHubAuthEngineConfig]`, immutable `spec.path`
  - [x] 4.2: Create `api/v1alpha1/githubauthengineteammap_webhook.go` — `admission.Defaulter[*GitHubAuthEngineTeamMap]`, `admission.Validator[*GitHubAuthEngineTeamMap]`, immutable `spec.path`/`spec.name`
  - [x] 4.3: Create `api/v1alpha1/githubauthenginerusermap_webhook.go` — `admission.Defaulter[*GitHubAuthEngineUserMap]`, `admission.Validator[*GitHubAuthEngineUserMap]`, immutable `spec.path`/`spec.name`

- [x] Task 5: Create controllers (AC: 1, 2, 3, 4, 5, 6, 7)
  - [x] 5.1: Create `internal/controller/githubauthengineconfig_controller.go` — embed `ReconcilerBase`, standard VaultResource reconcile, simple `For()` with default periodic reconcile predicate (no credential watches needed)
  - [x] 5.2: Create `internal/controller/githubauthengineteammap_controller.go` — simple `For()` with default periodic reconcile predicate
  - [x] 5.3: Create `internal/controller/githubauthenginerusermap_controller.go` — simple `For()` with default periodic reconcile predicate

- [x] Task 6: Register in main.go (AC: 1, 2, 3)
  - [x] 6.1: Add controller registrations for all 3 reconcilers
  - [x] 6.2: Add webhook registrations inside `ENABLE_WEBHOOKS` guard for all 3 types

- [x] Task 7: Unit tests (AC: 1, 2, 3, 7, 8)
  - [x] 7.1: Create `api/v1alpha1/githubauthengineconfig_test.go` — test `toMap()` output, test `IsEquivalentToDesiredState()` with independently constructed Vault-read-shaped fixtures, negative tests
  - [x] 7.2: Create `api/v1alpha1/githubauthengineteammap_test.go` — test `toMap()`, test `IsEquivalentToDesiredState()` with Vault-read fixture (containing extra `key` field), negative tests
  - [x] 7.3: Create `api/v1alpha1/githubauthenginerusermap_test.go` — test `toMap()`, test `IsEquivalentToDesiredState()` with Vault-read fixture, negative tests

- [x] Task 8: Test fixtures (AC: all)
  - [x] 8.1: Create test YAML fixtures in `test/githubauthengine/` — config, team map, and user map CRs
  - [x] 8.2: Integration tests — SKIP (GitHub is a cloud service, falls under "Skip it" per project integration test philosophy)

- [x] Task 9: CRD registration and code generation (AC: all)
  - [x] 9.1: Run `make manifests generate fmt vet test`
  - [x] 9.2: Add 3 new CRD YAML files to `config/crd/kustomization.yaml` (CRD registration checklist)
  - [x] 9.3: Verify all existing tests still pass

- [x] Task 10: Documentation (AC: 9)
  - [x] 10.1: Create `docs/auth-engines/github.md` following `docs/engine-doc-template.md`
  - [x] 10.2: Update `docs/auth-engines/index.md` with link to new doc

## Dev Notes

### Integration Test Classification: SKIP

Per the project's Integration Test Infrastructure Philosophy:
> **Skip it** — For cloud providers and services that cannot be installed in Kind and are hard to mock (e.g., AWS, Azure, GCP, GitHub App, Quay), the integration test must not be run. These types rely on unit test coverage only.

GitHub is a cloud service that cannot be installed in Kind. No integration tests. Comprehensive unit tests for `toMap()`, `IsEquivalentToDesiredState()`, and webhook validation are the primary quality gate.

### Story Intelligence Chain — Previous Story Context

**Story 14.2 (AWS Auth Engine)** is the most recently completed auth engine CRD story:
- Established the pattern for multi-CRD auth engines (config + role/mapping)
- Key learnings: `filterPayloadToDesiredKeys` for drift detection, simple For() controllers without watches when no credentials involved
- Review findings documented credential defaults bugs — avoid unconditional overwrites in defaulter webhook

**Story 14.1 (AppRole Auth Engine)** and **Epic 14 retrospective:**
- Action item (open): "Propagate durationToSeconds() to older auth engine types" — not relevant here since GitHub config uses standard token params
- Action item (open): "Document strict webhook validation philosophy" — apply here: reject invalid combinations at admission
- Action item (open): "Address 14.2 credential defaults bug" — avoid same pattern (this story has no credential defaults to worry about)

**LDAPAuthEngineGroup** is the closest pattern for team/user mapping CRDs:
- Simple type with `Path`, `Name`, `Policies`
- `GetPath()` constructs `auth/{path}/groups/{name}`
- `toMap()` emits `name` and `policies`
- `IsDeletable()=true`
- No PrepareInternalValues, no credential watches

### Design Decision: Three CRDs

The Vault GitHub auth method has three distinct API surfaces:
1. `auth/{path}/config` — single config object per mount (1:1 mapping → `GitHubAuthEngineConfig`)
2. `auth/{path}/map/teams/{name}` — per-team policy mapping (N objects → `GitHubAuthEngineTeamMap`)
3. `auth/{path}/map/users/{name}` — per-user policy mapping (N objects → `GitHubAuthEngineUserMap`)

Each maps 1:1 to a Vault API path. The team and user map types are structurally identical (both write `{"value": "policies"}`) but are separate CRDs because they target different Vault API paths.

**Note on naming:** The epic title mentions "OrgMap" but the Vault API actually provides "map/teams" and "map/users" endpoints (not "map/orgs"). The `organization` is set at the config level, not as a separate mapping. CRD names follow the Vault API surface: `GitHubAuthEngineTeamMap` and `GitHubAuthEngineUserMap`.

### Vault API Paths

| Operation | Method | Path |
|-----------|--------|------|
| Write config | POST | `auth/{path}/config` |
| Read config | GET | `auth/{path}/config` |
| Write team mapping | POST | `auth/{path}/map/teams/{name}` |
| Read team mapping | GET | `auth/{path}/map/teams/{name}` |
| Delete team mapping | DELETE | `auth/{path}/map/teams/{name}` |
| Write user mapping | POST | `auth/{path}/map/users/{name}` |
| Read user mapping | GET | `auth/{path}/map/users/{name}` |
| Delete user mapping | DELETE | `auth/{path}/map/users/{name}` |

### GitHubAuthEngineConfig — Vault API Field Reference

**Write (`POST auth/{path}/config`) fields:**
- `organization` (string: required) — GitHub organization users must be part of
- `organization_id` (int: 0) — Org ID; Vault auto-fetches if not provided
- `base_url` (string: "") — API endpoint for GitHub Enterprise
- `token_ttl` (integer: 0 or string: "") — Incremental lifetime for generated tokens
- `token_max_ttl` (integer: 0 or string: "") — Maximum lifetime for generated tokens
- `token_policies` (array: [] or comma-delimited string: "") — Policies on generated tokens
- `token_bound_cidrs` (array: [] or comma-delimited string: "") — CIDR blocks for authentication
- `token_explicit_max_ttl` (integer: 0 or string: "") — Hard cap max TTL
- `token_no_default_policy` (bool: false) — Exclude default policy
- `token_num_uses` (integer: 0) — Max token uses (0 = unlimited)
- `token_period` (integer: 0 or string: "") — Max allowed period for periodic tokens
- `token_type` (string: "") — Token type (service, batch, default)

**Read (`GET auth/{path}/config`) sample response:**
```json
{
  "data": {
    "organization": "acme-org",
    "organization_id": 12345,
    "base_url": "",
    "token_ttl": 0,
    "token_max_ttl": 0,
    "token_policies": [],
    "token_bound_cidrs": [],
    "token_explicit_max_ttl": 0,
    "token_no_default_policy": false,
    "token_num_uses": 0,
    "token_period": 0,
    "token_type": ""
  }
}
```

**Critical observation:** No write-only fields. Standard `filterPayloadToDesiredKeys` is sufficient for `IsEquivalentToDesiredState`. However, `organization_id` should only be included in `toMap()` when non-zero — otherwise Vault auto-populates it and creates false drift if the user didn't set it.

**TTL fields:** Vault returns TTL fields as integer seconds on read. Use `durationToSeconds()` in `toMap()` for `tokenTTL`, `tokenMaxTTL`, `tokenExplicitMaxTTL`, `tokenPeriod` to emit `json.Number` matching Vault's read format.

### GitHubAuthEngineTeamMap / UserMap — Vault API Field Reference

**Team map write (`POST auth/{path}/map/teams/{team_name}`):**
- `value` (string) — Comma-separated list of policies to assign

**Team map read (`GET auth/{path}/map/teams/{team_name}`) response:**
```json
{
  "data": {
    "key": "dev",
    "value": "dev-policy"
  }
}
```

**User map write (`POST auth/{path}/map/users/{user_name}`):**
- `value` (string) — Comma-separated list of policies to assign

**User map read (`GET auth/{path}/map/users/{user_name}`) response:**
```json
{
  "data": {
    "key": "sethvargo",
    "value": "sethvargo-policy"
  }
}
```

**Critical:** The read response includes a `key` field (the name from the URL path) that is NOT part of the write payload. `toMap()` should only emit `{"value": "..."}`. The `filterPayloadToDesiredKeys` helper will naturally exclude `key` from comparison since it's not in `desiredState`.

### CRD Field Spec — GitHubAuthEngineConfig

```go
type GitHubAuthConfig struct {
    // Organization is the GitHub organization users must be part of.
    // +kubebuilder:validation:Required
    // +kubebuilder:validation:MinLength=1
    Organization string `json:"organization"`

    // OrganizationID is the numeric ID of the organization. Vault auto-fetches if not provided.
    // +kubebuilder:validation:Optional
    // +kubebuilder:validation:Minimum=0
    OrganizationID int64 `json:"organizationID,omitempty"`

    // BaseURL is the API endpoint for GitHub Enterprise. Leave empty for public GitHub.
    // +kubebuilder:validation:Optional
    BaseURL string `json:"baseURL,omitempty"`

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

    // TokenBoundCIDRs are CIDR blocks that restrict authentication.
    // +kubebuilder:validation:Optional
    // +listType=set
    TokenBoundCIDRs []string `json:"tokenBoundCIDRs,omitempty"`

    // TokenExplicitMaxTTL is the hard cap max TTL for tokens.
    // +kubebuilder:validation:Optional
    TokenExplicitMaxTTL string `json:"tokenExplicitMaxTTL,omitempty"`

    // TokenNoDefaultPolicy if true, omits the default policy from generated tokens.
    // +kubebuilder:validation:Optional
    TokenNoDefaultPolicy bool `json:"tokenNoDefaultPolicy,omitempty"`

    // TokenNumUses is the max number of times a token may be used (0 = unlimited).
    // +kubebuilder:validation:Optional
    // +kubebuilder:validation:Minimum=0
    TokenNumUses int64 `json:"tokenNumUses,omitempty"`

    // TokenPeriod is the maximum allowed period for periodic tokens.
    // +kubebuilder:validation:Optional
    TokenPeriod string `json:"tokenPeriod,omitempty"`

    // TokenType is the type of token to generate (service, batch, default).
    // +kubebuilder:validation:Optional
    // +kubebuilder:validation:Enum={"service","batch","default","default-service","default-batch",""}
    TokenType string `json:"tokenType,omitempty"`
}
```

### CRD Field Spec — GitHubAuthEngineTeamMap

```go
type GitHubAuthEngineTeamMapSpec struct {
    // Connection represents the information needed to connect to Vault.
    // +kubebuilder:validation:Optional
    Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

    // Authentication is the kube auth configuration to be used to execute this request.
    // +kubebuilder:validation:Required
    Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

    // Path at which the GitHub auth engine is mounted.
    // The final path in Vault will be {[spec.authentication.namespace]}/auth/{spec.path}/map/teams/{spec.name}.
    // +kubebuilder:validation:Required
    Path vaultutils.Path `json:"path,omitempty"`

    // Name is the GitHub team name in slugified format.
    // +kubebuilder:validation:Required
    // +kubebuilder:validation:MinLength=1
    // +kubebuilder:validation:Pattern=`[a-z0-9]([-a-z0-9]*[a-z0-9])?`
    Name string `json:"name"`

    // Policies is a comma-separated list of policies to assign to this team.
    // +kubebuilder:validation:Optional
    Policies string `json:"policies,omitempty"`
}
```

### CRD Field Spec — GitHubAuthEngineUserMap

```go
type GitHubAuthEngineUserMapSpec struct {
    // Connection represents the information needed to connect to Vault.
    // +kubebuilder:validation:Optional
    Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

    // Authentication is the kube auth configuration to be used to execute this request.
    // +kubebuilder:validation:Required
    Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

    // Path at which the GitHub auth engine is mounted.
    // The final path in Vault will be {[spec.authentication.namespace]}/auth/{spec.path}/map/users/{spec.name}.
    // +kubebuilder:validation:Required
    Path vaultutils.Path `json:"path,omitempty"`

    // Name is the GitHub username.
    // +kubebuilder:validation:Required
    // +kubebuilder:validation:MinLength=1
    Name string `json:"name"`

    // Policies is a comma-separated list of policies to assign to this user.
    // +kubebuilder:validation:Optional
    Policies string `json:"policies,omitempty"`
}
```

### `toMap()` Implementation Notes

**GitHubAuthConfig.toMap():**
```go
func (c *GitHubAuthConfig) toMap() map[string]any {
    payload := map[string]any{}
    payload["organization"] = c.Organization
    if c.OrganizationID != 0 {
        payload["organization_id"] = json.Number(strconv.FormatInt(c.OrganizationID, 10))
    }
    if c.BaseURL != "" {
        payload["base_url"] = c.BaseURL
    }
    if c.TokenTTL != "" {
        payload["token_ttl"] = durationToSeconds(c.TokenTTL)
    }
    if c.TokenMaxTTL != "" {
        payload["token_max_ttl"] = durationToSeconds(c.TokenMaxTTL)
    }
    if len(c.TokenPolicies) > 0 {
        payload["token_policies"] = toInterfaceArray(c.TokenPolicies)
    }
    if len(c.TokenBoundCIDRs) > 0 {
        payload["token_bound_cidrs"] = toInterfaceArray(c.TokenBoundCIDRs)
    }
    if c.TokenExplicitMaxTTL != "" {
        payload["token_explicit_max_ttl"] = durationToSeconds(c.TokenExplicitMaxTTL)
    }
    if c.TokenNoDefaultPolicy {
        payload["token_no_default_policy"] = c.TokenNoDefaultPolicy
    }
    if c.TokenNumUses != 0 {
        payload["token_num_uses"] = json.Number(strconv.FormatInt(c.TokenNumUses, 10))
    }
    if c.TokenPeriod != "" {
        payload["token_period"] = durationToSeconds(c.TokenPeriod)
    }
    if c.TokenType != "" {
        payload["token_type"] = c.TokenType
    }
    return payload
}
```

**Critical:** Use conditional inclusion (`if != zero`) for optional fields. Vault omits unset fields from read responses — if we unconditionally include them in desiredState, `filterPayloadToDesiredKeys` would miss them and cause false "equivalent" results or leave them out of comparison. This follows the `removeUnsetFields` pattern implicitly by never including them.

**GitHubAuthEngineTeamMap.toMap() / GitHubAuthEngineUserMap.toMap():**
```go
func (d *GitHubAuthEngineTeamMap) toMap() map[string]any {
    payload := map[string]any{}
    payload["value"] = d.Spec.Policies
    return payload
}
```

### IsEquivalentToDesiredState — Config

No write-only secrets to strip. Standard pattern:
```go
func (d *GitHubAuthEngineConfig) IsEquivalentToDesiredState(payload map[string]any) bool {
    desiredState := d.Spec.GitHubAuthConfig.toMap()
    return reflect.DeepEqual(desiredState, filterPayloadToDesiredKeys(desiredState, payload))
}
```

### IsEquivalentToDesiredState — TeamMap / UserMap

Standard pattern — `filterPayloadToDesiredKeys` will exclude the `key` field returned by Vault:
```go
func (d *GitHubAuthEngineTeamMap) IsEquivalentToDesiredState(payload map[string]any) bool {
    desiredState := d.GetPayload()
    return reflect.DeepEqual(desiredState, filterPayloadToDesiredKeys(desiredState, payload))
}
```

### Controller Pattern

All three controllers are simple — no credential watches needed (GitHub auth config has no secrets resolved from K8s):

```go
func (r *GitHubAuthEngineConfigReconciler) SetupWithManager(mgr ctrl.Manager) error {
    return ctrl.NewControllerManagedBy(mgr).
        For(&redhatcopv1alpha1.GitHubAuthEngineConfig{}, builder.WithPredicates(vaultresourcecontroller.NewDefaultPeriodicReconcilePredicate())).
        Complete(r)
}
```

Same pattern for TeamMap and UserMap controllers.

### Kubebuilder Markers Checklist

- All exported Spec fields: `+kubebuilder:validation:Required` or `+kubebuilder:validation:Optional`
- Config `Organization`: `+kubebuilder:validation:Required`, `+kubebuilder:validation:MinLength=1` (no omitempty — required field)
- Config `TokenType`: `+kubebuilder:validation:Enum={"service","batch","default","default-service","default-batch",""}` (empty string allowed)
- TeamMap/UserMap `Name`: `+kubebuilder:validation:Required`, `+kubebuilder:validation:MinLength=1`
- Root types: `+kubebuilder:object:root=true`, `+kubebuilder:subresource:status`
- Conditions field: `+patchMergeKey=type`, `+patchStrategy=merge`, `+listType=map`, `+listMapKey=type`
- List fields: `+listType=set` for TokenPolicies, TokenBoundCIDRs

### RBAC Markers

Config controller:
```go
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=githubauthengineconfigs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=githubauthengineconfigs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=githubauthengineconfigs/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=serviceaccounts/token,verbs=create
//+kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;patch
```

Team map controller:
```go
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=githubauthengineteammaps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=githubauthengineteammaps/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=githubauthengineteammaps/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=serviceaccounts/token,verbs=create
//+kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;patch
```

User map controller:
```go
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=githubauthengineusermaps,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=githubauthengineusermaps/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=githubauthengineusermaps/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=serviceaccounts/token,verbs=create
//+kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;patch
```

### File Structure

| File | Action | Description |
|------|--------|-------------|
| `api/v1alpha1/githubauthengineconfig_types.go` | NEW | Config CRD type, VaultObject, ConditionsAware, toMap |
| `api/v1alpha1/githubauthengineteammap_types.go` | NEW | TeamMap CRD type, VaultObject, ConditionsAware, toMap |
| `api/v1alpha1/githubauthenginerusermap_types.go` | NEW | UserMap CRD type, VaultObject, ConditionsAware, toMap |
| `api/v1alpha1/githubauthengineconfig_webhook.go` | NEW | Config webhook — defaulter, validator, immutable path |
| `api/v1alpha1/githubauthengineteammap_webhook.go` | NEW | TeamMap webhook — defaulter, validator, immutable path/name |
| `api/v1alpha1/githubauthenginerusermap_webhook.go` | NEW | UserMap webhook — defaulter, validator, immutable path/name |
| `api/v1alpha1/githubauthengineconfig_test.go` | NEW | Unit tests for config toMap, IsEquivalentToDesiredState |
| `api/v1alpha1/githubauthengineteammap_test.go` | NEW | Unit tests for team map toMap, IsEquivalentToDesiredState |
| `api/v1alpha1/githubauthenginerusermap_test.go` | NEW | Unit tests for user map toMap, IsEquivalentToDesiredState |
| `internal/controller/githubauthengineconfig_controller.go` | NEW | Config reconciler — simple VaultResource |
| `internal/controller/githubauthengineteammap_controller.go` | NEW | TeamMap reconciler — simple VaultResource |
| `internal/controller/githubauthenginerusermap_controller.go` | NEW | UserMap reconciler — simple VaultResource |
| `cmd/main.go` | UPDATE | Register 3 controllers + 3 webhooks |
| `config/crd/kustomization.yaml` | UPDATE | Add 3 new CRD YAML files to resources list |
| `test/githubauthengine/` | NEW | Test YAML fixtures for all 3 types |
| `docs/auth-engines/github.md` | NEW | Engine documentation per DNFR5 |
| `docs/auth-engines/index.md` | UPDATE | Add link to github.md |

### Files Being Modified — Current State

**`cmd/main.go`**: Currently registers ~45+ controllers and webhooks (including Epic 14 additions). New registrations follow the exact same pattern:
- Controller: `(&controller.GitHubAuthEngineConfigReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "GitHubAuthEngineConfig")}).SetupWithManager(mgr)`
- Webhook (inside ENABLE_WEBHOOKS): `(&redhatcopv1alpha1.GitHubAuthEngineConfig{}).SetupWebhookWithManager(mgr)`

No existing behavior is changed — purely additive.

**`config/crd/kustomization.yaml`**: Add 3 new CRD YAML files to the `resources` list. Required for Helm chart build.

**`docs/auth-engines/index.md`**: Add a row to the Supported Auth Engines table for GitHub.

### Existing Patterns to Follow (Reference Files)

| Pattern | Reference File |
|---------|---------------|
| Auth config with IsDeletable=false, no credentials | `api/v1alpha1/gcpauthengineconfig_types.go` (simplify: no GCPCredentials) |
| Auth config GetPath (auth/{path}/config) | `api/v1alpha1/gcpauthengineconfig_types.go` |
| Simple mapping type (policies, name, path) | `api/v1alpha1/ldapauthenginegroup_types.go` |
| Mapping GetPath (auth/{path}/groups/{name}) | `api/v1alpha1/ldapauthenginegroup_types.go` |
| Standard filterPayloadToDesiredKeys | `api/v1alpha1/ldapauthenginegroup_types.go` |
| Simple controller (no credential watches) | `internal/controller/gcpauthenginerole_controller.go` |
| Auth config webhook (immutable path) | `api/v1alpha1/gcpauthengineconfig_webhook.go` |
| Mapping webhook (immutable path+name) | `api/v1alpha1/ldapauthenginegroup_webhook.go` |
| durationToSeconds for TTL fields | `api/v1alpha1/utils/vaultutils.go` |
| toInterfaceArray for list fields | `api/v1alpha1/utils/vaultutils.go` |
| json.Number for numeric Vault-facing fields | `api/v1alpha1/awsauthengineconfig_types.go` |
| Documentation template | `docs/engine-doc-template.md` |

### Unit Test Requirements

**Config tests (`githubauthengineconfig_test.go`):**
1. `TestGitHubAuthEngineConfig_toMap_AllFields` — verify organization, organization_id (json.Number), base_url, all token_* fields in snake_case with correct types (json.Number for TTLs, []any for arrays)
2. `TestGitHubAuthEngineConfig_toMap_MinimalFields` — verify only `organization` is emitted when all optional fields are zero-value
3. `TestGitHubAuthEngineConfig_IsEquivalentToDesiredState_Match` — Vault-read-shaped payload matching desiredState, verify returns `true`
4. `TestGitHubAuthEngineConfig_IsEquivalentToDesiredState_Mismatch` — change `organization`, verify returns `false`
5. `TestGitHubAuthEngineConfig_IsEquivalentToDesiredState_ExtraVaultFields` — add extra Vault-returned field (e.g., `organization_id` when not in spec), verify still returns `true` after filtering

**Team map tests (`githubauthengineteammap_test.go`):**
1. `TestGitHubAuthEngineTeamMap_toMap` — verify `{"value": "policy1,policy2"}`
2. `TestGitHubAuthEngineTeamMap_IsEquivalentToDesiredState_Match` — Vault-read fixture `{"key": "dev", "value": "dev-policy"}`, verify returns `true` (key filtered out)
3. `TestGitHubAuthEngineTeamMap_IsEquivalentToDesiredState_Mismatch` — change `value`, verify returns `false`
4. `TestGitHubAuthEngineTeamMap_GetPath` — verify path composition `auth/{path}/map/teams/{name}`

**User map tests (`githubauthenginerusermap_test.go`):**
1. `TestGitHubAuthEngineUserMap_toMap` — verify `{"value": "sethvargo-policy"}`
2. `TestGitHubAuthEngineUserMap_IsEquivalentToDesiredState_Match` — Vault-read fixture `{"key": "sethvargo", "value": "sethvargo-policy"}`, verify returns `true`
3. `TestGitHubAuthEngineUserMap_IsEquivalentToDesiredState_Mismatch` — change `value`, verify returns `false`
4. `TestGitHubAuthEngineUserMap_GetPath` — verify path composition `auth/{path}/map/users/{name}`

### Anti-Patterns / DO NOT

- **DO NOT** create integration tests — GitHub is a cloud service (per "Skip it" rule)
- **DO NOT** confuse `GitHubAuthEngineConfig` (auth engine, this story) with `GitHubSecretEngineConfig` (secret engine, existing) — different Vault API paths (`auth/` vs `secrets/`), different purposes, different Go files
- **DO NOT** include the `VAULT_AUTH_CONFIG_GITHUB_TOKEN` env var as a CRD field — it's a Vault server-side env var, not an operator-managed credential
- **DO NOT** include `organization_id` unconditionally in toMap() — only when non-zero. Vault auto-populates it; including zero would prevent drift detection from working correctly
- **DO NOT** modify shared framework behavior (reconcile_skeleton.go, vaultresourcereconciler.go, etc.) — only add new files
- **DO NOT** derive expected payloads from `toMap()` in unit tests — construct independent Vault-read-shaped fixtures
- **DO NOT** forget to add new CRD YAML files to `config/crd/kustomization.yaml` after `make manifests`
- **DO NOT** use Go `int` or `float64` in unit test Vault payloads — use `json.Number` to match real Vault client behavior
- **DO NOT** add a `deprecated policies` field — unlike some other auth methods, the GitHub config's `policies` parameter is fully deprecated in favor of `token_policies`; support only `token_policies`
- **DO NOT** include the `key` field in team/user map `toMap()` output — it's only returned by Vault on read and is derived from the URL path

### Novelty Risk: LOW

All three CRD types follow well-established patterns:
- Config: simpler than GCPAuthEngineConfig (no credentials), standard token parameters
- Team/User maps: nearly identical to LDAPAuthEngineGroup (name + policies string)
- No novel architectural patterns. No multi-mode validation. No credential resolution.

### Project Structure Notes

- All new files follow existing naming conventions: lowercase type name
- New files go in established directories (`api/v1alpha1/`, `internal/controller/`, `test/`, `docs/auth-engines/`)
- Test fixture directory `test/githubauthengine/` follows existing pattern (`test/gcpauthengine/`, `test/awsauthengine/`)
- No conflicts with existing code — purely additive
- Note: `api/v1alpha1/githubauthengineconfig_types.go` is for the **auth** engine config; `api/v1alpha1/githubsecretengineconfig_types.go` already exists for the **secret** engine config — these are different files for different purposes

### References

- [Source: _bmad-output/planning-artifacts/epics.md#Epic-15, Story 15.2 — requirements and acceptance criteria]
- [Source: _bmad-output/project-context.md — all project rules, patterns, and conventions]
- [Source: api/v1alpha1/ldapauthenginegroup_types.go — LDAP group mapping with policies, IsDeletable=true]
- [Source: api/v1alpha1/gcpauthengineconfig_types.go — GCP auth config with IsDeletable=false, GetPath pattern]
- [Source: api/v1alpha1/githubsecretengineconfig_types.go — existing GitHub SECRET engine (different from auth)]
- [Source: internal/controller/gcpauthenginerole_controller.go — simple controller with no watches]
- [Source: api/v1alpha1/gcpauthengineconfig_webhook.go — auth config webhook pattern]
- [Source: docs/engine-doc-template.md — documentation template]
- [Source: Vault GitHub Auth Method API — https://developer.hashicorp.com/vault/api-docs/auth/github]
- [Source: _bmad-output/implementation-artifacts/14-2-aws-auth-engine-config-and-role-crds.md — most recent predecessor auth engine story]

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6

### Debug Log References

None — clean implementation with no blocking issues.

### Completion Notes List

- Implemented 3 CRDs: GitHubAuthEngineConfig, GitHubAuthEngineTeamMap, GitHubAuthEngineUserMap
- Config type uses IsDeletable=false (tied to mount lifecycle), maps use IsDeletable=true
- toMap() uses conditional field inclusion to avoid false drift with Vault auto-populated fields (e.g., organization_id)
- TTL fields emit json.Number via durationToSeconds() matching Vault read format
- Team/User maps emit only {"value": policies} — the "key" field in Vault reads is correctly filtered by filterPayloadToDesiredKeys
- Webhooks enforce immutability on spec.path (all types) and spec.name (team/user map types)
- All unit tests pass including toMap output verification, IsEquivalentToDesiredState positive/negative tests, GetPath composition
- Integration tests skipped per project policy (GitHub is cloud service — "Skip it" classification)
- Documentation created following engine-doc-template.md

### Change Log

- 2026-08-17: Implemented story 15.2 — GitHub Auth Engine Config and Team/User Map CRDs (all 10 tasks)

### File List

- api/v1alpha1/githubauthengineconfig_types.go (NEW)
- api/v1alpha1/githubauthengineteammap_types.go (NEW)
- api/v1alpha1/githubauthenginerusermap_types.go (NEW)
- api/v1alpha1/githubauthengineconfig_webhook.go (NEW)
- api/v1alpha1/githubauthengineteammap_webhook.go (NEW)
- api/v1alpha1/githubauthenginerusermap_webhook.go (NEW)
- api/v1alpha1/githubauthengineconfig_test.go (NEW)
- api/v1alpha1/githubauthengineteammap_test.go (NEW)
- api/v1alpha1/githubauthenginerusermap_test.go (NEW)
- api/v1alpha1/zz_generated.deepcopy.go (MODIFIED — auto-generated)
- internal/controller/githubauthengineconfig_controller.go (NEW)
- internal/controller/githubauthengineteammap_controller.go (NEW)
- internal/controller/githubauthenginerusermap_controller.go (NEW)
- cmd/main.go (MODIFIED — added 3 controller + 3 webhook registrations)
- config/crd/bases/redhatcop.redhat.io_githubauthengineconfigs.yaml (NEW — auto-generated)
- config/crd/bases/redhatcop.redhat.io_githubauthengineteammaps.yaml (NEW — auto-generated)
- config/crd/bases/redhatcop.redhat.io_githubauthengineusermaps.yaml (NEW — auto-generated)
- config/crd/kustomization.yaml (MODIFIED — added 3 CRD resources)
- config/rbac/role.yaml (MODIFIED — auto-generated RBAC rules)
- test/githubauthengine/githubauthengineconfig.yaml (NEW)
- test/githubauthengine/githubauthengineteammap.yaml (NEW)
- test/githubauthengine/githubauthenginerusermap.yaml (NEW)
- docs/auth-engines/github.md (NEW)
- docs/auth-engines/index.md (MODIFIED — added GitHub row)
- _bmad-output/implementation-artifacts/15-2-github-auth-engine-config-and-team-org-mapping-crds.md (MODIFIED — status updates)
