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
	"github.com/redhat-cop/vault-config-operator/internal/controller"
	"github.com/redhat-cop/vault-config-operator/internal/controller/vaultresourcecontroller"
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
	var metricsAddr string
	var enableLeaderElection bool
	var probeAddr string
	var enableHTTP2 bool
	var secureMetrics bool
	var tlsOpts []func(*tls.Config)
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8443", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	flag.BoolVar(&secureMetrics, "metrics-secure", true,
		"If set, the metrics endpoint is served securely via HTTPS. Use --metrics-secure=false to use HTTP instead.")
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))
	// if the enable-http2 flag is false (the default), http/2 should be disabled
	// due to its vulnerabilities. More specifically, disabling http/2 will
	// prevent from being vulnerable to the HTTP/2 Stream Cancellation and
	// Rapid Reset CVEs. For more information see:
	// - https://github.com/advisories/GHSA-qppj-fm5r-hxr3
	// - https://github.com/advisories/GHSA-4374-p667-p6c8
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

	// Metrics endpoint is enabled in 'config/default/kustomization.yaml'. The Metrics options configure the server.
	// More info:
	// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.0/pkg/metrics/server
	// - https://book.kubebuilder.io/reference/metrics.html
	metricsServerOptions := metricsserver.Options{
		BindAddress:   metricsAddr,
		SecureServing: secureMetrics,
		TLSOpts:       tlsOpts,
	}

	if secureMetrics {
		metricsServerOptions.FilterProvider = filters.WithAuthenticationAndAuthorization
	}
	var syncPeriod = 36000 * time.Second //Defaults to every 10Hrs
	if syncPeriodSeconds, ok := os.LookupEnv("SYNC_PERIOD_SECONDS"); ok && syncPeriodSeconds != "" {
		if syncPeriodSecondsInt, err := strconv.ParseInt(syncPeriodSeconds, 10, 64); err == nil {
			syncPeriod = time.Duration(syncPeriodSecondsInt) * time.Second
		} else if err != nil {
			setupLog.Error(err, "unable to start manager")
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

	// --- Generic-migrated controllers (Categories 1-4) ---

	for _, setup := range []struct {
		name string
		fn   func(ctrl.Manager) error
	}{
		{"AliCloudAuthEngineRole", controller.SetupAliCloudAuthEngineRoleController},
		{"AppRoleAuthEngineRole", controller.SetupAppRoleAuthEngineRoleController},
		{"Audit", controller.SetupAuditController},
		{"AuditRequestHeader", controller.SetupAuditRequestHeaderController},
		{"AuthEngineMount", controller.SetupAuthEngineMountController},
		{"AWSAuthEngineClientConfig", controller.SetupAWSAuthEngineClientConfigController},
		{"AWSAuthEngineIdentityConfig", controller.SetupAWSAuthEngineIdentityConfigController},
		{"AWSAuthEngineRole", controller.SetupAWSAuthEngineRoleController},
		{"AWSSecretEngineRole", controller.SetupAWSSecretEngineRoleController},
		{"AzureAuthEngineConfig", controller.SetupAzureAuthEngineConfigController},
		{"AzureAuthEngineRole", controller.SetupAzureAuthEngineRoleController},
		{"AzureSecretEngineConfig", controller.SetupAzureSecretEngineConfigController},
		{"AzureSecretEngineRole", controller.SetupAzureSecretEngineRoleController},
		{"CertAuthEngineConfig", controller.SetupCertAuthEngineConfigController},
		{"CertAuthEngineRole", controller.SetupCertAuthEngineRoleController},
		{"CFAuthEngineConfig", controller.SetupCFAuthEngineConfigController},
		{"CFAuthEngineRole", controller.SetupCFAuthEngineRoleController},
		{"ConsulSecretEngineRole", controller.SetupConsulSecretEngineRoleController},
		{"CORSConfig", controller.SetupCORSConfigController},
		{"DatabaseSecretEngineRole", controller.SetupDatabaseSecretEngineRoleController},
		{"DatabaseSecretEngineStaticRole", controller.SetupDatabaseSecretEngineStaticRoleController},
		{"Entity", controller.SetupEntityController},
		{"EntityAlias", controller.SetupEntityAliasController},
		{"GCPAuthEngineConfig", controller.SetupGCPAuthEngineConfigController},
		{"GCPAuthEngineRole", controller.SetupGCPAuthEngineRoleController},
		{"GCPSecretEngineRoleset", controller.SetupGCPSecretEngineRolesetController},
		{"GCPSecretEngineStaticAccount", controller.SetupGCPSecretEngineStaticAccountController},
		{"GitHubAuthEngineConfig", controller.SetupGitHubAuthEngineConfigController},
		{"GitHubAuthEngineTeamMap", controller.SetupGitHubAuthEngineTeamMapController},
		{"GitHubAuthEngineUserMap", controller.SetupGitHubAuthEngineUserMapController},
		{"GitHubSecretEngineConfig", controller.SetupGitHubSecretEngineConfigController},
		{"GitHubSecretEngineRole", controller.SetupGitHubSecretEngineRoleController},
		{"Group", controller.SetupGroupController},
		{"GroupAlias", controller.SetupGroupAliasController},
		{"IdentityOIDCAssignment", controller.SetupIdentityOIDCAssignmentController},
		{"IdentityOIDCClient", controller.SetupIdentityOIDCClientController},
		{"IdentityOIDCProvider", controller.SetupIdentityOIDCProviderController},
		{"IdentityOIDCScope", controller.SetupIdentityOIDCScopeController},
		{"IdentityTokenConfig", controller.SetupIdentityTokenConfigController},
		{"IdentityTokenKey", controller.SetupIdentityTokenKeyController},
		{"IdentityTokenRole", controller.SetupIdentityTokenRoleController},
		{"JWTOIDCAuthEngineConfig", controller.SetupJWTOIDCAuthEngineConfigController},
		{"JWTOIDCAuthEngineRole", controller.SetupJWTOIDCAuthEngineRoleController},
		{"KerberosAuthEngineConfig", controller.SetupKerberosAuthEngineConfigController},
		{"KerberosAuthEngineLDAPConfig", controller.SetupKerberosAuthEngineLDAPConfigController},
		{"KerberosAuthEngineGroup", controller.SetupKerberosAuthEngineGroupController},
		{"KeyRotationConfig", controller.SetupKeyRotationConfigController},
		{"KubernetesAuthEngineConfig", controller.SetupKubernetesAuthEngineConfigController},
		{"KubernetesAuthEngineRole", controller.SetupKubernetesAuthEngineRoleController},
		{"KubernetesSecretEngineConfig", controller.SetupKubernetesSecretEngineConfigController},
		{"KubernetesSecretEngineRole", controller.SetupKubernetesSecretEngineRoleController},
		{"LDAPAuthEngineConfig", controller.SetupLDAPAuthEngineConfigController},
		{"LDAPAuthEngineGroup", controller.SetupLDAPAuthEngineGroupController},
		{"LDAPSecretEngineDynamicRole", controller.SetupLDAPSecretEngineDynamicRoleController},
		{"LDAPSecretEngineStaticRole", controller.SetupLDAPSecretEngineStaticRoleController},
		{"LeaseCountQuota", controller.SetupLeaseCountQuotaController},
		{"LoggerConfig", controller.SetupLoggerConfigController},
		{"MongoDBAtlasSecretEngineRole", controller.SetupMongoDBAtlasSecretEngineRoleController},
		{"Namespace", controller.SetupNamespaceController},
		{"NomadSecretEngineRole", controller.SetupNomadSecretEngineRoleController},
		{"OCIAuthEngineConfig", controller.SetupOCIAuthEngineConfigController},
		{"OCIAuthEngineRole", controller.SetupOCIAuthEngineRoleController},
		{"OktaAuthEngineConfig", controller.SetupOktaAuthEngineConfigController},
		{"OktaAuthEngineGroup", controller.SetupOktaAuthEngineGroupController},
		{"PasswordPolicy", controller.SetupPasswordPolicyController},
		{"PKISecretEngineConfig", controller.SetupPKISecretEngineConfigController},
		{"PKISecretEngineRole", controller.SetupPKISecretEngineRoleController},
		{"Policy", controller.SetupPolicyController},
		{"QuaySecretEngineConfig", controller.SetupQuaySecretEngineConfigController},
		{"QuaySecretEngineRole", controller.SetupQuaySecretEngineRoleController},
		{"QuaySecretEngineStaticRole", controller.SetupQuaySecretEngineStaticRoleController},
		{"QuotaConfig", controller.SetupQuotaConfigController},
		{"RabbitMQSecretEngineRole", controller.SetupRabbitMQSecretEngineRoleController},
		{"RADIUSAuthEngineConfig", controller.SetupRADIUSAuthEngineConfigController},
		{"RADIUSAuthEngineUser", controller.SetupRADIUSAuthEngineUserController},
		{"RateLimitQuota", controller.SetupRateLimitQuotaController},
		{"RotationPolicy", controller.SetupRotationPolicyController},
		{"SecretEngineMount", controller.SetupSecretEngineMountController},
		{"SSHSecretEngineConfig", controller.SetupSSHSecretEngineConfigController},
		{"SSHSecretEngineRole", controller.SetupSSHSecretEngineRoleController},
		{"TerraformCloudSecretEngineRole", controller.SetupTerraformCloudSecretEngineRoleController},
		{"UICustomHeader", controller.SetupUICustomHeaderController},
		{"UserpassAuthEngineUser", controller.SetupUserpassAuthEngineUserController},
	} {
		if err = setup.fn(mgr); err != nil {
			setupLog.Error(err, "unable to create controller", "controller", setup.name)
			os.Exit(1)
		}
	}

	// --- Custom controllers (Category 5 — hand-written Reconcile) ---

	if err = (&controller.DatabaseSecretEngineConfigReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "DatabaseSecretEngineConfig")}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "DatabaseSecretEngineConfig")
		os.Exit(1)
	}
	if err = (&controller.RandomSecretReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "RandomSecret")}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "RandomSecret")
		os.Exit(1)
	}
	if err = (&controller.VaultSecretReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "VaultSecret")}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "VaultSecret")
		os.Exit(1)
	}
	if err = (&controller.RabbitMQSecretEngineConfigReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "RabbitMQSecretEngineConfig")}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "RabbitMQSecretEngineConfig")
		os.Exit(1)
	}
	if err = (&controller.TransitSecretEngineKeyReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "TransitSecretEngineKey")}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "TransitSecretEngineKey")
		os.Exit(1)
	}
	if err = (&controller.TOTPSecretEngineKeyReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "TOTPSecretEngineKey")}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "TOTPSecretEngineKey")
		os.Exit(1)
	}
	if err = (&controller.AWSSecretEngineConfigReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "AWSSecretEngineConfig")}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "AWSSecretEngineConfig")
		os.Exit(1)
	}
	if err = (&controller.ConsulSecretEngineConfigReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "ConsulSecretEngineConfig")}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "ConsulSecretEngineConfig")
		os.Exit(1)
	}
	if err = (&controller.GCPSecretEngineConfigReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "GCPSecretEngineConfig")}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "GCPSecretEngineConfig")
		os.Exit(1)
	}
	if err = (&controller.LDAPSecretEngineConfigReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "LDAPSecretEngineConfig")}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "LDAPSecretEngineConfig")
		os.Exit(1)
	}
	if err = (&controller.MongoDBAtlasSecretEngineConfigReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "MongoDBAtlasSecretEngineConfig")}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "MongoDBAtlasSecretEngineConfig")
		os.Exit(1)
	}
	if err = (&controller.NomadSecretEngineConfigReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "NomadSecretEngineConfig")}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "NomadSecretEngineConfig")
		os.Exit(1)
	}
	if err = (&controller.TerraformCloudSecretEngineConfigReconciler{ReconcilerBase: vaultresourcecontroller.NewFromManager(mgr, "TerraformCloudSecretEngineConfig")}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "TerraformCloudSecretEngineConfig")
		os.Exit(1)
	}

	if webhooks, ok := os.LookupEnv("ENABLE_WEBHOOKS"); !ok || webhooks != "false" {
		for _, setup := range []struct {
			name string
			fn   func(ctrl.Manager) error
		}{
			// Category A — defaulter + validator (uncovered validation)
			{"AWSAuthEngineClientConfig", func(m ctrl.Manager) error {
				return (&redhatcopv1alpha1.AWSAuthEngineClientConfig{}).SetupWebhookWithManager(m)
			}},
			// Category B — defaulter-only (validator removed, VAP covers validation)
			{"CFAuthEngineConfig", func(m ctrl.Manager) error {
				return (&redhatcopv1alpha1.CFAuthEngineConfig{}).SetupWebhookWithManager(m)
			}},
			{"GCPSecretEngineConfig", func(m ctrl.Manager) error {
				return (&redhatcopv1alpha1.GCPSecretEngineConfig{}).SetupWebhookWithManager(m)
			}},
			{"KubernetesAuthEngineConfig", func(m ctrl.Manager) error {
				return (&redhatcopv1alpha1.KubernetesAuthEngineConfig{}).SetupWebhookWithManager(m)
			}},
			{"OktaAuthEngineConfig", func(m ctrl.Manager) error {
				return (&redhatcopv1alpha1.OktaAuthEngineConfig{}).SetupWebhookWithManager(m)
			}},
			{"RADIUSAuthEngineConfig", func(m ctrl.Manager) error {
				return (&redhatcopv1alpha1.RADIUSAuthEngineConfig{}).SetupWebhookWithManager(m)
			}},
			{"SSHSecretEngineConfig", func(m ctrl.Manager) error {
				return (&redhatcopv1alpha1.SSHSecretEngineConfig{}).SetupWebhookWithManager(m)
			}},
			{"TerraformCloudSecretEngineConfig", func(m ctrl.Manager) error {
				return (&redhatcopv1alpha1.TerraformCloudSecretEngineConfig{}).SetupWebhookWithManager(m)
			}},
			// Category C — validator-only (uncovered CRD-specific validation)
			{"AliCloudAuthEngineRole", func(m ctrl.Manager) error {
				return (&redhatcopv1alpha1.AliCloudAuthEngineRole{}).SetupWebhookWithManager(m)
			}},
			{"AppRoleAuthEngineRole", func(m ctrl.Manager) error {
				return (&redhatcopv1alpha1.AppRoleAuthEngineRole{}).SetupWebhookWithManager(m)
			}},
			{"AWSAuthEngineRole", func(m ctrl.Manager) error { return (&redhatcopv1alpha1.AWSAuthEngineRole{}).SetupWebhookWithManager(m) }},
			{"CORSConfig", func(m ctrl.Manager) error { return (&redhatcopv1alpha1.CORSConfig{}).SetupWebhookWithManager(m) }},
			{"GCPSecretEngineRoleset", func(m ctrl.Manager) error {
				return (&redhatcopv1alpha1.GCPSecretEngineRoleset{}).SetupWebhookWithManager(m)
			}},
			{"GCPSecretEngineStaticAccount", func(m ctrl.Manager) error {
				return (&redhatcopv1alpha1.GCPSecretEngineStaticAccount{}).SetupWebhookWithManager(m)
			}},
			{"IdentityOIDCClient", func(m ctrl.Manager) error {
				return (&redhatcopv1alpha1.IdentityOIDCClient{}).SetupWebhookWithManager(m)
			}},
			{"JWTOIDCAuthEngineRole", func(m ctrl.Manager) error {
				return (&redhatcopv1alpha1.JWTOIDCAuthEngineRole{}).SetupWebhookWithManager(m)
			}},
			{"KeyRotationConfig", func(m ctrl.Manager) error { return (&redhatcopv1alpha1.KeyRotationConfig{}).SetupWebhookWithManager(m) }},
			{"LDAPSecretEngineStaticRole", func(m ctrl.Manager) error {
				return (&redhatcopv1alpha1.LDAPSecretEngineStaticRole{}).SetupWebhookWithManager(m)
			}},
			{"LeaseCountQuota", func(m ctrl.Manager) error { return (&redhatcopv1alpha1.LeaseCountQuota{}).SetupWebhookWithManager(m) }},
			{"PKISecretEngineConfig", func(m ctrl.Manager) error {
				return (&redhatcopv1alpha1.PKISecretEngineConfig{}).SetupWebhookWithManager(m)
			}},
			{"RandomSecret", func(m ctrl.Manager) error { return (&redhatcopv1alpha1.RandomSecret{}).SetupWebhookWithManager(m) }},
			{"RateLimitQuota", func(m ctrl.Manager) error { return (&redhatcopv1alpha1.RateLimitQuota{}).SetupWebhookWithManager(m) }},
			{"UICustomHeader", func(m ctrl.Manager) error { return (&redhatcopv1alpha1.UICustomHeader{}).SetupWebhookWithManager(m) }},
		} {
			if err = setup.fn(mgr); err != nil {
				setupLog.Error(err, "unable to create webhook", "webhook", setup.name)
				os.Exit(1)
			}
		}

		// Category E — RabbitMQ raw handler (cross-resource validateUniquePath)
		mgr.GetWebhookServer().Register("/validate-redhatcop-redhat-io-v1alpha1-rabbitmqsecretengineconfig", &webhook.Admission{Handler: &redhatcopv1alpha1.RabbitMQSecretEngineConfigValidation{Client: mgr.GetClient()}})
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
