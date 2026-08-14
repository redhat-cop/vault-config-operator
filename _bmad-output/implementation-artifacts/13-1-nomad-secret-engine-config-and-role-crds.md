---
baseline_commit: 2ccad3a67685ec8bdd0c585247ee148143415e6c
---

# Story 13.1: Nomad Secret Engine — Config and Role CRDs

Status: done

## Story

As an operator developer,
I want CRDs for NomadSecretEngineConfig and NomadSecretEngineRole,
So that Vault's Nomad secret engine can be managed declaratively.

## Acceptance Criteria

1. **Given** a NomadSecretEngineConfig CR is created with a Nomad management token (via K8s Secret reference) **When** the reconciler processes it **Then** the Nomad access config is written to Vault at `{path}/config/access` and ReconcileSuccessful=True

2. **Given** a NomadSecretEngineRole CR is created with policies and token type **When** the reconciler processes it **Then** the role exists in Vault at `{path}/role/{name}` and can generate dynamic Nomad ACL tokens

3. **Given** the NomadSecretEngineConfig CR is deleted **When** the reconciler processes deletion **Then** the K8s object is removed but Vault config is **NOT** deleted (`IsDeletable=false` — Vault has no `DELETE /nomad/config/access` endpoint)

4. **Given** the NomadSecretEngineRole CR is deleted **When** the reconciler processes deletion **Then** the role is removed from Vault via `DELETE /nomad/role/{name}` and the CR is deleted from K8s

5. **Given** the NomadSecretEngineRole CR spec is updated (e.g., `policies` changed) **When** the reconciler processes the update **Then** the Vault role reflects the updated values

6. **Given** any Nomad CR is created or updated **When** the webhook validates it **Then** `spec.path` immutability is enforced on updates, `spec.name` immutability is enforced on updates, and credential source validation passes for config

7. **Given** the CRD types are implemented **When** the story is marked done **Then** a documentation file exists at `docs/secret-engines/nomad.md` following `docs/engine-doc-template.md` (DNFR5)

## Tasks / Subtasks

- [x] Task 1: Create `NomadSecretEngineConfig` type (AC: 1, 3, 6)
  - [x] 1.1: Create `api/v1alpha1/nomadsecretengineconfig_types.go` — Spec with `Connection`, `Authentication`, `Path`, inline `NomadSEConfig` struct, `RootCredentials` (RootCredentialConfig)
  - [x] 1.2: Implement `VaultObject` interface — `GetPath()` returns `{path}/config/access`, `GetPayload()`, `IsEquivalentToDesiredState()`, `PrepareInternalValues()`, `IsDeletable()=false`
  - [x] 1.3: Implement `ConditionsAware` interface — Status with `Conditions []metav1.Condition`
  - [x] 1.4: Implement `setInternalCredentials()` — resolve Nomad management token from K8s Secret, VaultSecret, or RandomSecret (follow ConsulSecretEngineConfig pattern)
  - [x] 1.5: Implement `toMap()` on `NomadSEConfig` — convert to Vault API snake_case fields, include resolved token
  - [x] 1.6: Implement `IsEquivalentToDesiredState()` — must delete `token`, `ca_cert`, `client_cert`, `client_key` from desired state (token is write-only, TLS certs not returned on read), then `removeUnsetFields` + `filterPayloadToDesiredKeys`

- [x] Task 2: Create `NomadSecretEngineRole` type (AC: 2, 4, 5)
  - [x] 2.1: Create `api/v1alpha1/nomadsecretenginerole_types.go` — Spec with `Connection`, `Authentication`, `Path`, inline `NomadSERole` struct, `Name`
  - [x] 2.2: Implement `VaultObject` interface — `GetPath()` returns `{path}/role/{name}`, `IsDeletable()=true`
  - [x] 2.3: Implement `ConditionsAware` interface
  - [x] 2.4: Implement `toMap()` on `NomadSERole` — handle `type` → `token_type` field mapping in IsEquivalentToDesiredState, `policies` as comma-joined string
  - [x] 2.5: Implement `IsEquivalentToDesiredState()` — remap `type` key to `token_type` (Vault returns `token_type` on read, not `type`), convert `policies` string to `[]any` array for comparison, then `removeUnsetFields` + `filterPayloadToDesiredKeys`

- [x] Task 3: Create webhooks (AC: 6)
  - [x] 3.1: Create `api/v1alpha1/nomadsecretengineconfig_webhook.go` — `admission.Defaulter[*NomadSecretEngineConfig]`, `admission.Validator[*NomadSecretEngineConfig]`, immutable `spec.path`, credential validation via `ValidateCredentialSource()`
  - [x] 3.2: Create `api/v1alpha1/nomadsecretenginerole_webhook.go` — `admission.Defaulter[*NomadSecretEngineRole]`, `admission.Validator[*NomadSecretEngineRole]`, immutable `spec.path`/`spec.name`

- [x] Task 4: Create controllers (AC: 1, 2, 3, 4, 5)
  - [x] 4.1: Create `internal/controller/nomadsecretengineconfig_controller.go` — embed `ReconcilerBase`, always-write reconcile logic (token is write-only like AWS secret_key/Consul token), watches on `corev1.Secret` and `RandomSecret`
  - [x] 4.2: Create `internal/controller/nomadsecretenginerole_controller.go` — standard `For()` with default periodic reconcile predicate

- [x] Task 5: Register in main.go (AC: 1, 2)
  - [x] 5.1: Add controller registrations for both reconcilers
  - [x] 5.2: Add webhook registrations inside `ENABLE_WEBHOOKS` guard for both types

- [x] Task 6: Unit tests (AC: 1, 2, 5, 6)
  - [x] 6.1: Create `api/v1alpha1/nomadsecretengineconfig_test.go` — test `toMap()`, test `IsEquivalentToDesiredState()` with independently constructed Vault-read-shaped fixtures (including token/TLS stripping), negative tests
  - [x] 6.2: Create `api/v1alpha1/nomadsecretenginerole_test.go` — test `toMap()`, test `IsEquivalentToDesiredState()` with `token_type` mapping and `policies` array conversion, negative tests

- [x] Task 7: Test fixtures and integration tests (AC: 1, 2, 4, 5)
  - [x] 7.1: Create test YAML fixtures in `test/nomadsecretengine/` — config and role CRs
  - [x] 7.2: Create integration tests — deploy Nomad in Kind cluster (see Integration Test Classification); tests: config create/verify, role create/verify/delete, config non-deletable verification

- [x] Task 8: CRD registration and code generation (AC: all)
  - [x] 8.1: Run `make manifests generate fmt vet test`
  - [x] 8.2: Add new CRD YAML files to `config/crd/kustomization.yaml` (CRD registration checklist)
  - [x] 8.3: Verify all existing tests still pass

- [x] Task 9: Documentation (AC: 7)
  - [x] 9.1: Create `docs/secret-engines/nomad.md` following `docs/engine-doc-template.md`
  - [x] 9.2: Update `docs/secret-engines/index.md` with link to new doc file

## Dev Notes

### Integration Test Classification: Install in Kind

Per the project's Integration Test Infrastructure Philosophy:
> **Install in Kind** — If the service can be installed in the Kind cluster and configured to work with Vault, the test **must** deploy it as a real service.

Nomad is a standalone binary (HashiCorp product, similar to Consul and Vault) that can run as a container in Kind. It needs to be deployed with ACL enabled and bootstrapped to obtain a management token for Vault's Nomad secrets engine to use.

**Nomad deployment for integration tests:**
- Deploy Nomad as a single-node server with ACL enabled (similar to how OpenLDAP/Consul are deployed for their respective tests)
- Create integration manifests in `integration/nomad/` directory
- Use the official `hashicorp/nomad` container image
- Bootstrap ACLs and capture the management token
- Create at least one Nomad ACL policy for role tests (e.g., a `readonly` policy)
- The Vault integration test setup script must enable a Nomad secrets engine mount and configure it to point to the Nomad instance

**Vault setup for Nomad secrets engine testing:**
- Enable a Nomad secrets engine mount: `vault secrets enable -path=nomad/test nomad`
- The config CR test should point to the Nomad server address in the Kind cluster (e.g., `http://nomad.nomad.svc.cluster.local:4646`)
- Role tests should reference the pre-created Nomad ACL policy

**Note on Vault secrets enable:** The integration test must ensure a `SecretEngineMount` CR exists at the test path before creating the config CR. Follow the pattern from Consul and LDAP secret engine integration tests for mount creation.

### Story Intelligence Chain — Previous Story Context

**Epic 12 stories (12.1 Consul, 12.2 GCP, 12.3 LDAP)** are the immediate predecessors:
- **ConsulSecretEngineConfig is the closest analog** — same `config/access` endpoint, same management-token credential pattern, same `IsDeletable=false`, same always-write controller pattern. Use it as the primary reference.
- **`removeUnsetFields` helper** (introduced in Epic 11): prevents false drift when the operator includes all managed fields with zero defaults but Vault omits fields that were never set
- **Always-write controller pattern** for config types with write-only credentials (ConsulSecretEngineConfig, AWSSecretEngineConfig — writes every reconcile because token is not readable from Vault)
- **Epic 12 retrospective action items** (all applied): `toMap()` normalization rule (emit Vault-read format), 5-iteration review cap escalation, sprint-status/story-file atomicity guard, final consistency check as blocking gate
- **`json.Number` requirement**: All numeric fields in `toMap()` must emit `json.Number` — applies to `max_token_name_length` if included
- **CRD registration checklist**: After `make manifests`, add new CRD YAMLs to `config/crd/kustomization.yaml`

### Vault API Paths

| Operation | Method | Path |
|-----------|--------|------|
| Write config | POST | `{path}/config/access` |
| Read config | GET | `{path}/config/access` |
| Configure lease | POST | `{path}/config/lease` |
| Read lease config | GET | `{path}/config/lease` |
| Delete lease config | DELETE | `{path}/config/lease` |
| Create/update role | POST | `{path}/role/{name}` |
| Read role | GET | `{path}/role/{name}` |
| Delete role | DELETE | `{path}/role/{name}` |
| List roles | LIST | `{path}/role` |
| Generate credential | GET | `{path}/creds/{name}` |

**No DELETE endpoint for `/nomad/config/access`** → `NomadSecretEngineConfig.IsDeletable()` must return `false`.

### NomadSecretEngineConfig — Vault API Field Reference

**Write (`POST {path}/config/access`) fields:**
- `address` (string: "") — Nomad instance address as `"protocol://host:port"` (e.g., `"http://127.0.0.1:4646"`)
- `token` (string: "") — Nomad management token (write-only, never returned on read)
- `max_token_name_length` (int: 0) — Max length for generated Nomad token names. 0 = use Nomad's default (64 for Nomad ≤0.8.3, 256 for ≥0.8.4)
- `ca_cert` (string: "") — CA certificate for verifying Nomad server certificate (x509 PEM)
- `client_cert` (string: "") — Client certificate for Nomad TLS communication (x509 PEM)
- `client_key` (string: "") — Client key for Nomad TLS communication (x509 PEM)

**Read (`GET {path}/config/access`) response — only `address` is returned:**
```json
{
  "data": {
    "address": "http://localhost:4646/"
  }
}
```

**Critical:** Vault returns only `address` on read. The `token`, `ca_cert`, `client_cert`, `client_key`, and `max_token_name_length` fields are NOT returned. This means:
1. `IsEquivalentToDesiredState` must delete `token`, `ca_cert`, `client_cert`, `client_key` from desired state before comparison
2. The always-write controller pattern is mandatory (token is write-only)
3. `max_token_name_length` — not confirmed in read response; use `removeUnsetFields` to handle absence gracefully

### Critical: `IsEquivalentToDesiredState` for Config — Token and TLS Stripping

Vault never returns `token`, `ca_cert`, `client_cert`, or `client_key` on read. The implementation must:
1. Build `desiredState` from `toMap()`
2. `delete(desiredState, "token")` — write-only
3. `delete(desiredState, "ca_cert")` — not returned on read
4. `delete(desiredState, "client_cert")` — not returned on read
5. `delete(desiredState, "client_key")` — not returned on read
6. `removeUnsetFields(desiredState, payload)` then `filterPayloadToDesiredKeys(desiredState, payload)` then `reflect.DeepEqual`

Follow the established pattern from `ConsulSecretEngineConfig` (deletes `token`, `ca_cert`, `client_cert`, `client_key`).

### NomadSecretEngineConfig — Credential Resolution

The Nomad management token must be resolved from one of three sources via `RootCredentialConfig`:
- **K8s Secret**: token from `PasswordKey` (default "password")
- **VaultSecret**: same key mapping from a Vault KV path
- **RandomSecret**: retrieve token from RandomSecret's Vault path

Pattern: follow `ConsulSecretEngineConfig.setInternalCredentials()` from `api/v1alpha1/consulsecretengineconfig_types.go` — this is the most direct reference since it handles a single token credential (no username). Store resolved value in an unexported field (`retrievedToken` with `json:"-"`) and include it in `toMap()` output.

Unlike AWS (which has accessKey + secretKey), Nomad only needs a single token value. The `RootCredentialConfig.PasswordKey` (default `"password"`) retrieves the token. No username resolution needed. Follow the Consul pattern exactly.

### NomadSecretEngineConfig — Controller Uses Always-Write Pattern

Because `token` is write-only (Vault never returns it on read), drift detection cannot observe token rotations. The config controller MUST use the always-write pattern from `ConsulSecretEngineConfigReconciler` / `AWSSecretEngineConfigReconciler`: call `PrepareInternalValues` then `vaultEndpoint.Create()` unconditionally.

### NomadSecretEngineConfig — `GetPath()` Is Fixed

The Nomad config endpoint is always at `{path}/config/access` (no per-name suffix). `GetPath()` returns `CleansePath(string(d.Spec.Path) + "/config/access")`.

No `spec.name` field needed for NomadSecretEngineConfig — the path is fixed. Omit the `Name` field entirely (unlike ConsulSecretEngineConfig which also omits it). The config endpoint has no name-based routing.

### NomadSecretEngineRole — Vault API Field Reference

**Write (`POST {path}/role/{name}`) fields:**
- `policies` (string: "") — Comma-separated list of Nomad ACL policies for the generated token
- `global` (bool: false) — Whether the token should be global (replicated across Nomad regions)
- `type` (string: "client") — Token type: `"client"` or `"management"`

**Read (`GET {path}/role/{name}`) response:**
```json
{
  "data": {
    "lease": "0s",
    "policies": ["example"],
    "token_type": "client"
  }
}
```

### Critical: `IsEquivalentToDesiredState` for Role — `type` → `token_type` Mapping

**Vault API gotcha**: The write endpoint accepts `type` but the read endpoint returns `token_type`. This is identical in pattern to the AWS `credential_type` → `credential_types` mapping.

The `IsEquivalentToDesiredState` implementation must:
1. Build `desiredState` from `toMap()` (which uses `type`)
2. If `type` key exists, copy its value to `token_type` and delete `type`
3. `removeUnsetFields` + `filterPayloadToDesiredKeys` + `reflect.DeepEqual`

Unit tests MUST verify this mapping with independently constructed Vault-read-shaped fixtures.

### Critical: `policies` — String Write vs Array Read

**Vault API gotcha**: The write endpoint accepts `policies` as a comma-separated string, but the read endpoint returns `policies` as an array of strings.

**CRD design**: Store `Policies` as `[]string` in the CRD (user-friendly). In `toMap()`, join with comma to produce a single string: `strings.Join(i.Policies, ",")`.

**`IsEquivalentToDesiredState` handling**: Convert the comma-joined string in desired state to `[]any{"policy1", "policy2"}` matching Vault's read format before comparison. Alternatively, use `toInterfaceArray()` in `toMap()` directly and rely on Vault accepting arrays on write (many Vault APIs accept both). If choosing the comma-string approach, the conversion must happen in `IsEquivalentToDesiredState`.

**Recommended approach** (simpler, matches Consul pattern): Emit `policies` as `toInterfaceArray(i.Policies)` in `toMap()`. Vault's Nomad plugin likely accepts arrays. Then `IsEquivalentToDesiredState` can compare arrays directly after `filterPayloadToDesiredKeys`. If Vault rejects arrays on write, fall back to comma-separated strings with array conversion in `IsEquivalentToDesiredState`.

### `lease` Extra Field in Read Response

The read response includes `"lease": "0s"` which is not a field in the write API. `filterPayloadToDesiredKeys` handles this automatically — the `lease` key is not in `desiredState` so it's excluded from the filtered payload.

### `global` Field — May Not Appear in Read Response

The sample read response does not include `global`. If Vault omits `global` when false, `removeUnsetFields` will remove it from `desiredState` (bool zero-value) when it's absent from `payload`. If Vault always returns it, comparison works normally. Either way, the standard helpers handle this correctly.

### NomadSecretEngineRole — `GetPath()`

Returns `CleansePath(string(d.Spec.Path) + "/role/" + name)` where name is `d.Spec.Name` if set, otherwise `d.Name`. Standard pattern. Note: Nomad uses `/role/` (singular), not `/roles/` (plural). Verify against the API docs.

### CRD Field Spec — NomadSecretEngineConfig

```go
type NomadSEConfig struct {
    // Address specifies the Nomad instance address as "protocol://host:port" (e.g., "http://127.0.0.1:4646")
    // +kubebuilder:validation:Required
    Address string `json:"address"`

    // MaxTokenNameLength specifies the maximum length for generated Nomad token names.
    // 0 uses Nomad's default (64 for ≤0.8.3, 256 for ≥0.8.4).
    // +kubebuilder:validation:Optional
    // +kubebuilder:validation:Minimum=0
    MaxTokenNameLength int `json:"maxTokenNameLength,omitempty"`

    // CACert is the CA certificate for verifying the Nomad server certificate (x509 PEM)
    // +kubebuilder:validation:Optional
    CACert string `json:"caCert,omitempty"`

    // ClientCert is the client certificate for Nomad TLS communication (x509 PEM).
    // If set, clientKey must also be set.
    // +kubebuilder:validation:Optional
    ClientCert string `json:"clientCert,omitempty"`

    // ClientKey is the client key for Nomad TLS communication (x509 PEM).
    // If set, clientCert must also be set.
    // +kubebuilder:validation:Optional
    ClientKey string `json:"clientKey,omitempty"`

    retrievedToken string `json:"-"`
}
```

Note: No `Name` field on NomadSecretEngineConfigSpec — the Vault path is fixed at `{path}/config/access`.

### CRD Field Spec — NomadSecretEngineRole

```go
type NomadSERole struct {
    // Policies is the list of Nomad ACL policies to assign to the generated token.
    // These must be created beforehand in Nomad.
    // +kubebuilder:validation:Optional
    // +listType=set
    Policies []string `json:"policies,omitempty"`

    // Global specifies if the token should be global (replicated across Nomad regions).
    // +kubebuilder:validation:Optional
    Global bool `json:"global,omitempty"`

    // Type specifies the type of Nomad token to create: "client" or "management".
    // +kubebuilder:validation:Optional
    // +kubebuilder:default="client"
    // +kubebuilder:validation:Enum:={"client","management"}
    Type string `json:"type"`
}
```

### Kubebuilder Markers Checklist

- All exported Spec fields: `+kubebuilder:validation:Required` or `+kubebuilder:validation:Optional`
- Config `Address`: `+kubebuilder:validation:Required`
- Config `MaxTokenNameLength`: `+kubebuilder:validation:Minimum=0`
- Role `Type`: `+kubebuilder:validation:Enum:={"client","management"}` with `+kubebuilder:default="client"`
- Role `Policies`: `+listType=set`
- Root type: `+kubebuilder:object:root=true`, `+kubebuilder:subresource:status`
- Conditions field: `+patchMergeKey=type`, `+patchStrategy=merge`, `+listType=map`, `+listMapKey=type`

### RBAC Markers

Config controller:
```go
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=nomadsecretengineconfigs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=nomadsecretengineconfigs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=nomadsecretengineconfigs/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=serviceaccounts/token,verbs=create
//+kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;patch
```

Role controller:
```go
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=nomadsecretengineroles,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=nomadsecretengineroles/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=nomadsecretengineroles/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=serviceaccounts/token,verbs=create
//+kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;patch
```

### File Structure

| File | Action | Description |
|------|--------|-------------|
| `api/v1alpha1/nomadsecretengineconfig_types.go` | NEW | Config CRD type, VaultObject, ConditionsAware, toMap, credential resolution |
| `api/v1alpha1/nomadsecretengineconfig_webhook.go` | NEW | Config webhook — defaulter, validator, immutable path |
| `api/v1alpha1/nomadsecretengineconfig_test.go` | NEW | Unit tests for config toMap, IsEquivalentToDesiredState |
| `api/v1alpha1/nomadsecretenginerole_types.go` | NEW | Role CRD type, VaultObject, ConditionsAware, toMap with type→token_type mapping |
| `api/v1alpha1/nomadsecretenginerole_webhook.go` | NEW | Role webhook — defaulter, validator, immutable path/name |
| `api/v1alpha1/nomadsecretenginerole_test.go` | NEW | Unit tests for role toMap, IsEquivalentToDesiredState |
| `internal/controller/nomadsecretengineconfig_controller.go` | NEW | Config reconciler with always-write pattern, Secret/RandomSecret watches |
| `internal/controller/nomadsecretenginerole_controller.go` | NEW | Role reconciler — standard VaultResource pattern |
| `internal/controller/nomadsecretengineconfig_controller_test.go` | NEW | Integration tests for config |
| `internal/controller/nomadsecretenginerole_controller_test.go` | NEW | Integration tests for role |
| `cmd/main.go` | UPDATE | Register 2 controllers + 2 webhooks |
| `config/crd/kustomization.yaml` | UPDATE | Add 2 new CRD YAML files to resources list |
| `test/nomadsecretengine/` | NEW | Test YAML fixtures for config, role |
| `integration/nomad/` | NEW | Nomad deployment manifests for Kind integration testing |
| `docs/secret-engines/nomad.md` | NEW | Documentation following engine-doc-template.md |
| `docs/secret-engines/index.md` | UPDATE | Add link to Nomad doc |

### Files Being Modified — Current State

**`cmd/main.go`**: Currently registers ~40+ controllers and ~40+ webhooks. New registrations follow the exact same pattern:
- Controller: `(&controller.NomadSecretEngineConfigReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "NomadSecretEngineConfig")}).SetupWithManager(mgr)`
- Webhook (inside ENABLE_WEBHOOKS): `(&redhatcopv1alpha1.NomadSecretEngineConfig{}).SetupWebhookWithManager(mgr)`

No existing behavior is changed — this is purely additive.

**`config/crd/kustomization.yaml`**: Add the 2 new CRD YAML files (`redhatcop.redhat.io_nomadsecretengineconfigs.yaml`, `redhatcop.redhat.io_nomadsecretengineroles.yaml`) to the `resources` list. Required for Helm chart build.

**`docs/secret-engines/index.md`**: Add a link entry for `[Nomad](nomad.md)`.

### Existing Patterns to Follow (Reference Files)

| Pattern | Reference File |
|---------|---------------|
| Config type with management token credential | `api/v1alpha1/consulsecretengineconfig_types.go` (closest analog — same config/access pattern, single token) |
| Config credential resolution (single token) | `ConsulSecretEngineConfig.setInternalCredentials()` in same file |
| Token-stripping + TLS-stripping in IsEquivalentToDesiredState | `api/v1alpha1/consulsecretengineconfig_types.go` (deletes `token`, `ca_cert`, `client_cert`, `client_key`) |
| Always-write controller for write-only credentials | `internal/controller/consulsecretengineconfig_controller.go` or `awssecretengineconfig_controller.go` |
| Controller with credential watches (Secret + RandomSecret) | `internal/controller/consulsecretengineconfig_controller.go` |
| Role type (simple, no credentials) | `api/v1alpha1/consulsecretenginerole_types.go` |
| Webhook pattern | `api/v1alpha1/consulsecretengineconfig_webhook.go` |
| Controller (simple role, no watches) | `internal/controller/consulsecretenginerole_controller.go` |
| Field rename in IsEquivalentToDesiredState (type→token_type) | `api/v1alpha1/awssecretenginerole_types.go` (credential_type → credential_types mapping) |
| filterPayloadToDesiredKeys helper | `api/v1alpha1/payload_filter.go` |
| removeUnsetFields helper | `api/v1alpha1/payload_filter.go` |
| Unit test payload construction | Project context: never derive expected from code under test |
| Integration test with in-Kind service | `internal/controller/consulsecretengine_controller_test.go` (Consul in Kind) |
| Test fixture creation (unstructured) | `decoder.CreateFromYAML(ctx, client, path, namespace)` |
| Documentation template | `docs/engine-doc-template.md` |

### Unit Test Requirements

**Config tests (`nomadsecretengineconfig_test.go`):**
1. `TestNomadSecretEngineConfig_toMap` — verify snake_case keys (`address`, `token`, `max_token_name_length`, `ca_cert`, `client_cert`, `client_key`), verify all set fields appear, verify resolved token is included
2. `TestNomadSecretEngineConfig_IsEquivalentToDesiredState_Match` — construct Vault-read-shaped payload (only `address`), verify returns `true`
3. `TestNomadSecretEngineConfig_IsEquivalentToDesiredState_Mismatch` — change `address`, verify returns `false`
4. `TestNomadSecretEngineConfig_IsEquivalentToDesiredState_TokenInPayload` — add `token` to Vault payload (shouldn't happen but defensive), verify still returns `true` after stripping
5. `TestNomadSecretEngineConfig_ClientCertPairValidation` — verify webhook rejects clientCert without clientKey and vice versa

**Role tests (`nomadsecretenginerole_test.go`):**
1. `TestNomadSecretEngineRole_toMap` — verify `policies` (comma-joined or array), `global`, `type`
2. `TestNomadSecretEngineRole_IsEquivalentToDesiredState_Match` — Vault-read fixture with `token_type` (not `type`) and `policies` as `[]any`, plus `lease` extra field. Verify returns `true`
3. `TestNomadSecretEngineRole_IsEquivalentToDesiredState_Mismatch` — change `policies`, verify returns `false`
4. `TestNomadSecretEngineRole_IsEquivalentToDesiredState_TypeMapping` — explicitly test the `type` → `token_type` field rename mapping with independent fixture
5. `TestNomadSecretEngineRole_IsEquivalentToDesiredState_ManagementType` — verify `type: management` maps to `token_type: management`

### Webhook Validation Rules

**Config webhook:**
- `ValidateCreate`: call `ValidateCredentialSource()`, validate `clientCert`/`clientKey` pair (both or neither)
- `ValidateUpdate`: reject changes to `spec.path` (immutable), then same validations as create
- `ValidateDelete`: no-op (return nil)

**Role webhook:**
- `ValidateCreate`: validate `type` enum is `"client"` or `"management"` (kubebuilder handles this, but belt-and-suspenders)
- `ValidateUpdate`: reject changes to `spec.path` and `spec.name` (immutable)
- `ValidateDelete`: no-op (return nil)

### Anti-Patterns / DO NOT

- **DO NOT** create a `NomadSecretEngineLease` CRD — the lease configuration (`config/lease`) is a separate concern and can be managed via the existing `SecretEngineMount` tune config or a future story. Out of scope.
- **DO NOT** reuse `ConsulSEConfig` struct — the Nomad config has different fields than Consul (no scheme, has max_token_name_length). Create a new `NomadSEConfig` inline struct.
- **DO NOT** modify existing Consul or AWS secret engine code — this is a new type, not a modification
- **DO NOT** modify shared framework behavior (reconcile_skeleton.go, vaultresourcereconciler.go, etc.) — only add new files
- **DO NOT** include the `Name` field on `NomadSecretEngineConfigSpec` — the Vault path is fixed at `{path}/config/access` with no name suffix
- **DO NOT** derive expected payloads from `toMap()` in unit tests — construct independent Vault-read-shaped fixtures
- **DO NOT** forget to add new CRD YAML files to `config/crd/kustomization.yaml` after `make manifests`
- **DO NOT** forget the `type` → `token_type` field rename in `IsEquivalentToDesiredState` — this is the most critical Vault API gotcha for this type

### Project Structure Notes

- All new files follow existing naming conventions: lowercase type name for files
- New files go in established directories (`api/v1alpha1/`, `internal/controller/`, `test/`)
- Test fixture directory `test/nomadsecretengine/` follows the existing pattern (`test/consulsecretengine/`, `test/awssecretengine/`)
- Integration manifests in `integration/nomad/` follow the pattern from `integration/ldap/`, `integration/consul/`
- No conflicts with existing code — purely additive

### References

- [Source: _bmad-output/planning-artifacts/epics.md#Epic-13, Story 13.1 — requirements and acceptance criteria]
- [Source: _bmad-output/project-context.md — all project rules, patterns, and conventions]
- [Source: api/v1alpha1/consulsecretengineconfig_types.go — closest analog: config/access with management token and TLS]
- [Source: api/v1alpha1/consulsecretenginerole_types.go — simple role pattern reference]
- [Source: internal/controller/consulsecretengineconfig_controller.go — always-write controller with credential watches]
- [Source: internal/controller/consulsecretenginerole_controller.go — standard role controller]
- [Source: api/v1alpha1/awssecretenginerole_types.go — field rename mapping pattern (credential_type → credential_types)]
- [Source: api/v1alpha1/payload_filter.go — filterPayloadToDesiredKeys and removeUnsetFields helpers]
- [Source: docs/engine-doc-template.md — documentation template for new engine docs]
- [Source: Vault Nomad Secrets API — https://developer.hashicorp.com/vault/api-docs/secret/nomad]

## Code Review Record

### Review Model Used

GPT-5.4 (Cursor)

### Review Findings

- [ ] [Review][Patch] Shared helper behavior changed outside story scope [`api/v1alpha1/payload_filter.go:23`]
- [ ] [Review][Patch] `deploy-nomad` bootstrap is still one-shot and can fail nondeterministically on a fresh cluster [`Makefile:220`]
- [ ] [Review][Patch] Role update integration test assumes condition ordering when checking `ObservedGeneration` [`internal/controller/nomadsecretengine_controller_test.go:161`]

#### Iteration 5 (Final)

- [ ] [Review][Patch] `max_token_name_length` handling still contradicts the story contract [`api/v1alpha1/nomadsecretengineconfig_types.go:105`]
- [ ] [Review][Patch] Credential delete-trigger reconcile path is still untested [`internal/controller/nomadsecretengine_controller_test.go:191`]

### Decisions Needed / Decisions Taken

None after triage. The acceptance-audit concern around AC2 was classified as a concrete test/verification gap rather than a requirements ambiguity.

### Fixes Applied

- Iteration 1 fixes verified present: the Nomad creds endpoint is now exercised, policies are compared order-insensitively, empty tokens are rejected during credential preparation, TLS redaction tests exist, and documentation/examples now use composed Vault paths.
- Iteration 2 fixes verified present: `deploy-nomad` now fails fast on token/bootstrap errors, the role update path is covered by integration tests, and `json.Number` zero-value handling was added for Nomad numeric drift comparison.
- Iteration 3 fixes verified present: the bootstrap path now retries, shared framework behavior was not changed, and the role update test now uses condition lookup instead of assuming condition ordering.
- Iteration 4 runtime fixes verified present: watched credential `Secret` and `RandomSecret` deletes now pass the controller predicates, and dedicated `max_token_name_length` tests were added. Final review still found one remaining semantic issue in the max-token drift expectation and one remaining coverage gap for delete-triggered reconcile behavior.

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6 (via Cursor)

### Debug Log References

- Initial `make test` failed: `global` bool field in NomadSERole not handled by `removeUnsetFields` (bools not covered). Fixed by adding explicit `global` false-value handling in `IsEquivalentToDesiredState`, following the Consul `local` field pattern.

### Completion Notes List

- NomadSecretEngineConfig implements the always-write controller pattern (token is write-only) with Secret/RandomSecret watches, following ConsulSecretEngineConfig exactly
- NomadSecretEngineRole implements standard VaultResource pattern with `type` → `token_type` field remapping in `IsEquivalentToDesiredState` (Vault API returns `token_type` on read but accepts `type` on write)
- `policies` emitted as `toInterfaceArray()` matching Vault's array read format
- `max_token_name_length` emitted as `json.Number` per project conventions, only when non-zero
- Config `IsDeletable()=false` (no DELETE endpoint for `/nomad/config/access`)
- Role `IsDeletable()=true` with standard delete via `DELETE /nomad/role/{name}`
- Integration test verifies config persists in Vault after CR deletion (non-deletable contract)
- All 30+ unit tests pass, no regressions in existing tests

### Change Log

- 2026-08-12: Implemented NomadSecretEngineConfig and NomadSecretEngineRole CRDs with webhooks, controllers, unit tests, integration tests, and documentation

### File List

- api/v1alpha1/nomadsecretengineconfig_types.go (NEW)
- api/v1alpha1/nomadsecretengineconfig_webhook.go (NEW)
- api/v1alpha1/nomadsecretengineconfig_test.go (NEW)
- api/v1alpha1/nomadsecretenginerole_types.go (NEW)
- api/v1alpha1/nomadsecretenginerole_webhook.go (NEW)
- api/v1alpha1/nomadsecretenginerole_test.go (NEW)
- api/v1alpha1/zz_generated.deepcopy.go (MODIFIED - auto-generated)
- internal/controller/nomadsecretengineconfig_controller.go (NEW)
- internal/controller/nomadsecretenginerole_controller.go (NEW)
- internal/controller/nomadsecretengine_controller_test.go (NEW)
- cmd/main.go (MODIFIED - added controller and webhook registrations)
- config/crd/bases/redhatcop.redhat.io_nomadsecretengineconfigs.yaml (NEW - auto-generated)
- config/crd/bases/redhatcop.redhat.io_nomadsecretengineroles.yaml (NEW - auto-generated)
- config/crd/kustomization.yaml (MODIFIED - added new CRD resources)
- config/rbac/role.yaml (MODIFIED - auto-generated RBAC for new controllers)
- config/webhook/manifests.yaml (MODIFIED - auto-generated webhook configs)
- test/nomadsecretengine/nomad-secret-engine-config.yaml (NEW)
- test/nomadsecretengine/nomad-secret-engine-role.yaml (NEW)
- test/nomadsecretengine/nomad-secret-engine-mount.yaml (NEW)
- integration/nomad/deployment.yaml (NEW)
- integration/nomad/service.yaml (NEW)
- integration/nomad/configmap.yaml (NEW)
- Makefile (MODIFIED - added deploy-nomad target)
- docs/secret-engines/nomad.md (NEW)
- docs/secret-engines/index.md (MODIFIED - added Nomad link)
