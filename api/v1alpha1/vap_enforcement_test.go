package v1alpha1

import (
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

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
})
