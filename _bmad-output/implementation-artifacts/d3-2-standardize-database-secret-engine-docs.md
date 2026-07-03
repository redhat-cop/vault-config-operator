# Story D3.2: Standardize Database Secret Engine Docs

Status: ready-for-dev

## Story

As a user configuring database dynamic secrets,
I want comprehensive docs covering config, role, and static role with credential resolution,
So that I can set up the most complex secret engine correctly.

## Acceptance Criteria

1. **Given** the existing Database content (Config, Role, StaticRole) in `docs/secret-engines.md` lines 52-179 **When** extracted to `docs/secret-engines/database.md` and standardized per the template **Then** it contains:
   - DatabaseSecretEngineConfig: complete YAML with `rootPasswordRotation` example, credential resolution (3 methods), `passwordAuthentication` field
   - DatabaseSecretEngineRole: complete YAML with `creationStatements`, Vault CLI equivalent
   - DatabaseSecretEngineStaticRole: complete YAML with `rotationStatements`, credential types
   - All three have Vault CLI equivalents

2. **Given** the new `database.md` file **When** validated against the template structure **Then** it follows the same structure as `docs/auth-engines/kubernetes.md` (Overview → Config CRD → Role CRD → Credential Resolution → See Also)

3. **Given** cross-references in `readme.md` lines 85-86 pointing to `secret-engines.md#DatabaseSecretEngineConfig` and `secret-engines.md#DatabaseSecretEngineRole` **When** the content is moved **Then** those links are updated to point to `secret-engines/database.md#databasesecretengineconfig` and `secret-engines/database.md#databasesecretenginerole`

## Tasks / Subtasks

- [ ] Task 1: Create `docs/secret-engines/database.md` (AC: 1, 2)
  - [ ] 1.1: Write Overview section — 2-3 sentences explaining Database secret engine, link to Vault docs, list the three CRDs (Config, Role, StaticRole)
  - [ ] 1.2: Write DatabaseSecretEngineConfig section with Example YAML (include `rootPasswordRotation`, `passwordAuthentication`, `rootCredentialsFromSecret`), Vault CLI Equivalent, and Field Descriptions table
  - [ ] 1.3: Write DatabaseSecretEngineRole section with Example YAML (include `creationStatements`), Vault CLI Equivalent, and Field Descriptions table
  - [ ] 1.4: Write DatabaseSecretEngineStaticRole section with Example YAML (include `rotationStatements`, `credentialType`, `passwordCredentialConfig`), Vault CLI Equivalent, and Field Descriptions table
  - [ ] 1.5: Write Credential Resolution section using Pattern A (`rootCredentialsFromSecret`, `rootCredentialsFromVaultSecret`, `rootCredentialsFromRandomSecret`)
  - [ ] 1.6: Add "See Also" section with links to `../auth-section.md`, `../contributing-vault-apis.md`, and Vault docs

- [ ] Task 2: Audit field names for camelCase consistency (AC: 1)
  - [ ] 2.1: Cross-reference all field names in the new doc against the Go CRD types (`databasesecretengineconfig_types.go`, `databasesecretenginerole_types.go`, `databasesecretenginestaticrole_types.go`) — field names in the doc MUST match the `json:` tag values exactly
  - [ ] 2.2: Fix any snake_case field names from the original `secret-engines.md` source

- [ ] Task 3: Update `readme.md` cross-references (AC: 3)
  - [ ] 3.1: Update line 85 from `./docs/secret-engines.md#DatabaseSecretEngineConfig` to `./docs/secret-engines/database.md#databasesecretengineconfig`
  - [ ] 3.2: Update line 86 from `./docs/secret-engines.md#DatabaseSecretEngineRole` to `./docs/secret-engines/database.md#databasesecretenginerole`

- [ ] Task 4: Verify links and structure (AC: 2)
  - [ ] 4.1: Verify relative links resolve correctly from `docs/secret-engines/database.md` (`../auth-section.md`, `../contributing-vault-apis.md`)
  - [ ] 4.2: Verify structure matches `kubernetes.md` pattern: heading hierarchy, section ordering, table format

## Dev Notes

### This Is a Documentation-Only Story

No Go code changes. No tests to run. No `make manifests generate`. The deliverables are:
- 1 new file: `docs/secret-engines/database.md`
- 1 modified file: `readme.md` (update 2 cross-reference links)

### Dependency on D3.1

This story assumes D3.1 has been completed (creating `docs/secret-engines/index.md` and the `docs/secret-engines/` directory). If D3.1 is NOT yet done, this story can still proceed — create the directory if it doesn't exist. The index will reference `database.md` via `[database.md](database.md)`.

### Source Content Location

The content to extract and standardize is in `docs/secret-engines.md` lines 52-179:
- `## DatabaseSecretEngineConfig` (lines 52-107)
- `## DatabaseSecretEngineRole` (lines 109-140)
- `## DatabaseSecretEngineStaticRole` (lines 142-179)

### Template to Follow

Use `docs/engine-doc-template.md` as the structural pattern. Use `docs/auth-engines/kubernetes.md` as the concrete reference implementation (most recently completed per-engine doc with same structure).

Key structural requirements from the template:
1. Title: `# Database Secret Engine`
2. Link to Vault docs immediately below title
3. `## Overview` — 2-3 sentences + CRD list (THREE CRDs for this engine)
4. `## DatabaseSecretEngineConfig` → `### Example` → `### Vault CLI Equivalent` → `### Field Descriptions`
5. `## DatabaseSecretEngineRole` → `### Example` → `### Vault CLI Equivalent` → `### Field Descriptions`
6. `## DatabaseSecretEngineStaticRole` → `### Example` → `### Vault CLI Equivalent` → `### Field Descriptions`
7. `## Credential Resolution` (Pattern A — flat prefix fields)
8. `## See Also`

### DatabaseSecretEngineConfig — Complete Field Reference

From `api/v1alpha1/databasesecretengineconfig_types.go`, the `DBSEConfig` struct has these fields:

| CRD Field (camelCase) | JSON tag | Vault API key (snake_case) | Type | Required | Default |
|---|---|---|---|---|---|
| pluginName | `json:"pluginName,omitempty"` | `plugin_name` | string | Yes | — |
| pluginVersion | `json:"pluginVersion,omitempty"` | `plugin_version` | string | No | — |
| verifyConnection | `json:"verifyConnection,omitempty"` | `verify_connection` | bool | No | true |
| allowedRoles | `json:"allowedRoles"` | `allowed_roles` | []string | No | `["*"]` |
| rootRotationStatements | `json:"rootRotationStatements,omitempty"` | `root_credentials_rotate_statements` | []string | No | — |
| passwordAuthentication | `json:"passwordAuthentication"` | `password_authentication` | string | No | `"password"` (Enum: `password`, `scram-sha-256`) |
| passwordPolicy | `json:"passwordPolicy,omitempty"` | `password_policy` | string | No | — |
| connectionURL | `json:"connectionURL,omitempty"` | `connection_url` | string | Yes | — |
| username | `json:"username,omitempty"` | `username` | string | No | — (retrieved from credentials if not set) |
| disableEscaping | `json:"disableEscaping,omitempty"` | `disable_escaping` | bool | No | false |
| databaseSpecificConfig | `json:"databaseSpecificConfig,omitempty"` | (keys added directly to payload) | map[string]string | No | — |
| rootPasswordRotation | `json:"rootPasswordRotation,omitempty"` | — (operator-side) | object | No | — |
| rootPasswordRotation.enable | `json:"enable,omitempty"` | — | bool | No | false |
| rootPasswordRotation.rotationPeriod | `json:"rotationPeriod,omitempty"` | — | duration | No | — |

Additional top-level spec fields:
- `path` (Required) — path of the secret engine mount. Final Vault path: `[namespace/]{path}/config/{name}`
- `authentication` (Required) — see shared auth-section.md
- `connection` (Optional) — override Vault connection
- `rootCredentials` (Required) — credential resolution. See Credential Resolution section
- `name` (Optional) — override Vault object name (defaults to `metadata.name`)

**Important:** `rootPasswordRotation` is an operator-side feature that triggers Vault's `rotate-root` endpoint. When `enable: true`, the root password is rotated immediately on first reconcile. When `rotationPeriod` is also set, periodic rotation is scheduled. There is NO way to recover the root password after rotation.

**Important:** `databaseSpecificConfig` is a freeform map for passing plugin-specific connection parameters (e.g., `tls_ca`, `tls_certificate_key` for MongoDB). Each key-value pair is added directly to the Vault write payload.

### DatabaseSecretEngineRole — Complete Field Reference

From `api/v1alpha1/databasesecretenginerole_types.go`, the `DBSERole` struct has these fields:

| CRD Field (camelCase) | JSON tag | Vault API key (snake_case) | Type | Required | Default |
|---|---|---|---|---|---|
| dBName | `json:"dBName,omitempty"` | `db_name` | string | Yes | — |
| defaultTTL | `json:"defaultTTL,omitempty"` | `default_ttl` | duration | No | system/engine default |
| maxTTL | `json:"maxTTL,omitempty"` | `max_ttl` | duration | No | system/mount default |
| creationStatements | `json:"creationStatements,omitempty"` | `creation_statements` | []string | No | — |
| revocationStatements | `json:"revocationStatements,omitempty"` | `revocation_statements` | []string | No | — |
| rollbackStatements | `json:"rollbackStatements,omitempty"` | `rollback_statements` | []string | No | — |
| renewStatements | `json:"renewStatements,omitempty"` | `renew_statements` | []string | No | — |

Additional top-level spec fields:
- `path` (Required) — path of the secret engine mount. Final Vault path: `[namespace/]{path}/roles/{name}`
- `authentication` (Required) — see shared auth-section.md
- `connection` (Optional) — override Vault connection
- `name` (Optional) — override Vault object name

### DatabaseSecretEngineStaticRole — Complete Field Reference

From `api/v1alpha1/databasesecretenginestaticrole_types.go`, the `DBSEStaticRole` struct has these fields:

| CRD Field (camelCase) | JSON tag | Vault API key (snake_case) | Type | Required | Default |
|---|---|---|---|---|---|
| dBName | `json:"dBName,omitempty"` | `db_name` | string | Yes | — |
| username | `json:"username,omitempty"` | `username` | string | Yes | — |
| rotationPeriod | `json:"rotationPeriod,omitempty"` | `rotation_period` | int (seconds) | Yes | — (minimum: 5) |
| rotationStatements | `json:"rotationStatements,omitempty"` | `rotation_statements` | []string | No | — |
| credentialType | `json:"credentialType,omitempty"` | `credential_type` | string | Yes | — (Enum: `password`, `rsa_private_key`) |
| passwordCredentialConfig | `json:"passwordCredentialConfig,omitempty"` | `credential_config` | object | No | — |
| passwordCredentialConfig.passwordPolicy | `json:"passwordPolicy,omitempty"` | `password_policy` | string | No | — |
| rsaPrivateKeyCredentialConfig | `json:"rsaPrivateKeyCredentialConfig,omitempty"` | `credential_config` | object | No | — |
| rsaPrivateKeyCredentialConfig.keyBits | `json:"keyBits,omitempty"` | `key_bits` | int | No | — (Enum: 2048, 3072, 4096) |
| rsaPrivateKeyCredentialConfig.format | `json:"format,omitempty"` | `format` | string | No | — (Enum: `pkcs8`) |

Additional top-level spec fields:
- `path` (Required) — path of the secret engine mount. Final Vault path: `[namespace/]{path}/static-roles/{name}`
- `authentication` (Required) — see shared auth-section.md
- `connection` (Optional) — override Vault connection
- `name` (Optional) — override Vault object name

**Important Validation Rule:** Exactly ONE of `passwordCredentialConfig` or `rsaPrivateKeyCredentialConfig` must be specified. The webhook rejects CRs with zero or both.

### Credential Resolution (Pattern A)

The Database secret engine uses Pattern A (flat prefix fields) for credential resolution. The credential field prefix is `rootCredentials`. Three methods:

1. `rootCredentials.secret.name` — reference a Kubernetes basic-auth Secret
2. `rootCredentials.vaultSecret.path` — reference a Vault KV secret (with `usernameKey`/`passwordKey`)
3. `rootCredentials.randomSecret.name` — reference a RandomSecret CR

**Important:** If `spec.username` is provided in the CRD, it takes precedence over the username from the credential source. If not provided, the username is retrieved from the credential source along with the password.

### Known Issues in Source Content

From the existing `docs/secret-engines.md`:
1. Line 85: typo "teh" → "the" in username field description
2. Line 93: "retrived" → "retrieved" typo
3. Field descriptions are informal prose (not tables) — convert to table format per template
4. No `pluginVersion`, `verifyConnection`, `passwordPolicy`, `disableEscaping`, or `databaseSpecificConfig` documented in existing content — add from CRD types
5. StaticRole section missing `rsaPrivateKeyCredentialConfig` documentation — add from CRD types
6. `passwordAuthentication` field mentioned in prose but not in YAML example — include in example

### readme.md Cross-References

Lines requiring update:

| Line | Current Link | New Link |
|------|-------------|----------|
| 85 | `./docs/secret-engines.md#DatabaseSecretEngineConfig` | `./docs/secret-engines/database.md#databasesecretengineconfig` |
| 86 | `./docs/secret-engines.md#DatabaseSecretEngineRole` | `./docs/secret-engines/database.md#databasesecretenginerole` |

Note: `DatabaseSecretEngineStaticRole` is NOT referenced in `readme.md` — no update needed for it.

### Relative Link Conventions

From `docs/secret-engines/database.md`:
- To shared docs: `../auth-section.md`, `../contributing-vault-apis.md`
- To secret management (for RandomSecret reference): `../secret-management.md#randomsecret`
- To other engine files: `pki.md`, `rabbitmq.md` (same directory)
- External: full URLs to Vault documentation

### Three CRDs — Unique for Secret Engines

Unlike auth engines (which have Config + Role), the Database secret engine has THREE CRDs:
- **DatabaseSecretEngineConfig** — connection/configuration for a database plugin
- **DatabaseSecretEngineRole** — dynamic credential generation role
- **DatabaseSecretEngineStaticRole** — password rotation for a pre-existing database user

The template only shows two sections (Config + Role). Extend it by adding a third section (`## DatabaseSecretEngineStaticRole`) after the Role section, following the same internal structure (Example → Vault CLI Equivalent → Field Descriptions).

### Vault Path Structure

Understanding the path hierarchy helps write accurate CLI equivalents:

```
{mount-path}/config/{connection-name}        ← DatabaseSecretEngineConfig
{mount-path}/roles/{role-name}               ← DatabaseSecretEngineRole
{mount-path}/static-roles/{static-role-name} ← DatabaseSecretEngineStaticRole
{mount-path}/rotate-root/{connection-name}   ← Root password rotation (operator-triggered)
```

Example with `path: postgresql-vault-demo/database`:
- Config: `postgresql-vault-demo/database/config/my-postgresql-database`
- Role: `postgresql-vault-demo/database/roles/read-only`
- StaticRole: `postgresql-vault-demo/database/static-roles/read-only-static`

### Previous Story Intelligence

**From D3.1 (Secret-Engines Directory Structure & Index — the direct predecessor):**
- Created `docs/secret-engines/index.md` with overview, SecretEngineMount section, engine table, common config, see also
- Replaced `docs/secret-engines.md` with redirect pointer
- Index page engine table has `database.md` link ready for this story
- Documented all readme.md cross-references for downstream stories (D3.2-D3.4)
- D3.1 was NOT yet implemented at time of story creation — check if directory exists before creating file

**From D2.2 (Kubernetes Auth Engine Docs — exact analogous pattern):**
- Established the extraction pattern: source content → template structure → field audit → CLI equivalents
- Field descriptions table uses camelCase names from JSON tags
- Vault CLI equivalents use snake_case names
- Review findings focused on: path confusion (`spec.authentication.path` vs `spec.path`), missing mutual exclusivity docs, incomplete field coverage
- Zero review findings on D2.5 — team fully internalized template by end of D2

**From D2.5 (GCP and Azure Auth Engine Docs — Pattern B credential resolution):**
- Credential resolution documentation patterns validated (Pattern A and Pattern B both work)
- Database uses Pattern A (flat prefix) — same as LDAP auth engine
- Zero review findings — pattern fully proven

**From D2 Retrospective:**
- D3 readiness: No preparation work needed, all preconditions satisfied
- Template proven across 6 auth engine docs — applies directly to secret engines
- Secret engines have more CRD variety (Config + Role + StaticRole) — this story establishes the three-CRD pattern for D3.3-D3.4
- Potential friction: more credential resolution patterns — Database uses standard Pattern A
- Recommendation: Continue using Opus 4.6 for all stories

### Project Structure Notes

```
docs/
├── secret-engines/
│   ├── index.md          ← D3.1
│   └── database.md       ← NEW (this story)
├── secret-engines.md     ← redirect pointer (D3.1)
├── auth-engines/
│   ├── index.md          ← D2.1
│   ├── kubernetes.md     ← D2.2 — reference implementation
│   └── ...
├── auth-section.md       ← shared auth config (unchanged)
├── engine-doc-template.md ← template (D1.1, review-patched)
└── ...
```

### References

- [Source: _bmad-output/planning-artifacts/epics.md#Story D3.2] — Story requirements and acceptance criteria
- [Source: docs/secret-engines.md:52-179] — Database content to extract and standardize
- [Source: docs/auth-engines/kubernetes.md] — Reference implementation for template pattern (D2.2)
- [Source: docs/engine-doc-template.md] — Template structure (D1.1, review-patched 4 times)
- [Source: api/v1alpha1/databasesecretengineconfig_types.go] — CRD field definitions for Config
- [Source: api/v1alpha1/databasesecretenginerole_types.go] — CRD field definitions for Role
- [Source: api/v1alpha1/databasesecretenginestaticrole_types.go] — CRD field definitions for StaticRole
- [Source: _bmad-output/implementation-artifacts/d3-1-create-secret-engines-directory-structure-and-index-page.md] — Previous story context
- [Source: _bmad-output/implementation-artifacts/d2-2-standardize-kubernetes-auth-engine-docs.md] — Analogous D2 story (reference)
- [Source: _bmad-output/implementation-artifacts/epic-d2-retro-2026-07-02.md] — D2 retro: D3 readiness assessment
- [Source: readme.md:85-86] — Cross-references that need updating
- [Source: _bmad-output/project-context.md] — Project conventions and coding standards

## Dev Agent Record

### Agent Model Used

{{agent_model_name_version}}

### Debug Log References

### Completion Notes List

### File List
