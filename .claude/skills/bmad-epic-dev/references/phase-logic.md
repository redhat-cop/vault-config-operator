commit c4164bbdc5898d252878314008cbbc57373c827c
Author: Raffaele Spazzoli <raffaele.spazzoli@gmail.com>
Date:   Sun Aug 2 18:18:53 2026 +0530

    chore(epic-10): improve bmad-epic-dev workflow and add review tracking
    
    Add Decision Relay Protocol to bmad-epic-dev phase-logic with explicit
    mechanics for detecting decision_needed, relaying to user via AskQuestion,
    resuming subagents, and recording decisions in story files.
    
    Strengthen worktree merge process: enforce --no-ff merges, add post-merge
    story file status updates, worktree cleanup verification, orphaned worktree
    detection on resume, and story file status consistency checks.
    
    Add Code Review Record section to story template (Review Model, Findings,
    Decisions Needed/Taken, Fixes Applied) across all three template copies.
    
    Record GPT 5.4 code review findings for stories 10.0 and 10.0a.
    
    Co-authored-by: Cursor <cursoragent@cursor.com>

diff --git a/.claude/skills/bmad-epic-dev/references/phase-logic.md b/.claude/skills/bmad-epic-dev/references/phase-logic.md
index c248a45..781214e 100644
--- a/.claude/skills/bmad-epic-dev/references/phase-logic.md
+++ b/.claude/skills/bmad-epic-dev/references/phase-logic.md
@@ -5,7 +5,7 @@ Communicate with the user in `{communication_language}`.
 **Contract anchor** (survives context compaction — re-read if uncertain):
 - You are an orchestrator, not an implementer — delegate all work to subagents
 - Every sub-skill (`bmad-create-story`, `bmad-dev-story`, `bmad-code-review`) runs in a **fresh subagent** with clean LLM context
-- Subagents run autonomously but you **relay "decision needed" questions** back to the user
+- Subagents run autonomously but you **relay "decision needed" questions** back to the user — see **Decision Relay Protocol** below
 - On any failure: **halt** with what failed, which story, and instructions to re-invoke
 - `sprint-status.yaml` is the sole checkpoint — story statuses determine resume position
 - Respect `--commit-policy` and `--skip-confirmations` args from SKILL.md
@@ -13,6 +13,50 @@ Communicate with the user in `{communication_language}`.
 
 ---
 
+## Decision Relay Protocol
+
+When a subagent reports `decision_needed`, the orchestrator MUST follow this exact sequence. **Do NOT skip steps or treat the decision as resolved without user input.**
+
+### Detection
+
+After every subagent completes, parse its output for these signals:
+- The string `decision_needed` in the status field
+- Questions phrased as "should we…", "which approach…", "is it acceptable…"
+- Explicit `[Decision]` tags in review findings
+
+### Relay to User
+
+1. **Present the decision clearly** using the `AskQuestion` tool with structured options when the decision has discrete choices, or as plain text when it's open-ended. Include:
+   - Which story raised the question (story key + title)
+   - Which phase raised it (create-story / dev-story / code-review)
+   - The full question text from the subagent
+   - Any context the subagent provided (code snippets, trade-offs)
+
+2. **Wait for the user's response.** Do NOT proceed with the story pipeline until the user answers. Other stories in the same layer that don't need decisions can continue.
+
+### Resume the Subagent
+
+3. After the user responds, **resume the subagent** using the Task tool with `resume: <agent_id>`:
+   ```
+   The user decided: {user's answer}
+   Please continue with this decision applied.
+   ```
+
+4. If the resumed subagent completes successfully, continue the story pipeline (next step in B.1).
+
+5. If the resumed subagent raises another decision, repeat this protocol.
+
+### Record the Decision
+
+6. After the story pipeline completes, update the story file's **Code Review Record** section:
+   - **Decisions Needed**: the original question(s)
+   - **Decisions Taken**: the user's answer(s) with rationale if provided
+   - Record the review model in **Review Model Used**
+   - Record all findings in **Review Findings**
+   - Record all applied fixes in **Fixes Applied**
+
+---
+
 ## Step 0: Validate Input and Parse Epic
 
 **Goal:** Identify the target epic, verify it's actionable, and build the execution plan.
@@ -131,10 +175,19 @@ For each dependency layer (in topological order):
    Process the story fully — implement all tasks, run all tests, mark complete.
    If you encounter a blocking issue requiring a human decision, describe it
    clearly and halt. Do NOT make assumptions on behalf of the user.
+
+   IMPORTANT: If you need a decision, your output MUST include:
+   - status: decision_needed
+   - decision_question: <the exact question for the user>
+   - decision_context: <relevant code/trade-offs to help the user decide>
+   - decision_options: <list of concrete options if applicable>
+
    When complete, report: story_key, status (success/failed/decision_needed),
    files_changed_count, and a one-line summary.
    ```
 
+   After Step 1 completes, check for `decision_needed`. If found, follow the **Decision Relay Protocol** before proceeding.
+
    **Step 2 — CODE REVIEW** (model: **ChatGPT 5.4**): Spawn a separate subagent for `bmad-code-review` in the same worktree:
    ```
    Run the bmad-code-review skill to review the changes for story {epicNum}.{storyNum}.
@@ -142,11 +195,20 @@ For each dependency layer (in topological order):
    Run the review fully and autonomously — complete all layers, produce the
    triage report.
    If the review raises a design/requirements question (not a code fix),
-   describe it and halt.
+   describe it and halt — do NOT resolve design questions yourself.
+
+   IMPORTANT: If you need a decision, your output MUST include:
+   - status: decision_needed
+   - decision_question: <the exact question for the user>
+   - decision_context: <relevant code/trade-offs to help the user decide>
+   - decision_options: <list of concrete options if applicable>
+
    When complete, report: story_key, status (approved/changes_requested/decision_needed),
    review_patches_count, and a one-line summary.
    ```
 
+   After Step 2 completes, check for `decision_needed`. If found, follow the **Decision Relay Protocol** before proceeding.
+
    **Step 3 — REVIEW FIXES**: If code review requested changes, re-run Step 1 (Opus 4.6) to address findings, then re-run Step 2 (ChatGPT 5.4). Repeat until approved or halted.
 
    **Step 4 — COMMIT** (respecting `--commit-policy`):
@@ -161,9 +223,20 @@ For each dependency layer (in topological order):
 
 ### B.2: Handle Subagent Results
 
-4. As each subagent completes:
-   - **Success:** Note the story's branch name and commit SHA.
-   - **Decision needed:** Relay the question to the user with full context. After getting the answer, resume that subagent.
+4. As each subagent completes, parse its output and act on the status:
+
+   - **Success / Approved:** Note the story's branch name and commit SHA. Continue to the next step in the story pipeline.
+
+   - **Decision needed:** Follow the **Decision Relay Protocol**:
+     1. Extract `decision_question`, `decision_context`, and `decision_options` from the subagent output
+     2. Present to the user using `AskQuestion` (for discrete choices) or text (for open-ended questions)
+     3. **HALT the story pipeline** — do NOT proceed until the user responds
+     4. After the user responds, resume the subagent with `resume: <agent_id>` passing the user's decision
+     5. Record the decision in the story file's Code Review Record
+     6. Continue the story pipeline with the resumed subagent's output
+
+   - **Changes requested (code review):** Proceed to Step 3 (review fixes) — this is NOT a decision, it's an actionable fix list.
+
    - **Failed:** Halt the entire epic. Report which story failed, the error, and resume instructions. Other in-flight subagents for the same layer can continue to completion (fail-forward within a layer) but no new layers start.
 
 5. Wait for all subagents in the layer to finish (or halt).
@@ -174,19 +247,46 @@ For each dependency layer (in topological order):
 
    For each completed story branch:
    ```
-   git checkout main
-   git merge --no-ff epic-{N}/story-{N.M}
+   git checkout {epic-branch}
+   git merge --no-ff epic-{N}/story-{N.M} -m "Merge story {N.M}: {title}"
    ```
 
+   **CRITICAL:** Always use `--no-ff` to create a merge commit, even if fast-forward is possible. This preserves the branch history and makes it clear which commits belong to which story.
+
    If merge conflicts occur: halt with details. The user resolves conflicts manually, then re-invokes to resume. The conflicting branch is preserved for inspection.
 
-7. Clean up worktrees after successful merge:
+7. **Update story file status on the main branch.** After each successful merge, on the main branch:
+   - Set the story file's `Status:` field to `done`
+   - Populate the **Code Review Record** section if not already filled by the review subagent:
+     - **Review Model Used**: the model slug used for code review
+     - **Review Findings**: summary of findings from the review
+     - **Decisions Needed / Taken**: any decisions that were relayed to the user
+     - **Fixes Applied**: list of fixes applied after review
+   - Commit these story-file updates as part of the merge or as a follow-up:
+     ```
+     chore(epic-{N}): update story {N.M} status and review record
+     ```
+
+   This ensures story file status is always consistent with `sprint-status.yaml`, regardless of whether the work was done in a worktree.
+
+8. Clean up worktrees after successful merge:
    ```
    git worktree remove <worktree-path>
    git branch -d epic-{N}/story-{N.M}
    ```
 
-8. Update `sprint-status.yaml`: mark each merged story as `done`.
+   **Verify cleanup:** Run `git worktree list` and confirm no orphaned worktrees remain for the merged story. If removal fails (e.g., untracked files), force-remove:
+   ```
+   git worktree remove --force <worktree-path>
+   ```
+
+9. Update `sprint-status.yaml`: mark each merged story as `done`.
+
+10. **Final consistency check:** After all stories in the layer are merged, verify that:
+    - Every merged story's file has `Status: done`
+    - Every merged story in `sprint-status.yaml` is `done`
+    - `git worktree list` shows only the main working tree (no orphaned worktrees)
+    If any inconsistency is found, fix it before proceeding to the next layer.
 
 ### B.4: Report Layer Progress
 
@@ -205,13 +305,20 @@ If `best-of-n-runner` subagents aren't available, process stories **sequentially
 
 For each story:
 1. Spawn a `generalPurpose` subagent for `bmad-dev-story`
-2. On success, spawn a `generalPurpose` subagent for `bmad-code-review`
-3. If changes requested: re-run dev-story then code-review until approved
-4. Commit (respecting `--commit-policy`):
+2. If `decision_needed`: follow the **Decision Relay Protocol** — present to user, wait, resume subagent
+3. On success, spawn a `generalPurpose` subagent for `bmad-code-review` (model: **ChatGPT 5.4**)
+4. If `decision_needed`: follow the **Decision Relay Protocol**
+5. If `changes_requested`: re-run dev-story (Opus 4.6) to address findings, then re-run code-review (ChatGPT 5.4). Repeat until approved or halted.
+6. **Update story file** on the main branch:
+   - Set `Status:` to `done`
+   - Populate the **Code Review Record** section (Review Model, Findings, Decisions, Fixes)
+7. Commit (respecting `--commit-policy`):
    ```
    feat(epic-{N}): {story_key} — {story title}
    ```
-5. Update sprint-status, report progress, continue to next story
+8. Update `sprint-status.yaml`: mark story as `done`
+9. **Verify consistency:** Confirm story file status and sprint-status agree before continuing
+10. Report progress, continue to next story
 
 ### B.6: Check for New Stories
 
@@ -259,4 +366,22 @@ When invoked for an epic that's already `in-progress`, the workflow detects the
 
 The dependency graph is re-parsed on resume. Stories whose dependencies aren't yet `done` are deferred until their dependencies complete — the topological sort handles this naturally.
 
-Check for orphaned worktree branches (e.g., `epic-{N}/story-*`) from a previous interrupted run. If found, report them and ask the user whether to resume from those branches or clean them up.
+### Orphaned Worktree Detection
+
+On resume, **always** run `git worktree list` and check for orphaned worktrees from a previous interrupted run (any worktree whose path contains `epic-{N}` or `story-{N}`).
+
+If orphaned worktrees are found:
+1. Report them to the user with their paths and HEAD commits
+2. Ask the user to choose:
+   - **Resume**: keep the worktrees and continue from where they left off
+   - **Clean up**: remove the worktrees and restart the affected stories from `ready-for-dev`
+3. If cleaning up:
+   ```
+   git worktree remove --force <worktree-path>
+   git branch -D <branch-name>   # only if the branch was not merged
+   ```
+4. Verify with `git worktree list` that only the main working tree remains
+
+### Story File Status Consistency
+
+On resume, also verify that story file `Status:` fields are consistent with `sprint-status.yaml`. If a story is marked `done` in sprint-status but its file shows a different status (e.g., `review`, `in-progress`), fix the story file to match sprint-status before proceeding.
