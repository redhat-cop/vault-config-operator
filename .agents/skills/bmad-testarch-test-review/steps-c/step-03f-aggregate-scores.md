---
name: 'step-03f-aggregate-scores'
description: 'Aggregate quality dimension scores into overall 0-100 score'
nextStepFile: '{skill-root}/steps-c/step-04-generate-report.md'
outputFile: '{test_artifacts}/test-review.md'
---

# Step 3F: Aggregate Quality Scores

## STEP GOAL

Read outputs from 4 quality subagents, aggregate violations by severity, and calculate the overall score (0-100) from the deduction ledger for report generation.

---

## MANDATORY EXECUTION RULES

- 📖 Read the entire step file before acting
- ✅ Speak in `{communication_language}`
- ✅ Read all 4 subagent outputs
- ✅ Aggregate violations by severity
- ✅ Calculate the overall score from the deduction ledger, never a weighted average
- ❌ Do NOT re-evaluate quality (use subagent outputs)

---

## EXECUTION PROTOCOLS:

- 🎯 Follow the MANDATORY SEQUENCE exactly
- 💾 Record outputs before proceeding
- 📖 Load the next step only when instructed

---

## MANDATORY SEQUENCE

### 1. Read All Subagent Outputs

```javascript
// Use the SAME timestamp generated in Step 3 (do not regenerate).
const timestamp = subagentContext?.timestamp;
if (!timestamp) {
  throw new Error('Missing timestamp from Step 3 context. Pass Step 3 timestamp into Step 3F.');
}
const dimensions = ['determinism', 'isolation', 'maintainability', 'performance'];
const results = {};

dimensions.forEach((dim) => {
  const outputPath = `/tmp/tea-test-review-${dim}-${timestamp}.json`;
  results[dim] = JSON.parse(fs.readFileSync(outputPath, 'utf8'));
});
```

**Verify all succeeded:**

```javascript
const allSucceeded = dimensions.every((dim) => results[dim].score !== undefined);
if (!allSucceeded) {
  throw new Error('One or more quality subagents failed!');
}
```

---

### 2. Aggregate Violations by Severity

**Collect all violations from all dimensions:**

```javascript
const allViolations = dimensions.flatMap((dim) =>
  results[dim].violations.map((v) => ({
    ...v,
    dimension: dim,
  })),
);

// Attribution first, because everything below depends on it. A violation with no
// registry `row` is a severity somebody chose, which is the thing the registry
// exists to prevent, and there is no safe fallback: keying dedup on `category`
// instead lets two workers describing one defect survive as two.
const unattributed = allViolations.filter((v) => !v.row);
if (unattributed.length > 0) {
  const dimensions = [...new Set(unattributed.map((v) => v.dimension))];
  throw new Error(`${unattributed.length} violation(s) carry no registry row; re-run these workers: ${dimensions.join(', ')}`);
}

// Deduplicate before counting. Some registry rows are detectable by more than one
// worker (M4 by isolation and maintainability, H5 by maintainability and
// performance), and counting the same defect twice deducts twice for it. Identity
// is the registry row at a location, never the prose description, which differs
// between workers describing the same line.
//
// File-level rows carry no meaningful line: H5 is a property of the whole file,
// and the pact config rows are properties of the whole config. Two workers each
// pick a plausible line for the same finding (1 and 341 for the same 341-line
// file), so the line is dropped from the key for those rows or the dedup this
// block exists for never fires on its own worked example.
const FILE_LEVEL_ROWS = new Set(['H5', 'H6', 'H7', 'H8', 'L4']);
const locationOf = (v) => (FILE_LEVEL_ROWS.has(v.row) ? 'file' : v.line);

const seenViolations = new Set();
const dedupedViolations = allViolations.filter((v) => {
  const key = `${v.file}:${locationOf(v)}:${v.row}`;
  if (seenViolations.has(key)) return false;
  seenViolations.add(key);
  return true;
});

// Group by severity (four tiers, matching the report template).
// CRITICAL: violations placed in the report's `## Critical Issues (Must Fix)`
// section (P0) count as CRITICAL; subagent HIGH/MEDIUM/LOW map to the
// report's Recommendations section (P1/P2/P3).
const criticalSeverity = dedupedViolations.filter((v) => v.severity === 'CRITICAL');
const highSeverity = dedupedViolations.filter((v) => v.severity === 'HIGH');
const mediumSeverity = dedupedViolations.filter((v) => v.severity === 'MEDIUM');
const lowSeverity = dedupedViolations.filter((v) => v.severity === 'LOW');

const violationSummary = {
  total: dedupedViolations.length,
  CRITICAL: criticalSeverity.length,
  HIGH: highSeverity.length,
  MEDIUM: mediumSeverity.length,
  LOW: lowSeverity.length,
};
```

**Every violation must carry the registry row that produced it**, which is what the
attribution guard above enforces. A violation with no `row` is a severity somebody
chose, which is the thing the registry exists to prevent. Reject the dimension
output and re-run that worker rather than scoring an unattributed violation, and
never substitute the prose `category` for a missing row.

**Everything downstream reads `dedupedViolations`.** The counts, the persisted
violation list, and every report-facing collection come from the same array, or the
report prints more findings than its own summary counts.

---

### 3. Calculate Quality Score

**This deduction ledger is the ONE scoring model for this workflow.** It is the
same arithmetic the `## Quality Score Breakdown` block in
`test-review-template.md` prints, so the published breakdown and the published
score are the same calculation and a reader can check one against the other.
Never substitute a weighted average, a per-dimension roll-up, or any other
formula, and never adjust the result by judgment after computing it.

```javascript
const deductions = violationSummary.CRITICAL * 10 + violationSummary.HIGH * 5 + violationSummary.MEDIUM * 2 + violationSummary.LOW * 1;
```

**Bonus points.** Exactly six categories, each worth `0` or `5` and nothing in
between: no partial credit, no invented categories, no category counted twice.
Award `5` only when the criterion holds across every reviewed file; otherwise
award `0`.

```javascript
const bonuses = {
  excellentBdd: 0, // 5: every test name states behavior, not implementation
  comprehensiveFixtures: 0, // 5: setup goes through fixtures, no inline duplication
  dataFactories: 0, // 5: test data comes from factories, not hardcoded literals
  networkFirst: 0, // 5: network interception is declared before the action that triggers it
  perfectIsolation: 0, // 5: no shared mutable state, any test can run alone or in parallel
  allTestIds: 0, // 5: every element lookup uses a stable test id, never a CSS or text selector
};

const bonusTotal = Object.values(bonuses).reduce((sum, value) => sum + value, 0);
```

**Final score**, clamped to the 0-100 range the report contract requires:

```javascript
const roundedScore = Math.max(0, Math.min(100, 100 - deductions + bonusTotal));
```

**Determine grade.** These five letters are the complete scale. Never emit a
modifier such as `A+`, `B-`, or any label outside this function.

```javascript
const getGrade = (score) => {
  if (score >= 90) return 'A';
  if (score >= 80) return 'B';
  if (score >= 70) return 'C';
  if (score >= 60) return 'D';
  return 'F';
};

const overallGrade = getGrade(roundedScore);
```

---

### 3b. Derive the Recommendation

**The recommendation is computed, never chosen.** Until this rule existed the score
was fully deterministic and the verdict beside it was free-form: the report template
offered four enum values and the reviewer picked one by judgment, while the CLI
checked only that the value was legal and that the two sections agreed. Nothing
bound the verdict to the findings.

That asymmetry is measurable. On couture-cast PR #103, two reviewers of the same
four files scored 82 and 85, a 3-point spread that is noise, and returned
`Request Changes` and "meets our quality bar for merge", which is the opposite
outcome. `--fail-on request-changes` acts on the verdict, so the gate was decided by
the unpinned half of the report.

```javascript
const deriveRecommendation = ({ CRITICAL, HIGH, MEDIUM, LOW }, score) => {
  if (CRITICAL > 0) return 'Block'; // a test that cannot fail is not a suggestion
  if (HIGH > 0) return 'Request Changes';
  if (score < 70) return 'Request Changes'; // volume of MEDIUM/LOW can also fail the bar
  if (MEDIUM + LOW > 0) return 'Approve with Comments';
  return 'Approve';
};

const recommendation = deriveRecommendation(violationSummary, roundedScore);
```

Why these boundaries:

- **`CRITICAL > 0` is `Block`.** A committed `.skip` on the one test that matters, a
  `expect(true).toBe(true)`, or an assertion against a self-configured mock means the
  suite reports green while proving nothing. That is worse than an absent test,
  because it buys false confidence, and it is not something a reviewer approves with
  a comment.
- **`HIGH > 0` is `Request Changes`.** Every HIGH row is either a test that can pass
  while the behavior is broken or one that fails at random. Both waste more
  engineering time downstream than fixing them costs now.
- **`score < 70` is `Request Changes` even with no HIGH.** Fifteen MEDIUM findings is
  a suite with a systemic problem, and a rule keyed only on severity tiers would wave
  it through.
- **Anything else with findings is `Approve with Comments`.** Real, worth fixing, not
  worth blocking a merge.

The reviewer's remaining judgment is which rows fired, which is where judgment
belongs. Write this value into **both** the `## Executive Summary` and the
`## Decision` section; the CLI rejects a report whose two copies disagree.

**A waiver is the only way past this**, it is recorded in the verdict payload, and it
never changes the computed value: `--waive` changes the exit code, not the
recommendation. Never soften the derived recommendation because context, a story, or
a focus note argued the findings were acceptable here.

**Before continuing, verify the ledger prints what it computed.** The breakdown
block in the report must show these exact deduction lines, this bonus total, and
this final score. A breakdown whose lines do not sum to the stated score is a
broken report; recompute rather than publishing the mismatch.

---

### 4. Prioritize Recommendations

**Extract recommendations from all dimensions:**

```javascript
const allRecommendations = dimensions.flatMap((dim) =>
  results[dim].recommendations.map((rec) => ({
    dimension: dim,
    recommendation: rec,
    impact: results[dim].score < 70 ? 'HIGH' : 'MEDIUM',
  })),
);

// Sort by impact (HIGH first)
const prioritizedRecommendations = allRecommendations.sort((a, b) => (a.impact === 'HIGH' ? -1 : 1)).slice(0, 10); // Top 10 recommendations
```

---

### 5. Create Review Summary Object

**Aggregate all results:**

```javascript
const reviewSummary = {
  overall_score: roundedScore,
  overall_grade: overallGrade,
  quality_assessment: getQualityAssessment(roundedScore),
  // Computed in 3b from the deduped violation counts. Publish this value in both
  // report sections verbatim; it is not a starting point for a judgment call.
  recommendation,
  // Carried through so the report can cite adoption counts on Convention rows and
  // say `PASS (n/a)` where a convention is absent rather than a bare WARN.
  convention_baseline: subagentContext.convention_baseline,

  dimension_scores: {
    determinism: results.determinism.score,
    isolation: results.isolation.score,
    maintainability: results.maintainability.score,
    performance: results.performance.score,
  },

  dimension_grades: {
    determinism: results.determinism.grade,
    isolation: results.isolation.grade,
    maintainability: results.maintainability.grade,
    performance: results.performance.grade,
  },

  violations_summary: violationSummary,

  // Deduped, not raw: violations_summary counts this same array, and a report
  // listing a defect twice beside a count of one is a report nobody can check.
  all_violations: dedupedViolations,

  critical_severity_violations: criticalSeverity,

  high_severity_violations: highSeverity,

  top_10_recommendations: prioritizedRecommendations,

  subagent_execution: 'PARALLEL (4 quality dimensions)',
  performance_gain: '~60% faster than sequential',
};

// Save for Step 4 (report generation)
fs.writeFileSync(`/tmp/tea-test-review-summary-${timestamp}.json`, JSON.stringify(reviewSummary, null, 2), 'utf8');
```

---

### 6. Display Summary to User

```
✅ Quality Evaluation Complete (Parallel Execution)

📊 Overall Quality Score: {roundedScore}/100 (Grade: {overallGrade})

📈 Dimension Scores:
- Determinism:      {determinism_score}/100 ({determinism_grade})
- Isolation:        {isolation_score}/100 ({isolation_grade})
- Maintainability:  {maintainability_score}/100 ({maintainability_grade})
- Performance:      {performance_score}/100 ({performance_grade})

ℹ️ Coverage is excluded from `test-review` scoring. Use `trace` for coverage analysis and gates.

⚠️ Violations Found:
- CRITICAL: {critical_count} violations
- HIGH:     {high_count} violations
- MEDIUM:   {medium_count} violations
- LOW:      {low_count} violations
- TOTAL:    {total_count} violations

🚀 Performance: Parallel execution ~60% faster than sequential

✅ Ready for report generation (Step 4)
```

---

---

### 7. Save Progress

**Save this step's accumulated work to `{outputFile}`.** When `output_file_override` is non-empty it IS `{outputFile}`, replacing the step frontmatter default.

- **If `{outputFile}` does not exist** (first save), create it using the workflow template (if available) with YAML frontmatter:

  ```yaml
  ---
  workflowType: 'testarch-test-review'
  stepsCompleted: ['step-03f-aggregate-scores']
  lastStep: 'step-03f-aggregate-scores'
  lastSaved: '{date}'
  ---
  ```

  Then write this step's output below the frontmatter.

- **If `{outputFile}` already exists**, update:
  - Add `'step-03f-aggregate-scores'` to `stepsCompleted` array (only if not already present)
  - Set `lastStep: 'step-03f-aggregate-scores'`
  - Set `lastSaved: '{date}'`
  - Append this step's output to the appropriate section of the document.

---

## EXIT CONDITION

Proceed to Step 4 when:

- ✅ All subagent outputs read successfully
- ✅ Overall score calculated
- ✅ Violations aggregated
- ✅ Recommendations prioritized
- ✅ Summary saved to temp file
- ✅ Output displayed to user
- ✅ Progress saved to output document

Load next step: `{nextStepFile}`

---

## 🚨 SYSTEM SUCCESS METRICS

### ✅ SUCCESS:

- All 4 subagent outputs read and parsed
- Violations aggregated correctly
- Overall score calculated from the deduction ledger, and the published breakdown sums to it
- Summary complete and saved

### ❌ FAILURE:

- Failed to read one or more subagent outputs
- Score calculation incorrect
- Summary missing or incomplete

**Master Rule:** Aggregate determinism, isolation, maintainability, and performance only.
