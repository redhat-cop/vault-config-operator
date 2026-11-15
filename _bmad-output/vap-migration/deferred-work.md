# Deferred Work

- source_plan: `_bmad-output/vap-migration/plan-18-1-vap-infrastructure-and-universal-immutability-policies.md`
  summary: Create parameterized vco-target-namespace-guard VAP with CleansePath/JoinNamespace CEL logic, vco-vap-params ConfigMap, and remove ValidateTargetNamespaceUpdate from all 91 webhook files and commons.go
  evidence: Split from Story 18.1 plan to reduce token count below 1600; the target-namespace-guard VAP is the most complex component (parameterized VAP, CEL string manipulation for CleansePath/JoinNamespace) and can ship independently after path/name VAPs prove the infrastructure
  resolved: Story 18-1-1 (`_bmad-output/implementation-artifacts/18-1-1-target-namespace-guard-vap.md`)
