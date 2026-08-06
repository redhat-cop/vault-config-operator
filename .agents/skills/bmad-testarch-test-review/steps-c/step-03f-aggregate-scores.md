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

// Group by severity (four tiers, matching the report template).
// CRITICAL: violations placed in the report's `## Critical Issues (Must Fix)`
// section (P0) count as CRITICAL; subagent HIGH/MEDIUM/LOW map to the
// report's Recommendations section (P1/P2/P3).
const criticalSeverity = allViolations.filter((v) => v.severity === 'CRITICAL');
const highSeverity = allViolations.filter((v) => v.severity === 'HIGH');
const mediumSeverity = allViolations.filter((v) => v.severity === 'MEDIUM');
const lowSeverity = allViolations.filter((v) => v.severity === 'LOW');

const violationSummary = {
  total: allViolations.length,
  CRITICAL: criticalSeverity.length,
  HIGH: highSeverity.length,
  MEDIUM: mediumSeverity.length,
  LOW: lowSeverity.length,
};
```

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

  all_violations: allViolations,

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
