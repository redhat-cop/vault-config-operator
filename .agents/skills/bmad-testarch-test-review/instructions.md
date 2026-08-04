# Test Quality Review

**Workflow:** `bmad-testarch-test-review`
**Version:** 5.0 (Step-File Architecture)

---

## Overview

Review test quality using TEA knowledge base and produce a 0–100 quality score with actionable findings.

Coverage assessment is intentionally out of scope for this workflow. Use `trace` for requirements coverage and coverage gate decisions.

---

## WORKFLOW ARCHITECTURE

This workflow uses **step-file architecture**:

- **Micro-file Design**: Each step is self-contained
- **JIT Loading**: Only the current step file is in memory
- **Sequential Enforcement**: Execute steps in order

---

## INITIALIZATION SEQUENCE

### 1. Configuration Loading

From `workflow.yaml`, resolve:

- `config_source`, `test_artifacts`, `user_name`, `communication_language`, `document_output_language`, `date`
- `test_dir`, `review_scope`
- `headless` — when `true`, skip the greeting and interactive menu, execute Create mode directly, and never prompt the user
- `review_files` — comma-separated authoritative review set; when non-empty it IS the complete review set (takes precedence over `review_scope` discovery)
- `context_files` — comma-separated read-only context artifacts (story, PRD, test design, changed source). Read for understanding, never reviewed and never scored. Step 1 resolves `{context_basis}` from what it actually read, and step 4 publishes it
- `output_file_override` — when non-empty, this IS `{outputFile}` for every step: it replaces both `default_output_file` and the `outputFile` declared in each step's frontmatter
- `generate_inline_comments` — when `true`, write `// TODO (TEA Review)` inline comments into reviewed test files; default `false` is report-only

### 2. First Step

Load, read completely, and execute:
`{skill-root}/steps-c/step-01-load-context.md`

### 3. Resume Support

If the user selects **Resume** mode, load, read completely, and execute:
`{skill-root}/steps-c/step-01b-resume.md`

This checks the output document for progress tracking frontmatter and routes to the next incomplete step.
