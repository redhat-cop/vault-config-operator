package v1alpha1

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	vault "github.com/hashicorp/vault/api"
	vaultutils "github.com/redhat-cop/vault-config-operator/api/v1alpha1/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func testScheme() *runtime.Scheme {
	s := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(s))
	utilruntime.Must(SchemeBuilder.AddToScheme(s))
	return s
}

func newFakeKubeClient(objs ...client.Object) client.Client {
	return fake.NewClientBuilder().
		WithScheme(testScheme()).
		WithObjects(objs...).
		WithStatusSubresource(&EntityAlias{}, &GroupAlias{}).
		Build()
}

type fakeVaultHandler struct {
	routes map[string]map[string]interface{}
}

func newFakeVaultHandler() *fakeVaultHandler {
	return &fakeVaultHandler{routes: make(map[string]map[string]interface{})}
}

func (h *fakeVaultHandler) setGet(path string, data map[string]interface{}) {
	h.routes[path] = data
}

func (h *fakeVaultHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path[len("/v1/"):]

	switch r.Method {
	case http.MethodGet:
		data, ok := h.routes[path]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		resp := map[string]interface{}{"data": data}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	case http.MethodPut, http.MethodPost:
		data, ok := h.routes[path]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		resp := map[string]interface{}{"data": data}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func newFakeVaultClient(t *testing.T, handler http.Handler) (*vault.Client, *httptest.Server) {
	t.Helper()
	ts := httptest.NewServer(handler)
	cfg := vault.DefaultConfig()
	cfg.Address = ts.URL
	vc, err := vault.NewClient(cfg)
	if err != nil {
		t.Fatalf("failed to create vault client: %v", err)
	}
	return vc, ts
}

func pivContext(kubeClient client.Client, vaultClient *vault.Client) context.Context {
	ctx := context.Background()
	ctx = vaultutils.ContextWithKubeClient(ctx, kubeClient)
	ctx = vaultutils.ContextWithVaultClient(ctx, vaultClient)
	return ctx
}

func pivContextWithRestConfig(kubeClient client.Client, vaultClient *vault.Client, restConfig *rest.Config) context.Context {
	ctx := pivContext(kubeClient, vaultClient)
	ctx = vaultutils.ContextWithRestConfig(ctx, restConfig)
	return ctx
}

func newK8sSecret(namespace, name string, data map[string][]byte) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: name},
		Data:       data,
	}
}

func newTypedK8sSecret(namespace, name string, secretType corev1.SecretType, data map[string][]byte) *corev1.Secret {
	return &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{Namespace: namespace, Name: name},
		Type:       secretType,
		Data:       data,
	}
}
