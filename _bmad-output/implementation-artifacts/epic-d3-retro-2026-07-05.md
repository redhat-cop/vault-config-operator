# Epic D3 Retrospective — Secret Engine Documentation — Per-Engine Split & Standardization

**Date:** 2026-07-05
**Facilitator:** Bob (Scrum Master)
**Participants:** Raffa (Project Lead), Alice (Product Owner), Charlie (Senior Dev), Dana (QA Engineer), Amelia (Developer Agent), Paige (Technical Writer)

---

## Epic Summary

| Metric | Value |
|--------|-------|
| Epic | D3: Secret Engine Documentation — Per-Engine Split & Standardization |
| Stories | 4 of 4 completed (100%) |
| Duration | ~3 days (July 2–5, 2026) |
| Story categories | 1 infrastructure/index (D3.1), 1 database engine (D3.2), 1 dual-engine (D3.3: PKI + RabbitMQ), 1 quad-engine (D3.4: GitHub + Quay + K8s + Azure) |
| Code review findings | 15 total (13 patched, 2 deferred pre-existing) |
| Source content bugs fixed | 4+ (Quay credential pattern, Azure OIDC copy-paste, K8s CLI duplicate field, various typos) |
| Technical debt created | 0 |
| Production incidents | 0 |

### AI Models Used

| Story | Model |
|-------|-------|
| D3.1 — Create Secret-Engines Directory Structure and Index Page | Opus 4.6 |
| D3.2 — Standardize Database Secret Engine Docs | Opus 4.6 |
| D3.3 — Standardize PKI and RabbitMQ Secret Engine Docs | Opus 4.6 |
| D3.4 — Standardize GitHub, Quay, Kubernetes, and Azure Secret Engine Docs | Opus 4.6 |

---

## Epic D2 Retrospective Follow-Through

| Action Item | Status |
|-------------|--------|
| Continue using Opus 4.6 for all stories | ✅ All 4 D3 stories used Opus 4.6 |
| Continue detailed dev notes and code review process | ✅ All 4 stories had comprehensive dev notes; 15 review findings across D3 |

**Completed 2/2.** Third consecutive perfect follow-through (D1→D2→D3). Both items graduated to embedded team practices — no longer tracked as action items.

---

## Successes

1. **Template-driven execution at scale.** The documentation template created in D1.1 and validated across 6 auth engines in D2 continued to perform in D3. All 7 secret engine docs follow the same structure: Overview → Config CRD → Role CRD(s) → Credential Resolution → See Also. The template is now validated across 13 engine docs total.

2. **Story intelligence chain — strongest yet.** D3.1 documented all cross-references for downstream stories. D3.2 established the three-CRD pattern (Config + Role + StaticRole). D3.3 handled the most field-heavy types (PKI: 67 fields). D3.4 handled four engines with four distinct credential patterns in a single story. Each story built effectively on predecessor context.

3. **Documentation-as-audit surfaced a pre-existing code bug.** The D3.3 code review discovered that `convertVhostsToJson` and `convertTopicsToJson` in `rabbitmqsecretenginerole_types.go` overwrite the map on each iteration — only the last vhost/topic entry survives when multiple are specified. This bug predates D3 and was found because thorough documentation forces deep code understanding.

4. **Code reviews caught 15 findings — 13 patched, 2 deferred.** Both deferred items are pre-existing issues (RabbitMQ serialization bug, README phrasing). Reviews caught real user-facing issues: missing field documentation, incomplete CLI examples, credential resolution edge cases, and type discrepancies.

5. **Four+ source content bugs fixed.** The original `secret-engines.md` monolith contained: (1) Quay credential docs using Pattern A names but CRD uses Pattern B, (2) Azure secret engine docs with OIDC copy-paste errors from auth engine docs, (3) K8s Role CLI example with duplicate `kubernetes_role_name`, (4) various typos. All caught during extraction and standardization.

6. **100% completion in ~3 days with zero blockers.** All 4 stories completed, zero new technical debt, zero production incidents. Third consecutive documentation epic with perfect delivery.

7. **D3.4 exceeded requirements.** Added 4 missing readme.md entries for KubernetesSecretEngineConfig, KubernetesSecretEngineRole, AzureSecretEngineConfig, and AzureSecretEngineRole — these were never in the README despite being in the operator. Also fixed 5 pre-existing "see the also the" typos in readme.md during the retrospective.

8. **Third consecutive perfect retro follow-through.** Both D2 action items followed. These are now embedded practices, not tracked action items.

---

## Challenges

1. **Review findings increased with story complexity.** D3.1 had 1 finding, D3.2 had 4, D3.3 had 4, D3.4 had 6. Unlike D2's decreasing trajectory (4→4→4→0), D3's trajectory increased. This is driven by complexity (PKI has 67 fields; D3.4 covered 4 engines with 4 credential patterns), not quality regression. The types of findings were accuracy issues in complex areas, not careless mistakes.

2. **Kind cluster infrastructure flakiness.** D3.4 debug logs mention USB device stale mapping issues causing Kind cluster local-path provisioner failures (3 occurrences). Fixed by restarting Kind node containers. Not D3-specific — recurring infrastructure issue. Not blocking documentation epics; will need addressing when code-touching epics resume.

---

## Key Insights

1. **Documentation-as-audit is a real quality tool.** Writing thorough documentation forces deep code understanding. D3.3's review discovered a pre-existing RabbitMQ serialization bug because the docs described multi-entry usage that the code doesn't actually support. This side benefit should be expected and valued in future documentation work.

2. **Template maturity validated at scale.** The D1 template has now been applied across 13 engine docs (6 auth + 7 secret) over 2 execution epics. Zero template-related debates or modifications were needed in D3. The template is stable and proven.

3. **Review effort scales with complexity, not story order.** D2 showed decreasing findings (learning curve effect). D3 showed increasing findings (complexity effect). Both are healthy patterns — the difference is that D3's later stories were genuinely more complex (more CRDs, more credential patterns, more fields).

4. **Graduate recurring action items to embedded practices.** After 3 consecutive perfect follow-throughs, "use Opus 4.6" and "detailed dev notes + code review" are no longer action items — they're how the team works. Stop tracking them to reduce ceremony.

---

## Action Items

### Process Improvements

1. **Graduate "use Opus 4.6" and "detailed dev notes + code review" to embedded team practices**
   - Owner: Bob (Scrum Master)
   - Effective: Immediately
   - Success criteria: No longer tracked as action items; violations would surface naturally in retros

2. **Budget review attention by story complexity**
   - Owner: Bob (Scrum Master)
   - Deadline: Starting with D4 story creation
   - Success criteria: Multi-engine or high-field-count stories get explicit complexity notes in story files

### Technical Debt (Pre-Existing, Tracked for Future Resolution)

1. **RabbitMQ vhost/vhostTopic multi-entry serialization bug**
   - File: `api/v1alpha1/rabbitmqsecretenginerole_types.go` lines 199-231
   - Bug: `convertVhostsToJson` and `convertTopicsToJson` overwrite map on each loop iteration instead of accumulating entries
   - Impact: Only last vhost/topic entry written to Vault when multiple are specified
   - Severity: Medium — single-entry usage (likely common case) works correctly
   - Fix: Change `vhostData = map[string]any{...}` to `vhostData[key] = value` (3 occurrences in both functions)
   - Test needed: Integration test with multiple vhost entries to verify fix
   - Owner: Track for future code-touching epic (Epic 8+ or dedicated hardening pass)

### Resolved During Retro

1. **README typo "see the also the" → "see also the"** ✅
   - 5 occurrences fixed in `readme.md` during retrospective session

### Dismissed

- ~~Preparation sprint for D4~~ — No prep needed; preconditions (D2 + D3 complete) satisfied
- ~~Template adaptation for examples~~ — D4 examples are simpler than field-reference docs; adapt within stories

### Team Agreements

- Template + reference implementation + code review = proven documentation quality system (keep all three)
- Story-to-story intelligence chain is critical — always reference predecessor context
- Documentation-as-audit is a real benefit — thorough docs surface pre-existing code bugs
- Graduate recurring action items to embedded practices after 3 consecutive follow-throughs
- Review effort scales with complexity — budget accordingly, don't mistake it for quality regression

---

## Epic D4 Preparation

### Readiness Assessment

- **Template:** `docs/engine-doc-template.md` is complete and proven across 13 engine docs
- **Pattern:** D2 and D3 validated the extract→standardize→verify workflow
- **Source quality:** Per-engine docs in `docs/auth-engines/` and `docs/secret-engines/` are standardized — examples can reference them directly
- **Infrastructure:** No Kind cluster, Vault, or build tooling needed — pure documentation epic
- **Preconditions:** D2 (auth engine docs) and D3 (secret engine docs) both complete
- **Blockers:** None

### Dependencies on Epic D3

- D4.2 secret engine example YAMLs should be consistent with per-engine docs created in D3
- Template-driven documentation pattern informs example quality and structure

### Potential Friction Points

- Examples are simpler than field-reference docs — lower review burden expected
- Existing `docs/examples/postgresql/` provides a reference implementation for example directory structure
- D4.1 and D4.2 cover many engines (6 auth + 7 secret) — stories may be large but straightforward

### Verdict

**Ready to proceed with Epic D4.** No preparation work needed. All preconditions satisfied.

---

## Readiness Assessment

| Area | Status |
|------|--------|
| Testing & Quality | ✅ Complete — code reviews served as quality gate; 13 of 15 findings patched, 2 deferred (pre-existing) |
| Deployment | ✅ Docs ready to push |
| Stakeholder Acceptance | ✅ No user-reported issues |
| Technical Health | ✅ Stable — docs-only epic, zero runtime risk |
| Unresolved Blockers | ✅ None |

---

## Team Performance

Epic D3 delivered 4 stories in ~3 days covering secret engine directory structure and index creation, plus standardization of Database, PKI, RabbitMQ, GitHub, Quay, Kubernetes, and Azure secret engine documentation — with 100% completion, zero production incidents, 15 code review findings (13 patched, 2 deferred pre-existing), and 4+ source content bugs corrected. The documentation-as-audit process additionally surfaced a pre-existing RabbitMQ serialization bug in the operator code. Combined with D2's auth engine work, the operator now has 13 standardized, template-compliant per-engine documentation files. All D2 retro action items were fully followed (third consecutive perfect follow-through), and both have been graduated to embedded team practices. The team is ready for Epic D4 with no preparation needed.
