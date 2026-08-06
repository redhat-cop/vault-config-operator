---
name: 'step-02-discover-tests'
description: 'Find and parse test files'
nextStepFile: '{skill-root}/steps-c/step-03-quality-evaluation.md'
outputFile: '{test_artifacts}/test-review.md'
---

# Step 2: Discover & Parse Tests

## STEP GOAL

Collect test files in scope and parse structure/metadata.

## MANDATORY EXECUTION RULES

- 📖 Read the entire step file before acting
- ✅ Speak in `{communication_language}`

---

## EXECUTION PROTOCOLS:

- 🎯 Follow the MANDATORY SEQUENCE exactly
- 💾 Record outputs before proceeding
- 📖 Load the next step only when instructed

## CONTEXT BOUNDARIES:

- Available context: config, loaded artifacts, and knowledge fragments
- Focus: this step's goal only
- Limits: do not execute future steps
- Dependencies: prior steps' outputs (if any)

## MANDATORY SEQUENCE

**CRITICAL:** Follow this sequence exactly. Do not skip, reorder, or improvise.

> **Exception — `review_files` supplied:** If `review_files` is non-empty, the discovered set equals `review_files` (comma-separated paths). Validate that each file exists — report missing files in the review report rather than silently dropping them — skip the glob in section 1, and continue the sequence from section 2. This is a first-class branch of the file-set source; the sequence remains mandatory.

> **Disclose every exclusion, on every branch.** The rule above is not specific to `review_files`: any file that a reader would expect in the reviewed set and that is not there gets named in the report's `## Excluded From Review Set` section with its reason, never omitted. One section, one entry shape — `path — reason` — and exactly three reasons are legal:
>
> - `path does not exist` — a `review_files` entry that is not on disk.
> - `file could not be parsed` — a discovered file the read or the parse failed on.
> - `format not scorable by the ledger` — a changed test artifact the registry has no criteria for (a Maestro flow, a `.feature` file, an `.http` collection).
>
> The third reason is the only one the runner can supply. When the run supplies an `---BEGIN UNSCORABLE---` block, reproduce every path in it verbatim with that exact reason, dropping none; the CLI rejects a report that dropped one. The first two you discover yourself, so add them to the same section with their own reason. No path may appear both here and in `## Reviewed Files`. A reviewed-files manifest that quietly omits a changed test artifact reads as "the diff held nothing else to review", which is a false statement the report makes on your behalf.

## 1. Discover Test Files

- **single**: use provided file path
- **directory**: glob under `{test_dir}` or selected folder
- **suite**: glob all tests in repo

Halt if no tests are found.

---

## 2. Parse Metadata (per file)

Collect:

- File size and line count
- Test framework detected
- Describe/test block counts
- Test IDs and priority markers
- Imports, fixtures, factories, network interception
- Waits/timeouts and control flow (if/try/catch)

---

## 2b. Derive the Convention Baseline

**Why this exists.** A criterion like "Priority Markers" used to fire as a bare
`4 violations, none present` in every repository on earth, including one that has
deliberately never used a priority marker. The violation was unfalsifiable: the
report could not say what the house standard was, so the reader could not tell a
real drift from the rubric's own preference. This pass measures the standard
before judging against it.

Sample the repository's **existing** test corpus and measure what it actually
does. Then `criteria-registry.md` scores each Convention row against the result.

### Sampling rules

- Sample test files that are **not in the review set**. A pull request adding four
  files must not be allowed to establish, or dilute, the convention it is judged
  against.
- Discover them the way `review_scope: suite` would, then cap the sample at **40
  files**, chosen closest-first by directory distance from the reviewed files, so
  the baseline describes the neighborhood the new tests live in rather than a
  distant corner of a monorepo.
- Record `corpusSize` (how many exist) and `sampled` (how many were read). When
  they differ, say so wherever the baseline is cited.
- Read only what the measurement needs: test names, locator calls, imports, and
  setup blocks. Do not evaluate quality here and do not produce violations; this
  step measures, `step-03` judges.

### Conventions to measure

For each key, count how many sampled files use it, and record the observed form
verbatim so the report can quote it back:

| Key               | Adopted when a sampled file…                                   | Record as `form`                                                                |
| ----------------- | -------------------------------------------------------------- | ------------------------------------------------------------------------------- |
| `priorityMarkers` | carries a priority marker on its tests                         | the observed shape, e.g. `[P0] in the test name`, `@P1 tag`, `{ tag: ['@p2'] }` |
| `testIds`         | locates elements by a stable test id                           | the attribute or helper, e.g. `data-testid`, `getByTestId`                      |
| `bddNaming`       | names tests by behavior rather than implementation             | e.g. `starts with a verb phrase`, `Given/When/Then`                             |
| `networkFirst`    | registers interception or a readiness signal before navigating | the helper, e.g. `interceptNetworkCall`, `page.route`                           |
| `dataFactories`   | builds domain payloads through a factory or builder            | e.g. `@couture/testing factories`, `build*` helpers                             |
| `fixtures`        | takes setup from a fixture rather than inline duplication      | e.g. `mergeTests`, `merged-fixtures`                                            |
| `assertionStyle`  | uses one assertion dialect consistently                        | e.g. `expect + vitest matchers`                                                 |

### Status thresholds, applied exactly

Deterministic on purpose. A model-chosen threshold reintroduces the variance this
whole pass removes.

```javascript
const conventionStatus = (adopted, sampled) => {
  if (sampled < 4) return 'unknown'; // corpus too small to infer a house rule
  if (adopted === 0) return 'absent';
  return adopted / sampled >= 0.5 ? 'established' : 'emerging';
};
```

### Output

Carry this object forward to `step-03` as `convention_baseline`. It travels in
`subagentContext` so all four workers score against the same measurement.

```javascript
const conventionBaseline = {
  corpusSize: 19,
  sampled: 19,
  conventions: {
    priorityMarkers: { adopted: 11, sampled: 19, status: 'established', form: '[P#] in the test name' },
    testIds: { adopted: 8, sampled: 19, status: 'emerging', form: 'data-testid' },
    networkFirst: { adopted: 5, sampled: 19, status: 'emerging', form: 'interceptNetworkCall' },
    // ...one entry per key in the table above, every key present
  },
};
```

Every key in the table must appear, including ones that came back `absent`: a
missing key is indistinguishable from an unmeasured one, and the registry needs
the difference to decide between deducting and passing as `n/a`.

**When the baseline cannot be measured** (no corpus outside the review set, a
shallow clone with nothing else checked out), set every status to `unknown` and
record `baselineUnavailable: true` with the reason. Every Convention row then
passes as `n/a` and the report says the baseline was unavailable. Guessing a
convention from the reviewed files themselves is circular; do not do it.

---

## 3. Evidence Collection (if `tea_browser_automation` is `cli` or `auto`)

> **Fallback:** If CLI is not installed, fall back to MCP (if available) or skip evidence collection.

**CLI Evidence Collection:**
All commands use the same named session to target the correct browser:

1. `playwright-cli -s=tea-review open <target_url>`
2. `playwright-cli -s=tea-review tracing-start`
3. Execute the flow under review (using `-s=tea-review` on each command)
4. `playwright-cli -s=tea-review tracing-stop` → saves trace.zip
5. `playwright-cli -s=tea-review screenshot --filename={test_artifacts}/review-evidence.png`
6. `playwright-cli -s=tea-review network` → capture network request log
7. `playwright-cli -s=tea-review close`

After capturing `trace.zip`, prefer Playwright's newer trace CLI for local or downloaded artifact analysis:

- `npx playwright trace open <trace.zip>` to start a trace session
- `npx playwright trace actions --grep="expect"` to jump to the failing assertion
- `npx playwright trace action <n>` / `trace snapshot <n> --name after` for root-cause details
- `npx playwright trace close` when done

> **Session Hygiene:** Always close sessions using `playwright-cli -s=tea-review close`. Do NOT use `close-all` — it kills every session on the machine and breaks parallel execution.

---

## 4. Save Progress

**Save this step's accumulated work to `{outputFile}`.** When `output_file_override` is non-empty it IS `{outputFile}`, replacing the step frontmatter default.

- **If `{outputFile}` does not exist** (first save), create it using the workflow template (if available) with YAML frontmatter:

  ```yaml
  ---
  workflowType: 'testarch-test-review'
  stepsCompleted: ['step-02-discover-tests']
  lastStep: 'step-02-discover-tests'
  lastSaved: '{date}'
  ---
  ```

  Then write this step's output below the frontmatter.

- **If `{outputFile}` already exists**, update:
  - Add `'step-02-discover-tests'` to `stepsCompleted` array (only if not already present)
  - Set `lastStep: 'step-02-discover-tests'`
  - Set `lastSaved: '{date}'`
  - Append this step's output to the appropriate section of the document.

Load next step: `{nextStepFile}`

## 🚨 SYSTEM SUCCESS/FAILURE METRICS:

### ✅ SUCCESS:

- Step completed in full with required outputs

### ❌ SYSTEM FAILURE:

- Skipped sequence steps or missing outputs
  **Master Rule:** Skipping steps is FORBIDDEN.
