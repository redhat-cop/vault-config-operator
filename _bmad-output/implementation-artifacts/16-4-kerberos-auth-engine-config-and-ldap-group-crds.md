# Story 16.4: Kerberos Auth Engine — Config and LDAP Group CRDs

Status: review

## Story

As an operator developer,
I want CRDs for KerberosAuthEngineConfig, KerberosAuthEngineLDAPConfig, and KerberosAuthEngineGroup,
So that Vault's Kerberos auth method can be managed declaratively.

## Acceptance Criteria

1. **Given** a KerberosAuthEngineConfig CR is created with a keytab (from K8s Secret), service account, and optional flags **When** the reconciler processes it **Then** the Kerberos config is written to Vault at `auth/{path}/config` and ReconcileSuccessful=True

2. **Given** a KerberosAuthEngineLDAPConfig CR is created with LDAP connection settings (url, binddn, bindpass from K8s Secret, groupdn, etc.) **When** the reconciler processes it **Then** the LDAP config is written to Vault at `auth/{path}/config/ldap` and ReconcileSuccessful=True

3. **Given** a KerberosAuthEngineGroup CR is created with group name and policies **When** the reconciler processes it **Then** the group mapping exists in Vault at `auth/{path}/groups/{name}` and ReconcileSuccessful=True

4. **Given** the KerberosAuthEngineConfig CR is deleted **When** the reconciler processes deletion **Then** the K8s object is removed but Vault config is NOT deleted (`IsDeletable=false` — auth engine configs are not deletable; the mount owns the config lifecycle)

5. **Given** the KerberosAuthEngineLDAPConfig CR is deleted **When** the reconciler processes deletion **Then** the K8s object is removed but Vault LDAP config is NOT deleted (`IsDeletable=false` — no DELETE endpoint for `auth/{path}/config/ldap`)

6. **Given** the KerberosAuthEngineGroup CR is deleted **When** the reconciler processes deletion **Then** the group mapping is removed from Vault via `DELETE auth/{path}/groups/{name}` (`IsDeletable=true`)

7. **Given** any Kerberos auth engine CR spec is updated **When** the reconciler processes the update **Then** the Vault resource reflects the updated values

8. **Given** any Kerberos auth engine CR is created or updated **When** the webhook validates it **Then** `spec.path` immutability is enforced on updates, and `spec.name` immutability is enforced on updates for KerberosAuthEngineGroup

9. **Given** the KerberosAuthEngineConfig CR has a `keytabSecret` field referencing a K8s Secret **When** the Secret is updated **Then** the reconciler re-reads the keytab and writes the updated config to Vault

10. **Given** the KerberosAuthEngineLDAPConfig CR has `LDAPCredentials` referencing a K8s Secret **When** the Secret is updated **Then** the reconciler re-reads the bind password and writes the updated config to Vault

11. **Given** the CRD types are implemented **When** the story is marked done **Then** a documentation file exists at `docs/auth-engines/kerberos.md` following `docs/engine-doc-template.md` (DNFR5)

## Tasks / Subtasks

- [x] Task 1: Create `KerberosAuthEngineConfig` type (AC: 1, 4, 7, 8, 9)
  - [x] 1.1: Create `api/v1alpha1/kerberosauthengineconfig_types.go` — Spec with `Connection`, `Authentication`, `Path`, inline `KerberosAuthConfig` struct, `KeytabSecret` (corev1.LocalObjectReference)
  - [x] 1.2: Implement `VaultObject` interface — `GetPath()` returns `auth/{path}/config`, `GetPayload()`, `IsEquivalentToDesiredState()`, `PrepareInternalValues()` (reads keytab from K8s Secret), `IsDeletable()=false`
  - [x] 1.3: Implement `ConditionsAware` interface — Status with `Conditions []metav1.Condition`
  - [x] 1.4: Implement `PrepareInternalValues()` — resolve base64 keytab from K8s Secret (`keytabSecret` reference), store in unexported `retrievedKeytab` field
  - [x] 1.5: Implement `toMap()` on `KerberosAuthConfig` — emit `keytab`, `service_account`, `remove_instance_name`, `add_group_aliases`
  - [x] 1.6: Implement `IsEquivalentToDesiredState()` — must delete `keytab` from desired state (Vault never returns it on read), then `filterPayloadToDesiredKeys`

- [x] Task 2: Create `KerberosAuthEngineLDAPConfig` type (AC: 2, 5, 7, 8, 10)
  - [x] 2.1: Create `api/v1alpha1/kerberosauthengineconfig_ldap_types.go` — Spec with `Connection`, `Authentication`, `Path`, inline `KerberosLDAPConfig` struct, `LDAPCredentials` (`RootCredentialConfig`), `TLSConfig`
  - [x] 2.2: Implement `VaultObject` interface — `GetPath()` returns `auth/{path}/config/ldap`, `IsDeletable()=false`, `PrepareInternalValues()` resolves binddn/bindpass credentials, `PrepareTLSConfig()` resolves certificate from K8s Secret
  - [x] 2.3: Implement `ConditionsAware` interface
  - [x] 2.4: Implement `toMap()` on `KerberosLDAPConfig` — emit all LDAP fields in snake_case; use `durationToSeconds()` for TTL fields, `json.Number` for `token_num_uses`
  - [x] 2.5: Implement `IsEquivalentToDesiredState()` — must delete `bindpass` from desired state (Vault returns empty string on read), then `filterPayloadToDesiredKeys`
  - [x] 2.6: Implement `setInternalCredentials()` — resolve binddn/bindpass from K8s Secret, VaultSecret, or RandomSecret. Follow `LDAPAuthEngineConfig.setInternalCredentials()` pattern
  - [x] 2.7: Implement `setTLSConfig()` — resolve certificate/clientTLSCert/clientTLSKey from K8s Secret. Follow `LDAPAuthEngineConfig.setTLSConfig()` pattern

- [x] Task 3: Create `KerberosAuthEngineGroup` type (AC: 3, 6, 7, 8)
  - [x] 3.1: Create `api/v1alpha1/kerberosauthenginegroup_types.go` — Spec with `Connection`, `Authentication`, `Path`, `Name` (group name override), `Policies` (comma-separated string)
  - [x] 3.2: Implement `VaultObject` interface — `GetPath()` returns `auth/{path}/groups/{name}`, `IsDeletable()=true`, no `PrepareInternalValues` needed
  - [x] 3.3: Implement `ConditionsAware` interface
  - [x] 3.4: Implement `toMap()` — emit `policies` field (string, NOT array — matches LDAPAuthEngineGroup pattern)
  - [x] 3.5: Implement `IsEquivalentToDesiredState()` — standard `filterPayloadToDesiredKeys`

- [x] Task 4: Create webhooks (AC: 8)
  - [x] 4.1: Create `api/v1alpha1/kerberosauthengineconfig_webhook.go` — `admission.Defaulter[*KerberosAuthEngineConfig]`, `admission.Validator[*KerberosAuthEngineConfig]`, immutable `spec.path`
  - [x] 4.2: Create `api/v1alpha1/kerberosauthengineconfig_ldap_webhook.go` — `admission.Defaulter[*KerberosAuthEngineLDAPConfig]`, `admission.Validator[*KerberosAuthEngineLDAPConfig]`, immutable `spec.path`, credential validation via `ValidateCredentialSource()`
  - [x] 4.3: Create `api/v1alpha1/kerberosauthenginegroup_webhook.go` — `admission.Defaulter[*KerberosAuthEngineGroup]`, `admission.Validator[*KerberosAuthEngineGroup]`, immutable `spec.path`/`spec.name`

- [x] Task 5: Create controllers (AC: 1, 2, 3, 4, 5, 6, 7, 9, 10)
  - [x] 5.1: Create `internal/controller/kerberosauthengineconfig_controller.go` — embed `ReconcilerBase`, standard VaultResource reconcile, watches on `corev1.Secret` for keytab rotation
  - [x] 5.2: Create `internal/controller/kerberosauthengineconfig_ldap_controller.go` — embed `ReconcilerBase`, standard VaultResource reconcile, watches on `corev1.Secret` and `RandomSecret` for bind credential rotation
  - [x] 5.3: Create `internal/controller/kerberosauthenginegroup_controller.go` — simple `For()` with default periodic reconcile predicate (no watches needed)

- [x] Task 6: Register in main.go (AC: 1, 2, 3)
  - [x] 6.1: Add controller registrations for all 3 reconcilers
  - [x] 6.2: Add webhook registrations inside `ENABLE_WEBHOOKS` guard for all 3 types

- [x] Task 7: Unit tests (AC: 1, 2, 3, 7, 8)
  - [x] 7.1: Create `api/v1alpha1/kerberosauthengineconfig_test.go` — test `toMap()`, test `IsEquivalentToDesiredState()` with independently constructed Vault-read-shaped fixtures (keytab stripped), negative tests
  - [x] 7.2: Create `api/v1alpha1/kerberosauthengineconfig_ldap_test.go` — test `toMap()`, test `IsEquivalentToDesiredState()` with Vault-read-shaped fixtures (bindpass stripped), negative tests
  - [x] 7.3: Create `api/v1alpha1/kerberosauthenginegroup_test.go` — test `toMap()`, test `IsEquivalentToDesiredState()` with Vault-read fixture, negative tests

- [x] Task 8: Test fixtures (AC: all)
  - [x] 8.1: Create test YAML fixtures in `test/kerberosauthengine/` — config, LDAP config, and group CRs
  - [x] 8.2: Integration tests — SKIP (Kerberos is a network auth protocol that cannot be installed in Kind, falls under "Skip it" per project integration test philosophy)

- [x] Task 9: CRD registration and code generation (AC: all)
  - [x] 9.1: Run `make manifests generate fmt vet test`
  - [x] 9.2: Add 3 new CRD YAML files to `config/crd/kustomization.yaml` (CRD registration checklist)
  - [x] 9.3: Verify all existing tests still pass

- [x] Task 10: Documentation (AC: 11)
  - [x] 10.1: Create `docs/auth-engines/kerberos.md` following `docs/engine-doc-template.md`
  - [x] 10.2: Update `docs/auth-engines/index.md` with link to new doc

## Dev Notes

### Integration Test Classification: SKIP

Per the project's Integration Test Infrastructure Philosophy:
> **Skip it** — For cloud providers and services that cannot be installed in Kind and are hard to mock (e.g., AWS, Azure, GCP, GitHub App, Quay), the integration test must not be run. These types rely on unit test coverage only.

Kerberos requires a KDC (Key Distribution Center), Active Directory or similar infrastructure, and LDAP server — none of which can be trivially installed in Kind. No integration tests. Comprehensive unit tests for `toMap()`, `IsEquivalentToDesiredState()`, and webhook validation are the primary quality gate.

### Epic 15 Retrospective — Mandatory Rules for This Story

1. **TTL/`toMap()`:** All duration fields Vault returns as integer seconds MUST use `durationToSeconds()` (string specs) or `json.Number` (integer specs). Do not copy raw-string patterns.
2. **Credential key defaulting:** If a type remaps `RootCredentialConfig` keys, default only keys omitted from the admission request. Preserve explicit values. Follow `AWSAuthEngineClientConfig.Default()`.
3. **Strict webhook validation:** Reject invalid field combinations at admission.
4. **Kerberos two-part config as two CRDs:** The 1:1 VaultObject contract requires `KerberosAuthEngineConfig` (for `auth/{path}/config`) and `KerberosAuthEngineLDAPConfig` (for `auth/{path}/config/ldap`) as separate CRDs. Same decision as AWS Auth ClientConfig + IdentityConfig.
5. **Keytab via PrepareInternalValues:** The base64 keytab content is sourced from a K8s Secret, not inline in the CR spec.

### Story Intelligence Chain — Previous Story Context

**Story 15.3 (Okta Auth Engine)** is the most recently completed auth engine config + group CRD story:
- Established OktaAuthEngineConfig (single credential via `RootCredentialConfig` with `passwordKey` default) + OktaAuthEngineGroup (follows LDAPAuthEngineGroup exactly)
- Key patterns: `api_token` stripping in `IsEquivalentToDesiredState`, webhook defaulter sets `passwordKey` only when empty, `durationToSeconds()` for all TTL fields after retro fix
- Group `toMap()` does NOT include `name` (Kerberos group read response also only returns `{"policies": [...]}`)

**Story 14.2 (AWS Auth Engine)** established the two-config-CRD split pattern:
- `AWSAuthEngineClientConfig` → `auth/{path}/config/client` and `AWSAuthEngineIdentityConfig` → `auth/{path}/config/identity`
- This is the same architectural decision for Kerberos: two Vault config paths → two CRDs
- Review findings: credential defaults overwrite bug, `json.Number` for numeric fields

**LDAP Auth Engine** (`LDAPAuthEngineConfig` + `LDAPAuthEngineGroup`) is the **closest structural analog** for both the LDAP config CRD and the group CRD:
- `KerberosAuthEngineLDAPConfig` reuses the same LDAP field set (url, binddn, bindpass, groupdn, groupattr, groupfilter, userdn, userattr, TLS settings, token_* fields) — the Kerberos LDAP config endpoint is backed by the same LDAP code as the standalone LDAP auth method
- `KerberosAuthEngineGroup` follows `LDAPAuthEngineGroup` exactly (same Vault API shape: `{"policies": "..."}`)

**Epic 15 Retrospective** completed the following prep items that apply here:
- `durationToSeconds()` propagated to all auth engine types
- Credential key defaulting fix applied (inspect admission request body)
- Strict webhook validation philosophy documented in `project-context.md`

### Design Decision: Three CRDs (Resolved)

Per Epic 15 retrospective and sprint-status action items: **Three separate CRDs** for Kerberos auth:
- `KerberosAuthEngineConfig` → `auth/{path}/config` (keytab + service account + flags)
- `KerberosAuthEngineLDAPConfig` → `auth/{path}/config/ldap` (full LDAP connection settings + token params)
- `KerberosAuthEngineGroup` → `auth/{path}/groups/{name}` (group-to-policy mapping)

This maps 1:1 to Vault's API surface. The epic title says "Config and LDAP Group" but the retro mandates two separate config CRDs (for the two distinct Vault config paths) plus a third group CRD.

### Vault API Paths

| Operation | Method | Path |
|-----------|--------|------|
| Write Kerberos config | POST | `auth/{path}/config` |
| Read Kerberos config | GET | `auth/{path}/config` |
| Write LDAP config | POST | `auth/{path}/config/ldap` |
| Read LDAP config | GET | `auth/{path}/config/ldap` |
| Create/update group | POST | `auth/{path}/groups/{name}` |
| Read group | GET | `auth/{path}/groups/{name}` |
| Delete group | DELETE | `auth/{path}/groups/{name}` |
| List groups | LIST | `auth/{path}/groups` |

### KerberosAuthEngineConfig — Vault API Field Reference

**Write (`POST auth/{path}/config`) fields:**
- `keytab` (string: "") — Base64-encoded keytab file contents for SPNEGO token verification
- `service_account` (string: "") — Service account associated with the keytab entry and LDAP service account
- `remove_instance_name` (bool: false) — Strip instance names from Kerberos service principal names in the keytab
- `add_group_aliases` (bool: false) — Add LDAP groups found for the user as group aliases

**Read (`GET auth/{path}/config`) response — `keytab` is NEVER returned:**
```json
{
  "data": {
    "add_group_aliases": false,
    "remove_instance_name": false,
    "service_account": "vault_svc"
  }
}
```

**Critical:** `keytab` is write-only. Must be deleted from `desiredState` before `IsEquivalentToDesiredState` comparison. Follow `OktaAuthEngineConfig` pattern (deletes `api_token`).

### Critical: `IsEquivalentToDesiredState` for Config — Keytab Stripping

Vault never returns `keytab` on read. The implementation must:
1. Build `desiredState` from `KerberosAuthConfig.toMap()`
2. `delete(desiredState, "keytab")` — remove before comparison
3. Use `filterPayloadToDesiredKeys(desiredState, payload)` → `reflect.DeepEqual`

### KerberosAuthEngineConfig — Keytab Resolution via PrepareInternalValues

The keytab is NOT sourced via `RootCredentialConfig` (it is not a username/password pair). Instead:
- Spec includes `KeytabSecret` (`corev1.LocalObjectReference`) — references a K8s Secret containing the base64-encoded keytab
- `PrepareInternalValues()` reads the Secret, extracts the keytab data from a configurable key (default: `keytab`), and stores it in the unexported `retrievedKeytab` field
- `toMap()` emits `"keytab": retrievedKeytab`

This follows the same pattern as `LDAPAuthEngineConfig.setTLSConfig()` (reads K8s Secret, stores in unexported field), but for keytab content instead of TLS certificates.

The controller watches `corev1.Secret` for keytab rotation detection, following `GCPAuthEngineConfigReconciler.SetupWithManager()`.

### KerberosAuthEngineConfig — Controller Pattern

Standard VaultResource reconcile (NOT always-write). The read endpoint returns `service_account`, `remove_instance_name`, `add_group_aliases` — enough for meaningful drift detection after keytab stripping. The controller MUST include `corev1.Secret` watches for keytab Secret updates.

### KerberosAuthEngineLDAPConfig — Vault API Field Reference

**Write (`POST auth/{path}/config/ldap`) fields:**
- `url` (string: "") — LDAP server URL (e.g., `ldaps://ldap.myorg.com:636`). Comma-separated for multiple servers
- `case_sensitive_names` (bool: false) — Case-sensitive user/group names for policy matching
- `starttls` (bool: false) — Issue StartTLS after establishing unencrypted connection
- `tls_min_version` (string: "tls12") — Minimum TLS version
- `tls_max_version` (string: "tls12") — Maximum TLS version
- `insecure_tls` (bool: false) — Skip LDAP server SSL certificate verification
- `certificate` (string: "") — CA certificate for LDAP server verification (x509 PEM)
- `binddn` (string: "") — Distinguished name for user search bind
- `bindpass` (string: "") — Password for bind DN
- `userdn` (string: "") — Base DN for user search
- `userattr` (string: "") — Attribute matching the username for auth
- `discoverdn` (bool: false) — Use anonymous bind to discover bind DN
- `deny_null_bind` (bool: true) — Prevent bypassing auth with empty password
- `upndomain` (string: "") — userPrincipalDomain for UPN string construction
- `groupfilter` (string: "") — Go template for group membership query
- `groupdn` (string: "") — LDAP search base for group membership
- `groupattr` (string: "") — LDAP attribute for group enumeration (default: `cn`)
- `token_ttl` (integer: 0 or string: "") — Incremental lifetime for generated tokens
- `token_max_ttl` (integer: 0 or string: "") — Maximum lifetime for generated tokens
- `token_policies` (array: [] or string: "") — Token policies
- `policies` (array: [] or string: "", DEPRECATED) — Use `token_policies`
- `token_bound_cidrs` (array: [] or string: "") — CIDR blocks restricting authentication
- `token_explicit_max_ttl` (integer: 0 or string: "") — Hard cap max TTL
- `token_no_default_policy` (bool: false) — Exclude default policy from tokens
- `token_num_uses` (integer: 0) — Max token uses (0 = unlimited)
- `token_period` (integer: 0 or string: "") — Max allowed period for periodic tokens
- `token_type` (string: "") — Token type (service, batch, default, default-service, default-batch)

**Read (`GET auth/{path}/config/ldap`) response — `bindpass` returns empty string:**
```json
{
  "data": {
    "binddn": "cn=vault,ou=Users,dc=example,dc=com",
    "bindpass": "",
    "certificate": "",
    "deny_null_bind": true,
    "discoverdn": false,
    "groupattr": "cn",
    "groupdn": "ou=Groups,dc=example,dc=com",
    "groupfilter": "(&(objectClass=group)(member:1.2.840.113556.1.4.1941:={{.UserDN}}))",
    "insecure_tls": false,
    "starttls": false,
    "tls_max_version": "tls12",
    "tls_min_version": "tls12",
    "upndomain": "",
    "url": "ldaps://ldap.myorg.com:636",
    "userattr": "samaccountname",
    "userdn": "ou=Users,dc=example,dc=com"
  }
}
```

**Critical:** `bindpass` is write-only (Vault returns empty string on read). Must be deleted from `desiredState` before `IsEquivalentToDesiredState` comparison. Follow `LDAPAuthEngineConfig.IsEquivalentToDesiredState()` pattern (deletes `bindpass`).

### KerberosAuthEngineLDAPConfig — Credential Resolution (Bind Password)

The LDAP bind credentials must be resolved using `RootCredentialConfig`, following `LDAPAuthEngineConfig` exactly:
- **K8s Secret**: usernameKey=`username`, passwordKey=`password` (defaults from `RootCredentialConfig`)
- **VaultSecret**: same key mapping from a Vault KV path
- **RandomSecret**: retrieve from RandomSecret's Vault path

The `setInternalCredentials()` implementation follows `LDAPAuthEngineConfig.setInternalCredentials()` exactly — resolve `binddn` and `bindpass` from the configured credential source. If `spec.bindDN` is set directly, it takes precedence over the username retrieved from the Secret.

### KerberosAuthEngineLDAPConfig — TLS Config Resolution

The TLS certificate fields (`certificate`, `clientTLSCert`, `clientTLSKey`) can be sourced from a K8s Secret via the `TLSConfig` field, following `LDAPAuthEngineConfig.setTLSConfig()` and `PrepareTLSConfig()` exactly.

### KerberosAuthEngineLDAPConfig — Controller Pattern

Standard VaultResource reconcile (NOT always-write). The read endpoint returns enough fields for meaningful drift detection (everything except `bindpass`). The controller MUST include `corev1.Secret` and `RandomSecret` watches for credential rotation, following `LDAPAuthEngineConfigReconciler.SetupWithManager()`.

### KerberosAuthEngineGroup — Vault API Field Reference

**Write (`POST auth/{path}/groups/{name}`) fields:**
- `name` (string: required) — Group name (from URL path)
- `policies` (string: "") — Comma-separated list of policies

**Read (`GET auth/{path}/groups/{name}`) response:**
```json
{
  "data": {
    "policies": ["admin", "default"]
  }
}
```

No write-only fields. Standard `filterPayloadToDesiredKeys` is sufficient.

### KerberosAuthEngineGroup — Pattern

Follow `LDAPAuthEngineGroup` exactly:
- Spec: `Connection`, `Authentication`, `Path`, `Name` (group name override), `Policies` (string — comma-separated list)
- `GetPath()`: `auth/{path}/groups/{name}` using `vaultutils.CleansePath`
- `IsDeletable()`: `true`
- `toMap()`: emit `policies` field (as string, matching LDAP group pattern). Note: the Vault read response for Kerberos groups returns `policies` as an array, but `LDAPAuthEngineGroup.toMap()` includes `name` — for Kerberos groups, do NOT include `name` since the Kerberos group read response only returns `{"policies": [...]}` (same behavior as Okta groups)
- `IsEquivalentToDesiredState()`: standard `filterPayloadToDesiredKeys`
- Simple controller: `For()` with default periodic reconcile predicate, no watches

### CRD Field Spec — KerberosAuthConfig

```go
type KerberosAuthConfig struct {
    // ServiceAccount is the service account associated with both the keytab entry and an LDAP service account.
    // +kubebuilder:validation:Required
    ServiceAccount string `json:"serviceAccount"`

    // RemoveInstanceName strips instance names from a Kerberos service principal name when parsing the keytab.
    // +kubebuilder:validation:Optional
    RemoveInstanceName bool `json:"removeInstanceName,omitempty"`

    // AddGroupAliases adds any LDAP groups found for the user as group aliases.
    // +kubebuilder:validation:Optional
    AddGroupAliases bool `json:"addGroupAliases,omitempty"`

    retrievedKeytab string `json:"-"`
}
```

### CRD Field Spec — KerberosAuthEngineConfigSpec (wrapper)

```go
type KerberosAuthEngineConfigSpec struct {
    // +kubebuilder:validation:Optional
    Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

    // +kubebuilder:validation:Required
    Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

    // Path at which the Kerberos auth engine is mounted.
    // +kubebuilder:validation:Required
    Path vaultutils.Path `json:"path,omitempty"`

    KerberosAuthConfig `json:",inline"`

    // KeytabSecret references a Kubernetes Secret containing the base64-encoded keytab file.
    // The Secret must contain a key (default: "keytab") with the base64 keytab content.
    // +kubebuilder:validation:Required
    KeytabSecret corev1.LocalObjectReference `json:"keytabSecret"`

    // KeytabKey is the key within the KeytabSecret that contains the base64-encoded keytab.
    // +kubebuilder:validation:Optional
    // +kubebuilder:default="keytab"
    KeytabKey string `json:"keytabKey"`
}
```

### CRD Field Spec — KerberosLDAPConfig

This struct mirrors the `LDAPConfig` struct from `LDAPAuthEngineConfig` but scoped to the Kerberos LDAP endpoint. The Kerberos LDAP config supports the same fields as the standalone LDAP auth method (the Vault documentation confirms: "The LDAP above relies upon the same code as the LDAP auth method").

Key differences from `LDAPConfig`:
- No `request_timeout`, `userFilter`, `anonymousGroupSearch`, `usernameAsAlias` — these fields are NOT in the Kerberos LDAP API
- Has `alias_metadata` — a `map[string]string` for custom metadata on entity aliases (not present in standalone LDAP auth config)

```go
type KerberosLDAPConfig struct {
    // URL is the LDAP server to connect to. Multiple URLs can be comma-separated.
    // +kubebuilder:validation:Required
    URL string `json:"url"`

    // CaseSensitiveNames if set, user and group names are case sensitive for policy matching.
    // +kubebuilder:validation:Optional
    CaseSensitiveNames bool `json:"caseSensitiveNames,omitempty"`

    // StartTLS issues a StartTLS command after establishing an unencrypted connection.
    // +kubebuilder:validation:Optional
    StartTLS bool `json:"startTLS,omitempty"`

    // TLSMinVersion is the minimum TLS version to use.
    // +kubebuilder:validation:Optional
    // +kubebuilder:default="tls12"
    // +kubebuilder:validation:Enum={"tls10","tls11","tls12","tls13"}
    TLSMinVersion string `json:"tlsMinVersion"`

    // TLSMaxVersion is the maximum TLS version to use.
    // +kubebuilder:validation:Optional
    // +kubebuilder:default="tls12"
    // +kubebuilder:validation:Enum={"tls10","tls11","tls12","tls13"}
    TLSMaxVersion string `json:"tlsMaxVersion"`

    // InsecureTLS skips LDAP server SSL certificate verification.
    // +kubebuilder:validation:Optional
    InsecureTLS bool `json:"insecureTLS,omitempty"`

    // Certificate is the CA certificate for verifying the LDAP server certificate (x509 PEM).
    // +kubebuilder:validation:Optional
    Certificate string `json:"certificate,omitempty"`

    // BindDN is the distinguished name used to bind when performing user search.
    // +kubebuilder:validation:Optional
    BindDN string `json:"bindDN,omitempty"`

    // UserDN is the base DN for user search.
    // +kubebuilder:validation:Optional
    UserDN string `json:"userDN,omitempty"`

    // UserAttr is the attribute on user objects matching the authenticating username.
    // +kubebuilder:validation:Optional
    UserAttr string `json:"userAttr,omitempty"`

    // DiscoverDN uses anonymous bind to discover the bind DN of a user.
    // +kubebuilder:validation:Optional
    DiscoverDN bool `json:"discoverDN,omitempty"`

    // DenyNullBind prevents users from bypassing authentication with an empty password.
    // +kubebuilder:validation:Optional
    // +kubebuilder:default=true
    DenyNullBind bool `json:"denyNullBind"`

    // UPNDomain is the userPrincipalDomain for constructing UPN strings.
    // +kubebuilder:validation:Optional
    UPNDomain string `json:"upnDomain,omitempty"`

    // GroupFilter is a Go template for constructing the group membership query.
    // +kubebuilder:validation:Optional
    GroupFilter string `json:"groupFilter,omitempty"`

    // GroupDN is the LDAP search base for group membership search.
    // +kubebuilder:validation:Optional
    GroupDN string `json:"groupDN,omitempty"`

    // GroupAttr is the LDAP attribute for enumerating user group membership.
    // +kubebuilder:validation:Optional
    GroupAttr string `json:"groupAttr,omitempty"`

    // TokenTTL is the incremental lifetime for generated tokens.
    // +kubebuilder:validation:Optional
    TokenTTL string `json:"tokenTTL,omitempty"`

    // TokenMaxTTL is the maximum lifetime for generated tokens.
    // +kubebuilder:validation:Optional
    TokenMaxTTL string `json:"tokenMaxTTL,omitempty"`

    // TokenPolicies are policies to encode onto generated tokens.
    // +kubebuilder:validation:Optional
    TokenPolicies string `json:"tokenPolicies,omitempty"`

    // TokenBoundCIDRs are CIDR blocks restricting authentication.
    // +kubebuilder:validation:Optional
    TokenBoundCIDRs string `json:"tokenBoundCIDRs,omitempty"`

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

    retrievedPassword string    `json:"-"`
    retrievedUsername string    `json:"-"`
    retrievedCertificate string `json:"-"`
    retrievedClientTLSCert string `json:"-"`
    retrievedClientTLSKey string `json:"-"`
}
```

### CRD Field Spec — KerberosAuthEngineLDAPConfigSpec (wrapper)

```go
type KerberosAuthEngineLDAPConfigSpec struct {
    // +kubebuilder:validation:Optional
    Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

    // +kubebuilder:validation:Required
    Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

    // Path at which the Kerberos auth engine is mounted.
    // +kubebuilder:validation:Required
    Path vaultutils.Path `json:"path,omitempty"`

    KerberosLDAPConfig `json:",inline"`

    // LDAPCredentials is used to connect to the LDAP service.
    // Consists in bindDN and bindPass, sourced from a K8s Secret, VaultSecret, or RandomSecret.
    // +kubebuilder:validation:Required
    LDAPCredentials vaultutils.RootCredentialConfig `json:"ldapCredentials,omitempty"`

    // TLSConfig represents the LDAP service certificate configuration.
    // +kubebuilder:validation:Optional
    TLSConfig vaultutils.TLSConfig `json:"tLSConfig,omitempty"`
}
```

### CRD Field Spec — KerberosAuthEngineGroup

Follow LDAPAuthEngineGroup pattern exactly:

```go
type KerberosAuthEngineGroupSpec struct {
    // +kubebuilder:validation:Optional
    Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

    // +kubebuilder:validation:Required
    Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

    // Path at which the Kerberos auth engine is mounted.
    // +kubebuilder:validation:Required
    Path vaultutils.Path `json:"path,omitempty"`

    // Name of the Kerberos LDAP group.
    // +kubebuilder:validation:Required
    Name string `json:"name,omitempty"`

    // Policies is a comma-separated list of policies associated with the group.
    // +kubebuilder:validation:Optional
    Policies string `json:"policies,omitempty"`
}
```

### `toMap()` Implementation Notes

**KerberosAuthConfig.toMap():**
```go
func (c *KerberosAuthConfig) toMap() map[string]any {
    payload := map[string]any{}
    payload["keytab"] = c.retrievedKeytab
    payload["service_account"] = c.ServiceAccount
    payload["remove_instance_name"] = c.RemoveInstanceName
    payload["add_group_aliases"] = c.AddGroupAliases
    return payload
}
```

**Critical:** `keytab` must come from `retrievedKeytab` (set by `PrepareInternalValues`), NOT from the spec directly. The keytab is never stored in the CR spec — it comes from a K8s Secret reference.

**KerberosLDAPConfig.toMap():**
```go
func (c *KerberosLDAPConfig) toMap() map[string]any {
    payload := map[string]any{}
    payload["url"] = c.URL
    payload["case_sensitive_names"] = c.CaseSensitiveNames
    payload["starttls"] = c.StartTLS
    payload["tls_min_version"] = c.TLSMinVersion
    payload["tls_max_version"] = c.TLSMaxVersion
    payload["insecure_tls"] = c.InsecureTLS
    payload["certificate"] = c.retrievedCertificate
    payload["binddn"] = c.retrievedUsername
    payload["bindpass"] = c.retrievedPassword
    payload["userdn"] = c.UserDN
    payload["userattr"] = c.UserAttr
    payload["discoverdn"] = c.DiscoverDN
    payload["deny_null_bind"] = c.DenyNullBind
    payload["upndomain"] = c.UPNDomain
    payload["groupfilter"] = c.GroupFilter
    payload["groupdn"] = c.GroupDN
    payload["groupattr"] = c.GroupAttr
    payload["token_ttl"] = durationToSeconds(c.TokenTTL)
    payload["token_max_ttl"] = durationToSeconds(c.TokenMaxTTL)
    payload["token_policies"] = c.TokenPolicies
    payload["token_bound_cidrs"] = c.TokenBoundCIDRs
    payload["token_explicit_max_ttl"] = durationToSeconds(c.TokenExplicitMaxTTL)
    payload["token_no_default_policy"] = c.TokenNoDefaultPolicy
    payload["token_num_uses"] = json.Number(strconv.FormatInt(c.TokenNumUses, 10))
    payload["token_period"] = durationToSeconds(c.TokenPeriod)
    payload["token_type"] = c.TokenType
    return payload
}
```

**Critical notes for LDAP config `toMap()`:**
- `binddn` and `bindpass` must come from resolved internal fields (set by `setInternalCredentials`)
- `certificate` comes from `retrievedCertificate` (set by `setTLSConfig`)
- Duration fields (`token_ttl`, `token_max_ttl`, `token_explicit_max_ttl`, `token_period`) MUST use `durationToSeconds()` — mandatory per Epic 15 retro
- `token_num_uses` MUST be `json.Number` — Vault returns integer seconds as `json.Number`
- `token_policies` and `token_bound_cidrs` are emitted as strings (matching `LDAPConfig.toMap()` pattern where these are string fields, NOT arrays)

**KerberosAuthEngineGroup.toMap():**
```go
func (g *KerberosAuthEngineGroup) toMap() map[string]any {
    payload := map[string]any{}
    payload["policies"] = g.Spec.Policies
    return payload
}
```

Do NOT include `name` in `toMap()` — the Kerberos group read response only returns `{"policies": [...]}`, same as Okta groups and unlike LDAP groups.

### Kubebuilder Markers Checklist

- All exported Spec fields: `+kubebuilder:validation:Required` or `+kubebuilder:validation:Optional`
- Config `ServiceAccount`: `+kubebuilder:validation:Required` (no default, no omitempty)
- Config `KeytabSecret`: `+kubebuilder:validation:Required`
- Config `KeytabKey`: `+kubebuilder:default="keytab"` (non-zero default → no `omitempty`)
- LDAP Config `URL`: `+kubebuilder:validation:Required`
- LDAP Config `TLSMinVersion`/`TLSMaxVersion`: `+kubebuilder:default="tls12"`, `+kubebuilder:validation:Enum` (non-zero default → no `omitempty`)
- LDAP Config `DenyNullBind`: `+kubebuilder:default=true` (non-zero default → no `omitempty`)
- LDAP Config `TokenType`: `+kubebuilder:validation:Enum={"service","batch","default","default-service","default-batch"}`
- LDAP Config `TokenNumUses`: `+kubebuilder:validation:Minimum=0`
- Root types: `+kubebuilder:object:root=true`, `+kubebuilder:subresource:status`
- Conditions field: `+patchMergeKey=type`, `+patchStrategy=merge`, `+listType=map`, `+listMapKey=type`
- Group `Name`: `+kubebuilder:validation:Required`

### RBAC Markers

Config controller:
```go
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=kerberosauthengineconfigs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=kerberosauthengineconfigs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=kerberosauthengineconfigs/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=serviceaccounts/token,verbs=create
//+kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;patch
```

LDAP config controller:
```go
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=kerberosauthengineldapconfigs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=kerberosauthengineldapconfigs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=kerberosauthengineldapconfigs/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=serviceaccounts/token,verbs=create
//+kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;patch
```

**Note:** The actual plural resource name for `KerberosAuthEngineLDAPConfig` is auto-generated by controller-gen. Run `make manifests` to verify the exact plural form matches. Expected: `kerberosauthengineldapconfigs`.

Group controller:
```go
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=kerberosauthenginegroups,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=kerberosauthenginegroups/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=kerberosauthenginegroups/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=serviceaccounts/token,verbs=create
//+kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;patch
```

### File Structure

| File | Action | Description |
|------|--------|-------------|
| `api/v1alpha1/kerberosauthengineconfig_types.go` | NEW | Config CRD type, VaultObject, ConditionsAware, toMap, keytab resolution from K8s Secret |
| `api/v1alpha1/kerberosauthengineconfig_ldap_types.go` | NEW | LDAP Config CRD type, VaultObject, ConditionsAware, toMap, credential + TLS resolution |
| `api/v1alpha1/kerberosauthenginegroup_types.go` | NEW | Group CRD type, VaultObject, ConditionsAware, toMap |
| `api/v1alpha1/kerberosauthengineconfig_webhook.go` | NEW | Config webhook — defaulter, validator, immutable path |
| `api/v1alpha1/kerberosauthengineconfig_ldap_webhook.go` | NEW | LDAP Config webhook — defaulter, validator, immutable path, credential validation |
| `api/v1alpha1/kerberosauthenginegroup_webhook.go` | NEW | Group webhook — defaulter, validator, immutable path/name |
| `api/v1alpha1/kerberosauthengineconfig_test.go` | NEW | Unit tests for config toMap, IsEquivalentToDesiredState |
| `api/v1alpha1/kerberosauthengineconfig_ldap_test.go` | NEW | Unit tests for LDAP config toMap, IsEquivalentToDesiredState |
| `api/v1alpha1/kerberosauthenginegroup_test.go` | NEW | Unit tests for group toMap, IsEquivalentToDesiredState |
| `internal/controller/kerberosauthengineconfig_controller.go` | NEW | Config reconciler with Secret watches |
| `internal/controller/kerberosauthengineconfig_ldap_controller.go` | NEW | LDAP Config reconciler with Secret/RandomSecret watches |
| `internal/controller/kerberosauthenginegroup_controller.go` | NEW | Group reconciler — simple VaultResource |
| `cmd/main.go` | UPDATE | Register 3 controllers + 3 webhooks |
| `config/crd/kustomization.yaml` | UPDATE | Add 3 new CRD YAML files to resources list |
| `test/kerberosauthengine/` | NEW | Test YAML fixtures for all 3 types |
| `docs/auth-engines/kerberos.md` | NEW | Engine documentation per DNFR5 |
| `docs/auth-engines/index.md` | UPDATE | Add link to kerberos.md |

### Files Being Modified — Current State

**`cmd/main.go`**: Currently registers ~48+ controllers and ~48+ webhooks (including Epic 15 additions). New registrations follow the exact same pattern:
- Controller: `(&controller.KerberosAuthEngineConfigReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "KerberosAuthEngineConfig")}).SetupWithManager(mgr)`
- Webhook (inside ENABLE_WEBHOOKS): `(&redhatcopv1alpha1.KerberosAuthEngineConfig{}).SetupWebhookWithManager(mgr)`

No existing behavior is changed — this is purely additive.

**`config/crd/kustomization.yaml`**: Add the 3 new CRD YAML files to the `resources` list. Required for Helm chart build.

**`docs/auth-engines/index.md`**: Add a row to the Supported Auth Engines table:
```
| Kerberos | KerberosAuthEngineConfig | KerberosAuthEngineLDAPConfig | KerberosAuthEngineGroup | [kerberos.md](kerberos.md) |
```

### Existing Patterns to Follow (Reference Files)

| Pattern | Reference File |
|---------|---------------|
| Auth engine config IsDeletable=false | `api/v1alpha1/gcpauthengineconfig_types.go` |
| Auth engine config GetPath (auth/{path}/config) | `api/v1alpha1/gcpauthengineconfig_types.go` |
| Two-config-CRD split (different Vault endpoints) | `api/v1alpha1/awsauthengineconfig_types.go` + `awsauthengineidentityconfig_types.go` |
| Write-only field stripping (keytab / api_token) | `api/v1alpha1/oktaauthengineconfig_types.go` (deletes `api_token`) |
| LDAP config with credential resolution + TLS | `api/v1alpha1/ldapauthengineconfig_types.go` |
| LDAP group with Name override | `api/v1alpha1/ldapauthenginegroup_types.go` |
| Auth group GetPath (auth/{path}/groups/{name}) | `api/v1alpha1/ldapauthenginegroup_types.go` |
| Bindpass stripping in IsEquivalentToDesiredState | `api/v1alpha1/ldapauthengineconfig_types.go` (deletes `bindpass`) |
| Config controller with Secret watches | `internal/controller/gcpauthengineconfig_controller.go` |
| Config controller with Secret+RandomSecret watches | `internal/controller/ldapauthengineconfig_controller.go` |
| Simple group controller (no watches) | `internal/controller/ldapauthenginegroup_controller.go` |
| Auth config webhook (immutable path only) | `api/v1alpha1/gcpauthengineconfig_webhook.go` |
| Auth group webhook (immutable path+name) | `api/v1alpha1/ldapauthenginegroup_webhook.go` |
| LDAP credential webhook with ValidateCredentialSource | `api/v1alpha1/ldapauthengineconfig_webhook.go` |
| durationToSeconds for TTL fields | `api/v1alpha1/ldapauthengineconfig_types.go` (toMap) |
| json.Number for integer fields | `api/v1alpha1/oktaauthengineconfig_types.go` (token_num_uses) |
| K8s Secret content resolution in PrepareInternalValues | `api/v1alpha1/ldapauthengineconfig_types.go` (setTLSConfig) |
| filterPayloadToDesiredKeys | `api/v1alpha1/payload_filter.go` |
| Documentation template | `docs/engine-doc-template.md` |

### Unit Test Requirements

**Config tests (`kerberosauthengineconfig_test.go`):**
1. `TestKerberosAuthEngineConfig_toMap` — verify `keytab` from resolved internal field, `service_account`, `remove_instance_name`, `add_group_aliases`
2. `TestKerberosAuthEngineConfig_IsEquivalentToDesiredState_Match` — construct Vault-read-shaped payload (without `keytab`, with `service_account`, `remove_instance_name`, `add_group_aliases`), verify returns `true`
3. `TestKerberosAuthEngineConfig_IsEquivalentToDesiredState_Mismatch` — change `service_account`, verify returns `false`
4. `TestKerberosAuthEngineConfig_IsEquivalentToDesiredState_KeytabStripping` — verify that a Vault payload without `keytab` still matches when desired state would have it (proves stripping works)

**LDAP config tests (`kerberosauthengineconfig_ldap_test.go`):**
1. `TestKerberosAuthEngineLDAPConfig_toMap` — verify all fields in snake_case, verify `binddn`/`bindpass` from resolved internal fields, verify `token_ttl`/`token_max_ttl`/`token_explicit_max_ttl`/`token_period` use `durationToSeconds`, verify `token_num_uses` is `json.Number`
2. `TestKerberosAuthEngineLDAPConfig_IsEquivalentToDesiredState_Match` — construct Vault-read-shaped payload (with `bindpass` as empty string, integer TTLs as `json.Number`), verify returns `true`
3. `TestKerberosAuthEngineLDAPConfig_IsEquivalentToDesiredState_Mismatch` — change `url`, verify returns `false`
4. `TestKerberosAuthEngineLDAPConfig_IsEquivalentToDesiredState_BindpassStripping` — verify that a Vault payload with empty `bindpass` still matches when desired state would have it (proves stripping works)
5. `TestKerberosAuthEngineLDAPConfig_IsEquivalentToDesiredState_ExtraVaultFields` — add extra Vault-returned field, verify still returns `true` after filtering

**Group tests (`kerberosauthenginegroup_test.go`):**
1. `TestKerberosAuthEngineGroup_toMap` — verify `policies` field matches spec
2. `TestKerberosAuthEngineGroup_IsEquivalentToDesiredState_Match` — Vault-read fixture with `policies` as string, verify returns `true`
3. `TestKerberosAuthEngineGroup_IsEquivalentToDesiredState_Mismatch` — change `policies`, verify returns `false`

### Anti-Patterns / DO NOT

- **DO NOT** create integration tests for these types — Kerberos requires KDC/AD infrastructure that cannot be installed in Kind (per "Skip it" rule)
- **DO NOT** merge the two config endpoints into a single CRD — the design decision is three separate CRDs for clean 1:1 Vault API mapping
- **DO NOT** put keytab content directly in the CR spec — keytab content MUST be sourced from a K8s Secret via `PrepareInternalValues`, never inline
- **DO NOT** use `RootCredentialConfig` for keytab resolution — keytab is binary (base64) content from a Secret, not a username/password pair. Use a direct `corev1.LocalObjectReference` + key approach
- **DO NOT** modify shared framework behavior (reconcile_skeleton.go, vaultresourcereconciler.go, etc.) — only add new files
- **DO NOT** derive expected payloads from `toMap()` in unit tests — construct independent Vault-read-shaped fixtures
- **DO NOT** forget to add new CRD YAML files to `config/crd/kustomization.yaml` after `make manifests`
- **DO NOT** use Go `int` or `float64` in unit test Vault payloads — use `json.Number` to match real Vault client behavior
- **DO NOT** emit raw duration strings for TTL fields — use `durationToSeconds()` (Epic 15 retro mandate)
- **DO NOT** include `name` in the Kerberos group `toMap()` output — the Kerberos group read response only returns `{"policies": [...]}` (same as Okta, unlike LDAP which includes `name`)
- **DO NOT** include `keytab` in the Vault-read fixture for Config `IsEquivalentToDesiredState` tests — Vault never returns it
- **DO NOT** include `bindpass` (with a non-empty value) in the Vault-read fixture for LDAP Config `IsEquivalentToDesiredState` tests — Vault returns empty string for `bindpass`
- **DO NOT** confuse `KerberosAuthEngineLDAPConfig` (Kerberos auth LDAP config at `auth/{path}/config/ldap`) with `LDAPAuthEngineConfig` (standalone LDAP auth at `auth/{path}/config`) — different auth engines with overlapping LDAP field sets
- **DO NOT** reuse the `LDAPConfig` struct from `ldapauthengineconfig_types.go` — the Kerberos LDAP endpoint has a different field set (no `request_timeout`, `userFilter`, `anonymousGroupSearch`, `usernameAsAlias`; has `alias_metadata`). Create a new `KerberosLDAPConfig` struct

- **Novelty Risk:** HIGH — Two Vault config paths mapped to two CRDs (novel for this project: only AWS Auth did this before), plus keytab content resolution from K8s Secret via `PrepareInternalValues` (similar to TLS config resolution but for binary keytab content, not certificates). Groups CRD is routine (copy of LDAP/Okta group pattern).
- Duration/TTL fields in `toMap()` must use `durationToSeconds()` or `json.Number` (Vault-read format). Never emit raw duration strings.

### What Is Novel vs Copy-Paste

| Component | Classification | Notes |
|-----------|---------------|-------|
| KerberosAuthEngineConfig type structure | Copy-paste | Follows OktaAuthEngineConfig (config at `auth/{path}/config`, IsDeletable=false) |
| Keytab resolution via PrepareInternalValues | **NOVEL** | Reads base64 keytab content from K8s Secret. Similar to `LDAPAuthEngineConfig.setTLSConfig()` but for keytab binary data, not TLS certificates. No existing type does exactly this |
| Keytab stripping in IsEquivalentToDesiredState | Copy-paste | Same as Okta api_token / AWS secret_key / LDAP bindpass stripping |
| Config controller with Secret watch (keytab rotation) | Copy-paste | Follows GCPAuthEngineConfig controller (Secret watches, no RandomSecret for keytab) |
| KerberosAuthEngineLDAPConfig type structure | Copy-paste | Follows LDAPAuthEngineConfig (same LDAP fields, credential resolution, TLS config) |
| LDAP Config controller with Secret+RandomSecret watches | Copy-paste | Follows LDAPAuthEngineConfigReconciler |
| KerberosAuthEngineGroup type + controller | Copy-paste | Exact copy of LDAPAuthEngineGroup / OktaAuthEngineGroup pattern |
| Two-CRD config split architecture | Copy-paste | Same decision as AWS Auth ClientConfig + IdentityConfig |
| Webhooks | Copy-paste | Standard immutable-path pattern for configs, immutable-path/name for groups, ValidateCredentialSource for LDAP config |

### Project Structure Notes

- All new files follow existing naming conventions: lowercase type name for files
- New files go in established directories (`api/v1alpha1/`, `internal/controller/`, `test/`, `docs/auth-engines/`)
- Test fixture directory `test/kerberosauthengine/` follows the existing pattern (`test/gcpauthengine/`, `test/ldapauthengine/`, `test/oktaauthengine/`)
- No conflicts with existing code — purely additive
- The LDAP config file uses `_ldap_` infix to distinguish from the main config: `kerberosauthengineconfig_ldap_types.go` (analogous to how AWS used `awsauthengineidentityconfig_types.go` as a separate file)

### References

- [Source: _bmad-output/planning-artifacts/epics.md#Epic-16, Story 16.4 — requirements and acceptance criteria]
- [Source: _bmad-output/project-context.md — all project rules, patterns, and conventions]
- [Source: _bmad-output/implementation-artifacts/epic-15-retro-2026-08-18.md — TTL, credential defaults, webhook validation, Kerberos two-CRD mandate]
- [Source: _bmad-output/implementation-artifacts/14-2-aws-auth-engine-config-and-role-crds.md — two-config-CRD split pattern]
- [Source: _bmad-output/implementation-artifacts/15-3-okta-auth-engine-config-and-group-crds.md — config + group pattern, api_token stripping]
- [Source: api/v1alpha1/ldapauthengineconfig_types.go — LDAP config with credential resolution, TLS config, bindpass stripping, toMap with durationToSeconds]
- [Source: api/v1alpha1/ldapauthenginegroup_types.go — LDAP group with Name override, policies string, toMap]
- [Source: api/v1alpha1/oktaauthengineconfig_types.go — Okta auth config with api_token stripping, durationToSeconds, json.Number]
- [Source: api/v1alpha1/gcpauthengineconfig_types.go — GCP auth config IsDeletable=false, Secret watches]
- [Source: api/v1alpha1/awsauthengineconfig_types.go — AWS auth config with RootCredentialConfig, two-CRD split]
- [Source: internal/controller/ldapauthengineconfig_controller.go — LDAP config controller with Secret/RandomSecret watches]
- [Source: internal/controller/ldapauthenginegroup_controller.go — LDAP group controller (simple)]
- [Source: docs/engine-doc-template.md — documentation template]
- [Source: Vault Kerberos Auth Method API — https://developer.hashicorp.com/vault/api-docs/auth/kerberos]

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6

### Debug Log References

### Completion Notes List

- Implemented all three CRDs (KerberosAuthEngineConfig, KerberosAuthEngineLDAPConfig, KerberosAuthEngineGroup) following established patterns
- KerberosAuthEngineConfig: keytab resolved via PrepareInternalValues from K8s Secret, keytab stripped in IsEquivalentToDesiredState, IsDeletable=false
- KerberosAuthEngineLDAPConfig: credentials via setInternalCredentials (K8s Secret, VaultSecret, RandomSecret), TLS via setTLSConfig, bindpass stripped in IsEquivalentToDesiredState, durationToSeconds for TTLs, json.Number for token_num_uses, IsDeletable=false
- KerberosAuthEngineGroup: simple toMap with policies only (no name), IsDeletable=true, follows Okta group pattern
- Config controller watches corev1.Secret for keytab rotation; LDAP controller watches Secret + RandomSecret for credential rotation; Group controller has no watches
- Webhooks enforce immutable spec.path (all three) and immutable spec.name (group only); LDAP config validates credential source via ValidateCredentialSource
- All unit tests pass: toMap, IsEquivalentToDesiredState (match, mismatch, keytab/bindpass stripping, extra vault fields)
- CRDs registered in config/crd/kustomization.yaml; controllers and webhooks registered in main.go
- Documentation created at docs/auth-engines/kerberos.md and index.md updated

### File List

- api/v1alpha1/kerberosauthengineconfig_types.go (NEW)
- api/v1alpha1/kerberosauthengineconfig_ldap_types.go (NEW)
- api/v1alpha1/kerberosauthenginegroup_types.go (NEW)
- api/v1alpha1/kerberosauthengineconfig_webhook.go (NEW)
- api/v1alpha1/kerberosauthengineconfig_ldap_webhook.go (NEW)
- api/v1alpha1/kerberosauthenginegroup_webhook.go (NEW)
- api/v1alpha1/kerberosauthengineconfig_test.go (NEW)
- api/v1alpha1/kerberosauthengineconfig_ldap_test.go (NEW)
- api/v1alpha1/kerberosauthenginegroup_test.go (NEW)
- api/v1alpha1/zz_generated.deepcopy.go (MODIFIED - auto-generated)
- internal/controller/kerberosauthengineconfig_controller.go (NEW)
- internal/controller/kerberosauthengineconfig_ldap_controller.go (NEW)
- internal/controller/kerberosauthenginegroup_controller.go (NEW)
- cmd/main.go (MODIFIED)
- config/crd/bases/redhatcop.redhat.io_kerberosauthengineconfigs.yaml (NEW - auto-generated)
- config/crd/bases/redhatcop.redhat.io_kerberosauthengineldapconfigs.yaml (NEW - auto-generated)
- config/crd/bases/redhatcop.redhat.io_kerberosauthenginegroups.yaml (NEW - auto-generated)
- config/crd/kustomization.yaml (MODIFIED)
- config/rbac/role.yaml (MODIFIED - auto-generated)
- test/kerberosauthengine/kerberosauthengineconfig.yaml (NEW)
- test/kerberosauthengine/kerberosauthengineldapconfig.yaml (NEW)
- test/kerberosauthengine/kerberosauthenginegroup.yaml (NEW)
- docs/auth-engines/kerberos.md (NEW)
- docs/auth-engines/index.md (MODIFIED)

## Code Review Record

### Review Model Used

gpt-5.4-medium

### Review Findings

- [x] [Review][Patch] LDAP credential Secret watch filters the wrong Secret type and the wrong change condition [internal/controller/kerberosauthengineconfig_ldap_controller.go:79] — The predicate excludes `kubernetes.io/basic-auth` Secrets and returns `bytes.Equal(...)` instead of change detection, so real bind-credential rotations are missed while unrelated or TLS Secret updates can enqueue reconciles spuriously, violating AC 10.
- [x] [Review][Patch] Kerberos group drift detection does not normalize Vault's array-shaped `policies` response [api/v1alpha1/kerberosauthenginegroup_types.go:70] — Vault reads group policies back as an array, but `IsEquivalentToDesiredState()` compares that payload directly against the CR's comma-delimited string, which will produce perpetual false drift and repeated writes.
- [x] [Review][Patch] Kerberos LDAP config omits required `clientTLSCert` / `clientTLSKey` / `aliasMetadata` support [api/v1alpha1/kerberosauthengineconfig_ldap_types.go:60] — The new struct and `toMap()` only model `certificate`, so the CRD cannot express the full Kerberos LDAP API surface described in the story and Vault docs, including client certificate authentication and alias metadata.
- [x] [Review][Patch] Keytab Secret watch ignores custom `keytabKey` [internal/controller/kerberosauthengineconfig_controller.go:86] — The update predicate hard-codes `Data["keytab"]`, so CRs using a non-default `spec.keytabKey` never reconcile on Secret rotation, violating AC 9 for supported custom key names.

### Decisions Needed / Decisions Taken

- Design decision (pre-resolved): Three separate CRDs (KerberosAuthEngineConfig + KerberosAuthEngineLDAPConfig + KerberosAuthEngineGroup) for 1:1 Vault API mapping — per Epic 15 retro mandate
- Design decision (pre-resolved): Keytab sourced from K8s Secret via PrepareInternalValues, not via RootCredentialConfig (keytab is binary data, not a username/password pair)
- Design decision (pre-resolved): Both config CRDs have `IsDeletable()=false` — no DELETE endpoints exist for `auth/{path}/config` or `auth/{path}/config/ldap`
- Design decision: KerberosLDAPConfig is a new struct (NOT reusing LDAPConfig from ldapauthengineconfig_types.go) because the Kerberos LDAP endpoint has a different field set
- Decision taken: none required — all fixes are straightforward corrections

### Fixes Applied

1. **LDAP Secret watch predicate (HIGH)**: Removed `kubernetes.io/basic-auth` type exclusion from both UpdateFunc and CreateFunc. Negated `bytes.Equal` calls so reconciliation triggers on credential CHANGE (`!bytes.Equal`), not on unchanged data. [internal/controller/kerberosauthengineconfig_ldap_controller.go]
2. **Group policies drift detection (HIGH)**: Added `normalizePoliciesField()` helper that converts both CR comma-delimited strings and Vault `[]interface{}` arrays to sorted comma-separated strings before DeepEqual comparison. Updated `IsEquivalentToDesiredState` to call it on both sides. Added three unit tests covering Vault-shaped array match, reordered array match, and array mismatch. [api/v1alpha1/kerberosauthenginegroup_types.go, api/v1alpha1/kerberosauthenginegroup_test.go]
3. **KerberosLDAPConfig missing fields (MEDIUM)**: Added `ClientTLSCert`, `ClientTLSKey`, and `AliasMetadata` spec fields to `KerberosLDAPConfig` struct. Updated `toMap()` to emit `client_tls_cert`, `client_tls_key`, and `alias_metadata`. Updated `setTLSConfig()` to copy spec-level TLS cert/key when no TLSSecret is configured (following LDAPAuthEngineConfig pattern). Added `client_tls_key` to write-only stripping in `IsEquivalentToDesiredState`. Updated all tests. [api/v1alpha1/kerberosauthengineconfig_ldap_types.go, api/v1alpha1/kerberosauthengineconfig_ldap_test.go]
4. **Keytab Secret watch custom key (MEDIUM)**: Replaced hard-coded `Data["keytab"]` comparison with `reflect.DeepEqual(oldSecret.Data, newSecret.Data)` so any data change in the keytab Secret triggers reconciliation regardless of which key stores the keytab. The reconciler's `setKeytabFromSecret` already reads the correct key from `spec.keytabKey`. [internal/controller/kerberosauthengineconfig_controller.go]

### Review Findings

- [x] [Review][Patch] Kerberos LDAP Secret watch still misses custom credential keys and TLS Secret rotations [internal/controller/kerberosauthengineconfig_ldap_controller.go:76] — The prior inversion bug is fixed, but the reconciler still only compares `Data["username"]` and `Data["password"]` in its sole Secret predicate while `findApplicableKLDAPForSecret()` also maps TLS Secrets. That means credentials using non-default `usernameKey`/`passwordKey` and TLS-only changes to `ca.crt`/`tls.crt`/`tls.key` will not enqueue reconciles, so AC 10 and the Secret-backed TLS support remain incomplete.
- [x] [Review][Patch] `aliasMetadata` is still modeled with the wrong shape [api/v1alpha1/kerberosauthengineconfig_ldap_types.go:101] — The story spec requires Kerberos LDAP `alias_metadata` to be a `map[string]string`, but the CRD type, generated manifest, and tests all implement it as a single string. Any user who sets alias metadata cannot express the Vault API's expected payload shape, and drift checks for that field will be incorrect.

### Fixes Applied (Iteration 3)

1. **LDAP Secret watch predicate (HIGH)**: Replaced `isBasicAuthSecret` predicate that only compared `Data["username"]` and `Data["password"]` with `isSecretDataChanged` predicate using `reflect.DeepEqual(oldSecret.Data, newSecret.Data)` — same approach as the keytab watch. This triggers reconciliation on any data change in credential or TLS Secrets, regardless of key names. Create/Delete enqueueing preserved as in sibling controllers. [internal/controller/kerberosauthengineconfig_ldap_controller.go]
2. **AliasMetadata type correction (MEDIUM)**: Changed `AliasMetadata` field type from `string` to `map[string]string` in `KerberosLDAPConfig` struct. Updated `toMap()` to only emit `alias_metadata` when the map is non-empty (omitempty semantics). Regenerated deepcopy via `make generate` (now uses `make(map[string]string, ...)` copy). Regenerated CRD manifests via `make manifests` (now `type: object` with `additionalProperties: type: string`). Updated test to use map literal and type-assert the result. [api/v1alpha1/kerberosauthengineconfig_ldap_types.go, api/v1alpha1/kerberosauthengineconfig_ldap_test.go, api/v1alpha1/zz_generated.deepcopy.go, config/crd/bases/redhatcop.redhat.io_kerberosauthengineldapconfigs.yaml]

### Review Findings (Iteration 4)

- [x] [Review][Patch] `toMap()` emits `alias_metadata` as `map[string]string`, but Vault reads it as `map[string]any`, so `DeepEqual` always fails when `aliasMetadata` is set [api/v1alpha1/kerberosauthengineconfig_ldap_types.go:386]

### Fixes Applied (Iteration 4)

1. **AliasMetadata toMap() type mismatch (MEDIUM)**: Changed `toMap()` to emit `alias_metadata` via `toAnyMapString(c.AliasMetadata)` (existing helper from `sshsecretenginerole_types.go`) so the payload is `map[string]any`, matching the shape Vault returns after JSON deserialization. Updated existing test to assert `map[string]any` type. Added two new unit tests: `TestKerberosAuthEngineLDAPConfig_AliasMetadata_DeepEqualWithVaultPayload` (verifies `reflect.DeepEqual` succeeds between `toMap()` output and a Vault-shaped `map[string]any`) and `TestKerberosAuthEngineLDAPConfig_IsEquivalentToDesiredState_WithAliasMetadata` (end-to-end drift check with aliasMetadata set). [api/v1alpha1/kerberosauthengineconfig_ldap_types.go, api/v1alpha1/kerberosauthengineconfig_ldap_test.go]

Status: review
