# Story R1.1: Correctness Fixes — Context Keys, PKI Bug, Webhook Loggers, Unsafe ToString

Status: done

## Story

As an operator developer,
I want correctness bugs fixed that can cause runtime panics or silent data corruption,
So that the operator is reliable before further refactoring.

## Acceptance Criteria

1. **Given** context values are passed using bare string keys (`"kubeClient"`, `"restConfig"`, `"vaultConnection"`, `"vaultClient"`) **When** replaced with typed context key constants and accessor functions **Then** a typo in a key name causes a compile error instead of a runtime panic
2. **Given** `VaultPKIEngineEndpoint.CreateOrUpdateConfig` accepts a `configPath` argument but always writes to `GetConfigCrlPath()` **When** the write path is changed to use the `configPath` argument **Then** `CreateOrUpdateConfigUrls` writes to the URLs endpoint (not the CRL endpoint)
3. **Given** `secretenginemount_webhook.go` `Default()` uses `authenginemountlog`, `databasesecretengineconfig_webhook.go` `Default()` uses `authenginemountlog`, and `azureauthengineconfig_webhook.go` `ValidateUpdate` uses `jwtoidcauthengineconfiglog` **When** each is replaced with the correct per-file logger variable **Then** log output attributes to the correct webhook
4. **Given** `utils.ToString(name interface{})` panics if `name` is not a string **When** replaced with a safe conversion (type switch or `fmt.Sprint`) **Then** non-string values produce a string instead of a panic
5. **Given** all changes **When** `make manifests generate fmt vet test` and `make integration` pass **Then** no regressions

## Tasks / Subtasks

- [x] Task 1: Define typed context key type and constants (AC: 1)
  - [x] 1.1: In `api/v1alpha1/utils/`, create an unexported `contextKey` type and 4 exported constants: `KubeClientKey`, `RestConfigKey`, `VaultConnectionKey`, `VaultClientKey`
  - [x] 1.2: Add 4 accessor functions: `KubeClientFromContext(ctx) client.Client`, `RestConfigFromContext(ctx) *rest.Config`, `VaultConnectionFromContext(ctx) *VaultConnection`, `VaultClientFromContext(ctx) *vault.Client`
  - [x] 1.3: Add 4 setter functions: `ContextWithKubeClient(ctx, client) context.Context`, `ContextWithRestConfig(ctx, cfg) context.Context`, `ContextWithVaultConnection(ctx, vc) context.Context`, `ContextWithVaultClient(ctx, vc) context.Context`
- [x] Task 2: Migrate all `context.WithValue` calls — 14 sites in 7 files (AC: 1)
  - [x] 2.1: `controllers/commons.go` (lines 20-23, 28) — 4 `WithValue` calls in `prepareContext`
  - [x] 2.2: `controllers/vaultsecret_controller.go` (lines 90-91) — 2 calls for `kubeClient`/`restConfig`
  - [x] 2.3: `controllers/vaultsecret_controller.go` (lines 337, 344) — 2 calls for `vaultConnection`/`vaultClient`
  - [x] 2.4: `controllers/vaultresourcecontroller/advanced-funcmap.go` (line 192) — 1 `restConfig` call
  - [x] 2.5: `api/v1alpha1/prepareinternalvalues_test_helpers_test.go` (lines 89-90, 96) — 3 calls across `pivContext` and `pivContextWithRestConfig`
  - [x] 2.6: `api/v1alpha1/kubernetesauthenginerole_test.go` (line 322) — 1 `kubeClient` call
  - [x] 2.7: `api/v1alpha1/utils/vaultobject_test.go` (line 85) — 1 `vaultClient` call in `newTestContext`
- [x] Task 3: Migrate all `context.Value("...")` calls — 43 sites in 18 files (AC: 1)
  - [x] 3.1: `api/v1alpha1/utils/commons.go` — 3 sites: `restConfig` (lines 142, 244), `vaultConnection` (line 281)
  - [x] 3.2: `api/v1alpha1/utils/vaultutils.go` — 4 sites: `vaultClient` (lines 36, 47, 66, 86)
  - [x] 3.3: `api/v1alpha1/utils/vaultobject.go` — 2 sites: `vaultClient` (lines 56, 76)
  - [x] 3.4: `api/v1alpha1/utils/vaultengineobject.go` — 1 site: `vaultClient` (line 50)
  - [x] 3.5: `api/v1alpha1/utils/vautlpkiengineobject.go` — 1 site: `vaultClient` (line 65)
  - [x] 3.6: `api/v1alpha1/utils/vaultauditobject.go` — 4 sites: `vaultClient` (lines 47, 67, 95, 120)
  - [x] 3.7: 16 `*_types.go` files — 26 sites total (see complete inventory below)
  - [x] 3.8: `controllers/vaultresourcecontroller/dynamicclientutils.go` — 2 sites: `restConfig` (lines 53, 83)
- [x] Task 4: Fix PKI `CreateOrUpdateConfig` write path (AC: 2)
  - [x] 4.1: In `api/v1alpha1/utils/vautlpkiengineobject.go` line 114: change `write(context, ve.vaultPKIEngineObject.GetConfigCrlPath(), payload)` to `write(context, configPath, payload)`
- [x] Task 5: Fix webhook logger copy-paste errors (AC: 3)
  - [x] 5.1: `api/v1alpha1/secretenginemount_webhook.go` line 47: change `authenginemountlog.Info(...)` to `secretenginemountlog.Info(...)`
  - [x] 5.2: `api/v1alpha1/databasesecretengineconfig_webhook.go` line 44: change `authenginemountlog.Info(...)` to `databasesecretengineconfiglog.Info(...)`
  - [x] 5.3: `api/v1alpha1/azureauthengineconfig_webhook.go` line 60: change `jwtoidcauthengineconfiglog.Info(...)` to `azureauthengineconfiglog.Info(...)`
- [x] Task 6: Replace `ToString` with safe conversion (AC: 4)
  - [x] 6.1: In `api/v1alpha1/utils/commons.go` (~line 354): replace unsafe `name.(string)` with `fmt.Sprintf("%v", name)` or a type switch
- [x] Task 7: Verify no regressions (AC: 5)
  - [x] 7.1: Run `make manifests generate fmt vet test`
  - [x] 7.2: Run `make integration`

## Dev Notes

### Bug 1: Context Keys — Full Call-Site Inventory

The epics doc estimates ~25 call sites; actual analysis found **57 total** (14 `WithValue` + 43 `.Value()`). All four keys must be migrated:

**Pattern for new code (in `api/v1alpha1/utils/`):**

```go
type contextKey int

const (
    KubeClientKey contextKey = iota
    RestConfigKey
    VaultConnectionKey
    VaultClientKey
)

func ContextWithKubeClient(ctx context.Context, c client.Client) context.Context {
    return context.WithValue(ctx, KubeClientKey, c)
}

func KubeClientFromContext(ctx context.Context) client.Client {
    return ctx.Value(KubeClientKey).(client.Client)
}
```

**Where to place these:** In `api/v1alpha1/utils/commons.go` (or a new `contextkeys.go` in the same package). This package is already imported by both `controllers/` and `api/v1alpha1/*_types.go`, so no new dependency edges are needed.

**Complete `.Value("...")` inventory for `*_types.go` files (26 sites):**

| File | Key | Line |
|------|-----|------|
| `policy_types.go` | `vaultClient` | 89 |
| `kubernetesauthenginerole_types.go` | `kubeClient` | 126 |
| `kubernetessecretengineconfig_types.go` | `kubeClient` | 148 |
| `kubernetessecretengineconfig_types.go` | `vaultClient` | 149 |
| `entityalias_types.go` | `vaultClient` | 198 |
| `entityalias_types.go` | `kubeClient` | 205 |
| `groupalias_types.go` | `vaultClient` | 177 |
| `groupalias_types.go` | `kubeClient` | 184 |
| `randomsecret_types.go` | `vaultClient` | 239 |
| `databasesecretengineconfig_types.go` | `kubeClient` | 148 |
| `databasesecretengineconfig_types.go` | `vaultClient` | 400 |
| `quaysecretengineconfig_types.go` | `kubeClient` | 99 |
| `ldapauthengineconfig_types.go` | `kubeClient` | 108 |
| `ldapauthengineconfig_types.go` | `kubeClient` | 172 |
| `jwtoidcauthengineconfig_types.go` | `kubeClient` | 227 |
| `azuresecretengineconfig_types.go` | `kubeClient` | 191 |
| `azureauthengineconfig_types.go` | `kubeClient` | 194 |
| `gcpauthengineconfig_types.go` | `kubeClient` | 207 |
| `rabbitmqsecretengineconfig_types.go` | `kubeClient` | 203 |
| `githubsecretengineconfig_types.go` | `kubeClient` | 135 |
| `githubsecretengineconfig_types.go` | `vaultClient` | 136 |
| `pkisecretengineconfig_types.go` | `kubeClient` | 284 |
| `pkisecretengineconfig_types.go` | `vaultClient` | 322 |
| `pkisecretengineconfig_types.go` | `kubeClient` | 327 |
| `pkisecretengineconfig_types.go` | `kubeClient` | 358 |
| `azureauthengineconfig_types.go` | `kubeClient` | 194 |

**Migration is mechanical:** search-and-replace each `context.WithValue(ctx, "kubeClient", ...)` → `vaultutils.ContextWithKubeClient(ctx, ...)` and each `context.Value("kubeClient").(client.Client)` → `vaultutils.KubeClientFromContext(context)`. Existing unsafe type assertions in the accessor functions preserve the current panic-on-nil contract (intentional — if context is not populated, the operator should panic early).

**Import alias:** This codebase imports `api/v1alpha1/utils` as `vaultutils` in controllers. In `*_types.go` files within the same `api/v1alpha1` package, the `utils` sub-package is imported directly — import it unqualified or as `vaultutils` depending on existing file convention.

### Bug 2: PKI CreateOrUpdateConfig

**File:** `api/v1alpha1/utils/vautlpkiengineobject.go` (note: typo in filename; rename is R1.3 scope, not this story)

**The bug (line 114):**
```go
func (ve *VaultPKIEngineEndpoint) CreateOrUpdateConfig(context context.Context, configPath string, payload map[string]interface{}) error {
    // ... reads from configPath correctly ...
    if !ve.vaultObject.IsEquivalentToDesiredState(currentConfigPayload) {
        return write(context, ve.vaultPKIEngineObject.GetConfigCrlPath(), payload)  // BUG: should be configPath
    }
    return nil
}
```

**Impact:** `CreateOrUpdateConfigUrls` (which passes `GetConfigUrlsPath()`) would write URL config to the CRL endpoint path when a drift is detected. This is a data-corruption bug.

**Fix:** Change `ve.vaultPKIEngineObject.GetConfigCrlPath()` to `configPath` on the write call. One-line fix.

### Bug 3: Webhook Logger Copy-Paste Errors

Three files, three one-line fixes:

| File | Line | Wrong Logger | Correct Logger |
|------|------|-------------|----------------|
| `secretenginemount_webhook.go` | 47 | `authenginemountlog` | `secretenginemountlog` |
| `databasesecretengineconfig_webhook.go` | 44 | `authenginemountlog` | `databasesecretengineconfiglog` |
| `azureauthengineconfig_webhook.go` | 60 | `jwtoidcauthengineconfiglog` | `azureauthengineconfiglog` |

Each file already declares the correct logger variable at the top — these are pure copy-paste errors where the log call references another file's logger.

### Bug 4: Unsafe ToString

**File:** `api/v1alpha1/utils/commons.go` (line ~354)

**Current code:**
```go
func ToString(name interface{}) string {
    if name != nil {
        return name.(string)  // PANICS on non-string
    }
    return ""
}
```

**Fix:** Replace with `fmt.Sprintf("%v", name)` which handles any type safely. Alternatively, use a type switch:
```go
func ToString(name interface{}) string {
    if name == nil {
        return ""
    }
    if s, ok := name.(string); ok {
        return s
    }
    return fmt.Sprintf("%v", name)
}
```

The type-switch approach is preferred because it avoids unnecessary `Sprintf` overhead for the common string case while still being safe for non-strings.

### Project Structure Notes

- All context key infrastructure goes in `api/v1alpha1/utils/` — this is the established home for cross-cutting Vault and K8s client utilities
- No new packages needed — existing import graph (`controllers/ → api/v1alpha1/utils/`) is preserved
- Do NOT rename `vautlpkiengineobject.go` (that's Story R1.3 scope)
- Do NOT touch `toMap()`, `IsEquivalentToDesiredState()`, or webhook validation logic beyond the logger fix
- The `make manifests generate` step may produce zero changes since this story touches no `*_types.go` Spec/Status structs or RBAC markers, but run it anyway as the verification gate

### Testing Strategy

- No new test files needed — the context key migration is a compile-time safety improvement
- Existing unit tests (`make test`) and integration tests (`make integration`) are the regression gate
- The PKI fix will be exercised by existing PKI integration tests (Story 5.3 infrastructure)
- Webhook logger changes are cosmetic (log attribution) — no behavioral test needed

### Previous Story Intelligence (Epic 7.5)

This is the first story in Epic R1. Key learnings from Epic 7.5 (the immediately preceding epic):

- **`make integration` takes ~576-579s** — budget accordingly
- **Kind cluster can degrade** (terminating namespaces) — if tests fail unexpectedly, try a fresh cluster
- **Run `make manifests generate` even for non-type changes** — catches unexpected diffs
- **The project-context.md** is the ground truth for coding patterns; this story's context key fix should be reflected there after completion
- **Epic 7.5 retro noted** the PKI `CreateOrUpdateConfig` bug as carried debt — this story fixes it

### References

- [Source: _bmad-output/planning-artifacts/epics.md#Story R1.1] — acceptance criteria and task list
- [Source: _bmad-output/planning-artifacts/epics.md#Epic R1] — epic ordering (R1.1 first, then R1.2a)
- [Source: _bmad-output/project-context.md#Context-Carried Values] — documents the 4 context keys and current unsafe pattern
- [Source: _bmad-output/project-context.md#Anti-Patterns to Avoid] — documents the logger and condition rules
- [Source: api/v1alpha1/utils/vautlpkiengineobject.go:105-118] — PKI write-path bug
- [Source: _bmad-output/implementation-artifacts/epic-7-5-retro-2026-05-12.md] — PKI bug noted as carried debt

## Dev Agent Record

### Agent Model Used

Opus 4.6 (Cursor Agent)

### Debug Log References

None — clean implementation, no debug cycles needed.

### Completion Notes List

- **AC1 (Context Keys):** Created `api/v1alpha1/utils/contextkeys.go` with typed `contextKey` int type, 4 exported constants, 4 getter functions, and 4 setter functions. Migrated all 14 `context.WithValue` call sites and all 43 `context.Value("...")` call sites across 25 files. Removed 8 now-unused imports (`vault`, `rest`). A context key typo will now cause a compile error instead of a runtime panic.
- **AC2 (PKI Bug):** Fixed `CreateOrUpdateConfig` in `vautlpkiengineobject.go` to write to `configPath` instead of hardcoded `GetConfigCrlPath()`. This was a data-corruption bug where URL config would overwrite CRL config on drift detection.
- **AC3 (Webhook Loggers):** Fixed 3 copy-paste logger errors: `secretenginemount_webhook.go`, `databasesecretengineconfig_webhook.go`, and `azureauthengineconfig_webhook.go` now use their own per-file logger variable.
- **AC4 (ToString):** Replaced unsafe `name.(string)` type assertion with type-switch approach: fast path for strings, `fmt.Sprintf` fallback for non-strings. Non-string values now produce a string instead of panicking.
- **AC5 (No Regressions):** `make manifests generate fmt vet test` and `make integration` both pass (integration: 576s).

### Change Log

- 2026-05-15: Implemented all 7 tasks for story R1.1 — typed context keys, PKI write-path fix, webhook logger fixes, safe ToString conversion. All unit and integration tests pass.

### File List

- api/v1alpha1/utils/contextkeys.go (NEW)
- api/v1alpha1/utils/commons.go (MODIFIED — removed bare-string context value lookups, safe ToString)
- api/v1alpha1/utils/vaultutils.go (MODIFIED — typed context accessors)
- api/v1alpha1/utils/vaultobject.go (MODIFIED — typed context accessors)
- api/v1alpha1/utils/vaultobject_test.go (MODIFIED — typed context setter)
- api/v1alpha1/utils/vaultengineobject.go (MODIFIED — typed context accessors, removed vault import)
- api/v1alpha1/utils/vautlpkiengineobject.go (MODIFIED — typed context accessors, PKI write-path fix)
- api/v1alpha1/utils/vaultauditobject.go (MODIFIED — typed context accessors)
- api/v1alpha1/policy_types.go (MODIFIED — typed context accessors, removed vault import)
- api/v1alpha1/randomsecret_types.go (MODIFIED — typed context accessors, removed vault import)
- api/v1alpha1/entityalias_types.go (MODIFIED — typed context accessors, removed vault import)
- api/v1alpha1/groupalias_types.go (MODIFIED — typed context accessors, removed vault import)
- api/v1alpha1/kubernetesauthenginerole_types.go (MODIFIED — typed context accessors)
- api/v1alpha1/kubernetessecretengineconfig_types.go (MODIFIED — typed context accessors, removed vault import)
- api/v1alpha1/databasesecretengineconfig_types.go (MODIFIED — typed context accessors, removed vault import)
- api/v1alpha1/pkisecretengineconfig_types.go (MODIFIED — typed context accessors)
- api/v1alpha1/githubsecretengineconfig_types.go (MODIFIED — typed context accessors, removed vault import)
- api/v1alpha1/quaysecretengineconfig_types.go (MODIFIED — typed context accessors)
- api/v1alpha1/ldapauthengineconfig_types.go (MODIFIED — typed context accessors)
- api/v1alpha1/jwtoidcauthengineconfig_types.go (MODIFIED — typed context accessors)
- api/v1alpha1/azureauthengineconfig_types.go (MODIFIED — typed context accessors)
- api/v1alpha1/gcpauthengineconfig_types.go (MODIFIED — typed context accessors)
- api/v1alpha1/azuresecretengineconfig_types.go (MODIFIED — typed context accessors)
- api/v1alpha1/rabbitmqsecretengineconfig_types.go (MODIFIED — typed context accessors)
- api/v1alpha1/secretenginemount_webhook.go (MODIFIED — logger fix)
- api/v1alpha1/databasesecretengineconfig_webhook.go (MODIFIED — logger fix)
- api/v1alpha1/azureauthengineconfig_webhook.go (MODIFIED — logger fix)
- api/v1alpha1/prepareinternalvalues_test_helpers_test.go (MODIFIED — typed context setters)
- api/v1alpha1/kubernetesauthenginerole_test.go (MODIFIED — typed context setter)
- controllers/commons.go (MODIFIED — typed context setters)
- controllers/vaultsecret_controller.go (MODIFIED — typed context setters)
- controllers/vaultresourcecontroller/advanced-funcmap.go (MODIFIED — typed context setter, added vaultutils import)
- controllers/vaultresourcecontroller/dynamicclientutils.go (MODIFIED — typed context accessors, added vaultutils import, removed rest import)
