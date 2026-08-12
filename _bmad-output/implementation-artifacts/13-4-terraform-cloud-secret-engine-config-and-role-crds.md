# Story 13.4: Terraform Cloud Secret Engine — Config and Role CRDs

Status: ready-for-dev

## Story

As an operator developer,
I want CRDs for TerraformCloudSecretEngineConfig and TerraformCloudSecretEngineRole,
So that Vault's Terraform Cloud secret engine can be managed declaratively.

## Acceptance Criteria

1. **Given** a TerraformCloudSecretEngineConfig CR is created with a Terraform Cloud API token (via K8s Secret reference) **When** the reconciler processes it **Then** the config is written to Vault at `{path}/config` and ReconcileSuccessful=True

2. **Given** a TerraformCloudSecretEngineRole CR is created with credential_type and the appropriate identifier (organization, team_id, or user_id) **When** the reconciler processes it **Then** the role exists in Vault at `{path}/role/{name}` and can generate dynamic Terraform Cloud API tokens

3. **Given** the TerraformCloudSecretEngineConfig CR is deleted **When** the reconciler processes deletion **Then** the K8s object is removed but Vault config is NOT deleted (`IsDeletable=false` — Vault has no explicit DELETE /terraform/config endpoint)

4. **Given** the TerraformCloudSecretEngineRole CR is deleted **When** the reconciler processes deletion **Then** the role is removed from Vault via `DELETE {path}/role/{name}`

5. **Given** any Terraform Cloud secret engine CR spec is updated **When** the reconciler processes the update **Then** the Vault resource reflects the updated values

6. **Given** any Terraform Cloud secret engine CR is created or updated **When** the webhook validates it **Then** `spec.path` immutability is enforced on updates, `spec.name` immutability is enforced on updates, and credential source validation passes for config

7. **Given** the CRD types are implemented **When** the story is marked done **Then** a documentation file exists at `docs/secret-engines/terraform-cloud.md` following `docs/engine-doc-template.md` (DNFR5)

## Tasks / Subtasks

- [ ] Task 1: Create `TerraformCloudSecretEngineConfig` type (AC: 1, 3, 5, 6)
  - [ ] 1.1: Create `api/v1alpha1/terraformcloudsecretengineconfig_types.go` — Spec with `Connection`, `Authentication`, `Path`, inline `TFCSEConfig` struct, `TFCCredentials` (credential config for token), `Name`
  - [ ] 1.2: Implement `VaultObject` interface — `GetPath()` returns `{path}/config`, `GetPayload()`, `IsEquivalentToDesiredState()`, `PrepareInternalValues()`, `IsDeletable()=false`
  - [ ] 1.3: Implement `ConditionsAware` interface — Status with `Conditions []metav1.Condition`
  - [ ] 1.4: Implement `setInternalCredentials()` — resolve token from K8s Secret, VaultSecret, or RandomSecret (single-field credential, passwordKey defaults to "token")
  - [ ] 1.5: Implement `toMap()` on `TFCSEConfig` — convert to Vault API snake_case fields (`address`, `token`)
  - [ ] 1.6: Implement `IsEquivalentToDesiredState()` — must delete `token` from desired state (Vault never returns it on read), then `filterPayloadToDesiredKeys`

- [ ] Task 2: Create `TerraformCloudSecretEngineRole` type (AC: 2, 4, 5, 6)
  - [ ] 2.1: Create `api/v1alpha1/terraformcloudsecretenginerole_types.go` — Spec with `Connection`, `Authentication`, `Path`, inline `TFCSERole` struct, `Name`
  - [ ] 2.2: Implement `VaultObject` interface — `GetPath()` returns `{path}/role/{name}`, `IsDeletable()=true`
  - [ ] 2.3: Implement `ConditionsAware` interface
  - [ ] 2.4: Implement `toMap()` on `TFCSERole` — handle `ttl`/`max_ttl` as duration strings, emit all identity fields
  - [ ] 2.5: Implement `IsEquivalentToDesiredState()` — Vault returns `ttl`/`max_ttl` as integer seconds and adds `name` field; use `durationToSeconds` for TTLs in toMap, then `removeUnsetFields` + `filterPayloadToDesiredKeys`

- [ ] Task 3: Create webhooks (AC: 6)
  - [ ] 3.1: Create `api/v1alpha1/terraformcloudsecretengineconfig_webhook.go` — `admission.Defaulter[*TerraformCloudSecretEngineConfig]`, `admission.Validator[*TerraformCloudSecretEngineConfig]`, immutable `spec.path`/`spec.name`, credential validation
  - [ ] 3.2: Create `api/v1alpha1/terraformcloudsecretenginerole_webhook.go` — `admission.Defaulter[*TerraformCloudSecretEngineRole]`, `admission.Validator[*TerraformCloudSecretEngineRole]`, immutable `spec.path`/`spec.name`

- [ ] Task 4: Create controllers (AC: 1, 2, 3, 4, 5)
  - [ ] 4.1: Create `internal/controller/terraformcloudsecretengineconfig_controller.go` — embed `ReconcilerBase`, always-write reconcile logic (token is write-only), watches on `corev1.Secret` and `RandomSecret`
  - [ ] 4.2: Create `internal/controller/terraformcloudsecretenginerole_controller.go` — standard `For()` with default periodic reconcile predicate

- [ ] Task 5: Register in main.go (AC: 1, 2)
  - [ ] 5.1: Add controller registrations for both reconcilers
  - [ ] 5.2: Add webhook registrations inside `ENABLE_WEBHOOKS` guard for both types

- [ ] Task 6: Unit tests (AC: 1, 2, 5, 6)
  - [ ] 6.1: Create `api/v1alpha1/terraformcloudsecretengineconfig_test.go` — test `toMap()`, test `IsEquivalentToDesiredState()` with independently constructed Vault-read-shaped fixtures (including token stripping), negative tests
  - [ ] 6.2: Create `api/v1alpha1/terraformcloudsecretenginerole_test.go` — test `toMap()` output with duration conversion, test `IsEquivalentToDesiredState()` with Vault-read fixture (ttl/max_ttl as json.Number), negative tests

- [ ] Task 7: Test fixtures (AC: all)
  - [ ] 7.1: Create test YAML fixtures in `test/terraformcloudsecretengine/` — config and role CRs
  - [ ] 7.2: Integration tests — SKIP (Terraform Cloud is a cloud service, falls under "Skip it" per project integration test philosophy)

- [ ] Task 8: CRD registration and code generation (AC: all)
  - [ ] 8.1: Run `make manifests generate fmt vet test`
  - [ ] 8.2: Add new CRD YAML files to `config/crd/kustomization.yaml` (CRD registration checklist)
  - [ ] 8.3: Verify all existing tests still pass

- [ ] Task 9: Documentation (AC: 7)
  - [ ] 9.1: Create `docs/secret-engines/terraform-cloud.md` following `docs/engine-doc-template.md`
  - [ ] 9.2: Update `docs/secret-engines/index.md` with link to new doc

## Dev Notes

### Integration Test Classification: SKIP

Per the project's Integration Test Infrastructure Philosophy:
> **Skip it** — For cloud providers and services that cannot be installed in Kind and are hard to mock (e.g., AWS, Azure, GCP, GitHub App, Quay), the integration test must not be run. These types rely on unit test coverage only.

Terraform Cloud is a SaaS service that cannot be installed in Kind and is not trivially mockable. No integration tests. Comprehensive unit tests for `toMap()`, `IsEquivalentToDesiredState()`, and webhook validation are the primary quality gate.

### Story Intelligence Chain — Previous Story Context

**Epic 12 stories (12.1 Consul, 12.2 GCP, 12.3 LDAP)** are the direct predecessors for secret engine CRD patterns:
- **12.2 GCP** is the closest analog — cloud service config with write-only credential (`credentials`), `IsDeletable=false` for config, role types with TTL duration handling
- **Established pattern:** Config type with credential resolution via custom credential config struct (not `RootCredentialConfig` when the credential shape differs), always-write controller for write-only tokens, role types with standard VaultResource reconcile
- **`removeUnsetFields` + `filterPayloadToDesiredKeys`** pipeline established as standard for all `IsEquivalentToDesiredState` implementations
- **`durationToSeconds` helper** converts duration strings to `json.Number` seconds for TTL fields to match Vault read format
- **Epic 12 retrospective action items** (all applied): `toMap()` normalization rule (emit Vault-read format), 5-iteration review cap, sprint-status atomicity guard, final consistency check

**Epic 11 patterns** (AWS, Transit, SSH):
- **Always-write controller** for config types with write-only credentials — `AWSSecretEngineConfigReconciler.manageReconcileLogic()` is the canonical reference
- **`json.Number` for all numeric fields** in `toMap()` — mandatory per project context
- **CRD registration checklist** — add to `config/crd/kustomization.yaml` after `make manifests`

### Vault API Paths

| Operation | Method | Path |
|-----------|--------|------|
| Write config | POST | `{path}/config` |
| Read config | GET | `{path}/config` |
| Create/update role | POST | `{path}/role/{name}` |
| Read role | GET | `{path}/role/{name}` |
| Delete role | DELETE | `{path}/role/{name}` |
| List roles | LIST | `{path}/role` |
| Generate credential | GET | `{path}/creds/{name}` |

**Note:** There is no `DELETE {path}/config` endpoint. Config is overwritten on POST, never deleted → `IsDeletable()=false`.

### TerraformCloudSecretEngineConfig — Vault API Field Reference

**Write (`POST {path}/config`) fields:**
- `address` (string, default "https://app.terraform.io") — Address of the Terraform Cloud/Enterprise server
- `token` (string) — Terraform Cloud API token with permissions to manage Organization, Team, and User tokens

**Read (`GET {path}/config`) response — `token` is NEVER returned:**
```json
{
  "data": {
    "address": "https://app.terraform.io",
    "base_path": "/api/v2/"
  }
}
```

**Critical:** The read response includes `base_path` (a Vault-computed field not in the write payload). This must be filtered out by `filterPayloadToDesiredKeys`. The `token` field is write-only and must be deleted from `desiredState` before comparison.

### Critical: `IsEquivalentToDesiredState` for Config — Token Stripping

Vault never returns `token` on read. The implementation must:
1. Build `desiredState` from `toMap()`
2. `delete(desiredState, "token")` — remove it before comparison
3. Use `filterPayloadToDesiredKeys(desiredState, payload)` then `reflect.DeepEqual`

Follow the established pattern from `GCPSecretEngineConfig` (deletes `credentials`), `AWSSecretEngineConfig` (deletes `secret_key`).

### TerraformCloudSecretEngineConfig — Credential Resolution

The Terraform Cloud API token must be resolved from one of three sources:
- **K8s Secret**: key from `PasswordKey` (default "token")
- **VaultSecret**: same key mapping from a Vault KV path
- **RandomSecret**: retrieve from RandomSecret's Vault path

Pattern: Use a custom `TFCCredentialConfig` struct similar to `GCPCredentialConfig` (from `gcpsecretengineconfig_types.go`) but with `passwordKey` defaulting to `"token"` instead of `"credentials"`. Store the resolved token in an unexported field (`retrievedToken string` with `json:"-"`) and include it in `toMap()` output as `"token"`.

### TerraformCloudSecretEngineConfig — Controller Uses Always-Write Pattern

Because `token` is write-only (Vault never returns it on read), drift detection cannot observe token rotations. The config controller MUST use the always-write pattern from `AWSSecretEngineConfigReconciler.manageReconcileLogic()`: call `PrepareInternalValues` then `vaultEndpoint.Create()` unconditionally (skip the standard `CreateOrUpdate` that reads first).

### TerraformCloudSecretEngineConfig — `GetPath()` Is Fixed

The config is always at `{path}/config` (no per-name suffix). `GetPath()` returns `CleansePath(string(d.Spec.Path) + "/config")`.

The `spec.name` field is kept for consistency but unused in path construction (same as GCPSecretEngineConfig/AWSSecretEngineConfig where `spec.name` exists but config path ignores it).

### TerraformCloudSecretEngineRole — Vault API Field Reference

**Write (`POST {path}/role/{name}`) fields:**
- `organization` (string: "") — Organization name. Conflicts with `user_id`.
- `team_id` (string: "") — Team ID. Conflicts with `user_id`.
- `user_id` (string: "") — User ID. Conflicts with `organization` and `team_id`.
- `credential_type` (string: "") — Type of credential: `team`, `team_legacy`, `user`, or `organization`. Vault sets it automatically if omitted based on which ID field is provided.
- `description` (string: "") — Description prefix for the token in HCP Terraform UI. Applies to User and Team tokens.
- `ttl` (duration: "") — TTL for this role. Uses duration format strings on write.
- `max_ttl` (duration: "") — Max TTL for this role. Uses duration format strings on write.

**Read (`GET {path}/role/{name}`) response:**
```json
{
  "data": {
    "credential_type": "user",
    "description": "description",
    "max_ttl": 86400,
    "name": "tfuser",
    "organization": "",
    "team_id": "",
    "ttl": 3600,
    "user_id": "user-glhf1234"
  }
}
```

**Critical observations:**
- Vault returns `ttl` and `max_ttl` as integer seconds (not duration strings)
- Vault adds a `name` field in the read response (redundant with the URL path) — must be filtered
- Vault returns empty strings for unset identity fields — `removeUnsetFields` must handle this

### TerraformCloudSecretEngineRole — TTL Duration Handling

Vault returns TTLs as integer seconds on read. The `toMap()` must emit TTL values as `json.Number` seconds using `durationToSeconds()` helper. This ensures `IsEquivalentToDesiredState` compares matching types (both `json.Number`).

If the user provides `"1h"` in the CRD, `durationToSeconds("1h")` converts to `json.Number("3600")` in `toMap()`. Vault read also returns `3600` as `json.Number` (via Vault client's `UseNumber()`). The comparison then succeeds.

### CRD Field Spec — TerraformCloudSecretEngineConfig

```go
type TFCSEConfig struct {
    // Address is the URL of the Terraform Cloud/Enterprise instance.
    // +kubebuilder:validation:Optional
    // +kubebuilder:default="https://app.terraform.io"
    Address string `json:"address"`

    retrievedToken string `json:"-"`
}
```

The `TFCSEConfig` struct is minimal — Terraform Cloud config only has `address` and `token` (token comes from credential resolution). This is simpler than GCP/AWS configs.

### CRD Field Spec — TerraformCloudSecretEngineRole

```go
type TFCSERole struct {
    // Organization is the name of the Terraform Cloud organization. Conflicts with UserID.
    // +kubebuilder:validation:Optional
    Organization string `json:"organization,omitempty"`

    // TeamID is the Terraform Cloud team ID. Conflicts with UserID.
    // +kubebuilder:validation:Optional
    TeamID string `json:"teamID,omitempty"`

    // UserID is the Terraform Cloud user ID. Conflicts with Organization and TeamID.
    // +kubebuilder:validation:Optional
    UserID string `json:"userID,omitempty"`

    // CredentialType specifies the type of credential to generate.
    // +kubebuilder:validation:Optional
    // +kubebuilder:validation:Enum:={"team","team_legacy","user","organization"}
    CredentialType string `json:"credentialType,omitempty"`

    // Description is a human-readable description used as a prefix in HCP Terraform UI. Applies to User and Team tokens.
    // +kubebuilder:validation:Optional
    Description string `json:"description,omitempty"`

    // TTL specifies the TTL for generated tokens (duration format, e.g. "1h")
    // +kubebuilder:validation:Optional
    TTL string `json:"ttl,omitempty"`

    // MaxTTL specifies the maximum TTL for generated tokens (duration format, e.g. "24h")
    // +kubebuilder:validation:Optional
    MaxTTL string `json:"maxTTL,omitempty"`
}
```

### TerraformCloudSecretEngineRole — `toMap()` Implementation Notes

```go
func (d *TFCSERole) toMap() map[string]any {
    payload := map[string]any{}
    if d.Organization != "" {
        payload["organization"] = d.Organization
    }
    if d.TeamID != "" {
        payload["team_id"] = d.TeamID
    }
    if d.UserID != "" {
        payload["user_id"] = d.UserID
    }
    if d.CredentialType != "" {
        payload["credential_type"] = d.CredentialType
    }
    if d.Description != "" {
        payload["description"] = d.Description
    }
    if d.TTL != "" {
        payload["ttl"] = durationToSeconds(d.TTL)
    }
    if d.MaxTTL != "" {
        payload["max_ttl"] = durationToSeconds(d.MaxTTL)
    }
    return payload
}
```

Key: TTL fields must use `durationToSeconds()` to emit `json.Number` in Vault-read format. Other string fields are conditional (only include if non-empty).

### TerraformCloudSecretEngineRole — `IsEquivalentToDesiredState` Notes

The read response includes a `name` field (the role name from the URL path). The `filterPayloadToDesiredKeys` helper automatically excludes this because `toMap()` never includes `name` in its output. Standard pipeline: `removeUnsetFields(desiredState, payload)` → `filterPayloadToDesiredKeys(desiredState, payload)` → `reflect.DeepEqual`.

### Kubebuilder Markers Checklist

- All exported Spec fields: `+kubebuilder:validation:Required` or `+kubebuilder:validation:Optional`
- Config `Address`: `+kubebuilder:default="https://app.terraform.io"` (non-zero default → no `omitempty`)
- Role `CredentialType`: `+kubebuilder:validation:Enum:={"team","team_legacy","user","organization"}`
- Root type: `+kubebuilder:object:root=true`, `+kubebuilder:subresource:status`
- Conditions field: `+patchMergeKey=type`, `+patchStrategy=merge`, `+listType=map`, `+listMapKey=type`

### RBAC Markers

Config controller:
```go
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=terraformcloudsecretengineconfigs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=terraformcloudsecretengineconfigs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=terraformcloudsecretengineconfigs/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=serviceaccounts/token,verbs=create
//+kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;patch
```

Role controller:
```go
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=terraformcloudsecretengineroles,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=terraformcloudsecretengineroles/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=terraformcloudsecretengineroles/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=serviceaccounts/token,verbs=create
//+kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;patch
```

### File Structure

| File | Action | Description |
|------|--------|-------------|
| `api/v1alpha1/terraformcloudsecretengineconfig_types.go` | NEW | Config CRD type, VaultObject, ConditionsAware, toMap, credential resolution |
| `api/v1alpha1/terraformcloudsecretengineconfig_webhook.go` | NEW | Config webhook — defaulter, validator, immutable path/name |
| `api/v1alpha1/terraformcloudsecretengineconfig_test.go` | NEW | Unit tests for config toMap, IsEquivalentToDesiredState |
| `api/v1alpha1/terraformcloudsecretenginerole_types.go` | NEW | Role CRD type, VaultObject, ConditionsAware, toMap |
| `api/v1alpha1/terraformcloudsecretenginerole_webhook.go` | NEW | Role webhook — defaulter, validator, immutable path/name |
| `api/v1alpha1/terraformcloudsecretenginerole_test.go` | NEW | Unit tests for role toMap, IsEquivalentToDesiredState |
| `internal/controller/terraformcloudsecretengineconfig_controller.go` | NEW | Config reconciler with always-write pattern, Secret/RandomSecret watches |
| `internal/controller/terraformcloudsecretenginerole_controller.go` | NEW | Role reconciler — standard VaultResource pattern |
| `cmd/main.go` | UPDATE | Register 2 controllers + 2 webhooks |
| `config/crd/kustomization.yaml` | UPDATE | Add 2 new CRD YAML files to resources list |
| `test/terraformcloudsecretengine/` | NEW | Test YAML fixtures for config and role |
| `docs/secret-engines/terraform-cloud.md` | NEW | Engine documentation per DNFR5 |
| `docs/secret-engines/index.md` | UPDATE | Add link to terraform-cloud.md |

### Files Being Modified — Current State

**`cmd/main.go`**: Currently registers ~40+ controllers and ~40+ webhooks. New registrations follow the exact same pattern:
- Controller: `(&controller.TerraformCloudSecretEngineConfigReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "TerraformCloudSecretEngineConfig")}).SetupWithManager(mgr)`
- Webhook (inside ENABLE_WEBHOOKS): `(&redhatcopv1alpha1.TerraformCloudSecretEngineConfig{}).SetupWebhookWithManager(mgr)`

No existing behavior is changed — this is purely additive.

**`config/crd/kustomization.yaml`**: Add the 2 new CRD YAML files to the `resources` list. Required for Helm chart build.

**`docs/secret-engines/index.md`**: Add a link entry for the new Terraform Cloud doc.

### Existing Patterns to Follow (Reference Files)

| Pattern | Reference File |
|---------|---------------|
| Config type with write-only credential (closest analog) | `api/v1alpha1/gcpsecretengineconfig_types.go` |
| Custom credential config struct (non-RootCredentialConfig) | `api/v1alpha1/gcpsecretengineconfig_types.go` (GCPCredentialConfig) |
| Token-stripping in IsEquivalentToDesiredState | `api/v1alpha1/gcpsecretengineconfig_types.go` (deletes `credentials`) |
| Always-write controller for write-only credentials | `internal/controller/awssecretengineconfig_controller.go` |
| Controller with credential watches (Secret + RandomSecret) | `internal/controller/awssecretengineconfig_controller.go` |
| IsDeletable=false config pattern | `api/v1alpha1/gcpsecretengineconfig_types.go` |
| Role type with TTL duration handling | `api/v1alpha1/gcpsecretengineroleset_types.go` |
| Role type (simple, no credentials) | `api/v1alpha1/awssecretenginerole_types.go` |
| Webhook pattern | `api/v1alpha1/gcpsecretengineconfig_webhook.go` |
| Controller (simple role, no watches) | `internal/controller/awssecretenginerole_controller.go` |
| filterPayloadToDesiredKeys helper | `api/v1alpha1/payload_filter.go` |
| removeUnsetFields helper | `api/v1alpha1/payload_filter.go` |
| durationToSeconds helper | `api/v1alpha1/utils/vaultutils.go` (or inline in types files) |
| Unit test payload construction | Project context: never derive expected from code under test |
| Documentation template | `docs/engine-doc-template.md` |

### Unit Test Requirements

**Config tests (`terraformcloudsecretengineconfig_test.go`):**
1. `TestTerraformCloudSecretEngineConfig_toMap` — verify `address` and `token` keys, verify token is from resolved credential
2. `TestTerraformCloudSecretEngineConfig_IsEquivalentToDesiredState_Match` — construct Vault-read-shaped payload (`{"address": "https://app.terraform.io", "base_path": "/api/v2/"}` — note: `base_path` is a Vault extra that gets filtered), verify returns `true`
3. `TestTerraformCloudSecretEngineConfig_IsEquivalentToDesiredState_Mismatch` — change `address`, verify returns `false`
4. `TestTerraformCloudSecretEngineConfig_IsEquivalentToDesiredState_TokenInPayload` — include `token` in Vault payload (shouldn't happen in practice but defensive), verify still returns `true` after filtering

**Role tests (`terraformcloudsecretenginerole_test.go`):**
1. `TestTerraformCloudSecretEngineRole_toMap` — verify `organization`, `team_id`, `user_id`, `credential_type`, `description`, `ttl`/`max_ttl` as `json.Number` seconds
2. `TestTerraformCloudSecretEngineRole_IsEquivalentToDesiredState_Match` — Vault-read fixture with `ttl`/`max_ttl` as `json.Number`, `name` field (filtered), verify returns `true`
3. `TestTerraformCloudSecretEngineRole_IsEquivalentToDesiredState_Mismatch` — change `credential_type`, verify returns `false`
4. `TestTerraformCloudSecretEngineRole_IsEquivalentToDesiredState_ExtraVaultFields` — add `name` field in Vault response, verify still returns `true` (filtered out)

### Anti-Patterns / DO NOT

- **DO NOT** create integration tests for this type — Terraform Cloud is a SaaS service that cannot be installed in Kind (per "Skip it" rule)
- **DO NOT** use `RootCredentialConfig` directly — the Terraform Cloud config only needs a single token field (no username/password pair). Use a custom `TFCCredentialConfig` struct similar to `GCPCredentialConfig`
- **DO NOT** include `base_path` in the CRD spec — this is a Vault-computed field returned on read but never sent on write
- **DO NOT** include the `name` field in `toMap()` for roles — Vault adds this to read responses from the URL path, but it's not a writable field
- **DO NOT** modify shared framework behavior (reconcile_skeleton.go, vaultresourcereconciler.go, etc.) — only add new files
- **DO NOT** derive expected payloads from `toMap()` in unit tests — construct independent Vault-read-shaped fixtures
- **DO NOT** forget to add new CRD YAML files to `config/crd/kustomization.yaml` after `make manifests`
- **DO NOT** use Go `int` or `float64` in unit test Vault payloads — use `json.Number` to match real Vault client behavior
- **DO NOT** emit TTL fields as duration strings in `toMap()` — use `durationToSeconds()` to emit `json.Number` seconds matching Vault's read format

### Project Structure Notes

- All new files follow existing naming conventions: lowercase type name for files
- New files go in established directories (`api/v1alpha1/`, `internal/controller/`, `test/`, `docs/secret-engines/`)
- Test fixture directory `test/terraformcloudsecretengine/` follows the existing pattern (`test/gcpsecretengine/`, `test/awssecretengine/`)
- No conflicts with existing code — purely additive

### References

- [Source: _bmad-output/planning-artifacts/epics.md#Epic-13, Story 13.4 — requirements and acceptance criteria]
- [Source: _bmad-output/project-context.md — all project rules, patterns, and conventions]
- [Source: api/v1alpha1/gcpsecretengineconfig_types.go — GCP secret engine config with credential resolution, IsDeletable=false, credential stripping]
- [Source: api/v1alpha1/awssecretengineconfig_types.go — AWS secret engine config with always-write pattern]
- [Source: internal/controller/awssecretengineconfig_controller.go — always-write controller with credential watches]
- [Source: api/v1alpha1/gcpsecretengineroleset_types.go — GCP roleset with TTL duration handling]
- [Source: api/v1alpha1/payload_filter.go — filterPayloadToDesiredKeys and removeUnsetFields helpers]
- [Source: Vault Terraform Cloud Secret Engine API — https://developer.hashicorp.com/vault/api-docs/secret/terraform]
- [Source: _bmad-output/implementation-artifacts/12-2-gcp-secret-engine-config-and-role-crds.md — closest predecessor story pattern]
- [Source: _bmad-output/implementation-artifacts/12-3-ldap-ad-secret-engine-config-and-role-crds.md — credential resolution and always-write patterns]

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
