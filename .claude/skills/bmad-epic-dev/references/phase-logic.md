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

**Subagent type:** Use `best-of-n-runner` subagents when available — they run in isolated git worktrees automatically. Fall back to sequential `generalPurpose` subagents on the main branch if worktree subagents aren't available.

### Integration Test Isolation Protocol

When multiple stories run integration tests in parallel (each in its own worktree), each worktree **must** use its own Kind cluster to avoid race conditions on shared Kubernetes namespaces, Vault state, and host port bindings.

**Port and cluster assignment:** Before launching parallel worktrees for a dependency layer, the orchestrator assigns each story a unique port offset based on its position within the layer (0-indexed):

| Variable | Formula | Example (story index 0, 1, 2) |
|---|---|---|
| `KIND_CLUSTER_NAME` | `vault-config-operator-{story_key}` | `vault-config-operator-s13-1` |
| `VAULT_HOST_PORT` | `8200 + (layer_story_index + 1) * 10` | `8210`, `8220`, `8230` |
| `HTTPS_HOST_PORT` | `VAULT_HOST_PORT + 1` | `8211`, `8221`, `8231` |

The base port `8200` (offset 0) is reserved for the default cluster (`vault-config-operator`) and is never assigned to a worktree. The formula assumes single-epic parallelism; concurrent epics could collide on ports.

**How to pass overrides:** Each `best-of-n-runner` subagent prompt must include an environment instruction block that the dev-story subagent applies when running `make integration`:

```
INTEGRATION TEST ISOLATION:
When running integration tests (make integration), use these overrides:
  KIND_CLUSTER_NAME={assigned_cluster_name}
  VAULT_HOST_PORT={assigned_vault_port}
  HTTPS_HOST_PORT={assigned_https_port}
Run as: make integration KIND_CLUSTER_NAME={assigned_cluster_name} VAULT_HOST_PORT={assigned_vault_port} HTTPS_HOST_PORT={assigned_https_port}
```

**Sequential fallback:** When stories run sequentially on the main branch (B.5), no overrides are needed — the default cluster name and ports are used.

For each dependency layer (in topological order):

### B.1: Launch Parallel Development

1. Identify stories in this layer where status is `ready-for-dev` (or `in-progress`/`review` for resumption).

2. For each story, the orchestrator manages a multi-step pipeline in the story's worktree. Each story gets its own worktree branch (e.g., `epic-{N}/story-{N.M}`).

   **Step 1 — IMPLEMENT** (model: **Opus 4.6**): Spawn a `best-of-n-runner` subagent for `bmad-dev-story`:
   ```
   Run the bmad-dev-story skill for story {epicNum}.{storyNum}.
   The story file is at: {implementation_artifacts}/{story_key}.md
   Process the story fully — implement all tasks, run all tests, mark complete.
   If you encounter a blocking issue requiring a human decision, describe it
   clearly and halt. Do NOT make assumptions on behalf of the user.

   INTEGRATION TEST ISOLATION:
   When running integration tests (make integration), use these overrides:
     KIND_CLUSTER_NAME={assigned_cluster_name}
     VAULT_HOST_PORT={assigned_vault_port}
     HTTPS_HOST_PORT={assigned_https_port}
   Run as: make integration KIND_CLUSTER_NAME={assigned_cluster_name} VAULT_HOST_PORT={assigned_vault_port} HTTPS_HOST_PORT={assigned_https_port}

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

   **CRITICAL:** Do NOT commit after applying fixes without a re-review pass. Every fix must be verified by a fresh code review subagent before it can be committed. Skipping re-review defeats the purpose of adversarial review — fix subagents can introduce new issues that only a second review would catch.

   **Step 4 — COMMIT** (respecting `--commit-policy`):
   ```
   git add -A && git commit -m "feat(epic-{N}): {story_key} — {title}"
   ```
   Report: story_key, review_patches_count, files_changed_count, commit_sha.

   Steps 1-4 run sequentially per story, but **multiple stories in the same layer run their pipelines in parallel** (each in its own worktree).

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

6. After all stories in the layer succeed, merge each story branch back to the main branch **sequentially** (to maintain a clean linear history):

   For each completed story branch:
   ```
   git checkout {epic-branch}
   git merge --no-ff epic-{N}/story-{N.M} -m "Merge story {N.M}: {title}"
   ```

   **CRITICAL:** Always use `--no-ff` to create a merge commit, even if fast-forward is possible. This preserves the branch history and makes it clear which commits belong to which story.

   If merge conflicts occur: halt with details. The user resolves conflicts manually, then re-invokes to resume. The conflicting branch is preserved for inspection.

7. **Update story file status on the main branch.** After each successful merge, on the main branch:
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

8. Clean up worktrees and Kind clusters after successful merge:

   **Worktree cleanup:**
   ```
   git worktree remove <worktree-path>
   git branch -d epic-{N}/story-{N.M}
   ```

   **Verify cleanup:** Run `git worktree list` and confirm no orphaned worktrees remain for the merged story. If removal fails (e.g., untracked files), force-remove:
   ```
   git worktree remove --force <worktree-path>
   ```

   **Kind cluster cleanup:** If the story used a dedicated Kind cluster (parallel worktree execution), tear it down to free resources:
   ```
   make kind-teardown KIND_CLUSTER_NAME={assigned_cluster_name}
   ```
   This deletes the Kind cluster and removes its kubeconfig file. Skip this step for the default cluster (`vault-config-operator`) or when running in sequential fallback mode.

9. Update `sprint-status.yaml`: mark each merged story as `done`.

10. **Final consistency check:** After all stories in the layer are merged, verify that:
    - Every merged story's file has `Status: done`
    - Every merged story in `sprint-status.yaml` is `done`
    - `git worktree list` shows only the main working tree (no orphaned worktrees)
    If any inconsistency is found, fix it before proceeding to the next layer.

### B.4: Report Layer Progress

9. Report to the user:
   ```
   Layer {L} complete: {count} stories merged
   {story_list with review patch counts and file counts}
   Progress: {done}/{total} stories in epic {N}
   ```

10. Continue to the next dependency layer.

### B.5: Sequential Fallback

If `best-of-n-runner` subagents aren't available, process stories **sequentially on the main branch** within each layer:

For each story:
1. Spawn a `generalPurpose` subagent for `bmad-dev-story`
2. If `decision_needed`: follow the **Decision Relay Protocol** — present to user, wait, resume subagent
3. On success, spawn a `generalPurpose` subagent for `bmad-code-review` (model: **ChatGPT 5.4**)
4. If `decision_needed`: follow the **Decision Relay Protocol**
5. If `changes_requested`: re-run dev-story (Opus 4.6) to address findings, then **ALWAYS** re-run code-review (ChatGPT 5.4) — never skip the re-review, even if the fixes seem trivial. Repeat until approved or halted.
6. **Update story file** on the main branch:
   - Set `Status:` to `done`
   - Populate the **Code Review Record** section (Review Model, Findings, Decisions, Fixes)
7. Commit (respecting `--commit-policy`):
   ```
   feat(epic-{N}): {story_key} — {story title}
   ```
8. Update `sprint-status.yaml`: mark story as `done`
9. **Verify consistency:** Confirm story file status and sprint-status agree before continuing
10. Report progress, continue to next story

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
