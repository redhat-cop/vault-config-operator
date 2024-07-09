# Authentication Engines APIs

- [Authentication Engines APIs](#authentication-engines-apis)
  - [AuthEngineMount](#authenginemount)
  - [KubernetesAuthEngineConfig](#kubernetesauthengineconfig)
  - [KubernetesAuthEngineRole](#kubernetesauthenginerole)
  - [LDAPAuthEngineConfig](#ldapauthengineconfig)
    - [LDAPAuthEngineGroup](#ldapauthenginegroup)
  - [JWTOIDCAuthEngineConfig](#jwtoidcauthengineconfig)
    - [JWTOIDCAuthEngineRole](#jwtoidcauthenginerole)
  - [GCPAuthEngineConfig](#gcpauthengineconfig)
    - [GCPAuthEngineRole](#gcpauthenginerole)
  - [AzureAuthEngineConfig](#azureauthengineconfig)
    - [AzureAuthEngineRole](#azureauthenginerole)

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

The `kubernetesCACert` field is the base64 encoded CA certificate that can be used to validate the connection to the master API. If passed, that CA bundle will be used. Consult the following table to see what happens when the field is not passed

| `kubernetesCACert`    | `disableLocalCAJWT` | `useOperatorPodCA` | Behaviour |
| -------- | ------- | -------- | ------- |
| set | ignored | ignored | the set CA is used |
| unset | false | ignored | the `/var/run/secrets/kubernetes.io/serviceaccount/ca.crt` of the Vault's pod is used. If Vault is not running in a pod, then the behavior is undefined |
| unset | true  | false | the default os CA where Vault is running is used |
| unset | true  | true  | the `/var/run/secrets/kubernetes.io/serviceaccount/ca.crt` the operator pod is inject and used |


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

  In addition, this CRD provides the `tLSConfig` field, where you can specify `certificate`, `clientTLSCert` and `clientTLSKey` by using a TLS Kubernetes secret, as shown below:

```yaml
  tLSConfig:
    tlsSecret:
     name: ldap-tls-certificate
```
  
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
  OIDCClientID: "12345-1234-1234-1234-1234567"
  OIDCCredentials:
    secret: 
      name: oidccredentials
    usernameKey: client_id
    passwordKey: client_secret
  OIDCDiscoveryURL: "https://login.microsoftonline.com/12345-1234-1234-1234-1234567/v2.0"
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
  roleType: "oidc"
  OIDCScopes: 
    - "https://graph.microsoft.com/.default"
  ...
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

## GCPAuthEngineConfig

The `GCPAuthEngineConfig` CRD allows a user to configure an authentication engine mount of type [Google Cloud](https://developer.hashicorp.com/vault/api-docs/auth/gcp).

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: GCPAuthEngineConfig
metadata:
  labels:
    app.kubernetes.io/name: gcpauthengineconfig
    app.kubernetes.io/instance: gcpauthengineconfig-sample
    app.kubernetes.io/part-of: vault-config-operator
    app.kubernetes.io/managed-by: kustomize
    app.kubernetes.io/created-by: vault-config-operator
  name: gcpauthengineconfig-sample
spec:
  authentication:
    path: vault-admin
    role: vault-admin
    serviceAccount:
      name: vault
  connection:
    address: 'https://vault.example.com'
  path: azure
  serviceAccount: ""
  IAMalias: "default"
  IAMmetadata: "default"
  GCEalias: "role_id"
  GCEmetadata: "default"
  customEndpoint: {}
  GCPCredentials:
    secret: 
      name: gcp-serviceaccount-credentials
    usernameKey: serviceaccount
    passwordKey: credentials
```

 The `serviceAccount` field - Service Account Name. A service account is a special kind of account typically used by an application or compute workload, such as a Compute Engine instance, rather than a person. 
 
 A service account is identified by its email address, which is unique to the account. 
 
 Applications use service accounts to make authorized API calls by authenticating as either the service account itself, or as Google Workspace or Cloud Identity users through domain-wide delegation. 
 
 When an application authenticates as a service account, it has access to all resources that the service account has permission to access.

 The `IAMalias` field - Must be either unique_id or role_id. If unique_id is specified, the service account's unique ID will be used for alias names during login. If role_id is specified, the ID of the Vault role will be used. Only used if role type is iam.

 The `IAMmetadata` field - The metadata to include on the token returned by the login endpoint. This metadata will be added to both audit logs, and on the iam_alias. By default, it includes project_id, role, service_account_id, and service_account_email. 
 - To include no metadata, set to "" via the CLI or [] via the API. 
 - To use only particular fields, select the explicit fields. 
 - To restore to defaults, send only a field of default. 
 
 Only select fields that will have a low rate of change for your iam_alias because each change triggers a storage write and can have a performance impact at scale. Only used if role type is iam.

 The `GCEalias` field - Must be either instance_id or role_id. If instance_id is specified, the GCE instance ID will be used for alias names during login. If role_id is specified, the ID of the Vault role will be used. Only used if role type is gce.

 The `GCEmetadata` field - The metadata to include on the token returned by the login endpoint. This metadata will be added to both audit logs, and on the gce_alias. By default, it includes instance_creation_timestamp, instance_id, instance_name, project_id, project_number, role, service_account_id, service_account_email, and zone. 
 - To include no metadata, set to "" via the CLI or [] via the API. 
 - To use only particular fields, select the explicit fields. 
 - To restore to defaults, send only a field of default. 
 
 Only select fields that will have a low rate of change for your gce_alias because each change triggers a storage write and can have a performance impact at scale. Only used if role type is gce.

 The `customEndpoint` field - Specifies overrides to service endpoints used when making API requests. This allows specific requests made during authentication to target alternative service endpoints for use in Private Google Access environments.
 Overrides are set at the subdomain level using the following keys:
 - api - Replaces the service endpoint used in API requests to https://www.googleapis.com.
 - iam - Replaces the service endpoint used in API requests to https://iam.googleapis.com.
 - crm - Replaces the service endpoint used in API requests to https://cloudresourcemanager.googleapis.com.
 - compute - Replaces the service endpoint used in API requests to https://compute.googleapis.com.
 
 The endpoint value provided for a given key has the form of scheme://host:port. The scheme:// and :port portions of the endpoint value are optional.

## GCPAuthEngineRole
 The `GCPAuthEngineRole` CRD allows a user to register a role in an authentication engine mount of type [Google Cloud](https://developer.hashicorp.com/vault/api-docs/auth/gcp#create-update-role).

 ```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: GCPAuthEngineRole
metadata:
  labels:
    app.kubernetes.io/name: gcpauthenginerole
    app.kubernetes.io/instance: gcpauthenginerole-sample
    app.kubernetes.io/part-of: vault-config-operator
    app.kubernetes.io/managed-by: kustomize
    app.kubernetes.io/created-by: vault-config-operator
  name: gcpauthenginerole-sample
spec:
  authentication:
    path: vault-admin
    role: vault-admin
    serviceAccount:
      name: vault
  connection:
    address: 'https://vault.example.com'
  path: gcp
  name: "gcp-iam-role"
  type: "iam"
  boundServiceAccounts: {}
  boundProjects: {}
  addGroupAliases: false
  tokenTTL: ""
  tokenMaxTTL: ""
  tokenPolicies: []
  policies: []
  tokenBoundCIDRs: []
  tokenExplicitMaxTTL: ""
  tokenNoDefaultPolicy: false
  tokenNumUses: 0
  tokenPeriod: 0
  tokenType: ""
  maxJWTExp: ""
  allowGCEInference: false
  boundZones: []
  boundRegions: []
  boundInstanceGroups: []
  boundLabels: []
 ```

The `name` field - The name of the role.

The `type` field - The type of this role. Certain fields correspond to specific roles and will be rejected otherwise. Please see below for more information.

The `bound_service_accounts` field - An array of service account emails or IDs that login is restricted to, either directly or through an associated instance. If set to *, all service accounts are allowed (you can bind this further using bound_projects.)

The `bound_projects` field - An array of GCP project IDs. Only entities belonging to this project can authenticate under the role.

The `add_group_aliases` field -  If true, any auth token generated under this token will have associated group aliases, namely project-$PROJECT_ID, folder-$PROJECT_ID, and organization-$ORG_ID for the entities project and all its folder or organization ancestors. This requires Vault to have IAM permission resourcemanager.projects.get.

The `token_ttl` field - The incremental lifetime for generated tokens. This current value of this will be referenced at renewal time.

The `token_max_ttl` field -  The maximum lifetime for generated tokens. This current value of this will be referenced at renewal time.

The `token_policies` field - List of token policies to encode onto generated tokens. Depending on the auth method, this list may be supplemented by user/group/other values.

The `policies` field - DEPRECATED: Please use the token_policies parameter instead. List of token policies to encode onto generated tokens. Depending on the auth method, this list may be supplemented by user/group/other values.

The `token_bound_cidrs` field - List of CIDR blocks; if set, specifies blocks of IP addresses which can authenticate successfully, and ties the resulting token to these blocks as well.

The `token_explicit_max_ttl` field - If set, will encode an explicit max TTL onto the token. This is a hard cap even if token_ttl and token_max_ttl would otherwise allow a renewal.

The `token_no_default_policy` field - If set, the default policy will not be set on generated tokens; otherwise it will be added to the policies set in token_policies.

The `token_num_uses` field - The maximum number of times a generated token may be used (within its lifetime); 0 means unlimited. If you require the token to have the ability to create child tokens, you will need to set this value to 0.

The `token_period` field - The maximum allowed period value when a periodic token is requested from this role.

The `token_type` field - The type of token that should be generated. Can be service, batch, or default to use the mount's tuned default (which unless changed will be service tokens). For token store roles, there are two additional possibilities: default-service and default-batch which specify the type to return unless the client requests a different type at generation time. For machine based authentication cases, you should use batch type tokens.

#### The following parameters are only valid when the role is of type "iam":

The `max_jwt_exp` field - The number of seconds past the time of authentication that the login param JWT must expire within. For example, if a user attempts to login with a token that expires within an hour and this is set to 15 minutes, Vault will return an error prompting the user to create a new signed JWT with a shorter exp. The GCE metadata tokens currently do not allow the exp claim to be customized.

The `allow_gce_inference` field - A flag to determine if this role should allow GCE instances to authenticate by inferring service accounts from the GCE identity metadata token.

#### The following parameters are only valid when the role is of type "gce":

The `bound_zones` field - The list of zones that a GCE instance must belong to in order to be authenticated. If bound_instance_groups is provided, it is assumed to be a zonal group and the group must belong to this zone.

The `bound_regions` field - The list of regions that a GCE instance must belong to in order to be authenticated. If bound_instance_groups is provided, it is assumed to be a regional group and the group must belong to this region. If bound_zones are provided, this attribute is ignored.

The `bound_instance_groups` field - The instance groups that an authorized instance must belong to in order to be authenticated. If specified, either bound_zones or bound_regions must be set too.

The `bound_labels` field - A comma-separated list of GCP labels formatted as "key:value" strings that must be set on authorized GCE instances. Because GCP labels are not currently ACL'd, we recommend that this be used in conjunction with other restrictions.

## AzureAuthEngineConfig

The `AzureAuthEngineConfig` CRD allows a user to configure an authentication engine mount of type [Azure](https://developer.hashicorp.com/vault/api-docs/auth/azure).

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: AzureAuthEngineConfig
metadata:
  labels:
    app.kubernetes.io/name: azureauthengineconfig
    app.kubernetes.io/instance: azureauthengineconfig-sample
    app.kubernetes.io/part-of: vault-config-operator
    app.kubernetes.io/managed-by: kustomize
    app.kubernetes.io/created-by: vault-config-operator
  name: azureauthengineconfig-sample
spec:
  authentication:
    path: vault-admin
    role: vault-admin
    serviceAccount:
      name: vault
  connection:
    address: 'https://vault.example.com'
  path: azure
  resource: "https://management.azure.com/"
  environment: "AzurePublicCloud"
  maxRetries: 3
  maxRetryDelay: 60
  retryDelay: 4
  tenantID: "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxx"
  clientID: "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxx"
  azureCredentials:
    secret: 
      name: aad-credentials
    usernameKey: client_id
    passwordKey: client_secret
```

 The `tenant_id` field - The tenant id for the Azure Active Directory organization. This value can also be provided with the AZURE_TENANT_ID environment variable.

 The `resource` field - The resource URL for the application registered in Azure Active Directory. 
 The value is expected to match the audience (aud claim) of the JWT provided to the login API. 
 See the resource parameter for how the audience is set when requesting a JWT access token from the Azure Instance Metadata Service (IMDS) endpoint. This value can also be provided with the AZURE_AD_RESOURCE environment variable.

 The `environment` field - The Azure cloud environment. Valid values: AzurePublicCloud, AzureUSGovernmentCloud, AzureChinaCloud, AzureGermanCloud. This value can also be provided with the AZURE_ENVIRONMENT environment variable.

 The `client_id` field - The client id for credentials to query the Azure APIs. Currently read permissions to query compute resources are required. This value can also be provided with the AZURE_CLIENT_ID environment variable.

 The `client_secret` field - The client secret for credentials to query the Azure APIs. This value can also be provided with the AZURE_CLIENT_SECRET environment variable.

 The `max_retries` field - The maximum number of attempts a failed operation will be retried before producing an error.

 The `max_retry_delay` field - The maximum delay, in seconds, allowed before retrying an operation.

 The `retry_delay` field - The initial amount of delay, in seconds, to use before retrying an operation. Increases exponentially.

 The `azureCredentials` field - The OAuth Client Secret from the provider for OIDC roles.
 The OIDCClientSecret and possibly the OIDCClientID can be retrived a three different ways :

1. From a Kubernetes secret, specifying the `azureCredentials` field as follows:
```yaml
  azureCredentials:
    secret: 
      name: aad-credentials
    usernameKey: clientid
    passwordKey: clientsecret
```
The secret must be of [basic auth type](https://kubernetes.io/docs/concepts/configuration/secret/#basic-authentication-secret). 

Example Secret : 
```bash
kubectl create secret generic aad-credentials --from-literal=clientid="123456-1234-1234-1234-123456789" --from-literal=clientsecret="saffsfsdfsfsdgsdgsdgsdgghdfhdhdgsjgjgjfj" -n vault-admin
```
If the secret is updated this connection will also be updated.

2. From a [Vault secret](https://developer.hashicorp.com/vault/docs/secrets/kv), specifying the `azureCredentials` field as follows :
```yaml
  azureCredentials:
    vaultSecret: 
      path: secret/foo
    usernameKey: clientid
    passwordKey: clientsecret
```
3. From a [RandomSecret](secret-management.md#RandomSecret), specifying the `azureCredentials` field as follows : 
```yaml
  azureCredentials:
    randomSecret: 
      name: aad-credentials
    usernameKey: clientid
    passwordKey: clientsecret
```
When the RandomSecret generates a new secret, this connection will also be updated.

## AzureAuthEngineRole
 The `AzureAuthEngineRole` CRD allows a user to register a role in an authentication engine mount of type [Azure](https://developer.hashicorp.com/vault/api-docs/auth/azure#create-update-role).

 ```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: AzureAuthEngineRole
metadata:
  labels:
    app.kubernetes.io/name: azureauthenginerole
    app.kubernetes.io/instance: azureauthenginerole-sample
    app.kubernetes.io/part-of: vault-config-operator
    app.kubernetes.io/managed-by: kustomize
    app.kubernetes.io/created-by: vault-config-operator
  name: azureauthenginerole-sample
spec:
  authentication:
    path: vault-admin
    role: vault-admin
    serviceAccount:
      name: vault
  connection:
    address: 'https://vault.example.com'
  path: azure
  name: dev-role
  boundServicePrincipalIDs:
    - sp1
    - sp2
  boundGroupIDs:
    - group1
    - group2
  boundLocations:
    - location1
    - location2
  boundSubscriptionIDs:
    - subscription1  
    - subscription2
  BoundResourceGroups:
    - resourcegroup1
    - resourcegroup2
  boundScaleSets:
    - scaleset1
    - scaleset1
  tokenTTL: ""
  tokenMaxTTL: ""
  tokenPolicies:
    - policy1
    - policy2
  policies:
    - policy1
    - policy2
  tokenBoundCIDRs:
    - CIDR1
    - CIDR2
  tokenExplicitMaxTTL: ""
  tokenNoDefaultPolicy: false
  tokenNumUses: 0
  tokenPeriod: 0
  tokenType: ""
 ```

  The `name` field - Name of the role

  The `bound_service_principal_ids` field - The list of Service Principal IDs that login is restricted to.

  The `bound_group_ids` field - The list of group ids that login is restricted to.

  The `bound_locations` field - The list of locations that login is restricted to.

  The `bound_subscription_ids` field - The list of subscription IDs that login is restricted to.

  The `bound_resource_groups` field - The list of resource groups that login is restricted to.

  The `bound_scale_sets` field - The list of scale set names that the login is restricted to.

  The `token_ttl` field - The incremental lifetime for generated tokens. This current value of this will be referenced at renewal time.

  The `token_max_ttl` field - The maximum lifetime for generated tokens. This current value of this will be referenced at renewal time.

  The `token_policies` field - List of token policies to encode onto generated tokens. Depending on the auth method, this list may be supplemented by user/group/other values.

  The `policies` field - DEPRECATED: Please use the token_policies parameter instead. List of token policies to encode onto generated tokens. Depending on the auth method, this list may be supplemented by user/group/other values.

  The `token_bound_cidrs` field - List of CIDR blocks; if set, specifies blocks of IP addresses which can authenticate successfully, and ties the resulting token to these blocks as well.

  The `token_explicit_max_ttl` field - If set, will encode an explicit max TTL onto the token. This is a hard cap even if token_ttl and token_max_ttl would otherwise allow a renewal.

  The `token_no_default_policy` field - If set, the default policy will not be set on generated tokens; otherwise it will be added to the policies set in token_policies.

  The `token_num_uses` field - The maximum number of times a generated token may be used (within its lifetime); 0 means unlimited. If you require the token to have the ability to create child tokens, you will need to set this value to 0.

  The `token_period` field - The maximum allowed period value when a periodic token is requested from this role.
  
  The `token_type` field - The type of token that should be generated. Can be service, batch, or default to use the mount's tuned default (which unless changed will be service tokens). For token store roles, there are two additional possibilities: default-service and default-batch which specify the type to return unless the client requests a different type at generation time. For machine based authentication cases, you should use batch type tokens.

