# JWT Auth + PKI Certificates — End-to-End Example

## Scenario

This example demonstrates a complete Vault configuration where CI/CD pipelines (e.g., GitHub Actions) authenticate using JWT tokens and then issue short-lived TLS certificates via a PKI secret engine. The JWT auth engine validates tokens against a remote JWKS endpoint, and a Policy connects the authenticated identity to the PKI certificate issuance paths.

## Prerequisites

- A running Vault instance with the vault-config-operator deployed
- The operator's Kubernetes auth configured at `kubernetes` with a `policy-admin` role
- A JWKS endpoint accessible from Vault (this example uses GitHub Actions' public JWKS URL)

## Resources Created

| # | Kind | Name | Purpose |
|---|------|------|---------|
| 1 | AuthEngineMount | jwt | Mounts the JWT auth engine at `ci/jwt` |
| 2 | JWTOIDCAuthEngineConfig | jwt-ci-config | Configures JWT validation via JWKS |
| 3 | Policy | ci-cert-issuer | Grants access to PKI certificate issuance |
| 4 | SecretEngineMount | pki | Mounts the PKI secret engine at `ci/pki` |
| 5 | PKISecretEngineConfig | pki-ci-root | Generates a root CA for CI certificates |
| 6 | PKISecretEngineRole | ci-service | Defines certificate issuance parameters |
| 7 | JWTOIDCAuthEngineRole | ci-runner-role | JWT auth role linking to the policy |

> **Path composition:** `AuthEngineMount` and `SecretEngineMount` compose their Vault mount path as `{spec.path}/{metadata.name}`. This example uses `path: ci` as a grouping prefix, producing auth mount `ci/jwt` and secret mount `ci/pki`.

## Apply

```shell
kubectl apply -f e2e-jwt-pki.yaml
```

## Verify

```shell
# Verify auth engine is mounted
vault auth list | grep ci/jwt

# Verify PKI CA is configured
vault read ci/pki/ca/pem

# Verify certificate issuance role exists
vault read ci/pki/roles/ci-service

# Issue a test certificate (after authenticating via JWT)
vault write ci/pki/issue/ci-service common_name="myapp.internal.example.com" ttl="1h"
```

## How It Works

1. **Auth mount** — The JWT auth engine is mounted at `ci/jwt` and configured to validate tokens using GitHub Actions' JWKS endpoint.
2. **Policy bridge** — The `ci-cert-issuer` policy grants `read` and `update` access to `ci/pki/issue/ci-service`, `ci/pki/sign/ci-service`, and `read` on `ci/pki/ca/pem`.
3. **Secret engine** — The PKI engine at `ci/pki` has a root CA and a role (`ci-service`) that issues short-lived certificates for `*.internal.example.com` and `*.svc.cluster.local`.
4. **Auth role** — The `ci-runner-role` binds JWT claims (audience, issuer, repository) to the `ci-cert-issuer` policy, completing the auth-to-secrets flow.

## Cleanup

```shell
kubectl delete -f e2e-jwt-pki.yaml
```

> **Note:** Deleting the CRs removes the Kubernetes objects. The PKI CA and auth engine configuration persist in Vault until their mounts are manually deleted via `vault secrets disable ci/pki` and `vault auth disable ci/jwt`.
