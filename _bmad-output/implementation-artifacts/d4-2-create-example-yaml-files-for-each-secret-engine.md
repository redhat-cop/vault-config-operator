# Story D4.2: Create Example YAML Files for Each Secret Engine

Status: done

## Story

As a user learning the operator,
I want ready-to-use example YAML files for each secret engine,
so that I can quickly bootstrap my configuration.

## Acceptance Criteria

1. **Given** the existing `docs/examples/postgresql/` as a reference, **when** example directories are created for each secret engine, **then** the following directories exist with complete, valid example CRs:
   - `docs/examples/secret-database/` — SecretEngineMount + DatabaseSecretEngineConfig + DatabaseSecretEngineRole + DatabaseSecretEngineStaticRole
   - `docs/examples/secret-pki/` — SecretEngineMount + PKISecretEngineConfig + PKISecretEngineRole
   - `docs/examples/secret-rabbitmq/` — SecretEngineMount + RabbitMQSecretEngineConfig + RabbitMQSecretEngineRole
   - `docs/examples/secret-github/` — SecretEngineMount + GitHubSecretEngineConfig + GitHubSecretEngineRole
   - `docs/examples/secret-quay/` — SecretEngineMount + QuaySecretEngineConfig + QuaySecretEngineRole + QuaySecretEngineStaticRole
   - `docs/examples/secret-kubernetes/` — SecretEngineMount + KubernetesSecretEngineConfig + KubernetesSecretEngineRole
   - `docs/examples/secret-azure/` — SecretEngineMount + AzureSecretEngineConfig + AzureSecretEngineRole

2. **Given** each example directory, **when** the YAML files are validated, **then** all examples use the correct `apiVersion: redhatcop.redhat.io/v1alpha1`, include required fields, and contain helpful inline comments explaining each field's purpose.

3. **Given** the existing `docs/examples/postgresql/` as the reference pattern, **when** examples are reviewed, **then** they follow the same multi-document YAML pattern with resources ordered by dependency (mount → config → role).

## Tasks / Subtasks

- [x] Task 1: Create `docs/examples/secret-database/` directory (AC: #1, #2, #3)
  - [x] Create `secret-database.yaml` with SecretEngineMount + DatabaseSecretEngineConfig + DatabaseSecretEngineRole + DatabaseSecretEngineStaticRole
- [x] Task 2: Create `docs/examples/secret-pki/` directory (AC: #1, #2, #3)
  - [x] Create `secret-pki.yaml` with SecretEngineMount + PKISecretEngineConfig (root CA) + PKISecretEngineRole
- [x] Task 3: Create `docs/examples/secret-rabbitmq/` directory (AC: #1, #2, #3)
  - [x] Create `secret-rabbitmq.yaml` with SecretEngineMount + RabbitMQSecretEngineConfig + RabbitMQSecretEngineRole
- [x] Task 4: Create `docs/examples/secret-github/` directory (AC: #1, #2, #3)
  - [x] Create `secret-github.yaml` with SecretEngineMount + GitHubSecretEngineConfig + GitHubSecretEngineRole
- [x] Task 5: Create `docs/examples/secret-quay/` directory (AC: #1, #2, #3)
  - [x] Create `secret-quay.yaml` with SecretEngineMount + QuaySecretEngineConfig + QuaySecretEngineRole + QuaySecretEngineStaticRole
- [x] Task 6: Create `docs/examples/secret-kubernetes/` directory (AC: #1, #2, #3)
  - [x] Create `secret-kubernetes.yaml` with SecretEngineMount + KubernetesSecretEngineConfig + KubernetesSecretEngineRole
- [x] Task 7: Create `docs/examples/secret-azure/` directory (AC: #1, #2, #3)
  - [x] Create `secret-azure.yaml` with SecretEngineMount + AzureSecretEngineConfig + AzureSecretEngineRole

## Dev Notes

### Relationship to Existing `docs/examples/postgresql/`

The epic mentions "rename/move existing postgresql" to `secret-database/`. **Do NOT move or rename** the existing `docs/examples/postgresql/` directory — it may be referenced from the README, external guides, or user bookmarks. Instead, create `docs/examples/secret-database/` as a **new, standalone** example that is more comprehensive (includes StaticRole) and better commented. The existing `postgresql/` directory remains untouched.

### Example File Pattern

Follow the existing `docs/examples/postgresql/postgresql-secret-engine.yaml` and the D4.1 auth engine example patterns:

- Multi-document YAML (resources separated by `---`)
- Resources ordered by dependency: SecretEngineMount → Config → Role(s) → StaticRole (if applicable)
- No `namespace` in metadata (user decides deployment namespace)
- Use realistic but clearly-placeholder values (e.g., `my-project`, `example.com`, `00000000-...`)
- Include inline YAML comments (`#`) explaining what each field does and when to change it
- One YAML file per engine directory (matching directory name: `secret-<engine>.yaml`)

### apiVersion and Kind for All CRDs

All CRDs use:
```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
```

### Authentication Block Pattern

Every CRD includes an `authentication` block — this is how the **operator** authenticates to Vault to perform the operation. Use this standard pattern in all examples:
```yaml
spec:
  authentication:
    path: kubernetes
    role: policy-admin
```

The `spec.path` field (separate from `spec.authentication.path`) is the Vault mount path of the **engine being configured**.

### SecretEngineMount Block Pattern

Every secret engine example starts with a SecretEngineMount. Use this pattern:
```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: SecretEngineMount
metadata:
  name: <engine-name>
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  type: <engine-type>
  path: <mount-path>
```

The `type` field values per engine:
- Database: `database`
- PKI: `pki`
- RabbitMQ: `rabbitmq`
- GitHub: `github` (third-party plugin: vault-plugin-secrets-github)
- Quay: `quay` (third-party plugin: vault-plugin-secrets-quay)
- Kubernetes: `kubernetes`
- Azure: `azure`

### Per-Engine CRD Details

#### secret-database/

**SecretEngineMount** — `type: database`, `path: database-demo`

**DatabaseSecretEngineConfig** — Configures the database connection:
- Required: `path`, `pluginName`, `connectionURL`, `rootCredentials`
- Show `pluginName: postgresql-database-plugin` (most common)
- Show `allowedRoles` with specific role names (not `["*"]`)
- `connectionURL` uses `{{username}}` and `{{password}}` placeholders
- `rootCredentials` uses Pattern A (Kubernetes Secret): `rootCredentials: { secret: { name: <secret-name> } }`
- Optional: `passwordAuthentication: scram-sha-256`, `rootPasswordRotation`

**DatabaseSecretEngineRole** — Dynamic credential role:
- Required: `path`, `dBName` (references Config name)
- Key: `defaultTTL`, `maxTTL`, `creationStatements`, `revocationStatements`
- `creationStatements` use Vault template variables: `{{name}}`, `{{password}}`, `{{expiration}}`

**DatabaseSecretEngineStaticRole** — Manages pre-existing user password rotation:
- Required: `path`, `dBName`, `username`, `rotationPeriod`, `credentialType`
- Show `credentialType: password` with `passwordCredentialConfig`
- `rotationPeriod` is in seconds (integer), not a duration string

#### secret-pki/

**SecretEngineMount** — `type: pki`, `path: pki-demo`

**PKISecretEngineConfig** — Configures root or intermediate CA:
- Required: `path`, `type` (root/intermediate), `privateKeyType` (internal/exported), `commonName`
- Show root CA: `type: root`, `privateKeyType: internal`
- Include `TTL`, `format`, `keyType`, `keyBits`, `maxPathLength`
- Include URL config: `issuingCertificates`, `CRLDistributionPoints`
- Include CRL config: `CRLExpiry`
- Include subject fields: `organization`, `country`, `province`, `locality`

**PKISecretEngineRole** — Certificate issuance role:
- Required: `path`
- Key: `allowedDomains`, `allowSubdomains`, `allowBareDomains`
- Include: `TTL`, `maxTTL`, `keyType`, `keyBits`
- Include: `keyUsage` (list of KeyUsage values), `extKeyUsage` (list of ExtKeyUsage values)
- Include: `useCSRCommonName`, `useCSRSans`, `requireCn`, `notBeforeDuration`

#### secret-rabbitmq/

**SecretEngineMount** — `type: rabbitmq`, `path: rabbitmq-demo`

**RabbitMQSecretEngineConfig** — Configures RabbitMQ connection:
- Required: `path`, `connectionURI`, `rootCredentials`
- `connectionURI` is the RabbitMQ management API URL (e.g., `https://rabbitmq.example.com:15672`)
- `rootCredentials` uses Pattern A (Kubernetes Secret) with `usernameKey` and `passwordKey`
- Optional: `verifyConnection`, `passwordPolicy`, `usernameTemplate`, `leaseTTL`, `leaseMaxTTL`
- Note: operator writes to two Vault paths: `{path}/config/connection` and `{path}/config/lease`

**RabbitMQSecretEngineRole** — Dynamic credential role:
- Required: `path`
- Key: `tags` (comma-separated string, e.g., `"administrator"`)
- `vhosts` is a list of objects with `vhostName` and `permissions` (configure/write/read regex patterns)
- `vhostTopics` for topic-level permissions (optional)
- Show single-vhost example (the multi-entry case has a known serialization bug — see D3 retro)

#### secret-github/

**SecretEngineMount** — `type: github`, `path: github-demo` (requires vault-plugin-secrets-github)

**GitHubSecretEngineConfig** — Configures GitHub App connection:
- Required: `path`, `applicationID`, `sSHKeyReference`
- `applicationID` is the GitHub App's numeric Application ID (int64, minimum 1)
- `sSHKeyReference` uses a Kubernetes Secret of type `kubernetes.io/ssh-auth` with the private key in `ssh-privatekey` field
- Optional: `gitHubAPIBaseURL` (defaults to `https://api.github.com`)
- Note: field name is `sSHKeyReference` (double-capital S), not `sshKeyReference`

**GitHubSecretEngineRole** — Creates a permission set (not a traditional "role"):
- Required: `path`
- Uses either `installationID` (int64) or `organizationName` — at least one required
- `repositories`: list of repo names within the org
- `permissions`: map of permission names to access types (`read` or `write`)
- Note: Vault path uses `permissionset` not `roles`: `{path}/permissionset/{name}`

#### secret-quay/

**SecretEngineMount** — `type: quay`, `path: quay-demo` (requires vault-plugin-secrets-quay)

**QuaySecretEngineConfig** — Configures Quay connection:
- Required: `path`, `url`, `rootCredentials`
- `url` is the Quay instance URL
- `rootCredentials` uses a Kubernetes Secret — **critical:** `passwordKey` must be set to the key containing the Quay API token (default key is `password`; set `passwordKey: token` if secret uses `token` as key name)
- Optional: `caCertificate`, `disableSslVerification`

**QuaySecretEngineRole** — Dynamic robot account credentials:
- Required: `path`, `namespaceName`
- `namespaceType`: `organization` (default) or `user`
- `repositories`: map of repo names → permissions (`admin`, `read`, `write`)
- `teams`: map of team names → roles (`admin`, `creator`, `member`)
- Optional: `createRepositories`, `defaultPermission`, `TTL`, `maxTTL`

**QuaySecretEngineStaticRole** — Static robot account (no TTL):
- Same fields as QuaySecretEngineRole except no `TTL`/`maxTTL`
- Credentials read via `vault read {path}/static-creds/{name}`

#### secret-kubernetes/

**SecretEngineMount** — `type: kubernetes`, `path: kubernetes-se-demo`

Use a different mount path than `kubernetes` to avoid confusion with the auth engine mount. Use `kubernetes-se-demo` or similar.

**KubernetesSecretEngineConfig** — Configures Kubernetes secret engine:
- Required: `path`, `kubernetesHost`, `jwtReference`
- `kubernetesHost` should default to `https://kubernetes.default.svc:443`
- `jwtReference` uses a Kubernetes Secret of service account token type: `jwtReference: { secret: { name: <secret-name> } }`
- Optional: `kubernetesCACert`, `disableLocalCAJWT`

**KubernetesSecretEngineRole** — Dynamic K8s credential role:
- Required: `path`, `allowedKubernetesNamespaces`
- Three mutually exclusive modes — show `kubernetesRoleName` (most common):
  1. `serviceAccountName` — use pre-existing SA (token only)
  2. `kubernetesRoleName` + `kubernetesRoleType` — bind to existing Role/ClusterRole
  3. `generateRoleRules` — inline RBAC rules
- Include: `defaultTTL`, `maxTTL`, `nameTemplate`
- `targetNamespaces` with `targetNamespaceSelector` (same pattern as auth engine role)

#### secret-azure/

**SecretEngineMount** — `type: azure`, `path: azure-se-demo`

Use a different mount path than `azure` to avoid confusion with the auth engine mount. Use `azure-se-demo` or similar.

**AzureSecretEngineConfig** — Configures Azure secret engine:
- Required: `path`, `subscriptionID`, `tenantID`
- `azureCredentials` uses Pattern B (Kubernetes Secret): `{ secret: { name: ... }, usernameKey: "clientid", passwordKey: "clientsecret" }`
- Optional: `environment` (defaults to `AzurePublicCloud`), `passwordPolicy`, `rootPasswordTTL`

**AzureSecretEngineRole** — Dynamic Azure SP credentials:
- Required: `path`
- `azureRoles`: **JSON-encoded string** (not a YAML list!) — e.g., `'[{"role_name":"Contributor","scope":"/subscriptions/.../resourceGroups/my-rg"}]'`
- `azureGroups`: also JSON-encoded string for group assignments
- Optional: `TTL`, `maxTTL`, `persistApp`, `signInAudience`, `tags`, `permanentlyDelete`

### Project Structure Notes

- All new files go under `docs/examples/secret-<engine>/`
- One YAML file per engine directory
- No Go code changes, no Makefile changes, no CRD changes
- Pure documentation/examples — no build, test, or runtime impact
- The existing `docs/examples/postgresql/` directory is **not modified**

### Naming Conventions

- Directory: `secret-<engine-name>/` (lowercase, hyphen-separated)
- Files: `secret-<engine-name>.yaml` (matches directory name)
- Resource metadata.name: descriptive but short (e.g., `database-config`, `pki-root-ca`, `rabbitmq-reader`)

### What NOT to Do

- Do NOT include `namespace` in metadata — let user decide
- Do NOT include `status` blocks — those are runtime-generated
- Do NOT use actual secrets/credentials — use placeholders with comments
- Do NOT copy test fixtures from `test/` directory — those are terse; examples should be user-friendly with comments
- Do NOT add a README.md in each directory — the YAML comments are sufficient
- Do NOT move or rename the existing `docs/examples/postgresql/` directory
- Do NOT use YAML lists for `azureRoles` or `azureGroups` — these are JSON-encoded strings in the CRD
- Do NOT use multiple vhost entries in the RabbitMQ role example — there is a known serialization bug with multi-entry (see D3 retro)

### Previous Story Intelligence (D4.1)

D4.1 created example YAML files for all 6 auth engines. Key patterns to replicate:
- Same authentication block pattern (`path: kubernetes`, `role: policy-admin`) across all examples
- Same file structure: one multi-document YAML per directory
- Comments are concise and explain the "what" and "when to change", not just field name repetition
- Resources ordered by dependency chain
- Placeholder values use consistent patterns (e.g., `example.com`, `my-org`, `my-app`)
- No `connection` block in examples (optional, most users don't need it)

### Git Intelligence

Recent commits are documentation-focused (Epic D2, D3). No code changes that affect this story. The codebase is stable for documentation work.

### References

- [Source: docs/secret-engines/index.md] — Supported secret engines index with CRD names and engine types
- [Source: docs/secret-engines/database.md] — DatabaseSecretEngineConfig, Role, and StaticRole field details
- [Source: docs/secret-engines/pki.md] — PKISecretEngineConfig and Role field details, root vs intermediate CA
- [Source: docs/secret-engines/rabbitmq.md] — RabbitMQSecretEngineConfig and Role field details, vhost permissions
- [Source: docs/secret-engines/github.md] — GitHubSecretEngineConfig and Role field details, permission sets
- [Source: docs/secret-engines/quay.md] — QuaySecretEngineConfig, Role, and StaticRole field details
- [Source: docs/secret-engines/kubernetes.md] — KubernetesSecretEngineConfig and Role field details, JWT reference
- [Source: docs/secret-engines/azure.md] — AzureSecretEngineConfig and Role field details, JSON-encoded roles
- [Source: docs/examples/postgresql/postgresql-secret-engine.yaml] — Reference pattern for multi-document example YAML
- [Source: _bmad-output/implementation-artifacts/d4-1-create-example-yaml-files-for-each-auth-engine.md] — D4.1 story with auth engine example conventions
- [Source: _bmad-output/implementation-artifacts/epic-d3-retro-2026-07-05.md] — D3 retro confirms D4 readiness; notes RabbitMQ multi-entry serialization bug

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6

### Debug Log References

- Integration tests skipped during pre-flight: Kind cluster container not running (infrastructure issue, not code regression). Story is pure documentation with no runtime impact.

### Completion Notes List

- Created 7 example YAML files for all secret engines supported by the operator
- Each file follows the multi-document YAML pattern from D4.1 auth engine examples and the existing `docs/examples/postgresql/` reference
- All files use `apiVersion: redhatcop.redhat.io/v1alpha1` and standard authentication block (`path: kubernetes`, `role: policy-admin`)
- Resources ordered by dependency: SecretEngineMount → Config → Role(s) → StaticRole
- No `namespace` in metadata, no `status` blocks, no actual credentials
- Inline YAML comments explain field purpose and when to change values
- RabbitMQ example uses single vhost entry per known serialization bug (D3 retro)
- Azure role uses JSON-encoded string for `azureRoles` (not YAML list) per CRD requirement
- GitHub example notes `sSHKeyReference` double-capital-S naming and `permissionset` Vault path
- Kubernetes and Azure secret engines use distinct mount paths (`kubernetes-se-demo`, `azure-se-demo`) to avoid auth engine confusion
- Quay example includes both dynamic (TTL) and static (no TTL) roles
- All 7 YAML files validated with Python YAML parser — no syntax errors

### Change Log

- 2026-07-07: Created all 7 secret engine example directories and YAML files

### File List

- docs/examples/secret-database/secret-database.yaml (new)
- docs/examples/secret-pki/secret-pki.yaml (new)
- docs/examples/secret-rabbitmq/secret-rabbitmq.yaml (new)
- docs/examples/secret-github/secret-github.yaml (new)
- docs/examples/secret-quay/secret-quay.yaml (new)
- docs/examples/secret-kubernetes/secret-kubernetes.yaml (new)
- docs/examples/secret-azure/secret-azure.yaml (new)
- _bmad-output/implementation-artifacts/d4-2-create-example-yaml-files-for-each-secret-engine.md (modified)
- _bmad-output/implementation-artifacts/sprint-status.yaml (modified)

### Review Findings

- [x] [Review][Patch] Secret engine dependent CRs use the mount prefix instead of the actual mounted path [docs/examples/secret-database/secret-database.yaml:20] — **Fixed**: all 7 examples now use `{mount spec.path}/{mount metadata.name}` pattern matching the existing postgresql reference
- [x] [Review][Patch] Database example allowlist omits the static role created later in the same file [docs/examples/secret-database/secret-database.yaml:47] — **Fixed**: added `app-user-rotation` to `allowedRoles`
- [x] [Review][Patch] GitHub example uses `github` instead of the registered plugin mount type [docs/examples/secret-github/secret-github.yaml:28] — **Fixed**: changed to `vault-plugin-secrets-github`; also fixed Quay to `vault-plugin-secrets-quay`
- [x] [Review][Dismiss] Sprint status history comments contradict the current story state transition — dismissed on deeper inspection; comments are the established append-only reverse-chronological audit trail format
