/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	vaultutils "github.com/redhat-cop/vault-config-operator/api/v1alpha1/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func defaultAuth() vaultutils.KubeAuthConfiguration {
	return vaultutils.KubeAuthConfiguration{
		Path:           "kubernetes",
		Role:           "test-role",
		ServiceAccount: &corev1.LocalObjectReference{Name: "default"},
	}
}

func authWithTargetNamespace(ns, targetNs string) vaultutils.KubeAuthConfiguration {
	return vaultutils.KubeAuthConfiguration{
		Path:            "kubernetes",
		Role:            "test-role",
		ServiceAccount:  &corev1.LocalObjectReference{Name: "default"},
		Namespace:       ns,
		TargetNamespace: targetNs,
	}
}

// createAndFetch creates a resource via the API server and fetches it back so
// that the returned object has a valid ResourceVersion for update operations.
func createAndFetch(obj client.Object) {
	ExpectWithOffset(1, k8sClient.Create(ctx, obj)).To(Succeed())
	DeferCleanup(func() { _ = k8sClient.Delete(ctx, obj) })
	ExpectWithOffset(1, k8sClient.Get(ctx, client.ObjectKeyFromObject(obj), obj)).To(Succeed())
}

// ── Immutable Path Tests ──────────────────────────────────────────────

var _ = Describe("VAP Immutable Path", func() {

	type pathTestCase struct {
		name   string
		obj    func() client.Object
		mutate func(obj client.Object)
	}

	entries := []pathTestCase{
		{
			name: "AuthEngineMount",
			obj: func() client.Object {
				return &AuthEngineMount{
					ObjectMeta: metav1.ObjectMeta{GenerateName: "vap-path-", Namespace: "default"},
					Spec: AuthEngineMountSpec{
						Authentication: defaultAuth(),
						Path:           "auth/original",
						AuthMount:      AuthMount{Type: "approle", Description: "test"},
					},
				}
			},
			mutate: func(obj client.Object) { obj.(*AuthEngineMount).Spec.Path = "auth/changed" },
		},
		{
			name: "SecretEngineMount",
			obj: func() client.Object {
				return &SecretEngineMount{
					ObjectMeta: metav1.ObjectMeta{GenerateName: "vap-path-", Namespace: "default"},
					Spec: SecretEngineMountSpec{
						Authentication: defaultAuth(),
						Path:           "secret/original",
						Mount:          Mount{Type: "kv"},
					},
				}
			},
			mutate: func(obj client.Object) { obj.(*SecretEngineMount).Spec.Path = "secret/changed" },
		},
		{
			name: "AzureAuthEngineRole (gap fix)",
			obj: func() client.Object {
				return &AzureAuthEngineRole{
					ObjectMeta: metav1.ObjectMeta{GenerateName: "vap-path-", Namespace: "default"},
					Spec: AzureAuthEngineRoleSpec{
						Authentication: defaultAuth(),
						Path:           "azure/original",
					},
				}
			},
			mutate: func(obj client.Object) { obj.(*AzureAuthEngineRole).Spec.Path = "azure/changed" },
		},
		{
			name: "GCPAuthEngineRole (gap fix)",
			obj: func() client.Object {
				return &GCPAuthEngineRole{
					ObjectMeta: metav1.ObjectMeta{GenerateName: "vap-path-", Namespace: "default"},
					Spec: GCPAuthEngineRoleSpec{
						Authentication: defaultAuth(),
						Path:           "gcp/original",
						GCPRole:        GCPRole{Name: "test-role", Type: "iam"},
					},
				}
			},
			mutate: func(obj client.Object) { obj.(*GCPAuthEngineRole).Spec.Path = "gcp/changed" },
		},
		{
			name: "JWTOIDCAuthEngineRole (gap fix)",
			obj: func() client.Object {
				return &JWTOIDCAuthEngineRole{
					ObjectMeta: metav1.ObjectMeta{GenerateName: "vap-path-", Namespace: "default"},
					Spec: JWTOIDCAuthEngineRoleSpec{
						Authentication: defaultAuth(),
						Path:           "jwt/original",
						JWTOIDCRole:    JWTOIDCRole{Name: "test-role", RoleType: "jwt"},
					},
				}
			},
			mutate: func(obj client.Object) { obj.(*JWTOIDCAuthEngineRole).Spec.Path = "jwt/changed" },
		},
		{
			name: "LDAPAuthEngineGroup (gap fix)",
			obj: func() client.Object {
				return &LDAPAuthEngineGroup{
					ObjectMeta: metav1.ObjectMeta{GenerateName: "vap-path-", Namespace: "default"},
					Spec: LDAPAuthEngineGroupSpec{
						Authentication: defaultAuth(),
						Path:           "ldap/original",
					},
				}
			},
			mutate: func(obj client.Object) { obj.(*LDAPAuthEngineGroup).Spec.Path = "ldap/changed" },
		},
		{
			name: "AWSSecretEngineRole",
			obj: func() client.Object {
				return &AWSSecretEngineRole{
					ObjectMeta: metav1.ObjectMeta{GenerateName: "vap-path-", Namespace: "default"},
					Spec: AWSSecretEngineRoleSpec{
						Authentication: defaultAuth(),
						Path:           "aws/original",
						AWSRole:        AWSRole{CredentialType: "iam_user"},
					},
				}
			},
			mutate: func(obj client.Object) { obj.(*AWSSecretEngineRole).Spec.Path = "aws/changed" },
		},
		{
			name: "DatabaseSecretEngineConfig",
			obj: func() client.Object {
				return &DatabaseSecretEngineConfig{
					ObjectMeta: metav1.ObjectMeta{GenerateName: "vap-path-", Namespace: "default"},
					Spec: DatabaseSecretEngineConfigSpec{
						Authentication:  defaultAuth(),
						Path:            "database/original",
						RootCredentials: vaultutils.RootCredentialConfig{Secret: &corev1.LocalObjectReference{Name: "db-creds"}},
					},
				}
			},
			mutate: func(obj client.Object) { obj.(*DatabaseSecretEngineConfig).Spec.Path = "database/changed" },
		},
		{
			name: "TransitSecretEngineKey",
			obj: func() client.Object {
				return &TransitSecretEngineKey{
					ObjectMeta: metav1.ObjectMeta{GenerateName: "vap-path-", Namespace: "default"},
					Spec: TransitSecretEngineKeySpec{
						Authentication: defaultAuth(),
						Path:           "transit/original",
					},
				}
			},
			mutate: func(obj client.Object) { obj.(*TransitSecretEngineKey).Spec.Path = "transit/changed" },
		},
		{
			name: "KubernetesAuthEngineRole",
			obj: func() client.Object {
				return &KubernetesAuthEngineRole{
					ObjectMeta: metav1.ObjectMeta{GenerateName: "vap-path-", Namespace: "default"},
					Spec: KubernetesAuthEngineRoleSpec{
						Authentication:   defaultAuth(),
						Path:             "k8s/original",
						VRole:            VRole{Policies: []string{"default"}},
						TargetNamespaces: vaultutils.TargetNamespaceConfig{TargetNamespaces: []string{"default"}},
					},
				}
			},
			mutate: func(obj client.Object) { obj.(*KubernetesAuthEngineRole).Spec.Path = "k8s/changed" },
		},
		{
			name: "SSHSecretEngineConfig",
			obj: func() client.Object {
				return &SSHSecretEngineConfig{
					ObjectMeta: metav1.ObjectMeta{GenerateName: "vap-path-", Namespace: "default"},
					Spec: SSHSecretEngineConfigSpec{
						Authentication: defaultAuth(),
						Path:           "ssh/original",
						SSHSEConfig:    SSHSEConfig{GenerateSigningKey: true},
					},
				}
			},
			mutate: func(obj client.Object) { obj.(*SSHSecretEngineConfig).Spec.Path = "ssh/changed" },
		},
		{
			name: "ConsulSecretEngineRole",
			obj: func() client.Object {
				return &ConsulSecretEngineRole{
					ObjectMeta: metav1.ObjectMeta{GenerateName: "vap-path-", Namespace: "default"},
					Spec: ConsulSecretEngineRoleSpec{
						Authentication: defaultAuth(),
						Path:           "consul/original",
					},
				}
			},
			mutate: func(obj client.Object) { obj.(*ConsulSecretEngineRole).Spec.Path = "consul/changed" },
		},
		{
			name: "NomadSecretEngineRole",
			obj: func() client.Object {
				return &NomadSecretEngineRole{
					ObjectMeta: metav1.ObjectMeta{GenerateName: "vap-path-", Namespace: "default"},
					Spec: NomadSecretEngineRoleSpec{
						Authentication: defaultAuth(),
						Path:           "nomad/original",
					},
				}
			},
			mutate: func(obj client.Object) { obj.(*NomadSecretEngineRole).Spec.Path = "nomad/changed" },
		},
		{
			name: "AWSSecretEngineConfig",
			obj: func() client.Object {
				return &AWSSecretEngineConfig{
					ObjectMeta: metav1.ObjectMeta{GenerateName: "vap-path-", Namespace: "default"},
					Spec: AWSSecretEngineConfigSpec{
						Authentication:  defaultAuth(),
						Path:            "aws-cfg/original",
						RootCredentials: vaultutils.RootCredentialConfig{Secret: &corev1.LocalObjectReference{Name: "aws-creds"}},
					},
				}
			},
			mutate: func(obj client.Object) { obj.(*AWSSecretEngineConfig).Spec.Path = "aws-cfg/changed" },
		},
		{
			name: "PKISecretEngineRole",
			obj: func() client.Object {
				return &PKISecretEngineRole{
					ObjectMeta: metav1.ObjectMeta{GenerateName: "vap-path-", Namespace: "default"},
					Spec: PKISecretEngineRoleSpec{
						Authentication: defaultAuth(),
						Path:           "pki/original",
					},
				}
			},
			mutate: func(obj client.Object) { obj.(*PKISecretEngineRole).Spec.Path = "pki/changed" },
		},
	}

	// Suite BeforeSuite already waits for VAP enforcement. Keep Eventually here
	// as a belt-and-suspenders probe for the immutable-path policy.
	It("should reject path change for AuthEngineMount (VAP readiness probe)", func() {
		obj := entries[0].obj()
		createAndFetch(obj)
		entries[0].mutate(obj)
		Eventually(func() error {
			return k8sClient.Update(ctx, obj)
		}, "30s", "1s").Should(HaveOccurred(), "VAP immutable-path not enforced after 30s")
		err := k8sClient.Update(ctx, obj)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("spec.path cannot be updated"))
	})

	for _, tc := range entries[1:] {
		tc := tc
		It(fmt.Sprintf("should reject path change for %s", tc.name), func() {
			obj := tc.obj()
			createAndFetch(obj)
			tc.mutate(obj)
			err := k8sClient.Update(ctx, obj)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("spec.path cannot be updated"))
		})
	}

	It("should allow update of non-path fields", func() {
		obj := &AuthEngineMount{
			ObjectMeta: metav1.ObjectMeta{GenerateName: "vap-path-allow-", Namespace: "default"},
			Spec: AuthEngineMountSpec{
				Authentication: defaultAuth(),
				Path:           "auth/keep",
				AuthMount:      AuthMount{Type: "approle", Description: "original desc"},
			},
		}
		createAndFetch(obj)
		obj.Spec.AuthMount.Config = AuthMountConfig{DefaultLeaseTTL: "1h"}
		Expect(k8sClient.Update(ctx, obj)).To(Succeed())
	})
})

// ── Immutable Name Tests ──────────────────────────────────────────────

var _ = Describe("VAP Immutable Name", func() {

	type nameTestCase struct {
		name   string
		obj    func() client.Object
		mutate func(obj client.Object)
	}

	entries := []nameTestCase{
		{
			name: "Policy",
			obj: func() client.Object {
				return &Policy{
					ObjectMeta: metav1.ObjectMeta{GenerateName: "vap-name-", Namespace: "default"},
					Spec: PolicySpec{
						Authentication: defaultAuth(),
						Name:           "original-name",
						Policy:         `path "secret/*" { capabilities = ["read"] }`,
					},
				}
			},
			mutate: func(obj client.Object) { obj.(*Policy).Spec.Name = "changed-name" },
		},
		{
			name: "AppRoleAuthEngineRole",
			obj: func() client.Object {
				return &AppRoleAuthEngineRole{
					ObjectMeta: metav1.ObjectMeta{GenerateName: "vap-name-", Namespace: "default"},
					Spec: AppRoleAuthEngineRoleSpec{
						Authentication: defaultAuth(),
						Path:           "approle",
						Name:           "original-name",
					},
				}
			},
			mutate: func(obj client.Object) { obj.(*AppRoleAuthEngineRole).Spec.Name = "changed-name" },
		},
		{
			name: "KubernetesAuthEngineRole",
			obj: func() client.Object {
				return &KubernetesAuthEngineRole{
					ObjectMeta: metav1.ObjectMeta{GenerateName: "vap-name-", Namespace: "default"},
					Spec: KubernetesAuthEngineRoleSpec{
						Authentication:   defaultAuth(),
						Path:             "k8s",
						Name:             "original-name",
						VRole:            VRole{Policies: []string{"default"}},
						TargetNamespaces: vaultutils.TargetNamespaceConfig{TargetNamespaces: []string{"default"}},
					},
				}
			},
			mutate: func(obj client.Object) { obj.(*KubernetesAuthEngineRole).Spec.Name = "changed-name" },
		},
		{
			name: "LeaseCountQuota (gap fix)",
			obj: func() client.Object {
				return &LeaseCountQuota{
					ObjectMeta: metav1.ObjectMeta{GenerateName: "vap-name-", Namespace: "default"},
					Spec: LeaseCountQuotaSpec{
						Authentication:        defaultAuth(),
						Name:                  "original-name",
						LeaseCountQuotaConfig: LeaseCountQuotaConfig{MaxLeases: 100},
					},
				}
			},
			mutate: func(obj client.Object) { obj.(*LeaseCountQuota).Spec.Name = "changed-name" },
		},
		{
			name: "RateLimitQuota (gap fix)",
			obj: func() client.Object {
				return &RateLimitQuota{
					ObjectMeta: metav1.ObjectMeta{GenerateName: "vap-name-", Namespace: "default"},
					Spec: RateLimitQuotaSpec{
						Authentication:       defaultAuth(),
						Name:                 "original-name",
						RateLimitQuotaConfig: RateLimitQuotaConfig{Rate: "100"},
					},
				}
			},
			mutate: func(obj client.Object) { obj.(*RateLimitQuota).Spec.Name = "changed-name" },
		},
		{
			name: "RotationPolicy (gap fix)",
			obj: func() client.Object {
				return &RotationPolicy{
					ObjectMeta: metav1.ObjectMeta{GenerateName: "vap-name-", Namespace: "default"},
					Spec: RotationPolicySpec{
						Authentication: defaultAuth(),
						Name:           "original-name",
					},
				}
			},
			mutate: func(obj client.Object) { obj.(*RotationPolicy).Spec.Name = "changed-name" },
		},
		{
			name: "UICustomHeader (gap fix)",
			obj: func() client.Object {
				return &UICustomHeader{
					ObjectMeta: metav1.ObjectMeta{GenerateName: "vap-name-", Namespace: "default"},
					Spec: UICustomHeaderSpec{
						Authentication:       defaultAuth(),
						Name:                 "original-name",
						UICustomHeaderConfig: UICustomHeaderConfig{Values: []string{"test-value"}},
					},
				}
			},
			mutate: func(obj client.Object) { obj.(*UICustomHeader).Spec.Name = "changed-name" },
		},
		{
			name: "LoggerConfig (gap fix)",
			obj: func() client.Object {
				return &LoggerConfig{
					ObjectMeta: metav1.ObjectMeta{GenerateName: "vap-name-", Namespace: "default"},
					Spec: LoggerConfigSpec{
						Authentication:     defaultAuth(),
						Name:               "original-name",
						LoggerConfigConfig: LoggerConfigConfig{Level: "info"},
					},
				}
			},
			mutate: func(obj client.Object) { obj.(*LoggerConfig).Spec.Name = "changed-name" },
		},
		{
			name: "TransitSecretEngineKey",
			obj: func() client.Object {
				return &TransitSecretEngineKey{
					ObjectMeta: metav1.ObjectMeta{GenerateName: "vap-name-", Namespace: "default"},
					Spec: TransitSecretEngineKeySpec{
						Authentication: defaultAuth(),
						Path:           "transit",
						Name:           "original-name",
					},
				}
			},
			mutate: func(obj client.Object) { obj.(*TransitSecretEngineKey).Spec.Name = "changed-name" },
		},
		{
			name: "DatabaseSecretEngineConfig (gap fix)",
			obj: func() client.Object {
				return &DatabaseSecretEngineConfig{
					ObjectMeta: metav1.ObjectMeta{GenerateName: "vap-name-", Namespace: "default"},
					Spec: DatabaseSecretEngineConfigSpec{
						Authentication:  defaultAuth(),
						Path:            "database",
						Name:            "original-name",
						RootCredentials: vaultutils.RootCredentialConfig{Secret: &corev1.LocalObjectReference{Name: "db-creds"}},
					},
				}
			},
			mutate: func(obj client.Object) { obj.(*DatabaseSecretEngineConfig).Spec.Name = "changed-name" },
		},
		{
			name: "AWSSecretEngineRole (gap fix)",
			obj: func() client.Object {
				return &AWSSecretEngineRole{
					ObjectMeta: metav1.ObjectMeta{GenerateName: "vap-name-", Namespace: "default"},
					Spec: AWSSecretEngineRoleSpec{
						Authentication: defaultAuth(),
						Path:           "aws",
						Name:           "original-name",
						AWSRole:        AWSRole{CredentialType: "iam_user"},
					},
				}
			},
			mutate: func(obj client.Object) { obj.(*AWSSecretEngineRole).Spec.Name = "changed-name" },
		},
		{
			name: "SecretEngineMount",
			obj: func() client.Object {
				return &SecretEngineMount{
					ObjectMeta: metav1.ObjectMeta{GenerateName: "vap-name-", Namespace: "default"},
					Spec: SecretEngineMountSpec{
						Authentication: defaultAuth(),
						Path:           "secret",
						Name:           "original-name",
						Mount:          Mount{Type: "kv"},
					},
				}
			},
			mutate: func(obj client.Object) { obj.(*SecretEngineMount).Spec.Name = "changed-name" },
		},
		{
			name: "ConsulSecretEngineRole (gap fix)",
			obj: func() client.Object {
				return &ConsulSecretEngineRole{
					ObjectMeta: metav1.ObjectMeta{GenerateName: "vap-name-", Namespace: "default"},
					Spec: ConsulSecretEngineRoleSpec{
						Authentication: defaultAuth(),
						Path:           "consul",
						Name:           "original-name",
					},
				}
			},
			mutate: func(obj client.Object) { obj.(*ConsulSecretEngineRole).Spec.Name = "changed-name" },
		},
		{
			name: "NomadSecretEngineRole (gap fix)",
			obj: func() client.Object {
				return &NomadSecretEngineRole{
					ObjectMeta: metav1.ObjectMeta{GenerateName: "vap-name-", Namespace: "default"},
					Spec: NomadSecretEngineRoleSpec{
						Authentication: defaultAuth(),
						Path:           "nomad",
						Name:           "original-name",
					},
				}
			},
			mutate: func(obj client.Object) { obj.(*NomadSecretEngineRole).Spec.Name = "changed-name" },
		},
		{
			name: "LDAPSecretEngineDynamicRole (gap fix)",
			obj: func() client.Object {
				return &LDAPSecretEngineDynamicRole{
					ObjectMeta: metav1.ObjectMeta{GenerateName: "vap-name-", Namespace: "default"},
					Spec: LDAPSecretEngineDynamicRoleSpec{
						Authentication: defaultAuth(),
						Path:           "ldap",
						Name:           "original-name",
					},
				}
			},
			mutate: func(obj client.Object) { obj.(*LDAPSecretEngineDynamicRole).Spec.Name = "changed-name" },
		},
	}

	for _, tc := range entries {
		tc := tc
		It(fmt.Sprintf("should reject name change for %s", tc.name), func() {
			obj := tc.obj()
			createAndFetch(obj)
			tc.mutate(obj)
			err := k8sClient.Update(ctx, obj)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("spec.name cannot be updated"))
		})
	}

	It("should allow update when name is unchanged", func() {
		obj := &Policy{
			ObjectMeta: metav1.ObjectMeta{GenerateName: "vap-name-allow-", Namespace: "default"},
			Spec: PolicySpec{
				Authentication: defaultAuth(),
				Name:           "keep-name",
				Policy:         `path "secret/*" { capabilities = ["read"] }`,
			},
		}
		createAndFetch(obj)
		obj.Spec.Policy = `path "secret/*" { capabilities = ["list"] }`
		Expect(k8sClient.Update(ctx, obj)).To(Succeed())
	})

	It("should reject adding spec.name when it was not set (has() edge case)", func() {
		obj := &TransitSecretEngineKey{
			ObjectMeta: metav1.ObjectMeta{GenerateName: "vap-name-edge-", Namespace: "default"},
			Spec: TransitSecretEngineKeySpec{
				Authentication: defaultAuth(),
				Path:           "transit",
			},
		}
		createAndFetch(obj)
		obj.Spec.Name = "newly-added-name"
		err := k8sClient.Update(ctx, obj)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("spec.name cannot be updated"))
	})
})

// ── Target Namespace Guard Tests ──────────────────────────────────────
// Additional target-namespace-guard tests. The file targetnamespace_webhook_test.go
// already covers the core scenarios; these extend coverage.

var _ = Describe("VAP Target Namespace Guard (extended)", func() {

	It("should reject update that changes authentication.namespace when targetNamespace is set", func() {
		obj := &AuthEngineMount{
			ObjectMeta: metav1.ObjectMeta{GenerateName: "vap-tns-", Namespace: "default"},
			Spec: AuthEngineMountSpec{
				Authentication: authWithTargetNamespace("org", "tenant-a"),
				Path:           "auth/tns",
				AuthMount:      AuthMount{Type: "approle", Description: "test"},
			},
		}
		createAndFetch(obj)
		obj.Spec.Authentication.Namespace = "org2"
		err := k8sClient.Update(ctx, obj)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("cannot be updated to a different Vault namespace"))
	})

	It("should allow update when targetNamespace is not set on either old or new", func() {
		obj := &AuthEngineMount{
			ObjectMeta: metav1.ObjectMeta{GenerateName: "vap-tns-allow-", Namespace: "default"},
			Spec: AuthEngineMountSpec{
				Authentication: vaultutils.KubeAuthConfiguration{
					Path:           "kubernetes",
					Role:           "test-role",
					ServiceAccount: &corev1.LocalObjectReference{Name: "default"},
					Namespace:      "org",
				},
				Path:      "auth/tns",
				AuthMount: AuthMount{Type: "approle", Description: "test"},
			},
		}
		createAndFetch(obj)
		obj.Spec.Authentication.Namespace = "org2"
		Expect(k8sClient.Update(ctx, obj)).To(Succeed())
	})
})

// ── Credential Source VAP Tests ───────────────────────────────────────

var _ = Describe("VAP Credential Source", func() {

	Context("rootCredentials (DatabaseSecretEngineConfig)", func() {
		It("should reject zero credential sources", func() {
			obj := &DatabaseSecretEngineConfig{
				ObjectMeta: metav1.ObjectMeta{GenerateName: "vap-cred-zero-", Namespace: "default"},
				Spec: DatabaseSecretEngineConfigSpec{
					Authentication:  defaultAuth(),
					Path:            "database",
					RootCredentials: vaultutils.RootCredentialConfig{},
				},
			}
			err := k8sClient.Create(ctx, obj)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("exactly one of"))
		})

		It("should reject two credential sources", func() {
			obj := &DatabaseSecretEngineConfig{
				ObjectMeta: metav1.ObjectMeta{GenerateName: "vap-cred-two-", Namespace: "default"},
				Spec: DatabaseSecretEngineConfigSpec{
					Authentication: defaultAuth(),
					Path:           "database",
					RootCredentials: vaultutils.RootCredentialConfig{
						Secret:      &corev1.LocalObjectReference{Name: "db-creds"},
						VaultSecret: &vaultutils.VaultSecretReference{Path: "secret/db"},
					},
				},
			}
			err := k8sClient.Create(ctx, obj)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("exactly one of"))
		})

		It("should allow exactly one credential source", func() {
			obj := &DatabaseSecretEngineConfig{
				ObjectMeta: metav1.ObjectMeta{GenerateName: "vap-cred-one-", Namespace: "default"},
				Spec: DatabaseSecretEngineConfigSpec{
					Authentication:  defaultAuth(),
					Path:            "database",
					RootCredentials: vaultutils.RootCredentialConfig{Secret: &corev1.LocalObjectReference{Name: "db-creds"}},
				},
			}
			Expect(k8sClient.Create(ctx, obj)).To(Succeed())
			DeferCleanup(func() { _ = k8sClient.Delete(ctx, obj) })
		})

		It("should reject update that adds a second credential source", func() {
			obj := &DatabaseSecretEngineConfig{
				ObjectMeta: metav1.ObjectMeta{GenerateName: "vap-cred-upd-", Namespace: "default"},
				Spec: DatabaseSecretEngineConfigSpec{
					Authentication:  defaultAuth(),
					Path:            "database",
					RootCredentials: vaultutils.RootCredentialConfig{Secret: &corev1.LocalObjectReference{Name: "db-creds"}},
				},
			}
			createAndFetch(obj)
			obj.Spec.RootCredentials.VaultSecret = &vaultutils.VaultSecretReference{Path: "secret/db"}
			err := k8sClient.Update(ctx, obj)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("exactly one of"))
		})
	})

	Context("bindCredentials (LDAPAuthEngineConfig)", func() {
		It("should reject zero bind credential sources", func() {
			obj := &LDAPAuthEngineConfig{
				ObjectMeta: metav1.ObjectMeta{GenerateName: "vap-bind-zero-", Namespace: "default"},
				Spec: LDAPAuthEngineConfigSpec{
					Authentication:  defaultAuth(),
					Path:            "ldap",
					BindCredentials: vaultutils.RootCredentialConfig{},
				},
			}
			err := k8sClient.Create(ctx, obj)
			Expect(err).To(HaveOccurred())
			Expect(err.Error()).To(ContainSubstring("exactly one of"))
		})

		It("should allow exactly one bind credential source", func() {
			obj := &LDAPAuthEngineConfig{
				ObjectMeta: metav1.ObjectMeta{GenerateName: "vap-bind-one-", Namespace: "default"},
				Spec: LDAPAuthEngineConfigSpec{
					Authentication:  defaultAuth(),
					Path:            "ldap",
					BindCredentials: vaultutils.RootCredentialConfig{Secret: &corev1.LocalObjectReference{Name: "ldap-creds"}},
				},
			}
			Expect(k8sClient.Create(ctx, obj)).To(Succeed())
			DeferCleanup(func() { _ = k8sClient.Delete(ctx, obj) })
		})
	})
})

// ── CRD-Specific VAP Tests ───────────────────────────────────────────

var _ = Describe("VAP CRD-Specific: TransitSecretEngineKey", func() {

	It("should reject convergentEncryption without derived", func() {
		obj := &TransitSecretEngineKey{
			ObjectMeta: metav1.ObjectMeta{GenerateName: "vap-transit-", Namespace: "default"},
			Spec: TransitSecretEngineKeySpec{
				Authentication:   defaultAuth(),
				Path:             "transit",
				TransitKeyConfig: TransitKeyConfig{ConvergentEncryption: true, Derived: false},
			},
		}
		err := k8sClient.Create(ctx, obj)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("convergentEncryption requires"))
	})

	It("should reject keySize on non-hmac type", func() {
		obj := &TransitSecretEngineKey{
			ObjectMeta: metav1.ObjectMeta{GenerateName: "vap-transit-", Namespace: "default"},
			Spec: TransitSecretEngineKeySpec{
				Authentication:   defaultAuth(),
				Path:             "transit",
				TransitKeyConfig: TransitKeyConfig{Type: "aes256-gcm96", KeySize: 64},
			},
		}
		err := k8sClient.Create(ctx, obj)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("keySize is only applicable"))
	})

	It("should reject hmac keySize < 32", func() {
		obj := &TransitSecretEngineKey{
			ObjectMeta: metav1.ObjectMeta{GenerateName: "vap-transit-", Namespace: "default"},
			Spec: TransitSecretEngineKeySpec{
				Authentication:   defaultAuth(),
				Path:             "transit",
				TransitKeyConfig: TransitKeyConfig{Type: "hmac", KeySize: 16},
			},
		}
		err := k8sClient.Create(ctx, obj)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("keySize must be at least 32"))
	})

	It("should reject autoRotatePeriod < 1h", func() {
		obj := &TransitSecretEngineKey{
			ObjectMeta: metav1.ObjectMeta{GenerateName: "vap-transit-", Namespace: "default"},
			Spec: TransitSecretEngineKeySpec{
				Authentication:   defaultAuth(),
				Path:             "transit",
				TransitKeyConfig: TransitKeyConfig{AutoRotatePeriod: "30m"},
			},
		}
		err := k8sClient.Create(ctx, obj)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("autoRotatePeriod must be at least 1h"))
	})

	It("should reject type change on update", func() {
		obj := &TransitSecretEngineKey{
			ObjectMeta: metav1.ObjectMeta{GenerateName: "vap-transit-", Namespace: "default"},
			Spec: TransitSecretEngineKeySpec{
				Authentication:   defaultAuth(),
				Path:             "transit",
				TransitKeyConfig: TransitKeyConfig{Type: "aes256-gcm96"},
			},
		}
		createAndFetch(obj)
		obj.Spec.Type = "chacha20-poly1305"
		err := k8sClient.Update(ctx, obj)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("spec.type cannot be updated"))
	})

	It("should reject exportable true→false", func() {
		obj := &TransitSecretEngineKey{
			ObjectMeta: metav1.ObjectMeta{GenerateName: "vap-transit-", Namespace: "default"},
			Spec: TransitSecretEngineKeySpec{
				Authentication:   defaultAuth(),
				Path:             "transit",
				TransitKeyConfig: TransitKeyConfig{Exportable: true},
			},
		}
		createAndFetch(obj)
		obj.Spec.Exportable = false
		err := k8sClient.Update(ctx, obj)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("spec.exportable cannot be changed from true to false"))
	})

	It("should allow valid create", func() {
		obj := &TransitSecretEngineKey{
			ObjectMeta: metav1.ObjectMeta{GenerateName: "vap-transit-ok-", Namespace: "default"},
			Spec: TransitSecretEngineKeySpec{
				Authentication:   defaultAuth(),
				Path:             "transit",
				TransitKeyConfig: TransitKeyConfig{Type: "aes256-gcm96", AutoRotatePeriod: "2h"},
			},
		}
		Expect(k8sClient.Create(ctx, obj)).To(Succeed())
		DeferCleanup(func() { _ = k8sClient.Delete(ctx, obj) })
	})

	It("should allow exportable false→true (one-way flag)", func() {
		obj := &TransitSecretEngineKey{
			ObjectMeta: metav1.ObjectMeta{GenerateName: "vap-transit-exp-", Namespace: "default"},
			Spec: TransitSecretEngineKeySpec{
				Authentication:   defaultAuth(),
				Path:             "transit",
				TransitKeyConfig: TransitKeyConfig{Exportable: false},
			},
		}
		createAndFetch(obj)
		obj.Spec.Exportable = true
		Expect(k8sClient.Update(ctx, obj)).To(Succeed())
	})
})

var _ = Describe("VAP CRD-Specific: AuthEngineMount", func() {

	It("should reject type change on update", func() {
		obj := &AuthEngineMount{
			ObjectMeta: metav1.ObjectMeta{GenerateName: "vap-aem-", Namespace: "default"},
			Spec: AuthEngineMountSpec{
				Authentication: defaultAuth(),
				Path:           "auth/mount-test",
				AuthMount:      AuthMount{Type: "approle", Description: "test mount"},
			},
		}
		createAndFetch(obj)
		obj.Spec.AuthMount.Type = "oidc"
		err := k8sClient.Update(ctx, obj)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("only .spec.config can be modified"))
	})

	It("should allow config-only update", func() {
		obj := &AuthEngineMount{
			ObjectMeta: metav1.ObjectMeta{GenerateName: "vap-aem-ok-", Namespace: "default"},
			Spec: AuthEngineMountSpec{
				Authentication: defaultAuth(),
				Path:           "auth/mount-test",
				AuthMount:      AuthMount{Type: "approle", Description: "test mount"},
			},
		}
		createAndFetch(obj)
		obj.Spec.AuthMount.Config = AuthMountConfig{DefaultLeaseTTL: "1h"}
		Expect(k8sClient.Update(ctx, obj)).To(Succeed())
	})
})

var _ = Describe("VAP CRD-Specific: SecretEngineMount", func() {

	It("should reject type change on update", func() {
		obj := &SecretEngineMount{
			ObjectMeta: metav1.ObjectMeta{GenerateName: "vap-sem-", Namespace: "default"},
			Spec: SecretEngineMountSpec{
				Authentication: defaultAuth(),
				Path:           "secret/mount-test",
				Mount:          Mount{Type: "kv", Description: "test"},
			},
		}
		createAndFetch(obj)
		obj.Spec.Mount.Type = "transit"
		err := k8sClient.Update(ctx, obj)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("only .spec.config can be modified"))
	})

	It("should allow config-only update", func() {
		obj := &SecretEngineMount{
			ObjectMeta: metav1.ObjectMeta{GenerateName: "vap-sem-ok-", Namespace: "default"},
			Spec: SecretEngineMountSpec{
				Authentication: defaultAuth(),
				Path:           "secret/mount-test",
				Mount:          Mount{Type: "kv", Description: "test"},
			},
		}
		createAndFetch(obj)
		obj.Spec.Mount.Config = MountConfig{DefaultLeaseTTL: "1h"}
		Expect(k8sClient.Update(ctx, obj)).To(Succeed())
	})
})

var _ = Describe("VAP CRD-Specific: AWSSecretEngineRole", func() {

	It("should reject externalID for non-assumed_role", func() {
		obj := &AWSSecretEngineRole{
			ObjectMeta: metav1.ObjectMeta{GenerateName: "vap-aws-", Namespace: "default"},
			Spec: AWSSecretEngineRoleSpec{
				Authentication: defaultAuth(),
				Path:           "aws",
				AWSRole: AWSRole{
					CredentialType: "iam_user",
					ExternalID:     "ext-123",
				},
			},
		}
		err := k8sClient.Create(ctx, obj)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("externalID is only allowed"))
	})

	It("should reject roleArns missing for assumed_role", func() {
		obj := &AWSSecretEngineRole{
			ObjectMeta: metav1.ObjectMeta{GenerateName: "vap-aws-", Namespace: "default"},
			Spec: AWSSecretEngineRoleSpec{
				Authentication: defaultAuth(),
				Path:           "aws",
				AWSRole:        AWSRole{CredentialType: "assumed_role"},
			},
		}
		err := k8sClient.Create(ctx, obj)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("roleArns is required"))
	})

	It("should reject adding externalID on update for iam_user", func() {
		obj := &AWSSecretEngineRole{
			ObjectMeta: metav1.ObjectMeta{GenerateName: "vap-aws-upd-", Namespace: "default"},
			Spec: AWSSecretEngineRoleSpec{
				Authentication: defaultAuth(),
				Path:           "aws",
				AWSRole:        AWSRole{CredentialType: "iam_user"},
			},
		}
		createAndFetch(obj)
		obj.Spec.ExternalID = "ext-123"
		err := k8sClient.Update(ctx, obj)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("externalID is only allowed"))
	})

	It("should allow valid assumed_role", func() {
		obj := &AWSSecretEngineRole{
			ObjectMeta: metav1.ObjectMeta{GenerateName: "vap-aws-ok-", Namespace: "default"},
			Spec: AWSSecretEngineRoleSpec{
				Authentication: defaultAuth(),
				Path:           "aws",
				AWSRole: AWSRole{
					CredentialType: "assumed_role",
					RoleArns:       []string{"arn:aws:iam::123456789:role/test"},
				},
			},
		}
		Expect(k8sClient.Create(ctx, obj)).To(Succeed())
		DeferCleanup(func() { _ = k8sClient.Delete(ctx, obj) })
	})
})

var _ = Describe("VAP CRD-Specific: RandomSecret", func() {

	It("should reject zero password policy sources", func() {
		obj := &RandomSecret{
			ObjectMeta: metav1.ObjectMeta{GenerateName: "vap-rand-", Namespace: "default"},
			Spec: RandomSecretSpec{
				Authentication: defaultAuth(),
				Path:           "secret/data/random",
				SecretKey:      "password",
				SecretFormat:   VaultPasswordPolicy{},
			},
		}
		err := k8sClient.Create(ctx, obj)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("only one of"))
	})

	It("should reject secretKey change on update", func() {
		obj := &RandomSecret{
			ObjectMeta: metav1.ObjectMeta{GenerateName: "vap-rand-", Namespace: "default"},
			Spec: RandomSecretSpec{
				Authentication: defaultAuth(),
				Path:           "secret/data/random",
				SecretKey:      "password",
				SecretFormat:   VaultPasswordPolicy{PasswordPolicyName: "default"},
			},
		}
		createAndFetch(obj)
		obj.Spec.SecretKey = "changed-key"
		err := k8sClient.Update(ctx, obj)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("secretKey cannot be updated"))
	})

	It("should allow valid create with passwordPolicyName", func() {
		obj := &RandomSecret{
			ObjectMeta: metav1.ObjectMeta{GenerateName: "vap-rand-ok-", Namespace: "default"},
			Spec: RandomSecretSpec{
				Authentication: defaultAuth(),
				Path:           "secret/data/random",
				SecretKey:      "password",
				SecretFormat:   VaultPasswordPolicy{PasswordPolicyName: "default"},
			},
		}
		Expect(k8sClient.Create(ctx, obj)).To(Succeed())
		DeferCleanup(func() { _ = k8sClient.Delete(ctx, obj) })
	})
})

var _ = Describe("VAP CRD-Specific: DatabaseSecretEngineStaticRole", func() {

	It("should reject both rotationPeriod and rotationSchedule", func() {
		rotPeriod := 3600
		obj := &DatabaseSecretEngineStaticRole{
			ObjectMeta: metav1.ObjectMeta{GenerateName: "vap-dbsr-", Namespace: "default"},
			Spec: DatabaseSecretEngineStaticRoleSpec{
				Authentication: defaultAuth(),
				Path:           "database",
				DBSEStaticRole: DBSEStaticRole{
					RotationPeriod:           &rotPeriod,
					RotationSchedule:         "0 0 * * SAT",
					PasswordCredentialConfig: &PasswordCredentialConfig{},
				},
			},
		}
		err := k8sClient.Create(ctx, obj)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("mutually exclusive"))
	})

	It("should reject neither rotationPeriod nor rotationSchedule", func() {
		obj := &DatabaseSecretEngineStaticRole{
			ObjectMeta: metav1.ObjectMeta{GenerateName: "vap-dbsr-", Namespace: "default"},
			Spec: DatabaseSecretEngineStaticRoleSpec{
				Authentication: defaultAuth(),
				Path:           "database",
				DBSEStaticRole: DBSEStaticRole{
					PasswordCredentialConfig: &PasswordCredentialConfig{},
				},
			},
		}
		err := k8sClient.Create(ctx, obj)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("one of spec.rotationPeriod or spec.rotationSchedule must be set"))
	})

	It("should allow valid rotationPeriod only", func() {
		rotPeriod := 3600
		obj := &DatabaseSecretEngineStaticRole{
			ObjectMeta: metav1.ObjectMeta{GenerateName: "vap-dbsr-ok-", Namespace: "default"},
			Spec: DatabaseSecretEngineStaticRoleSpec{
				Authentication: defaultAuth(),
				Path:           "database",
				DBSEStaticRole: DBSEStaticRole{
					RotationPeriod:           &rotPeriod,
					PasswordCredentialConfig: &PasswordCredentialConfig{},
				},
			},
		}
		Expect(k8sClient.Create(ctx, obj)).To(Succeed())
		DeferCleanup(func() { _ = k8sClient.Delete(ctx, obj) })
	})

	It("should reject adding rotationSchedule on update when rotationPeriod is set", func() {
		rotPeriod := 3600
		obj := &DatabaseSecretEngineStaticRole{
			ObjectMeta: metav1.ObjectMeta{GenerateName: "vap-dbsr-upd-", Namespace: "default"},
			Spec: DatabaseSecretEngineStaticRoleSpec{
				Authentication: defaultAuth(),
				Path:           "database",
				DBSEStaticRole: DBSEStaticRole{
					RotationPeriod:           &rotPeriod,
					PasswordCredentialConfig: &PasswordCredentialConfig{},
				},
			},
		}
		createAndFetch(obj)
		obj.Spec.RotationSchedule = "0 0 * * SAT"
		err := k8sClient.Update(ctx, obj)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("mutually exclusive"))
	})
})

var _ = Describe("VAP CRD-Specific: PKISecretEngineConfig", func() {

	It("should reject acme without cluster", func() {
		obj := &PKISecretEngineConfig{
			ObjectMeta: metav1.ObjectMeta{GenerateName: "vap-pki-", Namespace: "default"},
			Spec: PKISecretEngineConfigSpec{
				Authentication: defaultAuth(),
				Path:           "pki",
				PKIType:        PKIType{Type: "root", PrivateKeyType: "internal"},
				ACME:           &PKIConfigACME{Enabled: true},
			},
		}
		err := k8sClient.Create(ctx, obj)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("spec.acme requires spec.cluster"))
	})

	It("should reject adding acme without cluster on update", func() {
		obj := &PKISecretEngineConfig{
			ObjectMeta: metav1.ObjectMeta{GenerateName: "vap-pki-upd-", Namespace: "default"},
			Spec: PKISecretEngineConfigSpec{
				Authentication: defaultAuth(),
				Path:           "pki",
				PKIType:        PKIType{Type: "root", PrivateKeyType: "internal"},
			},
		}
		createAndFetch(obj)
		obj.Spec.ACME = &PKIConfigACME{Enabled: true}
		err := k8sClient.Update(ctx, obj)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("spec.acme requires spec.cluster"))
	})

	It("should allow acme with cluster", func() {
		obj := &PKISecretEngineConfig{
			ObjectMeta: metav1.ObjectMeta{GenerateName: "vap-pki-ok-", Namespace: "default"},
			Spec: PKISecretEngineConfigSpec{
				Authentication: defaultAuth(),
				Path:           "pki",
				PKIType:        PKIType{Type: "root", PrivateKeyType: "internal"},
				Cluster:        &PKIConfigCluster{Path: "https://vault.example.com/v1/pki"},
				ACME:           &PKIConfigACME{Enabled: true},
			},
		}
		Expect(k8sClient.Create(ctx, obj)).To(Succeed())
		DeferCleanup(func() { _ = k8sClient.Delete(ctx, obj) })
	})
})

var _ = Describe("VAP CRD-Specific: KubernetesAuthEngineRole", func() {

	It("should reject both targetNamespaceSelector and targetNamespaces", func() {
		obj := &KubernetesAuthEngineRole{
			ObjectMeta: metav1.ObjectMeta{GenerateName: "vap-kar-", Namespace: "default"},
			Spec: KubernetesAuthEngineRoleSpec{
				Authentication: defaultAuth(),
				Path:           "k8s",
				VRole:          VRole{Policies: []string{"default"}},
				TargetNamespaces: vaultutils.TargetNamespaceConfig{
					TargetNamespaceSelector: &metav1.LabelSelector{MatchLabels: map[string]string{"env": "test"}},
					TargetNamespaces:        []string{"default"},
				},
			},
		}
		err := k8sClient.Create(ctx, obj)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("only one of"))
	})

	It("should reject adding targetNamespaceSelector on update when targetNamespaces is set", func() {
		obj := &KubernetesAuthEngineRole{
			ObjectMeta: metav1.ObjectMeta{GenerateName: "vap-kar-upd-", Namespace: "default"},
			Spec: KubernetesAuthEngineRoleSpec{
				Authentication:   defaultAuth(),
				Path:             "k8s",
				VRole:            VRole{Policies: []string{"default"}},
				TargetNamespaces: vaultutils.TargetNamespaceConfig{TargetNamespaces: []string{"default"}},
			},
		}
		createAndFetch(obj)
		obj.Spec.TargetNamespaces.TargetNamespaceSelector = &metav1.LabelSelector{MatchLabels: map[string]string{"env": "test"}}
		err := k8sClient.Update(ctx, obj)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("only one of"))
	})

	It("should allow only targetNamespaces", func() {
		obj := &KubernetesAuthEngineRole{
			ObjectMeta: metav1.ObjectMeta{GenerateName: "vap-kar-ok-", Namespace: "default"},
			Spec: KubernetesAuthEngineRoleSpec{
				Authentication:   defaultAuth(),
				Path:             "k8s",
				VRole:            VRole{Policies: []string{"default"}},
				TargetNamespaces: vaultutils.TargetNamespaceConfig{TargetNamespaces: []string{"default"}},
			},
		}
		Expect(k8sClient.Create(ctx, obj)).To(Succeed())
		DeferCleanup(func() { _ = k8sClient.Delete(ctx, obj) })
	})
})

var _ = Describe("VAP CRD-Specific: TOTPSecretEngineKey", func() {

	It("should reject create without issuer", func() {
		obj := &TOTPSecretEngineKey{
			ObjectMeta: metav1.ObjectMeta{GenerateName: "vap-totp-", Namespace: "default"},
			Spec: TOTPSecretEngineKeySpec{
				Authentication: defaultAuth(),
				Path:           "totp",
				TOTPKeyConfig:  TOTPKeyConfig{Generate: true, AccountName: "test@example.com", Period: 30},
			},
		}
		err := k8sClient.Create(ctx, obj)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("spec.issuer is required"))
	})

	It("should reject create with key when generate is true", func() {
		obj := &TOTPSecretEngineKey{
			ObjectMeta: metav1.ObjectMeta{GenerateName: "vap-totp-", Namespace: "default"},
			Spec: TOTPSecretEngineKeySpec{
				Authentication: defaultAuth(),
				Path:           "totp",
				TOTPKeyConfig:  TOTPKeyConfig{Generate: true, Issuer: "Vault", AccountName: "test@example.com", Key: "JBSWY3DPEHPK3PXP", Period: 30},
			},
		}
		err := k8sClient.Create(ctx, obj)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("spec.key cannot be set when spec.generate is true"))
	})

	It("should reject keySize change on update", func() {
		obj := &TOTPSecretEngineKey{
			ObjectMeta: metav1.ObjectMeta{GenerateName: "vap-totp-", Namespace: "default"},
			Spec: TOTPSecretEngineKeySpec{
				Authentication: defaultAuth(),
				Path:           "totp",
				TOTPKeyConfig:  TOTPKeyConfig{Generate: true, Issuer: "Vault", AccountName: "test@example.com", Period: 30, KeySize: 20},
			},
		}
		createAndFetch(obj)
		obj.Spec.KeySize = 32
		err := k8sClient.Update(ctx, obj)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("spec.keySize cannot be updated"))
	})

	It("should allow valid generated key create", func() {
		obj := &TOTPSecretEngineKey{
			ObjectMeta: metav1.ObjectMeta{GenerateName: "vap-totp-ok-", Namespace: "default"},
			Spec: TOTPSecretEngineKeySpec{
				Authentication: defaultAuth(),
				Path:           "totp",
				TOTPKeyConfig:  TOTPKeyConfig{Generate: true, Issuer: "Vault", AccountName: "test@example.com", Period: 30},
			},
		}
		Expect(k8sClient.Create(ctx, obj)).To(Succeed())
		DeferCleanup(func() { _ = k8sClient.Delete(ctx, obj) })
	})
})
