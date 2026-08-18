# Story 15.1: Userpass Auth Engine — User CRD

Status: ready-for-dev

## Story

As an operator developer,
I want a CRD for UserpassAuthEngineUser,
so that Vault userpass accounts can be managed declaratively.

## Acceptance Criteria

1. **Given** a UserpassAuthEngineUser CR is created with policies and token settings **When** the reconciler processes it **Then** the user exists in Vault at `auth/{path}/users/{name}` (password from K8s Secret via PrepareInternalValues) and ReconcileSuccessful=True

2. **Given** a UserpassAuthEngineUser CR spec is updated (token settings, policies) **When** the reconciler processes the update **Then** the Vault user resource reflects the updated values (verified via `IsEquivalentToDesiredState` comparison)

3. **Given** a UserpassAuthEngineUser CR is deleted **When** the reconciler processes deletion **Then** the user is removed from Vault via `DELETE auth/{path}/users/{name}`

4. **Given** a UserpassAuthEngineUser CR is created or updated **When** the webhook validates it **Then** `spec.path` immutability is enforced on updates, `spec.name` immutability is enforced on updates, and credential source validation passes

5. **Given** the CRD type is implemented **When** the story is marked done **Then** a documentation file exists at `docs/auth-engines/userpass.md` following `docs/engine-doc-template.md` (DNFR5)

## Scope Note — Single CRD, No Config Endpoint

Userpass has **no separate config endpoint** — the mount itself (via `AuthEngineMount`) is the configuration. Users are the only API resource. This story produces a single CRD: `UserpassAuthEngineUser`. No `UserpassAuthEngineConfig` CRD is created.

Password **must** come from a K8s Secret reference (via `RootCredentialConfig`), not inline in the CR spec. This follows the same credential resolution pattern used by `AWSAuthEngineClientConfig`, `LDAPAuthEngineConfig`, etc.

## Tasks / Subtasks

- [ ] Task 1: Create `UserpassAuthEngineUser` type (AC: 1, 2, 4)
  - [ ] 1.1: Create `api/v1alpha1/userpassauthengineuser_types.go` — Spec with `Connection`, `Authentication`, `Path`, inline `UserpassUser` struct, `Name`, `PasswordCredentials` (`RootCredentialConfig` with passwordKey="password")
  - [ ] 1.2: Implement `VaultObject` interface — `GetPath()` returns `auth/{path}/users/{name}`, `GetPayload()`, `IsEquivalentToDesiredState()`, `PrepareInternalValues()`, `IsDeletable()=true`
  - [ ] 1.3: Implement `ConditionsAware` interface — Status with `Conditions []metav1.Condition`
  - [ ] 1.4: Implement `setInternalCredentials()` — resolve password from K8s Secret, VaultSecret, or RandomSecret using `RootCredentialConfig` pattern (follow `AWSAuthEngineClientConfig.setInternalCredentials()`)
  - [ ] 1.5: Implement `toMap()` on `UserpassUser` — emit all fields in Vault API snake_case; include `password` from resolved credential; use `durationToSeconds()` for duration fields, `json.Number` for integer fields
  - [ ] 1.6: Implement `IsEquivalentToDesiredState()` — must `delete(desiredState, "password")` (Vault never returns it on read), then `removeUnsetFields` + `filterPayloadToDesiredKeys` + `reflect.DeepEqual`

- [ ] Task 2: Create webhook (AC: 4)
  - [ ] 2.1: Create `api/v1alpha1/userpassauthengineuser_webhook.go` — `admission.Defaulter[*UserpassAuthEngineUser]`, `admission.Validator[*UserpassAuthEngineUser]`, immutable `spec.path`/`spec.name`, credential source validation via `ValidateCredentialSource()`

- [ ] Task 3: Create controller (AC: 1, 2, 3)
  - [ ] 3.1: Create `internal/controller/userpassauthengineuser_controller.go` — embed `ReconcilerBase`, standard `VaultResource` reconcile, include Secret and RandomSecret watches for password rotation (follow `GCPAuthEngineConfigReconciler.SetupWithManager()` pattern)

- [ ] Task 4: Register in main.go (AC: 1)
  - [ ] 4.1: Add controller registration for the reconciler
  - [ ] 4.2: Add webhook registration inside `ENABLE_WEBHOOKS` guard

- [ ] Task 5: Unit tests (AC: 1, 2, 4)
  - [ ] 5.1: Create `api/v1alpha1/userpassauthengineuser_test.go` — test `toMap()` output, `IsEquivalentToDesiredState()` match/mismatch/extra-fields (including password stripping), `GetPath()`, `IsDeletable()`, `GetConditions()`/`SetConditions()`

- [ ] Task 6: Integration tests (AC: 1, 2, 3)
  - [ ] 6.1: Create test YAML fixtures in `test/userpassauthengine/` — mount CR + user CR + updated user CR + password Secret
  - [ ] 6.2: Create `internal/controller/userpassauthengineuser_controller_test.go` with `//go:build integration` — create, verify ReconcileSuccessful, update, verify update, delete, verify Vault cleanup

- [ ] Task 7: CRD registration and code generation (AC: all)
  - [ ] 7.1: Run `make manifests generate fmt vet test`
  - [ ] 7.2: Add new CRD YAML file to `config/crd/kustomization.yaml` (CRD registration checklist)
  - [ ] 7.3: Verify all existing tests still pass

- [ ] Task 8: Documentation (AC: 5)
  - [ ] 8.1: Create `docs/auth-engines/userpass.md` following `docs/engine-doc-template.md`
  - [ ] 8.2: Update `docs/auth-engines/index.md` — add Userpass row to Supported Auth Engines table

## Dev Notes

### Integration Test Classification: INSTALL IN KIND (Full Integration Tests)

Per the project's Integration Test Infrastructure Philosophy:
> **Install in Kind** — If the service can be installed in the Kind cluster and configured to work with Vault, the test must deploy it as a real service.

Userpass is a Vault-native auth method with zero external dependencies. Only Vault itself is needed, and it's already running in the Kind cluster integration test infrastructure. Full integration tests are required.

### Story Intelligence Chain — Previous Story Context

**Story 14.1 (AppRole Auth Engine)** is the direct predecessor in this project phase:
- Established the pattern for single-CRD auth engine types (role-only, no config CRD)
- AppRole has no config endpoint — same as Userpass (no config endpoint; mount is the config)
- AppRole used `removeUnsetFields` + `filterPayloadToDesiredKeys` + `reflect.DeepEqual` for `IsEquivalentToDesiredState` — same pipeline applies here
- AppRole required no credential resolution — **this story differs** by requiring `PrepareInternalValues` for password resolution from K8s Secret
- AppRole's `toMap()` emits all fields unconditionally — this story's `toMap()` must additionally include `password` from resolved credentials
- Code review found that AppRole updates cannot clear previously-set optional fields back to default — same limitation will exist here; document but don't address (matches project-wide behavior)

**Story 14.2 (AWS Auth Engine)** established the credential resolution pattern for auth engines:
- `AWSAuthEngineClientConfig` uses `RootCredentialConfig` with custom key defaults (`usernameKey: "access_key"`, `passwordKey: "secret_key"`)
- `setInternalCredentials()` resolves credentials from K8s Secret/VaultSecret/RandomSecret via the shared `RootCredentialConfig.GetCredentialsFromSecret()` pipeline
- `IsEquivalentToDesiredState()` deletes `secret_key` from desired state before comparison — **this story must delete `password` analogously**
- Controller includes Secret and RandomSecret watches for credential rotation detection

**Epic 14 retrospective action items (open):**
- `durationToSeconds()` propagation to older auth engine types — not relevant to this new type (we'll use it from the start)
- Strict webhook validation philosophy — document and apply
- 14.2 credential defaults bug (webhook defaulter overwrites explicit usernameKey/passwordKey) — **avoid repeating this bug**: only set defaults if empty, don't overwrite

### Vault API Paths

| Operation | Method | Path |
|-----------|--------|------|
| Create/update user | POST | `auth/{path}/users/{name}` |
| Read user | GET | `auth/{path}/users/{name}` |
| Delete user | DELETE | `auth/{path}/users/{name}` |
| List users | LIST | `auth/{path}/users` |
| Update password | POST | `auth/{path}/users/{name}/password` |
| Update policies | POST | `auth/{path}/users/{name}/policies` |
| Login | POST | `auth/{path}/login/{name}` |

Only the first three operations are managed by this CRD. Password update, policy update, login, and list are operational, not declarative.

### UserpassAuthEngineUser — Vault API Field Reference

**Write (`POST auth/{path}/users/{name}`) fields:**

| Vault API Field | Type | Default | Description |
|-----------------|------|---------|-------------|
| `password` | string | `""` | Password for the user. Only required on create. Write-only — never returned on read. Mutually exclusive with `password_hash`. |
| `token_ttl` | duration | `0` | Incremental lifetime for generated tokens |
| `token_max_ttl` | duration | `0` | Maximum lifetime for generated tokens |
| `token_policies` | array | `[]` | Token policies |
| `token_bound_cidrs` | array | `[]` | IP blocks for token use |
| `token_explicit_max_ttl` | duration | `0` | Hard cap max TTL |
| `token_no_default_policy` | bool | `false` | Exclude default policy |
| `token_num_uses` | integer | `0` | Max token uses (0=unlimited) |
| `token_period` | duration | `0` | Period for periodic tokens |
| `token_type` | string | `""` | Token type: service, batch, default |

**Excluded fields:**
- `password_hash` — Pre-hashed bcrypt password. Not suitable for CRD spec (security concern, complex user experience). Users should provide passwords via K8s Secret reference.
- `alias_metadata` — Pre-populates custom metadata for entity aliases. Optional advanced feature; can be added in a future iteration if needed.
- `policies` — Deprecated; use `token_policies` instead.

**Read (`GET auth/{path}/users/{name}`) sample response:**
```json
{
  "data": {
    "token_bound_cidrs": ["127.0.0.1", "128.252.0.0/16"],
    "token_explicit_max_ttl": 0,
    "token_max_ttl": 0,
    "token_no_default_policy": false,
    "token_num_uses": 0,
    "token_period": 0,
    "token_policies": ["admin", "default"],
    "token_ttl": 0,
    "token_type": "default"
  }
}
```

**Critical observations for `IsEquivalentToDesiredState`:**
- `password` is **write-only** — Vault NEVER returns it on read. Must `delete(desiredState, "password")` before comparison. Follow the `AWSSecretEngineConfig` / `AWSAuthEngineClientConfig` pattern for `secret_key` stripping.
- Vault returns duration fields (`token_ttl`, `token_max_ttl`, `token_explicit_max_ttl`, `token_period`) as **integer seconds** — use `durationToSeconds()` in `toMap()`
- Vault returns integer fields (`token_num_uses`) as `json.Number` — use `json.Number(strconv.Itoa())` in `toMap()`
- Vault may omit zero-value bool fields (`token_no_default_policy`) — handle with zero-value elision in `IsEquivalentToDesiredState` (same as AppRole pattern)
- Vault returns list fields (`token_policies`, `token_bound_cidrs`) as `[]any` — use `toInterfaceArray()` in `toMap()`

### CRD Field Spec — UserpassAuthEngineUser

```go
type UserpassAuthEngineUserSpec struct {
    // Connection represents the information needed to connect to Vault. This operator uses the standard Vault environment variables to connect to Vault. If you need to override those settings and for example connect to a different Vault instance, you can do with this section of the CR.
    // +kubebuilder:validation:Optional
    Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

    // Authentication is the kube auth configuration to be used to execute this request
    // +kubebuilder:validation:Required
    Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

    // Path at which to make the configuration.
    // The final path in Vault will be {[spec.authentication.namespace]}/auth/{spec.path}/users/{metadata.name}.
    // The authentication role must have the following capabilities = [ "create", "read", "update", "delete"] on that path.
    // +kubebuilder:validation:Required
    Path vaultutils.Path `json:"path,omitempty"`

    UserpassUser `json:",inline"`

    // PasswordCredentials specifies where to retrieve the password for this user.
    // The password is resolved from a K8s Secret, VaultSecret, or RandomSecret.
    // Only the passwordKey is used (usernameKey is ignored — the username comes from metadata.name or spec.name).
    // +kubebuilder:validation:Required
    PasswordCredentials vaultutils.RootCredentialConfig `json:"passwordCredentials"`

    // The name of the object created in Vault. If this is specified it takes precedence over {metadata.name}
    // +kubebuilder:validation:Optional
    // +kubebuilder:validation:Pattern:=`[a-z0-9]([-a-z0-9._]*[a-z0-9])?`
    Name string `json:"name,omitempty"`
}

type UserpassUser struct {
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

    retrievedPassword string `json:"-"`
}
```

**Key field rules:**
- `Name` pattern allows underscores, hyphens, periods per Vault userpass username rules: `[a-z0-9]([-a-z0-9._]*[a-z0-9])?`
- `PasswordCredentials` is `Required` — a password source must always be specified
- All duration fields are `string` (user provides `"1h"`, `"10m"`); `toMap()` converts via `durationToSeconds()`
- Integer field `TokenNumUses` uses Go `int`; `toMap()` emits `json.Number`
- `TokenType` uses `+kubebuilder:validation:Enum` for admission-time validation
- `retrievedPassword` is unexported with `json:"-"` — populated by `PrepareInternalValues`
- All token fields have zero-value defaults → use `omitempty`

### `toMap()` Implementation Notes

```go
func (d *UserpassUser) toMap() map[string]any {
    payload := map[string]any{}
    if d.retrievedPassword != "" {
        payload["password"] = d.retrievedPassword
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

**Key:** `password` is included only when `retrievedPassword` is non-empty (populated by `PrepareInternalValues`). All other fields are conditional — only emitted when non-zero. This matches the AppRole `toMap()` style (conditional inclusion for zero-value defaults).

### `IsEquivalentToDesiredState` Implementation Notes

```go
func (d *UserpassAuthEngineUser) IsEquivalentToDesiredState(payload map[string]any) bool {
    desiredState := d.Spec.UserpassUser.toMap()
    // password is write-only — Vault never returns it on read
    delete(desiredState, "password")
    removeUnsetFields(desiredState, payload)

    // Handle zero-value bool fields Vault may omit
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

**Critical:** Password stripping is the primary custom logic. Everything else follows the AppRole `IsEquivalentToDesiredState` pattern exactly.

### `PrepareInternalValues` — Credential Resolution

```go
func (d *UserpassAuthEngineUser) PrepareInternalValues(ctx context.Context, object client.Object) error {
    return d.setInternalCredentials(ctx)
}

func (d *UserpassAuthEngineUser) setInternalCredentials(ctx context.Context) error {
    // Resolve password from K8s Secret, VaultSecret, or RandomSecret
    // Follow AWSAuthEngineClientConfig.setInternalCredentials() pattern
    // Only passwordKey is used — usernameKey is ignored
    // Store resolved password in d.Spec.UserpassUser.retrievedPassword
}
```

Follow `AWSAuthEngineClientConfig.setInternalCredentials()` or `LDAPAuthEngineConfig.setInternalCredentials()` as reference. Only the password is resolved — the username is `spec.name` or `metadata.name`, not from the credential source.

### `GetPath()` Implementation

```go
func (d *UserpassAuthEngineUser) GetPath() string {
    if d.Spec.Name != "" {
        return vaultutils.CleansePath("auth/" + string(d.Spec.Path) + "/users/" + d.Spec.Name)
    }
    return vaultutils.CleansePath("auth/" + string(d.Spec.Path) + "/users/" + d.Name)
}
```

### Webhook Rules

- `Default()` — set `PasswordCredentials.PasswordKey` to `"password"` **only if empty** (do NOT overwrite explicit values — avoid the 14.2 credential defaults bug)
- `ValidateCreate()` — call `credentials.ValidateCredentialSource()` on `PasswordCredentials`, call `isValid()`
- `ValidateUpdate()` — block `spec.path` changes, block `spec.name` changes, call `credentials.ValidateCredentialSource()`, call `isValid()`
- `ValidateDelete()` — no-op

### Controller Pattern

Standard `VaultResource` reconcile with Secret and RandomSecret watches for password rotation:
- Embed `ReconcilerBase`
- Fetch instance → `prepareContext` → `NewVaultResource` → `Reconcile`
- `SetupWithManager`: `For()` with `NewDefaultPeriodicReconcilePredicate()`, plus `Watches` on `corev1.Secret` and `RandomSecret` for credential rotation detection
- Follow `GCPAuthEngineConfigReconciler.SetupWithManager()` for watch pattern

### RBAC Markers

```go
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=userpassauthengineusers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=userpassauthengineusers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=userpassauthengineusers/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=serviceaccounts/token,verbs=create
//+kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;patch
```

### Kubebuilder Markers Checklist

- All exported Spec fields: `+kubebuilder:validation:Required` or `+kubebuilder:validation:Optional`
- `PasswordCredentials`: `+kubebuilder:validation:Required` (must always provide a password source)
- `TokenType`: `+kubebuilder:validation:Enum:={"service","batch","default","default-service","default-batch",""}`
- `TokenNumUses`: `+kubebuilder:validation:Minimum=0`
- Root type: `+kubebuilder:object:root=true`, `+kubebuilder:subresource:status`
- Conditions field: `+patchMergeKey=type`, `+patchStrategy=merge`, `+listType=map`, `+listMapKey=type`
- Webhook markers: mutating + validating paths for `userpassauthengineuser`

### Unit Test Requirements

**Tests (`userpassauthengineuser_test.go`):**

1. `TestUserpassAuthEngineUserGetPath` — with `spec.name` override and without (falls back to `metadata.name`)
2. `TestUserpassUserToMap` — verify all field keys match Vault API snake_case names, verify `password` included when `retrievedPassword` set, verify `json.Number` for `token_num_uses`, verify `durationToSeconds` output for TTL fields
3. `TestUserpassUserToMap_MinimalFields` — only `password` set, verify minimal output
4. `TestUserpassAuthEngineUserIsEquivalentToDesiredState_Match` — construct Vault-read-shaped payload with `json.Number` durations and integers (no `password` field), verify returns `true`
5. `TestUserpassAuthEngineUserIsEquivalentToDesiredState_Mismatch` — change `token_policies`, verify returns `false`
6. `TestUserpassAuthEngineUserIsEquivalentToDesiredState_ExtraVaultFields` — add extra Vault fields, verify still returns `true` (filtered out)
7. `TestUserpassAuthEngineUserIsEquivalentToDesiredState_PasswordStripped` — verify that even when `retrievedPassword` is set, `IsEquivalentToDesiredState` returns `true` against a Vault-read payload without `password` field (proves password stripping works)
8. `TestUserpassAuthEngineUserIsDeletable` — returns `true`
9. `TestUserpassAuthEngineUserConditions` — Get/SetConditions round-trip

**Critical:** All Vault payload fixtures must use `json.Number` for numeric values, not Go `int` or `float64`.

### Integration Test Plan

**Fixture files in `test/userpassauthengine/`:**
- `test-userpass-auth-mount.yaml` — `AuthEngineMount` with `type: userpass`, unique path for test isolation
- `test-userpass-password-secret.yaml` — `Secret` of type `Opaque` with `password` key (the credential source for the user CR)
- `test-userpass-auth-user.yaml` — `UserpassAuthEngineUser` with `passwordCredentials.secret.name`, `tokenPolicies`, `tokenTTL`, `tokenMaxTTL`
- `test-userpass-auth-user-updated.yaml` — updated version with changed `tokenPolicies` and `tokenTTL`

**Test file (`internal/controller/userpassauthengineuser_controller_test.go`):**

Test structure: `Describe("UserpassAuthEngineUser controller", Ordered, func() {...})` with cross-Context state sharing:

1. **Context "When creating a Userpass auth mount"** — create `AuthEngineMount` of `type: userpass`, wait for ReconcileSuccessful
2. **Context "When creating a password Secret"** — create the K8s Secret containing the password
3. **Context "When creating a UserpassAuthEngineUser"** — create user CR, wait for ReconcileSuccessful, verify user exists in Vault via direct API read
4. **Context "When updating a UserpassAuthEngineUser"** — update spec fields (tokenPolicies, tokenTTL), wait for re-reconcile, verify Vault state matches updated values
5. **Context "When deleting"** — delete user CR, wait for K8s NotFound, verify user is gone from Vault (404). Then delete the auth mount.

### File Structure

| File | Action | Description |
|------|--------|-------------|
| `api/v1alpha1/userpassauthengineuser_types.go` | NEW | User CRD type, VaultObject, ConditionsAware, toMap, credential resolution |
| `api/v1alpha1/userpassauthengineuser_webhook.go` | NEW | User webhook — defaulter, validator, immutable path/name, credential validation |
| `api/v1alpha1/userpassauthengineuser_test.go` | NEW | Unit tests for toMap, IsEquivalentToDesiredState (including password stripping), GetPath |
| `internal/controller/userpassauthengineuser_controller.go` | NEW | User reconciler with Secret/RandomSecret watches for password rotation |
| `internal/controller/userpassauthengineuser_controller_test.go` | NEW | Integration tests (create, update, delete lifecycle) |
| `cmd/main.go` | UPDATE | Register 1 controller + 1 webhook |
| `config/crd/kustomization.yaml` | UPDATE | Add 1 new CRD YAML file to resources list |
| `test/userpassauthengine/` | NEW | Test YAML fixtures (mount, secret, user, updated-user) |
| `docs/auth-engines/userpass.md` | NEW | Engine documentation per DNFR5 |
| `docs/auth-engines/index.md` | UPDATE | Add Userpass row to Supported Auth Engines table |

### Files Being Modified — Current State

**`cmd/main.go`**: Currently registers ~48+ controllers and webhooks (including Epics 13-14 additions). New registrations follow the exact same pattern:
- Controller: `(&controller.UserpassAuthEngineUserReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "UserpassAuthEngineUser")}).SetupWithManager(mgr)`
- Webhook (inside ENABLE_WEBHOOKS): `(&redhatcopv1alpha1.UserpassAuthEngineUser{}).SetupWebhookWithManager(mgr)`

No existing behavior is changed — purely additive.

**`config/crd/kustomization.yaml`**: Add 1 new CRD YAML file to the `resources` list. Required for Helm chart build.

**`docs/auth-engines/index.md`**: Add Userpass row to the Supported Auth Engines table:
```
| Userpass | — | UserpassAuthEngineUser | [userpass.md](userpass.md) |
```
Note: Config CRD column shows "—" because Userpass has no config CRD (mount-level config only).

### Existing Patterns to Follow (Reference Files)

| Pattern | Reference File |
|---------|---------------|
| Auth engine with single CRD (no config), closest structural analog | `api/v1alpha1/approleauthenginerole_types.go` |
| Auth engine with credential resolution (Secret/VaultSecret/RandomSecret) | `api/v1alpha1/awsauthengineconfig_types.go` |
| setInternalCredentials() for password resolution | `api/v1alpha1/awsauthengineconfig_types.go` (or `ldapauthengineconfig_types.go`) |
| Password/secret stripping in IsEquivalentToDesiredState | `api/v1alpha1/awsauthengineconfig_types.go` (deletes `secret_key`) |
| Config controller with Secret+RandomSecret watches | `internal/controller/gcpauthengineconfig_controller.go` |
| Auth webhook with credential validation (ValidateCredentialSource) | `api/v1alpha1/awsauthengineconfig_webhook.go` |
| Auth webhook (immutable path/name) | `api/v1alpha1/approleauthenginerole_webhook.go` |
| Standard VaultResource controller | `internal/controller/approleauthenginerole_controller.go` |
| Auth engine mount fixture | `test/approleauthengine/test-approle-auth-mount.yaml` |
| removeUnsetFields + filterPayloadToDesiredKeys | `api/v1alpha1/payload_filter.go` |
| durationToSeconds helper | `api/v1alpha1/utils/vaultutils.go` |
| toInterfaceArray helper | `api/v1alpha1/utils/vaultutils.go` |
| sortAnyStringSlice for set comparison | `api/v1alpha1/approleauthenginerole_types.go` |
| RootCredentialConfig type | `api/v1alpha1/utils/commons.go` |
| Documentation template | `docs/engine-doc-template.md` |

### Anti-Patterns / DO NOT

- **DO NOT** create a `UserpassAuthEngineConfig` CRD — Userpass has no separate config endpoint; the mount itself (via `AuthEngineMount`) is the configuration
- **DO NOT** put password inline in the CR spec — always resolve from `RootCredentialConfig` (K8s Secret, VaultSecret, or RandomSecret)
- **DO NOT** support `password_hash` in the CRD — pre-hashed passwords are a security risk (users should provide plaintext via K8s Secret, Vault encrypts it)
- **DO NOT** include deprecated `policies` field in the CRD — use `tokenPolicies` only
- **DO NOT** modify shared framework behavior (`reconcile_skeleton.go`, `vaultresourcereconciler.go`, etc.) — only add new files
- **DO NOT** derive expected payloads from `toMap()` in unit tests — construct independent Vault-read-shaped fixtures
- **DO NOT** use Go `int` or `float64` in unit test Vault payloads — use `json.Number` to match real Vault client behavior
- **DO NOT** emit TTL/duration fields as duration strings in `toMap()` — use `durationToSeconds()` to emit `json.Number` seconds matching Vault's read format
- **DO NOT** forget to add new CRD YAML file to `config/crd/kustomization.yaml` after `make manifests`
- **DO NOT** forget to `delete(desiredState, "password")` in `IsEquivalentToDesiredState` — Vault never returns password on read; forgetting this causes permanent drift and unnecessary rewrites
- **DO NOT** unconditionally overwrite `PasswordCredentials.PasswordKey` in webhook `Default()` — only set if empty (avoids 14.2 credential defaults bug)
- **DO NOT** add the username to `toMap()` output — Vault's user API uses the URL path for the username, not a body field. The username comes from `GetPath()` only.

### Novelty Risk: LOW

This type follows well-established patterns from AppRole (single CRD, no config endpoint) and AWS Auth (credential resolution via `RootCredentialConfig`, write-only field stripping in `IsEquivalentToDesiredState`). The Vault API surface for userpass users is simple — standard token parameters plus a write-only password. No novel architectural patterns required.

### Project Structure Notes

- All new files follow existing naming conventions: `userpassauthengineuser` lowercase for file names
- New files go in established directories (`api/v1alpha1/`, `internal/controller/`, `test/`, `docs/auth-engines/`)
- Test fixture directory `test/userpassauthengine/` follows existing pattern (`test/approleauthengine/`, `test/kubernetesauthengine/`)
- No conflicts with existing code — purely additive

### References

- [Source: _bmad-output/planning-artifacts/epics.md#Epic-15, Story 15.1 — requirements and acceptance criteria]
- [Source: _bmad-output/project-context.md — all project rules, patterns, and conventions]
- [Source: api/v1alpha1/approleauthenginerole_types.go — single-CRD auth engine type, closest analog]
- [Source: api/v1alpha1/awsauthengineconfig_types.go — auth engine with RootCredentialConfig, secret_key stripping]
- [Source: api/v1alpha1/ldapauthengineconfig_types.go — auth engine with credential resolution from K8s Secret]
- [Source: internal/controller/gcpauthengineconfig_controller.go — controller with Secret/RandomSecret watches]
- [Source: api/v1alpha1/utils/commons.go — RootCredentialConfig type definition]
- [Source: api/v1alpha1/payload_filter.go — filterPayloadToDesiredKeys and removeUnsetFields]
- [Source: docs/engine-doc-template.md — documentation template]
- [Source: Vault Userpass API — https://developer.hashicorp.com/vault/api-docs/auth/userpass]
- [Source: _bmad-output/implementation-artifacts/14-1-approle-auth-engine-config-and-role-crds.md — direct predecessor story]
- [Source: _bmad-output/implementation-artifacts/14-2-aws-auth-engine-config-and-role-crds.md — credential resolution pattern source]

## Code Review Record

### Review Model Used

### Review Findings

### Decisions Needed / Decisions Taken

- No config CRD — Userpass has no separate config endpoint (mount-level config only via `AuthEngineMount`)
- Password via `RootCredentialConfig` only — no inline password in CR spec, no `password_hash` support
- Integration tests required — Userpass is Vault-native, installs in Kind with zero external dependencies

### Fixes Applied

## Dev Agent Record

### Agent Model Used

{{agent_model_name_version}}

### Debug Log References

### Completion Notes List

### File List
