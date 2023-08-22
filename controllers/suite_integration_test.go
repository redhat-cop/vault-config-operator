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
package controllers

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"net/http"

	vault "github.com/hashicorp/vault/api"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	redhatcopv1alpha1 "github.com/redhat-cop/vault-config-operator/api/v1alpha1"
	controllertestutils "github.com/redhat-cop/vault-config-operator/controllers/controllertestutils"
	"github.com/redhat-cop/vault-config-operator/controllers/vaultresourcecontroller"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
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

	RunSpecs(t, "Controller Integration Suite")
}

var decoder = controllertestutils.NewDecoder()

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	ctx, cancel = context.WithCancel(context.TODO())

	Expect(os.Setenv("USE_EXISTING_CLUSTER", "true")).To(Succeed())

	vaultAddress, isSet := os.LookupEnv("VAULT_ADDR")
	if !isSet {
		Expect(os.Setenv("VAULT_ADDR", "http://localhost:8200")).To(Succeed())
	}

	_, err := http.Get(os.Getenv("VAULT_ADDR"))
	Expect(err).To(BeNil())

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
		CRDDirectoryPaths:     []string{filepath.Join("..", "config", "crd", "bases")},
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
		Scheme: scheme.Scheme,
	})
	Expect(err).ToNot(HaveOccurred())

	err = (&controllers.KubernetesAuthEngineRoleReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "KubernetesAuthEngineRole")}).SetupWithManager(mgr)
	Expect(err).ToNot(HaveOccurred())

	err = (&controllers.PolicyReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "Policy")}).SetupWithManager(mgr)
	Expect(err).ToNot(HaveOccurred())

	err = (&controllers.DatabaseSecretEngineConfigReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "DatabaseSecretEngineConfig")}).SetupWithManager(mgr)
	Expect(err).ToNot(HaveOccurred())

	err = (&controllers.DatabaseSecretEngineRoleReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "DatabaseSecretEngineRole")}).SetupWithManager(mgr)
	Expect(err).ToNot(HaveOccurred())

	err = (&controllers.SecretEngineMountReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "SecretEngineMount")}).SetupWithManager(mgr)
	Expect(err).ToNot(HaveOccurred())

	err = (&controllers.RandomSecretReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "RandomSecret")}).SetupWithManager(mgr)
	Expect(err).ToNot(HaveOccurred())

	err = (&controllers.AuthEngineMountReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "AuthEngineMount")}).SetupWithManager(mgr)
	Expect(err).ToNot(HaveOccurred())

	err = (&controllers.KubernetesAuthEngineConfigReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "KubernetesAuthEngineConfig")}).SetupWithManager(mgr)
	Expect(err).ToNot(HaveOccurred())

	err = (&controllers.LDAPAuthEngineConfigReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "LDAPAuthEngineConfig")}).SetupWithManager(mgr)
	Expect(err).ToNot(HaveOccurred())

	err = (&controllers.LDAPAuthEngineGroupReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "LDAPAuthEngineGroup")}).SetupWithManager(mgr)
	Expect(err).ToNot(HaveOccurred())

	err = (&controllers.JWTOIDCAuthEngineConfigReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "JWTOIDCAuthEngineConfig")}).SetupWithManager(mgr)
	Expect(err).ToNot(HaveOccurred())

	err = (&controllers.JWTOIDCAuthEngineRoleReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "JWTOIDCAuthEngineRole")}).SetupWithManager(mgr)
	Expect(err).ToNot(HaveOccurred())

	err = (&controllers.VaultSecretReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "VaultSecret")}).SetupWithManager(mgr)
	Expect(err).ToNot(HaveOccurred())

	err = (&controllers.PasswordPolicyReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "PasswordPolicy")}).SetupWithManager(mgr)
	Expect(err).ToNot(HaveOccurred())

	err = (&controllers.RabbitMQSecretEngineConfigReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "RabbitMQSecretEngineConfig")}).SetupWithManager(mgr)
	Expect(err).ToNot(HaveOccurred())

	err = (&controllers.RabbitMQSecretEngineRoleReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "RabbitMQSecretEngineRole")}).SetupWithManager(mgr)
	Expect(err).ToNot(HaveOccurred())

	err = (&controllers.PKISecretEngineConfigReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "PKISecretEngineConfig")}).SetupWithManager(mgr)
	Expect(err).ToNot(HaveOccurred())

	err = (&controllers.PKISecretEngineRoleReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "PKISecretEngineRole")}).SetupWithManager(mgr)
	Expect(err).ToNot(HaveOccurred())

	err = (&controllers.GitHubSecretEngineConfigReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "GitHubSecretEngineConfig")}).SetupWithManager(mgr)
	Expect(err).ToNot(HaveOccurred())

	err = (&controllers.GitHubSecretEngineRoleReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "GitHubSecretEngineRole")}).SetupWithManager(mgr)
	Expect(err).ToNot(HaveOccurred())

	err = (&controllers.QuaySecretEngineConfigReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "QuaySecretEngineConfig")}).SetupWithManager(mgr)
	Expect(err).ToNot(HaveOccurred())

	err = (&controllers.QuaySecretEngineRoleReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "QuaySecretEngineRole")}).SetupWithManager(mgr)
	Expect(err).ToNot(HaveOccurred())

	err = (&controllers.QuaySecretEngineStaticRoleReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "QuaySecretEngineStaticRole")}).SetupWithManager(mgr)
	Expect(err).ToNot(HaveOccurred())

	err = (&controllers.KubernetesSecretEngineConfigReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "KubernetesSecretEngineConfig")}).SetupWithManager(mgr)
	Expect(err).ToNot(HaveOccurred())

	err = (&controllers.KubernetesSecretEngineRoleReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "KubernetesSecretEngineRole")}).SetupWithManager(mgr)
	Expect(err).ToNot(HaveOccurred())

	err = (&controllers.DatabaseSecretEngineStaticRoleReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "DatabaseSecretEngineStaticRole")}).SetupWithManager(mgr)
	Expect(err).ToNot(HaveOccurred())

	err = (&controllers.GroupReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "Group")}).SetupWithManager(mgr)
	Expect(err).ToNot(HaveOccurred())

	err = (&controllers.GroupAliasReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "GroupAlias")}).SetupWithManager(mgr)
	Expect(err).ToNot(HaveOccurred())

	By(fmt.Sprintf("Creating the %v namespace", vaultAdminNamespaceName))
	vaultAdminNamespace = &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: vaultAdminNamespaceName,
		},
	}
	Expect(k8sIntegrationClient.Create(ctx, vaultAdminNamespace)).Should(Succeed())

	By(fmt.Sprintf("Creating the %v namespace", vaultTestNamespaceName))
	vaultTestNamespace = &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: vaultTestNamespaceName,
			Labels: map[string]string{
				"database-engine-admin": "true",
			},
		},
	}
	Expect(k8sIntegrationClient.Create(ctx, vaultTestNamespace)).Should(Succeed())

	go func() {
		defer GinkgoRecover()
		err = mgr.Start(ctx)
		Expect(err).ToNot(HaveOccurred(), "failed to run manager")
	}()
})

var _ = AfterSuite(func() {

	cancel()
	By("tearing down the test environment")
	err := testIntegrationEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})
