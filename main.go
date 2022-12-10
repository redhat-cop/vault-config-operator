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
	"flag"
	"os"

	// Import all Kubernetes client auth plugins (e.g. Azure, GCP, OIDC, etc.)
	// to ensure that exec-entrypoint and run can make use of them.
	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/healthz"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	"github.com/redhat-cop/operator-utils/pkg/util"

	"sigs.k8s.io/controller-runtime/pkg/webhook"

	redhatcopv1alpha1 "github.com/redhat-cop/vault-config-operator/api/v1alpha1"
	"github.com/redhat-cop/vault-config-operator/controllers"
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
	flag.StringVar(&metricsAddr, "metrics-bind-address", ":8080", "The address the metric endpoint binds to.")
	flag.StringVar(&probeAddr, "health-probe-bind-address", ":8081", "The address the probe endpoint binds to.")
	flag.BoolVar(&enableLeaderElection, "leader-elect", false,
		"Enable leader election for controller manager. "+
			"Enabling this will ensure there is only one active controller manager.")
	opts := zap.Options{
		Development: true,
	}
	opts.BindFlags(flag.CommandLine)
	flag.Parse()

	ctrl.SetLogger(zap.New(zap.UseFlagOptions(&opts)))

	mgr, err := ctrl.NewManager(ctrl.GetConfigOrDie(), ctrl.Options{
		Scheme:                 scheme,
		MetricsBindAddress:     metricsAddr,
		Port:                   9443,
		HealthProbeBindAddress: probeAddr,
		LeaderElection:         enableLeaderElection,
		LeaderElectionID:       "3d7d3a62.redhat.io",
	})
	if err != nil {
		setupLog.Error(err, "unable to start manager")
		os.Exit(1)
	}

	if err = (&controllers.KubernetesAuthEngineRoleReconciler{
		ReconcilerBase: util.NewReconcilerBase(mgr.GetClient(), mgr.GetScheme(), mgr.GetConfig(), mgr.GetEventRecorderFor("KubernetesAuthEngineRole"), mgr.GetAPIReader()),
		Log:            ctrl.Log.WithName("controllers").WithName("KubernetesAuthEngineRole"),
		ControllerName: "KubernetesAuthEngineRole",
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "KubernetesAuthEngineRole")
		os.Exit(1)
	}
	if err = (&controllers.PolicyReconciler{
		ReconcilerBase: util.NewReconcilerBase(mgr.GetClient(), mgr.GetScheme(), mgr.GetConfig(), mgr.GetEventRecorderFor("Policy"), mgr.GetAPIReader()),
		Log:            ctrl.Log.WithName("controllers").WithName("Policy"),
		ControllerName: "Policy",
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "Policy")
		os.Exit(1)
	}
	if err = (&controllers.DatabaseSecretEngineConfigReconciler{
		ReconcilerBase: util.NewReconcilerBase(mgr.GetClient(), mgr.GetScheme(), mgr.GetConfig(), mgr.GetEventRecorderFor("DatabaseSecretEngineConfig"), mgr.GetAPIReader()),
		Log:            ctrl.Log.WithName("controllers").WithName("DatabaseSecretEngineConfig"),
		ControllerName: "DatabaseSecretEngineConfig",
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "DatabaseSecretEngineConfig")
		os.Exit(1)
	}
	if err = (&controllers.DatabaseSecretEngineRoleReconciler{
		ReconcilerBase: util.NewReconcilerBase(mgr.GetClient(), mgr.GetScheme(), mgr.GetConfig(), mgr.GetEventRecorderFor("DatabaseSecretEngineRole"), mgr.GetAPIReader()),
		Log:            ctrl.Log.WithName("controllers").WithName("DatabaseSecretEngineRole"),
		ControllerName: "DatabaseSecretEngineRole",
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "DatabaseSecretEngineRole")
		os.Exit(1)
	}
	if err = (&controllers.SecretEngineMountReconciler{
		ReconcilerBase: util.NewReconcilerBase(mgr.GetClient(), mgr.GetScheme(), mgr.GetConfig(), mgr.GetEventRecorderFor("SecretEngineMount"), mgr.GetAPIReader()),
		Log:            ctrl.Log.WithName("controllers").WithName("SecretEngineMount"),
		ControllerName: "SecretEngineMount",
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "SecretEngineMount")
		os.Exit(1)
	}
	if err = (&controllers.RandomSecretReconciler{
		ReconcilerBase: util.NewReconcilerBase(mgr.GetClient(), mgr.GetScheme(), mgr.GetConfig(), mgr.GetEventRecorderFor("RandomSecret"), mgr.GetAPIReader()),
		Log:            ctrl.Log.WithName("controllers").WithName("RandomSecret"),
		ControllerName: "RandomSecret",
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "RandomSecret")
		os.Exit(1)
	}
	setupLog.Info("starting AuthEngineMountReconciler")
	if err = (&controllers.AuthEngineMountReconciler{
		ReconcilerBase: util.NewReconcilerBase(mgr.GetClient(), mgr.GetScheme(), mgr.GetConfig(), mgr.GetEventRecorderFor("AuthEngineMount"), mgr.GetAPIReader()),
		Log:            ctrl.Log.WithName("controllers").WithName("AuthEngineMount"),
		ControllerName: "AuthEngineMount",
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "AuthEngineMount")
		os.Exit(1)
	}
	setupLog.Info("started AuthEngineMountReconciler")
	if err = (&controllers.KubernetesAuthEngineConfigReconciler{
		ReconcilerBase: util.NewReconcilerBase(mgr.GetClient(), mgr.GetScheme(), mgr.GetConfig(), mgr.GetEventRecorderFor("KubernetesAuthEngineConfig"), mgr.GetAPIReader()),
		Log:            ctrl.Log.WithName("controllers").WithName("KubernetesAuthEngineConfig"),
		ControllerName: "KubernetesAuthEngineConfig",
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "KubernetesAuthEngineConfig")
		os.Exit(1)
	}

	if err = (&controllers.LDAPAuthEngineConfigReconciler{
		ReconcilerBase: util.NewReconcilerBase(mgr.GetClient(), mgr.GetScheme(), mgr.GetConfig(), mgr.GetEventRecorderFor("LDAPAuthEngineConfig"), mgr.GetAPIReader()),
		Log:            ctrl.Log.WithName("controllers").WithName("LDAPAuthEngineConfig"),
		ControllerName: "LDAPAuthEngineConfig",
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "LDAPAuthEngineConfig")
		os.Exit(1)
	}

	if err = (&controllers.LDAPAuthEngineGroupReconciler{
		ReconcilerBase: util.NewReconcilerBase(mgr.GetClient(), mgr.GetScheme(), mgr.GetConfig(), mgr.GetEventRecorderFor("LDAPAuthEngineGroup"), mgr.GetAPIReader()),
		Log:            ctrl.Log.WithName("controllers").WithName("LDAPAuthEngineGroup"),
		ControllerName: "LDAPAuthEngineGroup",
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "LDAPAuthEngineGroup")
		os.Exit(1)
	}

	if err = (&controllers.JWTOIDCAuthEngineConfigReconciler{
		ReconcilerBase: util.NewReconcilerBase(mgr.GetClient(), mgr.GetScheme(), mgr.GetConfig(), mgr.GetEventRecorderFor("JWTOIDCAuthEngineConfig"), mgr.GetAPIReader()),
		Log:            ctrl.Log.WithName("controllers").WithName("JWTOIDCAuthEngineConfig"),
		ControllerName: "JWTOIDCAuthEngineConfig",
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "JWTOIDCAuthEngineConfig")
		os.Exit(1)
	}

	if err = (&controllers.JWTOIDCAuthEngineRoleReconciler{
		ReconcilerBase: util.NewReconcilerBase(mgr.GetClient(), mgr.GetScheme(), mgr.GetConfig(), mgr.GetEventRecorderFor("JWTOIDCAuthEngineRole"), mgr.GetAPIReader()),
		Log:            ctrl.Log.WithName("controllers").WithName("JWTOIDCAuthEngineRole"),
		ControllerName: "JWTOIDCAuthEngineRole",
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "JWTOIDCAuthEngineRole")
		os.Exit(1)
	}

	if err = (&controllers.VaultSecretReconciler{
		ReconcilerBase: util.NewReconcilerBase(mgr.GetClient(), mgr.GetScheme(), mgr.GetConfig(), mgr.GetEventRecorderFor("VaultSecret"), mgr.GetAPIReader()),
		Log:            ctrl.Log.WithName("controllers").WithName("VaultSecret"),
		ControllerName: "VaultSecret",
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "VaultSecret")
		os.Exit(1)
	}
	if err = (&controllers.PasswordPolicyReconciler{
		ReconcilerBase: util.NewReconcilerBase(mgr.GetClient(), mgr.GetScheme(), mgr.GetConfig(), mgr.GetEventRecorderFor("PasswordPolicy"), mgr.GetAPIReader()),
		Log:            ctrl.Log.WithName("controllers").WithName("PasswordPolicy"),
		ControllerName: "PasswordPolicy",
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "PasswordPolicy")
		os.Exit(1)
	}
	if err = (&controllers.RabbitMQSecretEngineConfigReconciler{
		ReconcilerBase: util.NewReconcilerBase(mgr.GetClient(), mgr.GetScheme(), mgr.GetConfig(), mgr.GetEventRecorderFor("RabbitMQSecretEngineConfig"), mgr.GetAPIReader()),
		Log:            ctrl.Log.WithName("controllers").WithName("RabbitMQSecretEngineConfig"),
		ControllerName: "RabbitMQSecretEngineConfig",
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "RabbitMQSecretEngineConfig")
		os.Exit(1)
	}
	if err = (&controllers.RabbitMQSecretEngineRoleReconciler{
		ReconcilerBase: util.NewReconcilerBase(mgr.GetClient(), mgr.GetScheme(), mgr.GetConfig(), mgr.GetEventRecorderFor("RabbitMQSecretEngineRole"), mgr.GetAPIReader()),
		Log:            ctrl.Log.WithName("controllers").WithName("RabbitMQSecretEngineRole"),
		ControllerName: "RabbitMQSecretEngineRole",
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "RabbitMQSecretEngineRole")
		os.Exit(1)
	}

	if err = (&controllers.PKISecretEngineConfigReconciler{
		ReconcilerBase: util.NewReconcilerBase(mgr.GetClient(), mgr.GetScheme(), mgr.GetConfig(), mgr.GetEventRecorderFor("PKISecretEngineConfig"), mgr.GetAPIReader()),
		Log:            ctrl.Log.WithName("controllers").WithName("PKISecretEngineConfig"),
		ControllerName: "PKISecretEngineConfig",
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "PKISecretEngineConfig")
		os.Exit(1)
	}
	if err = (&controllers.PKISecretEngineRoleReconciler{
		ReconcilerBase: util.NewReconcilerBase(mgr.GetClient(), mgr.GetScheme(), mgr.GetConfig(), mgr.GetEventRecorderFor("PKISecretEngineRole"), mgr.GetAPIReader()),
		Log:            ctrl.Log.WithName("controllers").WithName("PKISecretEngineRole"),
		ControllerName: "PKISecretEngineRole",
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "PKISecretEngineRole")
		os.Exit(1)
	}

	if err = (&controllers.GitHubSecretEngineConfigReconciler{
		ReconcilerBase: util.NewReconcilerBase(mgr.GetClient(), mgr.GetScheme(), mgr.GetConfig(), mgr.GetEventRecorderFor("GitHubSecretEngineConfig"), mgr.GetAPIReader()),
		Log:            ctrl.Log.WithName("controllers").WithName("GitHubSecretEngineConfig"),
		ControllerName: "GitHubSecretEngineConfig",
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "GitHubSecretEngineConfig")
		os.Exit(1)
	}

	if err = (&controllers.GitHubSecretEngineRoleReconciler{
		ReconcilerBase: util.NewReconcilerBase(mgr.GetClient(), mgr.GetScheme(), mgr.GetConfig(), mgr.GetEventRecorderFor("GitHubSecretEngineRole"), mgr.GetAPIReader()),
		Log:            ctrl.Log.WithName("controllers").WithName("GitHubSecretEngineRole"),
		ControllerName: "GitHubSecretEngineRole",
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "GitHubSecretEngineRole")
		os.Exit(1)
	}

	if err = (&controllers.QuaySecretEngineConfigReconciler{
		ReconcilerBase: util.NewReconcilerBase(mgr.GetClient(), mgr.GetScheme(), mgr.GetConfig(), mgr.GetEventRecorderFor("QuaySecretEngineConfig"), mgr.GetAPIReader()),
		Log:            ctrl.Log.WithName("controllers").WithName("QuaySecretEngineConfig"),
		ControllerName: "QuaySecretEngineConfig",
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "QuaySecretEngineConfig")
		os.Exit(1)
	}

	if err = (&controllers.QuaySecretEngineRoleReconciler{
		ReconcilerBase: util.NewReconcilerBase(mgr.GetClient(), mgr.GetScheme(), mgr.GetConfig(), mgr.GetEventRecorderFor("QuaySecretEngineRole"), mgr.GetAPIReader()),
		Log:            ctrl.Log.WithName("controllers").WithName("QuaySecretEngineRole"),
		ControllerName: "QuaySecretEngineRole",
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "QuaySecretEngineRole")
		os.Exit(1)
	}

	if err = (&controllers.QuaySecretEngineStaticRoleReconciler{
		ReconcilerBase: util.NewReconcilerBase(mgr.GetClient(), mgr.GetScheme(), mgr.GetConfig(), mgr.GetEventRecorderFor("QuaySecretEngineStaticRole"), mgr.GetAPIReader()),
		Log:            ctrl.Log.WithName("controllers").WithName("QuaySecretEngineStaticRole"),
		ControllerName: "QuaySecretEngineStaticRole",
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "QuaySecretEngineStaticRole")
		os.Exit(1)
	}

	if err = (&controllers.KubernetesSecretEngineConfigReconciler{
		ReconcilerBase: util.NewReconcilerBase(mgr.GetClient(), mgr.GetScheme(), mgr.GetConfig(), mgr.GetEventRecorderFor("KubernetesSecretEngineConfig"), mgr.GetAPIReader()),
		Log:            ctrl.Log.WithName("controllers").WithName("KubernetesSecretEngineConfig"),
		ControllerName: "KubernetesSecretEngineConfig",
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "KubernetesSecretEngineConfig")
		os.Exit(1)
	}

	if err = (&controllers.KubernetesSecretEngineRoleReconciler{
		ReconcilerBase: util.NewReconcilerBase(mgr.GetClient(), mgr.GetScheme(), mgr.GetConfig(), mgr.GetEventRecorderFor("KubernetesSecretEngineRole"), mgr.GetAPIReader()),
		Log:            ctrl.Log.WithName("controllers").WithName("KubernetesSecretEngineRole"),
		ControllerName: "KubernetesSecretEngineRole",
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "KubernetesSecretEngineRole")
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
	}

	if err = (&controllers.DatabaseSecretEngineStaticRoleReconciler{
		ReconcilerBase: util.NewReconcilerBase(mgr.GetClient(), mgr.GetScheme(), mgr.GetConfig(), mgr.GetEventRecorderFor("DatabaseSecretEngineStaticRole"), mgr.GetAPIReader()),
		Log:            ctrl.Log.WithName("controllers").WithName("DatabaseSecretEngineStaticRole"),
		ControllerName: "DatabaseSecretEngineStaticRole",
	}).SetupWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create controller", "controller", "DatabaseSecretEngineStaticRole")
		os.Exit(1)
	}
	if err = (&redhatcopv1alpha1.DatabaseSecretEngineStaticRole{}).SetupWebhookWithManager(mgr); err != nil {
		setupLog.Error(err, "unable to create webhook", "webhook", "DatabaseSecretEngineStaticRole")
		os.Exit(1)
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
