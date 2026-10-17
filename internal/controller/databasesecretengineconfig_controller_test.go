//go:build !integration
// +build !integration

package controller

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	vault "github.com/hashicorp/vault/api"
	redhatcopv1alpha1 "github.com/redhat-cop/vault-config-operator/api/v1alpha1"
	vaultutils "github.com/redhat-cop/vault-config-operator/api/v1alpha1/utils"
	"github.com/redhat-cop/vault-config-operator/internal/controller/vaultresourcecontroller"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/client/interceptor"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	k8sschema "k8s.io/apimachinery/pkg/runtime/schema"
)

func controllerTestScheme() *runtime.Scheme {
	s := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(s))
	utilruntime.Must(redhatcopv1alpha1.SchemeBuilder.AddToScheme(s))
	return s
}

func newTestReconciler(k8sClient client.Client) *DatabaseSecretEngineConfigReconciler {
	logger := zap.New(zap.UseDevMode(true))
	recorder := record.NewFakeRecorder(10)
	base := vaultresourcecontroller.NewReconcilerBase(
		k8sClient, controllerTestScheme(), nil, recorder, k8sClient, logger, "test",
	)
	return &DatabaseSecretEngineConfigReconciler{ReconcilerBase: base}
}

func newFakeVaultServer(t *testing.T, rotateHandler http.HandlerFunc) (*vault.Client, *httptest.Server) {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/", rotateHandler)
	ts := httptest.NewServer(mux)
	cfg := vault.DefaultConfig()
	cfg.Address = ts.URL
	vc, err := vault.NewClient(cfg)
	if err != nil {
		t.Fatalf("failed to create vault client: %v", err)
	}
	return vc, ts
}

func vaultOKHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{})
}

func newTestInstance() *redhatcopv1alpha1.DatabaseSecretEngineConfig {
	return &redhatcopv1alpha1.DatabaseSecretEngineConfig{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-db-config",
			Namespace: "default",
		},
		Spec: redhatcopv1alpha1.DatabaseSecretEngineConfigSpec{
			DBSEConfig: redhatcopv1alpha1.DBSEConfig{
				PluginName:    "postgresql-database-plugin",
				ConnectionURL: "postgresql://{{username}}:{{password}}@localhost:5432",
				RootPasswordRotation: &redhatcopv1alpha1.RootPasswordRotation{
					Enable: true,
				},
			},
		},
	}
}

func TestRotateRootPassword_HappyPath(t *testing.T) {
	instance := newTestInstance()
	scheme := controllerTestScheme()

	k8sClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(instance).
		WithStatusSubresource(&redhatcopv1alpha1.DatabaseSecretEngineConfig{}).
		Build()

	reconciler := newTestReconciler(k8sClient)

	vc, ts := newFakeVaultServer(t, vaultOKHandler)
	defer ts.Close()

	ctx := context.Background()
	ctx = vaultutils.ContextWithVaultClient(ctx, vc)

	err := reconciler.rotateRootPassword(ctx, instance)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if instance.Status.LastRootPasswordRotation.IsZero() {
		t.Fatal("expected LastRootPasswordRotation to be set")
	}

	// Verify it was persisted
	fetched := &redhatcopv1alpha1.DatabaseSecretEngineConfig{}
	err = k8sClient.Get(ctx, types.NamespacedName{Name: "test-db-config", Namespace: "default"}, fetched)
	if err != nil {
		t.Fatalf("failed to fetch instance: %v", err)
	}
	if fetched.Status.LastRootPasswordRotation.IsZero() {
		t.Fatal("expected persisted LastRootPasswordRotation to be set")
	}
}

func TestRotateRootPassword_ConflictThenSuccess(t *testing.T) {
	instance := newTestInstance()
	scheme := controllerTestScheme()

	var statusUpdateCalls atomic.Int32

	// Build the real client first (for delegation), then wrap with interceptor.
	innerClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(instance).
		WithStatusSubresource(&redhatcopv1alpha1.DatabaseSecretEngineConfig{}).
		Build()

	wrappedClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(instance).
		WithStatusSubresource(&redhatcopv1alpha1.DatabaseSecretEngineConfig{}).
		WithInterceptorFuncs(interceptor.Funcs{
			SubResourceUpdate: func(ctx context.Context, c client.Client, subResourceName string, obj client.Object, opts ...client.SubResourceUpdateOption) error {
				call := statusUpdateCalls.Add(1)
				if call <= 2 {
					return apierrors.NewConflict(
						k8sschema.GroupResource{Group: "redhatcop.redhat.io", Resource: "databasesecretengineconfigs"},
						obj.GetName(),
						nil,
					)
				}
				return innerClient.Status().Update(ctx, obj, opts...)
			},
			Get: func(ctx context.Context, c client.WithWatch, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
				return innerClient.Get(ctx, key, obj, opts...)
			},
		}).
		Build()

	reconciler := newTestReconciler(wrappedClient)

	vc, ts := newFakeVaultServer(t, vaultOKHandler)
	defer ts.Close()

	ctx := context.Background()
	ctx = vaultutils.ContextWithVaultClient(ctx, vc)

	err := reconciler.rotateRootPassword(ctx, instance)
	if err != nil {
		t.Fatalf("expected no error after conflict retry, got: %v", err)
	}

	calls := statusUpdateCalls.Load()
	if calls < 3 {
		t.Fatalf("expected at least 3 status update attempts (2 conflicts + 1 success), got %d", calls)
	}

	if instance.Status.LastRootPasswordRotation.IsZero() {
		t.Fatal("expected LastRootPasswordRotation to be set after conflict retry")
	}
}

func TestRotateRootPassword_ExhaustedRetries(t *testing.T) {
	instance := newTestInstance()
	scheme := controllerTestScheme()

	// Always return conflict to exhaust retries.
	innerClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(instance).
		WithStatusSubresource(&redhatcopv1alpha1.DatabaseSecretEngineConfig{}).
		Build()

	wrappedClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(instance).
		WithStatusSubresource(&redhatcopv1alpha1.DatabaseSecretEngineConfig{}).
		WithInterceptorFuncs(interceptor.Funcs{
			SubResourceUpdate: func(ctx context.Context, c client.Client, subResourceName string, obj client.Object, opts ...client.SubResourceUpdateOption) error {
				return apierrors.NewConflict(
					k8sschema.GroupResource{Group: "redhatcop.redhat.io", Resource: "databasesecretengineconfigs"},
					obj.GetName(),
					nil,
				)
			},
			Get: func(ctx context.Context, c client.WithWatch, key client.ObjectKey, obj client.Object, opts ...client.GetOption) error {
				return innerClient.Get(ctx, key, obj, opts...)
			},
		}).
		Build()

	reconciler := newTestReconciler(wrappedClient)

	vc, ts := newFakeVaultServer(t, vaultOKHandler)
	defer ts.Close()

	ctx := context.Background()
	ctx = vaultutils.ContextWithVaultClient(ctx, vc)

	err := reconciler.rotateRootPassword(ctx, instance)
	if err == nil {
		t.Fatal("expected an error after exhausting retries, got nil")
	}
	if !apierrors.IsConflict(err) {
		t.Fatalf("expected conflict error, got: %v", err)
	}
}

func TestRotateRootPassword_VaultFailure(t *testing.T) {
	instance := newTestInstance()
	scheme := controllerTestScheme()

	k8sClient := fake.NewClientBuilder().
		WithScheme(scheme).
		WithObjects(instance).
		WithStatusSubresource(&redhatcopv1alpha1.DatabaseSecretEngineConfig{}).
		Build()

	reconciler := newTestReconciler(k8sClient)

	// Vault returns 500 on rotation attempt
	vaultErrorHandler := func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}
	vc, ts := newFakeVaultServer(t, vaultErrorHandler)
	defer ts.Close()

	ctx := context.Background()
	ctx = vaultutils.ContextWithVaultClient(ctx, vc)

	err := reconciler.rotateRootPassword(ctx, instance)
	if err == nil {
		t.Fatal("expected an error from Vault failure, got nil")
	}

	// Status should NOT have been updated since Vault call failed
	if !instance.Status.LastRootPasswordRotation.IsZero() {
		t.Fatal("expected LastRootPasswordRotation to remain zero after Vault failure")
	}
}
