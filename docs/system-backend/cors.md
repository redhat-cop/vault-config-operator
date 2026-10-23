# CORS Configuration

Vault's [CORS settings](https://developer.hashicorp.com/vault/api-docs/system/config-cors) control which origins and headers are permitted for cross-origin requests. The vault-config-operator supports managing these settings declaratively via the `CORSConfig` CRD.

## CORSConfig

The `CORSConfig` CRD manages the CORS configuration at `sys/config/cors`. This is a singleton resource — only one CR should exist per Vault instance. If multiple CRs target the same Vault, the last-reconciled wins.

**Important:** `CORSConfig` is deletable. Deleting the CR will disable CORS in Vault by issuing a DELETE to `sys/config/cors`.

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: CORSConfig
metadata:
  name: cors-config
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  allowedOrigins:
    - "https://example.com"
    - "https://app.example.com"
  allowedHeaders:
    - "X-Custom-Header"
```

### Wildcard Origin Example

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: CORSConfig
metadata:
  name: cors-config
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  allowedOrigins:
    - "*"
```

### Field Description

- `allowedOrigins`: The list of origins permitted to make cross-origin requests. Use `["*"]` to allow all origins. Required and must contain at least one entry.
- `allowedHeaders`: Additional headers permitted in cross-origin requests. Vault always includes its own default headers; specifying extras here is additive. Optional.

### Drift Detection

Vault appends its own default headers (e.g., `Content-Type`, `Authorization`, `X-Requested-With`) to the `allowed_headers` list in its read response. The operator uses a **subset check** for headers: as long as all headers specified in the CR are present in Vault's response, the state is considered equivalent. For `allowed_origins`, the operator uses an **order-insensitive set comparison**.
