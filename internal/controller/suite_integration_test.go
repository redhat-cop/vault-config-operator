//go:build integration
// +build integration

/*
Copyright 2021.

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
package controller

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	vault "github.com/hashicorp/vault/api"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	redhatcopv1alpha1 "github.com/redhat-cop/vault-config-operator/api/v1alpha1"
	controllertestutils "github.com/redhat-cop/vault-config-operator/internal/controller/controllertestutils"
	"github.com/redhat-cop/vault-config-operator/internal/controller/vaultresourcecontroller"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	//+kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var _ *rest.Config
var k8sIntegrationClient client.Client
var testIntegrationEnv *envtest.Environment
var ctx context.Context
var cancel context.CancelFunc
var vaultTestNamespace *corev1.Namespace
var vaultAdminNamespace *corev1.Namespace
var vaultClient *vault.Client

const (
	vaultTestNamespaceName  = "test-vault-config-operator"
	vaultAdminNamespaceName = "vault-admin"
)

func TestIntegrationAPIs(t *testing.T) {
	RegisterFailHandler(Fail)
	_, _ = fmt.Fprintf(GinkgoWriter, "Starting integration test suite\n")
	RunSpecs(t, "Controller Integration Suite")
}

var decoder = controllertestutils.NewDecoder()

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	ctx, cancel = context.WithCancel(context.TODO())

	Expect(os.Setenv("USE_EXISTING_CLUSTER", "true")).To(Succeed())

	vaultAddress, isSet := os.LookupEnv("VAULT_ADDR")
	if !isSet {
		vaultPort := os.Getenv("VAULT_HOST_PORT")
		if vaultPort == "" {
			vaultPort = "8200"
		}
		Expect(os.Setenv("VAULT_ADDR", "http://127.0.0.1:"+vaultPort)).To(Succeed())
		vaultAddress = os.Getenv("VAULT_ADDR")
	}

	// Vault health check with timeout and retry.
	// The cloud-provider-kind LoadBalancer port mapping may be ready before
	// Vault is fully serving, so we retry with a short per-request timeout.
	httpClient := &http.Client{Timeout: 5 * time.Second}
	var err error
	for attempt := 1; attempt <= 12; attempt++ {
		_, err = httpClient.Get(os.Getenv("VAULT_ADDR"))
		if err == nil {
			break
		}
		_, _ = fmt.Fprintf(GinkgoWriter, "Vault not reachable (attempt %d/12): %v\n", attempt, err)
		time.Sleep(5 * time.Second)
	}
	Expect(err).To(BeNil(), "Vault at %s did not become reachable", os.Getenv("VAULT_ADDR"))

	vaultToken, isSet := os.LookupEnv("VAULT_TOKEN")
	Expect(isSet).To(BeTrue())

	config := vault.DefaultConfig()
	config.Address = vaultAddress
	vaultClient, err = vault.NewClient(config)
	Expect(err).To(BeNil())
	// Authenticate
	vaultClient.SetToken(vaultToken)

	By("bootstrapping test environment")
	testIntegrationEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("..", "..", "config", "crd", "bases")},
		ErrorIfCRDPathMissing: true,
	}

	cfg, err := testIntegrationEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	err = redhatcopv1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	//+kubebuilder:scaffold:scheme

	k8sIntegrationClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sIntegrationClient).NotTo(BeNil())

	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme:  scheme.Scheme,
		Metrics: metricsserver.Options{BindAddress: "0"},
	})
	Expect(err).ToNot(HaveOccurred())

	// --- Generic-migrated controllers ---
	for _, setup := range []struct {
		name string
		fn   func(ctrl.Manager) error
	}{
		{"KubernetesAuthEngineRole", SetupKubernetesAuthEngineRoleController},
		{"Policy", SetupPolicyController},
		{"DatabaseSecretEngineRole", SetupDatabaseSecretEngineRoleController},
		{"SecretEngineMount", SetupSecretEngineMountController},
		{"AuthEngineMount", SetupAuthEngineMountController},
		{"KubernetesAuthEngineConfig", SetupKubernetesAuthEngineConfigController},
		{"LDAPAuthEngineConfig", SetupLDAPAuthEngineConfigController},
		{"LDAPAuthEngineGroup", SetupLDAPAuthEngineGroupController},
		{"JWTOIDCAuthEngineConfig", SetupJWTOIDCAuthEngineConfigController},
		{"JWTOIDCAuthEngineRole", SetupJWTOIDCAuthEngineRoleController},
		{"PasswordPolicy", SetupPasswordPolicyController},
		{"RabbitMQSecretEngineRole", SetupRabbitMQSecretEngineRoleController},
		{"PKISecretEngineConfig", SetupPKISecretEngineConfigController},
		{"PKISecretEngineRole", SetupPKISecretEngineRoleController},
		{"GitHubSecretEngineConfig", SetupGitHubSecretEngineConfigController},
		{"GitHubSecretEngineRole", SetupGitHubSecretEngineRoleController},
		{"QuaySecretEngineConfig", SetupQuaySecretEngineConfigController},
		{"QuaySecretEngineRole", SetupQuaySecretEngineRoleController},
		{"QuaySecretEngineStaticRole", SetupQuaySecretEngineStaticRoleController},
		{"KubernetesSecretEngineConfig", SetupKubernetesSecretEngineConfigController},
		{"KubernetesSecretEngineRole", SetupKubernetesSecretEngineRoleController},
		{"SSHSecretEngineConfig", SetupSSHSecretEngineConfigController},
		{"SSHSecretEngineRole", SetupSSHSecretEngineRoleController},
		{"LDAPSecretEngineStaticRole", SetupLDAPSecretEngineStaticRoleController},
		{"LDAPSecretEngineDynamicRole", SetupLDAPSecretEngineDynamicRoleController},
		{"DatabaseSecretEngineStaticRole", SetupDatabaseSecretEngineStaticRoleController},
		{"Group", SetupGroupController},
		{"GroupAlias", SetupGroupAliasController},
		{"Entity", SetupEntityController},
		{"EntityAlias", SetupEntityAliasController},
		{"IdentityOIDCScope", SetupIdentityOIDCScopeController},
		{"IdentityOIDCAssignment", SetupIdentityOIDCAssignmentController},
		{"IdentityOIDCClient", SetupIdentityOIDCClientController},
		{"IdentityOIDCProvider", SetupIdentityOIDCProviderController},
		{"IdentityTokenConfig", SetupIdentityTokenConfigController},
		{"IdentityTokenKey", SetupIdentityTokenKeyController},
		{"IdentityTokenRole", SetupIdentityTokenRoleController},
		{"Audit", SetupAuditController},
		{"AuditRequestHeader", SetupAuditRequestHeaderController},
		{"NomadSecretEngineRole", SetupNomadSecretEngineRoleController},
		{"AppRoleAuthEngineRole", SetupAppRoleAuthEngineRoleController},
		{"UserpassAuthEngineUser", SetupUserpassAuthEngineUserController},
		{"CORSConfig", SetupCORSConfigController},
		{"KeyRotationConfig", SetupKeyRotationConfigController},
		{"RotationPolicy", SetupRotationPolicyController},
		{"UICustomHeader", SetupUICustomHeaderController},
		{"LoggerConfig", SetupLoggerConfigController},
	} {
		err = setup.fn(mgr)
		Expect(err).ToNot(HaveOccurred())
	}

	// --- Custom controllers (Category 5 — hand-written Reconcile) ---
	err = (&DatabaseSecretEngineConfigReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "DatabaseSecretEngineConfig")}).SetupWithManager(mgr)
	Expect(err).ToNot(HaveOccurred())

	err = (&RandomSecretReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "RandomSecret")}).SetupWithManager(mgr)
	Expect(err).ToNot(HaveOccurred())

	err = (&VaultSecretReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "VaultSecret")}).SetupWithManager(mgr)
	Expect(err).ToNot(HaveOccurred())

	err = (&RabbitMQSecretEngineConfigReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "RabbitMQSecretEngineConfig")}).SetupWithManager(mgr)
	Expect(err).ToNot(HaveOccurred())

	err = (&TransitSecretEngineKeyReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "TransitSecretEngineKey")}).SetupWithManager(mgr)
	Expect(err).ToNot(HaveOccurred())

	err = (&TOTPSecretEngineKeyReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "TOTPSecretEngineKey")}).SetupWithManager(mgr)
	Expect(err).ToNot(HaveOccurred())

	err = (&NomadSecretEngineConfigReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "NomadSecretEngineConfig")}).SetupWithManager(mgr)
	Expect(err).ToNot(HaveOccurred())

	err = (&LDAPSecretEngineConfigReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "LDAPSecretEngineConfig")}).SetupWithManager(mgr)
	Expect(err).ToNot(HaveOccurred())

	// Temporarily disabled: RateLimitQuota creates a global quota that poisons
	// Vault on repeated runs (conflicting quota definitions detected).
	// err = SetupRateLimitQuotaController(mgr)
	// Expect(err).ToNot(HaveOccurred())

	// Temporarily disabled: QuotaConfig shares the same conflict issue.
	// err = SetupQuotaConfigController(mgr)
	// Expect(err).ToNot(HaveOccurred())

	By(fmt.Sprintf("Creating the %v namespace", vaultAdminNamespaceName))
	vaultAdminNamespace = &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: vaultAdminNamespaceName,
		},
	}
	err = k8sIntegrationClient.Create(ctx, vaultAdminNamespace)
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			Expect(k8sIntegrationClient.Get(ctx, types.NamespacedName{Name: vaultAdminNamespaceName}, vaultAdminNamespace)).Should(Succeed())
		} else {
			Expect(err).NotTo(HaveOccurred())
		}
	}

	By(fmt.Sprintf("Creating the %v namespace", vaultTestNamespaceName))
	vaultTestNamespace = &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: vaultTestNamespaceName,
			Labels: map[string]string{
				"database-engine-admin": "true",
			},
		},
	}
	err = k8sIntegrationClient.Create(ctx, vaultTestNamespace)
	if err != nil {
		if apierrors.IsAlreadyExists(err) {
			Expect(k8sIntegrationClient.Get(ctx, types.NamespacedName{Name: vaultTestNamespaceName}, vaultTestNamespace)).Should(Succeed())
			if vaultTestNamespace.Labels == nil {
				vaultTestNamespace.Labels = make(map[string]string)
			}
			if vaultTestNamespace.Labels["database-engine-admin"] != "true" {
				vaultTestNamespace.Labels["database-engine-admin"] = "true"
				Expect(k8sIntegrationClient.Update(ctx, vaultTestNamespace)).Should(Succeed())
			}
		} else {
			Expect(err).NotTo(HaveOccurred())
		}
	}

	go func() {
		defer GinkgoRecover()
		err = mgr.Start(ctx)
		Expect(err).ToNot(HaveOccurred(), "failed to run manager")
	}()
})

var _ = AfterSuite(func() {
	By("deleting test namespaces")
	if k8sIntegrationClient != nil {
		for _, nsName := range []string{vaultAdminNamespaceName, vaultTestNamespaceName} {
			ns := &corev1.Namespace{ObjectMeta: metav1.ObjectMeta{Name: nsName}}
			err := k8sIntegrationClient.Delete(ctx, ns, client.PropagationPolicy(metav1.DeletePropagationForeground))
			if err != nil && !apierrors.IsNotFound(err) {
				_, _ = fmt.Fprintf(GinkgoWriter, "Warning: failed to delete namespace %s: %v\n", nsName, err)
			}
		}
		By("waiting for test namespace deletion to complete")
		for _, nsName := range []string{vaultAdminNamespaceName, vaultTestNamespaceName} {
			Eventually(func() bool {
				ns := &corev1.Namespace{}
				err := k8sIntegrationClient.Get(ctx, types.NamespacedName{Name: nsName}, ns)
				return apierrors.IsNotFound(err)
			}, "60s", "2s").Should(BeTrue(), fmt.Sprintf("namespace %s was not fully deleted", nsName))
		}
	}

	cancel()
	By("tearing down the test environment")
	err := testIntegrationEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})
