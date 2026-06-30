# Story D2.4: Standardize JWT/OIDC Auth Engine Docs

Status: ready-for-dev

## Story

As a user configuring JWT/OIDC authentication,
I want clear documentation for both JWT and OIDC modes,
So that I can configure either mode correctly.

## Acceptance Criteria

1. **Given** the existing JWTOIDCAuthEngine content in `docs/auth-engines.md` (lines 256-423) **When** it is extracted to `docs/auth-engines/jwt-oidc.md` and standardized per the template **Then** it contains:
   - Overview linking to Vault JWT/OIDC auth docs
   - JWTOIDCAuthEngineConfig: complete YAML example, field descriptions (camelCase), Vault CLI equivalent
   - JWTOIDCAuthEngineRole: complete YAML example, field descriptions, Vault CLI equivalent
   - Credential Resolution section documenting the three `OIDCCredentials` methods

2. **Given** the new `jwt-oidc.md` file **When** validated against the template structure **Then** it follows the same structure as `docs/auth-engines/cert.md` (Overview → Config CRD → Role CRD → Credential Resolution → See Also)

3. **Given** the field descriptions **When** validated against the Go type definitions **Then** ALL field names use camelCase (matching `json:` tags exactly) with no residual snake_case

4. **Given** the `OIDCCredentials` section **When** reviewed **Then** all three methods (Kubernetes Secret, Vault Secret, RandomSecret) are clearly documented with YAML examples using Pattern B (nested credential object)

5. **Given** the dual-mode nature (JWT vs OIDC) **When** the documentation is reviewed **Then** the difference between JWT and OIDC modes is clearly explained, including which fields apply to each mode

## Tasks / Subtasks

- [ ] Task 1: Create `docs/auth-engines/jwt-oidc.md` (AC: 1, 2)
  - [ ] 1.1: Write Overview section — 2-3 sentences explaining JWT/OIDC auth, link to Vault docs, list the two CRDs (JWTOIDCAuthEngineConfig, JWTOIDCAuthEngineRole), explain dual-mode support
  - [ ] 1.2: Write JWTOIDCAuthEngineConfig section with Example YAML (OIDC mode with Azure provider), Vault CLI Equivalent, and Field Descriptions table
  - [ ] 1.3: Organize JWTOIDCAuthEngineConfig fields into logical groups: OIDC Discovery, JWKS Validation, JWT Validation, General Settings
  - [ ] 1.4: Write a "JWT vs OIDC Modes" subsection explaining mutually-exclusive validation source fields
  - [ ] 1.5: Write JWTOIDCAuthEngineRole section with Example YAML, Vault CLI Equivalent, and Field Descriptions table
  - [ ] 1.6: Organize JWTOIDCAuthEngineRole fields into logical groups: Role Identity, Claims & Binding, OIDC-Specific, JWT-Specific (leeway fields), Token Parameters
  - [ ] 1.7: Write Credential Resolution section documenting all three `OIDCCredentials` methods (Pattern B nested object) with YAML examples
  - [ ] 1.8: Add "See Also" section with links to auth-section.md, contributing-vault-apis.md, and Vault docs

- [ ] Task 2: Audit field names for camelCase consistency (AC: 3)
  - [ ] 2.1: Cross-reference all field names in the new doc against `jwtoidcauthengineconfig_types.go` and `jwtoidcauthenginerole_types.go` — field names MUST match the `json:` tag values exactly
  - [ ] 2.2: Fix any residual snake_case field names from the original `auth-engines.md` source
  - [ ] 2.3: Verify fields with non-standard casing are correct: `OIDCDiscoveryURL`, `OIDCDiscoveryCAPEM`, `OIDCClientID`, `OIDCResponseMode`, `OIDCResponseTypes`, `JWKSURL`, `JWKSCAPEM`, `JWTValidationPubKeys`, `JWTSupportedAlgs`, `OIDCScopes` (these use uppercase prefixes per Go convention)

- [ ] Task 3: Verify links and structure (AC: 2)
  - [ ] 3.1: Verify relative links resolve correctly from `docs/auth-engines/jwt-oidc.md` (`../auth-section.md`, `../contributing-vault-apis.md`, `../secret-management.md`)
  - [ ] 3.2: Verify structure matches `cert.md` pattern: heading hierarchy, section ordering, table format

## Dev Notes

### This Is a Documentation-Only Story

No Go code changes. No tests to run. No `make manifests generate`. The deliverables are:
- 1 new file: `docs/auth-engines/jwt-oidc.md`

### Dependency on D2.1

This story assumes D2.1 has been completed (creating `docs/auth-engines/index.md` and the redirect pointer in `docs/auth-engines.md`). If D2.1 is NOT yet done, this story can still proceed — the `jwt-oidc.md` file can be created independently. The index will reference it via `[jwt-oidc.md](jwt-oidc.md)`.

### Source Content Location

The content to extract and standardize is in `docs/auth-engines.md` lines 256-423:
- `## JWTOIDCAuthEngineConfig` (lines 256-344)
- `## JWTOIDCAuthEngineRole` (lines 347-423)

### Template to Follow

Use `docs/engine-doc-template.md` as the structural pattern. Use `docs/auth-engines/cert.md` as the concrete reference implementation.

Key structural requirements from the template:
1. Title: `# JWT/OIDC Auth Engine`
2. Link to Vault docs immediately below title
3. `## Overview` — 2-3 sentences + CRD list + dual-mode explanation
4. `## JWTOIDCAuthEngineConfig` → `### Example` → `### Vault CLI Equivalent` → `### Field Descriptions` → `### JWT vs OIDC Modes`
5. `## JWTOIDCAuthEngineRole` → `### Example` → `### Vault CLI Equivalent` → `### Field Descriptions`
6. `## Credential Resolution` — documenting the three `OIDCCredentials` methods (Pattern B nested object — only needed for OIDC mode)
7. `## See Also`

### JWT vs OIDC Modes — Critical Documentation Requirement

The JWT/OIDC engine supports TWO mutually-exclusive validation source configurations. This MUST be clearly explained:

| Mode | Validation Source | Required Config Fields |
|------|------------------|-----------------------|
| OIDC | OIDC Discovery URL | `OIDCDiscoveryURL`, optionally `OIDCDiscoveryCAPEM` |
| JWT (JWKS) | Remote JWKS URL | `JWKSURL`, optionally `JWKSCAPEM` |
| JWT (Public Keys) | Local public keys | `JWTValidationPubKeys` |

Only ONE of these three can be set. The original docs mention this but don't present it clearly as a table or decision point.

**OIDC mode additionally requires** `OIDCCredentials` (client ID + client secret) — the JWT modes do NOT need credentials.

### JWTOIDCAuthEngineConfig — Complete Field Reference

From `api/v1alpha1/jwtoidcauthengineconfig_types.go`, the `JWTOIDCConfig` struct has these fields:

**OIDC Discovery Fields:**

| CRD Field (camelCase) | JSON tag | Vault API key (snake_case) | Type | Required | Default |
|---|---|---|---|---|---|
| OIDCDiscoveryURL | `json:"OIDCDiscoveryURL,omitempty"` | `oidc_discovery_url` | string | No* | — |
| OIDCDiscoveryCAPEM | `json:"OIDCDiscoveryCAPEM,omitempty"` | `oidc_discovery_ca_pem` | string | No | — |
| OIDCClientID | `json:"OIDCClientID,omitempty"` | `oidc_client_id` | string | No* | — |
| OIDCResponseMode | `json:"OIDCResponseMode,omitempty"` | `oidc_response_mode` | string | No | — |
| OIDCResponseTypes | `json:"OIDCResponseTypes,omitempty"` | `oidc_response_types` | []string | No | — |

*Required when using OIDC mode

**JWKS/JWT Validation Fields:**

| CRD Field (camelCase) | JSON tag | Vault API key (snake_case) | Type | Required | Default |
|---|---|---|---|---|---|
| JWKSURL | `json:"JWKSURL,omitempty"` | `jwks_url` | string | No* | — |
| JWKSCAPEM | `json:"JWKSCAPEM,omitempty"` | `jwks_ca_pem` | string | No | — |
| JWTValidationPubKeys | `json:"JWTValidationPubKeys,omitempty"` | `jwt_validation_pubkeys` | []string | No* | — |

*One of `OIDCDiscoveryURL`, `JWKSURL`, or `JWTValidationPubKeys` must be set

**General Settings:**

| CRD Field (camelCase) | JSON tag | Vault API key (snake_case) | Type | Required | Default |
|---|---|---|---|---|---|
| boundIssuer | `json:"boundIssuer,omitempty"` | `bound_issuer` | string | No | — |
| JWTSupportedAlgs | `json:"JWTSupportedAlgs,omitempty"` | `jwt_supported_algs` | []string | No | [RS256] for OIDC |
| defaultRole | `json:"defaultRole,omitempty"` | `default_role` | string | No | — |
| providerConfig | `json:"providerConfig,omitempty"` | `provider_config` | JSON | No | — |
| namespaceInState | `json:"namespaceInState"` | `namespace_in_state` | bool | No | true |

Additional top-level spec fields (NOT in `JWTOIDCConfig` inline struct):
- `path` (Required) — mount path for the JWT/OIDC auth engine
- `authentication` (Required) — see shared auth-section.md
- `connection` (Optional) — override Vault connection
- `OIDCCredentials` (Optional) — credential resolution config for OIDC mode (see Credential Resolution below)

### JWTOIDCAuthEngineRole — Complete Field Reference

From `api/v1alpha1/jwtoidcauthenginerole_types.go`, the `JWTOIDCRole` struct has these fields:

**Role Identity Fields:**

| CRD Field (camelCase) | JSON tag | Vault API key (snake_case) | Type | Required | Default |
|---|---|---|---|---|---|
| name | `json:"name"` | `name` | string | Yes | — |
| roleType | `json:"roleType,omitempty"` | `role_type` | string | No | `"oidc"` |
| userClaim | `json:"userClaim"` | `user_claim` | string | Yes | — |
| userClaimJSONPointer | `json:"userClaimJSONPointer,omitempty"` | `user_claim_json_pointer` | bool | No | false |

**Claims & Binding Fields:**

| CRD Field (camelCase) | JSON tag | Vault API key (snake_case) | Type | Required | Default |
|---|---|---|---|---|---|
| boundAudiences | `json:"boundAudiences,omitempty"` | `bound_audiences` | []string | No* | — |
| boundSubject | `json:"boundSubject,omitempty"` | `bound_subject` | string | No | — |
| boundClaims | `json:"boundClaims,omitempty"` | `bound_claims` | JSON | No | — |
| boundClaimsType | `json:"boundClaimsType"` | `bound_claims_type` | string | No | `"string"` |
| groupsClaim | `json:"groupsClaim,omitempty"` | `groups_claim` | string | No | — |
| claimMappings | `json:"claimMappings,omitempty"` | `claim_mappings` | map[string]string | No | — |

*Required for "jwt" roles, optional for "oidc" roles

**OIDC-Specific Fields:**

| CRD Field (camelCase) | JSON tag | Vault API key (snake_case) | Type | Required | Default |
|---|---|---|---|---|---|
| OIDCScopes | `json:"OIDCScopes,omitempty"` | `oidc_scopes` | []string | No | — |
| allowedRedirectURIs | `json:"allowedRedirectURIs,omitempty"` | `allowed_redirect_uris` | []string | Yes* | — |
| verboseOIDCLogging | `json:"verboseOIDCLogging,omitempty"` | `verbose_oidc_logging` | bool | No | false |
| maxage | `json:"maxage,omitempty"` | `max_age` | int64 | No | 0 |

*Required for "oidc" roles

**JWT-Specific Fields (Leeway):**

| CRD Field (camelCase) | JSON tag | Vault API key (snake_case) | Type | Required | Default |
|---|---|---|---|---|---|
| clockSkewLeeway | `json:"clockSkewLeeway,omitempty"` | `clock_skew_leeway` | int64 | No | 60 |
| expirationLeeway | `json:"expirationLeeway,omitempty"` | `expiration_leeway` | int64 | No | 150 |
| notBeforeLeeway | `json:"notBeforeLeeway,omitempty"` | `not_before_leeway` | int64 | No | 150 |

**Token Parameters:**

| CRD Field (camelCase) | JSON tag | Vault API key (snake_case) | Type | Required | Default |
|---|---|---|---|---|---|
| tokenTTL | `json:"tokenTTL,omitempty"` | `token_ttl` | string | No | — |
| tokenMaxTTL | `json:"tokenMaxTTL,omitempty"` | `token_max_ttl` | string | No | — |
| tokenPolicies | `json:"tokenPolicies,omitempty"` | `token_policies` | []string | No | — |
| tokenBoundCIDRs | `json:"tokenBoundCIDRs,omitempty"` | `token_bound_cidrs` | []string | No | — |
| tokenExplicitMaxTTL | `json:"tokenExplicitMaxTTL,omitempty"` | `token_explicit_max_ttl` | string | No | — |
| tokenNoDefaultPolicy | `json:"tokenNoDefaultPolicy,omitempty"` | `token_no_default_policy` | bool | No | false |
| tokenNumUses | `json:"tokenNumUses,omitempty"` | `token_num_uses` | int64 | No | 0 |
| tokenPeriod | `json:"tokenPeriod,omitempty"` | `token_period` | int64 | No | 0 |
| tokenType | `json:"tokenType,omitempty"` | `token_type` | string | No | — |

Additional top-level spec fields:
- `path` (Required) — JWT/OIDC auth mount path
- `authentication` (Required) — see shared auth-section.md
- `connection` (Optional) — override Vault connection

**Vault path:** `auth/{spec.path}/role/{spec.name || metadata.name}`

### Credential Resolution (OIDCCredentials)

The JWT/OIDC engine uses `OIDCCredentials` field of type `*RootCredentialConfig`. This is **Pattern B** from the template (nested credential object).

The `OIDCCredentials` field resolves the OIDC Client ID and Client Secret used to communicate with the OIDC provider.

**Important behaviors:**
- `OIDCCredentials` is ONLY needed for OIDC mode (not JWT mode)
- If `OIDCClientID` is set directly in the spec, it takes precedence over the username from the referenced secret — the password (client secret) is always retrieved from the secret source
- If `OIDCCredentials` is nil/empty AND the config matches the zero-value check `{PasswordKey: "password", UsernameKey: "username"}`, credential resolution is skipped entirely (pure JWT mode)

From `api/v1alpha1/utils/commons.go`, `RootCredentialConfig` has:
- `secret` — Kubernetes Secret reference
- `vaultSecret` — Vault secret path reference
- `randomSecret` — RandomSecret reference
- `usernameKey` — key for client ID (default: `"username"`)
- `passwordKey` — key for client secret (default: `"password"`)

### IsDeletable Behavior

`JWTOIDCAuthEngineConfig` returns `IsDeletable() == false` — deleting the CR does NOT remove the JWT/OIDC config from Vault. The auth mount must be disabled separately.

`JWTOIDCAuthEngineRole` returns `IsDeletable() == true` — deleting the CR removes the role from Vault.

### Known Issues in Source Content

The original `auth-engines.md` JWT/OIDC section has these problems that MUST be fixed during extraction:

1. **Incorrect `tokenTTL` description for the role** — The source (line 408) has a copy-paste error: it describes `tokenTTL` with the `maxage` description ("Specifies the allowable elapsed time..."). The correct description is: "The incremental lifetime for generated tokens. This current value of this will be referenced at renewal time."

2. **Missing `JWKSCAPEM` from the source** — The source does not mention the `JWKSCAPEM` field at all but it exists in the Go type. Include it.

3. **Field names already appear as camelCase** — Unlike Kubernetes/LDAP sections, the JWT/OIDC section in `auth-engines.md` already uses the uppercase-prefix camelCase pattern. Verify but expect fewer corrections.

4. **`OIDCResponseMode` enum validation** — The Go type has `+kubebuilder:validation:Enum={"query","form_post"}` and the doc mentions these values. Ensure the doc clearly states these are the ONLY allowed values.

5. **`boundClaimsType` enum validation** — The Go type has `+kubebuilder:validation:Enum={"string","glob"}`. Document these as the only allowed values.

6. **`tokenType` enum validation** — The Go type has `+kubebuilder:validation:Enum={"service","batch","default","default-service","default-batch"}`. Document all five allowed values.

7. **`roleType` enum validation** — The Go type has `+kubebuilder:validation:Enum={"oidc","jwt"}`. Document these as the only allowed values.

### Relative Link Conventions

From `docs/auth-engines/jwt-oidc.md`:
- To shared docs: `../auth-section.md`, `../contributing-vault-apis.md`, `../secret-management.md`
- To other engine files: `cert.md`, `kubernetes.md`, `ldap.md` (same directory)
- External: full URLs to Vault documentation

### Previous Story Intelligence

**From D2.3 (LDAP Auth Engine Docs):**
- Established the credential resolution documentation pattern (Pattern A for LDAP; JWT/OIDC uses Pattern B)
- Organized fields into logical groups with separate tables — follow same approach
- Included TLS Configuration subsection — JWT/OIDC does NOT use TLS config, so skip
- Used `cert.md` as structural reference — continue this pattern

**From D2.2 (Kubernetes Auth Engine Docs):**
- Established the basic template pattern: Overview → Config → Role → See Also
- Included behavior table for complex multi-field interactions — JWT/OIDC needs similar for mode selection
- No credential resolution section for Kubernetes — JWT/OIDC DOES need one

**From D2.1 (Directory Structure & Index Page):**
- Created `docs/auth-engines/index.md` with engine table linking to `jwt-oidc.md`
- Replaced `docs/auth-engines.md` with redirect pointer
- AuthEngineMount section is in the index page (not per-engine files)

**From D1.1 (Template Creation):**
- Template was patched 4 times in review — always use current version at `docs/engine-doc-template.md`
- Template explicitly documents Pattern B (nested credential object) for JWT/OIDC — use this pattern

**From D1.2 (CertAuth Documentation):**
- First per-engine file at `docs/auth-engines/cert.md` — use as reference implementation
- Cert auth has inline credential field (Pattern A variant) — JWT/OIDC uses nested object (Pattern B)

**From D1.3 (Link and Naming Fixes):**
- Fixed snake_case→camelCase in GCP and Azure sections
- JWT/OIDC section was NOT explicitly mentioned as fixed, but fields already appear correct in source
- Fixed leading-space code fences — verify none remain in JWT/OIDC section

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
│   ├── ldap.md           ← D2.3
│   └── jwt-oidc.md       ← NEW (this story)
├── auth-engines.md       ← redirect pointer (D2.1)
├── auth-section.md       ← shared auth config docs (unchanged)
├── engine-doc-template.md ← template (D1.1, review-patched)
├── secret-management.md  ← link target for RandomSecret reference
└── ...
```

### References

- [Source: _bmad-output/planning-artifacts/epics.md#Story D2.4] — Story requirements and acceptance criteria
- [Source: docs/auth-engines.md:256-423] — JWT/OIDC auth content to extract and standardize
- [Source: docs/auth-engines/cert.md] — Reference implementation for template pattern
- [Source: docs/engine-doc-template.md] — Template structure (D1.1, review-patched 4 times)
- [Source: api/v1alpha1/jwtoidcauthengineconfig_types.go] — CRD field definitions for Config (JWTOIDCConfig struct)
- [Source: api/v1alpha1/jwtoidcauthenginerole_types.go] — CRD field definitions for Role (JWTOIDCRole struct)
- [Source: api/v1alpha1/utils/commons.go] — RootCredentialConfig struct definition
- [Source: _bmad-output/implementation-artifacts/d2-3-standardize-ldap-auth-engine-docs.md] — Previous story context
- [Source: _bmad-output/implementation-artifacts/d2-2-standardize-kubernetes-auth-engine-docs.md] — D2.2 story context
- [Source: _bmad-output/implementation-artifacts/d2-1-create-auth-engines-directory-structure-and-index-page.md] — D2.1 story context
- [Source: _bmad-output/project-context.md] — Project conventions and coding standards

## Dev Agent Record

### Agent Model Used

{{agent_model_name_version}}

### Debug Log References

### Completion Notes List

### File List
