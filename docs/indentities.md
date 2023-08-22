# Identites

The always present [Identity secret engine](https://developer.hashicorp.com/vault/docs/concepts/identity) powers some advanced features of Vault specifically in the space of human interactions.

The vault config operator supports the following API related to the Identity secret engine

  - [Group](#group)
  - [GroupAlias](#groupalias)


## Group

The Group CRD allows defining a [Vault Group](https://developer.hashicorp.com/vault/docs/concepts/identity#identity-groups).

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: Group
metadata:
  name: group-sample
spec:
  authentication: 
    path: kubernetes
    role: policy-admin
  type: external
  metadata: 
    team: team-abc
  policies: 
  - team-abc-access
```

## GroupAlias

The GroupAlias CRD allows defining a [Vault GroupAlias](https://developer.hashicorp.com/vault/api-docs/secret/identity/group-alias).

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: GroupAlias
metadata:
  name: groupalias-sample
spec:
  authentication: 
    path: kubernetes
    role: policy-admin
  authEngineMountPath: kubernetes
  groupName: group-sample 
```

Notice that we pass the auth engine mount path and the group name as opposed to the respctive IDs as expected by the Vault API. The vault-config-operator will resolved those values to teh relative IDs. This should keep things simpler for the user.