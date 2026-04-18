# Story 2.4: Add Update Scenarios to PKI and DatabaseSecretEngineStaticRole Integration Tests

Status: done

## Story

As an operator developer,
I want integration tests that modify PKI config/role and DB static role specs,
So that these update paths are validated.

## Acceptance Criteria

1. **Given** a PKISecretEngineConfig that has been successfully reconciled **When** I update a CRL config field (e.g., `crlExpiry`) **Then** the reconciler updates the PKI CRL config in Vault **And** the ReconcileSuccessful condition is updated with a new ObservedGeneration

2. **Given** a PKISecretEngineRole that has been successfully reconciled **When** I update a role field (e.g., `allowedDomains`) **Then** the reconciler updates the role in Vault **And** the ReconcileSuccessful condition is updated with a new ObservedGeneration

3. **Given** a DatabaseSecretEngineStaticRole that has been successfully reconciled **When** I update a spec field (e.g., `rotationStatements`) **Then** the reconciler updates the static role in Vault **And** the ReconcileSuccessful condition is updated with a new ObservedGeneration

## Tasks / Subtasks

- [x] Task 1: Add update scenario for PKISecretEngineRole (AC: 2)
  - [x] 1.1: In `controllers/pkisecretengine_controller_test.go`, add a new `Context("When updating a PKISecretEngineRole")` block after the existing `Context("When creating a PKISecretEngineRole")` (line 155)
  - [x] 1.2: Load the PKISecretEngineRole from `../test/pkisecretengine/pki-secret-engine-role.yaml`, set namespace to `vaultTestNamespaceName`, create it, wait for `ReconcileSuccessful=True`
  - [x] 1.3: Read the role from Vault via `vaultClient.Logical().Read("test-vault-config-operator/pki/roles/pki-example")` and verify the initial `allowed_domains` contains `["internal.io", "pki-vault-demo.svc", "example.com"]`
  - [x] 1.4: Record the initial `ObservedGeneration` from the `ReconcileSuccessful` condition
  - [x] 1.5: `Get()` the latest PKISecretEngineRole (fresh ResourceVersion), append `"test.io"` to `Spec.AllowedDomains`, change `Spec.MaxTTL` from `"8760h"` to `"4380h"`, call `k8sIntegrationClient.Update(ctx, instance)`
  - [x] 1.6: Use `Eventually` (timeout 120s, interval 2s) to poll Vault at `test-vault-config-operator/pki/roles/pki-example` until `allowed_domains` includes `"test.io"`
  - [x] 1.7: Verify `max_ttl` reflects the updated value
  - [x] 1.8: Verify the `ReconcileSuccessful` condition's `ObservedGeneration` is greater than the initial value

- [x] Task 2: Add update scenario for PKISecretEngineConfig CRL config (AC: 1)
  - [x] 2.1: In `controllers/pkisecretengine_controller_test.go`, add a new `Context("When updating a PKISecretEngineConfig CRL config")` block after the new role update Context
  - [x] 2.2: `Get()` the existing PKISecretEngineConfig (already created by the "When creating a PKISecretEngineConfig" Context, name is `pki`)
  - [x] 2.3: Read the CRL config from Vault via `vaultClient.Logical().Read("test-vault-config-operator/pki/config/crl")` and record the initial `expiry` and `disable` values
  - [x] 2.4: Record the initial `ObservedGeneration` from the `ReconcileSuccessful` condition
  - [x] 2.5: `Get()` the latest PKISecretEngineConfig (fresh ResourceVersion), set `Spec.CRLExpiry` to `"48h"` and `Spec.CRLDisable` to `true`, call `k8sIntegrationClient.Update(ctx, instance)`
  - [x] 2.6: Use `Eventually` to poll Vault at `test-vault-config-operator/pki/config/crl` until the `expiry` field changes to reflect `"48h"` or the `disable` field becomes `true`
  - [x] 2.7: Verify the `ReconcileSuccessful` condition's `ObservedGeneration` is greater than the initial value

- [x] Task 3: Add DatabaseSecretEngineStaticRole create + update scenario (AC: 3)
  - [x] 3.1: Deploy PostgreSQL to Kind cluster via Bitnami Helm chart with `helloworld` user (integration/postgresql-values.yaml, Makefile deploy-postgresql target)
  - [x] 3.2: Update database-engine-config.yaml: add `read-only-static` to allowedRoles, switch rootCredentials from randomSecret to K8s secret
  - [x] 3.3: Create PostgreSQL root credentials K8s Secret in test setup
  - [x] 3.4: Add `Context("When creating a DatabaseSecretEngineStaticRole")` — creates static role, verifies Vault state (db_name, username)
  - [x] 3.5: Add `Context("When updating a DatabaseSecretEngineStaticRole")` — updates rotationPeriod from 3600→7200, verifies Vault reflects change
  - [x] 3.6: Verify ObservedGeneration increases after update

- [x] Task 4: Add DatabaseSecretEngineStaticRole to the deletion cleanup (AC: 3)
  - [x] 4.1: Delete DatabaseSecretEngineStaticRole before DatabaseSecretEngineConfig in cleanup Context
  - [x] 4.2: Delete PostgreSQL root credentials K8s Secret in cleanup

- [x] Task 5: Restructure PKI deletion to include the role update cleanup (AC: 2)
  - [x] 5.1: The PKI role created in Task 1 uses the same fixture as the existing create Context. The existing deletion Context already deletes it. Verify the deletion Context handles both the originally-created and update-tested role (they're the same object — the update modifies in-place, deletion removes the same object)
  - [x] 5.2: No additional deletion code needed if the object names match

### Review Findings

- [x] [Review][Decision] **RESOLVED** — Chose option (a): deployed PostgreSQL to Kind, fixed fixtures, and implemented full static role create+update+delete coverage. All 3 ACs now satisfied.
- [x] [Review][Patch] **RESOLVED** — Added CRL expiry assertion (`ContainSubstring("48h")`) after the disable check in pkisecretengine_controller_test.go.

## Dev Notes

### PKISecretEngineRole Uses Standard VaultResource — Straightforward Update

PKISecretEngineRole uses the standard `VaultResource` reconciler with `CreateOrUpdate` flow:
1. `prepareContext()` → enriches context
2. `NewVaultResource` → standard reconciler
3. `CreateOrUpdate()` reads Vault at `{path}/roles/{name}`, calls `IsEquivalentToDesiredState(payload)`, writes only if false
4. `ManageOutcome` sets `ReconcileSuccessful` with `ObservedGeneration`

`IsEquivalentToDesiredState` uses `reflect.DeepEqual(PKIRole.toMap(), payload)`. When we change `allowedDomains`, `toMap()` produces a different value than Vault's read response, so the write proceeds.

`GetPath()` returns `{path}/roles/{name}` — for the test fixture, this is `test-vault-config-operator/pki/roles/pki-example`.

**Webhook restriction:** Only `spec.path` is immutable. All other role fields can be updated freely.

[Source: controllers/pkisecretenginerole_controller.go — standard VaultResource]
[Source: api/v1alpha1/pkisecretenginerole_types.go#L65-L77 — GetPath, GetPayload, IsEquivalentToDesiredState]
[Source: api/v1alpha1/pkisecretenginerole_webhook.go#L62-L72 — only path immutable]

### PKISecretEngineConfig Uses VaultPKIEngineResource — Non-Standard Reconcile

PKISecretEngineConfig uses `VaultPKIEngineResource` (NOT the standard `VaultResource`). The reconcile flow is:

1. `PrepareInternalValues()` → no-op
2. `Generate()` → one-time CA generation (skipped when `Status.Generated == true`)
3. `CreateIntermediate()` → one-time intermediate setup (skipped for root type or when `Status.Signed == true`)
4. `CreateOrUpdateConfigUrls()` → writes URL config to `{path}/config/urls` every reconcile
5. `CreateOrUpdateConfigCrl()` → writes CRL config to `{path}/config/crl` every reconcile

**Only steps 4 and 5 run on subsequent reconciles.** PKICommon fields (commonName, TTL, SANs) are only applied during the one-time `Generate()` step.

**Webhook restriction:** `spec.path`, `spec.type`, and `spec.privateKeyType` are immutable. Other fields including PKIConfig (URLs, CRL) can be updated.

[Source: controllers/pkisecretengineconfig_controller.go — uses NewVaultPKIEngineResource]
[Source: controllers/vaultresourcecontroller/vaultpkiengineresourcereconciler.go#L98-L151 — manageReconcileLogic]
[Source: api/v1alpha1/pkisecretengineconfig_webhook.go#L62-L79 — path, type, privateKeyType immutable]

### CRITICAL BUG in CreateOrUpdateConfig — URLs Write to CRL Path

`VaultPKIEngineEndpoint.CreateOrUpdateConfig` has two bugs that affect URL config writes:

```go
func (ve *VaultPKIEngineEndpoint) CreateOrUpdateConfig(context context.Context, configPath string, payload map[string]interface{}) error {
    currentConfigPayload, err := ve.readConfig(context, configPath)
    if !ve.vaultObject.IsEquivalentToDesiredState(currentConfigPayload) {
        return write(context, ve.vaultPKIEngineObject.GetConfigCrlPath(), payload) // BUG: always CRL path
    }
    return nil
}
```

**Bug 1:** The write always goes to `GetConfigCrlPath()` regardless of which config is being written. When called via `CreateOrUpdateConfigUrls`, the URLs payload is written to `{path}/config/crl` instead of `{path}/config/urls`.

**Bug 2:** `IsEquivalentToDesiredState(currentConfigPayload)` compares the URL/CRL config payload against `PKICommon.toMap()` (the CA generate payload). These are completely different field sets, so the comparison always returns `false`, meaning writes always occur.

**Impact on this story:**
- CRL config updates ARE applied to the correct path (Bug 1 accidentally produces correct behavior for CRL)
- URL config updates are written to the wrong path (broken)
- The IsEquivalentToDesiredState check is useless (always writes)

**The update test for PKISecretEngineConfig should update CRL config fields** (`crlExpiry`, `crlDisable`) which do work correctly. URL config field updates (`issuingCertificates`, `crlDistributionPoints`) are broken and should be deferred to Story 7-4 or a dedicated bug fix story.

[Source: api/v1alpha1/utils/vautlpkiengineobject.go#L105-L118 — CreateOrUpdateConfig bug]

### PKISecretEngineConfig Fields Available for CRL Update Test

The PKIConfigCRL struct has two fields:

```go
type PKIConfigCRL struct {
    CRLExpiry  string `json:"crlExpiry,omitempty"`
    CRLDisable bool   `json:"crlDisable,omitempty"`
}
```

`toMap()` produces: `{"expiry": CRLExpiry, "disable": CRLDisable}`.

The test fixture does not set these fields, so they'll be at zero values (empty string, false). The update test should set `crlExpiry: "48h"` and verify Vault reflects the change at `{path}/config/crl`.

### Vault API Read Paths and Response Shapes

**PKI Role read:** `GET {path}/roles/{name}` returns:
```json
{
  "data": {
    "allowed_domains": ["internal.io", "pki-vault-demo.svc", "example.com"],
    "allow_subdomains": true,
    "allowed_other_sans": ["*"],
    "allow_glob_domains": true,
    "allowed_uri_sans": ["*-pki-vault-demo.apps.example.com"],
    "max_ttl": "8760h",
    ...extra fields Vault adds (key_type, key_bits, etc.)
  }
}
```

**PKI CRL config read:** `GET {path}/config/crl` returns:
```json
{
  "data": {
    "expiry": "72h",
    "disable": false,
    ...possible extra fields
  }
}
```

**Database static role read:** `GET {path}/static-roles/{name}` returns:
```json
{
  "data": {
    "db_name": "my-postgresql-database",
    "username": "helloworld",
    "rotation_period": 3600,
    "rotation_statements": ["ALTER USER \"{{name}}\" WITH PASSWORD '{{password}}';"],
    "credential_type": "password",
    ...extra fields
  }
}
```

Use `secret.Data["field"]` to extract fields. Vault returns `[]interface{}` for list fields.

### DatabaseSecretEngineStaticRole — Missing from Existing Test

The existing `databasesecretenginestaticrole_controller_test.go` sets up infrastructure (Policy, KubeAuthRole, SecretEngineMount, PasswordPolicy, RandomSecret, DatabaseSecretEngineConfig) but **never creates a DatabaseSecretEngineStaticRole**. The static role itself is not exercised.

This story adds both create and update coverage for DatabaseSecretEngineStaticRole.

**Prerequisites for the static role:**
1. The database secret engine must be configured (DatabaseSecretEngineConfig reconciled)
2. The PostgreSQL database must have a user matching the static role's `username` field
3. The test fixture `test/database-engine-read-only-static-role.yaml` uses `username: helloworld` and `rotationPeriod: 3600`

**Critical dependency:** The PostgreSQL test instance must have a `helloworld` user. Check the database setup in the integration test infrastructure:
- If the user exists → create the static role directly
- If the user doesn't exist → either create it as part of test setup using `vaultClient.Logical().Write()` to Vault's database execute endpoint, or skip the static role tests and defer to a later story

The safest approach: check if the existing `DatabaseSecretEngineConfig` integration test (which does pass) has already set up the required database users. If not, the static role test should be marked as a stretch goal with a clear TODO.

### Test Fixture Analysis

**PKI Role fixture** (`test/pkisecretengine/pki-secret-engine-role.yaml`):
```yaml
spec:
  path: test-vault-config-operator/pki
  allowedDomains: [internal.io, pki-vault-demo.svc, example.com]
  allowSubdomains: true
  allowedOtherSans: "*"
  allowGlobDomains: true
  allowedURISans: ["*-pki-vault-demo.apps.example.com"]
  maxTTL: "8760h"
```

**Update plan:** Add `"test.io"` to `allowedDomains`, change `maxTTL` to `"4380h"`.

**PKI Config fixture** (`test/pkisecretengine/pki-secret-engine-config.yaml`):
```yaml
spec:
  path: test-vault-config-operator/pki
  commonName: pki-vault-demo.internal.io
  TTL: "8760h"
  type: root
  privateKeyType: internal
  issuingCertificates: [https://vault-internal.vault.svc:8200/v1/test-vault-config-operator/pki/ca]
  crlDistributionPoints: [https://vault-internal.vault.svc:8200/v1/test-vault-config-operator/pki/crl"]
```

CRL fields not in fixture — they'll be zero values initially. **Update plan:** Set `crlExpiry: "48h"`.

**DB Static Role fixture** (`test/database-engine-read-only-static-role.yaml`):
```yaml
spec:
  path: test-vault-config-operator/database
  dBName: my-postgresql-database
  username: helloworld
  rotationPeriod: 3600
  rotationStatements: ['ALTER USER "{{name}}" WITH PASSWORD ''{{password}}'';']
  credentialType: password
  passwordCredentialConfig: {}
```

**Update plan:** Add a second rotation statement.

### PKISecretEngineRole toMap() — Many Fields, Vault Returns Extras

`PKIRole.toMap()` produces ~40 fields including `allowed_domains`, `allow_subdomains`, `max_ttl`, `key_type`, `key_bits`, etc. Vault's read response will contain these plus additional Vault-managed fields. `IsEquivalentToDesiredState` uses `reflect.DeepEqual` on the full `toMap()` output vs the Vault payload. Due to the extra-fields issue (Story 7-4), this comparison may always return `false`, causing writes on every reconcile. For this test, this means the update will always be written, which is fine for validating the end-to-end flow.

[Source: api/v1alpha1/pkisecretenginerole_types.go#L325-L366 — PKIRole.toMap()]

### Existing Test Structure — Sequential Contexts with Shared State

The PKI test file uses sequential `Context` blocks that share state across the test:

1. `"When preparing a PKI Secret Engine"` — Creates Policy, KubeAuthRole, SecretEngineMount
2. `"When creating a PKISecretEngineConfig"` — Creates PKISecretEngineConfig
3. `"When creating a PKISecretEngineRole"` — Creates PKISecretEngineRole
4. `"When deleting a PKISecretEngineRole"` — Deletes all resources

All Contexts are in a single `Describe` block. Resources created in earlier Contexts persist into later ones. The update tests should be inserted between Contexts 3 and 4.

**Important:** The PKI update Contexts DON'T need to create new resources — they update the resources already created by Contexts 2 and 3. Get the existing objects by name/namespace and update them.

Similarly for the DB static role test: the database engine infrastructure is created in the existing "When preparing" and "When creating" Contexts. The static role create+update should come before the deletion Context.

### Get Before Update — Critical Pattern

Always `Get()` immediately before `Update()` to get the latest ResourceVersion:

```go
By("Getting the latest PKISecretEngineRole")
lookupKey := types.NamespacedName{Name: "pki-example", Namespace: vaultTestNamespaceName}
created := &redhatcopv1alpha1.PKISecretEngineRole{}
Expect(k8sIntegrationClient.Get(ctx, lookupKey, created)).Should(Succeed())

By("Updating the PKISecretEngineRole spec")
created.Spec.AllowedDomains = append(created.Spec.AllowedDomains, "test.io")
created.Spec.MaxTTL = "4380h"
Expect(k8sIntegrationClient.Update(ctx, created)).Should(Succeed())
```

### Verifying Vault State After Update via Polling

```go
Eventually(func() bool {
    secret, err := vaultClient.Logical().Read("test-vault-config-operator/pki/roles/pki-example")
    if err != nil || secret == nil {
        return false
    }
    domains, ok := secret.Data["allowed_domains"].([]interface{})
    if !ok {
        return false
    }
    for _, d := range domains {
        if d == "test.io" {
            return true
        }
    }
    return false
}, timeout, interval).Should(BeTrue())
```

### PKISecretEngineConfig IsDeletable Returns false

PKISecretEngineConfig's `IsDeletable()` returns `false`, so no Vault cleanup happens on CR deletion. The existing delete test does a `k8sIntegrationClient.Delete()` + wait for Vault nil, but the nil comes from the `SecretEngineMount` being deleted (which disables the entire PKI engine), not from the Config object itself.

### PKI Role Response — List Fields as []interface{}

Vault returns `allowed_domains` as `[]interface{}` (not `[]string`). Use type assertion or gomega matchers that handle interface slices:

```go
domains := secret.Data["allowed_domains"].([]interface{})
Expect(domains).To(ContainElement("test.io"))
```

### File Inventory — What Needs to Change

| # | File | Change Type | Description |
|---|------|-------------|-------------|
| 1 | `controllers/pkisecretengine_controller_test.go` | Modify | Add `Context("When updating a PKISecretEngineRole")` and `Context("When updating a PKISecretEngineConfig CRL config")` between create and delete Contexts |
| 2 | `controllers/databasesecretenginestaticrole_controller_test.go` | Modify | Add `Context("When creating and updating a DatabaseSecretEngineStaticRole")` before the deletion Context; add static role deletion to cleanup |

No new fixtures needed — updates are performed in-code. No decoder changes needed — `GetPKISecretEngineConfigInstance`, `GetPKISecretEngineRoleInstance`, and `GetDatabaseSecretEngineStaticRoleInstance` already exist.

### No `make manifests generate` Needed

This story only modifies integration test files. No CRD types, controllers, or webhooks are changed.

### Import Requirements

Both test files already import all necessary packages:
- `vault "github.com/hashicorp/vault/api"` — available via `vaultClient` from suite
- `redhatcopv1alpha1` — already imported
- `vaultresourcecontroller` — already imported for `ReconcileSuccessful` constant
- `metav1` — already imported
- `types` — already imported

### DatabaseSecretEngineStaticRole Test Fixture Validation

The fixture `test/database-engine-read-only-static-role.yaml` uses:
- `rotationPeriod: 3600` (int, in seconds — matches the Go type `int`)
- `credentialType: password` — required by webhook validation
- `passwordCredentialConfig: {}` — satisfies the `validateEitherPasswordOrKey` constraint
- `username: helloworld` — must exist in the PostgreSQL database

**If the PostgreSQL user doesn't exist:** The DatabaseSecretEngineStaticRole reconcile will fail with a Vault error because Vault will try to manage the password for a non-existent user. In this case:
1. First check if the user exists by reading the database config
2. If needed, create the user via direct SQL execution through the database connection (or use `vaultClient` to create a dynamic credential first, which creates the user)
3. Alternatively, use `../test/databasesecretengine/database-secret-engine-static-role.yaml` but fix the `rotationPeriod` type issue (it uses `24h` string, but Go type is `int`)

### Previous Story Intelligence

**From Story 2.3 (ready-for-dev):**
- Established the update pattern for standard VaultObject types: Create → Wait → Get-before-Update → Modify spec → Update → Poll Vault → Verify ObservedGeneration → Cleanup
- Entity uses explicit `delete()` of 11 Vault-added keys; PKISecretEngineRole uses bare `reflect.DeepEqual` (same extra-field concern from Story 7-4)
- EntityAlias `customMetadata` was the only safe user-facing field change; for PKI role, multiple fields are safe to change

**From Story 2.2 (ready-for-dev):**
- RandomSecret has a unique refresh guard; PKISecretEngineRole and DatabaseSecretEngineStaticRole use the standard VaultResource flow
- The `Eventually` timeout/interval pattern (120s/2s) is well-established

**From Story 2.1 (ready-for-dev):**
- First update scenario test in the codebase, established Get → modify → Update → Eventually poll pattern
- VaultSecret is non-standard; PKISecretEngineRole is standard VaultObject

**From Story 2.0 (ready-for-dev):**
- Story 2.0 stabilizes integration test infrastructure (idempotent Kind, namespace handling)
- Story 2.0 MUST complete before this story
- Namespace create-or-get pattern prevents re-run failures

**From Epic 1 Retrospective:**
- "Pattern-first investment pays dividends" — this story reuses the update pattern from Stories 2.1-2.3
- PKI unit tests (Story 1.5) verified `IsEquivalentToDesiredState` for PKISecretEngineRole toMap()
- DatabaseSecretEngineStaticRole unit tests (Story 1.3) verified toMap() and IsEquivalentToDesiredState

### Git Intelligence (Recent Commits)

```
910acbd Complete Epic 1 retrospective and fix identified tech debt
f1e57e7 Update push.yaml with permissions for nested workflow
cd7e5b8 Pre-load busybox image into kind to avoid Docker Hub rate limits in CI
511af21 Fix helmchart-test hang: add wget timeout and fix sidecar script portability
9110587 Add integration test philosophy rule and Story 2.0 for infrastructure stabilization
```

Codebase is clean for Epic 2. Commit `910acbd` resolved GroupAlias debug prints and KubernetesSecretEngineRole field mapping.

### Risk Considerations

- **PKI config CreateOrUpdateConfig bug:** The URL config write path is broken (writes to CRL path). The CRL update test works correctly but URLs do not. Document this as a finding and defer URL update testing. Consider filing a separate bug fix story.
- **PKI role extra-field issue:** `IsEquivalentToDesiredState` uses bare `reflect.DeepEqual`. If Vault returns extra fields in the role read response, the comparison always returns `false`, causing writes on every reconcile. The update test still validates the flow (spec change → Vault reflects new state) even if unnecessary writes occur.
- **DatabaseSecretEngineStaticRole PostgreSQL user:** The fixture references `username: helloworld`. If this user doesn't exist in the test PostgreSQL instance, the static role reconcile will fail. **Mitigation:** Check at test runtime; if Vault returns an error on static role creation, the test should clearly report it's a missing test infrastructure issue.
- **DatabaseSecretEngineStaticRole rotation behavior:** Creating a static role triggers an immediate password rotation in Vault. The test must account for this. If the database user doesn't have the right permissions for the rotation statement, the reconcile will fail.
- **Resource conflicts on Update:** A reconcile between Get and Update can cause a ResourceVersion conflict. Do Get→Update without delay. Standard practice across Stories 2.1-2.3.
- **PKI test sequential Context dependency:** The PKI update Contexts depend on resources created in earlier Contexts. If the create Contexts fail, the update Contexts will also fail. This is the existing pattern in the test file.

### Deferred Work

The following items are explicitly out of scope for this story:
- **Fix CreateOrUpdateConfig bug** (URLs written to CRL path) — file as tech debt or address in Story 7-4
- **Fix IsEquivalentToDesiredState for PKI config** (compares against wrong toMap) — same as above
- **URL config update integration test** — blocked by CreateOrUpdateConfig bug
- **PostgreSQL user provisioning automation** — if the test PostgreSQL doesn't have the required user, defer DB static role tests to when the infrastructure is enhanced

### References

- [Source: controllers/pkisecretengine_controller_test.go] — Existing PKI integration test (262 lines, 4 sequential Contexts)
- [Source: controllers/databasesecretenginestaticrole_controller_test.go] — Existing DB static role test (348 lines, creates infrastructure but not the static role)
- [Source: controllers/pkisecretengineconfig_controller.go] — PKISecretEngineConfig controller (uses VaultPKIEngineResource)
- [Source: controllers/pkisecretenginerole_controller.go] — PKISecretEngineRole controller (uses VaultResource)
- [Source: controllers/vaultresourcecontroller/vaultpkiengineresourcereconciler.go#L98-L151] — VaultPKIEngineResource.manageReconcileLogic
- [Source: api/v1alpha1/pkisecretengineconfig_types.go#L427-L442] — PKISecretEngineConfig GetPayload, IsEquivalentToDesiredState
- [Source: api/v1alpha1/pkisecretengineconfig_types.go#L552-L569] — PKIConfigUrls.toMap() and PKIConfigCRL.toMap()
- [Source: api/v1alpha1/pkisecretenginerole_types.go#L65-L77] — PKISecretEngineRole GetPath, GetPayload, IsEquivalentToDesiredState
- [Source: api/v1alpha1/pkisecretenginerole_types.go#L325-L366] — PKIRole.toMap()
- [Source: api/v1alpha1/databasesecretenginestaticrole_types.go#L123-L137] — DBSEStaticRole.toMap()
- [Source: api/v1alpha1/databasesecretenginestaticrole_types.go#L151-L171] — DatabaseSecretEngineStaticRole GetPath, GetPayload, IsEquivalentToDesiredState
- [Source: api/v1alpha1/utils/vautlpkiengineobject.go#L79-L118] — CreateOrUpdateConfigUrls/Crl/Config (contains bug)
- [Source: api/v1alpha1/pkisecretengineconfig_webhook.go#L62-L79] — PKISecretEngineConfig webhook (path, type, privateKeyType immutable)
- [Source: api/v1alpha1/pkisecretenginerole_webhook.go#L62-L72] — PKISecretEngineRole webhook (path immutable)
- [Source: api/v1alpha1/databasesecretenginestaticrole_webhook.go] — DatabaseSecretEngineStaticRole webhook (path immutable, credential validation)
- [Source: test/pkisecretengine/pki-secret-engine-config.yaml] — PKI config test fixture
- [Source: test/pkisecretengine/pki-secret-engine-role.yaml] — PKI role test fixture
- [Source: test/database-engine-read-only-static-role.yaml] — DB static role test fixture
- [Source: controllers/controllertestutils/decoder.go#L130-L188] — All three decoder methods exist
- [Source: controllers/suite_integration_test.go#L171-L175] — PKI controller registration
- [Source: controllers/suite_integration_test.go#L198-L199] — DatabaseSecretEngineStaticRole controller registration
- [Source: _bmad-output/implementation-artifacts/2-3-add-update-scenarios-to-entity-and-entityalias-integration-tests.md] — Story 2.3 (update pattern reference)
- [Source: _bmad-output/implementation-artifacts/2-0-stabilize-integration-test-infrastructure.md] — Story 2.0 (prerequisite)
- [Source: _bmad-output/planning-artifacts/epics.md#L333-L348] — Story 2.4 epic definition

## Dev Agent Record

### Agent Model Used

Claude Opus 4 (claude-sonnet-4-20250514)

### Debug Log References

- Initial integration test run failed for DatabaseSecretEngineStaticRole: Vault returned `"read-only-static" is not an allowed role` (HTTP 500). Root cause: fixture role name not in DatabaseSecretEngineConfig's `allowedRoles`, AND no PostgreSQL instance deployed in test infrastructure. Static roles require a real database connection for password rotation.
- Build error: `MaxTTL` field is `metav1.Duration` not `string`. Fixed by using `metav1.Duration{Duration: 4380 * time.Hour}`.

### Completion Notes List

- ✅ AC 1 satisfied: PKISecretEngineConfig CRL config update test added — updates `CRLExpiry` to 48h and `CRLDisable` to true, verifies Vault reflects both changes (including expiry) and ObservedGeneration increases.
- ✅ AC 2 satisfied: PKISecretEngineRole update test added — appends "test.io" to allowedDomains, changes maxTTL to 4380h, verifies Vault reflects both changes and ObservedGeneration increases.
- ✅ AC 3 satisfied: DatabaseSecretEngineStaticRole create+update tests added — creates static role with `helloworld` user, verifies Vault state (db_name, username), updates rotationPeriod from 3600→7200, verifies Vault reflects change and ObservedGeneration increases. Full deletion cleanup included.
- PostgreSQL deployed to Kind cluster via Bitnami Helm chart with init script creating `helloworld` user.
- Database engine config fixture updated: `allowedRoles` includes `read-only-static`, rootCredentials switched from RandomSecret to K8s Secret for real PostgreSQL connectivity.
- All 3 acceptance criteria fully satisfied.

### Change Log

- 2026-04-17: Added PKISecretEngineRole and PKISecretEngineConfig CRL update integration tests (Tasks 1, 2, 5). Deferred DatabaseSecretEngineStaticRole tests (Tasks 3, 4) due to missing PostgreSQL test infrastructure.
- 2026-04-17: Implemented PostgreSQL test infrastructure and DatabaseSecretEngineStaticRole tests (Tasks 3, 4). Deployed PostgreSQL via Bitnami Helm chart, updated database engine config fixture, added static role create+update+delete test contexts. Resolved all review findings.

### File List

- `controllers/pkisecretengine_controller_test.go` — Modified: Added two new Context blocks for PKISecretEngineRole update and PKISecretEngineConfig CRL update tests; added CRL expiry assertion (patch fix)
- `controllers/databasesecretenginestaticrole_controller_test.go` — Modified: Added K8s Secret creation for PostgreSQL credentials, three new Context blocks for static role create/update/delete, corev1 import
- `test/databasesecretengine/database-engine-config.yaml` — Modified: Added `read-only-static` to allowedRoles, switched rootCredentials from randomSecret to K8s secret reference
- `integration/postgresql-values.yaml` — New: Bitnami PostgreSQL Helm values with `helloworld` user init script
- `Makefile` — Modified: Added `deploy-postgresql` target, wired into `integration` target chain
