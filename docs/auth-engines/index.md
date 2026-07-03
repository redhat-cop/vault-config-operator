# Authentication Engines

The vault-config-operator manages Vault [authentication engine](https://www.vaultproject.io/docs/auth) configuration through Kubernetes Custom Resources. Each supported auth engine has a Config CRD (to configure the engine mount) and a Role or Group CRD (to define authentication roles or group policies). The operator reconciles these CRs against the Vault API, ensuring the desired auth configuration is always applied.

## AuthEngineMount

The `AuthEngineMount` CRD allows a user to define an [authentication engine endpoint](https://www.vaultproject.io/docs/auth)

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: AuthEngineMount
metadata:
  name: authenginemount-sample
spec:
  authentication: 
    path: kubernetes
    role: policy-admin
  type: kubernetes
  path: kube-authengine-mount-sample
```

The `type` field specifies the type of the authentication engine.

The `path` field specifies the path at which the auth engine is mounted. The complete path will be: `[namespace/]auth/{.spec.path}/{metadata.name}`

## Supported Auth Engines

| Engine | Config CRD | Role/Group CRD | File |
|--------|-----------|----------------|------|
| Kubernetes | KubernetesAuthEngineConfig | KubernetesAuthEngineRole | [kubernetes.md](kubernetes.md) |
| LDAP | LDAPAuthEngineConfig | LDAPAuthEngineGroup | [ldap.md](ldap.md) |
| JWT/OIDC | JWTOIDCAuthEngineConfig | JWTOIDCAuthEngineRole | [jwt-oidc.md](jwt-oidc.md) |
| GCP | GCPAuthEngineConfig | GCPAuthEngineRole | [gcp.md](gcp.md) |
| Azure | AzureAuthEngineConfig | AzureAuthEngineRole | [azure.md](azure.md) |
| TLS Certificate | CertAuthEngineConfig | CertAuthEngineRole | [cert.md](cert.md) |

## Common Configuration

- **[Authentication](../auth-section.md)** — Every auth engine CRD includes an `authentication` block that specifies how the operator authenticates to Vault (typically via Kubernetes auth). See the shared authentication section documentation for details.
- **[Vault Connection](../contributing-vault-apis.md)** — Auth engine CRDs support an optional `connection` block to override the default Vault address and TLS settings. See the contributing guide for the connection configuration pattern.

## See Also

- [vault-config-operator README](../../readme.md) — Project overview and getting started
- [Contributing a New Vault API](../contributing-vault-apis.md) — Developer guide for adding new CRD types
- [Secret Engines](../secret-engines/index.md) — Secret engine configuration documentation
