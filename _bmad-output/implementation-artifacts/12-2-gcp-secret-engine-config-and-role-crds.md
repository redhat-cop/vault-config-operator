# Story 12.2: GCP Secret Engine — Config and Role CRDs

Status: ready-for-dev

## Story

As an operator developer,
I want CRDs for GCPSecretEngineConfig, GCPSecretEngineRoleset, and GCPSecretEngineStaticAccount,
So that Vault's GCP secret engine can be managed declaratively (complementing the existing GCP auth engine CRDs).

## Acceptance Criteria

1. **Given** a GCPSecretEngineConfig CR is created with GCP credentials (via K8s Secret reference)
   **When** the reconciler processes it
   **Then** the config is written to Vault at `{path}/config` and ReconcileSuccessful=True

2. **Given** a GCPSecretEngineRoleset CR is created with project, bindings, and secret_type
   **When** the reconciler processes it
   **Then** the roleset exists in Vault at `{path}/roleset/{name}` and can generate service account keys or OAuth tokens

3. **Given** a GCPSecretEngineStaticAccount CR is created with service_account_email and bindings
   **When** the reconciler processes it
   **Then** the static account exists in Vault at `{path}/static-account/{name}` and can generate credentials

4. **Given** the GCPSecretEngineConfig CR is deleted
   **When** the reconciler processes deletion
   **Then** the K8s object is removed but Vault config is NOT deleted (`IsDeletable=false` — Vault has no `DELETE /gcp/config` endpoint)

5. **Given** the GCPSecretEngineRoleset or GCPSecretEngineStaticAccount CR is deleted
   **When** the reconciler processes deletion
   **Then** the resource is removed from Vault via the respective DELETE endpoint

6. **Given** any GCP Secret Engine CR is updated
   **When** the webhook validates it
   **Then** `spec.path` immutability is enforced; roleset additionally enforces `spec.name`, `spec.secretType`, `spec.project` immutability; static account additionally enforces `spec.name`, `spec.secretType`, `spec.serviceAccountEmail` immutability

## Tasks / Subtasks

- [ ] Task 1: Create `GCPSecretEngineConfig` type (AC: #1, #4)
  - [ ] 1.1: Create `api/v1alpha1/gcpsecretengineconfig_types.go` — Spec with `Connection`, `Authentication`, `Path`, inline `GCPSEConfig` struct, `GCPCredentials` (RootCredentialConfig)
  - [ ] 1.2: Implement `VaultObject` interface — `GetPath()` returns `{path}/config`, `GetPayload()`, `IsEquivalentToDesiredState()`, `PrepareInternalValues()`, `IsDeletable()=false`
  - [ ] 1.3: Implement `ConditionsAware` interface
  - [ ] 1.4: Implement `setInternalCredentials()` — resolve GCP credentials JSON from K8s Secret, VaultSecret, or RandomSecret (follow existing `GCPAuthEngineConfig.setInternalCredentials()` pattern)
  - [ ] 1.5: Implement `toMap()` on `GCPSEConfig` — convert to Vault API snake_case fields
  - [ ] 1.6: Implement `IsEquivalentToDesiredState()` — must delete `credentials` from desiredState (Vault never returns credentials on read), then `filterPayloadToDesiredKeys`

- [ ] Task 2: Create `GCPSecretEngineRoleset` type (AC: #2, #5, #6)
  - [ ] 2.1: Create `api/v1alpha1/gcpsecretengineroleset_types.go` — Spec with `Connection`, `Authentication`, `Path`, inline `GCPSERoleset` struct, `Name`
  - [ ] 2.2: Implement `VaultObject` interface — `GetPath()` returns `{path}/roleset/{name}`, `IsDeletable()=true`
  - [ ] 2.3: Implement `ConditionsAware` interface
  - [ ] 2.4: Implement `toMap()` on `GCPSERoleset`
  - [ ] 2.5: Implement `IsEquivalentToDesiredState()` — delete `bindings` (write=string, read=map format mismatch) and delete `project` (Vault returns as `service_account_project`), then `removeUnsetFields` + `filterPayloadToDesiredKeys`

- [ ] Task 3: Create `GCPSecretEngineStaticAccount` type (AC: #3, #5, #6)
  - [ ] 3.1: Create `api/v1alpha1/gcpsecretenginestaticaccount_types.go` — Spec with `Connection`, `Authentication`, `Path`, inline `GCPSEStaticAccount` struct, `Name`
  - [ ] 3.2: Implement `VaultObject` interface — `GetPath()` returns `{path}/static-account/{name}`, `IsDeletable()=true`
  - [ ] 3.3: Implement `ConditionsAware` interface
  - [ ] 3.4: Implement `toMap()` on `GCPSEStaticAccount`
  - [ ] 3.5: Implement `IsEquivalentToDesiredState()` — delete `bindings` (format mismatch), then `removeUnsetFields` + `filterPayloadToDesiredKeys`

- [ ] Task 4: Create webhooks (AC: #6)
  - [ ] 4.1: Create `api/v1alpha1/gcpsecretengineconfig_webhook.go` — credential validation, immutable `spec.path`
  - [ ] 4.2: Create `api/v1alpha1/gcpsecretengineroleset_webhook.go` — immutable `spec.path`, `spec.name`, `spec.secretType`, `spec.project`
  - [ ] 4.3: Create `api/v1alpha1/gcpsecretenginestaticaccount_webhook.go` — immutable `spec.path`, `spec.name`, `spec.secretType`, `spec.serviceAccountEmail`

- [ ] Task 5: Create controllers (AC: #1, #2, #3, #4, #5)
  - [ ] 5.1: Create `internal/controller/gcpsecretengineconfig_controller.go` — with K8s Secret and RandomSecret watches (same pattern as AWSSecretEngineConfig controller)
  - [ ] 5.2: Create `internal/controller/gcpsecretengineroleset_controller.go` — standard VaultResource pattern
  - [ ] 5.3: Create `internal/controller/gcpsecretenginestaticaccount_controller.go` — standard VaultResource pattern

- [ ] Task 6: Register in main.go (AC: all)
  - [ ] 6.1: Register all three controllers (outside webhook guard)
  - [ ] 6.2: Register all three webhooks (inside `ENABLE_WEBHOOKS` guard)

- [ ] Task 7: Unit tests (AC: #1, #2, #3, #6)
  - [ ] 7.1: Create `api/v1alpha1/gcpsecretengineconfig_test.go` — toMap, IsEquivalentToDesiredState (match, mismatch, credentials-in-payload)
  - [ ] 7.2: Create `api/v1alpha1/gcpsecretengineroleset_test.go` — toMap, IsEquivalentToDesiredState (match, mismatch, bindings excluded)
  - [ ] 7.3: Create `api/v1alpha1/gcpsecretenginestaticaccount_test.go` — toMap, IsEquivalentToDesiredState (match, mismatch, bindings excluded)

- [ ] Task 8: Test fixtures (AC: all)
  - [ ] 8.1: Create test YAML fixtures in `test/gcpsecretengine/`
  - [ ] 8.2: Integration tests — SKIP (GCP is a cloud provider, falls under "Skip it" per project integration test philosophy)

- [ ] Task 9: Code generation and validation (AC: all)
  - [ ] 9.1: Run `make manifests generate fmt vet`
  - [ ] 9.2: Run `make test` — unit tests pass
  - [ ] 9.3: Add new CRDs to `config/crd/kustomization.yaml`

## Dev Notes

### Integration Test Classification: SKIP

Per the project's Integration Test Infrastructure Philosophy:
> **Skip it** — For cloud providers and services that cannot be installed in Kind and are hard to mock (e.g., AWS, Azure, GCP, GitHub App, Quay), the integration test must not be run. These types rely on unit test coverage only.

GCP falls under "Skip it". No integration tests. Comprehensive unit tests for `toMap()`, `IsEquivalentToDesiredState()`, and webhook validation are the primary quality gate.

### Vault API Reference

**Config endpoint:** `POST /gcp/config`
- `credentials` (string: "") — JSON credentials for GCP service account (file contents or '@path/to/file'). Mutually exclusive with `identity_token_audience`.
- `ttl` (int/string: "0s") — default TTL for long-lived credentials (service account keys). Duration format.
- `max_ttl` (int/string: "0s") — max TTL for long-lived credentials. Duration format.

**Config read response (`GET /gcp/config`):** Returns `{"data":{"ttl":"1h","max_ttl":"4h"}}` — **credentials are OMITTED** from read response.

**No DELETE endpoint for config** → `IsDeletable()=false`.

---

**Roleset endpoint:** `POST /gcp/roleset/:name`
- `secret_type` (string: "access_token") — "access_token" or "service_account_key". **Cannot be updated.**
- `project` (string) — GCP project ID. **Cannot be updated.**
- `bindings` (string) — HCL or JSON format binding configuration.
- `token_scopes` (array: []) — OAuth scopes, only for access_token type.

**Roleset read response (`GET /gcp/roleset/:name`):**
```json
{
  "data": {
    "secret_type": "access_token",
    "service_account_email": "vault-myroleset-XXXXXXXXXX@myproject.iam.gserviceaccount.com",
    "service_account_project": "service-account-project",
    "bindings": {
      "project/mygcpproject": ["roles/viewer"]
    },
    "token_scopes": ["https://www.googleapis.com/auth/cloud-platform"]
  }
}
```

**Roleset DELETE:** `DELETE /gcp/roleset/:name` — supported → `IsDeletable()=true`.

---

**Static Account endpoint:** `POST /gcp/static-account/:name`
- `secret_type` (string: "access_token") — "access_token" or "service_account_key". **Cannot be updated.**
- `service_account_email` (string) — email of existing GCP service account. **Cannot be updated.**
- `bindings` (string) — HCL or JSON format. Optional.
- `token_scopes` (array: []) — OAuth scopes, only for access_token type.

**Static Account read response (`GET /gcp/static-account/:name`):**
```json
{
  "data": {
    "secret_type": "access_token",
    "service_account_email": "example@mygcpproject.iam.gserviceaccount.com",
    "service_account_project": "mygcpproject",
    "bindings": {
      "project/mygcpproject": ["roles/viewer"]
    },
    "token_scopes": ["https://www.googleapis.com/auth/cloud-platform"]
  }
}
```

**Static Account DELETE:** `DELETE /gcp/static-account/:name` — supported → `IsDeletable()=true`.

### Critical Vault API Gotchas

1. **Credentials never returned on config read** — `IsEquivalentToDesiredState` for config MUST delete `credentials` from desiredState before comparison. Same pattern as AWS `secret_key`, GitHub `prv_key`, Kubernetes SE `service_account_jwt`.

2. **`bindings` format mismatch** — Write sends `bindings` as a string (HCL/JSON), but read returns it as a **map** object. `reflect.DeepEqual(string, map)` will always return `false`. `IsEquivalentToDesiredState` for roleset and static account MUST delete `bindings` from desiredState before comparison. Binding drift is detected by spec generation changes (generation-based reconciliation).

3. **Roleset `project` → `service_account_project` key rename** — Write accepts `project` but read returns `service_account_project`. `IsEquivalentToDesiredState` for roleset MUST delete `project` from desiredState since the key name changes.

4. **Immutable fields** — Vault rejects updates to `secret_type`, `project` (roleset), `service_account_email` (static account). These MUST be enforced as immutable in `ValidateUpdate` to give users clear error messages instead of opaque Vault errors.

5. **Config always-write pattern** — Since credentials are write-only and never returned on read, the config controller should use the always-write pattern (like AWSSecretEngineConfig and RabbitMQSecretEngineConfig). The `manageReconcileLogic` should call `vaultEndpoint.Create(ctx)` directly instead of using `CreateOrUpdate` which would always see a drift.

### GCPSEConfig Struct Design

```go
type GCPSEConfig struct {
    // TTL specifies the default TTL for long-lived credentials (service account keys). Duration format.
    // +kubebuilder:validation:Optional
    TTL string `json:"ttl,omitempty"`

    // MaxTTL specifies the maximum TTL for long-lived credentials. Duration format.
    // +kubebuilder:validation:Optional
    MaxTTL string `json:"maxTTL,omitempty"`

    retrievedCredentials string `json:"-"`
}
```

### GCPSecretEngineConfig Spec Design

```go
type GCPSecretEngineConfigSpec struct {
    // Connection represents the information needed to connect to Vault.
    // +kubebuilder:validation:Optional
    Connection *vaultutils.VaultConnection `json:"connection,omitempty"`

    // Authentication is the kube auth configuration to be used to execute this request
    // +kubebuilder:validation:Required
    Authentication vaultutils.KubeAuthConfiguration `json:"authentication,omitempty"`

    // Path at which to make the configuration.
    // The final path in Vault will be {[spec.authentication.namespace]}/{spec.path}/config.
    // +kubebuilder:validation:Required
    Path vaultutils.Path `json:"path,omitempty"`

    // GCPCredentials specifies how to retrieve the GCP credentials JSON.
    // +kubebuilder:validation:Required
    GCPCredentials vaultutils.RootCredentialConfig `json:"gcpCredentials,omitempty"`

    GCPSEConfig `json:",inline"`

    // The name of the object created in Vault. If this is specified it takes precedence over {metadata.name}
    // +kubebuilder:validation:Optional
    // +kubebuilder:validation:Pattern:=`[a-z0-9]([-a-z0-9]*[a-z0-9])?`
    Name string `json:"name,omitempty"`
}
```

### GCPSERoleset Struct Design

```go
type GCPSERoleset struct {
    // SecretType specifies the type of secret generated. Accepted values: access_token, service_account_key.
    // Cannot be updated after creation.
    // +kubebuilder:validation:Required
    // +kubebuilder:validation:Enum:={"access_token","service_account_key"}
    // +kubebuilder:default="access_token"
    SecretType string `json:"secretType"`

    // Project is the GCP project ID that this roleset's service account belongs to. Cannot be updated.
    // +kubebuilder:validation:Required
    Project string `json:"project"`

    // Bindings is the bindings configuration string in HCL or JSON format.
    // +kubebuilder:validation:Optional
    Bindings string `json:"bindings,omitempty"`

    // TokenScopes is a list of OAuth scopes for access_token type rolesets only.
    // +kubebuilder:validation:Optional
    // +listType=set
    TokenScopes []string `json:"tokenScopes,omitempty"`
}
```

### GCPSEStaticAccount Struct Design

```go
type GCPSEStaticAccount struct {
    // SecretType specifies the type of secret generated. Accepted values: access_token, service_account_key.
    // Cannot be updated after creation.
    // +kubebuilder:validation:Required
    // +kubebuilder:validation:Enum:={"access_token","service_account_key"}
    // +kubebuilder:default="access_token"
    SecretType string `json:"secretType"`

    // ServiceAccountEmail is the email of the GCP service account to manage. Cannot be updated.
    // +kubebuilder:validation:Required
    ServiceAccountEmail string `json:"serviceAccountEmail"`

    // Bindings is the bindings configuration string in HCL or JSON format. Optional.
    // +kubebuilder:validation:Optional
    Bindings string `json:"bindings,omitempty"`

    // TokenScopes is a list of OAuth scopes for access_token type static accounts only.
    // +kubebuilder:validation:Optional
    // +listType=set
    TokenScopes []string `json:"tokenScopes,omitempty"`
}
```

### Config PrepareInternalValues — Credential Resolution Pattern

Follow the existing `GCPAuthEngineConfig.setInternalCredentials()` pattern from `api/v1alpha1/gcpauthengineconfig_types.go`. The GCP secrets engine config uses the same credential JSON format.

Key differences from the auth engine config:
- Auth engine uses `GCPCredentials` with `UsernameKey` defaulting to `"serviceaccount"` and `PasswordKey` defaulting to `"credentials"`. The secrets engine should use the same default keys for consistency.
- Auth engine stores both `retrievedServiceAccount` and `retrievedCredentials`. The secrets engine config only needs `retrievedCredentials` (the `credentials` field in `toMap()`), and optionally `ttl`/`max_ttl`.

The `toMap()` for config should output:
```go
func (i *GCPSEConfig) toMap() map[string]any {
    payload := map[string]any{}
    payload["credentials"] = i.retrievedCredentials
    payload["ttl"] = i.TTL
    payload["max_ttl"] = i.MaxTTL
    return payload
}
```

### Config IsEquivalentToDesiredState

```go
func (d *GCPSecretEngineConfig) IsEquivalentToDesiredState(payload map[string]any) bool {
    desiredState := d.Spec.GCPSEConfig.toMap()
    delete(desiredState, "credentials")
    removeUnsetFields(desiredState, payload)
    return reflect.DeepEqual(desiredState, filterPayloadToDesiredKeys(desiredState, payload))
}
```

### Config Controller — Always-Write Pattern

The config controller MUST use the always-write pattern because `credentials` is write-only (Vault never returns it on read). This means `IsEquivalentToDesiredState` will only compare `ttl`/`max_ttl`, and credentials updates are always written. Follow the `AWSSecretEngineConfigReconciler.manageReconcileLogic()` pattern:

```go
func (r *GCPSecretEngineConfigReconciler) manageReconcileLogic(ctx context.Context, instance client.Object) error {
    log := log.FromContext(ctx)
    if err := instance.(vaultutils.VaultObject).PrepareInternalValues(ctx, instance); err != nil {
        log.Error(err, "unable to prepare internal values", "instance", instance)
        return err
    }
    vaultEndpoint := vaultutils.NewVaultEndpoint(instance)
    if err := vaultEndpoint.Create(ctx); err != nil {
        log.Error(err, "unable to create/update vault resource", "instance", instance)
        return err
    }
    return nil
}
```

### Roleset IsEquivalentToDesiredState

```go
func (d *GCPSecretEngineRoleset) IsEquivalentToDesiredState(payload map[string]any) bool {
    desiredState := d.Spec.GCPSERoleset.toMap()
    delete(desiredState, "bindings")
    delete(desiredState, "project")
    removeUnsetFields(desiredState, payload)
    return reflect.DeepEqual(desiredState, filterPayloadToDesiredKeys(desiredState, payload))
}
```

### Static Account IsEquivalentToDesiredState

```go
func (d *GCPSecretEngineStaticAccount) IsEquivalentToDesiredState(payload map[string]any) bool {
    desiredState := d.Spec.GCPSEStaticAccount.toMap()
    delete(desiredState, "bindings")
    removeUnsetFields(desiredState, payload)
    return reflect.DeepEqual(desiredState, filterPayloadToDesiredKeys(desiredState, payload))
}
```

### Roleset and Static Account GetPath

```go
// Roleset
func (d *GCPSecretEngineRoleset) GetPath() string {
    if d.Spec.Name != "" {
        return vaultutils.CleansePath(string(d.Spec.Path) + "/" + "roleset" + "/" + d.Spec.Name)
    }
    return vaultutils.CleansePath(string(d.Spec.Path) + "/" + "roleset" + "/" + d.Name)
}

// Static Account
func (d *GCPSecretEngineStaticAccount) GetPath() string {
    if d.Spec.Name != "" {
        return vaultutils.CleansePath(string(d.Spec.Path) + "/" + "static-account" + "/" + d.Spec.Name)
    }
    return vaultutils.CleansePath(string(d.Spec.Path) + "/" + "static-account" + "/" + d.Name)
}
```

### Webhook Immutability Rules

**Config webhook:** Reject `spec.path` and `spec.name` changes in `ValidateUpdate`.

**Roleset webhook:** Reject `spec.path`, `spec.name`, `spec.secretType`, `spec.project` changes in `ValidateUpdate`. Vault itself rejects updates to these fields, but the webhook gives users clear error messages.

**Static Account webhook:** Reject `spec.path`, `spec.name`, `spec.secretType`, `spec.serviceAccountEmail` changes in `ValidateUpdate`.

### toMap Implementations

**Roleset toMap:**
```go
func (i *GCPSERoleset) toMap() map[string]any {
    payload := map[string]any{}
    payload["secret_type"] = i.SecretType
    payload["project"] = i.Project
    payload["bindings"] = i.Bindings
    payload["token_scopes"] = toInterfaceArray(i.TokenScopes)
    return payload
}
```

**Static Account toMap:**
```go
func (i *GCPSEStaticAccount) toMap() map[string]any {
    payload := map[string]any{}
    payload["secret_type"] = i.SecretType
    payload["service_account_email"] = i.ServiceAccountEmail
    payload["bindings"] = i.Bindings
    payload["token_scopes"] = toInterfaceArray(i.TokenScopes)
    return payload
}
```

### File Structure

| File | Action | Description |
|------|--------|-------------|
| `api/v1alpha1/gcpsecretengineconfig_types.go` | NEW | Config CRD type, VaultObject, ConditionsAware, toMap, credential resolution |
| `api/v1alpha1/gcpsecretengineconfig_webhook.go` | NEW | Config webhook — defaulter, validator, immutable path |
| `api/v1alpha1/gcpsecretengineconfig_test.go` | NEW | Unit tests for config toMap, IsEquivalentToDesiredState |
| `api/v1alpha1/gcpsecretengineroleset_types.go` | NEW | Roleset CRD type, VaultObject, ConditionsAware, toMap |
| `api/v1alpha1/gcpsecretengineroleset_webhook.go` | NEW | Roleset webhook — defaulter, validator, immutable fields |
| `api/v1alpha1/gcpsecretengineroleset_test.go` | NEW | Unit tests for roleset toMap, IsEquivalentToDesiredState |
| `api/v1alpha1/gcpsecretenginestaticaccount_types.go` | NEW | Static Account CRD type, VaultObject, ConditionsAware, toMap |
| `api/v1alpha1/gcpsecretenginestaticaccount_webhook.go` | NEW | Static Account webhook — defaulter, validator, immutable fields |
| `api/v1alpha1/gcpsecretenginestaticaccount_test.go` | NEW | Unit tests for static account toMap, IsEquivalentToDesiredState |
| `internal/controller/gcpsecretengineconfig_controller.go` | NEW | Config reconciler with Secret/RandomSecret watches |
| `internal/controller/gcpsecretengineroleset_controller.go` | NEW | Roleset reconciler — standard VaultResource pattern |
| `internal/controller/gcpsecretenginestaticaccount_controller.go` | NEW | Static Account reconciler — standard VaultResource pattern |
| `cmd/main.go` | UPDATE | Register 3 controllers + 3 webhooks |
| `config/crd/kustomization.yaml` | UPDATE | Add 3 new CRD YAML resources |
| `test/gcpsecretengine/` | NEW | Test YAML fixtures |

### Files Being Modified — Current State

**`cmd/main.go`**: Currently registers ~40+ controllers and webhooks (including AWS, Transit, SSH from Epic 11). New registrations follow the exact same pattern. No existing behavior changes — purely additive.

**`config/crd/kustomization.yaml`**: Must add the 3 new CRD YAML files after `make manifests`. Missing this step causes operator crash-loops in Helm-deployed environments (Epic 11 lesson learned).

### Existing Patterns to Follow (Reference Files)

| Pattern | Reference File |
|---------|---------------|
| Config type with RootCredentialConfig (GCP-specific) | `api/v1alpha1/gcpauthengineconfig_types.go` |
| GCP credential resolution (setInternalCredentials) | `api/v1alpha1/gcpauthengineconfig_types.go:204-266` |
| Config type with RootCredentialConfig (modern pattern) | `api/v1alpha1/awssecretengineconfig_types.go` |
| Config controller with credential watches | `internal/controller/awssecretengineconfig_controller.go` |
| Config always-write reconcile pattern | `internal/controller/awssecretengineconfig_controller.go:78-98` |
| Secret-stripping in IsEquivalentToDesiredState | `api/v1alpha1/awssecretengineconfig_types.go:109-113` |
| Role type with Name override | `api/v1alpha1/awssecretenginerole_types.go` |
| Simple role controller | `internal/controller/awssecretenginerole_controller.go` |
| Webhook pattern with multiple immutable fields | `api/v1alpha1/awssecretengineconfig_webhook.go` |
| filterPayloadToDesiredKeys helper | `api/v1alpha1/payload_filter.go` |
| removeUnsetFields helper | `api/v1alpha1/payload_filter.go:21-41` |
| toInterfaceArray helper | `api/v1alpha1/utils/commons.go` (or search for `toInterfaceArray`) |

### Unit Test Requirements

**Config tests (`gcpsecretengineconfig_test.go`):**
1. `TestGCPSecretEngineConfig_toMap` — verify snake_case keys, verify `credentials` populated from retrieved field
2. `TestGCPSecretEngineConfig_IsEquivalentToDesiredState_Match` — construct Vault-read-shaped payload (no `credentials`), verify returns `true`
3. `TestGCPSecretEngineConfig_IsEquivalentToDesiredState_Mismatch` — change `ttl`, verify returns `false`
4. `TestGCPSecretEngineConfig_IsEquivalentToDesiredState_CredentialsInPayload` — payload with `credentials` still returns `true` (extra fields ignored)

**Roleset tests (`gcpsecretengineroleset_test.go`):**
1. `TestGCPSecretEngineRoleset_toMap` — verify all fields
2. `TestGCPSecretEngineRoleset_IsEquivalentToDesiredState_Match` — construct Vault-read fixture with `bindings` as map (not string), without `project` (returns as `service_account_project`), verify returns `true`
3. `TestGCPSecretEngineRoleset_IsEquivalentToDesiredState_Mismatch` — change `secret_type`, verify returns `false`
4. `TestGCPSecretEngineRoleset_IsEquivalentToDesiredState_ExtraVaultFields` — add `service_account_email`, `service_account_project` (Vault-added), verify returns `true`

**Static Account tests (`gcpsecretenginestaticaccount_test.go`):**
1. `TestGCPSecretEngineStaticAccount_toMap` — verify all fields
2. `TestGCPSecretEngineStaticAccount_IsEquivalentToDesiredState_Match` — construct Vault-read fixture with `bindings` as map, verify returns `true`
3. `TestGCPSecretEngineStaticAccount_IsEquivalentToDesiredState_Mismatch` — change `service_account_email`, verify returns `false`

### RBAC Markers

Config controller:
```go
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=gcpsecretengineconfigs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=gcpsecretengineconfigs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=gcpsecretengineconfigs/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=serviceaccounts/token,verbs=create
//+kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;patch
```

Roleset controller:
```go
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=gcpsecretenginerolesets,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=gcpsecretenginerolesets/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=gcpsecretenginerolesets/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=serviceaccounts/token,verbs=create
//+kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;patch
```

Static Account controller:
```go
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=gcpsecretenginestaticaccounts,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=gcpsecretenginestaticaccounts/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=gcpsecretenginestaticaccounts/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=serviceaccounts/token,verbs=create
//+kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;patch
```

### Kubebuilder Markers Checklist

- All exported Spec fields: `+kubebuilder:validation:Required` or `+kubebuilder:validation:Optional`
- `SecretType`: `+kubebuilder:validation:Enum:={"access_token","service_account_key"}`, `+kubebuilder:default="access_token"`
- List fields (tokenScopes): `+listType=set`
- Root type: `+kubebuilder:object:root=true`, `+kubebuilder:subresource:status`
- Conditions field: `+patchMergeKey=type`, `+patchStrategy=merge`, `+listType=map`, `+listMapKey=type`

### Story Intelligence Chain

**Predecessor context (Epic 11 — Stories 11.1, 11.2, 11.3):**
- All 3 Epic 11 stories (AWS, Transit, SSH) followed the same pattern: config type + role/key type, controller, webhook, unit tests, test fixtures
- AWS (11.1) established: always-write config pattern for write-only credentials, `removeUnsetFields` for optional field handling, `credential_type` → `credential_types` singular-to-plural mapping
- Transit (11.2) established: key type with multiple sub-types (encryption, signing, etc.)
- SSH (11.3) established: conditional toMap based on key_type, private_key stripping pattern
- All 3 used the "Skip it" integration test classification for cloud providers
- CRD registration in `config/crd/kustomization.yaml` is mandatory — validated in Epic 11 retro
- `spec.name` immutability must be enforced alongside `spec.path`

**Story 12.1 (Consul, backlog):** Not yet created. No predecessor file in this epic.

**Patterns established by Epic 11 stories that MUST be followed:**
- Shared framework files are off-limits for behavioral changes
- `json.Number` for numeric Vault-facing fields (N/A here — GCP secrets engine only uses strings/arrays)
- Nil maps return empty `map[string]any{}` (N/A — no map fields in these types)
- `toInterfaceArray` for `[]string` → `[]any` conversion (used for `token_scopes`)

### Anti-Patterns / DO NOT

- **DO NOT** modify shared framework files (`reconcile_skeleton.go`, `vaultresourcereconciler.go`, `vaultobject.go`, `vaultutils.go`, `payload_filter.go`, `commons.go`) — only add new functions/types if needed
- **DO NOT** add integration tests — GCP is a cloud provider (Skip it category)
- **DO NOT** try to parse HCL bindings for IsEquivalentToDesiredState comparison — delete `bindings` from desiredState instead
- **DO NOT** compare `project` field in roleset IsEquivalentToDesiredState — Vault returns it as `service_account_project` (different key)
- **DO NOT** forget to add CRDs to `config/crd/kustomization.yaml` after `make manifests`
- **DO NOT** change `toInterfaceArray` or `filterPayloadToDesiredKeys` behavior — these are shared helpers
- **DO NOT** add `IsValid()` validation for `bindings` format (HCL/JSON parsing adds complexity; let Vault validate)
- **DO NOT** use `json.Number` for TTL fields — the GCP secrets engine `ttl`/`max_ttl` config fields are strings, not numeric

### Project Structure Notes

- All files follow existing naming conventions exactly
- Controllers in `internal/controller/` (go/v4 layout since Epic 10)
- CRD types and webhooks in `api/v1alpha1/`
- Test fixtures in `test/gcpsecretengine/` (new directory, follows `test/awssecretengine/` pattern)
- No conflicts with existing code — entirely additive
- Existing `gcpauthengineconfig_types.go` and `gcpauthenginerole_types.go` are for the GCP **auth** engine — these are separate types in a different Vault subsystem

### References

- [Source: _bmad-output/planning-artifacts/epics.md#Epic-12, Story 12.2 — requirements and acceptance criteria]
- [Source: _bmad-output/project-context.md — all project rules, patterns, and conventions]
- [Source: api/v1alpha1/gcpauthengineconfig_types.go — GCP auth engine config with credential resolution pattern]
- [Source: api/v1alpha1/awssecretengineconfig_types.go — AWS secret engine config with RootCredentialConfig, always-write pattern]
- [Source: api/v1alpha1/awssecretenginerole_types.go — AWS role type with removeUnsetFields pattern]
- [Source: internal/controller/awssecretengineconfig_controller.go — config controller with Secret/RandomSecret watches, always-write reconcile]
- [Source: internal/controller/awssecretenginerole_controller.go — simple role controller]
- [Source: api/v1alpha1/payload_filter.go — filterPayloadToDesiredKeys and removeUnsetFields helpers]
- [Vault API: https://developer.hashicorp.com/vault/api-docs/secret/gcp — GCP secrets engine API reference]
- [Vault Docs: https://developer.hashicorp.com/vault/docs/secrets/gcp — GCP secrets engine usage guide]

## Code Review Record

### Review Model Used

(To be filled after code review — must use a different model than the dev model)

### Review Findings

(To be filled after code review)

### Decisions Needed / Decisions Taken

(To be filled after code review)

### Fixes Applied

(To be filled after code review)

## Dev Agent Record

### Agent Model Used

{{agent_model_name_version}}

### Debug Log References

### Completion Notes List

### File List
