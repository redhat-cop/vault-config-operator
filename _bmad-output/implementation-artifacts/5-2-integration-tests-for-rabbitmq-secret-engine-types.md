# Story 5.2: Integration Tests for RabbitMQ Secret Engine Types

Status: ready-for-dev

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As an operator developer,
I want integration tests for RabbitMQSecretEngineConfig and RabbitMQSecretEngineRole covering create, reconcile success, Vault state verification, update, and delete,
So that the RabbitMQ secret engine lifecycle — with its custom reconcile flow, dual Vault paths (connection + lease), and IsDeletable=false config — is verified end-to-end against a real RabbitMQ instance in Kind.

## Acceptance Criteria

1. **Given** RabbitMQ is deployed in the Kind cluster via `deploy-rabbitmq` (Bitnami Helm) **And** a K8s Secret with RabbitMQ admin credentials exists **And** a SecretEngineMount (type=rabbitmq) has been created and reconciled **When** a RabbitMQSecretEngineConfig CR is created targeting the rabbitmq mount with `connectionURI` pointing at RabbitMQ **Then** the connection is configured in Vault at `{mount-path}/config/connection` with `connection_uri` and `username` verified, and ReconcileSuccessful=True

2. **Given** the RabbitMQSecretEngineConfig CR has `leaseTTL` and `leaseMaxTTL` set **When** the reconciler processes it **Then** the lease config exists in Vault at `{mount-path}/config/lease` with `ttl` and `max_ttl` values matching the spec

3. **Given** a RabbitMQSecretEngineRole CR is created with `tags`, `vhosts` permissions, and `path` referencing the rabbitmq mount **When** the reconciler processes it **Then** the role exists in Vault at `{mount-path}/roles/{name}` with correct field values and ReconcileSuccessful=True

4. **Given** the RabbitMQSecretEngineRole CR spec is updated (e.g., `tags` changed) **When** the reconciler processes the update **Then** the Vault role reflects the updated value and `ObservedGeneration` increases

5. **Given** the RabbitMQSecretEngineRole CR is deleted (IsDeletable=true) **When** the reconciler processes the deletion **Then** the role is removed from Vault and the CR is deleted from K8s

6. **Given** the RabbitMQSecretEngineConfig CR is deleted (IsDeletable=false) **When** the reconciler processes the deletion **Then** the CR is deleted from K8s **But** the Vault connection config persists (no Vault cleanup because IsDeletable=false)

## Tasks / Subtasks

- [ ] Task 1: Deploy RabbitMQ infrastructure in Kind (AC: 1)
  - [ ] 1.1: Create `integration/rabbitmq-values.yaml` — Bitnami RabbitMQ Helm values (management plugin enabled, admin credentials, service name override)
  - [ ] 1.2: Add `deploy-rabbitmq` Makefile target — Helm install to `test-vault-config-operator` namespace, wait for pod readiness
  - [ ] 1.3: Wire `deploy-rabbitmq` into the `integration` target dependency chain

- [ ] Task 2: Add decoder methods (AC: 1, 3)
  - [ ] 2.1: Add `GetRabbitMQSecretEngineConfigInstance` to `controllers/controllertestutils/decoder.go`
  - [ ] 2.2: Add `GetRabbitMQSecretEngineRoleInstance` to `controllers/controllertestutils/decoder.go`

- [ ] Task 3: Create test fixtures (AC: 1, 2, 3)
  - [ ] 3.1: Create `test/rabbitmqsecretengine/test-rmq-mount.yaml` — SecretEngineMount with `type: rabbitmq`, unique path prefix
  - [ ] 3.2: Create `test/rabbitmqsecretengine/test-rmq-config.yaml` — RabbitMQSecretEngineConfig with real RabbitMQ connection, `verifyConnection: true`, lease TTLs, K8s Secret credentials
  - [ ] 3.3: Create `test/rabbitmqsecretengine/test-rmq-role.yaml` — RabbitMQSecretEngineRole with tags and vhost permissions

- [ ] Task 4: Create integration test file (AC: 1, 2, 3, 4, 5, 6)
  - [ ] 4.1: Create `controllers/rabbitmqsecretengine_controller_test.go` with `//go:build integration` tag
  - [ ] 4.2: Add prerequisite context — create RabbitMQ admin credentials K8s Secret, create SecretEngineMount (type=rabbitmq), wait for reconcile, verify `sys/mounts`
  - [ ] 4.3: Add context for RabbitMQSecretEngineConfig — create, poll for ReconcileSuccessful=True, verify Vault state at both `{mount}/config/connection` and `{mount}/config/lease`
  - [ ] 4.4: Add context for RabbitMQSecretEngineRole — create, poll for ReconcileSuccessful=True, verify Vault state at `{mount}/roles/{name}`
  - [ ] 4.5: Add update context for RabbitMQSecretEngineRole — update `tags`, verify Vault reflects change, verify ObservedGeneration increased
  - [ ] 4.6: Add deletion context — delete role (IsDeletable=true, verify Vault cleanup), delete config (IsDeletable=false, verify Vault persistence), delete mount, delete secret

- [ ] Task 5: End-to-end verification (AC: 1, 2, 3, 4, 5, 6)
  - [ ] 5.1: Run `make integration` and verify new tests pass alongside all existing tests
  - [ ] 5.2: Verify no regressions — existing tests unaffected

## Dev Notes

### Infrastructure Scope — RabbitMQ Deployment Needed (New Infra)

Per the Epic 4 retrospective's readiness assessment:

> Story 5.2 — RabbitMQ secret engine types | RabbitMQ | Install in Kind | Medium — new `deploy-rabbitmq` target

Per the project's three-tier integration test infrastructure rule:

> 1. **Install in Kind** — If the service can be installed in the Kind cluster and configured to work with Vault, the test **must** deploy it as a real service (e.g., PostgreSQL via Helm, **RabbitMQ via Helm**, OpenLDAP via manifests).

RabbitMQ is explicitly listed as a "Install in Kind" example. A new `deploy-rabbitmq` Makefile target is needed, following the same pattern as `deploy-postgresql`.

[Source: _bmad-output/project-context.md#L134-L141 — Integration test infrastructure philosophy]
[Source: _bmad-output/implementation-artifacts/epic-4-retro-2026-04-23.md#L147 — Story 5.2 infra classification]

### RabbitMQ Helm Deployment

Deploy using the Bitnami RabbitMQ chart with a `rabbitmq-values.yaml` file:

```yaml
fullnameOverride: my-rabbitmq
auth:
  username: admin
  password: testpassword123
```

The Bitnami chart enables the management plugin by default (exposed on port 15672). The management API is what Vault's RabbitMQ secret engine uses to create/delete users.

RabbitMQ management API URL from within the Kind cluster: `http://my-rabbitmq.test-vault-config-operator.svc:15672`

Makefile target (follow `deploy-postgresql` pattern):

```makefile
.PHONY: deploy-rabbitmq
deploy-rabbitmq: kubectl helm
	$(HELM) repo add bitnami https://charts.bitnami.com/bitnami || true
	$(HELM) upgrade -i rabbitmq bitnami/rabbitmq \
		-n test-vault-config-operator --create-namespace --atomic \
		-f ./integration/rabbitmq-values.yaml
	$(KUBECTL) wait --for=condition=ready pod -l app.kubernetes.io/instance=rabbitmq \
		-n test-vault-config-operator --timeout=$(KUBECTL_WAIT_TIMEOUT)
```

Wire into integration target by adding `deploy-rabbitmq` to the dependency list.

### RabbitMQSecretEngineConfig — CRITICAL: Custom Reconcile Flow (NOT VaultResource)

The RabbitMQSecretEngineConfig controller does **NOT** use `NewVaultResource`. It has a custom `manageReconcileLogic` that uses `NewRabbitMQEngineConfigVaultEndpoint`:

1. `PrepareInternalValues()` — resolves root credentials from K8s Secret (same credential resolution as DatabaseSecretEngineConfig)
2. `rabbitMQVaultEndpoint.Create(context)` — **always writes** to `{path}/config/connection` via `write()` directly (no read-compare-write like VaultResource.CreateOrUpdate)
3. `rabbitMQVaultEndpoint.CreateOrUpdateLease(context)` — reads `{path}/config/lease`, compares via `IsEquivalentToDesiredState()` (which uses `leasesToMap()` → `{ttl, max_ttl}`), writes only if different

This means:
- **Connection config** is always written on every reconcile (no idempotency check)
- **Lease config** uses the standard read-compare-write pattern but compares only `ttl` and `max_ttl`
- There is also a custom deletion guard in the controller: `if !instance.DeletionTimestamp.IsZero() { return reconcile.Result{}, nil }` — returns immediately without calling `ManageOutcome` for deletion

[Source: controllers/rabbitmqsecretengineconfig_controller.go#L78-L89 — Custom reconcile flow]
[Source: api/v1alpha1/utils/vaultobject.go#L183-L216 — RabbitMQEngineConfigVaultEndpoint]

### Two Vault Paths for Config

RabbitMQSecretEngineConfig writes to TWO Vault paths:

1. **Connection:** `{spec.path}/config/connection` (via `GetPath()`)
   - Payload: `connection_uri`, `verify_connection`, `username`, `password`, `username_template`, `password_policy` (via `rabbitMQToMap()`)
   - Written via `Create()` — always writes, no read-compare

2. **Lease:** `{spec.path}/config/lease` (via `GetLeasePath()`)
   - Payload: `ttl`, `max_ttl` (via `leasesToMap()`)
   - Written via `CreateOrUpdateLease()` — reads first, compares via `IsEquivalentToDesiredState(leasesToMap())`, writes only if different
   - **Skipped entirely** if both `LeaseTTL` and `LeaseMaxTTL` are 0 (`CheckTTLValuesProvided()`)

[Source: api/v1alpha1/rabbitmqsecretengineconfig_types.go#L129-L139 — rabbitMQToMap]
[Source: api/v1alpha1/rabbitmqsecretengineconfig_types.go#L322-L327 — leasesToMap]
[Source: api/v1alpha1/rabbitmqsecretengineconfig_types.go#L171-L173 — GetPath: {spec.path}/config/connection]
[Source: api/v1alpha1/rabbitmqsecretengineconfig_types.go#L333-L335 — GetLeasePath: {spec.path}/config/lease]

### Config GetPath — Fixed Path, NOT Name-Based

Unlike DatabaseSecretEngineConfig which uses `{path}/config/{name}`, RabbitMQSecretEngineConfig uses a **fixed** path:

```go
func (rabbitMQ *RabbitMQSecretEngineConfig) GetPath() string {
    return string(rabbitMQ.Spec.Path) + "/config/connection"
}
```

This is because the RabbitMQ secret engine only supports ONE connection config per mount. The path doesn't include the metadata.name at all.

For fixture with `path: test-rmqse/test-rmq-mount` → Vault path is `test-rmqse/test-rmq-mount/config/connection`

Similarly, `GetLeasePath()` returns `test-rmqse/test-rmq-mount/config/lease`.

[Source: api/v1alpha1/rabbitmqsecretengineconfig_types.go#L171-L173]

### Config IsDeletable = false — Verify Vault Persistence After CR Deletion

`IsDeletable()` returns `false`. This means:
- No finalizer is added
- No Vault cleanup on CR deletion
- Controller has explicit guard: `if !instance.DeletionTimestamp.IsZero() { return reconcile.Result{}, nil }`

The delete test MUST verify that Vault config **persists** after the CR is deleted from Kubernetes. Read the connection config from Vault and assert key fields (like `connection_uri`) still have the expected values.

This follows the same pattern established in Story 4.2 (LDAPAuthEngineConfig) and Story 4.3 (JWTOIDCAuthEngineConfig).

[Source: api/v1alpha1/rabbitmqsecretengineconfig_types.go#L156-L158 — IsDeletable returns false]
[Source: controllers/rabbitmqsecretengineconfig_controller.go#L78-L81 — Deletion guard]

### Config IsEquivalentToDesiredState — Uses leasesToMap(), NOT rabbitMQToMap()

```go
func (rabbitMQ *RabbitMQSecretEngineConfig) IsEquivalentToDesiredState(payload map[string]interface{}) bool {
    desiredState := rabbitMQ.Spec.RMQSEConfig.leasesToMap()
    return reflect.DeepEqual(desiredState, payload)
}
```

This is used ONLY by the `CreateOrUpdateLease` method (for the lease path). The connection `Create()` method always writes without comparison. The `IsEquivalentToDesiredState` only compares `{ttl, max_ttl}` from `leasesToMap()`.

[Source: api/v1alpha1/rabbitmqsecretengineconfig_types.go#L179-L182]

### Config Webhook — Custom Admission Handler (NOT Standard Kubebuilder)

The RabbitMQSecretEngineConfig webhook is a **manually registered** admission handler, not the standard kubebuilder webhook pattern:

```go
mgr.GetWebhookServer().Register(
    "/validate-redhatcop-redhat-io-v1alpha1-rabbitmqsecretengineconfig",
    &webhook.Admission{Handler: &redhatcopv1alpha1.RabbitMQSecretEngineConfigValidation{Client: mgr.GetClient()}})
```

The handler:
- **CREATE:** Lists ALL existing RabbitMQSecretEngineConfig CRs and rejects if another config already uses the same `spec.path` (+ vault namespace). This enforces the one-config-per-mount constraint.
- **UPDATE:** Rejects changes to `spec.path` (immutable after creation).
- **DELETE:** Always allowed.

**Integration test implication:** Only ONE RabbitMQSecretEngineConfig can exist per `spec.path`. The test must use a unique mount path that doesn't collide with the existing `test/rabbitmq-engine-config.yaml` fixture (which uses `test-vault-config-operator/rabbitmq`). The new test will use `test-rmqse/test-rmq-mount`.

[Source: api/v1alpha1/rabbitmqsecretengineconfig_webhook.go — Custom admission handler]
[Source: main.go#L510 — Manual webhook registration]

### Config Credential Resolution (PrepareInternalValues)

Same pattern as DatabaseSecretEngineConfig. For the test, use the K8s Secret path (simplest):

When `RootCredentials.Secret` is set:
- If `Username != ""` in spec → `retrievedUsername = spec.Username`, `retrievedPassword = secret.Data[passwordKey]`
- If `Username == ""` → both from K8s Secret

The test fixture sets `username: admin` in spec, so `retrievedUsername = "admin"` and `retrievedPassword` comes from the K8s Secret.

[Source: api/v1alpha1/rabbitmqsecretengineconfig_types.go#L201-L265]

### RabbitMQSecretEngineRole — Standard VaultResource Reconciler

Uses `NewVaultResource` — standard reconcile flow (read → compare → write if different).

**GetPath():**
```go
func (d *RabbitMQSecretEngineRole) GetPath() string {
    if d.Spec.Name != "" {
        return vaultutils.CleansePath(string(d.Spec.Path) + "/" + "roles" + "/" + d.Spec.Name)
    }
    return vaultutils.CleansePath(string(d.Spec.Path) + "/" + "roles" + "/" + d.Name)
}
```

For fixture with `path: test-rmqse/test-rmq-mount`, `metadata.name: test-rmq-role` → `test-rmqse/test-rmq-mount/roles/test-rmq-role`

[Source: api/v1alpha1/rabbitmqsecretenginerole_types.go#L169-L174]
[Source: controllers/rabbitmqsecretenginerole_controller.go#L74 — NewVaultResource]

### Role IsDeletable = true — Verify Vault Cleanup After CR Deletion

Finalizer added, Vault role deleted on CR deletion. Standard delete-and-verify pattern: check K8s NotFound + Vault Read returns nil.

[Source: api/v1alpha1/rabbitmqsecretenginerole_types.go#L155-L157]

### Role toMap — JSON Serialization for Vhosts

```go
func (fields *RMQSERole) rabbitMQToMap() map[string]interface{} {
    payload := map[string]interface{}{}
    payload["tags"] = fields.Tags
    payload["vhosts"] = convertVhostsToJson(fields.Vhosts)
    payload["vhost_topics"] = convertTopicsToJson(fields.VhostTopics)
    return payload
}
```

`vhosts` and `vhost_topics` are **JSON-encoded strings** in the Vault payload, not nested maps. `convertVhostsToJson` serializes the vhosts slice to JSON string.

**Vault API response for roles returns the same JSON-string format.** When verifying the role in Vault, `vhosts` will be a string, not a map.

[Source: api/v1alpha1/rabbitmqsecretenginerole_types.go#L233-L239]
[Source: api/v1alpha1/rabbitmqsecretenginerole_types.go#L199-L231 — convertVhostsToJson/convertTopicsToJson]

### Role IsEquivalentToDesiredState — Bare DeepEqual

```go
func (rabbitMQ *RabbitMQSecretEngineRole) IsEquivalentToDesiredState(payload map[string]interface{}) bool {
    desiredState := rabbitMQ.Spec.RMQSERole.rabbitMQToMap()
    return reflect.DeepEqual(desiredState, payload)
}
```

No filtering of extra keys. Vault may return extra fields → potential write on every reconcile (known tech debt — Story 7-4). Does NOT affect ReconcileSuccessful=True or test correctness.

[Source: api/v1alpha1/rabbitmqsecretenginerole_types.go#L178-L181]

### Vault API Response Shapes

**GET `{mount}/config/connection`** — Returns RabbitMQ connection config:
```json
{
  "data": {
    "connection_uri": "http://my-rabbitmq.test-vault-config-operator.svc:15672",
    "verify_connection": true,
    "username": "admin",
    "password_policy": "",
    "username_template": ""
  }
}
```
Key: `password` is never returned by Vault. Other fields may appear.

**GET `{mount}/config/lease`** — Returns lease config:
```json
{
  "data": {
    "ttl": 3600,
    "max_ttl": 86400
  }
}
```

**GET `{mount}/roles/{name}`** — Returns role config:
```json
{
  "data": {
    "tags": "administrator",
    "vhosts": "{\"/{\"configure\":\".*\",\"write\":\".*\",\"read\":\".*\"}}",
    "vhost_topics": "{}"
  }
}
```
`vhosts` and `vhost_topics` are JSON strings, not nested maps.

### Verifying Vault State

**Connection config verification:**
```go
secret, err := vaultClient.Logical().Read("test-rmqse/test-rmq-mount/config/connection")
Expect(err).To(BeNil())
Expect(secret).NotTo(BeNil())

connURI, ok := secret.Data["connection_uri"].(string)
Expect(ok).To(BeTrue(), "expected connection_uri to be a string")
Expect(connURI).To(Equal("http://my-rabbitmq.test-vault-config-operator.svc:15672"))

username, ok := secret.Data["username"].(string)
Expect(ok).To(BeTrue(), "expected username to be a string")
Expect(username).To(Equal("admin"))
```

**Lease config verification:**
```go
leaseSecret, err := vaultClient.Logical().Read("test-rmqse/test-rmq-mount/config/lease")
Expect(err).To(BeNil())
Expect(leaseSecret).NotTo(BeNil())

ttl, ok := leaseSecret.Data["ttl"].(json.Number)
Expect(ok).To(BeTrue(), "expected ttl to be json.Number")
ttlVal, err := ttl.Int64()
Expect(err).To(BeNil())
Expect(ttlVal).To(Equal(int64(3600)))

maxTTL, ok := leaseSecret.Data["max_ttl"].(json.Number)
Expect(ok).To(BeTrue(), "expected max_ttl to be json.Number")
maxTTLVal, err := maxTTL.Int64()
Expect(err).To(BeNil())
Expect(maxTTLVal).To(Equal(int64(86400)))
```

**Role verification:**
```go
secret, err := vaultClient.Logical().Read("test-rmqse/test-rmq-mount/roles/test-rmq-role")
Expect(err).To(BeNil())
Expect(secret).NotTo(BeNil())

tags, ok := secret.Data["tags"].(string)
Expect(ok).To(BeTrue(), "expected tags to be a string")
Expect(tags).To(Equal("administrator"))
```

**Delete verification (IsDeletable=false for config):**
```go
// Config — IsDeletable=false: verify Vault persistence
Expect(k8sIntegrationClient.Delete(ctx, configInstance)).Should(Succeed())
configLookupKey := types.NamespacedName{Name: configInstance.Name, Namespace: configInstance.Namespace}
Eventually(func() bool {
    err := k8sIntegrationClient.Get(ctx, configLookupKey, &redhatcopv1alpha1.RabbitMQSecretEngineConfig{})
    return apierrors.IsNotFound(err)
}, timeout, interval).Should(BeTrue())

configSecret, err := vaultClient.Logical().Read("test-rmqse/test-rmq-mount/config/connection")
Expect(err).To(BeNil())
Expect(configSecret).NotTo(BeNil(), "expected connection config to persist in Vault after CR deletion")
Expect(configSecret.Data["connection_uri"]).To(Equal("http://my-rabbitmq.test-vault-config-operator.svc:15672"))
```

### Test Design — Dependency Chain

```
K8s Secret (test-rmq-creds) — admin credentials for RabbitMQ
  └── SecretEngineMount (test-rmq-mount, type=rabbitmq, path=test-rmqse)
        └── RabbitMQSecretEngineConfig (test-rmq-config)
              → test-rmqse/test-rmq-mount/config/connection
              → test-rmqse/test-rmq-mount/config/lease
        └── RabbitMQSecretEngineRole (test-rmq-role)
              → test-rmqse/test-rmq-mount/roles/test-rmq-role
```

Resources must be created in order: Secret → Mount → Config → Role. Deletion in reverse: Role → Config → Mount → Secret.

The SecretEngineMount must be reconciled before the config, because Vault rejects writes to `{mount}/config/connection` if the rabbitmq engine mount doesn't exist.

### RabbitMQ Admin Credentials K8s Secret — Created in Test

```go
rmqSecret := &corev1.Secret{
    ObjectMeta: metav1.ObjectMeta{
        Name:      "test-rmq-creds",
        Namespace: vaultAdminNamespaceName,
    },
    StringData: map[string]string{
        "password": "testpassword123",
    },
}
```

Only `password` is needed in the secret because the fixture sets `username: admin` directly in the spec. `setInternalCredentials` will use `spec.Username` for the username and read only `passwordKey` from the secret.

### Test Fixture Design

**Fixture 1: `test/rabbitmqsecretengine/test-rmq-mount.yaml`** — SecretEngineMount prerequisite:
```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: SecretEngineMount
metadata:
  name: test-rmq-mount
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  type: rabbitmq
  path: test-rmqse
```
Mounts at `sys/mounts/test-rmqse/test-rmq-mount`. Uses `type: rabbitmq` to enable the RabbitMQ secret engine.

**Fixture 2: `test/rabbitmqsecretengine/test-rmq-config.yaml`** — RabbitMQSecretEngineConfig:
```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: RabbitMQSecretEngineConfig
metadata:
  name: test-rmq-config
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  path: test-rmqse/test-rmq-mount
  connectionURI: "http://my-rabbitmq.test-vault-config-operator.svc:15672"
  rootCredentials:
    secret:
      name: test-rmq-creds
    passwordKey: password
  username: admin
  verifyConnection: true
  leaseTTL: 3600
  leaseMaxTTL: 86400
```
`GetPath()` returns `test-rmqse/test-rmq-mount/config/connection`.
`GetLeasePath()` returns `test-rmqse/test-rmq-mount/config/lease`.

Key: `verifyConnection: true` means Vault will actually attempt to connect to RabbitMQ's management API when writing the config. This validates the RabbitMQ deployment is reachable and the credentials work.

Key: `rootCredentials.secret` only needs `passwordKey` because `username: admin` is set in spec. `setInternalCredentials` will use `spec.Username` for the username (per the logic: if `Username != ""`, use spec value).

Key: `leaseTTL: 3600` and `leaseMaxTTL: 86400` trigger `CheckTTLValuesProvided() == true`, so the lease config path will be written.

**Fixture 3: `test/rabbitmqsecretengine/test-rmq-role.yaml`** — RabbitMQSecretEngineRole:
```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: RabbitMQSecretEngineRole
metadata:
  name: test-rmq-role
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  path: test-rmqse/test-rmq-mount
  tags: "administrator"
  vhosts:
  - vhostName: "/"
    permissions:
      read: ".*"
      write: ".*"
      configure: ".*"
```
`GetPath()` returns `test-rmqse/test-rmq-mount/roles/test-rmq-role`.

### Test Structure

```
Describe("RabbitMQSecretEngine controllers", Ordered)
  var rmqSecret *corev1.Secret
  var mountInstance *redhatcopv1alpha1.SecretEngineMount
  var configInstance *redhatcopv1alpha1.RabbitMQSecretEngineConfig
  var roleInstance *redhatcopv1alpha1.RabbitMQSecretEngineRole

  AfterAll: best-effort delete all instances + rmq secret (reverse order)

  Context("When creating prerequisite resources")
    It("Should create the RabbitMQ credentials secret and rabbitmq engine mount")
      - Create test-rmq-creds K8s Secret in vault-admin namespace
      - Load test-rmq-mount.yaml via decoder.GetSecretEngineMountInstance
      - Set namespace to vaultAdminNamespaceName, create
      - Eventually poll for ReconcileSuccessful=True
      - Verify mount exists via sys/mounts key "test-rmqse/test-rmq-mount/"

  Context("When creating a RabbitMQSecretEngineConfig")
    It("Should write the RabbitMQ connection and lease config to Vault")
      - Load test-rmq-config.yaml via decoder.GetRabbitMQSecretEngineConfigInstance
      - Set namespace to vaultAdminNamespaceName, create
      - Eventually poll for ReconcileSuccessful=True
      - Read test-rmqse/test-rmq-mount/config/connection from Vault
      - Verify connection_uri = "http://my-rabbitmq.test-vault-config-operator.svc:15672"
      - Verify username = "admin"
      - Read test-rmqse/test-rmq-mount/config/lease from Vault
      - Verify ttl = 3600 (json.Number)
      - Verify max_ttl = 86400 (json.Number)

  Context("When creating a RabbitMQSecretEngineRole")
    It("Should create the role in Vault with correct settings")
      - Load test-rmq-role.yaml via decoder.GetRabbitMQSecretEngineRoleInstance
      - Set namespace to vaultAdminNamespaceName, create
      - Eventually poll for ReconcileSuccessful=True
      - Read test-rmqse/test-rmq-mount/roles/test-rmq-role
      - Verify tags = "administrator"

  Context("When updating a RabbitMQSecretEngineRole")
    It("Should update the role in Vault and reflect updated ObservedGeneration")
      - Record initial ObservedGeneration
      - Get latest role CR, update tags to "management"
      - Eventually verify Vault reflects updated tags
      - Verify ObservedGeneration increased

  Context("When deleting RabbitMQSecretEngine resources")
    It("Should clean up role from Vault, preserve config in Vault, and remove all K8s resources")
      - Delete role CR (IsDeletable=true → Vault cleanup)
      - Eventually verify K8s deletion (NotFound)
      - Eventually verify role removed from Vault (Read returns nil)
      - Delete config CR (IsDeletable=false → NO Vault cleanup)
      - Eventually verify K8s deletion (NotFound)
      - Verify connection config STILL EXISTS in Vault (connection_uri field present)
      - Delete SecretEngineMount
      - Eventually verify K8s deletion and mount gone from sys/mounts
      - Delete RabbitMQ credentials secret
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

`encoding/json` needed for `json.Number` handling when checking lease TTL values from Vault.

### Name Collision Prevention

Fixture names use `test-rmq-` prefix with a unique mount path `test-rmqse`:
- `test-rmqse/test-rmq-mount` — secret engine mount (unique path prefix)
- `test-rmq-config` — RabbitMQSecretEngineConfig CR name
- `test-rmq-role` — RabbitMQSecretEngineRole CR name and Vault role name
- `test-rmq-creds` — RabbitMQ credentials K8s Secret

These don't collide with:
- Existing test fixtures at `test/rabbitmq-engine-*.yaml` (use `test-vault-config-operator/rabbitmq` path and `rabbitmq-engine-admin` auth role)
- Epic 4 test resources (`test-k8s-auth/*`, `test-ldap-auth/*`, `test-jwt-oidc-auth/*`)
- Story 5.1 resources (`test-dbse/*`, `test-db-*`)
- All other existing test resources

### Existing Test Coexistence

The new test uses completely separate resources at a different mount path (`test-rmqse`) in `vault-admin` namespace, avoiding any conflicts with existing tests. The existing `test/rabbitmq-engine-*.yaml` fixtures are not used by any integration test today. Since Ginkgo v2 randomizes top-level Describe blocks, both test files run independently.

### Controller Registration — Already Done

Both controllers are registered in `suite_integration_test.go`:
```go
err = (&RabbitMQSecretEngineConfigReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "RabbitMQSecretEngineConfig")}).SetupWithManager(mgr)
Expect(err).ToNot(HaveOccurred())

err = (&RabbitMQSecretEngineRoleReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "RabbitMQSecretEngineRole")}).SetupWithManager(mgr)
```

No changes needed to the test suite setup.

[Source: controllers/suite_integration_test.go#L172-L176]

### Decoder Methods — BOTH Must Be Added

Neither `GetRabbitMQSecretEngineConfigInstance` nor `GetRabbitMQSecretEngineRoleInstance` exist in the decoder:

```go
func (d *decoder) GetRabbitMQSecretEngineConfigInstance(filename string) (*redhatcopv1alpha1.RabbitMQSecretEngineConfig, error) {
    obj, groupKindVersion, err := d.decodeFile(filename)
    if err != nil {
        return nil, err
    }
    kind := reflect.TypeOf(redhatcopv1alpha1.RabbitMQSecretEngineConfig{}).Name()
    if groupKindVersion.Kind == kind {
        o := obj.(*redhatcopv1alpha1.RabbitMQSecretEngineConfig)
        return o, nil
    }
    return nil, errDecode
}

func (d *decoder) GetRabbitMQSecretEngineRoleInstance(filename string) (*redhatcopv1alpha1.RabbitMQSecretEngineRole, error) {
    obj, groupKindVersion, err := d.decodeFile(filename)
    if err != nil {
        return nil, err
    }
    kind := reflect.TypeOf(redhatcopv1alpha1.RabbitMQSecretEngineRole{}).Name()
    if groupKindVersion.Kind == kind {
        o := obj.(*redhatcopv1alpha1.RabbitMQSecretEngineRole)
        return o, nil
    }
    return nil, errDecode
}
```

[Source: controllers/controllertestutils/decoder.go — existing pattern at lines 175-188]

### Vault TTL Format Gotcha

Same as Story 5.1: Vault returns TTL values as `json.Number`, not Go `int`. Use the pattern:
```go
ttl, ok := secret.Data["ttl"].(json.Number)
Expect(ok).To(BeTrue(), "expected ttl to be json.Number")
val, err := ttl.Int64()
Expect(err).To(BeNil())
Expect(val).To(Equal(int64(3600)))
```

[Source: controllers/databasesecretenginestaticrole_controller_test.go#L292-L300 — json.Number pattern]

### RabbitMQ Connection Verification Risk

`verifyConnection: true` means Vault will attempt to connect to RabbitMQ's management API. If RabbitMQ is not yet ready when the config CR is created, Vault will reject the write. Mitigations:
- The `deploy-rabbitmq` target waits for pod readiness before proceeding to tests
- RabbitMQ management API is available as soon as the pod is ready
- The standard `Eventually` polling (120s timeout) provides additional buffer

### `policy-admin` Permissions

The test uses `policy-admin` auth role in `vault-admin` namespace — the standard broad-permissions role used by all integration tests. It has permissions for `sys/mounts/*` and full access to engine paths. This is sufficient for creating RabbitMQ mounts and writing configs/roles.

### Checked Type Assertions

Per Epic 3 retro action item and Epic 4 practice: always use two-value form `val, ok := x.(string)` with `Expect(ok).To(BeTrue())` for all Vault response field assertions.

### No `make manifests generate` Needed

This story adds an integration test file, YAML fixtures, decoder methods, infrastructure values, and a Makefile target. No CRD types, controllers, or webhooks are changed.

### File Inventory — What Needs to Change

| # | File | Change Type | Description |
|---|------|-------------|-------------|
| 1 | `integration/rabbitmq-values.yaml` | New | Bitnami RabbitMQ Helm values |
| 2 | `Makefile` | Modified | Add `deploy-rabbitmq` target, wire into `integration` |
| 3 | `controllers/controllertestutils/decoder.go` | Modified | Add `GetRabbitMQSecretEngineConfigInstance` and `GetRabbitMQSecretEngineRoleInstance` |
| 4 | `test/rabbitmqsecretengine/test-rmq-mount.yaml` | New | SecretEngineMount prerequisite (type=rabbitmq) |
| 5 | `test/rabbitmqsecretengine/test-rmq-config.yaml` | New | RabbitMQSecretEngineConfig with real RabbitMQ connection |
| 6 | `test/rabbitmqsecretengine/test-rmq-role.yaml` | New | RabbitMQSecretEngineRole with tags and vhost permissions |
| 7 | `controllers/rabbitmqsecretengine_controller_test.go` | New | Integration test — create mount, config, role; verify Vault state; update role; delete and verify |

No changes to suite setup, controllers, webhooks, types, or other infrastructure manifests.

### Previous Story Intelligence

**From Story 5.1 (DatabaseSecretEngine integration tests — immediate predecessor in Epic 5):**
- Established the full Epic 5 integration test pattern for secret engines
- Demonstrated K8s Secret created programmatically in the test
- Showed database prerequisite chain: Secret → Mount → Config → Role (same chain applies for RabbitMQ)
- Both types were IsDeletable=true — different from RabbitMQ config which is IsDeletable=false
- Used `json.Number` pattern for TTL assertions — reuse for lease TTL verification

**From Story 4.3 (JWTOIDCAuthEngine integration tests):**
- Established the IsDeletable=false config persistence verification pattern (config deleted from K8s but persists in Vault)
- This is the exact pattern needed for RabbitMQSecretEngineConfig delete test

**From Story 4.2 (LDAPAuthEngine integration tests):**
- Original establishment of the `IsDeletable=false` persistence verification rule (codified in project-context.md)
- Checked type assertions for Vault response fields

**From Epic 4 Retrospective:**
- Story 5.2 classified as "Medium — new `deploy-rabbitmq` target"
- Continue using Opus-class models for integration test stories
- Non-deletable config persistence rule — directly applicable to RabbitMQSecretEngineConfig

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
- **RabbitMQ:** Can be deployed in Kind via Bitnami Helm → **Tier 1: Install in Kind**
- **Vault API:** Already available in Kind
- **K8s Secrets:** Available via integration test client

**Classification: New infrastructure required — Medium scope story in Epic 5**

[Source: _bmad-output/project-context.md#L134-L141 — Integration test infrastructure philosophy]

### Project Structure Notes

- Helm values in `integration/rabbitmq-values.yaml` (alongside `postgresql-values.yaml`)
- Makefile `deploy-rabbitmq` target (alongside `deploy-postgresql`)
- Decoder changes in `controllers/controllertestutils/decoder.go` (add two methods)
- Test file goes in `controllers/rabbitmqsecretengine_controller_test.go`
- Test fixtures go in `test/rabbitmqsecretengine/` directory (new directory)
- All files follow existing naming conventions

### References

- [Source: api/v1alpha1/rabbitmqsecretengineconfig_types.go] — RabbitMQEngineConfigVaultObject implementation, GetPath ({spec.path}/config/connection), GetPayload (rabbitMQToMap), IsEquivalentToDesiredState (leasesToMap), GetLeasePath ({spec.path}/config/lease), GetLeasePayload (leasesToMap), CheckTTLValuesProvided, PrepareInternalValues (credential resolution), IsDeletable=false
- [Source: api/v1alpha1/rabbitmqsecretengineconfig_types.go#L171-L173] — GetPath: {spec.path}/config/connection (fixed, no name)
- [Source: api/v1alpha1/rabbitmqsecretengineconfig_types.go#L129-L139] — rabbitMQToMap (connection_uri, verify_connection, username, password, username_template, password_policy)
- [Source: api/v1alpha1/rabbitmqsecretengineconfig_types.go#L322-L327] — leasesToMap (ttl, max_ttl)
- [Source: api/v1alpha1/rabbitmqsecretengineconfig_types.go#L179-L182] — IsEquivalentToDesiredState uses leasesToMap(), NOT rabbitMQToMap()
- [Source: api/v1alpha1/rabbitmqsecretengineconfig_types.go#L156-L158] — IsDeletable returns false
- [Source: api/v1alpha1/rabbitmqsecretengineconfig_types.go#L201-L265] — setInternalCredentials (3 credential sources)
- [Source: api/v1alpha1/rabbitmqsecretengineconfig_webhook.go] — Custom admission handler: one-config-per-path on CREATE, immutable path on UPDATE
- [Source: api/v1alpha1/rabbitmqsecretenginerole_types.go] — VaultObject implementation, GetPath ({path}/roles/{name}), toMap (tags, vhosts as JSON string, vhost_topics as JSON string), IsDeletable=true
- [Source: api/v1alpha1/rabbitmqsecretenginerole_types.go#L169-L174] — GetPath: {spec.path}/roles/{name or metadata.name}
- [Source: api/v1alpha1/rabbitmqsecretenginerole_types.go#L178-L181] — IsEquivalentToDesiredState: bare DeepEqual, no filtering
- [Source: api/v1alpha1/rabbitmqsecretenginerole_types.go#L233-L239] — rabbitMQToMap (tags, vhosts=JSON string, vhost_topics=JSON string)
- [Source: api/v1alpha1/rabbitmqsecretenginerole_types.go#L199-L231] — convertVhostsToJson, convertTopicsToJson
- [Source: api/v1alpha1/rabbitmqsecretenginerole_webhook.go] — Standard webhook: immutable path on update
- [Source: controllers/rabbitmqsecretengineconfig_controller.go] — Custom reconcile flow: PrepareInternalValues → Create (always write connection) → CreateOrUpdateLease (read-compare-write lease)
- [Source: controllers/rabbitmqsecretengineconfig_controller.go#L78-L81] — Deletion guard: early return without ManageOutcome
- [Source: controllers/rabbitmqsecretenginerole_controller.go#L74] — Standard VaultResource reconciler
- [Source: controllers/suite_integration_test.go#L172-L176] — Both controllers registered
- [Source: controllers/controllertestutils/decoder.go] — Neither GetRabbitMQSecretEngineConfigInstance nor GetRabbitMQSecretEngineRoleInstance exist
- [Source: api/v1alpha1/utils/vaultobject.go#L176-L216] — RabbitMQEngineConfigVaultObject interface, RabbitMQEngineConfigVaultEndpoint struct, Create (always write), CreateOrUpdateLease (read-compare-write)
- [Source: main.go#L234-L239] — Controller registration
- [Source: main.go#L510] — Custom webhook registration for RabbitMQSecretEngineConfig
- [Source: test/rabbitmq-engine-config.yaml] — Existing fixture (verifyConnection: false, fake URL, different path)
- [Source: test/rabbitmq-engine-owner-role.yaml] — Existing fixture (different path and auth role)
- [Source: integration/postgresql-values.yaml] — PostgreSQL Helm values pattern to follow for RabbitMQ
- [Source: Makefile#L174-L181] — deploy-postgresql target pattern to follow for deploy-rabbitmq
- [Source: Makefile#L134-L135] — integration target dependency list (add deploy-rabbitmq)
- [Source: controllers/jwtoidcauthengine_controller_test.go] — IsDeletable=false persistence verification pattern
- [Source: controllers/databasesecretenginestaticrole_controller_test.go#L292-L300] — json.Number TTL pattern
- [Source: _bmad-output/implementation-artifacts/epic-4-retro-2026-04-23.md] — Epic 4 retrospective (infrastructure classification, readiness assessment)
- [Source: _bmad-output/project-context.md#L134-L141] — Integration test infrastructure philosophy

## Dev Agent Record

### Agent Model Used

{{agent_model_name_version}}

### Debug Log References

### Completion Notes List

### File List
