# Story D4.1: Create Example YAML Files for Each Auth Engine

Status: done

## Story

As a user learning the operator,
I want ready-to-use example YAML files for each auth engine,
so that I can quickly bootstrap my configuration.

## Acceptance Criteria

1. **Given** only `docs/examples/postgresql/` exists today, **when** example directories are created for each auth engine, **then** the following directories exist with complete, valid example CRs:
   - `docs/examples/auth-kubernetes/` — AuthEngineMount + KubernetesAuthEngineConfig + KubernetesAuthEngineRole
   - `docs/examples/auth-ldap/` — LDAPAuthEngineConfig + LDAPAuthEngineGroup
   - `docs/examples/auth-jwt-oidc/` — JWTOIDCAuthEngineConfig + JWTOIDCAuthEngineRole (both JWT and OIDC modes)
   - `docs/examples/auth-gcp/` — GCPAuthEngineConfig + GCPAuthEngineRole (both IAM and GCE types)
   - `docs/examples/auth-azure/` — AzureAuthEngineConfig + AzureAuthEngineRole
   - `docs/examples/auth-cert/` — CertAuthEngineConfig + CertAuthEngineRole

2. **Given** each example directory, **when** the YAML files are validated, **then** all examples use the correct `apiVersion: redhatcop.redhat.io/v1alpha1`, include required fields, and contain helpful inline comments explaining each field's purpose.

3. **Given** the existing `docs/examples/postgresql/` as the reference pattern, **when** examples are reviewed, **then** they follow the same multi-document-or-multi-file YAML pattern with resources ordered by dependency (mount → config → role/group).

## Tasks / Subtasks

- [x] Task 1: Create `docs/examples/auth-kubernetes/` directory (AC: #1, #2, #3)
  - [x] Create `auth-kubernetes.yaml` with AuthEngineMount + KubernetesAuthEngineConfig + KubernetesAuthEngineRole
- [x] Task 2: Create `docs/examples/auth-ldap/` directory (AC: #1, #2, #3)
  - [x] Create `auth-ldap.yaml` with LDAPAuthEngineConfig + LDAPAuthEngineGroup
- [x] Task 3: Create `docs/examples/auth-jwt-oidc/` directory (AC: #1, #2, #3)
  - [x] Create `auth-jwt-oidc.yaml` with JWTOIDCAuthEngineConfig (OIDC mode) + JWTOIDCAuthEngineRole
  - [x] Create `auth-jwt.yaml` with JWTOIDCAuthEngineConfig (JWT mode with JWKS) + JWTOIDCAuthEngineRole (roleType: jwt)
- [x] Task 4: Create `docs/examples/auth-gcp/` directory (AC: #1, #2, #3)
  - [x] Create `auth-gcp.yaml` with GCPAuthEngineConfig + GCPAuthEngineRole (IAM type) + GCPAuthEngineRole (GCE type)
- [x] Task 5: Create `docs/examples/auth-azure/` directory (AC: #1, #2, #3)
  - [x] Create `auth-azure.yaml` with AzureAuthEngineConfig + AzureAuthEngineRole
- [x] Task 6: Create `docs/examples/auth-cert/` directory (AC: #1, #2, #3)
  - [x] Create `auth-cert.yaml` with CertAuthEngineConfig + CertAuthEngineRole

### Review Findings

- [x] [Review][Patch] Align Kubernetes auth example mount/config/role paths so they target the same auth engine [docs/examples/auth-kubernetes/auth-kubernetes.yaml:13]
- [x] [Review][Patch] Correct the JWT example prerequisite to use the repo's actual JWT/OIDC auth mount type [docs/examples/auth-jwt-oidc/auth-jwt.yaml:13]
- [x] [Review][Patch] Fix the OIDC example's `namespaceInState` guidance for `form_post` flows [docs/examples/auth-jwt-oidc/auth-jwt-oidc.yaml:34]
- [x] [Review][Patch] Make the LDAP TLS Secret prerequisite consistent with the example's unconditional `tLSConfig` reference [docs/examples/auth-ldap/auth-ldap.yaml:10]

## Dev Notes

### Example File Pattern

Follow the existing `docs/examples/postgresql/postgresql-secret-engine.yaml` pattern:
- Multi-document YAML (resources separated by `---`)
- Resources ordered by dependency: mount first, then config, then role/group
- No namespace specified in metadata (user decides deployment namespace)
- Use realistic but clearly-placeholder values (e.g., `my-project`, `example.com`)
- Include inline YAML comments (`#`) explaining what each field does and when to change it

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

### Per-Engine CRD Details

#### auth-kubernetes/

**AuthEngineMount** — Creates the auth engine mount point:
```yaml
kind: AuthEngineMount
spec:
  type: kubernetes
  path: kubernetes  # Vault mount path for this auth engine
```

**KubernetesAuthEngineConfig** — Configures the engine:
- Required: `path`, `kubernetesHost`
- Optional: `tokenReviewerServiceAccount` (name reference), `kubernetesCACert`, `issuer`
- `kubernetesHost` should default to `https://kubernetes.default.svc:443`

**KubernetesAuthEngineRole** — Creates a role:
- Required: `path`, `policies`, `targetServiceAccounts`, `targetNamespaces`
- `targetNamespaces` has two mutually exclusive sub-fields: `targetNamespaceSelector` (label selector) or `targetNamespaces` (static list)
- Show the label selector approach in examples (more common production pattern)

#### auth-ldap/

**LDAPAuthEngineConfig** — Configures LDAP auth:
- Required: `path`, `url`, `bindCredentials` (secret reference)
- Key fields: `bindDN`, `userDN`, `userAttr`, `groupDN`, `groupAttr`, `groupFilter`
- TLS: show `tLSConfig.tlsSecret` approach (Kubernetes-native)
- `bindCredentials` uses Pattern B (with `secret.name`): `bindCredentials: { secret: { name: <secret-name> } }`

**LDAPAuthEngineGroup** — Maps LDAP groups to policies:
- Required: `path`, `name` (LDAP group name)
- `policies`: comma-separated string (not a list!)

#### auth-jwt-oidc/

Show **two separate files** demonstrating both modes:

**OIDC mode** (`auth-jwt-oidc.yaml`):
- Config: `OIDCDiscoveryURL`, `OIDCClientID`, `OIDCResponseMode`, `OIDCResponseTypes`, `OIDCCredentials` (secret ref)
- Role: `roleType: oidc` (default), `userClaim`, `boundAudiences`, `OIDCScopes`, `allowedRedirectURIs`, `groupsClaim`

**JWT mode** (`auth-jwt.yaml`):
- Config: `JWKSURL` (instead of `OIDCDiscoveryURL`), no credentials needed
- Role: `roleType: jwt` (MUST be explicit), `userClaim`, `boundAudiences`

#### auth-gcp/

**GCPAuthEngineConfig** — Configures GCP auth:
- Optional: `GCPCredentials` (secret ref), `serviceAccount`, `IAMalias`, `GCEalias`
- `GCPCredentials` uses Pattern B: `{ secret: { name: ... }, usernameKey: "serviceaccount", passwordKey: "credentials" }`

**GCPAuthEngineRole** — Show both IAM and GCE role types in same file:
- IAM role: `type: iam`, `boundServiceAccounts`, `boundProjects`, `maxJWTExp`
- GCE role: `type: gce`, `boundZones`, `boundRegions`, `boundLabels`

#### auth-azure/

**AzureAuthEngineConfig** — Configures Azure auth:
- Required: `path`, `tenantID`, `resource`
- Optional: `azureCredentials` (secret ref), `environment`, `maxRetries`
- `azureCredentials` uses Pattern B: `{ secret: { name: ... }, usernameKey: "clientid", passwordKey: "clientsecret" }`

**AzureAuthEngineRole** — Creates Azure role:
- Required: `path`, `name`
- Binding fields: `boundSubscriptionIDs`, `boundResourceGroups`, `boundServicePrincipalIDs`, `boundLocations`, `boundScalesets`
- At least one binding must be specified

#### auth-cert/

**CertAuthEngineConfig** — Configures TLS cert auth:
- Required: `path`
- Optional: `ocspCacheSize`, `roleCacheSize`, `disableBinding`, `enableIdentityAliasMetadata`
- Minimal config (engine has sensible defaults)

**CertAuthEngineRole** — Creates a cert role:
- Required: `path`, `certificate` (PEM CA cert)
- Optional: `allowedCommonNames`, `allowedDNSSANs`, `allowedEmailSANs`, `allowedURISANs`
- Token params: `tokenPolicies`, `tokenTTL`, `tokenMaxTTL`
- Use placeholder certificate (clearly marked `# Replace with your CA certificate`)

### Project Structure Notes

- All files go under `docs/examples/auth-<engine>/`
- One YAML file per engine (except JWT/OIDC which has two to demonstrate both modes)
- No Go code changes, no Makefile changes, no CRD changes
- Pure documentation/examples — no build, test, or runtime impact

### Naming Conventions

- Directory: `auth-<engine-name>/` (lowercase, hyphen-separated)
- Files: `auth-<engine-name>.yaml` (matches directory name)
- Resource metadata.name: descriptive but short (e.g., `kubernetes-config`, `ldap-admins`)

### What NOT to Do

- Do NOT include `namespace` in metadata — let user decide
- Do NOT include `status` blocks — those are runtime-generated
- Do NOT use actual secrets/credentials — use placeholders with comments
- Do NOT copy test fixtures from `test/` directory — those are terse; examples should be user-friendly with comments
- Do NOT add a README.md in each directory — the YAML comments are sufficient

### References

- [Source: docs/auth-engines/index.md] — Supported auth engines index with CRD names
- [Source: docs/auth-engines/kubernetes.md] — KubernetesAuthEngineConfig and Role field details
- [Source: docs/auth-engines/ldap.md] — LDAPAuthEngineConfig and Group field details
- [Source: docs/auth-engines/jwt-oidc.md] — JWTOIDCAuthEngineConfig and Role field details, JWT vs OIDC modes
- [Source: docs/auth-engines/gcp.md] — GCPAuthEngineConfig and Role field details, IAM vs GCE types
- [Source: docs/auth-engines/azure.md] — AzureAuthEngineConfig and Role field details
- [Source: docs/auth-engines/cert.md] — CertAuthEngineConfig and Role field details
- [Source: docs/examples/postgresql/postgresql-secret-engine.yaml] — Reference pattern for multi-document example YAML
- [Source: _bmad-output/implementation-artifacts/epic-d3-retro-2026-07-05.md] — D3 retro confirms D4 readiness, no prep needed

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6

### Debug Log References

(none — documentation-only story, no debugging needed)

### Completion Notes List

- Created 7 example YAML files across 6 directories for all supported auth engines
- All files follow the established postgresql example pattern: multi-document YAML, dependency ordering, inline comments, no namespace in metadata
- All CRs use `apiVersion: redhatcop.redhat.io/v1alpha1` with correct Kind names matching CRD definitions
- auth-kubernetes: AuthEngineMount + KubernetesAuthEngineConfig + KubernetesAuthEngineRole (demonstrates label selector for targetNamespaces)
- auth-ldap: LDAPAuthEngineConfig (with TLS secret reference) + LDAPAuthEngineGroup (comma-separated policies)
- auth-jwt-oidc: Two files — OIDC mode (OIDCDiscoveryURL + OIDCCredentials) and JWT mode (JWKSURL, explicit roleType: jwt)
- auth-gcp: GCPAuthEngineConfig + two GCPAuthEngineRole resources demonstrating both IAM and GCE types
- auth-azure: AzureAuthEngineConfig + AzureAuthEngineRole with subscription/resource group/principal bindings
- auth-cert: CertAuthEngineConfig + CertAuthEngineRole with placeholder CA certificate and SAN constraints
- No code changes, no Makefile changes, no CRD changes — pure documentation/examples

### File List

- docs/examples/auth-kubernetes/auth-kubernetes.yaml (new)
- docs/examples/auth-ldap/auth-ldap.yaml (new)
- docs/examples/auth-jwt-oidc/auth-jwt-oidc.yaml (new)
- docs/examples/auth-jwt-oidc/auth-jwt.yaml (new)
- docs/examples/auth-gcp/auth-gcp.yaml (new)
- docs/examples/auth-azure/auth-azure.yaml (new)
- docs/examples/auth-cert/auth-cert.yaml (new)

### Change Log

- 2026-07-07: Created example YAML files for all 6 auth engines (kubernetes, ldap, jwt-oidc, gcp, azure, cert) with complete CRD examples, inline documentation, and dependency ordering
