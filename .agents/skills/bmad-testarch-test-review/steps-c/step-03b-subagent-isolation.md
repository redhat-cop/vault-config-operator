---
name: 'step-03b-subagent-isolation'
description: 'Subagent: Check test isolation (no shared state/dependencies)'
subagent: true
outputFile: '/tmp/tea-test-review-isolation-{{timestamp}}.json'
---

# Subagent 3B: Isolation Quality Check

## SUBAGENT CONTEXT

This is an **isolated subagent** running in parallel with other quality dimension checks.

**Your task:** Analyze test files for ISOLATION violations only.

---

## MANDATORY EXECUTION RULES

- ✅ Check ISOLATION only (not other quality dimensions)
- ✅ Read `criteria_registry` before evaluating anything; severities come from it
- ✅ Output structured JSON to temp file
- ❌ Do NOT check determinism, maintainability, coverage, or performance
- ❌ Do NOT modify test files (read-only analysis)
- ❌ Do NOT choose a severity or invent a row

---

## SUBAGENT TASK

### 1. Identify Isolation Violations

Evaluate exactly these registry rows and no others. Load
`{skill-root}/steps-c/criteria-registry.md` for each row's firing predicate, its
pinned severity, and its gate.

| Row | Criterion                    | Severity | Gate     |
| --- | ---------------------------- | -------: | -------- |
| C5  | Mock asserted against itself | CRITICAL | Absolute |
| H4  | Unreset shared state         |     HIGH | Absolute |
| M4  | Ungrouped suite              |   MEDIUM | Absolute |

**H4 covers every shape the old prose list spread across three tiers**: a mutated
global, an order dependency, a shared record with no cleanup, a leaking
`beforeAll`/`afterAll`, an unrestored environment variable, a mutating shared
fixture. Each is the same defect, state surviving a test, and each makes the result
depend on execution order. The old list scored that one defect HIGH, MEDIUM, or LOW
depending on which sentence a reviewer happened to match it against, which is
exactly the variance this registry removes.

Two entries from the old list are deliberately gone:

- **"Tests sharing test data (but not mutating)"** is not a defect. Shared immutable
  data is what a fixture is for.
- **"Tests that could be more isolated"** is unfalsifiable. Every suite could be
  more isolated, so a row that always fires carries no information.

**M4 is published under maintainability but detectable here.** Emit it once; the
aggregation step deduplicates by `(file, line, row)` when two workers both find it.

### 2. Calculate Isolation Score

```javascript
const totalChecks = testFiles.length * checksPerFile;
const failedChecks = violations.length;
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
  "dimension": "isolation",
  "score": 90,
  "max_score": 100,
  "grade": "A-",
  "violations": [
    {
      "file": "tests/api/integration.spec.ts",
      "line": 15,
      "row": "H4",
      "severity": "HIGH",
      "category": "unreset-shared-state",
      "description": "Test reads a user record the previous test created, so the result depends on test order",
      "suggestion": "Create the record this test needs in beforeEach, and reset it after",
      "code_snippet": "test('updates the user', async () => { /* assumes the record from the test above */ });"
    }
  ],
  "passed_checks": 14,
  "failed_checks": 1,
  "total_checks": 15,
  "violation_summary": {
    "HIGH": 1,
    "MEDIUM": 0,
    "LOW": 0
  },
  "recommendations": [
    "Add beforeEach hooks to create test data",
    "Add afterEach hooks to cleanup created records",
    "Use test.describe.configure({ mode: 'parallel' }) to enforce isolation"
  ],
  "summary": "Tests are well isolated with 1 HIGH severity violation"
}
```

---

## EXIT CONDITION

Subagent completes when:

- ✅ All test files analyzed for isolation violations
- ✅ Score calculated
- ✅ JSON output written to temp file

**Subagent terminates here.**

---

## 🚨 SUBAGENT SUCCESS METRICS

### ✅ SUCCESS:

- Only isolation checked (not other dimensions)
- JSON output valid and complete

### ❌ FAILURE:

- Checked quality dimensions other than isolation
- Invalid or missing JSON output
