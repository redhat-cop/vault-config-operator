#!/usr/bin/env python3
"""Unit tests for finalize-story.py."""

# /// script
# requires-python = ">=3.9"
# ///

from __future__ import annotations

import textwrap
import unittest
from pathlib import Path
from tempfile import TemporaryDirectory

from importlib.util import module_from_spec, spec_from_file_location

_spec = spec_from_file_location(
    "finalize_story",
    str(Path(__file__).resolve().parent.parent / "finalize-story.py"),
)
fs = module_from_spec(_spec)
_spec.loader.exec_module(fs)


SPRINT_STATUS = textwrap.dedent("""\
    generated: 2026-04-06
    last_updated: 2026-08-01
    project: test-project

    development_status:
      epic-16: in-progress
      16-1-radius-auth: review
      16-10-other: done
      7-5-drift: review
      7-5-1-ldap: done
      epic-16-retrospective: optional

    action_items:
      - epic: 15
        action: "example"
        status: open
""")

STORY_READY = textwrap.dedent("""\
    # Story 16.1: RADIUS

    Status: review

    ## Code Review Record

    ### Review Model Used

    gpt-5.4-medium

    ### Review Findings

    - none
""")

STORY_NO_RECORD = textwrap.dedent("""\
    # Story 16.1: RADIUS

    Status: review

    ## Dev Agent Record

    Done implementing.
""")

STORY_EMPTY_MODEL = textwrap.dedent("""\
    # Story 16.1: RADIUS

    Status: review

    ## Code Review Record

    ### Review Model Used

    ### Review Findings

    - none
""")


def write_impl(d: str, sprint: str, stories: dict[str, str]) -> Path:
    impl = Path(d)
    (impl / "sprint-status.yaml").write_text(sprint)
    for name, body in stories.items():
        (impl / f"{name}.md").write_text(body)
    return impl


class TestKeyMatching(unittest.TestCase):
    def test_distinguishes_16_1_from_16_10(self):
        self.assertTrue(fs.is_story_key_for("16-1", "16-1-radius-auth"))
        self.assertFalse(fs.is_story_key_for("16-1", "16-10-other"))

    def test_distinguishes_7_5_from_7_5_1(self):
        self.assertTrue(fs.is_story_key_for("7-5", "7-5-drift"))
        self.assertFalse(fs.is_story_key_for("7-5", "7-5-1-ldap"))
        self.assertTrue(fs.is_story_key_for("7-5-1", "7-5-1-ldap"))

    def test_10_0_vs_10_0a(self):
        self.assertTrue(fs.is_story_key_for("10-0", "10-0-migrate"))
        self.assertFalse(fs.is_story_key_for("10-0", "10-0a-kustomize"))
        self.assertTrue(fs.is_story_key_for("10-0a", "10-0a-kustomize"))


class TestFinalize(unittest.TestCase):
    def test_sets_both_statuses_when_review_record_present(self):
        with TemporaryDirectory() as d:
            impl = write_impl(d, SPRINT_STATUS, {"16-1-radius-auth": STORY_READY})
            result = fs.finalize(impl, "16.1", today="2026-08-19")

            self.assertEqual(result["status"], "ok")
            self.assertEqual(result["changed"], ["story_status", "sprint_status"])
            story = (impl / "16-1-radius-auth.md").read_text()
            self.assertIn("Status: done", story)
            sprint = (impl / "sprint-status.yaml").read_text()
            self.assertIn("16-1-radius-auth: done", sprint)
            self.assertIn("16-10-other: done", sprint)
            self.assertIn("last_updated: 2026-08-19", sprint)
            self.assertIn("action_items:", sprint)

    def test_refuses_without_review_record(self):
        with TemporaryDirectory() as d:
            impl = write_impl(d, SPRINT_STATUS, {"16-1-radius-auth": STORY_NO_RECORD})
            with self.assertRaises(fs.FinalizeError) as ctx:
                fs.finalize(impl, "16-1")
            self.assertIn("Code Review Record", str(ctx.exception))
            story = (impl / "16-1-radius-auth.md").read_text()
            self.assertIn("Status: review", story)
            sprint = (impl / "sprint-status.yaml").read_text()
            self.assertIn("16-1-radius-auth: review", sprint)

    def test_refuses_empty_review_model(self):
        with TemporaryDirectory() as d:
            impl = write_impl(d, SPRINT_STATUS, {"16-1-radius-auth": STORY_EMPTY_MODEL})
            with self.assertRaises(fs.FinalizeError):
                fs.finalize(impl, "16-1")

    def test_idempotent_when_already_done(self):
        sprint = SPRINT_STATUS.replace("16-1-radius-auth: review", "16-1-radius-auth: done")
        story = STORY_READY.replace("Status: review", "Status: done")
        with TemporaryDirectory() as d:
            impl = write_impl(d, sprint, {"16-1-radius-auth": story})
            result = fs.finalize(impl, "16-1")
            self.assertEqual(result["changed"], [])
            self.assertTrue(result["already_done"])

    def test_does_not_touch_16_10_when_finalizing_16_1(self):
        with TemporaryDirectory() as d:
            impl = write_impl(
                d,
                SPRINT_STATUS,
                {
                    "16-1-radius-auth": STORY_READY,
                    "16-10-other": STORY_READY.replace("16.1", "16.10"),
                },
            )
            fs.finalize(impl, "16.1", today="2026-08-19")
            sprint = (impl / "sprint-status.yaml").read_text()
            self.assertRegex(sprint, r"16-1-radius-auth: done")
            self.assertRegex(sprint, r"16-10-other: done")


class TestCheck(unittest.TestCase):
    def test_check_reports_drift(self):
        sprint = SPRINT_STATUS.replace("16-1-radius-auth: review", "16-1-radius-auth: done")
        with TemporaryDirectory() as d:
            impl = write_impl(d, sprint, {"16-1-radius-auth": STORY_READY})
            info = fs.inspect_story(impl, "16.1")
            errors = fs.check_story(info)
            self.assertTrue(any("drift" in e for e in errors))

    def test_check_epic_passes_when_in_progress_has_no_record(self):
        with TemporaryDirectory() as d:
            impl = write_impl(
                d,
                SPRINT_STATUS,
                {
                    "16-1-radius-auth": STORY_NO_RECORD,
                    "16-10-other": STORY_READY.replace("Status: review", "Status: done").replace("16.1", "16.10"),
                },
            )
            result = fs.check_epic(impl, "16")
            self.assertEqual(result["status"], "ok")

    def test_check_epic_fails_on_done_without_record(self):
        sprint = SPRINT_STATUS.replace("16-1-radius-auth: review", "16-1-radius-auth: done")
        with TemporaryDirectory() as d:
            impl = write_impl(
                d,
                sprint,
                {
                    "16-1-radius-auth": STORY_NO_RECORD.replace("Status: review", "Status: done"),
                    "16-10-other": STORY_READY.replace("Status: review", "Status: done").replace("16.1", "16.10"),
                },
            )
            result = fs.check_epic(impl, "16")
            self.assertEqual(result["status"], "error")
            self.assertTrue(any("Code Review Record" in e for e in result["errors"]))


if __name__ == "__main__":
    unittest.main()
