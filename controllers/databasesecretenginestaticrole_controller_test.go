//go:build integration
// +build integration

package controllers

import (
	"encoding/json"
	"fmt"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	redhatcopv1alpha1 "github.com/redhat-cop/vault-config-operator/api/v1alpha1"
	"github.com/redhat-cop/vault-config-operator/controllers/vaultresourcecontroller"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("DatabaseSecretEngineStaticRole controller", func() {

	timeout := time.Second * 120
	interval := time.Second * 2

	Context("When preparing a Database Secret Engine", func() {
		It("Should create a Database Secret Engine when created", func() {
			By("By creating new Policies")
			pInstance, err := decoder.GetPolicyInstance("../test/databasesecretengine/database-engine-admin-policy.yaml")
			Expect(err).To(BeNil())
			pInstance.Namespace = vaultAdminNamespaceName
			Expect(k8sIntegrationClient.Create(ctx, pInstance)).Should(Succeed())

			pLookupKey := types.NamespacedName{Name: pInstance.Name, Namespace: pInstance.Namespace}
			pCreated := &redhatcopv1alpha1.Policy{}

			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, pLookupKey, pCreated)
				if err != nil {
					return false
				}

				for _, condition := range pCreated.Status.Conditions {
					if condition.Type == vaultresourcecontroller.ReconcileSuccessful && condition.Status == metav1.ConditionTrue {
						return true
					}
				}

				return false
			}, timeout, interval).Should(BeTrue())

			kaerInstance, err := decoder.GetKubernetesAuthEngineRoleInstance("../test/databasesecretengine/database-secret-engine-auth-role.yaml")
			Expect(err).To(BeNil())
			kaerInstance.Namespace = vaultAdminNamespaceName
			Expect(k8sIntegrationClient.Create(ctx, kaerInstance)).Should(Succeed())

			kaerLookupKey := types.NamespacedName{Name: kaerInstance.Name, Namespace: kaerInstance.Namespace}
			kaerCreated := &redhatcopv1alpha1.KubernetesAuthEngineRole{}

			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, kaerLookupKey, kaerCreated)
				if err != nil {
					return false
				}

				for _, condition := range kaerCreated.Status.Conditions {
					if condition.Type == vaultresourcecontroller.ReconcileSuccessful && condition.Status == metav1.ConditionTrue {
						return true
					}
				}

				return false
			}, timeout, interval).Should(BeTrue())

			By("By creating a new SecretEngineMount")

			semInstance, err := decoder.GetSecretEngineMountInstance("../test/database-secret-engine.yaml")
			Expect(err).To(BeNil())
			semInstance.Namespace = vaultTestNamespaceName
			Expect(k8sIntegrationClient.Create(ctx, semInstance)).Should(Succeed())

			semLookupKey := types.NamespacedName{Name: semInstance.Name, Namespace: semInstance.Namespace}
			semCreated := &redhatcopv1alpha1.SecretEngineMount{}

			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, semLookupKey, semCreated)
				if err != nil {
					return false
				}

				for _, condition := range semCreated.Status.Conditions {
					if condition.Type == vaultresourcecontroller.ReconcileSuccessful && condition.Status == metav1.ConditionTrue {
						return true
					}
				}

				return false
			}, timeout, interval).Should(BeTrue())

		})
	})

	Context("When creating a DatabaseSecretEngine", func() {
		It("Should configure the engine for the specific path when created", func() {
			By("By creating a new PasswordPolicy")
			ppInstance, err := decoder.GetPasswordPolicyInstance("../test/databasesecretengine/password-policy.yaml")
			Expect(err).To(BeNil())
			ppInstance.Namespace = vaultAdminNamespaceName
			Expect(k8sIntegrationClient.Create(ctx, ppInstance)).Should(Succeed())

			pplookupKey := types.NamespacedName{Name: ppInstance.Name, Namespace: ppInstance.Namespace}
			ppCreated := &redhatcopv1alpha1.PasswordPolicy{}

			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, pplookupKey, ppCreated)
				if err != nil {
					return false
				}

				for _, condition := range ppCreated.Status.Conditions {
					if condition.Type == vaultresourcecontroller.ReconcileSuccessful && condition.Status == metav1.ConditionTrue {
						return true
					}
				}

				return false
			}, timeout, interval).Should(BeTrue())

			By("By creating a new SecretEngineMount")

			semInstance, err := decoder.GetSecretEngineMountInstance("../test/databasesecretengine/database-kv-engine-mount.yaml")
			Expect(err).To(BeNil())
			semInstance.Namespace = vaultTestNamespaceName
			Expect(k8sIntegrationClient.Create(ctx, semInstance)).Should(Succeed())

			semLookupKey := types.NamespacedName{Name: semInstance.Name, Namespace: semInstance.Namespace}
			semCreated := &redhatcopv1alpha1.SecretEngineMount{}

			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, semLookupKey, semCreated)
				if err != nil {
					return false
				}

				for _, condition := range semCreated.Status.Conditions {
					if condition.Type == vaultresourcecontroller.ReconcileSuccessful && condition.Status == metav1.ConditionTrue {
						return true
					}
				}

				return false
			}, timeout, interval).Should(BeTrue())

			By("By creating a new RandomSecret")
			instance, err := decoder.GetRandomSecretInstance("../test/databasesecretengine/database-random-secret.yaml")
			Expect(err).To(BeNil())

			instance.Namespace = vaultTestNamespaceName
			Expect(k8sIntegrationClient.Create(ctx, instance)).Should(Succeed())

			lookupKey := types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}
			created := &redhatcopv1alpha1.RandomSecret{}

			// We'll need to retry getting this newly created RandomSecret, given that creation may not immediately happen.
			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, lookupKey, created)
				if err != nil {
					return false
				}

				for _, condition := range created.Status.Conditions {
					if condition.Type == vaultresourcecontroller.ReconcileSuccessful && condition.Status == metav1.ConditionTrue {
						return true
					}
				}

				return false
			}, timeout, interval).Should(BeTrue())

			By("Creating PostgreSQL root credentials secret")
			pgSecret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "postgresql-root-credentials",
					Namespace: vaultTestNamespaceName,
				},
				Type: corev1.SecretTypeBasicAuth,
				StringData: map[string]string{
					"username": "postgres",
					"password": "testpassword123",
				},
			}
			Expect(k8sIntegrationClient.Create(ctx, pgSecret)).Should(Succeed())

			By("By creating a database secret engine config")
			rsInstance, err := decoder.GetDatabaseSecretEngineConfigInstance("../test/databasesecretengine/database-engine-config.yaml")
			Expect(err).To(BeNil())
			rsInstance.Namespace = vaultTestNamespaceName
			Expect(k8sIntegrationClient.Create(ctx, rsInstance)).Should(Succeed())

			rslookupKey := types.NamespacedName{Name: rsInstance.Name, Namespace: rsInstance.Namespace}
			rsCreated := &redhatcopv1alpha1.DatabaseSecretEngineConfig{}

			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, rslookupKey, rsCreated)
				if err != nil {
					return false
				}

				for _, condition := range rsCreated.Status.Conditions {
					if condition.Type == vaultresourcecontroller.ReconcileSuccessful && condition.Status == metav1.ConditionTrue {
						return true
					}
				}

				return false
			}, timeout, interval).Should(BeTrue())

		})
	})

	Context("When creating a DatabaseSecretEngineStaticRole", func() {
		It("Should create the static role in Vault", func() {
			By("Creating a DatabaseSecretEngineStaticRole")
			srInstance, err := decoder.GetDatabaseSecretEngineStaticRoleInstance("../test/database-engine-read-only-static-role.yaml")
			Expect(err).To(BeNil())
			srInstance.Namespace = vaultTestNamespaceName
			Expect(k8sIntegrationClient.Create(ctx, srInstance)).Should(Succeed())

			srLookupKey := types.NamespacedName{Name: srInstance.Name, Namespace: srInstance.Namespace}
			srCreated := &redhatcopv1alpha1.DatabaseSecretEngineStaticRole{}

			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, srLookupKey, srCreated)
				if err != nil {
					return false
				}

				for _, condition := range srCreated.Status.Conditions {
					if condition.Type == vaultresourcecontroller.ReconcileSuccessful && condition.Status == metav1.ConditionTrue {
						return true
					}
				}

				return false
			}, timeout, interval).Should(BeTrue())

			By("Verifying the static role exists in Vault")
			vaultSecret, err := vaultClient.Logical().Read("test-vault-config-operator/database/static-roles/read-only-static")
			Expect(err).To(BeNil())
			Expect(vaultSecret).NotTo(BeNil())
			Expect(vaultSecret.Data["db_name"]).To(Equal("my-postgresql-database"))
			Expect(vaultSecret.Data["username"]).To(Equal("helloworld"))
		})
	})

	Context("When updating a DatabaseSecretEngineStaticRole", func() {
		It("Should update the static role in Vault and reflect updated ObservedGeneration", func() {

			By("Verifying initial Vault state for the static role")
			initialSecret, err := vaultClient.Logical().Read("test-vault-config-operator/database/static-roles/read-only-static")
			Expect(err).To(BeNil())
			Expect(initialSecret).NotTo(BeNil())
			initialStatements, ok := initialSecret.Data["rotation_statements"].([]interface{})
			Expect(ok).To(BeTrue(), "expected rotation_statements to be []interface{}")
			Expect(initialStatements).To(HaveLen(1))

			By("Recording initial ObservedGeneration")
			lookupKey := types.NamespacedName{Name: "read-only-static", Namespace: vaultTestNamespaceName}
			created := &redhatcopv1alpha1.DatabaseSecretEngineStaticRole{}
			Expect(k8sIntegrationClient.Get(ctx, lookupKey, created)).Should(Succeed())
			var initialObservedGeneration int64
			for _, condition := range created.Status.Conditions {
				if condition.Type == vaultresourcecontroller.ReconcileSuccessful && condition.Status == metav1.ConditionTrue {
					initialObservedGeneration = condition.ObservedGeneration
					break
				}
			}
			Expect(initialObservedGeneration).To(BeNumerically(">", 0))

			By("Getting the latest DatabaseSecretEngineStaticRole before update")
			Expect(k8sIntegrationClient.Get(ctx, lookupKey, created)).Should(Succeed())

			By("Updating the rotation period")
			created.Spec.RotationPeriod = 7200
			Expect(k8sIntegrationClient.Update(ctx, created)).Should(Succeed())

			By("Waiting for Vault to reflect the updated rotation_period")
			Eventually(func() bool {
				secret, err := vaultClient.Logical().Read("test-vault-config-operator/database/static-roles/read-only-static")
				if err != nil || secret == nil {
					return false
				}
				rotationPeriod, ok := secret.Data["rotation_period"].(json.Number)
				if !ok {
					return false
				}
				val, err := rotationPeriod.Int64()
				if err != nil {
					return false
				}
				return val == 7200
			}, timeout, interval).Should(BeTrue())

			By("Verifying ObservedGeneration increased")
			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, lookupKey, created)
				if err != nil {
					return false
				}
				for _, condition := range created.Status.Conditions {
					if condition.Type == vaultresourcecontroller.ReconcileSuccessful && condition.Status == metav1.ConditionTrue {
						return condition.ObservedGeneration > initialObservedGeneration
					}
				}
				return false
			}, timeout, interval).Should(BeTrue())
		})
	})

	Context("When deleting a DatabaseSecretEngineStaticRole and Config", func() {
		It("Should delete from Vault", func() {

			By("Deleting DatabaseSecretEngineStaticRole")
			srInstance, err := decoder.GetDatabaseSecretEngineStaticRoleInstance("../test/database-engine-read-only-static-role.yaml")
			Expect(err).To(BeNil())
			srInstance.Namespace = vaultTestNamespaceName

			Expect(k8sIntegrationClient.Delete(ctx, srInstance)).Should(Succeed())

			Eventually(func() error {
				secret, err := vaultClient.Logical().Read("test-vault-config-operator/database/static-roles/read-only-static")
				if secret == nil {
					return nil
				}
				out, err := json.Marshal(secret)
				if err != nil {
					panic(err)
				}
				return fmt.Errorf("secret is not nil %s", string(out))
			}, timeout, interval).Should(Succeed())

			By("Deleting DatabaseSecretEngineConfig")

			dsecInstance, err := decoder.GetDatabaseSecretEngineConfigInstance("../test/databasesecretengine/database-engine-config.yaml")
			Expect(err).To(BeNil())
			dsecInstance.Namespace = vaultTestNamespaceName

			Expect(k8sIntegrationClient.Delete(ctx, dsecInstance)).Should(Succeed())

			Eventually(func() error {
				secret, err := vaultClient.Logical().Read(dsecInstance.GetPath())
				if secret == nil {
					return nil
				}
				out, err := json.Marshal(secret)
				if err != nil {
					panic(err)
				}
				return fmt.Errorf("secret is not nil %s", string(out))
			}, timeout, interval).Should(Succeed())

			By("Deleting RandomSecret")
			rsInstance, err := decoder.GetRandomSecretInstance("../test/databasesecretengine/database-random-secret.yaml")
			Expect(err).To(BeNil())
			rsInstance.Namespace = vaultTestNamespaceName

			Expect(k8sIntegrationClient.Delete(ctx, rsInstance)).Should(Succeed())

			Eventually(func() error {
				secret, err := vaultClient.Logical().Read(rsInstance.GetPath())
				if secret == nil {
					return nil
				}
				out, err := json.Marshal(secret)
				if err != nil {
					panic(err)
				}
				return fmt.Errorf("secret is not nil %s", string(out))
			}, timeout, interval).Should(Succeed())

			By("Deleting KV SecretEngineMount")

			semKvDbInstance, err := decoder.GetSecretEngineMountInstance("../test/databasesecretengine/database-kv-engine-mount.yaml")
			Expect(err).To(BeNil())
			semKvDbInstance.Namespace = vaultTestNamespaceName

			Expect(k8sIntegrationClient.Delete(ctx, semKvDbInstance)).Should(Succeed())

			Eventually(func() error {
				secret, err := vaultClient.Logical().Read(semKvDbInstance.GetPath())
				if secret == nil {
					return nil
				}
				out, err := json.Marshal(secret)
				if err != nil {
					panic(err)
				}
				return fmt.Errorf("secret is not nil %s", string(out))
			}, timeout, interval).Should(Succeed())

			By("Deleting PasswordPolicy")
			ppInstance, err := decoder.GetPasswordPolicyInstance("../test/databasesecretengine/password-policy.yaml")
			Expect(err).To(BeNil())
			ppInstance.Namespace = vaultAdminNamespaceName

			Expect(k8sIntegrationClient.Delete(ctx, ppInstance)).Should(Succeed())

			Eventually(func() error {
				secret, err := vaultClient.Logical().Read(ppInstance.GetPath())
				if secret == nil {
					return nil
				}
				out, err := json.Marshal(secret)
				if err != nil {
					panic(err)
				}
				return fmt.Errorf("secret is not nil %s", string(out))
			}, timeout, interval).Should(Succeed())

			By("Deleting SecretEngineMount")

			semInstance, err := decoder.GetSecretEngineMountInstance("../test/database-secret-engine.yaml")
			Expect(err).To(BeNil())
			semInstance.Namespace = vaultTestNamespaceName

			Expect(k8sIntegrationClient.Delete(ctx, semInstance)).Should(Succeed())

			Eventually(func() error {
				secret, err := vaultClient.Logical().Read(semInstance.GetPath())
				if secret == nil {
					return nil
				}
				out, err := json.Marshal(secret)
				if err != nil {
					panic(err)
				}
				return fmt.Errorf("secret is not nil %s", string(out))
			}, timeout, interval).Should(Succeed())

			By("Deleting Policies")

			kaerInstance, err := decoder.GetKubernetesAuthEngineRoleInstance("../test/databasesecretengine/database-secret-engine-auth-role.yaml")
			Expect(err).To(BeNil())
			kaerInstance.Namespace = vaultAdminNamespaceName

			Expect(k8sIntegrationClient.Delete(ctx, kaerInstance)).Should(Succeed())

			Eventually(func() error {
				secret, err := vaultClient.Logical().Read(kaerInstance.GetPath())
				if secret == nil {
					return nil
				}
				out, err := json.Marshal(secret)
				if err != nil {
					panic(err)
				}
				return fmt.Errorf("secret is not nil %s", string(out))
			}, timeout, interval).Should(Succeed())

			pInstance, err := decoder.GetPolicyInstance("../test/databasesecretengine/database-engine-admin-policy.yaml")
			Expect(err).To(BeNil())
			pInstance.Namespace = vaultAdminNamespaceName

			Expect(k8sIntegrationClient.Delete(ctx, pInstance)).Should(Succeed())

			Eventually(func() error {
				secret, err := vaultClient.Logical().Read(pInstance.GetPath())
				if secret == nil {
					return nil
				}
				out, err := json.Marshal(secret)
				if err != nil {
					panic(err)
				}
				return fmt.Errorf("secret is not nil %s", string(out))
			}, timeout, interval).Should(Succeed())

			By("Deleting PostgreSQL root credentials secret")
			pgSecret := &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "postgresql-root-credentials",
					Namespace: vaultTestNamespaceName,
				},
			}
			Expect(k8sIntegrationClient.Delete(ctx, pgSecret)).Should(Succeed())
		})
	})

})
