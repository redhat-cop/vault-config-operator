#!/usr/bin/env python3
"""Unit tests for epic-plan.py."""

# /// script
# requires-python = ">=3.9"
# ///

from __future__ import annotations

import json
import sys
import textwrap
import unittest
from pathlib import Path
from tempfile import TemporaryDirectory

sys.path.insert(0, str(Path(__file__).resolve().parent.parent))

from importlib.util import module_from_spec, spec_from_file_location

_spec = spec_from_file_location(
    "epic_plan",
    str(Path(__file__).resolve().parent.parent / "epic-plan.py"),
)
epic_plan = module_from_spec(_spec)
_spec.loader.exec_module(epic_plan)

SAMPLE_SPRINT_STATUS = textwrap.dedent("""\
    generated: 2026-04-06
    last_updated: 2026-05-15
    project: test-project
    project_key: NOKEY
    tracking_system: file-system
    story_location: "{project-root}/_bmad-output/implementation-artifacts"

    development_status:
      epic-3: backlog
      3-0-first-story: backlog
      3-1-second-story: backlog
      3-2-third-story: backlog
      3-3-fourth-story: done
""")

SAMPLE_EPIC_FILE = textwrap.dedent("""\
    # Epic 3: Test Epic

    ## Stories

    | Story | Title | Dependency | Summary |
    |-------|-------|------------|---------|
    | 3.0 | First story | None | First |
    | 3.1 | Second story | 3.0 | Second |
    | 3.2 | Third story | 3.0 | Third |
    | 3.3 | Fourth story | 3.1, 3.2 | Fourth |
""")


class TestParseSprintStatus(unittest.TestCase):
    def test_extracts_epic_and_stories(self):
        with TemporaryDirectory() as d:
            p = Path(d) / "sprint-status.yaml"
            p.write_text(SAMPLE_SPRINT_STATUS)
            result = epic_plan.parse_sprint_status(p, "3")

        self.assertEqual(result["epic_status"], "backlog")
        self.assertEqual(len(result["stories"]), 4)
        self.assertEqual(result["stories"]["3-0-first-story"], "backlog")
        self.assertEqual(result["stories"]["3-3-fourth-story"], "done")

    def test_missing_epic_returns_none(self):
        with TemporaryDirectory() as d:
            p = Path(d) / "sprint-status.yaml"
            p.write_text(SAMPLE_SPRINT_STATUS)
            result = epic_plan.parse_sprint_status(p, "99")

        self.assertIsNone(result["epic_status"])
        self.assertEqual(len(result["stories"]), 0)


class TestParseStoriesTable(unittest.TestCase):
    def test_parses_stories_and_deps(self):
        with TemporaryDirectory() as d:
            p = Path(d) / "3-epic-test.md"
            p.write_text(SAMPLE_EPIC_FILE)
            stories = epic_plan.parse_stories_table(p, "3")

        self.assertEqual(len(stories), 4)
        self.assertEqual(stories[0]["story_num"], "3.0")
        self.assertEqual(stories[0]["dependencies"], [])
        self.assertEqual(stories[1]["dependencies"], ["3.0"])
        self.assertEqual(stories[3]["dependencies"], ["3.1", "3.2"])

    def test_empty_file_returns_empty(self):
        with TemporaryDirectory() as d:
            p = Path(d) / "empty.md"
            p.write_text("# No stories here\n")
            stories = epic_plan.parse_stories_table(p, "3")

        self.assertEqual(stories, [])


class TestBuildDependencyLayers(unittest.TestCase):
    def test_correct_layer_grouping(self):
        stories = [
            {"story_num": "3.0", "dependencies": []},
            {"story_num": "3.1", "dependencies": ["3.0"]},
            {"story_num": "3.2", "dependencies": ["3.0"]},
            {"story_num": "3.3", "dependencies": ["3.1", "3.2"]},
        ]
        layers = epic_plan.build_dependency_layers(stories)

        self.assertEqual(len(layers), 3)
        self.assertEqual(layers[0], ["3.0"])
        self.assertIn("3.1", layers[1])
        self.assertIn("3.2", layers[1])
        self.assertEqual(layers[2], ["3.3"])

    def test_all_independent(self):
        stories = [
            {"story_num": "1.0", "dependencies": []},
            {"story_num": "1.1", "dependencies": []},
            {"story_num": "1.2", "dependencies": []},
        ]
        layers = epic_plan.build_dependency_layers(stories)

        self.assertEqual(len(layers), 1)
        self.assertEqual(sorted(layers[0]), ["1.0", "1.1", "1.2"])

    def test_linear_chain(self):
        stories = [
            {"story_num": "2.0", "dependencies": []},
            {"story_num": "2.1", "dependencies": ["2.0"]},
            {"story_num": "2.2", "dependencies": ["2.1"]},
        ]
        layers = epic_plan.build_dependency_layers(stories)

        self.assertEqual(len(layers), 3)
        self.assertEqual(layers[0], ["2.0"])
        self.assertEqual(layers[1], ["2.1"])
        self.assertEqual(layers[2], ["2.2"])


class TestEndToEnd(unittest.TestCase):
    def test_full_plan_generation(self):
        with TemporaryDirectory() as d:
            impl_dir = Path(d)
            (impl_dir / "sprint-status.yaml").write_text(SAMPLE_SPRINT_STATUS)
            (impl_dir / "3-epic-test.md").write_text(SAMPLE_EPIC_FILE)
            out = impl_dir / "plan.json"

            epic_plan.find_sprint_status.__wrapped__ if hasattr(epic_plan.find_sprint_status, "__wrapped__") else None

            ss_path = impl_dir / "sprint-status.yaml"
            ss_data = epic_plan.parse_sprint_status(ss_path, "3")
            epic_file = epic_plan.find_epic_file("3", impl_dir, None)

            self.assertIsNotNone(epic_file)

            stories = epic_plan.parse_stories_table(epic_file, "3")
            self.assertEqual(len(stories), 4)

            layers = epic_plan.build_dependency_layers(stories)
            self.assertEqual(len(layers), 3)

            backlog_count = sum(
                1 for s in stories
                if ss_data["stories"].get(
                    epic_plan.story_key_from_num("3", s["story_num"], ss_data["stories"]),
                    "unknown",
                ) == "backlog"
            )
            self.assertEqual(backlog_count, 3)


class TestStoryKeyFromNum(unittest.TestCase):
    def test_finds_matching_key(self):
        statuses = {"3-0-first-story": "backlog", "3-1-second": "done"}
        self.assertEqual(
            epic_plan.story_key_from_num("3", "3.0", statuses),
            "3-0-first-story",
        )

    def test_returns_none_for_missing(self):
        statuses = {"3-0-first-story": "backlog"}
        self.assertIsNone(epic_plan.story_key_from_num("3", "9.9", statuses))


if __name__ == "__main__":
    unittest.main()
