# Story 12.3: LDAP/AD Secret Engine — Config and Role CRDs

Status: ready-for-dev

## Story

As an operator developer,
I want CRDs for LDAPSecretEngineConfig, LDAPSecretEngineStaticRole, and LDAPSecretEngineDynamicRole,
So that Vault's LDAP/AD secret engine (password rotation, dynamic credentials) can be managed declaratively.

## Acceptance Criteria

1. **Given** an LDAPSecretEngineConfig CR is created with LDAP bind credentials (via K8s Secret reference) **When** the reconciler processes it **Then** the LDAP config is written to Vault at `{path}/config` and ReconcileSuccessful=True

2. **Given** an LDAPSecretEngineStaticRole CR is created with username, dn, and rotation_period **When** the reconciler processes it **Then** the static role exists in Vault at `{path}/static-role/{name}` for managed password rotation and ReconcileSuccessful=True

3. **Given** an LDAPSecretEngineDynamicRole CR is created with creation_ldif and deletion_ldif **When** the reconciler processes it **Then** the dynamic role exists in Vault at `{path}/role/{name}` and ReconcileSuccessful=True

4. **Given** the LDAPSecretEngineConfig CR is deleted **When** the reconciler processes deletion **Then** the Vault config is deleted via `DELETE {path}/config` and the CR is removed (`IsDeletable=true`)

5. **Given** the LDAPSecretEngineStaticRole or LDAPSecretEngineDynamicRole CR is deleted **When** the reconciler processes deletion **Then** the role is removed from Vault and the CR is deleted from K8s

6. **Given** any LDAP secret engine CR spec is updated (e.g., schema changed, rotation_period changed) **When** the reconciler processes the update **Then** the Vault resource reflects the updated values

7. **Given** any LDAP secret engine CR is created or updated **When** the webhook validates it **Then** `spec.path` immutability is enforced on updates, `spec.name` immutability is enforced on updates, and credential source validation passes for config

## Tasks / Subtasks

- [ ] Task 1: Create `LDAPSecretEngineConfig` type (AC: 1, 4, 6, 7)
  - [ ] 1.1: Create `api/v1alpha1/ldapsecretengineconfig_types.go` — Spec with `Connection`, `Authentication`, `Path`, inline `LDAPSEConfig` struct, `BindCredentials` (RootCredentialConfig), `Name`
  - [ ] 1.2: Implement `VaultObject` interface — `GetPath()` returns `{path}/config`, `GetPayload()`, `IsEquivalentToDesiredState()`, `PrepareInternalValues()`, `IsDeletable()=true`
  - [ ] 1.3: Implement `ConditionsAware` interface — Status with `Conditions []metav1.Condition`
  - [ ] 1.4: Implement `setInternalCredentials()` — resolve binddn/bindpass from K8s Secret, VaultSecret, or RandomSecret (follow LDAPAuthEngineConfig pattern)
  - [ ] 1.5: Implement `toMap()` on `LDAPSEConfig` — convert to Vault API snake_case fields
  - [ ] 1.6: Implement `IsEquivalentToDesiredState()` — must delete `bindpass` from desired state (Vault never returns it on read), then `removeUnsetFields` + `filterPayloadToDesiredKeys`

- [ ] Task 2: Create `LDAPSecretEngineStaticRole` type (AC: 2, 5, 6)
  - [ ] 2.1: Create `api/v1alpha1/ldapsecretenginestaticrole_types.go` — Spec with `Connection`, `Authentication`, `Path`, inline `LDAPSEStaticRole` struct, `Name`
  - [ ] 2.2: Implement `VaultObject` interface — `GetPath()` returns `{path}/static-role/{name}`, `IsDeletable()=true`
  - [ ] 2.3: Implement `ConditionsAware` interface
  - [ ] 2.4: Implement `toMap()` on `LDAPSEStaticRole` — handle `rotation_period` as `json.Number`
  - [ ] 2.5: Implement `IsEquivalentToDesiredState()` — use `removeUnsetFields` + `filterPayloadToDesiredKeys` (Vault returns `last_vault_rotation`/`next_vault_rotation` extras that must be filtered)

- [ ] Task 3: Create `LDAPSecretEngineDynamicRole` type (AC: 3, 5, 6)
  - [ ] 3.1: Create `api/v1alpha1/ldapsecretengine_dynamicrole_types.go` — Spec with `Connection`, `Authentication`, `Path`, inline `LDAPSEDynamicRole` struct, `Name`
  - [ ] 3.2: Implement `VaultObject` interface — `GetPath()` returns `{path}/role/{name}`, `IsDeletable()=true`
  - [ ] 3.3: Implement `ConditionsAware` interface
  - [ ] 3.4: Implement `toMap()` on `LDAPSEDynamicRole`
  - [ ] 3.5: Implement `IsEquivalentToDesiredState()` — standard `removeUnsetFields` + `filterPayloadToDesiredKeys`

- [ ] Task 4: Create webhooks (AC: 7)
  - [ ] 4.1: Create `api/v1alpha1/ldapsecretengineconfig_webhook.go` — `admission.Defaulter[*LDAPSecretEngineConfig]`, `admission.Validator[*LDAPSecretEngineConfig]`, immutable `spec.path`/`spec.name`, credential validation via `ValidateCredentialSource()`
  - [ ] 4.2: Create `api/v1alpha1/ldapsecretenginestaticrole_webhook.go` — `admission.Defaulter[*LDAPSecretEngineStaticRole]`, `admission.Validator[*LDAPSecretEngineStaticRole]`, immutable `spec.path`/`spec.name`
  - [ ] 4.3: Create `api/v1alpha1/ldapsecretengine_dynamicrole_webhook.go` — `admission.Defaulter[*LDAPSecretEngineDynamicRole]`, `admission.Validator[*LDAPSecretEngineDynamicRole]`, immutable `spec.path`/`spec.name`

- [ ] Task 5: Create controllers (AC: 1, 2, 3, 4, 5, 6)
  - [ ] 5.1: Create `internal/controller/ldapsecretengineconfig_controller.go` — embed `ReconcilerBase`, always-write reconcile logic (bindpass is write-only like AWS secret_key), watches on `corev1.Secret` and `RandomSecret`
  - [ ] 5.2: Create `internal/controller/ldapsecretenginestaticrole_controller.go` — standard `For()` with default periodic reconcile predicate
  - [ ] 5.3: Create `internal/controller/ldapsecretengine_dynamicrole_controller.go` — standard `For()` with default periodic reconcile predicate

- [ ] Task 6: Register in main.go (AC: 1, 2, 3)
  - [ ] 6.1: Add controller registrations for all three reconcilers
  - [ ] 6.2: Add webhook registrations inside `ENABLE_WEBHOOKS` guard for all three types

- [ ] Task 7: Unit tests (AC: 1, 2, 3, 6, 7)
  - [ ] 7.1: Create `api/v1alpha1/ldapsecretengineconfig_test.go` — test `toMap()`, test `IsEquivalentToDesiredState()` with independently constructed Vault-read-shaped fixtures (including bindpass stripping), negative tests
  - [ ] 7.2: Create `api/v1alpha1/ldapsecretenginestaticrole_test.go` — test `toMap()` output, test `IsEquivalentToDesiredState()` filtering Vault extras like `last_vault_rotation`, negative tests
  - [ ] 7.3: Create `api/v1alpha1/ldapsecretengine_dynamicrole_test.go` — test `toMap()`, test `IsEquivalentToDesiredState()`, negative tests

- [ ] Task 8: Test fixtures and integration tests (AC: 1, 2, 4, 5)
  - [ ] 8.1: Create test YAML fixtures in `test/ldapsecretengine/` — config, static role, dynamic role CRs
  - [ ] 8.2: Create integration tests — OpenLDAP is already deployed in Kind (see Integration Test Classification below); the LDAP secrets engine needs a secrets mount enabled. Tests: config create/delete, static role create/verify/delete, dynamic role create/verify/delete

- [ ] Task 9: CRD registration and code generation (AC: all)
  - [ ] 9.1: Run `make manifests generate fmt vet test`
  - [ ] 9.2: Add new CRD YAML files to `config/crd/kustomization.yaml` (CRD registration checklist)
  - [ ] 9.3: Verify all existing tests still pass

## Dev Notes

### Integration Test Classification: Install in Kind

Per the project's Integration Test Infrastructure Philosophy:
> **Install in Kind** — If the service can be installed in the Kind cluster and configured to work with Vault, the test **must** deploy it as a real service.

OpenLDAP is already deployed in the Kind cluster for the LDAP auth engine integration tests (see `integration/ldap/` manifests — `osixia/openldap:1.3.0` in `ldap` namespace). The LDAP secrets engine can be tested against this same OpenLDAP instance. The integration test suite's `deploy-ldap` Makefile target already handles deployment.

**Vault setup for secrets engine testing:**
- The integration test Vault setup script (`integration/vault/vault-init.sh` or equivalent) must enable an LDAP secrets engine mount (e.g., `vault secrets enable -path=ldap/test openldap`)
- The config CR test should point to `ldap://ldap.ldap.svc.cluster.local` with admin bind credentials (`cn=admin,dc=example,dc=com` / `admin`)
- Static role tests can use pre-seeded LDAP users (e.g., `trevor`, `dev1`..`dev12` from `integration/ldap/configmap.yaml`)
- Dynamic role tests need creation_ldif/deletion_ldif templates targeting the same OpenLDAP instance

**Note on Vault secrets enable:** The integration test must ensure a `SecretEngineMount` CR exists at the test path before creating the config CR. Check the SSH and LDAP auth engine integration tests for the mount creation pattern.

### Story Intelligence Chain — Previous Story Context

**Epic 11 stories (11.1 AWS, 11.2 Transit, 11.3 SSH)** are the immediate predecessors and established the current secret engine CRD pattern:
- **Pattern:** Types file with inline config struct, `toMap()`, `IsEquivalentToDesiredState()`, `PrepareInternalValues()`, webhook, controller, unit tests, integration tests (where applicable)
- **`removeUnsetFields` helper** (introduced in Epic 11): prevents false drift when the operator includes all managed fields with zero defaults but Vault omits fields that were never set — used in `IsEquivalentToDesiredState` before `filterPayloadToDesiredKeys`
- **Always-write controller pattern** for config types with write-only credentials (AWS config writes every reconcile because `secret_key` is not readable from Vault, same applies to LDAP `bindpass`)
- **`json.Number` requirement**: All numeric fields in `toMap()` must emit `json.Number` (e.g., `json.Number(strconv.Itoa(v))` or `json.Number(strconv.FormatInt(v, 10))`) because Vault's Go client uses `json.Decoder.UseNumber()`
- **Epic 11 retrospective action items** (all applied): shared framework protection rule, CRD registration checklist, `spec.name` immutability rule, `json.Number` programming rule

**Existing LDAP auth engine** (`LDAPAuthEngineConfig` in `api/v1alpha1/ldapauthengineconfig_types.go`) provides the credential resolution pattern for bind credentials. The secrets engine config shares the bind credential concept but has different fields (schema, password_policy vs group/user/token fields).

### Vault API Paths

| Operation | Method | Path |
|-----------|--------|------|
| Write config | POST | `{path}/config` |
| Read config | GET | `{path}/config` |
| Delete config | DELETE | `{path}/config` |
| Create/update static role | POST | `{path}/static-role/{name}` |
| Read static role | GET | `{path}/static-role/{name}` |
| Delete static role | DELETE | `{path}/static-role/{name}` |
| Create/update dynamic role | POST | `{path}/role/{name}` |
| Read dynamic role | GET | `{path}/role/{name}` |
| Delete dynamic role | DELETE | `{path}/role/{name}` |

### LDAPSecretEngineConfig — Vault API Field Reference

**Write (`POST {path}/config`) fields (non-Enterprise, relevant to operator):**
- `binddn` (string) — DN of object to bind for managing user entries
- `bindpass` (string) — Password for bind DN
- `url` (string, default "ldap://127.0.0.1") — LDAP server URL(s), comma-separated
- `password_policy` (string) — Name of Vault password policy for generated passwords
- `schema` (string, default "openldap") — LDAP schema: `openldap`, `ad`, or `racf`
- `userdn` (string) — Base DN for user search in library/static roles
- `userattr` (string) — Attribute for user search (defaults vary by schema: `cn` for openldap, `userPrincipalName` for ad, `racfid` for racf)
- `upndomain` (string) — Domain for UPN string construction (Active Directory)
- `request_timeout` (integer/string, default "90s") — Connection timeout
- `starttls` (bool) — Issue StartTLS after unencrypted connection
- `insecure_tls` (bool) — Skip LDAP server certificate verification
- `certificate` (string) — CA cert for LDAP server verification (PEM)
- `client_tls_cert` (string) — Client cert for LDAP (PEM)
- `client_tls_key` (string) — Client key for LDAP (PEM)
- `connection_timeout` (integer/string, default "30s") — Connection timeout (deprecated, use `request_timeout`)
- `length` (int, default 64) — Generated password length (deprecated, use `password_policy`)
- `skip_static_role_import_rotation` (bool, default false) — Default skip_import_rotation for static roles
- `credential_type` (string, default "password") — Type: `password` or `phrase` (RACF phrase mode)

**Read (`GET {path}/config`) response — `bindpass` is NEVER returned:**
```json
{
  "data": {
    "binddn": "cn=admin,dc=hashicorp,dc=com",
    "case_sensitive_names": false,
    "certificate": "",
    "insecure_tls": false,
    "length": 64,
    "schema": "openldap",
    "starttls": false,
    "tls_max_version": "tls12",
    "tls_min_version": "tls12",
    "url": "ldap://127.0.0.1"
  }
}
```

### Critical: `IsEquivalentToDesiredState` for Config — Bind Password Stripping

Vault never returns `bindpass` on read. The implementation must:
1. Build `desiredState` from `toMap()`
2. `delete(desiredState, "bindpass")` — remove it before comparison
3. Use `removeUnsetFields(desiredState, payload)` then `filterPayloadToDesiredKeys(desiredState, payload)` then `reflect.DeepEqual`

Follow the established pattern from `LDAPAuthEngineConfig` (deletes `bindpass`), `AWSSecretEngineConfig` (deletes `secret_key`).

### LDAPSecretEngineConfig — Credential Resolution

Bind credentials (`binddn` + `bindpass`) must be resolved from one of three sources via `RootCredentialConfig`:
- **K8s Secret**: keys from `UsernameKey` (default "username") and `PasswordKey` (default "password")
- **VaultSecret**: same key mapping from a Vault KV path
- **RandomSecret**: retrieve password from RandomSecret's Vault path (binddn must be in `spec.bindDN` field)

Pattern: follow `LDAPAuthEngineConfig.setInternalCredentials()` from `api/v1alpha1/ldapauthengineconfig_types.go` — this is the most direct reference since it handles the same binddn/bindpass pattern. Store resolved values in unexported fields (`retrievedBindDN`, `retrievedBindPass` with `json:"-"`) and include them in `toMap()` output.

### LDAPSecretEngineConfig — Controller Uses Always-Write Pattern

Because `bindpass` is write-only (Vault never returns it on read), drift detection cannot observe credential rotations. The config controller MUST use the always-write pattern from `AWSSecretEngineConfigReconciler.manageReconcileLogic()`: call `PrepareInternalValues` then `vaultEndpoint.Create()` unconditionally (skip the standard `CreateOrUpdate` that reads first).

This follows the established precedent from AWSSecretEngineConfig and RabbitMQSecretEngineConfig.

### LDAPSecretEngineConfig — `GetPath()` Is Fixed

The LDAP secrets engine config is always at `{path}/config` (no per-name suffix). `GetPath()` returns `CleansePath(string(d.Spec.Path) + "/config")`.

The `spec.name` field is NOT needed for path construction, but keep it for consistency with the operator pattern (unused in path construction for config type, same as AWSSecretEngineConfig where `spec.name` is kept but not used).

### LDAPSecretEngineStaticRole — Vault API Field Reference

**Write (`POST {path}/static-role/{name}`) fields:**
- `username` (string, required) — Existing LDAP username to manage. Cannot be modified after creation.
- `dn` (string) — DN of existing LDAP entry. Takes precedence over username for password rotation. Cannot be modified after creation.
- `rotation_period` (string/integer, default 0) — Time in seconds before credential rotation. Minimum 10s.
- `skip_import_rotation` (boolean, default false) — Skip initial password rotation on role creation.

**Read (`GET {path}/static-role/{name}`) response:**
```json
{
  "data": {
    "dn": "uid=hashicorp,ou=Users,dc=hashicorp,dc=com",
    "last_vault_rotation": "2026-03-30T16:10:00Z",
    "next_vault_rotation": "2026-03-31T16:10:00Z",
    "rotation_period": 86400,
    "username": "hashicorp"
  }
}
```

**Critical:** Vault returns `rotation_period` as an integer (not string) and adds `last_vault_rotation`/`next_vault_rotation` which are not in the write payload. `filterPayloadToDesiredKeys` handles the extras. The `rotation_period` field must be emitted as `json.Number` in `toMap()`.

**Webhook:** `ValidateUpdate` must reject changes to `spec.username` and `spec.dn` in addition to `spec.path` and `spec.name` — these are immutable per Vault's API contract.

### LDAPSecretEngineDynamicRole — Vault API Field Reference

**Write (`POST {path}/role/{name}`) fields:**
- `creation_ldif` (string, required) — Templatized LDIF string for creating user accounts. May be base64 encoded.
- `deletion_ldif` (string, required) — Templatized LDIF string for deleting user accounts. May be base64 encoded.
- `rollback_ldif` (string, optional but recommended) — Templatized LDIF string for rollback on creation failure.
- `username_template` (string) — Go template for dynamic username generation.
- `default_ttl` (string/int) — Default TTL for leases. Accepts duration format strings.
- `max_ttl` (string/int) — Maximum TTL for leases.

**Read (`GET {path}/role/{name}`) response:**
```json
{
  "data": {
    "creation_ldif": "...",
    "deletion_ldif": "...",
    "rollback_ldif": "...",
    "username_template": "v_{{.DisplayName}}_{{.RoleName}}_{{random 10}}_{{unix_time}}",
    "default_ttl": 3600,
    "max_ttl": 86400
  }
}
```

**Note:** Vault returns decoded (non-base64) LDIF strings on read, even if they were submitted as base64. The `toMap()` must emit raw LDIF strings, not base64. If the user provides base64 in the CR, the operator should not decode — let Vault handle it (Vault accepts both).

**TTL fields:** Vault returns TTL values as integers (seconds). The `toMap()` must emit these as string duration values if set in the CRD (e.g., `"1h"`) — Vault accepts string durations on write and converts to seconds on read. For `IsEquivalentToDesiredState`, if the user specifies `"1h"` but Vault returns `3600`, these won't match. Two options: (a) always store/emit as integer seconds, or (b) convert duration strings to `json.Number` seconds in `toMap()`. Option (a) is simpler — the CRD field type should be `string` (accepting both "3600" and "1h"), and `toMap()` should pass the value as-is. Vault normalizes on write, and `removeUnsetFields` handles absent fields. Trust `filterPayloadToDesiredKeys` for comparison — if a duration string is sent on write, Vault converts it to seconds on read, and the next reconcile loop will see the integer from Vault matching after the always-write pattern (or drift will be detected and corrected).

### CRD Field Spec — LDAPSecretEngineConfig

```go
type LDAPSEConfig struct {
    // URL is the LDAP server to connect to. Examples: ldaps://ldap.myorg.com, ldaps://ldap.myorg.com:636
    // +kubebuilder:validation:Optional
    // +kubebuilder:default="ldap://127.0.0.1"
    URL string `json:"url"`

    // Schema is the LDAP schema to use when storing entry passwords
    // +kubebuilder:validation:Optional
    // +kubebuilder:default="openldap"
    // +kubebuilder:validation:Enum:={"openldap","ad","racf"}
    Schema string `json:"schema"`

    // BindDN is the Distinguished name of object to bind for managing user entries
    // +kubebuilder:validation:Optional
    BindDN string `json:"bindDN,omitempty"`

    // PasswordPolicy is the name of the password policy to use for generating passwords
    // +kubebuilder:validation:Optional
    PasswordPolicy string `json:"passwordPolicy,omitempty"`

    // UserDN is the base DN under which to perform user search
    // +kubebuilder:validation:Optional
    UserDN string `json:"userDN,omitempty"`

    // UserAttr is the attribute field name used to perform user search
    // +kubebuilder:validation:Optional
    UserAttr string `json:"userAttr,omitempty"`

    // UPNDomain is used to construct a UPN string for Active Directory
    // +kubebuilder:validation:Optional
    UPNDomain string `json:"upnDomain,omitempty"`

    // RequestTimeout is timeout in seconds for the connection
    // +kubebuilder:validation:Optional
    RequestTimeout string `json:"requestTimeout,omitempty"`

    // StartTLS issues a StartTLS command after establishing an unencrypted connection
    // +kubebuilder:validation:Optional
    StartTLS bool `json:"startTLS,omitempty"`

    // InsecureTLS skips LDAP server SSL certificate verification
    // +kubebuilder:validation:Optional
    InsecureTLS bool `json:"insecureTLS,omitempty"`

    // Certificate is the CA certificate for verifying LDAP server certificate (PEM)
    // +kubebuilder:validation:Optional
    Certificate string `json:"certificate,omitempty"`

    // ClientTLSCert is the client certificate for LDAP (PEM)
    // +kubebuilder:validation:Optional
    ClientTLSCert string `json:"clientTLSCert,omitempty"`

    // ClientTLSKey is the client key for LDAP (PEM)
    // +kubebuilder:validation:Optional
    ClientTLSKey string `json:"clientTLSKey,omitempty"`

    // SkipStaticRoleImportRotation is the default value for skip_import_rotation on static roles
    // +kubebuilder:validation:Optional
    SkipStaticRoleImportRotation bool `json:"skipStaticRoleImportRotation,omitempty"`

    // CredentialType is the type of password to generate (password or phrase for RACF)
    // +kubebuilder:validation:Optional
    // +kubebuilder:validation:Enum:={"password","phrase"}
    CredentialType string `json:"credentialType,omitempty"`

    // Length is the generated password string length (deprecated: use passwordPolicy)
    // +kubebuilder:validation:Optional
    // +kubebuilder:validation:Minimum=0
    Length *int `json:"length,omitempty"`

    // ConnectionTimeout is timeout before trying the next URL (deprecated: use requestTimeout)
    // +kubebuilder:validation:Optional
    ConnectionTimeout string `json:"connectionTimeout,omitempty"`

    retrievedBindDN   string `json:"-"`
    retrievedBindPass string `json:"-"`
}
```

### CRD Field Spec — LDAPSecretEngineStaticRole

```go
type LDAPSEStaticRole struct {
    // Username is the existing LDAP username to manage password rotation for. Cannot be modified after creation.
    // +kubebuilder:validation:Required
    Username string `json:"username"`

    // DN is the Distinguished Name of the existing LDAP entry. Takes precedence over username. Cannot be modified after creation.
    // +kubebuilder:validation:Optional
    DN string `json:"dn,omitempty"`

    // RotationPeriod is the time in seconds before credential rotation (minimum 10s)
    // +kubebuilder:validation:Required
    // +kubebuilder:validation:Minimum=10
    RotationPeriod int `json:"rotationPeriod"`

    // SkipImportRotation when true skips the initial password rotation on role creation
    // +kubebuilder:validation:Optional
    SkipImportRotation bool `json:"skipImportRotation,omitempty"`
}
```

### CRD Field Spec — LDAPSecretEngineDynamicRole

```go
type LDAPSEDynamicRole struct {
    // CreationLDIF is a templatized LDIF string for creating LDAP user accounts (may be base64 encoded)
    // +kubebuilder:validation:Required
    CreationLDIF string `json:"creationLDIF"`

    // DeletionLDIF is a templatized LDIF string for deleting LDAP user accounts (may be base64 encoded)
    // +kubebuilder:validation:Required
    DeletionLDIF string `json:"deletionLDIF"`

    // RollbackLDIF is a templatized LDIF string for rollback on creation failure (recommended)
    // +kubebuilder:validation:Optional
    RollbackLDIF string `json:"rollbackLDIF,omitempty"`

    // UsernameTemplate is a Go template for dynamic username generation
    // +kubebuilder:validation:Optional
    UsernameTemplate string `json:"usernameTemplate,omitempty"`

    // DefaultTTL specifies the default TTL for leases (duration format string, e.g. "1h")
    // +kubebuilder:validation:Optional
    DefaultTTL string `json:"defaultTTL,omitempty"`

    // MaxTTL specifies the maximum TTL for leases (duration format string, e.g. "24h")
    // +kubebuilder:validation:Optional
    MaxTTL string `json:"maxTTL,omitempty"`
}
```

### Kubebuilder Markers Checklist

- All exported Spec fields: `+kubebuilder:validation:Required` or `+kubebuilder:validation:Optional`
- Config `Schema`: `+kubebuilder:validation:Enum:={"openldap","ad","racf"}` with `+kubebuilder:default="openldap"`
- Config `CredentialType`: `+kubebuilder:validation:Enum:={"password","phrase"}`
- Config `URL`: `+kubebuilder:default="ldap://127.0.0.1"`
- StaticRole `RotationPeriod`: `+kubebuilder:validation:Minimum=10`
- Root type: `+kubebuilder:object:root=true`, `+kubebuilder:subresource:status`
- Conditions field: `+patchMergeKey=type`, `+patchStrategy=merge`, `+listType=map`, `+listMapKey=type`

### RBAC Markers

Config controller:
```go
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=ldapsecretengineconfigs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=ldapsecretengineconfigs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=ldapsecretengineconfigs/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=serviceaccounts/token,verbs=create
//+kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;patch
```

Static role controller:
```go
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=ldapsecretenginestaticroles,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=ldapsecretenginestaticroles/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=ldapsecretenginestaticroles/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=serviceaccounts/token,verbs=create
//+kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;patch
```

Dynamic role controller:
```go
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=ldapsecretengine_dynamicroles,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=ldapsecretengine_dynamicroles/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=ldapsecretengine_dynamicroles/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=serviceaccounts/token,verbs=create
//+kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;patch
```

### File Structure

| File | Action | Description |
|------|--------|-------------|
| `api/v1alpha1/ldapsecretengineconfig_types.go` | NEW | Config CRD type, VaultObject, ConditionsAware, toMap, credential resolution |
| `api/v1alpha1/ldapsecretengineconfig_webhook.go` | NEW | Config webhook — defaulter, validator, immutable path/name |
| `api/v1alpha1/ldapsecretengineconfig_test.go` | NEW | Unit tests for config toMap, IsEquivalentToDesiredState |
| `api/v1alpha1/ldapsecretenginestaticrole_types.go` | NEW | Static role CRD type, VaultObject, ConditionsAware, toMap |
| `api/v1alpha1/ldapsecretenginestaticrole_webhook.go` | NEW | Static role webhook — defaulter, validator, immutable path/name/username/dn |
| `api/v1alpha1/ldapsecretenginestaticrole_test.go` | NEW | Unit tests for static role toMap, IsEquivalentToDesiredState |
| `api/v1alpha1/ldapsecretengine_dynamicrole_types.go` | NEW | Dynamic role CRD type, VaultObject, ConditionsAware, toMap |
| `api/v1alpha1/ldapsecretengine_dynamicrole_webhook.go` | NEW | Dynamic role webhook — defaulter, validator, immutable path/name |
| `api/v1alpha1/ldapsecretengine_dynamicrole_test.go` | NEW | Unit tests for dynamic role toMap, IsEquivalentToDesiredState |
| `internal/controller/ldapsecretengineconfig_controller.go` | NEW | Config reconciler with always-write pattern, Secret/RandomSecret watches |
| `internal/controller/ldapsecretenginestaticrole_controller.go` | NEW | Static role reconciler — standard VaultResource pattern |
| `internal/controller/ldapsecretengine_dynamicrole_controller.go` | NEW | Dynamic role reconciler — standard VaultResource pattern |
| `internal/controller/ldapsecretengineconfig_controller_test.go` | NEW | Integration tests for config |
| `internal/controller/ldapsecretenginestaticrole_controller_test.go` | NEW | Integration tests for static role |
| `internal/controller/ldapsecretengine_dynamicrole_controller_test.go` | NEW | Integration tests for dynamic role |
| `cmd/main.go` | UPDATE | Register 3 controllers + 3 webhooks |
| `config/crd/kustomization.yaml` | UPDATE | Add 3 new CRD YAML files to resources list |
| `test/ldapsecretengine/` | NEW | Test YAML fixtures for config, static role, dynamic role |

### Files Being Modified — Current State

**`cmd/main.go`**: Currently registers ~35+ controllers and ~35+ webhooks. New registrations follow the exact same pattern:
- Controller: `(&controller.LDAPSecretEngineConfigReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "LDAPSecretEngineConfig")}).SetupWithManager(mgr)`
- Webhook (inside ENABLE_WEBHOOKS): `(&redhatcopv1alpha1.LDAPSecretEngineConfig{}).SetupWebhookWithManager(mgr)`

No existing behavior is changed — this is purely additive.

**`config/crd/kustomization.yaml`**: Add the 3 new CRD YAML files to the `resources` list. This is required for the Helm chart build to include the CRDs.

### Existing Patterns to Follow (Reference Files)

| Pattern | Reference File |
|---------|---------------|
| Config type with bind credentials | `api/v1alpha1/ldapauthengineconfig_types.go` (binddn/bindpass via RootCredentialConfig) |
| Config credential resolution | `LDAPAuthEngineConfig.setInternalCredentials()` in same file |
| Secret-stripping in IsEquivalentToDesiredState | `api/v1alpha1/ldapauthengineconfig_types.go` (deletes `bindpass`) |
| Always-write controller for write-only credentials | `internal/controller/awssecretengineconfig_controller.go` |
| Controller with credential watches (Secret + RandomSecret) | `internal/controller/awssecretengineconfig_controller.go` |
| Role type (simple, no credentials) | `api/v1alpha1/awssecretenginerole_types.go` |
| Webhook pattern | `api/v1alpha1/awssecretengineconfig_webhook.go` |
| Controller (simple role, no watches) | `internal/controller/awssecretenginerole_controller.go` |
| filterPayloadToDesiredKeys helper | `api/v1alpha1/payload_filter.go` |
| removeUnsetFields helper | `api/v1alpha1/payload_filter.go` |
| Unit test payload construction | Project context: never derive expected from code under test |
| Integration test with OpenLDAP | `internal/controller/ldapauthengine_controller_test.go` |
| Test fixture creation (unstructured) | `decoder.CreateFromYAML(ctx, client, path, namespace)` |

### Unit Test Requirements

**Config tests (`ldapsecretengineconfig_test.go`):**
1. `TestLDAPSecretEngineConfig_toMap` — verify snake_case keys, verify all set fields appear, verify bind credentials resolved
2. `TestLDAPSecretEngineConfig_IsEquivalentToDesiredState_Match` — construct Vault-read-shaped payload (no `bindpass`), verify returns `true`
3. `TestLDAPSecretEngineConfig_IsEquivalentToDesiredState_Mismatch` — change a managed field (e.g., `schema`), verify returns `false`
4. `TestLDAPSecretEngineConfig_IsEquivalentToDesiredState_ExtraVaultFields` — add extra Vault fields (e.g., `case_sensitive_names`), verify still returns `true`

**Static role tests (`ldapsecretenginestaticrole_test.go`):**
1. `TestLDAPSecretEngineStaticRole_toMap` — verify `username`, `dn`, `rotation_period` as `json.Number`, `skip_import_rotation`
2. `TestLDAPSecretEngineStaticRole_IsEquivalentToDesiredState_Match` — Vault-read fixture with `rotation_period` as `json.Number`, plus `last_vault_rotation`/`next_vault_rotation` extras
3. `TestLDAPSecretEngineStaticRole_IsEquivalentToDesiredState_Mismatch` — change `rotation_period`, verify returns `false`

**Dynamic role tests (`ldapsecretengine_dynamicrole_test.go`):**
1. `TestLDAPSecretEngineDynamicRole_toMap` — verify `creation_ldif`, `deletion_ldif`, `rollback_ldif`, `default_ttl`, `max_ttl`
2. `TestLDAPSecretEngineDynamicRole_IsEquivalentToDesiredState_Match` — standard comparison
3. `TestLDAPSecretEngineDynamicRole_IsEquivalentToDesiredState_Mismatch` — change a field, verify returns `false`

### Anti-Patterns / DO NOT

- **DO NOT** create an `LDAPSecretEngineLibrary` CRD — library sets are deferred to a future story per the epic note ("Consider library sets as a separate CRD")
- **DO NOT** reuse the `LDAPConfig` struct from `ldapauthengineconfig_types.go` — the secrets engine config has a completely different field set than the auth engine config (no group/user/token fields; has schema, password_policy, credential_type instead). Create a new `LDAPSEConfig` inline struct.
- **DO NOT** modify existing `LDAPAuthEngineConfig` code — this is a new secrets engine type, not a modification of the auth engine
- **DO NOT** modify shared framework behavior (reconcile_skeleton.go, vaultresourcereconciler.go, etc.) — only add new files
- **DO NOT** include Enterprise-only fields (rotation_schedule, rotation_window, disable_automated_rotation, rotation_policy, self_managed, password) — these require Vault Enterprise and are out of scope for the community operator
- **DO NOT** derive expected payloads from `toMap()` in unit tests — construct independent Vault-read-shaped fixtures
- **DO NOT** forget to add new CRD YAML files to `config/crd/kustomization.yaml` after `make manifests`

### Project Structure Notes

- All new files follow existing naming conventions: lowercase type name for files
- New files go in established directories (`api/v1alpha1/`, `internal/controller/`, `test/`)
- Test fixture directory `test/ldapsecretengine/` follows the existing pattern (`test/awssecretengine/`, `test/ssh/`)
- Dynamic role file uses underscore separator (`ldapsecretengine_dynamicrole`) to match the existing naming pattern for multi-word suffixes
- No conflicts with existing code — purely additive

### References

- [Source: _bmad-output/planning-artifacts/epics.md#Epic-12, Story 12.3 — requirements and acceptance criteria]
- [Source: _bmad-output/project-context.md — all project rules, patterns, and conventions]
- [Source: api/v1alpha1/ldapauthengineconfig_types.go — LDAP bind credential resolution pattern]
- [Source: api/v1alpha1/awssecretengineconfig_types.go — recent secret engine config with credential resolution and always-write]
- [Source: api/v1alpha1/awssecretenginerole_types.go — recent secret engine role pattern]
- [Source: internal/controller/awssecretengineconfig_controller.go — always-write controller with credential watches]
- [Source: api/v1alpha1/payload_filter.go — filterPayloadToDesiredKeys and removeUnsetFields helpers]
- [Source: Vault LDAP Secrets API — https://developer.hashicorp.com/vault/api-docs/secret/ldap]
- [Source: integration/ldap/ — OpenLDAP deployment manifests for Kind integration testing]
- [Source: test/ldapauthengine/ — existing LDAP test fixtures for reference]

## Code Review Record

### Review Model Used

(to be filled during review — must differ from dev model)

### Review Findings

(to be filled during review)

### Decisions Needed / Decisions Taken

(to be filled during review)

### Fixes Applied

(to be filled during review)

## Dev Agent Record

### Agent Model Used

{{agent_model_name_version}}

### Debug Log References

### Completion Notes List

### Change Log

### File List
