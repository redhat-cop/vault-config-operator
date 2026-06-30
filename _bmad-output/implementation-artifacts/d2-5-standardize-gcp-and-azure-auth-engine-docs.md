# Story D2.5: Standardize GCP and Azure Auth Engine Docs

Status: ready-for-dev

## Story

As a user configuring cloud provider authentication,
I want consistent docs for GCP and Azure auth,
So that I can configure cloud auth without confusion.

## Acceptance Criteria

1. **Given** the existing GCPAuthEngine content in `docs/auth-engines.md` (lines 425-588) **When** it is extracted to `docs/auth-engines/gcp.md` and standardized per the template **Then** it contains:
   - Overview linking to Vault GCP auth docs
   - GCPAuthEngineConfig: complete YAML example (with correct `path: gcp`), field descriptions (camelCase), Vault CLI equivalent
   - GCPAuthEngineRole: complete YAML example, field descriptions organized by role type (General, IAM-only, GCE-only), Vault CLI equivalent
   - Credential Resolution section documenting the three `GCPCredentials` methods

2. **Given** the existing AzureAuthEngine content in `docs/auth-engines.md` (lines 590-775) **When** it is extracted to `docs/auth-engines/azure.md` and standardized per the template **Then** it contains:
   - Overview linking to Vault Azure auth docs
   - AzureAuthEngineConfig: complete YAML example, field descriptions (camelCase), Vault CLI equivalent
   - AzureAuthEngineRole: complete YAML example, field descriptions, Vault CLI equivalent
   - Credential Resolution section documenting the three `azureCredentials` methods

3. **Given** the GCPAuthEngineRole field descriptions **When** validated against the Go type definitions **Then** ALL field names use camelCase (matching `json:` tags exactly) with no residual snake_case

4. **Given** the Azure credential resolution section **When** reviewed **Then** all three methods (Kubernetes Secret, Vault Secret, RandomSecret) are clearly documented with YAML examples using Pattern B (nested credential object)

5. **Given** both new files **When** validated against the template structure **Then** they follow the same structure as `docs/auth-engines/cert.md` (Overview → Config CRD → Role CRD → Credential Resolution → See Also)

6. **Given** both GCP and Azure docs **When** reviewed for Vault CLI equivalents **Then** each Config and Role section includes a Vault CLI equivalent command

## Tasks / Subtasks

- [ ] Task 1: Create `docs/auth-engines/gcp.md` (AC: 1, 3, 5, 6)
  - [ ] 1.1: Write Overview section — 2-3 sentences explaining GCP auth, link to Vault docs, list the two CRDs (GCPAuthEngineConfig, GCPAuthEngineRole), mention IAM and GCE role types
  - [ ] 1.2: Write GCPAuthEngineConfig section with Example YAML (fix `path: azure` → `path: gcp`), Vault CLI Equivalent, and Field Descriptions table
  - [ ] 1.3: Write GCPAuthEngineRole section with Example YAML, Vault CLI Equivalent, and Field Descriptions organized into groups: Role Identity, Token Parameters, IAM-Only Fields, GCE-Only Fields
  - [ ] 1.4: Write Credential Resolution section documenting all three `GCPCredentials` methods (Pattern B nested object) with YAML examples
  - [ ] 1.5: Add "See Also" section with links to auth-section.md, contributing-vault-apis.md, and Vault docs

- [ ] Task 2: Create `docs/auth-engines/azure.md` (AC: 2, 4, 5, 6)
  - [ ] 2.1: Write Overview section — 2-3 sentences explaining Azure auth, link to Vault docs, list the two CRDs (AzureAuthEngineConfig, AzureAuthEngineRole)
  - [ ] 2.2: Write AzureAuthEngineConfig section with Example YAML, Vault CLI Equivalent, and Field Descriptions table
  - [ ] 2.3: Write AzureAuthEngineRole section with Example YAML, Vault CLI Equivalent, and Field Descriptions table (Binding Fields + Token Parameters)
  - [ ] 2.4: Write Credential Resolution section documenting all three `azureCredentials` methods (Pattern B nested object) with YAML examples
  - [ ] 2.5: Add "See Also" section with links to auth-section.md, contributing-vault-apis.md, and Vault docs

- [ ] Task 3: Audit field names for camelCase consistency (AC: 3)
  - [ ] 3.1: Cross-reference all field names in `gcp.md` against `gcpauthengineconfig_types.go` and `gcpauthenginerole_types.go` — field names MUST match the `json:` tag values exactly
  - [ ] 3.2: Cross-reference all field names in `azure.md` against `azureauthengineconfig_types.go` and `azureauthenginerole_types.go` — field names MUST match the `json:` tag values exactly
  - [ ] 3.3: Fix any residual snake_case field names from the original `auth-engines.md` source

- [ ] Task 4: Verify links and structure (AC: 5)
  - [ ] 4.1: Verify relative links resolve correctly from `docs/auth-engines/gcp.md` and `docs/auth-engines/azure.md` (`../auth-section.md`, `../contributing-vault-apis.md`, `../secret-management.md`)
  - [ ] 4.2: Verify structure of both files matches `cert.md` pattern: heading hierarchy, section ordering, table format

## Dev Notes

### This Is a Documentation-Only Story

No Go code changes. No tests to run. No `make manifests generate`. The deliverables are:
- 1 new file: `docs/auth-engines/gcp.md`
- 1 new file: `docs/auth-engines/azure.md`

### Dependency on D2.1

This story assumes D2.1 has been completed (creating `docs/auth-engines/index.md` and the redirect pointer in `docs/auth-engines.md`). If D2.1 is NOT yet done, this story can still proceed — the files can be created independently. The index will reference them via `[gcp.md](gcp.md)` and `[azure.md](azure.md)`.

### Source Content Location

The content to extract and standardize is in `docs/auth-engines.md`:
- `## GCPAuthEngineConfig` (lines 425-496)
- `## GCPAuthEngineRole` (lines 498-588)
- `## AzureAuthEngineConfig` (lines 590-680)
- `## AzureAuthEngineRole` (lines 682-775)

### Template to Follow

Use `docs/engine-doc-template.md` as the structural pattern. Use `docs/auth-engines/cert.md` as the concrete reference implementation.

Key structural requirements from the template:
1. Title: `# GCP Auth Engine` / `# Azure Auth Engine`
2. Link to Vault docs immediately below title
3. `## Overview` — 2-3 sentences + CRD list
4. `## GCPAuthEngineConfig` → `### Example` → `### Vault CLI Equivalent` → `### Field Descriptions`
5. `## GCPAuthEngineRole` → `### Example` → `### Vault CLI Equivalent` → `### Field Descriptions`
6. `## Credential Resolution` — documenting the three credential methods (Pattern B nested object)
7. `## See Also`

### Known Issues in Source Content — MUST FIX

1. **GCPAuthEngineConfig example has wrong path** — The source (line 448) has `path: azure` which is a copy-paste error. It MUST be `path: gcp`.

2. **GCPAuthEngineConfig source uses unstructured field descriptions** — The source uses paragraphs instead of tables. Convert all descriptions to table format matching the cert.md pattern.

3. **GCPAuthEngineRole `type` field — Enum validation** — The Go type has `+kubebuilder:validation:Enum={"iam","gce"}`. Document these as the only allowed values.

4. **GCPAuthEngineRole `tokenType` field — Enum validation** — The Go type has `+kubebuilder:validation:Enum={"service","batch","default","default-service","default-batch"}`. Document all five allowed values.

5. **GCPAuthEngineConfig `GCEalias` field — Enum validation** — The Go type has `+kubebuilder:validation:Enum={"instance_id","role_id"}`. Document these as the only allowed values.

6. **AzureAuthEngineConfig `environment` field — Enum validation** — The Go type has `+kubebuilder:validation:Enum={"AzurePublicCloud","AzureUSGovernmentCloud","AzureChinaCloud","AzureGermanCloud"}`. Document all four allowed values.

7. **AzureAuthEngineRole `tokenType` field — Enum validation** — Same as GCP: `{"service","batch","default","default-service","default-batch"}`.

8. **GCPAuthEngineRole source uses snake_case in descriptions** — Descriptions reference `bound_projects`, `bound_instance_groups`, etc. Use camelCase field names in docs but retain the Vault API key references where needed for CLI equivalents.

### GCPAuthEngineConfig — Complete Field Reference

From `api/v1alpha1/gcpauthengineconfig_types.go`, the `GCPConfig` struct has these fields:

| CRD Field (camelCase) | JSON tag | Vault API key (snake_case) | Type | Required | Default |
|---|---|---|---|---|---|
| serviceAccount | `json:"serviceAccount,omitempty"` | — (used internally, not sent to Vault) | string | No | — |
| IAMalias | `json:"IAMalias"` | `iam_alias` | string | No | `"default"` |
| IAMmetadata | `json:"IAMmetadata"` | `iam_metadata` | string | No | `"default"` |
| GCEalias | `json:"GCEalias"` | `gce_alias` | string | No | `"role_id"` |
| GCEmetadata | `json:"GCEmetadata"` | `gce_metadata` | string | No | `"default"` |
| customEndpoint | `json:"customEndpoint,omitempty"` | `custom_endpoint` | JSON | No | `{}` |

Additional top-level spec fields (NOT in `GCPConfig` inline struct):
- `path` (Required) — mount path for the GCP auth engine
- `authentication` (Required) — see shared auth-section.md
- `connection` (Optional) — override Vault connection
- `GCPCredentials` (Optional) — credential resolution config for GCP service account JSON (see Credential Resolution below)

**Vault path:** `auth/{spec.path}/config`

**Important:** The `toMap()` method sends `credentials` (the resolved JSON string), `iam_alias`, `iam_metadata`, `gce_alias`, `gce_metadata`, and `custom_endpoint` to Vault. The `serviceAccount` field is used internally for credential resolution but is NOT sent as a separate key to the Vault config endpoint.

### GCPAuthEngineRole — Complete Field Reference

From `api/v1alpha1/gcpauthenginerole_types.go`, the `GCPRole` struct has these fields:

**Role Identity Fields:**

| CRD Field (camelCase) | JSON tag | Vault API key (snake_case) | Type | Required | Default |
|---|---|---|---|---|---|
| name | `json:"name"` | `name` | string | Yes | — |
| type | `json:"type"` | `type` | string | Yes | — (Enum: `iam`, `gce`) |
| boundServiceAccounts | `json:"boundServiceAccounts,omitempty"` | `bound_service_accounts` | []string | No | `[]` |
| boundProjects | `json:"boundProjects,omitempty"` | `bound_projects` | []string | No | `[]` |
| addGroupAliases | `json:"addGroupAliases,omitempty"` | `add_group_aliases` | bool | No | false |

**Token Parameters:**

| CRD Field (camelCase) | JSON tag | Vault API key (snake_case) | Type | Required | Default |
|---|---|---|---|---|---|
| tokenTTL | `json:"tokenTTL,omitempty"` | `token_ttl` | string | No | — |
| tokenMaxTTL | `json:"tokenMaxTTL,omitempty"` | `token_max_ttl` | string | No | — |
| tokenPolicies | `json:"tokenPolicies,omitempty"` | `token_policies` | []string | No | — |
| policies | `json:"policies,omitempty"` | `policies` | []string | No | — (DEPRECATED) |
| tokenBoundCIDRs | `json:"tokenBoundCIDRs,omitempty"` | `token_bound_cidrs` | []string | No | — |
| tokenExplicitMaxTTL | `json:"tokenExplicitMaxTTL,omitempty"` | `token_explicit_max_ttl` | string | No | — |
| tokenNoDefaultPolicy | `json:"tokenNoDefaultPolicy,omitempty"` | `token_no_default_policy` | bool | No | false |
| tokenNumUses | `json:"tokenNumUses,omitempty"` | `token_num_uses` | int64 | No | 0 |
| tokenPeriod | `json:"tokenPeriod,omitempty"` | `token_period` | int64 | No | 0 |
| tokenType | `json:"tokenType,omitempty"` | `token_type` | string | No | — (Enum: `service`, `batch`, `default`, `default-service`, `default-batch`) |

**IAM-Only Fields:**

| CRD Field (camelCase) | JSON tag | Vault API key (snake_case) | Type | Required | Default |
|---|---|---|---|---|---|
| maxJWTExp | `json:"maxJWTExp,omitempty"` | `max_jwt_exp` | string | No | — |
| allowGCEInference | `json:"allowGCEInference,omitempty"` | `allow_gce_inference` | bool | No | false |

**GCE-Only Fields:**

| CRD Field (camelCase) | JSON tag | Vault API key (snake_case) | Type | Required | Default |
|---|---|---|---|---|---|
| boundZones | `json:"boundZones,omitempty"` | `bound_zones` | []string | No | — |
| boundRegions | `json:"boundRegions,omitempty"` | `bound_regions` | []string | No | — |
| boundInstanceGroups | `json:"boundInstanceGroups,omitempty"` | `bound_instance_groups` | []string | No | — |
| boundLabels | `json:"boundLabels,omitempty"` | `bound_labels` | []string | No | — |

**Vault path:** `auth/{spec.path}/role/{spec.name}`

### AzureAuthEngineConfig — Complete Field Reference

From `api/v1alpha1/azureauthengineconfig_types.go`, the `AzureConfig` struct has these fields:

| CRD Field (camelCase) | JSON tag | Vault API key (snake_case) | Type | Required | Default |
|---|---|---|---|---|---|
| tenantID | `json:"tenantID"` | `tenant_id` | string | Yes | — |
| resource | `json:"resource"` | `resource` | string | Yes | — |
| environment | `json:"environment"` | `environment` | string | No | `"AzurePublicCloud"` (Enum: `AzurePublicCloud`, `AzureUSGovernmentCloud`, `AzureChinaCloud`, `AzureGermanCloud`) |
| clientID | `json:"clientID,omitempty"` | `client_id` | string | No | — |
| maxRetries | `json:"maxRetries"` | `max_retries` | int64 | No | `3` |
| maxRetryDelay | `json:"maxRetryDelay"` | `max_retry_delay` | int64 | No | `60` |
| retryDelay | `json:"retryDelay"` | `retry_delay` | int64 | No | `4` |

Additional top-level spec fields (NOT in `AzureConfig` inline struct):
- `path` (Required) — mount path for the Azure auth engine
- `authentication` (Required) — see shared auth-section.md
- `connection` (Optional) — override Vault connection
- `azureCredentials` (Optional) — credential resolution config for Azure Client ID + Client Secret (see Credential Resolution below)

**Vault path:** `auth/{spec.path}/config`

**Important:** The `toMap()` method sends `tenant_id`, `resource`, `environment`, `client_id` (resolved), `client_secret` (resolved), `max_retries`, `max_retry_delay`, and `retry_delay` to Vault. The `clientID` spec field takes precedence over the username from the referenced secret when resolving credentials.

### AzureAuthEngineRole — Complete Field Reference

From `api/v1alpha1/azureauthenginerole_types.go`, the `AzureRole` struct has these fields:

**Binding Fields:**

| CRD Field (camelCase) | JSON tag | Vault API key (snake_case) | Type | Required | Default |
|---|---|---|---|---|---|
| name | `json:"name"` | `name` | string | Yes | — |
| boundServicePrincipalIDs | `json:"boundServicePrincipalIDs,omitempty"` | `bound_service_principal_ids` | []string | No | — |
| boundGroupIDs | `json:"boundGroupIDs,omitempty"` | `bound_group_ids` | []string | No | — |
| boundLocations | `json:"boundLocations,omitempty"` | `bound_locations` | []string | No | — |
| boundSubscriptionIDs | `json:"boundSubscriptionIDs,omitempty"` | `bound_subscription_ids` | []string | No | — |
| boundResourceGroups | `json:"boundResourceGroups,omitempty"` | `bound_resource_groups` | []string | No | — |
| boundScaleSets | `json:"boundScaleSets,omitempty"` | `bound_scale_sets` | []string | No | — |

**Token Parameters:**

| CRD Field (camelCase) | JSON tag | Vault API key (snake_case) | Type | Required | Default |
|---|---|---|---|---|---|
| tokenTTL | `json:"tokenTTL,omitempty"` | `token_ttl` | string | No | — |
| tokenMaxTTL | `json:"tokenMaxTTL,omitempty"` | `token_max_ttl` | string | No | — |
| tokenPolicies | `json:"tokenPolicies,omitempty"` | `token_policies` | []string | No | — |
| policies | `json:"policies,omitempty"` | `policies` | []string | No | — (DEPRECATED) |
| tokenBoundCIDRs | `json:"tokenBoundCIDRs,omitempty"` | `token_bound_cidrs` | []string | No | — |
| tokenExplicitMaxTTL | `json:"tokenExplicitMaxTTL,omitempty"` | `token_explicit_max_ttl` | string | No | — |
| tokenNoDefaultPolicy | `json:"tokenNoDefaultPolicy,omitempty"` | `token_no_default_policy` | bool | No | false |
| tokenNumUses | `json:"tokenNumUses,omitempty"` | `token_num_uses` | int64 | No | 0 |
| tokenPeriod | `json:"tokenPeriod,omitempty"` | `token_period` | int64 | No | 0 |
| tokenType | `json:"tokenType,omitempty"` | `token_type` | string | No | — (Enum: `service`, `batch`, `default`, `default-service`, `default-batch`) |

**Vault path:** `auth/{spec.path}/role/{spec.name}`

### IsDeletable Behavior

- `GCPAuthEngineConfig` returns `IsDeletable() == false` — deleting the CR does NOT remove the GCP config from Vault. The auth mount must be disabled separately.
- `GCPAuthEngineRole` returns `IsDeletable() == true` — deleting the CR removes the role from Vault.
- `AzureAuthEngineConfig` returns `IsDeletable() == true` — deleting the CR DOES remove the Azure config from Vault.
- `AzureAuthEngineRole` returns `IsDeletable() == true` — deleting the CR removes the role from Vault.

**Document this difference** — GCP config is NOT cleaned up on CR deletion, but Azure config IS. This is an important behavioral difference users must be aware of.

### Credential Resolution (GCPCredentials)

The GCP engine uses `GCPCredentials` field of type `RootCredentialConfig`. This is **Pattern B** from the template (nested credential object).

The `GCPCredentials` field resolves the GCP service account JSON credentials used to authenticate with Google Cloud APIs.

**Important behaviors:**
- Default `usernameKey`: `"serviceaccount"` (maps to the service account email)
- Default `passwordKey`: `"credentials"` (maps to the service account JSON credentials)
- If `spec.serviceAccount` is set directly, it takes precedence over the username from the referenced secret
- If `GCPCredentials` is nil/empty AND matches the zero-value check `{UsernameKey: "serviceaccount", PasswordKey: "credentials"}`, credential resolution is skipped (GCP environment credentials will be used)

From `api/v1alpha1/utils/commons.go`, `RootCredentialConfig` has:
- `secret` — Kubernetes Secret reference
- `vaultSecret` — Vault secret path reference
- `randomSecret` — RandomSecret reference
- `usernameKey` — key for service account (default: `"serviceaccount"`)
- `passwordKey` — key for credentials JSON (default: `"credentials"`)

### Credential Resolution (azureCredentials)

The Azure engine uses `azureCredentials` field of type `RootCredentialConfig`. This is **Pattern B** from the template (nested credential object).

The `azureCredentials` field resolves the Azure Client ID and Client Secret used to authenticate with Azure Active Directory.

**Important behaviors:**
- Default `usernameKey`: `"clientid"` (maps to the Azure Client ID)
- Default `passwordKey`: `"clientsecret"` (maps to the Azure Client Secret)
- If `spec.clientID` is set directly, it takes precedence over the username from the referenced secret
- If `azureCredentials` is nil/empty AND matches the zero-value check `{PasswordKey: "clientsecret", UsernameKey: "clientid"}`, credential resolution is skipped (Azure environment variables will be used)

From `api/v1alpha1/utils/commons.go`, `RootCredentialConfig` has:
- `secret` — Kubernetes Secret reference
- `vaultSecret` — Vault secret path reference
- `randomSecret` — RandomSecret reference
- `usernameKey` — key for client ID (default: `"clientid"`)
- `passwordKey` — key for client secret (default: `"clientsecret"`)

### Relative Link Conventions

From `docs/auth-engines/gcp.md` and `docs/auth-engines/azure.md`:
- To shared docs: `../auth-section.md`, `../contributing-vault-apis.md`, `../secret-management.md`
- To other engine files: `cert.md`, `kubernetes.md`, `ldap.md`, `jwt-oidc.md` (same directory)
- External: full URLs to Vault documentation

### Previous Story Intelligence

**From D2.4 (JWT/OIDC Auth Engine Docs):**
- Documented dual-mode behavior (JWT vs OIDC) with decision table — GCP similarly has dual role types (IAM vs GCE); use the same approach
- Organized fields into logical groups with separate tables — follow same approach for GCP IAM/GCE split
- Documented Pattern B credential resolution (OIDCCredentials) — both GCP and Azure use the same pattern
- Used `cert.md` as structural reference — continue this pattern

**From D2.3 (LDAP Auth Engine Docs):**
- Established the credential resolution documentation pattern (Pattern A for LDAP with `BindCredentials`)
- GCP uses Pattern B (same as JWT/OIDC, Azure) — document as nested object

**From D2.2 (Kubernetes Auth Engine Docs):**
- Established the basic template pattern: Overview → Config → Role → See Also
- No credential resolution section for Kubernetes — both GCP and Azure DO need one

**From D2.1 (Directory Structure & Index Page):**
- Created `docs/auth-engines/` directory (currently only contains `cert.md`)
- AuthEngineMount section is in the index page (not per-engine files)

**From D1.3 (Link and Naming Fixes):**
- Fixed snake_case→camelCase in GCP and Azure sections specifically
- Verify the source content for any remaining inconsistencies

**From D1 Retrospective:**
- Documentation stories expect 3+ review findings — this is normal
- Opus 4.6 recommended for all stories

### Project Structure Notes

```
docs/
├── auth-engines/
│   ├── index.md          ← D2.1 (may not exist yet)
│   ├── cert.md           ← EXISTS (D1.2) — reference implementation
│   ├── kubernetes.md     ← D2.2 (may not exist yet)
│   ├── ldap.md           ← D2.3 (may not exist yet)
│   ├── jwt-oidc.md       ← D2.4 (may not exist yet)
│   ├── gcp.md            ← NEW (this story)
│   └── azure.md          ← NEW (this story)
├── auth-engines.md       ← original/redirect pointer
├── auth-section.md       ← shared auth config docs (unchanged)
├── engine-doc-template.md ← template (D1.1, review-patched)
├── secret-management.md  ← link target for RandomSecret reference
└── ...
```

### References

- [Source: _bmad-output/planning-artifacts/epics.md#Story D2.5] — Story requirements and acceptance criteria
- [Source: docs/auth-engines.md:425-588] — GCP auth content to extract and standardize
- [Source: docs/auth-engines.md:590-775] — Azure auth content to extract and standardize
- [Source: docs/auth-engines/cert.md] — Reference implementation for template pattern
- [Source: docs/engine-doc-template.md] — Template structure (D1.1, review-patched)
- [Source: api/v1alpha1/gcpauthengineconfig_types.go] — CRD field definitions for GCP Config (GCPConfig struct)
- [Source: api/v1alpha1/gcpauthenginerole_types.go] — CRD field definitions for GCP Role (GCPRole struct)
- [Source: api/v1alpha1/azureauthengineconfig_types.go] — CRD field definitions for Azure Config (AzureConfig struct)
- [Source: api/v1alpha1/azureauthenginerole_types.go] — CRD field definitions for Azure Role (AzureRole struct)
- [Source: api/v1alpha1/utils/commons.go] — RootCredentialConfig struct definition
- [Source: _bmad-output/implementation-artifacts/d2-4-standardize-jwt-oidc-auth-engine-docs.md] — Previous story context
- [Source: _bmad-output/project-context.md] — Project conventions and coding standards

## Dev Agent Record

### Agent Model Used

{{agent_model_name_version}}

### Debug Log References

### Completion Notes List

### File List
