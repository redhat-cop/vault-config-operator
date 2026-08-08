---
name: 'step-03c-subagent-maintainability'
description: 'Subagent: Check test maintainability (readability, structure, DRY)'
subagent: true
outputFile: '/tmp/tea-test-review-maintainability-{{timestamp}}.json'
---

# Subagent 3C: Maintainability Quality Check

## SUBAGENT CONTEXT

This is an **isolated subagent** running in parallel with other quality dimension checks.

**Your task:** Analyze test files for MAINTAINABILITY violations only.

---

## MANDATORY EXECUTION RULES

- ✅ Check MAINTAINABILITY only (not other quality dimensions)
- ✅ Read `criteria_registry` before evaluating anything; severities come from it
- ✅ Score Convention rows against `convention_baseline`, never against an absolute standard
- ✅ Output structured JSON to temp file
- ❌ Do NOT check determinism, isolation, coverage, or performance
- ❌ Do NOT choose a severity, invent a row, or step a severity outside the registry's Convention schedule

---

## SUBAGENT TASK

### 1. Identify Maintainability Violations

Evaluate exactly these registry rows and no others. Load
`{skill-root}/steps-c/criteria-registry.md` for each row's firing predicate, its
pinned severity, and its gate.

| Row | Criterion                       | Severity | Gate                         |
| --- | ------------------------------- | -------: | ---------------------------- |
| M2  | Repeated literal payload        |   MEDIUM | Applicability                |
| M3  | Multi-concern test              |   MEDIUM | Absolute                     |
| M4  | Ungrouped suite                 |   MEDIUM | Absolute                     |
| M5  | Low-level event dispatch        |   MEDIUM | Applicability                |
| M7  | Excessive nesting               |   MEDIUM | Absolute                     |
| H5  | Oversize test file (>300 lines) |     HIGH | Absolute                     |
| L1  | Fragile selector                |      LOW | Applicability                |
| L3  | Missing stable test id          |      LOW | Convention: `testIds`        |
| L5  | Implementation-shaped name      |      LOW | Convention: `bddNaming`      |
| L6  | Magic value                     |      LOW | Absolute                     |
| L7  | Inconsistent assertion style    |      LOW | Convention: `assertionStyle` |

Three rules this dimension used to get wrong, now fixed by the registry:

- **The 300-line threshold is the only length rule.** The old list deducted HIGH
  for "tests >100 lines", which contradicted both the published criteria table
  (`Test Length (≤300 lines)`) and the template. One threshold, one row: H5.
- **Naming and test ids are Convention rows.** A repo with no behavioral-naming
  convention and no test-id convention takes no deduction for either, and the
  report says `PASS (n/a)` with the adoption count. A role- or label-based locator
  satisfies L1 outright; it is not a missing test id.
- **"Could benefit from helper functions" and "minor code style issues" are gone.**
  Neither was falsifiable, so neither could be scored the same way twice. A real
  defect that matches no row goes in prose with no severity and no deduction.

### 2. Calculate Maintainability Score

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
  "dimension": "maintainability",
  "score": 90,
  "max_score": 100,
  "grade": "A",
  "violations": [
    {
      "file": "tests/e2e/complex-flow.spec.ts",
      "line": 1,
      "row": "H5",
      "severity": "HIGH",
      "category": "oversize-test-file",
      "description": "File is 341 lines, over the 300-line threshold",
      "suggestion": "Split by feature area. 250 lines would NOT fire this row; the threshold is 300 and there is only one",
      "code_snippet": "test.describe('Complex flow', () => { /* 341 lines */ });"
    }
  ],
  "passed_checks": 10,
  "failed_checks": 1,
  "violation_summary": {
    "CRITICAL": 0,
    "HIGH": 1,
    "MEDIUM": 0,
    "LOW": 0
  },
  "recommendations": [
    "Split large test files into smaller, focused files (<100 lines each)",
    "Add test.describe grouping for related tests",
    "Extract duplicate logic into helper functions"
  ],
  "summary": "1 maintainability violation (1 HIGH)"
}
```

---

## EXIT CONDITION

Subagent completes when JSON output written to temp file.

**Subagent terminates here.**
