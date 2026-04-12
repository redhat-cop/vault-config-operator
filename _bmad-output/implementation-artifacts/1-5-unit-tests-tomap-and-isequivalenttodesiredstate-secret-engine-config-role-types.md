# Story 1.5: Unit tests for `toMap()` and `IsEquivalentToDesiredState` — Secret Engine Config & Role Types

Status: ready-for-dev

## Story

As an operator developer,
I want unit tests for all secret engine configuration and role types,
So that field mappings for RabbitMQ, GitHub, Azure, PKI, Quay, and Kubernetes secret engines are verified.

## Acceptance Criteria

1. **Given** each secret engine type instance with representative field values **When** `toMap()` (or the type's equivalent map method) is called **Then** all fields map correctly to Vault API snake_case keys

2. **Given** each type and a matching Vault read response **When** `IsEquivalentToDesiredState(payload)` is called **Then** it returns `true`

3. **Given** each type and a response with one managed field changed **When** `IsEquivalentToDesiredState(payload)` is called **Then** it returns `false`

4. **Given** each type and a Vault response with extra fields not managed by the operator **When** `IsEquivalentToDesiredState(payload)` is called **Then** behavior is documented — `false` for types using bare `reflect.DeepEqual`, consistent with stories 1.1–1.4

5. **Given** types with custom `IsEquivalentToDesiredState` logic (GitHubSecretEngineConfig deletes `prv_key`, QuaySecretEngineConfig deletes `password`, KubernetesSecretEngineConfig deletes `service_account_jwt`, RabbitMQSecretEngineConfig uses `leasesToMap` not `rabbitMQToMap`) **When** `IsEquivalentToDesiredState` is called **Then** the custom behavior is verified explicitly

## Types Covered

| # | Type | File | Config Struct | Map Method | `IsEquivalentToDesiredState` | `IsDeletable` | Keys | Existing Tests |
|---|------|------|---------------|------------|------------------------------|---------------|------|----------------|
| 1 | RabbitMQSecretEngineConfig | `api/v1alpha1/rabbitmqsecretengineconfig_types.go` | `RMQSEConfig` | `rabbitMQToMap` (6 keys), `leasesToMap` (2 keys) | **Custom: uses `leasesToMap()` not `rabbitMQToMap()`** | `false` | 6+2 | None |
| 2 | RabbitMQSecretEngineRole | `api/v1alpha1/rabbitmqsecretenginerole_types.go` | `RMQSERole` | `rabbitMQToMap` | bare DeepEqual | `true` | 3 | None |
| 3 | GitHubSecretEngineConfig | `api/v1alpha1/githubsecretengineconfig_types.go` | `GHConfig` | `toMap` | **Custom: deletes `prv_key`** | `false` | 3 | None |
| 4 | GitHubSecretEngineRole | `api/v1alpha1/githubsecretenginerole_types.go` | `PermissionSet` | `toMap` | bare DeepEqual | `true` | 5 | None |
| 5 | AzureSecretEngineConfig | `api/v1alpha1/azuresecretengineconfig_types.go` | `AzureSEConfig` | `toMap` | bare DeepEqual | `true` | 7 | None |
| 6 | AzureSecretEngineRole | `api/v1alpha1/azuresecretenginerole_types.go` | `AzureSERole` | `toMap` | bare DeepEqual | `true` | 9 | None |
| 7 | PKISecretEngineConfig | `api/v1alpha1/pkisecretengineconfig_types.go` | `PKICommon` | `toMap` (22 keys), plus `PKIConfigUrls.toMap` (3), `PKIConfigCRL.toMap` (2) | bare DeepEqual (against `PKICommon` only) | `false` | 22+3+2 | None |
| 8 | PKISecretEngineRole | `api/v1alpha1/pkisecretenginerole_types.go` | `PKIRole` | `toMap` | bare DeepEqual | `true` | 38 | None |
| 9 | QuaySecretEngineConfig | `api/v1alpha1/quaysecretengineconfig_types.go` | `QuayConfig` | `toMap` | **Custom: deletes `password`** | `false` | 4 | None |
| 10 | QuaySecretEngineRole | `api/v1alpha1/quaysecretenginerole_types.go` | `QuayRole` | `toMap` | bare DeepEqual | `true` | 5–8 (conditional) | None |
| 11 | QuaySecretEngineStaticRole | `api/v1alpha1/quaysecretenginestaticrole_types.go` | `QuayBaseRole` | `toMap` | bare DeepEqual | `true` | 3–6 (conditional) | None |
| 12 | KubernetesSecretEngineConfig | `api/v1alpha1/kubernetessecretengineconfig_types.go` | `KubeSEConfig` | `toMap` | **Custom: deletes `service_account_jwt`** | `true` | 4 | None |
| 13 | KubernetesSecretEngineRole | `api/v1alpha1/kubernetessecretenginerole_types.go` | `KubeSERole` | `toMap` | bare DeepEqual | `true` | 12 | None |

## Tasks / Subtasks

- [ ] Task 1: Add RabbitMQSecretEngineConfig unit tests (AC: 1, 2, 3, 4, 5)
  - [ ] 1.1: Create `api/v1alpha1/rabbitmqsecretengineconfig_test.go`
  - [ ] 1.2: Add `TestRabbitMQSecretEngineConfigGetPath` — verify returns `{spec.path}/config/connection`
  - [ ] 1.3: Add `TestRMQSEConfigRabbitMQToMap` — verify all 6 keys: `connection_uri`, `verify_connection`, `username`, `password`, `username_template`, `password_policy`; set `retrievedUsername`/`retrievedPassword` directly (same package)
  - [ ] 1.4: Add `TestRMQSEConfigLeasesToMap` — verify 2 keys: `ttl`, `max_ttl`
  - [ ] 1.5: Add `TestRabbitMQSecretEngineConfigIsEquivalentMatching` — **critical**: uses `leasesToMap()` NOT `rabbitMQToMap()`, so payload must match lease keys only; matching → `true`
  - [ ] 1.6: Add `TestRabbitMQSecretEngineConfigIsEquivalentNonMatching` — change one lease field → `false`
  - [ ] 1.7: Add `TestRabbitMQSecretEngineConfigIsEquivalentExtraFields` — extra keys in payload → `false` (bare DeepEqual on lease map)
  - [ ] 1.8: Add `TestRabbitMQSecretEngineConfigIsDeletable` — returns `false`
  - [ ] 1.9: Add `TestRabbitMQSecretEngineConfigConditions` — GetConditions/SetConditions round-trip
  - [ ] 1.10: Add `TestRabbitMQSecretEngineConfigGetLeasePayload` — verify `GetLeasePayload()` matches `leasesToMap()`
  - [ ] 1.11: Add `TestRabbitMQSecretEngineConfigGetLeasePath` — verify returns `{spec.path}/config/lease`
- [ ] Task 2: Add RabbitMQSecretEngineRole unit tests (AC: 1, 2, 3, 4)
  - [ ] 2.1: Create `api/v1alpha1/rabbitmqsecretenginerole_test.go`
  - [ ] 2.2: Add `TestRabbitMQSecretEngineRoleGetPath` — with `spec.name`, without (fallback to `metadata.name`); verify path format `CleansePath({path}/roles/{name})`
  - [ ] 2.3: Add `TestRMQSERoleRabbitMQToMap` — verify all 3 keys: `tags`, `vhosts`, `vhost_topics`; verify `vhosts` and `vhost_topics` are JSON strings (from `convertVhostsToJson`/`convertTopicsToJson`)
  - [ ] 2.4: Add `TestRabbitMQSecretEngineRoleIsEquivalentMatching` — matching payload → `true`
  - [ ] 2.5: Add `TestRabbitMQSecretEngineRoleIsEquivalentNonMatching` — one field changed → `false`
  - [ ] 2.6: Add `TestRabbitMQSecretEngineRoleIsEquivalentExtraFields` — extra keys → `false`
  - [ ] 2.7: Add `TestRabbitMQSecretEngineRoleIsDeletable` — returns `true`
  - [ ] 2.8: Add `TestRabbitMQSecretEngineRoleConditions` — GetConditions/SetConditions round-trip
- [ ] Task 3: Add GitHubSecretEngineConfig unit tests (AC: 1, 2, 3, 4, 5)
  - [ ] 3.1: Create `api/v1alpha1/githubsecretengineconfig_test.go`
  - [ ] 3.2: Add `TestGitHubSecretEngineConfigGetPath` — verify returns `{spec.path}/config` (no name segment, no CleansePath)
  - [ ] 3.3: Add `TestGHConfigToMap` — verify all 3 keys: `app_id`, `prv_key`, `base_url`; set `retrievedSSHKey` directly (same package)
  - [ ] 3.4: Add `TestGitHubSecretEngineConfigIsEquivalentPrvKeyDeleted` — **critical test**: verify `prv_key` is deleted from desiredState before comparison; a payload without `prv_key` that matches `app_id` and `base_url` → `true`
  - [ ] 3.5: Add `TestGitHubSecretEngineConfigIsEquivalentMatching` — matching payload (without `prv_key`) → `true`
  - [ ] 3.6: Add `TestGitHubSecretEngineConfigIsEquivalentNonMatching` — one managed field changed → `false`
  - [ ] 3.7: Add `TestGitHubSecretEngineConfigIsEquivalentExtraFields` — extra keys → `false`
  - [ ] 3.8: Add `TestGitHubSecretEngineConfigIsDeletable` — returns `false`
  - [ ] 3.9: Add `TestGitHubSecretEngineConfigConditions` — GetConditions/SetConditions round-trip
- [ ] Task 4: Add GitHubSecretEngineRole unit tests (AC: 1, 2, 3, 4)
  - [ ] 4.1: Create `api/v1alpha1/githubsecretenginerole_test.go`
  - [ ] 4.2: Add `TestGitHubSecretEngineRoleGetPath` — with `spec.name`, without; verify path format `CleansePath({path}/permissionset/{name})` — **note: uses `permissionset` not `roles`**
  - [ ] 4.3: Add `TestPermissionSetToMap` — verify all 5 keys: `installation_id`, `org_name`, `repositories`, `repository_ids`, `permissions`; verify `permissions` is `map[string]string`
  - [ ] 4.4: Add `TestGitHubSecretEngineRoleIsEquivalentMatching` — matching payload → `true`
  - [ ] 4.5: Add `TestGitHubSecretEngineRoleIsEquivalentNonMatching` — one field changed → `false`
  - [ ] 4.6: Add `TestGitHubSecretEngineRoleIsEquivalentExtraFields` — extra keys → `false`
  - [ ] 4.7: Add `TestGitHubSecretEngineRoleIsDeletable` — returns `true`
  - [ ] 4.8: Add `TestGitHubSecretEngineRoleConditions` — GetConditions/SetConditions round-trip
- [ ] Task 5: Add AzureSecretEngineConfig unit tests (AC: 1, 2, 3, 4)
  - [ ] 5.1: Create `api/v1alpha1/azuresecretengineconfig_test.go`
  - [ ] 5.2: Add `TestAzureSecretEngineConfigGetPath` — verify returns `{spec.path}/config` (no name segment, no CleansePath)
  - [ ] 5.3: Add `TestAzureSEConfigToMap` — verify all 7 keys: `subscription_id`, `tenant_id`, `client_id`, `client_secret`, `environment`, `password_policy`, `root_password_ttl`; set `retrievedClientID`/`retrievedClientPassword` directly
  - [ ] 5.4: Add `TestAzureSecretEngineConfigIsEquivalentMatching` — matching payload → `true`
  - [ ] 5.5: Add `TestAzureSecretEngineConfigIsEquivalentNonMatching` — one field changed → `false`
  - [ ] 5.6: Add `TestAzureSecretEngineConfigIsEquivalentExtraFields` — extra keys → `false`
  - [ ] 5.7: Add `TestAzureSecretEngineConfigIsDeletable` — returns `true`
  - [ ] 5.8: Add `TestAzureSecretEngineConfigConditions` — GetConditions/SetConditions round-trip
- [ ] Task 6: Add AzureSecretEngineRole unit tests (AC: 1, 2, 3, 4)
  - [ ] 6.1: Create `api/v1alpha1/azuresecretenginerole_test.go`
  - [ ] 6.2: Add `TestAzureSecretEngineRoleGetPath` — with `spec.name`, without; verify path format `CleansePath({path}/roles/{name})`
  - [ ] 6.3: Add `TestAzureSERoleToMap` — verify all 9 keys: `azure_roles`, `azure_groups`, `application_object_id`, `persist_app`, `ttl`, `max_ttl`, `permanently_delete`, `sign_in_audience`, `tags`
  - [ ] 6.4: Add `TestAzureSecretEngineRoleIsEquivalentMatching` — matching payload → `true`
  - [ ] 6.5: Add `TestAzureSecretEngineRoleIsEquivalentNonMatching` — one field changed → `false`
  - [ ] 6.6: Add `TestAzureSecretEngineRoleIsEquivalentExtraFields` — extra keys → `false`
  - [ ] 6.7: Add `TestAzureSecretEngineRoleIsDeletable` — returns `true`
  - [ ] 6.8: Add `TestAzureSecretEngineRoleConditions` — GetConditions/SetConditions round-trip
- [ ] Task 7: Add PKISecretEngineConfig unit tests (AC: 1, 2, 3, 4)
  - [ ] 7.1: Create `api/v1alpha1/pkisecretengineconfig_test.go`
  - [ ] 7.2: Add `TestPKISecretEngineConfigGetPath` — verify returns `string(spec.path)` only (no `/config` in code — just the mount path)
  - [ ] 7.3: Add `TestPKICommonToMap` — verify all 22 keys: `common_name`, `alt_names`, `ip_sans`, `uri_sans`, `other_sans`, `ttl`, `format`, `private_key_format`, `key_type`, `key_bits`, `max_path_length`, `exclude_cn_from_sans`, `permitted_dns_domains`, `ou`, `organization`, `country`, `locality`, `province`, `street_address`, `postal_code`, `serial_number`; verify `type` is NOT in output (commented out)
  - [ ] 7.4: Add `TestPKIConfigUrlsToMap` — verify 3 keys: `issuing_certificates`, `crl_distribution_points`, `ocsp_servers`
  - [ ] 7.5: Add `TestPKIConfigCRLToMap` — verify 2 keys: `expiry`, `disable`
  - [ ] 7.6: Add `TestPKISecretEngineConfigIsEquivalentMatching` — uses `PKICommon.toMap()` only; matching → `true`
  - [ ] 7.7: Add `TestPKISecretEngineConfigIsEquivalentNonMatching` — one field changed → `false`
  - [ ] 7.8: Add `TestPKISecretEngineConfigIsEquivalentExtraFields` — extra keys → `false`
  - [ ] 7.9: Add `TestPKISecretEngineConfigIsDeletable` — returns `false`
  - [ ] 7.10: Add `TestPKISecretEngineConfigConditions` — GetConditions/SetConditions round-trip
- [ ] Task 8: Add PKISecretEngineRole unit tests (AC: 1, 2, 3, 4)
  - [ ] 8.1: Create `api/v1alpha1/pkisecretenginerole_test.go`
  - [ ] 8.2: Add `TestPKISecretEngineRoleGetPath` — with `spec.name`, without; verify path format `CleansePath({path}/roles/{name})`
  - [ ] 8.3: Add `TestPKIRoleToMap` — verify all 38 keys: `ttl`, `max_ttl`, `allow_localhost`, `allowed_domains`, `allowed_domains_template`, `allow_bare_domains`, `allow_subdomains`, `allow_glob_domains`, `allow_any_name`, `enforce_hostnames`, `allow_ip_sans`, `allowed_uri_sans`, `allowed_other_sans`, `server_flag`, `client_flag`, `code_signing_flag`, `email_protection_flag`, `key_type`, `key_bits`, `key_usage`, `ext_key_usage`, `ext_key_usage_oids`, `use_csr_common_name`, `use_csr_sans`, `ou`, `organization`, `country`, `locality`, `province`, `street_address`, `postal_code`, `serial_number`, `generate_lease`, `no_store`, `require_cn`, `policy_identifiers`, `basic_constraints_valid_for_non_ca`, `not_before_duration`
  - [ ] 8.4: Add `TestPKISecretEngineRoleIsEquivalentMatching` — matching payload → `true`
  - [ ] 8.5: Add `TestPKISecretEngineRoleIsEquivalentNonMatching` — one field changed → `false`
  - [ ] 8.6: Add `TestPKISecretEngineRoleIsEquivalentExtraFields` — extra keys → `false`
  - [ ] 8.7: Add `TestPKISecretEngineRoleIsDeletable` — returns `true`
  - [ ] 8.8: Add `TestPKISecretEngineRoleConditions` — GetConditions/SetConditions round-trip
- [ ] Task 9: Add QuaySecretEngineConfig unit tests (AC: 1, 2, 3, 4, 5)
  - [ ] 9.1: Create `api/v1alpha1/quaysecretengineconfig_test.go`
  - [ ] 9.2: Add `TestQuaySecretEngineConfigGetPath` — verify returns `{spec.path}/config`
  - [ ] 9.3: Add `TestQuayConfigToMap` — verify all 4 keys: `url`, `token`, `ca_certificate`, `disable_ssl_verification`; set `retrievedToken` directly
  - [ ] 9.4: Add `TestQuaySecretEngineConfigIsEquivalentPasswordDeleted` — **critical test**: verify `password` is deleted from desiredState before comparison; note that `toMap()` never sets a `password` key (it sets `token`), so this delete has no practical effect on the current `toMap()` output — document this behavior
  - [ ] 9.5: Add `TestQuaySecretEngineConfigIsEquivalentMatching` — matching payload → `true`
  - [ ] 9.6: Add `TestQuaySecretEngineConfigIsEquivalentNonMatching` — one field changed → `false`
  - [ ] 9.7: Add `TestQuaySecretEngineConfigIsEquivalentExtraFields` — extra keys → `false`
  - [ ] 9.8: Add `TestQuaySecretEngineConfigIsDeletable` — returns `false`
  - [ ] 9.9: Add `TestQuaySecretEngineConfigConditions` — GetConditions/SetConditions round-trip
- [ ] Task 10: Add QuaySecretEngineRole unit tests (AC: 1, 2, 3, 4)
  - [ ] 10.1: Create `api/v1alpha1/quaysecretenginerole_test.go`
  - [ ] 10.2: Add `TestQuaySecretEngineRoleGetPath` — with `spec.name`, without; verify path format `CleansePath({path}/roles/{name})`
  - [ ] 10.3: Add `TestQuayRoleToMapAllKeys` — set all optional fields (DefaultPermission, Teams, Repositories) to non-nil; verify 8 keys: `namespace_type`, `namespace_name`, `create_repositories`, `default_permission`, `teams`, `repositories`, `ttl`, `max_ttl`; verify `teams` and `repositories` are JSON strings (from `setMapJson`)
  - [ ] 10.4: Add `TestQuayRoleToMapMinimalKeys` — set optional fields to nil; verify only 5 keys present: `namespace_type`, `namespace_name`, `create_repositories`, `ttl`, `max_ttl`; verify `default_permission`, `teams`, `repositories` are absent
  - [ ] 10.5: Add `TestQuaySecretEngineRoleIsEquivalentMatching` — matching payload → `true`
  - [ ] 10.6: Add `TestQuaySecretEngineRoleIsEquivalentNonMatching` — one field changed → `false`
  - [ ] 10.7: Add `TestQuaySecretEngineRoleIsEquivalentExtraFields` — extra keys → `false`
  - [ ] 10.8: Add `TestQuaySecretEngineRoleIsDeletable` — returns `true`
  - [ ] 10.9: Add `TestQuaySecretEngineRoleConditions` — GetConditions/SetConditions round-trip
- [ ] Task 11: Add QuaySecretEngineStaticRole unit tests (AC: 1, 2, 3, 4)
  - [ ] 11.1: Create `api/v1alpha1/quaysecretenginestaticrole_test.go`
  - [ ] 11.2: Add `TestQuaySecretEngineStaticRoleGetPath` — with `spec.name`, without; verify path format `CleansePath({path}/static-roles/{name})` — **note: uses `static-roles` not `roles`**
  - [ ] 11.3: Add `TestQuayBaseRoleToMapAllKeys` — set all optional fields; verify 6 keys: `namespace_type`, `namespace_name`, `create_repositories`, `default_permission`, `teams`, `repositories`; verify NO `ttl`/`max_ttl` (unlike QuayRole)
  - [ ] 11.4: Add `TestQuayBaseRoleToMapMinimalKeys` — set optional fields to nil; verify only 3 keys: `namespace_type`, `namespace_name`, `create_repositories`
  - [ ] 11.5: Add `TestQuaySecretEngineStaticRoleIsEquivalentMatching` — matching payload → `true`
  - [ ] 11.6: Add `TestQuaySecretEngineStaticRoleIsEquivalentNonMatching` — one field changed → `false`
  - [ ] 11.7: Add `TestQuaySecretEngineStaticRoleIsEquivalentExtraFields` — extra keys → `false`
  - [ ] 11.8: Add `TestQuaySecretEngineStaticRoleIsDeletable` — returns `true`
  - [ ] 11.9: Add `TestQuaySecretEngineStaticRoleConditions` — GetConditions/SetConditions round-trip
- [ ] Task 12: Add KubernetesSecretEngineConfig unit tests (AC: 1, 2, 3, 4, 5)
  - [ ] 12.1: Create `api/v1alpha1/kubernetessecretengineconfig_test.go`
  - [ ] 12.2: Add `TestKubernetesSecretEngineConfigGetPath` — verify returns `{spec.path}/config`
  - [ ] 12.3: Add `TestKubeSEConfigToMap` — verify all 4 keys: `kubernetes_host`, `kubernetes_ca_cert`, `service_account_jwt`, `disable_local_ca_jwt`; set `retrievedServiceAccountJWT` directly
  - [ ] 12.4: Add `TestKubernetesSecretEngineConfigIsEquivalentJWTDeleted` — **critical test**: verify `service_account_jwt` is deleted from desiredState before comparison; a payload without `service_account_jwt` that matches remaining keys → `true`
  - [ ] 12.5: Add `TestKubernetesSecretEngineConfigIsEquivalentMatching` — matching payload (without JWT) → `true`
  - [ ] 12.6: Add `TestKubernetesSecretEngineConfigIsEquivalentNonMatching` — one managed field changed → `false`
  - [ ] 12.7: Add `TestKubernetesSecretEngineConfigIsEquivalentExtraFields` — extra keys → `false`
  - [ ] 12.8: Add `TestKubernetesSecretEngineConfigIsDeletable` — returns `true`
  - [ ] 12.9: Add `TestKubernetesSecretEngineConfigConditions` — GetConditions/SetConditions round-trip
- [ ] Task 13: Add KubernetesSecretEngineRole unit tests (AC: 1, 2, 3, 4)
  - [ ] 13.1: Create `api/v1alpha1/kubernetessecretenginerole_test.go`
  - [ ] 13.2: Add `TestKubernetesSecretEngineRoleGetPath` — with `spec.name`, without; verify path format `CleansePath({path}/roles/{name})`
  - [ ] 13.3: Add `TestKubeSERoleToMap` — verify all 12 keys: `allowed_kubernetes_namespaces`, `allowed_kubernetes_namespace_selector`, `token_max_ttl`, `token_default_ttl`, `token_default_audiences`, `service_account_name`, `kubernetes_role_name`, `kubernetes_role_type`, `generated_role_rules`, `name_template`, `extra_annotations`, `extra_labels`; **verify the field-to-key mapping: `DefaultTTL` → `token_max_ttl`, `MaxTTL` → `token_default_ttl`** (appears swapped — document and test the actual code behavior)
  - [ ] 13.4: Add `TestKubernetesSecretEngineRoleIsEquivalentMatching` — matching payload → `true`
  - [ ] 13.5: Add `TestKubernetesSecretEngineRoleIsEquivalentNonMatching` — one field changed → `false`
  - [ ] 13.6: Add `TestKubernetesSecretEngineRoleIsEquivalentExtraFields` — extra keys → `false`
  - [ ] 13.7: Add `TestKubernetesSecretEngineRoleIsDeletable` — returns `true`
  - [ ] 13.8: Add `TestKubernetesSecretEngineRoleConditions` — GetConditions/SetConditions round-trip
- [ ] Task 14: Verify all tests pass (AC: all)
  - [ ] 14.1: Run `go test ./api/v1alpha1/ -v -count=1` to confirm all new and existing tests pass
  - [ ] 14.2: Run `make test` to verify no regressions in full unit test suite

## Dev Notes

### Critical: RabbitMQSecretEngineConfig Has Unique `IsEquivalentToDesiredState` Using `leasesToMap()`

Unlike every other type in the operator, `RabbitMQSecretEngineConfig.IsEquivalentToDesiredState` does NOT use the same map as `GetPayload()`. It uses `leasesToMap()` (2 keys: `ttl`, `max_ttl`) instead of `rabbitMQToMap()` (6 keys for the connection config). This is because the controller uses a separate `CreateOrUpdateLease` flow for the lease endpoint — `IsEquivalentToDesiredState` is only called in that lease flow, not for the connection write which is unconditional.

```go
func (rabbitMQ *RabbitMQSecretEngineConfig) IsEquivalentToDesiredState(payload map[string]interface{}) bool {
    desiredState := rabbitMQ.Spec.RMQSEConfig.leasesToMap()
    return reflect.DeepEqual(desiredState, payload)
}
```

Test this explicitly:
- A payload matching `{ttl: X, max_ttl: Y}` → `true`
- A payload matching `{ttl: X, max_ttl: Y, extra: Z}` → `false`
- A payload from `rabbitMQToMap()` → `false` (different key sets)

[Source: api/v1alpha1/rabbitmqsecretengineconfig_types.go#L179-L182]

### Critical: RabbitMQSecretEngineConfig Has Two Map Methods and Two Paths

The config type has two distinct payloads and paths:
- **Connection**: `rabbitMQToMap()` → `GetPayload()`, path: `{path}/config/connection` via `GetPath()`
- **Lease**: `leasesToMap()` → `GetLeasePayload()`, path: `{path}/config/lease` via `GetLeasePath()`

Test both map methods independently plus both path methods.

[Source: api/v1alpha1/rabbitmqsecretengineconfig_types.go#L129-L140, L322-L335]

### Critical: RabbitMQ Map Method Is Named `rabbitMQToMap`, Not `toMap`

Both `RMQSEConfig` and `RMQSERole` use the method name `rabbitMQToMap()` instead of the standard `toMap()`. This is a naming convention difference — the tests should call it by its actual name.

### Critical: RabbitMQSecretEngineRole Uses JSON String Encoding for `vhosts` and `vhost_topics`

`RMQSERole.rabbitMQToMap()` produces `vhosts` and `vhost_topics` as **JSON strings** via `convertVhostsToJson()` and `convertTopicsToJson()`. These are not raw Go structs — they are serialized strings. For `reflect.DeepEqual` in tests, the payload must use `string` type values matching the JSON output.

```go
payload["vhosts"] = convertVhostsToJson(i.Vhosts)
payload["vhost_topics"] = convertTopicsToJson(i.VhostTopics)
```

[Source: api/v1alpha1/rabbitmqsecretenginerole_types.go#L233-L239]

### Critical: GitHubSecretEngineConfig Deletes `prv_key` Before Comparison

`IsEquivalentToDesiredState` deletes `prv_key` from the desired state (the SSH private key), because Vault redacts it in read responses. After deletion, only `app_id` and `base_url` are compared.

```go
func (d *GitHubSecretEngineConfig) IsEquivalentToDesiredState(payload map[string]interface{}) bool {
    desiredState := d.Spec.GHConfig.toMap()
    delete(desiredState, "prv_key")
    return reflect.DeepEqual(desiredState, payload)
}
```

Test:
- Payload matching `{app_id: X, base_url: Y}` → `true` (prv_key stripped from desired)
- Payload with `prv_key` present that otherwise matches → `false` (payload has key that desired no longer has)

[Source: api/v1alpha1/githubsecretengineconfig_types.go#L108-L112]

### GitHubSecretEngineConfig Uses `retrievedSSHKey` Unexported Field

The `prv_key` value in `GHConfig.toMap()` comes from `retrievedSSHKey` (set via `setInternalCredentials` from K8s Secret). In unit tests, set this directly since tests are in `package v1alpha1`.

[Source: api/v1alpha1/githubsecretengineconfig_types.go#L72, L78]

### GitHubSecretEngineRole Path Uses `permissionset` Not `roles`

`GitHubSecretEngineRole.GetPath()` uses `{path}/permissionset/{name}`, not the typical `{path}/roles/{name}`. This matches the GitHub app secrets engine Vault API.

[Source: api/v1alpha1/githubsecretenginerole_types.go#L99-L104]

### GitHubSecretEngineRole `Permissions` Is `map[string]string`

The `permissions` key in `PermissionSet.toMap()` stores `map[string]string` directly. For `reflect.DeepEqual`, if Vault returns `map[string]interface{}` instead, it would return `false`. Document this behavior.

[Source: api/v1alpha1/githubsecretenginerole_types.go#L91]

### AzureSecretEngineConfig Uses Unexported Credential Fields

`client_id` comes from `retrievedClientID` and `client_secret` from `retrievedClientPassword`. The exported `ClientID` field is NOT used in `toMap()`. Set the unexported fields directly in tests.

[Source: api/v1alpha1/azuresecretengineconfig_types.go#L259-L270]

### AzureSecretEngineConfig Has Conditional PrepareInternalValues

If `AzureCredentials` matches the default `RootCredentialConfig{PasswordKey: "clientsecret", UsernameKey: "clientid"}`, `PrepareInternalValues` returns early without calling `setInternalCredentials`, leaving `retrieved*` fields empty. Tests should set the retrieved fields directly.

[Source: api/v1alpha1/azuresecretengineconfig_types.go#L179-L181]

### PKISecretEngineConfig Has Three `toMap()` Methods

`PKISecretEngineConfig` has three different map methods on different embedded structs:
- `PKICommon.toMap()` — 22 keys (used by `GetPayload()` and `IsEquivalentToDesiredState`)
- `PKIConfigUrls.toMap()` — 3 keys (used by `GetConfigUrlsPayload()`)
- `PKIConfigCRL.toMap()` — 2 keys (used by `GetConfigCrlPayload()`)

`IsEquivalentToDesiredState` only compares against `PKICommon.toMap()`. Test all three map methods but note that only `PKICommon` is used for equivalence checking.

[Source: api/v1alpha1/pkisecretengineconfig_types.go#L523-L568]

### PKISecretEngineConfig `type` Is Commented Out in `toMap()`

`PKICommon.toMap()` has a commented-out line: `// payload["type"] = i.Type`. Verify that `type` is NOT present in the output.

[Source: api/v1alpha1/pkisecretengineconfig_types.go#L526-L527]

### PKISecretEngineConfig `GetPath()` Returns Only the Mount Path

Unlike most config types, `PKISecretEngineConfig.GetPath()` returns just `string(spec.path)` without appending `/config`. The controller uses additional path methods (`GetGeneratePath`, `GetConfigUrlsPath`, `GetConfigCrlPath`) for the various PKI endpoints.

[Source: api/v1alpha1/pkisecretengineconfig_types.go#L240-L242]

### PKISecretEngineRole Has 38 Keys — Largest `toMap()` in This Story

`PKIRole.toMap()` produces 38 keys covering certificate issuance parameters. All keys are unconditional. No unexported fields.

[Source: api/v1alpha1/pkisecretenginerole_types.go#L325-L366]

### QuaySecretEngineConfig Custom `IsEquivalentToDesiredState` — `password` Delete Has No Effect

`IsEquivalentToDesiredState` deletes `password` from the desired state, but `toMap()` never sets a `password` key (it sets `token`). The delete is effectively a no-op for the current `toMap()` output. This may be a defensive measure or a bug. Test and document this behavior.

```go
func (q *QuaySecretEngineConfig) IsEquivalentToDesiredState(payload map[string]interface{}) bool {
    desiredState := q.Spec.QuayConfig.toMap()
    delete(desiredState, "password")
    return reflect.DeepEqual(desiredState, payload)
}
```

[Source: api/v1alpha1/quaysecretengineconfig_types.go#L74-L78]

### QuaySecretEngineConfig Uses `retrievedToken` Unexported Field

The `token` value in `QuayConfig.toMap()` comes from `retrievedToken`, set via `SetToken()`. In tests, set it directly.

[Source: api/v1alpha1/quaysecretengineconfig_types.go#L167-L168, L189-L191]

### QuaySecretEngineConfig Has Typo in Field Name: `CACertertificate`

The exported field on `QuayConfig` is `CACertertificate` (double "er" in "certificate"). The map key is correctly spelled as `ca_certificate`. Tests should use the misspelled field name as it appears in the struct.

### Quay Role/StaticRole Have Conditional Keys

Both `QuayRole.toMap()` and `QuayBaseRole.toMap()` conditionally include `default_permission`, `teams`, and `repositories` only when non-nil. Test both paths:
- All optional fields set → verify all keys present
- All optional fields nil → verify keys absent

### Quay Role/StaticRole Use JSON String Encoding for `teams` and `repositories`

When present, `teams` and `repositories` are serialized to JSON strings via `setMapJson()`. The map values are `string` type, not Go maps/slices. For `reflect.DeepEqual`, the payload must use matching JSON strings.

[Source: api/v1alpha1/quaysecretenginerole_types.go#L133-L139]

### QuaySecretEngineStaticRole Path Uses `static-roles` Not `roles`

`GetPath()` uses `{path}/static-roles/{name}` vs `{path}/roles/{name}` for `QuaySecretEngineRole`. Test this difference explicitly.

[Source: api/v1alpha1/quaysecretenginestaticrole_types.go#L59-L64]

### QuaySecretEngineStaticRole Uses `QuayBaseRole.toMap()` — No `ttl`/`max_ttl`

The static role uses `QuayBaseRole.toMap()` which does NOT include `ttl`/`max_ttl` keys. The dynamic `QuayRole` extends `QuayBaseRole` and adds these fields. Verify the static role map has fewer keys.

### KubernetesSecretEngineConfig Deletes `service_account_jwt` Before Comparison

Like GitHubSecretEngineConfig with `prv_key`, this type strips the JWT from the desired state because Vault redacts it in read responses.

```go
func (d *KubernetesSecretEngineConfig) IsEquivalentToDesiredState(payload map[string]interface{}) bool {
    desiredState := d.Spec.KubeSEConfig.toMap()
    delete(desiredState, "service_account_jwt")
    return reflect.DeepEqual(desiredState, payload)
}
```

Test:
- Payload matching `{kubernetes_host: X, kubernetes_ca_cert: Y, disable_local_ca_jwt: Z}` → `true`
- Payload with `service_account_jwt` present → `false`

[Source: api/v1alpha1/kubernetessecretengineconfig_types.go#L119-L123]

### KubernetesSecretEngineConfig Uses `retrievedServiceAccountJWT` Unexported Field

Set `retrievedServiceAccountJWT` directly in tests to provide the JWT value for `toMap()`.

[Source: api/v1alpha1/kubernetessecretengineconfig_types.go#L197-L204]

### KubernetesSecretEngineRole `DefaultTTL`/`MaxTTL` Field-to-Key Mapping Appears Swapped

In `KubeSERole.toMap()`:
- `DefaultTTL` → `token_max_ttl`
- `MaxTTL` → `token_default_ttl`

The field names suggest the mapping should be reversed (`DefaultTTL` → `token_default_ttl`), but the code maps them as shown. **Test the actual code behavior, not the expected behavior.** Document this as a potential bug for future investigation.

```go
payload["token_max_ttl"] = i.DefaultTTL
payload["token_default_ttl"] = i.MaxTTL
```

[Source: api/v1alpha1/kubernetessecretenginerole_types.go#L163-L164]

### RabbitMQSecretEngineConfig Uses Unexported Credential Fields

`username` comes from `retrievedUsername` and `password` from `retrievedPassword`, set via `SetUsernameAndPassword()`. In tests, set these directly.

[Source: api/v1alpha1/rabbitmqsecretengineconfig_types.go#L90-L92, L129-L140]

### Slice Fields Stored as Go Types, Not `[]interface{}`

Like previous stories, all types in this story store slice fields as their Go types (`[]string`, etc.) rather than `[]interface{}`. If Vault returns `[]interface{}`, `reflect.DeepEqual` returns `false`. Document this behavior.

### Config Types That Return `IsDeletable() = false`

Four types return `false` for `IsDeletable()`: `RabbitMQSecretEngineConfig`, `GitHubSecretEngineConfig`, `PKISecretEngineConfig`, `QuaySecretEngineConfig`. These are "config" types that configure the engine itself.

Note: `AzureSecretEngineConfig` and `KubernetesSecretEngineConfig` return `true` for `IsDeletable()` — different from the typical config pattern.

### Implementation Pattern — Standard Go `testing` Package

All tests in `api/v1alpha1/` use standard Go `testing` package (NOT Ginkgo). Follow the exact pattern from existing tests in stories 1.1–1.4.

**Build tag**: No build tag needed — files in `api/v1alpha1/` run with default `go test`.

**Import pattern**:
```go
package v1alpha1

import (
    "reflect"
    "testing"

    metav1 "k8s.io/apimachinery/pkg/apis/meta-v1"
)
```

### Previous Story Intelligence (Stories 1.1–1.4)

Established patterns to follow:
- Table-driven tests with `t.Run` subtests
- `reflect.DeepEqual` for map comparisons
- Testing both positive (matching) and negative (non-matching) cases
- Extra-fields behavior: documented as-is (tests proving behavior without changing production code)
- Tests for `GetPath` with and without `spec.name` override (where applicable)
- Tests for `IsDeletable` and `GetConditions`/`SetConditions`
- Unexported fields accessed directly within same package
- No build tags needed for `api/v1alpha1/` test files

Key insight from stories 1.3/1.4: Custom `IsEquivalentToDesiredState` patterns vary per type:
- Story 1.3: `DatabaseSecretEngineConfig` remaps fields to match Vault's restructured read response
- Story 1.4: `LDAPAuthEngineConfig` deletes `bindpass` from desiredState
- This story: 4 types with custom logic (GitHub deletes `prv_key`, Quay deletes `password`, K8s deletes `service_account_jwt`, RabbitMQ uses a different map method entirely)

### Project Structure Notes

Create 13 new test files:
- `api/v1alpha1/rabbitmqsecretengineconfig_test.go`
- `api/v1alpha1/rabbitmqsecretenginerole_test.go`
- `api/v1alpha1/githubsecretengineconfig_test.go`
- `api/v1alpha1/githubsecretenginerole_test.go`
- `api/v1alpha1/azuresecretengineconfig_test.go`
- `api/v1alpha1/azuresecretenginerole_test.go`
- `api/v1alpha1/pkisecretengineconfig_test.go`
- `api/v1alpha1/pkisecretenginerole_test.go`
- `api/v1alpha1/quaysecretengineconfig_test.go`
- `api/v1alpha1/quaysecretenginerole_test.go`
- `api/v1alpha1/quaysecretenginestaticrole_test.go`
- `api/v1alpha1/kubernetessecretengineconfig_test.go`
- `api/v1alpha1/kubernetessecretenginerole_test.go`

No changes to `controllers/` directory. No decoder methods needed (unit tests only).

### References

- [Source: api/v1alpha1/rabbitmqsecretengineconfig_types.go#L129-L140] — RMQSEConfig.rabbitMQToMap() (6 keys)
- [Source: api/v1alpha1/rabbitmqsecretengineconfig_types.go#L179-L182] — IsEquivalentToDesiredState (uses leasesToMap)
- [Source: api/v1alpha1/rabbitmqsecretengineconfig_types.go#L322-L327] — leasesToMap() (2 keys)
- [Source: api/v1alpha1/rabbitmqsecretengineconfig_types.go#L329-L335] — GetLeasePayload, GetLeasePath
- [Source: api/v1alpha1/rabbitmqsecretenginerole_types.go#L178-L181] — IsEquivalentToDesiredState (bare DeepEqual)
- [Source: api/v1alpha1/rabbitmqsecretenginerole_types.go#L233-L239] — RMQSERole.rabbitMQToMap() (3 keys)
- [Source: api/v1alpha1/rabbitmqsecretenginerole_types.go#L199-L231] — convertVhostsToJson, convertTopicsToJson
- [Source: api/v1alpha1/githubsecretengineconfig_types.go#L75-L80] — GHConfig.toMap() (3 keys)
- [Source: api/v1alpha1/githubsecretengineconfig_types.go#L108-L112] — IsEquivalentToDesiredState (deletes prv_key)
- [Source: api/v1alpha1/githubsecretenginerole_types.go#L84-L92] — PermissionSet.toMap() (5 keys)
- [Source: api/v1alpha1/githubsecretenginerole_types.go#L109-L111] — IsEquivalentToDesiredState (bare DeepEqual)
- [Source: api/v1alpha1/githubsecretenginerole_types.go#L99-L104] — GetPath (permissionset path)
- [Source: api/v1alpha1/azuresecretengineconfig_types.go#L259-L270] — AzureSEConfig.toMap() (7 keys)
- [Source: api/v1alpha1/azuresecretengineconfig_types.go#L159-L162] — IsEquivalentToDesiredState
- [Source: api/v1alpha1/azuresecretenginerole_types.go#L196-L209] — AzureSERole.toMap() (9 keys)
- [Source: api/v1alpha1/azuresecretenginerole_types.go#L167-L170] — IsEquivalentToDesiredState
- [Source: api/v1alpha1/pkisecretengineconfig_types.go#L523-L549] — PKICommon.toMap() (22 keys)
- [Source: api/v1alpha1/pkisecretengineconfig_types.go#L552-L559] — PKIConfigUrls.toMap() (3 keys)
- [Source: api/v1alpha1/pkisecretengineconfig_types.go#L562-L568] — PKIConfigCRL.toMap() (2 keys)
- [Source: api/v1alpha1/pkisecretengineconfig_types.go#L439-L442] — IsEquivalentToDesiredState (PKICommon only)
- [Source: api/v1alpha1/pkisecretengineconfig_types.go#L240-L242] — GetPath (mount path only)
- [Source: api/v1alpha1/pkisecretenginerole_types.go#L325-L366] — PKIRole.toMap() (38 keys)
- [Source: api/v1alpha1/pkisecretenginerole_types.go#L74-L77] — IsEquivalentToDesiredState
- [Source: api/v1alpha1/quaysecretengineconfig_types.go#L214-L221] — QuayConfig.toMap() (4 keys)
- [Source: api/v1alpha1/quaysecretengineconfig_types.go#L74-L78] — IsEquivalentToDesiredState (deletes password)
- [Source: api/v1alpha1/quaysecretenginerole_types.go#L113-L131] — QuayRole.toMap() (5-8 keys, conditional)
- [Source: api/v1alpha1/quaysecretenginerole_types.go#L92-L95] — IsEquivalentToDesiredState
- [Source: api/v1alpha1/quaysecretenginerole_types.go#L133-L139] — setMapJson helper
- [Source: api/v1alpha1/quaysecretenginestaticrole_types.go#L93-L109] — QuayBaseRole.toMap() (3-6 keys, conditional)
- [Source: api/v1alpha1/quaysecretenginestaticrole_types.go#L68-L71] — IsEquivalentToDesiredState
- [Source: api/v1alpha1/quaysecretenginestaticrole_types.go#L59-L64] — GetPath (static-roles)
- [Source: api/v1alpha1/kubernetessecretengineconfig_types.go#L197-L204] — KubeSEConfig.toMap() (4 keys)
- [Source: api/v1alpha1/kubernetessecretengineconfig_types.go#L119-L123] — IsEquivalentToDesiredState (deletes service_account_jwt)
- [Source: api/v1alpha1/kubernetessecretenginerole_types.go#L158-L173] — KubeSERole.toMap() (12 keys, potential field swap)
- [Source: api/v1alpha1/kubernetessecretenginerole_types.go#L77-L80] — IsEquivalentToDesiredState
- [Source: _bmad-output/implementation-artifacts/1-4-unit-tests-tomap-and-isequivalenttodesiredstate-auth-engine-config-types.md] — Previous story patterns
- [Source: _bmad-output/project-context.md] — Testing rules and conventions

## Dev Agent Record

### Agent Model Used

### Debug Log References

### Completion Notes List

### File List
