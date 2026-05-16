---
name: bmad-epic-dev
description: "Orchestrate full epic development end-to-end. Use when the user says 'dev this epic' or 'implement epic [number]'."
---

# Epic Dev

## Overview

This skill drives an entire epic from backlog to done — creating story specs, implementing each story, and running adversarial code review — all with fresh LLM context per step. It parses the epic's dependency tree to optimize execution order and parallelize where safe.

Act as a disciplined build orchestrator. You don't write code or specs yourself — you delegate to specialized skills (`bmad-create-story`, `bmad-dev-story`, `bmad-code-review`) via subagents, track progress, and relay decisions back to the user.

**Args:**
- Epic number (e.g., `13`) — required
- `--commit-policy=<auto|ask|skip>` — controls git commit behavior (default: `auto`)
  - `auto`: commit automatically at prescribed points
  - `ask`: prompt user before each commit
  - `skip`: never commit (user handles commits manually)
- `--skip-confirmations` — suppress soft-gate pauses between phases (for experienced users resuming a known-good epic)

**Model configuration:**

Different phases use different LLMs by design — code review must run on a different model than the one that wrote the code to avoid self-review bias.

| Phase | Role | Model | Reasoning |
|---|---|---|---|
| Step 0 (planning) | Design & decisioning | Opus 4.6 | High reasoning for dependency analysis and orchestration |
| Phase A (create-story) | Story spec creation | Opus 4.6 | High reasoning for comprehensive context extraction |
| Phase B (dev-story) | Implementation | Opus 4.6 | High reasoning for code generation and problem solving |
| Phase B (code-review) | Adversarial review | ChatGPT 5.4 | Medium reasoning, different model prevents self-review blind spots |

When spawning subagents, pass the appropriate model. If a specified model isn't available in the environment, warn the user and fall back to whatever model is available — but always log when code review runs on the same model as development.

**Core contract:**
- Every sub-skill runs in a **fresh subagent** (clean LLM context)
- Sub-skill subagents run autonomously but **relay "decision needed" questions** back to you for the user
- **Code review subagents MUST use a different LLM model** than the development subagents
- On any failure: **halt with resume instructions** — re-invoking with the same epic number picks up where it left off
- `sprint-status.yaml` is the checkpoint — story statuses determine what's been done and what remains
- Interactive by default: report progress after each story, soft gates between phases
- Phase A (story specs) parallelizes by dependency layer via concurrent subagents
- Phase B (implementation) parallelizes by dependency layer via **git worktrees** — each story gets its own branch and working directory, merged back after review

## On Activation

Load available config from `{project-root}/_bmad/config.yaml` and `{project-root}/_bmad/config.user.yaml` (root level and `bmm` section). If config is missing, let the user know `bmad-bmb-setup` can configure the module at any time. Use sensible defaults for anything not configured.

Resolve:
- `implementation_artifacts` — where story files and sprint-status.yaml live
- `planning_artifacts` — where epic source files live
- `user_name`, `communication_language`

**Intent guard:** Before proceeding, confirm the user intends full epic orchestration. If the input looks like a single story reference (e.g., "13.2" without "epic"), clarify: "Did you mean to implement the full epic, or just story 13.2? For a single story, use `bmad-dev-story` instead."

Then read fully and follow `./references/phase-logic.md`.
