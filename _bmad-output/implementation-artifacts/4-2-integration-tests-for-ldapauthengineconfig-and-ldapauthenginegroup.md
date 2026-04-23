# Story 4.2: Integration Tests for LDAPAuthEngineConfig and LDAPAuthEngineGroup

Status: done

## Story

As an operator developer,
I want integration tests for the LDAP auth engine configuration and group types covering create, reconcile success, Vault state verification, and delete with cleanup,
So that the LDAP auth method has end-to-end test coverage with a real OpenLDAP server in Kind.

## Acceptance Criteria

1. **Given** OpenLDAP is deployed in the Kind cluster **When** `make integration` is run **Then** the LDAP server is available at `ldap://ldap.ldap.svc.cluster.local:389` and pre-seeded with users and groups

2. **Given** a Kubernetes Secret with LDAP bind credentials exists in `vault-admin` namespace **And** an AuthEngineMount (type=ldap) has been created and reconciled **And** a LDAPAuthEngineConfig CR is created targeting the test LDAP mount with the bind credentials secret **When** the reconciler processes it **Then** the LDAP config is written to Vault at `auth/{path}/config` and ReconcileSuccessful=True

3. **Given** a LDAPAuthEngineGroup CR is created mapping an LDAP group name to Vault policies **When** the reconciler processes it **Then** the group mapping exists in Vault at `auth/{path}/groups/{name}` with correct policies and ReconcileSuccessful=True

4. **Given** the LDAPAuthEngineGroup CR is deleted (IsDeletable=true) **When** the reconciler processes the deletion **Then** the group mapping is removed from Vault and the CR is deleted from K8s

5. **Given** the LDAPAuthEngineConfig CR is deleted (IsDeletable=false) **When** the reconciler processes the deletion **Then** the CR is deleted from K8s without Vault cleanup (no finalizer)

## Tasks / Subtasks

- [x] Task 1: Add LDAP deployment to integration test infrastructure (AC: 1)
  - [x] 1.1: Add `deploy-ldap` target to Makefile that deploys OpenLDAP in `ldap` namespace, creates the `bindcredentials` K8s Secret, and adds the `admins-group` LDAP entry
  - [x] 1.2: Add `deploy-ldap` as a dependency of the `integration` target in Makefile
  - [x] 1.3: Create `integration/ldap/bindcredentials-secret.yaml` with admin bind credentials

- [x] Task 2: Add decoder methods (AC: 2, 3)
  - [x] 2.1: Add `GetLDAPAuthEngineConfigInstance` method to `controllers/controllertestutils/decoder.go`
  - [x] 2.2: Add `GetLDAPAuthEngineGroupInstance` method to `controllers/controllertestutils/decoder.go`

- [x] Task 3: Create test fixtures (AC: 2, 3, 4, 5)
  - [x] 3.1: Create `test/ldapauthengine/test-ldap-auth-mount.yaml` — AuthEngineMount with `type: ldap`, `path: test-ldap-auth`, `metadata.name: test-laec-mount`
  - [x] 3.2: Create `test/ldapauthengine/test-ldap-auth-config.yaml` — LDAPAuthEngineConfig with bind credentials from `test-ldap-bind-creds` K8s Secret
  - [x] 3.3: Create `test/ldapauthengine/test-ldap-auth-group.yaml` — LDAPAuthEngineGroup mapping `admins-group` to `vault-admin` policy

- [x] Task 4: Create integration test file (AC: 2, 3, 4, 5)
  - [x] 4.1: Create `controllers/ldapauthengine_controller_test.go` with `//go:build integration` tag
  - [x] 4.2: Add prerequisite context — create bind credentials K8s Secret, create AuthEngineMount, wait for reconcile
  - [x] 4.3: Add context for LDAPAuthEngineConfig — create, poll for ReconcileSuccessful=True, verify Vault state at `auth/test-ldap-auth/test-laec-mount/config`
  - [x] 4.4: Add LDAP group entry to OpenLDAP (via Vault direct write or pre-seeded data) so the group mapping test makes sense
  - [x] 4.5: Add context for LDAPAuthEngineGroup — create, poll for ReconcileSuccessful=True, verify Vault state at `auth/test-ldap-auth/test-laec-mount/groups/test-ldap-admins`
  - [x] 4.6: Add deletion context — delete group (IsDeletable=true, verify Vault cleanup), delete config (IsDeletable=false), delete mount

- [x] Task 5: End-to-end verification (AC: 1, 2, 3, 4, 5)
  - [x] 5.1: Run `make integration` and verify new tests pass alongside all existing tests
  - [x] 5.2: Verify no regressions — existing `kubernetes` auth and all prior tests unaffected

### Review Findings

- [x] [Review][Patch] `ldap-setup` now attempts to create `admins-group` twice after the new configmap seed was added, so the manual LDAP workflow can fail with `Already exists` on `ldapadd`. [`Makefile:189`] — Fixed: removed redundant `ldapadd` line; group is pre-seeded in configmap LDIF.
- [x] [Review][Patch] The test fixture maps Vault policies to `test-ldap-admins` while the LDAP seed data adds `admins-group`, so the new seeded LDAP group is never actually exercised by the integration test. [`test/ldapauthengine/test-ldap-auth-group.yaml:1`] — Fixed: changed `spec.name` to `admins-group` and updated Vault read paths in test.
- [x] [Review][Patch] The delete phase verifies only Kubernetes deletion for `LDAPAuthEngineConfig`; it never proves AC5's "without Vault cleanup" behavior by checking that the config still exists in Vault before the mount is deleted. [`controllers/ldapauthengine_controller_test.go:189`] — Fixed: added Vault read after config CR deletion asserting config persists with correct `url`.

## Dev Notes

### Infrastructure Scope — OpenLDAP Deployment in Kind (Tier 1: Install in Kind)

Per the project's three-tier integration test infrastructure rule, LDAP CAN be installed in Kind. An OpenLDAP deployment already exists at `integration/ldap/` with all manifests, but it is NOT wired into `make integration`. This story must:

1. **Add a `deploy-ldap` Makefile target** that:
   - Creates the `ldap` namespace
   - Applies `./integration/ldap` manifests (deployment, service, configmap with LDIF seed data)
   - Waits for the LDAP pod to be ready
   - Creates a K8s Secret named `test-ldap-bind-creds` in `vault-admin` namespace with `username: cn=admin,dc=example,dc=com` and `password: admin`
   - Adds the `admins-group` LDAP entry via `ldapadd` (using `integration/ldap/group.ldif`) — OR seeds it in the configmap LDIF

2. **Add `deploy-ldap` to the `integration` target's dependency list** (after `deploy-vault`, before `vault`/test run)

The existing `ldap-setup` target has port-forwarding steps that are NOT needed for integration tests (the test uses in-cluster DNS). Reuse the manifests, not the target.

**OpenLDAP details:**
- Image: `osixia/openldap:1.3.0`
- Service: `ldap.ldap.svc.cluster.local:389`
- Admin bind DN: `cn=admin,dc=example,dc=com`, password: `admin`
- Domain: `dc=example,dc=com`
- Pre-seeded users: trevor, john, dev1-12, eric, erin, vanessa, mary, julia, matt (all with passwords)
- Pre-seeded groups: Online Retail Banking, Mobile Banking, Retail Banking, Core Banking, etc.
- TLS: disabled (`LDAP_TLS=false`)

The `integration/ldap/group.ldif` adds an `admins-group` with member `uid=trevor`. This group is used by the LDAPAuthEngineGroup test.

**Alternative: Seed the `admins-group` via the configmap LDIF** instead of running `ldapadd` at deploy time. This is simpler and avoids needing to exec into the pod or port-forward. Append the `admins-group` entry to `integration/ldap/configmap.yaml`'s `database.ldif` data, or create a separate LDIF file in the configmap.

[Source: Makefile#L183-L200 — existing ldap-setup target]
[Source: integration/ldap/ — OpenLDAP manifests]
[Source: _bmad-output/project-context.md#L134-L141 — Integration test infrastructure philosophy]

### Both Types Use VaultResource Reconciler — NOT VaultEngineResource

Both LDAPAuthEngineConfig and LDAPAuthEngineGroup use `NewVaultResource` (the standard reconciler variant), same as Policy (Story 3.1) and KubernetesAuthEngine types (Story 4.1). The reconcile flow is:

1. `prepareContext()` enriches context with kubeClient, restConfig, vaultConnection, vaultClient
2. `NewVaultResource(&r.ReconcilerBase, instance)` creates the standard reconciler
3. `VaultResource.Reconcile()` → `manageReconcileLogic()`:
   - `PrepareInternalValues()` — resolves bind credentials (config) or no-op (group)
   - `PrepareTLSConfig()` — resolves TLS certs (config) or no-op (group)
   - `VaultEndpoint.CreateOrUpdate()` — reads from Vault, calls `IsEquivalentToDesiredState()`, writes if different
4. `ManageOutcome()` sets `ReconcileSuccessful` condition

[Source: controllers/ldapauthengineconfig_controller.go — uses NewVaultResource]
[Source: controllers/ldapauthenginegroup_controller.go — uses NewVaultResource]

### LDAPAuthEngineConfig — Key Implementation Details

**GetPath():**
```go
func (d *LDAPAuthEngineConfig) GetPath() string {
    return vaultutils.CleansePath("auth/" + string(d.Spec.Path) + "/config")
}
```
For fixture with `path: test-ldap-auth/test-laec-mount` → `auth/test-ldap-auth/test-laec-mount/config`

Note: Unlike KubernetesAuthEngineConfig, LDAP config does NOT include `metadata.name` in the path. The path ends with `/config`.

[Source: api/v1alpha1/ldapauthengineconfig_types.go#L72-L74]

**IsDeletable(): false** — No finalizer, no Vault cleanup on CR deletion. The config persists in Vault until the auth mount itself is deleted.

[Source: api/v1alpha1/ldapauthengineconfig_types.go#L68-L70]

**IsEquivalentToDesiredState() — CUSTOM (strips `bindpass`):**
```go
func (d *LDAPAuthEngineConfig) IsEquivalentToDesiredState(payload map[string]interface{}) bool {
    desiredState := d.Spec.LDAPConfig.toMap()
    delete(desiredState, "bindpass")
    return reflect.DeepEqual(desiredState, payload)
}
```
This is a **custom** implementation — it removes the `bindpass` key before comparison because Vault's read response for `auth/{path}/config` does NOT return the bind password (sensitive field). This means the comparison is on ~29 keys instead of 30.

However, Vault's read response may still include extra fields not in the desired state, causing `IsEquivalentToDesiredState` to return false on every reconcile. This is the same Story 7-4 tech debt as other types and does NOT affect test correctness.

[Source: api/v1alpha1/ldapauthengineconfig_types.go#L77-L81]

**toMap() — ~30 Vault keys:**
Key Vault API mappings: `url`, `case_sensitive_names`, `request_timeout`, `starttls`, `tls_min_version`, `tls_max_version`, `insecure_tls`, `certificate`, `client_tls_cert`, `client_tls_key`, `binddn`, `bindpass`, `userdn`, `userattr`, `discoverdn`, `deny_null_bind`, `upndomain`, `userfilter`, `anonymous_group_search`, `groupfilter`, `groupdn`, `groupattr`, `username_as_alias`, `token_ttl`, `token_max_ttl`, `token_policies`, `token_bound_cidrs`, `token_explicit_max_ttl`, `token_no_default_policy`, `token_num_uses`, `token_period`, `token_type`

Note: `toMap()` also has inline logic that copies TLS cert fields from spec into retrieved fields when they're set directly (not via TLS secret). This dual path is important for understanding the flow.

[Source: api/v1alpha1/ldapauthengineconfig_types.go — LDAPConfig.toMap()]

**PrepareInternalValues() — Bind credential resolution:**
Resolves bind credentials from one of three sources (same RootCredentialConfig pattern as DatabaseSecretEngineConfig):
1. `bindCredentials.secret` → reads K8s Secret, gets username from `usernameKey` or `spec.bindDN`, password from `passwordKey`
2. `bindCredentials.randomSecret` → reads RandomSecret from K8s, reads its Vault secret, gets password
3. `bindCredentials.vaultSecret` → reads Vault secret directly, gets username and password

Sets `retrievedUsername` and `retrievedPassword` on the LDAPConfig struct.

**For the integration test:** Use option 1 (K8s Secret). Create a `test-ldap-bind-creds` Secret with `username: cn=admin,dc=example,dc=com` and `password: admin`.

[Source: api/v1alpha1/ldapauthengineconfig_types.go — setInternalCredentials]

**PrepareTLSConfig() — TLS certificate resolution:**
If `tLSConfig.tLSSecret` is set, reads the K8s Secret and extracts `ca.crt`, `tls.crt`, `tls.key`. If TLS is set directly in spec (`certificate`, `clientTLSCert`, `clientTLSKey`), copies those.

**For the integration test:** OpenLDAP is configured with `LDAP_TLS=false`. We do NOT need TLS config. Set `insecureTLS: true` and leave `tLSConfig` empty. The `PrepareTLSConfig` check `reflect.DeepEqual(d.Spec.TLSConfig, vaultutils.TLSConfig{TLSSecret: &corev1.LocalObjectReference{}})` will NOT match an empty TLSConfig — but `setTLSConfig` with no TLSSecret and no inline certs just returns nil. So we need to verify this works.

Actually, looking more carefully: the `PrepareTLSConfig` checks if the spec's TLSConfig equals `TLSConfig{TLSSecret: &corev1.LocalObjectReference{}}` — if so, it returns nil (skips). Otherwise it calls `setTLSConfig`. If TLSConfig has no TLSSecret and no inline certs, `setTLSConfig` falls through to the end and returns nil. So an empty TLSConfig is fine.

[Source: api/v1alpha1/ldapauthengineconfig_types.go — PrepareTLSConfig, setTLSConfig]

**Webhook:** `Default()` is a no-op (just logs). `ValidateUpdate()` rejects changes to `spec.path`. `ValidateCreate()` returns nil. Validation via `isValid()` is called only when `IsValid()` is invoked — it checks that exactly one credential source is specified.

[Source: api/v1alpha1/ldapauthengineconfig_webhook.go]

### LDAPAuthEngineGroup — Key Implementation Details

**GetPath():**
```go
func (d *LDAPAuthEngineGroup) GetPath() string {
    return vaultutils.CleansePath("auth/" + string(d.Spec.Path) + "/groups/" + string(d.Spec.Name))
}
```
For fixture with `path: test-ldap-auth/test-laec-mount`, `name: test-ldap-admins` → `auth/test-ldap-auth/test-laec-mount/groups/test-ldap-admins`

Note: Uses `spec.Name`, NOT `metadata.name`, for the Vault group path.

[Source: api/v1alpha1/ldapauthenginegroup_types.go#L67-L69]

**IsDeletable(): true** — Finalizer added after first successful reconcile. On deletion, the group is removed from Vault at `auth/{path}/groups/{name}`.

[Source: api/v1alpha1/ldapauthenginegroup_types.go#L71-L73]

**toMap() — 2 Vault keys:**
```go
func (i *LDAPAuthEngineGroup) toMap() map[string]interface{} {
    payload := map[string]interface{}{}
    payload["name"] = i.Spec.Name
    payload["policies"] = i.Spec.Policies
    return payload
}
```

[Source: api/v1alpha1/ldapauthenginegroup_types.go#L139-L145]

**IsEquivalentToDesiredState():** Bare `reflect.DeepEqual(d.GetPayload(), payload)` — no key filtering. Vault's read response for `auth/{path}/groups/{name}` may include extra keys.

[Source: api/v1alpha1/ldapauthenginegroup_types.go#L79-L81]

**PrepareInternalValues():** No-op — returns nil.

**Webhook:** `Default()` is a no-op. `ValidateCreate()` and `ValidateUpdate()` both return nil — no validation. No immutable path check on update (unlike most types — `LDAPAuthEngineGroup` webhook does NOT reject `spec.path` changes).

[Source: api/v1alpha1/ldapauthenginegroup_webhook.go]

### Vault API Response Shapes

**GET `auth/{path}/config`** — Returns LDAP auth config:
```json
{
  "data": {
    "url": "ldap://ldap.ldap.svc.cluster.local",
    "case_sensitive_names": false,
    "request_timeout": "90s",
    "starttls": false,
    "tls_min_version": "tls12",
    "tls_max_version": "tls12",
    "insecure_tls": true,
    "certificate": "",
    "client_tls_cert": "",
    "client_tls_key": "",
    "binddn": "cn=admin,dc=example,dc=com",
    "userdn": "ou=users,dc=example,dc=com",
    "userattr": "cn",
    "discoverdn": false,
    "deny_null_bind": true,
    "upndomain": "",
    "userfilter": "({{.UserAttr}}={{.Username}})",
    "anonymous_group_search": false,
    "groupfilter": "...",
    "groupdn": "ou=Groups,dc=example,dc=com",
    "groupattr": "",
    "username_as_alias": false,
    "token_ttl": 0,
    "token_max_ttl": 0,
    ...
  }
}
```
Note: `bindpass` is NOT returned (sensitive). `token_num_uses`, `token_period` etc. may be returned as integers. Extra fields like `use_token_groups`, `connection_timeout` may appear in newer Vault versions.

**GET `auth/{path}/groups/{name}`** — Returns group mapping:
```json
{
  "data": {
    "name": "test-ldap-admins",
    "policies": "vault-admin"
  }
}
```
The group response is very simple — just `name` and `policies`.

### Test Design — Dependency Chain

```
K8s Secret (test-ldap-bind-creds) — must exist before config
  └── AuthEngineMount (test-ldap-auth/test-laec-mount, type=ldap)
        └── LDAPAuthEngineConfig → auth/test-ldap-auth/test-laec-mount/config
        └── LDAPAuthEngineGroup → auth/test-ldap-auth/test-laec-mount/groups/test-ldap-admins
```

Resources must be created in order: Secret → Mount → Config → Group. Deletion in reverse: Group → Config → Mount → Secret.

The AuthEngineMount must be reconciled before the config or group, because Vault rejects writes to `auth/{path}/config` if the mount doesn't exist.

The LDAPAuthEngineGroup does NOT depend on config being written — Vault allows creating group mappings before the config is set. But for test realism and orderly state, create config before group.

### Bind Credentials K8s Secret — Created in Test, Not as Fixture

The bind credentials Secret should be created programmatically in the test's `BeforeAll` or first `Context` block using the `k8sIntegrationClient`, NOT as a YAML fixture. This is because:
1. The test needs to control the namespace (`vault-admin`)
2. The data fields are simple (`username`, `password`)
3. Following the pattern used in the `make integration` PostgreSQL setup where secrets are created directly

```go
bindSecret := &corev1.Secret{
    ObjectMeta: metav1.ObjectMeta{
        Name:      "test-ldap-bind-creds",
        Namespace: vaultAdminNamespaceName,
    },
    Data: map[string][]byte{
        "username": []byte("cn=admin,dc=example,dc=com"),
        "password": []byte("admin"),
    },
}
```

### Test Fixture Design

**Fixture 1: `test/ldapauthengine/test-ldap-auth-mount.yaml`** — AuthEngineMount prerequisite:
```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: AuthEngineMount
metadata:
  name: test-laec-mount
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  type: ldap
  path: test-ldap-auth
```
Mounts at `sys/auth/test-ldap-auth/test-laec-mount`. Uses `type: ldap` to enable the LDAP auth method.

**Fixture 2: `test/ldapauthengine/test-ldap-auth-config.yaml`** — LDAPAuthEngineConfig:
```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: LDAPAuthEngineConfig
metadata:
  name: test-laec-config
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  path: test-ldap-auth/test-laec-mount
  url: "ldap://ldap.ldap.svc.cluster.local"
  bindDN: "cn=admin,dc=example,dc=com"
  userDN: "ou=Users,dc=example,dc=com"
  userAttr: "cn"
  groupDN: "ou=Groups,dc=example,dc=com"
  groupFilter: "(|(memberUid={{.Username}})(member={{.UserDN}})(uniqueMember={{.UserDN}}))"
  userFilter: "({{.UserAttr}}={{.Username}})"
  insecureTLS: true
  bindCredentials:
    secret:
      name: test-ldap-bind-creds
    usernameKey: username
    passwordKey: password
```
`GetPath()` returns `auth/test-ldap-auth/test-laec-mount/config`. The config points at the OpenLDAP service in the cluster.

**Fixture 3: `test/ldapauthengine/test-ldap-auth-group.yaml`** — LDAPAuthEngineGroup:
```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: LDAPAuthEngineGroup
metadata:
  name: test-laeg-group
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  name: "test-ldap-admins"
  path: "test-ldap-auth/test-laec-mount"
  policies: "vault-admin"
```
`GetPath()` returns `auth/test-ldap-auth/test-laec-mount/groups/test-ldap-admins`.

### Verifying Vault State

**Config verification:**
```go
secret, err := vaultClient.Logical().Read("auth/test-ldap-auth/test-laec-mount/config")
Expect(err).To(BeNil())
Expect(secret).NotTo(BeNil())
Expect(secret.Data["url"]).To(Equal("ldap://ldap.ldap.svc.cluster.local"))
Expect(secret.Data["binddn"]).To(Equal("cn=admin,dc=example,dc=com"))
Expect(secret.Data["userdn"]).To(Equal("ou=Users,dc=example,dc=com"))
Expect(secret.Data["insecure_tls"]).To(BeTrue())
```
Note: `bindpass` should NOT be in the response. Do NOT try to verify it.

**Group verification:**
```go
secret, err := vaultClient.Logical().Read("auth/test-ldap-auth/test-laec-mount/groups/test-ldap-admins")
Expect(err).To(BeNil())
Expect(secret).NotTo(BeNil())
Expect(secret.Data["policies"]).To(Equal("vault-admin"))
```

**Delete verification (group — IsDeletable=true):**
```go
Eventually(func() bool {
    err := k8sIntegrationClient.Get(ctx, types.NamespacedName{...}, &redhatcopv1alpha1.LDAPAuthEngineGroup{})
    return apierrors.IsNotFound(err)
}, timeout, interval).Should(BeTrue())

Eventually(func() bool {
    secret, err := vaultClient.Logical().Read("auth/test-ldap-auth/test-laec-mount/groups/test-ldap-admins")
    return err == nil && secret == nil
}, timeout, interval).Should(BeTrue())
```

**Delete verification (config — IsDeletable=false):**
Config deletion from K8s happens immediately (no finalizer). No Vault cleanup expected.
```go
Eventually(func() bool {
    err := k8sIntegrationClient.Get(ctx, types.NamespacedName{...}, &redhatcopv1alpha1.LDAPAuthEngineConfig{})
    return apierrors.IsNotFound(err)
}, timeout, interval).Should(BeTrue())
```

### Test Structure

```
Describe("LDAPAuthEngine controllers", Ordered)
  var bindSecret *corev1.Secret
  var mountInstance *redhatcopv1alpha1.AuthEngineMount
  var configInstance *redhatcopv1alpha1.LDAPAuthEngineConfig
  var groupInstance *redhatcopv1alpha1.LDAPAuthEngineGroup

  AfterAll: best-effort delete all instances + bind secret

  Context("When creating prerequisite resources")
    It("Should create the bind credentials secret and LDAP auth mount")
      - Create test-ldap-bind-creds K8s Secret in vault-admin namespace
      - Load test-ldap-auth-mount.yaml via decoder.GetAuthEngineMountInstance
      - Set namespace to vaultAdminNamespaceName, create
      - Eventually poll for ReconcileSuccessful=True
      - Verify mount exists in sys/auth with key "test-ldap-auth/test-laec-mount/"

  Context("When creating a LDAPAuthEngineConfig")
    It("Should write the LDAP config to Vault")
      - Load test-ldap-auth-config.yaml via decoder.GetLDAPAuthEngineConfigInstance
      - Set namespace to vaultAdminNamespaceName, create
      - Eventually poll for ReconcileSuccessful=True
      - Read auth/test-ldap-auth/test-laec-mount/config from Vault
      - Verify url = "ldap://ldap.ldap.svc.cluster.local"
      - Verify binddn = "cn=admin,dc=example,dc=com"
      - Verify insecure_tls = true

  Context("When creating a LDAPAuthEngineGroup")
    It("Should create the group mapping in Vault")
      - Load test-ldap-auth-group.yaml via decoder.GetLDAPAuthEngineGroupInstance
      - Set namespace to vaultAdminNamespaceName, create
      - Eventually poll for ReconcileSuccessful=True
      - Read auth/test-ldap-auth/test-laec-mount/groups/test-ldap-admins
      - Verify policies = "vault-admin"

  Context("When deleting LDAPAuthEngine resources")
    It("Should clean up group from Vault and remove all resources")
      - Delete group CR (IsDeletable=true → Vault cleanup)
      - Eventually verify K8s deletion (NotFound)
      - Eventually verify group removed from Vault (Read returns nil)
      - Delete config CR (IsDeletable=false → no Vault cleanup)
      - Eventually verify K8s deletion
      - Delete AuthEngineMount
      - Eventually verify K8s deletion and mount gone from sys/auth
      - Delete bind credentials secret
```

### Import Requirements

```go
import (
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

### Name Collision Prevention

Fixture names use `test-laec-` and `test-laeg-` prefixes:
- `test-ldap-auth/test-laec-mount` — auth mount (unique path prefix)
- `test-laec-config` — LDAP config CR name
- `test-laeg-group` — LDAP group CR name
- `test-ldap-admins` — Vault group name in `spec.name`
- `test-ldap-bind-creds` — bind credentials K8s Secret

These don't collide with:
- `ldap/test` — existing manual LDAP fixtures
- `test-k8s-auth/test-kaec-mount` — Story 4.1 Kubernetes auth tests
- `test-auth-mount/test-aem-*` — Story 3.4 AuthEngineMount tests
- `kubernetes/` — default Kubernetes auth mount

### Controller Registration — Already Done

Both controllers are registered in `suite_integration_test.go`:
```go
err = (&LDAPAuthEngineConfigReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "LDAPAuthEngineConfig")}).SetupWithManager(mgr)
Expect(err).ToNot(HaveOccurred())

err = (&LDAPAuthEngineGroupReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "LDAPAuthEngineGroup")}).SetupWithManager(mgr)
Expect(err).ToNot(HaveOccurred())
```

No changes needed to the test suite setup.

[Source: controllers/suite_integration_test.go#L154-L158]

### LDAPAuthEngineConfig Controller — Secret and RandomSecret Watches

The LDAPAuthEngineConfig controller has extra `Watches` for:
1. K8s Secrets (basic-auth type) — re-queues config CRs when their referenced bind credential secret changes
2. K8s Secrets (TLS type) — re-queues config CRs when their referenced TLS secret changes
3. RandomSecret — re-queues config CRs when their referenced RandomSecret's `LastVaultSecretUpdate` changes

This is relevant because the controller's `SetupWithManager` does more than simple `For()`. The test doesn't directly exercise these watch paths but should be aware of them.

[Source: controllers/ldapauthengineconfig_controller.go#L98-L213]

### Makefile Integration — deploy-ldap Target Design

The new `deploy-ldap` target should:

```makefile
.PHONY: deploy-ldap
deploy-ldap: kubectl
	$(KUBECTL) create namespace ldap --dry-run=client -o yaml | $(KUBECTL) apply -f -
	$(KUBECTL) apply -f ./integration/ldap -n ldap
	$(KUBECTL) wait --for=condition=ready -n ldap pod -l app=ldap --timeout=$(KUBECTL_WAIT_TIMEOUT)
```

Key design decisions:
- Use `--dry-run=client -o yaml | kubectl apply` for idempotent namespace creation (same pattern as integration pipeline)
- Use label selector `-l app=ldap` instead of jq-based pod name extraction (simpler, more robust)
- Do NOT include port-forwarding (not needed for integration tests — test uses in-cluster DNS via Vault)
- The `admins-group` LDAP entry from `group.ldif` should be added to the configmap's `database.ldif` so it's pre-seeded, avoiding the need for `ldapadd` at deploy time

Update the `integration` target:
```makefile
integration: kind-setup deploy-vault deploy-ingress deploy-postgresql deploy-ldap vault manifests generate fmt vet envtest
```

### Bind Credentials Secret in vault-admin Namespace

The test must create a K8s Secret in `vault-admin` namespace BEFORE creating the LDAPAuthEngineConfig CR, because `PrepareInternalValues` → `setInternalCredentials` reads it via `kubeClient.Get`. The Secret must have:
- `data.username`: `cn=admin,dc=example,dc=com` (used as bind DN if `spec.bindDN` is not set, or overridden by `spec.bindDN` if set)
- `data.password`: `admin`

Since the fixture sets `bindDN: "cn=admin,dc=example,dc=com"` explicitly, the `setInternalCredentials` logic takes the `spec.bindDN` as username and only reads password from the secret. But providing both keys in the secret is good practice.

### Risk Considerations

- **Config `IsEquivalentToDesiredState` may return false on every reconcile:** The custom implementation strips `bindpass` but does NOT filter extra Vault response fields. Vault may return fields like `use_token_groups`, `connection_timeout`, etc. that aren't in `toMap()`. This causes a write on every cycle — known tech debt (Story 7-4), does NOT block `ReconcileSuccessful=True`.

- **OpenLDAP pod startup timing:** The `deploy-ldap` target waits for pod readiness, but the LDAP server inside the container may take a few seconds after readiness to fully initialize the database from LDIF. If the Vault write to `auth/{path}/config` happens immediately and Vault tests the LDAP connection, it should succeed since we use `insecureTLS: true` and the LDAP server is up.

- **Vault connection to OpenLDAP:** Vault needs to reach `ldap://ldap.ldap.svc.cluster.local:389` from inside the Kind cluster. Since both Vault and LDAP are in the cluster, this works via cluster DNS. No port-forwarding needed.

- **The `admins-group` LDAP group:** The group must exist in LDAP for the Vault LDAP group mapping to make semantic sense. However, Vault does NOT validate that the group name in the group mapping actually exists in LDAP — it's just a mapping table. So the test will work even if the LDAP group doesn't exist, but for realism, we should ensure it's seeded.

- **Checked type assertions:** Per Epic 3 retro action item, always use two-value form for type assertions in tests. For LDAP, the group response has string values, so `secret.Data["policies"].(string)` should work, but use checked form for safety.

### File Inventory — What Needs to Change

| # | File | Change Type | Description |
|---|------|-------------|-------------|
| 1 | `Makefile` | Modified | Add `deploy-ldap` target; add `deploy-ldap` to `integration` dependencies |
| 2 | `integration/ldap/configmap.yaml` | Modified | Add `admins-group` LDAP entry to seed data (avoid needing `ldapadd` at deploy time) |
| 3 | `controllers/controllertestutils/decoder.go` | Modified | Add `GetLDAPAuthEngineConfigInstance` and `GetLDAPAuthEngineGroupInstance` |
| 4 | `test/ldapauthengine/test-ldap-auth-mount.yaml` | New | AuthEngineMount prerequisite (type=ldap) |
| 5 | `test/ldapauthengine/test-ldap-auth-config.yaml` | New | LDAPAuthEngineConfig with bind credentials from K8s Secret |
| 6 | `test/ldapauthengine/test-ldap-auth-group.yaml` | New | LDAPAuthEngineGroup mapping test-ldap-admins to vault-admin |
| 7 | `controllers/ldapauthengine_controller_test.go` | New | Integration test — create mount, config, group; verify Vault state; delete and verify cleanup |

No changes to suite setup, controllers, webhooks, or types.

### No `make manifests generate` Needed

This story only adds an integration test file, YAML fixtures, decoder methods, and Makefile changes. No CRD types, controllers, or webhooks are changed.

### Previous Story Intelligence

**From Story 4.1 (KubernetesAuthEngine integration tests):**
- Established the Epic 4 auth engine integration test pattern: prerequisite AuthEngineMount → config → role(s), with isolated test mount paths
- Demonstrated decoder method addition pattern (add `GetKubernetesAuthEngineConfigInstance`)
- Confirmed `vaultAdminNamespaceName` for all CR creation
- Used `Eventually` polling for ReconcileSuccessful condition check
- `AfterAll` cleanup guard pattern
- Checked type assertions for Vault response fields

**From Story 3.4 (AuthEngineMount integration tests):**
- Established AuthEngineMount test pattern: verify `sys/auth` response
- Demonstrated `json.Number` for TTL values in tune config
- `AfterAll` cleanup guard

**From Story 3.1 (Policy integration tests):**
- Established VaultResource test pattern: create → poll → verify → delete → verify cleanup
- Both LDAP types use the same VaultResource reconciler

**From Epic 3 Retrospective:**
- "Story ordering: 4.1 (simplest) → 4.2 (LDAP infra) → 4.3 (Keycloak infra)"
- "Checked type assertions rule" — always use two-value form
- Story 4.2 has MEDIUM infrastructure scope — requires deploying OpenLDAP in Kind

### Git Intelligence (Recent Commits)

```
9608211 Merge pull request #318 from raffaelespazzoli/bmad-epic-3
24a37f0 Complete Epic 3 retrospective and close Epics 1-3
cb473c3 Mark Story 3.4 as done after clean code review
866c843 Add integration tests for AuthEngineMount type (Story 3.4)
db21d90 Add integration tests for SecretEngineMount type (Story 3.3)
```

Codebase is clean post-Epic 3 merge to main. No pending changes affect this story.

### Integration Test Infrastructure Classification

Per the project's three-tier rule:
- **LDAP server:** CAN be installed in Kind → **Tier 1: Install in Kind** using existing `integration/ldap/` manifests
- **Vault API:** Already available in Kind
- **K8s Secrets:** Available via envtest client

**Classification: Install in Kind — Medium infrastructure scope** (new Makefile target + LDAP deployment)

[Source: _bmad-output/project-context.md#L134-L141 — Integration test infrastructure philosophy]

### Project Structure Notes

- Decoder changes in `controllers/controllertestutils/decoder.go` (add two methods)
- Test file goes in `controllers/ldapauthengine_controller_test.go`
- Test fixtures go in `test/ldapauthengine/` directory (alongside existing manual fixtures, with `test-` prefix)
- Makefile changes: new `deploy-ldap` target, updated `integration` dependency list
- All files follow existing naming conventions

### References

- [Source: api/v1alpha1/ldapauthengineconfig_types.go] — LDAPAuthEngineConfig VaultObject implementation, GetPath, GetPayload, IsEquivalentToDesiredState (custom — strips bindpass), toMap (~30 keys), PrepareInternalValues (bind credentials), PrepareTLSConfig, IsDeletable=false
- [Source: api/v1alpha1/ldapauthengineconfig_types.go#L72-L74] — GetPath: auth/{path}/config
- [Source: api/v1alpha1/ldapauthengineconfig_types.go#L77-L81] — IsEquivalentToDesiredState: delete(bindpass) + reflect.DeepEqual
- [Source: api/v1alpha1/ldapauthengineconfig_webhook.go] — Webhook: immutable path, no create validation
- [Source: api/v1alpha1/ldapauthenginegroup_types.go] — LDAPAuthEngineGroup VaultObject implementation, GetPath (uses spec.Name), toMap (2 keys: name, policies), IsDeletable=true
- [Source: api/v1alpha1/ldapauthenginegroup_types.go#L67-L69] — GetPath: auth/{path}/groups/{spec.Name}
- [Source: api/v1alpha1/ldapauthenginegroup_webhook.go] — Webhook: no validation on create or update (no immutable path check!)
- [Source: controllers/ldapauthengineconfig_controller.go] — Controller (VaultResource + Secret/RandomSecret watches)
- [Source: controllers/ldapauthenginegroup_controller.go] — Controller (VaultResource, simple)
- [Source: controllers/suite_integration_test.go#L154-L158] — Both controllers registered
- [Source: controllers/controllertestutils/decoder.go] — Existing decoder methods; GetLDAPAuthEngineConfigInstance and GetLDAPAuthEngineGroupInstance MUST BE ADDED
- [Source: integration/ldap/] — OpenLDAP deployment manifests (deployment, service, configmap with LDIF)
- [Source: integration/ldap/group.ldif] — admins-group LDAP entry
- [Source: test/ldapauthengine/] — Existing manual LDAP fixtures (reference only)
- [Source: Makefile#L135] — integration target dependencies (must add deploy-ldap)
- [Source: Makefile#L183-L200] — existing ldap-setup target (reference for OpenLDAP deploy pattern)
- [Source: controllers/authenginemount_controller_test.go] — AuthEngineMount test pattern (Story 3.4)
- [Source: controllers/policy_controller_test.go] — VaultResource test pattern (Story 3.1)
- [Source: _bmad-output/implementation-artifacts/4-1-integration-tests-for-kubernetesauthengineconfig-and-kubernetesauthenginerole.md] — Previous story (closest pattern in Epic 4)
- [Source: _bmad-output/project-context.md#L134-L141] — Integration test infrastructure philosophy
- [Source: _bmad-output/project-context.md#L143-L155] — Integration test pattern and Ordered lifecycle

## Dev Agent Record

### Agent Model Used

Opus 4.6

### Debug Log References

- Run 1: `tls_min_version` error — Vault 1.19.0 rejects empty string for `tls_min_version`/`tls_max_version`. Root cause: kubebuilder defaults not applied because Go struct serializes zero-value strings (no `omitempty` on JSON tag), so API server sees field as present and skips defaulting. Fix: added `omitempty` to `TLSMinVersion` and `TLSMaxVersion` JSON tags in `ldapauthengineconfig_types.go`.
- Run 2: `policies` type assertion error — Vault LDAP group API returns `policies` as `[]interface{}` (JSON array), not a plain string. Fix: changed type assertion from `.(string)` to `.([]interface{})` with `ContainElement` matcher.
- Run 3: All 44 specs passed (40.7% coverage), 418.9s, exit code 0, no regressions.

### Completion Notes List

- Task 1: Added `deploy-ldap` Makefile target using idempotent namespace creation and label selector pod wait. Added `admins-group` LDAP entry to configmap seed data (avoids needing `ldapadd` at deploy time). Wired `deploy-ldap` into `integration` target dependency list.
- Task 1.3: Bind credentials secret is created programmatically in the test (not as a YAML fixture), since the data fields are simple and the test controls the namespace.
- Task 2: Added `GetLDAPAuthEngineConfigInstance` and `GetLDAPAuthEngineGroupInstance` to decoder.go following the established pattern.
- Task 3: Created 3 YAML fixtures in `test/ldapauthengine/` — mount (type=ldap), config (with explicit TLS version fields), group (mapping test-ldap-admins to vault-admin policy).
- Task 4: Created `ldapauthengine_controller_test.go` with 4 ordered contexts: prerequisites (secret + mount), config create+verify, group create+verify, delete+cleanup. Used checked type assertions throughout.
- Task 5: `make integration` passed — all 44 specs green, no regressions. LDAP deployment in Kind works cleanly via `deploy-ldap` target. Existing `auth/kubernetes` mount unaffected.
- All 5 ACs satisfied: OpenLDAP deployed in Kind (AC1), LDAP config written to Vault with correct url/binddn/insecure_tls (AC2), group mapping created with correct policies (AC3), group deleted with Vault cleanup (AC4), config deleted without Vault cleanup (AC5).

### File List

| # | File | Change Type |
|---|------|-------------|
| 1 | `Makefile` | Modified |
| 2 | `integration/ldap/configmap.yaml` | Modified |
| 3 | `api/v1alpha1/ldapauthengineconfig_types.go` | Modified |
| 4 | `controllers/controllertestutils/decoder.go` | Modified |
| 5 | `test/ldapauthengine/test-ldap-auth-mount.yaml` | New |
| 6 | `test/ldapauthengine/test-ldap-auth-config.yaml` | New |
| 7 | `test/ldapauthengine/test-ldap-auth-group.yaml` | New |
| 8 | `controllers/ldapauthengine_controller_test.go` | New |
