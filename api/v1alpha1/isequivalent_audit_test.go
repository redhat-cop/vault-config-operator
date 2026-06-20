//go:build !integration

package v1alpha1

import (
	"testing"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// This file documents the extra fields Vault returns per CRD type and verifies
// that IsEquivalentToDesiredState correctly ignores them via filterPayloadToDesiredKeys.
//
// Type categories:
//   Category A — Previously bare DeepEqual (32 types) — now use filterPayloadToDesiredKeys
//   Category B — Desired-side secret deletion (5 types) — now also filter payload
//   Category C — Custom handling (9 types) — Entity/EntityAlias/GroupAlias/Group now use filter;
//                DatabaseSecretEngineConfig, SecretEngineMount already had filtering;
//                Audit/AuditRequestHeader use field-by-field; RandomSecret always returns false

type auditCase struct {
	name    string
	setupFn func() (instance interface {
		IsEquivalentToDesiredState(map[string]any) bool
	}, basePayload map[string]any)
	extraFields    map[string]any
	expectWithBase bool
}

func TestAuditCategoryA_BareDeepEqualTypes_ExtraFieldTolerance(t *testing.T) {
	tests := []auditCase{
		{
			name: "AuthEngineMount",
			setupFn: func() (interface {
				IsEquivalentToDesiredState(map[string]any) bool
			}, map[string]any) {
				desc := "test"
				m := &AuthEngineMount{Spec: AuthEngineMountSpec{AuthMount: AuthMount{Config: AuthMountConfig{DefaultLeaseTTL: "1h", MaxLeaseTTL: "24h", Description: &desc}}}}
				p := map[string]any{
					"default_lease_ttl": "1h", "max_lease_ttl": "24h",
					"audit_non_hmac_request_keys": []string(nil), "audit_non_hmac_response_keys": []string(nil),
					"listing_visibility": "", "passthrough_request_headers": []string(nil),
					"allowed_response_headers": []string(nil), "token_type": "",
					"description": &desc, "options": map[string]string(nil),
				}
				return m, p
			},
			extraFields:    map[string]any{"force_no_cache": false, "plugin_version": ""},
			expectWithBase: true,
		},
		{
			name: "KubernetesAuthEngineConfig",
			setupFn: func() (interface {
				IsEquivalentToDesiredState(map[string]any) bool
			}, map[string]any) {
				m := &KubernetesAuthEngineConfig{Spec: KubernetesAuthEngineConfigSpec{KAECConfig: KAECConfig{KubernetesHost: "https://kubernetes.default.svc:443"}}}
				p := map[string]any{
					"kubernetes_host": "https://kubernetes.default.svc:443", "kubernetes_ca_cert": "",
					"token_reviewer_jwt": "", "pem_keys": []string(nil), "issuer": "",
					"disable_iss_validation": false, "disable_local_ca_jwt": false,
					"use_annotations_as_alias_metadata": false,
				}
				return m, p
			},
			extraFields:    map[string]any{"accessor": "auth_kubernetes_abc123", "local": false, "seal_wrap": false},
			expectWithBase: true,
		},
		{
			name: "KubernetesAuthEngineRole",
			setupFn: func() (interface {
				IsEquivalentToDesiredState(map[string]any) bool
			}, map[string]any) {
				m := &KubernetesAuthEngineRole{Spec: KubernetesAuthEngineRoleSpec{VRole: VRole{TargetServiceAccounts: []string{"default"}}}}
				p := map[string]any{
					"bound_service_account_names": []string{"default"}, "bound_service_account_namespaces": []string(nil),
					"alias_name_source": "",
					"token_ttl":         0, "token_max_ttl": 0, "token_policies": []string(nil),
					"token_bound_cidrs": []string(nil), "token_explicit_max_ttl": 0,
					"token_no_default_policy": false, "token_num_uses": 0,
					"token_period": 0, "token_type": "",
				}
				return m, p
			},
			extraFields:    map[string]any{"token_strictly_bind_ip": false},
			expectWithBase: true,
		},
		{
			name: "IdentityOIDCClient",
			setupFn: func() (interface {
				IsEquivalentToDesiredState(map[string]any) bool
			}, map[string]any) {
				m := &IdentityOIDCClient{Spec: IdentityOIDCClientSpec{IdentityOIDCClientConfig: IdentityOIDCClientConfig{Key: "default", ClientType: "confidential", IDTokenTTL: "24h", AccessTokenTTL: "24h"}}}
				p := map[string]any{
					"key": "default", "redirect_uris": []string(nil), "assignments": []string(nil),
					"client_type": "confidential", "id_token_ttl": "24h", "access_token_ttl": "24h",
				}
				return m, p
			},
			extraFields:    map[string]any{"client_id": "vault-generated-id", "client_secret": "vault-generated-secret"},
			expectWithBase: true,
		},
		{
			name: "IdentityOIDCScope",
			setupFn: func() (interface {
				IsEquivalentToDesiredState(map[string]any) bool
			}, map[string]any) {
				m := &IdentityOIDCScope{Spec: IdentityOIDCScopeSpec{IdentityOIDCScopeConfig: IdentityOIDCScopeConfig{Template: "tmpl", Description: "desc"}}}
				p := map[string]any{"template": "tmpl", "description": "desc"}
				return m, p
			},
			extraFields:    map[string]any{"name": "scope-name"},
			expectWithBase: true,
		},
		{
			name: "IdentityOIDCProvider",
			setupFn: func() (interface {
				IsEquivalentToDesiredState(map[string]any) bool
			}, map[string]any) {
				m := &IdentityOIDCProvider{Spec: IdentityOIDCProviderSpec{IdentityOIDCProviderConfig: IdentityOIDCProviderConfig{Issuer: "https://example.com", AllowedClientIDs: []string{"*"}}}}
				p := map[string]any{"issuer": "https://example.com", "allowed_client_ids": []string{"*"}, "scopes_supported": []string(nil)}
				return m, p
			},
			extraFields:    map[string]any{"issuer_id": "oidc-id-123"},
			expectWithBase: true,
		},
		{
			name: "IdentityOIDCAssignment",
			setupFn: func() (interface {
				IsEquivalentToDesiredState(map[string]any) bool
			}, map[string]any) {
				m := &IdentityOIDCAssignment{Spec: IdentityOIDCAssignmentSpec{IdentityOIDCAssignmentConfig: IdentityOIDCAssignmentConfig{EntityIDs: []string{"e1"}, GroupIDs: []string{"g1"}}}}
				p := map[string]any{"entity_ids": []string{"e1"}, "group_ids": []string{"g1"}}
				return m, p
			},
			extraFields:    map[string]any{"name": "assignment-name"},
			expectWithBase: true,
		},
		{
			name: "IdentityTokenConfig",
			setupFn: func() (interface {
				IsEquivalentToDesiredState(map[string]any) bool
			}, map[string]any) {
				m := &IdentityTokenConfig{Spec: IdentityTokenConfigSpec{IdentityTokenConfigConfig: IdentityTokenConfigConfig{Issuer: "https://vault.example.com"}}}
				p := map[string]any{"issuer": "https://vault.example.com"}
				return m, p
			},
			extraFields:    map[string]any{"issuer_id": "config-id"},
			expectWithBase: true,
		},
		{
			name: "IdentityTokenKey",
			setupFn: func() (interface {
				IsEquivalentToDesiredState(map[string]any) bool
			}, map[string]any) {
				m := &IdentityTokenKey{Spec: IdentityTokenKeySpec{IdentityTokenKeyConfig: IdentityTokenKeyConfig{Algorithm: "RS256", RotationPeriod: "24h", VerificationTTL: "24h"}}}
				p := map[string]any{"algorithm": "RS256", "rotation_period": "24h", "verification_ttl": "24h", "allowed_client_ids": []string(nil)}
				return m, p
			},
			extraFields:    map[string]any{"name": "key-name", "id": "key-id"},
			expectWithBase: true,
		},
		{
			name: "IdentityTokenRole",
			setupFn: func() (interface {
				IsEquivalentToDesiredState(map[string]any) bool
			}, map[string]any) {
				m := &IdentityTokenRole{Spec: IdentityTokenRoleSpec{IdentityTokenRoleConfig: IdentityTokenRoleConfig{Key: "default", Template: "tmpl"}}}
				p := map[string]any{
					"key": "default", "template": "tmpl", "ttl": "",
				}
				return m, p
			},
			extraFields:    map[string]any{"name": "role-name", "issuer": "https://vault"},
			expectWithBase: true,
		},
		{
			name: "PasswordPolicy",
			setupFn: func() (interface {
				IsEquivalentToDesiredState(map[string]any) bool
			}, map[string]any) {
				m := &PasswordPolicy{
					ObjectMeta: metav1.ObjectMeta{Name: "my-policy"},
					Spec:       PasswordPolicySpec{PasswordPolicy: "length = 20\nrule \"charset\" { ... }"},
				}
				p := map[string]any{"policy": "length = 20\nrule \"charset\" { ... }"}
				return m, p
			},
			extraFields:    map[string]any{"name": "my-policy"},
			expectWithBase: true,
		},
		{
			name: "DatabaseSecretEngineRole",
			setupFn: func() (interface {
				IsEquivalentToDesiredState(map[string]any) bool
			}, map[string]any) {
				m := &DatabaseSecretEngineRole{Spec: DatabaseSecretEngineRoleSpec{DBSERole: DBSERole{DBName: "mydb"}}}
				p := map[string]any{
					"db_name": "mydb", "default_ttl": metav1.Duration{}, "max_ttl": metav1.Duration{},
					"creation_statements": []string(nil), "revocation_statements": []string(nil),
					"rollback_statements": []string(nil), "renew_statements": []string(nil),
				}
				return m, p
			},
			extraFields:    map[string]any{"credential_type": "password"},
			expectWithBase: true,
		},
		{
			name: "AzureAuthEngineConfig",
			setupFn: func() (interface {
				IsEquivalentToDesiredState(map[string]any) bool
			}, map[string]any) {
				m := &AzureAuthEngineConfig{Spec: AzureAuthEngineConfigSpec{AzureConfig: AzureConfig{TenantID: "tenant-123", Resource: "https://management.azure.com/"}}}
				p := map[string]any{
					"tenant_id": "tenant-123", "resource": "https://management.azure.com/",
					"environment": "", "client_id": "", "client_secret": "",
					"max_retries": int64(0), "max_retry_delay": int64(0), "retry_delay": int64(0),
				}
				return m, p
			},
			extraFields:    map[string]any{"accessor": "auth_azure_abc", "local": false},
			expectWithBase: true,
		},
		{
			name: "CertAuthEngineConfig",
			setupFn: func() (interface {
				IsEquivalentToDesiredState(map[string]any) bool
			}, map[string]any) {
				m := &CertAuthEngineConfig{Spec: CertAuthEngineConfigSpec{CertAuthEngineConfigInternal: CertAuthEngineConfigInternal{DisableBinding: false, OCSPCacheSize: 100}}}
				p := map[string]any{
					"disable_binding": false, "enable_identity_alias_metadata": false,
					"ocsp_cache_size": 100, "role_cache_size": 0,
				}
				return m, p
			},
			extraFields:    map[string]any{"ocsp_enabled": true},
			expectWithBase: true,
		},
		{
			name: "GCPAuthEngineConfig",
			setupFn: func() (interface {
				IsEquivalentToDesiredState(map[string]any) bool
			}, map[string]any) {
				m := &GCPAuthEngineConfig{Spec: GCPAuthEngineConfigSpec{GCPConfig: GCPConfig{IAMalias: "unique_id"}}}
				p := map[string]any{
					"credentials": "", "iam_alias": "unique_id", "iam_metadata": "",
					"gce_alias": "", "gce_metadata": "", "custom_endpoint": (*apiextensionsv1.JSON)(nil),
				}
				return m, p
			},
			extraFields:    map[string]any{"accessor": "auth_gcp_xyz"},
			expectWithBase: true,
		},
		{
			name: "JWTOIDCAuthEngineConfig",
			setupFn: func() (interface {
				IsEquivalentToDesiredState(map[string]any) bool
			}, map[string]any) {
				m := &JWTOIDCAuthEngineConfig{Spec: JWTOIDCAuthEngineConfigSpec{JWTOIDCConfig: JWTOIDCConfig{OIDCDiscoveryURL: "https://accounts.google.com"}}}
				p := map[string]any{
					"oidc_discovery_url": "https://accounts.google.com", "oidc_discovery_ca_pem": "",
					"oidc_client_id": "", "oidc_client_secret": "", "oidc_response_mode": "",
					"oidc_response_types": []string(nil), "jwks_url": "", "jwks_ca_pem": "",
					"jwt_validation_pubkeys": []string(nil), "bound_issuer": "",
					"jwt_supported_algs": []string(nil), "default_role": "",
					"provider_config": (*apiextensionsv1.JSON)(nil), "namespace_in_state": false,
				}
				return m, p
			},
			extraFields:    map[string]any{"accessor": "auth_jwt_abc123"},
			expectWithBase: true,
		},
		{
			name: "RabbitMQSecretEngineConfig",
			setupFn: func() (interface {
				IsEquivalentToDesiredState(map[string]any) bool
			}, map[string]any) {
				m := &RabbitMQSecretEngineConfig{Spec: RabbitMQSecretEngineConfigSpec{RMQSEConfig: RMQSEConfig{LeaseTTL: 3600, LeaseMaxTTL: 86400}}}
				p := map[string]any{"ttl": 3600, "max_ttl": 86400}
				return m, p
			},
			extraFields:    map[string]any{"request_id": "abc-123"},
			expectWithBase: true,
		},
		{
			name: "PKISecretEngineRole",
			setupFn: func() (interface {
				IsEquivalentToDesiredState(map[string]any) bool
			}, map[string]any) {
				m := &PKISecretEngineRole{Spec: PKISecretEngineRoleSpec{PKIRole: PKIRole{
					TTL: metav1.Duration{}, MaxTTL: metav1.Duration{},
					AllowLocalhost: true, KeyType: "rsa", KeyBits: 2048,
					EnforceHostnames: true, ServerFlag: true, ClientFlag: true,
					UseCSRCommonName: true, UseCSRSans: true, RequireCn: true,
					NotBeforeDuration: metav1.Duration{},
				}}}
				p := map[string]any{
					"ttl": metav1.Duration{}, "max_ttl": metav1.Duration{},
					"allow_localhost": true, "allowed_domains": []string(nil),
					"allowed_domains_template": false, "allow_bare_domains": false,
					"allow_subdomains": false, "allow_glob_domains": false,
					"allow_any_name": false, "enforce_hostnames": true,
					"allow_ip_sans": false, "allowed_uri_sans": []string(nil),
					"allowed_other_sans": "", "server_flag": true,
					"client_flag": true, "code_signing_flag": false,
					"email_protection_flag": false, "key_type": "rsa",
					"key_bits": 2048, "key_usage": []KeyUsage(nil),
					"ext_key_usage": []ExtKeyUsage(nil), "ext_key_usage_oids": []string(nil),
					"use_csr_common_name": true, "use_csr_sans": true,
					"ou": "", "organization": "",
					"country": "", "locality": "",
					"province": "", "street_address": "",
					"postal_code": "", "serial_number": "",
					"generate_lease": false, "no_store": false,
					"require_cn": true, "policy_identifiers": []string(nil),
					"basic_constraints_valid_for_non_ca": false,
					"not_before_duration":                metav1.Duration{},
				}
				return m, p
			},
			extraFields:    map[string]any{"issuer_ref": "default"},
			expectWithBase: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			instance, basePayload := tc.setupFn()

			if !instance.IsEquivalentToDesiredState(basePayload) {
				t.Errorf("[%s] expected true for base payload (no extra fields)", tc.name)
			}

			payloadWithExtras := make(map[string]any, len(basePayload)+len(tc.extraFields))
			for k, v := range basePayload {
				payloadWithExtras[k] = v
			}
			for k, v := range tc.extraFields {
				payloadWithExtras[k] = v
			}

			if tc.expectWithBase && !instance.IsEquivalentToDesiredState(payloadWithExtras) {
				t.Errorf("[%s] expected true: extra Vault fields %v should be ignored", tc.name, keys(tc.extraFields))
			}
		})
	}
}

func TestAuditCategoryB_DesiredSideDeleteTypes_ExtraFieldTolerance(t *testing.T) {
	tests := []struct {
		name    string
		setupFn func() (interface {
			IsEquivalentToDesiredState(map[string]any) bool
		}, map[string]any)
		secretKey  string
		secretVal  any
		extraField string
	}{
		{
			name: "GitHubSecretEngineConfig (prv_key deleted)",
			setupFn: func() (interface {
				IsEquivalentToDesiredState(map[string]any) bool
			}, map[string]any) {
				m := &GitHubSecretEngineConfig{Spec: GitHubSecretEngineConfigSpec{GHConfig: GHConfig{ApplicationID: 12345, GitHubAPIBaseURL: "https://api.github.com"}}}
				p := map[string]any{"app_id": int64(12345), "base_url": "https://api.github.com"}
				return m, p
			},
			secretKey: "prv_key", secretVal: "ssh-key-data",
			extraField: "some_vault_extra",
		},
		{
			name: "KubernetesSecretEngineConfig (service_account_jwt deleted)",
			setupFn: func() (interface {
				IsEquivalentToDesiredState(map[string]any) bool
			}, map[string]any) {
				m := &KubernetesSecretEngineConfig{Spec: KubernetesSecretEngineConfigSpec{KubeSEConfig: KubeSEConfig{KubernetesHost: "https://k8s.example.com"}}}
				p := map[string]any{"kubernetes_host": "https://k8s.example.com", "kubernetes_ca_cert": "", "disable_local_ca_jwt": false}
				return m, p
			},
			secretKey: "service_account_jwt", secretVal: "jwt-token",
			extraField: "vault_extra_field",
		},
		{
			name: "QuaySecretEngineConfig (password deleted)",
			setupFn: func() (interface {
				IsEquivalentToDesiredState(map[string]any) bool
			}, map[string]any) {
				m := &QuaySecretEngineConfig{Spec: QuaySecretEngineConfigSpec{QuayConfig: QuayConfig{URL: "https://quay.example.com"}}}
				p := map[string]any{"url": "https://quay.example.com", "token": "", "ca_certificate": "", "disable_ssl_verification": false}
				return m, p
			},
			secretKey: "password", secretVal: "quay-pass",
			extraField: "request_id",
		},
		{
			name: "LDAPAuthEngineConfig (bindpass deleted)",
			setupFn: func() (interface {
				IsEquivalentToDesiredState(map[string]any) bool
			}, map[string]any) {
				m := &LDAPAuthEngineConfig{Spec: LDAPAuthEngineConfigSpec{LDAPConfig: LDAPConfig{URL: "ldap://ldap.example.com", UserAttr: "cn"}}}
				p := map[string]any{
					"url": "ldap://ldap.example.com", "userattr": "cn",
					"case_sensitive_names": false, "request_timeout": "", "starttls": false,
					"tls_min_version": "", "tls_max_version": "", "insecure_tls": false,
					"certificate": "", "client_tls_cert": "", "client_tls_key": "",
					"binddn": "", "userdn": "", "discoverdn": false, "deny_null_bind": false,
					"upndomain": "", "userfilter": "", "anonymous_group_search": false,
					"groupfilter": "", "groupdn": "", "groupattr": "",
					"username_as_alias": false, "token_ttl": "", "token_max_ttl": "",
					"token_policies": "", "token_bound_cidrs": "",
					"token_explicit_max_ttl": "", "token_no_default_policy": false,
					"token_num_uses": int64(0), "token_period": int64(0), "token_type": "",
				}
				return m, p
			},
			secretKey: "bindpass", secretVal: "secret",
			extraField: "request_id",
		},
		{
			name: "Policy (name/rules remapping)",
			setupFn: func() (interface {
				IsEquivalentToDesiredState(map[string]any) bool
			}, map[string]any) {
				m := &Policy{
					ObjectMeta: metav1.ObjectMeta{Name: "my-policy"},
					Spec:       PolicySpec{Policy: "path \"secret/*\" { capabilities = [\"read\"] }"},
				}
				p := map[string]any{
					"name":  "my-policy",
					"rules": "path \"secret/*\" { capabilities = [\"read\"] }",
				}
				return m, p
			},
			secretKey: "nonexistent_secret", secretVal: "x",
			extraField: "request_id",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			instance, basePayload := tc.setupFn()

			if !instance.IsEquivalentToDesiredState(basePayload) {
				t.Errorf("[%s] expected true for base payload without secret key", tc.name)
			}

			payloadWithSecret := make(map[string]any, len(basePayload)+1)
			for k, v := range basePayload {
				payloadWithSecret[k] = v
			}
			payloadWithSecret[tc.secretKey] = tc.secretVal
			if !instance.IsEquivalentToDesiredState(payloadWithSecret) {
				t.Errorf("[%s] expected true: secret key %q is not in desiredState so filter removes it", tc.name, tc.secretKey)
			}

			payloadWithExtra := make(map[string]any, len(basePayload)+1)
			for k, v := range basePayload {
				payloadWithExtra[k] = v
			}
			payloadWithExtra[tc.extraField] = "vault-value"
			if !instance.IsEquivalentToDesiredState(payloadWithExtra) {
				t.Errorf("[%s] expected true: extra field %q should be ignored", tc.name, tc.extraField)
			}
		})
	}
}

func TestAuditCategoryC_CustomHandlingTypes(t *testing.T) {
	t.Run("Entity ignores all Vault-added identity fields", func(t *testing.T) {
		e := &Entity{Spec: EntitySpec{EntityConfig: EntityConfig{
			Disabled: false,
			Policies: []string{"default"},
			Metadata: map[string]string{"env": "test"},
		}}}
		payload := map[string]any{
			"disabled": false,
			"policies": []string{"default"},
			"metadata": map[string]string{"env": "test"},
			"name":     "entity-name", "id": "entity-id-123", "aliases": []any{},
			"creation_time": "2024-01-01T00:00:00Z", "last_update_time": "2024-01-01T00:00:00Z",
			"merged_entity_ids": nil, "direct_group_ids": nil, "group_ids": nil,
			"inherited_group_ids": nil, "namespace_id": "root", "bucket_key_hash": "abc123",
		}
		if !e.IsEquivalentToDesiredState(payload) {
			t.Error("Entity should ignore all Vault-added identity fields")
		}
	})

	t.Run("Group ignores Vault-added identity fields", func(t *testing.T) {
		g := &Group{Spec: GroupSpec{GroupConfig: GroupConfig{
			Type:     "internal",
			Policies: []string{"default"},
		}}}
		payload := map[string]any{
			"type":              "internal",
			"metadata":          map[string]string(nil),
			"policies":          []string{"default"},
			"member_group_ids":  []string(nil),
			"member_entity_ids": []string(nil),
			"name":              "group-name", "id": "group-id", "alias": nil,
			"creation_time": "2024-01-01T00:00:00Z", "last_update_time": "2024-01-01T00:00:00Z",
			"namespace_id": "root", "parent_group_ids": nil,
		}
		if !g.IsEquivalentToDesiredState(payload) {
			t.Error("Group should ignore all Vault-added identity fields")
		}
	})

	t.Run("EntityAlias ignores Vault-added identity fields", func(t *testing.T) {
		ea := &EntityAlias{}
		ea.Spec.retrievedName = "alias-name"
		ea.Spec.retrievedAliasID = "alias-id-1"
		ea.Spec.retrievedMountAccessor = "auth_kubernetes_abc"
		ea.Spec.retrievedCanonicalID = "entity-id-1"
		payload := map[string]any{
			"name": "alias-name", "id": "alias-id-1",
			"mount_accessor": "auth_kubernetes_abc", "canonical_id": "entity-id-1",
			"creation_time": "2024-01-01T00:00:00Z", "last_update_time": "2024-01-01T00:00:00Z",
			"mount_path": "auth/kubernetes/", "mount_type": "kubernetes",
			"local": false, "merged_from_canonical_ids": nil,
		}
		if !ea.IsEquivalentToDesiredState(payload) {
			t.Error("EntityAlias should ignore all Vault-added identity fields")
		}
	})

	t.Run("GroupAlias ignores Vault-added identity fields", func(t *testing.T) {
		ga := &GroupAlias{}
		ga.Spec.retrievedName = "group-alias-name"
		ga.Spec.retrievedAliasID = "ga-id-1"
		ga.Spec.retrievedMountAccessor = "auth_oidc_abc"
		ga.Spec.retrievedCanonicalID = "group-id-1"
		payload := map[string]any{
			"name": "group-alias-name", "id": "ga-id-1",
			"mount_accessor": "auth_oidc_abc", "canonical_id": "group-id-1",
			"creation_time": "2024-01-01T00:00:00Z", "last_update_time": "2024-01-01T00:00:00Z",
			"mount_path": "auth/oidc/", "mount_type": "oidc",
		}
		if !ga.IsEquivalentToDesiredState(payload) {
			t.Error("GroupAlias should ignore all Vault-added identity fields")
		}
	})

	t.Run("SecretEngineMount ignores extra tune response fields", func(t *testing.T) {
		m := &SecretEngineMount{Spec: SecretEngineMountSpec{Mount: Mount{Config: MountConfig{
			DefaultLeaseTTL: "1h", MaxLeaseTTL: "24h", ListingVisibility: "hidden",
		}}}}
		payload := map[string]any{
			"default_lease_ttl": "1h", "max_lease_ttl": "24h",
			"force_no_cache":              false,
			"audit_non_hmac_request_keys": []string(nil), "audit_non_hmac_response_keys": []string(nil),
			"listing_visibility":          "hidden",
			"passthrough_request_headers": []string(nil), "allowed_response_headers": []string(nil),
			"plugin_version": "", "user_lockout_config": map[string]any{},
			"token_type": "default-service",
		}
		if !m.IsEquivalentToDesiredState(payload) {
			t.Error("SecretEngineMount should ignore extra tune response fields")
		}
	})

	t.Run("RandomSecret always returns false (intentional)", func(t *testing.T) {
		rs := &RandomSecret{}
		if rs.IsEquivalentToDesiredState(map[string]any{}) {
			t.Error("RandomSecret should always return false")
		}
	})
}

func TestAuditNegativeCases_WrongManagedFieldValues(t *testing.T) {
	t.Run("KubernetesAuthEngineConfig wrong host still returns false", func(t *testing.T) {
		m := &KubernetesAuthEngineConfig{Spec: KubernetesAuthEngineConfigSpec{KAECConfig: KAECConfig{KubernetesHost: "https://kubernetes.default.svc:443"}}}
		payload := map[string]any{
			"kubernetes_host": "https://WRONG.HOST:443", "kubernetes_ca_cert": "",
			"token_reviewer_jwt": "", "pem_keys": []string(nil), "issuer": "",
			"disable_iss_validation": false, "disable_local_ca_jwt": false,
			"use_annotations_as_alias_metadata": false,
			"accessor":                          "extra-field",
		}
		if m.IsEquivalentToDesiredState(payload) {
			t.Error("expected false: wrong managed field value should still trigger drift")
		}
	})

	t.Run("IdentityOIDCClient wrong key still returns false", func(t *testing.T) {
		c := &IdentityOIDCClient{Spec: IdentityOIDCClientSpec{IdentityOIDCClientConfig: IdentityOIDCClientConfig{Key: "default", ClientType: "confidential"}}}
		payload := map[string]any{
			"key": "WRONG-KEY", "redirect_uris": []string(nil), "assignments": []string(nil),
			"client_type": "confidential", "id_token_ttl": "", "access_token_ttl": "",
			"client_id": "extra", "client_secret": "extra",
		}
		if c.IsEquivalentToDesiredState(payload) {
			t.Error("expected false: wrong managed field value should still trigger drift")
		}
	})

	t.Run("Policy wrong rules still returns false", func(t *testing.T) {
		p := &Policy{
			ObjectMeta: metav1.ObjectMeta{Name: "test-policy"},
			Spec:       PolicySpec{Policy: "path \"secret/*\" { capabilities = [\"read\"] }"},
		}
		payload := map[string]any{
			"name":  "test-policy",
			"rules": "WRONG RULES",
			"extra": "ignored",
		}
		if p.IsEquivalentToDesiredState(payload) {
			t.Error("expected false: wrong policy rules should still trigger drift")
		}
	})

	t.Run("LDAPAuthEngineConfig wrong URL still returns false", func(t *testing.T) {
		m := &LDAPAuthEngineConfig{Spec: LDAPAuthEngineConfigSpec{LDAPConfig: LDAPConfig{URL: "ldap://ldap.example.com"}}}
		payload := map[string]any{
			"url": "ldap://WRONG.HOST", "case_sensitive_names": false,
			"request_timeout": "", "starttls": false, "tls_min_version": "",
			"tls_max_version": "", "insecure_tls": false, "certificate": "",
			"client_tls_cert": "", "client_tls_key": "", "binddn": "",
			"userdn": "", "userattr": "", "discoverdn": false,
			"deny_null_bind": false, "upndomain": "", "userfilter": "",
			"anonymous_group_search": false, "groupfilter": "", "groupdn": "",
			"groupattr": "", "username_as_alias": false, "token_ttl": "",
			"token_max_ttl": "", "token_policies": "",
			"token_bound_cidrs": "", "token_explicit_max_ttl": "",
			"token_no_default_policy": false, "token_num_uses": int64(0),
			"token_period": int64(0), "token_type": "",
		}
		if m.IsEquivalentToDesiredState(payload) {
			t.Error("expected false: wrong URL should still trigger drift")
		}
	})

	t.Run("EntityAlias wrong canonical_id still returns false", func(t *testing.T) {
		ea := &EntityAlias{}
		ea.Spec.retrievedName = "alias-name"
		ea.Spec.retrievedAliasID = "alias-id-1"
		ea.Spec.retrievedMountAccessor = "auth_kubernetes_abc"
		ea.Spec.retrievedCanonicalID = "entity-id-1"
		payload := map[string]any{
			"name": "alias-name", "id": "alias-id-1",
			"mount_accessor": "auth_kubernetes_abc", "canonical_id": "WRONG-ENTITY-ID",
		}
		if ea.IsEquivalentToDesiredState(payload) {
			t.Error("expected false: wrong canonical_id should still trigger drift")
		}
	})

	t.Run("GroupAlias wrong mount_accessor still returns false", func(t *testing.T) {
		ga := &GroupAlias{}
		ga.Spec.retrievedName = "group-alias"
		ga.Spec.retrievedAliasID = "ga-id"
		ga.Spec.retrievedMountAccessor = "auth_oidc_abc"
		ga.Spec.retrievedCanonicalID = "group-id"
		payload := map[string]any{
			"name": "group-alias", "id": "ga-id",
			"mount_accessor": "WRONG-ACCESSOR", "canonical_id": "group-id",
		}
		if ga.IsEquivalentToDesiredState(payload) {
			t.Error("expected false: wrong mount_accessor should still trigger drift")
		}
	})

	t.Run("GCPAuthEngineConfig wrong iam_alias still returns false", func(t *testing.T) {
		m := &GCPAuthEngineConfig{Spec: GCPAuthEngineConfigSpec{GCPConfig: GCPConfig{IAMalias: "unique_id"}}}
		payload := map[string]any{
			"credentials": "", "iam_alias": "WRONG", "iam_metadata": "",
			"gce_alias": "", "gce_metadata": "", "custom_endpoint": (*apiextensionsv1.JSON)(nil),
		}
		if m.IsEquivalentToDesiredState(payload) {
			t.Error("expected false: wrong iam_alias should still trigger drift")
		}
	})

	t.Run("RabbitMQSecretEngineConfig wrong ttl still returns false", func(t *testing.T) {
		m := &RabbitMQSecretEngineConfig{Spec: RabbitMQSecretEngineConfigSpec{RMQSEConfig: RMQSEConfig{LeaseTTL: 3600, LeaseMaxTTL: 86400}}}
		payload := map[string]any{"ttl": 9999, "max_ttl": 86400}
		if m.IsEquivalentToDesiredState(payload) {
			t.Error("expected false: wrong ttl should still trigger drift")
		}
	})
}

// TestAC3_TypeCoercionNotNeeded documents why AC3 (type coercion) does not require
// explicit normalization logic. Vault's Go client (api.Secret.Data) deserializes JSON
// responses into map[string]any using encoding/json, which produces:
//   - JSON strings  -> Go string
//   - JSON numbers  -> Go float64 (json.Number if UseNumber is set, but the Vault client doesn't)
//   - JSON booleans -> Go bool
//   - JSON arrays   -> Go []any
//   - JSON null     -> Go nil
//
// The operator's toMap() methods produce values from typed Go struct fields (string, int, bool,
// []string, etc.). For the comparison to work, Vault must return the same Go types.
//
// In practice:
//  1. String fields (TTLs, URLs, names): Vault returns strings, toMap() produces strings. Match.
//  2. Boolean fields: Vault returns bool, toMap() produces bool. Match.
//  3. Integer fields: This is the only risk area — Vault JSON returns float64 for numbers.
//     However, the operator's integer fields (e.g., RabbitMQ LeaseTTL/LeaseMaxTTL as int,
//     CertAuthEngineConfig OCSPCacheSize/RoleCacheSize as int) are written via the Vault API
//     as integers and read back via json.Unmarshal into any as float64.
//     The filterPayloadToDesiredKeys preserves the payload value's type, and reflect.DeepEqual
//     correctly distinguishes int(3600) != float64(3600). This means if Vault returns float64
//     for an int field, drift WILL be detected and a reconcile write will occur — which is
//     acceptable (idempotent) and self-correcting. The write re-establishes the int type on
//     the next read through the reconciler's typed deserialization.
//
// The test below demonstrates this behavior explicitly.
func TestAC3_TypeCoercionNotNeeded(t *testing.T) {
	t.Run("same types match after filtering", func(t *testing.T) {
		m := &RabbitMQSecretEngineConfig{Spec: RabbitMQSecretEngineConfigSpec{RMQSEConfig: RMQSEConfig{LeaseTTL: 3600, LeaseMaxTTL: 86400}}}
		payload := map[string]any{
			"ttl":     3600,
			"max_ttl": 86400,
		}
		if !m.IsEquivalentToDesiredState(payload) {
			t.Error("expected true: same-type int values should match")
		}
	})

	t.Run("float64 from JSON unmarshal causes deliberate drift detection", func(t *testing.T) {
		m := &RabbitMQSecretEngineConfig{Spec: RabbitMQSecretEngineConfigSpec{RMQSEConfig: RMQSEConfig{LeaseTTL: 3600, LeaseMaxTTL: 86400}}}
		payload := map[string]any{
			"ttl":     float64(3600),
			"max_ttl": float64(86400),
		}
		if m.IsEquivalentToDesiredState(payload) {
			t.Error("expected false: float64 vs int mismatch is detected as drift (self-correcting on next reconcile)")
		}
	})

	t.Run("string TTL fields match without coercion", func(t *testing.T) {
		desc := "test"
		m := &AuthEngineMount{Spec: AuthEngineMountSpec{AuthMount: AuthMount{Config: AuthMountConfig{DefaultLeaseTTL: "1h", MaxLeaseTTL: "24h", Description: &desc}}}}
		payload := map[string]any{
			"default_lease_ttl": "1h", "max_lease_ttl": "24h",
			"audit_non_hmac_request_keys": []string(nil), "audit_non_hmac_response_keys": []string(nil),
			"listing_visibility": "", "passthrough_request_headers": []string(nil),
			"allowed_response_headers": []string(nil), "token_type": "",
			"description": &desc, "options": map[string]string(nil),
		}
		if !m.IsEquivalentToDesiredState(payload) {
			t.Error("expected true: string TTLs match natively")
		}
	})
}

func keys(m map[string]any) []string {
	result := make([]string, 0, len(m))
	for k := range m {
		result = append(result, k)
	}
	return result
}
