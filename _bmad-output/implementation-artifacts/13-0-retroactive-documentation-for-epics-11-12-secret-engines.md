---
baseline_commit: 2ccad3a67685ec8bdd0c585247ee148143415e6c
---

# Story 13.0: Retroactive Documentation for Epics 11-12 Secret Engines

Status: done

## Story

As a user of the vault-config-operator,
I want documentation for the 6 secret engine types shipped in Epics 11-12,
So that I can discover and correctly use AWS, Transit, SSH, Consul, GCP, and LDAP secret engines.

## Acceptance Criteria

1. **Given** Epics 11 and 12 shipped without documentation (violating DNFR5) **When** documentation is created following `docs/engine-doc-template.md` **Then** the following files exist and comply with DNFR1-DNFR3:
   - `docs/secret-engines/aws.md` — AWSSecretEngineConfig, AWSSecretEngineRole
   - `docs/secret-engines/transit.md` — TransitSecretEngineKey
   - `docs/secret-engines/ssh.md` — SSHSecretEngineConfig, SSHSecretEngineRole
   - `docs/secret-engines/consul.md` — ConsulSecretEngineConfig, ConsulSecretEngineRole
   - `docs/secret-engines/gcp.md` — GCPSecretEngineConfig, GCPSecretEngineRoleset, GCPSecretEngineStaticAccount
   - `docs/secret-engines/ldap.md` — LDAPSecretEngineConfig, LDAPSecretEngineStaticRole, LDAPSecretEngineDynamicRole

2. **Given** each doc file is created **When** reviewed **Then** each contains: overview with Vault docs link, config CRD section (YAML example, Vault CLI equivalent, field descriptions table), role/key CRD section(s), credential resolution section (where applicable), and See Also links

3. **Given** `docs/secret-engines/index.md` exists **When** the new doc files are created **Then** the index page's "Supported Secret Engines" table is updated to include links to all 6 new engine docs

## Tasks / Subtasks

- [x] Task 1: Create `docs/secret-engines/aws.md` (AC: 1, 2)
  - [x] 1.1: Overview section — Vault AWS secrets engine generates IAM credentials on-demand
  - [x] 1.2: AWSSecretEngineConfig section — YAML example, Vault CLI equivalent, field descriptions from `awssecretengineconfig_types.go`
  - [x] 1.3: AWSSecretEngineRole section — YAML example, Vault CLI equivalent, field descriptions from `awssecretenginerole_types.go`
  - [x] 1.4: Credential Resolution section — Pattern A (`rootCredentials` field: K8s Secret, VaultSecret, RandomSecret)
  - [x] 1.5: See Also links

- [x] Task 2: Create `docs/secret-engines/transit.md` (AC: 1, 2)
  - [x] 2.1: Overview section — Vault Transit engine provides encryption-as-a-service
  - [x] 2.2: TransitSecretEngineKey section — YAML example, Vault CLI equivalent, field descriptions from `transitsecretenginekey_types.go`
  - [x] 2.3: Note on create-time vs config-time field mutability
  - [x] 2.4: See Also links (no credential resolution needed)

- [x] Task 3: Create `docs/secret-engines/ssh.md` (AC: 1, 2)
  - [x] 3.1: Overview section — Vault SSH engine provides signed certificates for SSH access
  - [x] 3.2: SSHSecretEngineConfig section — YAML example, Vault CLI equivalent, field descriptions from `sshsecretengineconfig_types.go`
  - [x] 3.3: SSHSecretEngineRole section — YAML example, Vault CLI equivalent, field descriptions from `sshsecretenginerole_types.go`
  - [x] 3.4: Credential Resolution section — Pattern A (optional `caKeyReference` for externally-managed CA key)
  - [x] 3.5: See Also links

- [x] Task 4: Create `docs/secret-engines/consul.md` (AC: 1, 2)
  - [x] 4.1: Overview section — Vault Consul engine generates Consul ACL tokens on-demand
  - [x] 4.2: ConsulSecretEngineConfig section — YAML example, Vault CLI equivalent, field descriptions from `consulsecretengineconfig_types.go`
  - [x] 4.3: ConsulSecretEngineRole section — YAML example, Vault CLI equivalent, field descriptions from `consulsecretenginerole_types.go`
  - [x] 4.4: Credential Resolution section — Pattern A (`rootCredentials` for Consul ACL management token)
  - [x] 4.5: See Also links

- [x] Task 5: Create `docs/secret-engines/gcp.md` (AC: 1, 2)
  - [x] 5.1: Overview section — Vault GCP engine generates GCP service account credentials
  - [x] 5.2: GCPSecretEngineConfig section — YAML example, Vault CLI equivalent, field descriptions from `gcpsecretengineconfig_types.go`
  - [x] 5.3: GCPSecretEngineRoleset section — YAML example, Vault CLI equivalent, field descriptions from `gcpsecretengineroleset_types.go`
  - [x] 5.4: GCPSecretEngineStaticAccount section — YAML example, Vault CLI equivalent, field descriptions from `gcpsecretenginestaticaccount_types.go`
  - [x] 5.5: Credential Resolution section — Pattern B (nested `gcpCredentials` object with passwordKey defaulting to "credentials")
  - [x] 5.6: See Also links

- [x] Task 6: Create `docs/secret-engines/ldap.md` (AC: 1, 2)
  - [x] 6.1: Overview section — Vault LDAP secrets engine manages LDAP passwords (static rotation + dynamic credentials)
  - [x] 6.2: LDAPSecretEngineConfig section — YAML example, Vault CLI equivalent, field descriptions from `ldapsecretengineconfig_types.go`
  - [x] 6.3: LDAPSecretEngineStaticRole section — YAML example, Vault CLI equivalent, field descriptions from `ldapsecretenginestaticrole_types.go`
  - [x] 6.4: LDAPSecretEngineDynamicRole section — YAML example, Vault CLI equivalent, field descriptions from `ldapsecretengine_dynamicrole_types.go`
  - [x] 6.5: Credential Resolution section — Pattern A (`bindCredentials` for LDAP bind credentials)
  - [x] 6.6: See Also links

- [x] Task 7: Update `docs/secret-engines/index.md` (AC: 3)
  - [x] 7.1: Add rows to "Supported Secret Engines" table for all 6 new engines
  - [x] 7.2: Maintain alphabetical order within the table

## Dev Notes

This is a **documentation-only story** — no Go code changes, no `make manifests`, no tests. All CRD types are already implemented and merged in Epics 11-12.

### Story Intelligence Chain

- **Epic 11** (done 2026-08-08): Implemented AWS, Transit, SSH secret engine CRDs — all fully functional with unit + integration tests.
- **Epic 12** (done 2026-08-10): Implemented Consul, GCP, LDAP secret engine CRDs — all fully functional with unit + integration tests.
- **Neither epic created documentation** — this was explicitly deferred to Story 13.0 as a DNFR5 remediation story.
- **Existing docs pattern established**: Epics D3-D4 created the `docs/secret-engines/` directory, standardized the template (`docs/engine-doc-template.md`), and documented the original 7 engines (Database, PKI, RabbitMQ, GitHub, Quay, Kubernetes, Azure).

### Quality Bar Reference

Use `docs/secret-engines/database.md` or `docs/secret-engines/kubernetes.md` as the **quality reference** for structure, level of detail, YAML examples, and field tables.

### Source Files for Field Definitions

| Engine | Types File(s) |
|--------|---------------|
| AWS | `api/v1alpha1/awssecretengineconfig_types.go`, `api/v1alpha1/awssecretenginerole_types.go` |
| Transit | `api/v1alpha1/transitsecretenginekey_types.go` |
| SSH | `api/v1alpha1/sshsecretengineconfig_types.go`, `api/v1alpha1/sshsecretenginerole_types.go` |
| Consul | `api/v1alpha1/consulsecretengineconfig_types.go`, `api/v1alpha1/consulsecretenginerole_types.go` |
| GCP | `api/v1alpha1/gcpsecretengineconfig_types.go`, `api/v1alpha1/gcpsecretengineroleset_types.go`, `api/v1alpha1/gcpsecretenginestaticaccount_types.go` |
| LDAP | `api/v1alpha1/ldapsecretengineconfig_types.go`, `api/v1alpha1/ldapsecretenginestaticrole_types.go`, `api/v1alpha1/ldapsecretengine_dynamicrole_types.go` |

### Key Implementation Details per Engine

**AWS Secret Engine:**
- Config path: `{path}/config/root`
- Role path: `{path}/roles/{name}`
- `IsDeletable() = false` (config persists after CR deletion)
- Credential resolution: `rootCredentials` (Pattern A) — resolves access_key + secret_key
- Vault docs: https://developer.hashicorp.com/vault/docs/secrets/aws
- Role `credentialType` has enum: `iam_user`, `assumed_role`, `federation_token`, `session_token`

**Transit Secret Engine:**
- Key path: `{path}/keys/{name}`
- No credentials needed (uses Vault's internal encryption)
- Has create-time immutable fields (`type`, `derived`, `convergentEncryption`, `keySize`) and config-time mutable fields (`minDecryptionVersion`, `minEncryptionVersion`, `deletionAllowed`, `exportable`, `allowPlaintextBackup`, `autoRotatePeriod`)
- Vault docs: https://developer.hashicorp.com/vault/docs/secrets/transit
- Key `type` has enum: aes128-gcm96, aes256-gcm96, chacha20-poly1305, ed25519, ecdsa-p256, ecdsa-p384, ecdsa-p521, rsa-2048, rsa-3072, rsa-4096, hmac

**SSH Secret Engine:**
- Config path: `{path}/config/ca`
- Role path: `{path}/roles/{name}`
- Config `IsDeletable() = true`
- Credential resolution: optional `caKeyReference` (Pattern A) — only needed when `generateSigningKey=false`
- Role `keyType` enum: `otp`, `ca`
- Vault docs: https://developer.hashicorp.com/vault/docs/secrets/ssh

**Consul Secret Engine:**
- Config path: `{path}/config/access`
- Role path: `{path}/roles/{name}`
- Config has no `Name` override (uses fixed `config/access` path)
- Credential resolution: `rootCredentials` (Pattern A) — resolves Consul ACL management token
- Vault docs: https://developer.hashicorp.com/vault/docs/secrets/consul

**GCP Secret Engine:**
- Config path: `{path}/config`
- Roleset path: `{path}/roleset/{name}`
- Static account path: `{path}/static-account/{name}`
- Credential resolution: `gcpCredentials` (Pattern B — nested object, `passwordKey` defaults to "credentials")
- `GCPCredentialConfig` has its own `ValidateCredentialSource()` method
- Vault docs: https://developer.hashicorp.com/vault/docs/secrets/gcp

**LDAP Secret Engine:**
- Config path: `{path}/config`
- Static role path: `{path}/static-role/{name}`
- Dynamic role path: `{path}/role/{name}`
- Credential resolution: `bindCredentials` (Pattern A) — resolves bindDN + bindPass
- Three CRD types: Config, StaticRole (managed password rotation), DynamicRole (on-demand credentials via LDIF templates)
- Vault docs: https://developer.hashicorp.com/vault/docs/secrets/ldap

### Index Table Format

The current `docs/secret-engines/index.md` table uses this format:
```
| Engine | Config CRD | Role CRD(s) | File |
```

New rows to add (maintain logical grouping):
- AWS: `AWSSecretEngineConfig` | `AWSSecretEngineRole` | [aws.md](aws.md)
- Transit: `—` (no config CRD) | `TransitSecretEngineKey` | [transit.md](transit.md)
- SSH: `SSHSecretEngineConfig` | `SSHSecretEngineRole` | [ssh.md](ssh.md)
- Consul: `ConsulSecretEngineConfig` | `ConsulSecretEngineRole` | [consul.md](consul.md)
- GCP: `GCPSecretEngineConfig` | `GCPSecretEngineRoleset`, `GCPSecretEngineStaticAccount` | [gcp.md](gcp.md)
- LDAP: `LDAPSecretEngineConfig` | `LDAPSecretEngineStaticRole`, `LDAPSecretEngineDynamicRole` | [ldap.md](ldap.md)

### Anti-Patterns / DO NOT

- **DO NOT modify any Go code** — this is pure documentation, no `_types.go`, controller, webhook, or test changes.
- **DO NOT run `make manifests` or `make generate`** — there are no type changes to process.
- **DO NOT create example YAML files in `examples/`** — that is out of scope (example files are a separate concern from docs).
- **DO NOT reference internal implementation details** (like `toMap()`, `IsEquivalentToDesiredState`, `filterPayloadToDesiredKeys`) in user-facing docs — those are developer internals.
- **DO NOT duplicate the full template comments** in output files — use the template as a structure guide, fill in real content.
- **DO NOT add auth-section or connection docs inline** — link to the shared `../auth-section.md` and `../contributing-vault-apis.md` as all existing docs do.
- **DO NOT use snake_case field names in field description tables** — use the camelCase CRD field names (e.g., `accessKey` not `access_key`).

### Project Structure Notes

- All new files go in `docs/secret-engines/` (existing directory, well-established)
- Template to follow: `docs/engine-doc-template.md`
- Quality reference files: `docs/secret-engines/database.md`, `docs/secret-engines/kubernetes.md`
- Index file to update: `docs/secret-engines/index.md`
- No new directories needed

### References

- [Source: docs/engine-doc-template.md] — Template structure and section requirements
- [Source: docs/secret-engines/index.md] — Existing index table format
- [Source: docs/secret-engines/database.md] — Quality bar reference (has Config, Role, StaticRole + credential resolution)
- [Source: _bmad-output/planning-artifacts/epics.md#Story 13.0] — AC and implementation notes
- [Source: _bmad-output/project-context.md#Documentation Gate (DNFR5)] — DNFR5 rule requiring docs for every new CRD type

## Code Review Record

### Review Model Used

GPT-5.4

### Review Findings

- [x] [Review][Patch] Rename the `Role CRD(s)` column in the secret-engine index table so it also fits key and static-account resources [`docs/secret-engines/index.md:36`]
- [x] [Review][Patch] Correct the `SSHSecretEngineConfig` README description so it describes CA configuration instead of an SSH "Connection" [`readme.md:110`]
- [x] [Review][Patch] Add `readme.md` to the story artifact `File List` so the review record matches the actual patch [`_bmad-output/implementation-artifacts/13-0-retroactive-documentation-for-epics-11-12-secret-engines.md:251`]

### Decisions Needed / Decisions Taken

- Iteration 4: No decision-needed findings remain. The previously discussed GCP `randomSecret` scope is treated as a follow-up bug outside this documentation story.

### Fixes Applied

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6

### Debug Log References

No debug issues — documentation-only story with straightforward implementation.

### Completion Notes List

- Created 6 secret engine documentation files following the established template and quality bar (database.md reference):
  - `docs/secret-engines/aws.md` — AWSSecretEngineConfig + AWSSecretEngineRole with credential resolution (Pattern A, rootCredentials)
  - `docs/secret-engines/transit.md` — TransitSecretEngineKey with create-time vs config-time field mutability note (no credential resolution)
  - `docs/secret-engines/ssh.md` — SSHSecretEngineConfig + SSHSecretEngineRole with credential resolution (Pattern A, optional caKeyReference); includes separate OTP and CA examples
  - `docs/secret-engines/consul.md` — ConsulSecretEngineConfig + ConsulSecretEngineRole with credential resolution (Pattern A, rootCredentials)
  - `docs/secret-engines/gcp.md` — GCPSecretEngineConfig + GCPSecretEngineRoleset + GCPSecretEngineStaticAccount with credential resolution (Pattern B, gcpCredentials with passwordKey defaulting to "credentials")
  - `docs/secret-engines/ldap.md` — LDAPSecretEngineConfig + LDAPSecretEngineStaticRole + LDAPSecretEngineDynamicRole with credential resolution (Pattern A, bindCredentials)
- Updated `docs/secret-engines/index.md` — added all 6 engines to the Supported Secret Engines table in alphabetical order (table now has 13 engines total)
- All field descriptions derived directly from CRD type source files using camelCase field names
- All YAML examples include authentication block and engine-specific fields
- All Vault CLI equivalents show the correct Vault API path for each CRD
- IsDeletable behavior documented where relevant (AWS config and Consul config persist after CR deletion)
- Documentation-only story: no Go code changes, no tests, no make manifests/generate required

### Change Log

- 2026-08-12: Created documentation for 6 secret engines (AWS, Transit, SSH, Consul, GCP, LDAP) and updated index — DNFR5 remediation for Epics 11-12

### File List

- `docs/secret-engines/aws.md` (new)
- `docs/secret-engines/transit.md` (new)
- `docs/secret-engines/ssh.md` (new)
- `docs/secret-engines/consul.md` (new)
- `docs/secret-engines/gcp.md` (new)
- `docs/secret-engines/ldap.md` (new)
- `docs/secret-engines/index.md` (modified)
- `readme.md` (modified)
- `_bmad-output/implementation-artifacts/sprint-status.yaml` (modified)
- `_bmad-output/implementation-artifacts/13-0-retroactive-documentation-for-epics-11-12-secret-engines.md` (modified)
