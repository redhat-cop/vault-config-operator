# Secret Engines

The vault-config-operator manages Vault [secret engine](https://www.vaultproject.io/docs/secrets) configuration through Kubernetes Custom Resources. Each supported secret engine has a Config CRD (to configure the engine connection) and one or more Role CRDs (to define dynamic credential generation roles). The operator reconciles these CRs against the Vault API, ensuring the desired secret engine configuration is always applied.

## SecretEngineMount

The `SecretEngineMount` CRD allows a user to create a Secret Engine mount point, here is an example:

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: SecretEngineMount
metadata:
  name: database
spec:
  authentication: 
    path: kubernetes
    role: database-engine-admin
  type: database
  path: postgresql-vault-demo
```

The `type` field specifies the secret engine type.

The `path` field specifies the path at which to mount the secret engine

Many other standard Secret Engine Mount fields are available for fine tuning, see the [Vault Documentation](https://www.vaultproject.io/api-docs/system/mounts#enable-secrets-engine)

This CR is roughly equivalent to this Vault CLI command:

```shell
vault secrets enable -path [namespace/]postgresql-vault-demo/database database
```

## Supported Secret Engines

| Engine | Config CRD | Role CRD(s) | File |
|--------|-----------|-------------|------|
| Database | DatabaseSecretEngineConfig | DatabaseSecretEngineRole, DatabaseSecretEngineStaticRole | [database.md](database.md) |
| PKI | PKISecretEngineConfig | PKISecretEngineRole | [pki.md](pki.md) |
| RabbitMQ | RabbitMQSecretEngineConfig | RabbitMQSecretEngineRole | [rabbitmq.md](rabbitmq.md) |
| GitHub | GitHubSecretEngineConfig | GitHubSecretEngineRole | [github.md](github.md) |
| Quay | QuaySecretEngineConfig | QuaySecretEngineRole, QuaySecretEngineStaticRole | [quay.md](quay.md) |
| Kubernetes | KubernetesSecretEngineConfig | KubernetesSecretEngineRole | [kubernetes.md](kubernetes.md) |
| Azure | AzureSecretEngineConfig | AzureSecretEngineRole | [azure.md](azure.md) |

## Common Configuration

- **[Authentication](../auth-section.md)** — Every secret engine CRD includes an `authentication` block that specifies how the operator authenticates to Vault (typically via Kubernetes auth). See the shared authentication section documentation for details.
- **[Vault Connection](../contributing-vault-apis.md)** — Secret engine CRDs support an optional `connection` block to override the default Vault address and TLS settings. See the contributing guide for the connection configuration pattern.

## See Also

- [vault-config-operator README](../../readme.md) — Project overview and getting started
- [Contributing a New Vault API](../contributing-vault-apis.md) — Developer guide for adding new CRD types
- [Authentication Engines](../auth-engines/index.md) — Auth engine configuration documentation
