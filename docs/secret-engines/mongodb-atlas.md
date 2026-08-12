# MongoDB Atlas Secret Engine

The vault-config-operator supports Vault's [MongoDB Atlas secrets engine](https://developer.hashicorp.com/vault/docs/secrets/mongodbatlas) through two CRDs:

- **MongoDBAtlasSecretEngineConfig** — Configures the MongoDB Atlas connection credentials
- **MongoDBAtlasSecretEngineRole** — Defines roles for dynamic Atlas Programmatic API Key generation

## MongoDBAtlasSecretEngineConfig

Configures the MongoDB Atlas secrets engine with Programmatic API Key credentials. The config path in Vault is `{path}/config`.

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: MongoDBAtlasSecretEngineConfig
metadata:
  name: mongodb-atlas-config
spec:
  authentication:
    path: kubernetes
    role: atlas-engine-admin
  path: mongodbatlas
  rootCredentials:
    secret:
      name: atlas-api-credentials
    usernameKey: public_key
    passwordKey: private_key
```

### Credential Sources

The `rootCredentials` field supports three sources for resolving the MongoDB Atlas API key pair (`public_key` / `private_key`):

| Source | Description |
|--------|-------------|
| `secret` | Kubernetes Secret reference — `usernameKey` maps to public_key, `passwordKey` maps to private_key |
| `vaultSecret` | Vault KV secret path — same key mapping |
| `randomSecret` | RandomSecret CR — provides private_key only; `spec.publicKey` must be set in the spec |

### Important Notes

- **Not deletable**: Vault has no `DELETE /mongodbatlas/config` endpoint, so deleting the CR removes it from Kubernetes but does not alter Vault configuration.
- **Always-write pattern**: The `private_key` is write-only (never returned on read), so the controller always writes the config to Vault on each reconcile to ensure credential rotations propagate.
- **RandomSecret usage**: When using `randomSecret`, `spec.publicKey` must be set directly in the spec since RandomSecret only provides the private key.

## MongoDBAtlasSecretEngineRole

Defines a role for generating dynamic MongoDB Atlas Programmatic API Keys. The role path in Vault is `{path}/roles/{name}`.

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: MongoDBAtlasSecretEngineRole
metadata:
  name: my-atlas-role
spec:
  authentication:
    path: kubernetes
    role: atlas-engine-admin
  path: mongodbatlas
  organizationID: "5cf5a45a9ccf6400e60981b6"
  roles:
    - ORG_READ_ONLY
  ipAddresses:
    - "192.168.1.3"
    - "192.168.1.4"
  ttl: "30m"
  maxTTL: "1h"
```

### Role Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `organizationID` | string | * | Atlas organization ID. Required if `projectID` is not set |
| `projectID` | string | * | Atlas project ID. Required if `organizationID` is not set |
| `roles` | []string | Yes | Atlas roles for the API Key (min 1 required) |
| `ipAddresses` | []string | No | IP whitelist entries |
| `cidrBlocks` | []string | No | CIDR whitelist entries |
| `projectRoles` | []string | No | Roles for org key assigned to a project |
| `ttl` | string | No | Credential TTL (e.g., "30m", "2h") |
| `maxTTL` | string | No | Maximum credential TTL |

### Valid Organization Roles

`ORG_OWNER`, `ORG_MEMBER`, `ORG_GROUP_CREATOR`, `ORG_BILLING_ADMIN`, `ORG_READ_ONLY`

### Valid Project Roles

`GROUP_CHARTS_ADMIN`, `GROUP_CLUSTER_MANAGER`, `GROUP_DATA_ACCESS_ADMIN`, `GROUP_DATA_ACCESS_READ_ONLY`, `GROUP_DATA_ACCESS_READ_WRITE`, `GROUP_OWNER`, `GROUP_READ_ONLY`

### Important Notes

- **Deletable**: Deleting the CR removes the role from Vault via `DELETE /roles/{name}`.
- **Name override**: Use `spec.name` to override the Vault role name (defaults to `metadata.name`).
- **Immutable fields**: `spec.path` and `spec.name` cannot be changed after creation.

## Prerequisites

1. Mount the MongoDB Atlas secrets engine:

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: SecretEngineMount
metadata:
  name: mongodbatlas
spec:
  authentication:
    path: kubernetes
    role: atlas-engine-admin
  type: mongodbatlas
  path: mongodbatlas
```

2. Create a Kubernetes Secret with your Atlas Programmatic API Key:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: atlas-api-credentials
type: Opaque
stringData:
  public_key: "your-atlas-public-key"
  private_key: "your-atlas-private-key"
```

## See Also

- [Vault MongoDB Atlas Secrets Engine Documentation](https://developer.hashicorp.com/vault/docs/secrets/mongodbatlas)
- [Vault MongoDB Atlas API Reference](https://developer.hashicorp.com/vault/api-docs/secret/mongodbatlas)
