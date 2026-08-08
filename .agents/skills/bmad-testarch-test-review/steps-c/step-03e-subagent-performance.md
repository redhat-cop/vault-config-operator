---
name: 'step-03e-subagent-performance'
description: 'Subagent: Check test performance (speed, efficiency, parallelization)'
subagent: true
outputFile: '/tmp/tea-test-review-performance-{{timestamp}}.json'
---

# Subagent 3E: Performance Quality Check

## SUBAGENT CONTEXT

This is an **isolated subagent** running in parallel with other quality dimension checks.

**Your task:** Analyze test files for PERFORMANCE violations only.

---

## MANDATORY EXECUTION RULES

- ✅ Check PERFORMANCE only (not other quality dimensions)
- ✅ Read `criteria_registry` before evaluating anything; severities come from it
- ✅ Output structured JSON to temp file
- ❌ Do NOT check determinism, isolation, maintainability, or coverage
- ❌ Do NOT choose a severity or invent a row
- ❌ Do NOT emit a hard-wait violation; H1 belongs to the determinism worker

---

## SUBAGENT TASK

### 1. Identify Performance Violations

Evaluate exactly these registry rows and no others. Load
`{skill-root}/steps-c/criteria-registry.md` for each row's firing predicate, its
pinned severity, and its gate.

| Row | Criterion                       | Severity | Gate          |
| --- | ------------------------------- | -------: | ------------- |
| M1  | Network-first violated          |   MEDIUM | Applicability |
| M6  | Unawaited async                 |   MEDIUM | Absolute      |
| H5  | Oversize test file (>300 lines) |     HIGH | Absolute      |

**The hard-wait double count is fixed here.** This worker used to score
`waitForTimeout(5000)` as MEDIUM while the determinism worker scored the identical
line as HIGH and the published criteria table called it a FAIL. One `waitForTimeout`
could therefore deduct 5 and 2 from the same file, at two severities, in one run.
H1 now lives only in the determinism worker. Report a genuinely slow wait as
evidence in your notes, never as a second violation.

Four entries from the old list are deliberately gone:

- **"Slow setup/teardown (creating fresh DB for every test)"** describes correct
  isolation. Deducting for it pushed reviewers toward shared mutable state, which
  H4 then penalizes. The rubric was arguing with itself.
- **"Tests not parallelizable (`describe.serial`)"** is often the right call, and
  for pact suites it is mandatory (H6-H8 require serialization). A blanket
  deduction contradicted the contract-testing rules in the determinism worker.
- **"Missing performance optimizations"** and **"minor inefficiencies"** are
  unfalsifiable.
- **"Excessive logging"** is a style preference with no risk behind it.

Test duration is published in the criteria table but is not independently
measurable from a static read. Report `PASS` with the note that no excessive loops,
sleeps, or repeated navigation were found, or cite the specific M1/H1 evidence that
suggests otherwise. Never assert a measured runtime the run did not measure.

### 2. Calculate Performance Score

```javascript
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
const score = Math.max(0, 100 - totalPenalty);
```

---

## OUTPUT FORMAT

```json
{
  "dimension": "performance",
  "score": 90,
  "max_score": 100,
  "grade": "A",
  "violations": [
    {
      "file": "tests/e2e/search.spec.ts",
      "line": 10,
      "row": "M1",
      "severity": "MEDIUM",
      "category": "network-first-violated",
      "description": "Navigates and then reads result rows with no intercept or readiness signal registered first",
      "suggestion": "Register the search response before page.goto, then await it before asserting",
      "code_snippet": "await page.goto('/search'); await expect(page.getByRole('row')).toHaveCount(3);"
    },
    {
      "file": "tests/api/bulk-operations.spec.ts",
      "line": 35,
      "row": "M6",
      "severity": "MEDIUM",
      "category": "unawaited-async",
      "description": "A promise-returning call is neither awaited nor returned, so the assertion may run before the effect",
      "suggestion": "Await the call, or return the promise from the test",
      "code_snippet": "service.bulkCreate(payload); expect(repository.create).toHaveBeenCalled();"
    }
  ],
  "passed_checks": 13,
  "failed_checks": 2,
  "violation_summary": {
    "CRITICAL": 0,
    "HIGH": 0,
    "MEDIUM": 2,
    "LOW": 0
  },
  "performance_metrics": {
    "parallelizable_tests": 80,
    "serial_tests": 20,
    "avg_test_duration_estimate": "~2 seconds",
    "slow_tests": ["bulk-operations.spec.ts (>30s)"]
  },
  "recommendations": [
    "Enable parallel mode where possible",
    "Reduce setup data to minimum needed",
    "Use fixtures to share expensive setup across tests",
    "Remove unnecessary .serial constraints"
  ],
  "summary": "Good performance with 2 violations - 80% tests can run in parallel"
}
```

---

## EXIT CONDITION

Subagent completes when JSON output written to temp file.

**Subagent terminates here.**
