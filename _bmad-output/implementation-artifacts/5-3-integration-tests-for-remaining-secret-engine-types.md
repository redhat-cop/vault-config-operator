# Story 5.3: Integration Tests for Remaining Secret Engine Types

Status: ready-for-dev

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As an operator developer,
I want integration tests for KubernetesSecretEngineConfig and KubernetesSecretEngineRole covering create, reconcile success, Vault state verification, update, and delete,
So that the Kubernetes secret engine lifecycle — with its JWT credential resolution, `service_account_jwt` stripping in `IsEquivalentToDesiredState`, and IsDeletable=true for both types — is verified end-to-end using the Kind cluster's own Kubernetes API as the target.

## Scope Decision: Tier Classification

Per the Epic 4 retrospective's Story 5.3 infrastructure classification and the project's three-tier integration test rule:

| Type | Dependency | Classification | Action |
|------|-----------|---------------|--------|
| KubernetesSecretEngineConfig/Role | Kubernetes API | **Tier 1: Already available** | **Test in Kind** — Kind cluster IS the Kubernetes instance |
| GitHubSecretEngineConfig/Role | GitHub App + custom Vault plugin | **Tier 3: Skip** | Unit test coverage only (no GitHub App in test env) |
| AzureSecretEngineConfig/Role | Azure cloud | **Tier 3: Skip** | Unit test coverage only (cloud provider) |
| QuaySecretEngineConfig/Role/StaticRole | Quay + custom Vault plugin | **Tier 3: Skip** | Unit test coverage only (heavy stack + plugin not in test env) |

This story implements integration tests ONLY for KubernetesSecretEngineConfig and KubernetesSecretEngineRole. The other types cannot be integration-tested per the three-tier rule. Azure controllers are not even registered in `suite_integration_test.go`.

[Source: _bmad-output/implementation-artifacts/epic-4-retro-2026-04-23.md — Story 5.3 Type Classification]
[Source: _bmad-output/project-context.md#L134-L141 — Integration test infrastructure philosophy]

## Acceptance Criteria

1. **Given** a ServiceAccount with cluster-admin permissions exists in the test namespace **And** a K8s Secret of type `kubernetes.io/service-account-token` has been created for that SA **And** a SecretEngineMount (type=kubernetes) has been created and reconciled **When** a KubernetesSecretEngineConfig CR is created with `kubernetesHost` pointing at the Kind cluster API and `jwtReference` pointing at the SA token secret **Then** the config exists in Vault at `{mount-path}/config` with `kubernetes_host` verified, `service_account_jwt` NOT present in Vault read response, and ReconcileSuccessful=True

2. **Given** a KubernetesSecretEngineRole CR is created with `allowedKubernetesNamespaces`, `kubernetesRoleName`, and `kubernetesRoleType` **When** the reconciler processes it **Then** the role exists in Vault at `{mount-path}/roles/{name}` with correct field values and ReconcileSuccessful=True

3. **Given** the KubernetesSecretEngineRole CR spec is updated (e.g., `kubernetesRoleName` changed) **When** the reconciler processes the update **Then** the Vault role reflects the updated value and `ObservedGeneration` increases

4. **Given** the KubernetesSecretEngineRole CR is deleted (IsDeletable=true) **When** the reconciler processes the deletion **Then** the role is removed from Vault and the CR is deleted from K8s

5. **Given** the KubernetesSecretEngineConfig CR is deleted (IsDeletable=true) **When** the reconciler processes the deletion **Then** the config is removed from Vault and the CR is deleted from K8s

## Tasks / Subtasks

- [ ] Task 1: Create ServiceAccount infrastructure for JWT credential resolution (AC: 1)
  - [ ] 1.1: Create `test/kubernetessecretengine/test-kubese-sa-rbac.yaml` — ServiceAccount + ClusterRoleBinding (cluster-admin) in vault-admin namespace
  - [ ] 1.2: The SA token K8s Secret is created programmatically in the test (type `kubernetes.io/service-account-token` with annotation `kubernetes.io/service-account.name`)

- [ ] Task 2: Add decoder methods (AC: 1, 2)
  - [ ] 2.1: Add `GetKubernetesSecretEngineConfigInstance` to `controllers/controllertestutils/decoder.go`
  - [ ] 2.2: Add `GetKubernetesSecretEngineRoleInstance` to `controllers/controllertestutils/decoder.go`

- [ ] Task 3: Create test fixtures (AC: 1, 2)
  - [ ] 3.1: Create `test/kubernetessecretengine/test-kubese-mount.yaml` — SecretEngineMount with `type: kubernetes`, unique path prefix
  - [ ] 3.2: Create `test/kubernetessecretengine/test-kubese-config.yaml` — KubernetesSecretEngineConfig pointing at Kind cluster API with `jwtReference.secret` referencing the SA token secret
  - [ ] 3.3: Create `test/kubernetessecretengine/test-kubese-role.yaml` — KubernetesSecretEngineRole with `allowedKubernetesNamespaces`, `kubernetesRoleName`, `kubernetesRoleType`

- [ ] Task 4: Create integration test file (AC: 1, 2, 3, 4, 5)
  - [ ] 4.1: Create `controllers/kubernetessecretengine_controller_test.go` with `//go:build integration` tag
  - [ ] 4.2: Add prerequisite context — apply SA RBAC manifest, create SA token K8s Secret, wait for token population, create SecretEngineMount (type=kubernetes), wait for reconcile, verify `sys/mounts`
  - [ ] 4.3: Add context for KubernetesSecretEngineConfig — create, poll for ReconcileSuccessful=True, verify Vault state at `{mount}/config` including `kubernetes_host`
  - [ ] 4.4: Add context for KubernetesSecretEngineRole — create, poll for ReconcileSuccessful=True, verify Vault state at `{mount}/roles/{name}`
  - [ ] 4.5: Add update context for KubernetesSecretEngineRole — update `kubernetesRoleName`, verify Vault reflects change, verify ObservedGeneration increased
  - [ ] 4.6: Add deletion context — delete role (IsDeletable=true, verify Vault cleanup), delete config (IsDeletable=true, verify Vault cleanup), delete mount, delete SA token secret, delete SA RBAC

- [ ] Task 5: End-to-end verification (AC: 1, 2, 3, 4, 5)
  - [ ] 5.1: Run `make integration` and verify new tests pass alongside all existing tests
  - [ ] 5.2: Verify no regressions — existing tests unaffected

## Dev Notes

### Infrastructure Scope — No New Infrastructure Required

The Kubernetes secret engine needs a Kubernetes API endpoint. The Kind cluster running the integration tests IS that Kubernetes instance. No new Helm charts, Makefile targets, or external services needed.

The Kubernetes API is reachable from within the cluster at `https://kubernetes.default.svc:443`.

[Source: _bmad-output/implementation-artifacts/epic-4-retro-2026-04-23.md — Story 5.3 Kubernetes = "Tier 1: Already available"]

### KubernetesSecretEngineConfig — VaultResource Reconciler

Uses `NewVaultResource` — standard reconcile flow (read → compare → write if different).

**GetPath():**
```go
func (d *KubernetesSecretEngineConfig) GetPath() string {
    return string(d.Spec.Path) + "/" + "config"
}
```

For fixture with `path: test-kubese/test-kubese-mount` → Vault path is `test-kubese/test-kubese-mount/config`

This is a FIXED path (no name appended) — one config per mount, like RabbitMQSecretEngineConfig.

[Source: api/v1alpha1/kubernetessecretengineconfig_types.go#L113-L115]

### Config IsDeletable = true — Verify Vault Cleanup After CR Deletion

Both KubernetesSecretEngineConfig AND KubernetesSecretEngineRole have `IsDeletable() == true`. This means:
- Finalizer is added by ManageOutcome
- Vault resource is deleted on CR deletion
- Delete test must verify BOTH K8s NotFound AND Vault Read returns nil

This differs from auth engine configs (which were IsDeletable=false) and RabbitMQSecretEngineConfig (IsDeletable=false). Both types here follow the same delete pattern as DatabaseSecretEngineConfig (Story 5.1).

[Source: api/v1alpha1/kubernetessecretengineconfig_types.go#L109-L111 — IsDeletable returns true]
[Source: api/v1alpha1/kubernetessecretenginerole_types.go#L64-L66 — IsDeletable returns true]

### Config IsEquivalentToDesiredState — Custom: Strips `service_account_jwt`

```go
func (d *KubernetesSecretEngineConfig) IsEquivalentToDesiredState(payload map[string]interface{}) bool {
    desiredState := d.Spec.KubeSEConfig.toMap()
    delete(desiredState, "service_account_jwt")
    return reflect.DeepEqual(desiredState, payload)
}
```

This is the same pattern as GitHubSecretEngineConfig (strips `prv_key`) — Vault never returns the JWT in read responses. The comparison is done after removing the secret field from the desired state.

[Source: api/v1alpha1/kubernetessecretengineconfig_types.go#L119-L123]

### Config toMap — 4 Fields

```go
func (i *KubeSEConfig) toMap() map[string]interface{} {
    payload := map[string]interface{}{}
    payload["kubernetes_host"] = i.KubernetesHost
    payload["kubernetes_ca_cert"] = i.KubernetesCACert
    payload["service_account_jwt"] = i.retrievedServiceAccountJWT
    payload["disable_local_ca_jwt"] = i.DisableLocalCAJWT
    return payload
}
```

[Source: api/v1alpha1/kubernetessecretengineconfig_types.go#L197-L204]

### Config Credential Resolution (PrepareInternalValues → setInternalCredentials)

The `setInternalCredentials` method supports TWO JWT sources:

1. **K8s Secret** (used in test): Must be of type `kubernetes.io/service-account-token`. Reads `secret.Data["token"]` (the `corev1.ServiceAccountTokenKey`).
2. **VaultSecret**: Reads from Vault KV path, extracts `secret.Data["key"]`.

**CRITICAL**: The K8s Secret MUST have `Type: corev1.SecretTypeServiceAccountToken`. The code explicitly rejects other secret types:
```go
if secret.Type != corev1.SecretTypeServiceAccountToken {
    err := errors.New("secret must be of type: " + string(corev1.SecretTypeServiceAccountToken))
    return err
}
```

[Source: api/v1alpha1/kubernetessecretengineconfig_types.go#L146-L178]

### Config Webhook — Standard Kubebuilder

- **ValidateCreate**: Calls `isValid()` → `JWTReference.ValidateEitherFromVaultSecretOrFromSecret()` (must have exactly one JWT source)
- **ValidateUpdate**: Rejects changes to `spec.path` (immutable), then calls `isValid()`
- **ValidateDelete**: No-op
- **Default**: Log-only

Webhook validation verbs include `update` only (not `create;update`), but `ValidateCreate` still has `isValid()` logic. The kubebuilder marker is:
```
verbs=update
```
But the `ValidateCreate` function body does call `isValid()`. This means webhook validation on create may NOT be active (verbs only lists update). The test should NOT rely on webhook rejection on create.

[Source: api/v1alpha1/kubernetessecretengineconfig_webhook.go]

### ServiceAccount Token Secret — Created Programmatically

In Kubernetes 1.24+, service account token secrets are NOT auto-created. The test must create one manually:

```go
saTokenSecret := &corev1.Secret{
    ObjectMeta: metav1.ObjectMeta{
        Name:      "test-kubese-sa-token",
        Namespace: vaultAdminNamespaceName,
        Annotations: map[string]string{
            corev1.ServiceAccountNameAnnotation: "test-kubese-sa",
        },
    },
    Type: corev1.SecretTypeServiceAccountToken,
}
```

After creating this secret, the K8s token controller will populate `Data["token"]`, `Data["ca.crt"]`, and `Data["namespace"]`. The test MUST wait for the `token` field to be populated before creating the KubernetesSecretEngineConfig CR:

```go
Eventually(func() bool {
    err := k8sIntegrationClient.Get(ctx, types.NamespacedName{Name: "test-kubese-sa-token", Namespace: vaultAdminNamespaceName}, saTokenSecret)
    if err != nil {
        return false
    }
    _, hasToken := saTokenSecret.Data[corev1.ServiceAccountTokenKey]
    return hasToken
}, timeout, interval).Should(BeTrue())
```

### ServiceAccount RBAC — Needs Cluster-Admin

The ServiceAccount whose JWT is used by Vault to interact with the K8s API needs broad permissions. Vault will create service accounts, role bindings, etc. The test should apply a YAML manifest:

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: test-kubese-sa
  namespace: vault-admin
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: test-kubese-sa-cluster-admin
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cluster-admin
subjects:
- kind: ServiceAccount
  name: test-kubese-sa
  namespace: vault-admin
```

This mirrors what the manual readme instructs: `oc adm policy add-cluster-role-to-user cluster-admin -z default -n vault-admin`

**IMPORTANT**: The ClusterRoleBinding is cluster-scoped. The AfterAll cleanup should delete both the SA and the ClusterRoleBinding.

[Source: test/kubernetessecretengine/readme.md — Manual setup with cluster-admin]

### KubernetesSecretEngineRole — Standard VaultResource Reconciler

Uses `NewVaultResource` — simple standard reconcile flow. No extra watches, no credential resolution.

**GetPath():**
```go
func (d *KubernetesSecretEngineRole) GetPath() string {
    if d.Spec.Name != "" {
        return vaultutils.CleansePath(string(d.Spec.Path) + "/" + "roles" + "/" + d.Spec.Name)
    }
    return vaultutils.CleansePath(string(d.Spec.Path) + "/" + "roles" + "/" + d.Name)
}
```

For fixture with `path: test-kubese/test-kubese-mount`, `metadata.name: test-kubese-role` → `test-kubese/test-kubese-mount/roles/test-kubese-role`

[Source: api/v1alpha1/kubernetessecretenginerole_types.go#L68-L73]
[Source: controllers/kubernetessecretenginerole_controller.go#L69 — NewVaultResource]

### Role toMap — 11 Fields

```go
func (i *KubeSERole) toMap() map[string]interface{} {
    payload := map[string]interface{}{}
    payload["allowed_kubernetes_namespaces"] = i.AllowedKubernetesNamespaces
    payload["allowed_kubernetes_namespace_selector"] = i.AllowedKubernetesNamespaceSelector
    payload["token_max_ttl"] = i.MaxTTL
    payload["token_default_ttl"] = i.DefaultTTL
    payload["token_default_audiences"] = i.DefaultAudiences
    payload["service_account_name"] = i.ServiceAccountName
    payload["kubernetes_role_name"] = i.KubernetesRoleName
    payload["kubernetes_role_type"] = i.KubernetesRoleType
    payload["generated_role_rules"] = i.GenerateRoleRules
    payload["name_template"] = i.NameTemplate
    payload["extra_annotations"] = i.ExtraAnnotations
    payload["extra_labels"] = i.ExtraLabels
    return payload
}
```

**NOTE**: `TargetNamespaces` (from spec) is NOT part of the Vault payload — it's a Kubernetes-side concept used by the operator, not sent to Vault.

**NOTE**: `toMap()` sends `token_max_ttl` and `token_default_ttl` as `metav1.Duration` values, not strings or ints. Vault may return these as `json.Number` (seconds).

[Source: api/v1alpha1/kubernetessecretenginerole_types.go#L158-L173]

### Role IsEquivalentToDesiredState — Bare DeepEqual

```go
func (d *KubernetesSecretEngineRole) IsEquivalentToDesiredState(payload map[string]interface{}) bool {
    desiredState := d.Spec.KubeSERole.toMap()
    return reflect.DeepEqual(desiredState, payload)
}
```

No filtering of extra keys. Vault may return extra fields → potential write on every reconcile (known tech debt — Story 7-4). Does NOT affect ReconcileSuccessful=True or test correctness.

[Source: api/v1alpha1/kubernetessecretenginerole_types.go#L77-L80]

### Role Webhook — ValidateUpdate Only

```
verbs=update
```

Only immutable `spec.path` check on update. No validation on create or delete.

[Source: api/v1alpha1/kubernetessecretenginerole_webhook.go]

### Vault API Response Shapes

**GET `{mount}/config`** — Returns Kubernetes secret engine config:
```json
{
  "data": {
    "kubernetes_host": "https://kubernetes.default.svc:443",
    "kubernetes_ca_cert": "...",
    "disable_local_ca_jwt": false
  }
}
```
Key: `service_account_jwt` is NEVER returned by Vault. Other fields like `kubernetes_host` and `disable_local_ca_jwt` are returned.

**GET `{mount}/roles/{name}`** — Returns role config:
```json
{
  "data": {
    "allowed_kubernetes_namespaces": ["default"],
    "allowed_kubernetes_namespace_selector": "",
    "token_max_ttl": 0,
    "token_default_ttl": 0,
    "token_default_audiences": "",
    "service_account_name": "",
    "kubernetes_role_name": "edit",
    "kubernetes_role_type": "ClusterRole",
    "generated_role_rules": "",
    "name_template": "",
    "extra_annotations": null,
    "extra_labels": null
  }
}
```
Key: TTL values returned as `json.Number`. `allowed_kubernetes_namespaces` returned as `[]interface{}` not `[]string`. Extra fields may appear depending on Vault version.

### Verifying Vault State

**Config verification:**
```go
secret, err := vaultClient.Logical().Read("test-kubese/test-kubese-mount/config")
Expect(err).To(BeNil())
Expect(secret).NotTo(BeNil())

kubeHost, ok := secret.Data["kubernetes_host"].(string)
Expect(ok).To(BeTrue(), "expected kubernetes_host to be a string")
Expect(kubeHost).To(Equal("https://kubernetes.default.svc:443"))

disableLocalCA, ok := secret.Data["disable_local_ca_jwt"].(bool)
Expect(ok).To(BeTrue(), "expected disable_local_ca_jwt to be a bool")
Expect(disableLocalCA).To(BeFalse())
```

**Role verification:**
```go
secret, err := vaultClient.Logical().Read("test-kubese/test-kubese-mount/roles/test-kubese-role")
Expect(err).To(BeNil())
Expect(secret).NotTo(BeNil())

roleName, ok := secret.Data["kubernetes_role_name"].(string)
Expect(ok).To(BeTrue(), "expected kubernetes_role_name to be a string")
Expect(roleName).To(Equal("edit"))

roleType, ok := secret.Data["kubernetes_role_type"].(string)
Expect(ok).To(BeTrue(), "expected kubernetes_role_type to be a string")
Expect(roleType).To(Equal("ClusterRole"))

allowedNs, ok := secret.Data["allowed_kubernetes_namespaces"].([]interface{})
Expect(ok).To(BeTrue(), "expected allowed_kubernetes_namespaces to be []interface{}")
Expect(allowedNs).To(ContainElement("default"))
```

**Delete verification (both IsDeletable=true):**
```go
// Role — IsDeletable=true: verify Vault cleanup
Expect(k8sIntegrationClient.Delete(ctx, roleInstance)).Should(Succeed())
roleLookupKey := types.NamespacedName{Name: roleInstance.Name, Namespace: roleInstance.Namespace}
Eventually(func() bool {
    err := k8sIntegrationClient.Get(ctx, roleLookupKey, &redhatcopv1alpha1.KubernetesSecretEngineRole{})
    return apierrors.IsNotFound(err)
}, timeout, interval).Should(BeTrue())

Eventually(func() bool {
    secret, err := vaultClient.Logical().Read("test-kubese/test-kubese-mount/roles/test-kubese-role")
    return err == nil && secret == nil
}, timeout, interval).Should(BeTrue())

// Config — IsDeletable=true: verify Vault cleanup
Expect(k8sIntegrationClient.Delete(ctx, configInstance)).Should(Succeed())
configLookupKey := types.NamespacedName{Name: configInstance.Name, Namespace: configInstance.Namespace}
Eventually(func() bool {
    err := k8sIntegrationClient.Get(ctx, configLookupKey, &redhatcopv1alpha1.KubernetesSecretEngineConfig{})
    return apierrors.IsNotFound(err)
}, timeout, interval).Should(BeTrue())

Eventually(func() bool {
    secret, err := vaultClient.Logical().Read("test-kubese/test-kubese-mount/config")
    return err == nil && secret == nil
}, timeout, interval).Should(BeTrue())
```

### Test Design — Dependency Chain

```
ServiceAccount (test-kubese-sa) + ClusterRoleBinding (cluster-admin)
  └── K8s Secret (test-kubese-sa-token, type=kubernetes.io/service-account-token)
        └── SecretEngineMount (test-kubese-mount, type=kubernetes, path=test-kubese)
              └── KubernetesSecretEngineConfig (test-kubese-config)
                    → test-kubese/test-kubese-mount/config
              └── KubernetesSecretEngineRole (test-kubese-role)
                    → test-kubese/test-kubese-mount/roles/test-kubese-role
```

Resources must be created in order: SA + RBAC → Token Secret (wait for population) → Mount → Config → Role.
Deletion in reverse: Role → Config → Mount → Token Secret → SA + ClusterRoleBinding.

### Test Fixture Design

**Fixture 1: `test/kubernetessecretengine/test-kubese-sa-rbac.yaml`** — SA + ClusterRoleBinding:
```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: test-kubese-sa
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: test-kubese-sa-cluster-admin
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: cluster-admin
subjects:
- kind: ServiceAccount
  name: test-kubese-sa
  namespace: vault-admin
```
Namespace is set by the test to `vault-admin`. The ClusterRoleBinding is cluster-scoped.

**Fixture 2: `test/kubernetessecretengine/test-kubese-mount.yaml`** — SecretEngineMount prerequisite:
```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: SecretEngineMount
metadata:
  name: test-kubese-mount
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  type: kubernetes
  path: test-kubese
```
Mounts at `sys/mounts/test-kubese/test-kubese-mount`. Uses `type: kubernetes` to enable the Kubernetes secret engine.

**Fixture 3: `test/kubernetessecretengine/test-kubese-config.yaml`** — KubernetesSecretEngineConfig:
```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: KubernetesSecretEngineConfig
metadata:
  name: test-kubese-config
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  path: test-kubese/test-kubese-mount
  kubernetesHost: "https://kubernetes.default.svc:443"
  jwtReference:
    secret:
      name: test-kubese-sa-token
```
`GetPath()` returns `test-kubese/test-kubese-mount/config`.

Key: `disableLocalCAJWT` defaults to `false` (kubebuilder default). This means Vault CAN fall back to its own JWT, but the test provides an explicit one via `jwtReference.secret`.

Key: `kubernetesHost` is the in-cluster Kubernetes API URL.

Key: `kubernetesCACert` is NOT set — Vault will use its own CA when `disableLocalCAJWT` is false.

**Fixture 4: `test/kubernetessecretengine/test-kubese-role.yaml`** — KubernetesSecretEngineRole:
```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: KubernetesSecretEngineRole
metadata:
  name: test-kubese-role
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  path: test-kubese/test-kubese-mount
  allowedKubernetesNamespaces:
  - default
  kubernetesRoleName: "edit"
  kubernetesRoleType: "ClusterRole"
```
`GetPath()` returns `test-kubese/test-kubese-mount/roles/test-kubese-role`.

`allowedKubernetesNamespaces: ["default"]` limits credential generation to the `default` namespace.
`kubernetesRoleName: "edit"` with `kubernetesRoleType: "ClusterRole"` means Vault will bind the generated SA to the built-in `edit` ClusterRole.

### Test Structure

```
Describe("KubernetesSecretEngine controllers", Ordered)
  var saTokenSecret *corev1.Secret
  var mountInstance *redhatcopv1alpha1.SecretEngineMount
  var configInstance *redhatcopv1alpha1.KubernetesSecretEngineConfig
  var roleInstance *redhatcopv1alpha1.KubernetesSecretEngineRole

  AfterAll: best-effort delete all instances (reverse order):
    role → config → mount → sa token secret → SA → ClusterRoleBinding

  Context("When creating prerequisite resources")
    It("Should create the SA, RBAC, token secret, and kubernetes engine mount")
      - Apply SA and ClusterRoleBinding (from YAML or programmatically)
      - Create SA token K8s Secret (type=kubernetes.io/service-account-token, annotation referencing SA)
      - Eventually wait for token population (Data["token"] exists)
      - Load test-kubese-mount.yaml via decoder.GetSecretEngineMountInstance
      - Set namespace to vaultAdminNamespaceName, create
      - Eventually poll for ReconcileSuccessful=True
      - Verify mount exists via sys/mounts key "test-kubese/test-kubese-mount/"

  Context("When creating a KubernetesSecretEngineConfig")
    It("Should write the Kubernetes config to Vault")
      - Load test-kubese-config.yaml via decoder.GetKubernetesSecretEngineConfigInstance
      - Set namespace to vaultAdminNamespaceName, create
      - Eventually poll for ReconcileSuccessful=True
      - Read test-kubese/test-kubese-mount/config from Vault
      - Verify kubernetes_host = "https://kubernetes.default.svc:443"
      - Verify disable_local_ca_jwt = false

  Context("When creating a KubernetesSecretEngineRole")
    It("Should create the role in Vault with correct settings")
      - Load test-kubese-role.yaml via decoder.GetKubernetesSecretEngineRoleInstance
      - Set namespace to vaultAdminNamespaceName, create
      - Eventually poll for ReconcileSuccessful=True
      - Read test-kubese/test-kubese-mount/roles/test-kubese-role
      - Verify kubernetes_role_name = "edit"
      - Verify kubernetes_role_type = "ClusterRole"
      - Verify allowed_kubernetes_namespaces contains "default"

  Context("When updating a KubernetesSecretEngineRole")
    It("Should update the role in Vault and reflect updated ObservedGeneration")
      - Record initial ObservedGeneration
      - Get latest role CR, update kubernetesRoleName to "view"
      - Eventually verify Vault reflects updated kubernetes_role_name
      - Verify ObservedGeneration increased

  Context("When deleting KubernetesSecretEngine resources")
    It("Should clean up role and config from Vault and remove all K8s resources")
      - Delete role CR (IsDeletable=true → Vault cleanup)
      - Eventually verify K8s deletion (NotFound)
      - Eventually verify role removed from Vault (Read returns nil)
      - Delete config CR (IsDeletable=true → Vault cleanup)
      - Eventually verify K8s deletion (NotFound)
      - Eventually verify config removed from Vault (Read returns nil)
      - Delete SecretEngineMount
      - Eventually verify K8s deletion and mount gone from sys/mounts
      - Delete SA token secret
      - Delete ServiceAccount and ClusterRoleBinding
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
    rbacv1 "k8s.io/api/rbac/v1"
    apierrors "k8s.io/apimachinery/pkg/api/errors"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/apimachinery/pkg/types"
)
```

`rbacv1` is needed for programmatic ClusterRoleBinding creation/deletion if not using YAML fixture.

### Name Collision Prevention

Fixture names use `test-kubese-` prefix with a unique mount path `test-kubese`:
- `test-kubese/test-kubese-mount` — secret engine mount (unique path prefix)
- `test-kubese-config` — KubernetesSecretEngineConfig CR name
- `test-kubese-role` — KubernetesSecretEngineRole CR name and Vault role name
- `test-kubese-sa` — ServiceAccount name
- `test-kubese-sa-token` — SA token K8s Secret
- `test-kubese-sa-cluster-admin` — ClusterRoleBinding name

These don't collide with:
- Existing fixtures at `test/kubernetessecretengine/` (use `kubese-test` names and empty path)
- Story 5.1 resources (`test-dbse/*`, `test-db-*`)
- Story 5.2 resources (`test-rmqse/*`, `test-rmq-*`)
- Epic 4 test resources (`test-k8s-auth/*`, `test-ldap-auth/*`, `test-jwt-oidc-auth/*`)

### Existing Test Coexistence

The new test uses completely separate resources at a different mount path (`test-kubese`) in `vault-admin` namespace. The existing `test/kubernetessecretengine/kubese-*.yaml` fixtures use `kubese-test` as the path and metadata name. Since Ginkgo v2 randomizes top-level Describe blocks, both test files run independently.

### Controller Registration — Already Done

Both controllers are registered in `suite_integration_test.go`:
```go
err = (&KubernetesSecretEngineConfigReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "KubernetesSecretEngineConfig")}).SetupWithManager(mgr)
Expect(err).ToNot(HaveOccurred())

err = (&KubernetesSecretEngineRoleReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "KubernetesSecretEngineRole")}).SetupWithManager(mgr)
```

No changes needed to the test suite setup.

[Source: controllers/suite_integration_test.go#L199-L203]

### Decoder Methods — BOTH Must Be Added

Neither `GetKubernetesSecretEngineConfigInstance` nor `GetKubernetesSecretEngineRoleInstance` exist in the decoder:

```go
func (d *decoder) GetKubernetesSecretEngineConfigInstance(filename string) (*redhatcopv1alpha1.KubernetesSecretEngineConfig, error) {
    obj, groupKindVersion, err := d.decodeFile(filename)
    if err != nil {
        return nil, err
    }
    kind := reflect.TypeOf(redhatcopv1alpha1.KubernetesSecretEngineConfig{}).Name()
    if groupKindVersion.Kind == kind {
        o := obj.(*redhatcopv1alpha1.KubernetesSecretEngineConfig)
        return o, nil
    }
    return nil, errDecode
}

func (d *decoder) GetKubernetesSecretEngineRoleInstance(filename string) (*redhatcopv1alpha1.KubernetesSecretEngineRole, error) {
    obj, groupKindVersion, err := d.decodeFile(filename)
    if err != nil {
        return nil, err
    }
    kind := reflect.TypeOf(redhatcopv1alpha1.KubernetesSecretEngineRole{}).Name()
    if groupKindVersion.Kind == kind {
        o := obj.(*redhatcopv1alpha1.KubernetesSecretEngineRole)
        return o, nil
    }
    return nil, errDecode
}
```

[Source: controllers/controllertestutils/decoder.go — existing pattern]

### SA Token Population Timing

The K8s token controller populates `Data["token"]` asynchronously after the secret is created with `Type: kubernetes.io/service-account-token` and the `kubernetes.io/service-account.name` annotation. In Kind, this typically happens within seconds. The `Eventually` with 120s timeout provides ample buffer.

If the SA doesn't exist when the secret is created, the token controller will NOT populate the token until the SA exists. Create the SA BEFORE the token secret.

### Vault Kubernetes Secret Engine Config — `service_account_jwt` NOT in Read Response

When reading `{mount}/config`, Vault does NOT return `service_account_jwt` in the response (it's a secret). This is why `IsEquivalentToDesiredState` deletes it from `desiredState` before comparison. The test should verify that `kubernetes_host` and `disable_local_ca_jwt` are present but should NOT check for `service_account_jwt`.

### Vault Role TTL Format

The role fixture uses `kubernetesRoleType` and `kubernetesRoleName` (string fields). TTL fields (`defaultTTL`, `maxTTL`) default to `0s`. Vault returns these as `json.Number` (value `0`). The test does NOT need to verify TTLs since they use defaults, but if checking, use the `json.Number` pattern from Story 5.1.

### `policy-admin` Permissions

The test uses `policy-admin` auth role in `vault-admin` namespace — the standard broad-permissions role used by all integration tests. It has permissions for `sys/mounts/*` and full access to engine paths.

### Checked Type Assertions

Per Epic 3 retro action item and Epic 4 practice: always use two-value form `val, ok := x.(string)` with `Expect(ok).To(BeTrue())` for all Vault response field assertions.

### No `make manifests generate` Needed

This story adds an integration test file, YAML fixtures, and decoder methods. No CRD types, controllers, or webhooks are changed. No Makefile changes needed.

### File Inventory — What Needs to Change

| # | File | Change Type | Description |
|---|------|-------------|-------------|
| 1 | `controllers/controllertestutils/decoder.go` | Modified | Add `GetKubernetesSecretEngineConfigInstance` and `GetKubernetesSecretEngineRoleInstance` |
| 2 | `test/kubernetessecretengine/test-kubese-sa-rbac.yaml` | New | ServiceAccount + ClusterRoleBinding for JWT credential |
| 3 | `test/kubernetessecretengine/test-kubese-mount.yaml` | New | SecretEngineMount prerequisite (type=kubernetes) |
| 4 | `test/kubernetessecretengine/test-kubese-config.yaml` | New | KubernetesSecretEngineConfig pointing at Kind cluster API |
| 5 | `test/kubernetessecretengine/test-kubese-role.yaml` | New | KubernetesSecretEngineRole with ClusterRole binding |
| 6 | `controllers/kubernetessecretengine_controller_test.go` | New | Integration test — SA setup, create mount, config, role; verify Vault state; update role; delete and verify |

No changes to suite setup, controllers, webhooks, types, Makefile, or other infrastructure manifests.

### Previous Story Intelligence

**From Story 5.2 (RabbitMQ secret engine integration tests — immediate predecessor):**
- Established infrastructure deployment pattern with Helm (RabbitMQ via Bitnami)
- Demonstrated IsDeletable=false config persistence verification pattern
- Story 5.3's config is IsDeletable=true, so deletion test should verify Vault CLEANUP (not persistence)
- Used `json.Number` pattern for TTL assertions

**From Story 5.1 (DatabaseSecretEngine integration tests):**
- Established the full Epic 5 integration test pattern for secret engines
- Demonstrated K8s Secret created programmatically in the test
- Both types were IsDeletable=true — SAME pattern applies here for both types
- Showed `connection_details` nesting verification — not applicable here (Kubernetes engine has flat config)

**From Story 4.1 (KubernetesAuthEngine integration tests):**
- Established KubernetesAuth integration test pattern — the auth ENGINE version of what this story tests on the SECRET ENGINE side
- Different types but similar Kubernetes API interaction concepts
- AfterAll cleanup guard pattern

**From Epic 4 Retrospective:**
- Story 5.3 classified: Kubernetes = Tier 1 (test in Kind); GitHub/Azure/Quay = Tier 3 (skip)
- Story ordering: 5.1 → 5.2 → 5.3 (Kubernetes secret engine from Kind + skip cloud/Quay types)
- Continue using Opus-class models for integration test stories

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
- **Kubernetes API:** Already available in Kind → **Tier 1: No new infrastructure**
- **Vault API:** Already available in Kind
- **K8s Secrets/RBAC:** Available via integration test client
- **GitHub/Azure/Quay:** Cloud services or heavy plugins → **Tier 3: Skip**

**Classification: No new infrastructure — Lowest scope story in Epic 5 (final)**

[Source: _bmad-output/project-context.md#L134-L141 — Integration test infrastructure philosophy]

### Project Structure Notes

- Decoder changes in `controllers/controllertestutils/decoder.go` (add two methods)
- Test file goes in `controllers/kubernetessecretengine_controller_test.go`
- Test fixtures go in `test/kubernetessecretengine/` directory (alongside existing fixtures, with `test-kubese-` prefix)
- No Makefile changes needed
- No new infrastructure directories
- All files follow existing naming conventions

### References

- [Source: api/v1alpha1/kubernetessecretengineconfig_types.go] — VaultObject implementation, GetPath ({spec.path}/config), GetPayload, IsEquivalentToDesiredState (strips service_account_jwt), toMap (4 keys), PrepareInternalValues (JWT from K8s Secret or VaultSecret), IsDeletable=true
- [Source: api/v1alpha1/kubernetessecretengineconfig_types.go#L113-L115] — GetPath: {spec.path}/config (fixed, no name)
- [Source: api/v1alpha1/kubernetessecretengineconfig_types.go#L119-L123] — IsEquivalentToDesiredState: strips service_account_jwt then DeepEqual
- [Source: api/v1alpha1/kubernetessecretengineconfig_types.go#L146-L178] — setInternalCredentials (K8s Secret type=SA token or VaultSecret)
- [Source: api/v1alpha1/kubernetessecretengineconfig_types.go#L197-L204] — KubeSEConfig.toMap (kubernetes_host, kubernetes_ca_cert, service_account_jwt, disable_local_ca_jwt)
- [Source: api/v1alpha1/kubernetessecretengineconfig_types.go#L109-L111] — IsDeletable returns true
- [Source: api/v1alpha1/kubernetessecretengineconfig_webhook.go] — Standard webhook: isValid on create/update, immutable path on update
- [Source: api/v1alpha1/kubernetessecretenginerole_types.go] — VaultObject implementation, GetPath ({path}/roles/{name}), toMap (11 keys), IsDeletable=true
- [Source: api/v1alpha1/kubernetessecretenginerole_types.go#L68-L73] — GetPath: {spec.path}/roles/{name or metadata.name}
- [Source: api/v1alpha1/kubernetessecretenginerole_types.go#L77-L80] — IsEquivalentToDesiredState: bare DeepEqual, no filtering
- [Source: api/v1alpha1/kubernetessecretenginerole_types.go#L158-L173] — KubeSERole.toMap (11 Vault keys)
- [Source: api/v1alpha1/kubernetessecretenginerole_webhook.go] — Standard webhook: immutable path on update only
- [Source: controllers/kubernetessecretengineconfig_controller.go#L77] — VaultResource reconciler, watches SA token Secrets
- [Source: controllers/kubernetessecretenginerole_controller.go#L69] — Standard VaultResource reconciler
- [Source: controllers/suite_integration_test.go#L199-L203] — Both controllers registered
- [Source: controllers/controllertestutils/decoder.go] — Neither GetKubernetesSecretEngineConfigInstance nor GetKubernetesSecretEngineRoleInstance exist
- [Source: test/kubernetessecretengine/kubese-config.yaml] — Existing fixture (references legacy SA token, different path)
- [Source: test/kubernetessecretengine/kubese-mount.yaml] — Existing fixture (empty path, different names)
- [Source: test/kubernetessecretengine/kubese-role.yaml] — Existing fixture (different names and path)
- [Source: test/kubernetessecretengine/readme.md] — Manual setup instructions (cluster-admin for SA)
- [Source: controllers/jwtoidcauthengine_controller_test.go] — Most recent Epic 4 test pattern reference (Ordered, AfterAll, checked assertions)
- [Source: controllers/databasesecretengine_controller_test.go] — Story 5.1 pattern (IsDeletable=true delete verification)
- [Source: _bmad-output/implementation-artifacts/epic-4-retro-2026-04-23.md] — Epic 4 retrospective (Story 5.3 type classification, infrastructure assessment)
- [Source: _bmad-output/project-context.md#L134-L141] — Integration test infrastructure philosophy

## Dev Agent Record

### Agent Model Used

{{agent_model_name_version}}

### Debug Log References

### Completion Notes List

### File List
