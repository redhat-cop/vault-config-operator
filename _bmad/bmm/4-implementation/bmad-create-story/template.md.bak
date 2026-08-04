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

diff --git a/_bmad/bmm/4-implementation/bmad-create-story/template.md b/_bmad/bmm/4-implementation/bmad-create-story/template.md
index c4e129f..3012727 100644
--- a/_bmad/bmm/4-implementation/bmad-create-story/template.md
+++ b/_bmad/bmm/4-implementation/bmad-create-story/template.md
@@ -47,3 +47,17 @@ so that {{benefit}}.
 ### Completion Notes List
 
 ### File List
+
+## Code Review Record
+
+### Review Model Used
+
+{{review_model_name_version}}
+
+### Review Findings
+
+### Decisions Needed
+
+### Decisions Taken
+
+### Fixes Applied
