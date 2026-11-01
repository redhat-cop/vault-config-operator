---
title: 'Credential Source + CRD-Specific VAPs'
type: 'feature'
ticket: ''
created: '2026-11-01'
status: 'built'
baseline_revision: '6de75b5f30c918cb70cb9d4857cec052b623aec0'
route: 'full'
route_source: 'auto'
risk: 'medium'
review: 'quick'
review_source: 'pinned'
lenses_ran: ['quick']
review_loop_iteration: 1
context:
  - '_bmad-output/implementation-artifacts/18-2-credential-source-and-crd-specific-vaps.md'
  - '_bmad-output/vap-migration/plan-18-1-vap-infrastructure-and-universal-immutability-policies.md'
  - '_bmad-output/vap-migration/plan-18-1-1-target-namespace-guard-vap.md'
---

<frozen-after-approval reason="human-owned intent — do not modify unless human renegotiates">

## Intent

**Problem:** Credential source mutual-exclusivity validation and CRD-specific business rules are scattered across 20+ Go webhook and `_types.go` files. The Go code uses `reflect.DeepEqual`, `time.ParseDuration`, and other runtime constructs that cannot be audited declaratively.

**Approach:** Create ~30 ValidatingAdmissionPolicies + bindings under `config/vap/` using CEL expressions, covering: (A) shared and single-CRD credential source exactly-one rules, and (B) per-CRD validations (immutability, conditional field restrictions, one-way flags, mutual exclusivity). No Go code changes — webhook removal is deferred to Story 18.3.

## Boundaries & Constraints

**Always:**
- Use `admissionregistration.k8s.io/v1` (GA since K8s 1.30)
- Follow naming convention `vco-validate-<lowercase-kind>` for CRD-specific VAPs; `vco-exactly-one-<credential-group>` for shared credential VAPs
- Fold single-CRD credential validation into the CRD-specific VAP when one exists for that CRD (reduces total policy count)
- Match exact CEL `has()` semantics: pointer fields with `omitempty` → `has()` is false when nil; booleans with `omitempty` → `has()` is false when `false`; fields with `+kubebuilder:default` → `has()` is always true (API server populates default)
- `failurePolicy: Fail` on all VAPs
- All VAP and binding YAMLs added to `config/vap/kustomization.yaml`

**Never:**
- Do not modify Go webhook files, controller code, or test files
- Do not modify existing VAP manifests from Stories 18.1/18.1.1
- Do not attempt HCL parsing in CEL (RandomSecret inline password policy validation stays in Go webhook)
- Do not use CEL `duration()` for autoRotatePeriod — the Go `time.ParseDuration` format allows units CEL may not handle; validate via string comparison against "0" and empty only

## I/O & Edge-Case Matrix

| Scenario | Input / State | Expected Output / Behavior | Error Handling |
|----------|--------------|---------------------------|----------------|
| Exactly-one credential | CREATE with only `vaultSecret` set | Allowed | N/A |
| Zero credentials | CREATE with none of vaultSecret/secret/randomSecret | VAP rejection | "exactly one of ... must be specified" |
| Two credentials | CREATE with both `vaultSecret` and `secret` | VAP rejection | "exactly one of ... must be specified" |
| Transit immutable type change | UPDATE changing `spec.type` from `aes256-gcm96` to `ed25519` | VAP rejection | "spec.type cannot be updated" |
| Transit one-way flag OK | UPDATE changing `exportable` false→true | Allowed | N/A |
| Transit one-way flag blocked | UPDATE changing `exportable` true→false | VAP rejection | "one-way flag" |
| AuthEngineMount config-only update | UPDATE changing only `spec.authMount.config.defaultLeaseTTL` | Allowed | N/A |
| AuthEngineMount type change | UPDATE changing `spec.authMount.type` | VAP rejection | "only .spec.config can be modified" |
| AWS role externalID with assumed_role | CREATE with `credentialType: assumed_role`, `externalID` set | Allowed | N/A |
| AWS role externalID without assumed_role | CREATE with `credentialType: iam_user`, `externalID` set | VAP rejection | "only allowed when credentialType is assumed_role" |
| KubernetesAuthEngineRole both targetNamespace fields | CREATE with both `targetNamespaceSelector` and `targetNamespaces` | VAP rejection | "exactly one required" |

</frozen-after-approval>

## Code Map

- `config/vap/kustomization.yaml` -- ADD all new VAP+binding resources
- `config/vap/vco-exactly-one-root-credential.yaml` -- NEW shared VAP for 7 CRDs with `spec.rootCredentials`
- `config/vap/vco-exactly-one-bind-credential.yaml` -- NEW shared VAP for LDAPAuthEngineConfig + LDAPSecretEngineConfig (`spec.bindCredentials`)
- `config/vap/vco-validate-kerberosauthengineconfig-ldap.yaml` -- NEW credential VAP for KerberosAuthEngineLDAPConfig (`spec.ldapCredentials` — different field path)
- `config/vap/vco-validate-transitsecretenginekey.yaml` -- NEW complex VAP: convergent/derived, keySize/hmac, immutable fields, one-way flags
- `config/vap/vco-validate-awssecretenginerole.yaml` -- NEW: credentialType-dependent field restrictions
- `config/vap/vco-validate-randomsecret.yaml` -- NEW: passwordPolicy mutual exclusivity, secretKey immutability, KVv2 /data/ check
- `config/vap/vco-validate-databasesecretenginestaticrole.yaml` -- NEW: rotation mutual exclusivity, password/key credential config
- `config/vap/vco-validate-totpsecretenginekey.yaml` -- NEW: generate-mode conditional rules
- `config/vap/vco-validate-terraformcloudsecretenginerole.yaml` -- NEW: credentialType enum + type-dependent required fields
- `config/vap/vco-validate-authenginemount.yaml` -- NEW UPDATE-only: description, type, local, sealwrap immutable (compare each non-config field)
- `config/vap/vco-validate-secretenginemount.yaml` -- NEW UPDATE-only: type, description, local, sealWrap, externalEntropyAccess, options immutable
- `config/vap/vco-validate-pkisecretengineconfig.yaml` -- NEW: ACME requires cluster
- `config/vap/vco-validate-kubernetesauthengineconfig.yaml` -- NEW: tokenReviewer mutual exclusivity
- `config/vap/vco-validate-kubernetesauthenginerole.yaml` -- NEW: targetNamespaceSelector vs targetNamespaces
- `config/vap/vco-validate-sshsecretengineconfig.yaml` -- NEW: generateSigningKey conditional + caKeyReference constraints
- `config/vap/vco-validate-awssecretengineconfig.yaml` -- NEW combined: rootCredentials + accessKey/randomSecret
- `config/vap/vco-validate-consulsecretengineconfig.yaml` -- NEW combined: rootCredentials + clientCert/clientKey pairing
- `config/vap/vco-validate-nomadsecretengineconfig.yaml` -- NEW combined: rootCredentials + clientCert/clientKey pairing
- `config/vap/vco-validate-mongodbatlassecretengineconfig.yaml` -- NEW combined: rootCredentials + publicKey/randomSecret
- `config/vap/vco-validate-ldapsecretengineconfig.yaml` -- NEW combined: bindCredentials + bindDN/randomSecret
- `config/vap/vco-validate-terraformcloudsecretengineconfig.yaml` -- NEW combined: tfcCredentials + passwordKey constraint
- `config/vap/vco-validate-cfauthengineconfig.yaml` -- NEW combined: cfCredentials + cfUsername/randomSecret
- `config/vap/vco-validate-githubsecretengineconfig.yaml` -- NEW: sshKeyReference exactly one of vaultSecret/secret (no randomSecret)
- `config/vap/vco-validate-kubernetessecretengineconfig.yaml` -- NEW: jwtReference no randomSecret
- `config/vap/vco-validate-userpassauthengineuser.yaml` -- NEW: passwordCredentials source validation
- `config/vap/vco-validate-awsauthengineconfig.yaml` -- NEW credential-only: `spec.AWSCredentials` (note capital-A JSON key)
- `config/vap/vco-validate-azuresecretengineconfig.yaml` -- NEW credential-only: `spec.azureCredentials`
- `config/vap/vco-validate-gcpsecretengineconfig.yaml` -- NEW credential-only: `spec.gcpCredentials`
- `config/vap/vco-validate-oktaauthengineconfig.yaml` -- NEW credential-only: `spec.oktaCredentials`
- `config/vap/vco-validate-radiusauthengineconfig.yaml` -- NEW credential-only: `spec.radiusCredentials`
- `api/v1alpha1/authenginemount_types.go` (L62-83) -- AuthMount struct: immutable fields are description, type, local, sealwrap; Config is the only mutable sub-struct
- `api/v1alpha1/secretenginemount_types.go` (L100-129) -- Mount struct: immutable fields are type, description, local, sealWrap, externalEntropyAccess, options; Config is the only mutable sub-struct
- `api/v1alpha1/transitsecretenginekey_webhook.go` -- Go validation source of truth for transit key rules
- `api/v1alpha1/utils/commons.go` (L451-498) -- RootCredentialConfig struct and ValidateCredentialSource()

## Tasks & Acceptance

**Execution:**

- [ ] `config/vap/vco-exactly-one-root-credential.yaml` + binding -- CREATE shared VAP: `(has(object.spec.rootCredentials.vaultSecret) ? 1 : 0) + ... == 1` for resources: awssecretengineconfigs, consulsecretengineconfigs, databasesecretengineconfigs, mongodbatlassecretengineconfigs, nomadsecretengineconfigs, quaysecretengineconfigs, rabbitmqsecretengineconfigs; ops: CREATE, UPDATE
- [ ] `config/vap/vco-exactly-one-bind-credential.yaml` + binding -- CREATE shared VAP for ldapauthengineconfigs, ldapsecretengineconfigs (field: `spec.bindCredentials`); ops: CREATE, UPDATE
- [ ] `config/vap/vco-validate-kerberosauthengineconfig-ldap.yaml` + binding -- CREATE for kerberosauthengineldapconfigs (field: `spec.ldapCredentials`); ops: CREATE, UPDATE
- [ ] `config/vap/vco-validate-awsauthengineconfig.yaml` + binding -- CREATE for awsauthengineclientconfigs (field: `spec.AWSCredentials`, capital-A JSON key); ops: CREATE, UPDATE
- [ ] `config/vap/vco-validate-azuresecretengineconfig.yaml` + binding -- CREATE for azuresecretengineconfigs (field: `spec.azureCredentials`)
- [ ] `config/vap/vco-validate-gcpsecretengineconfig.yaml` + binding -- CREATE for gcpsecretengineconfigs (field: `spec.gcpCredentials`)
- [ ] `config/vap/vco-validate-oktaauthengineconfig.yaml` + binding -- CREATE for oktaauthengineconfigs (field: `spec.oktaCredentials`)
- [ ] `config/vap/vco-validate-radiusauthengineconfig.yaml` + binding -- CREATE for radiusauthengineconfigs (field: `spec.radiusCredentials`)
- [ ] `config/vap/vco-validate-transitsecretenginekey.yaml` + binding -- CREATE: CREATE+UPDATE ops; 10 validation rules (see Design Notes)
- [ ] `config/vap/vco-validate-awssecretenginerole.yaml` + binding -- CREATE: credentialType enum validation + field restriction matrix per credential type
- [ ] `config/vap/vco-validate-randomsecret.yaml` + binding -- CREATE: passwordPolicy/inlinePasswordPolicy mutual exclusivity, secretKey immutability on UPDATE, KVv2 `/data/` path check
- [ ] `config/vap/vco-validate-databasesecretenginestaticrole.yaml` + binding -- CREATE: rotationPeriod/rotationSchedule exactly-one, rotationWindow only with schedule, passwordCredentialConfig/rsaPrivateKeyCredentialConfig exactly-one
- [ ] `config/vap/vco-validate-totpsecretenginekey.yaml` + binding -- CREATE: issuer+accountName required, generate-mode conditional rules
- [ ] `config/vap/vco-validate-terraformcloudsecretenginerole.yaml` + binding -- CREATE: credentialType enum {team,team_legacy,user,organization} + type-dependent required fields
- [ ] `config/vap/vco-validate-authenginemount.yaml` + binding -- CREATE UPDATE-only: compare each non-config AuthMount field individually
- [ ] `config/vap/vco-validate-secretenginemount.yaml` + binding -- CREATE UPDATE-only: compare each non-config Mount field individually + map comparison for options
- [ ] `config/vap/vco-validate-pkisecretengineconfig.yaml` + binding -- CREATE: `has(object.spec.acme) && !has(object.spec.cluster)` → reject
- [ ] `config/vap/vco-validate-kubernetesauthengineconfig.yaml` + binding -- CREATE: tokenReviewerServiceAccount/tokenReviewerJWTSecret mutual exclusivity
- [ ] `config/vap/vco-validate-kubernetesauthenginerole.yaml` + binding -- CREATE: targetNamespaceSelector vs targetNamespaces exactly-one
- [ ] `config/vap/vco-validate-sshsecretengineconfig.yaml` + binding -- CREATE: generateSigningKey=false → caKeyReference required + no randomSecret
- [ ] `config/vap/vco-validate-awssecretengineconfig.yaml` + binding -- CREATE combined: credential source + accessKey required when randomSecret
- [ ] `config/vap/vco-validate-consulsecretengineconfig.yaml` + binding -- CREATE combined: credential source + clientCert/clientKey both-or-neither
- [ ] `config/vap/vco-validate-nomadsecretengineconfig.yaml` + binding -- CREATE combined: credential source + clientCert/clientKey both-or-neither
- [ ] `config/vap/vco-validate-mongodbatlassecretengineconfig.yaml` + binding -- CREATE combined: credential source + publicKey required when randomSecret
- [ ] `config/vap/vco-validate-ldapsecretengineconfig.yaml` + binding -- CREATE combined: bindCredentials + bindDN required when randomSecret
- [ ] `config/vap/vco-validate-terraformcloudsecretengineconfig.yaml` + binding -- CREATE combined: tfcCredentials + passwordKey required when secret/vaultSecret
- [ ] `config/vap/vco-validate-cfauthengineconfig.yaml` + binding -- CREATE combined: cfCredentials + cfUsername required when randomSecret
- [ ] `config/vap/vco-validate-githubsecretengineconfig.yaml` + binding -- CREATE: exactly one of sshKeyReference.vaultSecret/secret (no randomSecret)
- [ ] `config/vap/vco-validate-kubernetessecretengineconfig.yaml` + binding -- CREATE: jwtReference no randomSecret (only vaultSecret/secret)
- [ ] `config/vap/vco-validate-userpassauthengineuser.yaml` + binding -- CREATE: passwordCredentials exactly-one source
- [ ] `config/vap/kustomization.yaml` -- UPDATE: add all new VAP+binding files to resources list

**Acceptance Criteria:**
- Given `kustomize build config/default/`, when rendered, then output includes all new VAP manifests and bindings
- Given any of the 7 shared rootCredentials CRDs, when CREATE/UPDATE sets zero or >1 credential source, then API server rejects
- Given TransitSecretEngineKey, when UPDATE changes spec.type, then API server rejects
- Given AuthEngineMount, when UPDATE changes spec.authMount.type, then API server rejects (only config is mutable)
- Given AWSSecretEngineRole, when CREATE sets externalID with credentialType iam_user, then API server rejects
- Given all YAML manifests, when `kubectl apply --dry-run=server`, then all are valid

## Implementation Notes

## Plan Change Log

**Loopback 1** — Triggered by review finding: AuthEngineMount and SecretEngineMount VAPs used `spec.authMount.*` and `spec.mount.*` field paths, but both structs use `json:",inline"` so their fields appear directly under `spec.*` in the CRD schema. CEL would fail to compile → `failurePolicy: Fail` → all UPDATEs denied.
- Amended: Design Notes section — corrected CEL examples to use `object.spec.description` instead of `object.spec.authMount.description` (and all other inlined fields for both mount types). Added CRITICAL annotation about `json:",inline"`.
- Known-bad state avoided: VAP compile error rejecting all AuthEngineMount and SecretEngineMount UPDATE operations.
- KEEP: All other VAPs (28 out of 30) are correct and should not be regenerated. Credential source patterns, CRD-specific validations, naming conventions, kustomization structure all correct.

## Review Triage Log

**Pass 1 (quick lens):** 2 high, 0 medium, 1 low, 3 false, 0 maybe-false

| Finding | Verdict | Route | Evidence |
|---------|---------|-------|----------|
| AuthEngineMount VAP uses `spec.authMount.*` but AuthMount is `json:",inline"` — fields live at `spec.*` in CRD schema | high | bad_plan | CRD schema confirms no `authMount` property; CEL compile error → `failurePolicy: Fail` → all UPDATEs denied |
| SecretEngineMount VAP uses `spec.mount.*` but Mount is `json:",inline"` — fields live at `spec.*` in CRD schema | high | bad_plan | Same root cause: CRD schema confirms no `mount` property |
| TransitSecretEngineKey reads omitempty bools without `has()` | false | reject | K8s CEL engine returns zero values (false/0) for schema-declared optional fields absent from JSON; documented at kubernetes.io/docs CRD validation-rules |
| TOTPSecretEngineKey reads omitempty bools without `has()` | false | reject | Same as Transit: K8s CEL provides zero values for missing schema-declared optional primitives |
| RandomSecret reads omitempty bool/string without `has()`; KVv2 checks spec.path vs Go GetPath() | false | reject | Omitempty: same as Transit. KVv2 path: spec.path is what user controls; Go's GetPath() includes name suffix making it more permissive for `path=kv/data` edge case, but CEL check is arguably more correct |
| TFC credential VAP doesn't check inner field emptiness (vaultSecret.path, secret.name) | low | reject | Go webhook still runs and catches this; fix would add >3 validation rules across multiple credential VAPs — beyond direct correction; Story 18.3/18.4 will address |

## Design Notes

**CEL credential source exactly-one pattern:**
```cel
(has(object.spec.rootCredentials.vaultSecret) ? 1 : 0) +
(has(object.spec.rootCredentials.secret) ? 1 : 0) +
(has(object.spec.rootCredentials.randomSecret) ? 1 : 0) == 1
```
All three fields are pointers with `omitempty` — `has()` is false when nil.

**TransitSecretEngineKey CEL rules (CREATE+UPDATE):**
1. `!object.spec.convergentEncryption || object.spec.derived` — convergent requires derived
2. `object.spec.keySize == 0 || object.spec.type == "hmac"` — keySize only for hmac
3. `object.spec.keySize == 0 || object.spec.type != "hmac" || object.spec.keySize >= 32` — hmac keySize ≥ 32
4. `!has(object.spec.autoRotatePeriod) || object.spec.autoRotatePeriod == "" || object.spec.autoRotatePeriod == "0" || duration(object.spec.autoRotatePeriod) >= duration("1h")` — autoRotatePeriod ≥ 1h

**TransitSecretEngineKey UPDATE-only rules** (guard with `request.operation == "CREATE" ||`):
5. `object.spec.type == oldObject.spec.type`
6. `object.spec.derived == oldObject.spec.derived`
7. `object.spec.convergentEncryption == oldObject.spec.convergentEncryption`
8. `object.spec.keySize == oldObject.spec.keySize`
9. `!oldObject.spec.exportable || object.spec.exportable` — one-way
10. `!oldObject.spec.allowPlaintextBackup || object.spec.allowPlaintextBackup` — one-way

**Note on `spec.type` default:** TransitSecretEngineKey has `+kubebuilder:default="aes256-gcm96"` on `type`, so `has(object.spec.type)` is always true. Bool fields `derived`, `convergentEncryption` have `omitempty`, so `has()` is false when `false`. But in CEL, comparing `object.spec.derived == oldObject.spec.derived` works regardless because both sides have the same presence — the API server serializes consistently.

**AuthEngineMount UPDATE-only CEL (field-by-field comparison):**
CRITICAL: `AuthMount` uses `json:",inline"` — its fields appear directly under `spec.`, NOT under `spec.authMount.`.
```cel
object.spec.description == oldObject.spec.description &&
object.spec.type == oldObject.spec.type &&
object.spec.local == oldObject.spec.local &&
object.spec.sealwrap == oldObject.spec.sealwrap
```
Note: `description` and `type` are strings with `omitempty`; `local` and `sealwrap` are booleans with `omitempty`. K8s CEL returns zero values for missing schema-declared optional primitives, so direct `==` comparison works.

**SecretEngineMount UPDATE-only CEL:**
CRITICAL: `Mount` uses `json:",inline"` — its fields appear directly under `spec.`, NOT under `spec.mount.`.
Same pattern: compare `type`, `description`, `local`, `sealWrap`, `externalEntropyAccess` directly on `object.spec.*`. For `options` (map): `(!has(object.spec.options) && !has(oldObject.spec.options)) || (has(object.spec.options) && has(oldObject.spec.options) && object.spec.options == oldObject.spec.options)`.

**CRD plural names:** Verify against `config/crd/bases/` filenames (format `redhatcop.redhat.io_<plural>.yaml`). Key exception: `KerberosAuthEngineLDAPConfig` → `kerberosauthengineldapconfigs`, and `AWSAuthEngineConfig` → `awsauthengineclientconfigs` (note: **client** in plural name — the CRD is named `AWSAuthEngineClientConfig` not `AWSAuthEngineConfig`).

**Binding structure (standard):**
```yaml
apiVersion: admissionregistration.k8s.io/v1
kind: ValidatingAdmissionPolicyBinding
metadata:
  name: vco-validate-<kind>
spec:
  policyName: vco-validate-<kind>
  validationActions: ["Deny"]
```

## Verification

**Commands:**
- `GOTOOLCHAIN=auto make test` -- expected: all pass (no Go changes, but validates manifests generation)
- `kustomize build config/default/ | grep -c ValidatingAdmissionPolicy` -- expected: count reflects 6 existing + all new VAPs
- `ls config/vap/*.yaml | wc -l` -- expected: 7 existing + ~60 new (30 VAPs + 30 bindings) = ~67
- `kubectl apply --dry-run=server -f config/vap/` -- expected: all manifests valid YAML (requires K8s 1.30+ cluster)
