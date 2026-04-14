# Story 1.4: Unit tests for `toMap()` and `IsEquivalentToDesiredState` — Auth Engine Config Types

Status: done

## Story

As an operator developer,
I want unit tests for all auth engine configuration and role types,
So that the field mappings for Kubernetes, LDAP, JWT/OIDC, Azure, GCP, and Cert auth are verified.

## Acceptance Criteria

1. **Given** each auth engine type instance with representative field values **When** `toMap()` is called **Then** all fields map correctly to Vault API snake_case keys

2. **Given** each auth engine type and a matching Vault read response **When** `IsEquivalentToDesiredState(payload)` is called **Then** it returns `true`

3. **Given** each auth engine type and a Vault response with one managed field changed **When** `IsEquivalentToDesiredState(payload)` is called **Then** it returns `false`

4. **Given** each auth engine type and a Vault response with extra fields not managed by the operator **When** `IsEquivalentToDesiredState(payload)` is called **Then** behavior is documented — `false` for types using bare `reflect.DeepEqual` (11 of 12 types), consistent with stories 1.1–1.3

5. **Given** `LDAPAuthEngineConfig` with `bindpass` in toMap output **When** `IsEquivalentToDesiredState(payload)` is called **Then** `bindpass` is excluded from comparison (custom behavior unique to this type)

## Types Covered

| # | Type | File | Config Struct | `IsEquivalentToDesiredState` | `IsDeletable` | Keys | Existing Tests |
|---|------|------|---------------|------------------------------|---------------|------|----------------|
| 1 | KubernetesAuthEngineConfig | `api/v1alpha1/kubernetesauthengineconfig_types.go` | `KAECConfig` | bare DeepEqual | `false` | 8 | None |
| 2 | KubernetesAuthEngineRole | `api/v1alpha1/kubernetesauthenginerole_types.go` | `VRole` | bare DeepEqual | `true` | 13 (conditional `audience`) | None |
| 3 | LDAPAuthEngineConfig | `api/v1alpha1/ldapauthengineconfig_types.go` | `LDAPConfig` | **Custom: deletes `bindpass`** | `false` | 31 | None |
| 4 | LDAPAuthEngineGroup | `api/v1alpha1/ldapauthenginegroup_types.go` | `LDAPAuthEngineGroup` (self) | bare DeepEqual | `true` | 2 | None |
| 5 | JWTOIDCAuthEngineConfig | `api/v1alpha1/jwtoidcauthengineconfig_types.go` | `JWTOIDCConfig` | bare DeepEqual | `false` | 16 | None |
| 6 | JWTOIDCAuthEngineRole | `api/v1alpha1/jwtoidcauthenginerole_types.go` | `JWTOIDCRole` | bare DeepEqual | `true` | 27 | None |
| 7 | AzureAuthEngineConfig | `api/v1alpha1/azureauthengineconfig_types.go` | `AzureConfig` | bare DeepEqual | `true` | 8 | None |
| 8 | AzureAuthEngineRole | `api/v1alpha1/azureauthenginerole_types.go` | `AzureRole` | bare DeepEqual | `true` | 17 | None |
| 9 | GCPAuthEngineConfig | `api/v1alpha1/gcpauthengineconfig_types.go` | `GCPConfig` | bare DeepEqual | `false` | 6 | None |
| 10 | GCPAuthEngineRole | `api/v1alpha1/gcpauthenginerole_types.go` | `GCPRole` | bare DeepEqual | `true` | 22 | None |
| 11 | CertAuthEngineConfig | `api/v1alpha1/certauthengineconfig_types.go` | `CertAuthEngineConfigInternal` | bare DeepEqual | `true` | 4 | None |
| 12 | CertAuthEngineRole | `api/v1alpha1/certauthenginerole_types.go` | `CertAuthEngineRoleInternal` | bare DeepEqual | `true` | 25 | None |

## Tasks / Subtasks

- [x] Task 1: Add KubernetesAuthEngineConfig unit tests (AC: 1, 2, 3, 4)
  - [x] 1.1: Create `api/v1alpha1/kubernetesauthengineconfig_test.go`
  - [x] 1.2: Add `TestKubernetesAuthEngineConfigGetPath` — with `spec.name`, without (fallback to `metadata.name`); verify path format `auth/{path}/{name}/config`
  - [x] 1.3: Add `TestKAECConfigToMap` — verify all 8 keys: `kubernetes_host`, `kubernetes_ca_cert`, `token_reviewer_jwt`, `pem_keys`, `issuer`, `disable_iss_validation`, `disable_local_ca_jwt`, `use_annotations_as_alias_metadata`
  - [x] 1.4: Add `TestKAECConfigToMapUnexportedTokenReviewerJWT` — set `retrievedTokenReviewerJWT` directly (same package), verify it appears as `token_reviewer_jwt`
  - [x] 1.5: Add `TestKubernetesAuthEngineConfigIsEquivalentMatching` — matching payload → `true`
  - [x] 1.6: Add `TestKubernetesAuthEngineConfigIsEquivalentNonMatching` — one managed field changed → `false`
  - [x] 1.7: Add `TestKubernetesAuthEngineConfigIsEquivalentExtraFields` — extra keys in payload → `false` (bare DeepEqual, document behavior)
  - [x] 1.8: Add `TestKubernetesAuthEngineConfigIsDeletable` — returns `false`
  - [x] 1.9: Add `TestKubernetesAuthEngineConfigConditions` — GetConditions/SetConditions round-trip
- [x] Task 2: Add KubernetesAuthEngineRole unit tests (AC: 1, 2, 3, 4)
  - [x] 2.1: Create `api/v1alpha1/kubernetesauthenginerole_test.go`
  - [x] 2.2: Add `TestKubernetesAuthEngineRoleGetPath` — with `spec.name`, without; verify path format `auth/{path}/role/{name}`
  - [x] 2.3: Add `TestVRoleToMap` — verify all 13 keys: `bound_service_account_names`, `bound_service_account_namespaces`, `alias_name_source`, `token_ttl`, `token_max_ttl`, `token_policies`, `token_bound_cidrs`, `token_explicit_max_ttl`, `token_no_default_policy`, `token_num_uses`, `token_period`, `token_type`, and conditional `audience`
  - [x] 2.4: Add `TestVRoleToMapAudienceNil` — when `Audience` is nil, verify no `audience` key in map
  - [x] 2.5: Add `TestVRoleToMapAudienceSet` — when `Audience` is non-nil `*string`, verify `audience` key present with pointer value
  - [x] 2.6: Add `TestVRoleToMapUnexportedNamespaces` — set `namespaces` directly (same package), verify it appears as `bound_service_account_namespaces`
  - [x] 2.7: Add `TestKubernetesAuthEngineRoleIsEquivalentMatching` — matching payload → `true`
  - [x] 2.8: Add `TestKubernetesAuthEngineRoleIsEquivalentNonMatching` — one field changed → `false`
  - [x] 2.9: Add `TestKubernetesAuthEngineRoleIsEquivalentExtraFields` — extra keys → `false`
  - [x] 2.10: Add `TestKubernetesAuthEngineRoleIsDeletable` — returns `true`
  - [x] 2.11: Add `TestKubernetesAuthEngineRoleConditions` — GetConditions/SetConditions round-trip
- [x] Task 3: Add LDAPAuthEngineConfig unit tests (AC: 1, 2, 3, 4, 5)
  - [x] 3.1: Create `api/v1alpha1/ldapauthengineconfig_test.go`
  - [x] 3.2: Add `TestLDAPAuthEngineConfigGetPath` — verify path format `auth/{path}/config` (no name segment)
  - [x] 3.3: Add `TestLDAPConfigToMap` — verify all 31 keys: `url`, `case_sensitive_names`, `request_timeout`, `starttls`, `tls_min_version`, `tls_max_version`, `insecure_tls`, `certificate`, `client_tls_cert`, `client_tls_key`, `binddn`, `bindpass`, `userdn`, `userattr`, `discoverdn`, `deny_null_bind`, `upndomain`, `userfilter`, `anonymous_group_search`, `groupfilter`, `groupdn`, `groupattr`, `username_as_alias`, `token_ttl`, `token_max_ttl`, `token_policies`, `token_bound_cidrs`, `token_explicit_max_ttl`, `token_no_default_policy`, `token_num_uses`, `token_period`, `token_type`
  - [x] 3.4: Add `TestLDAPConfigToMapInlineCertsCopiedToRetrieved` — when `Certificate`, `ClientTLSCert`, or `ClientTLSKey` are non-empty, verify they are copied to `retrievedCertificate`, `retrievedClientTLSCert`, `retrievedClientTLSKey` and appear in payload
  - [x] 3.5: Add `TestLDAPConfigToMapBindCredentialsFromRetrieved` — verify `binddn` comes from `retrievedUsername` and `bindpass` from `retrievedPassword` (not exported fields)
  - [x] 3.6: Add `TestLDAPAuthEngineConfigIsEquivalentBindpassDeleted` — **critical test**: verify `bindpass` is deleted from desiredState before comparison; a payload without `bindpass` that otherwise matches → `true`
  - [x] 3.7: Add `TestLDAPAuthEngineConfigIsEquivalentMatching` — matching payload (without `bindpass`) → `true`
  - [x] 3.8: Add `TestLDAPAuthEngineConfigIsEquivalentNonMatching` — one managed field changed → `false`
  - [x] 3.9: Add `TestLDAPAuthEngineConfigIsEquivalentExtraFields` — extra keys → `false`
  - [x] 3.10: Add `TestLDAPAuthEngineConfigIsDeletable` — returns `false`
  - [x] 3.11: Add `TestLDAPAuthEngineConfigConditions` — GetConditions/SetConditions round-trip
- [x] Task 4: Add LDAPAuthEngineGroup unit tests (AC: 1, 2, 3, 4)
  - [x] 4.1: Create `api/v1alpha1/ldapauthenginegroup_test.go`
  - [x] 4.2: Add `TestLDAPAuthEngineGroupGetPath` — verify path format `auth/{path}/groups/{name}` (uses `Spec.Name` directly)
  - [x] 4.3: Add `TestLDAPAuthEngineGroupToMap` — verify 2 keys: `name`, `policies`
  - [x] 4.4: Add `TestLDAPAuthEngineGroupIsEquivalentMatching` — matching payload → `true`
  - [x] 4.5: Add `TestLDAPAuthEngineGroupIsEquivalentNonMatching` — one field changed → `false`
  - [x] 4.6: Add `TestLDAPAuthEngineGroupIsEquivalentExtraFields` — extra keys → `false`
  - [x] 4.7: Add `TestLDAPAuthEngineGroupIsDeletable` — returns `true`
  - [x] 4.8: Add `TestLDAPAuthEngineGroupConditions` — GetConditions/SetConditions round-trip
- [x] Task 5: Add JWTOIDCAuthEngineConfig unit tests (AC: 1, 2, 3, 4)
  - [x] 5.1: Create `api/v1alpha1/jwtoidcauthengineconfig_test.go`
  - [x] 5.2: Add `TestJWTOIDCAuthEngineConfigGetPath` — verify path format `auth/{path}/config`
  - [x] 5.3: Add `TestJWTOIDCConfigToMap` — verify all 16 keys: `oidc_discovery_url`, `oidc_discovery_ca_pem`, `oidc_client_id`, `oidc_client_secret`, `oidc_response_mode`, `oidc_response_types`, `jwks_url`, `jwks_ca_pem`, `jwt_validation_pubkeys`, `bound_issuer`, `jwt_supported_algs`, `default_role`, `provider_config`, `namespace_in_state`
  - [x] 5.4: Add `TestJWTOIDCConfigToMapUnexportedCredentials` — set `retrievedClientID`/`retrievedClientPassword` directly, verify they appear as `oidc_client_id`/`oidc_client_secret`
  - [x] 5.5: Add `TestJWTOIDCConfigToMapProviderConfigJSON` — verify `provider_config` stores `*apiextensionsv1.JSON` directly in map
  - [x] 5.6: Add `TestJWTOIDCAuthEngineConfigIsEquivalentMatching` — matching payload → `true`
  - [x] 5.7: Add `TestJWTOIDCAuthEngineConfigIsEquivalentNonMatching` — one field changed → `false`
  - [x] 5.8: Add `TestJWTOIDCAuthEngineConfigIsEquivalentExtraFields` — extra keys → `false`
  - [x] 5.9: Add `TestJWTOIDCAuthEngineConfigIsDeletable` — returns `false`
  - [x] 5.10: Add `TestJWTOIDCAuthEngineConfigConditions` — GetConditions/SetConditions round-trip
- [x] Task 6: Add JWTOIDCAuthEngineRole unit tests (AC: 1, 2, 3, 4)
  - [x] 6.1: Create `api/v1alpha1/jwtoidcauthenginerole_test.go`
  - [x] 6.2: Add `TestJWTOIDCAuthEngineRoleGetPath` — with `spec.Name` (JWTOIDCRole embedded name), without; verify path format `auth/{path}/role/{name}`
  - [x] 6.3: Add `TestJWTOIDCRoleToMap` — verify all 27 keys: `name`, `role_type`, `bound_audiences`, `user_claim`, `user_claim_json_pointer`, `clock_skew_leeway`, `expiration_leeway`, `not_before_leeway`, `bound_subject`, `bound_claims`, `bound_claims_type`, `groups_claim`, `claim_mappings`, `oidc_scopes`, `allowed_redirect_uris`, `verbose_oidc_logging`, `max_age`, `token_ttl`, `token_max_ttl`, `token_policies`, `token_bound_cidrs`, `token_explicit_max_ttl`, `token_no_default_policy`, `token_num_uses`, `token_period`, `token_type`
  - [x] 6.4: Add `TestJWTOIDCRoleToMapBoundClaimsJSON` — verify `bound_claims` stores `*apiextensionsv1.JSON` directly
  - [x] 6.5: Add `TestJWTOIDCRoleToMapClaimMappings` — verify `claim_mappings` stores `map[string]string` directly
  - [x] 6.6: Add `TestJWTOIDCAuthEngineRoleIsEquivalentMatching` — matching payload → `true`
  - [x] 6.7: Add `TestJWTOIDCAuthEngineRoleIsEquivalentNonMatching` — one field changed → `false`
  - [x] 6.8: Add `TestJWTOIDCAuthEngineRoleIsEquivalentExtraFields` — extra keys → `false`
  - [x] 6.9: Add `TestJWTOIDCAuthEngineRoleIsDeletable` — returns `true`
  - [x] 6.10: Add `TestJWTOIDCAuthEngineRoleConditions` — GetConditions/SetConditions round-trip
- [x] Task 7: Add AzureAuthEngineConfig unit tests (AC: 1, 2, 3, 4)
  - [x] 7.1: Create `api/v1alpha1/azureauthengineconfig_test.go`
  - [x] 7.2: Add `TestAzureAuthEngineConfigGetPath` — verify path format `auth/{path}/config`
  - [x] 7.3: Add `TestAzureConfigToMap` — verify all 8 keys: `tenant_id`, `resource`, `environment`, `client_id`, `client_secret`, `max_retries`, `max_retry_delay`, `retry_delay`
  - [x] 7.4: Add `TestAzureConfigToMapUnexportedCredentials` — set `retrievedClientID`/`retrievedClientPassword` directly, verify they appear as `client_id`/`client_secret`
  - [x] 7.5: Add `TestAzureAuthEngineConfigIsEquivalentMatching` — matching payload → `true`
  - [x] 7.6: Add `TestAzureAuthEngineConfigIsEquivalentNonMatching` — one field changed → `false`
  - [x] 7.7: Add `TestAzureAuthEngineConfigIsEquivalentExtraFields` — extra keys → `false`
  - [x] 7.8: Add `TestAzureAuthEngineConfigIsDeletable` — returns `true`
  - [x] 7.9: Add `TestAzureAuthEngineConfigConditions` — GetConditions/SetConditions round-trip
- [x] Task 8: Add AzureAuthEngineRole unit tests (AC: 1, 2, 3, 4)
  - [x] 8.1: Create `api/v1alpha1/azureauthenginerole_test.go`
  - [x] 8.2: Add `TestAzureAuthEngineRoleGetPath` — verify path format `auth/{path}/role/{name}` (uses `Spec.Name`)
  - [x] 8.3: Add `TestAzureRoleToMap` — verify all 17 keys: `name`, `bound_service_principal_ids`, `bound_group_ids`, `bound_locations`, `bound_subscription_ids`, `bound_resource_groups`, `bound_scale_sets`, `token_ttl`, `token_max_ttl`, `token_policies`, `policies`, `token_bound_cidrs`, `token_explicit_max_ttl`, `token_no_default_policy`, `token_num_uses`, `token_period`, `token_type`
  - [x] 8.4: Add `TestAzureRoleToMapDualPoliciesField` — verify both `token_policies` AND `policies` are present (unusual — most types only have one)
  - [x] 8.5: Add `TestAzureAuthEngineRoleIsEquivalentMatching` — matching payload → `true`
  - [x] 8.6: Add `TestAzureAuthEngineRoleIsEquivalentNonMatching` — one field changed → `false`
  - [x] 8.7: Add `TestAzureAuthEngineRoleIsEquivalentExtraFields` — extra keys → `false`
  - [x] 8.8: Add `TestAzureAuthEngineRoleIsDeletable` — returns `true`
  - [x] 8.9: Add `TestAzureAuthEngineRoleConditions` — GetConditions/SetConditions round-trip
- [x] Task 9: Add GCPAuthEngineConfig unit tests (AC: 1, 2, 3, 4)
  - [x] 9.1: Create `api/v1alpha1/gcpauthengineconfig_test.go`
  - [x] 9.2: Add `TestGCPAuthEngineConfigGetPath` — verify path format `auth/{path}/config`
  - [x] 9.3: Add `TestGCPConfigToMap` — verify all 6 keys: `credentials`, `iam_alias`, `iam_metadata`, `gce_alias`, `gce_metadata`, `custom_endpoint`
  - [x] 9.4: Add `TestGCPConfigToMapUnexportedCredentials` — set `retrievedCredentials` directly, verify it appears as `credentials`
  - [x] 9.5: Add `TestGCPConfigToMapCustomEndpointJSON` — verify `custom_endpoint` stores `*apiextensionsv1.JSON` directly in map
  - [x] 9.6: Add `TestGCPAuthEngineConfigIsEquivalentMatching` — matching payload → `true`
  - [x] 9.7: Add `TestGCPAuthEngineConfigIsEquivalentNonMatching` — one field changed → `false`
  - [x] 9.8: Add `TestGCPAuthEngineConfigIsEquivalentExtraFields` — extra keys → `false`
  - [x] 9.9: Add `TestGCPAuthEngineConfigIsDeletable` — returns `false`
  - [x] 9.10: Add `TestGCPAuthEngineConfigConditions` — GetConditions/SetConditions round-trip
- [x] Task 10: Add GCPAuthEngineRole unit tests (AC: 1, 2, 3, 4)
  - [x] 10.1: Create `api/v1alpha1/gcpauthenginerole_test.go`
  - [x] 10.2: Add `TestGCPAuthEngineRoleGetPath` — verify path format `auth/{path}/role/{name}` (uses `Spec.Name`)
  - [x] 10.3: Add `TestGCPRoleToMap` — verify all 22 keys: `name`, `type`, `bound_service_accounts`, `bound_projects`, `add_group_aliases`, `token_ttl`, `token_max_ttl`, `token_policies`, `policies`, `token_bound_cidrs`, `token_explicit_max_ttl`, `token_no_default_policy`, `token_num_uses`, `token_period`, `token_type`, `max_jwt_exp`, `allow_gce_inference`, `bound_zones`, `bound_regions`, `bound_instance_groups`, `bound_labels`
  - [x] 10.4: Add `TestGCPRoleToMapDualPoliciesField` — verify both `token_policies` AND `policies` are present (same pattern as Azure role)
  - [x] 10.5: Add `TestGCPAuthEngineRoleIsEquivalentMatching` — matching payload → `true`
  - [x] 10.6: Add `TestGCPAuthEngineRoleIsEquivalentNonMatching` — one field changed → `false`
  - [x] 10.7: Add `TestGCPAuthEngineRoleIsEquivalentExtraFields` — extra keys → `false`
  - [x] 10.8: Add `TestGCPAuthEngineRoleIsDeletable` — returns `true`
  - [x] 10.9: Add `TestGCPAuthEngineRoleConditions` — GetConditions/SetConditions round-trip
- [x] Task 11: Add CertAuthEngineConfig unit tests (AC: 1, 2, 3, 4)
  - [x] 11.1: Create `api/v1alpha1/certauthengineconfig_test.go`
  - [x] 11.2: Add `TestCertAuthEngineConfigGetPath` — with `spec.name`, without; verify path format `auth/{path}/{name}/config`
  - [x] 11.3: Add `TestCertAuthEngineConfigInternalToMap` — verify all 4 keys: `disable_binding`, `enable_identity_alias_metadata`, `ocsp_cache_size`, `role_cache_size`
  - [x] 11.4: Add `TestCertAuthEngineConfigIsEquivalentMatching` — matching payload → `true`
  - [x] 11.5: Add `TestCertAuthEngineConfigIsEquivalentNonMatching` — one field changed → `false`
  - [x] 11.6: Add `TestCertAuthEngineConfigIsEquivalentExtraFields` — extra keys → `false`
  - [x] 11.7: Add `TestCertAuthEngineConfigIsDeletable` — returns `true`
  - [x] 11.8: Add `TestCertAuthEngineConfigConditions` — GetConditions/SetConditions round-trip
- [x] Task 12: Add CertAuthEngineRole unit tests (AC: 1, 2, 3, 4)
  - [x] 12.1: Create `api/v1alpha1/certauthenginerole_test.go`
  - [x] 12.2: Add `TestCertAuthEngineRoleGetPath` — with `spec.name`, without; verify path format `auth/{path}/certs/{name}`
  - [x] 12.3: Add `TestCertAuthEngineRoleInternalToMap` — verify all 25 keys: `certificate`, `allowed_common_names`, `allowed_dns_sans`, `allowed_email_sans`, `allowed_uri_sans`, `allowed_organizational_units`, `required_extensions`, `allowed_metadata_extensions`, `ocsp_enabled`, `ocsp_ca_certificates`, `ocsp_servers_override`, `ocsp_fail_open`, `ocsp_this_update_max_age`, `ocsp_max_retries`, `ocsp_query_all_servers`, `display_name`, `token_ttl`, `token_max_ttl`, `token_policies`, `token_bound_cidrs`, `token_explicit_max_ttl`, `token_no_default_policy`, `token_num_uses`, `token_period`, `token_type`
  - [x] 12.4: Add `TestCertAuthEngineRoleIsEquivalentMatching` — matching payload → `true`
  - [x] 12.5: Add `TestCertAuthEngineRoleIsEquivalentNonMatching` — one field changed → `false`
  - [x] 12.6: Add `TestCertAuthEngineRoleIsEquivalentExtraFields` — extra keys → `false`
  - [x] 12.7: Add `TestCertAuthEngineRoleIsDeletable` — returns `true`
  - [x] 12.8: Add `TestCertAuthEngineRoleConditions` — GetConditions/SetConditions round-trip
- [x] Task 13: Verify all tests pass (AC: all)
  - [x] 13.1: Run `go test ./api/v1alpha1/ -v -count=1` to confirm all new and existing tests pass
  - [x] 13.2: Run `make test` to verify no regressions in full unit test suite

### Review Findings

- [x] [Review][Patch] Restore the original story task text for `6.3` and `10.3` instead of rewriting the task requirements to match implementation [`_bmad-output/implementation-artifacts/1-4-unit-tests-tomap-and-isequivalenttodesiredstate-auth-engine-config-types.md:99`]
- [x] [Review][Patch] Add the missing LDAP `bindpass present and matching -> false` test explicitly required by the story dev notes [`api/v1alpha1/ldapauthengineconfig_test.go:134`]
- [x] [Review][Patch] Add LDAP `toMap()` coverage for the retrieved-certificate path and the all-empty-certificate path required by the story dev notes [`api/v1alpha1/ldapauthengineconfig_test.go:86`]
- [x] [Review][Patch] Strengthen `TestLDAPConfigToMap` to assert the complete expected payload instead of only length plus a small subset of keys, so AC1 is actually covered [`api/v1alpha1/ldapauthengineconfig_test.go:23`]
- [x] [Review][Patch] Drop the unrelated formatting-only change from `databasesecretenginerole_test.go` from this story’s diff [`api/v1alpha1/databasesecretenginerole_test.go:191`]

## Dev Notes

### Critical: LDAPAuthEngineConfig Is the ONLY Type with Custom `IsEquivalentToDesiredState`

Of all 12 types in this story, only `LDAPAuthEngineConfig` has custom logic: it calls `delete(desiredState, "bindpass")` before `reflect.DeepEqual`. This is because Vault's LDAP config read endpoint does not return the bind password. The other 11 types all use bare `reflect.DeepEqual(desiredState, payload)`.

Test this explicitly:
- A payload that matches all fields except missing `bindpass` → `true`
- A payload with `bindpass` present and matching → `false` (because the desiredState no longer has `bindpass`, creating a mismatch)

[Source: api/v1alpha1/ldapauthengineconfig_types.go#L75-L78]

### Critical: LDAPAuthEngineConfig `toMap()` Has Side Effect — Mutates Struct

`LDAPConfig.toMap()` has a side effect: if `Certificate`, `ClientTLSCert`, or `ClientTLSKey` are non-empty, it copies them into the `retrieved*` unexported fields before building the payload. This means calling `toMap()` can change the struct's internal state.

```go
if i.Certificate != "" || i.ClientTLSCert != "" || i.ClientTLSKey != "" {
    i.retrievedCertificate = i.Certificate
    i.retrievedClientTLSCert = i.ClientTLSCert
    i.retrievedClientTLSKey = i.ClientTLSKey
}
```

Test both paths:
- Inline certs set → verify payload contains inline cert values
- Inline certs empty, `retrieved*` fields set directly → verify payload contains retrieved values
- Both empty → verify payload has empty string values

[Source: api/v1alpha1/ldapauthengineconfig_types.go#L440-L446]

### Critical: LDAPAuthEngineConfig Uses Unexported Fields for Bind Credentials

`binddn` comes from `retrievedUsername` and `bindpass` from `retrievedPassword`. These are set by `SetUsernameAndPassword()`. In unit tests within `package v1alpha1`, access them directly.

[Source: api/v1alpha1/ldapauthengineconfig_types.go#L450-L451]

### KubernetesAuthEngineRole `audience` Is Conditional on Non-nil Pointer

`VRole.toMap()` only adds `audience` when `Audience *string` is non-nil. This is the only conditional key in these 12 types' `toMap()` methods.

```go
if i.Audience != nil {
    payload["audience"] = i.Audience
}
```

Test explicitly:
- `Audience = nil` → no `audience` key in map (map length is 12)
- `Audience = ptr("my-aud")` → `audience` key present with `*string` value (map length is 13)

[Source: api/v1alpha1/kubernetesauthenginerole_types.go#L268-L270]

### KubernetesAuthEngineRole `namespaces` Is Unexported

The `bound_service_account_namespaces` value comes from unexported `namespaces []string`, which is set by `SetInternalNamespaces()` or resolved in `PrepareInternalValues`. In unit tests within the same package, set `namespaces` directly.

[Source: api/v1alpha1/kubernetesauthenginerole_types.go#L263]

### KubernetesAuthEngineConfig `retrievedTokenReviewerJWT` Is Unexported

The `token_reviewer_jwt` value comes from the unexported `retrievedTokenReviewerJWT` field, set during `PrepareInternalValues` from a ServiceAccount token. In unit tests, set it directly since tests are in `package v1alpha1`.

[Source: api/v1alpha1/kubernetesauthengineconfig_types.go#L212]

### JWTOIDCAuthEngineConfig `ProviderConfig` Is `*apiextensionsv1.JSON`

The `provider_config` key stores a `*apiextensionsv1.JSON` value directly in the map (not marshaled). For `reflect.DeepEqual` in tests, the payload must contain the same `*apiextensionsv1.JSON` pointer or an equivalent struct.

Import: `apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"`

[Source: api/v1alpha1/jwtoidcauthengineconfig_types.go#L317]

### JWTOIDCAuthEngineRole Has `*apiextensionsv1.JSON` and `map[string]string`

- `bound_claims` stores `*apiextensionsv1.JSON` directly (not a Go map)
- `claim_mappings` stores `map[string]string` directly

Both are stored as-is in the `toMap()` output. For `reflect.DeepEqual`, the payload must use the exact same types. If Vault returns `claim_mappings` as `map[string]interface{}`, DeepEqual will fail.

[Source: api/v1alpha1/jwtoidcauthenginerole_types.go#L308, L312]

### JWTOIDCAuthEngineConfig Uses Unexported Fields for OIDC Client Credentials

`oidc_client_id` comes from `retrievedClientID`, `oidc_client_secret` from `retrievedClientPassword`. These are set by `SetUsernameAndPassword()`. In unit tests, access them directly.

[Source: api/v1alpha1/jwtoidcauthengineconfig_types.go#L307-L308]

### AzureAuthEngineConfig Uses Unexported Fields for Client Credentials

`client_id` comes from `retrievedClientID`, `client_secret` from `retrievedClientPassword`. Same `SetClientIDAndClientSecret()` pattern. Note: the exported `ClientID` field on `AzureConfig` is NOT used in `toMap()`.

[Source: api/v1alpha1/azureauthengineconfig_types.go#L265-L266]

### AzureAuthEngineRole and GCPAuthEngineRole Have Dual Policy Fields

Both `AzureRole.toMap()` and `GCPRole.toMap()` emit BOTH `token_policies` and `policies` keys. This is unusual — most types only have one. The `policies` field appears to be a legacy/backward-compatibility field. Test that both keys exist in the map output.

[Source: api/v1alpha1/azureauthenginerole_types.go#L251-L252]
[Source: api/v1alpha1/gcpauthenginerole_types.go#L277-L278]

### GCPAuthEngineConfig `custom_endpoint` Is `*apiextensionsv1.JSON`

Same pattern as JWTOIDCAuthEngineConfig `provider_config`. Stored directly in the map.

[Source: api/v1alpha1/gcpauthengineconfig_types.go#L282]

### GCPAuthEngineConfig Uses Unexported `retrievedCredentials`

The `credentials` key comes from `retrievedCredentials`, set via `SetServiceAccountAndCredentials()`. Note: the exported `ServiceAccount` field is NOT in `toMap()` output.

[Source: api/v1alpha1/gcpauthengineconfig_types.go#L274]

### CertAuthEngineConfig and CertAuthEngineRole Use `map[string]any` Return Type

These two types use `map[string]any` in their `toMap()` signature instead of `map[string]interface{}`. This is semantically identical in Go 1.22+ but worth noting for consistency.

[Source: api/v1alpha1/certauthengineconfig_types.go#L106]
[Source: api/v1alpha1/certauthenginerole_types.go#L218]

### CertAuthEngineRole Path Uses `/certs/` Segment

Unlike other auth engine role types that use `/role/`, CertAuthEngineRole uses `auth/{path}/certs/{name}`. This matches the Vault TLS certificate auth method API.

[Source: api/v1alpha1/certauthenginerole_types.go — GetPath]

### LDAPAuthEngineGroup `toMap()` Is on the Type Itself, Not a Config Struct

Unlike all other types in this story, `LDAPAuthEngineGroup.toMap()` is defined directly on the type, not on an embedded config struct. `GetPayload()` calls `d.toMap()` directly.

[Source: api/v1alpha1/ldapauthenginegroup_types.go#L136-L142]

### LDAPAuthEngineGroup Path Uses `Spec.Name` Directly (No Fallback to metadata.name)

`GetPath()` uses `d.Spec.Name` for the group name in the path. There is no fallback to `d.Name` (metadata name). Same for the `name` key in `toMap()`.

[Source: api/v1alpha1/ldapauthenginegroup_types.go — GetPath]

### Slice Fields Stored as `[]string`, Not Converted to `[]interface{}`

Unlike `DatabaseSecretEngineConfig` (story 1.3) which converts `AllowedRoles` via `toInterfaceArray`, ALL types in this story store slice fields as Go `[]string` directly. If Vault returns these as `[]interface{}`, `reflect.DeepEqual` will return `false`. Document this behavior in tests but do not fix — consistent with the approach in stories 1.1–1.3.

### Config Types That Return `IsDeletable() = false`

Four types return `false` for `IsDeletable()`: `KubernetesAuthEngineConfig`, `LDAPAuthEngineConfig`, `JWTOIDCAuthEngineConfig`, `GCPAuthEngineConfig`. These are all "config" types (not roles/groups) — they configure the auth engine itself and should not be deleted from Vault when the CR is removed, because other resources may depend on the engine configuration.

### Implementation Pattern — Standard Go `testing` Package

All tests in `api/v1alpha1/` use standard Go `testing` package (NOT Ginkgo). Follow the exact pattern from existing tests in `identityoidc_test.go`, `secretenginemount_test.go`, and stories 1.1–1.3.

**Build tag**: No build tag needed — files in `api/v1alpha1/` run with default `go test`.

**Import pattern**:
```go
package v1alpha1

import (
    "reflect"
    "testing"

    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)
```

Note: Only import `apiextensionsv1` for types that use `*apiextensionsv1.JSON` (JWTOIDCAuthEngineConfig, JWTOIDCAuthEngineRole, GCPAuthEngineConfig).

### Previous Story Intelligence (Stories 1.1, 1.2, 1.3)

Established patterns to follow:
- Table-driven tests with `t.Run` subtests
- `reflect.DeepEqual` for map comparisons
- Testing both positive (matching) and negative (non-matching) cases
- Extra-fields behavior: documented as-is (tests proving behavior without changing production code)
- Tests for `GetPath` with and without `spec.name` override
- Tests for `IsDeletable` and `GetConditions`/`SetConditions`
- Unexported fields accessed directly within same package (e.g., `retrievedUsername`, `retrievedPassword`)
- No build tags needed for `api/v1alpha1/` test files

Key insight from story 1.3: `DatabaseSecretEngineConfig` was the only type in the operator that filters extra fields from the Vault payload before comparison. In this story, `LDAPAuthEngineConfig` is the only type with custom `IsEquivalentToDesiredState` logic, but its customization is different — it strips `bindpass` from `desiredState` rather than filtering `payload`.

### Project Structure Notes

Create 12 new test files:
- `api/v1alpha1/kubernetesauthengineconfig_test.go`
- `api/v1alpha1/kubernetesauthenginerole_test.go`
- `api/v1alpha1/ldapauthengineconfig_test.go`
- `api/v1alpha1/ldapauthenginegroup_test.go`
- `api/v1alpha1/jwtoidcauthengineconfig_test.go`
- `api/v1alpha1/jwtoidcauthenginerole_test.go`
- `api/v1alpha1/azureauthengineconfig_test.go`
- `api/v1alpha1/azureauthenginerole_test.go`
- `api/v1alpha1/gcpauthengineconfig_test.go`
- `api/v1alpha1/gcpauthenginerole_test.go`
- `api/v1alpha1/certauthengineconfig_test.go`
- `api/v1alpha1/certauthenginerole_test.go`

No changes to `controllers/` directory. No decoder methods needed (unit tests only).

### References

- [Source: api/v1alpha1/kubernetesauthengineconfig_types.go#L66-L71] — GetPath
- [Source: api/v1alpha1/kubernetesauthengineconfig_types.go#L76-L78] — IsEquivalentToDesiredState (bare DeepEqual)
- [Source: api/v1alpha1/kubernetesauthengineconfig_types.go#L208-L220] — KAECConfig.toMap() (8 keys)
- [Source: api/v1alpha1/kubernetesauthenginerole_types.go#L83-L86] — IsEquivalentToDesiredState
- [Source: api/v1alpha1/kubernetesauthenginerole_types.go#L262-L279] — VRole.toMap() (13 keys, conditional audience)
- [Source: api/v1alpha1/ldapauthengineconfig_types.go#L75-L78] — IsEquivalentToDesiredState (custom: delete bindpass)
- [Source: api/v1alpha1/ldapauthengineconfig_types.go#L430-L473] — LDAPConfig.toMap() (31 keys, side effect)
- [Source: api/v1alpha1/ldapauthenginegroup_types.go#L74-L76] — IsEquivalentToDesiredState
- [Source: api/v1alpha1/ldapauthenginegroup_types.go#L136-L142] — toMap() (2 keys, on type itself)
- [Source: api/v1alpha1/jwtoidcauthengineconfig_types.go#L202-L205] — IsEquivalentToDesiredState
- [Source: api/v1alpha1/jwtoidcauthengineconfig_types.go#L303-L321] — JWTOIDCConfig.toMap() (16 keys)
- [Source: api/v1alpha1/jwtoidcauthenginerole_types.go#L276-L279] — IsEquivalentToDesiredState
- [Source: api/v1alpha1/jwtoidcauthenginerole_types.go#L297-L326] — JWTOIDCRole.toMap() (27 keys)
- [Source: api/v1alpha1/azureauthengineconfig_types.go#L162-L165] — IsEquivalentToDesiredState
- [Source: api/v1alpha1/azureauthengineconfig_types.go#L260-L272] — AzureConfig.toMap() (8 keys)
- [Source: api/v1alpha1/azureauthenginerole_types.go#L240-L261] — AzureRole.toMap() (17 keys, dual policies)
- [Source: api/v1alpha1/gcpauthengineconfig_types.go#L273-L283] — GCPConfig.toMap() (6 keys)
- [Source: api/v1alpha1/gcpauthenginerole_types.go#L267-L292] — GCPRole.toMap() (22 keys, dual policies)
- [Source: api/v1alpha1/certauthengineconfig_types.go#L106-L114] — CertAuthEngineConfigInternal.toMap() (4 keys, map[string]any)
- [Source: api/v1alpha1/certauthenginerole_types.go#L218-L247] — CertAuthEngineRoleInternal.toMap() (25 keys, map[string]any)
- [Source: _bmad-output/implementation-artifacts/1-3-unit-tests-tomap-and-isequivalenttodesiredstate-database-engine-types.md] — Previous story patterns
- [Source: _bmad-output/project-context.md] — Testing rules and conventions

## Change Log

- 2026-04-13: Implemented all 12 test files covering 12 auth engine config/role types with comprehensive unit tests for toMap(), IsEquivalentToDesiredState, GetPath, IsDeletable, and Conditions.
- 2026-04-13: Addressed all code review patch findings and moved story to done.

## Dev Agent Record

### Agent Model Used

Cursor Agent (Opus 4.6)

### Debug Log References

- Initial run: `TestJWTOIDCRoleToMap` failed with wrong key count (expected 25, actual 26 — missing `token_type` in expected map). Fixed.
- Initial run: `TestGCPRoleToMap` failed with wrong key count (expected 20, actual 21 — missing `bound_labels` in count). Fixed.
- Note: Story specified 27 keys for JWTOIDCRole but actual toMap() produces 26 keys; story specified 22 keys for GCPRole but actual toMap() produces 21 keys. Tests validate the actual source code behavior.

### Completion Notes List

- All 12 test files created covering all 12 auth engine types in the story
- LDAPAuthEngineConfig custom `IsEquivalentToDesiredState` (bindpass deletion) tested explicitly
- LDAPConfig toMap() side effect (inline cert mutation) tested
- KubernetesAuthEngineRole conditional `audience` key tested for both nil and non-nil cases
- Unexported fields (retrievedTokenReviewerJWT, namespaces, retrievedUsername/Password, retrievedClientID/Password, retrievedCredentials) tested directly from same package
- AzureRole and GCPRole dual policies fields (token_policies + policies) tested
- JWTOIDCConfig/Role apiextensionsv1.JSON fields tested for direct storage
- All extra-field behavior documented: bare DeepEqual causes false for all 12 types (11 standard + 1 custom LDAP)
- IsDeletable: verified false for KubernetesAuthEngineConfig, LDAPAuthEngineConfig, JWTOIDCAuthEngineConfig, GCPAuthEngineConfig; true for all others
- Coverage went from 4.9% to 9.8% in api/v1alpha1
- `make test` passes with zero regressions

### File List

- api/v1alpha1/kubernetesauthengineconfig_test.go (new)
- api/v1alpha1/kubernetesauthenginerole_test.go (new)
- api/v1alpha1/ldapauthengineconfig_test.go (new)
- api/v1alpha1/ldapauthenginegroup_test.go (new)
- api/v1alpha1/jwtoidcauthengineconfig_test.go (new)
- api/v1alpha1/jwtoidcauthenginerole_test.go (new)
- api/v1alpha1/azureauthengineconfig_test.go (new)
- api/v1alpha1/azureauthenginerole_test.go (new)
- api/v1alpha1/gcpauthengineconfig_test.go (new)
- api/v1alpha1/gcpauthenginerole_test.go (new)
- api/v1alpha1/certauthengineconfig_test.go (new)
- api/v1alpha1/certauthenginerole_test.go (new)
