//go:build integration
// +build integration

package controllers

import (
	"encoding/json"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	redhatcopv1alpha1 "github.com/redhat-cop/vault-config-operator/api/v1alpha1"
	"github.com/redhat-cop/vault-config-operator/controllers/vaultresourcecontroller"

	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("DatabaseSecretEngine controllers", Ordered, func() {

	timeout := 120 * time.Second
	interval := 2 * time.Second

	var pgSecret *corev1.Secret
	var mountInstance *redhatcopv1alpha1.SecretEngineMount
	var configInstance *redhatcopv1alpha1.DatabaseSecretEngineConfig
	var roleInstance *redhatcopv1alpha1.DatabaseSecretEngineRole

	AfterAll(func() {
		if roleInstance != nil {
			k8sIntegrationClient.Delete(ctx, roleInstance) //nolint:errcheck
		}
		if configInstance != nil {
			k8sIntegrationClient.Delete(ctx, configInstance) //nolint:errcheck
		}
		if mountInstance != nil {
			k8sIntegrationClient.Delete(ctx, mountInstance) //nolint:errcheck
		}
		if pgSecret != nil {
			k8sIntegrationClient.Delete(ctx, pgSecret) //nolint:errcheck
		}
	})

	Context("When creating prerequisite resources", func() {
		It("Should create the PostgreSQL credentials secret and database engine mount", func() {

			By("Creating the PostgreSQL credentials K8s Secret")
			pgSecret = &corev1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-db-pg-creds",
					Namespace: vaultAdminNamespaceName,
				},
				StringData: map[string]string{
					"username": "postgres",
					"password": "testpassword123",
				},
			}
			Expect(k8sIntegrationClient.Create(ctx, pgSecret)).Should(Succeed())

			By("Loading and creating the SecretEngineMount fixture")
			var err error
			mountInstance, err = decoder.GetSecretEngineMountInstance("../test/databasesecretengine/test-db-mount.yaml")
			Expect(err).To(BeNil())
			mountInstance.Namespace = vaultAdminNamespaceName
			Expect(k8sIntegrationClient.Create(ctx, mountInstance)).Should(Succeed())

			lookupKey := types.NamespacedName{Name: mountInstance.Name, Namespace: mountInstance.Namespace}
			created := &redhatcopv1alpha1.SecretEngineMount{}

			By("Waiting for ReconcileSuccessful=True")
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

			By("Verifying the mount exists in Vault")
			secret, err := vaultClient.Logical().Read("sys/mounts")
			Expect(err).To(BeNil())
			Expect(secret).NotTo(BeNil())
			_, exists := secret.Data["test-dbse/test-db-mount/"]
			Expect(exists).To(BeTrue(), "expected mount 'test-dbse/test-db-mount/' in sys/mounts")
		})
	})

	Context("When creating a DatabaseSecretEngineConfig", func() {
		It("Should write the database config to Vault", func() {

			By("Loading and creating the DatabaseSecretEngineConfig fixture")
			var err error
			configInstance, err = decoder.GetDatabaseSecretEngineConfigInstance("../test/databasesecretengine/test-db-config.yaml")
			Expect(err).To(BeNil())
			configInstance.Namespace = vaultAdminNamespaceName
			Expect(k8sIntegrationClient.Create(ctx, configInstance)).Should(Succeed())

			lookupKey := types.NamespacedName{Name: configInstance.Name, Namespace: configInstance.Namespace}
			created := &redhatcopv1alpha1.DatabaseSecretEngineConfig{}

			By("Waiting for ReconcileSuccessful=True")
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

			By("Verifying the config in Vault")
			secret, err := vaultClient.Logical().Read("test-dbse/test-db-mount/config/test-db-config")
			Expect(err).To(BeNil())
			Expect(secret).NotTo(BeNil())

			pluginName, ok := secret.Data["plugin_name"].(string)
			Expect(ok).To(BeTrue(), "expected plugin_name to be a string")
			Expect(pluginName).To(Equal("postgresql-database-plugin"))

			connDetails, ok := secret.Data["connection_details"].(map[string]interface{})
			Expect(ok).To(BeTrue(), "expected connection_details to be a map")
			Expect(connDetails["connection_url"]).To(Equal("postgresql://{{username}}:{{password}}@my-postgresql-database.test-vault-config-operator.svc:5432"))
			Expect(connDetails["username"]).To(Equal("postgres"))

			allowedRoles, ok := secret.Data["allowed_roles"].([]interface{})
			Expect(ok).To(BeTrue(), "expected allowed_roles to be []interface{}")
			Expect(allowedRoles).To(ContainElement("test-db-role"))
		})
	})

	Context("When creating a DatabaseSecretEngineRole", func() {
		It("Should create the role in Vault with correct database settings", func() {

			By("Loading and creating the DatabaseSecretEngineRole fixture")
			var err error
			roleInstance, err = decoder.GetDatabaseSecretEngineRoleInstance("../test/databasesecretengine/test-db-role.yaml")
			Expect(err).To(BeNil())
			roleInstance.Namespace = vaultAdminNamespaceName
			Expect(k8sIntegrationClient.Create(ctx, roleInstance)).Should(Succeed())

			lookupKey := types.NamespacedName{Name: roleInstance.Name, Namespace: roleInstance.Namespace}
			created := &redhatcopv1alpha1.DatabaseSecretEngineRole{}

			By("Waiting for ReconcileSuccessful=True")
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

			By("Verifying the role in Vault")
			secret, err := vaultClient.Logical().Read("test-dbse/test-db-mount/roles/test-db-role")
			Expect(err).To(BeNil())
			Expect(secret).NotTo(BeNil())

			dbName, ok := secret.Data["db_name"].(string)
			Expect(ok).To(BeTrue(), "expected db_name to be a string")
			Expect(dbName).To(Equal("test-db-config"))

			creationStatements, ok := secret.Data["creation_statements"].([]interface{})
			Expect(ok).To(BeTrue(), "expected creation_statements to be []interface{}")
			Expect(creationStatements).To(HaveLen(1))
			Expect(creationStatements[0]).To(Equal("CREATE ROLE \"{{name}}\" WITH LOGIN PASSWORD '{{password}}' VALID UNTIL '{{expiration}}';"))

			defaultTTL, ok := secret.Data["default_ttl"].(json.Number)
			Expect(ok).To(BeTrue(), "expected default_ttl to be json.Number")
			defaultTTLVal, err := defaultTTL.Int64()
			Expect(err).To(BeNil())
			Expect(defaultTTLVal).To(Equal(int64(3600)))

			maxTTL, ok := secret.Data["max_ttl"].(json.Number)
			Expect(ok).To(BeTrue(), "expected max_ttl to be json.Number")
			maxTTLVal, err := maxTTL.Int64()
			Expect(err).To(BeNil())
			Expect(maxTTLVal).To(Equal(int64(86400)))
		})
	})

	Context("When updating a DatabaseSecretEngineRole", func() {
		It("Should update the role in Vault and reflect updated ObservedGeneration", func() {

			Expect(roleInstance).NotTo(BeNil(), "expected role to be created before update phase")

			By("Recording initial ObservedGeneration from ReconcileSuccessful condition")
			lookupKey := types.NamespacedName{Name: roleInstance.Name, Namespace: roleInstance.Namespace}
			current := &redhatcopv1alpha1.DatabaseSecretEngineRole{}
			Expect(k8sIntegrationClient.Get(ctx, lookupKey, current)).Should(Succeed())
			var initialGeneration int64
			for _, condition := range current.Status.Conditions {
				if condition.Type == vaultresourcecontroller.ReconcileSuccessful && condition.Status == metav1.ConditionTrue {
					initialGeneration = condition.ObservedGeneration
					break
				}
			}
			Expect(initialGeneration).To(BeNumerically(">", 0))

			By("Updating maxTTL to 48h")
			current.Spec.DBSERole.MaxTTL = metav1.Duration{Duration: 48 * time.Hour}
			Expect(k8sIntegrationClient.Update(ctx, current)).Should(Succeed())

			By("Waiting for Vault to reflect the updated max_ttl")
			Eventually(func() bool {
				secret, err := vaultClient.Logical().Read("test-dbse/test-db-mount/roles/test-db-role")
				if err != nil || secret == nil {
					return false
				}
				maxTTL, ok := secret.Data["max_ttl"].(json.Number)
				if !ok {
					return false
				}
				val, err := maxTTL.Int64()
				if err != nil {
					return false
				}
				return val == int64(172800) // 48h in seconds
			}, timeout, interval).Should(BeTrue())

			By("Verifying ObservedGeneration increased on ReconcileSuccessful condition")
			updated := &redhatcopv1alpha1.DatabaseSecretEngineRole{}
			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, lookupKey, updated)
				if err != nil {
					return false
				}
				for _, condition := range updated.Status.Conditions {
					if condition.Type == vaultresourcecontroller.ReconcileSuccessful {
						return condition.ObservedGeneration > initialGeneration
					}
				}
				return false
			}, timeout, interval).Should(BeTrue())
		})
	})

	Context("When deleting DatabaseSecretEngine resources", func() {
		It("Should clean up role and config from Vault and remove all resources", func() {

			Expect(mountInstance).NotTo(BeNil(), "expected mount to be created before delete phase")
			Expect(configInstance).NotTo(BeNil(), "expected config to be created before delete phase")
			Expect(roleInstance).NotTo(BeNil(), "expected role to be created before delete phase")

			By("Deleting the role CR (IsDeletable=true)")
			Expect(k8sIntegrationClient.Delete(ctx, roleInstance)).Should(Succeed())
			roleLookupKey := types.NamespacedName{Name: roleInstance.Name, Namespace: roleInstance.Namespace}
			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, roleLookupKey, &redhatcopv1alpha1.DatabaseSecretEngineRole{})
				return apierrors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue())

			By("Verifying the role is removed from Vault")
			Eventually(func() bool {
				secret, err := vaultClient.Logical().Read("test-dbse/test-db-mount/roles/test-db-role")
				return err == nil && secret == nil
			}, timeout, interval).Should(BeTrue())

			By("Deleting the config CR (IsDeletable=true)")
			Expect(k8sIntegrationClient.Delete(ctx, configInstance)).Should(Succeed())
			configLookupKey := types.NamespacedName{Name: configInstance.Name, Namespace: configInstance.Namespace}
			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, configLookupKey, &redhatcopv1alpha1.DatabaseSecretEngineConfig{})
				return apierrors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue())

			By("Verifying the config is removed from Vault")
			Eventually(func() bool {
				secret, err := vaultClient.Logical().Read("test-dbse/test-db-mount/config/test-db-config")
				return err == nil && secret == nil
			}, timeout, interval).Should(BeTrue())

			By("Deleting the SecretEngineMount")
			Expect(k8sIntegrationClient.Delete(ctx, mountInstance)).Should(Succeed())
			mountLookupKey := types.NamespacedName{Name: mountInstance.Name, Namespace: mountInstance.Namespace}
			Eventually(func() bool {
				err := k8sIntegrationClient.Get(ctx, mountLookupKey, &redhatcopv1alpha1.SecretEngineMount{})
				return apierrors.IsNotFound(err)
			}, timeout, interval).Should(BeTrue())

			By("Verifying the mount is removed from Vault")
			Eventually(func() bool {
				secret, err := vaultClient.Logical().Read("sys/mounts")
				if err != nil || secret == nil {
					return false
				}
				_, exists := secret.Data["test-dbse/test-db-mount/"]
				return !exists
			}, timeout, interval).Should(BeTrue())

			By("Deleting the PostgreSQL credentials secret")
			Expect(k8sIntegrationClient.Delete(ctx, pgSecret)).Should(Succeed())
		})
	})
})
