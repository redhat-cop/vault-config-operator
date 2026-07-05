# PKI Secret Engine

[PKI engine documentation](https://developer.hashicorp.com/vault/docs/secrets/pki)

## Overview

The PKI secret engine generates X.509 certificates dynamically, eliminating the manual process of generating private keys and CSRs, submitting to a CA, and distributing signed certificates. It supports both root and intermediate certificate authorities, enabling a complete PKI hierarchy managed declaratively through Kubernetes CRDs.

The vault-config-operator supports the following CRDs for the PKI engine:

- [PKISecretEngineConfig](#pkisecretengineconfig)
- [PKISecretEngineRole](#pkisecretenginerole)

## PKISecretEngineConfig

The `PKISecretEngineConfig` CRD allows you to configure a [PKI secret engine](https://developer.hashicorp.com/vault/api-docs/secret/pki#generate-root) root or intermediate certificate authority.

### Example

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: PKISecretEngineConfig
metadata:
  name: my-pki
spec:
  authentication:
    path: kubernetes
    role: pki-engine-admin
  path: pki-vault-demo/pki
  type: root
  privateKeyType: internal
  commonName: pki-vault-demo.internal.io
  TTL: "8760h"
  format: pem
  keyType: rsa
  keyBits: 4096
  maxPathLength: -1
  organization: "My Company"
  country: "US"
  province: "California"
  locality: "San Francisco"
  issuingCertificates:
    - "https://vault.example.com/v1/pki-vault-demo/pki/ca"
  CRLDistributionPoints:
    - "https://vault.example.com/v1/pki-vault-demo/pki/crl"
  CRLExpiry: "72h"
```

> **`spec.authentication.path` vs `spec.path`:** `spec.authentication.path` is the auth mount the operator itself uses to authenticate with Vault. `spec.path` is the mount path of the PKI secret engine being configured. They may point to different mounts.

### Vault CLI Equivalent

```shell
vault write [namespace/]<path>/<type>/generate/<privateKeyType> \
    common_name="pki-vault-demo.internal.io" \
    ttl="8760h" \
    format=pem \
    key_type=rsa \
    key_bits=4096 \
    max_path_length=-1 \
    organization="My Company" \
    country="US" \
    province="California" \
    locality="San Francisco"

vault write [namespace/]<path>/config/urls \
    issuing_certificates="https://vault.example.com/v1/pki-vault-demo/pki/ca" \
    crl_distribution_points="https://vault.example.com/v1/pki-vault-demo/pki/crl"

vault write [namespace/]<path>/config/crl \
    expiry="72h"
```

> **Note:** The operator writes to three separate Vault paths during reconciliation: the generate endpoint for the root/intermediate CA, `config/urls` for URL configuration, and `config/crl` for CRL configuration.

### Field Descriptions

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| path | string | Yes | Mount path of the PKI secret engine. The operator writes to multiple Vault paths: `[namespace/]{path}/{type}/generate/{privateKeyType}`, `{path}/config/urls`, and `{path}/config/crl` |
| authentication | object | Yes | Kubernetes auth configuration. See [Authentication](../auth-section.md) |
| connection | object | No | Override Vault connection settings. See [Vault Connection](../contributing-vault-apis.md) |
| type | string | Yes | Type of certificate authority: `root` or `intermediate`. Defaults to `root` |
| privateKeyType | string | Yes | Key generation mode: `internal` (key stays in Vault) or `exported` (key returned in response). Defaults to `internal` |
| commonName | string | Yes | Requested CN for the certificate |
| altNames | string | No | Subject Alternative Names (hostnames or emails), comma-delimited |
| IPSans | string | No | IP Subject Alternative Names, comma-delimited |
| URISans | string | No | URI Subject Alternative Names, comma-delimited |
| otherSans | string | No | Custom OID/UTF8-string SANs in OpenSSL format: `<oid>;<type>:<value>` |
| TTL | duration | No | Requested Time To Live for the certificate |
| format | string | No | Output format: `pem`, `pem_bundle`, or `der`. Defaults to `pem` |
| privateKeyFormat | string | No | Private key marshaling format: `der` or `pkcs8` |
| keyType | string | No | Key algorithm: `rsa` or `ec`. Defaults to `rsa` |
| keyBits | int | No | Key size in bits (e.g., `2048`, `4096` for RSA; `224`, `256`, `384`, `521` for EC). Defaults to `2048` |
| maxPathLength | int | No | Maximum path length in the certificate. `-1` means no limit, `0` means literal zero. Defaults to `-1` |
| excludeCnFromSans | bool | No | If `true`, the CN is not included in DNS or Email SANs. Defaults to `false` |
| permittedDnsDomains | []string | No | DNS domains for which certificates are allowed to be issued by this CA |
| ou | string | No | OrganizationalUnit value in the certificate subject |
| organization | string | No | Organization value in the certificate subject |
| country | string | No | Country value in the certificate subject |
| locality | string | No | Locality value in the certificate subject |
| province | string | No | Province/State value in the certificate subject |
| streetAddress | string | No | Street Address value in the certificate subject |
| postalCode | string | No | Postal Code value in the certificate subject |
| serialNumber | string | No | Serial number. If not set, Vault generates a random one |
| issuingCertificates | []string | No | URLs for the Issuing Certificate field (written to `{path}/config/urls`) |
| CRLDistributionPoints | []string | No | URLs for CRL Distribution Points (written to `{path}/config/urls`) |
| ocspServers | []string | No | URLs for OCSP Servers (written to `{path}/config/urls`) |
| CRLExpiry | duration | No | CRL expiration time. Defaults to `72h` (written to `{path}/config/crl`) |
| CRLDisable | bool | No | Disable CRL building. Defaults to `false` (written to `{path}/config/crl`) |
| externalSignSecret | object | No | Kubernetes Secret containing the externally-signed intermediate certificate. See [Intermediate CA](#intermediate-ca) |
| certificateKey | string | No | Key name in the `externalSignSecret` to read the signed certificate from. Defaults to `tls.crt` |
| internalSign | object | No | Vault mount path of the signing CA. The `name` field is used directly as the Vault path prefix (e.g., `name: my-root-pki` signs via `my-root-pki/root/sign-intermediate`). See [Intermediate CA](#intermediate-ca) |

> **Note:** Deleting the `PKISecretEngineConfig` CR does **not** remove the root/intermediate CA from Vault. The configuration can only be removed by deleting the entire PKI engine mount.

### Intermediate CA

When `type` is set to `intermediate`, the operator generates a CSR instead of a root CA. The CSR must then be signed by a parent CA. The operator supports two workflows for signing the intermediate certificate:

**Internal Signing** — The CSR is signed by another `PKISecretEngineConfig` already managed by the operator:

```yaml
spec:
  type: intermediate
  privateKeyType: internal
  commonName: intermediate.example.com
  internalSign:
    name: my-root-pki
```

The operator automatically submits the CSR to `{internalSign.name}/root/sign-intermediate` (where `internalSign.name` is the Vault mount path of the signing CA) and sets the signed certificate at `{path}/intermediate/set-signed`.

**External Signing** — The CSR is signed by an external CA and the signed certificate is provided via a Kubernetes Secret:

```yaml
spec:
  type: intermediate
  privateKeyType: exported
  commonName: intermediate.example.com
  externalSignSecret:
    name: signed-intermediate-cert
  certificateKey: tls.crt
```

When `privateKeyType` is `exported`, the private key and CSR are stored in a Kubernetes Secret with the same name as the CR. The operator reads the signed certificate from the referenced Secret (key: `certificateKey`, default: `tls.crt`) and writes it to `{path}/intermediate/set-signed`.

## PKISecretEngineRole

The `PKISecretEngineRole` CRD allows you to create a [PKI secret engine role](https://developer.hashicorp.com/vault/api-docs/secret/pki#create-update-role) for issuing certificates.

### Example

```yaml
apiVersion: redhatcop.redhat.io/v1alpha1
kind: PKISecretEngineRole
metadata:
  name: web-server
spec:
  authentication:
    path: kubernetes
    role: pki-engine-admin
  path: pki-vault-demo/pki
  allowedDomains:
    - internal.io
    - pki-vault-demo.svc
  allowSubdomains: true
  allowBareDomains: false
  TTL: "24h"
  maxTTL: "720h"
  keyType: rsa
  keyBits: 2048
  keyUsage:
    - DigitalSignature
    - KeyEncipherment
  extKeyUsage:
    - ServerAuth
  useCSRCommonName: true
  useCSRSans: true
  requireCn: true
  notBeforeDuration: "30s"
```

### Vault CLI Equivalent

```shell
vault write [namespace/]<path>/roles/<role-name> \
    allowed_domains="internal.io,pki-vault-demo.svc" \
    allow_subdomains=true \
    allow_bare_domains=false \
    ttl="24h" \
    max_ttl="720h" \
    key_type=rsa \
    key_bits=2048 \
    key_usage="DigitalSignature,KeyEncipherment" \
    ext_key_usage="ServerAuth" \
    use_csr_common_name=true \
    use_csr_sans=true \
    require_cn=true \
    not_before_duration="30s"
```

### Field Descriptions

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| path | string | Yes | Mount path of the PKI secret engine. Full Vault path: `[namespace/]{path}/roles/{name}` |
| authentication | object | Yes | Kubernetes auth configuration. See [Authentication](../auth-section.md) |
| connection | object | No | Override Vault connection settings. See [Vault Connection](../contributing-vault-apis.md) |
| name | string | No | Override the Vault object name. Defaults to `metadata.name` |
| TTL | duration | No | Default TTL for issued certificates. Uses system default or `maxTTL`, whichever is shorter |
| maxTTL | duration | No | Maximum TTL for issued certificates. Defaults to system maximum lease TTL |
| allowLocalhost | bool | No | Allow `localhost` as a valid CN/SAN. Defaults to `false` |
| allowedDomains | []string | No | Domains allowed for this role. Used with `allowBareDomains` and `allowSubdomains` |
| allowedDomainsTemplate | bool | No | Allow ACL path templating in `allowedDomains`. Defaults to `false` |
| allowBareDomains | bool | No | Allow clients to request certificates matching the exact domain from `allowedDomains`. Defaults to `false` |
| allowSubdomains | bool | No | Allow certificates for subdomains of `allowedDomains`, including wildcards. Defaults to `false` |
| allowGlobDomains | bool | No | Allow glob patterns in `allowedDomains` (e.g., `ftp*.example.com`). Defaults to `false` |
| allowAnyName | bool | No | Allow any CN to be requested. Defaults to `false` |
| enforceHostnames | bool | No | Only allow valid hostnames for CNs and DNS SANs. Defaults to `false` |
| allowIPSans | bool | No | Allow IP Subject Alternative Names. Defaults to `false` |
| allowedURISans | []string | No | Allowed URI SANs. Values can contain glob patterns (e.g., `spiffe://hostname/*`) |
| allowedOtherSans | string | No | Allowed custom OID/UTF8-string SANs. Use `*` to allow any |
| serverFlag | bool | No | Flag certificates for server use. Defaults to `false` |
| clientFlag | bool | No | Flag certificates for client use. Defaults to `false` |
| codeSigningFlag | bool | No | Flag certificates for code signing. Defaults to `false` |
| emailProtectionFlag | bool | No | Flag certificates for email protection. Defaults to `false` |
| keyType | string | No | Key type for generated keys: `rsa`, `ec`, or `any` (CSR signing only). Defaults to `rsa` |
| keyBits | int | No | Key size in bits. Defaults to `2048` |
| keyUsage | []KeyUsage | No | Allowed key usage constraints. Values: `DigitalSignature`, `KeyAgreement`, `KeyEncipherment`, `ContentCommitment`, `DataEncipherment`, `CertSign`, `CRLSign`, `EncipherOnly`, `DecipherOnly` |
| extKeyUsage | []ExtKeyUsage | No | Allowed extended key usage constraints. Values: `ServerAuth`, `ClientAuth`, `CodeSigning`, `EmailProtection`, `IPSECEndSystem`, `IPSECTunnel`, `IPSECUser`, `TimeStamping`, `OCSPSigning`, `MicrosoftServerGatedCrypto`, `NetscapeServerGatedCrypto`, `MicrosoftCommercialCodeSigning`, `MicrosoftKernelCodeSigning` |
| extKeyUsageOids | []string | No | Extended key usage OIDs, comma-separated |
| useCSRCommonName | bool | No | Use the CN from a submitted CSR instead of the JSON data. Defaults to `true` |
| useCSRSans | bool | No | Use SANs from a submitted CSR instead of the JSON data. Defaults to `true` |
| ou | string | No | OrganizationalUnit in issued certificate subjects |
| organization | string | No | Organization in issued certificate subjects |
| country | string | No | Country in issued certificate subjects |
| locality | string | No | Locality in issued certificate subjects |
| province | string | No | Province/State in issued certificate subjects |
| streetAddress | string | No | Street Address in issued certificate subjects |
| postalCode | string | No | Postal Code in issued certificate subjects |
| serialNumber | string | No | Serial number. If not set, Vault generates a random one |
| generateLease | bool | No | Attach Vault leases to issued certificates (enables revocation via `vault revoke`). Defaults to `false` |
| noStore | bool | No | Do not store issued certificates in Vault storage. Improves performance but prevents enumeration and revocation. Defaults to `false` |
| requireCn | bool | No | Require a CN when generating certificates. Defaults to `false` |
| policyIdentifiers | []string | No | Policy OIDs to include in issued certificates |
| basicConstraintsValidForNonCa | bool | No | Mark Basic Constraints valid when issuing non-CA certificates. Defaults to `false` |
| notBeforeDuration | duration | No | Duration to backdate the NotBefore property. Defaults to `30s` |

## See Also

- [Authentication](../auth-section.md) — Common authentication section configuration
- [Contributing a New Vault API](../contributing-vault-apis.md) — Developer guide for adding new CRD types
- [Vault PKI Secret Engine](https://developer.hashicorp.com/vault/docs/secrets/pki) — Vault documentation
- [Vault PKI Secret Engine API](https://developer.hashicorp.com/vault/api-docs/secret/pki) — Vault API reference
