# Namespaces

## Namespace management

Note: this feature requires the vault enterprise edition: [namespaces](https://developer.hashicorp.com/vault/docs/enterprise/namespaces)

You can create Vault namespaces using the Namespace resource:

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: namespace
metadata:
  name: my-namespace
spec:
  authentication: 
    path: kubernetes
    role: policy-admin
```

If nothing else is provided, the namespaces will be named using the Kubernetes resource name, here `my-namespace`.

You can specify another name using the `name` field.

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: namespace
metadata:
  name: my-namespace
spec:
  authentication: 
    path: kubernetes
    role: policy-admin
  name: my-other-name
```

This will create the namespace `my-other-name` under the authentication namespace.

For example, if you are authenticated under root, it will create the namespace directly under root.
If you are authenticated in another namespace like `my-root-namespace`, it'll create it at `my-root-namespace/my-other-name`.

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: namespace
metadata:
  name: my-namespace
spec:
  authentication: 
    path: kubernetes
    role: policy-admin
    namespace: my-root-namespace
  name: my-other-name
```

## Nested namespaces

You can also specify nested namespace path under which you would like to create the namespaces, independently from the authentication namespace, using the `path` field.

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: namespace
metadata:
  name: my-namespace
spec:
  authentication: 
    path: kubernetes
    role: policy-admin
    namespace: my-root-namespace
  name: my-other-name
  path: my-first-level-namespace
```

The result here would be a namespace created under `my-root-namespace/my-first-level-namespace/my-other-name`.