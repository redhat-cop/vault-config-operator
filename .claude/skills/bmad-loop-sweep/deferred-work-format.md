# Deferred Work Format

Canonical entry format for `{implementation_artifacts}/deferred-work.md`. On the
inner dev path the bmad-dev-auto session appends its own flat entries (review
defers, multi-goal splits, token splits); the orchestrator owns the ledger and
normalizes those flat entries into this canonical form on sweep, and a
`bmad-loop sweep` migration rewrites freeform pre-DW-format content from older
projects into it wholesale (see `./migration-mode.md`; the TUI displays such
legacy items read-only until that happens). The file is append-only — never
rewrite or delete existing entries.

## Before appending: dedupe check

Scan the existing file for an entry describing the same issue or goal (same
location and same substance, even if worded differently). If one exists, do
NOT append a duplicate — add a `seen-again:` line to the existing entry
instead:

```markdown
seen-again: 2026-06-12 (code review of spec-3-3-export.md)
```

## Entry format

Number entries sequentially (`DW-1`, `DW-2`, …) by scanning the file for the
highest existing number. One entry per deferred item:

```markdown
### DW-<seq>: <one-line title>

origin: <workflow + artifact + date, e.g. "code review of spec-3-2-digest.md, 2026-06-12">
location: <file:line or component, or "n/a" for deferred goals>
severity: <critical | high | medium | low — how much it matters if never done>
reason: <why this was deferred rather than done now, one or two sentences>
status: open
```

`location:` is always written. Use `n/a` whenever there is nothing to open — a
deferred goal, but equally a finding whose reporter recorded no place. The field
says "no location was recorded", not "this item has none": a reader that finds
`n/a` should fall back to `reason:`, which often names the file even when
`location:` is empty. Entries written before this rule may omit the line; read
an absent `location:` as `n/a`, never as "not yet known".

Entries the orchestrator harvests from a spec carry one extra line,
`source_spec:`, directly after `location:` — the spec the deferral came from. It
is half the dedupe key (with `origin:`), so never edit or drop it when touching
an entry; entries written by hand do not need it.

**Every field line is exactly one line, and so is the title.** The format is
line-oriented: readers find each field by scanning for `<name>:` at the start of
a line, and an entry ends at whichever comes first — the next `### DW-<n>`
entry, any other `#` .. `######` heading, or a `- source_spec:` flat-append
bullet. A value carrying a line break therefore does not wrap; it becomes new
ledger content, and three things can follow:

- a break followed by `### ` mints an entry nobody filed;
- a break before a `status:` line leaves one entry carrying two, so the ledger
  no longer says one thing about it;
- a break followed by `- source_spec:` cuts the entry short at that bullet, and
  everything after it re-surfaces as a phantom _legacy_ item.

Keep breaks out of field values, along with `### ` and a leading
`- source_spec:`. If a reason needs two sentences, write them on one line.

`severity:` is optional — entries written before this field existed have none
and that is fine; readers must treat a missing or unrecognized value as
"unspecified". Use `critical` for correctness/security issues, `high` for
likely user-visible problems, `medium` for quality and robustness gaps, `low`
for polish and nice-to-haves.

When a deferred item is later completed, set its `status:` to `done` with the
date (e.g. `status: done 2026-06-20`) — do not delete the entry.

## Sweep annotations

`bmad-loop sweep` runs (the orchestrator and its bundle dev sessions) add two
optional field lines to existing entries — both directly after `status:`:

```markdown
resolution: <one line: what was built or why the entry was closed>
decision: <date> <chosen option label> — <detail>
```

- `resolution:` accompanies every sweep close (`status: done <date>`). Bundle
  dev sessions write it when finishing a bundle's entries; the orchestrator
  writes it when closing entries triage proved already resolved.
- `decision:` records a human's sweep-time choice on an entry. It does not by
  itself change `status:` — a `keep-open` decision leaves the entry open.

## Closure declared by a story

A sweep bundle is not the only thing that closes an entry. A regular story may
declare the entries its work closes — on its `stories.yaml` entry (stories mode),
or in its story spec's frontmatter. The two are unioned:

```yaml
closes_deferred: [DW-5, DW-6] # DW-<n> ids this story closes
```

Both are written by a human, and breakdown time — with this file open — is where
it belongs, though not a deadline: the declaration is read when the story
commits, so one added to a spec's frontmatter mid-run still counts. No upstream
skill emits the field yet, and re-deriving `stories.yaml` will drop it unless the
intent is recorded in `.memlog.md` first.

When the story commits, the orchestrator annotates each declared id exactly as a
bundle close does — `status: done <date>` plus a `resolution:` line naming the
story:

```markdown
status: done 2026-07-23
resolution: resolved by story 3-2-export
```

The rules that keep this safe:

- **Declared, never inferred.** Closure comes only from this field; the
  orchestrator does not guess it from a diff.
- **Only once the story actually lands.** The annotation is written at the
  commit boundary — after verification, the review loop and every checkpoint,
  and just before the story's commit is squashed. A story that fails, blocks, is
  rejected by review, or escalates closes nothing; a commit that then fails
  takes the annotation back with it, restored to the pre-close text.
- **In the story's own commit**, when this file lives inside the repo —
  worktree isolation included: the unit's copy rides the unit commit and
  reaches the target branch with the merge. If the artifacts dir is configured
  outside the repo, the file is shared between worktrees and no commit can
  carry it; the annotation is written all the same and the run journals
  `deferred-close-external-ledger`. A location that cannot be read or written
  when the write comes due closes nothing and is journaled — the entries stay
  `open` for a sweep to re-verify, and an outage is never read as "no such
  entries", never allowed to fail the story or crash the run.
- **Idempotent.** An id already `done` is left untouched, so a resumed run
  re-driving the same close neither doubles the `resolution:` line nor warns.
- **Never a gate.** An id that matches no entry, an entry whose `status:` reads
  as neither `open` nor `done`, and a story spec declaring a bare
  `closes_deferred: DW-5` where a list belongs are each journaled and dropped —
  none can fail the story. `bmad-loop validate` reports the same mismatches as
  warnings before the run starts. The one exception is that same wrong container
  in `stories.yaml`: the manifest is a schema the parser owns, so it fails to
  load there like any other field of the wrong type — before any story runs, and
  reported by `validate` up front.
- **Read at the commit.** The declaration that counts is the one on disk when the
  story commits, not the one it was implemented from — edit it late and the edit
  is honored, in both directions.

Keep the ids stable when editing this file: a reworded title is fine, but
renumbering an entry orphans any declaration that already references it.
