# Identities

The always present [Identity secret engine](https://developer.hashicorp.com/vault/docs/concepts/identity) powers some advanced features of Vault specifically in the space of human interactions.

The vault config operator supports the following API related to the Identity secret engine

  - [Group](#group)
  - [GroupAlias](#groupalias)
  - [IdentityOIDCProvider](#identityoidcprovider)
  - [IdentityOIDCScope](#identityoidcscope)
  - [IdentityOIDCClient](#identityoidcclient)
  - [IdentityOIDCAssignment](#identityoidcassignment)
  - [IdentityTokenConfig](#identitytokenconfig)
  - [IdentityTokenKey](#identitytokenkey)
  - [IdentityTokenRole](#identitytokenrole)


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

## IdentityOIDCProvider

The IdentityOIDCProvider CRD allows defining a [Vault OIDC Provider](https://developer.hashicorp.com/vault/api-docs/secret/identity/oidc-provider#create-or-update-a-provider).

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: IdentityOIDCProvider
metadata:
  name: identityoidcprovider-sample
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  allowedClientIDs:
  - "*"
  scopesSupported:
  - test-scope
```

The following fields are available:

- `issuer` - (optional) What will be used as the `scheme://host:port` component for the `iss` claim of ID tokens. Defaults to a URL with Vault's `api_addr`.
- `allowedClientIDs` - (optional) List of client IDs permitted to use the provider. Use `"*"` to allow all clients.
- `scopesSupported` - (optional) List of scopes available for requesting on the provider.

## IdentityOIDCScope

The IdentityOIDCScope CRD allows defining a [Vault OIDC Scope](https://developer.hashicorp.com/vault/api-docs/secret/identity/oidc-provider#create-or-update-a-scope).

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: IdentityOIDCScope
metadata:
  name: identityoidcscope-sample
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  template: '{ "groups": {{identity.entity.groups.names}} }'
  description: A simple scope example.
```

The following fields are available:

- `template` - (optional) The JSON template string for the scope. May be provided as escaped JSON or base64 encoded JSON.
- `description` - (optional) A description of the scope.

## IdentityOIDCClient

The IdentityOIDCClient CRD allows defining a [Vault OIDC Client](https://developer.hashicorp.com/vault/api-docs/secret/identity/oidc-provider#create-or-update-a-client).

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: IdentityOIDCClient
metadata:
  name: identityoidcclient-sample
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  key: default
  clientType: confidential
  redirectURIs:
  - https://example.com/callback
  assignments:
  - allow_all
  idTokenTTL: "24h"
  accessTokenTTL: "24h"
```

The following fields are available:

- `key` - (optional, default: `"default"`) Reference to a named key resource used to sign ID tokens. Cannot be modified after creation.
- `redirectURIs` - (optional) List of redirection URI values used by the client.
- `assignments` - (optional) List of assignment resources associated with the client. Use `"allow_all"` to allow all Vault entities.
- `clientType` - (optional, default: `"confidential"`) Client type: `"confidential"` or `"public"`. Cannot be modified after creation.
- `idTokenTTL` - (optional, default: `"24h"`) Time-to-live for ID tokens.
- `accessTokenTTL` - (optional, default: `"24h"`) Time-to-live for access tokens.

## IdentityOIDCAssignment

The IdentityOIDCAssignment CRD allows defining a [Vault OIDC Assignment](https://developer.hashicorp.com/vault/api-docs/secret/identity/oidc-provider#create-or-update-an-assignment).

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: IdentityOIDCAssignment
metadata:
  name: identityoidcassignment-sample
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  entityIDs:
  - b6094ac6-baf4-6520-b05a-2bd9f07c66da
  groupIDs:
  - 262ca5b9-7b69-0a84-446a-303dc7d778af
```

The following fields are available:

- `entityIDs` - (optional) List of Vault entity IDs.
- `groupIDs` - (optional) List of Vault group IDs.

## IdentityTokenConfig

The IdentityTokenConfig CRD allows configuring the [Identity Tokens backend](https://developer.hashicorp.com/vault/api-docs/secret/identity/tokens#configure-the-identity-tokens-backend). This is a singleton configuration resource — deleting the CR will not remove the configuration from Vault.

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: IdentityTokenConfig
metadata:
  name: identitytokenconfig-sample
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  issuer: "https://example.com:1234"
```

The following fields are available:

- `issuer` - (optional) Issuer URL to be used in the `iss` claim of the token. If not set, Vault's `api_addr` will be used.

## IdentityTokenKey

The IdentityTokenKey CRD allows creating or updating a [named key](https://developer.hashicorp.com/vault/api-docs/secret/identity/tokens#create-a-named-key) used by a role to sign tokens.

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: IdentityTokenKey
metadata:
  name: identitytokenkey-sample
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  rotationPeriod: "24h"
  verificationTTL: "24h"
  allowedClientIDs:
  - "*"
  algorithm: RS256
```

The following fields are available:

- `rotationPeriod` - (optional, default: `"24h"`) How often to generate a new signing key.
- `verificationTTL` - (optional, default: `"24h"`) How long the public portion of a signing key will be available for verification after being rotated.
- `allowedClientIDs` - (optional) List of role client IDs allowed to use this key. Use `"*"` to allow all roles.
- `algorithm` - (optional, default: `"RS256"`) Signing algorithm. Allowed values: RS256, RS384, RS512, ES256, ES384, ES512, EdDSA.

## IdentityTokenRole

The IdentityTokenRole CRD allows creating or updating a [role](https://developer.hashicorp.com/vault/api-docs/secret/identity/tokens#create-or-update-a-role). ID tokens are generated against a role and signed against a named key.

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: IdentityTokenRole
metadata:
  name: identitytokenrole-sample
spec:
  authentication:
    path: kubernetes
    role: policy-admin
  key: identitytokenkey-sample
  ttl: "24h"
```

The following fields are available:

- `key` - (required) A configured named key; the key must already exist.
- `template` - (optional) The template string to use for generating tokens. May be in string-ified JSON or base64 format.
- `clientID` - (optional) Client ID. A random ID will be generated if left unset.
- `ttl` - (optional, default: `"24h"`) TTL of the tokens generated against the role.