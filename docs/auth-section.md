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

The `targetNamespace` field specifies the Vault namespace in which the resource is managed, relative to the namespace in which the operator authenticates, for example `tenant-a` or `org/tenant-a`. It cannot start or end with `/`, and it cannot contain empty, `.` or `..` segments. This is optional. See [Managing resources in a child namespace](#managing-resources-in-a-child-namespace). `targetNamespace` is a Vault namespace. Do not confuse it with `spec.targetNamespaces` on `KubernetesAuthEngineRole` and `KubernetesSecretEngineRole`, which holds Kubernetes namespaces.

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

When `VAULT_NAMESPACE` is not set, the operator authenticates at `auth/kubernetes/login` in the root namespace, and then manages the resource in the `tenant-a` namespace. `tenant-a` does not need its own auth method.

`targetNamespace` is relative to the namespace in which the operator authenticates. With `namespace: org` and `targetNamespace: tenant-a`, the operator authenticates in `org` and manages the resource in `org/tenant-a`. When `namespace` is empty, the operator authenticates in the namespace in the `VAULT_NAMESPACE` environment variable of the operator. With `VAULT_NAMESPACE=admin` and `targetNamespace: tenant-a`, the operator authenticates in `admin` and manages the resource in `admin/tenant-a`.

When a resource sets `targetNamespace`, you cannot change `namespace` or `targetNamespace` in a way that moves the resource to a different Vault namespace, because the change leaves the old resource in Vault. To move a resource, delete it and create it again. You can change the fields when the Vault namespace stays the same. For example, you can change `namespace: org/tenant-a` to `namespace: org` and `targetNamespace: tenant-a`, and the operator keeps the same resource.

The validating webhook enforces this limit. The limit does not apply to `VaultSecret`, because it only reads from Vault. `Audit`, `AuditRequestHeader`, `Entity`, `EntityAlias` and `RabbitMQSecretEngineConfig` have no active validating webhook, so do not change `targetNamespace` on these kinds. For a resource without `targetNamespace`, a change to only `namespace` still moves the resource, as in earlier versions.

A resource that refers to a `RandomSecret`, for example through `rootCredentials.randomSecret`, reads the secret in the Vault namespace of that resource, not in the Vault namespace of the `RandomSecret`. Give the `RandomSecret` the same `namespace` and `targetNamespace` as the resource that refers to it.

The policy of the role must grant the resource paths in the child namespace. A policy in a parent namespace grants a child path when the path starts with the child namespace:

```hcl
path "tenant-a/sys/policy/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}
```

This feature requires Vault Enterprise. Vault Community Edition ignores the namespace, so every resource with a `targetNamespace` is managed in the root namespace.
