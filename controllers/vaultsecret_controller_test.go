package controllers

import (
	"context"
	"errors"
	"fmt"
	"log"
	"reflect"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	redhatcopv1alpha1 "github.com/redhat-cop/vault-config-operator/api/v1alpha1"

	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/types"

	"io/ioutil"
)

//TODO: Example: https://github.com/kubernetes-sigs/kubebuilder/blob/master/docs/book/src/cronjob-tutorial/testdata/project/controllers/cronjob_controller_test.go
// Define utility constants for object names and testing timeouts/durations and intervals.
const (
	vaultSecretName = "randomsecret-it"
	timeout         = time.Second * 10
	interval        = time.Millisecond * 250
)

var _ = Describe("VaultSecret controller", func() {

	Context(fmt.Sprintf("When creating the %v VaultSecret", vaultSecretName), func() {
		It("Should be Successful when created", func() {
			By("By creating a new VaultSecret")
			ctx := context.Background()

			instance, err := getVaultSecretInstance()

			Expect(err).To(BeNil())

			Expect(k8sClient.Create(ctx, instance)).Should(Succeed())

			lookupKey := types.NamespacedName{Name: vaultSecretName, Namespace: vaultTestNamespaceName}
			createdVaultSecret := &redhatcopv1alpha1.VaultSecret{}

			// We'll need to retry getting this newly created VaultSecret, given that creation may not immediately happen.
			Eventually(func() bool {
				err := k8sClient.Get(ctx, lookupKey, createdVaultSecret)
				if err != nil {
					return false
				}
				if createdVaultSecret.Status.LastVaultSecretUpdate == nil {
					return false
				}
				return true
			}, timeout, interval).Should(BeTrue())

		})
	})
})

var files = []string{
	"../test/password-policy.yaml",
	"../test/kv-engine-admin-policy.yaml",
	"../test/secret-writer-policy.yaml",
	"../test/kv-engine-admin-role.yaml",
	"../test/secret-writer-role.yaml",
	"../test/kv-secret-engine.yaml",
	"../test/random-secret.yaml",
}

func decodeFile(filename string) (runtime.Object, *schema.GroupVersionKind, error) {
	scheme := runtime.NewScheme()
	redhatcopv1alpha1.AddToScheme(scheme)
	decode := serializer.NewCodecFactory(scheme).UniversalDeserializer().Decode

	stream, ferr := ioutil.ReadFile(filename)
	if ferr != nil {
		log.Fatal(ferr)
	}
	return decode(stream, nil, nil)
}

func getVaultSecretInstance() (*redhatcopv1alpha1.VaultSecret, error) {

	obj, groupKindVersion, err := decodeFile("../test/vaultsecret/vaultsecret-randomsecret.yaml")
	if err != nil {
		return nil, err
	}

	kind := reflect.TypeOf(redhatcopv1alpha1.VaultSecret{}).Name()
	if groupKindVersion.Kind == kind {
		o := obj.(*redhatcopv1alpha1.VaultSecret)
		o.ObjectMeta.Namespace = vaultTestNamespaceName
		return o, nil
	}

	return nil, errors.New("Failed to decode")

	// duration, _ := time.ParseDuration("3m0s")

	// return &redhatcopv1alpha1.VaultSecret{
	// 	TypeMeta: metav1.TypeMeta{
	// 		APIVersion: "redhatcop.redhat.io/v1alpha1",
	// 		Kind:       "VaultSecret",
	// 	},
	// 	ObjectMeta: metav1.ObjectMeta{
	// 		Name:      vaultSecretName,
	// 		Namespace: vaultTestNamespaceName,
	// 	},
	// 	Spec: redhatcopv1alpha1.VaultSecretSpec{
	// 		VaultSecretDefinitions: []redhatcopv1alpha1.VaultSecretDefinition{
	// 			{
	// 				Name: "randomsecret",
	// 				Authentication: redhatcopv1alpha1.KubeAuthConfiguration{
	// 					Path: "kubernetes",
	// 					Role: "secret-reader",
	// 					ServiceAccount: &corev1.LocalObjectReference{
	// 						Name: "default",
	// 					},
	// 				},
	// 				Path: "test-vault-config-operator/kv/randomsecret-password",
	// 			},
	// 			{
	// 				Name: "anotherrandomsecret",
	// 				Authentication: redhatcopv1alpha1.KubeAuthConfiguration{
	// 					Path: "kubernetes",
	// 					Role: "secret-reader",
	// 					ServiceAccount: &corev1.LocalObjectReference{
	// 						Name: "default",
	// 					},
	// 				},
	// 				Path: "test-vault-config-operator/kv/another-password",
	// 			},
	// 		},
	// 		TemplatizedK8sSecret: redhatcopv1alpha1.TemplatizedK8sSecret{
	// 			Name: vaultSecretName,
	// 			StringData: map[string]string{
	// 				"password":        "{{ .randomsecret.password }}",
	// 				"anotherpassword": "{{ .anotherrandomsecret.password }}",
	// 			},
	// 			Type: "Opaque",
	// 			Labels: map[string]string{
	// 				"app": "test-vault-config-operator",
	// 			},
	// 			Annotations: map[string]string{
	// 				"refresh": "every-minute",
	// 			},
	// 		},
	// 		RefreshPeriod: &metav1.Duration{
	// 			Duration: duration,
	// 		},
	// 	},
	// }
}

func getPasswordPolicyInstance() (*redhatcopv1alpha1.PasswordPolicy, error) {
	obj, groupKindVersion, err := decodeFile("../test/vaultsecret/vaultsecret-randomsecret.yaml")
	if err != nil {
		return nil, err
	}

	kind := reflect.TypeOf(redhatcopv1alpha1.PasswordPolicy{}).Name()
	if groupKindVersion.Kind == kind {
		o := obj.(*redhatcopv1alpha1.PasswordPolicy)
		o.ObjectMeta.Namespace = vaultTestNamespaceName
		return o, nil
	}

	return nil, errors.New("Failed to decode")
}

func getPolicyInstance() (*redhatcopv1alpha1.Policy, error) {
	obj, groupKindVersion, err := decodeFile("../test/vaultsecret/vaultsecret-randomsecret.yaml")
	if err != nil {
		return nil, err
	}

	kind := reflect.TypeOf(redhatcopv1alpha1.Policy{}).Name()
	if groupKindVersion.Kind == kind {
		o := obj.(*redhatcopv1alpha1.Policy)
		o.ObjectMeta.Namespace = vaultTestNamespaceName
		return o, nil
	}

	return nil, errors.New("Failed to decode")
}

func getKubernetesAuthEngineRoleInstance() (*redhatcopv1alpha1.KubernetesAuthEngineRole, error) {
	obj, groupKindVersion, err := decodeFile("../test/vaultsecret/vaultsecret-randomsecret.yaml")
	if err != nil {
		return nil, err
	}

	kind := reflect.TypeOf(redhatcopv1alpha1.KubernetesAuthEngineRole{}).Name()
	if groupKindVersion.Kind == kind {
		o := obj.(*redhatcopv1alpha1.KubernetesAuthEngineRole)
		o.ObjectMeta.Namespace = vaultTestNamespaceName
		return o, nil
	}

	return nil, errors.New("Failed to decode")
}
