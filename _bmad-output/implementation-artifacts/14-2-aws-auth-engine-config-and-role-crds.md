# Story 14.2: AWS Auth Engine — Config and Role CRDs

Status: ready-for-dev

## Story

As an operator developer,
I want CRDs for AWSAuthEngineClientConfig, AWSAuthEngineIdentityConfig, and AWSAuthEngineRole,
So that Vault's AWS auth method can be managed declaratively.

## Acceptance Criteria

1. **Given** an AWSAuthEngineClientConfig CR is created with AWS credentials and STS endpoint **When** the reconciler processes it **Then** the client config is written to Vault at `auth/{path}/config/client` and ReconcileSuccessful=True

2. **Given** an AWSAuthEngineIdentityConfig CR is created with iam_alias/ec2_alias settings **When** the reconciler processes it **Then** the identity config is written to Vault at `auth/{path}/config/identity` and ReconcileSuccessful=True

3. **Given** an AWSAuthEngineRole CR is created with auth_type (iam | ec2) and bound constraints **When** the reconciler processes it **Then** the role exists in Vault at `auth/{path}/role/{name}` and ReconcileSuccessful=True

4. **Given** the AWSAuthEngineClientConfig CR is deleted **When** the reconciler processes deletion **Then** the client config is removed from Vault via `DELETE auth/{path}/config/client` (`IsDeletable=true` — Vault has an explicit DELETE endpoint)

5. **Given** the AWSAuthEngineIdentityConfig CR is deleted **When** the reconciler processes deletion **Then** the K8s object is removed but Vault identity config is NOT deleted (`IsDeletable=false` — no DELETE endpoint for config/identity)

6. **Given** the AWSAuthEngineRole CR is deleted **When** the reconciler processes deletion **Then** the role is removed from Vault via `DELETE auth/{path}/role/{name}`

7. **Given** any AWS auth engine CR spec is updated **When** the reconciler processes the update **Then** the Vault resource reflects the updated values

8. **Given** any AWS auth engine CR is created or updated **When** the webhook validates it **Then** `spec.path` immutability is enforced on updates, `spec.name` immutability is enforced on updates (for types with Name field), and credential source validation passes for client config

9. **Given** the CRD types are implemented **When** the story is marked done **Then** a documentation file exists at `docs/auth-engines/aws.md` following `docs/engine-doc-template.md` (DNFR5)

## Tasks / Subtasks

- [ ] Task 1: Create `AWSAuthEngineClientConfig` type (AC: 1, 4, 7, 8)
  - [ ] 1.1: Create `api/v1alpha1/awsauthengineconfig_types.go` — Spec with `Connection`, `Authentication`, `Path`, inline `AWSAuthClientConfig` struct, `AWSCredentials` (`RootCredentialConfig` with usernameKey="access_key", passwordKey="secret_key")
  - [ ] 1.2: Implement `VaultObject` interface — `GetPath()` returns `auth/{path}/config/client`, `GetPayload()`, `IsEquivalentToDesiredState()`, `PrepareInternalValues()`, `IsDeletable()=true`
  - [ ] 1.3: Implement `ConditionsAware` interface — Status with `Conditions []metav1.Condition`
  - [ ] 1.4: Implement `setInternalCredentials()` — resolve access_key/secret_key from K8s Secret, VaultSecret, or RandomSecret (follow `AWSSecretEngineConfig` pattern)
  - [ ] 1.5: Implement `toMap()` on `AWSAuthClientConfig` — convert to Vault API snake_case fields
  - [ ] 1.6: Implement `IsEquivalentToDesiredState()` — must delete `secret_key` from desired state (Vault never returns it on read), then `filterPayloadToDesiredKeys`

- [ ] Task 2: Create `AWSAuthEngineIdentityConfig` type (AC: 2, 5, 7, 8)
  - [ ] 2.1: Create `api/v1alpha1/awsauthengineidentityconfig_types.go` — Spec with `Connection`, `Authentication`, `Path`, inline `AWSAuthIdentityConfig` struct
  - [ ] 2.2: Implement `VaultObject` interface — `GetPath()` returns `auth/{path}/config/identity`, `IsDeletable()=false`, no `PrepareInternalValues` needed
  - [ ] 2.3: Implement `ConditionsAware` interface
  - [ ] 2.4: Implement `toMap()` on `AWSAuthIdentityConfig` — emit iam_alias, iam_metadata, ec2_alias, ec2_metadata
  - [ ] 2.5: Implement `IsEquivalentToDesiredState()` — standard `filterPayloadToDesiredKeys`

- [ ] Task 3: Create `AWSAuthEngineRole` type (AC: 3, 6, 7, 8)
  - [ ] 3.1: Create `api/v1alpha1/awsauthenginerole_types.go` — Spec with `Connection`, `Authentication`, `Path`, inline `AWSAuthRole` struct, `Name`
  - [ ] 3.2: Implement `VaultObject` interface — `GetPath()` returns `auth/{path}/role/{name}`, `IsDeletable()=true`
  - [ ] 3.3: Implement `ConditionsAware` interface
  - [ ] 3.4: Implement `toMap()` on `AWSAuthRole` — handle auth_type, bound constraints (all as `toInterfaceArray`), token fields; use conditional inclusion for optional fields
  - [ ] 3.5: Implement `IsEquivalentToDesiredState()` — `removeUnsetFields` + `filterPayloadToDesiredKeys`

- [ ] Task 4: Create webhooks (AC: 8)
  - [ ] 4.1: Create `api/v1alpha1/awsauthengineconfig_webhook.go` — `admission.Defaulter[*AWSAuthEngineClientConfig]`, `admission.Validator[*AWSAuthEngineClientConfig]`, immutable `spec.path`, credential validation via `ValidateCredentialSource()`
  - [ ] 4.2: Create `api/v1alpha1/awsauthengineidentityconfig_webhook.go` — `admission.Defaulter[*AWSAuthEngineIdentityConfig]`, `admission.Validator[*AWSAuthEngineIdentityConfig]`, immutable `spec.path`
  - [ ] 4.3: Create `api/v1alpha1/awsauthenginerole_webhook.go` — `admission.Defaulter[*AWSAuthEngineRole]`, `admission.Validator[*AWSAuthEngineRole]`, immutable `spec.path`/`spec.name`, `auth_type` validation

- [ ] Task 5: Create controllers (AC: 1, 2, 3, 4, 5, 6, 7)
  - [ ] 5.1: Create `internal/controller/awsauthengineconfig_controller.go` — embed `ReconcilerBase`, standard VaultResource reconcile, watches on `corev1.Secret` and `RandomSecret` for credential rotation
  - [ ] 5.2: Create `internal/controller/awsauthengineidentityconfig_controller.go` — standard `For()` with default periodic reconcile predicate (simple, no watches)
  - [ ] 5.3: Create `internal/controller/awsauthenginerole_controller.go` — standard `For()` with default periodic reconcile predicate (simple, no watches)

- [ ] Task 6: Register in main.go (AC: 1, 2, 3)
  - [ ] 6.1: Add controller registrations for all 3 reconcilers
  - [ ] 6.2: Add webhook registrations inside `ENABLE_WEBHOOKS` guard for all 3 types

- [ ] Task 7: Unit tests (AC: 1, 2, 3, 7, 8)
  - [ ] 7.1: Create `api/v1alpha1/awsauthengineconfig_test.go` — test `toMap()`, test `IsEquivalentToDesiredState()` with independently constructed Vault-read-shaped fixtures (including secret_key stripping), negative tests
  - [ ] 7.2: Create `api/v1alpha1/awsauthengineidentityconfig_test.go` — test `toMap()`, test `IsEquivalentToDesiredState()`, negative tests
  - [ ] 7.3: Create `api/v1alpha1/awsauthenginerole_test.go` — test `toMap()` output for both iam and ec2 auth_type roles, test `IsEquivalentToDesiredState()` with Vault-read fixture, negative tests for constraint validation

- [ ] Task 8: Test fixtures (AC: all)
  - [ ] 8.1: Create test YAML fixtures in `test/awsauthengine/` — client config, identity config, and role CRs
  - [ ] 8.2: Integration tests — SKIP (AWS is a cloud provider, falls under "Skip it" per project integration test philosophy)

- [ ] Task 9: CRD registration and code generation (AC: all)
  - [ ] 9.1: Run `make manifests generate fmt vet test`
  - [ ] 9.2: Add 3 new CRD YAML files to `config/crd/kustomization.yaml` (CRD registration checklist)
  - [ ] 9.3: Verify all existing tests still pass

- [ ] Task 10: Documentation (AC: 9)
  - [ ] 10.1: Create `docs/auth-engines/aws.md` following `docs/engine-doc-template.md`
  - [ ] 10.2: Update `docs/auth-engines/index.md` with link to new doc

## Dev Notes

### Integration Test Classification: SKIP

Per the project's Integration Test Infrastructure Philosophy:
> **Skip it** — For cloud providers and services that cannot be installed in Kind and are hard to mock (e.g., AWS, Azure, GCP, GitHub App, Quay), the integration test must not be run. These types rely on unit test coverage only.

AWS is a cloud provider that cannot be installed in Kind. No integration tests. Comprehensive unit tests for `toMap()`, `IsEquivalentToDesiredState()`, and webhook validation are the primary quality gate.

### Story Intelligence Chain — Previous Story Context

**Story 14.1 (AppRole Auth Engine)** is the direct predecessor in Epic 14:
- Established the pattern for auth engine CRDs in this epic
- AppRole had no config endpoint (mount is the config) — this story differs significantly with **two** config endpoints plus a role
- No predecessor story file exists yet (14.1 may be in-progress or not started), so patterns come from prior epics

**Epic 13 stories (13.0–13.4)** are the most recent completed CRD implementation stories:
- **13.4 Terraform Cloud** is the closest completed analog — established: custom credential config, token-stripping in IsEquivalentToDesiredState, always-write controller for write-only credentials, `durationToSeconds` for TTL fields
- **Retrospective action items (pending):** validation matrix for multi-mode types, novelty risk field, Phase B finalization — adopt the validation matrix for AWSAuthEngineRole which has two auth_types with different valid constraints

**Epic 11 (AWS Secret Engine)** provides the closest AWS-specific patterns:
- `AWSSecretEngineConfig` — credential resolution via `RootCredentialConfig` with `access_key`/`secret_key`, `SetAccessKeyAndSecretKey()`, `secret_key` stripping in IsEquivalentToDesiredState
- `AWSSecretEngineRole` — conditional field validation based on `credentialType`, `sortAnyStringSlice` for set comparison, `removeUnsetFields + filterPayloadToDesiredKeys` pipeline

**Epic 12 (GCP Secret Engine)** and existing auth engines (GCP, Azure):
- `GCPAuthEngineConfig` — auth engine config with `IsDeletable()=false`, credential resolution via `GCPCredentials` (`RootCredentialConfig`), `GetPath()` returns `auth/{path}/config`
- `GCPAuthEngineRole` — auth engine role with `GetPath()` returns `auth/{path}/role/{name}`, `IsDeletable()=true`, `Name` field for Vault object name override
- Controller pattern: config controller has Secret/RandomSecret watches; role controller is simple `For()` only

### Design Decision: Two Config CRDs (Resolved)

Per sprint-status action item (Epic 13, status: done): **Two separate CRDs** for AWS auth config:
- `AWSAuthEngineClientConfig` → `auth/{path}/config/client` (AWS credentials, STS/IAM endpoints)
- `AWSAuthEngineIdentityConfig` → `auth/{path}/config/identity` (alias and metadata settings)

This maps 1:1 to Vault's API surface. A single merged CRD would require the operator to write to two separate Vault endpoints, complicating the reconcile flow and breaking the standard VaultObject contract where one CR maps to one Vault path.

### Vault API Paths

| Operation | Method | Path |
|-----------|--------|------|
| Write client config | POST | `auth/{path}/config/client` |
| Read client config | GET | `auth/{path}/config/client` |
| Delete client config | DELETE | `auth/{path}/config/client` |
| Write identity config | POST | `auth/{path}/config/identity` |
| Read identity config | GET | `auth/{path}/config/identity` |
| Create/update role | POST | `auth/{path}/role/{name}` |
| Read role | GET | `auth/{path}/role/{name}` |
| Delete role | DELETE | `auth/{path}/role/{name}` |
| List roles | LIST | `auth/{path}/roles` |

### AWSAuthEngineClientConfig — Vault API Field Reference

**Write (`POST auth/{path}/config/client`) fields:**
- `access_key` (string: "") — AWS Access key for API calls
- `secret_key` (string: "") — AWS Secret key for API calls
- `endpoint` (string: "") — Custom endpoint for EC2 API calls
- `iam_endpoint` (string: "") — Custom endpoint for IAM API calls
- `sts_endpoint` (string: "") — Custom endpoint for STS API calls
- `sts_region` (string: "") — Region for STS API calls (should be set with sts_endpoint)
- `use_sts_region_from_client` (bool: false) — Use client request region for STS
- `iam_server_id_header_value` (string: "") — Required X-Vault-AWS-IAM-Server-ID header value
- `allowed_sts_header_values` (string: "") — Additional permitted STS headers
- `max_retries` (int: -1) — Max retries for recoverable errors

**Read (`GET auth/{path}/config/client`) response — `secret_key` is NEVER returned:**
```json
{
  "data": {
    "access_key": "VKIAJBRHKH6EVTTNXDHA",
    "endpoint": "",
    "iam_endpoint": "",
    "sts_endpoint": "",
    "sts_region": "",
    "use_sts_region_from_client": false,
    "iam_server_id_header_value": ""
  }
}
```

**Critical:** `secret_key` is write-only. Must be deleted from `desiredState` before `IsEquivalentToDesiredState` comparison. Follow `AWSSecretEngineConfig` pattern (deletes `secret_key`).

### Critical: `IsEquivalentToDesiredState` for ClientConfig — Secret Key Stripping

Vault never returns `secret_key` on read. The implementation must:
1. Build `desiredState` from `AWSAuthClientConfig.toMap()`
2. `delete(desiredState, "secret_key")` — remove before comparison
3. Use `removeUnsetFields(desiredState, payload)` → `filterPayloadToDesiredKeys(desiredState, payload)` → `reflect.DeepEqual`

Follow the established pattern from `AWSSecretEngineConfig.IsEquivalentToDesiredState()`.

### AWSAuthEngineClientConfig — Credential Resolution

AWS credentials (access_key/secret_key) must be resolved from one of three sources:
- **K8s Secret**: usernameKey defaults to "access_key", passwordKey defaults to "secret_key"
- **VaultSecret**: same key mapping from a Vault KV path
- **RandomSecret**: retrieve from RandomSecret's Vault path (access_key must be in spec if using RandomSecret)

Pattern: Use `RootCredentialConfig` with custom default keys (`usernameKey: "access_key"`, `passwordKey: "secret_key"`). Store resolved values in unexported fields (`retrievedAccessKey`, `retrievedSecretKey` with `json:"-"`). Follow `AWSSecretEngineConfig.setInternalCredentials()` exactly.

### AWSAuthEngineClientConfig — Controller Pattern

The client config controller uses the **standard VaultResource reconcile** (NOT always-write) because the read endpoint returns enough fields for meaningful drift detection (everything except secret_key). The `IsEquivalentToDesiredState` strips `secret_key` to avoid false drift.

The controller MUST include Secret and RandomSecret watches for credential rotation detection, following `GCPAuthEngineConfigReconciler.SetupWithManager()`.

### AWSAuthEngineIdentityConfig — Vault API Field Reference

**Write (`POST auth/{path}/config/identity`) fields:**
- `iam_alias` (string: "role_id") — Identity alias for IAM auth. Choices: role_id, unique_id, canonical_arn, full_arn
- `iam_metadata` (string: "default") — Metadata on login token for IAM auth
- `ec2_alias` (string: "role_id") — Identity alias for EC2 auth. Choices: role_id, instance_id, image_id
- `ec2_metadata` (string: "default") — Metadata on login token for EC2 auth

**Read (`GET auth/{path}/config/identity`) response:**
```json
{
  "data": {
    "iam_alias": "full_arn",
    "iam_metadata": "default",
    "ec2_alias": "role_id",
    "ec2_metadata": "default"
  }
}
```

No write-only fields. Standard `filterPayloadToDesiredKeys` is sufficient.

### AWSAuthEngineIdentityConfig — Simple Controller

No credentials, no watches. Simple `For()` with default periodic reconcile predicate (like `GCPAuthEngineRoleReconciler`).

### AWSAuthEngineRole — Vault API Field Reference

**Write (`POST auth/{path}/role/{name}`) fields:**
- `auth_type` (string: "iam") — Choices: iam, ec2
- `bound_ami_id` (list: []) — EC2/IAM+infer constraint
- `bound_account_id` (list: []) — EC2/IAM+infer constraint
- `bound_region` (list: []) — EC2/IAM+infer constraint
- `bound_vpc_id` (list: []) — EC2/IAM+infer constraint
- `bound_subnet_id` (list: []) — EC2/IAM+infer constraint
- `bound_iam_role_arn` (list: []) — EC2/IAM+infer constraint
- `bound_iam_instance_profile_arn` (list: []) — EC2/IAM+infer constraint
- `bound_ec2_instance_id` (list: []) — EC2/IAM+infer constraint
- `role_tag` (string: "") — EC2 only
- `bound_iam_principal_arn` (list: []) — IAM only
- `inferred_entity_type` (string: "") — IAM only (value: "ec2_instance")
- `inferred_aws_region` (string: "") — IAM only (required with inferred_entity_type)
- `resolve_aws_unique_ids` (bool: true) — IAM only
- `allow_instance_migration` (bool: false) — EC2 only
- `disallow_reauthentication` (bool: false) — EC2 only
- `token_ttl` (integer/string: 0) — Standard token param
- `token_max_ttl` (integer/string: 0) — Standard token param
- `token_policies` (array: []) — Standard token param
- `policies` (array: [], deprecated) — Legacy, keep for compat
- `token_bound_cidrs` (array: []) — Standard token param
- `token_explicit_max_ttl` (integer/string: 0) — Standard token param
- `token_no_default_policy` (bool: false) — Standard token param
- `token_num_uses` (integer: 0) — Standard token param
- `token_period` (integer/string: 0) — Standard token param
- `token_type` (string: "") — Standard token param

**Read (`GET auth/{path}/role/{name}`) sample response:**
```json
{
  "data": {
    "auth_type": "ec2",
    "bound_ami_id": ["ami-fce36987"],
    "role_tag": "",
    "policies": ["default", "dev", "prod"],
    "max_ttl": 1800000,
    "disallow_reauthentication": false,
    "allow_instance_migration": false
  }
}
```

**Critical observations:**
- Vault returns TTL fields (`token_ttl`, `token_max_ttl`, `token_explicit_max_ttl`, `token_period`) as integer seconds — use `durationToSeconds()` in `toMap()`
- Vault returns list constraints as `[]any` — use `toInterfaceArray()` in `toMap()`
- Vault may omit unset fields — use `removeUnsetFields` before comparison

### AWSAuthEngineRole — Validation Matrix (auth_type Constraints)

| Field | EC2 | IAM | IAM+infer |
|-------|-----|-----|-----------|
| bound_ami_id | valid | - | valid |
| bound_account_id | valid | - | valid |
| bound_region | valid | - | valid |
| bound_vpc_id | valid | - | valid |
| bound_subnet_id | valid | - | valid |
| bound_iam_role_arn | valid | - | valid |
| bound_iam_instance_profile_arn | valid | - | valid |
| bound_ec2_instance_id | valid | - | valid |
| role_tag | valid | - | - |
| allow_instance_migration | valid | - | - |
| disallow_reauthentication | valid | - | - |
| bound_iam_principal_arn | - | valid | valid |
| inferred_entity_type | - | valid | - |
| inferred_aws_region | - | valid (with inferred_entity_type) | - |
| resolve_aws_unique_ids | - | valid | valid |

**Webhook validation for role:**
- `auth_type` must be "iam" or "ec2" (enforce via `+kubebuilder:validation:Enum`)
- EC2-only fields (`role_tag`, `allow_instance_migration`, `disallow_reauthentication`) must be empty/false when `auth_type` is "iam" (unless `inferred_entity_type` enables EC2 inference)
- IAM-only fields (`bound_iam_principal_arn`, `inferred_entity_type`, `inferred_aws_region`) must be empty when `auth_type` is "ec2"
- `allow_instance_migration` and `disallow_reauthentication` are mutually exclusive

### AWSAuthEngineRole — `toMap()` Implementation Notes

```go
func (r *AWSAuthRole) toMap() map[string]any {
    payload := map[string]any{}
    payload["auth_type"] = r.AuthType
    payload["bound_ami_id"] = toInterfaceArray(r.BoundAmiID)
    payload["bound_account_id"] = toInterfaceArray(r.BoundAccountID)
    payload["bound_region"] = toInterfaceArray(r.BoundRegion)
    payload["bound_vpc_id"] = toInterfaceArray(r.BoundVpcID)
    payload["bound_subnet_id"] = toInterfaceArray(r.BoundSubnetID)
    payload["bound_iam_role_arn"] = toInterfaceArray(r.BoundIAMRoleARN)
    payload["bound_iam_instance_profile_arn"] = toInterfaceArray(r.BoundIAMInstanceProfileARN)
    payload["bound_ec2_instance_id"] = toInterfaceArray(r.BoundEC2InstanceID)
    payload["role_tag"] = r.RoleTag
    payload["bound_iam_principal_arn"] = toInterfaceArray(r.BoundIAMPrincipalARN)
    payload["inferred_entity_type"] = r.InferredEntityType
    payload["inferred_aws_region"] = r.InferredAWSRegion
    payload["resolve_aws_unique_ids"] = r.ResolveAWSUniqueIDs
    payload["allow_instance_migration"] = r.AllowInstanceMigration
    payload["disallow_reauthentication"] = r.DisallowReauthentication
    payload["token_ttl"] = r.TokenTTL
    payload["token_max_ttl"] = r.TokenMaxTTL
    payload["token_policies"] = toInterfaceArray(r.TokenPolicies)
    payload["policies"] = toInterfaceArray(r.Policies)
    payload["token_bound_cidrs"] = toInterfaceArray(r.TokenBoundCIDRs)
    payload["token_explicit_max_ttl"] = r.TokenExplicitMaxTTL
    payload["token_no_default_policy"] = r.TokenNoDefaultPolicy
    payload["token_num_uses"] = r.TokenNumUses
    payload["token_period"] = r.TokenPeriod
    payload["token_type"] = r.TokenType
    return payload
}
```

**Note:** Unlike GCP role where TTL fields are duration strings, the AWS auth role uses `token_ttl`/`token_max_ttl`/`token_explicit_max_ttl`/`token_period` as strings that the user provides as duration format. Follow the GCPAuthEngineRole pattern: emit these as-is (Vault accepts both formats). Use `removeUnsetFields` to handle fields that Vault doesn't return when unset.

### CRD Field Spec — AWSAuthEngineClientConfig

```go
type AWSAuthClientConfig struct {
    // AccessKey is the AWS access key ID for API calls.
    // +kubebuilder:validation:Optional
    AccessKey string `json:"accessKey,omitempty"`

    // Endpoint is a custom URL for EC2 API calls.
    // +kubebuilder:validation:Optional
    Endpoint string `json:"endpoint,omitempty"`

    // IAMEndpoint is a custom URL for IAM API calls.
    // +kubebuilder:validation:Optional
    IAMEndpoint string `json:"iamEndpoint,omitempty"`

    // STSEndpoint is a custom URL for STS API calls.
    // +kubebuilder:validation:Optional
    STSEndpoint string `json:"stsEndpoint,omitempty"`

    // STSRegion is the region for STS API calls (set with stsEndpoint).
    // +kubebuilder:validation:Optional
    STSRegion string `json:"stsRegion,omitempty"`

    // UseSTSRegionFromClient overrides stsEndpoint/stsRegion to use client request region.
    // +kubebuilder:validation:Optional
    UseSTSRegionFromClient bool `json:"useSTSRegionFromClient,omitempty"`

    // IAMServerIDHeaderValue is the value to require in the X-Vault-AWS-IAM-Server-ID header.
    // +kubebuilder:validation:Optional
    IAMServerIDHeaderValue string `json:"iamServerIDHeaderValue,omitempty"`

    // AllowedSTSHeaderValues is a comma-separated list of additional permitted STS request headers.
    // +kubebuilder:validation:Optional
    AllowedSTSHeaderValues string `json:"allowedSTSHeaderValues,omitempty"`

    // MaxRetries is the max retries for recoverable errors (-1 = AWS SDK default).
    // +kubebuilder:validation:Optional
    MaxRetries *int `json:"maxRetries,omitempty"`

    retrievedAccessKey string `json:"-"`
    retrievedSecretKey string `json:"-"`
}
```

### CRD Field Spec — AWSAuthEngineIdentityConfig

```go
type AWSAuthIdentityConfig struct {
    // IAMAlias controls identity alias generation for IAM auth.
    // +kubebuilder:validation:Optional
    // +kubebuilder:default="role_id"
    // +kubebuilder:validation:Enum={"role_id","unique_id","canonical_arn","full_arn"}
    IAMAlias string `json:"iamAlias"`

    // IAMMetadata is the metadata to include on the login token for IAM auth.
    // +kubebuilder:validation:Optional
    // +kubebuilder:default="default"
    IAMMetadata string `json:"iamMetadata"`

    // EC2Alias controls identity alias generation for EC2 auth.
    // +kubebuilder:validation:Optional
    // +kubebuilder:default="role_id"
    // +kubebuilder:validation:Enum={"role_id","instance_id","image_id"}
    EC2Alias string `json:"ec2Alias"`

    // EC2Metadata is the metadata to include on the login token for EC2 auth.
    // +kubebuilder:validation:Optional
    // +kubebuilder:default="default"
    EC2Metadata string `json:"ec2Metadata"`
}
```

### CRD Field Spec — AWSAuthEngineRole

```go
type AWSAuthRole struct {
    // Name of the role.
    // +kubebuilder:validation:Required
    Name string `json:"name"`

    // AuthType specifies the auth type for this role (iam or ec2).
    // +kubebuilder:validation:Optional
    // +kubebuilder:default="iam"
    // +kubebuilder:validation:Enum={"iam","ec2"}
    AuthType string `json:"authType"`

    // BoundAmiID constrains EC2 instances to specific AMI IDs.
    // +kubebuilder:validation:Optional
    // +listType=set
    BoundAmiID []string `json:"boundAmiID,omitempty"`

    // BoundAccountID constrains EC2 instances to specific account IDs.
    // +kubebuilder:validation:Optional
    // +listType=set
    BoundAccountID []string `json:"boundAccountID,omitempty"`

    // BoundRegion constrains EC2 instances to specific regions.
    // +kubebuilder:validation:Optional
    // +listType=set
    BoundRegion []string `json:"boundRegion,omitempty"`

    // BoundVpcID constrains EC2 instances to specific VPC IDs.
    // +kubebuilder:validation:Optional
    // +listType=set
    BoundVpcID []string `json:"boundVpcID,omitempty"`

    // BoundSubnetID constrains EC2 instances to specific subnet IDs.
    // +kubebuilder:validation:Optional
    // +listType=set
    BoundSubnetID []string `json:"boundSubnetID,omitempty"`

    // BoundIAMRoleARN constrains EC2 instances to specific IAM role ARNs.
    // +kubebuilder:validation:Optional
    // +listType=set
    BoundIAMRoleARN []string `json:"boundIAMRoleARN,omitempty"`

    // BoundIAMInstanceProfileARN constrains EC2 instances to specific instance profile ARNs.
    // +kubebuilder:validation:Optional
    // +listType=set
    BoundIAMInstanceProfileARN []string `json:"boundIAMInstanceProfileARN,omitempty"`

    // BoundEC2InstanceID constrains to specific EC2 instance IDs.
    // +kubebuilder:validation:Optional
    // +listType=set
    BoundEC2InstanceID []string `json:"boundEC2InstanceID,omitempty"`

    // RoleTag enables role tags for EC2 auth. Value is the tag key on the instance.
    // +kubebuilder:validation:Optional
    RoleTag string `json:"roleTag,omitempty"`

    // BoundIAMPrincipalARN constrains IAM auth to specific principal ARNs. Wildcards supported.
    // +kubebuilder:validation:Optional
    // +listType=set
    BoundIAMPrincipalARN []string `json:"boundIAMPrincipalARN,omitempty"`

    // InferredEntityType enables IAM role inferencing. Only valid value: "ec2_instance".
    // +kubebuilder:validation:Optional
    InferredEntityType string `json:"inferredEntityType,omitempty"`

    // InferredAWSRegion is the region to search for inferred entities. Required with inferredEntityType.
    // +kubebuilder:validation:Optional
    InferredAWSRegion string `json:"inferredAWSRegion,omitempty"`

    // ResolveAWSUniqueIDs resolves bound_iam_principal_arn to AWS unique IDs.
    // +kubebuilder:validation:Optional
    // +kubebuilder:default=true
    ResolveAWSUniqueIDs bool `json:"resolveAWSUniqueIDs"`

    // AllowInstanceMigration allows migration of the underlying EC2 instance. EC2 auth only.
    // +kubebuilder:validation:Optional
    AllowInstanceMigration bool `json:"allowInstanceMigration,omitempty"`

    // DisallowReauthentication if true, only allows a single token per instance ID. EC2 auth only.
    // +kubebuilder:validation:Optional
    DisallowReauthentication bool `json:"disallowReauthentication,omitempty"`

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

    // TokenBoundCIDRs are CIDR blocks that restrict authentication and tie the token.
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
    // +kubebuilder:validation:Minimum=0
    TokenPeriod int64 `json:"tokenPeriod,omitempty"`

    // TokenType is the type of token to generate (service, batch, default).
    // +kubebuilder:validation:Optional
    // +kubebuilder:validation:Enum={"service","batch","default","default-service","default-batch"}
    TokenType string `json:"tokenType,omitempty"`
}
```

### Kubebuilder Markers Checklist

- All exported Spec fields: `+kubebuilder:validation:Required` or `+kubebuilder:validation:Optional`
- Identity config defaults: `+kubebuilder:default="role_id"` for alias fields (non-zero default → no `omitempty`), `+kubebuilder:default="default"` for metadata fields
- Role `AuthType`: `+kubebuilder:default="iam"` (non-zero default → no `omitempty`), `+kubebuilder:validation:Enum={"iam","ec2"}`
- Role `ResolveAWSUniqueIDs`: `+kubebuilder:default=true` (non-zero default → no `omitempty`)
- Root types: `+kubebuilder:object:root=true`, `+kubebuilder:subresource:status`
- Conditions field: `+patchMergeKey=type`, `+patchStrategy=merge`, `+listType=map`, `+listMapKey=type`
- List fields on role: `+listType=set`

### RBAC Markers

Client config controller:
```go
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=awsauthengineclientconfigs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=awsauthengineclientconfigs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=awsauthengineclientconfigs/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=serviceaccounts/token,verbs=create
//+kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;patch
```

Identity config controller:
```go
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=awsauthengineidentityconfigs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=awsauthengineidentityconfigs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=awsauthengineidentityconfigs/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=serviceaccounts/token,verbs=create
//+kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;patch
```

Role controller:
```go
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=awsauthengineroles,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=awsauthengineroles/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=redhatcop.redhat.io,resources=awsauthengineroles/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=serviceaccounts/token,verbs=create
//+kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch;create;patch
```

### File Structure

| File | Action | Description |
|------|--------|-------------|
| `api/v1alpha1/awsauthengineconfig_types.go` | NEW | ClientConfig CRD type, VaultObject, ConditionsAware, toMap, credential resolution |
| `api/v1alpha1/awsauthengineidentityconfig_types.go` | NEW | IdentityConfig CRD type, VaultObject, ConditionsAware, toMap |
| `api/v1alpha1/awsauthenginerole_types.go` | NEW | Role CRD type, VaultObject, ConditionsAware, toMap |
| `api/v1alpha1/awsauthengineconfig_webhook.go` | NEW | ClientConfig webhook — defaulter, validator, immutable path, credential validation |
| `api/v1alpha1/awsauthengineidentityconfig_webhook.go` | NEW | IdentityConfig webhook — defaulter, validator, immutable path |
| `api/v1alpha1/awsauthenginerole_webhook.go` | NEW | Role webhook — defaulter, validator, immutable path/name, auth_type constraint validation |
| `api/v1alpha1/awsauthengineconfig_test.go` | NEW | Unit tests for client config toMap, IsEquivalentToDesiredState |
| `api/v1alpha1/awsauthengineidentityconfig_test.go` | NEW | Unit tests for identity config toMap, IsEquivalentToDesiredState |
| `api/v1alpha1/awsauthenginerole_test.go` | NEW | Unit tests for role toMap, IsEquivalentToDesiredState, webhook validation |
| `internal/controller/awsauthengineconfig_controller.go` | NEW | ClientConfig reconciler with Secret/RandomSecret watches |
| `internal/controller/awsauthengineidentityconfig_controller.go` | NEW | IdentityConfig reconciler — simple VaultResource |
| `internal/controller/awsauthenginerole_controller.go` | NEW | Role reconciler — simple VaultResource |
| `cmd/main.go` | UPDATE | Register 3 controllers + 3 webhooks |
| `config/crd/kustomization.yaml` | UPDATE | Add 3 new CRD YAML files to resources list |
| `test/awsauthengine/` | NEW | Test YAML fixtures for all 3 types |
| `docs/auth-engines/aws.md` | NEW | Engine documentation per DNFR5 |
| `docs/auth-engines/index.md` | UPDATE | Add link to aws.md |

### Files Being Modified — Current State

**`cmd/main.go`**: Currently registers ~42+ controllers and ~42+ webhooks (including Epic 13 additions). New registrations follow the exact same pattern:
- Controller: `(&controller.AWSAuthEngineClientConfigReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "AWSAuthEngineClientConfig")}).SetupWithManager(mgr)`
- Webhook (inside ENABLE_WEBHOOKS): `(&redhatcopv1alpha1.AWSAuthEngineClientConfig{}).SetupWebhookWithManager(mgr)`

No existing behavior is changed — this is purely additive.

**`config/crd/kustomization.yaml`**: Add the 3 new CRD YAML files to the `resources` list. Required for Helm chart build.

**`docs/auth-engines/index.md`**: Add a link entry for the new AWS auth doc.

### Existing Patterns to Follow (Reference Files)

| Pattern | Reference File |
|---------|---------------|
| AWS credential resolution (access_key/secret_key) | `api/v1alpha1/awssecretengineconfig_types.go` |
| Auth engine config with credential resolution | `api/v1alpha1/gcpauthengineconfig_types.go` |
| Auth engine config IsDeletable=false | `api/v1alpha1/gcpauthengineconfig_types.go` |
| Auth engine config GetPath (auth/{path}/config) | `api/v1alpha1/gcpauthengineconfig_types.go` |
| Auth engine role with Name override | `api/v1alpha1/gcpauthenginerole_types.go` |
| Auth engine role GetPath (auth/{path}/role/{name}) | `api/v1alpha1/gcpauthenginerole_types.go` |
| Secret-key stripping in IsEquivalentToDesiredState | `api/v1alpha1/awssecretengineconfig_types.go` (deletes `secret_key`) |
| Config controller with Secret+RandomSecret watches | `internal/controller/gcpauthengineconfig_controller.go` |
| Simple auth role controller (no watches) | `internal/controller/gcpauthenginerole_controller.go` |
| Auth config webhook (immutable path only) | `api/v1alpha1/gcpauthengineconfig_webhook.go` |
| Auth role webhook (immutable path+name) | `api/v1alpha1/gcpauthenginerole_webhook.go` |
| Identity config with enum defaults (alias fields) | `api/v1alpha1/gcpauthengineconfig_types.go` (GCPConfig alias fields) |
| removeUnsetFields + filterPayloadToDesiredKeys | `api/v1alpha1/payload_filter.go` |
| toInterfaceArray helper | `api/v1alpha1/utils/vaultutils.go` |
| Conditional field validation (auth_type-specific) | `api/v1alpha1/awssecretenginerole_types.go` (credentialType-specific) |
| Documentation template | `docs/engine-doc-template.md` |

### Unit Test Requirements

**Client config tests (`awsauthengineconfig_test.go`):**
1. `TestAWSAuthEngineClientConfig_toMap` — verify all fields in snake_case, verify credentials from resolved fields
2. `TestAWSAuthEngineClientConfig_IsEquivalentToDesiredState_Match` — construct Vault-read-shaped payload (without `secret_key`, with `access_key`, `endpoint`, etc.), verify returns `true`
3. `TestAWSAuthEngineClientConfig_IsEquivalentToDesiredState_Mismatch` — change `sts_endpoint`, verify returns `false`
4. `TestAWSAuthEngineClientConfig_IsEquivalentToDesiredState_ExtraVaultFields` — add extra Vault-returned field, verify still returns `true` after filtering

**Identity config tests (`awsauthengineidentityconfig_test.go`):**
1. `TestAWSAuthEngineIdentityConfig_toMap` — verify `iam_alias`, `iam_metadata`, `ec2_alias`, `ec2_metadata`
2. `TestAWSAuthEngineIdentityConfig_IsEquivalentToDesiredState_Match` — Vault-read fixture, verify returns `true`
3. `TestAWSAuthEngineIdentityConfig_IsEquivalentToDesiredState_Mismatch` — change `iam_alias`, verify returns `false`

**Role tests (`awsauthenginerole_test.go`):**
1. `TestAWSAuthEngineRole_toMap_IAM` — verify all fields for IAM role, verify list fields are `[]any`
2. `TestAWSAuthEngineRole_toMap_EC2` — verify all fields for EC2 role
3. `TestAWSAuthEngineRole_IsEquivalentToDesiredState_Match` — Vault-read fixture with mixed types, verify returns `true`
4. `TestAWSAuthEngineRole_IsEquivalentToDesiredState_Mismatch` — change `auth_type`, verify returns `false`
5. `TestAWSAuthEngineRole_IsEquivalentToDesiredState_ExtraVaultFields` — extra fields from Vault, verify filtering
6. `TestAWSAuthEngineRole_Webhook_IAMOnlyFields_EC2Rejected` — verify EC2 role rejects IAM-only fields
7. `TestAWSAuthEngineRole_Webhook_EC2OnlyFields_IAMRejected` — verify IAM role rejects EC2-only fields

### Anti-Patterns / DO NOT

- **DO NOT** create integration tests for these types — AWS is a cloud provider that cannot be installed in Kind (per "Skip it" rule)
- **DO NOT** merge the two config endpoints into a single CRD — the design decision is two separate CRDs (AWSAuthEngineClientConfig + AWSAuthEngineIdentityConfig) for clean 1:1 Vault API mapping
- **DO NOT** use the always-write controller pattern for client config — unlike AWSSecretEngineConfig, the client config read response returns enough fields for meaningful drift detection; only `secret_key` is write-only
- **DO NOT** modify shared framework behavior (reconcile_skeleton.go, vaultresourcereconciler.go, etc.) — only add new files
- **DO NOT** derive expected payloads from `toMap()` in unit tests — construct independent Vault-read-shaped fixtures
- **DO NOT** forget to add new CRD YAML files to `config/crd/kustomization.yaml` after `make manifests`
- **DO NOT** use Go `int` or `float64` in unit test Vault payloads — use `json.Number` to match real Vault client behavior
- **DO NOT** include Enterprise-only fields (role_arn, identity_token_audience, identity_token_ttl, rotation_*) in the CRD — these are behind Vault Enterprise license and not applicable to the open-source operator
- **DO NOT** add the `AWSAuthEngineRole.Name` field to `toMap()` output — Vault's role API uses the URL path for the role name, not a body field. The `Name` field is only for `GetPath()` override.
- **DO NOT** confuse `AWSAuthEngineClientConfig` (auth engine, this story) with `AWSSecretEngineConfig` (secret engine, Epic 11) — different Vault API paths, different purposes

### Novelty Risk: LOW

All three CRD types follow well-established patterns from existing auth engine implementations (GCP, Azure). The AWS-specific complexity (two config endpoints, auth_type-specific constraints) is fully addressed by the design decision to split into two config CRDs and by the validation matrix. No novel architectural patterns required.

### Project Structure Notes

- All new files follow existing naming conventions: lowercase type name for files
- New files go in established directories (`api/v1alpha1/`, `internal/controller/`, `test/`, `docs/auth-engines/`)
- Test fixture directory `test/awsauthengine/` follows the existing pattern (`test/gcpauthengine/`, `test/azureauthengine/`)
- No conflicts with existing code — purely additive
- Note: `api/v1alpha1/awsauthengineconfig_types.go` is for the **auth** engine client config; `api/v1alpha1/awssecretengineconfig_types.go` already exists for the **secret** engine config — these are different files for different purposes

### References

- [Source: _bmad-output/planning-artifacts/epics.md#Epic-14, Story 14.2 — requirements and acceptance criteria]
- [Source: _bmad-output/project-context.md — all project rules, patterns, and conventions]
- [Source: api/v1alpha1/gcpauthengineconfig_types.go — GCP auth config with credential resolution, IsDeletable=false]
- [Source: api/v1alpha1/gcpauthenginerole_types.go — GCP auth role with Name override, auth_type enum]
- [Source: api/v1alpha1/awssecretengineconfig_types.go — AWS credential resolution, secret_key stripping]
- [Source: api/v1alpha1/awssecretenginerole_types.go — conditional field validation per credential_type]
- [Source: api/v1alpha1/azureauthengineconfig_types.go — Azure auth config with RootCredentialConfig]
- [Source: internal/controller/gcpauthengineconfig_controller.go — auth config controller with Secret/RandomSecret watches]
- [Source: internal/controller/gcpauthenginerole_controller.go — simple auth role controller]
- [Source: api/v1alpha1/gcpauthengineconfig_webhook.go — auth config webhook pattern]
- [Source: docs/engine-doc-template.md — documentation template]
- [Source: Vault AWS Auth Method API — https://developer.hashicorp.com/vault/api-docs/auth/aws]
- [Source: _bmad-output/implementation-artifacts/13-4-terraform-cloud-secret-engine-config-and-role-crds.md — most recent predecessor story]
- [Source: Sprint-status action item Epic 13: "AWS Auth config endpoint decision — Two separate CRDs" (status: done)]

## Code Review Record

### Review Model Used

(to be filled during code review)

### Review Findings

(to be filled during code review)

### Decisions Needed / Decisions Taken

- Design decision (pre-resolved): Two separate config CRDs (AWSAuthEngineClientConfig + AWSAuthEngineIdentityConfig) for 1:1 Vault API mapping
- Design decision (pre-resolved): Standard VaultResource reconcile for client config (NOT always-write), since read response provides meaningful drift detection

### Fixes Applied

(to be filled during code review)

## Dev Agent Record

### Agent Model Used

{{agent_model_name_version}}

### Debug Log References

### Completion Notes List

### Change Log

### File List
