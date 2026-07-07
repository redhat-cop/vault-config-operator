# Story D4.3: Create Additional End-to-End Examples

Status: ready-for-dev

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As a user designing a complete Vault configuration,
I want end-to-end examples beyond the PostgreSQL one,
so that I can see how different engines and auth methods work together.

## Acceptance Criteria

1. **Given** only one end-to-end example exists (PostgreSQL with Kubernetes auth), **when** additional end-to-end examples are created, **then** at least two new examples exist:
   - `docs/examples/e2e-jwt-pki/` — JWT/OIDC auth + PKI secret engine: complete walkthrough for certificate issuance
   - `docs/examples/e2e-azure/` — Azure auth + Azure secret engine: complete walkthrough for Azure service principal provisioning

2. **Given** each end-to-end example, **when** reviewed for completeness, **then** it includes: prerequisites, Vault setup (operator CRs), verification commands, and cleanup instructions — all in a single well-commented YAML file with a companion README.md.

3. **Given** each end-to-end example, **when** validated, **then** all CRs use the correct `apiVersion: redhatcop.redhat.io/v1alpha1`, include required fields, and contain inline comments explaining how the resources connect to each other.

## Tasks / Subtasks

- [ ] Task 1: Create `docs/examples/e2e-jwt-pki/` directory (AC: #1, #2, #3)
  - [ ] Create `README.md` with scenario description, prerequisites, walkthrough steps, verification, and cleanup
  - [ ] Create `e2e-jwt-pki.yaml` with: AuthEngineMount (JWT) + JWTOIDCAuthEngineConfig (JWT mode with JWKS) + JWTOIDCAuthEngineRole + SecretEngineMount (PKI) + PKISecretEngineConfig (root CA) + PKISecretEngineRole + Policy (connecting auth to secrets)
- [ ] Task 2: Create `docs/examples/e2e-azure/` directory (AC: #1, #2, #3)
  - [ ] Create `README.md` with scenario description, prerequisites, walkthrough steps, verification, and cleanup
  - [ ] Create `e2e-azure.yaml` with: AuthEngineMount (Azure) + AzureAuthEngineConfig + AzureAuthEngineRole + SecretEngineMount (Azure) + AzureSecretEngineConfig + AzureSecretEngineRole + Policy (connecting auth to secrets)

## Dev Notes

### Key Difference from D4.1/D4.2

D4.1 and D4.2 created **single-engine** examples (one auth OR one secret engine per directory). This story creates **cross-engine** examples that show how auth and secret engines **work together** in a real deployment scenario. The distinguishing factor is:

- Each example includes a **Policy** CRD granting the auth role access to the secret engine paths
- Resources reference each other via mount paths (the auth role's policy grants access to the secret engine's mount path)
- The README provides a narrative walkthrough explaining how the pieces connect

### End-to-End Example File Pattern

Each example directory contains:

1. **README.md** — Narrative documentation:
   - Scenario description (what problem this solves)
   - Prerequisites (what must exist before applying)
   - Step-by-step walkthrough (apply resources, explain what happens)
   - Verification commands (vault CLI commands to verify the configuration works)
   - Cleanup instructions (delete CRs, note what persists in Vault)

2. **e2e-<name>.yaml** — All CRs in a single multi-document YAML:
   - Resources ordered by dependency: auth mount → auth config → policy → secret mount → secret config → secret role → auth role (role last because it references the policy)
   - Inline comments explaining how resources connect to each other (e.g., "this policy grants access to the PKI engine mounted at the path above")

### apiVersion and Kind for All CRDs

All CRDs use:
```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
```

### Authentication Block Pattern

Every CRD includes an `authentication` block — this is how the **operator** authenticates to Vault to perform the operation. Use this standard pattern:
```yaml
spec:
  authentication:
    path: kubernetes
    role: policy-admin
```

The `spec.path` field (separate from `spec.authentication.path`) is the Vault mount path of the **engine being configured**.

### Example 1: JWT/OIDC Auth + PKI Secret Engine (`e2e-jwt-pki/`)

**Scenario:** An organization uses JWT tokens (from a CI/CD system like GitHub Actions) to authenticate to Vault, and needs to issue short-lived TLS certificates for internal services via PKI.

**CRD Stack (in order):**

1. **AuthEngineMount** — Mounts JWT auth at `jwt-ci`
   - `type: jwt`, `path: jwt-ci`

2. **JWTOIDCAuthEngineConfig** — Configures JWT validation via JWKS URL
   - `path: jwt-ci`
   - Use `JWKSURL` mode (no credentials needed — JWT mode, not OIDC)
   - Example JWKS URL: `https://token.actions.githubusercontent.com/.well-known/jwks` (GitHub Actions)
   - Set `boundIssuer: https://token.actions.githubusercontent.com`
   - No `OIDCCredentials` needed (JWT mode)

3. **Policy** — Grants certificate issuance access
   - Name: `ci-cert-issuer`
   - Policy content: allow `read` on `pki-ci/issue/ci-service` and `pki-ci/sign/ci-service`
   - Also allow `read` on `pki-ci/ca/pem` (for CA cert retrieval)

4. **SecretEngineMount** — Mounts PKI engine at `pki-ci`
   - `type: pki`, `path: pki-ci`

5. **PKISecretEngineConfig** — Creates root CA
   - `path: pki-ci`
   - `type: root`, `privateKeyType: internal`
   - `commonName: ci.internal.example.com`
   - `TTL: "87600h"` (10 years)
   - `organization`, `country`, etc. with placeholder values
   - `issuingCertificates` and `CRLDistributionPoints` URLs

6. **PKISecretEngineRole** — Certificate issuance role
   - `path: pki-ci`
   - `allowedDomains: ["internal.example.com", "svc.cluster.local"]`
   - `allowSubdomains: true`
   - `TTL: "1h"`, `maxTTL: "24h"` (short-lived CI certificates)
   - `keyType: rsa`, `keyBits: 2048`
   - `keyUsage: [DigitalSignature, KeyEncipherment]`
   - `extKeyUsage: [ServerAuth, ClientAuth]`

7. **JWTOIDCAuthEngineRole** — JWT auth role (references policy)
   - `path: jwt-ci`
   - `roleType: jwt` — **MUST be explicit** (default is `oidc`)
   - `userClaim: sub`
   - `boundAudiences: ["https://vault.example.com"]`
   - `boundClaims` to restrict to specific repo/workflow
   - `tokenPolicies: ["ci-cert-issuer"]` — connects to Policy above
   - `tokenTTL: "10m"`, `tokenMaxTTL: "30m"`

**Verification commands in README:**
```shell
# Verify auth engine is mounted
vault auth list | grep jwt-ci

# Verify PKI CA is configured
vault read pki-ci/ca/pem

# Verify role exists
vault read pki-ci/roles/ci-service

# Issue a test certificate (after authenticating via JWT)
vault write pki-ci/issue/ci-service common_name="myapp.internal.example.com" ttl="1h"
```

### Example 2: Azure Auth + Azure Secret Engine (`e2e-azure/`)

**Scenario:** Azure VMs authenticate to Vault using their managed identity, then request dynamic Azure service principal credentials for cross-subscription resource access.

**CRD Stack (in order):**

1. **AuthEngineMount** — Mounts Azure auth at `azure-auth`
   - `type: azure`, `path: azure-auth`

2. **AzureAuthEngineConfig** — Configures Azure auth
   - `path: azure-auth`
   - `tenantID: 00000000-0000-0000-0000-000000000000` (placeholder)
   - `resource: https://management.azure.com/`
   - `azureCredentials`: reference Kubernetes Secret with `usernameKey: clientid`, `passwordKey: clientsecret`

3. **Policy** — Grants access to Azure secret engine credentials
   - Name: `azure-sp-reader`
   - Policy: allow `read` on `azure-se/creds/contributor-role`

4. **SecretEngineMount** — Mounts Azure secret engine at `azure-se`
   - `type: azure`, `path: azure-se`
   - Use `azure-se` (not `azure`) to avoid path collision with the auth mount

5. **AzureSecretEngineConfig** — Configures Azure secret engine
   - `path: azure-se`
   - `subscriptionID`, `tenantID` with placeholder UUIDs
   - `azureCredentials`: reference Kubernetes Secret with `usernameKey: clientid`, `passwordKey: clientsecret`

6. **AzureSecretEngineRole** — Dynamic SP credential role
   - `path: azure-se`
   - `azureRoles`: JSON-encoded string — `'[{"role_name":"Contributor","scope":"/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/my-rg"}]'`
   - `TTL: "1h"`, `maxTTL: "4h"`

7. **AzureAuthEngineRole** — Auth role (references policy)
   - `path: azure-auth`
   - `name: azure-vm-role`
   - `boundSubscriptionIDs` with placeholder UUID
   - `boundResourceGroups: ["my-resource-group"]`
   - `tokenPolicies: ["azure-sp-reader"]` — connects to Policy above
   - `tokenTTL: "1h"`, `tokenMaxTTL: "4h"`

**Verification commands in README:**
```shell
# Verify auth engine is mounted
vault auth list | grep azure-auth

# Verify secret engine is configured
vault read azure-se/config

# Verify role exists
vault read azure-se/roles/contributor-role

# Generate credentials (after authenticating via Azure auth)
vault read azure-se/creds/contributor-role
```

### Naming Conventions

- Directory: `e2e-<descriptive-name>/` (the `e2e-` prefix distinguishes these from single-engine examples)
- YAML file: `e2e-<name>.yaml` (matches directory name)
- Resource metadata.name: descriptive, showing the scenario context (e.g., `jwt-ci`, `pki-ci-root`, `ci-service`)

### What NOT to Do

- Do NOT include `namespace` in metadata — let user decide
- Do NOT include `status` blocks — those are runtime-generated
- Do NOT use actual secrets/credentials — use placeholders with comments
- Do NOT copy test fixtures from `test/` directory — those are terse; examples should be user-friendly
- Do NOT use `roleType: oidc` (the default) in the JWT example — this is a JWT-mode example, must set `roleType: jwt` explicitly
- Do NOT use YAML lists for `azureRoles` — this is a JSON-encoded string in the CRD
- Do NOT use the same mount path (`azure`) for both the Azure auth engine and Azure secret engine — use `azure-auth` and `azure-se` to avoid collision
- Do NOT make README.md excessively long — keep it focused and scannable, under 100 lines
- Do NOT include a `connection` block in examples (optional, most users don't need it)
- Do NOT use `OIDCDiscoveryURL` or `OIDCCredentials` in the JWT example — this is JWT mode (JWKS), not OIDC mode

### README.md Template Pattern

Each README.md should follow this structure:
```markdown
# <Scenario Title>

## Scenario
<2-3 sentence description of what this example demonstrates>

## Prerequisites
- <bullet list of requirements>

## Resources Created
<table or list of CRDs and what each one does>

## Apply
<kubectl apply commands>

## Verify
<vault CLI verification commands>

## How It Works
<brief explanation of how auth → policy → secret engine flow works>

## Cleanup
<kubectl delete commands + note about what persists in Vault>
```

### Previous Story Intelligence (D4.1 and D4.2)

D4.1 and D4.2 established these conventions for D4 examples:
- Same authentication block pattern (`path: kubernetes`, `role: policy-admin`) for the operator's own auth
- Multi-document YAML with `---` separators
- Comments are concise: explain "what" and "when to change", not just field name repetition
- Placeholder values use consistent patterns: `example.com`, `00000000-...`, `my-<thing>`
- No `connection` block (optional)
- No `namespace` in metadata

### Project Structure Notes

- All files go under `docs/examples/e2e-<name>/`
- Two files per directory: one README.md + one YAML file
- No Go code changes, no Makefile changes, no CRD changes
- Pure documentation/examples — no build, test, or runtime impact
- The existing `docs/examples/postgresql/` directory is **not modified**

### Git Intelligence

Recent commits are documentation-focused (Epic D2, D3). No code changes affecting this story. The codebase is stable for documentation work.

### References

- [Source: docs/auth-engines/jwt-oidc.md] — JWTOIDCAuthEngineConfig and Role field details, JWT vs OIDC modes, JWKS validation
- [Source: docs/secret-engines/pki.md] — PKISecretEngineConfig and Role field details, root CA generation, URL/CRL config
- [Source: docs/auth-engines/azure.md] — AzureAuthEngineConfig and Role field details, credential resolution
- [Source: docs/secret-engines/azure.md] — AzureSecretEngineConfig and Role field details, JSON-encoded azureRoles
- [Source: docs/examples/postgresql/postgresql-secret-engine.yaml] — Reference pattern for multi-document example YAML
- [Source: docs/examples/postgresql/namespace-config.yaml] — Reference for Policy CRD usage in end-to-end context
- [Source: _bmad-output/implementation-artifacts/d4-1-create-example-yaml-files-for-each-auth-engine.md] — D4.1 story with auth engine example conventions
- [Source: _bmad-output/implementation-artifacts/d4-2-create-example-yaml-files-for-each-secret-engine.md] — D4.2 story with secret engine example conventions

## Dev Agent Record

### Agent Model Used

{{agent_model_name_version}}

### Debug Log References

### Completion Notes List

### File List
