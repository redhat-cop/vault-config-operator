package v1alpha1

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	vaultutils "github.com/redhat-cop/vault-config-operator/api/v1alpha1/utils"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("spec.authentication.targetNamespace", func() {
	newPolicy := func(namespace string, targetNamespace string) *Policy {
		return &Policy{
			ObjectMeta: metav1.ObjectMeta{GenerateName: "target-namespace-", Namespace: "default"},
			Spec: PolicySpec{
				Policy:         `path "secret/*" { capabilities = ["read"] }`,
				Authentication: vaultutils.KubeAuthConfiguration{Path: "kubernetes", Role: "test", Namespace: namespace, TargetNamespace: targetNamespace},
			},
		}
	}
	create := func(policy *Policy) error {
		err := k8sClient.Create(ctx, policy)
		if err == nil {
			DeferCleanup(func() { Expect(k8sClient.Delete(ctx, policy)).To(Succeed()) })
		}
		return err
	}

	It("rejects a value that is not a Vault namespace path", func() {
		for _, value := range []string{"tenant a", "org//tenant-a", "/"} {
			err := create(newPolicy("", value))
			Expect(apierrors.IsInvalid(err)).To(BeTrue(), "value %q: %v", value, err)
		}
	})

	It("cannot be updated to a different Vault namespace", func() {
		policy := newPolicy("", "org/tenant-a")
		Expect(create(policy)).To(Succeed())

		policy.Spec.Policy = `path "secret/*" { capabilities = ["list"] }`
		Expect(k8sClient.Update(ctx, policy)).To(Succeed())

		policy.Spec.Authentication.TargetNamespace = "org/tenant-b"
		err := k8sClient.Update(ctx, policy)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("spec.authentication.targetNamespace cannot be updated to a different Vault namespace"))
	})

	It("can be updated when the Vault namespace stays the same", func() {
		policy := newPolicy("org/tenant-a", "")
		Expect(create(policy)).To(Succeed())

		policy.Spec.Authentication.Namespace = "org"
		policy.Spec.Authentication.TargetNamespace = "tenant-a"
		Expect(k8sClient.Update(ctx, policy)).To(Succeed())
	})
})
