# Epic D4 Retrospective — Examples Directory Expansion

**Date:** 2026-07-08
**Facilitator:** Bob (Scrum Master)
**Participants:** Raffa (Project Lead), Alice (Product Owner), Charlie (Senior Dev), Dana (QA Engineer), Amelia (Developer Agent), Paige (Technical Writer)

---

## Epic Summary

| Metric | Value |
|--------|-------|
| Epic | D4: Examples Directory Expansion |
| Stories | 3 of 3 completed (100%) |
| Duration | ~1 day (July 7, 2026) |
| Story categories | 1 auth engine examples (D4.1: 6 engines), 1 secret engine examples (D4.2: 7 engines), 1 end-to-end examples (D4.3: 2 scenarios) |
| Files created | 18 new files (7 auth YAML + 7 secret YAML + 2 e2e YAML + 2 e2e README) |
| Code review findings | 9 total (8 patched, 1 dismissed) |
| Technical debt created | 0 |
| Production incidents | 0 |
| Code changes | 0 (pure documentation/examples) |

### AI Models Used

| Story | Model |
|-------|-------|
| D4.1 — Create Example YAML Files for Each Auth Engine | Opus 4.6 |
| D4.2 — Create Example YAML Files for Each Secret Engine | Opus 4.6 |
| D4.3 — Create Additional End-to-End Examples | Opus 4.6 |

---

## Epic D3 Retrospective Follow-Through

| Action Item | Status |
|-------------|--------|
| Graduate "use Opus 4.6" and "detailed dev notes + code review" to embedded practices | ✅ All 3 D4 stories used Opus 4.6; all had dev notes and review |
| Budget review attention by story complexity | ✅ D4 stories included per-engine CRD detail sections scaled to complexity |

**Completed 2/2.** Fourth consecutive perfect follow-through (D1→D2→D3→D4). Both items were already graduated to embedded practices after D3. Follow-through tracking itself is now graduated — stop tracking as ceremony.

---

## Successes

1. **Fastest epic yet — 3 stories in a single day.** All 3 stories completed on July 7, 2026. Expanded the `docs/examples/` directory from 1 engine (PostgreSQL) to 15 directories covering every supported auth and secret engine plus 2 end-to-end walkthroughs.

2. **Story intelligence chain at peak.** D4.1 established conventions, D4.2's dev notes explicitly referenced D4.1 patterns under "Previous Story Intelligence," and D4.3 referenced both predecessors. This chain drove the single-day velocity.

3. **E2e examples introduced a new documentation tier.** D4.3 added README.md companion files for end-to-end examples — narrative walkthroughs explaining how auth, policy, and secret engine connect. A higher tier than inline YAML comments alone.

4. **Code review continued to catch real issues.** 8 of 9 findings were path accuracy issues (mount path composition). All caught and fixed before merge. Reviews remain the primary quality gate for documentation work.

5. **Zero code changes, zero debt, zero incidents.** Fourth consecutive clean documentation epic (D1–D4). Phase 1.5 Documentation Improvement completed with perfect delivery across all four epics.

6. **Fourth consecutive perfect retro follow-through.** Both D3 action items followed. Follow-through tracking itself is now graduated to embedded practice.

7. **Phase 1.5 Documentation Improvement completed.** D4 closes out the documentation phase: D1 (standards + CertAuth), D2 (auth engine docs), D3 (secret engine docs), D4 (examples). The operator now has comprehensive, template-standardized documentation across all supported engines.

---

## Challenges

1. **Mount path composition was the dominant review theme.** 8 of 9 review findings across D4 were about the `{spec.path}/{metadata.name}` Vault path composition convention. D4.2 had all 7 examples initially wrong; D4.3 had both e2e examples wrong. The convention is non-obvious — it lives in operator runtime code, not in CRD field docs or inline comments. Dev agents reading CRD schemas naturally assume `spec.path` is the full mount path.

2. **Kind cluster intermittent failures continued.** D4.1 and D4.2 skipped integration tests due to Kind cluster container not running (USB device stale mapping issues). D4.3 ran tests successfully. Not blocking for documentation epics, but becomes critical for Epic 8 (dependency upgrades).

---

## Key Insights

1. **Mount path composition needs explicit documentation.** The `{spec.path}/{metadata.name}` convention caused 89% of review findings. This is a non-obvious operator convention that must be documented in project-context.md to prevent repeat issues in future work.

2. **Writing examples is a different skill than writing field-reference docs.** Examples must demonstrate cross-resource relationships and runtime behavior (path composition, policy-to-engine connections), not just field definitions. D4.3's e2e examples surfaced this distinction clearly.

3. **Documentation epic velocity reached optimal level.** Entire epics completing in a single day with zero debt. Diminishing returns on further velocity optimization — the bottleneck is now review quality, not dev speed.

4. **Dependency upgrade epic targets go stale.** Epic 8 was planned in April 2026 with Go 1.24 as target. By July 2026, Go 1.24 was EOL and the K8s ecosystem had moved to the v0.36 generation. Just-in-time version audits during pre-epic retros are essential.

5. **Just-in-time version audits are a new team practice.** Established during this retro: check dependency targets when an epic is about to start, not months ahead. Applied immediately to Epic 8.

---

## Significant Discovery: Epic 8 Version Targets Stale

During this retrospective, the team discovered that Epic 8's version targets (planned April 2026) were outdated:

| Component | Epic 8 Original Target | Updated Target (July 2026) | Reason |
|-----------|----------------------|---------------------------|--------|
| Go | 1.24 | **1.26** | Go 1.24 is EOL; only 1.25/1.26 maintained |
| controller-runtime | v0.23.x | **v0.24.1** | v0.24 is current; requires Go 1.26 |
| K8s libs (client-go etc.) | v0.35.x | **v0.36.2** | Coupled to controller-runtime v0.24 |
| controller-gen | v0.20.1 | **v0.21.0** | Coupled to K8s v0.36 generation |
| ENVTEST_VERSION | release-0.23 | **release-0.24** | Coupled to controller-runtime v0.24 |
| ENVTEST_K8S_VERSION | 1.35.0 | **1.36.x** | Current K8s version |
| kubectl | v1.35.x | **v1.36.2** | Current K8s version |
| Kind | v0.31.0 | **v0.32.0** | Breaking changes: Envoy replaces HAProxy, kubeadm v1beta4 |

**Action taken during retro:** Updated `epics.md` (all 4 stories + summary + DU requirements) and `sprint-status.yaml` with corrected version targets. Epic 8 is now aligned with the current v0.36 / Go 1.26 ecosystem generation.

---

## Action Items

### Process Improvements

1. **Document mount path composition convention in project-context.md**
   - Owner: Bob (Scrum Master) / next story author
   - Deadline: Before Epic 8 starts
   - Success criteria: `project-context.md` contains a clear explanation that engine mounts compose their Vault path as `{spec.path}/{metadata.name}`

2. **Adopt just-in-time version target refresh for dependency upgrade epics**
   - Owner: Bob (Scrum Master)
   - Effective: Immediately
   - Success criteria: Each dependency upgrade epic gets a version audit during its pre-epic retrospective

### Technical Debt (Pre-Existing, Tracked for Future Resolution)

1. **RabbitMQ vhost/vhostTopic multi-entry serialization bug** (carried from D3 retro)
   - File: `api/v1alpha1/rabbitmqsecretenginerole_types.go` lines 199-231
   - Owner: Track for future code-touching epic (Epic 8+)

### Resolved During Retro

1. **Epic 8 version targets updated** ✅
   - All 4 stories, epic summary, DU requirements, and sprint-status updated to v0.36 / Go 1.26 generation

### Team Agreements

- Story intelligence chain is a proven velocity multiplier — always reference predecessor context
- Code review remains the primary quality gate for documentation epics
- Graduate retro follow-through tracking itself — after 4 consecutive perfect follow-throughs, stop tracking as ceremony
- Just-in-time version audits for dependency epics — check targets at epic start, not months ahead
- Document non-obvious operator conventions in project-context.md proactively

---

## Readiness Assessment

| Area | Status |
|------|--------|
| Testing & Quality | ✅ Complete — code reviews served as quality gate; 8 of 9 findings patched, 1 dismissed |
| Deployment | ⏳ Pending — release planned after retrospective |
| Stakeholder Acceptance | ✅ No issues |
| Technical Health | ✅ Stable — docs-only epic, zero runtime risk |
| Unresolved Blockers | ✅ None (Kind cluster flakiness is pre-existing, not D4-specific) |

---

## Team Performance

Epic D4 delivered 3 stories in a single day, creating 18 new files that expanded the `docs/examples/` directory from 1 engine to 15 directories — covering all 6 auth engines, all 7 secret engines, and 2 end-to-end walkthroughs (JWT+PKI, Azure+Azure). This completes Phase 1.5 Documentation Improvement: four documentation epics (D1–D4) delivered with 100% completion, zero technical debt, and zero production incidents across all four. The retrospective surfaced 4 key insights and 1 significant discovery (Epic 8 version targets stale), which was resolved during the session by updating all targets to the current v0.36 / Go 1.26 ecosystem generation. The team established a new practice — just-in-time version audits for dependency upgrade epics. The team is ready for Phase 2 (Epic 8: Go + Kubernetes Stack Upgrade) with updated version targets and no preparation work needed beyond documenting the mount path composition convention.
