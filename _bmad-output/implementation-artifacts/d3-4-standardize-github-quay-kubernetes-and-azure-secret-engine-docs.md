# Story D3.4: Standardize GitHub, Quay, Kubernetes, and Azure Secret Engine Docs

Status: done

## Story

As a user configuring any of these secret engines,
I want consistent, complete documentation following the standard pattern,
So that switching between engines feels familiar.

## Acceptance Criteria

1. **Given** the existing content for GitHub, Quay, Kubernetes, and Azure secret engines in `docs/secret-engines.md` (lines 181-740) **When** extracted and standardized per the template **Then** each file follows the template structure (Overview → Config CRD → Role CRD(s) → Credential Resolution → See Also)

2. **Given** the new `github.md` file **When** validated against the template **Then** it documents GitHubSecretEngineConfig and GitHubSecretEngineRole with the SSH key credential pattern (not Pattern A/B — this engine uses `sSHKeyReference` with Kubernetes SSH secret or Vault secret)

3. **Given** the new `quay.md` file **When** validated **Then** it documents QuaySecretEngineConfig, QuaySecretEngineRole, AND QuaySecretEngineStaticRole (three CRDs), with Pattern B credential resolution (`rootCredentials.{secret,vaultSecret,randomSecret}`)

4. **Given** the new `kubernetes.md` file **When** validated **Then** it documents KubernetesSecretEngineConfig and KubernetesSecretEngineRole, with `jwtReference` credential pattern (Kubernetes secret or Vault secret only — RandomSecret is explicitly disallowed)

5. **Given** the new `azure.md` file **When** validated **Then** it documents AzureSecretEngineConfig and AzureSecretEngineRole, with Pattern B credential resolution (`azureCredentials.{secret,vaultSecret,randomSecret}`), and all "OIDC" references from the current text are removed (copy-paste error from auth engine docs)

6. **Given** cross-references in `readme.md` lines 87-93 pointing to `secret-engines.md#...` **When** the content is moved **Then** those links are updated to point to the new per-engine files

## Tasks / Subtasks

- [x] Task 1: Create `docs/secret-engines/github.md` (AC: 1, 2)
  - [x] 1.1: Write Overview — 2-3 sentences: third-party plugin (vault-plugin-secrets-github v2.0.0), generates GitHub installation access tokens, link to plugin repo
  - [x] 1.2: Write GitHubSecretEngineConfig section — Example YAML (include `applicationID`, `sSHKeyReference`, `gitHubAPIBaseURL`), Vault CLI Equivalent, Field Descriptions table
  - [x] 1.3: Write GitHubSecretEngineRole section — Example YAML (include `repositories`, `organizationName`, `permissions`), Vault CLI Equivalent (uses `permissionset` path, NOT `roles`), Field Descriptions table
  - [x] 1.4: Write SSH Key Resolution section (custom pattern — NOT Pattern A/B from template). Two methods: Kubernetes SSH-type secret (`sSHKeyReference.secret`) or Vault secret (`sSHKeyReference.vaultSecret`). Document mutual exclusivity (webhook rejects 0 or 2 sources)
  - [x] 1.5: Add See Also section

- [x] Task 2: Create `docs/secret-engines/quay.md` (AC: 1, 3)
  - [x] 2.1: Write Overview — 2-3 sentences: third-party plugin (vault-plugin-secrets-quay), manages Quay robot accounts, link to plugin repo
  - [x] 2.2: Write QuaySecretEngineConfig section — Example YAML (include `url`, `rootCredentials`, `caCertificate`, `disableSslVerification`), Vault CLI Equivalent, Field Descriptions table
  - [x] 2.3: Write QuaySecretEngineRole section — Example YAML (include `namespaceName`, `namespaceType`, `createRepositories`, `defaultPermission`, `teams`, `repositories`, `TTL`, `maxTTL`), Vault CLI Equivalent, Field Descriptions table
  - [x] 2.4: Write QuaySecretEngineStaticRole section — Example YAML (same fields as Role minus TTL/maxTTL), Vault CLI Equivalent, Field Descriptions table
  - [x] 2.5: Write Credential Resolution section using Pattern B (`rootCredentials.{secret,vaultSecret,randomSecret}`)
  - [x] 2.6: Add See Also section

- [x] Task 3: Create `docs/secret-engines/kubernetes.md` (AC: 1, 4)
  - [x] 3.1: Write Overview — 2-3 sentences: built-in Vault engine, generates Kubernetes service accounts/tokens with scoped RBAC, link to Vault docs
  - [x] 3.2: Write KubernetesSecretEngineConfig section — Example YAML (include `kubernetesHost`, `jwtReference`, `kubernetesCACert`, `disableLocalCAJWT`), Vault CLI Equivalent, Field Descriptions table
  - [x] 3.3: Write KubernetesSecretEngineRole section — Example YAML (include `allowedKubernetesNamespaces`, `kubernetesRoleName`, `kubernetesRoleType`, `nameTemplate`, full field set), Vault CLI Equivalent, Field Descriptions table
  - [x] 3.4: Write JWT Reference section (custom credential pattern — only Kubernetes service-account-token secret or Vault secret; RandomSecret is explicitly rejected by webhook)
  - [x] 3.5: Add See Also section

- [x] Task 4: Create `docs/secret-engines/azure.md` (AC: 1, 5)
  - [x] 4.1: Write Overview — 2-3 sentences: built-in Vault engine, generates Azure service principals and credentials, link to Vault docs
  - [x] 4.2: Write AzureSecretEngineConfig section — Example YAML (include `tenantID`, `subscriptionID`, `clientID`, `environment`, `azureCredentials`, `passwordPolicy`, `rootPasswordTTL`), Vault CLI Equivalent, Field Descriptions table
  - [x] 4.3: Write AzureSecretEngineRole section — Example YAML (include `azureRoles`, `azureGroups`, `applicationObjectID`, `persistApp`, `TTL`, `maxTTL`, `permanentlyDelete`, `signInAudience`, `tags`), Vault CLI Equivalent, Field Descriptions table
  - [x] 4.4: Write Credential Resolution section using Pattern B (`azureCredentials.{secret,vaultSecret,randomSecret}`). Document that `clientID` set in spec takes precedence over clientID from credential source
  - [x] 4.5: Add See Also section

- [x] Task 5: Audit field names for camelCase consistency (AC: 1, 2, 3, 4, 5)
  - [x] 5.1: Cross-reference ALL field names in the four new docs against Go CRD type files — field names in docs MUST match `json:` tag values exactly
  - [x] 5.2: Fix any snake_case field names from the original `secret-engines.md` source

- [x] Task 6: Update `readme.md` cross-references (AC: 6)
  - [x] 6.1: Update line 87 from `./docs/secret-engines.md#GitHubSecretEngineConfig` to `./docs/secret-engines/github.md#githubsecretengineconfig`
  - [x] 6.2: Update line 88 from `./docs/secret-engines.md#GitHubSecretEngineRole` to `./docs/secret-engines/github.md#githubsecretenginerole`
  - [x] 6.3: Update line 91 from `./docs/secret-engines.md#QuaySecretEngineConfig` to `./docs/secret-engines/quay.md#quaysecretengineconfig`
  - [x] 6.4: Update line 92 from `./docs/secret-engines.md#QuaySecretEngineRole` to `./docs/secret-engines/quay.md#quaysecretenginerole`
  - [x] 6.5: Update line 93 from `./docs/secret-engines.md#QuaySecretEngineStaticRole` to `./docs/secret-engines/quay.md#quaysecretenginestaticrole`
  - [x] 6.6: (Optional) Add missing `readme.md` entries for KubernetesSecretEngineConfig, KubernetesSecretEngineRole, AzureSecretEngineConfig, AzureSecretEngineRole — these were never added to the README

- [x] Task 7: Verify links and structure (AC: 1)
  - [x] 7.1: Verify relative links resolve correctly from `docs/secret-engines/*.md` (`../auth-section.md`, `../contributing-vault-apis.md`, `../secret-management.md#randomsecret`)
  - [x] 7.2: Verify heading hierarchy and section ordering matches `docs/auth-engines/kubernetes.md` template pattern

## Dev Notes

### This Is a Documentation-Only Story

No Go code changes. No tests to run. No `make manifests generate`. The deliverables are:
- 4 new files: `docs/secret-engines/github.md`, `docs/secret-engines/quay.md`, `docs/secret-engines/kubernetes.md`, `docs/secret-engines/azure.md`
- 1 modified file: `readme.md` (update 5 cross-reference links, optionally add 4 missing entries)

### Dependency on D3.1

This story assumes D3.1 has been completed (creating `docs/secret-engines/index.md` and the `docs/secret-engines/` directory). If D3.1 is NOT yet done, create the directory if it doesn't exist.

### Template to Follow

Use `docs/engine-doc-template.md` as the structural pattern. Use `docs/auth-engines/kubernetes.md` as the primary reference implementation. Use `docs/auth-engines/azure.md` for the Pattern B credential resolution reference.

Key structural requirements:
1. Title: `# {EngineName} Secret Engine`
2. Link to Vault docs / plugin repo immediately below title
3. `## Overview` — 2-3 sentences + CRD list
4. `## {CRDName}` → `### Example` → `### Vault CLI Equivalent` → `### Field Descriptions` (repeat for each CRD)
5. `## Credential Resolution` or `## SSH Key Resolution` or `## JWT Reference` (engine-specific)
6. `## See Also`

### GitHubSecretEngineConfig — Complete Field Reference

From `api/v1alpha1/githubsecretengineconfig_types.go`:

**GHConfig fields:**

| CRD Field (camelCase) | JSON tag | Vault API key (snake_case) | Type | Required | Default |
|---|---|---|---|---|---|
| applicationID | `json:"applicationID,omitempty"` | `app_id` | int64 | Yes | — (Minimum: 1) |
| gitHubAPIBaseURL | `json:"gitHubAPIBaseURL"` | `base_url` | string | No | `"https://api.github.com"` |

**SSHKeyReference fields (SSHKeyConfig struct):**

| CRD Field | JSON tag | Type | Required | Description |
|---|---|---|---|---|
| sSHKeyReference.secret | `json:"secret,omitempty"` | LocalObjectReference | One of | Kubernetes SSH-type secret (`kubernetes.io/ssh-auth`) |
| sSHKeyReference.vaultSecret | `json:"vaultSecret,omitempty"` | VaultSecretReference | One of | Vault secret (read from `key` field) |

Additional top-level spec fields:
- `path` (Required) — mount path. Final Vault path: `[namespace/]{path}/config`
- `authentication` (Required) — see shared auth-section.md
- `connection` (Optional) — override Vault connection

**Important Behaviors:**
- `IsDeletable() = false` — deleting the CR does NOT delete the config from Vault
- `IsEquivalentToDesiredState` strips `prv_key` (SSH key is write-only, never returned by Vault read)
- Webhook validates exactly ONE of `sSHKeyReference.secret` or `sSHKeyReference.vaultSecret` (rejects 0 or 2)
- Secret must be of type `kubernetes.io/ssh-auth` — the private key is read from the `ssh-privatekey` data field
- Third-party plugin: [vault-plugin-secrets-github](https://github.com/martinbaillie/vault-plugin-secrets-github) v2.0.0

### GitHubSecretEngineRole — Complete Field Reference

From `api/v1alpha1/githubsecretenginerole_types.go`:

**PermissionSet fields:**

| CRD Field (camelCase) | JSON tag | Vault API key (snake_case) | Type | Required | Default |
|---|---|---|---|---|---|
| installationID | `json:"installationID,omitempty"` | `installation_id` | int64 | Optional | — |
| organizationName | `json:"organizationName,omitempty"` | `org_name` | string | Optional | — |
| repositories | `json:"repositories,omitempty"` | `repositories` | []string | No | — |
| repositoriesIDs | `json:"repositoriesIDs,omitempty"` | `repository_ids` | []string | No | — |
| permissions | `json:"permissions,omitempty"` | `permissions` | map[string]string | No | — |

Additional top-level spec fields:
- `path` (Required) — mount path. Final Vault path: `[namespace/]{path}/permissionset/{name}`
- `authentication` (Required), `connection` (Optional), `name` (Optional override)

**Important:** Only ONE of `installationID` or `organizationName` is required. If both are provided, `installationID` takes precedence.

**Important:** Path uses `permissionset` (NOT `roles`) — this is unique to the GitHub plugin. Vault CLI read: `vault read {path}/token/{name}` to generate a token.

### QuaySecretEngineConfig — Complete Field Reference

From `api/v1alpha1/quaysecretengineconfig_types.go`:

**QuayConfig fields:**

| CRD Field (camelCase) | JSON tag | Vault API key (snake_case) | Type | Required | Default |
|---|---|---|---|---|---|
| url | `json:"url,omitempty"` | `url` | string | Yes | — |
| caCertificate | `json:"caCertificate,omitempty"` | `ca_certificate` | string | No | — |
| disableSslVerification | `json:"disableSslVerification,omitempty"` | `disable_ssl_verification` | bool | No | false |

Additional top-level spec fields:
- `path` (Required) — mount path. Final Vault path: `[namespace/]{path}/config`
- `authentication` (Required), `connection` (Optional)
- `rootCredentials` (Required) — Pattern B credential resolution

**Important Behaviors:**
- `IsDeletable() = false` — deleting the CR does NOT delete the config from Vault
- `IsEquivalentToDesiredState` strips `password` (token is write-only, never returned by Vault read)
- Uses `rootCredentials` (RootCredentialConfig) with Pattern B nested structure
- Quay requires a token, NOT username/password. The `passwordKey` maps to the token field
- Third-party plugin: [vault-plugin-secrets-quay](https://github.com/redhat-cop/vault-plugin-secrets-quay)

**Bug in existing docs:** The current `docs/secret-engines.md` lines 282-286 incorrectly show Pattern A field names (`rootCredentialsFromSecret`, `rootCredentialsFromVaultSecret`, `rootCredentialsFromRandomSecret`) but the CRD actually uses Pattern B nested structure (`rootCredentials.secret`, `rootCredentials.vaultSecret`, `rootCredentials.randomSecret`). The YAML example on line 271-273 is correct (shows `rootCredentials.secret`). Fix this inconsistency.

### QuaySecretEngineRole — Complete Field Reference

From `api/v1alpha1/quaysecretenginerole_types.go` (inherits `QuayBaseRole` + adds TTL fields):

**QuayRole fields (QuayBaseRole + QuayRole):**

| CRD Field (camelCase) | JSON tag | Vault API key (snake_case) | Type | Required | Default |
|---|---|---|---|---|---|
| namespaceType | `json:"namespaceType"` | `namespace_type` | string (Enum) | No | `"organization"` (Enum: `organization`, `user`) |
| namespaceName | `json:"namespaceName,omitempty"` | `namespace_name` | string | Yes | — |
| createRepositories | `json:"createRepositories"` | `create_repositories` | *bool | No | `false` |
| defaultPermission | `json:"defaultPermission,omitempty"` | `default_permission` | *Permission (Enum) | No | — (Enum: `admin`, `read`, `write`) |
| teams | `json:"teams,omitempty"` | `teams` (JSON string) | *map[string]TeamRole | No | — (TeamRole Enum: `admin`, `creator`, `member`) |
| repositories | `json:"repositories,omitempty"` | `repositories` (JSON string) | *map[string]Permission | No | — |
| TTL | `json:"TTL,omitempty"` | `ttl` | *Duration | No | — |
| maxTTL | `json:"maxTTL,omitempty"` | `max_ttl` | *Duration | No | — |

Path: `[namespace/]{path}/roles/{name}`

### QuaySecretEngineStaticRole — Complete Field Reference

From `api/v1alpha1/quaysecretenginestaticrole_types.go` (inherits `QuayBaseRole` only — no TTL fields):

Same fields as QuayBaseRole (namespaceType, namespaceName, createRepositories, defaultPermission, teams, repositories). No TTL or maxTTL.

Path: `[namespace/]{path}/static-roles/{name}`

Vault CLI read: `vault read {path}/static-creds/{name}` for static credentials.

### KubernetesSecretEngineConfig — Complete Field Reference

From `api/v1alpha1/kubernetessecretengineconfig_types.go`:

**KubeSEConfig fields:**

| CRD Field (camelCase) | JSON tag | Vault API key (snake_case) | Type | Required | Default |
|---|---|---|---|---|---|
| kubernetesHost | `json:"kubernetesHost,omitempty"` | `kubernetes_host` | string | Yes | — |
| kubernetesCACert | `json:"kubernetesCACert,omitempty"` | `kubernetes_ca_cert` | string | No | — |
| disableLocalCAJWT | `json:"disableLocalCAJWT,omitempty"` | `disable_local_ca_jwt` | bool | No | false |

Additional top-level spec fields:
- `path` (Required) — mount path. Final Vault path: `[namespace/]{path}/config`
- `authentication` (Required), `connection` (Optional)
- `jwtReference` (Required) — service account JWT credential reference (RootCredentialConfig but RandomSecret is disallowed)

**Important Behaviors:**
- `IsDeletable() = true` — deleting the CR DOES delete the config from Vault
- `IsEquivalentToDesiredState` strips `service_account_jwt` (JWT is write-only, never returned by Vault read)
- `jwtReference` uses RootCredentialConfig struct but webhook explicitly rejects `randomSecret` — only `secret` (must be type `kubernetes.io/service-account-token`) or `vaultSecret` (reads `key` field from Vault)
- Built-in Vault engine: [Vault Kubernetes Secrets Engine](https://developer.hashicorp.com/vault/docs/secrets/kubernetes)

### KubernetesSecretEngineRole — Complete Field Reference

From `api/v1alpha1/kubernetessecretenginerole_types.go`:

**KubeSERole fields:**

| CRD Field (camelCase) | JSON tag | Vault API key (snake_case) | Type | Required | Default |
|---|---|---|---|---|---|
| allowedKubernetesNamespaces | `json:"allowedKubernetesNamespaces,omitempty"` | `allowed_kubernetes_namespaces` | []string | No | — |
| allowedKubernetesNamespaceSelector | `json:"allowedKubernetesNamespaceSelector,omitempty"` | `allowed_kubernetes_namespace_selector` | string | No | — |
| defaultTTL | `json:"defaultTTL,omitempty"` | `token_default_ttl` | Duration | No | — |
| maxTTL | `json:"maxTTL,omitempty"` | `token_max_ttl` | Duration | No | — |
| defaultAudiences | `json:"defaultAudiences,omitempty"` | `token_default_audiences` | string | No | — |
| serviceAccountName | `json:"serviceAccountName,omitempty"` | `service_account_name` | string | No | — |
| kubernetesRoleName | `json:"kubernetesRoleName,omitempty"` | `kubernetes_role_name` | string | No | — |
| kubernetesRoleType | `json:"kubernetesRoleType"` | `kubernetes_role_type` | string | No | `"Role"` (Enum: `Role`, `ClusterRole`) |
| generateRoleRules | `json:"generateRoleRules,omitempty"` | `generated_role_rules` | string | No | — |
| nameTemplate | `json:"nameTemplate,omitempty"` | `name_template` | string | No | — |
| extraAnnotations | `json:"extraAnnotations,omitempty"` | `extra_annotations` | map[string]string | No | — |
| extraLabels | `json:"extraLabels,omitempty"` | `extra_labels` | map[string]string | No | — |

Also uses `targetNamespaces` (TargetNamespaceConfig) at the spec level — same pattern as KubernetesAuthEngineRole.

Path: `[namespace/]{path}/roles/{name}`

**Three mutually exclusive credential generation modes:**
1. `serviceAccountName` — use a pre-existing service account (only token generated)
2. `kubernetesRoleName` — bind to a pre-existing Role/ClusterRole (SA + binding + token created)
3. `generateRoleRules` — generate the Role/ClusterRole from inline rules (entire RBAC chain created)

**Bug in existing docs:** Line 600-601 has `kubernetes_role_name` used twice in CLI example — the second instance should be `kubernetes_role_type`.

### AzureSecretEngineConfig — Complete Field Reference

From `api/v1alpha1/azuresecretengineconfig_types.go`:

**AzureSEConfig fields:**

| CRD Field (camelCase) | JSON tag | Vault API key (snake_case) | Type | Required | Default |
|---|---|---|---|---|---|
| subscriptionID | `json:"subscriptionID"` | `subscription_id` | string | Yes | — |
| tenantID | `json:"tenantID"` | `tenant_id` | string | Yes | — |
| clientID | `json:"clientID,omitempty"` | `client_id` (via retrievedClientID) | string | No | — |
| environment | `json:"environment"` | `environment` | string | No | `"AzurePublicCloud"` (Enum: `AzurePublicCloud`, `AzureUSGovernmentCloud`, `AzureChinaCloud`, `AzureGermanCloud`) |
| passwordPolicy | `json:"passwordPolicy,omitempty"` | `password_policy` | string | No | — |
| rootPasswordTTL | `json:"rootPasswordTTL"` | `root_password_ttl` | string | No | `"182d"` |

Additional top-level spec fields:
- `path` (Required) — mount path. Final Vault path: `[namespace/]{path}/config`
- `authentication` (Required), `connection` (Optional)
- `azureCredentials` (Optional) — Pattern B credential resolution for ClientID + ClientSecret

**Important Behaviors:**
- `IsDeletable() = true` — deleting the CR DOES delete the config from Vault
- `IsEquivalentToDesiredState` does NOT strip any fields (but `client_id` and `client_secret` are from internal retrieval, not directly from spec fields)
- If `azureCredentials` is omitted/empty (defaults to `{passwordKey: "clientsecret", usernameKey: "clientid"}`), the operator skips credential retrieval — Azure uses environment variables instead
- If `clientID` is set directly in the spec, it takes precedence over the username from the credential source
- Built-in Vault engine: [Vault Azure Secrets Engine](https://developer.hashicorp.com/vault/docs/secrets/azure)

**Bug in existing docs:** Lines 654-655 reference "OIDC" — `OIDCClientSecret` and `OIDCClientID` — which are copy-pasted from the auth engine docs. These should be `clientSecret` and `clientID`.

### AzureSecretEngineRole — Complete Field Reference

From `api/v1alpha1/azuresecretenginerole_types.go`:

**AzureSERole fields:**

| CRD Field (camelCase) | JSON tag | Vault API key (snake_case) | Type | Required | Default |
|---|---|---|---|---|---|
| azureRoles | `json:"azureRoles,omitempty"` | `azure_roles` | string (JSON) | No | — |
| azureGroups | `json:"azureGroups,omitempty"` | `azure_groups` | string (JSON) | No | — |
| applicationObjectID | `json:"applicationObjectID,omitempty"` | `application_object_id` | string | No | — |
| persistApp | `json:"persistApp,omitempty"` | `persist_app` | bool | No | false |
| TTL | `json:"TTL,omitempty"` | `ttl` | string | No | — |
| maxTTL | `json:"maxTTL,omitempty"` | `max_ttl` | string | No | — |
| permanentlyDelete | `json:"permanentlyDelete,omitempty"` | `permanently_delete` | string | No | — |
| signInAudience | `json:"signInAudience,omitempty"` | `sign_in_audience` | string | No | — (Enum: `AzureADMyOrg`, `AzureADMultipleOrgs`, `AzureADandPersonalMicrosoftAccount`, `PersonalMicrosoftAccount`) |
| tags | `json:"tags,omitempty"` | `tags` | string | No | — |

Path: `[namespace/]{path}/roles/{name}` (Note: the Go source comment says "groups" but the code actually uses "roles")

**Important:** `azureRoles` and `azureGroups` are JSON-encoded strings, not native arrays. They must be properly escaped in the YAML.

### Credential Resolution Patterns Summary

Each engine in this story uses a DIFFERENT credential pattern:

| Engine | CRD Field | Pattern | Allowed Sources | Notes |
|--------|-----------|---------|----------------|-------|
| GitHub | `sSHKeyReference` | Custom (SSH) | K8s SSH secret, Vault secret | NOT RootCredentialConfig. Webhook validates mutual exclusivity |
| Quay | `rootCredentials` | B (nested) | K8s secret, Vault secret, RandomSecret | Standard RootCredentialConfig. Token-only (no username) |
| Kubernetes | `jwtReference` | B (nested, restricted) | K8s SA-token secret, Vault secret | RootCredentialConfig but RandomSecret explicitly rejected by webhook |
| Azure | `azureCredentials` | B (nested) | K8s secret, Vault secret, RandomSecret | Standard RootCredentialConfig. ClientID can be overridden in spec |

### Known Issues in Source Content

From `docs/secret-engines.md`:

1. **Quay credential docs (lines 282-286):** Uses Pattern A field names (`rootCredentialsFromSecret`) but CRD uses Pattern B nested structure (`rootCredentials.secret`). The YAML example (line 271) is correct. Fix the prose to match the CRD
2. **Azure credential docs (lines 654-655):** References "OIDC" (`OIDCClientSecret`, `OIDCClientID`) — copy-paste error from auth engine docs. Replace with `clientSecret` and `clientID`
3. **Kubernetes Role CLI example (line 600-601):** `kubernetes_role_name` appears twice — second should be `kubernetes_role_type="ClusterRole"`
4. **Kubernetes Secret Engine Config CLI (line 561):** `kube-setest` has wrong hyphenation — should match the spec example `kubese-test`
5. **Quay existing doc (line 319):** "type of name of the namespace" — awkward phrasing, should be "name of the Quay organization or user"
6. **GitHub role existing doc (line 593):** "sued" → "used" typo

### readme.md Cross-References

Lines requiring update:

| Line | Current Link | New Link |
|------|-------------|----------|
| 87 | `./docs/secret-engines.md#GitHubSecretEngineConfig` | `./docs/secret-engines/github.md#githubsecretengineconfig` |
| 88 | `./docs/secret-engines.md#GitHubSecretEngineRole` | `./docs/secret-engines/github.md#githubsecretenginerole` |
| 91 | `./docs/secret-engines.md#QuaySecretEngineConfig` | `./docs/secret-engines/quay.md#quaysecretengineconfig` |
| 92 | `./docs/secret-engines.md#QuaySecretEngineRole` | `./docs/secret-engines/quay.md#quaysecretenginerole` |
| 93 | `./docs/secret-engines.md#QuaySecretEngineStaticRole` | `./docs/secret-engines/quay.md#quaysecretenginestaticrole` |

Note: `KubernetesSecretEngine*` and `AzureSecretEngine*` are NOT in `readme.md` — they were added to the operator after the README was written. Optionally add entries for them after line 95.

### Relative Link Conventions

From `docs/secret-engines/{engine}.md`:
- To shared docs: `../auth-section.md`, `../contributing-vault-apis.md`
- To secret management (RandomSecret): `../secret-management.md#randomsecret`
- To other engine files: `database.md`, `pki.md` (same directory)
- External: full URLs to Vault documentation or plugin repos

### Vault Path Structure

Understanding the path hierarchy for accurate CLI equivalents:

```
GitHub (third-party plugin):
{mount-path}/config                      ← GitHubSecretEngineConfig
{mount-path}/permissionset/{name}        ← GitHubSecretEngineRole (NOT /roles/)
{mount-path}/token/{name}               ← Read credential

Quay (third-party plugin):
{mount-path}/config                      ← QuaySecretEngineConfig
{mount-path}/roles/{name}               ← QuaySecretEngineRole
{mount-path}/creds/{name}               ← Read dynamic credential
{mount-path}/static-roles/{name}        ← QuaySecretEngineStaticRole
{mount-path}/static-creds/{name}        ← Read static credential

Kubernetes (built-in):
{mount-path}/config                      ← KubernetesSecretEngineConfig
{mount-path}/roles/{name}               ← KubernetesSecretEngineRole
{mount-path}/creds/{name}               ← Read credential

Azure (built-in):
{mount-path}/config                      ← AzureSecretEngineConfig
{mount-path}/roles/{name}               ← AzureSecretEngineRole
{mount-path}/creds/{name}               ← Read credential
```

### Previous Story Intelligence

**From D3.2 (Database Secret Engine Docs — same epic, closest precedent):**
- Established the three-CRD documentation pattern (Config + Role + StaticRole) — reuse for Quay
- Field descriptions table uses camelCase names from JSON tags
- Vault CLI equivalents use snake_case names
- Pattern A credential resolution documented — Quay and Azure use Pattern B instead
- Source content extracted from `docs/secret-engines.md` and cross-referenced against Go CRD types
- readme.md cross-references updated with lowercase anchor names

**From D3.1 (Secret-Engines Directory Structure & Index):**
- Created `docs/secret-engines/index.md` with engine table — github.md, quay.md, kubernetes.md, azure.md links ready
- Replaced `docs/secret-engines.md` with redirect pointer
- Documented all readme.md cross-references for downstream stories

**From D2.5 (GCP and Azure Auth Engine Docs — Pattern B credential resolution):**
- Established Pattern B credential resolution docs (`azureCredentials.{secret,vaultSecret,randomSecret}`)
- `docs/auth-engines/azure.md` is the exact reference for Azure Secret Engine credential docs
- Zero review findings — pattern fully proven

**From D2 Retrospective:**
- D3 readiness: No preparation work needed, all preconditions satisfied
- Template proven across 6 auth engine docs — applies directly to secret engines
- Recommendation: Continue using Opus 4.6 for all stories

### Project Structure Notes

```
docs/
├── secret-engines/
│   ├── index.md          ← D3.1
│   ├── database.md       ← D3.2
│   ├── pki.md            ← D3.3
│   ├── rabbitmq.md       ← D3.3
│   ├── github.md         ← NEW (this story)
│   ├── quay.md           ← NEW (this story)
│   ├── kubernetes.md     ← NEW (this story)
│   └── azure.md          ← NEW (this story)
├── secret-engines.md     ← redirect pointer (D3.1)
├── auth-engines/
│   ├── index.md          ← D2.1
│   ├── kubernetes.md     ← D2.2 — reference implementation
│   ├── azure.md          ← D2.5 — Pattern B credential resolution reference
│   └── ...
├── auth-section.md       ← shared auth config (unchanged)
├── engine-doc-template.md ← template (D1.1, review-patched)
└── ...
```

### References

- [Source: _bmad-output/planning-artifacts/epics.md#Story D3.4] — Story requirements and acceptance criteria
- [Source: docs/secret-engines.md:181-257] — GitHub content to extract and standardize
- [Source: docs/secret-engines.md:258-380] — Quay content to extract and standardize
- [Source: docs/secret-engines.md:532-603] — Kubernetes content to extract and standardize
- [Source: docs/secret-engines.md:605-740] — Azure content to extract and standardize
- [Source: docs/auth-engines/kubernetes.md] — Reference implementation for template pattern (D2.2)
- [Source: docs/auth-engines/azure.md] — Pattern B credential resolution reference (D2.5)
- [Source: docs/engine-doc-template.md] — Template structure (D1.1, review-patched)
- [Source: api/v1alpha1/githubsecretengineconfig_types.go] — GitHub Config CRD fields
- [Source: api/v1alpha1/githubsecretenginerole_types.go] — GitHub Role CRD fields
- [Source: api/v1alpha1/quaysecretengineconfig_types.go] — Quay Config CRD fields
- [Source: api/v1alpha1/quaysecretenginerole_types.go] — Quay Role CRD fields
- [Source: api/v1alpha1/quaysecretenginestaticrole_types.go] — Quay StaticRole CRD fields
- [Source: api/v1alpha1/kubernetessecretengineconfig_types.go] — Kubernetes Config CRD fields
- [Source: api/v1alpha1/kubernetessecretenginerole_types.go] — Kubernetes Role CRD fields
- [Source: api/v1alpha1/azuresecretengineconfig_types.go] — Azure Config CRD fields
- [Source: api/v1alpha1/azuresecretenginerole_types.go] — Azure Role CRD fields
- [Source: _bmad-output/implementation-artifacts/d3-2-standardize-database-secret-engine-docs.md] — Previous story context (same epic)
- [Source: _bmad-output/implementation-artifacts/d3-1-create-secret-engines-directory-structure-and-index-page.md] — D3.1 predecessor
- [Source: readme.md:87-93] — Cross-references that need updating
- [Source: _bmad-output/project-context.md] — Project conventions and coding standards

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6

### Debug Log References

- Pre-flight: Kind cluster local-path provisioner failure diagnosed (USB device `/dev/bus/usb/002/006` disappeared from host after container creation; stale device mapping caused helper pod `StartError`). Fixed by restarting Kind node container and re-applying iptables-nft configuration.
- Post-completion: Same infrastructure issue recurred twice (stale device `/dev/hidraw3`). Each recurrence fixed by restarting Kind node container, re-applying iptables-nft, and cleaning local-path-storage pods. Final `make integration` passed (exit 0, 582s controller tests, all packages OK).

### Completion Notes List

- Created `docs/secret-engines/github.md` with SSH Key Resolution (custom pattern, not Pattern A/B). Documents GitHubSecretEngineConfig and GitHubSecretEngineRole with `permissionset` path (not `/roles/`).
- Created `docs/secret-engines/quay.md` with three CRDs (Config, Role, StaticRole). Uses Pattern B credential resolution (`rootCredentials.{secret,vaultSecret,randomSecret}`). Fixed known bug from source: credential docs now correctly use Pattern B nested structure, not Pattern A flat field names.
- Created `docs/secret-engines/kubernetes.md` with JWT Reference section (custom restricted pattern — only K8s SA-token secret or Vault secret; RandomSecret explicitly rejected by webhook).
- Created `docs/secret-engines/azure.md` with Pattern B credential resolution (`azureCredentials.{secret,vaultSecret,randomSecret}`). Documents `clientID` precedence behavior. No OIDC references (fixed copy-paste error from source).
- All field names in all four docs verified against Go CRD `json:` tags — 100% match, no snake_case issues.
- Updated 5 existing readme.md cross-references from `secret-engines.md#...` to per-engine files with lowercase anchors.
- Added 4 new readme.md entries for KubernetesSecretEngineConfig, KubernetesSecretEngineRole, AzureSecretEngineConfig, AzureSecretEngineRole (previously missing from README).
- All relative links verified: `../auth-section.md`, `../contributing-vault-apis.md`, `../secret-management.md#randomsecret` all resolve correctly.
- Section ordering matches template pattern across all four files.

### Change Log

- 2026-07-05: Created github.md, quay.md, kubernetes.md, azure.md; updated readme.md cross-references; added missing K8s/Azure README entries
- 2026-07-05: Code review — fixed 5 patch findings (Azure credential defaults, Quay passwordKey docs, K8s JWT wording, Azure permanentlyDelete type); 1 deferred (pre-existing README phrasing)

### File List

- docs/secret-engines/github.md (new)
- docs/secret-engines/quay.md (new)
- docs/secret-engines/kubernetes.md (new)
- docs/secret-engines/azure.md (new)
- readme.md (modified — updated 5 cross-references, added 4 new entries)
- _bmad-output/implementation-artifacts/sprint-status.yaml (modified — story status)
- _bmad-output/implementation-artifacts/d3-4-standardize-github-quay-kubernetes-and-azure-secret-engine-docs.md (modified — task checkboxes, dev agent record)

### Review Findings

- [x] [Review][Patch] Azure docs describe an omitted/empty `azureCredentials` fallback mode that the current CRD/controller does not support [docs/secret-engines/azure.md] — fixed: removed incorrect defaults parenthetical, added exact-one-source constraint, aligned with auth-engine doc wording
- [x] [Review][Patch] Quay credential docs omit the `rootCredentials.passwordKey` requirement/default even though the examples rely on a token key override [docs/secret-engines/quay.md] — fixed: documented passwordKey default and override requirement
- [x] [Review][Patch] Azure credential docs omit the `usernameKey`/`passwordKey` defaults and the exact-one-source constraint for `azureCredentials` [docs/secret-engines/azure.md] — fixed: added key defaults note
- [x] [Review][Patch] Kubernetes JWT docs do not clearly state that exactly one of `jwtReference.secret` or `jwtReference.vaultSecret` must be set [docs/secret-engines/kubernetes.md] — fixed: changed to "exactly one ... must be specified"
- [x] [Review][Patch] Azure role docs present `permanentlyDelete` as a boolean-style flag even though the CRD field type is `string` [docs/secret-engines/azure.md] — fixed: clarified string type with `"true"`/`"false"` values
- [x] [Review][Defer] Existing README wording still says "see the also the" in several secret-engine entries [readme.md] — deferred, pre-existing
