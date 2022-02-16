# Authentication Engines APIs

- [Authentication Engines APIs](#authentication-engines-apis)
  - [AuthEngineMount](#authenginemount)
  - [KubernetesAuthEngineConfig](#kubernetesauthengineconfig)
  - [KubernetesAuthEngineRole](#kubernetesauthenginerole)

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

## KubernetesAuthEngineConfig

The `KubernetesAuthEngineConfig` CRD allows a user to configure an authentication engine mount of [type Kubernetes](https://www.vaultproject.io/docs/auth/kubernetes).

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: KubernetesAuthEngineConfig
metadata:
  name: authenginemount-sample
spec:
  authentication: 
    path: kubernetes
    role: policy-admin
  path: kube-authengine-mount-sample
  tokenReviewerServiceAccount:
    name: token-review-sa
  kubernetesHost:   
  kubernetesCACert: ...  
```

The `path` field specifies the path to configure. the complete path of the configuration will be: `[namespace/]auth/{.spec.path}/{metadata.name}/config`

The `tokenReviewerServiceAccount.name` field specifies the service account to be used to perform the token review. This account must exists and must be granted the TokenReviews create permission. If not specified it will default to `default`.

The `kubernetesCACert` field is the base64 encoded CA certificate that can be used to validate the connection to the master API. It will default to the content of the file `/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"`. This default should work for most cases.

The `kubernetesHost` field defines the master api endpoint. It defaults to `https://kubernetes.default.svc:443` and it should work most cases.

## KubernetesAuthEngineRole

The `KubernetesAuthEngineRole` creates a [Vault Authentication Role](https://www.vaultproject.io/docs/auth/kubernetes#configuration) for a Kubernetes Authentication mount, here is an example:

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: KubernetesAuthEngineRole
metadata:
  name: database-engine-admin
spec:
  authentication: 
    path: kubernetes
    role: policy-admin
  path: kubernetes  
  policies:
    - database-engine-admin
  targetServiceAccounts: 
  - vaultsa  
  targetNamespaceSelector:
    matchLabels:
      postgresql-enabled: "true"
```

The `path` field specifies the path of the Kubernetes Authentication Mount at which the role will be mounted.

The `policies` field specifies which Vault policies will be associated with this role.

The `targetServiceAccounts` field specifies which service accounts can authenticate. If not specified, it defaults to `default`.

The `targetNamespaceSelector` field specifies from which kubernetes namespaces it is possible to authenticate. Notice as the set of namespaces selected by the selector varies, this configuration will be updated. It is also possible to specify a static set of namespaces.

Many other standard Kubernetes Authentication Role fields are available for fine tuning, see the [Vault Documentation](https://www.vaultproject.io/api-docs/auth/kubernetes#create-role)

This CR is roughly equivalent to this Vault CLI command:

```shell
vault write [namespace/]auth/kubernetes/role/database-engine-admin bound_service_account_names=vaultsa bound_service_account_namespaces=<dynamically generated> policies=database-engine-admin
```
