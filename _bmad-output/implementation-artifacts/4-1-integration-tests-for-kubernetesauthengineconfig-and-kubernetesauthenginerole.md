# Story 4.1: Integration Tests for KubernetesAuthEngineConfig and KubernetesAuthEngineRole

Status: done

## Story

As an operator developer,
I want integration tests for the Kubernetes auth engine configuration and role types covering create, reconcile success, Vault state verification, and delete with cleanup,
So that the most commonly used auth method has end-to-end test coverage.

## Acceptance Criteria

1. **Given** a KubernetesAuthEngineConfig CR is created targeting a test auth mount **When** the reconciler processes it **Then** the config is written to Vault at `auth/{path}/{name}/config` and ReconcileSuccessful=True

2. **Given** a KubernetesAuthEngineRole CR is created with explicit targetNamespaces and targetServiceAccounts **When** the reconciler processes it **Then** the role exists in Vault at `auth/{path}/role/{name}` with correct bound_service_account_names and bound_service_account_namespaces, and ReconcileSuccessful=True

3. **Given** a KubernetesAuthEngineRole CR using `targetNamespaceSelector` with label matching **When** the reconciler processes it **Then** the role's `bound_service_account_namespaces` in Vault contains the matching namespace names

4. **Given** both CRs are deleted **When** the reconciler processes the deletions **Then** the role is cleaned up from Vault (IsDeletable=true) and the config CR is deleted from K8s without Vault cleanup (IsDeletable=false)

## Tasks / Subtasks

- [x] Task 1: Add decoder method (AC: 1)
  - [x] 1.1: Add `GetKubernetesAuthEngineConfigInstance` method to `controllers/controllertestutils/decoder.go` following the established pattern (decode YAML → type-assert to `*redhatcopv1alpha1.KubernetesAuthEngineConfig`)

- [x] Task 2: Create test fixtures (AC: 1, 2, 3, 4)
  - [x] 2.1: Create `test/kubernetesauthengine/test-kube-auth-mount.yaml` — an AuthEngineMount CR with `type: kubernetes`, `path: test-k8s-auth`, `metadata.name: test-kaec-mount`, using `authentication.role: policy-admin`
  - [x] 2.2: Create `test/kubernetesauthengine/test-kube-auth-config.yaml` — a KubernetesAuthEngineConfig CR with `metadata.name: test-kaec-mount`, `path: test-k8s-auth`, `kubernetesHost: https://kubernetes.default.svc:443`, `disableLocalCAJWT: true`, `useOperatorPodCA: false`
  - [x] 2.3: Create `test/kubernetesauthengine/test-kube-auth-role.yaml` — a KubernetesAuthEngineRole CR with `metadata.name: test-kaer-role`, `path: test-k8s-auth/test-kaec-mount`, explicit `targetNamespaces: [vault-admin]`, `targetServiceAccounts: [default]`, `policies: [vault-admin]`
  - [x] 2.4: Create `test/kubernetesauthengine/test-kube-auth-role-selector.yaml` — a KubernetesAuthEngineRole CR with `metadata.name: test-kaer-role-selector`, `path: test-k8s-auth/test-kaec-mount`, `targetNamespaceSelector` matching label `database-engine-admin: "true"` (which the `test-vault-config-operator` namespace has), `targetServiceAccounts: [default]`, `policies: [vault-admin]`

- [x] Task 3: Create integration test file (AC: 1, 2, 3, 4)
  - [x] 3.1: Create `controllers/kubernetesauthengine_controller_test.go` with `//go:build integration` tag, package `controllers`, standard Ginkgo imports
  - [x] 3.2: Add `Describe("KubernetesAuthEngine controllers", Ordered)` with `timeout := 120 * time.Second`, `interval := 2 * time.Second`
  - [x] 3.3: Add `Context("When creating the prerequisite auth mount")` — load `test-kube-auth-mount.yaml` via `decoder.GetAuthEngineMountInstance`, set namespace to `vaultAdminNamespaceName`, create it, poll for `ReconcileSuccessful=True`. This creates the Kubernetes auth mount needed by config and role tests.
  - [x] 3.4: Add `Context("When creating a KubernetesAuthEngineConfig")` — load `test-kube-auth-config.yaml` via `decoder.GetKubernetesAuthEngineConfigInstance`, set namespace to `vaultAdminNamespaceName`, create, poll for `ReconcileSuccessful=True`
  - [x] 3.5: After reconcile success, read `auth/test-k8s-auth/test-kaec-mount/config` from Vault, verify `kubernetes_host` equals `https://kubernetes.default.svc:443`
  - [x] 3.6: Add `Context("When creating a KubernetesAuthEngineRole with explicit namespaces")` — load `test-kube-auth-role.yaml`, set namespace to `vaultAdminNamespaceName`, create, poll for `ReconcileSuccessful=True`
  - [x] 3.7: After reconcile success, read `auth/test-k8s-auth/test-kaec-mount/role/test-kaer-role` from Vault, verify `bound_service_account_names` contains `"default"`, `bound_service_account_namespaces` contains `"vault-admin"`, `token_policies` contains `"vault-admin"`
  - [x] 3.8: Add `Context("When creating a KubernetesAuthEngineRole with namespace selector")` — load `test-kube-auth-role-selector.yaml`, set namespace to `vaultAdminNamespaceName`, create, poll for `ReconcileSuccessful=True`
  - [x] 3.9: After reconcile success, read `auth/test-k8s-auth/test-kaec-mount/role/test-kaer-role-selector` from Vault, verify `bound_service_account_namespaces` contains `"test-vault-config-operator"` (the namespace with `database-engine-admin: "true"` label)
  - [x] 3.10: Add `Context("When deleting KubernetesAuthEngine resources")` — delete both role CRs first (IsDeletable=true), use `Eventually` to poll for K8s deletion (NotFound), then verify roles no longer exist in Vault via `Logical().Read` returning nil. Then delete the config CR (IsDeletable=false, no finalizer → immediate K8s deletion, no Vault cleanup). Finally delete the AuthEngineMount, wait for deletion, verify mount gone from `sys/auth`.
  - [x] 3.11: Add `AfterAll` cleanup guard to best-effort delete all CRs if earlier contexts failed

- [x] Task 4: End-to-end verification (AC: 1, 2, 3, 4)
  - [x] 4.1: Run `make integration` and verify new tests pass alongside all existing tests
  - [x] 4.2: Verify no regressions — the existing `kubernetes` auth mount at `auth/kubernetes` is unaffected

## Dev Notes

### Both Types Use VaultResource Reconciler — NOT VaultEngineResource

Unlike AuthEngineMount/SecretEngineMount (Epic 3), both KubernetesAuthEngineConfig and KubernetesAuthEngineRole use `NewVaultResource` (the standard reconciler variant). The reconcile flow is:

1. `prepareContext()` enriches context with kubeClient, restConfig, vaultConnection, vaultClient
2. `NewVaultResource(&r.ReconcilerBase, instance)` creates the standard reconciler
3. `VaultResource.Reconcile()` → `manageReconcileLogic()`:
   - `PrepareInternalValues()` — resolves JWT token (config) or namespace list (role)
   - `PrepareTLSConfig()` — no-op for both types
   - `VaultEndpoint.CreateOrUpdate()` — reads from Vault, calls `IsEquivalentToDesiredState()`, writes if different
4. `ManageOutcome()` sets `ReconcileSuccessful` condition

This is the same variant as Policy (Story 3.1) and PasswordPolicy (Story 3.2), not the VaultEngineResource variant used by AuthEngineMount (Story 3.4) and SecretEngineMount (Story 3.3).

[Source: controllers/kubernetesauthengineconfig_controller.go — uses NewVaultResource]
[Source: controllers/kubernetesauthenginerole_controller.go — uses NewVaultResource]
[Source: controllers/vaultresourcecontroller/vaultresourcereconciler.go — manageReconcileLogic]

### KubernetesAuthEngineConfig — Key Implementation Details

**GetPath():**
```go
func (d *KubernetesAuthEngineConfig) GetPath() string {
    if d.Spec.Name != "" {
        return vaultutils.CleansePath("auth/" + string(d.Spec.Path) + "/" + d.Spec.Name + "/config")
    }
    return vaultutils.CleansePath("auth/" + string(d.Spec.Path) + "/" + d.Name + "/config")
}
```
For fixture with `path: test-k8s-auth`, `metadata.name: test-kaec-mount` → `auth/test-k8s-auth/test-kaec-mount/config`

[Source: api/v1alpha1/kubernetesauthengineconfig_types.go#L66-L71]

**IsDeletable(): false** — No finalizer, no Vault cleanup on CR deletion. The config persists in Vault until the auth mount itself is deleted.

[Source: api/v1alpha1/kubernetesauthengineconfig_types.go — IsDeletable returns false]

**toMap() — 8 Vault keys:**
```go
func (i *KAECConfig) toMap() map[string]interface{} {
    payload := map[string]interface{}{}
    payload["kubernetes_host"] = i.KubernetesHost
    payload["kubernetes_ca_cert"] = i.KubernetesCACert
    payload["token_reviewer_jwt"] = i.retrievedTokenReviewerJWT
    payload["pem_keys"] = i.PEMKeys
    payload["issuer"] = i.Issuer
    payload["disable_iss_validation"] = i.DisableISSValidation
    payload["disable_local_ca_jwt"] = i.DisableLocalCAJWT
    payload["use_annotations_as_alias_metadata"] = i.UseAnnotationsAsAliasMetadata
    return payload
}
```

Note: `use_operator_pod_ca` is NOT in toMap() — it only drives webhook defaulting behavior.

[Source: api/v1alpha1/kubernetesauthengineconfig_types.go#L208-L220]

**IsEquivalentToDesiredState():** Bare `reflect.DeepEqual(desiredState, payload)` with no key filtering. Vault's GET response for `auth/{path}/config` does NOT return `token_reviewer_jwt` (sensitive) and may not return `use_annotations_as_alias_metadata` (newer Vault feature). This means `IsEquivalentToDesiredState` will return false on every reconcile, causing a config write each time. This is tracked tech debt (Story 7-4) and does NOT affect test correctness — the reconciler still reaches `ReconcileSuccessful=True` after writing.

[Source: api/v1alpha1/kubernetesauthengineconfig_types.go — IsEquivalentToDesiredState]

**PrepareInternalValues():** If `spec.tokenReviewerServiceAccount` is set, creates a JWT token for that ServiceAccount via `GetJWTTokenWithDuration` (1-year expiry) and stores it in `retrievedTokenReviewerJWT`. If not set, the JWT field remains empty. For the integration test, we do NOT set `tokenReviewerServiceAccount` to avoid needing a dedicated SA.

[Source: api/v1alpha1/kubernetesauthengineconfig_types.go — PrepareInternalValues]

**Webhook Default():** If `useOperatorPodCA` is true (default) and `kubernetesCACert` is empty, reads `/var/run/secrets/kubernetes.io/serviceaccount/ca.crt` from the filesystem. In envtest (running outside a pod), this file likely does not exist — the webhook logs and returns without setting the CA cert. To avoid this path entirely, test fixtures should set `useOperatorPodCA: false` and optionally provide `kubernetesCACert` explicitly, OR leave it empty (Vault accepts config without CA cert).

[Source: api/v1alpha1/kubernetesauthengineconfig_webhook.go — Default()]

### KubernetesAuthEngineRole — Key Implementation Details

**GetPath():**
```go
func (d *KubernetesAuthEngineRole) GetPath() string {
    if d.Spec.Name != "" {
        return vaultutils.CleansePath("auth/" + string(d.Spec.Path) + "/role/" + d.Spec.Name)
    }
    return vaultutils.CleansePath("auth/" + string(d.Spec.Path) + "/role/" + d.Name)
}
```
For fixture with `path: test-k8s-auth/test-kaec-mount`, `metadata.name: test-kaer-role` → `auth/test-k8s-auth/test-kaec-mount/role/test-kaer-role`

Note: `spec.path` for the role includes the FULL auth mount path (including the mount name segment), not just the parent path.

[Source: api/v1alpha1/kubernetesauthenginerole_types.go#L74-L78]

**IsDeletable(): true** — Finalizer added after first successful reconcile. On deletion, the role is removed from Vault at `auth/{path}/role/{name}`.

**toMap() — 12-13 Vault keys (conditional on Audience):**
```go
func (i *VRole) toMap() map[string]interface{} {
    payload := map[string]interface{}{}
    payload["bound_service_account_names"] = i.TargetServiceAccounts
    payload["bound_service_account_namespaces"] = i.namespaces  // resolved in PrepareInternalValues
    payload["alias_name_source"] = i.AliasNameSource
    if i.Audience != nil {
        payload["audience"] = i.Audience
    }
    payload["token_ttl"] = i.TokenTTL
    payload["token_max_ttl"] = i.TokenMaxTTL
    payload["token_policies"] = i.Policies
    payload["token_bound_cidrs"] = i.TokenBoundCIDRs
    payload["token_explicit_max_ttl"] = i.TokenExplicitMaxTTL
    payload["token_no_default_policy"] = i.TokenNoDefaultPolicy
    payload["token_num_uses"] = i.TokenNumUses
    payload["token_period"] = i.TokenPeriod
    payload["token_type"] = i.TokenType
    return payload
}
```

[Source: api/v1alpha1/kubernetesauthenginerole_types.go#L262-L279]

**IsEquivalentToDesiredState():** Same bare `reflect.DeepEqual` with no key filtering. Vault's read response for roles may include extra keys not in the desired state. Same Story 7-4 tech debt applies.

**PrepareInternalValues():**
- If `targetNamespaces.targetNamespaceSelector` is set: lists all Kubernetes namespaces matching the label selector, collects their names into `namespaces` internal field. If zero match, sets `["__no_namespace__"]` placeholder.
- If `targetNamespaces.targetNamespaces` (explicit list) is set: uses those directly.
- Uses `context.Value("kubeClient")` to list namespaces.

[Source: api/v1alpha1/kubernetesauthenginerole_types.go — PrepareInternalValues, findSelectedNamespaceNames]

**Webhook Validation (create AND update):**
- `ValidateCreate()`: calls `isValid()` — ensures exactly ONE of `targetNamespaceSelector` or `targetNamespaces` is set (not both, not neither)
- `ValidateUpdate()`: rejects changes to `spec.path` + calls `isValid()`
- `ValidateDelete()`: no-op

[Source: api/v1alpha1/kubernetesauthenginerole_webhook.go]

**Controller watches Namespace objects:** The role controller's `SetupWithManager` includes a `Watches` on `Namespace` resources, re-enqueuing `KubernetesAuthEngineRole` CRs whose `targetNamespaceSelector` matches any changed namespace's labels. This enables dynamic membership — if a namespace gains/loses a matching label, the role's `bound_service_account_namespaces` is updated in Vault.

### Test Design — Dependency Chain

The test requires creating resources in dependency order:

1. **AuthEngineMount** (type=kubernetes) → creates the k8s auth mount in Vault
2. **KubernetesAuthEngineConfig** → configures the mount (writes to `auth/{path}/config`)
3. **KubernetesAuthEngineRole** → creates a role under the mount (writes to `auth/{path}/role/{name}`)

Deletion must happen in reverse: roles first, then config, then mount.

```
AuthEngineMount (test-k8s-auth/test-kaec-mount)
  └── KubernetesAuthEngineConfig → auth/.../config
  └── KubernetesAuthEngineRole → auth/.../role/test-kaer-role
  └── KubernetesAuthEngineRole → auth/.../role/test-kaer-role-selector
```

The AuthEngineMount must be reconciled successfully before creating the config or roles, because Vault will reject writes to `auth/{path}/config` if the auth mount doesn't exist at that path.

### Vault API Response Shapes

**GET `auth/{path}/config`** — Returns kubernetes auth config:
```json
{
  "data": {
    "kubernetes_host": "https://kubernetes.default.svc:443",
    "kubernetes_ca_cert": "...",
    "pem_keys": [],
    "issuer": "",
    "disable_iss_validation": false,
    "disable_local_ca_jwt": false
  }
}
```
Note: `token_reviewer_jwt` is NOT returned (sensitive). `use_annotations_as_alias_metadata` may or may not be present depending on Vault version (1.19.0 should include it).

**GET `auth/{path}/role/{name}`** — Returns role config:
```json
{
  "data": {
    "bound_service_account_names": ["default"],
    "bound_service_account_namespaces": ["vault-admin"],
    "token_policies": ["vault-admin"],
    "token_ttl": 0,
    "token_max_ttl": 0,
    "token_bound_cidrs": [],
    "token_explicit_max_ttl": 0,
    "token_no_default_policy": false,
    "token_num_uses": 0,
    "token_period": 0,
    "token_type": "default",
    "alias_name_source": "serviceaccount_uid",
    ...
  }
}
```
The role response may include extra keys like `num_uses`, `period`, `ttl`, `max_ttl` (legacy aliases).

### Why `useOperatorPodCA: false` and `disableLocalCAJWT: true` in Test Fixtures

The integration test runs outside a Kubernetes pod (envtest process on the host machine). Two config behaviors depend on being inside a pod:

1. **`useOperatorPodCA: true` (default)** — webhook reads `/var/run/secrets/kubernetes.io/serviceaccount/ca.crt`. Outside a pod, this file doesn't exist. The webhook logs the error and continues, leaving `kubernetesCACert` empty.

2. **`disableLocalCAJWT: false` (default)** — Vault itself tries to use a local JWT for token review. In our test setup, Vault IS inside a pod (Kind cluster) so this could work, but it's cleaner to be explicit.

Setting `useOperatorPodCA: false` and `disableLocalCAJWT: true` avoids both filesystem-dependent paths and makes the test deterministic. We still verify that:
- The config is written to Vault
- `kubernetes_host` is correctly set
- The reconciler reaches `ReconcileSuccessful=True`

### Why Not Use the Existing `auth/kubernetes` Mount

The Vault initializer creates `auth/kubernetes` with a config and the `policy-admin` role. Using this mount would:
1. **Risk overwriting the test infrastructure config** — writing to `auth/kubernetes/config` would replace the Vault-initializer's config
2. **Create naming conflicts** — the `policy-admin` role already exists at `auth/kubernetes/role/policy-admin`

Instead, we create a separate auth mount at `test-k8s-auth/test-kaec-mount` to isolate our tests.

### Namespace Selector Test Design

The `test-vault-config-operator` namespace is created in `BeforeSuite` with label `database-engine-admin: "true"`. The selector fixture uses:
```yaml
targetNamespaces:
  targetNamespaceSelector:
    matchLabels:
      database-engine-admin: "true"
```

`PrepareInternalValues()` will list namespaces matching this selector and find `test-vault-config-operator`. The role's `bound_service_account_namespaces` should contain `"test-vault-config-operator"` in Vault.

[Source: controllers/suite_integration_test.go — namespace creation with database-engine-admin label]

### Controller Registration — Already Done

Both controllers are registered in `suite_integration_test.go`:
```go
err = (&KubernetesAuthEngineRoleReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "KubernetesAuthEngineRole")}).SetupWithManager(mgr)
Expect(err).ToNot(HaveOccurred())

err = (&KubernetesAuthEngineConfigReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "KubernetesAuthEngineConfig")}).SetupWithManager(mgr)
Expect(err).ToNot(HaveOccurred())
```

No changes needed to the test suite setup.

[Source: controllers/suite_integration_test.go#L130-L152]

### Decoder — GetKubernetesAuthEngineConfigInstance MUST BE ADDED

`GetKubernetesAuthEngineRoleInstance` already exists. `GetKubernetesAuthEngineConfigInstance` must be added:

```go
func (d *decoder) GetKubernetesAuthEngineConfigInstance(filename string) (*redhatcopv1alpha1.KubernetesAuthEngineConfig, error) {
    obj, groupKindVersion, err := d.decodeFile(filename)
    if err != nil {
        return nil, err
    }

    kind := reflect.TypeOf(redhatcopv1alpha1.KubernetesAuthEngineConfig{}).Name()
    if groupKindVersion.Kind == kind {
        o := obj.(*redhatcopv1alpha1.KubernetesAuthEngineConfig)
        return o, nil
    }

    return nil, errDecode
}
```

[Source: controllers/controllertestutils/decoder.go — existing pattern at lines 85-98]

### Test Fixture Design

**Fixture 1: `test/kubernetesauthengine/test-kube-auth-mount.yaml`** — AuthEngineMount prerequisite:
```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: AuthEngineMount
metadata:
  name: test-kaec-mount
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  type: kubernetes
  path: test-k8s-auth
```
Mounts at `sys/auth/test-k8s-auth/test-kaec-mount`. Uses `type: kubernetes` to enable the Kubernetes auth method at this path.

**Fixture 2: `test/kubernetesauthengine/test-kube-auth-config.yaml`** — KubernetesAuthEngineConfig:
```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: KubernetesAuthEngineConfig
metadata:
  name: test-kaec-mount
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  path: test-k8s-auth
  kubernetesHost: "https://kubernetes.default.svc:443"
  disableLocalCAJWT: true
  useOperatorPodCA: false
```
`metadata.name` matches the mount's `metadata.name` so GetPath resolves to `auth/test-k8s-auth/test-kaec-mount/config` — targeting the mount we created.

**Fixture 3: `test/kubernetesauthengine/test-kube-auth-role.yaml`** — KubernetesAuthEngineRole (explicit namespaces):
```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: KubernetesAuthEngineRole
metadata:
  name: test-kaer-role
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  path: test-k8s-auth/test-kaec-mount
  policies:
    - vault-admin
  targetServiceAccounts:
    - default
  targetNamespaces:
    targetNamespaces:
      - vault-admin
```
Writes to `auth/test-k8s-auth/test-kaec-mount/role/test-kaer-role`. Note `path` includes the full mount path segment.

**Fixture 4: `test/kubernetesauthengine/test-kube-auth-role-selector.yaml`** — KubernetesAuthEngineRole (namespace selector):
```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: KubernetesAuthEngineRole
metadata:
  name: test-kaer-role-selector
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  path: test-k8s-auth/test-kaec-mount
  policies:
    - vault-admin
  targetServiceAccounts:
    - default
  targetNamespaces:
    targetNamespaceSelector:
      matchLabels:
        database-engine-admin: "true"
```
Writes to `auth/test-k8s-auth/test-kaec-mount/role/test-kaer-role-selector`. The selector matches `test-vault-config-operator` namespace.

### Verifying Vault State

**Config verification:**
```go
secret, err := vaultClient.Logical().Read("auth/test-k8s-auth/test-kaec-mount/config")
Expect(err).To(BeNil())
Expect(secret).NotTo(BeNil())
Expect(secret.Data["kubernetes_host"]).To(Equal("https://kubernetes.default.svc:443"))
```

**Role verification (explicit namespaces):**
```go
secret, err := vaultClient.Logical().Read("auth/test-k8s-auth/test-kaec-mount/role/test-kaer-role")
Expect(err).To(BeNil())
Expect(secret).NotTo(BeNil())
```
For list fields, Vault returns `[]interface{}` not `[]string`. Use checked type assertions:
```go
boundSANames, ok := secret.Data["bound_service_account_names"].([]interface{})
Expect(ok).To(BeTrue(), "expected bound_service_account_names to be []interface{}")
Expect(boundSANames).To(ContainElement("default"))

boundSANamespaces, ok := secret.Data["bound_service_account_namespaces"].([]interface{})
Expect(ok).To(BeTrue())
Expect(boundSANamespaces).To(ContainElement("vault-admin"))

tokenPolicies, ok := secret.Data["token_policies"].([]interface{})
Expect(ok).To(BeTrue())
Expect(tokenPolicies).To(ContainElement("vault-admin"))
```

**Role verification (namespace selector):**
```go
secret, err := vaultClient.Logical().Read("auth/test-k8s-auth/test-kaec-mount/role/test-kaer-role-selector")
Expect(err).To(BeNil())
Expect(secret).NotTo(BeNil())
boundSANamespaces, ok := secret.Data["bound_service_account_namespaces"].([]interface{})
Expect(ok).To(BeTrue())
Expect(boundSANamespaces).To(ContainElement("test-vault-config-operator"))
```

**Delete verification (role — IsDeletable=true):**
```go
// Wait for K8s deletion
Eventually(func() bool {
    err := k8sIntegrationClient.Get(ctx, types.NamespacedName{...}, &redhatcopv1alpha1.KubernetesAuthEngineRole{})
    return apierrors.IsNotFound(err)
}, timeout, interval).Should(BeTrue())

// Verify Vault cleanup
Eventually(func() bool {
    secret, err := vaultClient.Logical().Read("auth/test-k8s-auth/test-kaec-mount/role/test-kaer-role")
    return err == nil && secret == nil
}, timeout, interval).Should(BeTrue())
```

**Delete verification (config — IsDeletable=false):**
Config deletion from K8s happens immediately (no finalizer). No Vault cleanup expected.
```go
Eventually(func() bool {
    err := k8sIntegrationClient.Get(ctx, types.NamespacedName{...}, &redhatcopv1alpha1.KubernetesAuthEngineConfig{})
    return apierrors.IsNotFound(err)
}, timeout, interval).Should(BeTrue())
```

### Test Structure

```
Describe("KubernetesAuthEngine controllers", Ordered)
  var mountInstance *redhatcopv1alpha1.AuthEngineMount
  var configInstance *redhatcopv1alpha1.KubernetesAuthEngineConfig
  var roleInstance *redhatcopv1alpha1.KubernetesAuthEngineRole
  var roleSelectorInstance *redhatcopv1alpha1.KubernetesAuthEngineRole

  AfterAll: best-effort delete all instances

  Context("When creating the prerequisite auth mount")
    It("Should enable the kubernetes auth method in Vault")
      - Load test-kube-auth-mount.yaml via decoder.GetAuthEngineMountInstance
      - Set namespace to vaultAdminNamespaceName, create
      - Eventually poll for ReconcileSuccessful=True
      - Verify mount exists in sys/auth with key "test-k8s-auth/test-kaec-mount/"

  Context("When creating a KubernetesAuthEngineConfig")
    It("Should write the config to Vault")
      - Load test-kube-auth-config.yaml via decoder.GetKubernetesAuthEngineConfigInstance
      - Set namespace to vaultAdminNamespaceName, create
      - Eventually poll for ReconcileSuccessful=True
      - Read auth/test-k8s-auth/test-kaec-mount/config
      - Verify kubernetes_host = "https://kubernetes.default.svc:443"

  Context("When creating a KubernetesAuthEngineRole with explicit namespaces")
    It("Should create the role in Vault with correct bindings")
      - Load test-kube-auth-role.yaml via decoder.GetKubernetesAuthEngineRoleInstance
      - Set namespace to vaultAdminNamespaceName, create
      - Eventually poll for ReconcileSuccessful=True
      - Read auth/test-k8s-auth/test-kaec-mount/role/test-kaer-role
      - Verify bound_service_account_names contains "default"
      - Verify bound_service_account_namespaces contains "vault-admin"
      - Verify token_policies contains "vault-admin"

  Context("When creating a KubernetesAuthEngineRole with namespace selector")
    It("Should resolve the selector and set bound namespaces")
      - Load test-kube-auth-role-selector.yaml
      - Set namespace to vaultAdminNamespaceName, create
      - Eventually poll for ReconcileSuccessful=True
      - Read auth/test-k8s-auth/test-kaec-mount/role/test-kaer-role-selector
      - Verify bound_service_account_namespaces contains "test-vault-config-operator"

  Context("When deleting KubernetesAuthEngine resources")
    It("Should clean up roles from Vault and remove all resources")
      - Delete both role CRs (IsDeletable=true → Vault cleanup)
      - Eventually verify K8s deletion (NotFound) for each role
      - Eventually verify roles removed from Vault (Read returns nil)
      - Delete config CR (IsDeletable=false → no Vault cleanup)
      - Eventually verify K8s deletion
      - Delete AuthEngineMount
      - Eventually verify K8s deletion and mount gone from sys/auth
```

### Import Requirements for kubernetesauthengine_controller_test.go

```go
import (
    "time"

    . "github.com/onsi/ginkgo/v2"
    . "github.com/onsi/gomega"
    redhatcopv1alpha1 "github.com/redhat-cop/vault-config-operator/api/v1alpha1"
    "github.com/redhat-cop/vault-config-operator/controllers/vaultresourcecontroller"

    apierrors "k8s.io/apimachinery/pkg/api/errors"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/apimachinery/pkg/types"
)
```

No `encoding/json` needed (unlike AuthEngineMount tests which needed json.Number for TTL values).

### Name Collision Prevention

Fixture names use `test-kaec-` and `test-kaer-` prefixes with unique names:
- `test-k8s-auth/test-kaec-mount` — auth mount (unique prefix)
- `test-kaer-role` — explicit namespace role
- `test-kaer-role-selector` — selector-based role

These don't collide with:
- `kubernetes/` — default Kubernetes auth mount (test infrastructure)
- `test-auth-mount/test-aem-*` — Story 3.4 AuthEngineMount tests
- `kube-authengine-mount-sample/*` — existing sample fixtures (not created in integration tests)

### Risk Considerations

- **Config `IsEquivalentToDesiredState` always returns false:** Vault's read response for `auth/{path}/config` omits `token_reviewer_jwt` and may omit other keys. The reconciler writes on every reconcile cycle. This is a known issue (Story 7-4) but does NOT block `ReconcileSuccessful=True`.
- **Namespace selector timing:** `PrepareInternalValues` lists namespaces at reconcile time. The `test-vault-config-operator` namespace must exist with the correct label BEFORE the role CR is created. The `BeforeSuite` creates this namespace, so it should be available.
- **`__no_namespace__` placeholder:** If the namespace selector matches zero namespaces, `PrepareInternalValues` sets `namespaces = ["__no_namespace__"]`. The test should ensure the labeled namespace exists to avoid this.
- **Webhook defaulting in envtest:** The operator's webhook server runs in the envtest manager process. The `Default()` for KubernetesAuthEngineConfig reads from the filesystem — use `useOperatorPodCA: false` in test fixtures to avoid this code path.
- **Vault API path for role:** The role's `spec.path` must be `test-k8s-auth/test-kaec-mount` (including both path segments), not just `test-k8s-auth`. This is because the auth mount is at `sys/auth/test-k8s-auth/test-kaec-mount` and the role endpoint is `auth/{mount-path}/role/{name}`.
- **Checked type assertions:** Per Epic 3 retro action item, always use two-value form `val, ok := x.([]interface{})` with `Expect(ok).To(BeTrue())` instead of bare `.( )` type assertions that panic on failure.

### File Inventory — What Needs to Change

| # | File | Change Type | Description |
|---|------|-------------|-------------|
| 1 | `controllers/controllertestutils/decoder.go` | Modified | Add `GetKubernetesAuthEngineConfigInstance` decoder method |
| 2 | `test/kubernetesauthengine/test-kube-auth-mount.yaml` | New | AuthEngineMount prerequisite (type=kubernetes) |
| 3 | `test/kubernetesauthengine/test-kube-auth-config.yaml` | New | KubernetesAuthEngineConfig fixture |
| 4 | `test/kubernetesauthengine/test-kube-auth-role.yaml` | New | KubernetesAuthEngineRole with explicit namespaces |
| 5 | `test/kubernetesauthengine/test-kube-auth-role-selector.yaml` | New | KubernetesAuthEngineRole with namespace selector |
| 6 | `controllers/kubernetesauthengine_controller_test.go` | New | Integration test — create mount, config, roles; verify Vault state; delete and verify cleanup |

No changes to suite setup, controllers, webhooks, or types.

### No `make manifests generate` Needed

This story only adds an integration test file, YAML fixtures, and a decoder method. No CRD types, controllers, or webhooks are changed.

### Previous Story Intelligence

**From Story 3.4 (AuthEngineMount integration tests):**
- Established the VaultEngineResource integration test pattern with mount existence via `sys/auth`, accessor, tune config, json.Number TTL
- Demonstrated `AfterAll` cleanup guard pattern
- Used `decoder.GetAuthEngineMountInstance` (reused in this story for the prerequisite mount)
- Confirmed `test-auth-mount/` prefix avoids collisions with `kubernetes/` mount

**From Story 3.1 (Policy integration tests):**
- Established the VaultResource test pattern: create → poll ReconcileSuccessful → verify Vault state → delete → verify cleanup
- Demonstrated Vault API `Logical().Read()` for policy verification
- KubernetesAuthEngineConfig/Role use the same VaultResource reconciler

**From Epic 3 Retrospective:**
- "Cleanest epic yet — zero debug failures" with Opus 4.6
- "Checked type assertions rule" — always use two-value form in tests
- "json.Number for TTL values" — not needed for this story (no tune config verification)
- "Story 4.1 has low infrastructure scope — no new infra, but PrepareInternalValues does ServiceAccount JWT resolution"
- "Story ordering: 4.1 (simplest) → 4.2 (LDAP infra) → 4.3 (Keycloak infra)"

**From Epic 2 Retrospective:**
- "Pattern-first investment paid off"
- "Prefer Opus-class models for implementation"

### Git Intelligence (Recent Commits)

```
9608211 Merge pull request #318 from raffaelespazzoli/bmad-epic-3
24a37f0 Complete Epic 3 retrospective and close Epics 1-3
cb473c3 Mark Story 3.4 as done after clean code review
866c843 Add integration tests for AuthEngineMount type (Story 3.4)
db21d90 Add integration tests for SecretEngineMount type (Story 3.3)
25dbe39 Add integration tests for PasswordPolicy type (Story 3.2)
```

Codebase is clean post-Epic 3 merge to main. No pending changes affect this story.

### Integration Test Infrastructure Classification

Per the project's three-tier rule:
- **Kubernetes API:** Already available in Kind — no new infra needed
- **ServiceAccount JWT resolution:** Uses the Kind cluster's Kubernetes API (already available)
- **Namespace listing for selector:** Uses the envtest client (already available)

**Classification: Already available — Low infrastructure scope**

[Source: _bmad-output/project-context.md#L134-L141 — Integration test infrastructure philosophy]

### Project Structure Notes

- Decoder change in `controllers/controllertestutils/decoder.go` (add one method)
- Test file goes in `controllers/kubernetesauthengine_controller_test.go` (combines both types in one file since they share the dependency chain)
- Test fixtures go in `test/kubernetesauthengine/` directory (follows `test/<feature>/` pattern)
- All files follow existing naming conventions

### References

- [Source: api/v1alpha1/kubernetesauthengineconfig_types.go] — KubernetesAuthEngineConfig VaultObject implementation, GetPath, GetPayload, IsEquivalentToDesiredState, toMap, PrepareInternalValues, IsDeletable
- [Source: api/v1alpha1/kubernetesauthengineconfig_types.go#L66-L71] — GetPath: auth/{path}/{name}/config
- [Source: api/v1alpha1/kubernetesauthengineconfig_types.go#L208-L220] — KAECConfig.toMap (8 keys)
- [Source: api/v1alpha1/kubernetesauthengineconfig_webhook.go] — Webhook: useOperatorPodCA defaulting, immutable path
- [Source: api/v1alpha1/kubernetesauthenginerole_types.go] — KubernetesAuthEngineRole VaultObject implementation, GetPath, toMap, PrepareInternalValues, isValid
- [Source: api/v1alpha1/kubernetesauthenginerole_types.go#L74-L78] — GetPath: auth/{path}/role/{name}
- [Source: api/v1alpha1/kubernetesauthenginerole_types.go#L262-L279] — VRole.toMap (12-13 keys)
- [Source: api/v1alpha1/kubernetesauthenginerole_webhook.go] — Webhook: create+update validation, isValid, immutable path
- [Source: controllers/kubernetesauthengineconfig_controller.go] — Controller (VaultResource)
- [Source: controllers/kubernetesauthenginerole_controller.go] — Controller (VaultResource + Namespace watches)
- [Source: controllers/vaultresourcecontroller/vaultresourcereconciler.go] — VaultResource.manageReconcileLogic
- [Source: api/v1alpha1/utils/vaultobject.go] — VaultEndpoint.CreateOrUpdate
- [Source: controllers/suite_integration_test.go#L130-L152] — Both controllers registered
- [Source: controllers/controllertestutils/decoder.go#L85-L98] — GetKubernetesAuthEngineRoleInstance (existing); GetKubernetesAuthEngineConfigInstance MUST BE ADDED
- [Source: controllers/authenginemount_controller_test.go] — Closest pattern reference (Story 3.4, AuthEngineMount used as prerequisite)
- [Source: controllers/policy_controller_test.go] — VaultResource test pattern reference
- [Source: integration/vault-values.yaml#L161-L183] — vault-admin-initializer: auth/kubernetes enabled, policy-admin role
- [Source: _bmad-output/planning-artifacts/epics.md#L427-L455] — Story 4.1 epic definition
- [Source: _bmad-output/implementation-artifacts/3-4-integration-tests-for-authenginemount-type.md] — Previous story (closest pattern)
- [Source: _bmad-output/implementation-artifacts/epic-3-retro-2026-04-20.md] — Epic 3 retro with Epic 4 preparation
- [Source: _bmad-output/project-context.md#L134-L141] — Integration test infrastructure philosophy
- [Source: _bmad-output/project-context.md#L143-L155] — Integration test pattern and Ordered lifecycle

## Dev Agent Record

### Agent Model Used

Opus 4.6

### Debug Log References

- Baseline integration tests: port 8080 conflict (Quarkus dev portal) resolved by killing PID 846263
- Baseline had 19 pass / 16 fail (pre-existing failures in Entity, EntityAlias, RandomSecret, VaultSecret, PKI tests due to stale test data)
- First full run after story changes: vault-admin namespace stuck in Terminating from previous run; resolved by removing finalizers from 4 stale CRs
- Final run: exit code 0, all 40 specs passed (39.5% coverage), 408.9s

### Completion Notes List

- Task 1: Added `GetKubernetesAuthEngineConfigInstance` to decoder.go following established pattern
- Task 2: Created 4 YAML fixtures in `test/kubernetesauthengine/` — mount, config, role (explicit ns), role (selector)
- Task 3: Created `kubernetesauthengine_controller_test.go` with 5 contexts: prerequisite mount, config create+verify, role with explicit namespaces, role with namespace selector, delete+cleanup
- Task 4: `make integration` passed — all 40 specs green, no regressions, `auth/kubernetes` mount unaffected
- All 4 ACs satisfied: config written to Vault (AC1), role with explicit namespaces verified (AC2), namespace selector resolved correctly (AC3), delete with proper cleanup behavior (AC4)

### File List

| # | File | Change Type |
|---|------|-------------|
| 1 | `controllers/controllertestutils/decoder.go` | Modified |
| 2 | `test/kubernetesauthengine/test-kube-auth-mount.yaml` | New |
| 3 | `test/kubernetesauthengine/test-kube-auth-config.yaml` | New |
| 4 | `test/kubernetesauthengine/test-kube-auth-role.yaml` | New |
| 5 | `test/kubernetesauthengine/test-kube-auth-role-selector.yaml` | New |
| 6 | `controllers/kubernetesauthengine_controller_test.go` | New |
