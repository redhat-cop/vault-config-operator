# Phase 2: Expansion Analysis Report

**Project:** vault-config-operator
**Date:** 2026-04-11
**Status:** Analysis complete

---

## Part A: Dependency Upgrade Inventory & Risk Assessment

### Direct Dependencies

| Dependency | Current | Latest | Gap | Risk | Notes |
|------------|---------|--------|-----|------|-------|
| **Go** | 1.22.0 | 1.24+ | 2 minor | **HIGH** | Latest controller-runtime and ginkgo require Go 1.24+ |
| **controller-runtime** | v0.17.3 | v0.23.3 | 6 minor | **HIGH** | K8s version coupling, API surface changes, webhook registration changes |
| **k8s.io/api** | v0.29.2 | v0.35.3 | 6 minor | **HIGH** | Must move in lockstep with controller-runtime |
| **k8s.io/apimachinery** | v0.29.2 | v0.35.3 | 6 minor | **HIGH** | Must move in lockstep with controller-runtime |
| **k8s.io/client-go** | v0.29.2 | v0.35.2 | 6 minor | **HIGH** | Must move in lockstep with controller-runtime |
| **k8s.io/apiextensions-apiserver** | v0.29.2 | v0.35.x | 6 minor | **HIGH** | Must move in lockstep |
| **vault/api** | v1.14.0 | v1.23.0 | 9 minor | **MEDIUM** | API additions unlikely to break; may require Go 1.24 |
| **Operator SDK** | v1.31.0 | v1.42.2 | 11 minor | **HIGH** | Scaffolding, Makefile targets, bundle generation, Dockerfile |
| **ginkgo/v2** | v2.19.0 | v2.28.1 | 9 minor | **LOW** | Test-only, backward compatible; v2.28+ requires Go 1.24 |
| **gomega** | v1.33.1 | v1.39.1 | 6 minor | **LOW** | Test-only, backward compatible; v1.39+ requires Go 1.24 |
| **go-logr/logr** | v1.4.2 | v1.4.3 | 1 patch | **LOW** | Minor fix |
| **hcl/v2** | v2.21.0 | v2.24.0 | 3 minor | **LOW** | Used for HCL policy parsing |
| **sprig/v3** | v3.2.3 | v3.3.0 | 1 minor | **LOW** | Template functions |
| **pkg/errors** | v0.9.1 | v0.9.1 (archived) | 0 | **INFO** | Repository archived; consider migration to `fmt.Errorf` with `%w` |
| **go-multierror** | v1.1.1 | v1.1.1 | 0 | **LOW** | Stable |
| **BurntSushi/toml** | v1.4.0 | v1.4.0 | 0 | **LOW** | Current |
| **scylladb/go-set** | v1.0.2 | v1.0.2 | 0 | **LOW** | Current |
| **sigs.k8s.io/yaml** | v1.4.0 | v1.4.0 | 0 | **LOW** | Current |

### Key Indirect Dependencies (Security/Compatibility)

| Dependency | Current | Notes |
|------------|---------|-------|
| golang.org/x/crypto | v0.23.0 | Should upgrade for security patches |
| golang.org/x/net | v0.25.0 | Should upgrade for security patches |
| google.golang.org/protobuf | v1.33.0 | Will be pulled up by K8s lib upgrade |
| prometheus/client_golang | v1.18.0 | Will be pulled up by controller-runtime upgrade |
| golang.org/x/oauth2 | v0.12.0 | Will be pulled up by K8s lib upgrade |

### Dockerfile Dependencies

| Item | Current | Latest | Gap | Notes |
|------|---------|--------|-----|-------|
| Builder image | `golang:1.22` | `golang:1.24+` | 2 minor | Must match go.mod Go version |
| Runtime base | `ubi9/ubi-minimal` | Current | OK | Red Hat UBI9 is current |
| Build GOARCH | Hardcoded `amd64` | Multi-arch | **Issue** | Should support arm64 for multi-arch builds |

### bundle.Dockerfile

| Item | Current | Notes |
|------|---------|-------|
| `operator-sdk-v1.31.0` label | v1.31.0 | Must update when Operator SDK is upgraded |
| `go.kubebuilder.io/v3` layout label | v3 | May change with Operator SDK upgrade |

### Makefile Tool Dependencies

| Tool | Current | Latest | Gap | Risk |
|------|---------|--------|-----|------|
| **HELM_VERSION** | v3.11.0 | v4.1.4 | **MAJOR (v3→v4)** | **HIGH** — Helm 4 has breaking changes; kustomize v5.8+ has Helm 4 compat fixes |
| **GOLANGCI_LINT_VERSION** | v1.59.1 | v2.11.4 | **MAJOR (v1→v2)** | **HIGH** — Config format changes, linter renames |
| **OPM** (hardcoded in Makefile) | v1.23.0 | v1.65.0 | **42 minor** | **MEDIUM** — OLM catalog builder |
| CONTROLLER_TOOLS_VERSION | v0.14.0 | v0.20.1 | 6 minor | **HIGH** — Must match controller-runtime version; new CRD markers |
| ENVTEST_VERSION | release-0.17 | release-0.23 | 6 minor | **HIGH** — Must match controller-runtime version |
| ENVTEST_K8S_VERSION | 1.29.0 | 1.35.0 | 6 minor | **HIGH** — Must match K8s lib version |
| KIND_VERSION | v0.27.0 | v0.31.0 | 4 minor | **MEDIUM** — Integration test cluster |
| KUBECTL_VERSION | v1.29.0 | v1.35.x | 6 minor | **MEDIUM** — Must be compatible with K8s version |
| KUSTOMIZE_VERSION | v5.4.3 | v5.8.1 | 4 minor | **LOW** — Backward compatible |
| VAULT_VERSION | 1.19.0 | 1.21.4 | 2 minor | **MEDIUM** — Integration test Vault server |
| VAULT_CHART_VERSION | 0.30.0 | Needs check | Unknown | Must match VAULT_VERSION |

### CI Workflow Dependencies

| Item | Location | Current | Latest | Gap |
|------|----------|---------|--------|-----|
| GO_VERSION | pr.yaml, push.yaml | ~1.22 | ~1.24+ | 2 minor |
| OPERATOR_SDK_VERSION | pr.yaml, push.yaml | v1.31.0 | v1.42.2 | 11 minor |
| redhat-cop/github-workflows-operators | pr.yaml, push.yaml | @5b04934 (v1.1.6) | Needs check | Pinned SHA |
| cert-manager (helmchart-test) | Makefile | v1.7.1 | Very old | **HIGH** — v1.7 is long EOL |

### Integration Test Vault Images

| File | Current | Latest |
|------|---------|--------|
| integration/vault-values.yaml | `hashicorp/vault:1.19.0` (4 occurrences) | `hashicorp/vault:1.21.4` |
| config/local-development/vault-values.yaml | `hashicorp/vault:1.19.2-ubi` (3 occurrences) | Needs UBI equivalent of 1.21.4 |

### Upgrade Strategy

**Critical path:** Go → controller-runtime + K8s libs (lockstep) → Operator SDK → vault/api → everything else.

The Go + controller-runtime + K8s libs upgrade is the **single largest risk** — it touches every controller, webhook, test suite, envtest setup, CI pipeline, and Dockerfile. This must be done as one coordinated effort and validated against the full test suite (which is why Phase 1 stabilization comes first).

**Recommended upgrade sequence:**
1. Go 1.22 → 1.24+ (minimum needed by latest deps)
2. controller-runtime v0.17 → v0.23 + K8s libs v0.29 → v0.35 + controller-gen v0.14 → v0.20 + envtest release-0.17 → release-0.23 (must be simultaneous)
3. Operator SDK v1.31 → v1.42 (Makefile, Dockerfile, bundle changes)
4. vault/api v1.14 → v1.23 (can be independent but may benefit from Go 1.24)
5. Test deps (ginkgo, gomega) — can ride along with step 1 or 2
6. Peripheral deps (hcl, sprig, logr) — low risk, any time
7. Tooling upgrades: Helm v3→v4, golangci-lint v1→v2, OPM, Kind, kubectl, kustomize
8. Integration infra: Vault version, Vault Helm chart, cert-manager, Kind node image

---

## Part B: Secret Engine Gap Analysis

### Current Operator Secret Engine Coverage

| Vault Secret Engine | Operator CRD Coverage | Config | Role | Static Role | Status |
|--------------------|-----------------------|--------|------|-------------|--------|
| **KV (v1/v2)** | SecretEngineMount + RandomSecret + VaultSecret | Mount | N/A | N/A | **Covered** (mount + data via VaultSecret/RandomSecret) |
| **Database** | DatabaseSecretEngineConfig/Role/StaticRole | Yes | Yes | Yes | **Full** |
| **PKI** | PKISecretEngineConfig/Role | Yes | Yes | N/A | **Full** |
| **RabbitMQ** | RabbitMQSecretEngineConfig/Role | Yes | Yes | N/A | **Full** |
| **Azure** | AzureSecretEngineConfig/Role | Yes | Yes | N/A | **Full** |
| **GitHub** | GitHubSecretEngineConfig/Role | Yes | Yes | N/A | **Full** |
| **Kubernetes** | KubernetesSecretEngineConfig/Role | Yes | Yes | N/A | **Full** |
| **Quay** | QuaySecretEngineConfig/Role/StaticRole | Yes | Yes | Yes | **Full** |

### Missing Secret Engines

| Vault Secret Engine | Type | Demand | Complexity | Priority |
|---------------------|------|--------|------------|----------|
| **AWS** | Builtin | **Very High** — most popular cloud | High — IAM users, STS assumed roles, federation tokens | **P1** |
| **Transit** | Builtin | **Very High** — encryption as a service | Medium — config + key management, no roles per se | **P1** |
| **SSH** | Builtin | **High** — infrastructure SSH management | Medium — signed keys + OTP modes | **P1** |
| **Consul** | Builtin | **High** — HashiCorp ecosystem | Low — straightforward config + role | **P2** |
| **GCP (secrets)** | External plugin | **High** — operator already has GCP auth | Medium — service account keys + OAuth tokens | **P2** |
| **LDAP/AD (secrets)** | External plugin | **Medium** — enterprise environments | Medium — static creds, dynamic creds, check-out library | **P2** |
| **Nomad** | Builtin | **Medium** — HashiCorp ecosystem | Low — config + role | **P3** |
| **TOTP** | Builtin | **Low** — niche use case | Low — key generation | **P3** |
| **MongoDB Atlas** | External plugin | **Low** — niche | Low — config + role | **P3** |
| **Terraform Cloud** | External plugin | **Low** — niche | Low — config + role | **P3** |

### Enterprise-Only (Out of Scope for OSS Operator)

| Engine | Notes |
|--------|-------|
| Transform | Format-preserving encryption — Enterprise + ADP module |
| KMIP | Key Management Interoperability Protocol — Enterprise + ADP module |
| Key Management (keymgmt) | Cloud KMS key distribution — Enterprise + ADP module |

---

## Part C: Auth Engine Gap Analysis

### Current Operator Auth Engine Coverage

| Vault Auth Method | Operator CRD Coverage | Config | Role/Group | Status |
|-------------------|-----------------------|--------|------------|--------|
| **Any (mount)** | AuthEngineMount | Mount | N/A | **Partial** (mount-only) |
| **Kubernetes** | KubernetesAuthEngineConfig/Role | Yes | Yes | **Full** |
| **LDAP** | LDAPAuthEngineConfig/Group | Yes | Group | **Full** |
| **JWT/OIDC** | JWTOIDCAuthEngineConfig/Role | Yes | Yes | **Full** |
| **Azure** | AzureAuthEngineConfig/Role | Yes | Yes | **Full** |
| **GCP** | GCPAuthEngineConfig/Role | Yes | Yes | **Full** |
| **Cert (TLS)** | CertAuthEngineConfig/Role | Yes | Yes | **Full** |

### Missing Auth Methods

| Vault Auth Method | Type | Demand | Complexity | Priority |
|-------------------|------|--------|------------|----------|
| **AppRole** | Builtin | **Very High** — #1 machine-to-machine auth | Medium — config + role + secret-id management | **P1** |
| **AWS** | Builtin | **Very High** — #1 cloud auth | Medium — config + role (IAM + EC2 types) | **P1** |
| **Userpass** | Builtin | **Medium** — basic auth, dev/test use | Low — config + user management | **P2** |
| **GitHub** | Builtin | **Medium** — developer environments | Low — config + team/org mapping | **P2** |
| **Okta** | Builtin | **Medium** — enterprise SSO | Medium — config + group mapping | **P2** |
| **RADIUS** | Builtin | **Low** — niche enterprise | Low — config + user mapping | **P3** |
| **AliCloud** | External plugin | **Low** — regional cloud | Low — standard config + role | **P3** |
| **OCI** | External plugin | **Low** — regional cloud | Low — standard config + role | **P3** |
| **Kerberos** | External plugin | **Low** — legacy enterprise | Medium — complex SPNEGO setup | **P3** |
| **Cloud Foundry** | External plugin | **Low** — declining platform | Low — standard config + role | **P3** |

### Out of Scope

| Auth Method | Notes |
|-------------|-------|
| Token | Always enabled, no config/role CRDs needed — authentication handled by operator's own Vault client |
| SAML | Enterprise-only |

---

## Summary

| Area | Current | Gap | Action Items |
|------|---------|-----|-------------|
| **Dependencies** | Go 1.22, CR v0.17, K8s v0.29, Vault API v1.14, Operator SDK v1.31 | 6 minor versions behind on K8s stack, 9 on Vault API, 11 on Operator SDK | 3 epics: K8s stack, Vault+peripheral, Operator SDK |
| **Secret Engines** | 8 engines covered (KV, Database, PKI, RabbitMQ, Azure, GitHub, K8s, Quay) | 10 OSS engines missing (AWS, Transit, SSH, Consul, GCP, LDAP/AD, Nomad, TOTP, MongoDB Atlas, Terraform Cloud) | 3 epics by priority tier |
| **Auth Methods** | 6 methods covered + generic mount (K8s, LDAP, JWT/OIDC, Azure, GCP, Cert) | 9 methods missing (AppRole, AWS, Userpass, GitHub, Okta, RADIUS, AliCloud, OCI, Kerberos, CF) | 3 epics by priority tier |
