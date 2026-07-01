# Story D2.3: Standardize LDAP Auth Engine Docs

Status: done

## Story

As a user configuring LDAP authentication,
I want well-structured LDAP auth docs that are comprehensive but not overwhelming,
So that I can configure LDAP auth without drowning in field descriptions.

## Acceptance Criteria

1. **Given** the existing LDAPAuthEngine content in `docs/auth-engines.md` (lines 114-254) **When** it is extracted to `docs/auth-engines/ldap.md` and standardized per the template **Then** it contains:
   - Overview linking to Vault LDAP auth docs
   - LDAPAuthEngineConfig: complete YAML example, field descriptions (camelCase), credential resolution section, TLS configuration section, Vault CLI equivalent
   - LDAPAuthEngineGroup: complete YAML example, field descriptions, Vault CLI equivalent

2. **Given** the new `ldap.md` file **When** validated against the template structure **Then** it follows the same structure as `docs/auth-engines/cert.md` (Overview → Config CRD → Group CRD → Credential Resolution → See Also)

3. **Given** `docs/auth-engines.md` (the redirect pointer, post-D2.1) **When** the LDAP content is moved **Then** no LDAP-specific content remains in `auth-engines.md` (it should already be a redirect after D2.1)

4. **Given** the field descriptions **When** validated against the Go type definitions **Then** ALL field names use camelCase (matching `json:` tags exactly) with no residual snake_case

5. **Given** the credential resolution section **When** reviewed **Then** all three methods (Kubernetes Secret, Vault Secret, RandomSecret) are clearly documented with YAML examples

6. **Given** the TLS configuration **When** reviewed **Then** it is clearly separated and documents both inline cert fields and the `tLSConfig` secret-based approach

## Tasks / Subtasks

- [x] Task 1: Create `docs/auth-engines/ldap.md` (AC: 1, 2)
  - [x] 1.1: Write Overview section — 2-3 sentences explaining LDAP auth, link to Vault docs, list the two CRDs (LDAPAuthEngineConfig, LDAPAuthEngineGroup)
  - [x] 1.2: Write LDAPAuthEngineConfig section with Example YAML, Vault CLI Equivalent, and Field Descriptions table
  - [x] 1.3: Organize LDAPAuthEngineConfig fields into logical groups: Connection & TLS, User Search, Group Search, Token Parameters
  - [x] 1.4: Write TLS Configuration subsection documenting both inline fields (`certificate`, `clientTLSCert`, `clientTLSKey`) and the `tLSConfig` secret-based approach
  - [x] 1.5: Write LDAPAuthEngineGroup section with Example YAML, Vault CLI Equivalent, and Field Descriptions table
  - [x] 1.6: Write Credential Resolution section documenting all three `bindCredentials` methods with YAML examples
  - [x] 1.7: Add "See Also" section with links to auth-section.md, contributing-vault-apis.md, and Vault docs

- [x] Task 2: Audit field names for camelCase consistency (AC: 4)
  - [x] 2.1: Cross-reference all field names in the new doc against `ldapauthengineconfig_types.go` and `ldapauthenginegroup_types.go` — field names MUST match the `json:` tag values exactly
  - [x] 2.2: Fix any residual snake_case field names from the original `auth-engines.md` source (D1.3 did NOT audit LDAP section — confirmed in D1 retro)
  - [x] 2.3: Verify fields with non-standard casing are correct: `TLSMinVersion`, `TLSMaxVersion`, `UPNDomain` (these use uppercase per Go convention)

- [x] Task 3: Verify links and structure (AC: 2)
  - [x] 3.1: Verify relative links resolve correctly from `docs/auth-engines/ldap.md` (`../auth-section.md`, `../contributing-vault-apis.md`)
  - [x] 3.2: Verify structure matches `cert.md` pattern: heading hierarchy, section ordering, table format

### Review Findings

- [x] [Review][Patch] Clarify `spec.authentication.path` vs `spec.path` in the LDAP examples [`docs/auth-engines/ldap.md:20-49`]
- [x] [Review][Patch] Explain that the Vault CLI example shows the resolved TLS fields for `tLSConfig`, and document the exact Secret keys required by `tLSConfig.tlsSecret` [`docs/auth-engines/ldap.md:51-71`]
- [x] [Review][Patch] Document that exactly one of `bindCredentials.secret`, `bindCredentials.vaultSecret`, or `bindCredentials.randomSecret` may be set [`docs/auth-engines/ldap.md:217-255`]
- [x] [Review][Patch] Correct the documented default for `groupAttr` to `cn` [`docs/auth-engines/ldap.md:116-121`]

## Dev Notes

### This Is a Documentation-Only Story

No Go code changes. No tests to run. No `make manifests generate`. The deliverables are:
- 1 new file: `docs/auth-engines/ldap.md`

### Dependency on D2.1

This story assumes D2.1 has been completed (creating `docs/auth-engines/index.md` and the redirect pointer in `docs/auth-engines.md`). If D2.1 is NOT yet done, this story can still proceed — the `ldap.md` file can be created independently. The index will reference it via `[ldap.md](ldap.md)`.

### Source Content Location

The content to extract and standardize is in `docs/auth-engines.md` lines 114-254:
- `## LDAPAuthEngineConfig` (lines 114-228)
- `## LDAPAuthEngineGroup` (lines 230-254)

### Template to Follow

Use `docs/engine-doc-template.md` as the structural pattern. Use `docs/auth-engines/cert.md` as the concrete reference implementation.

Key structural requirements from the template:
1. Title: `# LDAP Auth Engine`
2. Link to Vault docs immediately below title
3. `## Overview` — 2-3 sentences + CRD list
4. `## LDAPAuthEngineConfig` → `### Example` → `### Vault CLI Equivalent` → `### Field Descriptions` → `### TLS Configuration`
5. `## LDAPAuthEngineGroup` → `### Example` → `### Vault CLI Equivalent` → `### Field Descriptions`
6. `## Credential Resolution` — documenting the three `bindCredentials` methods (this engine DOES use credentials, unlike Kubernetes)
7. `## See Also`

### LDAPAuthEngineConfig — Complete Field Reference

From `api/v1alpha1/ldapauthengineconfig_types.go`, the `LDAPConfig` struct has these fields:

**Connection & TLS Fields:**

| CRD Field (camelCase) | JSON tag | Vault API key (snake_case) | Type | Required | Default |
|---|---|---|---|---|---|
| url | `json:"url"` | `url` | string | Yes | `ldap://127.0.0.1` |
| startTLS | `json:"startTLS,omitempty"` | `starttls` | bool | No | false |
| TLSMinVersion | `json:"TLSMinVersion"` | `tls_min_version` | string | No | `tls12` |
| TLSMaxVersion | `json:"TLSMaxVersion"` | `tls_max_version` | string | No | `tls12` |
| insecureTLS | `json:"insecureTLS,omitempty"` | `insecure_tls` | bool | No | false |
| certificate | `json:"certificate,omitempty"` | `certificate` | string | No | — |
| clientTLSCert | `json:"clientTLSCert,omitempty"` | `client_tls_cert` | string | No | — |
| clientTLSKey | `json:"clientTLSKey,omitempty"` | `client_tls_key` | string | No | — |
| requestTimeout | `json:"requestTimeout"` | `request_timeout` | string | No | `90s` |

**User Search Fields:**

| CRD Field (camelCase) | JSON tag | Vault API key (snake_case) | Type | Required | Default |
|---|---|---|---|---|---|
| bindDN | `json:"bindDN,omitempty"` | `binddn` | string | No | — |
| userDN | `json:"userDN,omitempty"` | `userdn` | string | No | — |
| userAttr | `json:"userAttr"` | `userattr` | string | No | `cn` |
| userFilter | `json:"userFilter,omitempty"` | `userfilter` | string | No | — |
| discoverDN | `json:"discoverDN,omitempty"` | `discoverdn` | bool | No | false |
| denyNullBind | `json:"denyNullBind"` | `deny_null_bind` | bool | No | true |
| UPNDomain | `json:"UPNDomain,omitempty"` | `upndomain` | string | No | — |
| caseSensitiveNames | `json:"caseSensitiveNames,omitempty"` | `case_sensitive_names` | bool | No | false |
| usernameAsAlias | `json:"usernameAsAlias,omitempty"` | `username_as_alias` | bool | No | false |

**Group Search Fields:**

| CRD Field (camelCase) | JSON tag | Vault API key (snake_case) | Type | Required | Default |
|---|---|---|---|---|---|
| groupDN | `json:"groupDN,omitempty"` | `groupdn` | string | No | — |
| groupFilter | `json:"groupFilter,omitempty"` | `groupfilter` | string | No | — |
| groupAttr | `json:"groupAttr,omitempty"` | `groupattr` | string | No | — |
| anonymousGroupSearch | `json:"anonymousGroupSearch,omitempty"` | `anonymous_group_search` | bool | No | false |

**Token Parameters:**

| CRD Field (camelCase) | JSON tag | Vault API key (snake_case) | Type | Required | Default |
|---|---|---|---|---|---|
| tokenTTL | `json:"tokenTTL,omitempty"` | `token_ttl` | string | No | — |
| tokenMaxTTL | `json:"tokenMaxTTL,omitempty"` | `token_max_ttl` | string | No | — |
| tokenPolicies | `json:"tokenPolicies,omitempty"` | `token_policies` | string | No | — |
| tokenBoundCIDRs | `json:"tokenBoundCIDRs,omitempty"` | `token_bound_cidrs` | string | No | — |
| tokenExplicitMaxTTL | `json:"tokenExplicitMaxTTL,omitempty"` | `token_explicit_max_ttl` | string | No | — |
| tokenNoDefaultPolicy | `json:"tokenNoDefaultPolicy,omitempty"` | `token_no_default_policy` | bool | No | false |
| tokenNumUses | `json:"tokenNumUses,omitempty"` | `token_num_uses` | int64 | No | 0 |
| tokenPeriod | `json:"tokenPeriod,omitempty"` | `token_period` | int64 | No | 0 |
| tokenType | `json:"tokenType,omitempty"` | `token_type` | string | No | — |

Additional top-level spec fields (NOT in `LDAPConfig` inline struct):
- `path` (Required) — mount path for the LDAP auth engine
- `authentication` (Required) — see shared auth-section.md
- `connection` (Optional) — override Vault connection
- `bindCredentials` (Required) — credential resolution config (see below)
- `tLSConfig` (Optional) — TLS certificate via Kubernetes Secret

### LDAPAuthEngineGroup — Complete Field Reference

From `api/v1alpha1/ldapauthenginegroup_types.go`:

| CRD Field (camelCase) | JSON tag | Vault API key (snake_case) | Type | Required | Default |
|---|---|---|---|---|---|
| name | `json:"name,omitempty"` | `name` | string | Yes | — |
| policies | `json:"policies,omitempty"` | `policies` | string | No | — |

Additional top-level spec fields:
- `path` (Required) — LDAP auth mount path
- `authentication` (Required) — see shared auth-section.md
- `connection` (Optional) — override Vault connection

**Vault path:** `auth/{spec.path}/groups/{spec.name}`

### Credential Resolution (bindCredentials)

The LDAP engine uses `bindCredentials` field of type `RootCredentialConfig`. This is **Pattern A** from the template (flat credential config — not a nested object).

The `bindCredentials` field resolves the `bindDN` (username) and `bindPass` (password) used to connect to the LDAP server.

From `api/v1alpha1/utils/commons.go`, `RootCredentialConfig` has:
- `secret` — Kubernetes Secret reference (basic auth type)
- `vaultSecret` — Vault secret path reference
- `randomSecret` — RandomSecret reference
- `usernameKey` — key for username (default: `"username"`)
- `passwordKey` — key for password (default: `"password"`)

**Important behavior:** If `bindDN` is set in the spec, it takes precedence over the username from the referenced secret. The password is always retrieved from the secret source.

### TLS Configuration

The LDAP engine supports two TLS configuration methods:

1. **Inline fields** (in the `LDAPConfig` struct): `certificate`, `clientTLSCert`, `clientTLSKey` — set directly in the CR spec
2. **`tLSConfig` field** — references a Kubernetes TLS Secret (`ca.crt`, `tls.crt`, `tls.key` keys)

The `tLSConfig` approach is the Kubernetes-native way:
```yaml
spec:
  tLSConfig:
    tlsSecret:
      name: ldap-tls-certificate
```

When `tLSConfig.tlsSecret` is set, the operator reads `ca.crt` → `certificate`, `tls.crt` → `clientTLSCert`, `tls.key` → `clientTLSKey` from the referenced Secret.

### IsDeletable Behavior

`LDAPAuthEngineConfig` returns `IsDeletable() == false` — deleting the CR does NOT remove the LDAP config from Vault. The auth mount must be disabled separately.

`LDAPAuthEngineGroup` returns `IsDeletable() == true` — deleting the CR removes the group from Vault.

### Known Issues in Source Content

From D1 retrospective (section "Potential Friction Points"):
> Kubernetes and LDAP sections were NOT explicitly audited for snake_case field names in D1.3 — D2.2 and D2.3 will handle during extraction

Action: When extracting from `auth-engines.md`, carefully verify ALL field names use camelCase (matching JSON tags). The original source likely contains mixed usage. In the field descriptions table, always use camelCase. Vault API names belong only in the "Vault CLI Equivalent" section.

Specific casing notes from the Go types:
- `TLSMinVersion` and `TLSMaxVersion` — uppercase TLS prefix (json tags: `"TLSMinVersion"`, `"TLSMaxVersion"`)
- `UPNDomain` — uppercase UPN prefix (json tag: `"UPNDomain"`)
- `bindDN`, `userDN`, `groupDN`, `discoverDN` — uppercase DN suffix
- `tLSConfig` — lowercase t, uppercase LS (json tag: `"tLSConfig"`)

### Relative Link Conventions

From `docs/auth-engines/ldap.md`:
- To shared docs: `../auth-section.md`, `../contributing-vault-apis.md`, `../secret-management.md`
- To other engine files: `cert.md`, `kubernetes.md` (same directory)
- External: full URLs to Vault documentation

### Previous Story Intelligence

**From D2.2 (Kubernetes Auth Engine Docs):**
- Established the pattern for this epic: extract from `auth-engines.md`, standardize per template, verify camelCase
- No credential resolution section needed for Kubernetes — LDAP DOES need one
- Used `cert.md` as structural reference
- D2.1 created the directory structure and index page

**From D2.1 (Directory Structure & Index Page):**
- Created `docs/auth-engines/index.md` with engine table linking to `ldap.md`
- Replaced `docs/auth-engines.md` with redirect pointer
- AuthEngineMount section is in the index page (not per-engine files)

**From D1.1 (Template Creation):**
- Template was patched 4 times in review — always use current version at `docs/engine-doc-template.md`
- DNFR1-DNFR5 requirements define documentation standards

**From D1.2 (CertAuth Documentation):**
- First per-engine file at `docs/auth-engines/cert.md` — use as reference implementation
- Validates the template pattern works; established relative link patterns
- Cert auth uses inline certificate field directly — LDAP uses `bindCredentials` + optional `tLSConfig`

**From D1.3 (Link and Naming Fixes):**
- Fixed snake_case→camelCase in GCP and Azure sections
- LDAP section was explicitly NOT in scope — residual snake_case is EXPECTED
- Fixed leading-space code fences and broken cross-references

**From D1 Retrospective:**
- Documentation stories expect 3+ review findings — this is normal
- Opus 4.6 recommended for all stories
- D2 assessed as ready — no preparation needed

### Project Structure Notes

```
docs/
├── auth-engines/
│   ├── index.md          ← D2.1
│   ├── cert.md           ← EXISTS (D1.2) — reference implementation
│   ├── kubernetes.md     ← D2.2
│   └── ldap.md           ← NEW (this story)
├── auth-engines.md       ← redirect pointer (D2.1)
├── auth-section.md       ← shared auth config docs (unchanged)
├── engine-doc-template.md ← template (D1.1, review-patched)
├── secret-management.md  ← link target for RandomSecret reference
└── ...
```

### References

- [Source: _bmad-output/planning-artifacts/epics.md#Story D2.3] — Story requirements and acceptance criteria
- [Source: docs/auth-engines.md:114-254] — LDAP auth content to extract and standardize
- [Source: docs/auth-engines/cert.md] — Reference implementation for template pattern
- [Source: docs/engine-doc-template.md] — Template structure (D1.1, review-patched 4 times)
- [Source: api/v1alpha1/ldapauthengineconfig_types.go] — CRD field definitions for Config (LDAPConfig struct)
- [Source: api/v1alpha1/ldapauthenginegroup_types.go] — CRD field definitions for Group
- [Source: api/v1alpha1/utils/commons.go:366-396] — RootCredentialConfig struct definition
- [Source: api/v1alpha1/utils/commons.go:93-109] — TLSConfig struct definition
- [Source: _bmad-output/implementation-artifacts/d2-2-standardize-kubernetes-auth-engine-docs.md] — Previous story context
- [Source: _bmad-output/implementation-artifacts/d2-1-create-auth-engines-directory-structure-and-index-page.md] — D2.1 story context
- [Source: _bmad-output/implementation-artifacts/epic-d1-retro-2026-06-28.md] — D1 retro: known friction points for D2.3
- [Source: _bmad-output/project-context.md] — Project conventions and coding standards

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6 (claude-4.6-opus)

### Debug Log References

- Integration tests (`make integration`) failed twice during Step 0 pre-flight due to Vault Helm chart timeout in kind cluster (podman provider). Error: "release vault failed, and has been uninstalled due to atomic being set: timed out waiting for the condition". This is a pre-existing infrastructure issue unrelated to documentation changes. Story Dev Notes explicitly state "No Go code changes. No tests to run."

### Completion Notes List

- Created `docs/auth-engines/ldap.md` with complete LDAP auth engine documentation
- Structured per template: Overview → LDAPAuthEngineConfig (Example, CLI, Fields, TLS Config) → LDAPAuthEngineGroup (Example, CLI, Fields) → Credential Resolution → See Also
- Organized LDAPAuthEngineConfig fields into 4 logical groups: Connection & TLS, User Search, Group Search, Token Parameters
- Documented both TLS methods: inline fields and `tLSConfig` secret-based approach with mapping table
- Documented all three `bindCredentials` methods (Kubernetes Secret, Vault Secret, RandomSecret) with YAML examples
- All field names verified against `json:` tags in Go types — using exact camelCase (no snake_case)
- Non-standard casing verified: `TLSMinVersion`, `TLSMaxVersion`, `UPNDomain`, `tLSConfig`
- All relative links verified to exist: `../auth-section.md`, `../contributing-vault-apis.md`, `../secret-management.md`
- Structure validated against `cert.md` reference implementation

### File List

- docs/auth-engines/ldap.md (new)
- _bmad-output/implementation-artifacts/d2-3-standardize-ldap-auth-engine-docs.md (modified — task checkboxes, status, dev agent record)
- _bmad-output/implementation-artifacts/sprint-status.yaml (modified — story status update)
