# Story 11.1: AWS Secret Engine — Config and Role CRDs

Status: ready-for-dev

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As an operator developer,
I want CRDs for AWSSecretEngineConfig (root credentials config) and AWSSecretEngineRole (IAM user, assumed role, federation token),
So that Vault's AWS secret engine can be managed declaratively.

## Acceptance Criteria

1. **Given** an AWSSecretEngineConfig CR is created with AWS access key credentials (via K8s Secret reference) **When** the reconciler processes it **Then** the root config is written to Vault at `{path}/config/root` and ReconcileSuccessful=True

2. **Given** an AWSSecretEngineRole CR is created with `credential_type` set to one of `iam_user`, `assumed_role`, `federation_token`, or `session_token` **When** the reconciler processes it **Then** the role exists in Vault at `{path}/roles/{name}` and can generate dynamic AWS credentials

3. **Given** the AWSSecretEngineConfig CR is deleted **When** the reconciler processes deletion **Then** the K8s object is removed but Vault root config is **NOT** deleted (`IsDeletable=false` — Vault has no `DELETE /aws/config/root` endpoint)

4. **Given** the AWSSecretEngineRole CR is deleted **When** the reconciler processes deletion **Then** the role is removed from Vault via `DELETE /aws/roles/{name}` and the CR is deleted from K8s

5. **Given** the AWSSecretEngineRole CR spec is updated (e.g., `policyArns` changed) **When** the reconciler processes the update **Then** the Vault role reflects the updated value

6. **Given** any AWS CR is created or updated **When** the webhook validates it **Then** `spec.path` immutability is enforced on updates, and credential source validation passes for config

## Tasks / Subtasks

- [ ] Task 1: Create `AWSSecretEngineConfig` type (AC: 1, 3, 6)
  - [ ] 1.1: Create `api/v1alpha1/awssecretengineconfig_types.go` — Spec with `Connection`, `Authentication`, `Path`, inline `AWSRootConfig` struct, `RootCredentials` (RootCredentialConfig), `Name`
  - [ ] 1.2: Implement `VaultObject` interface — `GetPath()` returns `{path}/config/root`, `GetPayload()`, `IsEquivalentToDesiredState()`, `PrepareInternalValues()`, `IsDeletable()=false`
  - [ ] 1.3: Implement `ConditionsAware` interface — Status with `Conditions []metav1.Condition`
  - [ ] 1.4: Implement `setInternalCredentials()` — resolve AWS access_key/secret_key from K8s Secret, VaultSecret, or RandomSecret
  - [ ] 1.5: Implement `toMap()` on `AWSRootConfig` — convert to Vault API snake_case fields
  - [ ] 1.6: Implement `IsEquivalentToDesiredState()` — must delete `secret_key` from desired state (Vault never returns it on read), then `filterPayloadToDesiredKeys`

- [ ] Task 2: Create `AWSSecretEngineRole` type (AC: 2, 4, 5)
  - [ ] 2.1: Create `api/v1alpha1/awssecretenginerole_types.go` — Spec with `Connection`, `Authentication`, `Path`, inline `AWSRole` struct, `Name`
  - [ ] 2.2: Implement `VaultObject` interface — `GetPath()` returns `{path}/roles/{name}`, `IsDeletable()=true`
  - [ ] 2.3: Implement `ConditionsAware` interface
  - [ ] 2.4: Implement `toMap()` on `AWSRole` — handle conditional fields based on `credential_type`
  - [ ] 2.5: Implement `IsEquivalentToDesiredState()` — handle `credential_types` (plural array) vs `credential_type` (singular string) mapping from Vault read response

- [ ] Task 3: Create webhooks (AC: 6)
  - [ ] 3.1: Create `api/v1alpha1/awssecretengineconfig_webhook.go` — `admission.Defaulter[*AWSSecretEngineConfig]`, `admission.Validator[*AWSSecretEngineConfig]`, immutable `spec.path`, credential validation via `ValidateCredentialSource()`
  - [ ] 3.2: Create `api/v1alpha1/awssecretenginerole_webhook.go` — `admission.Defaulter[*AWSSecretEngineRole]`, `admission.Validator[*AWSSecretEngineRole]`, immutable `spec.path`

- [ ] Task 4: Create controllers (AC: 1, 2, 3, 4, 5)
  - [ ] 4.1: Create `internal/controller/awssecretengineconfig_controller.go` — embed `ReconcilerBase`, standard reconcile flow with `NewVaultResource`, watches on `corev1.Secret` and `RandomSecret`
  - [ ] 4.2: Create `internal/controller/awssecretenginerole_controller.go` — embed `ReconcilerBase`, standard reconcile flow with `NewVaultResource`

- [ ] Task 5: Register in main.go (AC: 1, 2)
  - [ ] 5.1: Add controller registrations for `AWSSecretEngineConfigReconciler` and `AWSSecretEngineRoleReconciler`
  - [ ] 5.2: Add webhook registrations inside `ENABLE_WEBHOOKS` guard for both types

- [ ] Task 6: Unit tests (AC: 1, 2, 5, 6)
  - [ ] 6.1: Create `api/v1alpha1/awssecretengineconfig_test.go` — test `toMap()` output, test `IsEquivalentToDesiredState()` with independently constructed Vault-read-shaped fixtures (including secret_key stripping), negative test proving managed field mismatch returns `false`
  - [ ] 6.2: Create `api/v1alpha1/awssecretenginerole_test.go` — test `toMap()` for each credential_type, test `IsEquivalentToDesiredState()` with `credential_types` (plural array) mapping, negative tests
  - [ ] 6.3: Webhook validation tests (immutable path, credential source validation)

- [ ] Task 7: Test fixtures and integration tests (AC: 1, 2, 4, 5)
  - [ ] 7.1: Create test YAML fixtures in `test/awssecretengine/` — config and role CRs
  - [ ] 7.2: Integration tests — **SKIP** (see Dev Notes: AWS is a cloud provider, cannot run in Kind, falls under "Skip it" per integration test philosophy)

- [ ] Task 8: Code generation and validation (AC: all)
  - [ ] 8.1: Run `make manifests generate fmt vet test`
  - [ ] 8.2: Verify all existing tests still pass

## Dev Notes

### Integration Test Classification: SKIP

Per the project's Integration Test Infrastructure Philosophy:
> **Skip it** — For cloud providers and services that cannot be installed in Kind and are hard to mock (e.g., AWS, Azure, GCP, GitHub App, Quay), the integration test must not be run. These types rely on unit test coverage only.

AWS falls squarely under "Skip it". No integration tests should be written. Comprehensive unit tests for `toMap()`, `IsEquivalentToDesiredState()`, and webhook validation are the primary quality gate.

### Vault API Paths

| Operation | Method | Path |
|-----------|--------|------|
| Write root config | POST | `/aws/config/root` |
| Read root config | GET | `/aws/config/root` |
| Write lease config | POST | `/aws/config/lease` |
| Read lease config | GET | `/aws/config/lease` |
| Create/update role | POST | `/aws/roles/:name` |
| Read role | GET | `/aws/roles/:name` |
| Delete role | DELETE | `/aws/roles/:name` |

**No DELETE endpoint exists for `/aws/config/root`** → `AWSSecretEngineConfig.IsDeletable()` must return `false`.

### AWSSecretEngineConfig — Vault API Field Reference

**Write (`POST /aws/config/root`) fields:**
- `access_key` (string) — AWS access key ID
- `secret_key` (string) — AWS secret access key
- `region` (string) — AWS region, defaults to `us-east-1`
- `iam_endpoint` (string) — custom IAM endpoint
- `sts_endpoint` (string) — custom STS endpoint
- `max_retries` (int, default -1) — max retries for recoverable errors
- `username_template` (string) — Go template for dynamic username generation

**Read (`GET /aws/config/root`) response — `secret_key` is NEVER returned:**
```json
{
  "data": {
    "access_key": "AKIAEXAMPLE",
    "region": "us-west-2",
    "iam_endpoint": "https://iam.amazonaws.com",
    "sts_endpoint": "https://sts.us-west-2.amazonaws.com",
    "max_retries": -1
  }
}
```

### Critical: `IsEquivalentToDesiredState` for Config — Secret Key Stripping

Vault never returns `secret_key` on read. The `IsEquivalentToDesiredState` implementation must:
1. Build `desiredState` from `toMap()`
2. `delete(desiredState, "secret_key")` — remove it before comparison
3. Also `delete(desiredState, "access_key")` — the access_key may differ if credentials were rotated; this field should be excluded from drift detection (similar to how DatabaseSecretEngineConfig deletes `password`)
4. Use `filterPayloadToDesiredKeys(desiredState, payload)` then `reflect.DeepEqual`

Follow the established pattern from `GitHubSecretEngineConfig` (deletes `prv_key`), `KubernetesSecretEngineConfig` (deletes `service_account_jwt`), `LDAPAuthEngineConfig` (deletes `bindpass`).

### AWSSecretEngineConfig — Credential Resolution via RootCredentialConfig

The `access_key` and `secret_key` must be resolved from one of three sources via `RootCredentialConfig`:
- **K8s Secret**: keys `accessKey` (or custom via `UsernameKey`) and `secretKey` (or custom via `PasswordKey`)
- **VaultSecret**: same key mapping from a Vault KV path
- **RandomSecret**: retrieve password from RandomSecret's Vault path (access_key must be in `spec.accessKey` field)

Pattern: follow `DatabaseSecretEngineConfig.setInternalCredentials()` exactly. Store resolved values in unexported fields (`retrievedAccessKey`, `retrievedSecretKey` with `json:"-"`) and include them in `toMap()` output.

The `RootCredentialConfig.UsernameKey` default is `"username"` and `PasswordKey` default is `"password"`. For AWS, the user may want different key names — the existing `RootCredentialConfig` supports custom keys via `usernameKey`/`passwordKey` fields.

### AWSSecretEngineConfig — `GetPath()` Is Fixed

Unlike DatabaseSecretEngineConfig which uses `{path}/config/{name}`, the AWS root config endpoint is always `{path}/config/root` (no per-name suffix). `GetPath()` must return `CleansePath(string(d.Spec.Path) + "/config/root")`.

The `spec.name` field is NOT needed for AWSSecretEngineConfig since the path is fixed. However, keep the `Name` field for consistency with the operator pattern (it will simply be unused in path construction for this type).

### AWSSecretEngineRole — Vault API Field Reference

**Write (`POST /aws/roles/:name`) fields:**
- `credential_type` (string, required) — one of: `iam_user`, `assumed_role`, `federation_token`, `session_token`
- `role_arns` (list) — required when `credential_type=assumed_role`, prohibited otherwise
- `policy_arns` (list) — AWS managed policy ARNs; valid for `iam_user`, `assumed_role`, `federation_token`; disallowed for `session_token`
- `policy_document` (string) — IAM policy JSON; valid for `iam_user`, `assumed_role`, `federation_token`; disallowed for `session_token`
- `iam_groups` (list) — IAM group names; valid for all credential types that support policies
- `iam_tags` (list) — key=value tags for `iam_user` type
- `default_sts_ttl` (string) — default TTL for STS creds; valid only for `assumed_role` or `federation_token`
- `max_sts_ttl` (string) — max TTL for STS creds; valid only for `assumed_role` or `federation_token`
- `user_path` (string) — IAM user path; valid only for `iam_user`
- `permissions_boundary_arn` (string) — permissions boundary; valid only for `iam_user`
- `external_id` (string) — external ID for assume role; valid only for `assumed_role`
- `session_tags` (list) — STS session tags; valid only for `assumed_role`
- `mfa_serial_number` (string) — MFA device ARN

### Critical: `IsEquivalentToDesiredState` for Role — `credential_types` Mapping

**Vault API gotcha**: The write endpoint accepts `credential_type` (singular string), but the read endpoint returns `credential_types` (plural, as an array):
```json
{
  "data": {
    "credential_types": ["assumed_role"],
    "role_arns": ["arn:aws:iam::123456789012:role/DeveloperRole"],
    "policy_arns": [],
    "policy_document": "",
    "iam_groups": []
  }
}
```

The `IsEquivalentToDesiredState` implementation must:
1. Build `desiredState` from `toMap()` (which uses `credential_type` singular)
2. Convert the `credential_type` key to `credential_types` as `[]any{credentialType}` to match Vault's read format
3. Delete the original `credential_type` key from desiredState
4. Use `filterPayloadToDesiredKeys(desiredState, payload)` then `reflect.DeepEqual`

Unit tests MUST verify this mapping with independently constructed Vault-read-shaped fixtures.

### AWSSecretEngineRole — `toMap()` Conditional Fields

`toMap()` must only include fields relevant to the `credential_type`. For example:
- `role_arns` should only be included if `credential_type=assumed_role` and the field is non-empty
- `default_sts_ttl`/`max_sts_ttl` should only be included if `credential_type` is `assumed_role` or `federation_token`
- `user_path`/`permissions_boundary_arn` should only be included if `credential_type=iam_user`

However, follow the project pattern: include all fields in `toMap()` and let `filterPayloadToDesiredKeys` handle the comparison. Vault ignores irrelevant fields on write, and `filterPayloadToDesiredKeys` will only compare keys that Vault returns on read. The simpler approach is: always emit all set fields in `toMap()`, and rely on `filterPayloadToDesiredKeys` for drift detection.

Actually, the safer approach: only include non-zero-value optional fields in `toMap()`. This avoids sending empty arrays or empty strings that Vault might interpret differently. Follow the conditional inclusion pattern used in `DatabaseSecretEngineConfig.toMap()` where `password` and `username` are conditionally included.

### AWSSecretEngineRole — `GetPath()`

Returns `CleansePath(string(d.Spec.Path) + "/roles/" + name)` where name is `d.Spec.Name` if set, otherwise `d.Name`. Standard pattern from `DatabaseSecretEngineRole`.

### File Structure

| File | Action | Description |
|------|--------|-------------|
| `api/v1alpha1/awssecretengineconfig_types.go` | NEW | Config CRD type, VaultObject, ConditionsAware, toMap, credential resolution |
| `api/v1alpha1/awssecretengineconfig_webhook.go` | NEW | Config webhook — defaulter, validator, immutable path |
| `api/v1alpha1/awssecretenginerole_types.go` | NEW | Role CRD type, VaultObject, ConditionsAware, toMap with conditional fields |
| `api/v1alpha1/awssecretenginerole_webhook.go` | NEW | Role webhook — defaulter, validator, immutable path |
| `internal/controller/awssecretengineconfig_controller.go` | NEW | Config reconciler with Secret/RandomSecret watches |
| `internal/controller/awssecretenginerole_controller.go` | NEW | Role reconciler — standard VaultResource pattern |
| `cmd/main.go` | UPDATE | Register both controllers + both webhooks |
| `api/v1alpha1/awssecretengineconfig_test.go` | NEW | Unit tests for config toMap, IsEquivalentToDesiredState |
| `api/v1alpha1/awssecretenginerole_test.go` | NEW | Unit tests for role toMap, IsEquivalentToDesiredState |
| `test/awssecretengine/` | NEW | Test YAML fixtures for unit tests |

### Files Being Modified — Current State

**`cmd/main.go`**: Currently registers ~30+ controllers and ~30+ webhooks. New registrations follow the exact same pattern:
- Controller: `(&controller.AWSSecretEngineConfigReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "AWSSecretEngineConfig")}).SetupWithManager(mgr)`
- Webhook (inside ENABLE_WEBHOOKS): `(&redhatcopv1alpha1.AWSSecretEngineConfig{}).SetupWebhookWithManager(mgr)`

No existing behavior is changed — this is purely additive.

### Existing Patterns to Follow (Reference Files)

| Pattern | Reference File |
|---------|---------------|
| Config type with RootCredentialConfig | `api/v1alpha1/databasesecretengineconfig_types.go` |
| Config credential resolution | `DatabaseSecretEngineConfig.setInternalCredentials()` |
| Secret-stripping in IsEquivalentToDesiredState | `api/v1alpha1/githubsecretengineconfig_types.go` (deletes `prv_key`) |
| Role type (simple, no credentials) | `api/v1alpha1/databasesecretenginerole_types.go` |
| Webhook pattern | `api/v1alpha1/databasesecretengineconfig_webhook.go` |
| Controller with credential watches | `internal/controller/quaysecretengineconfig_controller.go` |
| Controller (simple role) | `internal/controller/databasesecretenginerole_controller.go` |
| filterPayloadToDesiredKeys helper | `api/v1alpha1/payload_filter.go` |
| Unit test payload construction | Project context: never derive expected from code under test |

### CRD Field Spec — AWSSecretEngineConfig

```go
type AWSRootConfig struct {
    // AccessKey is the AWS access key ID (set via spec or resolved from credentials)
    AccessKey string `json:"accessKey,omitempty"`
    // Region is the AWS region (defaults to us-east-1 if not set)
    Region string `json:"region,omitempty"`
    // IAMEndpoint is a custom HTTP IAM endpoint
    IAMEndpoint string `json:"iamEndpoint,omitempty"`
    // STSEndpoint is a custom HTTP STS endpoint
    STSEndpoint string `json:"stsEndpoint,omitempty"`
    // MaxRetries is the max retries for recoverable errors (-1 = SDK default)
    MaxRetries *int `json:"maxRetries,omitempty"`
    // UsernameTemplate is a Go template for dynamic username generation
    UsernameTemplate string `json:"usernameTemplate,omitempty"`

    // internal fields (not serialized)
    retrievedAccessKey string `json:"-"`
    retrievedSecretKey string `json:"-"`
}
```

### CRD Field Spec — AWSSecretEngineRole

```go
type AWSRole struct {
    // CredentialType specifies the type of credential (iam_user, assumed_role, federation_token, session_token)
    CredentialType string `json:"credentialType"`
    // RoleArns specifies ARNs of AWS roles (required for assumed_role)
    RoleArns []string `json:"roleArns,omitempty"`
    // PolicyArns specifies AWS managed policy ARNs
    PolicyArns []string `json:"policyArns,omitempty"`
    // PolicyDocument is the IAM policy document JSON
    PolicyDocument string `json:"policyDocument,omitempty"`
    // IAMGroups specifies IAM group names
    IAMGroups []string `json:"iamGroups,omitempty"`
    // IAMTags specifies key=value tags for iam_user
    IAMTags []string `json:"iamTags,omitempty"`
    // DefaultSTSTTL is the default TTL for STS credentials
    DefaultSTSTTL string `json:"defaultSTSTTL,omitempty"`
    // MaxSTSTTL is the max TTL for STS credentials
    MaxSTSTTL string `json:"maxSTSTTL,omitempty"`
    // UserPath is the IAM user path (iam_user only)
    UserPath string `json:"userPath,omitempty"`
    // PermissionsBoundaryARN is the permissions boundary (iam_user only)
    PermissionsBoundaryARN string `json:"permissionsBoundaryARN,omitempty"`
    // ExternalID is the external ID for assume role
    ExternalID string `json:"externalID,omitempty"`
    // SessionTags are STS session tags (assumed_role only)
    SessionTags []string `json:"sessionTags,omitempty"`
    // MFASerialNumber is the MFA device ARN
    MFASerialNumber string `json:"mfaSerialNumber,omitempty"`
}
```

### Kubebuilder Markers Checklist

- All exported Spec fields: `+kubebuilder:validation:Required` or `+kubebuilder:validation:Optional`
- `CredentialType`: `+kubebuilder:validation:Enum:={"iam_user","assumed_role","federation_token","session_token"}`
- List fields (roleArns, policyArns, etc.): `+listType=set`
- Root type: `+kubebuilder:object:root=true`, `+kubebuilder:subresource:status`
- Conditions field: `+patchMergeKey=type`, `+patchStrategy=merge`, `+listType=map`, `+listMapKey=type`

### Unit Test Requirements

**Config tests (`awssecretengineconfig_test.go`):**
1. `TestAWSSecretEngineConfig_toMap` — verify snake_case keys, verify all set fields appear
2. `TestAWSSecretEngineConfig_IsEquivalentToDesiredState_Match` — construct Vault-read-shaped payload (no `secret_key`, no `access_key`), verify returns `true`
3. `TestAWSSecretEngineConfig_IsEquivalentToDesiredState_Mismatch` — change a managed field (e.g., `region`), verify returns `false`
4. `TestAWSSecretEngineConfig_IsEquivalentToDesiredState_ExtraVaultFields` — add extra Vault fields, verify still returns `true` (extra fields ignored)

**Role tests (`awssecretenginerole_test.go`):**
1. `TestAWSSecretEngineRole_toMap_IAMUser` — verify fields for iam_user type
2. `TestAWSSecretEngineRole_toMap_AssumedRole` — verify fields for assumed_role type
3. `TestAWSSecretEngineRole_toMap_FederationToken` — verify fields for federation_token type
4. `TestAWSSecretEngineRole_IsEquivalentToDesiredState_Match` — construct Vault-read fixture with `credential_types` (plural array), verify returns `true`
5. `TestAWSSecretEngineRole_IsEquivalentToDesiredState_Mismatch` — change a field, verify returns `false`
6. `TestAWSSecretEngineRole_IsEquivalentToDesiredState_CredentialTypesMapping` — explicitly test the singular-to-plural mapping

### RBAC Markers

Config controller:
```go
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=awssecretengineconfigs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=awssecretengineconfigs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=awssecretengineconfigs/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=serviceaccounts/token,verbs=create
//+kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;patch
```

Role controller:
```go
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=awssecretengineroles,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=awssecretengineroles/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=awssecretengineroles/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=serviceaccounts/token,verbs=create
//+kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;patch
```

### Project Structure Notes

- All new files follow existing naming conventions: lowercase type name for files
- New files go in established directories (`api/v1alpha1/`, `internal/controller/`, `test/`)
- Test fixture directory `test/awssecretengine/` follows the existing pattern (`test/databasesecretengine/`, `test/rabbitmqsecretengine/`)
- No conflicts with existing code — purely additive

### References

- [Source: _bmad-output/planning-artifacts/epics.md#Epic-11, Story 11.1 — requirements and acceptance criteria]
- [Source: _bmad-output/project-context.md — all project rules, patterns, and conventions]
- [Source: api/v1alpha1/databasesecretengineconfig_types.go — config type with RootCredentialConfig pattern]
- [Source: api/v1alpha1/databasesecretenginerole_types.go — role type pattern]
- [Source: api/v1alpha1/databasesecretengineconfig_webhook.go — webhook pattern]
- [Source: internal/controller/quaysecretengineconfig_controller.go — controller with credential watches pattern]
- [Source: api/v1alpha1/payload_filter.go — filterPayloadToDesiredKeys helper]
- [Source: api/v1alpha1/utils/commons.go — RootCredentialConfig, VaultConnection, KubeAuthConfiguration]
- [Source: Vault AWS API docs — https://developer.hashicorp.com/vault/api-docs/secret/aws]

## Dev Agent Record

### Agent Model Used

### Debug Log References

### Completion Notes List

### File List
