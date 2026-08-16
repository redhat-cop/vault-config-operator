# Epic Dev — Phase Logic

Communicate with the user in `{communication_language}`.

**Contract anchor** (survives context compaction — re-read if uncertain):
- You are an orchestrator, not an implementer — delegate all work to subagents
- Every sub-skill (`bmad-create-story`, `bmad-dev-story`, `bmad-code-review`) runs in a **fresh subagent** with clean LLM context
- Subagents run autonomously but you **relay "decision needed" questions** back to the user — see **Decision Relay Protocol** below
- On any failure: **halt** with what failed, which story, and instructions to re-invoke
- `sprint-status.yaml` is the sole checkpoint — story statuses determine resume position
- Respect `--commit-policy` and `--skip-confirmations` args from SKILL.md
- **Model assignment:** `bmad-create-story` and `bmad-dev-story` subagents use **Opus 4.6** (high reasoning). `bmad-code-review` subagents use **ChatGPT 5.4** (medium reasoning, different model to prevent self-review bias). Pass the model when spawning each subagent. If a model isn't available, warn the user and fall back — but always log when review runs on the same model as development.

---

## Decision Relay Protocol

When a subagent reports `decision_needed`, the orchestrator MUST follow this exact sequence. **Do NOT skip steps or treat the decision as resolved without user input.**

### Detection

After every subagent completes, parse its output for these signals:
- The string `decision_needed` in the status field
- Questions phrased as "should we…", "which approach…", "is it acceptable…"
- Explicit `[Decision]` tags in review findings

### Relay to User

1. **Present the decision clearly** using the `AskQuestion` tool with structured options when the decision has discrete choices, or as plain text when it's open-ended. Include:
   - Which story raised the question (story key + title)
   - Which phase raised it (create-story / dev-story / code-review)
   - The full question text from the subagent
   - Any context the subagent provided (code snippets, trade-offs)

2. **Wait for the user's response.** Do NOT proceed with the story pipeline until the user answers. Other stories in the same layer that don't need decisions can continue.

### Resume the Subagent

3. After the user responds, **resume the subagent** using the Task tool with `resume: <agent_id>`:
   ```
   The user decided: {user's answer}
   Please continue with this decision applied.
   ```

4. If the resumed subagent completes successfully, continue the story pipeline (next step in B.1).

5. If the resumed subagent raises another decision, repeat this protocol.

### Record the Decision

6. After the story pipeline completes, update the story file's **Code Review Record** section:
   - **Decisions Needed**: the original question(s)
   - **Decisions Taken**: the user's answer(s) with rationale if provided
   - Record the review model in **Review Model Used**
   - Record all findings in **Review Findings**
   - Record all applied fixes in **Fixes Applied**

---

## Step 0: Validate Input and Parse Epic

**Goal:** Identify the target epic, verify it's actionable, and build the execution plan.

If an `epic-plan.py` script exists at `./scripts/epic-plan.py`, run it to extract the execution plan as JSON (faster, deterministic). Otherwise, perform the steps below manually.

1. If the user didn't provide an epic number, check `sprint-status.yaml` for epics in `backlog` state. If exactly one, use it. If multiple, ask the user to choose.

2. Load `sprint-status.yaml` fully. Verify the epic status is `backlog` or `in-progress`. If `done`, halt — nothing to do.

3. Find the epic source file. Search in this order:
   - `{implementation_artifacts}/{epicNum}-epic-*.md` (implementation artifact — preferred, has dependency table)
   - `{planning_artifacts}/*epic*.md` (planning artifact — search for the epic section)

4. Parse the **Stories table** from the epic file. Extract for each story:
   - Story number (e.g., `13.1`)
   - Title
   - Dependency (e.g., `None`, `13.1`, or `13.1, 13.4` for multiple)

5. Build a dependency graph (DAG) and compute a **topological sort**. Group stories into **dependency layers** — stories in the same layer have all dependencies satisfied by earlier layers.

6. Cross-reference with `sprint-status.yaml` to determine current state of each story:
   - `backlog` → needs Phase A (create-story)
   - `ready-for-dev` → needs Phase B (dev-story + code-review)
   - `in-progress` or `review` → needs Phase B continuation
   - `done` → skip

7. Present the execution plan to the user:
   - Epic name and story count
   - Dependency layers with parallelism opportunities (stories per layer)
   - Stories already completed (if resuming)
   - Estimated phases remaining

   Unless `--skip-confirmations`, ask: "Ready to proceed, or would you like to adjust anything?"

8. **Create the epic branch** (if not already on it). All epic work happens on a dedicated branch — never directly on `main`:
   ```
   git checkout -b epic-{N}
   ```
   If the branch already exists (resuming), check it out:
   ```
   git checkout epic-{N}
   ```
   Set `{epic-branch}` = `epic-{N}` (e.g., `epic-14`). All subsequent commits in Phase A and merge targets in Phase B use this branch.

   **CRITICAL:** Verify the current branch before proceeding:
   ```
   git rev-parse --abbrev-ref HEAD
   ```
   Must output `epic-{N}`. If it outputs `main` or anything else, halt — branch creation failed.

---

## Phase A: Create Story Specifications

**Goal:** Transform all `backlog` stories into comprehensive spec files via `bmad-create-story`.

**Parallelism:** Stories in the same dependency layer with all dependencies satisfied can be specced in parallel via concurrent subagents. Story specs are independent markdown files — no code conflicts.

For each dependency layer (in order):

1. Identify stories in this layer still in `backlog` state.

2. Launch subagents — one `bmad-create-story` per story. If multiple stories are ready, launch them in parallel:

   **Subagent prompt template:**
   ```
   Run the bmad-create-story skill for story {epicNum}.{storyNum} in epic {epicNum}.
   The story title is "{title}".
   Process the story fully and autonomously — no user interaction needed.
   If you encounter a blocking issue that requires a decision, describe the
   decision needed clearly and halt.
   When complete, report: story_key, output_file_path, status (success/failed/decision_needed),
   and a one-line summary.
   ```

3. Wait for all subagents in the batch to complete.
   - If any reports `decision_needed`: relay the question to the user, get the answer, and resume that subagent.
   - If any fails: halt the epic and report which story failed and why.

4. Fall back to sequential execution if parallel subagents aren't available.

5. After each batch, verify sprint-status updates (stories → `ready-for-dev`, epic → `in-progress`).

6. Report progress: "Created specs for stories {list}. {remaining} remaining."

When all `backlog` stories have specs:

7. **Review gate** (always, even with `--skip-confirmations`): List all created story spec files with paths and ask the user to review them before development begins. This is the last chance to adjust requirements, acceptance criteria, or scope before code is written.

   ```
   Phase A complete — {count} story specs created:
   {list of story files with paths}

   Please review the story specs. When satisfied, confirm to proceed.
   (You can edit any story file now — changes will be picked up by dev-story.)
   ```

   Wait for explicit user confirmation before continuing.

8. **Commit** (respecting `--commit-policy`):
   ```
   chore(epic-{N}): create story specifications
   ```

9. Proceed to Phase B.

---

## Phase B: Implement, Review, and Commit Stories

**Goal:** For each dependency layer, develop and review stories in parallel using **git worktrees**, then merge and commit.

### Worktree-Based Parallel Execution

Stories within the same dependency layer are independent — their dependencies are all `done` from previous layers. Each story runs in its own git worktree (separate branch, separate working directory), so parallel development is safe with no working-tree conflicts.

**Subagent type:** Use `best-of-n-runner` subagents when available — they run in isolated git worktrees automatically. Fall back to sequential `generalPurpose` subagents on the epic branch if worktree subagents aren't available.

### Integration Test Isolation Protocol

Integration tests are an **orchestrator responsibility** — dev subagents run only unit tests (`make test`). The orchestrator runs `make integration` at two points: once per layer as a baseline (on the epic branch before launching worktrees), and once per story as a pre-merge gate (in each worktree after code review is approved).

When running pre-merge integration tests in parallel (multiple stories in the same layer completing around the same time), each worktree **must** use its own Kind cluster to avoid race conditions on shared Kubernetes namespaces, Vault state, and host port bindings.

**Port and cluster assignment:** Before launching parallel worktrees for a dependency layer, the orchestrator assigns each story a unique port offset based on its position within the layer (0-indexed):

| Variable | Formula | Example (story index 0, 1, 2) |
|---|---|---|
| `KIND_CLUSTER_NAME` | `vault-config-operator-{story_key}` | `vault-config-operator-s13-1` |
| `VAULT_HOST_PORT` | `8200 + (layer_story_index + 1) * 10` | `8210`, `8220`, `8230` |
| `HTTPS_HOST_PORT` | `VAULT_HOST_PORT + 1` | `8211`, `8221`, `8231` |

The base port `8200` (offset 0) is reserved for the default cluster (`vault-config-operator`) and is used for the pre-layer baseline test on the epic branch. The formula assumes single-epic parallelism; concurrent epics could collide on ports.

**Sequential fallback:** When stories run sequentially on the epic branch (B.5), no per-story overrides are needed — the default cluster name and ports are used. The orchestrator still runs integration tests before and after each story.

### Worktree Path Isolation (CRITICAL)

The `best-of-n-runner` creates a real git worktree (separate branch + separate directory), but does **not** sandbox file operations. If the subagent receives absolute paths pointing to the original repository, it will read/write there — defeating isolation entirely.

**Rules for the orchestrator when spawning worktree subagents:**

1. **Use relative paths only.** Never pass an absolute path to the original repository in subagent prompts. Use paths relative to the project root (e.g., `{implementation_artifacts}/{story_key}.md`).

2. **Include the worktree isolation block** (below) in every `best-of-n-runner` subagent prompt. This instructs the subagent to resolve `{project-root}` from its current working directory, not from any hardcoded path.

3. **The commit (Step 4) happens on the story branch** inside the worktree. The orchestrator must NOT commit on behalf of the subagent on the epic branch — the worktree subagent owns its own branch.

**Worktree isolation block** (include verbatim in every `best-of-n-runner` prompt):
```
WORKTREE ISOLATION (CRITICAL):
You are running in an isolated git worktree on branch epic-{N}/story-{N.M}.
Your project root is your current working directory — do NOT use absolute
paths to any other repository copy. All file reads and writes MUST use paths
relative to your CWD or resolved from your CWD. If the skill resolves
{project-root}, it MUST resolve to your current working directory.
Run `git rev-parse --show-toplevel` if you need to confirm your project root.
```

For each dependency layer (in topological order):

### B.0: Pre-Layer Integration Baseline

0. **Before launching any worktree subagents for this layer**, run integration tests on the current epic branch state to establish a green baseline. This confirms the foundation is solid before anyone starts coding.

   ```
   make integration
   ```

   Use the default Kind cluster (`vault-config-operator`) on the default port (`8200`). If a default cluster is already running from a previous layer, reuse it.

   - If tests **pass**: proceed to B.1. Log: "Layer {L} baseline: integration tests pass."
   - If tests **fail**: **HALT the entire epic.** Report the failure and ask the user to resolve. Do NOT launch any worktree subagents on a broken baseline.

   **On resume after a baseline failure:** The orchestrator re-runs the baseline before continuing.

### B.1: Launch Parallel Development

1. Identify stories in this layer where status is `ready-for-dev` (or `in-progress`/`review` for resumption).

2. For each story, the orchestrator manages a multi-step pipeline in the story's worktree. Each story gets its own worktree branch (e.g., `epic-{N}/story-{N.M}`).

   **Step 1 — IMPLEMENT** (model: **Opus 4.6**): Spawn a `best-of-n-runner` subagent for `bmad-dev-story`:
   ```
   Run the bmad-dev-story skill for story {epicNum}.{storyNum}.
   The story file is at: {implementation_artifacts}/{story_key}.md
   Process the story fully — implement all tasks, run unit tests (`make test`),
   mark complete. Integration tests will be run by the orchestrator after
   code review — do NOT run `make integration` yourself.
   If you encounter a blocking issue requiring a human decision, describe it
   clearly and halt. Do NOT make assumptions on behalf of the user.

   WORKTREE ISOLATION (CRITICAL):
   You are running in an isolated git worktree on branch epic-{N}/story-{N.M}.
   Your project root is your current working directory — do NOT use absolute
   paths to any other repository copy. All file reads and writes MUST use paths
   relative to your CWD or resolved from your CWD. If the skill resolves
   {project-root}, it MUST resolve to your current working directory.
   Run `git rev-parse --show-toplevel` if you need to confirm your project root.

   IMPORTANT: If you need a decision, your output MUST include:
   - status: decision_needed
   - decision_question: <the exact question for the user>
   - decision_context: <relevant code/trade-offs to help the user decide>
   - decision_options: <list of concrete options if applicable>

   When complete, report: story_key, status (success/failed/decision_needed),
   files_changed_count, and a one-line summary.
   ```

   After Step 1 completes, check for `decision_needed`. If found, follow the **Decision Relay Protocol** before proceeding.

   **Step 2 — CODE REVIEW** (model: **ChatGPT 5.4**): Spawn a separate subagent for `bmad-code-review` in the same worktree:
   ```
   Run the bmad-code-review skill to review the changes for story {epicNum}.{storyNum}.
   The story file is at: {implementation_artifacts}/{story_key}.md
   Run the review fully and autonomously — complete all layers, produce the
   triage report.
   If the review raises a design/requirements question (not a code fix),
   describe it and halt — do NOT resolve design questions yourself.

   WORKTREE ISOLATION (CRITICAL):
   You are running in an isolated git worktree on branch epic-{N}/story-{N.M}.
   Your project root is your current working directory — do NOT use absolute
   paths to any other repository copy. All file reads and writes MUST use paths
   relative to your CWD or resolved from your CWD. If the skill resolves
   {project-root}, it MUST resolve to your current working directory.
   Run `git rev-parse --show-toplevel` if you need to confirm your project root.

   IMPORTANT: If you need a decision, your output MUST include:
   - status: decision_needed
   - decision_question: <the exact question for the user>
   - decision_context: <relevant code/trade-offs to help the user decide>
   - decision_options: <list of concrete options if applicable>

   When complete, report: story_key, status (approved/changes_requested/decision_needed),
   review_patches_count, and a one-line summary.
   ```

   After Step 2 completes, check for `decision_needed`. If found, follow the **Decision Relay Protocol** before proceeding.

   **Step 3 — REVIEW FIXES**: If code review requested changes:
   1. Spawn a new Step 1 subagent (Opus 4.6) to address the findings
   2. **ALWAYS re-run Step 2** (ChatGPT 5.4 code review) after fixes are applied — never skip the re-review even if the fixes seem trivial
   3. Repeat Steps 1→2 until the review returns `approved` or is halted
   4. **HARD CAP: maximum 5 review iterations per story.** Track the iteration count (first review = iteration 1). If iteration 5 still returns `changes_requested`: **HALT the story pipeline immediately.** Do NOT proceed to Step 4. Instead:
      - Summarize ALL open/unresolved review findings for this story in a clear list
      - Include the iteration history (what was fixed in each round, what persists)
      - Present this summary to the user with the message: "⚠️ Review cap reached (5/5 iterations) for story {story_key}. Human intervention needed."
      - Wait for the user to decide: resolve the issues manually, accept as-is, or abandon
      - Only proceed to Step 4 after explicit user approval

   **CRITICAL:** Do NOT commit after applying fixes without a re-review pass. Every fix must be verified by a fresh code review subagent before it can be committed. Skipping re-review defeats the purpose of adversarial review — fix subagents can introduce new issues that only a second review would catch.

   **Step 4 — COMMIT** (respecting `--commit-policy`):
   ```
   git add -A && git commit -m "feat(epic-{N}): {story_key} — {title}"
   ```
   Report: story_key, review_patches_count, files_changed_count, commit_sha.

   **Step 5 — PRE-MERGE INTEGRATION TEST** (orchestrator-owned, runs in worktree):

   After commit, the **orchestrator itself** (not a subagent) runs integration tests in the story's worktree to verify the new code doesn't break anything before merging to main.

   ```
   cd <worktree-path>
   make integration KIND_CLUSTER_NAME={assigned_cluster_name} VAULT_HOST_PORT={assigned_vault_port} HTTPS_HOST_PORT={assigned_https_port}
   ```

   - If tests **pass**: the story is ready for merge. Tear down the per-story Kind cluster:
     ```
     make kind-teardown KIND_CLUSTER_NAME={assigned_cluster_name}
     ```

   - If tests **fail**: spawn a **fix agent** (model: **Opus 4.6**) in the same worktree with the failure details:
     ```
     Integration tests failed in the worktree for story {epicNum}.{storyNum}.
     The story file is at: {implementation_artifacts}/{story_key}.md

     FAILURE OUTPUT:
     {paste the relevant test failure output}

     Fix the failing tests. The failures may be caused by:
     - Incorrect test fixtures (wrong field values, missing config)
     - Missing controller/webhook registrations in test suite setup
     - Field name mismatches between CRD spec and Vault API responses
     - Missing or incorrect test infrastructure (Makefile targets, deploy scripts)

     After fixing, run `make test` to verify unit tests still pass.
     Do NOT run `make integration` yourself — the orchestrator will re-run it.

     WORKTREE ISOLATION (CRITICAL):
     {include the standard worktree isolation block}

     When complete, report: story_key, status (success/failed),
     files_changed, and a summary of what was fixed.
     ```

     After the fix agent completes, commit the fixes and **re-run `make integration`** in the worktree. Repeat until tests pass or the cap is reached.

     **HARD CAP: 3 integration-fix iterations per story.** If iteration 3 still fails:
     - **HALT the story pipeline immediately.**
     - Preserve the worktree and Kind cluster for manual debugging.
     - Present the failure summary to the user: "Integration tests still failing after 3 fix attempts for story {story_key}. Human intervention needed."
     - Wait for explicit user decision before proceeding.

   Steps 1-5 run sequentially per story, but **multiple stories in the same layer run their pipelines in parallel** (each in its own worktree).

3. Launch all stories in the layer simultaneously. Monitor for completion.

### B.2: Handle Subagent Results

4. As each subagent completes, parse its output and act on the status:

   - **Success / Approved:** Note the story's branch name and commit SHA. Continue to the next step in the story pipeline.

   - **Decision needed:** Follow the **Decision Relay Protocol**:
     1. Extract `decision_question`, `decision_context`, and `decision_options` from the subagent output
     2. Present to the user using `AskQuestion` (for discrete choices) or text (for open-ended questions)
     3. **HALT the story pipeline** — do NOT proceed until the user responds
     4. After the user responds, resume the subagent with `resume: <agent_id>` passing the user's decision
     5. Record the decision in the story file's Code Review Record
     6. Continue the story pipeline with the resumed subagent's output

   - **Changes requested (code review):** Proceed to Step 3 (review fixes) — this is NOT a decision, it's an actionable fix list.

   - **Failed:** Halt the entire epic. Report which story failed, the error, and resume instructions. Other in-flight subagents for the same layer can continue to completion (fail-forward within a layer) but no new layers start.

5. Wait for all subagents in the layer to finish (or halt).

### B.3: Merge Branches

6. After all stories in the layer succeed, **verify each worktree branch** before merging:

   For each completed story branch, run from the **epic branch**:
   ```
   git log --oneline epic-{N}/story-{N.M} --not HEAD | head -20
   ```
   Confirm at least one commit exists on the story branch beyond the merge base. If the branch has no commits (empty branch), the worktree subagent failed to commit — halt and report the issue.

   Also verify the worktree is clean (no uncommitted changes that were missed):
   ```
   git -C <worktree-path> status --porcelain
   ```
   If dirty files exist, either the subagent forgot to commit or the worktree isolation failed. Halt and report.

7. Merge each story branch back to the epic branch **sequentially** (to maintain a clean linear history):

   For each completed story branch:
   ```
   git checkout {epic-branch}
   git merge --no-ff epic-{N}/story-{N.M} -m "Merge story {N.M}: {title}"
   ```

   **CRITICAL:** Always use `--no-ff` to create a merge commit, even if fast-forward is possible. This preserves the branch history and makes it clear which commits belong to which story.

   If merge conflicts occur: halt with details. The user resolves conflicts manually, then re-invokes to resume. The conflicting branch is preserved for inspection.

8. **Post-merge verification.** Integration tests already passed in the worktree (Step 5). After merge, run the fast unit test suite (`make test`) on the epic branch as a quick sanity check that the merge itself didn't introduce issues. If tests fail, halt with the merge SHA and failure details — the user must resolve before continuing.

9. **Update story file status on the epic branch.** After each successful merge, on the epic branch:
   - Set the story file's `Status:` field to `done`
   - Populate the **Code Review Record** section if not already filled by the review subagent:
     - **Review Model Used**: the model slug used for code review
     - **Review Findings**: summary of findings from the review
     - **Decisions Needed / Taken**: any decisions that were relayed to the user
     - **Fixes Applied**: list of fixes applied after review
   - Commit these story-file updates as part of the merge or as a follow-up:
     ```
     chore(epic-{N}): update story {N.M} status and review record
     ```

   This ensures story file status is always consistent with `sprint-status.yaml`, regardless of whether the work was done in a worktree.

10. Clean up worktrees after successful merge:

   **Worktree cleanup:**
   ```
   git worktree remove <worktree-path>
   git branch -d epic-{N}/story-{N.M}
   ```

   **Verify cleanup:** Run `git worktree list` and confirm no orphaned worktrees remain for the merged story. If removal fails (e.g., untracked files), force-remove:
   ```
   git worktree remove --force <worktree-path>
   ```

   **Kind cluster note:** Per-story Kind clusters are torn down in Step 5 after integration tests pass. If a story was halted before reaching Step 5, its Kind cluster may still be running — see the orphan detection protocol in the Resume Behavior section.

11. Update `sprint-status.yaml`: mark each merged story as `done`.

   **HARD GUARD — Status Atomicity:** The orchestrator MUST NOT update sprint-status.yaml to `done` for a story unless the story file's `Status:` field has ALREADY been set to `done` in the same commit or an earlier commit on the epic branch. If the story file still shows `in-progress`, `review`, or any other non-done status, HALT and fix the story file first. This prevents drift between the two tracking systems.

12. **Final consistency check (BLOCKING GATE):** After all stories in the layer are merged, verify that:
    - Every merged story's file has `Status: done`
    - Every merged story in `sprint-status.yaml` is `done`
    - `git worktree list` shows only the main working tree (no orphaned worktrees)
    If ANY inconsistency is found: **HALT immediately.** Do NOT proceed to the next layer. Report the inconsistency to the user and fix it before continuing. This is a blocking gate, not advisory.

### B.4: Report Layer Progress

13. Report to the user:
   ```
   Layer {L} complete: {count} stories merged
   {story_list with review patch counts and file counts}
   Progress: {done}/{total} stories in epic {N}
   ```

14. Continue to the next dependency layer.

### B.5: Sequential Fallback

If `best-of-n-runner` subagents aren't available, process stories **sequentially on the epic branch (`{epic-branch}`)** within each layer:

**B.0 still applies:** Run the pre-layer integration baseline (`make integration`) on the epic branch before starting any story in the layer.

For each story:
1. Spawn a `generalPurpose` subagent for `bmad-dev-story` (unit tests only — the subagent runs `make test`, not `make integration`)
2. If `decision_needed`: follow the **Decision Relay Protocol** — present to user, wait, resume subagent
3. On success, spawn a `generalPurpose` subagent for `bmad-code-review` (model: **ChatGPT 5.4**)
4. If `decision_needed`: follow the **Decision Relay Protocol**
5. If `changes_requested`: re-run dev-story (Opus 4.6) to address findings, then **ALWAYS** re-run code-review (ChatGPT 5.4) — never skip the re-review, even if the fixes seem trivial. Repeat until approved or halted. **HARD CAP: 5 review iterations max.** If iteration 5 still returns `changes_requested`: **HALT immediately.** Summarize all open/unresolved findings, present to the user with "Review cap reached (5/5 iterations) for story {story_key}. Human intervention needed.", and wait for explicit user decision before proceeding.
6. **Pre-commit integration test** (orchestrator-owned): Run `make integration` on the epic branch (default cluster, default ports — sequential mode uses only one cluster). If tests fail, spawn a fix agent (Opus 4.6) to address failures. **HARD CAP: 3 integration-fix iterations.** If still failing after 3 attempts, HALT and escalate to user.
7. Commit (respecting `--commit-policy`):
   ```
   feat(epic-{N}): {story_key} — {story title}
   ```
8. **Update story file** on the epic branch:
   - Set `Status:` to `done`
   - Populate the **Code Review Record** section (Review Model, Findings, Decisions, Fixes)
9. Update `sprint-status.yaml`: mark story as `done`. **HARD GUARD:** Do NOT update sprint-status until the story file's `Status:` field is already `done` in a committed state.
10. **Verify consistency (BLOCKING GATE):** Confirm story file status and sprint-status agree before continuing. If they disagree, HALT and fix — do NOT proceed.
11. Report progress, continue to next story

### B.6: Check for New Stories

After all stories in the current set are done:

11. Re-read `sprint-status.yaml`. Check if any **new** `backlog` stories appeared in this epic (code review can spawn stories).

12. If new stories exist:
    - Report: "Code review spawned {count} new stories: {list}. Looping back to create specs."
    - Return to **Phase A** for the new stories, then back to Phase B.

13. If no new stories: proceed to completion.

---

## Completion

Report the epic summary:
- Total stories completed (including review-spawned)
- Total review patches across all stories
- Total commits and merge operations
- Dependency layers processed
- Any notable findings from code reviews

Update `sprint-status.yaml`: epic status → `done` (only if ALL stories are `done`).

```
Epic {N} complete: "{epic title}"
{total} stories implemented, reviewed, and committed across {layers} dependency layers.
```

---

## Resume Behavior

When invoked for an epic that's already `in-progress`, the workflow detects the current state from `sprint-status.yaml` and resumes:

| Story State | Action |
|---|---|
| `backlog` | Include in Phase A |
| `ready-for-dev` | Include in Phase B (dev + review) |
| `in-progress` | Include in Phase B (dev-story resumes from last task) |
| `review` | Include in Phase B (code-review, or dev-story if review follow-ups exist) |
| `done` | Skip |

The dependency graph is re-parsed on resume. Stories whose dependencies aren't yet `done` are deferred until their dependencies complete — the topological sort handles this naturally.

### Orphaned Worktree and Kind Cluster Detection

On resume, **always** run `git worktree list` and check for orphaned worktrees from a previous interrupted run (any worktree whose path contains `epic-{N}` or `story-{N}`).

Also check for orphaned Kind clusters by running `kind get clusters` and looking for any cluster whose name starts with `vault-config-operator-s` (the per-worktree naming convention).

If orphaned worktrees or Kind clusters are found:
1. Report them to the user with their paths/names and HEAD commits
2. Ask the user to choose:
   - **Resume**: keep the worktrees and clusters and continue from where they left off
   - **Clean up**: remove the worktrees and clusters and restart the affected stories from `ready-for-dev`
3. If cleaning up:
   ```
   git worktree remove --force <worktree-path>
   git branch -D <branch-name>   # only if the branch was not merged
   make kind-teardown KIND_CLUSTER_NAME=<cluster-name>   # for each orphaned cluster
   ```
4. Verify with `git worktree list` that only the main working tree remains
5. Verify with `kind get clusters` that no orphaned per-story clusters remain

### Story File Status Consistency

On resume, also verify that story file `Status:` fields are consistent with `sprint-status.yaml`. If a story is marked `done` in sprint-status but its file shows a different status (e.g., `review`, `in-progress`), fix the story file to match sprint-status before proceeding.
