package v1alpha1

import (
	"context"
	"reflect"
	"strings"
	"testing"

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
		for _, value := range []string{"tenant a", "org//tenant-a", "/", "/tenant-a", "tenant-a/", ".", "..", "org/../tenant-b"} {
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
		Expect(err.Error()).To(ContainSubstring("cannot be updated to a different Vault namespace"))

		policy.Spec.Authentication.TargetNamespace = "org/tenant-a"
		policy.Spec.Authentication.Namespace = "admin"
		err = k8sClient.Update(ctx, policy)
		Expect(err).To(HaveOccurred())
		Expect(err.Error()).To(ContainSubstring("cannot be updated to a different Vault namespace"))
	})

	It("keeps the old behavior for a resource without targetNamespace", func() {
		policy := newPolicy("org", "")
		Expect(create(policy)).To(Succeed())

		policy.Spec.Authentication.Namespace = "org2"
		Expect(k8sClient.Update(ctx, policy)).To(Succeed())
	})

	It("can be updated when the Vault namespace stays the same", func() {
		policy := newPolicy("org/tenant-a", "")
		Expect(create(policy)).To(Succeed())

		policy.Spec.Authentication.Namespace = "org"
		policy.Spec.Authentication.TargetNamespace = "tenant-a"
		Expect(k8sClient.Update(ctx, policy)).To(Succeed())
	})
})

var targetNamespaceLockExemptKinds = map[string]string{
	"VaultSecret":                "only reads from Vault",
	"Audit":                      "has no webhook",
	"AuditRequestHeader":         "has no webhook",
	"Entity":                     "has no webhook",
	"EntityAlias":                "has no webhook",
	"RabbitMQSecretEngineConfig": "uses a raw admission handler, tested in rabbitmqsecretengineconfig_webhook_test.go",
}

func TestUpdateWebhooksRejectTargetNamespaceMove(t *testing.T) {
	t.Setenv("VAULT_NAMESPACE", "")
	apiPackage := reflect.TypeOf(Policy{}).PkgPath()
	checked := 0
	for kind, objType := range testScheme().KnownTypes(GroupVersion) {
		if objType.PkgPath() != apiPackage || strings.HasSuffix(kind, "List") {
			continue
		}
		if reason, ok := targetNamespaceLockExemptKinds[kind]; ok {
			t.Logf("%s is exempt: %s", kind, reason)
			continue
		}
		checked++
		t.Run(kind, func(t *testing.T) {
			oldObj, newObj := reflect.New(objType), reflect.New(objType)
			validateUpdate := newObj.MethodByName("ValidateUpdate")
			if !validateUpdate.IsValid() {
				t.Fatalf("%s has no ValidateUpdate: add the targetNamespace check, or add the kind to targetNamespaceLockExemptKinds", kind)
			}
			targetNamespace := newObj.Elem().FieldByName("Spec").FieldByName("Authentication").FieldByName("TargetNamespace")
			if !targetNamespace.IsValid() {
				t.Fatalf("%s has no spec.authentication.targetNamespace", kind)
			}
			targetNamespace.SetString("tenant-a")

			out := validateUpdate.Call([]reflect.Value{reflect.ValueOf(context.Background()), oldObj, newObj})
			err, _ := out[1].Interface().(error)
			if err == nil || !strings.Contains(err.Error(), "cannot be updated to a different Vault namespace") {
				t.Errorf("expected the targetNamespace check to reject the update, got %v", err)
			}
		})
	}
	if checked < 80 {
		t.Fatalf("checked %d kinds, expected at least 80", checked)
	}
}
