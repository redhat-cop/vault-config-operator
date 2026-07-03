# Story D3.3: Standardize PKI and RabbitMQ Secret Engine Docs

Status: ready-for-dev

## Story

As a user configuring PKI certificates or RabbitMQ dynamic credentials,
I want complete docs with all common fields and examples,
So that I can configure these engines without guesswork.

## Acceptance Criteria

1. **Given** the existing PKI content in `docs/secret-engines.md` lines 466-530 **When** extracted to `docs/secret-engines/pki.md` and standardized per the template **Then** it contains:
   - PKISecretEngineConfig: complete YAML with all common fields (type, privateKeyType, format, keyType, keyBits, maxPathLength, config URLs, CRL config, intermediate CA support), Vault CLI equivalent, Field Descriptions table
   - PKISecretEngineRole: complete YAML with all common role fields (allowedDomains, allowSubdomains, allowBareDomains, keyType, keyBits, keyUsage, extKeyUsage, TTL, maxTTL, useCSRCommonName, useCSRSans), Vault CLI equivalent, Field Descriptions table
   - Intermediate CA section explaining the `externalSignSecret` and `internalSign` workflows

2. **Given** the existing RabbitMQ content in `docs/secret-engines.md` lines 382-464 **When** extracted to `docs/secret-engines/rabbitmq.md` and standardized per the template **Then** it contains:
   - RabbitMQSecretEngineConfig: complete YAML with all fields (connectionURI, username, verifyConnection, passwordPolicy, usernameTemplate, leaseTTL, leaseMaxTTL), Vault CLI equivalent, Field Descriptions table
   - RabbitMQSecretEngineRole: complete YAML with vhost permissions AND vhostTopics examples, Vault CLI equivalent, Field Descriptions table
   - Credential Resolution section (nested `rootCredentials` object with secret/vaultSecret/randomSecret)

3. **Given** cross-references in `readme.md` lines 89-90 pointing to `secret-engines.md#pkisecretengineconfig` and `secret-engines.md#pkisecretenginerole`, and lines 94-95 pointing to `secret-engines.md#rabbitmqsecretengineconfig` and `secret-engines.md#rabbitmqsecretenginerole` **When** the content is moved **Then** those links are updated to point to `secret-engines/pki.md#pkisecretengineconfig`, `secret-engines/pki.md#pkisecretenginerole`, `secret-engines/rabbitmq.md#rabbitmqsecretengineconfig`, and `secret-engines/rabbitmq.md#rabbitmqsecretenginerole`

4. **Given** the new PKI and RabbitMQ doc files **When** validated against the template structure **Then** they follow the same structure as `docs/auth-engines/kubernetes.md` (Overview → Config CRD → Role CRD → Credential Resolution [RabbitMQ only] → See Also)

## Tasks / Subtasks

- [ ] Task 1: Create `docs/secret-engines/pki.md` (AC: 1, 4)
  - [ ] 1.1: Write Overview section — 2-3 sentences explaining the PKI secret engine, link to Vault docs, list the two CRDs (Config, Role)
  - [ ] 1.2: Write PKISecretEngineConfig section with Example YAML (include `type`, `privateKeyType`, `commonName`, `TTL`, `format`, `keyType`, `keyBits`, `maxPathLength`, `issuingCertificates`, `CRLDistributionPoints`, `CRLExpiry`), Vault CLI Equivalent, and Field Descriptions table
  - [ ] 1.3: Write "Intermediate CA" subsection explaining the two intermediate workflows (`externalSignSecret` and `internalSign`)
  - [ ] 1.4: Write PKISecretEngineRole section with Example YAML (include `allowedDomains`, `allowSubdomains`, `allowBareDomains`, `maxTTL`, `keyType`, `keyBits`, `keyUsage`, `extKeyUsage`, `useCSRCommonName`, `useCSRSans`), Vault CLI Equivalent, and Field Descriptions table
  - [ ] 1.5: Add "See Also" section with links to `../auth-section.md`, `../contributing-vault-apis.md`, and Vault PKI docs

- [ ] Task 2: Create `docs/secret-engines/rabbitmq.md` (AC: 2, 4)
  - [ ] 2.1: Write Overview section — 2-3 sentences explaining the RabbitMQ secret engine, link to Vault docs, list the two CRDs (Config, Role)
  - [ ] 2.2: Write RabbitMQSecretEngineConfig section with Example YAML (include `connectionURI`, `username`, `verifyConnection`, `passwordPolicy`, `usernameTemplate`, `leaseTTL`, `leaseMaxTTL`, `rootCredentials`), Vault CLI Equivalent, and Field Descriptions table
  - [ ] 2.3: Write RabbitMQSecretEngineRole section with Example YAML (include `tags`, complete `vhosts` with multiple vhost entries, complete `vhostTopics` with multiple topic entries), Vault CLI Equivalent, and Field Descriptions table
  - [ ] 2.4: Write Credential Resolution section using the nested `rootCredentials` object pattern (3 methods: `rootCredentials.secret`, `rootCredentials.vaultSecret`, `rootCredentials.randomSecret`)
  - [ ] 2.5: Add "See Also" section with links to `../auth-section.md`, `../contributing-vault-apis.md`, and Vault RabbitMQ docs

- [ ] Task 3: Audit field names for camelCase consistency (AC: 1, 2)
  - [ ] 3.1: Cross-reference all PKI field names against Go types (`pkisecretengineconfig_types.go`, `pkisecretenginerole_types.go`) — field names MUST match `json:` tag values exactly
  - [ ] 3.2: Cross-reference all RabbitMQ field names against Go types (`rabbitmqsecretengineconfig_types.go`, `rabbitmqsecretenginerole_types.go`) — field names MUST match `json:` tag values exactly

- [ ] Task 4: Update `readme.md` cross-references (AC: 3)
  - [ ] 4.1: Update line 89 from `./docs/secret-engines.md#pkisecretengineconfig` to `./docs/secret-engines/pki.md#pkisecretengineconfig`
  - [ ] 4.2: Update line 90 from `./docs/secret-engines.md#pkisecretenginerole` to `./docs/secret-engines/pki.md#pkisecretenginerole`
  - [ ] 4.3: Update line 94 from `./docs/secret-engines.md#rabbitmqsecretengineconfig` to `./docs/secret-engines/rabbitmq.md#rabbitmqsecretengineconfig`
  - [ ] 4.4: Update line 95 from `./docs/secret-engines.md#rabbitmqsecretenginerole` to `./docs/secret-engines/rabbitmq.md#rabbitmqsecretenginerole`

- [ ] Task 5: Verify links and structure (AC: 4)
  - [ ] 5.1: Verify relative links resolve correctly from `docs/secret-engines/pki.md` and `docs/secret-engines/rabbitmq.md` (`../auth-section.md`, `../contributing-vault-apis.md`)
  - [ ] 5.2: Verify structure matches `kubernetes.md` pattern: heading hierarchy, section ordering, table format

## Dev Notes

### This Is a Documentation-Only Story

No Go code changes. No tests to run. No `make manifests generate`. The deliverables are:
- 2 new files: `docs/secret-engines/pki.md`, `docs/secret-engines/rabbitmq.md`
- 1 modified file: `readme.md` (update 4 cross-reference links)

### Dependency on D3.1

This story assumes D3.1 has been completed (creating `docs/secret-engines/index.md` and the `docs/secret-engines/` directory). If D3.1 is NOT yet done, create the directory if it doesn't exist. The index references `pki.md` and `rabbitmq.md` via `[pki.md](pki.md)` and `[rabbitmq.md](rabbitmq.md)`.

### Source Content Location

The content to extract and standardize:
- PKI: `docs/secret-engines.md` lines 466-530 (`## PKISecretEngineConfig` and `## PKISecretEngineRole`)
- RabbitMQ: `docs/secret-engines.md` lines 382-464 (`## RabbitMQSecretEngineConfig` and `## RabbitMQSecretEngineRole`)

### Template to Follow

Use `docs/engine-doc-template.md` as the structural pattern. Use `docs/auth-engines/kubernetes.md` as the concrete reference implementation.

Key structural requirements from the template:
1. Title: `# PKI Secret Engine` / `# RabbitMQ Secret Engine`
2. Link to Vault docs immediately below title
3. `## Overview` — 2-3 sentences + CRD list
4. `## PKISecretEngineConfig` / `## RabbitMQSecretEngineConfig` → `### Example` → `### Vault CLI Equivalent` → `### Field Descriptions`
5. `## PKISecretEngineRole` / `## RabbitMQSecretEngineRole` → `### Example` → `### Vault CLI Equivalent` → `### Field Descriptions`
6. `## Credential Resolution` (RabbitMQ only — PKI has no external credentials)
7. `## See Also`

### PKISecretEngineConfig — Complete Field Reference

From `api/v1alpha1/pkisecretengineconfig_types.go`:

**PKIType struct:**

| CRD Field (camelCase) | JSON tag | Vault API key | Type | Required | Default |
|---|---|---|---|---|---|
| type | `json:"type"` | (URL path segment) | string | Yes | `"root"` (Enum: `root`, `intermediate`) |
| privateKeyType | `json:"privateKeyType"` | (URL path segment) | string | Yes | `"internal"` (Enum: `internal`, `exported`) |

**PKICommon struct (used for generate payload):**

| CRD Field (camelCase) | JSON tag | Vault API key | Type | Required | Default |
|---|---|---|---|---|---|
| commonName | `json:"commonName,omitempty"` | `common_name` | string | Yes | — |
| altNames | `json:"altNames,omitempty"` | `alt_names` | string | No | — |
| IPSans | `json:"IPSans,omitempty"` | `ip_sans` | string | No | — |
| URISans | `json:"URISans,omitempty"` | `uri_sans` | string | No | — |
| otherSans | `json:"otherSans,omitempty"` | `other_sans` | string | No | — |
| TTL | `json:"TTL,omitempty"` | `ttl` | duration | No | — |
| format | `json:"format"` | `format` | string | No | `"pem"` (Enum: `pem`, `pem_bundle`, `der`) |
| privateKeyFormat | `json:"privateKeyFormat,omitempty"` | `private_key_format` | string | No | — (Enum: `der`, `pkcs8`) |
| keyType | `json:"keyType"` | `key_type` | string | No | `"rsa"` (Enum: `rsa`, `ec`) |
| keyBits | `json:"keyBits"` | `key_bits` | int | No | `2048` |
| maxPathLength | `json:"maxPathLength"` | `max_path_length` | int | No | `-1` (Min: -1) |
| excludeCnFromSans | `json:"excludeCnFromSans,omitempty"` | `exclude_cn_from_sans` | bool | No | false |
| permittedDnsDomains | `json:"permittedDnsDomains,omitempty"` | `permitted_dns_domains` | []string | No | — |
| ou | `json:"ou,omitempty"` | `ou` | string | No | — |
| organization | `json:"organization,omitempty"` | `organization` | string | No | — |
| country | `json:"country,omitempty"` | `country` | string | No | — |
| locality | `json:"locality,omitempty"` | `locality` | string | No | — |
| province | `json:"province,omitempty"` | `province` | string | No | — |
| streetAddress | `json:"streetAddress,omitempty"` | `street_address` | string | No | — |
| postalCode | `json:"postalCode,omitempty"` | `postal_code` | string | No | — |
| serialNumber | `json:"serialNumber,omitempty"` | `serial_number` | string | No | — |

**PKIConfigUrls struct (written to `{path}/config/urls`):**

| CRD Field (camelCase) | JSON tag | Vault API key | Type | Required | Default |
|---|---|---|---|---|---|
| issuingCertificates | `json:"issuingCertificates,omitempty"` | `issuing_certificates` | []string | No | — |
| CRLDistributionPoints | `json:"CRLDistributionPoints,omitempty"` | `crl_distribution_points` | []string | No | — |
| ocspServers | `json:"ocspServers,omitempty"` | `ocsp_servers` | []string | No | — |

**PKIConfigCRL struct (written to `{path}/config/crl`):**

| CRD Field (camelCase) | JSON tag | Vault API key | Type | Required | Default |
|---|---|---|---|---|---|
| CRLExpiry | `json:"CRLExpiry"` | `expiry` | duration | No | `"72h"` |
| CRLDisable | `json:"CRLDisable,omitempty"` | `disable` | bool | No | false |

**PKIIntermediate struct:**

| CRD Field (camelCase) | JSON tag | Vault API key | Type | Required | Default |
|---|---|---|---|---|---|
| externalSignSecret | `json:"externalSignSecret,omitempty"` | — (operator-side) | LocalObjectReference | No | — |
| certificateKey | `json:"certificateKey"` | — (operator-side) | string | No | `"tls.crt"` |
| internalSign | `json:"internalSign,omitempty"` | — (operator-side) | LocalObjectReference | No | — |

Additional top-level spec fields:
- `path` (Required) — path of the secret engine mount. Final Vault paths:
  - Generate: `{path}/{type}/generate/{privateKeyType}` (e.g., `pki/root/generate/internal`)
  - Config URLs: `{path}/config/urls`
  - Config CRL: `{path}/config/crl`
- `authentication` (Required) — see shared auth-section.md
- `connection` (Optional) — override Vault connection

**Important — PKI has a unique reconciler (`VaultPKIEngineResource`):**
- The reconcile flow writes to THREE separate Vault paths: generate (root/intermediate CA), config/urls, and config/crl
- `IsDeletable() == false` — deleting the CR does NOT remove the PKI root/intermediate from Vault
- `IsEquivalentToDesiredState` only checks `PKICommon.toMap()` (the generate payload), not URLs or CRL config
- Intermediate CA mode requires either `externalSignSecret` (externally-signed CSR) or `internalSign` (signed by another PKISecretEngineConfig)

**Important — Intermediate CA Workflow:**
1. When `type: intermediate`, the reconciler generates a CSR (not a root CA)
2. If `privateKeyType: exported`, the CSR is stored in a Kubernetes Secret (same namespace, same name as CR)
3. If `internalSign` is set, the operator uses the referenced PKISecretEngineConfig to sign the CSR (writes to `{internalSign.name}/root/sign-intermediate`)
4. If `externalSignSecret` is set, the operator reads the signed certificate from the referenced Kubernetes Secret (key: `certificateKey`, default: `tls.crt`)
5. The signed certificate is then written to `{path}/intermediate/set-signed`

### PKISecretEngineRole — Complete Field Reference

From `api/v1alpha1/pkisecretenginerole_types.go`, the `PKIRole` struct:

| CRD Field (camelCase) | JSON tag | Vault API key | Type | Required | Default |
|---|---|---|---|---|---|
| TTL | `json:"TTL,omitempty"` | `ttl` | duration | No | — |
| maxTTL | `json:"maxTTL,omitempty"` | `max_ttl` | duration | No | — |
| allowLocalhost | `json:"allowLocalhost,omitempty"` | `allow_localhost` | bool | No | false |
| allowedDomains | `json:"allowedDomains,omitempty"` | `allowed_domains` | []string | No | — |
| allowedDomainsTemplate | `json:"allowedDomainsTemplate,omitempty"` | `allowed_domains_template` | bool | No | false |
| allowBareDomains | `json:"allowBareDomains,omitempty"` | `allow_bare_domains` | bool | No | false |
| allowSubdomains | `json:"allowSubdomains,omitempty"` | `allow_subdomains` | bool | No | false |
| allowGlobDomains | `json:"allowGlobDomains,omitempty"` | `allow_glob_domains` | bool | No | false |
| allowAnyName | `json:"allowAnyName,omitempty"` | `allow_any_name` | bool | No | false |
| enforceHostnames | `json:"enforceHostnames,omitempty"` | `enforce_hostnames` | bool | No | false |
| allowIPSans | `json:"allowIPSans,omitempty"` | `allow_ip_sans` | bool | No | false |
| allowedURISans | `json:"allowedURISans,omitempty"` | `allowed_uri_sans` | []string | No | — |
| allowedOtherSans | `json:"allowedOtherSans,omitempty"` | `allowed_other_sans` | string | No | — |
| serverFlag | `json:"serverFlag,omitempty"` | `server_flag` | bool | No | false |
| clientFlag | `json:"clientFlag,omitempty"` | `client_flag` | bool | No | false |
| codeSigningFlag | `json:"codeSigningFlag,omitempty"` | `code_signing_flag` | bool | No | false |
| emailProtectionFlag | `json:"emailProtectionFlag,omitempty"` | `email_protection_flag` | bool | No | false |
| keyType | `json:"keyType"` | `key_type` | string | No | `"rsa"` (Enum: `rsa`, `ec`, `any`) |
| keyBits | `json:"keyBits"` | `key_bits` | int | No | `2048` |
| keyUsage | `json:"keyUsage,omitempty"` | `key_usage` | []KeyUsage | No | — (Enum values: DigitalSignature, KeyAgreement, KeyEncipherment, ContentCommitment, DataEncipherment, CertSign, CRLSign, EncipherOnly, DecipherOnly) |
| extKeyUsage | `json:"extKeyUsage,omitempty"` | `ext_key_usage` | []ExtKeyUsage | No | — (Enum values: ServerAuth, ClientAuth, CodeSigning, EmailProtection, IPSECEndSystem, IPSECTunnel, IPSECUser, TimeStamping, OCSPSigning, MicrosoftServerGatedCrypto, NetscapeServerGatedCrypto, MicrosoftCommercialCodeSigning, MicrosoftKernelCodeSigning) |
| extKeyUsageOids | `json:"extKeyUsageOids,omitempty"` | `ext_key_usage_oids` | []string | No | — |
| useCSRCommonName | `json:"useCSRCommonName"` | `use_csr_common_name` | bool | No | `true` |
| useCSRSans | `json:"useCSRSans"` | `use_csr_sans` | bool | No | `true` |
| ou | `json:"ou,omitempty"` | `ou` | string | No | — |
| organization | `json:"organization,omitempty"` | `organization` | string | No | — |
| country | `json:"country,omitempty"` | `country` | string | No | — |
| locality | `json:"locality,omitempty"` | `locality` | string | No | — |
| province | `json:"province,omitempty"` | `province` | string | No | — |
| streetAddress | `json:"streetAddress,omitempty"` | `street_address` | string | No | — |
| postalCode | `json:"postalCode,omitempty"` | `postal_code` | string | No | — |
| serialNumber | `json:"serialNumber,omitempty"` | `serial_number` | string | No | — |
| generateLease | `json:"generateLease,omitempty"` | `generate_lease` | bool | No | false |
| noStore | `json:"noStore,omitempty"` | `no_store` | bool | No | false |
| requireCn | `json:"requireCn,omitempty"` | `require_cn` | bool | No | false |
| policyIdentifiers | `json:"policyIdentifiers,omitempty"` | `policy_identifiers` | []string | No | — |
| basicConstraintsValidForNonCa | `json:"basicConstraintsValidForNonCa,omitempty"` | `basic_constraints_valid_for_non_ca` | bool | No | false |
| notBeforeDuration | `json:"notBeforeDuration"` | `not_before_duration` | duration | No | `"30s"` |

Additional top-level spec fields:
- `path` (Required) — path of the secret engine mount. Final Vault path: `[namespace/]{path}/roles/{name}`
- `authentication` (Required) — see shared auth-section.md
- `connection` (Optional) — override Vault connection
- `name` (Optional) — override Vault object name (defaults to `metadata.name`)

**Important:** `IsDeletable() == true` — deleting the CR removes the role from Vault.

### RabbitMQSecretEngineConfig — Complete Field Reference

From `api/v1alpha1/rabbitmqsecretengineconfig_types.go`, the `RMQSEConfig` struct:

| CRD Field (camelCase) | JSON tag | Vault API key | Type | Required | Default |
|---|---|---|---|---|---|
| connectionURI | `json:"connectionURI,omitempty"` | `connection_uri` | string | Yes | — (Pattern: `^(http|https):\/\/.+$`) |
| username | `json:"username,omitempty"` | `username` | string | No | — (from credential source if not set) |
| verifyConnection | `json:"verifyConnection,omitempty"` | `verify_connection` | bool | No | false (Vault default is true) |
| passwordPolicy | `json:"passwordPolicy,omitempty"` | `password_policy` | string | No | — |
| usernameTemplate | `json:"usernameTemplate,omitempty"` | `username_template` | string | No | — |
| leaseTTL | `json:"leaseTTL,omitempty"` | `ttl` | int (seconds) | No | — |
| leaseMaxTTL | `json:"leaseMaxTTL,omitempty"` | `max_ttl` | int (seconds) | No | — |

Additional top-level spec fields:
- `path` (Required) — path of the secret engine mount. Final Vault path: `[namespace/]{path}/config/connection`
- `authentication` (Required) — see shared auth-section.md
- `connection` (Optional) — override Vault connection
- `rootCredentials` (Required) — credential resolution. See Credential Resolution section

**Important — RabbitMQ uses a dual-write reconciler pattern:**
- The `GetPayload()` method calls `rabbitMQToMap()` which produces the connection config (written to `{path}/config/connection`)
- The lease config (leaseTTL, leaseMaxTTL) is written SEPARATELY to `{path}/config/lease` via `GetLeasePayload()` / `GetLeasePath()`
- `IsEquivalentToDesiredState` checks ONLY `leasesToMap()` (ttl and max_ttl), NOT the connection config. This is because the connection config contains credentials (password) that Vault never returns on read
- `CheckTTLValuesProvided()` returns true only if leaseTTL or leaseMaxTTL is non-zero — if neither is set, the lease write is skipped
- `IsDeletable() == false` — deleting the CR does NOT remove the RabbitMQ connection from Vault

**Important — Credential field `username`:** If `spec.username` is provided, it takes precedence over the username from the credential source. If not provided, the username is retrieved from the credential source.

### RabbitMQSecretEngineRole — Complete Field Reference

From `api/v1alpha1/rabbitmqsecretenginerole_types.go`, the `RMQSERole` struct:

| CRD Field (camelCase) | JSON tag | Vault API key | Type | Required | Default |
|---|---|---|---|---|---|
| tags | `json:"tags,omitempty"` | `tags` | string | No | — (comma-separated RabbitMQ management tags) |
| vhosts | `json:"vhosts,omitempty"` | `vhosts` | []Vhost | No | — |
| vhostTopics | `json:"vhostTopics,omitempty"` | `vhost_topics` | []VhostTopic | No | — (requires RabbitMQ 3.7.0+) |

**Vhost struct:**

| CRD Field | JSON tag | Type | Required | Description |
|---|---|---|---|---|
| vhostName | `json:"vhostName,omitempty"` | string | Yes | Name of an existing vhost |
| permissions | `json:"permissions,omitempty"` | VhostPermissions | Yes | Permissions for this vhost |

**VhostTopic struct:**

| CRD Field | JSON tag | Type | Required | Description |
|---|---|---|---|---|
| vhostName | `json:"vhostName,omitempty"` | string | Yes | Name of an existing vhost |
| topics | `json:"topics,omitempty"` | []Topic | Yes | List of topic permissions |

**Topic struct:**

| CRD Field | JSON tag | Type | Required | Description |
|---|---|---|---|---|
| topicName | `json:"topicName,omitempty"` | string | Yes | Name of an existing topic/exchange |
| permissions | `json:"permissions,omitempty"` | VhostPermissions | Yes | Permissions for this topic |

**VhostPermissions struct:**

| CRD Field | JSON tag | Type | Required | Description |
|---|---|---|---|---|
| configure | `json:"configure,omitempty"` | string | No | Regex pattern for configure permission |
| write | `json:"write,omitempty"` | string | No | Regex pattern for write permission |
| read | `json:"read,omitempty"` | string | No | Regex pattern for read permission |

Additional top-level spec fields:
- `path` (Required) — path of the secret engine mount. Final Vault path: `[namespace/]{path}/roles/{name}`
- `authentication` (Required) — see shared auth-section.md
- `connection` (Optional) — override Vault connection
- `name` (Optional) — override Vault object name (defaults to `metadata.name`)

**Important:** `IsDeletable() == true` — deleting the CR removes the role from Vault.

**Important — Payload serialization:** The `rabbitMQToMap()` method converts vhosts and vhostTopics to JSON strings before writing to Vault. The Vault API expects `vhosts` and `vhost_topics` as JSON-encoded strings, not nested objects.

### Credential Resolution (RabbitMQ)

RabbitMQ uses a nested `rootCredentials` object (similar to Database). The struct is `RootCredentialConfig` with:
- `rootCredentials.secret.name` — reference a Kubernetes Secret
- `rootCredentials.vaultSecret.path` — reference a Vault KV secret
- `rootCredentials.randomSecret.name` — reference a RandomSecret CR
- `rootCredentials.passwordKey` — key for password retrieval (default: `"password"`)
- `rootCredentials.usernameKey` — key for username retrieval (default: `"username"`)

The Kubernetes Secret for `rootCredentials.secret` is NOT basic-auth type for RabbitMQ — it uses `passwordKey` and `usernameKey` to specify custom keys in any secret type.

**Important:** When using `randomSecret`, a username MUST be specified in `spec.username` because RandomSecret only generates a single value (the password).

### PKI Does NOT Have Credential Resolution

Unlike Database and RabbitMQ, the PKI secret engine does not require external credentials (no `rootCredentials` field). PKI generates its own keys internally. Do NOT add a Credential Resolution section to `pki.md`.

### Known Issues in Source Content

From the existing `docs/secret-engines.md`:
1. PKI docs only show `commonName` and `TTL` — missing all other PKICommon fields, PKIConfigUrls, PKIConfigCRL, PKIIntermediate
2. PKI Role docs only show `allowedDomains` and `maxTTL` — missing 30+ other role fields
3. RabbitMQ Config docs say "rootCredentialsFromSecret" (flat pattern wording) but the CRD uses nested `rootCredentials.secret` — use the correct nested pattern in new docs
4. RabbitMQ Role docs don't show a Vault CLI equivalent — add one
5. Neither PKI nor RabbitMQ docs have Field Descriptions tables — add per template

### readme.md Cross-References

Lines requiring update:

| Line | Current Link | New Link |
|------|-------------|----------|
| 89 | `./docs/secret-engines.md#pkisecretengineconfig` | `./docs/secret-engines/pki.md#pkisecretengineconfig` |
| 90 | `./docs/secret-engines.md#pkisecretenginerole` | `./docs/secret-engines/pki.md#pkisecretenginerole` |
| 94 | `./docs/secret-engines.md#rabbitmqsecretengineconfig` | `./docs/secret-engines/rabbitmq.md#rabbitmqsecretengineconfig` |
| 95 | `./docs/secret-engines.md#rabbitmqsecretenginerole` | `./docs/secret-engines/rabbitmq.md#rabbitmqsecretenginerole` |

### Relative Link Conventions

From `docs/secret-engines/pki.md` and `docs/secret-engines/rabbitmq.md`:
- To shared docs: `../auth-section.md`, `../contributing-vault-apis.md`
- To secret management (for RandomSecret reference): `../secret-management.md#randomsecret`
- To other engine files: `database.md`, `rabbitmq.md` (same directory)
- External: full URLs to Vault documentation

### Vault Path Structure

**PKI:**
```
{path}/{type}/generate/{privateKeyType}  ← PKISecretEngineConfig (generate root/intermediate)
{path}/config/urls                        ← PKISecretEngineConfig (URL config)
{path}/config/crl                         ← PKISecretEngineConfig (CRL config)
{path}/roles/{role-name}                  ← PKISecretEngineRole
{path}/root/sign-intermediate             ← Used by internalSign workflow
{path}/intermediate/set-signed            ← Set signed intermediate certificate
```

Example with `path: pki-vault-demo/pki`:
- Generate Root: `pki-vault-demo/pki/root/generate/internal`
- Config URLs: `pki-vault-demo/pki/config/urls`
- Config CRL: `pki-vault-demo/pki/config/crl`
- Role: `pki-vault-demo/pki/roles/my-role`

**RabbitMQ:**
```
{path}/config/connection      ← RabbitMQSecretEngineConfig (connection)
{path}/config/lease           ← RabbitMQSecretEngineConfig (lease TTLs)
{path}/roles/{role-name}      ← RabbitMQSecretEngineRole
```

Example with `path: vault-config-operator/rabbitmq`:
- Connection: `vault-config-operator/rabbitmq/config/connection`
- Lease: `vault-config-operator/rabbitmq/config/lease`
- Role: `vault-config-operator/rabbitmq/roles/my-role`

### Previous Story Intelligence

**From D3.2 (Database Secret Engine Docs — the direct predecessor):**
- Established the extraction pattern for secret engines: source content → template structure → field audit → CLI equivalents
- Field descriptions table uses camelCase names from JSON tags
- Vault CLI equivalents use snake_case names
- Three-CRD pattern (Config + Role + StaticRole) was established — PKI/RabbitMQ only have Config + Role
- Credential Resolution section format validated with nested `rootCredentials` object
- Path structure documentation pattern established

**From D3.1 (Secret-Engines Directory Structure & Index):**
- Created `docs/secret-engines/index.md` with overview, SecretEngineMount section, engine table
- Index page engine table has `pki.md` and `rabbitmq.md` links ready for this story
- Documented all readme.md cross-references for downstream stories (lines 89-90 for PKI, 94-95 for RabbitMQ)
- Directory `docs/secret-engines/` created

**From D2 Retrospective:**
- Template proven across 6 auth engine docs — applies directly to secret engines
- Zero review findings on D2.5 — team fully internalized template
- Recommendation: Continue using Opus 4.6 for all stories

### PKI YAML Example (Comprehensive)

Include this complete YAML in the doc (not the minimal 3-field example from the old docs):

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: PKISecretEngineConfig
metadata:
  name: my-pki
spec:
  authentication:
    path: kubernetes
    role: pki-engine-admin
  path: pki-vault-demo/pki
  type: root
  privateKeyType: internal
  commonName: pki-vault-demo.internal.io
  TTL: "8760h"
  format: pem
  keyType: rsa
  keyBits: 4096
  maxPathLength: -1
  organization: "My Company"
  country: "US"
  province: "California"
  locality: "San Francisco"
  issuingCertificates:
    - "https://vault.example.com/v1/pki-vault-demo/pki/ca"
  CRLDistributionPoints:
    - "https://vault.example.com/v1/pki-vault-demo/pki/crl"
  CRLExpiry: "72h"
```

### PKI Role YAML Example (Comprehensive)

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: PKISecretEngineRole
metadata:
  name: web-server
spec:
  authentication:
    path: kubernetes
    role: pki-engine-admin
  path: pki-vault-demo/pki
  allowedDomains:
    - internal.io
    - pki-vault-demo.svc
  allowSubdomains: true
  allowBareDomains: false
  maxTTL: "720h"
  keyType: rsa
  keyBits: 2048
  keyUsage:
    - DigitalSignature
    - KeyEncipherment
  extKeyUsage:
    - ServerAuth
  useCSRCommonName: true
  useCSRSans: true
  requireCn: true
  notBeforeDuration: "30s"
```

### RabbitMQ Config YAML Example (Comprehensive)

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: RabbitMQSecretEngineConfig
metadata:
  name: my-rabbitmq-config
spec:
  authentication:
    path: kubernetes
    role: rabbitmq-engine-admin
  path: vault-config-operator/rabbitmq
  connectionURI: https://my-rabbitmq.example.com:15672
  username: admin
  verifyConnection: true
  passwordPolicy: my-password-policy
  usernameTemplate: "v-{{.RoleName}}-{{unix_time}}"
  leaseTTL: 86400
  leaseMaxTTL: 172800
  rootCredentials:
    secret:
      name: rabbitmq-admin-password
    passwordKey: password
    usernameKey: username
```

### RabbitMQ Role YAML Example (Comprehensive with Topics)

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: RabbitMQSecretEngineRole
metadata:
  name: my-rabbitmq-role
spec:
  authentication:
    path: kubernetes
    role: rabbitmq-engine-admin
  path: vault-config-operator/rabbitmq
  tags: "administrator"
  vhosts:
    - vhostName: "/"
      permissions:
        read: ".*"
        write: ".*"
        configure: ".*"
    - vhostName: "my-vhost"
      permissions:
        read: "my-queue"
        write: "my-exchange"
        configure: ""
  vhostTopics:
    - vhostName: "/"
      topics:
        - topicName: "my-topic"
          permissions:
            read: ".*"
            write: ".*"
        - topicName: "audit-topic"
          permissions:
            read: ".*"
            write: ""
```

### Project Structure Notes

```
docs/
├── secret-engines/
│   ├── index.md          ← D3.1
│   ├── database.md       ← D3.2
│   ├── pki.md            ← NEW (this story)
│   └── rabbitmq.md       ← NEW (this story)
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

- [Source: _bmad-output/planning-artifacts/epics.md#Story D3.3] — Story requirements and acceptance criteria
- [Source: docs/secret-engines.md:466-530] — PKI content to extract and standardize
- [Source: docs/secret-engines.md:382-464] — RabbitMQ content to extract and standardize
- [Source: docs/auth-engines/kubernetes.md] — Reference implementation for template pattern (D2.2)
- [Source: docs/engine-doc-template.md] — Template structure (D1.1, review-patched)
- [Source: api/v1alpha1/pkisecretengineconfig_types.go] — CRD field definitions for PKI Config
- [Source: api/v1alpha1/pkisecretenginerole_types.go] — CRD field definitions for PKI Role
- [Source: api/v1alpha1/rabbitmqsecretengineconfig_types.go] — CRD field definitions for RabbitMQ Config
- [Source: api/v1alpha1/rabbitmqsecretenginerole_types.go] — CRD field definitions for RabbitMQ Role
- [Source: api/v1alpha1/utils/commons.go:366-396] — RootCredentialConfig struct definition
- [Source: _bmad-output/implementation-artifacts/d3-2-standardize-database-secret-engine-docs.md] — Previous story context
- [Source: _bmad-output/implementation-artifacts/d3-1-create-secret-engines-directory-structure-and-index-page.md] — Directory structure story
- [Source: readme.md:89-90,94-95] — Cross-references that need updating
- [Source: _bmad-output/project-context.md] — Project conventions and coding standards

## Dev Agent Record

### Agent Model Used

{{agent_model_name_version}}

### Debug Log References

### Completion Notes List

### File List
