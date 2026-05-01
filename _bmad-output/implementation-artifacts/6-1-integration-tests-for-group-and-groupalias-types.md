# Story 6.1: Integration Tests for Group and GroupAlias Types

Status: ready-for-dev

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As an operator developer,
I want integration tests for Group and GroupAlias covering create, reconcile success, Vault state verification, update, and delete,
So that identity group management — including GroupAlias's non-trivial `PrepareInternalValues` for accessor lookup and asymmetric alias creation API — is verified end-to-end.

## Scope Decision: Tier Classification

Per the project's three-tier integration test rule:

| Type | Dependency | Classification | Action |
|------|-----------|---------------|--------|
| Group | Vault identity API (internal) | **Tier 1: Already available** | **Test** — no external service |
| GroupAlias | Vault identity API + auth mount accessor | **Tier 1: Already available** | **Test** — auth mount already exists in test env |

No new infrastructure needed. Both types interact with Vault's internal identity subsystem.

[Source: _bmad-output/project-context.md#L134-L141 — Integration test infrastructure philosophy]

## Acceptance Criteria

1. **Given** a Group CR is created with `type: external`, metadata, and policies **When** the reconciler processes it **Then** the group exists in Vault at `/identity/group/name/{name}` with `type=external`, correct metadata map, correct policies list, and ReconcileSuccessful=True

2. **Given** the Group CR spec is updated (e.g., `policies` list changed) **When** the reconciler processes the update **Then** the Vault group reflects the updated policies and `ObservedGeneration` increases

3. **Given** a GroupAlias CR referencing the group (via `groupName`) and the `kubernetes` auth engine mount (via `authEngineMountPath`) **When** the reconciler processes it **Then** the alias is created in Vault, `Status.ID` is populated on the CR, and ReconcileSuccessful=True

4. **Given** the GroupAlias CR is deleted (IsDeletable=true) **When** the reconciler processes the deletion **Then** the alias is removed from Vault and the CR is deleted from K8s

5. **Given** the Group CR is deleted (IsDeletable=true) **When** the reconciler processes the deletion **Then** the group is removed from Vault and the CR is deleted from K8s

## Tasks / Subtasks

- [ ] Task 1: Add decoder methods (AC: 1, 3)
  - [ ] 1.1: Add `GetGroupInstance` to `controllers/controllertestutils/decoder.go`
  - [ ] 1.2: Add `GetGroupAliasInstance` to `controllers/controllertestutils/decoder.go`

- [ ] Task 2: Create test fixtures (AC: 1, 3)
  - [ ] 2.1: Create `test/groups/test-group.yaml` — Group with `type: external`, metadata, policies (unique names to avoid collision with existing `group-sample` fixture)
  - [ ] 2.2: Create `test/groups/test-groupalias.yaml` — GroupAlias referencing the test group and `kubernetes` auth mount

- [ ] Task 3: Create integration test file (AC: 1, 2, 3, 4, 5)
  - [ ] 3.1: Create `controllers/group_controller_test.go` with `//go:build integration` tag
  - [ ] 3.2: Add context for Group creation — create, poll for ReconcileSuccessful=True, verify Vault state at `/identity/group/name/{name}`
  - [ ] 3.3: Add context for Group update — update policies, verify Vault reflects change, verify ObservedGeneration increased
  - [ ] 3.4: Add context for GroupAlias creation — create alias CR, poll for ReconcileSuccessful=True, verify Status.ID is populated, verify alias exists in Vault at `/identity/group-alias/id/{status.id}`
  - [ ] 3.5: Add deletion context — delete alias (IsDeletable=true, verify Vault cleanup), delete group (IsDeletable=true, verify Vault cleanup)

- [ ] Task 4: End-to-end verification (AC: 1, 2, 3, 4, 5)
  - [ ] 4.1: Run `make integration` and verify new tests pass alongside all existing tests
  - [ ] 4.2: Verify no regressions — existing tests unaffected

## Dev Notes

### Group — VaultResource Reconciler (Standard)

Uses `NewVaultResource` — standard reconcile flow (read → compare → write if different).

**GetPath():**
```go
func (d *Group) GetPath() string {
    if d.Spec.Name != "" {
        return vaultutils.CleansePath(string("/identity/group/name/" + d.Spec.Name))
    }
    return vaultutils.CleansePath(string("/identity/group/name/" + d.Name))
}
```

For fixture with `metadata.name: test-group` → Vault path is `identity/group/name/test-group`

[Source: api/v1alpha1/group_types.go#L138-L143]

### Group IsDeletable = true — Verify Vault Cleanup After CR Deletion

`IsDeletable() == true` means:
- Finalizer is added by ManageOutcome
- Vault resource is deleted on CR deletion
- Delete test must verify BOTH K8s NotFound AND Vault Read returns nil

[Source: api/v1alpha1/group_types.go#L123]

### Group toMap — Conditional Fields

```go
func (i *GroupSpec) toMap() map[string]interface{} {
    payload := map[string]interface{}{}
    payload["type"] = i.Type
    payload["metadata"] = i.Metadata
    payload["policies"] = i.Policies
    if i.Type == "internal" {
        payload["member_group_ids"] = i.MemberGroupIDs
        payload["member_entity_ids"] = i.MemberEntityIDs
    }
    return payload
}
```

**CRITICAL**: For `type: external`, `member_group_ids` and `member_entity_ids` are NOT included in the payload. The test fixture uses `type: external` so these fields should not appear in the Vault write.

[Source: api/v1alpha1/group_types.go#L148-L158]

### Group IsEquivalentToDesiredState — Deletes `name` from Payload

```go
func (d *Group) IsEquivalentToDesiredState(payload map[string]interface{}) bool {
    desiredState := d.Spec.toMap()
    delete(payload, "name")
    return reflect.DeepEqual(desiredState, payload)
}
```

Vault returns a `name` field in the read response that isn't in the write payload. The custom logic deletes it before comparison. Vault also likely returns additional identity-related fields (e.g., `id`, `creation_time`, `last_update_time`, `alias`, `member_group_ids`, `member_entity_ids`, `parent_group_ids`, `namespace_id`, `modify_index`). However, these are NOT currently filtered — this means `IsEquivalentToDesiredState` may return `false` on every reconcile (known tech debt — Story 7-4). This does NOT affect ReconcileSuccessful=True or test correctness.

[Source: api/v1alpha1/group_types.go#L176-L180]

### Group Webhook — No Validation Logic

Both `ValidateCreate` and `ValidateUpdate` are scaffold-only (return `nil, nil`). No immutable path rule (Group doesn't have a `spec.path` field). No defaulting logic.

[Source: api/v1alpha1/group_webhook.go]

### GroupAlias — Asymmetric API (Most Complex Part)

GroupAlias uses a different API pattern from most types:
1. **Create** is done in `PrepareInternalValues` by writing to `/identity/group-alias` (POST with no ID)
2. **Read/Update** uses `/identity/group-alias/id/{status.id}` (GET/PUT with ID)
3. The `Status.ID` field is populated by `PrepareInternalValues` on first reconcile

This is the same asymmetric pattern as EntityAlias.

[Source: api/v1alpha1/groupalias_types.go#L143-L194]

### GroupAlias PrepareInternalValues — Two Vault Lookups + Conditional Write

```go
func (d *GroupAlias) PrepareInternalValues(context context.Context, object client.Object) error {
    // 1. Lookup auth mount accessor from sys/auth/{authEngineMountPath}
    secret, found, err := vaultutils.ReadSecret(context, "sys/auth/"+d.Spec.AuthEngineMountPath)
    mountAccessor := secret.Data["accessor"].(string)

    // 2. Lookup group canonical ID from /identity/group/name/{groupName}
    secret, found, err = vaultutils.ReadSecret(context, "/identity/group/name/"+d.Spec.GroupName)
    canonicalID := secret.Data["id"].(string)

    // 3. If status.id is empty → CREATE the alias (first reconcile)
    if d.Status.ID == "" {
        payload := map[string]interface{}{
            "name":           d.Name (or d.Spec.Name),
            "mount_accessor": mountAccessor,
            "canonical_id":   canonicalID,
        }
        result, err := vaultClient.Logical().Write("/identity/group-alias", payload)
        d.Status.ID = result.Data["id"].(string)
        kubeClient.Status().Update(context, d)
    }

    // 4. Set retrieved fields (after any status update)
    d.Spec.retrievedMountAccessor = mountAccessor
    d.Spec.retrievedCanonicalID = canonicalID
    d.Spec.retrievedAliasID = d.Status.ID
    d.Spec.retrievedName = d.Name (or d.Spec.Name)
}
```

**CRITICAL IMPLICATIONS FOR TESTING:**
- The Group MUST exist in Vault BEFORE creating the GroupAlias (otherwise `PrepareInternalValues` will fail with "group not found")
- The `kubernetes` auth mount MUST be accessible at `sys/auth/kubernetes` (it already is in the integration test environment)
- After first reconcile, `Status.ID` will be populated — subsequent reconciles use `GetPath()` → `/identity/group-alias/id/{status.id}`

[Source: api/v1alpha1/groupalias_types.go#L143-L194]

### GroupAlias GetPath — Uses Status.ID

```go
func (d *GroupAlias) GetPath() string {
    return vaultutils.CleansePath("/identity/group-alias/id/" + d.Status.ID)
}
```

The path is dynamic and depends on `Status.ID` being populated. On first reconcile, the path would be `/identity/group-alias/id/` (empty ID) — but by that point `PrepareInternalValues` has already created the alias and set `Status.ID`.

[Source: api/v1alpha1/groupalias_types.go#L128-L130]

### GroupAlias toMap — 4 Internal Fields

```go
func (i *GroupAliasSpec) toMap() map[string]interface{} {
    payload := map[string]interface{}{}
    payload["name"] = i.retrievedName
    payload["id"] = i.retrievedAliasID
    payload["mount_accessor"] = i.retrievedMountAccessor
    payload["canonical_id"] = i.retrievedCanonicalID
    return payload
}
```

All fields are populated from `PrepareInternalValues`. The `GetPayload()` returns the results of dynamic lookups, not static spec fields.

[Source: api/v1alpha1/groupalias_types.go#L132-L139]

### GroupAlias IsEquivalentToDesiredState — Deletes 6 Extra Fields

```go
func (d *GroupAlias) IsEquivalentToDesiredState(payload map[string]interface{}) bool {
    desiredState := d.Spec.toMap()
    delete(payload, "creation_time")
    delete(payload, "last_update_time")
    delete(payload, "merged_from_canonical_ids")
    delete(payload, "metadata")
    delete(payload, "mount_path")
    delete(payload, "mount_type")
    return reflect.DeepEqual(desiredState, payload)
}
```

Vault returns extra fields for group aliases that are not managed by the operator.

[Source: api/v1alpha1/groupalias_types.go#L217-L227]

### GroupAlias IsDeletable = true — Verify Vault Cleanup

Both Group and GroupAlias have `IsDeletable() == true`. After deletion:
- GroupAlias: verify `/identity/group-alias/id/{captured_id}` returns nil
- Group: verify `/identity/group/name/{name}` returns nil

[Source: api/v1alpha1/groupalias_types.go#L117]

### GroupAlias Webhook — No Validation Logic

Scaffold-only like Group. No immutable path rule (GroupAlias doesn't have `spec.path`).

[Source: api/v1alpha1/groupalias_webhook.go]

### Vault API Response Shapes

**GET `/identity/group/name/{name}`** — Returns group data:
```json
{
  "data": {
    "id": "abc123-...",
    "name": "test-group",
    "type": "external",
    "metadata": {"team": "team-abc"},
    "policies": ["team-abc-access"],
    "member_group_ids": [],
    "member_entity_ids": [],
    "parent_group_ids": [],
    "alias": {},
    "creation_time": "2024-...",
    "last_update_time": "2024-...",
    "namespace_id": "root",
    "modify_index": 2
  }
}
```
Key: Many extra fields are returned but NOT currently filtered by `IsEquivalentToDesiredState` (only `name` is deleted). This means the reconciler may write on every cycle but will still achieve `ReconcileSuccessful=True`.

**POST `/identity/group-alias`** — Creates alias, returns:
```json
{
  "data": {
    "id": "def456-...",
    "canonical_id": "abc123-..."
  }
}
```

**GET `/identity/group-alias/id/{id}`** — Returns alias data:
```json
{
  "data": {
    "id": "def456-...",
    "name": "test-groupalias",
    "canonical_id": "abc123-...",
    "mount_accessor": "auth_kubernetes_xyz",
    "mount_path": "auth/kubernetes/",
    "mount_type": "kubernetes",
    "creation_time": "2024-...",
    "last_update_time": "2024-...",
    "merged_from_canonical_ids": null,
    "metadata": null
  }
}
```

### Verifying Vault State

**Group verification:**
```go
secret, err := vaultClient.Logical().Read("identity/group/name/test-group")
Expect(err).To(BeNil())
Expect(secret).NotTo(BeNil())

groupType, ok := secret.Data["type"].(string)
Expect(ok).To(BeTrue(), "expected type to be a string")
Expect(groupType).To(Equal("external"))

metadata, ok := secret.Data["metadata"].(map[string]interface{})
Expect(ok).To(BeTrue(), "expected metadata to be map[string]interface{}")
Expect(metadata["team"]).To(Equal("team-abc"))

policies, ok := secret.Data["policies"].([]interface{})
Expect(ok).To(BeTrue(), "expected policies to be []interface{}")
Expect(policies).To(ContainElement("team-abc-access"))
```

**GroupAlias verification:**
```go
// After reconcile, Status.ID should be populated
Expect(created.Status.ID).NotTo(BeEmpty())
capturedAliasID := created.Status.ID

// Read alias from Vault using the captured ID
secret, err := vaultClient.Logical().Read("identity/group-alias/id/" + capturedAliasID)
Expect(err).To(BeNil())
Expect(secret).NotTo(BeNil())

aliasName, ok := secret.Data["name"].(string)
Expect(ok).To(BeTrue(), "expected name to be a string")
Expect(aliasName).To(Equal("test-groupalias"))

mountAccessor, ok := secret.Data["mount_accessor"].(string)
Expect(ok).To(BeTrue(), "expected mount_accessor to be a string")
Expect(mountAccessor).NotTo(BeEmpty())
```

**Delete verification (both IsDeletable=true):**
```go
// GroupAlias — IsDeletable=true: verify Vault cleanup
Expect(k8sIntegrationClient.Delete(ctx, aliasCreated)).Should(Succeed())
aliasLookupKey := types.NamespacedName{Name: aliasCreated.Name, Namespace: aliasCreated.Namespace}
Eventually(func() bool {
    err := k8sIntegrationClient.Get(ctx, aliasLookupKey, &redhatcopv1alpha1.GroupAlias{})
    return apierrors.IsNotFound(err)
}, timeout, interval).Should(BeTrue())

Eventually(func() bool {
    secret, err := vaultClient.Logical().Read("identity/group-alias/id/" + capturedAliasID)
    if err != nil {
        return false
    }
    return secret == nil
}, timeout, interval).Should(BeTrue())

// Group — IsDeletable=true: verify Vault cleanup
Expect(k8sIntegrationClient.Delete(ctx, groupCreated)).Should(Succeed())
groupLookupKey := types.NamespacedName{Name: groupCreated.Name, Namespace: groupCreated.Namespace}
Eventually(func() bool {
    err := k8sIntegrationClient.Get(ctx, groupLookupKey, &redhatcopv1alpha1.Group{})
    return apierrors.IsNotFound(err)
}, timeout, interval).Should(BeTrue())

Eventually(func() bool {
    secret, err := vaultClient.Logical().Read("identity/group/name/test-group")
    if err != nil {
        return false
    }
    return secret == nil
}, timeout, interval).Should(BeTrue())
```

### Test Design — Dependency Chain

```
kubernetes auth mount (pre-existing in test env)
  └── Group (test-group, type=external)
        └── GroupAlias (test-groupalias, references group + kubernetes auth mount)
```

The `kubernetes` auth mount is already available in the integration test environment (used by all tests for authentication). No additional prerequisite setup needed.

Resources must be created in order: Group → GroupAlias.
Deletion in reverse: GroupAlias → Group.

### Test Fixture Design

**Fixture 1: `test/groups/test-group.yaml`** — Group with external type:
```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: Group
metadata:
  name: test-group
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  type: external
  metadata:
    team: team-abc
  policies:
  - team-abc-access
```
Uses `type: external` because GroupAlias can ONLY be attached to external groups (Vault constraint — aliases map to external identity providers, not internal groups).

**Fixture 2: `test/groups/test-groupalias.yaml`** — GroupAlias:
```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: GroupAlias
metadata:
  name: test-groupalias
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  authEngineMountPath: kubernetes
  groupName: test-group
```
`authEngineMountPath: kubernetes` — references the already-existing kubernetes auth mount.
`groupName: test-group` — references the Group created in this test (must exist before alias is created).

### Test Structure

```
Describe("Group and GroupAlias controllers", Ordered)
  var groupInstance *redhatcopv1alpha1.Group
  var aliasInstance *redhatcopv1alpha1.GroupAlias
  var capturedAliasID string

  AfterAll: best-effort delete all instances (reverse order):
    alias → group

  Context("When creating a Group")
    It("Should create the group in Vault with correct settings")
      - Load test-group.yaml via decoder.GetGroupInstance
      - Set namespace to vaultAdminNamespaceName, create
      - Eventually poll for ReconcileSuccessful=True
      - Read identity/group/name/test-group from Vault
      - Verify type = "external"
      - Verify metadata["team"] = "team-abc"
      - Verify policies contains "team-abc-access"

  Context("When updating a Group")
    It("Should update the group in Vault and reflect updated ObservedGeneration")
      - Record initial ObservedGeneration
      - Get latest Group CR, add "kv-reader" to policies
      - Eventually verify Vault reflects updated policies
      - Verify ObservedGeneration increased

  Context("When creating a GroupAlias")
    It("Should create the alias in Vault with Status.ID populated")
      - Load test-groupalias.yaml via decoder.GetGroupAliasInstance
      - Set namespace to vaultAdminNamespaceName, create
      - Eventually poll for ReconcileSuccessful=True
      - Verify Status.ID is not empty
      - Capture Status.ID for later assertions
      - Read identity/group-alias/id/{status.id} from Vault
      - Verify name = "test-groupalias"
      - Verify mount_accessor is non-empty
      - Verify canonical_id matches the group's ID

  Context("When deleting GroupAlias and Group resources")
    It("Should clean up alias and group from Vault and remove all K8s resources")
      - Delete alias CR (IsDeletable=true → Vault cleanup)
      - Eventually verify K8s deletion (NotFound)
      - Eventually verify alias removed from Vault (Read returns nil)
      - Delete Group CR (IsDeletable=true → Vault cleanup)
      - Eventually verify K8s deletion (NotFound)
      - Eventually verify group removed from Vault (Read returns nil)
```

### Import Requirements

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

No additional imports needed — no RBAC, no ServiceAccounts, no infrastructure setup.

### Name Collision Prevention

Fixture names use `test-group` and `test-groupalias` prefix:
- `test-group` — Group CR name and Vault identity group name
- `test-groupalias` — GroupAlias CR name and Vault alias name

These don't collide with:
- Existing `test/groups/group-sample` and `test/groups/groupalias-sample` fixtures
- Entity/EntityAlias tests (`test-entity`, `test-entity-alias`)
- Other epic tests (database, RabbitMQ, Kubernetes secret engine)

### Controller Registration — Already Done

Both controllers are registered in `suite_integration_test.go`:
```go
err = (&GroupReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "Group")}).SetupWithManager(mgr)
Expect(err).ToNot(HaveOccurred())

err = (&GroupAliasReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "GroupAlias")}).SetupWithManager(mgr)
Expect(err).ToNot(HaveOccurred())
```

No changes needed to the test suite setup.

[Source: controllers/suite_integration_test.go#L208-L212]

### Decoder Methods — BOTH Must Be Added

Neither `GetGroupInstance` nor `GetGroupAliasInstance` exist in the decoder:

```go
func (d *decoder) GetGroupInstance(filename string) (*redhatcopv1alpha1.Group, error) {
    obj, groupKindVersion, err := d.decodeFile(filename)
    if err != nil {
        return nil, err
    }
    kind := reflect.TypeOf(redhatcopv1alpha1.Group{}).Name()
    if groupKindVersion.Kind == kind {
        o := obj.(*redhatcopv1alpha1.Group)
        return o, nil
    }
    return nil, errDecode
}

func (d *decoder) GetGroupAliasInstance(filename string) (*redhatcopv1alpha1.GroupAlias, error) {
    obj, groupKindVersion, err := d.decodeFile(filename)
    if err != nil {
        return nil, err
    }
    kind := reflect.TypeOf(redhatcopv1alpha1.GroupAlias{}).Name()
    if groupKindVersion.Kind == kind {
        o := obj.(*redhatcopv1alpha1.GroupAlias)
        return o, nil
    }
    return nil, errDecode
}
```

[Source: controllers/controllertestutils/decoder.go — existing pattern]

### GroupAlias canonical_id Verification

After creating the GroupAlias, you can verify the `canonical_id` matches the Group's ID by reading the group first:
```go
groupSecret, err := vaultClient.Logical().Read("identity/group/name/test-group")
Expect(err).To(BeNil())
groupID := groupSecret.Data["id"].(string)

aliasSecret, err := vaultClient.Logical().Read("identity/group-alias/id/" + capturedAliasID)
canonicalID, ok := aliasSecret.Data["canonical_id"].(string)
Expect(ok).To(BeTrue())
Expect(canonicalID).To(Equal(groupID))
```

This proves that `PrepareInternalValues` correctly resolved the group lookup.

### GroupAlias — Must Use External Group Type

Vault only allows group aliases on **external** groups. If the fixture uses `type: internal`, the alias creation will fail with a Vault error. The test fixture MUST use `type: external`.

### Vault's `kubernetes` Auth Mount Accessor

The `kubernetes` auth mount is pre-configured in the integration test environment. `PrepareInternalValues` reads `sys/auth/kubernetes` to get the accessor. This works because the test environment's `policy-admin` role has broad permissions including `sys/auth/*` reads.

### ObservedGeneration Baseline Assertion

Per Epic 5 retrospective action item: When testing updates, record initial ObservedGeneration BEFORE the update, then assert the post-update value is strictly greater than the recorded baseline.

### Checked Type Assertions

Per project convention: always use two-value form `val, ok := x.(string)` with `Expect(ok).To(BeTrue())` for all Vault response field assertions.

### No `make manifests generate` Needed

This story adds an integration test file, YAML fixtures, and decoder methods. No CRD types, controllers, or webhooks are changed. No Makefile changes needed.

### File Inventory — What Needs to Change

| # | File | Change Type | Description |
|---|------|-------------|-------------|
| 1 | `controllers/controllertestutils/decoder.go` | Modified | Add `GetGroupInstance` and `GetGroupAliasInstance` |
| 2 | `test/groups/test-group.yaml` | New | Group fixture (type=external, metadata, policies) |
| 3 | `test/groups/test-groupalias.yaml` | New | GroupAlias fixture (references test-group + kubernetes auth mount) |
| 4 | `controllers/group_controller_test.go` | New | Integration test — Group CRUD + GroupAlias create/delete with Vault verification |

No changes to suite setup, controllers, webhooks, types, Makefile, or other infrastructure manifests.

### Previous Story Intelligence

**From Story 5.3 (KubernetesSecretEngine integration tests — most recent):**
- Established Tier 1 "no new infrastructure" pattern — same applies here
- Demonstrated IsDeletable=true Vault cleanup verification
- Demonstrated Ordered Describe block with shared state across Contexts
- Confirmed 63 integration test specs passing (target: all pass + new specs)

**From Entity/EntityAlias integration tests (closest analogy for GroupAlias):**
- EntityAlias test at `controllers/entityalias_controller_test.go` demonstrates the **exact same asymmetric API pattern**
- EntityAlias has `Status.ID` populated by `PrepareInternalValues` — same as GroupAlias
- EntityAlias test verifies `Status.ID` is not empty after first reconcile
- EntityAlias test captures the ID for Vault read/delete verification
- GroupAlias is almost identical: replace "entity" with "group" and "entityName" with "groupName"

**From Epic 5 Retrospective:**
- Epic 6 Story 6.1 noted as "GroupAlias has non-trivial `PrepareInternalValues` (accessor lookup)"
- No infrastructure needed for any Epic 6 story
- ObservedGeneration baseline assertion guidance: record before update, assert strictly greater after
- Continue using Opus-class models

[Source: _bmad-output/implementation-artifacts/epic-5-retro-2026-04-29.md]

### Git Intelligence (Recent Commits)

```
7ce3e42 Merge pull request #320 from raffaelespazzoli/bmad-epic-5
d64b2b1 Complete Epic 5 retrospective and close epic
e5e982c Add integration tests for KubernetesSecretEngineConfig and KubernetesSecretEngineRole (Story 5.3)
168e7e0 Fix RabbitMQ role vhosts assertion type mismatch
c13227f Add integration tests for RabbitMQSecretEngineConfig and RabbitMQSecretEngineRole (Story 5.2)
```

Codebase is clean post-Epic 5 merge to main. All 63 integration tests passing with 46.0% coverage.

### Project Structure Notes

- Decoder changes in `controllers/controllertestutils/decoder.go` (add two methods)
- Test file goes in `controllers/group_controller_test.go`
- Test fixtures go in `test/groups/` directory (alongside existing `group-sample` and `groupalias-sample` fixtures, with `test-` prefix)
- No Makefile changes needed
- No new infrastructure directories
- All files follow existing naming conventions

### References

- [Source: api/v1alpha1/group_types.go] — VaultObject implementation, GetPath (/identity/group/name/{name}), GetPayload, IsEquivalentToDesiredState (deletes "name"), toMap (conditional member fields for internal type), IsDeletable=true, no PrepareInternalValues logic
- [Source: api/v1alpha1/group_types.go#L138-L143] — GetPath: /identity/group/name/{spec.name or metadata.name}
- [Source: api/v1alpha1/group_types.go#L148-L158] — GroupSpec.toMap (type, metadata, policies, conditional member fields)
- [Source: api/v1alpha1/group_types.go#L176-L180] — IsEquivalentToDesiredState: deletes "name" from payload then DeepEqual
- [Source: api/v1alpha1/group_webhook.go] — Scaffold-only webhook, no validation logic
- [Source: api/v1alpha1/groupalias_types.go] — VaultObject implementation, GetPath (/identity/group-alias/id/{status.id}), asymmetric API, PrepareInternalValues (accessor + canonical_id lookups + conditional create), IsDeletable=true
- [Source: api/v1alpha1/groupalias_types.go#L128-L130] — GetPath: /identity/group-alias/id/{status.id}
- [Source: api/v1alpha1/groupalias_types.go#L132-L139] — toMap (name, id, mount_accessor, canonical_id — all from retrieved internal values)
- [Source: api/v1alpha1/groupalias_types.go#L143-L194] — PrepareInternalValues: auth mount accessor lookup, group canonical_id lookup, conditional alias creation, status update
- [Source: api/v1alpha1/groupalias_types.go#L217-L227] — IsEquivalentToDesiredState: deletes 6 extra fields (creation_time, last_update_time, merged_from_canonical_ids, metadata, mount_path, mount_type)
- [Source: api/v1alpha1/groupalias_webhook.go] — Scaffold-only webhook, no validation logic
- [Source: controllers/group_controller.go] — Standard VaultResource reconciler
- [Source: controllers/groupalias_controller.go] — Standard VaultResource reconciler
- [Source: controllers/suite_integration_test.go#L208-L212] — Both controllers registered
- [Source: controllers/controllertestutils/decoder.go] — Neither GetGroupInstance nor GetGroupAliasInstance exist
- [Source: test/groups/group.yaml] — Existing fixture (group-sample, external type)
- [Source: test/groups/groupalias.yaml] — Existing fixture (groupalias-sample, references group-sample)
- [Source: controllers/entity_controller_test.go] — Entity integration test pattern (create, update with Vault verification, delete with Vault cleanup)
- [Source: controllers/entityalias_controller_test.go] — EntityAlias integration test pattern (Status.ID verification, captured ID for Vault assertions, asymmetric API)
- [Source: _bmad-output/implementation-artifacts/epic-5-retro-2026-04-29.md] — Epic 5 retrospective (Epic 6 readiness, no infrastructure needed, model guidance)
- [Source: _bmad-output/project-context.md#L134-L141] — Integration test infrastructure philosophy

## Dev Agent Record

### Agent Model Used

{{agent_model_name_version}}

### Debug Log References

### Completion Notes List

### File List
