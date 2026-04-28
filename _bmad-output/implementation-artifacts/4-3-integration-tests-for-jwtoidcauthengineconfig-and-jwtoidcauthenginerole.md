# Story 4.3: Integration Tests for JWTOIDCAuthEngineConfig and JWTOIDCAuthEngineRole

Status: done

## Story

As an operator developer,
I want integration tests for the JWT/OIDC auth engine configuration and role types covering create, reconcile success, Vault state verification, and delete with cleanup,
So that JWT/OIDC authentication is verified end-to-end with a real Keycloak OIDC provider in Kind.

## Acceptance Criteria

1. **Given** Keycloak is deployed in the Kind cluster with a pre-configured realm and OIDC client **When** `make integration` is run **Then** the Keycloak server is available at `http://keycloak.keycloak.svc.cluster.local:8080` and the OIDC discovery endpoint at `/realms/test-realm/.well-known/openid-configuration` is reachable from Vault

2. **Given** a Kubernetes Secret with OIDC client credentials exists in `vault-admin` namespace **And** an AuthEngineMount (type=oidc) has been created and reconciled **And** a JWTOIDCAuthEngineConfig CR is created targeting the test OIDC mount with OIDC discovery URL pointing at Keycloak **When** the reconciler processes it **Then** the OIDC config is written to Vault at `auth/{path}/config` and ReconcileSuccessful=True

3. **Given** a JWTOIDCAuthEngineRole CR is created with `roleType: oidc`, `userClaim`, `allowedRedirectURIs`, and `tokenPolicies` **When** the reconciler processes it **Then** the role exists in Vault at `auth/{path}/role/{name}` with correct field values and ReconcileSuccessful=True

4. **Given** the JWTOIDCAuthEngineRole CR is deleted (IsDeletable=true) **When** the reconciler processes the deletion **Then** the role is removed from Vault and the CR is deleted from K8s

5. **Given** the JWTOIDCAuthEngineConfig CR is deleted (IsDeletable=false) **When** the reconciler processes the deletion **Then** the CR is deleted from K8s without Vault cleanup (no finalizer)

## Tasks / Subtasks

- [x] Task 1: Add Keycloak deployment to integration test infrastructure (AC: 1)
  - [x] 1.1: Create `integration/keycloak/deployment.yaml` with Keycloak Deployment + Service in dev mode with realm import
  - [x] 1.2: Create `integration/keycloak/configmap.yaml` with realm JSON (realm `test-realm`, client `vault-oidc`, secret `test-client-secret`)
  - [x] 1.3: Add `deploy-keycloak` target to Makefile that deploys Keycloak to `keycloak` namespace and waits for pod readiness
  - [x] 1.4: Add `deploy-keycloak` as a dependency of the `integration` target in Makefile (after `deploy-vault`, before test run)

- [x] Task 2: Add decoder methods (AC: 2, 3)
  - [x] 2.1: Add `GetJWTOIDCAuthEngineConfigInstance` method to `controllers/controllertestutils/decoder.go`
  - [x] 2.2: Add `GetJWTOIDCAuthEngineRoleInstance` method to `controllers/controllertestutils/decoder.go`

- [x] Task 3: Create test fixtures (AC: 2, 3, 4, 5)
  - [x] 3.1: Create `test/jwtoidcauthengine/test-jwtoidc-auth-mount.yaml` — AuthEngineMount with `type: oidc`, `path: test-jwt-oidc-auth`, `metadata.name: test-joaec-mount`
  - [x] 3.2: Create `test/jwtoidcauthengine/test-jwtoidc-auth-config.yaml` — JWTOIDCAuthEngineConfig with OIDC credentials from K8s Secret, `OIDCDiscoveryURL` pointing at Keycloak
  - [x] 3.3: Create `test/jwtoidcauthengine/test-jwtoidc-auth-role.yaml` — JWTOIDCAuthEngineRole with `roleType: oidc`, `userClaim: email`, `tokenPolicies: [default]`

- [x] Task 4: Create integration test file (AC: 2, 3, 4, 5)
  - [x] 4.1: Create `controllers/jwtoidcauthengine_controller_test.go` with `//go:build integration` tag
  - [x] 4.2: Add prerequisite context — create OIDC credentials K8s Secret, create AuthEngineMount (type=oidc), wait for reconcile
  - [x] 4.3: Add context for JWTOIDCAuthEngineConfig — create, poll for ReconcileSuccessful=True, verify Vault state at `auth/test-jwt-oidc-auth/test-joaec-mount/config`
  - [x] 4.4: Add context for JWTOIDCAuthEngineRole — create, poll for ReconcileSuccessful=True, verify Vault state at `auth/test-jwt-oidc-auth/test-joaec-mount/role/test-oidc-role`
  - [x] 4.5: Add deletion context — delete role (IsDeletable=true, verify Vault cleanup), delete config (IsDeletable=false), delete mount, delete secret

- [x] Task 5: End-to-end verification (AC: 1, 2, 3, 4, 5)
  - [x] 5.1: Run `make integration` and verify new tests pass alongside all existing tests
  - [x] 5.2: Verify no regressions — existing `kubernetes` auth and all prior tests unaffected

### Review Findings

- [x] [Review][Patch] Verify the JWT/OIDC config remains in Vault after deleting the non-deletable config CR [controllers/jwtoidcauthengine_controller_test.go:206]

## Dev Notes

### Infrastructure Scope — Keycloak Deployment in Kind (Tier 1: Install in Kind)

Per the project's three-tier integration test infrastructure rule and the Epic 3 retro decision, JWT/OIDC uses a **real OIDC provider** (Keycloak) deployed in Kind. This is the **highest infrastructure scope** story in Epic 4.

The Keycloak deployment must:

1. **Create `integration/keycloak/` directory** with:
   - `deployment.yaml` — Keycloak Deployment + Service
   - `configmap.yaml` — Realm import JSON as a ConfigMap

2. **Keycloak deployment details:**
   - Image: `quay.io/keycloak/keycloak:26.2` (verify latest stable at implementation time)
   - Namespace: `keycloak`
   - Mode: `start-dev --import-realm` (development mode: no HTTPS, auto-imports realm from `/opt/keycloak/data/import/`)
   - Service: `keycloak.keycloak.svc.cluster.local:8080` (HTTP, no TLS)
   - Resource limits: `memory: 512Mi`, `cpu: 500m` (Keycloak is Java-based, needs more than OpenLDAP)
   - Readiness probe: HTTP GET `/health/ready` on port 8080
   - Environment: `KC_HEALTH_ENABLED=true`, `KC_HTTP_ENABLED=true`, `KEYCLOAK_ADMIN=admin`, `KEYCLOAK_ADMIN_PASSWORD=admin`

3. **Realm import JSON (`test-realm`):**
   - Realm: `test-realm`, enabled=true
   - Client: `vault-oidc`, `clientAuthenticatorType: client-secret`, `secret: test-client-secret`, `protocol: openid-connect`, `standardFlowEnabled: true`, `redirectUris: ["http://localhost:8250/oidc/callback", "*"]`, `webOrigins: ["*"]`
   - No test users needed (we only verify config/role write to Vault, not the login flow)

4. **Add `deploy-keycloak` Makefile target:**
```makefile
.PHONY: deploy-keycloak
deploy-keycloak: kubectl
	$(KUBECTL) create namespace keycloak --dry-run=client -o yaml | $(KUBECTL) apply -f -
	$(KUBECTL) apply -f ./integration/keycloak -n keycloak
	$(KUBECTL) wait --for=condition=ready -n keycloak pod -l app=keycloak --timeout=$(KUBECTL_WAIT_TIMEOUT)
```

5. **Update the `integration` target** to include `deploy-keycloak`:
```makefile
integration: kind-setup deploy-vault deploy-ingress deploy-postgresql deploy-ldap deploy-keycloak vault manifests generate fmt vet envtest
```

**Startup timing:** Keycloak takes 30-60 seconds to start (Java JVM + realm import). The `$(KUBECTL_WAIT_TIMEOUT)` (120s) should be sufficient. The readiness probe ensures the pod is ready before the target completes.

**OIDC Discovery URL:** Once Keycloak is running, the well-known endpoint is: `http://keycloak.keycloak.svc.cluster.local:8080/realms/test-realm/.well-known/openid-configuration`

Vault can reach this URL from inside the Kind cluster via cluster DNS. No port-forwarding or ingress needed.

[Source: _bmad-output/implementation-artifacts/epic-3-retro-2026-04-20.md#L139-L146 — Keycloak decision]
[Source: _bmad-output/project-context.md#L134-L141 — Integration test infrastructure philosophy]

### Both Types Use VaultResource Reconciler — NOT VaultEngineResource

Both JWTOIDCAuthEngineConfig and JWTOIDCAuthEngineRole use `NewVaultResource` (the standard reconciler variant), same as Policy (Story 3.1), KubernetesAuthEngine types (Story 4.1), and LDAP types (Story 4.2). The reconcile flow is:

1. `prepareContext()` enriches context with kubeClient, restConfig, vaultConnection, vaultClient
2. `NewVaultResource(&r.ReconcilerBase, instance)` creates the standard reconciler
3. `VaultResource.Reconcile()` → `manageReconcileLogic()`:
   - `PrepareInternalValues()` — resolves OIDC credentials (config) or no-op (role)
   - `PrepareTLSConfig()` — no-op for both types
   - `VaultEndpoint.CreateOrUpdate()` — reads from Vault, calls `IsEquivalentToDesiredState()`, writes if different
4. `ManageOutcome()` sets `ReconcileSuccessful` condition

[Source: controllers/jwtoidcauthengineconfig_controller.go — uses NewVaultResource at line ~78-80]
[Source: controllers/jwtoidcauthenginerole_controller.go — uses NewVaultResource at line ~70-72]

### JWTOIDCAuthEngineConfig — Key Implementation Details

**GetPath():**
```go
func (r *JWTOIDCAuthEngineConfig) GetPath() string {
    return vaultutils.CleansePath("auth/" + string(r.Spec.Path) + "/config")
}
```
For fixture with `path: test-jwt-oidc-auth/test-joaec-mount` → `auth/test-jwt-oidc-auth/test-joaec-mount/config`

Note: Unlike KubernetesAuthEngineConfig (which includes `metadata.name` in the path), JWTOIDCAuthEngineConfig uses only `spec.path`. Same pattern as LDAPAuthEngineConfig.

[Source: api/v1alpha1/jwtoidcauthengineconfig_types.go#L194-L196]

**IsDeletable(): false** — No finalizer, no Vault cleanup on CR deletion. The config persists in Vault until the auth mount itself is deleted.

[Source: api/v1alpha1/jwtoidcauthengineconfig_types.go#L177-L179]

**toMap() — 14 Vault keys:**
```go
func (i *JWTOIDCConfig) toMap() map[string]interface{} {
    payload := map[string]interface{}{}
    payload["oidc_discovery_url"] = i.OIDCDiscoveryURL
    payload["oidc_discovery_ca_pem"] = i.OIDCDiscoveryCAPEM
    payload["oidc_client_id"] = i.retrievedClientID        // set by PrepareInternalValues
    payload["oidc_client_secret"] = i.retrievedClientPassword // set by PrepareInternalValues
    payload["oidc_response_mode"] = i.OIDCResponseMode
    payload["oidc_response_types"] = i.OIDCResponseTypes
    payload["jwks_url"] = i.JWKSURL
    payload["jwks_ca_pem"] = i.JWKSCAPEM
    payload["jwt_validation_pubkeys"] = i.JWTValidationPubKeys
    payload["bound_issuer"] = i.BoundIssuer
    payload["jwt_supported_algs"] = i.JWTSupportedAlgs
    payload["default_role"] = i.DefaultRole
    payload["provider_config"] = i.ProviderConfig
    payload["namespace_in_state"] = i.NamespaceInState
    return payload
}
```

[Source: api/v1alpha1/jwtoidcauthengineconfig_types.go#L303-L320]

**IsEquivalentToDesiredState():** Bare `reflect.DeepEqual(desiredState, payload)` — NO custom logic, no key filtering. Vault's GET response for `auth/{path}/config` does NOT return `oidc_client_secret` (sensitive) and may include extra fields. This means `IsEquivalentToDesiredState` returns false on every reconcile, causing a config write each cycle. Same Story 7-4 tech debt as other types — does NOT affect test correctness.

[Source: api/v1alpha1/jwtoidcauthengineconfig_types.go#L202-L205]

**PrepareInternalValues():**
```go
func (r *JWTOIDCAuthEngineConfig) PrepareInternalValues(context context.Context, object client.Object) error {
    if reflect.DeepEqual(r.Spec.OIDCCredentials, &vaultutils.RootCredentialConfig{PasswordKey: "password", UsernameKey: "username"}) {
        return nil
    }
    return r.setInternalCredentials(context)
}
```
If `OIDCCredentials` equals the empty default (`{PasswordKey: "password", UsernameKey: "username"}`), returns nil immediately (no credential fetch). Otherwise calls `setInternalCredentials` which resolves OIDC client ID and secret from one of three sources (K8s Secret, RandomSecret, VaultSecret).

**For the K8s Secret path:**
- If `OIDCClientID` is set in spec → `retrievedClientID = spec.OIDCClientID`, `retrievedClientPassword = secret.Data[passwordKey]`
- If `OIDCClientID` is empty → both ID and password come from the secret

[Source: api/v1alpha1/jwtoidcauthengineconfig_types.go#L219-L297]

**PrepareTLSConfig():** Returns `nil` (no-op). No TLS configuration support.

**Webhook:**
- `Default()`: Log-only, no field defaults
- `ValidateCreate()`: No-op (returns nil)
- `ValidateUpdate()`: Checks immutable `spec.path` — rejects changes with `errors.New("spec.path cannot be updated")`
- `ValidateDelete()`: No-op

[Source: api/v1alpha1/jwtoidcauthengineconfig_webhook.go]

### JWTOIDCAuthEngineRole — Key Implementation Details

**GetPath():**
```go
func (r *JWTOIDCAuthEngineRole) GetPath() string {
    return vaultutils.CleansePath("auth/" + string(r.Spec.Path) + "/role/" + r.getName())
}
func (r *JWTOIDCAuthEngineRole) getName() string {
    if r.Spec.Name != "" {
        return string(r.Spec.Name)
    }
    return r.Name
}
```
For fixture with `path: test-jwt-oidc-auth/test-joaec-mount`, `name: test-oidc-role` → `auth/test-jwt-oidc-auth/test-joaec-mount/role/test-oidc-role`

Note: Uses `spec.name` (not `metadata.name`) for the Vault role name, same pattern as LDAPAuthEngineGroup.

[Source: api/v1alpha1/jwtoidcauthenginerole_types.go#L268-L270, L329-L335]

**IsDeletable(): true** — Finalizer added after first successful reconcile. On deletion, the role is removed from Vault at `auth/{path}/role/{name}`.

[Source: api/v1alpha1/jwtoidcauthenginerole_types.go#L260-L262]

**toMap() — ~23 Vault keys:**
```go
func (i *JWTOIDCRole) toMap() map[string]interface{} {
    payload := map[string]interface{}{}
    payload["name"] = i.Name
    payload["role_type"] = i.RoleType
    payload["bound_audiences"] = i.BoundAudiences
    payload["user_claim"] = i.UserClaim
    payload["user_claim_json_pointer"] = i.UserClaimJSONPointer
    payload["clock_skew_leeway"] = i.ClockSkewLeeway
    payload["expiration_leeway"] = i.ExpirationLeeway
    payload["not_before_leeway"] = i.NotBeforeLeeway
    payload["bound_subject"] = i.BoundSubject
    payload["bound_claims"] = i.BoundClaims
    payload["bound_claims_type"] = i.BoundClaimsType
    payload["groups_claim"] = i.GroupsClaim
    payload["claim_mappings"] = i.ClaimMappings
    payload["oidc_scopes"] = i.OIDCScopes
    payload["allowed_redirect_uris"] = i.AllowedRedirectURIs
    payload["verbose_oidc_logging"] = i.VerboseOIDCLogging
    payload["max_age"] = i.MaxAge
    payload["token_ttl"] = i.TokenTTL
    payload["token_max_ttl"] = i.TokenMaxTTL
    payload["token_policies"] = i.TokenPolicies
    payload["token_bound_cidrs"] = i.TokenBoundCIDRs
    payload["token_explicit_max_ttl"] = i.TokenExplicitMaxTTL
    payload["token_no_default_policy"] = i.TokenNoDefaultPolicy
    payload["token_num_uses"] = i.TokenNumUses
    payload["token_period"] = i.TokenPeriod
    payload["token_type"] = i.TokenType
    return payload
}
```

[Source: api/v1alpha1/jwtoidcauthenginerole_types.go#L297-L325]

**IsEquivalentToDesiredState():** Bare `reflect.DeepEqual(desiredState, payload)` — no key filtering. Vault's read response for roles may include extra keys. Same Story 7-4 tech debt.

[Source: api/v1alpha1/jwtoidcauthenginerole_types.go#L276-L279]

**PrepareInternalValues():** Returns `nil` (no-op). No credential resolution needed for roles.

**PrepareTLSConfig():** Returns `nil` (no-op).

**Webhook — NOTABLE: No immutable path check on update!**
- `Default()`: Log-only, no defaults
- `ValidateCreate()`: No-op (returns nil)
- `ValidateUpdate()`: Returns `(nil, nil)` — **does NOT check immutable `spec.path`**. This is different from most types (KubernetesAuthEngineConfig, LDAPAuthEngineConfig both enforce immutable path). This is existing behavior, NOT something we need to fix in this story.
- `ValidateDelete()`: No-op

[Source: api/v1alpha1/jwtoidcauthenginerole_webhook.go]

### JWTOIDCAuthEngineConfig Controller — Watches for Credential Changes

The JWTOIDCAuthEngineConfig controller has extra `Watches` beyond the basic `For()`:
1. **K8s Secrets** — re-queues config CRs when their referenced `OIDCCredentials.Secret` changes (username or password data change)
2. **RandomSecret** — re-queues config CRs when their referenced RandomSecret's `LastVaultSecretUpdate` changes

The JWTOIDCAuthEngineRole controller has **no** extra watches — just the basic `For()` + periodic reconcile predicate.

These watches don't directly affect the integration test but the dev should be aware that the config controller's `SetupWithManager` is more complex than the role controller's.

[Source: controllers/jwtoidcauthengineconfig_controller.go#L83-L190]
[Source: controllers/jwtoidcauthenginerole_controller.go#L76-L79]

### Vault API Response Shapes

**GET `auth/{path}/config`** — Returns JWT/OIDC auth config:
```json
{
  "data": {
    "oidc_discovery_url": "http://keycloak.keycloak.svc.cluster.local:8080/realms/test-realm",
    "oidc_discovery_ca_pem": "",
    "oidc_client_id": "vault-oidc",
    "oidc_response_mode": "",
    "oidc_response_types": [],
    "jwks_url": "",
    "jwks_ca_pem": "",
    "jwt_validation_pubkeys": [],
    "bound_issuer": "",
    "jwt_supported_algs": [],
    "default_role": "",
    "provider_config": {},
    "namespace_in_state": true
  }
}
```
Note: `oidc_client_secret` is NOT returned (sensitive). Extra fields like `jwt_auto_remove_keys` may appear depending on Vault version.

**GET `auth/{path}/role/{name}`** — Returns role config:
```json
{
  "data": {
    "name": "test-oidc-role",
    "role_type": "oidc",
    "bound_audiences": null,
    "user_claim": "email",
    "user_claim_json_pointer": false,
    "clock_skew_leeway": 0,
    "expiration_leeway": 0,
    "not_before_leeway": 0,
    "bound_subject": "",
    "bound_claims": null,
    "bound_claims_type": "string",
    "groups_claim": "",
    "claim_mappings": null,
    "oidc_scopes": ["openid", "email"],
    "allowed_redirect_uris": ["http://localhost:8250/oidc/callback"],
    "verbose_oidc_logging": false,
    "max_age": 0,
    "token_ttl": 0,
    "token_max_ttl": 0,
    "token_policies": ["default"],
    "token_bound_cidrs": [],
    "token_explicit_max_ttl": 0,
    "token_no_default_policy": false,
    "token_num_uses": 0,
    "token_period": 0,
    "token_type": "default",
    ...
  }
}
```
The role response may include extra keys like `num_uses`, `period`, `ttl`, `max_ttl` (legacy aliases).

### Test Design — Dependency Chain

```
K8s Secret (test-oidc-creds) — must exist before config
  └── AuthEngineMount (test-jwt-oidc-auth/test-joaec-mount, type=oidc)
        └── JWTOIDCAuthEngineConfig → auth/test-jwt-oidc-auth/test-joaec-mount/config
        └── JWTOIDCAuthEngineRole → auth/test-jwt-oidc-auth/test-joaec-mount/role/test-oidc-role
```

Resources must be created in order: Secret → Mount → Config → Role. Deletion in reverse: Role → Config → Mount → Secret.

The AuthEngineMount must be reconciled before the config or role, because Vault rejects writes to `auth/{path}/config` if the mount doesn't exist.

The JWTOIDCAuthEngineRole does NOT depend on config being written — Vault allows creating roles before the config is set. But for test realism, create config before role. More importantly, if the role has `role_type: oidc`, Vault may validate that the config has OIDC discovery configured.

### OIDC Credentials K8s Secret — Created in Test, Not as Fixture

The OIDC credentials Secret should be created programmatically in the test's first `Context` block:

```go
oidcSecret := &corev1.Secret{
    ObjectMeta: metav1.ObjectMeta{
        Name:      "test-oidc-creds",
        Namespace: vaultAdminNamespaceName,
    },
    Data: map[string][]byte{
        "oidc_client_id":     []byte("vault-oidc"),
        "oidc_client_secret": []byte("test-client-secret"),
    },
}
```

The fixture sets `OIDCClientID: "vault-oidc"` in the spec. The `setInternalCredentials` logic uses the spec value for client ID and reads only the password from the secret. Providing both keys in the secret is good practice.

### Test Fixture Design

**Fixture 1: `test/jwtoidcauthengine/test-jwtoidc-auth-mount.yaml`** — AuthEngineMount prerequisite:
```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: AuthEngineMount
metadata:
  name: test-joaec-mount
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  type: oidc
  path: test-jwt-oidc-auth
```
Mounts at `sys/auth/test-jwt-oidc-auth/test-joaec-mount`. Uses `type: oidc` to enable the OIDC auth method. Note: Vault's `oidc` type enables the combined JWT/OIDC auth method — there is no separate `jwt` mount type.

**Fixture 2: `test/jwtoidcauthengine/test-jwtoidc-auth-config.yaml`** — JWTOIDCAuthEngineConfig:
```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: JWTOIDCAuthEngineConfig
metadata:
  name: test-joaec-config
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  path: test-jwt-oidc-auth/test-joaec-mount
  OIDCDiscoveryURL: "http://keycloak.keycloak.svc.cluster.local:8080/realms/test-realm"
  OIDCClientID: "vault-oidc"
  OIDCCredentials:
    secret:
      name: test-oidc-creds
    usernameKey: oidc_client_id
    passwordKey: oidc_client_secret
```
`GetPath()` returns `auth/test-jwt-oidc-auth/test-joaec-mount/config`. The config points at the Keycloak OIDC discovery endpoint in the cluster.

Key: `OIDCCredentials` is set with custom key names (`oidc_client_id`/`oidc_client_secret`), NOT the defaults (`username`/`password`). This ensures `PrepareInternalValues` does NOT skip credential resolution (the default check compares against `{PasswordKey: "password", UsernameKey: "username"}`).

**Fixture 3: `test/jwtoidcauthengine/test-jwtoidc-auth-role.yaml`** — JWTOIDCAuthEngineRole:
```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: JWTOIDCAuthEngineRole
metadata:
  name: test-joaer-role
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  path: test-jwt-oidc-auth/test-joaec-mount
  name: "test-oidc-role"
  roleType: "oidc"
  userClaim: "email"
  allowedRedirectURIs:
    - "http://localhost:8250/oidc/callback"
  tokenPolicies:
    - "default"
  OIDCScopes:
    - "openid"
    - "email"
```
`GetPath()` returns `auth/test-jwt-oidc-auth/test-joaec-mount/role/test-oidc-role` (uses `spec.name`, not `metadata.name`).

### Verifying Vault State

**Config verification:**
```go
secret, err := vaultClient.Logical().Read("auth/test-jwt-oidc-auth/test-joaec-mount/config")
Expect(err).To(BeNil())
Expect(secret).NotTo(BeNil())
Expect(secret.Data["oidc_discovery_url"]).To(Equal("http://keycloak.keycloak.svc.cluster.local:8080/realms/test-realm"))
Expect(secret.Data["oidc_client_id"]).To(Equal("vault-oidc"))
```
Note: `oidc_client_secret` should NOT be in the response. Do NOT try to verify it.

**Role verification:**
```go
secret, err := vaultClient.Logical().Read("auth/test-jwt-oidc-auth/test-joaec-mount/role/test-oidc-role")
Expect(err).To(BeNil())
Expect(secret).NotTo(BeNil())

roleType, ok := secret.Data["role_type"].(string)
Expect(ok).To(BeTrue(), "expected role_type to be a string")
Expect(roleType).To(Equal("oidc"))

userClaim, ok := secret.Data["user_claim"].(string)
Expect(ok).To(BeTrue(), "expected user_claim to be a string")
Expect(userClaim).To(Equal("email"))

tokenPolicies, ok := secret.Data["token_policies"].([]interface{})
Expect(ok).To(BeTrue(), "expected token_policies to be []interface{}")
Expect(tokenPolicies).To(ContainElement("default"))

oidcScopes, ok := secret.Data["oidc_scopes"].([]interface{})
Expect(ok).To(BeTrue(), "expected oidc_scopes to be []interface{}")
Expect(oidcScopes).To(ContainElement("openid"))
Expect(oidcScopes).To(ContainElement("email"))

allowedRedirects, ok := secret.Data["allowed_redirect_uris"].([]interface{})
Expect(ok).To(BeTrue(), "expected allowed_redirect_uris to be []interface{}")
Expect(allowedRedirects).To(ContainElement("http://localhost:8250/oidc/callback"))
```

**Delete verification (role — IsDeletable=true):**
```go
Eventually(func() bool {
    err := k8sIntegrationClient.Get(ctx, types.NamespacedName{...}, &redhatcopv1alpha1.JWTOIDCAuthEngineRole{})
    return apierrors.IsNotFound(err)
}, timeout, interval).Should(BeTrue())

Eventually(func() bool {
    secret, err := vaultClient.Logical().Read("auth/test-jwt-oidc-auth/test-joaec-mount/role/test-oidc-role")
    return err == nil && secret == nil
}, timeout, interval).Should(BeTrue())
```

**Delete verification (config — IsDeletable=false):**
Config deletion from K8s happens immediately (no finalizer). No Vault cleanup expected.
```go
Eventually(func() bool {
    err := k8sIntegrationClient.Get(ctx, types.NamespacedName{...}, &redhatcopv1alpha1.JWTOIDCAuthEngineConfig{})
    return apierrors.IsNotFound(err)
}, timeout, interval).Should(BeTrue())
```

### Test Structure

```
Describe("JWTOIDCAuthEngine controllers", Ordered)
  var oidcSecret *corev1.Secret
  var mountInstance *redhatcopv1alpha1.AuthEngineMount
  var configInstance *redhatcopv1alpha1.JWTOIDCAuthEngineConfig
  var roleInstance *redhatcopv1alpha1.JWTOIDCAuthEngineRole

  AfterAll: best-effort delete all instances + oidc secret

  Context("When creating prerequisite resources")
    It("Should create the OIDC credentials secret and JWT/OIDC auth mount")
      - Create test-oidc-creds K8s Secret in vault-admin namespace
      - Load test-jwtoidc-auth-mount.yaml via decoder.GetAuthEngineMountInstance
      - Set namespace to vaultAdminNamespaceName, create
      - Eventually poll for ReconcileSuccessful=True
      - Verify mount exists in sys/auth with key "test-jwt-oidc-auth/test-joaec-mount/"

  Context("When creating a JWTOIDCAuthEngineConfig")
    It("Should write the OIDC config to Vault")
      - Load test-jwtoidc-auth-config.yaml via decoder.GetJWTOIDCAuthEngineConfigInstance
      - Set namespace to vaultAdminNamespaceName, create
      - Eventually poll for ReconcileSuccessful=True
      - Read auth/test-jwt-oidc-auth/test-joaec-mount/config from Vault
      - Verify oidc_discovery_url = Keycloak URL
      - Verify oidc_client_id = "vault-oidc"

  Context("When creating a JWTOIDCAuthEngineRole")
    It("Should create the role in Vault with correct OIDC settings")
      - Load test-jwtoidc-auth-role.yaml via decoder.GetJWTOIDCAuthEngineRoleInstance
      - Set namespace to vaultAdminNamespaceName, create
      - Eventually poll for ReconcileSuccessful=True
      - Read auth/test-jwt-oidc-auth/test-joaec-mount/role/test-oidc-role
      - Verify role_type = "oidc"
      - Verify user_claim = "email"
      - Verify token_policies contains "default"
      - Verify oidc_scopes contains "openid" and "email"
      - Verify allowed_redirect_uris contains callback URL

  Context("When deleting JWTOIDCAuthEngine resources")
    It("Should clean up role from Vault and remove all resources")
      - Delete role CR (IsDeletable=true → Vault cleanup)
      - Eventually verify K8s deletion (NotFound)
      - Eventually verify role removed from Vault (Read returns nil)
      - Delete config CR (IsDeletable=false → no Vault cleanup)
      - Eventually verify K8s deletion
      - Delete AuthEngineMount
      - Eventually verify K8s deletion and mount gone from sys/auth
      - Delete OIDC credentials secret
```

### Import Requirements

```go
import (
    "time"

    . "github.com/onsi/ginkgo/v2"
    . "github.com/onsi/gomega"
    redhatcopv1alpha1 "github.com/redhat-cop/vault-config-operator/api/v1alpha1"
    "github.com/redhat-cop/vault-config-operator/controllers/vaultresourcecontroller"

    corev1 "k8s.io/api/core/v1"
    apierrors "k8s.io/apimachinery/pkg/api/errors"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/apimachinery/pkg/types"
)
```

### Name Collision Prevention

Fixture names use `test-joaec-` (JWT/OIDC Auth Engine Config) and `test-joaer-` (JWT/OIDC Auth Engine Role) prefixes:
- `test-jwt-oidc-auth/test-joaec-mount` — auth mount (unique path prefix)
- `test-joaec-config` — JWT/OIDC config CR name
- `test-joaer-role` — JWT/OIDC role CR name
- `test-oidc-role` — Vault role name in `spec.name`
- `test-oidc-creds` — OIDC credentials K8s Secret

These don't collide with:
- `oidc/azuread-oidc` — existing sample fixtures (not created in integration tests)
- `test-k8s-auth/test-kaec-mount` — Story 4.1 Kubernetes auth tests
- `test-ldap-auth/test-laec-mount` — Story 4.2 LDAP auth tests
- `test-auth-mount/test-aem-*` — Story 3.4 AuthEngineMount tests
- `kubernetes/` — default Kubernetes auth mount

### Controller Registration — Already Done

Both controllers are registered in `suite_integration_test.go`:
```go
err = (&JWTOIDCAuthEngineConfigReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "JWTOIDCAuthEngineConfig")}).SetupWithManager(mgr)
Expect(err).ToNot(HaveOccurred())

err = (&JWTOIDCAuthEngineRoleReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "JWTOIDCAuthEngineRole")}).SetupWithManager(mgr)
Expect(err).ToNot(HaveOccurred())
```

No changes needed to the test suite setup.

[Source: controllers/suite_integration_test.go#L160-L164]

### Decoder Methods — Both Must Be Added

Neither `GetJWTOIDCAuthEngineConfigInstance` nor `GetJWTOIDCAuthEngineRoleInstance` exist in the decoder. Both must be added following the established pattern:

```go
func (d *decoder) GetJWTOIDCAuthEngineConfigInstance(filename string) (*redhatcopv1alpha1.JWTOIDCAuthEngineConfig, error) {
    obj, groupKindVersion, err := d.decodeFile(filename)
    if err != nil {
        return nil, err
    }
    kind := reflect.TypeOf(redhatcopv1alpha1.JWTOIDCAuthEngineConfig{}).Name()
    if groupKindVersion.Kind == kind {
        o := obj.(*redhatcopv1alpha1.JWTOIDCAuthEngineConfig)
        return o, nil
    }
    return nil, errDecode
}

func (d *decoder) GetJWTOIDCAuthEngineRoleInstance(filename string) (*redhatcopv1alpha1.JWTOIDCAuthEngineRole, error) {
    obj, groupKindVersion, err := d.decodeFile(filename)
    if err != nil {
        return nil, err
    }
    kind := reflect.TypeOf(redhatcopv1alpha1.JWTOIDCAuthEngineRole{}).Name()
    if groupKindVersion.Kind == kind {
        o := obj.(*redhatcopv1alpha1.JWTOIDCAuthEngineRole)
        return o, nil
    }
    return nil, errDecode
}
```

[Source: controllers/controllertestutils/decoder.go — existing pattern at lines 85-98]

### Keycloak Realm Import JSON Design

The ConfigMap should contain a Keycloak realm export JSON with minimal configuration:

```json
{
  "realm": "test-realm",
  "enabled": true,
  "sslRequired": "none",
  "clients": [
    {
      "clientId": "vault-oidc",
      "enabled": true,
      "clientAuthenticatorType": "client-secret",
      "secret": "test-client-secret",
      "protocol": "openid-connect",
      "standardFlowEnabled": true,
      "directAccessGrantsEnabled": true,
      "redirectUris": ["http://localhost:8250/oidc/callback", "*"],
      "webOrigins": ["*"],
      "publicClient": false
    }
  ]
}
```

Key fields:
- `sslRequired: "none"` — required for HTTP (non-TLS) mode in dev
- `publicClient: false` — makes the client "confidential" (requires client secret)
- `directAccessGrantsEnabled: true` — enables Resource Owner Password Credentials grant (useful for testing login)
- `redirectUris: ["*"]` — permissive for testing (Vault needs redirect URI match)

### Vault OIDC Discovery Validation

When writing the config to Vault with `oidc_discovery_url` set, Vault will attempt to fetch the well-known OpenID configuration from:
`http://keycloak.keycloak.svc.cluster.local:8080/realms/test-realm/.well-known/openid-configuration`

If Keycloak is not reachable, the Vault write will fail with an error like:
```
error fetching certificates from discovery endpoint: ...
```

This is why Keycloak must be fully started and the realm imported BEFORE the integration test creates the JWTOIDCAuthEngineConfig CR. The `deploy-keycloak` Makefile target with its readiness probe ensures this.

### Risk Considerations

- **Keycloak startup time:** Keycloak is Java-based and takes 30-60 seconds to start. The `$(KUBECTL_WAIT_TIMEOUT)` of 120s should be sufficient. If CI is resource-constrained, this may take longer. Consider bumping to 180s for the Keycloak-specific wait if needed.

- **Keycloak memory requirements:** Keycloak needs ~512Mi RAM minimum. Kind clusters in CI may have limited resources. Set `resources.requests.memory: 512Mi` and `resources.limits.memory: 768Mi` in the deployment.

- **Vault-to-Keycloak connectivity:** Vault must reach Keycloak via in-cluster DNS (`keycloak.keycloak.svc.cluster.local`). Both are in the Kind cluster so this should work. No port-forwarding or ingress needed.

- **Config `IsEquivalentToDesiredState` always returns false:** Vault's read response omits `oidc_client_secret` (sensitive) and may include extra fields. The reconciler writes on every cycle. Known tech debt (Story 7-4), does NOT block `ReconcileSuccessful=True`.

- **Role `ValidateUpdate` has NO immutable path check:** Unlike most types, `JWTOIDCAuthEngineRole` webhook does NOT enforce immutable `spec.path` on update. This is existing behavior and not a test concern.

- **OIDCCredentials default detection:** `PrepareInternalValues` skips credential resolution if `OIDCCredentials` matches the exact default `{PasswordKey: "password", UsernameKey: "username"}`. The test fixture uses custom key names (`oidc_client_id`/`oidc_client_secret`) to ensure credential resolution runs.

- **Existing test fixtures at `test/jwtoidcauthengine/`:** The directory already contains sample fixtures (`jwtoidc-auth-engine-config.yaml`, etc.) that reference Azure AD. These are NOT used in integration tests — they are reference examples. New test fixtures use `test-` prefix to distinguish.

- **Checked type assertions:** Per Epic 3 retro action item, always use two-value form `val, ok := x.(string)` with `Expect(ok).To(BeTrue())` for all Vault response field assertions.

### File Inventory — What Needs to Change

| # | File | Change Type | Description |
|---|------|-------------|-------------|
| 1 | `integration/keycloak/deployment.yaml` | New | Keycloak Deployment + Service manifest |
| 2 | `integration/keycloak/configmap.yaml` | New | Keycloak realm import JSON as ConfigMap |
| 3 | `Makefile` | Modified | Add `deploy-keycloak` target; add `deploy-keycloak` to `integration` dependencies |
| 4 | `controllers/controllertestutils/decoder.go` | Modified | Add `GetJWTOIDCAuthEngineConfigInstance` and `GetJWTOIDCAuthEngineRoleInstance` |
| 5 | `test/jwtoidcauthengine/test-jwtoidc-auth-mount.yaml` | New | AuthEngineMount prerequisite (type=oidc) |
| 6 | `test/jwtoidcauthengine/test-jwtoidc-auth-config.yaml` | New | JWTOIDCAuthEngineConfig with OIDC credentials from K8s Secret |
| 7 | `test/jwtoidcauthengine/test-jwtoidc-auth-role.yaml` | New | JWTOIDCAuthEngineRole with roleType=oidc |
| 8 | `controllers/jwtoidcauthengine_controller_test.go` | New | Integration test — create mount, config, role; verify Vault state; delete and verify cleanup |

No changes to suite setup, controllers, webhooks, or types.

### No `make manifests generate` Needed

This story only adds an integration test file, YAML fixtures, decoder methods, Makefile changes, and Keycloak infrastructure manifests. No CRD types, controllers, or webhooks are changed.

### Previous Story Intelligence

**From Story 4.2 (LDAPAuthEngine integration tests):**
- Established the "install in Kind" infrastructure pattern for Epic 4: new Makefile target + K8s manifests in `integration/` + wire into `integration` target
- Demonstrated bind credentials K8s Secret created programmatically in the test
- Same `GetPath()` pattern — config uses `spec.path` only (no `metadata.name`)
- `IsDeletable` split: config=false (no cleanup), group/role=true (Vault cleanup)
- Deploy-ldap target pattern: `create namespace --dry-run=client | apply`, `apply -f ./integration/ldap`, `wait --for=condition=ready`

**From Story 4.1 (KubernetesAuthEngine integration tests):**
- Established the Epic 4 auth engine test pattern: prerequisite AuthEngineMount → config → role, with isolated test mount paths
- Demonstrated decoder method addition pattern
- `AfterAll` cleanup guard pattern
- Checked type assertions for Vault response fields

**From Story 3.4 (AuthEngineMount integration tests):**
- AuthEngineMount test pattern: verify `sys/auth` response after create
- `AfterAll` cleanup for best-effort delete

**From Story 3.1 (Policy integration tests):**
- VaultResource test pattern: create → poll ReconcileSuccessful → verify Vault state → delete → verify cleanup
- Both JWT/OIDC types use the same VaultResource reconciler

**From Epic 3 Retrospective:**
- "Story ordering: 4.1 (simplest) → 4.2 (LDAP infra) → 4.3 (Keycloak infra)"
- "Checked type assertions rule" — always use two-value form
- Story 4.3 classified as "High infra scope — OIDC provider / Install in Kind"
- "Deploy Keycloak to Kind using the Keycloak Operator" — refined here to use direct deployment (simpler than operator)

### Git Intelligence (Recent Commits)

```
9608211 Merge pull request #318 from raffaelespazzoli/bmad-epic-3
24a37f0 Complete Epic 3 retrospective and close Epics 1-3
cb473c3 Mark Story 3.4 as done after clean code review
866c843 Add integration tests for AuthEngineMount type (Story 3.4)
db21d90 Add integration tests for SecretEngineMount type (Story 3.3)
```

Codebase is clean post-Epic 3 merge to main. Stories 4.1 and 4.2 have story specs but are not yet implemented.

### Integration Test Infrastructure Classification

Per the project's three-tier rule:
- **OIDC provider:** CAN be installed in Kind → **Tier 1: Install in Kind** using Keycloak
- **Vault API:** Already available in Kind
- **K8s Secrets:** Available via envtest client

**Classification: Install in Kind — High infrastructure scope** (new Makefile target + Keycloak deployment + realm configuration)

[Source: _bmad-output/project-context.md#L134-L141 — Integration test infrastructure philosophy]
[Source: _bmad-output/implementation-artifacts/epic-3-retro-2026-04-20.md#L139-L146 — Keycloak decision]

### Project Structure Notes

- Decoder changes in `controllers/controllertestutils/decoder.go` (add two methods)
- Test file goes in `controllers/jwtoidcauthengine_controller_test.go`
- Test fixtures go in `test/jwtoidcauthengine/` directory (alongside existing sample fixtures, with `test-` prefix)
- Keycloak manifests go in `integration/keycloak/` directory (follows `integration/ldap/` pattern)
- Makefile changes: new `deploy-keycloak` target, updated `integration` dependency list
- All files follow existing naming conventions

### References

- [Source: api/v1alpha1/jwtoidcauthengineconfig_types.go] — JWTOIDCAuthEngineConfig VaultObject implementation, GetPath (auth/{path}/config), GetPayload, IsEquivalentToDesiredState (bare DeepEqual), toMap (14 keys), PrepareInternalValues (OIDCCredentials resolution), IsDeletable=false
- [Source: api/v1alpha1/jwtoidcauthengineconfig_types.go#L194-L196] — GetPath: auth/{spec.path}/config
- [Source: api/v1alpha1/jwtoidcauthengineconfig_types.go#L202-L205] — IsEquivalentToDesiredState: bare reflect.DeepEqual
- [Source: api/v1alpha1/jwtoidcauthengineconfig_types.go#L219-L297] — PrepareInternalValues + setInternalCredentials (3 credential sources)
- [Source: api/v1alpha1/jwtoidcauthengineconfig_types.go#L303-L320] — JWTOIDCConfig.toMap (14 keys)
- [Source: api/v1alpha1/jwtoidcauthengineconfig_webhook.go] — Webhook: immutable path on update, no create validation
- [Source: api/v1alpha1/jwtoidcauthenginerole_types.go] — JWTOIDCAuthEngineRole VaultObject implementation, GetPath (auth/{path}/role/{name}), toMap (~23 keys), IsDeletable=true
- [Source: api/v1alpha1/jwtoidcauthenginerole_types.go#L268-L270] — GetPath: auth/{path}/role/{getName()}
- [Source: api/v1alpha1/jwtoidcauthenginerole_types.go#L297-L325] — JWTOIDCRole.toMap (~23 keys)
- [Source: api/v1alpha1/jwtoidcauthenginerole_types.go#L329-L335] — getName(): spec.Name if set, else metadata.Name
- [Source: api/v1alpha1/jwtoidcauthenginerole_webhook.go] — Webhook: NO immutable path check on update!
- [Source: controllers/jwtoidcauthengineconfig_controller.go] — Controller (VaultResource + Secret/RandomSecret watches)
- [Source: controllers/jwtoidcauthenginerole_controller.go] — Controller (VaultResource, simple, no extra watches)
- [Source: controllers/suite_integration_test.go#L160-L164] — Both controllers registered
- [Source: controllers/controllertestutils/decoder.go] — Existing decoder methods; GetJWTOIDCAuthEngineConfigInstance and GetJWTOIDCAuthEngineRoleInstance MUST BE ADDED
- [Source: test/jwtoidcauthengine/] — Existing sample fixtures (reference only; test fixtures use test- prefix)
- [Source: Makefile#L135] — integration target dependencies (must add deploy-keycloak)
- [Source: controllers/policy_controller_test.go] — VaultResource test pattern (Story 3.1)
- [Source: controllers/authenginemount_controller_test.go] — AuthEngineMount test pattern (Story 3.4)
- [Source: _bmad-output/implementation-artifacts/4-1-integration-tests-for-kubernetesauthengineconfig-and-kubernetesauthenginerole.md] — Story 4.1 spec (closest Epic 4 pattern)
- [Source: _bmad-output/implementation-artifacts/4-2-integration-tests-for-ldapauthengineconfig-and-ldapauthenginegroup.md] — Story 4.2 spec (infrastructure pattern reference)
- [Source: _bmad-output/implementation-artifacts/epic-3-retro-2026-04-20.md#L139-L146] — Keycloak infrastructure decision
- [Source: _bmad-output/project-context.md#L134-L141] — Integration test infrastructure philosophy
- [Source: _bmad-output/project-context.md#L143-L155] — Integration test pattern and Ordered lifecycle
- [Source: _bmad-output/planning-artifacts/epics.md#L477-L496] — Story 4.3 epic definition

## Dev Agent Record

### Agent Model Used

Claude Opus 4 (via Cursor)

### Debug Log References

- Keycloak 26.2 requires realm import files to be named `{realm-name}-realm.json` (not `{realm-name}.json`) — fixed ConfigMap filename from `test-realm.json` to `test-realm-realm.json`
- Keycloak 26.2 serves health endpoints on management port 9000, not the application port 8080 — fixed readiness probe port from 8080 to 9000
- Initial `make integration` attempts hit wrong kubeconfig context (OpenShift instead of Kind) — resolved by switching to `kind-kind` context

### Completion Notes List

- Deployed Keycloak 26.2 in Kind cluster as real OIDC provider (Tier 1: Install in Kind)
- Keycloak realm `test-realm` with confidential client `vault-oidc` auto-imported on startup
- JWTOIDCAuthEngineConfig integration test verifies: CR creation → ReconcileSuccessful=True → Vault state at `auth/test-jwt-oidc-auth/test-joaec-mount/config` with correct `oidc_discovery_url` and `oidc_client_id`
- JWTOIDCAuthEngineRole integration test verifies: CR creation → ReconcileSuccessful=True → Vault state at `auth/test-jwt-oidc-auth/test-joaec-mount/role/test-oidc-role` with correct `role_type`, `user_claim`, `token_policies`, `oidc_scopes`, `allowed_redirect_uris`
- Deletion test verifies: role cleanup from Vault (IsDeletable=true), config K8s-only deletion (IsDeletable=false), mount cleanup
- All checked type assertions use two-value form per Epic 3 retro rule
- Coverage increased from 40.7% to 42.0%
- All existing tests pass — no regressions

### Change Log

- 2026-04-23: Implemented Story 4.3 — Added Keycloak integration infrastructure, decoder methods, test fixtures, and integration tests for JWTOIDCAuthEngineConfig and JWTOIDCAuthEngineRole

### File List

- integration/keycloak/deployment.yaml (new) — Keycloak Deployment + Service manifest
- integration/keycloak/configmap.yaml (new) — Keycloak realm import JSON as ConfigMap
- Makefile (modified) — Added `deploy-keycloak` target; added `deploy-keycloak` to `integration` dependencies
- controllers/controllertestutils/decoder.go (modified) — Added `GetJWTOIDCAuthEngineConfigInstance` and `GetJWTOIDCAuthEngineRoleInstance`
- test/jwtoidcauthengine/test-jwtoidc-auth-mount.yaml (new) — AuthEngineMount prerequisite (type=oidc)
- test/jwtoidcauthengine/test-jwtoidc-auth-config.yaml (new) — JWTOIDCAuthEngineConfig with OIDC credentials from K8s Secret
- test/jwtoidcauthengine/test-jwtoidc-auth-role.yaml (new) — JWTOIDCAuthEngineRole with roleType=oidc
- controllers/jwtoidcauthengine_controller_test.go (new) — Integration test file
