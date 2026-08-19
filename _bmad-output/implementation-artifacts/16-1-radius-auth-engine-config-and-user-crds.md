# Story 16.1: RADIUS Auth Engine — Config and User CRDs

Status: ready-for-dev

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As an operator developer,
I want CRDs for RADIUSAuthEngineConfig and RADIUSAuthEngineUser,
so that Vault's RADIUS auth method can be managed declaratively.

## Acceptance Criteria

1. **Given** a RADIUSAuthEngineConfig CR is created with RADIUS host, shared secret (from K8s Secret), and optional port/timeout settings **When** the reconciler processes it **Then** the config is written to Vault at `auth/{path}/config` and ReconcileSuccessful=True

2. **Given** a RADIUSAuthEngineUser CR is created with username and policies **When** the reconciler processes it **Then** the user mapping exists in Vault at `auth/{path}/users/{name}` and ReconcileSuccessful=True

3. **Given** the RADIUSAuthEngineConfig CR is deleted **When** the reconciler processes deletion **Then** the K8s object is removed but Vault config is NOT deleted (`IsDeletable=false` — auth engine configs are not deletable; the mount owns the config lifecycle)

4. **Given** the RADIUSAuthEngineUser CR is deleted **When** the reconciler processes deletion **Then** the user mapping is removed from Vault via `DELETE auth/{path}/users/{name}` (`IsDeletable=true`)

5. **Given** any RADIUS auth engine CR spec is updated **When** the reconciler processes the update **Then** the Vault resource reflects the updated values

6. **Given** any RADIUS auth engine CR is created or updated **When** the webhook validates it **Then** `spec.path` immutability is enforced on updates, and `spec.name` immutability is enforced on updates for RADIUSAuthEngineUser

7. **Given** the RADIUSAuthEngineConfig CR has a `RADIUSCredentials` field referencing a K8s Secret **When** the Secret is updated **Then** the reconciler re-reads the shared secret and writes the updated config to Vault

8. **Given** the CRD types are implemented **When** the story is marked done **Then** a documentation file exists at `docs/auth-engines/radius.md` following `docs/engine-doc-template.md` (DNFR5)

## Tasks / Subtasks

- [ ] Task 1: Create `RADIUSAuthEngineConfig` type (AC: 1, 3, 5, 6, 7)
  - [ ] 1.1: Create `api/v1alpha1/radiusauthengineconfig_types.go` — Spec with `Connection`, `Authentication`, `Path`, inline `RADIUSAuthConfig` struct, `RADIUSCredentials` (`RootCredentialConfig` with passwordKey="secret")
  - [ ] 1.2: Implement `VaultObject` interface — `GetPath()` returns `auth/{path}/config`, `GetPayload()`, `IsEquivalentToDesiredState()`, `PrepareInternalValues()`, `IsDeletable()=false`
  - [ ] 1.3: Implement `ConditionsAware` interface — Status with `Conditions []metav1.Condition`
  - [ ] 1.4: Implement `setInternalCredentials()` — resolve RADIUS shared secret from K8s Secret, VaultSecret, or RandomSecret
  - [ ] 1.5: Implement `toMap()` on `RADIUSAuthConfig` — convert to Vault API snake_case fields, include resolved `secret` from internal field, use `json.Number` for integer fields
  - [ ] 1.6: Implement `IsEquivalentToDesiredState()` — must delete `secret` from desired state (Vault never returns it on read), then `filterPayloadToDesiredKeys` + `reflect.DeepEqual`

- [ ] Task 2: Create `RADIUSAuthEngineUser` type (AC: 2, 4, 5, 6)
  - [ ] 2.1: Create `api/v1alpha1/radiusauthengineuser_types.go` — Spec with `Connection`, `Authentication`, `Path`, `Name` (username override), `Policies` (comma-separated string)
  - [ ] 2.2: Implement `VaultObject` interface — `GetPath()` returns `auth/{path}/users/{name}`, `IsDeletable()=true`, no `PrepareInternalValues` needed
  - [ ] 2.3: Implement `ConditionsAware` interface
  - [ ] 2.4: Implement `toMap()` — emit `policies` field
  - [ ] 2.5: Implement `IsEquivalentToDesiredState()` — standard `filterPayloadToDesiredKeys`

- [ ] Task 3: Create webhooks (AC: 6)
  - [ ] 3.1: Create `api/v1alpha1/radiusauthengineconfig_webhook.go` — `admission.Defaulter[*RADIUSAuthEngineConfig]`, `admission.Validator[*RADIUSAuthEngineConfig]`, immutable `spec.path`, credential validation via `ValidateCredentialSource()`
  - [ ] 3.2: Create `api/v1alpha1/radiusauthengineuser_webhook.go` — `admission.Defaulter[*RADIUSAuthEngineUser]`, `admission.Validator[*RADIUSAuthEngineUser]`, immutable `spec.path`/`spec.name`

- [ ] Task 4: Create controllers (AC: 1, 2, 3, 4, 5, 7)
  - [ ] 4.1: Create `internal/controller/radiusauthengineconfig_controller.go` — embed `ReconcilerBase`, standard VaultResource reconcile, watches on `corev1.Secret` and `RandomSecret` for shared secret rotation
  - [ ] 4.2: Create `internal/controller/radiusauthengineuser_controller.go` — simple `For()` with default periodic reconcile predicate (no watches needed)

- [ ] Task 5: Register in main.go (AC: 1, 2)
  - [ ] 5.1: Add controller registrations for both reconcilers
  - [ ] 5.2: Add webhook registrations inside `ENABLE_WEBHOOKS` guard for both types

- [ ] Task 6: Unit tests (AC: 1, 2, 5, 6)
  - [ ] 6.1: Create `api/v1alpha1/radiusauthengineconfig_test.go` — test `toMap()`, test `IsEquivalentToDesiredState()` with independently constructed Vault-read-shaped fixtures (secret stripped), negative tests
  - [ ] 6.2: Create `api/v1alpha1/radiusauthengineuser_test.go` — test `toMap()`, test `IsEquivalentToDesiredState()` with Vault-read fixture, negative tests

- [ ] Task 7: Test fixtures (AC: all)
  - [ ] 7.1: Create test YAML fixtures in `test/radiusauthengine/` — config and user CRs
  - [ ] 7.2: Integration tests — SKIP (RADIUS server cannot be trivially installed in Kind; see Integration Test Classification below)

- [ ] Task 8: CRD registration and code generation (AC: all)
  - [ ] 8.1: Run `make manifests generate fmt vet test`
  - [ ] 8.2: Add 2 new CRD YAML files to `config/crd/kustomization.yaml` (CRD registration checklist)
  - [ ] 8.3: Verify all existing tests still pass

- [ ] Task 9: Documentation (AC: 8)
  - [ ] 9.1: Create `docs/auth-engines/radius.md` following `docs/engine-doc-template.md`
  - [ ] 9.2: Update `docs/auth-engines/index.md` — add RADIUS row to Supported Auth Engines table

## Dev Notes

### Integration Test Classification: SKIP

Per the project's Integration Test Infrastructure Philosophy:
> **Skip it** — For cloud providers and services that cannot be installed in Kind and are hard to mock (e.g., AWS, Azure, GCP, GitHub App, Quay), the integration test must not be run. These types rely on unit test coverage only.

RADIUS requires an external RADIUS server (e.g., FreeRADIUS). While FreeRADIUS *can* technically run in a container, it requires PAP-scheme configuration, user database seeding, and UDP connectivity from Vault to the RADIUS container — making it fragile and non-trivial to set up in Kind compared to Vault-native methods like Userpass. The RADIUS server is the *auth backend* (Vault forwards login requests to it), but the operator only manages config and user-policy mappings — it never performs RADIUS authentication itself. The config write and user mapping are Vault-native API calls that don't require RADIUS server connectivity at all.

**Decision: SKIP integration tests.** Comprehensive unit tests for `toMap()`, `IsEquivalentToDesiredState()`, and webhook validation are the primary quality gate. The config and user CRUD paths are standard Vault API calls identical in shape to Okta (config with credential stripping) and LDAP/Okta (user/group policy mapping).

### Novelty Risk: MEDIUM

RADIUS config includes RADIUS-specific integer fields (`port`, `dial_timeout`, `read_timeout`, `nas_port`) that are uncommon in other auth engine configs. The `secret` field follows the established write-only credential pattern but with a different key name ("secret" instead of "api_token" or "password"). The user CRD is a trivial copy of the Okta group pattern (`policies` string). Overall the architectural pattern is well-established but the config has more integer fields than typical, requiring care with `json.Number` emission.

### Story Intelligence Chain — Previous Story Context

**Story 15.3 (Okta Auth Engine Config + Group)** is the closest structural analog for this story:
- Two-CRD pattern: Config (`IsDeletable=false`, credential resolution, `GetPath()` → `auth/{path}/config`) + mapping type (`IsDeletable=true`, simple policies string)
- OktaAuthEngineConfig uses `RootCredentialConfig` for API token with write-only stripping in `IsEquivalentToDesiredState`
- OktaAuthEngineGroup follows LDAPAuthEngineGroup exactly — policies as string, `IsDeletable=true`
- Okta config webhook defaults `OktaCredentials.PasswordKey` to `"api_token"` only when empty
- Integration tests skipped (cloud provider)

**Story 15.1 (Userpass Auth Engine User)** established:
- Credential resolution for auth engine user types via `RootCredentialConfig`
- Write-only field stripping (`password`) in `IsEquivalentToDesiredState`
- Controller with Secret/RandomSecret watches for credential rotation
- Kind lifecycle integration tests (Vault-native type)

**Epic 15 retrospective action items (all completed):**
- `durationToSeconds()` propagation — done for all auth engine types
- Strict webhook validation — documented in project-context.md
- Credential key defaulting — Default() inspects admission request, remaps only omitted keys

### Vault API Paths

| Operation | Method | Path |
|-----------|--------|------|
| Write config | POST | `auth/{path}/config` |
| Read config | GET | `auth/{path}/config` |
| Register user | POST | `auth/{path}/users/{username}` |
| Read user | GET | `auth/{path}/users/{username}` |
| Delete user | DELETE | `auth/{path}/users/{username}` |
| List users | LIST | `auth/{path}/users` |

### RADIUSAuthEngineConfig — Vault API Field Reference

**Write (`POST auth/{path}/config`) fields:**

| Vault API Field | Type | Default | Description |
|-----------------|------|---------|-------------|
| `host` | string | "" | RADIUS server hostname or IP. Required. |
| `port` | integer | 1812 | UDP port the RADIUS server listens on |
| `secret` | string | "" | RADIUS shared secret. Write-only — never returned on read. Required. |
| `unregistered_user_policies` | string | "" | Comma-separated policies for users who authenticate via RADIUS but have no explicit mapping |
| `dial_timeout` | integer | 10 | Seconds to wait for backend connection |
| `read_timeout` | integer | 10 | Seconds before response times out |
| `nas_port` | integer | 10 | NAS-Port attribute of the RADIUS request |
| `token_ttl` | integer/string | 0 | Incremental lifetime for generated tokens |
| `token_max_ttl` | integer/string | 0 | Maximum lifetime for generated tokens |
| `token_policies` | array/string | [] | Token policies |
| `token_bound_cidrs` | array/string | [] | CIDR blocks for token authentication |
| `token_explicit_max_ttl` | integer/string | 0 | Hard cap max TTL |
| `token_no_default_policy` | bool | false | Exclude default policy |
| `token_num_uses` | integer | 0 | Max token uses (0=unlimited) |
| `token_period` | integer/string | 0 | Max allowed period for periodic tokens |
| `token_type` | string | "" | Token type: service, batch, default, default-service, default-batch |

**Excluded fields:**
- `policies` — Deprecated; use `token_policies` instead
- `alias_metadata` — Optional advanced feature; can be added in a future iteration

**Read (`GET auth/{path}/config`) expected response:**
```json
{
  "data": {
    "host": "radius.example.com",
    "port": 1812,
    "unregistered_user_policies": "",
    "dial_timeout": 10,
    "read_timeout": 10,
    "nas_port": 10,
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

**Critical observations for `IsEquivalentToDesiredState`:**
- `secret` is **write-only** — Vault NEVER returns it on read. Must `delete(desiredState, "secret")` before comparison. Follow `OktaAuthEngineConfig` pattern (deletes `api_token`).
- Vault returns integer fields (`port`, `dial_timeout`, `read_timeout`, `nas_port`, `token_num_uses`) as `json.Number` — `toMap()` must emit all integers as `json.Number`
- Vault returns duration fields (`token_ttl`, `token_max_ttl`, `token_explicit_max_ttl`, `token_period`) as integer seconds — use `durationToSeconds()` in `toMap()`
- Vault returns list fields (`token_policies`, `token_bound_cidrs`) as `[]any` — use `toInterfaceArray()` in `toMap()`

### Critical: `IsEquivalentToDesiredState` for Config — Secret Stripping

Vault never returns `secret` on read. The implementation must:
1. Build `desiredState` from `RADIUSAuthConfig.toMap()`
2. `delete(desiredState, "secret")` — remove before comparison
3. Use `filterPayloadToDesiredKeys(desiredState, payload)` → `reflect.DeepEqual`

Follow the established pattern from `OktaAuthEngineConfig.IsEquivalentToDesiredState()` (deletes `api_token`).

### RADIUSAuthEngineConfig — Credential Resolution (Shared Secret)

The RADIUS shared secret must be resolved from one of three sources using `RootCredentialConfig`:
- **K8s Secret**: passwordKey defaults to "secret"
- **VaultSecret**: same key mapping from a Vault KV path
- **RandomSecret**: retrieve from RandomSecret's Vault path

Pattern: Use `RootCredentialConfig` with custom default key (`passwordKey: "secret"`). The usernameKey is not used for RADIUS config (there is no username credential — `host` is specified directly in the spec). Store the resolved shared secret in an unexported field (`retrievedSecret` with `json:"-"`). Follow `OktaAuthEngineConfig.setInternalCredentials()` exactly — adapted for a single credential (shared secret only).

**Webhook Defaulter:** Set `RADIUSCredentials.PasswordKey` to `"secret"` when empty. Do NOT unconditionally overwrite — only set when the field is empty string (lesson from 14.2 / Epic 15 retro).

### RADIUSAuthEngineConfig — Controller Pattern

Standard VaultResource reconcile (NOT always-write) because the read endpoint returns enough fields for meaningful drift detection (everything except `secret`). The `IsEquivalentToDesiredState` strips `secret` to avoid false drift.

The controller MUST include Secret and RandomSecret watches for shared secret rotation detection, following `OktaAuthEngineConfigReconciler.SetupWithManager()`.

### RADIUSAuthEngineUser — Vault API Field Reference

**Write (`POST auth/{path}/users/{username}`) fields:**
- `username` (string: required) — Username (from URL path, not body)
- `policies` (string: "") — Comma-separated list of policies

**Read (`GET auth/{path}/users/{username}`) response:**
```json
{
  "data": {
    "policies": "default,dev"
  }
}
```

No write-only fields. Standard `filterPayloadToDesiredKeys` is sufficient.

### RADIUSAuthEngineUser — Pattern

Follow `OktaAuthEngineGroup` exactly:
- Spec: `Connection`, `Authentication`, `Path`, `Name` (username override), `Policies` (string — comma-separated list)
- `GetPath()`: `auth/{path}/users/{name}` using `vaultutils.CleansePath`
- `IsDeletable()`: `true`
- `toMap()`: emit `policies` field (as string, NOT array — matches Okta/LDAP group pattern)
- `IsEquivalentToDesiredState()`: standard `filterPayloadToDesiredKeys`
- Simple controller: `For()` with default periodic reconcile predicate, no watches

**Important:** The Vault API accepts `policies` as a comma-separated string. The Okta group / LDAP group pattern uses a single string. Follow the same pattern. Do NOT include `username` in `toMap()` — Vault's user API uses the URL path for the username (the read response only returns `{"policies": "..."}`, not the username).

### CRD Field Spec — RADIUSAuthConfig

```go
type RADIUSAuthConfig struct {
    // Host is the RADIUS server to connect to (e.g. "radius.myorg.com", "127.0.0.1").
    // +kubebuilder:validation:Required
    Host string `json:"host"`

    // Port is the UDP port where the RADIUS server is listening. Defaults to 1812.
    // +kubebuilder:validation:Optional
    // +kubebuilder:validation:Minimum=1
    // +kubebuilder:validation:Maximum=65535
    Port int `json:"port,omitempty"`

    // UnregisteredUserPolicies is a comma-separated list of policies to grant to
    // users who authenticate via RADIUS but have no explicit user mapping.
    // +kubebuilder:validation:Optional
    UnregisteredUserPolicies string `json:"unregisteredUserPolicies,omitempty"`

    // DialTimeout is the number of seconds to wait for a backend connection before timing out.
    // +kubebuilder:validation:Optional
    // +kubebuilder:validation:Minimum=0
    DialTimeout int `json:"dialTimeout,omitempty"`

    // ReadTimeout is the number of seconds before a response times out.
    // +kubebuilder:validation:Optional
    // +kubebuilder:validation:Minimum=0
    ReadTimeout int `json:"readTimeout,omitempty"`

    // NASPort is the NAS-Port attribute of the RADIUS request.
    // +kubebuilder:validation:Optional
    // +kubebuilder:validation:Minimum=0
    NASPort int `json:"nasPort,omitempty"`

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
    TokenNumUses int `json:"tokenNumUses,omitempty"`

    // TokenPeriod is the maximum allowed period for periodic tokens.
    // +kubebuilder:validation:Optional
    TokenPeriod string `json:"tokenPeriod,omitempty"`

    // TokenType is the type of token to generate.
    // +kubebuilder:validation:Optional
    // +kubebuilder:validation:Enum={"service","batch","default","default-service","default-batch",""}
    TokenType string `json:"tokenType,omitempty"`

    retrievedSecret string `json:"-"`
}
```

### CRD Field Spec — RADIUSAuthEngineConfigSpec (wrapper)

```go
type RADIUSAuthEngineConfigSpec struct {
    // +kubebuilder:validation:Optional
    Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

    // +kubebuilder:validation:Required
    Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

    // Path at which the RADIUS auth engine is mounted.
    // +kubebuilder:validation:Required
    Path vaultutils.Path `json:"path,omitempty"`

    RADIUSAuthConfig `json:",inline"`

    // RADIUSCredentials is used to provide the RADIUS shared secret.
    // The shared secret can be sourced from a Kubernetes Secret, VaultSecret, or RandomSecret.
    // +kubebuilder:validation:Optional
    RADIUSCredentials vaultutils.RootCredentialConfig `json:"radiusCredentials,omitempty"`
}
```

### CRD Field Spec — RADIUSAuthEngineUser

Follow OktaAuthEngineGroup pattern exactly:

```go
type RADIUSAuthEngineUserSpec struct {
    // +kubebuilder:validation:Optional
    Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

    // +kubebuilder:validation:Required
    Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

    // Path at which the RADIUS auth engine is mounted.
    // +kubebuilder:validation:Required
    Path vaultutils.Path `json:"path,omitempty"`

    // Name of the RADIUS user. If specified, takes precedence over metadata.name.
    // +kubebuilder:validation:Optional
    Name string `json:"name,omitempty"`

    // Policies is a comma-separated list of policies associated with this user.
    // +kubebuilder:validation:Optional
    Policies string `json:"policies,omitempty"`
}
```

### `toMap()` Implementation Notes

**RADIUSAuthConfig.toMap():**
```go
func (c *RADIUSAuthConfig) toMap() map[string]any {
    payload := map[string]any{}
    payload["host"] = c.Host
    payload["secret"] = c.retrievedSecret
    payload["port"] = json.Number(strconv.Itoa(c.Port))
    payload["unregistered_user_policies"] = c.UnregisteredUserPolicies
    payload["dial_timeout"] = json.Number(strconv.Itoa(c.DialTimeout))
    payload["read_timeout"] = json.Number(strconv.Itoa(c.ReadTimeout))
    payload["nas_port"] = json.Number(strconv.Itoa(c.NASPort))
    payload["token_ttl"] = durationToSeconds(c.TokenTTL)
    payload["token_max_ttl"] = durationToSeconds(c.TokenMaxTTL)
    payload["token_policies"] = toInterfaceArray(c.TokenPolicies)
    payload["token_bound_cidrs"] = toInterfaceArray(c.TokenBoundCIDRs)
    payload["token_explicit_max_ttl"] = durationToSeconds(c.TokenExplicitMaxTTL)
    payload["token_no_default_policy"] = c.TokenNoDefaultPolicy
    payload["token_num_uses"] = json.Number(strconv.Itoa(c.TokenNumUses))
    payload["token_period"] = durationToSeconds(c.TokenPeriod)
    payload["token_type"] = c.TokenType
    return payload
}
```

**Critical notes for `toMap()`:**
- `secret` must come from `retrievedSecret` (internal field set by PrepareInternalValues), NOT from the spec directly
- `port`, `dial_timeout`, `read_timeout`, `nas_port`, `token_num_uses` must all be emitted as `json.Number` (Vault returns integers as `json.Number`)
- Duration fields (`token_ttl`, `token_max_ttl`, `token_explicit_max_ttl`, `token_period`) use `durationToSeconds()` to normalize to Vault-read integer-seconds format
- `toInterfaceArray()` for list fields to emit `[]any` matching Vault's read format

**RADIUSAuthEngineUser.toMap():**
```go
func (d *RADIUSAuthEngineUser) toMap() map[string]any {
    payload := map[string]any{}
    payload["policies"] = d.Spec.Policies
    return payload
}
```

Follow OktaAuthEngineGroup.toMap() exactly — emit `policies` as a string. Do NOT include `username` — Vault's user read response only returns `{"policies": "..."}`.

### Kubebuilder Markers Checklist

- All exported Spec fields: `+kubebuilder:validation:Required` or `+kubebuilder:validation:Optional`
- Config `Host`: `+kubebuilder:validation:Required` (no default, no omitempty — required field)
- Config `Port`: `+kubebuilder:validation:Minimum=1`, `+kubebuilder:validation:Maximum=65535`
- Config `DialTimeout`, `ReadTimeout`, `NASPort`: `+kubebuilder:validation:Minimum=0`
- Config `TokenType`: `+kubebuilder:validation:Enum={"service","batch","default","default-service","default-batch",""}`
- Config `TokenNumUses`: `+kubebuilder:validation:Minimum=0`
- Config `TokenNoDefaultPolicy`: zero-value default → use `omitempty`
- Config list fields (TokenPolicies, TokenBoundCIDRs): `+listType=set`
- Root types: `+kubebuilder:object:root=true`, `+kubebuilder:subresource:status`
- Conditions field: `+patchMergeKey=type`, `+patchStrategy=merge`, `+listType=map`, `+listMapKey=type`
- User `Name`: `+kubebuilder:validation:Optional` (overrides metadata.name)

### RBAC Markers

Config controller:
```go
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=radiusauthengineconfigs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=radiusauthengineconfigs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=radiusauthengineconfigs/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=serviceaccounts/token,verbs=create
//+kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;patch
```

User controller:
```go
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=radiusauthengineusers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=radiusauthengineusers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=radiusauthengineusers/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=serviceaccounts/token,verbs=create
//+kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;patch
```

### File Structure

| File | Action | Description |
|------|--------|-------------|
| `api/v1alpha1/radiusauthengineconfig_types.go` | NEW | Config CRD type, VaultObject, ConditionsAware, toMap, credential resolution |
| `api/v1alpha1/radiusauthengineuser_types.go` | NEW | User CRD type, VaultObject, ConditionsAware, toMap |
| `api/v1alpha1/radiusauthengineconfig_webhook.go` | NEW | Config webhook — defaulter, validator, immutable path, credential validation |
| `api/v1alpha1/radiusauthengineuser_webhook.go` | NEW | User webhook — defaulter, validator, immutable path/name |
| `api/v1alpha1/radiusauthengineconfig_test.go` | NEW | Unit tests for config toMap, IsEquivalentToDesiredState |
| `api/v1alpha1/radiusauthengineuser_test.go` | NEW | Unit tests for user toMap, IsEquivalentToDesiredState |
| `internal/controller/radiusauthengineconfig_controller.go` | NEW | Config reconciler with Secret/RandomSecret watches |
| `internal/controller/radiusauthengineuser_controller.go` | NEW | User reconciler — simple VaultResource |
| `cmd/main.go` | UPDATE | Register 2 controllers + 2 webhooks |
| `config/crd/kustomization.yaml` | UPDATE | Add 2 new CRD YAML files to resources list |
| `test/radiusauthengine/` | NEW | Test YAML fixtures for both types |
| `docs/auth-engines/radius.md` | NEW | Engine documentation per DNFR5 |
| `docs/auth-engines/index.md` | UPDATE | Add RADIUS row to Supported Auth Engines table |

### Files Being Modified — Current State

**`cmd/main.go`**: Currently registers ~50+ controllers and webhooks (including Epics 13-15 additions). New registrations follow the exact same pattern:
- Controller: `(&controller.RADIUSAuthEngineConfigReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "RADIUSAuthEngineConfig")}).SetupWithManager(mgr)`
- Webhook (inside ENABLE_WEBHOOKS): `(&redhatcopv1alpha1.RADIUSAuthEngineConfig{}).SetupWebhookWithManager(mgr)`

No existing behavior is changed — purely additive.

**`config/crd/kustomization.yaml`**: Add 2 new CRD YAML files to the `resources` list. Required for Helm chart build.

**`docs/auth-engines/index.md`**: Add a row to the Supported Auth Engines table:
```
| RADIUS | RADIUSAuthEngineConfig | RADIUSAuthEngineUser | [radius.md](radius.md) |
```

### Existing Patterns to Follow (Reference Files)

| Pattern | Reference File |
|---------|---------------|
| Auth engine config with credential resolution (closest analog) | `api/v1alpha1/oktaauthengineconfig_types.go` |
| Auth engine config IsDeletable=false | `api/v1alpha1/oktaauthengineconfig_types.go` |
| Auth engine config GetPath (auth/{path}/config) | `api/v1alpha1/oktaauthengineconfig_types.go` |
| Auth engine user/group with Name override | `api/v1alpha1/oktaauthenginegroup_types.go` |
| Auth engine user/group GetPath (auth/{path}/users/{name}) | `api/v1alpha1/oktaauthenginegroup_types.go` (change `groups` → `users`) |
| Write-only credential stripping in IsEquivalentToDesiredState | `api/v1alpha1/oktaauthengineconfig_types.go` (deletes `api_token`) |
| Config controller with Secret+RandomSecret watches | `internal/controller/oktaauthengineconfig_controller.go` |
| Simple user controller (no watches) | `internal/controller/oktaauthenginegroup_controller.go` |
| Auth config webhook (immutable path only) | `api/v1alpha1/oktaauthengineconfig_webhook.go` |
| Auth user webhook (immutable path+name) | `api/v1alpha1/oktaauthenginegroup_webhook.go` |
| RootCredentialConfig usage for credentials | `api/v1alpha1/oktaauthengineconfig_types.go` |
| filterPayloadToDesiredKeys | `api/v1alpha1/payload_filter.go` |
| toInterfaceArray helper | `api/v1alpha1/utils/vaultutils.go` |
| durationToSeconds helper | `api/v1alpha1/utils/vaultutils.go` |
| json.Number for integer fields | `api/v1alpha1/oktaauthengineconfig_types.go` |
| Documentation template | `docs/engine-doc-template.md` |
| Auth engine doc example | `docs/auth-engines/userpass.md` |

### Unit Test Requirements

**Config tests (`radiusauthengineconfig_test.go`):**
1. `TestRADIUSAuthEngineConfig_toMap` — verify all fields in snake_case, verify `secret` from resolved internal field, verify `port`/`dial_timeout`/`read_timeout`/`nas_port`/`token_num_uses` are `json.Number`, verify `durationToSeconds()` for TTL fields, verify list fields are `[]any`
2. `TestRADIUSAuthEngineConfig_IsEquivalentToDesiredState_Match` — construct Vault-read-shaped payload (without `secret`, with `host`, `port` as `json.Number(1812)`, integer 0 values as `json.Number`, `token_policies` as `[]any`, `token_type` as "default"), verify returns `true`
3. `TestRADIUSAuthEngineConfig_IsEquivalentToDesiredState_Mismatch` — change `host` value, verify returns `false`
4. `TestRADIUSAuthEngineConfig_IsEquivalentToDesiredState_ExtraVaultFields` — add extra Vault-returned field, verify still returns `true` after filtering
5. `TestRADIUSAuthEngineConfig_IsEquivalentToDesiredState_SecretStripping` — verify that a Vault payload without `secret` still matches when desired state would have it (proves stripping works)
6. `TestRADIUSAuthEngineConfig_GetPath` — verify returns `auth/{path}/config`
7. `TestRADIUSAuthEngineConfig_IsDeletable` — returns `false`
8. `TestRADIUSAuthEngineConfig_Conditions` — Get/SetConditions round-trip

**User tests (`radiusauthengineuser_test.go`):**
1. `TestRADIUSAuthEngineUser_toMap` — verify `policies` field matches spec
2. `TestRADIUSAuthEngineUser_IsEquivalentToDesiredState_Match` — Vault-read fixture with `policies` as string, verify returns `true`
3. `TestRADIUSAuthEngineUser_IsEquivalentToDesiredState_Mismatch` — change `policies`, verify returns `false`
4. `TestRADIUSAuthEngineUser_GetPath` — with `spec.name` override and without (falls back to `metadata.name`)
5. `TestRADIUSAuthEngineUser_IsDeletable` — returns `true`
6. `TestRADIUSAuthEngineUser_Conditions` — Get/SetConditions round-trip

**Critical:** All Vault payload fixtures must use `json.Number` for numeric values, not Go `int` or `float64`.

### Webhook Rules

**Config webhook (`radiusauthengineconfig_webhook.go`):**
- `Default()` — set `RADIUSCredentials.PasswordKey` to `"secret"` **only if empty** (do NOT overwrite explicit values — avoid the 14.2 credential defaults bug)
- `ValidateCreate()` — call `credentials.ValidateCredentialSource()` on `RADIUSCredentials`, call `isValid()`
- `ValidateUpdate()` — block `spec.path` changes, call `credentials.ValidateCredentialSource()`, call `isValid()`
- `ValidateDelete()` — no-op

**User webhook (`radiusauthengineuser_webhook.go`):**
- `Default()` — no-op (no credential fields to default)
- `ValidateCreate()` — no-op (no complex validation needed)
- `ValidateUpdate()` — block `spec.path` changes, block `spec.name` changes
- `ValidateDelete()` — no-op

### Anti-Patterns / DO NOT

- **DO NOT** create integration tests for these types — RADIUS cannot be trivially installed in Kind (per "Skip" classification)
- **DO NOT** modify shared framework behavior (`reconcile_skeleton.go`, `vaultresourcereconciler.go`, etc.) — only add new files
- **DO NOT** derive expected payloads from `toMap()` in unit tests — construct independent Vault-read-shaped fixtures
- **DO NOT** use Go `int` or `float64` in unit test Vault payloads — use `json.Number` to match real Vault client behavior
- **DO NOT** emit TTL/duration fields as duration strings in `toMap()` — use `durationToSeconds()` to emit `json.Number` seconds matching Vault's read format
- **DO NOT** forget to add new CRD YAML files to `config/crd/kustomization.yaml` after `make manifests`
- **DO NOT** include `secret` in the Vault-read fixture for `IsEquivalentToDesiredState` tests — Vault never returns it
- **DO NOT** unconditionally overwrite `RADIUSCredentials.PasswordKey` in the webhook defaulter — only set when empty (lesson from 14.2/Epic 15 retro)
- **DO NOT** include `username` in the `toMap()` output for the user type — Vault's user read response only returns `{"policies": "..."}`, the username is part of the URL path, not the body
- **DO NOT** include Enterprise-only fields in the CRDs
- **DO NOT** include deprecated `policies` config field — use `token_policies` instead on the config; `policies` on the user type is the Vault API field name (not deprecated there)
- **DO NOT** add the `host` field as `omitempty` — it is Required and must always be present in serialized JSON

### Project Structure Notes

- All new files follow existing naming conventions: `radiusauthengineconfig` and `radiusauthengineuser` lowercase for file names
- New files go in established directories (`api/v1alpha1/`, `internal/controller/`, `test/`, `docs/auth-engines/`)
- Test fixture directory `test/radiusauthengine/` follows existing pattern (`test/oktaauthengine/`, `test/userpassauthengine/`)
- No conflicts with existing code — purely additive

### References

- [Source: _bmad-output/planning-artifacts/epics.md#Epic-16, Story 16.1 — requirements and acceptance criteria]
- [Source: _bmad-output/project-context.md — all project rules, patterns, and conventions]
- [Source: api/v1alpha1/oktaauthengineconfig_types.go — auth engine config with RootCredentialConfig, api_token stripping, IsDeletable=false]
- [Source: api/v1alpha1/oktaauthenginegroup_types.go — auth engine group/user with policies string, IsDeletable=true]
- [Source: api/v1alpha1/userpassauthengineuser_types.go — auth engine user with credential resolution, password stripping]
- [Source: internal/controller/oktaauthengineconfig_controller.go — auth config controller with Secret/RandomSecret watches]
- [Source: internal/controller/oktaauthenginegroup_controller.go — simple auth group controller]
- [Source: api/v1alpha1/oktaauthengineconfig_webhook.go — auth config webhook with credential defaulting]
- [Source: api/v1alpha1/oktaauthenginegroup_webhook.go — auth group webhook (immutable path+name)]
- [Source: api/v1alpha1/payload_filter.go — filterPayloadToDesiredKeys]
- [Source: api/v1alpha1/utils/vaultutils.go — durationToSeconds, toInterfaceArray helpers]
- [Source: docs/engine-doc-template.md — documentation template]
- [Source: docs/auth-engines/userpass.md — auth engine doc example (same credential resolution pattern)]
- [Source: Vault RADIUS Auth Method API — https://developer.hashicorp.com/vault/api-docs/auth/radius]
- [Source: _bmad-output/implementation-artifacts/15-3-okta-auth-engine-config-and-group-crds.md — closest structural analog (config+group two-CRD pattern)]
- [Source: _bmad-output/implementation-artifacts/15-1-userpass-auth-engine-user-crd.md — credential resolution and write-only stripping pattern]
- [Source: _bmad-output/implementation-artifacts/epic-15-retro-2026-08-18.md — TTL, credential defaulting, webhook rules]

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

- RADIUS config `IsDeletable() = false` — consistent with all other auth engine configs (Okta, GCP, LDAP, JWT/OIDC, Azure, Kubernetes). No DELETE endpoint for `auth/{path}/config`.
- RADIUS shared secret via `RootCredentialConfig` only — no inline secret in CR spec. Default key: `"secret"` (matching Vault API field name).
- Integration tests SKIP — RADIUS server cannot be trivially installed in Kind. Unit tests are the quality gate.
- RADIUSAuthEngineUser follows OktaAuthEngineGroup pattern (policies as comma-separated string, IsDeletable=true).
- Config integer fields (`port`, `dial_timeout`, `read_timeout`, `nas_port`) use Go `int` with `json.Number` emission in `toMap()` — zero-value omission not needed because these have Vault-side defaults and will match on read even at zero.

### Fixes Applied
