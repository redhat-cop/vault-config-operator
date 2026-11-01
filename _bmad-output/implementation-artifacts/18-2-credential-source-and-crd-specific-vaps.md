# Story 18.2: Credential Source + CRD-Specific VAPs

Status: done
## Story

As an operator developer,
I want the credential source validation and per-CRD unique validation rules migrated to VAPs,
So that all validating admission logic is expressed declaratively in CEL, not in Go webhook code.

## Acceptance Criteria

### Part A — Credential Source VAPs

1. **Given** any of the 7 CRDs with `spec.rootCredentials` (AWSSecretEngineConfig, ConsulSecretEngineConfig, DatabaseSecretEngineConfig, MongoDBAtlasSecretEngineConfig, NomadSecretEngineConfig, QuaySecretEngineConfig, RabbitMQSecretEngineConfig)
   **When** a user submits a CREATE or UPDATE with zero or more than one of `spec.rootCredentials.{vaultSecret, secret, randomSecret}` set
   **Then** the API server rejects the request

2. **Given** other credential field groups (bindCredentials, GCPCredentials, AWSCredentials, AzureCredentials, OktaCredentials, RADIUSCredentials, CFCredentials, LDAPCredentials, PasswordCredentials, TFCCredentials, JWTReference, CAKeyReference, SSHKeyReference)
   **When** the same mutual-exclusivity rule is violated
   **Then** the corresponding per-field VAP rejects the request

### Part B — CRD-Specific VAPs

3. **TransitSecretEngineKey:** convergentEncryption requires derived, keySize only for hmac (≥32), autoRotatePeriod minimum 1h, and 5 immutable fields (type, derived, convergentEncryption, keySize) + 2 one-way flags (exportable, allowPlaintextBackup)
4. **AWSSecretEngineRole:** credentialType-dependent field restrictions (externalID/sessionTags only for assumed_role, iamTags/userPath/permissionsBoundaryARN only for iam_user, STS TTLs only for assumed_role/federation_token, roleArns required for assumed_role, policyArns/policyDocument/roleArns forbidden for session_token, etc.)
5. **RandomSecret:** passwordPolicy mutual exclusivity, secretKey immutability, KVv2 path `/data/` check
6. **DatabaseSecretEngineStaticRole:** rotationPeriod/rotationSchedule mutual exclusivity (exactly one required), rotationWindow only with rotationSchedule, passwordCredentialConfig/rsaPrivateKeyCredentialConfig exactly-one
7. **TOTPSecretEngineKey:** issuer and accountName required, generate-mode conditional rules (key/url forbidden when generate=true; key-or-url required when generate=false; exported/keySize forbidden when generate=false)
8. **TerraformCloudSecretEngineRole:** credentialType required and enum-validated {team, team_legacy, user, organization}, type-dependent required fields (userID for user, organization+teamID for team/team_legacy, organization for organization), userID conflicts with organization/teamID
9. **AuthEngineMount:** only `.spec.config` can be modified on UPDATE (description, type, local, sealWrap are immutable after creation)
10. **SecretEngineMount:** only `.spec.config` can be modified on UPDATE (type, description, local, sealWrap, externalEntropyAccess, options are immutable after creation)
11. **PKISecretEngineConfig:** ACME requires cluster (`spec.acme` set but `spec.cluster` nil → reject)
12. **KubernetesAuthEngineConfig:** tokenReviewerServiceAccount and tokenReviewerJWTSecret are mutually exclusive
13. **KubernetesAuthEngineRole:** targetNamespaceSelector vs targetNamespaces mutual exclusivity (exactly one required)
14. **SSHSecretEngineConfig:** when generateSigningKey=false: caKeyReference required, caKeyReference.randomSecret forbidden, caKeyReference must pass credential source validation
15. **AWSSecretEngineConfig:** rootCredentials + credential source validation, plus accessKey required when using randomSecret
16. **ConsulSecretEngineConfig:** rootCredentials + credential source validation, plus clientCert and clientKey must both be set or both empty
17. **NomadSecretEngineConfig:** rootCredentials + credential source validation, plus clientCert and clientKey must both be set or both empty
18. **MongoDBAtlasSecretEngineConfig:** rootCredentials + credential source validation, plus publicKey required when using randomSecret
19. **LDAPSecretEngineConfig:** bindCredentials + credential source validation, plus bindDN required when using randomSecret
20. **TerraformCloudSecretEngineConfig:** tfcCredentials + credential source validation, plus passwordKey required when secret or vaultSecret is used
21. **CFAuthEngineConfig:** cfCredentials + credential source validation, plus cfUsername required when using randomSecret
22. **GitHubSecretEngineConfig:** exactly one of sshKeyReference.vaultSecret or sshKeyReference.secret (no randomSecret allowed)
23. **KubernetesSecretEngineConfig:** jwtReference.randomSecret is forbidden (only vaultSecret or secret allowed)
24. **UserpassAuthEngineUser:** passwordCredentials must pass credential source validation (exactly one source)

## Tasks / Subtasks

- [ ] Task 1: Create credential source shared VAPs (AC: #1, #2)
  - [ ] 1.1: Create `vco-exactly-one-root-credential` VAP for 7 CRDs sharing `spec.rootCredentials`
  - [ ] 1.2: Create `vco-exactly-one-bind-credential` VAP for LDAPAuthEngineConfig + LDAPSecretEngineConfig + KerberosAuthEngineLDAPConfig (`spec.bindCredentials` / `spec.ldapCredentials`)
  - [ ] 1.3: Create individual credential source VAPs for single-CRD credential groups: AWSCredentials, AzureCredentials, GCPCredentials, OktaCredentials, RADIUSCredentials, CFCredentials, PasswordCredentials, TFCCredentials, JWTReference, SSHKeyReference(CAKeyReference)
- [ ] Task 2: Create CRD-specific VAPs (AC: #3–#24)
  - [ ] 2.1: `vco-validate-transitsecretenginekey` — convergent/derived, keySize/hmac, autoRotatePeriod, immutable fields, one-way flags
  - [ ] 2.2: `vco-validate-awssecretenginerole` — credentialType-dependent field restrictions
  - [ ] 2.3: `vco-validate-randomsecret` — passwordPolicy mutual exclusivity, secretKey immutability, KVv2 data path
  - [ ] 2.4: `vco-validate-databasesecretenginestaticrole` — rotation mutual exclusivity, password/key either-or
  - [ ] 2.5: `vco-validate-totpsecretenginekey` — issuer/accountName required, generate-mode rules
  - [ ] 2.6: `vco-validate-terraformcloudsecretenginerole` — credentialType enum + type-dependent fields
  - [ ] 2.7: `vco-validate-authenginemount` — only config modifiable on update
  - [ ] 2.8: `vco-validate-secretenginemount` — only config modifiable on update
  - [ ] 2.9: `vco-validate-pkisecretengineconfig` — ACME/cluster constraint
  - [ ] 2.10: `vco-validate-kubernetesauthengineconfig` — tokenReviewer mutual exclusivity
  - [ ] 2.11: `vco-validate-kubernetesauthenginerole` — targetNamespace selector/list mutual exclusivity
  - [ ] 2.12: `vco-validate-sshsecretengineconfig` — generateSigningKey conditional + caKeyReference constraints
  - [ ] 2.13: Combined VAPs for types with credential-source + extra constraint (AWSSecretEngineConfig, ConsulSecretEngineConfig, NomadSecretEngineConfig, MongoDBAtlasSecretEngineConfig, LDAPSecretEngineConfig, TerraformCloudSecretEngineConfig, CFAuthEngineConfig)
  - [ ] 2.14: `vco-validate-githubsecretengineconfig` — exactly one of vaultSecret/secret (no randomSecret)
  - [ ] 2.15: `vco-validate-kubernetessecretengineconfig` — no randomSecret in jwtReference
  - [ ] 2.16: `vco-validate-userpassauthengineuser` — passwordCredentials source validation
- [ ] Task 3: Create ValidatingAdmissionPolicyBinding manifests for all new VAPs (AC: #1–#24)
  - [ ] 3.1: Bindings for shared credential source VAPs (one binding per VAP listing applicable CRDs)
  - [ ] 3.2: Bindings for CRD-specific VAPs (one binding per VAP)
- [ ] Task 4: Kustomize integration (AC: all)
  - [ ] 4.1: Add all new VAP and binding YAMLs to `config/vap/kustomization.yaml`

## Dev Notes

### Dependency on Story 18.1

This story **depends on Story 18.1** which sets up the `config/vap/` directory, kustomize overlay, envtest VAP configuration, and the three universal shared VAPs (immutable-path, immutable-name, target-namespace-guard). The dev agent must verify that `config/vap/` already exists with the Tier 1 VAPs before starting this work.

If `config/vap/` does not exist yet, **stop and report** that Story 18.1 must be completed first.

### Credential Source VAP Grouping Strategy

Group credential source VAPs by field path pattern to minimize policy count:

| VAP Name | Field Path | Applicable CRDs | Operation |
|----------|-----------|------------------|-----------|
| `vco-exactly-one-root-credential` | `spec.rootCredentials` | AWSSecretEngineConfig, ConsulSecretEngineConfig, DatabaseSecretEngineConfig, MongoDBAtlasSecretEngineConfig, NomadSecretEngineConfig, QuaySecretEngineConfig, RabbitMQSecretEngineConfig | CREATE, UPDATE |
| `vco-exactly-one-bind-credential` | `spec.bindCredentials` or `spec.ldapCredentials` | LDAPAuthEngineConfig (`bindCredentials`), LDAPSecretEngineConfig (`bindCredentials`), KerberosAuthEngineLDAPConfig (`ldapCredentials`) | CREATE, UPDATE |
| Individual per-CRD credential VAPs | Various paths | See each CRD below | CREATE, UPDATE |

**Note on bind/LDAP credential grouping:** LDAPAuthEngineConfig and LDAPSecretEngineConfig both use `spec.bindCredentials` so they can share a VAP. KerberosAuthEngineLDAPConfig uses `spec.ldapCredentials` — a different JSON path — so it needs its own expression or a separate binding. The CEL expression structure is identical; only the field path differs. Consider using `spec.bindCredentials` and `spec.ldapCredentials` in a combined VAP with `||` if the Kerberos LDAP type also has `.vaultSecret`, `.secret`, `.randomSecret` at the same nesting depth.

**Single-CRD credential groups** (each needs its own VAP or can be folded into the CRD-specific VAP):
- `AWSAuthEngineClientConfig` → `spec.AWSCredentials` (note: capital-A JSON key)
- `AzureSecretEngineConfig` → `spec.azureCredentials`
- `GCPSecretEngineConfig` → `spec.gcpCredentials` (custom `GCPCredentialConfig` type, same structure)
- `OktaAuthEngineConfig` → `spec.oktaCredentials`
- `RADIUSAuthEngineConfig` → `spec.radiusCredentials`
- `CFAuthEngineConfig` → `spec.cfCredentials`
- `UserpassAuthEngineUser` → `spec.passwordCredentials`
- `TerraformCloudSecretEngineConfig` → `spec.tfcCredentials` (custom `TFCCredentialConfig`, same structure)
- `KubernetesSecretEngineConfig` → `spec.jwtReference` (with extra constraint: no randomSecret)
- `SSHSecretEngineConfig` → `spec.caKeyReference` (pointer field, with extra constraint: no randomSecret)
- `GitHubSecretEngineConfig` → `spec.sSHKeyReference` (custom: no randomSecret field at all, must have exactly one of vaultSecret/secret)

For single-CRD credential groups, the preferred approach is to **fold the credential validation into the CRD-specific VAP** rather than creating a separate shared VAP. This reduces total policy count.

### CEL Expression Patterns

#### Credential Source Mutual Exclusivity (exactly-one)
```yaml
# For CREATE and UPDATE — spec.rootCredentials must have exactly one source
validations:
  - expression: |
      (has(object.spec.rootCredentials.vaultSecret) ? 1 : 0) +
      (has(object.spec.rootCredentials.secret) ? 1 : 0) +
      (has(object.spec.rootCredentials.randomSecret) ? 1 : 0) == 1
    message: "exactly one of spec.rootCredentials.vaultSecret, spec.rootCredentials.secret, or spec.rootCredentials.randomSecret must be specified"
```

#### Immutable Fields on Update (TransitSecretEngineKey example)
```yaml
# For UPDATE only — spec.type cannot change after creation
validations:
  - expression: "object.spec.type == oldObject.spec.type"
    message: "spec.type cannot be updated after key creation"
```

#### One-Way Flag (can go false→true, but not true→false)
```yaml
# TransitSecretEngineKey: exportable and allowPlaintextBackup are one-way flags
validations:
  - expression: "!oldObject.spec.exportable || object.spec.exportable"
    message: "spec.exportable cannot be changed from true to false (one-way flag)"
```

#### Conditional Field Restriction (credentialType-dependent)
```yaml
# AWSSecretEngineRole: externalID only when credentialType is assumed_role
validations:
  - expression: |
      object.spec.credentialType == "assumed_role" ||
      !has(object.spec.externalID) || object.spec.externalID == ""
    message: "spec.externalID is only allowed when credentialType is assumed_role"
```

#### Duration Validation in CEL
```yaml
# TransitSecretEngineKey: autoRotatePeriod must be >= 1h when set and non-zero
validations:
  - expression: |
      !has(object.spec.autoRotatePeriod) ||
      object.spec.autoRotatePeriod == "" ||
      object.spec.autoRotatePeriod == "0" ||
      duration(object.spec.autoRotatePeriod) >= duration("1h")
    message: "spec.autoRotatePeriod must be at least 1h when enabled"
```

**CEL duration() caveat:** CEL's `duration()` function parses a subset of Go duration format. If Vault/Go duration strings use formats not supported by CEL (e.g., days), the expression may need adjustment. Test with actual field values.

#### Mount "Only Config Modifiable" Guard
```yaml
# AuthEngineMount: on UPDATE, only .spec.config fields can change
validations:
  - expression: |
      (!has(object.spec.authMount.description) && !has(oldObject.spec.authMount.description) ||
       has(object.spec.authMount.description) && has(oldObject.spec.authMount.description) &&
       object.spec.authMount.description == oldObject.spec.authMount.description) &&
      (!has(object.spec.authMount.type) && !has(oldObject.spec.authMount.type) ||
       has(object.spec.authMount.type) && has(oldObject.spec.authMount.type) &&
       object.spec.authMount.type == oldObject.spec.authMount.type) &&
      (!has(object.spec.authMount.local) && !has(oldObject.spec.authMount.local) ||
       object.spec.authMount.local == oldObject.spec.authMount.local) &&
      (!has(object.spec.authMount.sealwrap) && !has(oldObject.spec.authMount.sealwrap) ||
       object.spec.authMount.sealwrap == oldObject.spec.authMount.sealwrap)
    message: "only .spec.config can be modified"
```
**Note:** CEL cannot use `reflect.DeepEqual`. Each immutable sub-field must be compared explicitly. The Go webhook zeroes out the Config struct and uses `reflect.DeepEqual` — the CEL equivalent checks every non-config field individually.

### File Structure Requirements

All new VAP manifests go under `config/vap/`:
```
config/vap/
├── kustomization.yaml                              # (already exists from 18.1)
├── vco-immutable-path.yaml                         # (from 18.1)
├── vco-immutable-path-binding.yaml                 # (from 18.1)
├── vco-immutable-name.yaml                         # (from 18.1)
├── vco-immutable-name-binding.yaml                 # (from 18.1)
├── vco-target-namespace-guard.yaml                 # (from 18.1)
├── vco-target-namespace-guard-binding.yaml         # (from 18.1)
├── vco-target-namespace-guard-params.yaml          # (from 18.1)
│
│   # NEW — Credential Source VAPs (this story)
├── vco-exactly-one-root-credential.yaml
├── vco-exactly-one-root-credential-binding.yaml
├── vco-exactly-one-bind-credential.yaml
├── vco-exactly-one-bind-credential-binding.yaml
│
│   # NEW — CRD-Specific VAPs (this story)
├── vco-validate-transitsecretenginekey.yaml
├── vco-validate-transitsecretenginekey-binding.yaml
├── vco-validate-awssecretenginerole.yaml
├── vco-validate-awssecretenginerole-binding.yaml
├── vco-validate-randomsecret.yaml
├── vco-validate-randomsecret-binding.yaml
├── vco-validate-databasesecretenginestaticrole.yaml
├── vco-validate-databasesecretenginestaticrole-binding.yaml
├── vco-validate-totpsecretenginekey.yaml
├── vco-validate-totpsecretenginekey-binding.yaml
├── vco-validate-terraformcloudsecretenginerole.yaml
├── vco-validate-terraformcloudsecretenginerole-binding.yaml
├── vco-validate-authenginemount.yaml
├── vco-validate-authenginemount-binding.yaml
├── vco-validate-secretenginemount.yaml
├── vco-validate-secretenginemount-binding.yaml
├── vco-validate-pkisecretengineconfig.yaml
├── vco-validate-pkisecretengineconfig-binding.yaml
├── vco-validate-kubernetesauthengineconfig.yaml
├── vco-validate-kubernetesauthengineconfig-binding.yaml
├── vco-validate-kubernetesauthenginerole.yaml
├── vco-validate-kubernetesauthenginerole-binding.yaml
├── vco-validate-sshsecretengineconfig.yaml
├── vco-validate-sshsecretengineconfig-binding.yaml
├── vco-validate-awssecretengineconfig.yaml
├── vco-validate-awssecretengineconfig-binding.yaml
├── vco-validate-consulsecretengineconfig.yaml
├── vco-validate-consulsecretengineconfig-binding.yaml
├── vco-validate-nomadsecretengineconfig.yaml
├── vco-validate-nomadsecretengineconfig-binding.yaml
├── vco-validate-mongodbatlassecretengineconfig.yaml
├── vco-validate-mongodbatlassecretengineconfig-binding.yaml
├── vco-validate-ldapsecretengineconfig.yaml
├── vco-validate-ldapsecretengineconfig-binding.yaml
├── vco-validate-terraformcloudsecretengineconfig.yaml
├── vco-validate-terraformcloudsecretengineconfig-binding.yaml
├── vco-validate-cfauthengineconfig.yaml
├── vco-validate-cfauthengineconfig-binding.yaml
├── vco-validate-githubsecretengineconfig.yaml
├── vco-validate-githubsecretengineconfig-binding.yaml
├── vco-validate-kubernetessecretengineconfig.yaml
├── vco-validate-kubernetessecretengineconfig-binding.yaml
├── vco-validate-userpassauthengineuser.yaml
├── vco-validate-userpassauthengineuser-binding.yaml
│
│   # Credential-only VAPs for single-CRD groups (alternative: fold into CRD-specific VAPs above)
├── vco-validate-awsauthengineconfig.yaml           # spec.AWSCredentials (note capital A)
├── vco-validate-awsauthengineconfig-binding.yaml
├── vco-validate-azuresecretengineconfig.yaml       # spec.azureCredentials
├── vco-validate-azuresecretengineconfig-binding.yaml
├── vco-validate-gcpsecretengineconfig.yaml         # spec.gcpCredentials
├── vco-validate-gcpsecretengineconfig-binding.yaml
├── vco-validate-oktaauthengineconfig.yaml          # spec.oktaCredentials
├── vco-validate-oktaauthengineconfig-binding.yaml
├── vco-validate-radiusauthengineconfig.yaml        # spec.radiusCredentials
├── vco-validate-radiusauthengineconfig-binding.yaml
└── vco-validate-kerberosauthengineconfig-ldap.yaml # spec.ldapCredentials
    vco-validate-kerberosauthengineconfig-ldap-binding.yaml
```

**Naming convention:** `vco-validate-<lowercase-kind>` for CRD-specific VAPs, matching the epics file specification.

### VAP Manifest Structure

Each VAP manifest follows this structure:
```yaml
apiVersion: admissionregistration.k8s.io/v1
kind: ValidatingAdmissionPolicy
metadata:
  name: vco-validate-<kind>
spec:
  failurePolicy: Fail
  matchConstraints:
    resourceRules:
      - apiGroups: ["redhatcop.redhat.io"]
        apiVersions: ["v1alpha1"]
        operations: ["CREATE", "UPDATE"]  # or ["UPDATE"] for update-only rules
        resources: ["<plural-resource-name>"]
  validations:
    - expression: "<CEL expression>"
      message: "<error message>"
```

Each binding:
```yaml
apiVersion: admissionregistration.k8s.io/v1
kind: ValidatingAdmissionPolicyBinding
metadata:
  name: vco-validate-<kind>
spec:
  policyName: vco-validate-<kind>
  validationActions: ["Deny"]
```

**For shared credential VAPs** (covering multiple CRDs), the `matchConstraints.resourceRules` lists all applicable plural resource names:
```yaml
resources: ["awssecretengineconfigs", "consulsecretengineconfigs", "databasesecretengineconfigs", ...]
```

### CRD Plural Resource Names Reference

The plural resource names (for VAP `resources` fields) are derived from the Kind by lowercasing and appending 's':
- AWSSecretEngineConfig → `awssecretengineconfigs`
- AWSSecretEngineRole → `awssecretengineroles`
- AuthEngineMount → `authenginemounts`
- AzureSecretEngineConfig → `azuresecretengineconfigs`
- CFAuthEngineConfig → `cfauthengineconfigs`
- ConsulSecretEngineConfig → `consulsecretengineconfigs`
- DatabaseSecretEngineConfig → `databasesecretengineconfigs`
- DatabaseSecretEngineStaticRole → `databasesecretenginestaticroles`
- GCPSecretEngineConfig → `gcpsecretengineconfigs`
- GitHubSecretEngineConfig → `githubsecretengineconfigs`
- KerberosAuthEngineLDAPConfig → `kerberosauthengineconfigs` (verify via `make manifests` output — Kerberos LDAP uses a different resource name)
- KubernetesAuthEngineConfig → `kubernetesauthengineconfigs`
- KubernetesAuthEngineRole → `kubernetesauthengineroles`
- KubernetesSecretEngineConfig → `kubernetessecretengineconfigs`
- LDAPAuthEngineConfig → `ldapauthengineconfigs`
- LDAPSecretEngineConfig → `ldapsecretengineconfigs`
- MongoDBAtlasSecretEngineConfig → `mongodbatlassecretengineconfigs`
- NomadSecretEngineConfig → `nomadsecretengineconfigs`
- OktaAuthEngineConfig → `oktaauthengineconfigs`
- PKISecretEngineConfig → `pkisecretengineconfigs`
- QuaySecretEngineConfig → `quaysecretengineconfigs`
- RabbitMQSecretEngineConfig → `rabbitmqsecretengineconfigs`
- RADIUSAuthEngineConfig → `radiusauthengineconfigs`
- RandomSecret → `randomsecrets`
- SSHSecretEngineConfig → `sshsecretengineconfigs`
- SecretEngineMount → `secretenginemounts`
- TOTPSecretEngineKey → `totpsecretenginekeys`
- TerraformCloudSecretEngineConfig → `terraformcloudsecretengineconfigs`
- TerraformCloudSecretEngineRole → `terraformcloudsecretengineroles`
- TransitSecretEngineKey → `transitsecretenginekeys`
- UserpassAuthEngineUser → `userpassauthengineusers`

**IMPORTANT:** Verify all plural names against `config/crd/bases/` YAML file names. The CRD filename contains the plural: `redhatcop.redhat.io_<plural>.yaml`.

### CEL `has()` Semantics for Kubernetes CRD Fields

- `has(object.spec.field)` returns `false` when the field is absent from the JSON payload OR when the field has a zero value **and** is tagged `omitempty`.
- For pointer fields (`*corev1.LocalObjectReference`), `has()` returns `false` when nil, `true` when non-nil (even if the struct is empty).
- For string fields with `omitempty`: `has()` returns `false` when empty string, `true` when non-empty.
- For non-omitempty fields (e.g., fields with `+kubebuilder:default`): `has()` always returns `true` since the API server populates the default.

This affects how credential source checks are written. For `RootCredentialConfig`, `.vaultSecret`, `.secret`, and `.randomSecret` are pointer fields with `omitempty`, so `has()` works correctly for nil checks.

### Specific Validation Logic Reference (Go → CEL Translation)

#### TransitSecretEngineKey — Full Validation Set

**CREATE validations:**
1. `convergentEncryption → derived` must be true
2. `keySize > 0` → `type` must be `"hmac"` (and keySize ≥ 32)
3. `autoRotatePeriod` set and non-zero → duration ≥ 1h

**UPDATE validations (in addition to CREATE):**
4. `spec.type` immutable
5. `spec.derived` immutable
6. `spec.convergentEncryption` immutable
7. `spec.keySize` immutable
8. `spec.exportable` one-way (false→true OK, true→false rejected)
9. `spec.allowPlaintextBackup` one-way (false→true OK, true→false rejected)
10. `convergentEncryption && !derived` (recheck on update)

**Note:** The existing Go webhook in `transitsecretenginekey_webhook.go` calls `validateTransitKeyConstraints()` for both create and update. The CEL expressions must cover both operation types where applicable, using `request.operation == "UPDATE"` guards for update-only rules.

#### RandomSecret — Inline Password Policy Validation

The Go code validates `InlinePasswordPolicy` via HCL parsing (`hclsimple.Decode`). **CEL cannot parse HCL.** Options:
1. **Skip HCL validation in VAP** — keep the webhook for this specific check only (not recommended since Story 18.3 removes webhook code)
2. **Keep a residual validating webhook** for RandomSecret specifically for inline HCL validation
3. **Drop the HCL validation** — Vault will reject invalid HCL at write time anyway

**Recommended:** Keep the RandomSecret webhook as a residual for inline HCL validation. Document this in Story 18.3 residual webhook list. The VAP handles everything else (mutual exclusivity, secretKey immutability, KVv2 path).

#### AuthEngineMount / SecretEngineMount — Field-by-Field Comparison

The Go code uses `reflect.DeepEqual` after zeroing the Config struct. In CEL, compare each non-config field individually:

**AuthEngineMount immutable fields:** `description`, `type`, `local`, `sealwrap`
**SecretEngineMount immutable fields:** `type`, `description`, `local`, `sealWrap`, `externalEntropyAccess`, `options`

For `options` (map type), CEL comparison: `object.spec.mount.options == oldObject.spec.mount.options` works for maps if both sides are present. Handle absent maps with `has()` checks.

#### AWSSecretEngineRole — Complete credentialType Matrix

| Field | assumed_role | iam_user | federation_token | session_token |
|-------|:---:|:---:|:---:|:---:|
| externalID | ✅ | ❌ | ❌ | ❌ |
| sessionTags | ✅ | ❌ | ❌ | ❌ |
| iamTags | ❌ | ✅ | ❌ | ❌ |
| userPath | ❌ | ✅ | ❌ | ❌ |
| permissionsBoundaryARN | ❌ | ✅ | ❌ | ❌ |
| defaultSTSTTL | ✅ | ❌ | ✅ | ❌ |
| maxSTSTTL | ✅ | ❌ | ✅ | ❌ |
| roleArns | ✅ (required) | ❌ | ❌ | ❌ |
| policyArns | ✅ | ✅ | ✅ | ❌ |
| policyDocument | ✅ | ✅ | ✅ | ❌ |
| iamGroups | ✅ | ✅ | ✅ | ❌ |

### Testing Requirements

**This story creates VAP manifests only.** Testing is covered by Story 18.4 (VAP Test Suite). However:
- Manually verify each VAP manifest is valid YAML by running `kubectl apply --dry-run=server -f config/vap/<file>.yaml` against a K8s 1.30+ cluster, OR by running `make manifests` to ensure kustomize can build the full overlay without errors.
- Verify the kustomize build includes all new VAP files: `kustomize build config/default | grep ValidatingAdmissionPolicy | wc -l` should reflect the expected count.

### Architecture Compliance

- **API group:** `admissionregistration.k8s.io/v1` — VAPs are GA in Kubernetes 1.30+
- **CEL runtime:** Kubernetes 1.30 supports CEL v1.16 — all used CEL functions (`has()`, `duration()`, string comparisons, `size()`, ternary `? :`) are available
- **Shared framework protection:** This story does NOT modify any shared framework files. All changes are new YAML manifests under `config/vap/`.
- **No Go code changes in this story.** Webhook code removal happens in Story 18.3 after VAP tests pass in Story 18.4.

### Library/Framework Requirements

- **Kubernetes API:** `admissionregistration.k8s.io/v1` — `ValidatingAdmissionPolicy` and `ValidatingAdmissionPolicyBinding`
- **CEL:** Common Expression Language as implemented in the Kubernetes API server (kube-cel-lib). Reference: https://kubernetes.io/docs/reference/using-api/cel/
- **Kustomize:** Standard kustomize overlay under `config/vap/`

### Novelty Risk

**MEDIUM** — CEL expressions are well-documented, but some validations require careful translation from Go:
1. Duration parsing: CEL `duration()` supports a subset of Go's `time.ParseDuration` — test with actual field values
2. Map comparison: CEL `==` on maps works for simple types but may need care with nested structs
3. Optional field semantics: `has()` behavior with `omitempty` and pointer types needs careful testing
4. The TransitSecretEngineKey VAP is the most complex with ~10 validation rules combining create and update semantics

### Project Structure Notes

- All manifests under `config/vap/` following the pattern established by Story 18.1
- Kustomize overlay updated to include all new manifests
- No changes to Go source code, controllers, or webhook files

### References

- [Source: `_bmad-output/planning-artifacts/epics.md` lines 3386–3430 — Story 18.2 specification]
- [Source: `api/v1alpha1/utils/commons.go` lines 469–516 — RootCredentialConfig struct and ValidateCredentialSource()]
- [Source: `api/v1alpha1/transitsecretenginekey_webhook.go` lines 64–124 — TransitSecretEngineKey validation logic]
- [Source: `api/v1alpha1/awssecretenginerole_types.go` lines 175–230 — AWSSecretEngineRole credentialType validation]
- [Source: `api/v1alpha1/randomsecret_types.go` lines 301–352 — RandomSecret validation (passwordPolicy, secretKey, KVv2)]
- [Source: `api/v1alpha1/databasesecretenginestaticrole_types.go` lines 168–203 — rotation and password/key validation]
- [Source: `api/v1alpha1/totpsecretenginekey_webhook.go` lines 58–84 — TOTP generate-mode validation]
- [Source: `api/v1alpha1/terraformcloudsecretenginerole_types.go` lines 114–150 — TFC role credentialType validation]
- [Source: `api/v1alpha1/authenginemount_webhook.go` lines 65–85 — only-config-modifiable guard]
- [Source: `api/v1alpha1/secretenginemount_webhook.go` lines 65–85 — only-config-modifiable guard]
- [Source: `api/v1alpha1/pkisecretengineconfig_types.go` lines 517–522 — ACME/cluster validation]
- [Source: `api/v1alpha1/kubernetesauthengineconfig_types.go` lines 134–139 — tokenReviewer mutual exclusivity]
- [Source: `api/v1alpha1/kubernetesauthenginerole_types.go` lines 250–266 — targetNamespace validation]
- [Source: `api/v1alpha1/sshsecretengineconfig_types.go` lines 170–181 — generateSigningKey validation]
- [Source: `api/v1alpha1/awssecretengineconfig_types.go` lines 231–239 — accessKey+randomSecret constraint]
- [Source: `api/v1alpha1/consulsecretengineconfig_types.go` lines 199–207 — clientCert/clientKey pairing]
- [Source: `api/v1alpha1/nomadsecretengineconfig_types.go` lines 203–211 — clientCert/clientKey pairing]
- [Source: `api/v1alpha1/mongodbatlassecretengineconfig_types.go` lines 203–211 — publicKey+randomSecret constraint]
- [Source: `api/v1alpha1/ldapsecretengineconfig_types.go` lines 176–184 — bindDN+randomSecret constraint]
- [Source: `api/v1alpha1/terraformcloudsecretengineconfig_types.go` lines 148–156 — passwordKey constraint]
- [Source: `api/v1alpha1/cfauthengineconfig_types.go` lines 268–276 — cfUsername+randomSecret constraint]
- [Source: `api/v1alpha1/githubsecretengineconfig_types.go` lines 214–230 — vaultSecret/secret exclusivity]
- [Source: `api/v1alpha1/kubernetessecretengineconfig_types.go` lines 133–143 — no randomSecret for jwtReference]
- [Source: `api/v1alpha1/userpassauthengineuser_types.go` lines 267–268 — passwordCredentials validation]
- [Source: `_bmad-output/project-context.md` — Admission webhook rules, CEL reference]
- Kubernetes CEL documentation: https://kubernetes.io/docs/reference/using-api/cel/
- ValidatingAdmissionPolicy API: https://kubernetes.io/docs/reference/access-authn-authz/validating-admission-policy/

## Dev Agent Record

### Agent Model Used
Claude Opus 4.6 (implementation), Grok 4.7 xhigh (review — quick lens)

### Summary of Changes
61 files changed: created 30 ValidatingAdmissionPolicies + 30 ValidatingAdmissionPolicyBindings under `config/vap/`, updated `config/vap/kustomization.yaml`. No Go code changes.

### Outcome
All gates passed: unit tests, integration tests (144 specs), helmchart-test. `kustomize build config/default/` produces 66 VAP resources (6 existing + 60 new). Review found 2 high-severity bugs (AuthEngineMount/SecretEngineMount using wrong field paths due to `json:",inline"`) — fixed in loopback.

### Notes
- RandomSecret HCL inline password policy validation intentionally NOT in CEL (stays in Go webhook until Story 18.3)
- CEL `duration()` for TransitSecretEngineKey autoRotatePeriod needs integration testing in Story 18.4
- TFC credential inner field emptiness check not replicated in VAP (Go webhook still catches it)

## Code Review Record

### Review Model Used
Grok 4.7 xhigh (quick lens)

### Review Findings
6 findings: 2 high (AuthEngineMount/SecretEngineMount inline field paths), 1 low (TFC empty credential), 3 false (omitempty field access — K8s CEL returns zero values for missing schema-declared optional primitives)

### Decisions Needed / Decisions Taken
AuthEngineMount/SecretEngineMount field paths fixed (bad_plan loopback). TFC empty credential gap accepted (Go webhook still catches it; deferred to Story 18.3/18.4).

### Fixes Applied
- `vco-validate-authenginemount.yaml`: `spec.authMount.*` → `spec.*` (AuthMount is `json:",inline"`)
- `vco-validate-secretenginemount.yaml`: `spec.mount.*` → `spec.*` (Mount is `json:",inline"`)
