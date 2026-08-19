---
baseline_commit: 2856f66ecd326a158496e9ffbe68c89c0c42b15c
---

# Story 16.5: Cloud Foundry Auth Engine — Config and Role CRDs

Status: review

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As an operator developer,
I want CRDs for CFAuthEngineConfig and CFAuthEngineRole,
So that Vault's Cloud Foundry auth method can be managed declaratively.

## Acceptance Criteria

1. **Given** a CFAuthEngineConfig CR is created with CF API address, credentials (cf_username/cf_password from K8s Secret), identity CA certificates, and optional API trusted certificates **When** the reconciler processes it **Then** the config is written to Vault at `auth/{path}/config` and ReconcileSuccessful=True

2. **Given** a CFAuthEngineRole CR is created with bound CF constraints (application IDs, space IDs, organization IDs, instance IDs) **When** the reconciler processes it **Then** the role exists in Vault at `auth/{path}/roles/{name}` and ReconcileSuccessful=True

3. **Given** the CFAuthEngineConfig CR is deleted **When** the reconciler processes deletion **Then** the Vault config is removed via `DELETE auth/{path}/config` (`IsDeletable=true` — Vault has an explicit DELETE endpoint for CF config)

4. **Given** the CFAuthEngineRole CR is deleted **When** the reconciler processes deletion **Then** the role is removed from Vault via `DELETE auth/{path}/roles/{name}` (`IsDeletable=true`)

5. **Given** any CF auth engine CR spec is updated **When** the reconciler processes the update **Then** the Vault resource reflects the updated values

6. **Given** any CF auth engine CR is created or updated **When** the webhook validates it **Then** `spec.path` immutability is enforced on updates, `spec.name` immutability is enforced on updates for CFAuthEngineRole, and credential source validation passes for config

7. **Given** the CFAuthEngineConfig CR has a `CFCredentials` field referencing a K8s Secret **When** the Secret is updated **Then** the reconciler re-reads the credentials and writes the updated config to Vault

8. **Given** the CRD types are implemented **When** the story is marked done **Then** a documentation file exists at `docs/auth-engines/cloud-foundry.md` following `docs/engine-doc-template.md` (DNFR5)

## Tasks / Subtasks

- [x] Task 1: Create `CFAuthEngineConfig` type (AC: 1, 3, 5, 6, 7)
  - [x] 1.1: Create `api/v1alpha1/cfauthengineconfig_types.go` — Spec with `Connection`, `Authentication`, `Path`, inline `CFAuthConfig` struct, `CFCredentials` (`RootCredentialConfig` with usernameKey="cf_username", passwordKey="cf_password")
  - [x] 1.2: Implement `VaultObject` interface — `GetPath()` returns `auth/{path}/config`, `GetPayload()`, `IsEquivalentToDesiredState()`, `PrepareInternalValues()`, `IsDeletable()=true`
  - [x] 1.3: Implement `ConditionsAware` interface — Status with `Conditions []metav1.Condition`
  - [x] 1.4: Implement `setInternalCredentials()` — resolve cf_username/cf_password from K8s Secret, VaultSecret, or RandomSecret (follow `AWSAuthEngineClientConfig` two-credential pattern)
  - [x] 1.5: Implement `toMap()` on `CFAuthConfig` — convert to Vault API snake_case fields; use `json.Number` for integer fields (`login_max_seconds_not_before`, `login_max_seconds_not_after`); emit `toInterfaceArray()` for `identity_ca_certificates` and `cf_api_trusted_certificates`
  - [x] 1.6: Implement `IsEquivalentToDesiredState()` — must delete `cf_password` from desired state (Vault never returns it on read), then `filterPayloadToDesiredKeys`

- [x] Task 2: Create `CFAuthEngineRole` type (AC: 2, 4, 5, 6)
  - [x] 2.1: Create `api/v1alpha1/cfauthenginerole_types.go` — Spec with `Connection`, `Authentication`, `Path`, `Name` (role name override), inline `CFAuthRole` struct
  - [x] 2.2: Implement `VaultObject` interface — `GetPath()` returns `auth/{path}/roles/{name}`, `IsDeletable()=true`, no `PrepareInternalValues` needed
  - [x] 2.3: Implement `ConditionsAware` interface
  - [x] 2.4: Implement `toMap()` on `CFAuthRole` — emit bound constraint arrays via `toInterfaceArray()`, `disable_ip_matching` bool, token fields with `durationToSeconds()` for TTLs, `json.Number` for `token_num_uses`
  - [x] 2.5: Implement `IsEquivalentToDesiredState()` — `removeUnsetFields` + `filterPayloadToDesiredKeys`

- [x] Task 3: Create webhooks (AC: 6)
  - [x] 3.1: Create `api/v1alpha1/cfauthengineconfig_webhook.go` — `admission.Defaulter[*CFAuthEngineConfig]`, `admission.Validator[*CFAuthEngineConfig]`, immutable `spec.path`, credential validation via `ValidateCredentialSource()`, default `CFCredentials.UsernameKey` to `"cf_username"` and `PasswordKey` to `"cf_password"` only when empty
  - [x] 3.2: Create `api/v1alpha1/cfauthenginerole_webhook.go` — `admission.Defaulter[*CFAuthEngineRole]`, `admission.Validator[*CFAuthEngineRole]`, immutable `spec.path`/`spec.name`

- [x] Task 4: Create controllers (AC: 1, 2, 3, 4, 5, 7)
  - [x] 4.1: Create `internal/controller/cfauthengineconfig_controller.go` — embed `ReconcilerBase`, standard VaultResource reconcile, watches on `corev1.Secret` and `RandomSecret` for credential rotation
  - [x] 4.2: Create `internal/controller/cfauthenginerole_controller.go` — simple `For()` with default periodic reconcile predicate (no watches needed)

- [x] Task 5: Register in main.go (AC: 1, 2)
  - [x] 5.1: Add controller registrations for both reconcilers
  - [x] 5.2: Add webhook registrations inside `ENABLE_WEBHOOKS` guard for both types

- [x] Task 6: Unit tests (AC: 1, 2, 5, 6)
  - [x] 6.1: Create `api/v1alpha1/cfauthengineconfig_test.go` — test `toMap()`, test `IsEquivalentToDesiredState()` with independently constructed Vault-read-shaped fixtures (cf_password stripped), negative tests
  - [x] 6.2: Create `api/v1alpha1/cfauthenginerole_test.go` — test `toMap()`, test `IsEquivalentToDesiredState()` with Vault-read fixture (TTLs as `json.Number` integer seconds), negative tests

- [x] Task 7: Test fixtures (AC: all)
  - [x] 7.1: Create test YAML fixtures in `test/cfauthengine/` — config and role CRs
  - [x] 7.2: Integration tests — SKIP (CF is a cloud platform, no CF test double in Kind)

- [x] Task 8: CRD registration and code generation (AC: all)
  - [x] 8.1: Run `make manifests generate fmt vet test`
  - [x] 8.2: Add 2 new CRD YAML files to `config/crd/kustomization.yaml` (CRD registration checklist)
  - [x] 8.3: Verify all existing tests still pass

- [x] Task 9: Documentation (AC: 8)
  - [x] 9.1: Create `docs/auth-engines/cloud-foundry.md` following `docs/engine-doc-template.md`
  - [x] 9.2: Update `docs/auth-engines/index.md` with link to new doc

### Review Findings

- [ ] [Review][Patch] Explicit empty credential keys skip CF defaulting and later break reconciliation [`api/v1alpha1/cfauthengineconfig_webhook.go:45`]
- [ ] [Review][Patch] CF role drift detection still treats set-valued lists as order-sensitive [`api/v1alpha1/cfauthenginerole_types.go:202`]
- [ ] [Review][Patch] RandomSecret `cfUsername` contract is inconsistent across shipped docs and generated schema [`docs/auth-engines/cloud-foundry.md:158`]

## Dev Notes

### Integration Test Classification: SKIP

Per the project's Integration Test Infrastructure Philosophy:
> **Skip it** — For cloud providers and services that cannot be installed in Kind and are hard to mock (e.g., AWS, Azure, GCP, GitHub App, Quay), the integration test must not be run. These types rely on unit test coverage only.

Cloud Foundry is a cloud platform that cannot be installed in Kind. No CF test double is available. No integration tests. Comprehensive unit tests for `toMap()`, `IsEquivalentToDesiredState()`, and webhook validation are the primary quality gate.

### Story Intelligence Chain — Previous Story Context

**Story 15.3 (Okta Auth Engine Config + Group)** is the closest structural analog for CFAuthEngineConfig:
- Auth engine config with credential resolution via `RootCredentialConfig`, `IsDeletable()=false` (Okta) but CF differs — CF has DELETE endpoint so `IsDeletable()=true`
- `api_token` stripping in `IsEquivalentToDesiredState` — same pattern applies: CF strips `cf_password`
- Webhook defaulter sets `PasswordKey` only when empty (lesson from 14.2 review)
- `durationToSeconds()` for all TTL fields in `toMap()` (fixed during Epic 15 retro)
- `json.Number` for integer fields (`token_num_uses`)

**Story 14.2 (AWS Auth Engine Config + Role)** provides the two-credential resolution pattern:
- `RootCredentialConfig` with custom keys (`usernameKey: "access_key"`, `passwordKey: "secret_key"`)
- CF needs the same: `usernameKey: "cf_username"`, `passwordKey: "cf_password"`
- `setInternalCredentials()` retrieves both username and password — follow `AWSAuthEngineClientConfig` exactly
- Controller with Secret/RandomSecret watches for credential rotation

**Epic 15 Retrospective (2026-08-18)** — mandatory rules:
- All `toMap()` duration fields MUST use `durationToSeconds()` — CF role has `token_ttl`, `token_max_ttl`, `token_explicit_max_ttl`, `token_period`
- Credential key defaulting: only set when empty, never overwrite explicit keys
- Strict webhook validation philosophy upheld
- All these patterns are now in `project-context.md`

**GCPAuthEngineRole** is the closest pattern for CFAuthEngineRole:
- Role with `Name` override, `GetPath()` returns `auth/{path}/role/{name}` (CF uses `roles` not `role`)
- `IsDeletable()=true`, simple controller (no watches)
- `removeUnsetFields` + `filterPayloadToDesiredKeys` pipeline

### Vault API Paths

| Operation | Method | Path |
|-----------|--------|------|
| Write config | POST | `auth/{path}/config` |
| Read config | GET | `auth/{path}/config` |
| Delete config | DELETE | `auth/{path}/config` |
| Create/update role | POST | `auth/{path}/roles/{name}` |
| Read role | GET | `auth/{path}/roles/{name}` |
| Delete role | DELETE | `auth/{path}/roles/{name}` |
| List roles | LIST | `auth/{path}/roles` |

**Note:** CF uses `roles` (plural) in the path, not `role` (singular) as in most other auth engines (GCP, Azure, AWS use `role`). Verify in the implementation that `GetPath()` uses `roles/{name}`.

### CFAuthEngineConfig — Vault API Field Reference

**Write (`POST auth/{path}/config`) fields:**
- `identity_ca_certificates` (array: [], required) — Root CA certificate(s) for verifying CF_INSTANCE_CERT
- `cf_api_addr` (string: required) — CF's full API address for instance verification
- `cf_username` (string: required) — Username for authenticating to CF API
- `cf_password` (string: required) — Password for authenticating to CF API. **Write-only.**
- `cf_api_trusted_certificates` (array: []) — Certificate(s) presented by CF API for trust
- `login_max_seconds_not_before` (int: 300) — Max seconds in the past for signature creation
- `login_max_seconds_not_after` (int: 60) — Max seconds in the future for signature creation
- `cf_api_mutual_tls_certificate` (string: "") — Client certificate for mutual TLS with CF API
- `cf_api_mutual_tls_key` (string: "") — Client key for mutual TLS with CF API. **Write-only.**

**Read (`GET auth/{path}/config`) response — `cf_password` and `cf_api_mutual_tls_key` are NEVER returned:**
```json
{
  "identity_ca_certificates": ["-----BEGIN CERTIFICATE-----\n..."],
  "cf_api_addr": "https://api.sys.somewhere.cf-app.com",
  "cf_username": "vault",
  "cf_api_trusted_certificates": ["-----BEGIN CERTIFICATE-----\n..."],
  "login_max_seconds_not_before": 5,
  "login_max_seconds_not_after": 1
}
```

**Critical:** `cf_password` is write-only. Must be deleted from `desiredState` before `IsEquivalentToDesiredState` comparison. Follow `OktaAuthEngineConfig` pattern (deletes `api_token`) and `AWSAuthEngineClientConfig` (deletes `secret_key`).

### Critical: `IsEquivalentToDesiredState` for Config — Password Stripping

Vault never returns `cf_password` on read. The implementation must:
1. Build `desiredState` from `CFAuthConfig.toMap()`
2. `delete(desiredState, "cf_password")` — remove before comparison
3. Use `filterPayloadToDesiredKeys(desiredState, payload)` → `reflect.DeepEqual`

Follow the established pattern from `OktaAuthEngineConfig.IsEquivalentToDesiredState()`.

### CFAuthEngineConfig — IsDeletable = true

Unlike most other auth engine configs (GCP, LDAP, Okta, JWT/OIDC have `IsDeletable=false`), the CF auth method **has an explicit DELETE endpoint** for `auth/{path}/config`. Therefore `IsDeletable()` returns `true`, matching `AWSAuthEngineClientConfig` which also has a DELETE config endpoint.

### CFAuthEngineConfig — Credential Resolution (CF Username + Password)

CF credentials (cf_username/cf_password) must be resolved from one of three sources:
- **K8s Secret**: usernameKey defaults to "cf_username", passwordKey defaults to "cf_password"
- **VaultSecret**: same key mapping from a Vault KV path
- **RandomSecret**: retrieve from RandomSecret's Vault path

Pattern: Use `RootCredentialConfig` with custom default keys (`usernameKey: "cf_username"`, `passwordKey: "cf_password"`). Store resolved values in unexported fields (`retrievedCFUsername`, `retrievedCFPassword` with `json:"-"`). Follow `AWSAuthEngineClientConfig.setInternalCredentials()` for the two-credential resolution pattern.

**Webhook Defaulter:** Set `CFCredentials.UsernameKey` to `"cf_username"` and `CFCredentials.PasswordKey` to `"cf_password"` when empty. Do NOT unconditionally overwrite — only set when the field is empty string (lesson from 14.2 review and Epic 15 retro).

### CFAuthEngineConfig — Controller Pattern

The config controller uses the **standard VaultResource reconcile** (NOT always-write) because the read endpoint returns enough fields for meaningful drift detection (everything except `cf_password` and `cf_api_mutual_tls_key`). The `IsEquivalentToDesiredState` strips write-only fields to avoid false drift.

The controller MUST include Secret and RandomSecret watches for credential rotation detection, following `OktaAuthEngineConfigReconciler.SetupWithManager()`.

### CFAuthEngineConfig — Mutual TLS Fields

The CF auth method supports mutual TLS with the CF API via `cf_api_mutual_tls_certificate` and `cf_api_mutual_tls_key`. Include these as optional string fields in the CRD spec. `cf_api_mutual_tls_key` is write-only (never returned by Vault read) — strip it from `desiredState` alongside `cf_password` in `IsEquivalentToDesiredState`.

### CFAuthEngineRole — Vault API Field Reference

**Write (`POST auth/{path}/roles/{name}`) fields:**
- `bound_application_ids` (array: []) — Application IDs to constrain membership
- `bound_space_ids` (array: []) — Space IDs to constrain membership
- `bound_organization_ids` (array: []) — Organization IDs to constrain membership
- `bound_instance_ids` (array: []) — Instance IDs to constrain membership (changes on `cf push`)
- `disable_ip_matching` (bool: false) — Disable IP-to-cert matching for proxied logins
- `token_ttl` (integer: 0 or string: "") — Incremental lifetime for generated tokens
- `token_max_ttl` (integer: 0 or string: "") — Maximum lifetime for generated tokens
- `token_policies` (array: [] or comma-delimited string: "") — Token policies
- `policies` (array: [], DEPRECATED) — Use `token_policies` instead
- `token_bound_cidrs` (array: [] or comma-delimited string: "") — CIDR blocks for IP restriction
- `token_explicit_max_ttl` (integer: 0 or string: "") — Hard cap max TTL
- `token_no_default_policy` (bool: false) — Exclude default policy
- `token_num_uses` (integer: 0) — Max token uses (0 = unlimited)
- `token_period` (integer: 0 or string: "") — Max allowed period for periodic tokens
- `token_type` (string: "") — Token type: service, batch, default, default-service, default-batch

**Read (`GET auth/{path}/roles/{name}`) sample response:**
```json
{
  "bound_application_ids": ["09d7eb6a-afc2-49a0-bb32-858c22f2b346"],
  "bound_space_ids": ["21005ebb-8943-433e-84e6-d9d9d7338853"],
  "bound_organization_ids": ["9785a884-5e93-49bd-97ee-57bf7c2b20e0"],
  "bound_instance_ids": ["f3e0f176-3f83-4efb-5842-2ff4"],
  "bound_cidrs": ["127.0.0.1/32", "128.252.0.0/16"],
  "policies": ["default"],
  "ttl": 2764790,
  "max_ttl": 2764790,
  "period": 2764790
}
```

**Critical observations:**
- Vault returns TTL fields as integer seconds — use `durationToSeconds()` in `toMap()`
- Vault returns list constraints as `[]any` — use `toInterfaceArray()` in `toMap()`
- Vault may omit unset fields — use `removeUnsetFields` before comparison
- No write-only fields on the role (no secret stripping needed)
- `token_num_uses` must be emitted as `json.Number` — Vault returns integers as `json.Number`

### CRD Field Spec — CFAuthConfig

```go
type CFAuthConfig struct {
    // IdentityCACertificates is the root CA certificate(s) for verifying CF_INSTANCE_CERT.
    // +kubebuilder:validation:Required
    // +listType=set
    IdentityCACertificates []string `json:"identityCACertificates"`

    // CFAPIAddr is the full API address of the CF deployment.
    // +kubebuilder:validation:Required
    CFAPIAddr string `json:"cfAPIAddr"`

    // CFAPITrustedCertificates is the certificate(s) presented by the CF API.
    // +kubebuilder:validation:Optional
    // +listType=set
    CFAPITrustedCertificates []string `json:"cfAPITrustedCertificates,omitempty"`

    // LoginMaxSecondsNotBefore is the max seconds in the past for signature creation.
    // +kubebuilder:validation:Optional
    // +kubebuilder:validation:Minimum=0
    // +kubebuilder:default=300
    LoginMaxSecondsNotBefore int `json:"loginMaxSecondsNotBefore"`

    // LoginMaxSecondsNotAfter is the max seconds in the future for signature creation.
    // +kubebuilder:validation:Optional
    // +kubebuilder:validation:Minimum=0
    // +kubebuilder:default=60
    LoginMaxSecondsNotAfter int `json:"loginMaxSecondsNotAfter"`

    // CFAPIMutualTLSCertificate is the client certificate for mutual TLS with the CF API.
    // +kubebuilder:validation:Optional
    CFAPIMutualTLSCertificate string `json:"cfAPIMutualTLSCertificate,omitempty"`

    // CFAPIMutualTLSKey is the client key for mutual TLS with the CF API. Write-only.
    // +kubebuilder:validation:Optional
    CFAPIMutualTLSKey string `json:"cfAPIMutualTLSKey,omitempty"`

    retrievedCFUsername string `json:"-"`
    retrievedCFPassword string `json:"-"`
}
```

### CRD Field Spec — CFAuthEngineConfigSpec (wrapper)

```go
type CFAuthEngineConfigSpec struct {
    // +kubebuilder:validation:Optional
    Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

    // +kubebuilder:validation:Required
    Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

    // Path at which the CF auth engine is mounted.
    // +kubebuilder:validation:Required
    Path vaultutils.Path `json:"path,omitempty"`

    CFAuthConfig `json:",inline"`

    // CFCredentials is used to provide the CF API username and password.
    // The credentials can be sourced from a Kubernetes Secret, VaultSecret, or RandomSecret.
    // +kubebuilder:validation:Optional
    CFCredentials vaultutils.RootCredentialConfig `json:"cfCredentials,omitempty"`
}
```

### CRD Field Spec — CFAuthEngineRole

```go
type CFAuthRole struct {
    // BoundApplicationIDs constrains instances to specific application IDs.
    // +kubebuilder:validation:Optional
    // +listType=set
    BoundApplicationIDs []string `json:"boundApplicationIDs,omitempty"`

    // BoundSpaceIDs constrains instances to specific space IDs.
    // +kubebuilder:validation:Optional
    // +listType=set
    BoundSpaceIDs []string `json:"boundSpaceIDs,omitempty"`

    // BoundOrganizationIDs constrains instances to specific organization IDs.
    // +kubebuilder:validation:Optional
    // +listType=set
    BoundOrganizationIDs []string `json:"boundOrganizationIDs,omitempty"`

    // BoundInstanceIDs constrains to specific instance IDs. Changes on cf push.
    // +kubebuilder:validation:Optional
    // +listType=set
    BoundInstanceIDs []string `json:"boundInstanceIDs,omitempty"`

    // DisableIPMatching if true, disables IP-to-cert matching for proxied logins.
    // +kubebuilder:validation:Optional
    DisableIPMatching bool `json:"disableIPMatching,omitempty"`

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
    TokenPeriod string `json:"tokenPeriod,omitempty"`

    // TokenType is the type of token to generate.
    // +kubebuilder:validation:Optional
    // +kubebuilder:validation:Enum={"service","batch","default","default-service","default-batch"}
    TokenType string `json:"tokenType,omitempty"`
}
```

### CRD Field Spec — CFAuthEngineRoleSpec (wrapper)

```go
type CFAuthEngineRoleSpec struct {
    // +kubebuilder:validation:Optional
    Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

    // +kubebuilder:validation:Required
    Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

    // Path at which the CF auth engine is mounted.
    // +kubebuilder:validation:Required
    Path vaultutils.Path `json:"path,omitempty"`

    // Name of the CF role. Overrides metadata.name for the Vault object name.
    // +kubebuilder:validation:Optional
    // +kubebuilder:validation:Pattern:=`[a-z0-9]([-a-z0-9]*[a-z0-9])?`
    Name string `json:"name,omitempty"`

    CFAuthRole `json:",inline"`
}
```

### `toMap()` Implementation Notes

**CFAuthConfig.toMap():**
```go
func (c *CFAuthConfig) toMap() map[string]any {
    payload := map[string]any{}
    payload["identity_ca_certificates"] = toInterfaceArray(c.IdentityCACertificates)
    payload["cf_api_addr"] = c.CFAPIAddr
    payload["cf_username"] = c.retrievedCFUsername
    payload["cf_password"] = c.retrievedCFPassword
    payload["cf_api_trusted_certificates"] = toInterfaceArray(c.CFAPITrustedCertificates)
    payload["login_max_seconds_not_before"] = json.Number(strconv.Itoa(c.LoginMaxSecondsNotBefore))
    payload["login_max_seconds_not_after"] = json.Number(strconv.Itoa(c.LoginMaxSecondsNotAfter))
    payload["cf_api_mutual_tls_certificate"] = c.CFAPIMutualTLSCertificate
    payload["cf_api_mutual_tls_key"] = c.CFAPIMutualTLSKey
    return payload
}
```

**Critical notes for `toMap()`:**
- `cf_username` and `cf_password` must come from `retrievedCFUsername`/`retrievedCFPassword` (internal fields set by `PrepareInternalValues`), NOT from the spec directly
- `login_max_seconds_not_before` and `login_max_seconds_not_after` must be emitted as `json.Number` (Vault returns these as integers)
- `identity_ca_certificates` and `cf_api_trusted_certificates` use `toInterfaceArray()` for `[]any` matching Vault's read format
- `cf_api_mutual_tls_key` is write-only; stripped in `IsEquivalentToDesiredState`

**CFAuthRole.toMap():**
```go
func (r *CFAuthRole) toMap() map[string]any {
    payload := map[string]any{}
    payload["bound_application_ids"] = toInterfaceArray(r.BoundApplicationIDs)
    payload["bound_space_ids"] = toInterfaceArray(r.BoundSpaceIDs)
    payload["bound_organization_ids"] = toInterfaceArray(r.BoundOrganizationIDs)
    payload["bound_instance_ids"] = toInterfaceArray(r.BoundInstanceIDs)
    payload["disable_ip_matching"] = r.DisableIPMatching
    payload["token_ttl"] = durationToSeconds(r.TokenTTL)
    payload["token_max_ttl"] = durationToSeconds(r.TokenMaxTTL)
    payload["token_policies"] = toInterfaceArray(r.TokenPolicies)
    payload["policies"] = toInterfaceArray(r.Policies)
    payload["token_bound_cidrs"] = toInterfaceArray(r.TokenBoundCIDRs)
    payload["token_explicit_max_ttl"] = durationToSeconds(r.TokenExplicitMaxTTL)
    payload["token_no_default_policy"] = r.TokenNoDefaultPolicy
    payload["token_num_uses"] = json.Number(strconv.FormatInt(r.TokenNumUses, 10))
    payload["token_period"] = durationToSeconds(r.TokenPeriod)
    payload["token_type"] = r.TokenType
    return payload
}
```

### Kubebuilder Markers Checklist

- All exported Spec fields: `+kubebuilder:validation:Required` or `+kubebuilder:validation:Optional`
- Config `IdentityCACertificates`: `+kubebuilder:validation:Required` (no default, no omitempty)
- Config `CFAPIAddr`: `+kubebuilder:validation:Required` (no default, no omitempty)
- Config `LoginMaxSecondsNotBefore`: `+kubebuilder:default=300` (non-zero default → no `omitempty`), `+kubebuilder:validation:Minimum=0`
- Config `LoginMaxSecondsNotAfter`: `+kubebuilder:default=60` (non-zero default → no `omitempty`), `+kubebuilder:validation:Minimum=0`
- Config list fields (IdentityCACertificates, CFAPITrustedCertificates): `+listType=set`
- Role `TokenType`: `+kubebuilder:validation:Enum={"service","batch","default","default-service","default-batch"}`
- Role `TokenNumUses`: `+kubebuilder:validation:Minimum=0`
- Role list fields (all bound ID arrays, TokenPolicies, Policies, TokenBoundCIDRs): `+listType=set`
- Root types: `+kubebuilder:object:root=true`, `+kubebuilder:subresource:status`
- Conditions field: `+patchMergeKey=type`, `+patchStrategy=merge`, `+listType=map`, `+listMapKey=type`
- Role `Name`: `+kubebuilder:validation:Pattern:=[a-z0-9]([-a-z0-9]*[a-z0-9])?`

### RBAC Markers

Config controller:
```go
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=cfauthengineconfigs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=cfauthengineconfigs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=cfauthengineconfigs/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=serviceaccounts/token,verbs=create
//+kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;patch
```

Role controller:
```go
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=cfauthengineroles,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=cfauthengineroles/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=cfauthengineroles/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=serviceaccounts/token,verbs=create
//+kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;patch
```

### File Structure

| File | Action | Description |
|------|--------|-------------|
| `api/v1alpha1/cfauthengineconfig_types.go` | NEW | Config CRD type, VaultObject, ConditionsAware, toMap, credential resolution |
| `api/v1alpha1/cfauthenginerole_types.go` | NEW | Role CRD type, VaultObject, ConditionsAware, toMap |
| `api/v1alpha1/cfauthengineconfig_webhook.go` | NEW | Config webhook — defaulter, validator, immutable path, credential validation |
| `api/v1alpha1/cfauthenginerole_webhook.go` | NEW | Role webhook — defaulter, validator, immutable path/name |
| `api/v1alpha1/cfauthengineconfig_test.go` | NEW | Unit tests for config toMap, IsEquivalentToDesiredState |
| `api/v1alpha1/cfauthenginerole_test.go` | NEW | Unit tests for role toMap, IsEquivalentToDesiredState |
| `internal/controller/cfauthengineconfig_controller.go` | NEW | Config reconciler with Secret/RandomSecret watches |
| `internal/controller/cfauthenginerole_controller.go` | NEW | Role reconciler — simple VaultResource |
| `cmd/main.go` | UPDATE | Register 2 controllers + 2 webhooks |
| `config/crd/kustomization.yaml` | UPDATE | Add 2 new CRD YAML files to resources list |
| `test/cfauthengine/` | NEW | Test YAML fixtures for both types |
| `docs/auth-engines/cloud-foundry.md` | NEW | Engine documentation per DNFR5 |
| `docs/auth-engines/index.md` | UPDATE | Add link to cloud-foundry.md |

### Files Being Modified — Current State

**`cmd/main.go`**: Currently registers ~46+ controllers and ~46+ webhooks (including Epic 15 additions and any earlier Epic 16 stories). New registrations follow the exact same pattern:
- Controller: `(&controller.CFAuthEngineConfigReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "CFAuthEngineConfig")}).SetupWithManager(mgr)`
- Webhook (inside ENABLE_WEBHOOKS): `(&redhatcopv1alpha1.CFAuthEngineConfig{}).SetupWebhookWithManager(mgr)`

No existing behavior is changed — this is purely additive.

**`config/crd/kustomization.yaml`**: Add the 2 new CRD YAML files to the `resources` list. Required for Helm chart build.

**`docs/auth-engines/index.md`**: Add a row to the Supported Auth Engines table:
```
| Cloud Foundry | CFAuthEngineConfig | CFAuthEngineRole | [cloud-foundry.md](cloud-foundry.md) |
```

### Existing Patterns to Follow (Reference Files)

| Pattern | Reference File |
|---------|---------------|
| Auth engine config with two-credential resolution (username+password) | `api/v1alpha1/awsauthengineconfig_types.go` (access_key/secret_key) |
| Auth engine config with single-credential resolution | `api/v1alpha1/oktaauthengineconfig_types.go` |
| Auth engine config IsDeletable=true (has DELETE endpoint) | `api/v1alpha1/awsauthengineconfig_types.go` |
| Auth engine config GetPath (auth/{path}/config) | `api/v1alpha1/gcpauthengineconfig_types.go` |
| Auth engine role with Name override | `api/v1alpha1/gcpauthenginerole_types.go` |
| Auth engine role GetPath (auth/{path}/role/{name}) | `api/v1alpha1/gcpauthenginerole_types.go` (adapt to `roles/{name}` for CF) |
| Password stripping in IsEquivalentToDesiredState | `api/v1alpha1/oktaauthengineconfig_types.go` (deletes `api_token`) |
| Config controller with Secret+RandomSecret watches | `internal/controller/oktaauthengineconfig_controller.go` |
| Simple role controller (no watches) | `internal/controller/oktaauthenginegroup_controller.go` |
| Auth config webhook (immutable path, credential validation) | `api/v1alpha1/oktaauthengineconfig_webhook.go` |
| Auth role webhook (immutable path+name) | `api/v1alpha1/gcpauthenginerole_webhook.go` |
| removeUnsetFields + filterPayloadToDesiredKeys | `api/v1alpha1/payload_filter.go` |
| toInterfaceArray helper | `api/v1alpha1/utils/vaultutils.go` |
| durationToSeconds helper | `api/v1alpha1/utils/vaultutils.go` |
| Documentation template | `docs/engine-doc-template.md` |

### Unit Test Requirements

**Config tests (`cfauthengineconfig_test.go`):**
1. `TestCFAuthEngineConfig_toMap` — verify all fields in snake_case, verify `cf_username`/`cf_password` from resolved internal fields, verify `login_max_seconds_not_before`/`login_max_seconds_not_after` are `json.Number`, verify cert arrays are `[]any`
2. `TestCFAuthEngineConfig_IsEquivalentToDesiredState_Match` — construct Vault-read-shaped payload (without `cf_password` or `cf_api_mutual_tls_key`, with `cf_api_addr`, `cf_username`, integer values as `json.Number`, cert arrays as `[]any`), verify returns `true`
3. `TestCFAuthEngineConfig_IsEquivalentToDesiredState_Mismatch` — change `cf_api_addr`, verify returns `false`
4. `TestCFAuthEngineConfig_IsEquivalentToDesiredState_ExtraVaultFields` — add extra Vault-returned field, verify still returns `true` after filtering
5. `TestCFAuthEngineConfig_IsEquivalentToDesiredState_PasswordStripping` — verify payload without `cf_password` matches when desired state would have it (proves stripping works)

**Role tests (`cfauthenginerole_test.go`):**
1. `TestCFAuthEngineRole_toMap` — verify all fields, verify bound arrays are `[]any`, verify TTLs use `durationToSeconds()`, verify `token_num_uses` is `json.Number`
2. `TestCFAuthEngineRole_IsEquivalentToDesiredState_Match` — Vault-read fixture with integer TTLs (as `json.Number`), verify returns `true`
3. `TestCFAuthEngineRole_IsEquivalentToDesiredState_Mismatch` — change `bound_application_ids`, verify returns `false`
4. `TestCFAuthEngineRole_IsEquivalentToDesiredState_ExtraVaultFields` — extra fields from Vault, verify filtering

### Anti-Patterns / DO NOT

- **DO NOT** create integration tests for these types — CF is a cloud platform that cannot be installed in Kind (per "Skip it" rule). No CF test double in Kind. Document Skip explicitly.
- **DO NOT** modify shared framework behavior (reconcile_skeleton.go, vaultresourcereconciler.go, etc.) — only add new files
- **DO NOT** derive expected payloads from `toMap()` in unit tests — construct independent Vault-read-shaped fixtures
- **DO NOT** forget to add new CRD YAML files to `config/crd/kustomization.yaml` after `make manifests`
- **DO NOT** use Go `int` or `float64` in unit test Vault payloads — use `json.Number` to match real Vault client behavior
- **DO NOT** include Enterprise-only fields in the CRD
- **DO NOT** include `cf_password` or `cf_api_mutual_tls_key` in the Vault-read fixture for `IsEquivalentToDesiredState` tests — Vault never returns them
- **DO NOT** unconditionally overwrite `CFCredentials.UsernameKey`/`PasswordKey` in the webhook defaulter — only set when empty (lesson from 14.2 review, enforced in Epic 15 retro)
- **DO NOT** emit raw duration strings in `toMap()` for TTL fields — use `durationToSeconds()` (Epic 15 retro mandate)
- **DO NOT** use `role` (singular) in the Vault path — CF uses `roles` (plural): `auth/{path}/roles/{name}`
- **DO NOT** add the `CFAuthEngineRole.Name` field to `toMap()` output — Vault's role API uses the URL path for the role name. The `Name` field is only for `GetPath()` override.
- **DO NOT** set `IsDeletable()=false` for CFAuthEngineConfig — unlike other auth configs, CF **has** a DELETE endpoint

### Novelty Risk: LOW

Both CRD types follow well-established patterns from existing auth engine implementations. CFAuthEngineConfig follows Okta/AWS auth config patterns with credential resolution. CFAuthEngineRole follows GCP/AWS auth role patterns with bound constraints and token parameters. The only CF-specific difference is `IsDeletable()=true` for the config (due to Vault's DELETE endpoint) and the plural `roles` path. No novel architectural patterns required. Skip integration tests — no CF test double in Kind.

### Duration/TTL fields in `toMap()` must use `durationToSeconds()` or `json.Number` (Vault-read format). Never emit raw duration strings.

Config integer fields (`login_max_seconds_not_before`, `login_max_seconds_not_after`) → `json.Number(strconv.Itoa(...))`.
Role TTL fields (`token_ttl`, `token_max_ttl`, `token_explicit_max_ttl`, `token_period`) → `durationToSeconds()`.
Role integer field (`token_num_uses`) → `json.Number(strconv.FormatInt(..., 10))`.

### Project Structure Notes

- All new files follow existing naming conventions: lowercase type name for files
- New files go in established directories (`api/v1alpha1/`, `internal/controller/`, `test/`, `docs/auth-engines/`)
- Test fixture directory `test/cfauthengine/` follows the existing pattern (`test/oktaauthengine/`, `test/awsauthengine/`)
- No conflicts with existing code — purely additive

### References

- [Source: _bmad-output/planning-artifacts/epics.md#Epic-16, Story 16.5 — requirements and acceptance criteria]
- [Source: _bmad-output/project-context.md — all project rules, patterns, and conventions]
- [Source: _bmad-output/implementation-artifacts/epic-15-retro-2026-08-18.md — TTL/credential/webhook mandates]
- [Source: api/v1alpha1/oktaauthengineconfig_types.go — Okta auth config with credential resolution, api_token stripping]
- [Source: api/v1alpha1/awsauthengineconfig_types.go — AWS auth config with two-credential resolution, IsDeletable=true]
- [Source: api/v1alpha1/gcpauthenginerole_types.go — GCP auth role with Name override]
- [Source: internal/controller/oktaauthengineconfig_controller.go — auth config controller with Secret/RandomSecret watches]
- [Source: internal/controller/oktaauthenginegroup_controller.go — simple auth group/role controller]
- [Source: api/v1alpha1/oktaauthengineconfig_webhook.go — auth config webhook with credential defaulting]
- [Source: api/v1alpha1/gcpauthenginerole_webhook.go — auth role webhook (immutable path+name)]
- [Source: docs/engine-doc-template.md — documentation template]
- [Source: Vault CF Auth Method API — https://developer.hashicorp.com/vault/api-docs/auth/cf]
- [Source: _bmad-output/implementation-artifacts/14-2-aws-auth-engine-config-and-role-crds.md — AWS auth story reference]
- [Source: _bmad-output/implementation-artifacts/15-3-okta-auth-engine-config-and-group-crds.md — Okta auth story reference]

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6 (via Cursor)

### Debug Log References

No debug issues encountered. All tests passed on first run after implementation.

### Completion Notes List

- Implemented CFAuthEngineConfig with full VaultObject interface, two-credential resolution (cf_username/cf_password via RootCredentialConfig), IsDeletable()=true, and write-only field stripping (cf_password, cf_api_mutual_tls_key) in IsEquivalentToDesiredState
- Implemented CFAuthEngineRole with VaultObject interface, GetPath() using plural `roles/{name}` path, removeUnsetFields + filterPayloadToDesiredKeys for drift detection, durationToSeconds for all TTL fields, json.Number for integer fields
- Created webhooks: config defaults UsernameKey/PasswordKey only when empty (per Epic 15 retro), immutable path; role validates immutable path and name
- Created config controller with Secret and RandomSecret watches for credential rotation, and simple role controller with periodic reconcile
- Registered both controllers and webhooks in main.go
- Created comprehensive unit tests: 8 config tests (toMap, IsEquivalentToDesiredState match/mismatch/extra fields/password stripping, GetPath, IsDeletable, Default empty/custom keys), 7 role tests (toMap, IsEquivalentToDesiredState match/mismatch/extra fields, GetPath with name/metadata, IsDeletable)
- Created test YAML fixtures in test/cfauthengine/
- Generated CRDs via `make manifests generate fmt vet`, added to config/crd/kustomization.yaml
- Created documentation at docs/auth-engines/cloud-foundry.md and updated index.md
- All tests pass (make test exit code 0)
- Integration tests: SKIP — CF is a cloud platform, no CF test double in Kind

### File List

- api/v1alpha1/cfauthengineconfig_types.go (NEW)
- api/v1alpha1/cfauthenginerole_types.go (NEW)
- api/v1alpha1/cfauthengineconfig_webhook.go (NEW)
- api/v1alpha1/cfauthenginerole_webhook.go (NEW)
- api/v1alpha1/cfauthengineconfig_test.go (NEW)
- api/v1alpha1/cfauthenginerole_test.go (NEW)
- api/v1alpha1/zz_generated.deepcopy.go (MODIFIED — generated)
- internal/controller/cfauthengineconfig_controller.go (NEW)
- internal/controller/cfauthenginerole_controller.go (NEW)
- cmd/main.go (MODIFIED — added controller + webhook registrations)
- config/crd/bases/redhatcop.redhat.io_cfauthengineconfigs.yaml (NEW — generated)
- config/crd/bases/redhatcop.redhat.io_cfauthengineroles.yaml (NEW — generated)
- config/crd/kustomization.yaml (MODIFIED — added 2 CRD entries)
- config/rbac/role.yaml (MODIFIED — generated RBAC rules)
- test/cfauthengine/cf-auth-engine-config.yaml (NEW)
- test/cfauthengine/cf-auth-engine-role.yaml (NEW)
- docs/auth-engines/cloud-foundry.md (NEW)
- docs/auth-engines/index.md (MODIFIED — added Cloud Foundry row)

### Change Log

- 2026-08-19: Initial implementation of CFAuthEngineConfig and CFAuthEngineRole CRDs with full VaultObject interface, webhooks, controllers, unit tests, test fixtures, CRD generation, and documentation. All 9 tasks completed. All unit tests pass.

## Code Review Record

### Review Model Used

gpt-5.4-medium

### Review Findings

- [x] [Review][Patch] HIGH: CFAuthEngineConfig RandomSecret branch never resolves a CF username — only password; retrievedCFUsername stays empty [`api/v1alpha1/cfauthengineconfig_types.go:188`]
- [x] [Review][Patch] HIGH: CFAuthEngineConfig webhook defaulting does not remap inherited `username`/`password` defaults to `cf_username`/`cf_password` when keys are omitted from the admission request [`api/v1alpha1/cfauthengineconfig_webhook.go:41`]
- [x] [Review][Patch] HIGH: CFAuthEngineRole drift detection compares against write-shape keys (`token_ttl`, `token_max_ttl`, `token_period`, `token_bound_cidrs`) but Vault CF role READ returns aliases (`ttl`, `max_ttl`, `period`, `bound_cidrs`) [`api/v1alpha1/cfauthenginerole_types.go:193`]
- [x] [Review][Patch] MEDIUM: CFAuthEngineConfig.IsEquivalentToDesiredState() never calls removeUnsetFields() — empty optional fields cause false drift [`api/v1alpha1/cfauthengineconfig_types.go:160`]

### Decisions Needed / Decisions Taken

- Design decision (pre-resolved): CFAuthEngineConfig `IsDeletable()=true` — CF has an explicit DELETE config endpoint, unlike most other auth engine configs
- Design decision (pre-resolved): CF uses `roles/{name}` (plural) path, not `role/{name}` — confirmed from Vault API docs
- Design decision (pre-resolved): Two-credential resolution pattern (cf_username + cf_password) via `RootCredentialConfig` with custom default keys
- Design decision (pre-resolved): Skip integration tests — no CF test double can run in Kind

### Fixes Applied

1. **RandomSecret username resolution (HIGH):** Added `CFUsername` field to `CFAuthConfig` (mirroring AWS `AccessKey` pattern). RandomSecret branch now uses `r.Spec.CFUsername` instead of empty `retrievedCFUsername`. Added validation requiring `cfUsername` when using randomSecret. Updated `toMap()` to use `retrievedCFUsername` with fallback to `CFUsername` (matching AWS `toMap()` pattern).

2. **Credential-key defaulting (HIGH):** Replaced simple empty-string check with `cfCredentialKeyOmitted()` function that inspects the raw admission request to distinguish truly omitted keys from explicitly set values. Omitted keys default to `cf_username`/`cf_password`; explicit values (even `"username"`/`"password"`) are preserved. Fallback path (no admission request) also remaps inherited schema defaults. Added test for inherited "username"/"password" defaults.

3. **Vault read aliases for CFAuthEngineRole (HIGH):** Added `cfRoleVaultReadAliases` map (`ttl`→`token_ttl`, `max_ttl`→`token_max_ttl`, `period`→`token_period`, `bound_cidrs`→`token_bound_cidrs`) and call to `normalizeVaultReadAliases()` in `IsEquivalentToDesiredState()`. Updated all role drift-detection tests to use Vault-read-shaped payloads (with read aliases) so they exercise the normalization.

4. **removeUnsetFields for CFAuthEngineConfig (MEDIUM):** Added `removeUnsetFields(desiredState, payload)` call in `IsEquivalentToDesiredState()` after write-only field deletion. Updated config drift-detection tests to omit empty optional fields from Vault payloads (matching real Vault read behavior). Added dedicated test for unset optionals.

### Review Iteration 2 Fixes Applied

5. **Default() skips remapping for explicit empty string (HIGH):** Updated `cfCredentialKeyOmitted()` to treat a key present with an empty-string value as omitted for remapping purposes. Previously, sending `usernameKey: ""` or `passwordKey: ""` in the admission request would bypass remapping. Added unit test `TestCFAuthEngineConfig_Default_ExplicitEmptyStringGetsRemapped` that constructs an admission context with explicit empty values and verifies remapping to `cf_username`/`cf_password`.

6. **CFAuthEngineRole.IsEquivalentToDesiredState order-sensitive (MEDIUM):** Added `sortAnyStringSlice` calls for all set-like fields (`bound_application_ids`, `bound_space_ids`, `bound_organization_ids`, `bound_instance_ids`, `token_policies`, `policies`, `token_bound_cidrs`) on both `desiredState` and `filteredPayload` before `DeepEqual`, following the same pattern used by other types (AWS, AppRole, Consul, etc.). Added unit test `TestCFAuthEngineRole_IsEquivalentToDesiredState_DifferentOrderEquivalent`.

7. **Document cfUsername for RandomSecret (MEDIUM):** Updated `docs/auth-engines/cloud-foundry.md` RandomSecret guidance to require `spec.cfUsername` (not `spec.username`), added `cfUsername` to the field descriptions table, and fixed the kubebuilder comment on `CFUsername` in `cfauthengineconfig_types.go` to explicitly state "Use spec.cfUsername (not spec.username)". Ran `make manifests` to regenerate CRDs.

### Review Iteration 3 Findings

- [x] [Review][Patch] Update admission still allows blank credential keys, so remapping only holds on create [`api/v1alpha1/cfauthengineconfig_webhook.go:109`]
- [x] [Review][Patch] Config drift detection still treats certificate set fields as order-sensitive [`api/v1alpha1/cfauthengineconfig_types.go:167`]
- [x] [Review][Patch] Generated CF config CRD schema still documents `spec.username` instead of `spec.cfUsername` [`config/crd/bases/redhatcop.redhat.io_cfauthengineconfigs.yaml:112`]

### Review Iteration 3 Fixes Applied

8. **Default() now runs on update (MEDIUM):** Changed the mutating webhook marker from `verbs=create` to `verbs=create;update` so `Default()` remaps empty/omitted usernameKey/passwordKey to `cf_username`/`cf_password` on both create and update. Added test `TestCFAuthEngineConfig_Default_UpdateRemapsEmptyKeys`.

9. **Sort cert set fields in IsEquivalentToDesiredState (MEDIUM):** Added `sortAnyStringSlice` calls for `identity_ca_certificates` and `cf_api_trusted_certificates` on both `desiredState` and `filteredPayload` before `DeepEqual`, matching the pattern used for CFAuthEngineRole set fields. Added test `TestCFAuthEngineConfig_IsEquivalentToDesiredState_ReorderedCertsEquivalent`.

10. **CRD `spec.username` references — inherited description limitation (LOW):** The `spec.username` text in the generated CRD `cfCredentials.randomSecret`, `cfCredentials.secret`, and `cfCredentials.vaultSecret` descriptions is inherited from the shared `RootCredentialConfig` type comments in `api/v1alpha1/utils/commons.go`. Kubebuilder does not support overriding sub-field descriptions of a referenced type from the embedding site. The CF-local `cfUsername` field description already explicitly states "Use spec.cfUsername (not spec.username)" in the generated CRD, and `docs/auth-engines/cloud-foundry.md` documents the correct usage. Changing `commons.go` would affect all other types that embed `RootCredentialConfig` and is out of scope. Ran `make manifests` to confirm no further local fixes are possible.
