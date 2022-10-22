# Authentication Engines APIs

- [Authentication Engines APIs](#authentication-engines-apis)
  - [AuthEngineMount](#authenginemount)
  - [KubernetesAuthEngineConfig](#kubernetesauthengineconfig)
  - [KubernetesAuthEngineRole](#kubernetesauthenginerole)
  - [LDAPAuthEngineConfig](#ldapauthengineconfig)
    - [LDAPAuthEngineGroup](#ldapauthenginegroup)
  - [JWTOIDCAuthEngineConfig](#jwtoidcauthengineconfig)
    - [JWTOIDCAuthEngineRole](#jwtoidcauthenginerole)

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

## LDAPAuthEngineConfig

The `LDAPAuthEngineConfig` CRD allows a user to configure an authentication engine mount of [type LDAP](https://www.vaultproject.io/docs/auth/ldap).

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: LDAPAuthEngineConfig
metadata:
  name: authenginemount-sample
spec:
  UPNDomain:
  anonymousGroupSearch:
  authentication: 
    path: kubernetes
    role: policy-admin
    serviceAccount:
      name: admin-sa
  bindDN: cn=vault,ou=Users,dc=example,dc=com
  bindCredentials:
    secret:
      name: bindcredentials
    passwordKey: password
    usernameKey: username
  caseSensitiveNames: false
  groupDN:
  groupFilter:
  insecureTLS: true
  path: ldap
  url: ldaps://ldap.myorg.com:636
  userAttr: "samaccountname"
  userDN: ou=Users,dc=example,dc=com 
  userFilter: ({{.UserAttr}}={{.Username}})
  usernameAsAlias: false
  ...
```
  The `UPNDomain` field - userPrincipalDomain used to construct the UPN string for the authenticating user. The constructed UPN will appear as [username]@UPNDomain. Example: example.com, which will cause vault to bind as username@example.com.

  The `anonymousGroupSearch` field - Use Anonymous binds when performing LDAP group searches (note: even when true, the initial credentials will still be used for the initial connection test).

  The `bindDN` field - Username used to connect to the LDAP service on the specified LDAP Server.
  If in the form accountname@domain.com, the username is transformed into a proper LDAP bind DN, for example, CN=accountname,CN=users,DC=domain,DC=com, when accessing the LDAP server.

  The `bindPass` field - Password to use along with bindDN when performing user search.
  The bindPass and possibly the bindDN can be retrived a three different ways:

  1. From a Kubernetes secret, specifying the `bindCredentialsFromSecret` field. The secret must be of [basic auth type](https://kubernetes.io/docs/concepts/configuration/secret/#basic-authentication-secret). If the secret is updated this connection will also be updated.
  2. From a Vault secret, specifying the `bindCredentialsFromVaultSecret` field.
  3. From a [RandomSecret](#RandomSecret), specifying the `bindCredentialsFromRandomSecret` field. When the RandomSecret generates a new secret, this connection will also be updated.
  
  The `caseSensitiveNames` field -  If set, user and group names assigned to policies within the backend will be case sensitive. Otherwise, names will be normalized to lower case. Case will still be preserved when sending the username to the LDAP server at login time; this is only for matching local user/group definitions.
  
  The `certificate` field – CA certificate to use when verifying LDAP server certificate, must be x509 PEM encoded.
  
  The `clientTLSCert` field - Client certificate to provide to the LDAP server, must be x509 PEM encoded (optional).
  
  The `clientTLSKey` field - Client certificate key to provide to the LDAP server, must be x509 PEM encoded (optional).
  
  The `denyNullBind` field - This option prevents users from bypassing authentication when providing an empty password.
  
  The `discoverDN` field - Use anonymous bind to discover the bind DN of a user.
  
  The `groupAttr` field - LDAP attribute to follow on objects returned by groupfilter in order to enumerate user group membership. Examples: for groupfilter queries returning group objects, use: cn. For queries returning user objects, use: memberOf. The default is cn.
  
  The `groupDN` field – LDAP search base to use for group membership search. This can be the root containing either groups or users. Example: ou=Groups,dc=example,dc=com
  
  The `groupFilter` field – Go template used when constructing the group membership query. The template can access the following context variables: [UserDN, Username]. The default is (|(memberUid={{.Username}})(member={{.UserDN}})(uniqueMember={{.UserDN}})), which is compatible with several common directory schemas. To support nested group resolution for Active Directory, instead use the following query: (&(objectClass=group)(member:1.2.840.113556.1.4.1941:={{.UserDN}})).
  
  The `insecureTLS` field - If true, skips LDAP server SSL certificate verification - insecure, use with caution!
  
  The `path` field - The path field specifies the path to configure. the complete path of the configuration will be: [namespace/]auth/{.spec.path}/{metadata.name}/config
  
  The `requestTimeout` field - Timeout, in seconds, for the connection when making requests against the server before returning back an error.
  
  The `startTLS` field - If true, issues a StartTLS command after establishing an unencrypted connection.
  
  The `TLSMaxVersion` field - Minimum TLS version to use. Accepted values are tls10, tls11, tls12 or tls13.
  
  The `TLSMinVersion` field - Maximum TLS version to use. Accepted values are tls10, tls11, tls12 or tls13.
  
  The `tokenBoundCIDRs` field - List of CIDR blocks; if set, specifies blocks of IP addresses which can authenticate successfully, and ties the resulting token to these blocks as well.
  
  The `tokenExplicitMaxTTL` field - If set, will encode an explicit max TTL onto the token. This is a hard cap even if token_ttl and token_max_ttl would otherwise allow a renewal.
  
  The `tokenMaxTTL` field - The maximum lifetime for generated tokens. This current value of this will be referenced at renewal time.
  
  The `tokenNoDefaultPolicy` field - If set, the default policy will not be set on generated tokens; otherwise it will be added to the policies set in token_policies.
  
  The `tokenNumUses` field - The maximum number of times a generated token may be used (within its lifetime); 0 means unlimited. If you require the token to have the ability to create child tokens, you will need to set this value to 0.
  
  The `tokenPeriod` field - The period, if any, to set on the token.
  
  The `tokenPolicies` field - List of policies to encode onto generated tokens. Depending on the auth method, this list may be supplemented by user/group/other values.

  The `tokenTTL` field - The incremental lifetime for generated tokens. This current value of this will be referenced at renewal time.
  
  The `tokenType` field - The type of token that should be generated. Can be service, batch, or default to use the mount's tuned default (which unless changed will be service tokens). For token store roles, there are two additional possibilities: default-service and default-batch which specify the type to return unless the client requests a different type at generation time.
  
  The `url` field - The LDAP server to connect to. Examples: ldap://ldap.myorg.com, ldaps://ldap.myorg.com:636. Multiple URLs can be specified with commas, e.g. ldap://ldap.myorg.com,ldap://ldap2.myorg.com; these will be tried in-order.
  
  The `userAttr` field - Attribute on user attribute object matching the username passed when authenticating. Examples: sAMAccountName, cn, uid
  
  The `userDN` field - Base DN under which to perform user search. Example: ou=Users,dc=example,dc=com
  
  The `userFilter` field - An optional LDAP user search filter. The template can access the following context variables: UserAttr, Username. The default is ({{.UserAttr}}={{.Username}}), or ({{.UserAttr}}={{.Username@.upndomain}}) if upndomain is set.
  
  The `usernameAsAlias` field - If set to true, forces the auth method to use the username passed by the user as the alias name.


## LDAPAuthEngineGroup

The `LDAPAuthEngineGroup` CRD allows a user to create or update [LDAP group policies](https://www.vaultproject.io/api-docs/auth/ldap#create-update-ldap-group)

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: LDAPAuthEngineGroup
metadata:
  name: ldapauthenginegroup-sample3
spec:
  authentication: 
    path: kubernetes
    role: policy-admin
    serviceAccount:
      name: default
  name: "test-3"
  path: "ldap/test"
  policies: "admin, audit, users"
```
  The `name` field - The name of the LDAP group

  The `path` field - The path field specifies the LDAP auth path where to create the Group. The complete path of the configuration will be: [namespace/]auth/{.spec.path}/groups/name

  The `policies` field - Comma-separated list of policies associated to the group


## JWTOIDCAuthEngineConfig

The `JWTOIDCAuthEngineConfig` CRD allows a user to configure an authentication engine mount of type [JWT/OIDC](https://developer.hashicorp.com/vault/api-docs/auth/jwt).

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: JWTOIDCAuthEngineConfig
metadata:
  name: azure-oidc
spec:
  authentication: 
    path: kubernetes
    role: policy-admin
  path: oidc-aad
  OIDCClientID: "xxxxxx-xxxx-xxxx-xxxx-xxxxxxxxx"
  OIDCCredentials:
    secret: 
      name: oidccredentials
    usernameKey: client_id
    passwordKey: client_secret
  OIDCDiscoveryURL: "https://login.microsoftonline.com/xxxxxx-xxxx-xxxx-xxxxx-xxxxxxxxxx/v2.0"
  providerConfig: 
      {
      "provider": "azure"
      }
  ...
```
 The `OIDCDiscoveryURL` field - The OIDC Discovery URL, without any .well-known component (base path). Cannot be used with "jwks_url" or "jwt_validation_pubkeys"

 The `OIDCDiscoveryCAPEM` field - The CA certificate or chain of certificates, in PEM format, to use to validate connections to the OIDC Discovery URL. If not set, system certificates are used.

 The `OIDCClientID` field - The OAuth Client ID from the provider for OIDC roles.

 The `OIDCClientSecret` field - The OAuth Client Secret from the provider for OIDC roles.
 The OIDCClientSecret and possibly the OIDCClientID can be retrived a three different ways :

1. From a Kubernetes secret, specifying the `OIDCCredentials` field as follows:
```yaml
  OIDCCredentials:
    secret: 
      name: oidccredentials
    usernameKey: client_id
    passwordKey: client_secret
```
The secret must be of [basic auth type](https://kubernetes.io/docs/concepts/configuration/secret/#basic-authentication-secret). 

Example Secret : 
```bash
kubectl create secret generic oidccredentials --from-literal=oidc_client_id="123456-1234-1234-1234-123456789" --from-literal=oidc_client_secret="saffsfsdfsfsdgsdgsdgsdgghdfhdhdgsjgjgjfj" -n vault-admin
```
If the secret is updated this connection will also be updated.

2. From a [Vault secret](https://developer.hashicorp.com/vault/docs/secrets/kv), specifying the `OIDCCredentials` field as follows :
```yaml
  OIDCCredentials:
    vaultSecret: 
      path: secret/foo
    usernameKey: client_id
    passwordKey: client_secret
```
3. From a [RandomSecret](secret-management.md#RandomSecret), specifying the `OIDCCredentials` field as follows : 
```yaml
  OIDCCredentials:
    randomSecret: 
      name: oidccredentials
    usernameKey: client_id
    passwordKey: client_secret
```
When the RandomSecret generates a new secret, this connection will also be updated.

 The `OIDCResponseMode` field - The response mode to be used in the OAuth2 request. Allowed values are "query" and "form_post". Defaults to "query". If using Vault namespaces, and oidc_response_mode is "form_post", then "namespace_in_state" should be set to false.

 The `OIDCResponseTypes` field - The response types to request. Allowed values are "code" and "id_token". Defaults to "code". Note: "id_token" may only be used if "oidc_response_mode" is set to "form_post".

 The `JWKSURL` field - JWKS URL to use to authenticate signatures. Cannot be used with "oidc_discovery_url" or "jwt_validation_pubkeys".

 The `JWKSCAPEM` field - The CA certificate or chain of certificates, in PEM format, to use to validate connections to the JWKS URL. If not set, system certificates are used.

 The `JWTValidationPubKeys` field - A list of PEM-encoded public keys to use to authenticate signatures locally. Cannot be used with "jwks_url" or "oidc_discovery_url".

 The `boundIssuer` field - The value against which to match the iss claim in a JWT.

 The `JWTSupportedAlgs` field - A list of supported signing algorithms. Defaults to [RS256] for OIDC roles. Defaults to all available algorithms for JWT roles.

 The `defaultRole` field - The default role to use if none is provided during login. 

 The `providerConfig` field - Configuration options for provider-specific handling. Providers with specific handling include: Azure, Google. The options are described in each provider's section in OIDC Provider Setup. 

 The `namespaceInState` field - Pass namespace in the OIDC state parameter instead of as a separate query parameter. With this setting, the allowed redirect URL(s) in Vault and on the provider side should not contain a namespace query parameter. This means only one redirect URL entry needs to be maintained on the provider side for all vault namespaces that will be authenticating against it. Defaults to true for new configs. 


## JWTOIDCAuthEngineRole

The `JWTOIDCAuthEngineRole` CRD allows a user to register a role in an authentication engine mount of type [JWT/OIDC](https://developer.hashicorp.com/vault/api-docs/auth/jwt#create-role).

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: JWTOIDCAuthEngineRole
metadata:
  name: azure-oidc-dev-role
spec:
  authentication: 
    path: kubernetes
    role: policy-admin
  path: oidc-aad/azuread-oidc
  name: dev-role
  userClaim: "email"
  allowedRedirectURIs: 
    - "http://localhost:8250/oidc/callback"
    - "http://localhost:8200/ui/vault/auth/oidc/azuread-oidc/oidc/callback"
  groupsClaim: "groups"
  tokenPolicies: 
    - "dev"
    - "prod"
  roleType: "oidc"
  OIDCScopes: 
    - "https://graph.microsoft.com/.default"

```
 The `name` field - Name of the role

 The `roleType` field - Type of role, either "oidc" (default) or "jwt"

 The `boundAudiences` field - List of aud claims to match against. Any match is sufficient. Required for "jwt" roles, optional for "oidc" roles

 The `userClaim` field - The claim to use to uniquely identify the user; this will be used as the name for the Identity entity alias created due to a successful login. The claim value must be a string

 The `userClaimJSONPointer` field - Specifies if the user_claim value uses JSON pointer syntax for referencing claims. By default, the user_claim value will not use JSON pointer

 The `clockSkewLeeway` field - The amount of leeway to add to all claims to account for clock skew, in seconds. Defaults to 60 seconds if set to 0 and can be disabled if set to -1. Accepts an integer number of seconds, or a Go duration format string. Only applicable with "jwt" roles

 The `expirationLeeway` field - The amount of leeway to add to expiration (exp) claims to account for clock skew, in seconds. Defaults to 150 seconds if set to 0 and can be disabled if set to -1. Accepts an integer number of seconds, or a Go duration format string. Only applicable with "jwt" roles

 The `notBeforeLeeway` field - The amount of leeway to add to not before (nbf) claims to account for clock skew, in seconds. Defaults to 150 seconds if set to 0 and can be disabled if set to -1. Accepts an integer number of seconds, or a Go duration format string. Only applicable with "jwt" roles

 The `boundSubject` field - If set, requires that the sub claim matches this value

 The `boundClaims` field - If set, a map of claims (keys) to match against respective claim values (values). The expected value may be a single string or a list of strings. The interpretation of the bound claim values is configured with bound_claims_type. Keys support JSON pointer syntax for referencing claims

 The `boundClaimsType` field - Configures the interpretation of the bound_claims values. If "string" (the default), the values will treated as string literals and must match exactly. If set to "glob", the values will be interpreted as globs, with * matching any number of characters

 The `groupsClaim` field - The claim to use to uniquely identify the set of groups to which the user belongs; this will be used as the names for the Identity group aliases created due to a successful login. The claim value must be a list of strings. Supports JSON pointer syntax for referencing claims

 The `claimMappings` field - If set, a map of claims (keys) to be copied to specified metadata fields (values). Keys support JSON pointer syntax for referencing claims

 The `OIDCScopes` field - If set, a list of OIDC scopes to be used with an OIDC role. The standard scope "openid" is automatically included and need not be specified

 The `allowedRedirectURIs` field - The list of allowed values for redirect_uri during OIDC logins

 The `verboseOIDCLogging` field - Log received OIDC tokens and claims when debug-level logging is active. Not recommended in production since sensitive information may be present in OIDC responses

 The `maxage` field - Specifies the allowable elapsed time in seconds since the last time the user was actively authenticated with the OIDC provider. If set, the max_age request parameter will be included in the authentication request. See AuthRequest for additional details. Accepts an integer number of seconds, or a Go duration format string

 The `tokenTTL` field - Specifies the allowable elapsed time in seconds since the last time the user was actively authenticated with the OIDC provider. If set, the max_age request parameter will be included in the authentication request. See AuthRequest for additional details. Accepts an integer number of seconds, or a Go duration format string

 The `tokenMaxTTL` field - The maximum lifetime for generated tokens. This current value of this will be referenced at renewal time
 The `tokenPolicies` field - List of policies to encode onto generated tokens. Depending on the auth method, this list may be supplemented by user/group/other values

 The `tokenBoundCIDRs` field - List of CIDR blocks; if set, specifies blocks of IP addresses which can authenticate successfully, and ties the resulting token to these blocks as well

 The `tokenExplicitMaxTTL` field - If set, will encode an explicit max TTL onto the token. This is a hard cap even if token_ttl and token_max_ttl would otherwise allow a renewal

 The `tokenNoDefaultPolicy` field - If set, the default policy will not be set on generated tokens; otherwise it will be added to the policies set in token_policies

 The `tokenNumUses` field - The maximum number of times a generated token may be used (within its lifetime); 0 means unlimited. If you require the token to have the ability to create child tokens, you will need to set this value to 0

 The `tokenPeriod` field - The period, if any, to set on the token

 The `tokenType` field - The type of token that should be generated. Can be service, batch, or default to use the mount's tuned default (which unless changed will be service tokens). For token store roles, there are two additional possibilities: default-service and default-batch which specify the type to return unless the client requests a different type at generation time