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

package v1alpha1

import (
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/util/yaml"
	"k8s.io/client-go/rest"

	// +kubebuilder:scaffold:imports
	apimachineryruntime "k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

var cfg *rest.Config
var k8sClient client.Client
var testEnv *envtest.Environment
var ctx context.Context
var cancel context.CancelFunc

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Webhook Suite")
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	ctx, cancel = context.WithCancel(context.TODO())

	By("bootstrapping test environment")
	binaryAssetsDir := os.Getenv("KUBEBUILDER_ASSETS")
	if binaryAssetsDir == "" {
		binaryAssetsDir = filepath.Join("..", "..", "bin", "k8s", "k8s",
			fmt.Sprintf("1.31.0-%s-%s", runtime.GOOS, runtime.GOARCH))
	}
	testEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("..", "..", "config", "crd", "bases")},
		ErrorIfCRDPathMissing: false,

		// BinaryAssetsDirectory is only required if you want to run the tests directly
		// without calling the makefile target test. When we run make test,
		// KUBEBUILDER_ASSETS is set automatically via setup-envtest.
		BinaryAssetsDirectory: binaryAssetsDir,

		WebhookInstallOptions: envtest.WebhookInstallOptions{
			Paths: []string{filepath.Join("..", "..", "config", "webhook")},
		},
	}

	// Ensure the ValidatingAdmissionPolicy admission plugin is enabled.
	// VAP is GA since K8s 1.30 and enabled by default, but we explicitly
	// add it to guarantee enforcement in envtest.
	testEnv.ControlPlane.GetAPIServer().Configure().
		Append("enable-admission-plugins", "ValidatingAdmissionPolicy")

	var err error
	// cfg is defined in this file globally.
	cfg, err = testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	scheme := apimachineryruntime.NewScheme()
	err = AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())

	err = admissionv1.AddToScheme(scheme)
	Expect(err).NotTo(HaveOccurred())

	//+kubebuilder:scaffold:scheme

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	By("loading VAP manifests from config/vap/")
	vapDir := filepath.Join("..", "..", "config", "vap")
	vapFiles, err := filepath.Glob(filepath.Join(vapDir, "*.yaml"))
	Expect(err).NotTo(HaveOccurred())
	for _, vapFile := range vapFiles {
		if filepath.Base(vapFile) == "kustomization.yaml" {
			continue
		}
		data, err := os.ReadFile(vapFile)
		Expect(err).NotTo(HaveOccurred())
		reader := yaml.NewYAMLReader(bufio.NewReader(bytes.NewReader(data)))
		for {
			doc, err := reader.Read()
			if err == io.EOF {
				break
			}
			Expect(err).NotTo(HaveOccurred())
			if len(bytes.TrimSpace(doc)) == 0 {
				continue
			}
			obj := &unstructured.Unstructured{}
			Expect(yaml.NewYAMLOrJSONDecoder(bytes.NewReader(doc), 4096).Decode(obj)).To(Succeed())
			Expect(k8sClient.Create(ctx, obj)).To(Succeed())
		}
	}

	// Ginkgo randomizes specs, so Deny Its can run before any in-test readiness
	// probe. Wait here until representative VAPs actually enforce.
	waitForVAPEnforcement()

	// start webhook server using Manager
	webhookInstallOptions := &testEnv.WebhookInstallOptions
	mgr, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme: scheme,
		WebhookServer: webhook.NewServer(webhook.Options{
			Host:    webhookInstallOptions.LocalServingHost,
			Port:    webhookInstallOptions.LocalServingPort,
			CertDir: webhookInstallOptions.LocalServingCertDir,
		}),
		LeaderElection: false,
		Metrics:        metricsserver.Options{BindAddress: "0"},
	})
	Expect(err).NotTo(HaveOccurred())

	// Category A — defaulter + validator (uncovered validation)
	err = (&AWSAuthEngineClientConfig{}).SetupWebhookWithManager(mgr)
	Expect(err).NotTo(HaveOccurred())

	// Category B — defaulter-only (validator removed, VAP covers validation)
	err = (&CFAuthEngineConfig{}).SetupWebhookWithManager(mgr)
	Expect(err).NotTo(HaveOccurred())

	err = (&GCPSecretEngineConfig{}).SetupWebhookWithManager(mgr)
	Expect(err).NotTo(HaveOccurred())

	err = (&KubernetesAuthEngineConfig{}).SetupWebhookWithManager(mgr)
	Expect(err).NotTo(HaveOccurred())

	err = (&OktaAuthEngineConfig{}).SetupWebhookWithManager(mgr)
	Expect(err).NotTo(HaveOccurred())

	err = (&RADIUSAuthEngineConfig{}).SetupWebhookWithManager(mgr)
	Expect(err).NotTo(HaveOccurred())

	err = (&SSHSecretEngineConfig{}).SetupWebhookWithManager(mgr)
	Expect(err).NotTo(HaveOccurred())

	err = (&TerraformCloudSecretEngineConfig{}).SetupWebhookWithManager(mgr)
	Expect(err).NotTo(HaveOccurred())

	// Category C — validator-only (uncovered CRD-specific validation)
	err = (&AliCloudAuthEngineRole{}).SetupWebhookWithManager(mgr)
	Expect(err).NotTo(HaveOccurred())

	err = (&AppRoleAuthEngineRole{}).SetupWebhookWithManager(mgr)
	Expect(err).NotTo(HaveOccurred())

	err = (&AWSAuthEngineRole{}).SetupWebhookWithManager(mgr)
	Expect(err).NotTo(HaveOccurred())

	err = (&CORSConfig{}).SetupWebhookWithManager(mgr)
	Expect(err).NotTo(HaveOccurred())

	err = (&GCPSecretEngineRoleset{}).SetupWebhookWithManager(mgr)
	Expect(err).NotTo(HaveOccurred())

	err = (&GCPSecretEngineStaticAccount{}).SetupWebhookWithManager(mgr)
	Expect(err).NotTo(HaveOccurred())

	err = (&IdentityOIDCClient{}).SetupWebhookWithManager(mgr)
	Expect(err).NotTo(HaveOccurred())

	err = (&JWTOIDCAuthEngineRole{}).SetupWebhookWithManager(mgr)
	Expect(err).NotTo(HaveOccurred())

	err = (&KeyRotationConfig{}).SetupWebhookWithManager(mgr)
	Expect(err).NotTo(HaveOccurred())

	err = (&LDAPSecretEngineStaticRole{}).SetupWebhookWithManager(mgr)
	Expect(err).NotTo(HaveOccurred())

	err = (&LeaseCountQuota{}).SetupWebhookWithManager(mgr)
	Expect(err).NotTo(HaveOccurred())

	err = (&PKISecretEngineConfig{}).SetupWebhookWithManager(mgr)
	Expect(err).NotTo(HaveOccurred())

	err = (&RandomSecret{}).SetupWebhookWithManager(mgr)
	Expect(err).NotTo(HaveOccurred())

	err = (&RateLimitQuota{}).SetupWebhookWithManager(mgr)
	Expect(err).NotTo(HaveOccurred())

	err = (&UICustomHeader{}).SetupWebhookWithManager(mgr)
	Expect(err).NotTo(HaveOccurred())

	//+kubebuilder:scaffold:webhook

	go func() {
		defer GinkgoRecover()
		err = mgr.Start(ctx)
		Expect(err).NotTo(HaveOccurred())
	}()

	// wait for the webhook server to get ready
	dialer := &net.Dialer{Timeout: time.Second}
	addrPort := fmt.Sprintf("%s:%d", webhookInstallOptions.LocalServingHost, webhookInstallOptions.LocalServingPort)
	Eventually(func() error {
		conn, err := tls.DialWithDialer(dialer, "tcp", addrPort, &tls.Config{InsecureSkipVerify: true})
		if err != nil {
			return err
		}
		return conn.Close()
	}).Should(Succeed())

})

var _ = AfterSuite(func() {
	cancel()
	By("tearing down the test environment")
	err := testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})

// waitForVAPEnforcement blocks until key ValidatingAdmissionPolicies are
// compiled and denying invalid requests. Suite-level wait is required because
// Ginkgo randomizes specs and in-test Eventually probes are not guaranteed to
// run first. Functional deny probes are used (not status.observedGeneration)
// because enforcement is what the suite asserts.
func waitForVAPEnforcement() {
	By("waiting for TransitSecretEngineKey VAP to enforce")
	Eventually(func() error {
		obj := &TransitSecretEngineKey{
			ObjectMeta: metav1.ObjectMeta{GenerateName: "vap-ready-transit-", Namespace: "default"},
			Spec: TransitSecretEngineKeySpec{
				Authentication:   defaultAuth(),
				Path:             "transit",
				TransitKeyConfig: TransitKeyConfig{ConvergentEncryption: true, Derived: false},
			},
		}
		err := k8sClient.Create(ctx, obj)
		if err == nil {
			_ = k8sClient.Delete(ctx, obj)
			return fmt.Errorf("TransitSecretEngineKey VAP not enforcing yet")
		}
		return nil
	}, "60s", "1s").Should(Succeed())

	By("waiting for AuthEngineMount type-immutability VAP to enforce")
	Eventually(func() error {
		obj := &AuthEngineMount{
			ObjectMeta: metav1.ObjectMeta{GenerateName: "vap-ready-aem-", Namespace: "default"},
			Spec: AuthEngineMountSpec{
				Authentication: defaultAuth(),
				Path:           "auth/vap-ready",
				AuthMount:      AuthMount{Type: "approle", Description: "vap readiness"},
			},
		}
		if err := k8sClient.Create(ctx, obj); err != nil {
			return fmt.Errorf("create AuthEngineMount for readiness probe: %w", err)
		}
		defer func() { _ = k8sClient.Delete(ctx, obj) }()
		if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(obj), obj); err != nil {
			return fmt.Errorf("get AuthEngineMount for readiness probe: %w", err)
		}
		obj.Spec.AuthMount.Type = "oidc"
		if err := k8sClient.Update(ctx, obj); err == nil {
			return fmt.Errorf("AuthEngineMount type-immutability VAP not enforcing yet")
		}
		return nil
	}, "60s", "1s").Should(Succeed())

	By("waiting for immutable-path VAP to enforce")
	Eventually(func() error {
		obj := &AuthEngineMount{
			ObjectMeta: metav1.ObjectMeta{GenerateName: "vap-ready-path-", Namespace: "default"},
			Spec: AuthEngineMountSpec{
				Authentication: defaultAuth(),
				Path:           "auth/vap-ready-path",
				AuthMount:      AuthMount{Type: "approle", Description: "vap path readiness"},
			},
		}
		if err := k8sClient.Create(ctx, obj); err != nil {
			return fmt.Errorf("create AuthEngineMount for path readiness probe: %w", err)
		}
		defer func() { _ = k8sClient.Delete(ctx, obj) }()
		if err := k8sClient.Get(ctx, client.ObjectKeyFromObject(obj), obj); err != nil {
			return fmt.Errorf("get AuthEngineMount for path readiness probe: %w", err)
		}
		obj.Spec.Path = "auth/vap-ready-path-changed"
		if err := k8sClient.Update(ctx, obj); err == nil {
			return fmt.Errorf("immutable-path VAP not enforcing yet")
		}
		return nil
	}, "60s", "1s").Should(Succeed())
}
