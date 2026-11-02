# Admission Control

- [Admission Control](#admission-control)
  - [Overview](#overview)
  - [Architecture — Three-Tier VAP Strategy](#architecture--three-tier-vap-strategy)
    - [Tier 1: Universal Shared Policies](#tier-1-universal-shared-policies)
    - [Tier 2: Credential Source Shared Policies](#tier-2-credential-source-shared-policies)
    - [Tier 3: CRD-Specific Policies](#tier-3-crd-specific-policies)
  - [Parameterized VAP — Target Namespace Guard](#parameterized-vap--target-namespace-guard)
  - [Residual Webhooks](#residual-webhooks)
  - [Kubernetes Version Requirements](#kubernetes-version-requirements)
  - [Manifest Location](#manifest-location)

## Overview

Kubernetes [admission control](https://kubernetes.io/docs/reference/access-authn-authz/admission-controllers/) intercepts API requests before objects are persisted to etcd. vault-config-operator uses **ValidatingAdmissionPolicies (VAPs)** to enforce validation rules declaratively via [CEL expressions](https://kubernetes.io/docs/reference/using-api/cel/). Residual Go webhooks handle mutation (defaulting), cross-resource checks, and type-specific validation that has not yet been migrated to VAPs.

VAPs replaced the per-type Go `ValidateUpdate` webhook code that previously existed in every `*_webhook.go` file. The migration removed thousands of lines of repetitive Go validation and consolidated all validation rules into YAML manifests under `config/vap/`.

## Architecture — Three-Tier VAP Strategy

Validation rules are organized into three tiers based on how many CRD types each policy applies to.

### Tier 1: Universal Shared Policies

These policies apply to large groups of CRDs that share a common structural pattern.

| Policy | Rule | CRDs | Operation |
|--------|------|------|-----------|
| `vco-immutable-path` | Prevents `spec.path` changes on UPDATE | 72 | UPDATE |
| `vco-immutable-name` | Prevents `spec.name` changes on UPDATE (including adding `spec.name` after creation) | 65 | UPDATE |
| `vco-target-namespace-guard` | Prevents Vault namespace changes when `targetNamespace` is set on either old or new resource | 96 (parameterized) | UPDATE |

**`vco-immutable-path`** — Changing `spec.path` after creation would orphan the Vault object at the old mount path. The CEL expression is:

```cel
object.spec.path == oldObject.spec.path
```

**`vco-immutable-name`** — Changing `spec.name` after creation would orphan the Vault object at the old name. Adding `spec.name` when it was previously empty also changes the effective Vault path (from `metadata.name` to `spec.name`). The CEL expression handles the absent-field case using `has()`:

```cel
!has(oldObject.spec.name) && !has(object.spec.name) ||
has(oldObject.spec.name) && has(object.spec.name) && object.spec.name == oldObject.spec.name
```

**`vco-target-namespace-guard`** — Replicates the Go `ValidateTargetNamespaceUpdate` logic. See [Parameterized VAP — Target Namespace Guard](#parameterized-vap--target-namespace-guard) below.

### Tier 2: Credential Source Shared Policies

These policies enforce mutual exclusivity for credential source fields shared across multiple CRD types.

| Policy | Field Path | CRDs | Rule |
|--------|-----------|------|------|
| `vco-exactly-one-root-credential` | `spec.rootCredentials` | DatabaseSecretEngineConfig, QuaySecretEngineConfig, RabbitMQSecretEngineConfig | Exactly one of `.vaultSecret`, `.secret`, `.randomSecret` must be set |
| `vco-exactly-one-bind-credential` | `spec.bindCredentials` | LDAPAuthEngineConfig | Exactly one credential source must be set |

Other CRDs with root/bind credential validation (AWSSecretEngineConfig, ConsulSecretEngineConfig, MongoDBAtlasSecretEngineConfig, NomadSecretEngineConfig, LDAPSecretEngineConfig, KerberosAuthEngineLDAPConfig) have their credential checks folded into their CRD-specific VAPs in Tier 3 alongside additional type-specific constraints.

The CEL pattern for credential mutual exclusivity:

```cel
(has(object.spec.rootCredentials.vaultSecret) ? 1 : 0) +
(has(object.spec.rootCredentials.secret) ? 1 : 0) +
(has(object.spec.rootCredentials.randomSecret) ? 1 : 0) == 1
```

### Tier 3: CRD-Specific Policies

These policies enforce validation rules unique to individual CRD types. Each follows the naming pattern `vco-validate-<lowercase-kind>`.

| Policy | Description |
|--------|-------------|
| `vco-validate-authenginemount` | Only `.spec.config` can be modified on UPDATE; description, type, local, sealWrap are immutable |
| `vco-validate-awsauthengineconfig` | AWS credential source mutual exclusivity |
| `vco-validate-awssecretengineconfig` | Root credentials + accessKey required when using randomSecret |
| `vco-validate-awssecretenginerole` | credentialType-dependent field restrictions (assumed_role, iam_user, federation_token, session_token) |
| `vco-validate-azuresecretengineconfig` | Azure credential source mutual exclusivity |
| `vco-validate-cfauthengineconfig` | CF credential source + cfUsername required when using randomSecret |
| `vco-validate-consulsecretengineconfig` | Root credentials + clientCert/clientKey must both be set or both empty |
| `vco-validate-databasesecretenginestaticrole` | rotationPeriod/rotationSchedule mutual exclusivity, password/key credential config exactly-one |
| `vco-validate-gcpsecretengineconfig` | GCP credential source mutual exclusivity |
| `vco-validate-githubsecretengineconfig` | Exactly one of sshKeyReference.vaultSecret or sshKeyReference.secret (no randomSecret) |
| `vco-validate-kerberosauthengineconfig-ldap` | LDAP credential source mutual exclusivity |
| `vco-validate-kubernetesauthengineconfig` | tokenReviewerServiceAccount and tokenReviewerJWTSecret are mutually exclusive |
| `vco-validate-kubernetesauthenginerole` | targetNamespaceSelector vs targetNamespaces mutual exclusivity |
| `vco-validate-kubernetessecretengineconfig` | jwtReference.randomSecret is forbidden (only vaultSecret or secret) |
| `vco-validate-ldapsecretengineconfig` | Bind credentials + bindDN required when using randomSecret |
| `vco-validate-mongodbatlassecretengineconfig` | Root credentials + publicKey required when using randomSecret |
| `vco-validate-nomadsecretengineconfig` | Root credentials + clientCert/clientKey must both be set or both empty |
| `vco-validate-oktaauthengineconfig` | Okta credential source mutual exclusivity |
| `vco-validate-pkisecretengineconfig` | ACME requires cluster (`spec.acme` set but `spec.cluster` nil → reject) |
| `vco-validate-radiusauthengineconfig` | RADIUS credential source mutual exclusivity |
| `vco-validate-randomsecret` | passwordPolicy mutual exclusivity, secretKey immutability on UPDATE, KVv2 `/data/` path check |
| `vco-validate-secretenginemount` | Only `.spec.config` can be modified on UPDATE; type, description, local, sealWrap, externalEntropyAccess, options are immutable |
| `vco-validate-sshsecretengineconfig` | When generateSigningKey=false: caKeyReference required, caKeyReference.randomSecret forbidden |
| `vco-validate-terraformcloudsecretengineconfig` | TFC credential source + passwordKey required when using secret or vaultSecret |
| `vco-validate-terraformcloudsecretenginerole` | credentialType enum {team, team_legacy, user, organization} + type-dependent required fields |
| `vco-validate-totpsecretenginekey` | issuer and accountName required; generate-mode conditional rules |
| `vco-validate-transitsecretenginekey` | convergent/derived, keySize/hmac, autoRotatePeriod ≥ 1h, immutable fields (type, derived, convergentEncryption, keySize), one-way flags (exportable, allowPlaintextBackup) |
| `vco-validate-userpassauthengineuser` | Password credential source validation (exactly one source) |

## Parameterized VAP — Target Namespace Guard

The `vco-target-namespace-guard` policy is a **parameterized VAP** — it reads runtime configuration from a ConfigMap to replicate the Vault namespace fallback logic previously implemented in the Go `ValidateTargetNamespaceUpdate` method (now removed).

### How It Works

The Go code computes an "effective target namespace" by joining the login namespace with the target namespace:
1. **Login namespace** = `spec.authentication.namespace` (if set), otherwise the `VAULT_NAMESPACE` environment variable
2. **Effective target namespace** = `CleansePath(loginNamespace + "/" + targetNamespace)`

The VAP replicates this using CEL variables and a ConfigMap parameter:

```yaml
paramKind:
  apiVersion: v1
  kind: ConfigMap
variables:
- name: oldLoginNs
  expression: >-
    has(oldObject.spec.authentication.namespace) ? oldObject.spec.authentication.namespace
    : (has(params.data) && 'vaultNamespace' in params.data ? params.data.vaultNamespace : '')
```

### ConfigMap Setup

The `vco-vap-params` ConfigMap provides the `VAULT_NAMESPACE` fallback value:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: vco-vap-params
  namespace: vault-config-operator  # must match operator namespace
data:
  vaultNamespace: ""  # set to VAULT_NAMESPACE env var value if used
```

The ConfigMap's namespace is injected by kustomize via a `replacements:` entry in `config/default/kustomization.yaml` — it uses the operator namespace as the source.

When `VAULT_NAMESPACE` is not used (the common case), leave `vaultNamespace` empty.

### Guard Behavior

The policy **only enforces** when `targetNamespace` is set on either the old or new resource. When neither has `targetNamespace`, namespace changes are allowed (preserving existing behavior for single-namespace deployments).

## Residual Webhooks

Of the 96 CRD types, 72 have no webhook file at all. The remaining 24 webhook files fall into three categories:

### Cross-Resource Validator

| Type | Webhook Purpose | Why Not VAP |
|------|----------------|-------------|
| RabbitMQSecretEngineConfig | `validateUniquePath` — ensures no two RabbitMQ configs share the same mount path | Requires `client.List()` to query existing resources; CEL has no API server access |

### Mutating Defaulters (Credential Key Remapping)

These 7 webhook files were gutted to defaulter-only — their `admission.Validator` interface was removed and validation is handled entirely by VAPs:

| Type | Default Applied | Why Not VAP |
|------|----------------|-------------|
| CFAuthEngineConfig | Remaps `usernameKey`/`passwordKey` via raw admission request inspection | Mutation (not validation) |
| GCPSecretEngineConfig | Remaps `passwordKey` from `"password"` to `"credentials"` | Mutation (not validation) |
| KubernetesAuthEngineConfig | Reads CA cert from `/var/run/secrets/.../ca.crt` filesystem path | Mutation (not validation); requires filesystem access |
| OktaAuthEngineConfig | Defaults `passwordKey` to `"api_token"` | Mutation (not validation) |
| RADIUSAuthEngineConfig | Remaps `passwordKey` from `"password"` to `"secret"` | Mutation (not validation) |
| SSHSecretEngineConfig | Conditional remap of `passwordKey`/`usernameKey` when `caKeyReference` is set | Mutation (not validation) |
| TerraformCloudSecretEngineConfig | Remaps `passwordKey` from `"password"` to `"token"` | Mutation (not validation) |

### Retained Webhook Validators

These 16 types retain Go `admission.Validator` implementations for type-specific validation that was not migrated to VAPs. Some also have defaulting logic:

| Type | Validation Logic | Defaulter? |
|------|-----------------|------------|
| AWSAuthEngineClientConfig | Credential source `isValid()` check | Yes — credential key remapping via raw admission request |
| AliCloudAuthEngineRole | ARN role name validation via `IsValid()` | No |
| AppRoleAuthEngineRole | `spec.localSecretIDs` immutability on UPDATE | No |
| AWSAuthEngineRole | `authType` (`ec2`/`iam`) field combination validation | No |
| CORSConfig | `spec.allowedOrigins` must not be empty | No |
| GCPSecretEngineRoleset | `spec.secretType` and `spec.project` immutability on UPDATE | No |
| GCPSecretEngineStaticAccount | `spec.secretType` and `spec.serviceAccountEmail` immutability on UPDATE | No |
| IdentityOIDCClient | `spec.key` and `spec.clientType` immutability on UPDATE | No |
| JWTOIDCAuthEngineRole | `allowedRedirectURIs` validation | No |
| KeyRotationConfig | `spec.interval` must be ≥ 24h | No |
| LDAPSecretEngineStaticRole | `spec.username` and `spec.dn` immutability on UPDATE | No |
| LeaseCountQuota | `spec.maxLeases` must be ≥ 1 | No |
| PKISecretEngineConfig | `spec.type` and `spec.privateKeyType` immutability + `isValid()` (supplements `vco-validate-pkisecretengineconfig` VAP) | No |
| RandomSecret | Password policy, secretKey immutability, inline HCL parse (supplements `vco-validate-randomsecret` VAP) | No |
| RateLimitQuota | `spec.rate` format validation | No |
| UICustomHeader | `spec.values` must not be empty | No |

**Key principle:** VAPs can only **reject** requests. They cannot **modify** the object being admitted. Any logic that sets default field values must remain in a Go mutating webhook (or use `+kubebuilder:default` CRD schema markers when applicable).

## Kubernetes Version Requirements

ValidatingAdmissionPolicies are **GA in Kubernetes 1.30** ([KEP-3488](https://github.com/kubernetes/enhancements/tree/master/keps/sig-api-machinery/3488-cel-admission-control)).

- **Minimum Kubernetes version:** 1.30.0 (`minKubeVersion` in the CSV)
- **Minimum OpenShift version:** 4.17 (ships Kubernetes 1.30)

The operator's `config/manifests/bases/vault-config-operator.clusterserviceversion.yaml` declares `minKubeVersion: 1.30.0` and `config/manifests/annotations.yaml` declares `com.redhat.openshift.versions: v4.17`.

## Manifest Location

All VAP manifests live under `config/vap/`:

```
config/vap/
├── kustomization.yaml                                    # lists all VAP resources
│
│   # Tier 1 — Universal shared policies
├── vco-immutable-path.yaml                               # ValidatingAdmissionPolicy (72 CRDs)
├── vco-immutable-path-binding.yaml                       # ValidatingAdmissionPolicyBinding
├── vco-immutable-name.yaml                               # ValidatingAdmissionPolicy (65 CRDs)
├── vco-immutable-name-binding.yaml                       # ValidatingAdmissionPolicyBinding
├── vco-target-namespace-guard.yaml                       # ValidatingAdmissionPolicy (96 CRDs, parameterized)
├── vco-target-namespace-guard-binding.yaml               # ValidatingAdmissionPolicyBinding
├── vco-vap-params-configmap.yaml                         # ConfigMap for VAULT_NAMESPACE fallback
│
│   # Tier 2 — Credential source shared policies
├── vco-exactly-one-root-credential.yaml                  # 3 CRDs sharing rootCredentials
├── vco-exactly-one-root-credential-binding.yaml
├── vco-exactly-one-bind-credential.yaml                  # 1 CRD sharing bindCredentials
├── vco-exactly-one-bind-credential-binding.yaml
│
│   # Tier 3 — CRD-specific policies (28 VAPs + 28 bindings)
├── vco-validate-<lowercase-kind>.yaml                    # ValidatingAdmissionPolicy
└── vco-validate-<lowercase-kind>-binding.yaml            # ValidatingAdmissionPolicyBinding
```

**Naming convention:** All policies use the `vco-` prefix (vault-config-operator). Shared policies use descriptive names (`vco-immutable-path`). CRD-specific policies use `vco-validate-<lowercase-kind>`.

**Kustomize integration:** The `config/vap/kustomization.yaml` lists all VAP and binding YAMLs as `resources:`. The `config/default/kustomization.yaml` includes `- ../vap` alongside `- ../webhook`, `- ../crd`, and `- ../rbac`.

**Total VAP resource count:** 33 ValidatingAdmissionPolicies + 33 ValidatingAdmissionPolicyBindings + 1 ConfigMap = **67 resources**.
