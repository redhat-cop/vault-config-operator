# Story 5.1: Integration Tests for DatabaseSecretEngineConfig and DatabaseSecretEngineRole

Status: ready-for-dev

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As an operator developer,
I want integration tests for the Database secret engine config and role types covering create, reconcile success, Vault state verification, update, and delete with cleanup,
So that the most complex secret engine (with credential resolution and `IsEquivalentToDesiredState` field remapping) is verified end-to-end against a real PostgreSQL database in Kind.

## Acceptance Criteria

1. **Given** PostgreSQL is deployed in the Kind cluster via `deploy-postgresql` (Bitnami Helm) **And** a K8s Secret with PostgreSQL root credentials exists **And** a SecretEngineMount (type=database) has been created and reconciled **And** a DatabaseSecretEngineConfig CR is created targeting the database mount with `connectionURL` pointing at PostgreSQL **When** the reconciler processes it **Then** the database connection is configured in Vault at `{mount}/config/{name}` with `plugin_name`, `connection_details.connection_url`, and `connection_details.username` verified, and ReconcileSuccessful=True

2. **Given** a DatabaseSecretEngineRole CR is created with `dBName` referencing the config, `creationStatements`, `defaultTTL`, and `maxTTL` **When** the reconciler processes it **Then** the role exists in Vault at `{mount}/roles/{name}` with correct field values and ReconcileSuccessful=True

3. **Given** the DatabaseSecretEngineRole CR spec is updated (e.g., `maxTTL` changed) **When** the reconciler processes the update **Then** the Vault role reflects the updated value and `ObservedGeneration` increases

4. **Given** the DatabaseSecretEngineRole CR is deleted (IsDeletable=true) **When** the reconciler processes the deletion **Then** the role is removed from Vault and the CR is deleted from K8s

5. **Given** the DatabaseSecretEngineConfig CR is deleted (IsDeletable=true) **When** the reconciler processes the deletion **Then** the config is removed from Vault and the CR is deleted from K8s

## Tasks / Subtasks

- [ ] Task 1: Add decoder method for DatabaseSecretEngineRole (AC: 2, 3, 4)
  - [ ] 1.1: Add `GetDatabaseSecretEngineRoleInstance` to `controllers/controllertestutils/decoder.go`

- [ ] Task 2: Create test fixtures (AC: 1, 2)
  - [ ] 2.1: Create `test/databasesecretengine/test-db-mount.yaml` — SecretEngineMount with `type: database`, `path: test-dbse`, unique metadata name
  - [ ] 2.2: Create `test/databasesecretengine/test-db-config.yaml` — DatabaseSecretEngineConfig with PostgreSQL connection and K8s Secret credentials
  - [ ] 2.3: Create `test/databasesecretengine/test-db-role.yaml` — DatabaseSecretEngineRole with `dBName` referencing the config, creation statements, TTLs

- [ ] Task 3: Create integration test file (AC: 1, 2, 3, 4, 5)
  - [ ] 3.1: Create `controllers/databasesecretengine_controller_test.go` with `//go:build integration` tag
  - [ ] 3.2: Add prerequisite context — create PostgreSQL root credentials K8s Secret, create SecretEngineMount (type=database), wait for reconcile, verify `sys/mounts`
  - [ ] 3.3: Add context for DatabaseSecretEngineConfig — create, poll for ReconcileSuccessful=True, verify Vault state at `{mount}/config/{name}` including `connection_details` nested fields
  - [ ] 3.4: Add context for DatabaseSecretEngineRole — create, poll for ReconcileSuccessful=True, verify Vault state at `{mount}/roles/{name}`
  - [ ] 3.5: Add update context for DatabaseSecretEngineRole — update `maxTTL`, verify Vault reflects change, verify ObservedGeneration increased
  - [ ] 3.6: Add deletion context — delete role (IsDeletable=true, verify Vault cleanup), delete config (IsDeletable=true, verify Vault cleanup), delete mount, delete secret

- [ ] Task 4: End-to-end verification (AC: 1, 2, 3, 4, 5)
  - [ ] 4.1: Run `make integration` and verify new tests pass alongside all existing tests
  - [ ] 4.2: Verify no regressions — existing DatabaseSecretEngineStaticRole tests and all prior tests unaffected

## Dev Notes

### Infrastructure Scope — PostgreSQL Already Deployed (No New Infra)

Per the Epic 4 retrospective's readiness assessment:

> Story 5.1 — DatabaseSecretEngineConfig/Role | PostgreSQL | Already deployed (`deploy-postgresql`) | Low — no new infra

PostgreSQL is deployed to `test-vault-config-operator` namespace via Bitnami Helm chart. No new Makefile targets, manifests, or infrastructure changes needed.

PostgreSQL details (from `integration/postgresql-values.yaml`):
- Helm release: `postgresql`, chart `bitnami/postgresql`
- `fullnameOverride: my-postgresql-database` → Service: `my-postgresql-database.test-vault-config-operator.svc`
- Port: 5432
- `auth.postgresPassword: testpassword123`
- `auth.database: testdb`
- Init script creates user `helloworld` with password `helloworld`

The `deploy-postgresql` Makefile target is already wired into the `integration` target.

[Source: integration/postgresql-values.yaml — PostgreSQL Helm values]
[Source: Makefile#L174-L181 — deploy-postgresql target]
[Source: _bmad-output/implementation-artifacts/epic-4-retro-2026-04-23.md#L146 — Story 5.1 infra classification]

### Both Types Use VaultResource Reconciler — NOT VaultEngineResource

Both DatabaseSecretEngineConfig and DatabaseSecretEngineRole use `NewVaultResource` (not `NewVaultEngineResource`). The reconcile flow is:

1. `prepareContext()` enriches context with kubeClient, restConfig, vaultConnection, vaultClient
2. `NewVaultResource(&r.ReconcilerBase, instance)` creates the standard reconciler
3. `VaultResource.Reconcile()` → `manageReconcileLogic()`:
   - `PrepareInternalValues()` — resolves root credentials from K8s Secret (config) or no-op (role)
   - `PrepareTLSConfig()` — no-op for both types
   - `VaultEndpoint.CreateOrUpdate()` — reads from Vault, calls `IsEquivalentToDesiredState()`, writes if different
4. `ManageOutcome()` sets `ReconcileSuccessful` condition

**Extra logic in DatabaseSecretEngineConfig controller:** After successful reconcile, there is root password rotation logic (`RotateRootPassword`) that runs if `RootPasswordRotation.Enable == true`. The test does NOT need root password rotation (just basic config/role CRUD).

The DatabaseSecretEngineConfig controller also watches K8s Secrets and RandomSecrets for credential changes (re-queues config CRs on credential updates). The role controller has no extra watches.

[Source: controllers/databasesecretengineconfig_controller.go#L85-L87 — NewVaultResource]
[Source: controllers/databasesecretenginerole_controller.go#L70-L77 — NewVaultResource]

### DatabaseSecretEngineConfig — Key Implementation Details

**GetPath():**
```go
func (d *DatabaseSecretEngineConfig) GetPath() string {
    if d.Spec.Name != "" {
        return vaultutils.CleansePath(string(d.Spec.Path) + "/" + "config" + "/" + d.Spec.Name)
    }
    return vaultutils.CleansePath(string(d.Spec.Path) + "/" + "config" + "/" + d.Name)
}
```
For fixture with `path: test-dbse/test-db-mount`, `metadata.name: test-db-config` → `test-dbse/test-db-mount/config/test-db-config`

Note: Uses `metadata.name` (not `spec.name`) for the Vault path when `spec.name` is empty. Different from auth engine configs that use only `spec.path`.

[Source: api/v1alpha1/databasesecretengineconfig_types.go#L78-L83]

**IsDeletable(): true** — Finalizer added, Vault config deleted on CR deletion. Different from auth engine configs (which are IsDeletable=false).

[Source: api/v1alpha1/databasesecretengineconfig_types.go#L74-L76]

**toMap() — Vault write payload keys:**
`plugin_name`, `plugin_version`, `verify_connection`, `allowed_roles`, `root_credentials_rotate_statements`, `password_policy`, `connection_url`, `DatabaseSpecificConfig` entries, `username` (from spec or retrieved), `disable_escaping`, `password` (only if retrieved).

[Source: api/v1alpha1/databasesecretengineconfig_types.go#L364-L388]

**IsEquivalentToDesiredState() — CRITICAL: Field remapping for `connection_details`**

This is the most complex `IsEquivalentToDesiredState` in the entire codebase. Vault's read response for database configs restructures fields:
- Write sends: `connection_url`, `username`, `disable_escaping`, `root_credentials_rotate_statements` at top level
- Read returns: these fields nested under `connection_details` sub-map

```go
func (d *DatabaseSecretEngineConfig) IsEquivalentToDesiredState(payload map[string]interface{}) bool {
    // Forces re-reconcile when root rotation enabled but no rotation done yet
    if d.Spec.DBSEConfig.RootPasswordRotation != nil && d.Spec.DBSEConfig.RootPasswordRotation.Enable && d.Status.LastRootPasswordRotation.IsZero() {
        return false
    }
    desiredState := d.Spec.DBSEConfig.toMap()
    connectionDetails := map[string]interface{}{}
    connectionDetails["connection_url"] = desiredState["connection_url"]
    connectionDetails["disable_escaping"] = desiredState["disable_escaping"]
    connectionDetails["root_credentials_rotate_statements"] = desiredState["root_credentials_rotate_statements"]
    connectionDetails["username"] = desiredState["username"]
    desiredState["connection_details"] = connectionDetails
    delete(desiredState, "password")
    delete(desiredState, "connection_url")
    delete(desiredState, "username")
    delete(desiredState, "disable_escaping")

    filteredPayload := make(map[string]interface{})
    for key, value := range payload {
        if _, exists := desiredState[key]; exists || key == "connection_details" {
            filteredPayload[key] = value
        }
    }
    return reflect.DeepEqual(desiredState, filteredPayload)
}
```

[Source: api/v1alpha1/databasesecretengineconfig_types.go#L93-L119]

**PrepareInternalValues() — Root credential resolution:**

Always calls `setInternalCredentials`. Supports 3 credential sources:
1. **K8s Secret** (`RootCredentials.Secret`) — reads `usernameKey`/`passwordKey` from K8s Secret
2. **RandomSecret** (`RootCredentials.RandomSecret`) — reads from Vault KV via RandomSecret reference
3. **VaultSecret** (`RootCredentials.VaultSecret`) — reads directly from Vault path

For the test, use the K8s Secret path (simplest, same pattern as Epic 4 tests).

When `RootCredentials.Secret` is set:
- If `Username` is set in spec → `retrievedUsername = spec.Username`, `retrievedPassword = secret.Data[passwordKey]`
- If `Username` is empty → both username and password come from the K8s Secret

[Source: api/v1alpha1/databasesecretengineconfig_types.go#L133-L215]

**PrepareTLSConfig():** Returns nil (no-op).

**Webhook:**
- `Default()`: Log-only, no field defaults (Note: uses wrong logger name `authenginemountlog` — existing bug, do not fix in this story)
- `ValidateCreate()`: Calls `r.isValid()` → validates `RootCredentials` has exactly one credential source
- `ValidateUpdate()`: Checks immutable `spec.path`, then calls `r.isValid()`
- `ValidateDelete()`: No-op

[Source: api/v1alpha1/databasesecretengineconfig_webhook.go]

### DatabaseSecretEngineRole — Key Implementation Details

**GetPath():**
```go
func (d *DatabaseSecretEngineRole) GetPath() string {
    if d.Spec.Name != "" {
        return vaultutils.CleansePath(string(d.Spec.Path) + "/" + "roles" + "/" + d.Spec.Name)
    }
    return vaultutils.CleansePath(string(d.Spec.Path) + "/" + "roles" + "/" + d.Name)
}
```
For fixture with `path: test-dbse/test-db-mount`, `metadata.name: test-db-role` → `test-dbse/test-db-mount/roles/test-db-role`

[Source: api/v1alpha1/databasesecretenginerole_types.go#L70-L75]

**IsDeletable(): true** — Finalizer added, Vault role deleted on CR deletion.

[Source: api/v1alpha1/databasesecretenginerole_types.go#L62-L64]

**toMap() — 7 Vault keys:**
`db_name`, `default_ttl`, `max_ttl`, `creation_statements`, `revocation_statements`, `rollback_statements`, `renew_statements`

[Source: api/v1alpha1/databasesecretenginerole_types.go#L183-L193]

**IsEquivalentToDesiredState():** Bare `reflect.DeepEqual(desiredState, payload)` — NO filtering of extra keys. Vault's read response for roles may include extra keys not in `toMap()`. This means the comparison may return false on every reconcile. Same Story 7-4 tech debt as other types — does NOT affect test correctness.

[Source: api/v1alpha1/databasesecretenginerole_types.go#L79-L82]

**PrepareInternalValues():** Returns nil (no-op). No credential resolution needed for roles.

**PrepareTLSConfig():** Returns nil (no-op).

**Webhook:**
- `Default()`: Log-only (uses wrong logger name `authenginemountlog` — existing bug, do not fix)
- `ValidateCreate()`: No-op (NOTE: kubebuilder marker has `verbs=update` only — ValidateCreate is not registered for admission on create)
- `ValidateUpdate()`: Checks immutable `spec.path`
- `ValidateDelete()`: No-op

[Source: api/v1alpha1/databasesecretenginerole_webhook.go]

### Vault API Response Shapes

**GET `{mount}/config/{name}`** — Returns database config with `connection_details` nesting:
```json
{
  "data": {
    "plugin_name": "postgresql-database-plugin",
    "plugin_version": "",
    "connection_details": {
      "connection_url": "postgresql://{{username}}:{{password}}@my-postgresql-database.test-vault-config-operator.svc:5432",
      "username": "postgres",
      "disable_escaping": false,
      "root_credentials_rotate_statements": []
    },
    "allowed_roles": ["test-db-role"],
    "root_credentials_rotate_statements": [],
    "password_policy": "",
    "verify_connection": true
  }
}
```
Key: Fields are nested under `connection_details`, not at top level. The `password` is never returned. Extra fields may appear depending on Vault version.

**GET `{mount}/roles/{name}`** — Returns dynamic role config:
```json
{
  "data": {
    "db_name": "test-db-config",
    "default_ttl": 3600,
    "max_ttl": 86400,
    "creation_statements": ["CREATE ROLE ..."],
    "revocation_statements": [],
    "rollback_statements": [],
    "renew_statements": []
  }
}
```
Vault returns TTLs as `json.Number` (not int). Extra fields like `credential_type`, `credential_config` may appear.

### Verifying Vault State

**Config verification — MUST use `connection_details` nesting:**
```go
secret, err := vaultClient.Logical().Read("test-dbse/test-db-mount/config/test-db-config")
Expect(err).To(BeNil())
Expect(secret).NotTo(BeNil())

pluginName, ok := secret.Data["plugin_name"].(string)
Expect(ok).To(BeTrue(), "expected plugin_name to be a string")
Expect(pluginName).To(Equal("postgresql-database-plugin"))

connDetails, ok := secret.Data["connection_details"].(map[string]interface{})
Expect(ok).To(BeTrue(), "expected connection_details to be a map")
Expect(connDetails["connection_url"]).To(Equal("postgresql://{{username}}:{{password}}@my-postgresql-database.test-vault-config-operator.svc:5432"))
Expect(connDetails["username"]).To(Equal("postgres"))

allowedRoles, ok := secret.Data["allowed_roles"].([]interface{})
Expect(ok).To(BeTrue(), "expected allowed_roles to be []interface{}")
Expect(allowedRoles).To(ContainElement("test-db-role"))
```

**Role verification:**
```go
secret, err := vaultClient.Logical().Read("test-dbse/test-db-mount/roles/test-db-role")
Expect(err).To(BeNil())
Expect(secret).NotTo(BeNil())

dbName, ok := secret.Data["db_name"].(string)
Expect(ok).To(BeTrue(), "expected db_name to be a string")
Expect(dbName).To(Equal("test-db-config"))

creationStatements, ok := secret.Data["creation_statements"].([]interface{})
Expect(ok).To(BeTrue(), "expected creation_statements to be []interface{}")
Expect(creationStatements).To(HaveLen(1))
```

**Delete verification (both IsDeletable=true):**
```go
// Role
Eventually(func() bool {
    err := k8sIntegrationClient.Get(ctx, types.NamespacedName{...}, &redhatcopv1alpha1.DatabaseSecretEngineRole{})
    return apierrors.IsNotFound(err)
}, timeout, interval).Should(BeTrue())

Eventually(func() bool {
    secret, err := vaultClient.Logical().Read("test-dbse/test-db-mount/roles/test-db-role")
    return err == nil && secret == nil
}, timeout, interval).Should(BeTrue())

// Config
Eventually(func() bool {
    err := k8sIntegrationClient.Get(ctx, types.NamespacedName{...}, &redhatcopv1alpha1.DatabaseSecretEngineConfig{})
    return apierrors.IsNotFound(err)
}, timeout, interval).Should(BeTrue())

Eventually(func() bool {
    secret, err := vaultClient.Logical().Read("test-dbse/test-db-mount/config/test-db-config")
    return err == nil && secret == nil
}, timeout, interval).Should(BeTrue())
```

### Test Design — Dependency Chain

```
K8s Secret (test-db-pg-creds) — root credentials for PostgreSQL
  └── SecretEngineMount (test-db-mount, type=database, path=test-dbse)
        └── DatabaseSecretEngineConfig (test-db-config) → test-dbse/test-db-mount/config/test-db-config
        └── DatabaseSecretEngineRole (test-db-role) → test-dbse/test-db-mount/roles/test-db-role
```

Resources must be created in order: Secret → Mount → Config → Role. Deletion in reverse: Role → Config → Mount → Secret.

The SecretEngineMount must be reconciled before the config, because Vault rejects writes to `{mount}/config/{name}` if the engine mount doesn't exist.

The DatabaseSecretEngineRole depends on the config — the `dBName` field references the config name. Vault validates that `dBName` references an existing config when creating the role.

### PostgreSQL Root Credentials K8s Secret — Created in Test

The root credentials Secret should be created programmatically in the test's first `Context` block:

```go
pgSecret := &corev1.Secret{
    ObjectMeta: metav1.ObjectMeta{
        Name:      "test-db-pg-creds",
        Namespace: vaultAdminNamespaceName,
    },
    StringData: map[string]string{
        "username": "postgres",
        "password": "testpassword123",
    },
}
```

The fixture's `rootCredentials` references this secret with `usernameKey: username` and `passwordKey: password`. `PrepareInternalValues` will read both fields from the secret and set `retrievedUsername` and `retrievedPassword`, which are then included in the `toMap()` payload.

### Test Fixture Design

**Fixture 1: `test/databasesecretengine/test-db-mount.yaml`** — SecretEngineMount prerequisite:
```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: SecretEngineMount
metadata:
  name: test-db-mount
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  type: database
  path: test-dbse
```
Mounts at `sys/mounts/test-dbse/test-db-mount`. Uses `type: database` to enable the database secret engine. Uses `policy-admin` auth role (standard for integration tests in `vault-admin` namespace).

**Fixture 2: `test/databasesecretengine/test-db-config.yaml`** — DatabaseSecretEngineConfig:
```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: DatabaseSecretEngineConfig
metadata:
  name: test-db-config
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  path: test-dbse/test-db-mount
  pluginName: postgresql-database-plugin
  allowedRoles:
    - test-db-role
  connectionURL: "postgresql://{{username}}:{{password}}@my-postgresql-database.test-vault-config-operator.svc:5432"
  rootCredentials:
    secret:
      name: test-db-pg-creds
    usernameKey: username
    passwordKey: password
  username: postgres
  verifyConnection: true
```
`GetPath()` returns `test-dbse/test-db-mount/config/test-db-config` (uses `metadata.name` since no `spec.name`).

Key: `verifyConnection: true` means Vault will actually attempt to connect to PostgreSQL when writing the config. This validates the PostgreSQL deployment is reachable and the credentials work.

Key: `rootCredentials.secret` has custom key names and `username` is set in spec. `setInternalCredentials` will read `retrievedPassword` from the secret but use `spec.Username` for the username (per the logic: if `Username != ""`, use spec value).

**Fixture 3: `test/databasesecretengine/test-db-role.yaml`** — DatabaseSecretEngineRole:
```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: DatabaseSecretEngineRole
metadata:
  name: test-db-role
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  path: test-dbse/test-db-mount
  dBName: test-db-config
  defaultTTL: 1h
  maxTTL: 24h
  creationStatements:
    - "CREATE ROLE \"{{name}}\" WITH LOGIN PASSWORD '{{password}}' VALID UNTIL '{{expiration}}';"
```
`GetPath()` returns `test-dbse/test-db-mount/roles/test-db-role` (uses `metadata.name` since no `spec.name`).

`dBName: test-db-config` matches the DatabaseSecretEngineConfig's `metadata.name`, linking the role to the connection.

### Test Structure

```
Describe("DatabaseSecretEngine controllers", Ordered)
  var pgSecret *corev1.Secret
  var mountInstance *redhatcopv1alpha1.SecretEngineMount
  var configInstance *redhatcopv1alpha1.DatabaseSecretEngineConfig
  var roleInstance *redhatcopv1alpha1.DatabaseSecretEngineRole

  AfterAll: best-effort delete all instances + pg secret (reverse order)

  Context("When creating prerequisite resources")
    It("Should create the PostgreSQL credentials secret and database engine mount")
      - Create test-db-pg-creds K8s Secret in vault-admin namespace
      - Load test-db-mount.yaml via decoder.GetSecretEngineMountInstance
      - Set namespace to vaultAdminNamespaceName, create
      - Eventually poll for ReconcileSuccessful=True
      - Verify mount exists via sys/mounts key "test-dbse/test-db-mount/"

  Context("When creating a DatabaseSecretEngineConfig")
    It("Should write the database config to Vault")
      - Load test-db-config.yaml via decoder.GetDatabaseSecretEngineConfigInstance
      - Set namespace to vaultAdminNamespaceName, create
      - Eventually poll for ReconcileSuccessful=True
      - Read test-dbse/test-db-mount/config/test-db-config from Vault
      - Verify plugin_name = "postgresql-database-plugin"
      - Verify connection_details.connection_url contains PostgreSQL URL
      - Verify connection_details.username = "postgres"
      - Verify allowed_roles contains "test-db-role"

  Context("When creating a DatabaseSecretEngineRole")
    It("Should create the role in Vault with correct database settings")
      - Load test-db-role.yaml via decoder.GetDatabaseSecretEngineRoleInstance
      - Set namespace to vaultAdminNamespaceName, create
      - Eventually poll for ReconcileSuccessful=True
      - Read test-dbse/test-db-mount/roles/test-db-role
      - Verify db_name = "test-db-config"
      - Verify creation_statements has length 1

  Context("When updating a DatabaseSecretEngineRole")
    It("Should update the role in Vault and reflect updated ObservedGeneration")
      - Record initial ObservedGeneration
      - Get latest role CR, update maxTTL to 48h
      - Eventually verify Vault reflects updated max_ttl
      - Verify ObservedGeneration increased

  Context("When deleting DatabaseSecretEngine resources")
    It("Should clean up role and config from Vault and remove all resources")
      - Delete role CR (IsDeletable=true → Vault cleanup)
      - Eventually verify K8s deletion (NotFound)
      - Eventually verify role removed from Vault (Read returns nil)
      - Delete config CR (IsDeletable=true → Vault cleanup)
      - Eventually verify K8s deletion
      - Eventually verify config removed from Vault (Read returns nil)
      - Delete SecretEngineMount
      - Eventually verify K8s deletion and mount gone from sys/mounts
      - Delete PostgreSQL credentials secret
```

### Import Requirements

```go
import (
    "encoding/json"
    "time"

    . "github.com/onsi/ginkgo/v2"
    . "github.com/onsi/gomega"
    redhatcopv1alpha1 "github.com/redhat-cop/vault-config-operator/api/v1alpha1"
    "github.com/redhat-cop/vault-config-operator/controllers/vaultresourcecontroller"

    corev1 "k8s.io/api/core/v1"
    apierrors "k8s.io/apimachinery/pkg/api/errors"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/apimachinery/pkg/types"
)
```

`encoding/json` may be needed if checking `json.Number` for TTL values from Vault responses.

### Name Collision Prevention

Fixture names use `test-db-` prefix:
- `test-dbse/test-db-mount` — secret engine mount (unique path prefix)
- `test-db-config` — DatabaseSecretEngineConfig CR name and Vault config name
- `test-db-role` — DatabaseSecretEngineRole CR name and Vault role name
- `test-db-pg-creds` — PostgreSQL credentials K8s Secret

These don't collide with:
- `my-postgresql-database` — existing DatabaseSecretEngineConfig used by the static role test
- `test-vault-config-operator/database` — existing mount path used by static role test
- `read-only` / `read-only-static` — existing role/static-role names
- `postgresql-root-credentials` — K8s Secret used by static role test (different namespace too)
- Epic 4 test resources (`test-k8s-auth/*`, `test-ldap-auth/*`, `test-jwt-oidc-auth/*`)

### Existing Test Coexistence

The existing `databasesecretenginestaticrole_controller_test.go` (from Story 2.4) creates its own infrastructure (Policy, KubernetesAuthEngineRole, SecretEngineMount at `test-vault-config-operator/database`, DatabaseSecretEngineConfig `my-postgresql-database` in `test-vault-config-operator` namespace). The new test creates completely separate resources at a different mount path (`test-dbse`) in a different namespace (`vault-admin`), so there are zero conflicts.

Since Ginkgo v2 randomizes top-level Describe blocks, both tests will run independently regardless of ordering.

### Controller Registration — Already Done

Both controllers are registered in `suite_integration_test.go`:
```go
err = (&DatabaseSecretEngineConfigReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "DatabaseSecretEngineConfig")}).SetupWithManager(mgr)
Expect(err).ToNot(HaveOccurred())

err = (&DatabaseSecretEngineRoleReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "DatabaseSecretEngineRole")}).SetupWithManager(mgr)
```

No changes needed to the test suite setup.

[Source: controllers/suite_integration_test.go#L136-L139]

### Decoder Methods

**`GetDatabaseSecretEngineConfigInstance` — Already exists:**

[Source: controllers/controllertestutils/decoder.go#L175-L188]

**`GetDatabaseSecretEngineRoleInstance` — MUST BE ADDED:**

```go
func (d *decoder) GetDatabaseSecretEngineRoleInstance(filename string) (*redhatcopv1alpha1.DatabaseSecretEngineRole, error) {
    obj, groupKindVersion, err := d.decodeFile(filename)
    if err != nil {
        return nil, err
    }
    kind := reflect.TypeOf(redhatcopv1alpha1.DatabaseSecretEngineRole{}).Name()
    if groupKindVersion.Kind == kind {
        o := obj.(*redhatcopv1alpha1.DatabaseSecretEngineRole)
        return o, nil
    }
    return nil, errDecode
}
```

[Source: controllers/controllertestutils/decoder.go — existing pattern at lines 175-188]

### Vault TTL Format Gotcha

Vault returns TTL values as `json.Number`, not Go `int`. When verifying TTLs in the Vault response:

```go
maxTTL, ok := secret.Data["max_ttl"].(json.Number)
Expect(ok).To(BeTrue(), "expected max_ttl to be json.Number")
val, err := maxTTL.Int64()
Expect(err).To(BeNil())
Expect(val).To(Equal(int64(86400))) // 24h in seconds
```

This pattern was established in Story 2.4 (`databasesecretenginestaticrole_controller_test.go` line 292-300) for the rotation_period check.

[Source: controllers/databasesecretenginestaticrole_controller_test.go#L292-L300]

### `connection_details` Assertion Gotcha

When reading database config from Vault, `connection_details` is a nested map. The Vault Go client returns it as `map[string]interface{}`. Use checked type assertion:

```go
connDetails, ok := secret.Data["connection_details"].(map[string]interface{})
Expect(ok).To(BeTrue(), "expected connection_details to be a map")
```

Then access nested fields via the map. This nesting is unique to the database secret engine — auth engine configs don't have it.

### Risk Considerations

- **PostgreSQL connectivity from Vault:** Vault must reach PostgreSQL via cluster DNS (`my-postgresql-database.test-vault-config-operator.svc:5432`). Both services are in the Kind cluster so this should work. The existing static role test already proves this connectivity works.

- **`verifyConnection: true`:** The config fixture sets `verifyConnection: true`. If PostgreSQL is not yet ready when the config CR is created, the Vault write will fail. However, `deploy-postgresql` runs before tests in the `integration` Makefile target and waits for pod readiness, so PostgreSQL should be available.

- **`policy-admin` permissions:** The test uses `policy-admin` auth role in `vault-admin` namespace. This role must have permissions to create database engine mounts and write configs/roles. This is the standard integration test auth role with broad permissions. If permissions are insufficient, the reconciler will set `ReconcileFailed` condition. The existing SecretEngineMount tests (Story 3.3) already use `policy-admin` for secret engine operations.

- **Config `IsEquivalentToDesiredState` with credential resolution:** The `password` field is included in `toMap()` after `PrepareInternalValues` resolves it from the K8s Secret. But `IsEquivalentToDesiredState` deletes `password` from `desiredState` (since Vault never returns the password). This means the password comparison is handled correctly.

- **Role `IsEquivalentToDesiredState` strict equality:** The role's `IsEquivalentToDesiredState` does bare `DeepEqual` without filtering extra keys. Vault may return extra fields not in `toMap()`. This causes a write on every reconcile (known tech debt — Story 7-4). Does NOT affect `ReconcileSuccessful=True` or test correctness.

- **Checked type assertions:** Per Epic 3 retro action item and Epic 4 practice, always use two-value form `val, ok := x.(string)` with `Expect(ok).To(BeTrue())` for all Vault response field assertions.

### File Inventory — What Needs to Change

| # | File | Change Type | Description |
|---|------|-------------|-------------|
| 1 | `controllers/controllertestutils/decoder.go` | Modified | Add `GetDatabaseSecretEngineRoleInstance` method |
| 2 | `test/databasesecretengine/test-db-mount.yaml` | New | SecretEngineMount prerequisite (type=database) |
| 3 | `test/databasesecretengine/test-db-config.yaml` | New | DatabaseSecretEngineConfig with PostgreSQL connection |
| 4 | `test/databasesecretengine/test-db-role.yaml` | New | DatabaseSecretEngineRole with creation statements |
| 5 | `controllers/databasesecretengine_controller_test.go` | New | Integration test — create mount, config, role; verify Vault state; update role; delete and verify cleanup |

No changes to suite setup, controllers, webhooks, types, Makefile, or infrastructure manifests.

### No `make manifests generate` Needed

This story only adds an integration test file, YAML fixtures, and a decoder method. No CRD types, controllers, or webhooks are changed.

### Previous Story Intelligence

**From Story 4.3 (JWTOIDCAuthEngine integration tests — most recent):**
- Established the full Epic 4 integration test pattern with Ordered Describe, AfterAll cleanup, checked type assertions
- Demonstrated K8s Secret created programmatically in the test
- Non-deletable config persistence verification pattern (config was IsDeletable=false)
- Story 5.1's config is IsDeletable=true, so deletion test should verify Vault cleanup (not persistence)

**From Story 4.2 (LDAPAuthEngine integration tests):**
- Established the `IsDeletable=false` persistence verification rule (codified in project-context.md)
- Not needed here since both database types are IsDeletable=true

**From Story 4.1 (KubernetesAuthEngine integration tests):**
- AfterAll cleanup guard pattern
- Checked type assertions for Vault response fields

**From Story 2.4 (DatabaseSecretEngineStaticRole integration tests):**
- Demonstrates the full database engine prerequisite chain (Policy → K8s auth role → SecretEngineMount → credentials → config → static role)
- Shows PostgreSQL root credentials secret creation pattern
- Shows `json.Number` handling for Vault TTL responses
- Uses the OLD test pattern (no Ordered, no AfterAll, no checked type assertions on most fields) — the new test should use the modern Epic 4 patterns

**From Epic 4 Retrospective:**
- Story 5.1 classified as "Low — no new infra" (PostgreSQL already deployed)
- Continue using Opus-class models for integration test stories
- Story ordering: 5.1 (already there) → 5.2 (RabbitMQ) → 5.3 (remaining)
- Non-deletable config persistence rule (not needed here — both types are deletable)

[Source: _bmad-output/implementation-artifacts/epic-4-retro-2026-04-23.md]

### Git Intelligence (Recent Commits)

```
af8e4c4 Bmad epic 4 (#319)
9608211 Merge pull request #318 from raffaelespazzoli/bmad-epic-3
24a37f0 Complete Epic 3 retrospective and close Epics 1-3
cb473c3 Mark Story 3.4 as done after clean code review
866c843 Add integration tests for AuthEngineMount type (Story 3.4)
```

Codebase is clean post-Epic 4 merge to main.

### Integration Test Infrastructure Classification

Per the project's three-tier rule:
- **PostgreSQL:** Already deployed via `deploy-postgresql` Helm target → **No new infrastructure**
- **Vault API:** Already available in Kind
- **K8s Secrets:** Available via integration test client

**Classification: No new infrastructure — Lowest scope story in Epic 5**

[Source: _bmad-output/project-context.md#L134-L141 — Integration test infrastructure philosophy]
[Source: _bmad-output/implementation-artifacts/epic-4-retro-2026-04-23.md#L146 — Infrastructure classification]

### Project Structure Notes

- Decoder change in `controllers/controllertestutils/decoder.go` (add one method)
- Test file goes in `controllers/databasesecretengine_controller_test.go`
- Test fixtures go in `test/databasesecretengine/` directory (alongside existing fixtures, with `test-` prefix)
- No Makefile changes needed
- No new infrastructure directories
- All files follow existing naming conventions

### References

- [Source: api/v1alpha1/databasesecretengineconfig_types.go] — VaultObject implementation, GetPath ({path}/config/{name}), GetPayload, IsEquivalentToDesiredState (connection_details remapping + filtered payload), toMap, PrepareInternalValues (root credential resolution), IsDeletable=true
- [Source: api/v1alpha1/databasesecretengineconfig_types.go#L78-L83] — GetPath: {spec.path}/config/{name or metadata.name}
- [Source: api/v1alpha1/databasesecretengineconfig_types.go#L93-L119] — IsEquivalentToDesiredState: connection_details remapping, password deletion, filtered comparison
- [Source: api/v1alpha1/databasesecretengineconfig_types.go#L133-L215] — PrepareInternalValues + setInternalCredentials (3 credential sources)
- [Source: api/v1alpha1/databasesecretengineconfig_types.go#L364-L388] — DBSEConfig.toMap
- [Source: api/v1alpha1/databasesecretengineconfig_webhook.go] — Webhook: immutable path, isValid on create/update
- [Source: api/v1alpha1/databasesecretenginerole_types.go] — VaultObject implementation, GetPath ({path}/roles/{name}), toMap (7 keys), IsDeletable=true
- [Source: api/v1alpha1/databasesecretenginerole_types.go#L70-L75] — GetPath: {spec.path}/roles/{name or metadata.name}
- [Source: api/v1alpha1/databasesecretenginerole_types.go#L79-L82] — IsEquivalentToDesiredState: bare DeepEqual, no filtering
- [Source: api/v1alpha1/databasesecretenginerole_types.go#L183-L193] — DBSERole.toMap (7 keys)
- [Source: api/v1alpha1/databasesecretenginerole_webhook.go] — Webhook: immutable path on update, no create validation registered
- [Source: controllers/databasesecretengineconfig_controller.go#L85-L87] — Controller (VaultResource + root rotation logic + Secret/RandomSecret watches)
- [Source: controllers/databasesecretenginerole_controller.go#L70-L77] — Controller (VaultResource, simple)
- [Source: controllers/suite_integration_test.go#L136-L139] — Both controllers registered
- [Source: controllers/controllertestutils/decoder.go#L175-L188] — GetDatabaseSecretEngineConfigInstance exists; GetDatabaseSecretEngineRoleInstance MUST BE ADDED
- [Source: controllers/databasesecretenginestaticrole_controller_test.go] — Existing static role test (Story 2.4 pattern, shows database engine prerequisite chain)
- [Source: test/databasesecretengine/] — Existing fixtures (used by static role test); new test fixtures use test- prefix
- [Source: integration/postgresql-values.yaml] — PostgreSQL Helm values (fullnameOverride, credentials, init script)
- [Source: Makefile#L174-L181] — deploy-postgresql target (already in integration dependency chain)
- [Source: controllers/jwtoidcauthengine_controller_test.go] — Most recent Epic 4 test pattern reference
- [Source: controllers/secretenginemount_controller_test.go] — SecretEngineMount test pattern (sys/mounts verification)
- [Source: _bmad-output/implementation-artifacts/epic-4-retro-2026-04-23.md] — Epic 4 retrospective (readiness assessment, infrastructure classification)
- [Source: _bmad-output/project-context.md#L134-L141] — Integration test infrastructure philosophy
- [Source: _bmad-output/project-context.md#L148-L155] — Integration test pattern and Ordered lifecycle

## Dev Agent Record

### Agent Model Used

{{agent_model_name_version}}

### Debug Log References

### Completion Notes List

### File List
