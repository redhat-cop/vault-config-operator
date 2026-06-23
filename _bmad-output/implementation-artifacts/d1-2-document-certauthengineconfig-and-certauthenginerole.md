# Story D1.2: Document CertAuthEngineConfig and CertAuthEngineRole

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a user of the vault-config-operator,
I want documentation for the CertAuthEngineConfig and CertAuthEngineRole CRDs,
So that I can discover and use the TLS certificate authentication method that the operator already supports.

## Acceptance Criteria

1. **Given** CertAuthEngineConfig and CertAuthEngineRole are implemented in `api/v1alpha1/` with no documentation **When** a new doc file `docs/auth-engines/cert.md` is created following the template from D1.1 **Then** it contains:
   - CertAuthEngineConfig: full YAML example, all field descriptions, link to [Vault TLS cert auth docs](https://developer.hashicorp.com/vault/docs/auth/cert)
   - CertAuthEngineRole: full YAML example, all field descriptions, Vault CLI equivalent
   - Credential resolution for TLS certificates (certificate and key references)

2. **Given** the new file exists **When** the auth-engines index page (`docs/auth-engines.md`) references it **Then** CertAuth is discoverable alongside all other auth engines

## Tasks / Subtasks

- [x] Task 1: Create `docs/auth-engines/` directory (AC: 2)
  - [x] 1.1: `mkdir -p docs/auth-engines`
- [x] Task 2: Create `docs/auth-engines/cert.md` following the D1.1 template (AC: 1)
  - [x] 2.1: Write Title section with link to Vault TLS cert auth docs
  - [x] 2.2: Write Overview paragraph describing the TLS cert auth method
  - [x] 2.3: Write CertAuthEngineConfig section — description, full YAML example, Vault CLI equivalent, field descriptions table
  - [x] 2.4: Write CertAuthEngineRole section — description, full YAML example, Vault CLI equivalent, field descriptions table
  - [x] 2.5: Write Credential Resolution section with certificate/key reference examples (Kubernetes Secret, Vault Secret, RandomSecret)
  - [x] 2.6: Write See Also links section
- [x] Task 3: Add CertAuth entry to the `docs/auth-engines.md` TOC (AC: 2)
  - [x] 3.1: Add `[CertAuthEngineConfig](#certauthengineconfig)` and `[CertAuthEngineRole](#certauthenginerole)` entries to the TOC
  - [x] 3.2: Add a short CertAuth section at the end of `docs/auth-engines.md` pointing to `auth-engines/cert.md`
- [x] Task 4: Validate (AC: 1, 2)
  - [x] 4.1: Verify all internal links resolve
  - [x] 4.2: Verify YAML examples use `apiVersion: redhatcop.redhat.io/v1alpha1`
  - [x] 4.3: Verify all field names are camelCase (matching CRD json tags)

### Review Findings

- [x] [Review][Patch] Document the correct `CertAuthEngineConfig` Vault path and CLI shape [`docs/auth-engines/cert.md:36`]
- [x] [Review][Patch] Replace the misleading Kubernetes Secret `stringData` reference in credential resolution [`docs/auth-engines/cert.md:150`]
- [x] [Review][Patch] Avoid the empty `CertAuthEngineConfig` section in the auth engines index page [`docs/auth-engines.md:777`]

## Dev Notes

### This Is a Documentation-Only Story

No Go code changes. No tests to run. No `make manifests generate`. The deliverables are:
1. New file: `docs/auth-engines/cert.md`
2. Minor edit to: `docs/auth-engines.md` (add TOC entry and cross-reference)

### Dependency on D1.1

This story uses the template defined in D1.1 (`docs/engine-doc-template.md`). If that file doesn't exist yet when you start, follow the template structure defined in the D1.1 story notes:

```
# {EngineName} Auth Engine
  → Title with link to Vault documentation

## Overview
  → 2-3 sentence description

## CertAuthEngineConfig
  ### Description (paragraph with Vault API doc link)
  ### Example (full YAML)
  ### Vault CLI Equivalent (shell code block)
  ### Field Descriptions (markdown table: Field | Type | Required | Description)

## CertAuthEngineRole
  ### Description (paragraph with Vault API doc link)
  ### Example (full YAML)
  ### Vault CLI Equivalent (shell code block)
  ### Field Descriptions (markdown table: Field | Type | Required | Description)

## Credential Resolution
  ### Using a Kubernetes Secret
  ### Using a Vault Secret
  ### Using a RandomSecret

## See Also
  → Links to auth-section.md, contributing-vault-apis.md
```

### CertAuthEngineConfig — Complete Field Reference

Source: `api/v1alpha1/certauthengineconfig_types.go`

Vault API path: `auth/{spec.path}/{name}/config`

| camelCase Field | Type | Required | Default | Vault API Key | Description |
|---|---|---|---|---|---|
| connection | VaultConnection | No | — | — | Override Vault connection settings |
| authentication | KubeAuthConfiguration | Yes | — | — | Kube auth configuration (see auth-section.md) |
| path | string | Yes | — | — | Mount path for the cert auth engine |
| name | string | No | metadata.name | — | Override Vault object name |
| disableBinding | bool | No | false | `disable_binding` | Skip client identity matching during renewal |
| enableIdentityAliasMetadata | bool | No | false | `enable_identity_alias_metadata` | Store certificate metadata in identity alias |
| ocspCacheSize | int | No | 100 | `ocsp_cache_size` | Size of the OCSP response LRU cache (min: 0) |
| roleCacheSize | int | No | 200 | `role_cache_size` | Size of the role cache; -1 disables caching (min: -1) |

### CertAuthEngineRole — Complete Field Reference

Source: `api/v1alpha1/certauthenginerole_types.go`

Vault API path: `auth/{spec.path}/certs/{name}`

| camelCase Field | Type | Required | Default | Vault API Key | Description |
|---|---|---|---|---|---|
| connection | VaultConnection | No | — | — | Override Vault connection settings |
| authentication | KubeAuthConfiguration | Yes | — | — | Kube auth configuration (see auth-section.md) |
| path | string | Yes | — | — | Mount path for the cert auth engine |
| name | string | No | metadata.name | — | Override Vault object name |
| certificate | string | Yes | — | `certificate` | PEM-format CA certificate |
| allowedCommonNames | []string | No | [] (allow all) | `allowed_common_names` | Glob patterns for allowed Common Names |
| allowedDNSSANs | []string | No | [] (allow all) | `allowed_dns_sans` | Glob patterns for allowed DNS SANs |
| allowedEmailSANs | []string | No | [] (allow all) | `allowed_email_sans` | Glob patterns for allowed Email SANs |
| allowedURISANs | []string | No | [] (allow all) | `allowed_uri_sans` | Glob patterns for allowed URI SANs |
| allowedOrganizationalUnits | []string | No | [] (allow all) | `allowed_organizational_units` | Glob patterns for allowed OUs |
| requiredExtensions | []string | No | [] | `required_extensions` | Required Custom Extension OIDs (format: `oid:value` or `hex:oid:value`) |
| allowedMetadataExtensions | []string | No | [] | `allowed_metadata_extensions` | OID extensions to add as metadata on successful auth |
| ocspEnabled | bool | No | false | `ocsp_enabled` | Validate certificate revocation via OCSP |
| ocspCACertificates | string | No | "" | `ocsp_ca_certificates` | Additional OCSP responder certs (base64 PEM) |
| ocspServersOverride | []string | No | [] | `ocsp_servers_override` | Override OCSP server addresses |
| ocspFailOpen | bool | No | false | `ocsp_fail_open` | Allow login if OCSP response unavailable/unknown |
| ocspThisUpdateMaxAge | string | No | "" | `ocsp_this_update_max_age` | Max age of OCSP thisUpdate field (duration) |
| ocspMaxRetries | int64 | No | 4 | `ocsp_max_retries` | OCSP request retry count; 0 disables retries (min: 0) |
| ocspQueryAllServers | bool | No | false | `ocsp_query_all_servers` | Query all OCSP servers and require unanimous agreement |
| displayName | string | No | role name | `display_name` | Display name on tokens issued via this role |
| tokenTTL | string | No | "" | `token_ttl` | Incremental token lifetime (e.g., "1h") |
| tokenMaxTTL | string | No | "" | `token_max_ttl` | Maximum token lifetime |
| tokenPolicies | []string | No | [] | `token_policies` | Policies to attach to generated tokens |
| tokenBoundCIDRs | []string | No | [] | `token_bound_cidrs` | CIDR blocks restricting token usage |
| tokenExplicitMaxTTL | string | No | "" | `token_explicit_max_ttl` | Hard cap TTL overriding tokenTTL/tokenMaxTTL |
| tokenNoDefaultPolicy | bool | No | false | `token_no_default_policy` | Exclude the `default` policy from generated tokens |
| tokenNumUses | int64 | No | 0 (unlimited) | `token_num_uses` | Max token usage count (min: 0) |
| tokenPeriod | string | No | "" | `token_period` | Maximum allowed period for periodic token requests |
| tokenType | string | No | "" | `token_type` | Token type: `service`, `batch`, `default`, `default-service`, `default-batch` |

### Vault CLI Equivalents

**CertAuthEngineConfig:**
```shell
vault write auth/<path>/config \
  disable_binding=false \
  enable_identity_alias_metadata=false \
  ocsp_cache_size=100 \
  role_cache_size=200
```

**CertAuthEngineRole:**
```shell
vault write auth/<path>/certs/<role-name> \
  certificate=@ca.pem \
  allowed_common_names="*.example.com" \
  token_policies="app-policy" \
  token_ttl="1h" \
  token_max_ttl="24h"
```

### YAML Example Snippets

**CertAuthEngineConfig:**
```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: CertAuthEngineConfig
metadata:
  name: cert-config
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  path: cert
  ocspCacheSize: 100
  roleCacheSize: 200
  disableBinding: false
  enableIdentityAliasMetadata: false
```

**CertAuthEngineRole:**
```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: CertAuthEngineRole
metadata:
  name: my-app-cert-role
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  path: cert
  certificate: |
    -----BEGIN CERTIFICATE-----
    MIIDQjCCAiqgAwIBAgI...
    -----END CERTIFICATE-----
  allowedCommonNames:
    - "*.example.com"
  allowedDNSSANs:
    - "*.internal.example.com"
  tokenPolicies:
    - app-policy
    - default
  tokenTTL: "1h"
  tokenMaxTTL: "24h"
```

### CertAuth Does NOT Use Credential Resolution in the Usual Sense

Unlike LDAP or Database engines that resolve credentials (passwords, tokens) from Kubernetes Secrets or Vault Secrets before writing to the Vault API, the CertAuth engine's sensitive field is the `certificate` PEM itself — which is a string field directly in the spec.

The "Credential Resolution" section in the doc should explain that:
- The `certificate` field is set directly in the CR spec as a PEM string
- For production usage, users typically store the CA cert in a Kubernetes Secret and reference it via `stringData`, then include the PEM inline in the CR
- There is no `spec.credentialSecret` or `spec.vaultSecretRef` pattern for this type — CertAuth is one of the simpler patterns

This is different from engines like LDAPAuthEngineConfig which have `spec.bindCredentials` with Kubernetes Secret, Vault Secret, and RandomSecret options.

### Existing Patterns to Follow

The doc should match the style of existing engine sections in `docs/auth-engines.md`:
- KubernetesAuthEngineConfig section (lines 37-71) for the Config CRD pattern
- KubernetesAuthEngineRole section (lines 73-end of that section) for the Role CRD pattern
- LDAPAuthEngineConfig section for credential resolution reference

Key style points from existing docs:
- Description paragraph includes a link to Vault docs in parentheses or inline
- YAML examples show realistic values (not just placeholders)
- Field descriptions are prose paragraphs in existing docs, but the template (D1.1) specifies a **markdown table** format instead
- The `authentication` block in examples uses `path: kubernetes` and `role: policy-admin` as standard values

### Directory Structure Decision

The story specifies `docs/auth-engines/cert.md` — this is the **first** file in the `docs/auth-engines/` directory. The directory does not exist yet. Story D2.1 will later create the full directory structure with an index page and move other engine docs there.

For now, this story:
1. Creates `docs/auth-engines/` directory
2. Creates `docs/auth-engines/cert.md`
3. Adds a reference in the existing `docs/auth-engines.md` monolith

### What NOT to Do

- Do NOT split existing engine docs out of `docs/auth-engines.md` — that's D2 scope
- Do NOT create an `index.md` in `docs/auth-engines/` — that's D2.1 scope
- Do NOT modify any Go code, CRD types, or controllers
- Do NOT modify `docs/engine-doc-template.md` (just follow its pattern)
- Do NOT run `make manifests generate` or `make test`
- Do NOT document CertAuth in `docs/auth-engines.md` with the full content — only add a TOC entry and a brief reference pointing to `auth-engines/cert.md`

### Phase 1.5 Non-Functional Requirements to Follow

- **DNFR1:** Follow the template structure from D1.1
- **DNFR2:** All YAML examples must use `apiVersion: redhatcop.redhat.io/v1alpha1`
- **DNFR3:** Field descriptions must use camelCase field names (matching Go struct json tags)
- **DNFR4:** Internal cross-references must use relative paths that work from `docs/auth-engines/`

### Project Structure Notes

- New directory: `docs/auth-engines/`
- New file: `docs/auth-engines/cert.md`
- Modified file: `docs/auth-engines.md` (TOC addition only)
- Docs directory currently has 9 markdown files (flat structure)
- This is the first file to live in a subdirectory under `docs/`

### References

- [Source: _bmad-output/planning-artifacts/epics.md#Story D1.2] — Story requirements and acceptance criteria
- [Source: _bmad-output/planning-artifacts/epics.md#Phase 1.5 Requirements] — DOC1-DOC10, DNFR1-DNFR5
- [Source: api/v1alpha1/certauthengineconfig_types.go] — CertAuthEngineConfig type definition, toMap(), field comments
- [Source: api/v1alpha1/certauthenginerole_types.go] — CertAuthEngineRole type definition, toMap(), field comments
- [Source: _bmad-output/implementation-artifacts/d1-1-create-documentation-template-and-pattern-guide.md] — Previous story (D1.1) defining the template structure
- [Source: docs/auth-engines.md] — Current monolith doc to add TOC reference
- [Source: docs/auth-section.md] — Common authentication section docs (link target)
- [Source: docs/contributing-vault-apis.md] — Developer guide (link target for See Also)
- [Vault Docs: https://developer.hashicorp.com/vault/docs/auth/cert] — Vault TLS cert auth method documentation
- [Vault API: https://developer.hashicorp.com/vault/api-docs/auth/cert] — Vault TLS cert auth API reference

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6

### Debug Log References

None — documentation-only story with no code changes or test runs.

### Completion Notes List

- Created `docs/auth-engines/cert.md` following the D1.1 template structure with all required sections
- CertAuthEngineConfig section includes: description with Vault API link, full YAML example, Vault CLI equivalent, field descriptions table (8 fields including path, authentication, connection, name, disableBinding, enableIdentityAliasMetadata, ocspCacheSize, roleCacheSize)
- CertAuthEngineRole section includes: description with Vault API link, full YAML example, Vault CLI equivalent, field descriptions table (28 fields covering certificate matching, OCSP validation, and token parameters)
- Credential Resolution section explains the simpler inline PEM pattern (no credentialSecret/vaultSecretRef pattern for CertAuth)
- See Also section links to auth-section.md, contributing-vault-apis.md, and Vault external docs
- Added CertAuth TOC entries and cross-reference section to docs/auth-engines.md
- All field names verified as camelCase matching Go struct json tags
- All YAML examples use apiVersion: redhatcop.redhat.io/v1alpha1
- All internal links verified to resolve correctly from the docs/auth-engines/ subdirectory

### Change Log

- 2026-06-23: Created docs/auth-engines/cert.md and updated docs/auth-engines.md TOC (Story D1.2)
- 2026-06-23: Code review — fixed CertAuthEngineConfig Vault path (added `{name}` segment), clarified credential resolution wording, demoted CertAuthEngineRole to `###` in index page

### File List

- docs/auth-engines/cert.md (new)
- docs/auth-engines.md (modified — TOC and CertAuth cross-reference section added)
