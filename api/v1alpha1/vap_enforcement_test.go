package v1alpha1

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"

	vaultutils "github.com/redhat-cop/vault-config-operator/api/v1alpha1/utils"
)

var corev1GVK = schema.GroupVersionKind{Group: "", Version: "v1", Kind: "ConfigMap"}

var _ = Describe("VAP enforcement smoke tests", func() {

	Context("vco-immutable-path", func() {
		It("should reject an UPDATE that changes spec.path", func() {
			obj := &PKISecretEngineRole{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "vap-path-test",
					Namespace: "default",
				},
				Spec: PKISecretEngineRoleSpec{
					Path: "pki/original",
				},
			}
			Expect(k8sClient.Create(ctx, obj)).To(Succeed())

			fetched := &PKISecretEngineRole{}
			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Name:      obj.Name,
				Namespace: obj.Namespace,
			}, fetched)).To(Succeed())

			fetched.Spec.Path = "pki/changed"
			err := k8sClient.Update(ctx, fetched)
			Expect(err).To(HaveOccurred())
			Expect(strings.Contains(err.Error(), "spec.path cannot be updated")).To(BeTrue(),
				"expected VAP rejection for path change, got: %v", err)
		})
	})

	Context("vco-immutable-name", func() {
		It("should reject an UPDATE that changes spec.name", func() {
			obj := &PKISecretEngineRole{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "vap-name-test",
					Namespace: "default",
				},
				Spec: PKISecretEngineRoleSpec{
					Path: "pki/roles",
					Name: "original-name",
				},
			}
			Expect(k8sClient.Create(ctx, obj)).To(Succeed())

			fetched := &PKISecretEngineRole{}
			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Name:      obj.Name,
				Namespace: obj.Namespace,
			}, fetched)).To(Succeed())

			fetched.Spec.Name = "changed-name"
			err := k8sClient.Update(ctx, fetched)
			Expect(err).To(HaveOccurred())
			Expect(strings.Contains(err.Error(), "spec.name cannot be updated")).To(BeTrue(),
				"expected VAP rejection for name change, got: %v", err)
		})

		It("should reject adding spec.name when it was not originally set", func() {
			obj := &PKISecretEngineRole{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "vap-name-add-test",
					Namespace: "default",
				},
				Spec: PKISecretEngineRoleSpec{
					Path: "pki/roles",
				},
			}
			Expect(k8sClient.Create(ctx, obj)).To(Succeed())

			fetched := &PKISecretEngineRole{}
			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Name:      obj.Name,
				Namespace: obj.Namespace,
			}, fetched)).To(Succeed())

			fetched.Spec.Name = "newly-added-name"
			err := k8sClient.Update(ctx, fetched)
			Expect(err).To(HaveOccurred())
			Expect(strings.Contains(err.Error(), "spec.name cannot be updated")).To(BeTrue(),
				"expected VAP rejection for adding name, got: %v", err)
		})

		It("should allow an UPDATE when spec.name is not set on either old or new", func() {
			obj := &PKISecretEngineRole{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "vap-name-absent-test",
					Namespace: "default",
				},
				Spec: PKISecretEngineRoleSpec{
					Path: "pki/roles",
				},
			}
			Expect(k8sClient.Create(ctx, obj)).To(Succeed())

			fetched := &PKISecretEngineRole{}
			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Name:      obj.Name,
				Namespace: obj.Namespace,
			}, fetched)).To(Succeed())

			// Update a non-immutable field with name still absent
			if fetched.Labels == nil {
				fetched.Labels = map[string]string{}
			}
			fetched.Labels["test"] = "value"
			Expect(k8sClient.Update(ctx, fetched)).To(Succeed())
		})
	})

	Context("vco-target-namespace-guard", func() {
		It("should reject an UPDATE that changes targetNamespace", func() {
			obj := &PKISecretEngineRole{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "vap-tns-change-test",
					Namespace: "default",
				},
				Spec: PKISecretEngineRoleSpec{
					Path: "pki/roles",
					Authentication: vaultutils.KubeAuthConfiguration{
						TargetNamespace: "tenant-a",
					},
				},
			}
			Expect(k8sClient.Create(ctx, obj)).To(Succeed())

			fetched := &PKISecretEngineRole{}
			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Name:      obj.Name,
				Namespace: obj.Namespace,
			}, fetched)).To(Succeed())

			fetched.Spec.Authentication.TargetNamespace = "tenant-b"
			err := k8sClient.Update(ctx, fetched)
			Expect(err).To(HaveOccurred())
			Expect(strings.Contains(err.Error(), "spec.authentication cannot be updated to a different Vault namespace")).To(BeTrue(),
				"expected VAP rejection for targetNamespace change, got: %v", err)
		})

		It("should reject an UPDATE that adds targetNamespace", func() {
			obj := &PKISecretEngineRole{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "vap-tns-add-test",
					Namespace: "default",
				},
				Spec: PKISecretEngineRoleSpec{
					Path: "pki/roles",
				},
			}
			Expect(k8sClient.Create(ctx, obj)).To(Succeed())

			fetched := &PKISecretEngineRole{}
			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Name:      obj.Name,
				Namespace: obj.Namespace,
			}, fetched)).To(Succeed())

			fetched.Spec.Authentication.TargetNamespace = "tenant-a"
			err := k8sClient.Update(ctx, fetched)
			Expect(err).To(HaveOccurred())
			Expect(strings.Contains(err.Error(), "spec.authentication cannot be updated to a different Vault namespace")).To(BeTrue(),
				"expected VAP rejection for adding targetNamespace, got: %v", err)
		})

		It("should reject an UPDATE that changes namespace while targetNamespace is set", func() {
			obj := &PKISecretEngineRole{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "vap-tns-ns-change-test",
					Namespace: "default",
				},
				Spec: PKISecretEngineRoleSpec{
					Path: "pki/roles",
					Authentication: vaultutils.KubeAuthConfiguration{
						Namespace:       "org",
						TargetNamespace: "tenant-a",
					},
				},
			}
			Expect(k8sClient.Create(ctx, obj)).To(Succeed())

			fetched := &PKISecretEngineRole{}
			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Name:      obj.Name,
				Namespace: obj.Namespace,
			}, fetched)).To(Succeed())

			fetched.Spec.Authentication.Namespace = "org2"
			err := k8sClient.Update(ctx, fetched)
			Expect(err).To(HaveOccurred())
			Expect(strings.Contains(err.Error(), "spec.authentication cannot be updated to a different Vault namespace")).To(BeTrue(),
				"expected VAP rejection for namespace change with targetNamespace set, got: %v", err)
		})

		It("should allow an UPDATE when no targetNamespace is set on either old or new", func() {
			obj := &PKISecretEngineRole{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "vap-tns-absent-test",
					Namespace: "default",
				},
				Spec: PKISecretEngineRoleSpec{
					Path: "pki/roles",
				},
			}
			Expect(k8sClient.Create(ctx, obj)).To(Succeed())

			fetched := &PKISecretEngineRole{}
			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Name:      obj.Name,
				Namespace: obj.Namespace,
			}, fetched)).To(Succeed())

			fetched.Spec.Authentication.Namespace = "changed-ns"
			Expect(k8sClient.Update(ctx, fetched)).To(Succeed())
		})

		It("should reject an UPDATE that removes targetNamespace", func() {
			obj := &PKISecretEngineRole{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "vap-tns-remove-test",
					Namespace: "default",
				},
				Spec: PKISecretEngineRoleSpec{
					Path: "pki/roles",
					Authentication: vaultutils.KubeAuthConfiguration{
						TargetNamespace: "tenant-a",
					},
				},
			}
			Expect(k8sClient.Create(ctx, obj)).To(Succeed())

			fetched := &PKISecretEngineRole{}
			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Name:      obj.Name,
				Namespace: obj.Namespace,
			}, fetched)).To(Succeed())

			fetched.Spec.Authentication.TargetNamespace = ""
			err := k8sClient.Update(ctx, fetched)
			Expect(err).To(HaveOccurred())
			Expect(strings.Contains(err.Error(), "spec.authentication cannot be updated to a different Vault namespace")).To(BeTrue(),
				"expected VAP rejection for removing targetNamespace, got: %v", err)
		})

		It("should allow moving from namespace to targetNamespace when effective namespace is the same", func() {
			obj := &PKISecretEngineRole{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "vap-tns-move-test",
					Namespace: "default",
				},
				Spec: PKISecretEngineRoleSpec{
					Path: "pki/roles",
					Authentication: vaultutils.KubeAuthConfiguration{
						Namespace: "org/tenant-a",
					},
				},
			}
			Expect(k8sClient.Create(ctx, obj)).To(Succeed())

			fetched := &PKISecretEngineRole{}
			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Name:      obj.Name,
				Namespace: obj.Namespace,
			}, fetched)).To(Succeed())

			fetched.Spec.Authentication.Namespace = "org"
			fetched.Spec.Authentication.TargetNamespace = "tenant-a"
			Expect(k8sClient.Update(ctx, fetched)).To(Succeed())
		})

		It("should use vaultNamespace from ConfigMap as fallback when namespace is empty", func() {
			cm := &unstructured.Unstructured{}
			cm.SetGroupVersionKind(corev1GVK)
			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Name:      "vco-vap-params",
				Namespace: "default",
			}, cm)).To(Succeed())
			cmData := cm.Object["data"].(map[string]interface{})
			cmData["vaultNamespace"] = "admin"
			Expect(k8sClient.Update(ctx, cm)).To(Succeed())

			obj := &PKISecretEngineRole{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "vap-tns-fallback-test",
					Namespace: "default",
				},
				Spec: PKISecretEngineRoleSpec{
					Path: "pki/roles",
					Authentication: vaultutils.KubeAuthConfiguration{
						TargetNamespace: "tenant-a",
					},
				},
			}
			Expect(k8sClient.Create(ctx, obj)).To(Succeed())

			fetched := &PKISecretEngineRole{}
			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Name:      obj.Name,
				Namespace: obj.Namespace,
			}, fetched)).To(Succeed())

			// Set namespace to "admin" explicitly — same effective namespace as fallback
			fetched.Spec.Authentication.Namespace = "admin"
			Expect(k8sClient.Update(ctx, fetched)).To(Succeed())

			// Restore ConfigMap
			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Name:      "vco-vap-params",
				Namespace: "default",
			}, cm)).To(Succeed())
			cmData = cm.Object["data"].(map[string]interface{})
			cmData["vaultNamespace"] = ""
			Expect(k8sClient.Update(ctx, cm)).To(Succeed())
		})

		It("should allow an UPDATE when CleansePath-equivalent paths resolve to the same namespace", func() {
			obj := &PKISecretEngineRole{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "vap-tns-cleanpath-test",
					Namespace: "default",
				},
				Spec: PKISecretEngineRoleSpec{
					Path: "pki/roles",
					Authentication: vaultutils.KubeAuthConfiguration{
						Namespace:       "org/",
						TargetNamespace: "tenant-a",
					},
				},
			}
			Expect(k8sClient.Create(ctx, obj)).To(Succeed())

			fetched := &PKISecretEngineRole{}
			Expect(k8sClient.Get(ctx, types.NamespacedName{
				Name:      obj.Name,
				Namespace: obj.Namespace,
			}, fetched)).To(Succeed())

			fetched.Spec.Authentication.Namespace = "org"
			Expect(k8sClient.Update(ctx, fetched)).To(Succeed())
		})
	})
})
