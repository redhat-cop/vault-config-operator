---
name: 'criteria-registry'
description: 'The single rule registry: every criterion, its firing predicate, its pinned severity, and what gates its applicability'
---

# Criteria Registry

## WHY THIS FILE EXISTS

Two vendors reviewing the same file used to be able to agree on the defect and
disagree on its severity, and severity is what the CI gate acts on. A `HIGH`
instead of a `MEDIUM` moves the score by 3 and can flip a verdict. So severity
is not a judgment call here: every row below carries a fixed severity, and the
only decision left to the reviewer is whether the predicate fires.

Three rules bind every evaluation:

1. **Severity is read from this table, never chosen.** A violation's severity is
   whatever its row says. If a defect matches no row, report it in prose under
   Best Practices or Recommendations without a severity and without a deduction,
   and say the registry has no row for it. Inventing a severity is a defect in
   the review, not a finding about the tests.
2. **A criterion fires only when its gate is open.** The `Gate` column says what
   has to be true before the row can produce a violation at all. A closed gate
   is `PASS (n/a)` with the reason stated, never a `WARN` and never a deduction.
3. **Context and convention may raise, never waive.** No repo habit, story, or
   focus note lowers a severity in this table or excuses an Absolute row.

## THE THREE GATE CLASSES

**Absolute.** Applies to every reviewed test file, always. These are correctness
and stability properties. A repo where every existing test sleeps on a timer does
not thereby earn a waiver for hard waits; it earns the same violation on every
file. Repo adoption is irrelevant to an Absolute row and must never be consulted
for one.

**Applicability.** Applies when the reviewed file exercises the thing the
criterion protects. A criterion about navigation races cannot fire on a file that
never navigates. The gate is a property of the file under review, decided by
reading it, not a popularity measurement.

**Convention.** Applies when the repository demonstrably has a house convention,
measured by the baseline that `step-02-discover-tests` computes over the existing
corpus outside the review set. This is the class that used to produce the review's
worst noise: a bare `Priority Markers: 4 violations, none present` fires
identically in a repo that has the convention and a repo that has never used one.
Read the deduction schedule below before scoring any Convention row.

### Convention deduction schedule

`step-02` classifies each convention as `established`, `emerging`, `absent`, or
`unknown` using fixed thresholds. Apply this schedule exactly:

| Baseline status | Meaning                                                  | Effect on a reviewed file that lacks it                                                                                    |
| --------------- | -------------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------- |
| `established`   | adopted in ≥ 50% of the sampled corpus, corpus ≥ 4 files | Violation at the row's stated severity. Cite the adoption count.                                                           |
| `emerging`      | adopted in ≥ 1 file but < 50%                            | Violation one severity step lower, floored at `LOW`. Cite the adoption count and say the convention is not yet house-wide. |
| `absent`        | adopted in 0 files                                       | **No violation and no deduction.** Status `✅ PASS (n/a)`, note "the repo uses no such convention (0 of N sampled)".       |
| `unknown`       | corpus < 4 files, too small to infer                     | **No violation and no deduction.** Status `✅ PASS (n/a)`, note the corpus was too small to establish a convention.        |

One step lower means `CRITICAL`→`HIGH`, `HIGH`→`MEDIUM`, `MEDIUM`→`LOW`,
`LOW`→`LOW`. Never step a severity in any other circumstance.

---

## CRITICAL rows — the test cannot fail, or never reaches the system under test

Before this registry existed, no subagent defined a single `CRITICAL` violation
while the aggregation step counted them at `-10` each and the report reserved a
`## Critical Issues (Must Fix)` section for them. Every critical finding was
therefore improvised. These are the rows. A test matching any of them provides no
evidence at all, which is worse than having no test, because the suite reports
green.

| ID  | Criterion                    | Fires when                                                                                                                                                                                                                      | Severity | Gate     |
| --- | ---------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------: | -------- |
| C1  | Disabled test                | A reviewed test is skipped or excluded: `.skip`, `xit`, `xdescribe`, `test.todo` over a previously real body, `@Ignore`, `@Disabled`, `pytest.mark.skip` without a documented, still-true reason on the line or the line above. | CRITICAL | Absolute |
| C2  | Focused test                 | `.only`, `fdescribe`, `fit`, or `test.only` is committed, silently disabling every sibling test in the file.                                                                                                                    | CRITICAL | Absolute |
| C3  | Tautological assertion       | An assertion compares a value to itself or to a literal that cannot differ: `expect(true).toBe(true)`, `expect(1).toBe(1)`, `assert x == x`, `expect(y).toBe(y)`.                                                               | CRITICAL | Absolute |
| C4  | No assertion                 | A test body contains zero assertions and zero explicit failure paths. A test that only performs setup and navigation asserts nothing.                                                                                           | CRITICAL | Absolute |
| C5  | Mock asserted against itself | The only assertion targets a mock, stub, or spy that the same test configured, with no call into the system under test between configuration and assertion. The test proves the mocking library works.                          | CRITICAL | Absolute |
| C6  | Assertion unreachable        | An assertion sits after an unconditional `return`, inside a `catch` that the happy path never enters, or inside a callback the test never awaits, so it cannot execute.                                                         | CRITICAL | Absolute |

## HIGH rows — the test can pass while the behavior is broken, or fails at random

| ID  | Criterion                   | Fires when                                                                                                                                                                                                                                                                                          | Severity | Gate                                                           |
| --- | --------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------: | -------------------------------------------------------------- |
| H1  | Hard wait                   | `waitForTimeout`, `sleep(`, `time.sleep(`, `Thread.sleep(`, `cy.wait(<number>)`, or any bare timer used to order steps.                                                                                                                                                                             |     HIGH | Absolute                                                       |
| H2  | Wall-clock fixture          | A time-sensitive fixture is derived from the live clock (`Date.now()`, `new Date()` with no argument, `time.time()`) without fake timers, and the value governs an expiry, token lifetime, TTL, or scheduling boundary. Security-relevant lifetimes are the reason this is HIGH rather than MEDIUM. |     HIGH | Applicability: the file builds or asserts a time-bounded value |
| H3  | Conditional assertion       | Control flow decides whether or what to assert: an `if`/ternary selecting the expected value, a `try`/`catch` that swallows a failure, an assertion inside a loop that may run zero times.                                                                                                          |     HIGH | Absolute                                                       |
| H4  | Unreset shared state        | Module- or suite-level mutable state is written inside a test with no `beforeEach`/`afterEach` reset, so test order changes the outcome.                                                                                                                                                            |     HIGH | Absolute                                                       |
| H5  | Oversize test file          | The reviewed file exceeds 300 lines.                                                                                                                                                                                                                                                                |     HIGH | Absolute                                                       |
| H6  | Pact worker parallelism     | `vitest.config.pact.ts` omits `fileParallelism: false`. Parallel workers race on the shared pact JSON.                                                                                                                                                                                              |     HIGH | Absolute                                                       |
| H7  | Pact pool isolation         | `pool: 'forks'` or `poolOptions.forks.singleFork: true` is missing **and** the repo has ≥ 2 `.pacttest.ts` files for the same consumer+provider pair. Single-file suites take L4 instead.                                                                                                           |     HIGH | Applicability: ≥ 2 `.pacttest.ts` for the pair                 |
| H8  | Pact serialization defeated | Any of `sequence.concurrent: true`, `maxConcurrency > 1`, `maxWorkers > 1`, `isolate: false` in a pact vitest config.                                                                                                                                                                               |     HIGH | Absolute                                                       |

## MEDIUM rows — the test works but diagnoses poorly or will erode

| ID  | Criterion                | Fires when                                                                                                                                                                                                                                                                                                           | Severity | Gate                                                                    |
| --- | ------------------------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------: | ----------------------------------------------------------------------- |
| M1  | Network-first violated   | The file navigates (`page.goto`, `cy.visit`, a router push) and then interacts with or asserts on data-dependent content, with no intercept, route stub, or explicit readiness signal registered before the navigation. A generic post-navigation helper is not a readiness signal for the data the test then reads. |   MEDIUM | Applicability: the file navigates and then reads data-dependent content |
| M2  | Repeated literal payload | The same domain payload shape is constructed inline three or more times in the file, or a factory for that shape already exists in the repo and the file bypasses it.                                                                                                                                                |   MEDIUM | Applicability: the file constructs domain payloads                      |
| M3  | Multi-concern test       | One test asserts against three or more unrelated subjects, so a failure does not localize. Count subjects, not `expect` calls: three assertions about one response is one concern.                                                                                                                                   |   MEDIUM | Absolute                                                                |
| M4  | Ungrouped suite          | A file with three or more tests has no `describe`/`context` grouping, so failures print without a subject.                                                                                                                                                                                                           |   MEDIUM | Absolute                                                                |
| M5  | Low-level event dispatch | `fireEvent` (or an equivalent raw dispatch) is used where the project already depends on a user-level API such as `userEvent`, so the test skips the real interaction sequence.                                                                                                                                      |   MEDIUM | Applicability: a user-level interaction API is a project dependency     |
| M6  | Unawaited async          | A promise-returning call in a test body is not awaited and not explicitly returned, so the assertion may run before the effect.                                                                                                                                                                                      |   MEDIUM | Absolute                                                                |
| M7  | Excessive nesting        | `describe`/`context` nesting deeper than three levels, or a test body nested more than three blocks deep.                                                                                                                                                                                                            |   MEDIUM | Absolute                                                                |

## LOW rows — real, cheap to fix, no risk to the verdict on their own

| ID  | Criterion                      | Fires when                                                                                                                                                                                                                                                           | Severity | Gate                                                 |
| --- | ------------------------------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------: | ---------------------------------------------------- |
| L1  | Fragile selector               | An element is located by CSS id or class, by XPath, or by presentation text that copy or layout work would change, where a role-, label-, or test-id-based locator is available. A role- or label-based locator **satisfies** this row: it is not a missing test id. |      LOW | Applicability: the file locates DOM elements         |
| L2  | Missing priority marker        | A reviewed test carries no priority marker in the form the baseline recorded.                                                                                                                                                                                        |      LOW | Convention: `priorityMarkers`                        |
| L3  | Missing stable test id         | An element lookup uses neither a stable test id nor a role/label locator, in a repo whose baseline shows a test-id convention. Redundant with L1 where L1 already fired on the same line; report once.                                                               |      LOW | Convention: `testIds`                                |
| L4  | Pact single-file pool advisory | `pool: 'forks'` / `singleFork: true` missing on a suite with exactly one `.pacttest.ts` for the pair. Future-proofing, not a live flake.                                                                                                                             |      LOW | Applicability: exactly 1 `.pacttest.ts` for the pair |
| L5  | Implementation-shaped name     | A test name states the implementation rather than the behavior (names a method, a selector, or "works correctly") in a repo whose baseline shows a behavioral naming convention.                                                                                     |      LOW | Convention: `bddNaming`                              |
| L6  | Magic value                    | An unexplained numeric or string literal carries domain meaning and appears with no name or comment.                                                                                                                                                                 |      LOW | Absolute                                             |
| L7  | Inconsistent assertion style   | The file mixes assertion dialects (`assert` and `expect`, or two matcher styles for the same check) against a baseline-established house style.                                                                                                                      |      LOW | Convention: `assertionStyle`                         |

---

## MAPPING TO THE CRITERIA TABLE

The report's `## Quality Criteria Assessment` table has one row per published
criterion. Each table row draws from the registry rows below it, and its `Basis`
column states the gate that decided it.

| Report criterion                     | Registry rows          | Basis                         |
| ------------------------------------ | ---------------------- | ----------------------------- |
| BDD Format (Given-When-Then)         | L5                     | Convention: `bddNaming`       |
| Test IDs                             | L3                     | Convention: `testIds`         |
| Priority Markers (P0/P1/P2/P3)       | L2                     | Convention: `priorityMarkers` |
| Hard Waits (sleep, waitForTimeout)   | H1                     | Absolute                      |
| Determinism (no conditionals)        | H2, H3, C6             | Absolute + Applicability      |
| Isolation (cleanup, no shared state) | H4, C5                 | Absolute                      |
| Fixture Patterns                     | M2, M5                 | Applicability                 |
| Data Factories                       | M2                     | Applicability                 |
| Network-First Pattern                | M1                     | Applicability                 |
| Explicit Assertions                  | C3, C4, M6             | Absolute                      |
| Test Length (≤300 lines)             | H5                     | Absolute                      |
| Test Duration (≤1.5 min)             | H1, M1                 | Absolute                      |
| Flakiness Patterns                   | H1, H2, H3, H4, M1, M6 | Absolute + Applicability      |
| Disabled or Focused Tests            | C1, C2                 | Absolute                      |

`Disabled or Focused Tests` is a new published row. It existed as a scoring
possibility with no rule and no table line, which is how a committed `.skip` on
the most important test in a module could be found by attention rather than by
the rubric.

## STATUS SYMBOLS

- `✅ PASS` — the gate was open and no row fired.
- `✅ PASS (n/a)` — the gate was closed. State why in the note. Never deducts.
- `⚠️ WARN` — one or more rows fired at `MEDIUM` or `LOW`.
- `❌ FAIL` — one or more rows fired at `CRITICAL` or `HIGH`.

A `WARN` with zero violations is a contradiction; so is a `PASS` with a nonzero
count. Reconcile before publishing.
