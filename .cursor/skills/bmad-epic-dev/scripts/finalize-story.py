#!/usr/bin/env python3
"""Atomic Phase B story finalization for bmad-epic-dev.

Marks a story done in BOTH the story file and sprint-status.yaml, or in
neither. Refuses if the story file has no Code Review Record with a
non-empty Review Model Used.

This is the code-level gate that documentation-only HARD GUARD rules
failed to enforce (Epics 12–15). Orchestrators MUST run this script
instead of hand-editing sprint-status.yaml to `done`.

Usage:
    python3 finalize-story.py <story-id> [--impl-dir <path>]
    python3 finalize-story.py <story-id> --check [--impl-dir <path>]
    python3 finalize-story.py --check-epic <epic-num> [--impl-dir <path>]

story-id may be a story number (16.1, 16-1) or a full sprint-status key.
"""

# /// script
# requires-python = ">=3.9"
# ///

from __future__ import annotations

import argparse
import json
import re
import sys
from datetime import date
from pathlib import Path


STATUS_LINE_RE = re.compile(r"^Status:\s*(\S+)\s*$", re.MULTILINE)
REVIEW_HEADING_RE = re.compile(r"^##\s+Code Review Record\s*$", re.MULTILINE)
REVIEW_MODEL_HEADING_RE = re.compile(
    r"^###\s+Review Model Used\s*$", re.MULTILINE
)
DEV_STATUS_KEY_RE = re.compile(r"^(\s+)([\w.-]+)(\s*:\s*)(\S+)(.*)$")


class FinalizeError(Exception):
    """User-facing finalization failure."""


def normalize_story_id(raw: str) -> str:
    return raw.strip().replace(".", "-")


def is_story_key_for(normalized: str, key: str) -> bool:
    """True if sprint-status key belongs to story number `normalized`.

    Distinguishes 16-1 from 16-10, and 7-5 from 7-5-1.
    """
    if key == normalized:
        return True
    if not key.startswith(normalized):
        return False
    rest = key[len(normalized) :]
    if not rest.startswith("-"):
        return False
    next_seg = rest[1:].split("-", 1)[0]
    if re.fullmatch(r"[0-9]+[a-z]?", next_seg):
        return False
    return True


def parse_development_status_keys(text: str) -> dict[str, str]:
    """Return development_status keys → first token of value (status)."""
    keys: dict[str, str] = {}
    in_dev = False
    for line in text.splitlines():
        stripped = line.strip()
        if stripped.startswith("development_status:"):
            in_dev = True
            continue
        if not in_dev:
            continue
        if stripped.startswith("action_items:"):
            break
        if not stripped or stripped.startswith("#"):
            continue
        match = DEV_STATUS_KEY_RE.match(line)
        if not match:
            continue
        key, raw_value = match.group(2), match.group(4)
        status = raw_value.split()[0] if raw_value else ""
        keys[key] = status
    return keys


def find_sprint_status_key(normalized: str, keys: dict[str, str]) -> str:
    matches = [k for k in keys if is_story_key_for(normalized, k)]
    if len(matches) == 1:
        return matches[0]
    if normalized in keys:
        return normalized
    if not matches:
        raise FinalizeError(
            f"No sprint-status key found for story '{normalized}'"
        )
    raise FinalizeError(
        f"Ambiguous story id '{normalized}' matches: {', '.join(sorted(matches))}"
    )


def find_story_file(impl_dir: Path, key: str, normalized: str) -> Path:
    exact = impl_dir / f"{key}.md"
    if exact.exists():
        return exact
    candidates = []
    for path in impl_dir.glob("*.md"):
        if is_story_key_for(normalized, path.stem) or path.stem == key:
            candidates.append(path)
    if len(candidates) == 1:
        return candidates[0]
    if not candidates:
        raise FinalizeError(
            f"No story file found in {impl_dir} for '{normalized}'"
        )
    raise FinalizeError(
        f"Ambiguous story file for '{normalized}': "
        + ", ".join(p.name for p in sorted(candidates))
    )


def read_story_status(text: str) -> str | None:
    match = STATUS_LINE_RE.search(text)
    return match.group(1) if match else None


def review_record_ok(text: str) -> tuple[bool, str]:
    if not REVIEW_HEADING_RE.search(text):
        return False, "story file is missing '## Code Review Record'"
    model_heading = REVIEW_MODEL_HEADING_RE.search(text)
    if not model_heading:
        return False, "Code Review Record is missing '### Review Model Used'"
    after = text[model_heading.end() :]
    for line in after.splitlines():
        stripped = line.strip()
        if not stripped:
            continue
        if stripped.startswith("#"):
            return False, "Review Model Used is empty"
        return True, stripped
    return False, "Review Model Used is empty"


def set_story_status(text: str, status: str) -> str:
    if not STATUS_LINE_RE.search(text):
        raise FinalizeError("story file has no 'Status:' line")
    return STATUS_LINE_RE.sub(f"Status: {status}", text, count=1)


def set_sprint_status_key(text: str, key: str, status: str) -> str:
    in_dev = False
    found = False
    lines = text.splitlines(keepends=True)
    out: list[str] = []
    for line in lines:
        stripped = line.strip()
        if stripped.startswith("development_status:"):
            in_dev = True
            out.append(line)
            continue
        if in_dev and stripped.startswith("action_items:"):
            in_dev = False
        if in_dev:
            match = DEV_STATUS_KEY_RE.match(line.rstrip("\n"))
            if match and match.group(2) == key:
                indent, name, sep, _old, rest = match.groups()
                newline = "\n" if line.endswith("\n") else ""
                out.append(f"{indent}{name}{sep}{status}{rest}{newline}")
                found = True
                continue
        out.append(line)
    if not found:
        raise FinalizeError(f"sprint-status key '{key}' not found under development_status")
    return "".join(out)


def stamp_last_updated(text: str, today: str) -> str:
    return re.sub(
        r"^last_updated:\s*\S+",
        f"last_updated: {today}",
        text,
        count=1,
        flags=re.MULTILINE,
    )


def inspect_story(impl_dir: Path, story_id: str) -> dict:
    sprint_path = impl_dir / "sprint-status.yaml"
    if not sprint_path.exists():
        raise FinalizeError(f"sprint-status.yaml not found in {impl_dir}")
    sprint_text = sprint_path.read_text(encoding="utf-8")
    keys = parse_development_status_keys(sprint_text)
    normalized = normalize_story_id(story_id)
    key = find_sprint_status_key(normalized, keys)
    story_path = find_story_file(impl_dir, key, normalized)
    story_text = story_path.read_text(encoding="utf-8")
    story_status = read_story_status(story_text)
    ok, model_or_err = review_record_ok(story_text)
    return {
        "sprint_path": sprint_path,
        "sprint_text": sprint_text,
        "story_path": story_path,
        "story_text": story_text,
        "key": key,
        "story_status": story_status,
        "sprint_status": keys[key],
        "review_ok": ok,
        "review_model": model_or_err if ok else "",
        "review_error": "" if ok else model_or_err,
    }


def check_story(info: dict) -> list[str]:
    errors: list[str] = []
    if info["story_status"] is None:
        errors.append(f"{info['story_path'].name}: missing Status: line")
    if not info["review_ok"]:
        errors.append(f"{info['story_path'].name}: {info['review_error']}")
    if info["sprint_status"] == "done" and info["story_status"] != "done":
        errors.append(
            f"drift: sprint-status '{info['key']}' is done but "
            f"{info['story_path'].name} Status is {info['story_status']!r}"
        )
    if info["story_status"] == "done" and info["sprint_status"] != "done":
        errors.append(
            f"drift: {info['story_path'].name} is done but "
            f"sprint-status '{info['key']}' is {info['sprint_status']!r}"
        )
    return errors


def finalize(impl_dir: Path, story_id: str, today: str | None = None) -> dict:
    info = inspect_story(impl_dir, story_id)
    if not info["review_ok"]:
        raise FinalizeError(
            f"Cannot mark '{info['key']}' done: {info['review_error']}. "
            "Populate Code Review Record first, then re-run."
        )
    if info["story_status"] is None:
        raise FinalizeError(
            f"{info['story_path'].name} has no Status: line"
        )

    story_text = info["story_text"]
    sprint_text = info["sprint_text"]
    changed = []

    if info["story_status"] != "done":
        story_text = set_story_status(story_text, "done")
        info["story_path"].write_text(story_text, encoding="utf-8")
        changed.append("story_status")

    if info["sprint_status"] != "done":
        sprint_text = set_sprint_status_key(sprint_text, info["key"], "done")
        sprint_text = stamp_last_updated(sprint_text, today or date.today().isoformat())
        info["sprint_path"].write_text(sprint_text, encoding="utf-8")
        changed.append("sprint_status")

    return {
        "status": "ok",
        "key": info["key"],
        "story_file": info["story_path"].name,
        "review_model": info["review_model"],
        "changed": changed,
        "already_done": not changed,
    }


def iter_epic_story_keys(keys: dict[str, str], epic_num: str) -> list[str]:
    prefix = epic_num.replace(".", "-") + "-"
    epic_key = f"epic-{epic_num}"
    retro_key = f"epic-{epic_num}-retrospective"
    out = []
    for key in keys:
        if key in (epic_key, retro_key):
            continue
        if key.startswith(prefix) or key.startswith(epic_num.replace(".", "-") + "-"):
            out.append(key)
    return out


def check_epic(impl_dir: Path, epic_num: str) -> dict:
    sprint_path = impl_dir / "sprint-status.yaml"
    if not sprint_path.exists():
        raise FinalizeError(f"sprint-status.yaml not found in {impl_dir}")
    keys = parse_development_status_keys(sprint_path.read_text(encoding="utf-8"))
    story_keys = iter_epic_story_keys(keys, str(epic_num))
    if not story_keys:
        raise FinalizeError(f"No stories found for epic {epic_num}")

    errors: list[str] = []
    checked: list[dict] = []
    for key in story_keys:
        info = inspect_story(impl_dir, key)
        story_errors = check_story(info)
        checked.append(
            {
                "key": key,
                "story_status": info["story_status"],
                "sprint_status": info["sprint_status"],
                "review_ok": info["review_ok"],
                "errors": story_errors,
            }
        )
        # Only enforce atomicity for stories that claim to be done on either side.
        if info["sprint_status"] == "done" or info["story_status"] == "done":
            errors.extend(story_errors)

    return {
        "status": "ok" if not errors else "error",
        "epic": str(epic_num),
        "stories": checked,
        "errors": errors,
    }


def _print(data: dict) -> None:
    print(json.dumps(data, indent=2))


def main() -> None:
    parser = argparse.ArgumentParser(description="Atomic Phase B story finalization")
    parser.add_argument("story_id", nargs="?", help="Story number (16.1) or sprint-status key")
    parser.add_argument("--check", action="store_true", help="Verify only; do not write")
    parser.add_argument("--check-epic", metavar="EPIC", help="Verify all done stories in an epic")
    parser.add_argument("--impl-dir", default=None, help="Implementation artifacts directory")
    args = parser.parse_args()

    project_root = Path.cwd()
    impl_dir = (
        Path(args.impl_dir)
        if args.impl_dir
        else project_root / "_bmad-output" / "implementation-artifacts"
    )

    try:
        if args.check_epic:
            result = check_epic(impl_dir, args.check_epic)
            _print(result)
            sys.exit(0 if result["status"] == "ok" else 1)

        if not args.story_id:
            parser.error("story_id is required unless --check-epic is set")

        if args.check:
            info = inspect_story(impl_dir, args.story_id)
            errors = check_story(info)
            result = {
                "status": "ok" if not errors else "error",
                "key": info["key"],
                "story_file": info["story_path"].name,
                "story_status": info["story_status"],
                "sprint_status": info["sprint_status"],
                "review_ok": info["review_ok"],
                "review_model": info["review_model"] or None,
                "errors": errors,
            }
            _print(result)
            sys.exit(0 if not errors else 1)

        result = finalize(impl_dir, args.story_id)
        _print(result)
        sys.exit(0)
    except FinalizeError as exc:
        _print({"status": "error", "errors": [str(exc)]})
        sys.exit(1)


if __name__ == "__main__":
    main()
