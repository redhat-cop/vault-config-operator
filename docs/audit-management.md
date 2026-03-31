# Audit Management

Vault's [audit devices](https://developer.hashicorp.com/vault/docs/audit) keep detailed logs of all requests and responses made to Vault. 

The vault-config-operator supports the following APIs related to Vault audit management:

- [Audit](#audit)
- [AuditRequestHeader](#auditrequestheader)

## Audit

The `Audit` CRD allows you to enable and configure [Vault audit devices](https://developer.hashicorp.com/vault/docs/audit). Audit devices are the components in Vault that keep a detailed log of all authenticated requests and responses to Vault.

Here is an example of enabling a file audit device:

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: Audit
metadata:
  name: file-audit-device
spec:
  authentication: 
    path: kubernetes
    role: policy-admin
  path: file
  type: file
  description: "File audit device for logging"
  options:
    file_path: /vault/audit/vault-audit.log
  local: false
```

### Field Description

- `path`: The path where the audit device will be mounted (e.g., "file", "file2", "syslog"). This appears in the `sys/audit` list.
- `type`: The type of audit device (e.g., "file", "socket", "syslog"). Different types support different configuration options.
- `description`: A human-friendly description of the audit device.
- `options`: Configuration options specific to the audit device type. For example:
  - For `file` type: `file_path` specifies where audit logs are written
  - For `socket` type: `address` specifies the socket address
  - For `syslog` type: `facility` and `tag` configure syslog behavior
- `local`: If true, the audit device is local to the cluster. Local audit devices are not replicated and are removed upon replication setup.

## AuditRequestHeader

The `AuditRequestHeader` CRD allows you to configure which HTTP request headers should be audited by Vault. By default, Vault does not audit headers in requests, but you can configure specific headers to be captured in audit logs.

Here is an example:

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: AuditRequestHeader
metadata:
  name: x-forwarded-for-header
spec:
  authentication: 
    path: kubernetes
    role: policy-admin
  name: X-Forwarded-For
  hmac: false
```

### Field Description

- `name`: The name of the HTTP request header to audit (e.g., "X-Forwarded-For", "X-Request-ID"). This is case-insensitive.
- `hmac`: If `true`, the header value will be HMAC'd in the audit logs for security. If `false`, the header value will appear in plain text in the audit logs.
