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
```

Notice that in this policy we have parametrized the path based on the namespace of the connecting service account.

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