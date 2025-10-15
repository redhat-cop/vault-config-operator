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

package main

import (
	"crypto/tls"
	"flag"
	"os"
	"strconv"
	"time"

	controller "github.com/redhat-cop/vault-config-operator/internal/controller"
	"github.com/redhat-cop/vault-config-operator/internal/controller/vaultresourcecontroller"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"sigs.k8s.io/controller-runtime/pkg/metrics/filters"
	metricsserver "sigs.k8s.io/controller-runtime/pkg/metrics/server"

	//"github.com/redhat-cop/operator-utils/pkg/util"

	"sigs.k8s.io/controller-runtime/pkg/webhook"

	redhatcopv1alpha1 "github.com/redhat-cop/vault-config-operator/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	//+kubebuilder:scaffold:imports
)

var (
	scheme   = runtime.NewScheme()
	setupLog = ctrl.Log.WithName("setup")
)

func init() {
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))

	utilruntime.Must(redhatcopv1alpha1.AddToScheme(scheme))
	//+kubebuilder:scaffold:scheme
}

func main() {

	var (
		metricsAddr          string
		enableLeaderElection bool
		probeAddr            string
		enableHTTP2          bool
		secureMetrics        bool
		tlsOpts              []func(*tls.Config)
	)

	// v1.38 / controller-runtime v0.18 metrics flags
	flag.StringVar(&metricsAddr, "metrics-bind-address", "0",
		"The address the metrics endpoint binds to. Use :8443 for HTTPS or :8080 for HTTP, or leave as 0 to disable the metrics service.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081",
		"The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. Enabling this will ensure there is only one active controller manager.")
	flag.BoolVar(&secureMetrics, "metrics-secure", true,
		"If set, the metrics endpoint is served securely via HTTPS. Use --metrics-secure=false to use HTTP instead.")

	opts := zap.Options{Development: true}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	// Disable HTTP/2 unless explicitly enabled.
	disableHTTP2 := func(c *tls.Config) {
		setupLog.Info("disabling http/2")
		c.NextProtos = []string{"http/1.1"}
	}
	if !enableHTTP2 {
		tlsOpts = append(tlsOpts, disableHTTP2)
	}

	webhookServer := webhook.NewServer(webhook.Options{
		TLSOpts: tlsOpts,
	})

	// Metrics server options (no kube-rbac-proxy in v1.38).
	metricsServerOptions := metricsserver.Options{
		BindAddress:   metricsAddr,
		SecureServing: secureMetrics,
		TLSOpts:       tlsOpts,
	}
	if secureMetrics {
		// Protect /metrics with authn/authz. RBAC rules are in config/rbac/*metrics* files.
		metricsServerOptions.FilterProvider = filters.WithAuthenticationAndAuthorization
	}

	// Optional resync period from env
	var syncPeriod = 36000 * time.Second // Defaults to every 10 hours
	if v, ok := os.LookupEnv("SYNC_PERIOD_SECONDS"); ok && v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			syncPeriod = time.Duration(n) * time.Second
		} else {
			setupLog.Error(err, "invalid SYNC_PERIOD_SECONDS")
			os.Exit(1)
		}
	}

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		Metrics:                metricsServerOptions,
		WebhookServer:          webhookServer,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "3d7d3a62.redhat.io",
		Cache: cache.Options{
			SyncPeriod: &syncPeriod,
		},
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	// Set the sync period for use in predicates
	vaultresourcecontroller.SetSyncPeriod(syncPeriod)

	if err = (&controller.KubernetesAuthEngineRoleReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "KubernetesAuthEngineRole")}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "KubernetesAuthEngineRole")
		os.Exit(1)
	}
	if err = (&controller.PolicyReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "Policy")}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Policy")
		os.Exit(1)
	}
	if err = (&controller.DatabaseSecretEngineConfigReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "DatabaseSecretEngineConfig")}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "DatabaseSecretEngineConfig")
		os.Exit(1)
	}
	if err = (&controller.DatabaseSecretEngineRoleReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "DatabaseSecretEngineRole")}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "DatabaseSecretEngineRole")
		os.Exit(1)
	}
	if err = (&controller.SecretEngineMountReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "SecretEngineMount")}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "SecretEngineMount")
		os.Exit(1)
	}
	if err = (&controller.RandomSecretReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "RandomSecret")}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "RandomSecret")
		os.Exit(1)
	}
	setupLog.Info("starting AuthEngineMountReconciler")
	if err = (&controller.AuthEngineMountReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "AuthEngineMount")}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "AuthEngineMount")
		os.Exit(1)
	}
	setupLog.Info("started AuthEngineMountReconciler")
	if err = (&controller.KubernetesAuthEngineConfigReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "KubernetesAuthEngineConfig")}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "KubernetesAuthEngineConfig")
		os.Exit(1)
	}

	if err = (&controller.LDAPAuthEngineConfigReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "LDAPAuthEngineConfig")}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "LDAPAuthEngineConfig")
		os.Exit(1)
	}

	if err = (&controller.LDAPAuthEngineGroupReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "LDAPAuthEngineGroup")}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "LDAPAuthEngineGroup")
		os.Exit(1)
	}

	if err = (&controller.JWTOIDCAuthEngineConfigReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "JWTOIDCAuthEngineConfig")}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "JWTOIDCAuthEngineConfig")
		os.Exit(1)
	}

	if err = (&controller.JWTOIDCAuthEngineRoleReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "JWTOIDCAuthEngineRole")}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "JWTOIDCAuthEngineRole")
		os.Exit(1)
	}

	if err = (&controller.AzureAuthEngineConfigReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "AzureAuthEngineConfig")}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "AzureAuthEngineConfig")
		os.Exit(1)
	}

	if err = (&controller.AzureAuthEngineRoleReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "AzureAuthEngineRole")}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "AzureAuthEngineRole")
		os.Exit(1)
	}

	if err = (&controller.GCPAuthEngineConfigReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "GCPAuthEngineConfig")}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "GCPAuthEngineConfig")
		os.Exit(1)
	}
	if err = (&controller.GCPAuthEngineRoleReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "GCPAuthEngineRole")}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "GCPAuthEngineRole")
		os.Exit(1)
	}

	if err = (&controller.CertAuthEngineConfigReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "CertAuthEngineConfig")}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "CertAuthEngineConfig")
		os.Exit(1)
	}
	if err = (&controller.CertAuthEngineRoleReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "CertAuthEngineRole")}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "CertAuthEngineRole")
		os.Exit(1)
	}

	if err = (&controller.VaultSecretReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "VaultSecret")}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "VaultSecret")
		os.Exit(1)
	}
	if err = (&controller.PasswordPolicyReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "PasswordPolicy")}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "PasswordPolicy")
		os.Exit(1)
	}
	if err = (&controller.RabbitMQSecretEngineConfigReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "RabbitMQSecretEngineConfig")}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "RabbitMQSecretEngineConfig")
		os.Exit(1)
	}
	if err = (&controller.RabbitMQSecretEngineRoleReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "RabbitMQSecretEngineRole")}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "RabbitMQSecretEngineRole")
		os.Exit(1)
	}

	if err = (&controller.PKISecretEngineConfigReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "PKISecretEngineConfig")}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "PKISecretEngineConfig")
		os.Exit(1)
	}
	if err = (&controller.PKISecretEngineRoleReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "PKISecretEngineRole")}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "PKISecretEngineRole")
		os.Exit(1)
	}

	if err = (&controller.GitHubSecretEngineConfigReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "GitHubSecretEngineConfig")}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "GitHubSecretEngineConfig")
		os.Exit(1)
	}

	if err = (&controller.GitHubSecretEngineRoleReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "GitHubSecretEngineRole")}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "GitHubSecretEngineRole")
		os.Exit(1)
	}

	if err = (&controller.AzureSecretEngineRoleReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "AzureSecretEngineRole")}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "AzureSecretEngineRole")
		os.Exit(1)
	}

	if err = (&controller.AzureSecretEngineConfigReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "AzureSecretEngineConfig")}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "AzureSecretEngineConfig")
		os.Exit(1)
	}

	if err = (&controller.QuaySecretEngineConfigReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "QuaySecretEngineConfig")}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "QuaySecretEngineConfig")
		os.Exit(1)
	}

	if err = (&controller.QuaySecretEngineRoleReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "QuaySecretEngineRole")}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "QuaySecretEngineRole")
		os.Exit(1)
	}

	if err = (&controller.QuaySecretEngineStaticRoleReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "QuaySecretEngineStaticRole")}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "QuaySecretEngineStaticRole")
		os.Exit(1)
	}

	if err = (&controller.KubernetesSecretEngineConfigReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "KubernetesSecretEngineConfig")}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "KubernetesSecretEngineConfig")
		os.Exit(1)
	}

	if err = (&controller.KubernetesSecretEngineRoleReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "KubernetesSecretEngineRole")}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "KubernetesSecretEngineRole")
		os.Exit(1)
	}

	if err = (&controller.DatabaseSecretEngineStaticRoleReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "DatabaseSecretEngineStaticRole")}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "DatabaseSecretEngineStaticRole")
		os.Exit(1)
	}

	if err = (&controller.GroupReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "Group")}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Group")
		os.Exit(1)
	}
	if err = (&controller.GroupAliasReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "GroupAlias")}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "GroupAlias")
		os.Exit(1)
	}

	if webhooks, ok := os.LookupEnv("ENABLE_WEBHOOKS"); !ok || webhooks != "false" {
		if err = (&redhatcopv1alpha1.RandomSecret{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "RandomSecret")
			os.Exit(1)
		}
		if err = (&redhatcopv1alpha1.SecretEngineMount{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "SecretEngineMount")
			os.Exit(1)
		}
		if err = (&redhatcopv1alpha1.DatabaseSecretEngineRole{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "DatabaseSecretEngineRole")
			os.Exit(1)
		}
		if err = (&redhatcopv1alpha1.DatabaseSecretEngineConfig{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "DatabaseSecretEngineConfig")
			os.Exit(1)
		}
		if err = (&redhatcopv1alpha1.KubernetesAuthEngineRole{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "KubernetesAuthEngineRole")
			os.Exit(1)
		}
		if err = (&redhatcopv1alpha1.AuthEngineMount{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "AuthEngineMount")
			os.Exit(1)
		}
		if err = (&redhatcopv1alpha1.KubernetesAuthEngineConfig{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "KubernetesAuthEngineConfig")
			os.Exit(1)
		}
		if err = (&redhatcopv1alpha1.LDAPAuthEngineConfig{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "LDAPAuthEngineConfig")
			os.Exit(1)
		}
		if err = (&redhatcopv1alpha1.LDAPAuthEngineGroup{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "LDAPAuthEngineGroup")
			os.Exit(1)
		}
		if err = (&redhatcopv1alpha1.JWTOIDCAuthEngineConfig{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "JWTOIDCAuthEngineConfig")
			os.Exit(1)
		}
		if err = (&redhatcopv1alpha1.JWTOIDCAuthEngineRole{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "JWTOIDCAuthEngineRole")
			os.Exit(1)
		}
		if err = (&redhatcopv1alpha1.AzureAuthEngineConfig{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "AzureAuthEngineConfig")
			os.Exit(1)
		}
		if err = (&redhatcopv1alpha1.AzureAuthEngineRole{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "AzureAuthEngineRole")
			os.Exit(1)
		}

		if err = (&redhatcopv1alpha1.GCPAuthEngineConfig{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "GCPAuthEngineConfig")
			os.Exit(1)
		}
		if err = (&redhatcopv1alpha1.GCPAuthEngineRole{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "GCPAuthEngineRole")
			os.Exit(1)
		}

		if err = (&redhatcopv1alpha1.CertAuthEngineConfig{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "CertAuthEngineConfig")
			os.Exit(1)
		}
		if err = (&redhatcopv1alpha1.CertAuthEngineRole{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "CertAuthEngineRole")
			os.Exit(1)
		}

		if err = (&redhatcopv1alpha1.VaultSecret{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "VaultSecret")
			os.Exit(1)
		}
		if err = (&redhatcopv1alpha1.PasswordPolicy{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "PasswordPolicy")
			os.Exit(1)
		}
		if err = (&redhatcopv1alpha1.Policy{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "Policy")
			os.Exit(1)
		}
		if err = (&redhatcopv1alpha1.GitHubSecretEngineConfig{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "GitHubSecretEngineConfig")
			os.Exit(1)
		}

		if err = (&redhatcopv1alpha1.GitHubSecretEngineRole{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "GitHubSecretEngineRole")
			os.Exit(1)
		}

		if err = (&redhatcopv1alpha1.AzureSecretEngineRole{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "AzureSecretEngineRole")
			os.Exit(1)
		}

		if err = (&redhatcopv1alpha1.AzureSecretEngineConfig{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "AzureSecretEngineConfig")
			os.Exit(1)
		}

		if err = (&redhatcopv1alpha1.PKISecretEngineConfig{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "PKISecretEngineConfig")
			os.Exit(1)
		}

		if err = (&redhatcopv1alpha1.PKISecretEngineRole{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "PKISecretEngineRole")
			os.Exit(1)
		}

		if err = (&redhatcopv1alpha1.RabbitMQSecretEngineRole{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "RabbitMQSecretEngineRole")
			os.Exit(1)
		}

		if err = (&redhatcopv1alpha1.QuaySecretEngineConfig{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "QuaySecretEngineConfig")
			os.Exit(1)
		}

		if err = (&redhatcopv1alpha1.QuaySecretEngineRole{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "QuaySecretEngineRole")
			os.Exit(1)
		}

		if err = (&redhatcopv1alpha1.QuaySecretEngineStaticRole{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "QuaySecretEngineStaticRole")
			os.Exit(1)
		}

		if err = (&redhatcopv1alpha1.KubernetesSecretEngineConfig{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "KubernetesSecretEngineConfig")
			os.Exit(1)
		}

		if err = (&redhatcopv1alpha1.KubernetesSecretEngineRole{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "KubernetesSecretEngineRole")
			os.Exit(1)
		}

		mgr.GetWebhookServer().Register("/validate-redhatcop-redhat-io-v1alpha1-rabbitmqsecretengineconfig", &webhook.Admission{Handler: &redhatcopv1alpha1.RabbitMQSecretEngineConfigValidation{Client: mgr.GetClient()}})

		if err = (&redhatcopv1alpha1.DatabaseSecretEngineStaticRole{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "DatabaseSecretEngineStaticRole")
			os.Exit(1)
		}

		if err = (&redhatcopv1alpha1.Group{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "Group")
			os.Exit(1)
		}
		if err = (&redhatcopv1alpha1.GroupAlias{}).SetupWebhookWithManager(mgr); err != nil {
			setupLog.Error(err, "unable to create webhook", "webhook", "GroupAlias")
			os.Exit(1)
		}
	}

	//+kubebuilder:scaffold:builder

	if err := mgr.AddHealthzCheck("healthz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up health check")
		os.Exit(1)
	}
	if err := mgr.AddReadyzCheck("readyz", healthz.Ping); err != nil {
		setupLog.Error(err, "unable to set up ready check")
		os.Exit(1)
	}

	setupLog.Info("starting manager")
	if err := mgr.Start(ctrl.SetupSignalHandler()); err != nil {
		setupLog.Error(err, "problem running manager")
		os.Exit(1)
	}
}
