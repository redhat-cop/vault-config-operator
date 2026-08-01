# Story 9.0: Fix RabbitMQ vhost/vhostTopic multi-entry serialization bug

Status: done

<!-- Note: Validation is optional. Run validate-create-story for quality check before dev-story. -->

## Story

As an operator developer,
I want to fix the RabbitMQ secret engine role serialization bug for vhost and vhostTopic multi-entry configurations,
So that users can correctly configure multiple vhost or vhostTopic entries on a RabbitMQ secret engine role.

## Acceptance Criteria

1. **Given** a RabbitMQSecretEngineRole CR with multiple vhost entries, **When** the `toMap()` method serializes the configuration, **Then** all vhost entries are correctly serialized as a JSON string map (not just the last entry).

2. **Given** a RabbitMQSecretEngineRole CR with multiple vhostTopic entries, **When** the `toMap()` method serializes the configuration, **Then** all vhostTopic entries are correctly serialized as a JSON string map (not just the last entry).

3. **Given** the fix is applied, **When** `make test` is run, **Then** all unit tests pass including new multi-entry serialization tests.

4. **Given** the fix is applied, **When** `make integration` is run, **Then** all integration tests pass (existing RabbitMQ integration test with single-entry fixture continues to work).

## Tasks / Subtasks

- [x] Task 1: Fix `convertVhostsToJson` map reassignment bug (AC: #1)
  - [x] 1.1 Change `vhostData = map[string]any{...}` to `vhostData[vhost.VhostName] = vhost.Permissions` in `api/v1alpha1/rabbitmqsecretenginerole_types.go:201-203`
- [x] Task 2: Fix `convertTopicsToJson` map reassignment bugs (AC: #2)
  - [x] 2.1 Move `topicData` creation inside the outer loop (scope per vhost) — line 214 must move to inside the `for _, vhost := range vhosts` loop
  - [x] 2.2 Change `topicData = map[string]any{...}` to `topicData[topic.TopicName] = topic.Permissions` (line 217-219)
  - [x] 2.3 Change `vhostData = map[string]any{...}` to `vhostData[vhost.VhostName] = topicData` (line 221-223)
- [x] Task 3: Replace `log.Fatal` with proper error handling (AC: #3)
  - [x] 3.1 Remove the `"log"` stdlib import
  - [x] 3.2 In both `convertVhostsToJson` and `convertTopicsToJson`, replace `log.Fatal(err)` with a panic or structured error propagation — `json.Marshal` of a `map[string]any` with string keys and simple struct values will never fail in practice, but the stdlib `log.Fatal` calls `os.Exit(1)` which would kill the operator process
- [x] Task 4: Add multi-entry unit tests with independently-constructed expected values (AC: #1, #2, #3)
  - [x] 4.1 Add `TestConvertVhostsToJsonMultipleEntries` — construct 2+ vhosts, call `convertVhostsToJson`, unmarshal the result, assert ALL entries are present (not just the last one)
  - [x] 4.2 Add `TestConvertTopicsToJsonMultipleEntries` — construct 2+ vhostTopics with 2+ topics each, call `convertTopicsToJson`, unmarshal the result, assert all vhosts and all topics within each vhost are present
  - [x] 4.3 Add `TestRMQSERoleRabbitMQToMapMultipleVhosts` — construct a role with multiple vhosts and vhostTopics, call `rabbitMQToMap()`, parse the resulting JSON strings, and verify against independently-constructed Vault API payloads (do NOT use `convertVhostsToJson`/`convertTopicsToJson` to build expected values)
  - [x] 4.4 Add `TestRabbitMQSecretEngineRoleIsEquivalentMultipleVhosts` — verify `IsEquivalentToDesiredState` works correctly with multi-entry payloads using independently-constructed Vault-read-shaped fixtures
- [x] Task 5: Run `make test` and `make integration` (AC: #3, #4)
  - [x] 5.1 Run `make test` — verify all unit tests pass
  - [x] 5.2 Run `make integration` — verify all integration tests pass (existing single-entry fixture)

## Dev Notes

### Root Cause Analysis

The bug is in two functions in `api/v1alpha1/rabbitmqsecretenginerole_types.go` (lines 198-230):

**Bug 1 — `convertVhostsToJson` (lines 198-210):**
```go
func convertVhostsToJson(vhosts []Vhost) string {
    vhostData := make(map[string]any)
    for _, vhost := range vhosts {
        vhostData = map[string]any{              // BUG: reassigns entire map
            vhost.VhostName: vhost.Permissions,
        }
    }
    // ...
}
```
Each loop iteration creates a **new** map and assigns it to `vhostData`, discarding all previous entries. Only the last vhost survives. Fix: `vhostData[vhost.VhostName] = vhost.Permissions`.

**Bug 2 — `convertTopicsToJson` (lines 212-230):**
```go
func convertTopicsToJson(vhosts []VhostTopic) string {
    vhostData := make(map[string]any)
    topicData := make(map[string]any)          // BUG: scoped outside loop
    for _, vhost := range vhosts {
        for _, topic := range vhost.Topics {
            topicData = map[string]any{         // BUG: reassigns
                topic.TopicName: topic.Permissions,
            }
        }
        vhostData = map[string]any{             // BUG: reassigns
            vhost.VhostName: topicData,
        }
    }
    // ...
}
```
Three problems: (a) `topicData` is scoped outside the outer loop so topics bleed across vhosts; (b) inner loop reassigns `topicData` (loses all but last topic); (c) outer loop reassigns `vhostData` (loses all but last vhost).

**Why existing tests didn't catch it:** All existing unit tests and the integration test fixture (`test/rabbitmqsecretengine/test-rmq-role.yaml`) use only **single** vhost/vhostTopic entries, so the reassignment produces the same output as appending. Additionally, `TestRMQSERoleRabbitMQToMap` compares against `convertVhostsToJson(vhosts)` — using the buggy function to build expected values, which violates the project's "never derive expected payloads from the code under test" rule.

### Vault API Format Reference

Vault's RabbitMQ secret engine `/roles/:name` endpoint expects `vhosts` and `vhost_topics` as **JSON strings**:

```json
{
  "tags": "tag1,tag2",
  "vhosts": "{\"/\": {\"configure\":\".*\", \"write\":\".*\", \"read\": \".*\"}}",
  "vhost_topics": "{\"/\": {\"amq.topic\": {\"write\":\".*\", \"read\": \".*\"}}}"
}
```

The read response returns the same format — strings, not parsed objects. This means `IsEquivalentToDesiredState` compares string values via `reflect.DeepEqual`, which works correctly as long as serialization is deterministic (Go's `json.Marshal` sorts map keys).

[Source: https://developer.hashicorp.com/vault/api-docs/secret/rabbitmq#create-role]

### Correct Fixed Code

**`convertVhostsToJson`:**
```go
func convertVhostsToJson(vhosts []Vhost) string {
	vhostData := make(map[string]any)
	for _, vhost := range vhosts {
		vhostData[vhost.VhostName] = vhost.Permissions
	}
	result, err := json.Marshal(vhostData)
	if err != nil {
		panic(fmt.Sprintf("failed to marshal vhost data: %v", err))
	}
	return string(result)
}
```

**`convertTopicsToJson`:**
```go
func convertTopicsToJson(vhosts []VhostTopic) string {
	vhostData := make(map[string]any)
	for _, vhost := range vhosts {
		topicData, _ := vhostData[vhost.VhostName].(map[string]any)
		if topicData == nil {
			topicData = make(map[string]any)
		}
		for _, topic := range vhost.Topics {
			topicData[topic.TopicName] = topic.Permissions
		}
		vhostData[vhost.VhostName] = topicData
	}
	result, err := json.Marshal(vhostData)
	if err != nil {
		panic(fmt.Sprintf("failed to marshal vhost topic data: %v", err))
	}
	return string(result)
}
```

### `log.Fatal` Remediation

Both functions currently import `"log"` (stdlib) and call `log.Fatal(err)` for JSON marshaling errors. `log.Fatal` calls `os.Exit(1)`, which would terminate the entire operator process — violating the project's error management pattern. Since `json.Marshal` of a `map[string]any` with string keys and simple struct values will never realistically fail, replace with `panic` (signals a programming error if it ever fires, and is recoverable by the controller-runtime recovery middleware). After the change, the `"log"` stdlib import should be removable — verify no other usages exist in the file.

Replace `"log"` import with `"fmt"` (for `fmt.Sprintf` in the panic message). The existing import of `"encoding/json"` stays.

### Anti-Pattern Prevention

- **DO NOT** use `convertVhostsToJson()` or `convertTopicsToJson()` to build expected values in tests — this only proves `x == x`. Build expected payloads independently.
- **DO NOT** change the Vault API payload format — `vhosts` and `vhost_topics` must remain JSON **strings** (not parsed maps).
- **DO NOT** change `rabbitMQToMap()` behavior — it calls the two fixed converter functions. The fix is entirely within `convertVhostsToJson` and `convertTopicsToJson`.
- **DO NOT** modify the CRD types (`RMQSERole`, `Vhost`, `VhostTopic`, `Topic`, `VhostPermissions`) — the bug is in serialization only.
- **DO NOT** modify the integration test fixture (`test/rabbitmqsecretengine/test-rmq-role.yaml`) — it correctly tests the single-entry case and must continue to work.

### Testing Strategy

**Unit tests** in `api/v1alpha1/rabbitmqsecretenginerole_test.go`:

1. **Multi-entry `convertVhostsToJson`:** Create 3 vhosts (`/`, `/staging`, `/production`) with different permissions. Call `convertVhostsToJson`. Unmarshal the result string into `map[string]any`. Assert all 3 vhost keys exist with correct permission values.

2. **Multi-entry `convertTopicsToJson`:** Create 2 vhostTopics (`/` with `amq.topic` + `amq.fanout`, `/staging` with `amq.direct`). Call `convertTopicsToJson`. Unmarshal the result. Assert both vhosts exist, and `/` contains both topic entries.

3. **Multi-entry `rabbitMQToMap`:** Construct an `RMQSERole` with multiple vhosts and vhostTopics. Call `rabbitMQToMap()`. Parse the `vhosts` and `vhost_topics` string values. Compare against independently-hardcoded expected JSON strings matching the Vault API format.

4. **Multi-entry `IsEquivalentToDesiredState`:** Construct a role with multiple vhosts. Build a Vault-read-shaped payload independently (hardcoded JSON strings). Assert `IsEquivalentToDesiredState` returns true for matching and false for mismatching payloads.

**Integration tests** — no changes needed. The existing test in `controllers/rabbitmqsecretengine_controller_test.go` with a single-entry fixture validates the end-to-end flow. Multi-entry correctness is covered by unit tests since the bug is in pure serialization logic, not in the controller or Vault interaction.

### Files to Modify

| File | Change |
|------|--------|
| `api/v1alpha1/rabbitmqsecretenginerole_types.go` | Fix `convertVhostsToJson` and `convertTopicsToJson` map reassignment bugs; replace `log.Fatal` with `panic`; replace `"log"` import with `"fmt"` |
| `api/v1alpha1/rabbitmqsecretenginerole_test.go` | Add multi-entry tests for `convertVhostsToJson`, `convertTopicsToJson`, `rabbitMQToMap`, and `IsEquivalentToDesiredState` |

**Files NOT modified:**
- `api/v1alpha1/rabbitmqsecretenginerole_webhook.go` — no serialization logic
- `controllers/rabbitmqsecretenginerole_controller.go` — no serialization logic
- `controllers/rabbitmqsecretengine_controller_test.go` — integration test, single-entry fixture is correct
- `test/rabbitmqsecretengine/test-rmq-role.yaml` — single-entry fixture, must continue to work
- CRD type structs — no changes needed
- No CRD regeneration needed (`make manifests generate` not required — no type changes)

### Previous Story Intelligence

**Epic 8 Retrospective (2026-07-20):** This bug was first identified during the Epic D3 retrospective (2026-07-05) and formalized as Story 9.0 during the Epic 8 retrospective. The retro also identified 3 immediate remediation items (broken webhook markers, copy-pasted loggers, mount path documentation) that were supposed to be done before Epic 9 — these are independent of this story but the dev should not attempt to fix them here (they are separate scope).

**Epic 8 Learnings:**
- All 4 Epic 8 stories completed in 4 days with zero production incidents
- Story intelligence chain worked well — each story documented what was deferred to the next
- Code review caught 8 findings across 4 stories
- Just-in-time version audit practice validated twice

### Project Structure Notes

- All changes are within `api/v1alpha1/` — the types package
- No new files created
- No controller changes needed
- No CRD changes needed (bug is in runtime serialization, not CRD schema)
- No webhook changes needed
- Follows existing test file organization (`*_test.go` alongside `*_types.go`)

### References

- [Source: api/v1alpha1/rabbitmqsecretenginerole_types.go:198-231 — buggy convertVhostsToJson and convertTopicsToJson]
- [Source: api/v1alpha1/rabbitmqsecretenginerole_types.go:232-238 — rabbitMQToMap calling the buggy functions]
- [Source: api/v1alpha1/rabbitmqsecretenginerole_test.go — existing single-entry tests]
- [Source: api/v1alpha1/payload_filter.go:7-15 — filterPayloadToDesiredKeys helper used by IsEquivalentToDesiredState]
- [Source: test/rabbitmqsecretengine/test-rmq-role.yaml — single-entry integration test fixture]
- [Source: controllers/rabbitmqsecretengine_controller_test.go — integration test for RabbitMQ secret engine]
- [Source: _bmad-output/implementation-artifacts/epic-8-retro-2026-07-20.md — bug tracking history]
- [Source: _bmad-output/planning-artifacts/epics.md — Epic 9, Story 9.0]
- [Source: _bmad-output/project-context.md — project rules and conventions]
- [Source: https://developer.hashicorp.com/vault/api-docs/secret/rabbitmq — Vault RabbitMQ API docs]

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6

### Debug Log References

None — implementation was straightforward with no debugging needed.

### Completion Notes List

- Fixed `convertVhostsToJson` map reassignment bug: changed `vhostData = map[string]any{...}` to `vhostData[vhost.VhostName] = vhost.Permissions` so all vhost entries are preserved across loop iterations.
- Fixed `convertTopicsToJson` with three changes: (1) moved `topicData` creation inside the outer loop to scope it per-vhost, (2) changed topic map reassignment to entry assignment, (3) changed vhost map reassignment to entry assignment.
- Replaced `log.Fatal(err)` with `panic(fmt.Sprintf(...))` in both converter functions to avoid `os.Exit(1)` killing the operator process. Replaced `"log"` import with `"fmt"`.
- Added 4 new multi-entry unit tests with independently-constructed expected values (not derived from code under test): `TestConvertVhostsToJsonMultipleEntries` (3 vhosts), `TestConvertTopicsToJsonMultipleEntries` (2 vhosts with 2+1 topics), `TestRMQSERoleRabbitMQToMapMultipleVhosts` (hardcoded expected JSON strings), `TestRabbitMQSecretEngineRoleIsEquivalentMultipleVhosts` (positive + negative matching with independently-constructed Vault-read-shaped payloads).
- All existing tests continue to pass — no regressions.
- `make test` passes (unit tests including new multi-entry tests).
- `make integration` passes (existing single-entry RabbitMQ fixture continues to work).

**Review fixes (2026-07-23):**
- `convertTopicsToJson` now merges repeated `vhostName` entries instead of overwriting, so topics from duplicate vhost items are additive.
- Replaced helper-derived expected values in `TestRMQSERoleRabbitMQToMap` with hardcoded JSON strings so the test no longer uses `convertVhostsToJson` / `convertTopicsToJson` to build expectations.
- `TestRMQSERoleRabbitMQToMapMultipleVhosts` now parses JSON and compares via `reflect.DeepEqual` instead of raw string equality.
- Added `"reflect"` import to the test file.

### File List

- `api/v1alpha1/rabbitmqsecretenginerole_types.go` (modified) — fixed `convertVhostsToJson` and `convertTopicsToJson` map reassignment bugs; replaced `log.Fatal` with `panic`; replaced `"log"` import with `"fmt"`; `convertTopicsToJson` merges duplicate `vhostName` entries
- `api/v1alpha1/rabbitmqsecretenginerole_test.go` (modified) — added 4 multi-entry unit tests with independently-constructed expected values; replaced helper-derived values in original test; multi-entry test uses parsed JSON comparison; added `"encoding/json"` and `"reflect"` imports

## Change Log

- 2026-07-21: Fixed RabbitMQ vhost/vhostTopic multi-entry serialization bug in `convertVhostsToJson` and `convertTopicsToJson`; replaced `log.Fatal` with `panic`; added 4 multi-entry unit tests
- 2026-07-23: Code review fixes — merge duplicate `vhostTopics` entries; replace helper-derived test expectations with hardcoded JSON; switch multi-entry test to parsed JSON comparison

### Review Findings

- [x] [Review][Patch] Merge repeated `vhostTopics` entries with the same `vhostName` during serialization so later items add topics instead of overwriting earlier ones [api/v1alpha1/rabbitmqsecretenginerole_types.go:210] — fixed
- [x] [Review][Dismiss] Align RabbitMQ equivalence handling and tests with actual Vault read payloads — dismissed as false positive; Vault RabbitMQ API returns `vhosts` and `vhost_topics` as JSON strings (not structured maps), confirmed by API docs and passing integration test
- [x] [Review][Patch] Replace helper-derived expected values in the original single-entry `rabbitMQToMap` test so it no longer builds expectations with `convertVhostsToJson` / `convertTopicsToJson` [api/v1alpha1/rabbitmqsecretenginerole_test.go:93] — fixed
- [x] [Review][Patch] Parse and compare the `vhosts` / `vhost_topics` JSON payloads in `TestRMQSERoleRabbitMQToMapMultipleVhosts` instead of asserting raw JSON string equality [api/v1alpha1/rabbitmqsecretenginerole_test.go:409] — fixed
