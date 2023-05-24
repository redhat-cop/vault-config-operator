# Policy Management APIs

- [Policy Management APIs](#policy-management-apis)
  - [Policy](#policy)
  - [PasswordPolicy](#passwordpolicy)

## Policy

The `Policy` CRD allows a user to create a [Vault Policy](https://www.vaultproject.io/docs/concepts/policies), here is an example:

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: Policy
metadata:
  name: database-creds-reader
spec:
  authentication: 
    path: kubernetes
    role: policy-admin
  policy: |
    # Configure read secrets
    path "/{{identity.entity.aliases.auth_kubernetes_804f1655.metadata.service_account_namespace}}/database/creds/+" {
      capabilities = ["read"]
    }
  type: acl  
```

Notice that in this policy we have parametrized the path based on the namespace of the connecting service account.

This creates a policy at this path `/sys/policies/acl/<name>`

Automatically resolving an authentication engine accessor into a templated policy HCL can also be achieved for a more declarative style:
the operator will automatically replace all placeholders with format `${auth/<auth engine path>/@accessor}` with the accessor of the authentication
engine mounted at path `<auth engine path>`.

See [Vault Templated Policies](https://developer.hashicorp.com/vault/docs/concepts/policies#templated-policies) for more details on templated policies.

Note: the Vault role used for authentication (specified in the `authentication` section) must have `read` and `list` access
to Vault's `sys/auth` API endpoint for this automated resolution to work.

Any unresolved placeholder is left as-is into the configured policy.

The example above then becomes:

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: Policy
metadata:
  name: database-creds-reader
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  policy: |
    # Configure read secrets
    path "/{{identity.entity.aliases.${auth/kubernetes/@accessor}.metadata.service_account_namespace}}/database/creds/+" {
      capabilities = ["read"]
    }
  type: acl
```

## PasswordPolicy

The `PasswordPolicy` CRD allows a user to create a [Vault Password Policy](https://www.vaultproject.io/docs/concepts/password-policies), here is an example:

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: PasswordPolicy
metadata:
  name: simple-password-policy
spec:
  authentication: 
    path: kubernetes
    role: policy-admin  
  passwordPolicy: |
    length = 20
    rule "charset" {
      charset = "abcdefghijklmnopqrstuvwxyz"
    }
```