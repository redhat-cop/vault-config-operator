# Deferred Work

## ~~Deferred from: code review of story 1-1 (2026-04-12)~~ RESOLVED

- ~~**AC #4 contradiction**: Story 1.1 AC #4 states that extra unmanaged Vault fields should be ignored (return `true`), but 7 of 9 types use `reflect.DeepEqual` which returns `false` when extra keys are present.~~ — Resolved by Epic 7, Story 7-4 ("audit Vault API responses and harden IsEquivalentToDesiredState extra-field handling") which introduced `filterPayloadToDesiredKeys` across all types. Verified during Epic 10 retrospective (2026-08-04).

## ~~Deferred from: code review of story 1-6 (2026-04-13)~~ RESOLVED

- ~~**GroupAlias debug prints leak into test and runtime output**: `api/v1alpha1/groupalias_types.go` still contains `fmt.Print("desired state", ...)` and `fmt.Print("actual state", ...)` inside `GroupAlias.IsEquivalentToDesiredState()`.~~ — Resolved: `fmt.Print` calls no longer present in `groupalias_types.go`. Verified by source code inspection during Epic 10 retrospective (2026-08-04).

## Deferred from: code review of 2-3-add-update-scenarios-to-entity-and-entityalias-integration-tests (2026-04-17)

- **ObservedGeneration assertion could be stronger**: Tests assert `ObservedGeneration > initial` (satisfies spec), but a stronger assertion `ObservedGeneration == metadata.generation` would more precisely verify the controller contract. Beyond spec requirement; defer to a test hardening pass.
- **Redundant Get before update**: Both update tests call `Get()` twice before `Update()` (once to record ObservedGeneration, once labeled "before update"). Minor clean-up; not a bug.
- **Post-update `disabled` field not re-verified in Vault**: After the Entity update, only policies and metadata are checked; `disabled` is not re-asserted. Spec doesn't require it; low risk.
- **Vault paths hardcoded as `"identity/entity/name/test-entity"`**: Should ideally derive from the CR `entityInstance.Name`. Not a bug (name is fixture-driven), but coupling concern for when fixtures change.
- **Duplicated `wait for ReconcileSuccessful` loop**: The Eventually polling pattern appears 3+ times per file. Pre-existing codebase pattern; extract to a helper when refactoring integration tests.
- **No conflict/retry on `k8sIntegrationClient.Update()`**: Single-shot update can produce 409 Conflict flakes under concurrent reconciliation. Pre-existing pattern across all Epic 2 stories; address in a broader test hardening story.
- **EntityAlias Vault deletion verified after both K8s deletes**: Timing dependency between K8s finalizer processing and Vault deletion is implicit. Low risk given sequential Ginkgo execution; document or add a sequential Vault delete check if flakiness observed.

## Deferred from: code review of R1-2c-lint-green-gate-verify-full-compliance (2026-06-14)

- **R1.2c scope expansion accepted**: The verification-only gate pragmatically included a 6-line fix in `controllers/vaultresourcecontroller/utils.go` for an R1.3 regression (`apimeta.SetStatusCondition` stopped advancing `LastTransitionTime` on same-status reconciles). Accepted as scope expansion rather than retroactively reopening R1.3.
- **`PeriodicReconcilePredicate` drift timer abuses `LastTransitionTime` as a reconcile heartbeat**: `utils.go:252-261` reads `ReconcileSuccessful.LastTransitionTime` to compute time-since-last-reconcile for drift detection gating. Standard Kubernetes condition conventions define `LastTransitionTime` as "last time the condition's Status field changed", not "last time the controller ran". The forced timestamp override at `utils.go:157-164` is a temporary workaround. **Requires a dedicated story/epic** to redesign `PeriodicReconcilePredicate` to use a different signal (status annotation, dedicated status field, or controller-internal bookkeeping), then remove the forced override so `apimeta.SetStatusCondition` operates with standard semantics. Consumers: only `PeriodicReconcilePredicate` (production) and `driftdetection_controller_test.go` (test) read the timestamp; all other `ReconcileSuccessful` consumers check boolean presence only.

## ~~Deferred from: code review of d3-3-standardize-pki-and-rabbitmq-secret-engine-docs (2026-07-05)~~ RESOLVED

- ~~**RabbitMQ multi-entry examples contradict current implementation**: `docs/secret-engines/rabbitmq.md` documents multiple `vhosts` and multiple `vhostTopics`, but `api/v1alpha1/rabbitmqsecretenginerole_types.go` overwrites the serialized map on each loop iteration, so only the last vhost and last topic set survive.~~ — Resolved by Story 9.0 (fix RabbitMQ vhost/vhostTopic multi-entry serialization bug). Verified during Epic 9 retrospective (2026-08-01).

## ~~Deferred from: code review of d3-4-standardize-github-quay-kubernetes-and-azure-secret-engine-docs (2026-07-05)~~ RESOLVED

- ~~**Existing README wording still says "see the also the"**~~ — Resolved: all 5 instances fixed during D3 retrospective session (2026-07-05). Verified by source code grep during Epic 10 retrospective (2026-08-04): zero occurrences remain.

## ~~Deferred from: code review of 8-2-upgrade-controller-runtime-and-k8s-libs (2026-07-17)~~ RESOLVED

- ~~**Broken validating webhook markers remain in unchanged Policy and PasswordPolicy webhooks**~~ — Fixed during Epic 8 retrospective session (2026-07-20). Verified by source code inspection during Epic 9 retrospective (2026-08-01).
- ~~**Copy-pasted `authenginemountlog` usage remains in migrated defaulters**~~ — Fixed during Epic 8 retrospective session (2026-07-20). All three files now use correct type-specific loggers. Verified during Epic 9 retrospective (2026-08-01).

## Deferred from: code review of 9-5-upgrade-vault-version-in-integration-test-infrastructure (2026-07-31)

- **Define the supported `deploy-vault` upgrade path across existing 1.x data**: Deferred because integration tests are expected to run on fresh environments, so Helm chart upgrade handling is out of scope for this story.
