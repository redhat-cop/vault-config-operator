# The Authentication Section

Each API has an Authentication section that specifies how to authenticate to Vault. Here is an example:

```yaml
  authentication: 
    path: kubernetes
    role: policy-admin
    namespace: tenant-namespace
    serviceAccount:
      name: vaultsa
```

The `path` field specifies the path at which the Kubernetes authentication role is mounted.

The `role` field specifies which role to request when authenticating

The `namespace` field specifies the Vault namespace (not related to Kubernetes namespace) in which the operator authenticates. The resource is also managed in this namespace, unless `targetNamespace` is set. This is optional.

The `targetNamespace` field specifies the Vault namespace in which the resource is managed, relative to `namespace`. This is optional. See [Managing resources in a child namespace](#managing-resources-in-a-child-namespace).

The `serviceAccount.name` specifies the token of which service account to use during the authentication process.

So the above configuration roughly correspond to the following command:

```shell
vault write [tenant-namespace/]auth/kubernetes/login role=policy-admin jwt=<vaultsa jwt token>
```

## Managing resources in a child namespace

By default, the operator authenticates and manages the resource in the same Vault namespace. This requires an auth method in every namespace that the operator manages.

Set `targetNamespace` to authenticate in a parent namespace and manage the resource in a child namespace:

```yaml
  authentication:
    path: kubernetes
    role: tenant-admin
    targetNamespace: tenant-a
```

The operator logs in at `auth/kubernetes/login` in the root namespace, and then manages the resource in the `tenant-a` namespace. `tenant-a` does not need its own auth method.

`targetNamespace` is relative to `namespace`. With `namespace: org` and `targetNamespace: tenant-a`, the operator logs in to `org` and manages the resource in `org/tenant-a`.

The policy of the role must grant the resource paths in the child namespace. A policy in a parent namespace grants a child path when the path starts with the child namespace:

```hcl
path "tenant-a/sys/policy/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}
```

This feature requires Vault Enterprise. Vault Community Edition ignores the namespace, so every resource with a `targetNamespace` is managed in the root namespace.
