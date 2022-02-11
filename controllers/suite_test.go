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
	"os"
	"path/filepath"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/redhat-cop/operator-utils/pkg/util"
	redhatcopv1alpha1 "github.com/redhat-cop/vault-config-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	"sigs.k8s.io/controller-runtime/pkg/envtest/printer"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	//+kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var _ *rest.Config
var k8sClient client.Client
var testEnv *envtest.Environment
var ctx context.Context
var cancel context.CancelFunc

const (
	vaultTestNamespaceName  = "test-vault-config-operator"
	vaultAdminNamespaceName = "vault-admin"
)

var vaultTestNamespace *corev1.Namespace
var vaultAdminNamespace *corev1.Namespace

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecsWithDefaultAndCustomReporters(t,
		"Controller Integration Suite",
		[]Reporter{printer.NewlineReporter{}})
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	ctx, cancel = context.WithCancel(context.TODO())

	Expect(os.Setenv("USE_EXISTING_CLUSTER", "true")).To(Succeed())
	Expect(os.Setenv("VAULT_ADDR", "http://localhost:8200")).To(Succeed())

	By("bootstrapping test environment")
	testEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("..", "config", "crd", "bases")},
		ErrorIfCRDPathMissing: true,
	}

	cfg, err := testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	err = redhatcopv1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	//+kubebuilder:scaffold:scheme

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	vaultTestNamespace = &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: vaultTestNamespaceName,
			Labels: map[string]string{
				"database-engine-admin": "true",
			},
		},
	}

	vaultAdminNamespace = &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: vaultAdminNamespaceName,
		},
	}

	Expect(k8sClient.Create(ctx, vaultTestNamespace)).Should(Succeed())

	Expect(k8sClient.Create(ctx, vaultAdminNamespace)).Should(Succeed())

	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme: scheme.Scheme,
	})
	Expect(err).ToNot(HaveOccurred())

	err = (&VaultSecretReconciler{
		ReconcilerBase: util.NewReconcilerBase(mgr.GetClient(), mgr.GetScheme(), mgr.GetConfig(), mgr.GetEventRecorderFor("VaultSecret"), mgr.GetAPIReader()),
		Log:            ctrl.Log.WithName("controllers").WithName("VaultSecret"),
		ControllerName: "VaultSecret",
	}).SetupWithManager(mgr)
	Expect(err).ToNot(HaveOccurred())

	err = (&PasswordPolicyReconciler{
		ReconcilerBase: util.NewReconcilerBase(mgr.GetClient(), mgr.GetScheme(), mgr.GetConfig(), mgr.GetEventRecorderFor("PasswordPolicy"), mgr.GetAPIReader()),
		Log:            ctrl.Log.WithName("controllers").WithName("PasswordPolicy"),
		ControllerName: "PasswordPolicy",
	}).SetupWithManager(mgr)
	Expect(err).ToNot(HaveOccurred())

	err = (&PolicyReconciler{
		ReconcilerBase: util.NewReconcilerBase(mgr.GetClient(), mgr.GetScheme(), mgr.GetConfig(), mgr.GetEventRecorderFor("Policy"), mgr.GetAPIReader()),
		Log:            ctrl.Log.WithName("controllers").WithName("Policy"),
		ControllerName: "Policy",
	}).SetupWithManager(mgr)
	Expect(err).ToNot(HaveOccurred())

	err = (&KubernetesAuthEngineRoleReconciler{
		ReconcilerBase: util.NewReconcilerBase(mgr.GetClient(), mgr.GetScheme(), mgr.GetConfig(), mgr.GetEventRecorderFor("KubernetesAuthEngineRole"), mgr.GetAPIReader()),
		Log:            ctrl.Log.WithName("controllers").WithName("KubernetesAuthEngineRole"),
		ControllerName: "KubernetesAuthEngineRole",
	}).SetupWithManager(mgr)
	Expect(err).ToNot(HaveOccurred())

	err = (&SecretEngineMountReconciler{
		ReconcilerBase: util.NewReconcilerBase(mgr.GetClient(), mgr.GetScheme(), mgr.GetConfig(), mgr.GetEventRecorderFor("SecretEngineMount"), mgr.GetAPIReader()),
		Log:            ctrl.Log.WithName("controllers").WithName("SecretEngineMount"),
		ControllerName: "SecretEngineMount",
	}).SetupWithManager(mgr)
	Expect(err).ToNot(HaveOccurred())

	err = (&RandomSecretReconciler{
		ReconcilerBase: util.NewReconcilerBase(mgr.GetClient(), mgr.GetScheme(), mgr.GetConfig(), mgr.GetEventRecorderFor("RandomSecret"), mgr.GetAPIReader()),
		Log:            ctrl.Log.WithName("controllers").WithName("RandomSecret"),
		ControllerName: "RandomSecret",
	}).SetupWithManager(mgr)
	Expect(err).ToNot(HaveOccurred())

	go func() {
		defer GinkgoRecover()
		err = mgr.Start(ctx)
		Expect(err).ToNot(HaveOccurred(), "failed to run manager")
	}()
}, 60)

var _ = AfterSuite(func() {
	Expect(k8sClient.Delete(ctx, vaultTestNamespace)).Should(Succeed())
	Expect(k8sClient.Delete(ctx, vaultAdminNamespace)).Should(Succeed())

	cancel()
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})
