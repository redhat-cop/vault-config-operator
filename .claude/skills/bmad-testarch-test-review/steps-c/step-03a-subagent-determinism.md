---
name: 'step-03a-subagent-determinism'
description: 'Subagent: Check test determinism (no random/time dependencies)'
subagent: true
outputFile: '/tmp/tea-test-review-determinism-{{timestamp}}.json'
---

# Subagent 3A: Determinism Quality Check

## SUBAGENT CONTEXT

This is an **isolated subagent** running in parallel with other quality dimension checks.

**What you have from parent workflow:**

- Test files discovered in Step 2
- Knowledge fragment: test-quality (determinism criteria)
- Config: test framework

**Your task:** Analyze test files for DETERMINISM violations only.

---

## MANDATORY EXECUTION RULES

- 📖 Read this entire subagent file before acting
- ✅ Check DETERMINISM only (not other quality dimensions)
- ✅ Read `criteria_registry` before evaluating anything; severities come from it
- ✅ Output structured JSON to temp file
- ❌ Do NOT check isolation, maintainability, coverage, or performance (other subagents)
- ❌ Do NOT modify test files (read-only analysis)
- ❌ Do NOT run tests (just analyze code)
- ❌ Do NOT choose a severity or invent a row

---

## SUBAGENT TASK

### 1. Identify Determinism Violations

Evaluate exactly these registry rows and no others. Load
`{skill-root}/steps-c/criteria-registry.md` for each row's firing predicate, its
pinned severity, and its gate. This worker owns every `CRITICAL` row except C5.

| Row | Criterion                                 | Severity | Gate          |
| --- | ----------------------------------------- | -------: | ------------- |
| C1  | Disabled test (`.skip`, `xit`, `@Ignore`) | CRITICAL | Absolute      |
| C2  | Focused test (`.only`, `fit`)             | CRITICAL | Absolute      |
| C3  | Tautological assertion                    | CRITICAL | Absolute      |
| C4  | No assertion                              | CRITICAL | Absolute      |
| C6  | Assertion unreachable                     | CRITICAL | Absolute      |
| H1  | Hard wait                                 |     HIGH | Absolute      |
| H2  | Wall-clock fixture                        |     HIGH | Applicability |
| H3  | Conditional assertion                     |     HIGH | Absolute      |
| H6  | Pact worker parallelism                   |     HIGH | Absolute      |
| H7  | Pact pool isolation (≥2 pacttest files)   |     HIGH | Applicability |
| H8  | Pact serialization defeated               |     HIGH | Absolute      |
| L4  | Pact single-file pool advisory            |      LOW | Applicability |

**Three scoring conflicts this replaces, all of which produced different numbers
for identical code depending on which worker saw it first:**

- **`waitForTimeout` was MEDIUM here** while the published criteria table calls Hard
  Waits a `❌ FAIL` and the performance worker scored the same line MEDIUM again. One
  timer could deduct twice at two severities. H1 is HIGH, owned here alone.
- **Shared state was LOW here and HIGH in the isolation worker.** It is H4 and it
  belongs to isolation. Do not emit it.
- **Test order dependency was MEDIUM here and HIGH in isolation.** Same defect, same
  fix: it is H4, and it is not yours.

**`Date.now()` is no longer an unconditional HIGH.** Judged as written, it fired on
every test that stamped a timestamp for a value nothing depended on. H2 gates it on
what the value governs: an expiry, token lifetime, TTL, or scheduling boundary,
which is where a wall-clock race actually costs something. A timestamp used as
opaque test data is not a violation. Say which case you found.

**Non-determinism with no registry row** (`Math.random()` without a seed, an
unmocked external call, a filesystem write to a random path, an unordered database
read) is real and worth reporting. Report it in prose with the risk named, and
without a severity or a deduction, until it earns a row. A finding whose severity
you invented cannot be compared against the same finding next week.

### 1b. Pact and contract-testing predicates

These carry their own predicates and severities, already deterministic, and they
stay verbatim. H6, H7, H8 and L4 above are their registry identities.

- **PactV4 consumer tests: multiple `pact.addInteraction()` in a single `it()` block** — the Rust FFI non-deterministically drops interactions (see `pactjs-utils-consumer-helpers.md` Example 6). Flag any `.pacttest.ts` file where a single `it()`/`test()` contains more than one `addInteraction()` chain.
- **PactV4 consumer Vitest config missing `fileParallelism: false`** in `vitest.config.pact.ts` — parallel workers race on the shared pact JSON file (see `pact-consumer-framework-setup.md` Example 2). HIGH regardless of file count.
- **PactV4 consumer Vitest config missing `pool: 'forks'` + `poolOptions.forks.singleFork: true`** in `vitest.config.pact.ts` — best current understanding is that the `@pact-foundation/pact` napi-rs binding is not robust across Vitest worker threads sharing a process; once a consumer+provider pair has ≥2 `.pacttest.ts` files, default threads pool produces reproducible "request was expected but not received" flakes on Linux CI. **Severity: HIGH if the repo has ≥2 `.pacttest.ts` files for the same consumer+provider pair; LOW (future-proof advisory) for single-file suites.** See `pact-consumer-framework-setup.md` Example 2.
- **Pact provider Vitest config missing `pool: 'forks'` + `poolOptions.forks.singleFork: true`** in `vitest.config.contract.ts` for multi-file provider suites (especially message providers) — same pool rule as the consumer side (see `pactjs-utils-provider-verifier.md` Example 7).
- **Consumer or provider Vitest config sets any of: `sequence.concurrent: true`, `maxConcurrency > 1`, `maxWorkers > 1`, `isolate: false`** in `vitest.config.pact.ts` / `vitest.config.contract.ts` — each defeats the serialization the forks-singleFork rule relies on. HIGH.
- **Consumer repo lacks a determinism gate** — if `tea_use_pactjs_utils` is enabled, flag any `package.json` whose `test:pact:consumer` script does not run `scripts/check-pact-determinism.sh` (see `pact-consumer-framework-setup.md` Example 10).

### 2. Analyze Each Test File

For each test file from Step 2:

Every violation carries the registry `row` that produced it, and its `severity` is
copied from that row rather than written by hand. The aggregation step rejects a
violation with no `row`, because an unattributed severity is one somebody chose.

```javascript
const violations = [];

// Every predicate below reports the line of the syntax that ACTUALLY matched,
// never a hardcoded literal. A file skipped with `xit` and a file skipped with
// `.skip` are the same row; a violation citing a line the reader cannot find is
// a finding they have to re-derive by hand.

// C1 — a skipped test, in every form the registry row names. The most expensive
// row in the registry: the suite reports green while the case nobody wants to
// lose goes unrun.
const skipMatch = /\b(?:test|it|describe)\.(?:skip|todo)\b|\bxit\b|\bxdescribe\b|@Ignore\b|@Disabled\b|pytest\.mark\.skip\w*/.exec(
  testFileContent,
);
// A skip carrying a documented reason that is still true is exempt: check the
// matched line and the line above it for one before reporting.
if (skipMatch && !hasStillTrueReason(testFileContent, skipMatch)) {
  violations.push({
    file: testFile,
    line: findLineNumber(skipMatch[0]),
    row: 'C1',
    severity: 'CRITICAL', // from the registry, not chosen here
    category: 'disabled-test',
    description: `Test is disabled with \`${skipMatch[0]}\` and no still-true reason recorded`,
    suggestion: 'Re-enable it, or record why it is skipped and what re-enables it',
  });
}

// C3 — an assertion that cannot fail. Both shapes the registry row names: a
// literal compared to itself, and any operand compared to itself.
const selfComparison =
  /expect\(\s*([^()]+?)\s*\)\.(?:toBe|toEqual)\(\s*\1\s*\)/.exec(testFileContent) ??
  /\bassert\s+([A-Za-z_$][\w.$]*)\s*==\s*\1\b/.exec(testFileContent);
if (selfComparison) {
  violations.push({
    file: testFile,
    line: findLineNumber(selfComparison[0]),
    row: 'C3',
    severity: 'CRITICAL',
    category: 'tautological-assertion',
    description: `\`${selfComparison[0]}\` compares a value to itself, so it can never fail`,
    suggestion: 'Assert the behavior the test name claims',
  });
}

// H1 — hard wait. HIGH, and owned by this worker alone: the performance worker
// must not emit it again, or one timer deducts twice.
const hardWait = /waitForTimeout|\bsleep\(|time\.sleep\(|Thread\.sleep\(|cy\.wait\(\s*\d/.exec(testFileContent);
if (hardWait) {
  violations.push({
    file: testFile,
    line: findLineNumber(hardWait[0]),
    row: 'H1',
    severity: 'HIGH',
    category: 'hard-wait',
    description: `A bare timer (\`${hardWait[0]}\`) orders steps instead of a condition`,
    suggestion: 'Await the observable state: expect(locator).toBeVisible(), or a network-first intercept',
  });
}

// H2 — wall-clock fixture. GATED: only when the value governs an expiry, token
// lifetime, TTL, or scheduling boundary. A timestamp used as opaque test data is
// not a violation, which is why the old unconditional Date.now() check over-fired.
const wallClock = /Date\.now\(\)|new Date\(\s*\)|time\.time\(\)/.exec(testFileContent);
if (wallClock && governsATimeBoundary(testFileContent) && !usesFakeTimers(testFileContent)) {
  violations.push({
    file: testFile,
    line: findLineNumber(wallClock[0]),
    row: 'H2',
    severity: 'HIGH',
    category: 'time-dependency',
    description: `An expiry or lifetime is derived from the live clock (\`${wallClock[0]}\`) with no fake timers`,
    suggestion: 'Freeze time (vi.useFakeTimers / setSystemTime) and add explicit expired and still-valid boundary cases',
  });
}

// Math.random(), unmocked external calls, and random filesystem paths have no
// registry row yet. Report them in prose with the risk named, and with no severity
// and no deduction, rather than inventing a tier for them here.
```

**Detecting Pact Vitest config violations (`vitest.config.pact.ts` / `vitest.config.contract.ts`)**

Vitest configs vary widely — `defineConfig({ test: { ... } })`, `mergeConfig(base, overrides)`, `satisfies UserConfig`, imported constants, TS spreads. A full AST parse is out of scope; use this fallback heuristic and accept false-negatives only for the `mergeConfig` case, which the subagent must flag separately:

```javascript
// Resolve the config file(s). For consumer: scripts.test:pact:consumer:run in package.json
// usually points at `vitest run --config <path>`. For provider: `vitest run --config <path>`.
// If neither script exists but `.pacttest.ts` files exist, default to 'vitest.config.pact.ts'.
const configPath = resolveVitestConfigPath({ scriptName: 'test:pact:consumer:run', fallback: 'vitest.config.pact.ts' });
const src = fs.readFileSync(configPath, 'utf8');

// 1. Literal-match the two mandatory lines. Tolerate single or double quotes and whitespace.
const hasFileParallelismFalse = /\bfileParallelism\s*:\s*false\b/.test(src);
const hasPoolForks = /\bpool\s*:\s*['"]forks['"]/.test(src);
const hasSingleForkTrue = /\bsingleFork\s*:\s*true\b/.test(src);

// 2. Flag settings that would defeat the rule if a human added them.
const hasSequenceConcurrent = /\bsequence\s*:\s*\{[^}]*\bconcurrent\s*:\s*true/.test(src);
const hasHighMaxConcurrency = /\bmaxConcurrency\s*:\s*([2-9]|\d{2,})/.test(src);
const hasHighMaxWorkers = /\bmaxWorkers\s*:\s*([2-9]|\d{2,})/.test(src);
const hasIsolateFalse = /\bisolate\s*:\s*false\b/.test(src);

// 3. mergeConfig / extends fallback — we cannot reliably follow imports. Emit LOW advisory.
const usesMergeConfig = /\bmergeConfig\s*\(/.test(src) || /\bextends\s*:/.test(src);

// 4. File-count gating for the pool-forks rule.
const pactTestCount = glob.sync('tests/contract/**/*.pacttest.ts').length;
```

**Violation emission rules** (apply in order; exit on first match per check):

- Missing `fileParallelism: false` → HIGH (always)
- Missing `pool: 'forks'` OR missing `singleFork: true`, AND `pactTestCount >= 2` → HIGH
- Missing `pool: 'forks'` OR missing `singleFork: true`, AND `pactTestCount < 2` → LOW (future-proof advisory)
- Any of `sequence.concurrent: true`, `maxConcurrency > 1`, `maxWorkers > 1`, `isolate: false` present → HIGH
- `usesMergeConfig` AND any of the three mandatory matches missing → LOW + `category: "pact-config-unverifiable"` with a suggestion to inline the pool settings at the leaf config or provide a `// tea:pact-ffi-safe` marker comment the subagent can trust

### 3. Calculate Determinism Score

**Scoring Logic**:

```javascript
const totalChecks = testFiles.length * checksPerFile;
const failedChecks = violations.length;
const passedChecks = totalChecks - failedChecks;

// Weight violations by severity
// CRITICAL is present because the registry now defines CRITICAL rows. Without the
// key, `sum + undefined` makes this dimension score NaN the first time a reviewer
// finds a skipped test. This per-dimension number is informational; step-03f's
// deduction ledger remains the authoritative score.
const severityWeights = { CRITICAL: 20, HIGH: 10, MEDIUM: 5, LOW: 2 };
const totalPenalty = violations.reduce((sum, v) => {
  const weight = severityWeights[v.severity];
  if (weight === undefined) throw new Error(`unknown severity "${v.severity}" on ${v.row ?? 'an unattributed violation'}`);
  return sum + weight;
}, 0);

// Score: 100 - (penalty points)
const score = Math.max(0, 100 - totalPenalty);
```

---

## OUTPUT FORMAT

Write JSON to temp file: `/tmp/tea-test-review-determinism-{{timestamp}}.json`

```json
{
  "dimension": "determinism",
  "score": 85,
  "max_score": 100,
  "grade": "B",
  "violations": [
    {
      "file": "tests/api/user.spec.ts",
      "line": 42,
      "severity": "HIGH",
      "category": "random-generation",
      "description": "Test uses Math.random() - non-deterministic",
      "suggestion": "Use faker.seed(12345) for deterministic random data",
      "code_snippet": "const userId = Math.random() * 1000;"
    },
    {
      "file": "tests/e2e/checkout.spec.ts",
      "line": 78,
      "severity": "MEDIUM",
      "category": "hard-wait",
      "description": "Test uses waitForTimeout - creates flakiness",
      "suggestion": "Replace with expect(locator).toBeVisible()",
      "code_snippet": "await page.waitForTimeout(5000);"
    }
  ],
  "passed_checks": 12,
  "failed_checks": 3,
  "total_checks": 15,
  "violation_summary": {
    "HIGH": 1,
    "MEDIUM": 1,
    "LOW": 1
  },
  "recommendations": [
    "Use faker with fixed seed for all random data",
    "Replace all waitForTimeout with conditional waits",
    "Mock Date.now() in tests that use current time"
  ],
  "summary": "Tests are mostly deterministic with 3 violations (1 HIGH, 1 MEDIUM, 1 LOW)"
}
```

**On Error:**

```json
{
  "dimension": "determinism",
  "success": false,
  "error": "Error message describing what went wrong"
}
```

---

## EXIT CONDITION

Subagent completes when:

- ✅ All test files analyzed for determinism violations
- ✅ Score calculated (0-100)
- ✅ Violations categorized by severity
- ✅ Recommendations generated
- ✅ JSON output written to temp file

**Subagent terminates here.** Parent workflow will read output and aggregate with other quality dimensions.

---

## 🚨 SUBAGENT SUCCESS METRICS

### ✅ SUCCESS:

- All test files scanned for determinism violations
- Score calculated with proper severity weighting
- JSON output valid and complete
- Only determinism checked (not other dimensions)

### ❌ FAILURE:

- Checked quality dimensions other than determinism
- Invalid or missing JSON output
- Score calculation incorrect
- Modified test files (should be read-only)
