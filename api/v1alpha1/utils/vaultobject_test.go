package utils

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	vault "github.com/hashicorp/vault/api"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// mockVaultObject implements VaultObject for testing.
type mockVaultObject struct {
	path    string
	payload map[string]interface{}
}

func (m *mockVaultObject) GetPath() string                                    { return m.path }
func (m *mockVaultObject) GetPayload() map[string]interface{}                 { return m.payload }
func (m *mockVaultObject) IsEquivalentToDesiredState(_ map[string]interface{}) bool { return false }
func (m *mockVaultObject) IsInitialized() bool                                { return true }
func (m *mockVaultObject) IsValid() (bool, error)                             { return true, nil }
func (m *mockVaultObject) IsDeletable() bool                                  { return true }
func (m *mockVaultObject) PrepareInternalValues(_ context.Context, _ client.Object) error {
	return nil
}
func (m *mockVaultObject) PrepareTLSConfig(_ context.Context, _ client.Object) error { return nil }
func (m *mockVaultObject) GetKubeAuthConfiguration() *KubeAuthConfiguration        { return nil }
func (m *mockVaultObject) GetVaultConnection() *VaultConnection                    { return nil }

// fakeVaultStore holds in-memory KV data and serves Vault-compatible HTTP responses.
type fakeVaultStore struct {
	mu   sync.Mutex
	data map[string]map[string]interface{}
}

func newFakeVaultStore() *fakeVaultStore {
	return &fakeVaultStore{data: make(map[string]map[string]interface{})}
}

func (s *fakeVaultStore) set(path string, payload map[string]interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[path] = payload
}

func (s *fakeVaultStore) get(path string) (map[string]interface{}, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	v, ok := s.data[path]
	return v, ok
}

func (s *fakeVaultStore) handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path[len("/v1/"):]
		switch r.Method {
		case http.MethodGet:
			data, ok := s.get(path)
			if !ok {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			resp := map[string]interface{}{"data": data}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		case http.MethodPut, http.MethodPost:
			var body map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			s.set(path, body)
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})
}

func newTestContext(client *vault.Client) context.Context {
	return context.WithValue(context.Background(), "vaultClient", client)
}

func newTestClient(t *testing.T, store *fakeVaultStore) (*vault.Client, *httptest.Server) {
	t.Helper()
	ts := httptest.NewServer(store.handler())
	cfg := vault.DefaultConfig()
	cfg.Address = ts.URL
	client, err := vault.NewClient(cfg)
	if err != nil {
		t.Fatalf("failed to create vault client: %v", err)
	}
	return client, ts
}

// --- KVv1 tests ---

func TestCreateOrMergeKV_KVv1_CreatesWhenPathNotFound(t *testing.T) {
	store := newFakeVaultStore()
	client, ts := newTestClient(t, store)
	defer ts.Close()

	obj := &mockVaultObject{
		path:    "secret/myapp",
		payload: map[string]interface{}{"password": "abc123"},
	}
	ve := &VaultEndpoint{vaultObject: obj}
	ctx := newTestContext(client)

	err := ve.CreateOrMergeKV(ctx, false, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	stored, ok := store.get("secret/myapp")
	if !ok {
		t.Fatal("expected secret to be created")
	}
	if stored["password"] != "abc123" {
		t.Errorf("expected password=abc123, got %v", stored["password"])
	}
}

func TestCreateOrMergeKV_KVv1_MergesNewKey(t *testing.T) {
	store := newFakeVaultStore()
	store.set("secret/myapp", map[string]interface{}{"password": "existing-pw"})
	client, ts := newTestClient(t, store)
	defer ts.Close()

	obj := &mockVaultObject{
		path:    "secret/myapp",
		payload: map[string]interface{}{"username": "admin"},
	}
	ve := &VaultEndpoint{vaultObject: obj}
	ctx := newTestContext(client)

	err := ve.CreateOrMergeKV(ctx, false, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	stored, _ := store.get("secret/myapp")
	if stored["password"] != "existing-pw" {
		t.Errorf("existing key should be preserved, got %v", stored["password"])
	}
	if stored["username"] != "admin" {
		t.Errorf("new key should be added, got %v", stored["username"])
	}
}

func TestCreateOrMergeKV_KVv1_OverwritesExistingKey(t *testing.T) {
	store := newFakeVaultStore()
	store.set("secret/myapp", map[string]interface{}{"password": "old-value"})
	client, ts := newTestClient(t, store)
	defer ts.Close()

	obj := &mockVaultObject{
		path:    "secret/myapp",
		payload: map[string]interface{}{"password": "new-value"},
	}
	ve := &VaultEndpoint{vaultObject: obj}
	ctx := newTestContext(client)

	err := ve.CreateOrMergeKV(ctx, false, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	stored, _ := store.get("secret/myapp")
	if stored["password"] != "new-value" {
		t.Errorf("expected overwritten password=new-value, got %v", stored["password"])
	}
}

func TestCreateOrMergeKV_KVv1_PreservesExistingKey(t *testing.T) {
	store := newFakeVaultStore()
	store.set("secret/myapp", map[string]interface{}{"password": "keep-this"})
	client, ts := newTestClient(t, store)
	defer ts.Close()

	obj := &mockVaultObject{
		path:    "secret/myapp",
		payload: map[string]interface{}{"password": "would-overwrite"},
	}
	ve := &VaultEndpoint{vaultObject: obj}
	ctx := newTestContext(client)

	err := ve.CreateOrMergeKV(ctx, false, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	stored, _ := store.get("secret/myapp")
	if stored["password"] != "keep-this" {
		t.Errorf("expected preserved password=keep-this, got %v", stored["password"])
	}
}

func TestCreateOrMergeKV_KVv1_PreserveAddsNewKeyButKeepsExisting(t *testing.T) {
	store := newFakeVaultStore()
	store.set("secret/myapp", map[string]interface{}{"password": "keep-this"})
	client, ts := newTestClient(t, store)
	defer ts.Close()

	obj := &mockVaultObject{
		path: "secret/myapp",
		payload: map[string]interface{}{
			"password": "would-overwrite",
			"username": "new-user",
		},
	}
	ve := &VaultEndpoint{vaultObject: obj}
	ctx := newTestContext(client)

	err := ve.CreateOrMergeKV(ctx, false, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	stored, _ := store.get("secret/myapp")
	if stored["password"] != "keep-this" {
		t.Errorf("existing key should be preserved, got %v", stored["password"])
	}
	if stored["username"] != "new-user" {
		t.Errorf("new key should be added, got %v", stored["username"])
	}
}

func TestCreateOrMergeKV_KVv1_PreserveCreatesWhenPathNotFound(t *testing.T) {
	store := newFakeVaultStore()
	client, ts := newTestClient(t, store)
	defer ts.Close()

	obj := &mockVaultObject{
		path:    "secret/myapp",
		payload: map[string]interface{}{"password": "first-value"},
	}
	ve := &VaultEndpoint{vaultObject: obj}
	ctx := newTestContext(client)

	err := ve.CreateOrMergeKV(ctx, false, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	stored, ok := store.get("secret/myapp")
	if !ok {
		t.Fatal("expected secret to be created")
	}
	if stored["password"] != "first-value" {
		t.Errorf("expected password=first-value, got %v", stored["password"])
	}
}

// --- KVv2 tests ---

func TestCreateOrMergeKV_KVv2_CreatesWhenPathNotFound(t *testing.T) {
	store := newFakeVaultStore()
	client, ts := newTestClient(t, store)
	defer ts.Close()

	obj := &mockVaultObject{
		path: "secret/data/myapp",
		payload: map[string]interface{}{
			"data": map[string]interface{}{"password": "abc123"},
		},
	}
	ve := &VaultEndpoint{vaultObject: obj}
	ctx := newTestContext(client)

	err := ve.CreateOrMergeKV(ctx, true, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	stored, ok := store.get("secret/data/myapp")
	if !ok {
		t.Fatal("expected secret to be created")
	}
	data, _ := stored["data"].(map[string]interface{})
	if data["password"] != "abc123" {
		t.Errorf("expected password=abc123, got %v", data["password"])
	}
}

func TestCreateOrMergeKV_KVv2_MergesNewKey(t *testing.T) {
	store := newFakeVaultStore()
	store.set("secret/data/myapp", map[string]interface{}{
		"data": map[string]interface{}{"password": "existing-pw"},
	})
	client, ts := newTestClient(t, store)
	defer ts.Close()

	obj := &mockVaultObject{
		path: "secret/data/myapp",
		payload: map[string]interface{}{
			"data": map[string]interface{}{"username": "admin"},
		},
	}
	ve := &VaultEndpoint{vaultObject: obj}
	ctx := newTestContext(client)

	err := ve.CreateOrMergeKV(ctx, true, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	stored, _ := store.get("secret/data/myapp")
	data, _ := stored["data"].(map[string]interface{})
	if data["password"] != "existing-pw" {
		t.Errorf("existing key should be preserved during merge, got %v", data["password"])
	}
	if data["username"] != "admin" {
		t.Errorf("new key should be added, got %v", data["username"])
	}
}

func TestCreateOrMergeKV_KVv2_OverwritesExistingKey(t *testing.T) {
	store := newFakeVaultStore()
	store.set("secret/data/myapp", map[string]interface{}{
		"data": map[string]interface{}{"password": "old-value"},
	})
	client, ts := newTestClient(t, store)
	defer ts.Close()

	obj := &mockVaultObject{
		path: "secret/data/myapp",
		payload: map[string]interface{}{
			"data": map[string]interface{}{"password": "new-value"},
		},
	}
	ve := &VaultEndpoint{vaultObject: obj}
	ctx := newTestContext(client)

	err := ve.CreateOrMergeKV(ctx, true, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	stored, _ := store.get("secret/data/myapp")
	data, _ := stored["data"].(map[string]interface{})
	if data["password"] != "new-value" {
		t.Errorf("expected overwritten password=new-value, got %v", data["password"])
	}
}

func TestCreateOrMergeKV_KVv2_PreservesExistingKey(t *testing.T) {
	store := newFakeVaultStore()
	store.set("secret/data/myapp", map[string]interface{}{
		"data": map[string]interface{}{"password": "keep-this"},
	})
	client, ts := newTestClient(t, store)
	defer ts.Close()

	obj := &mockVaultObject{
		path: "secret/data/myapp",
		payload: map[string]interface{}{
			"data": map[string]interface{}{"password": "would-overwrite"},
		},
	}
	ve := &VaultEndpoint{vaultObject: obj}
	ctx := newTestContext(client)

	err := ve.CreateOrMergeKV(ctx, true, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	stored, _ := store.get("secret/data/myapp")
	data, _ := stored["data"].(map[string]interface{})
	if data["password"] != "keep-this" {
		t.Errorf("expected preserved password=keep-this, got %v", data["password"])
	}
}

func TestCreateOrMergeKV_KVv2_PreserveAddsNewKeyButKeepsExisting(t *testing.T) {
	store := newFakeVaultStore()
	store.set("secret/data/myapp", map[string]interface{}{
		"data": map[string]interface{}{"password": "keep-this"},
	})
	client, ts := newTestClient(t, store)
	defer ts.Close()

	obj := &mockVaultObject{
		path: "secret/data/myapp",
		payload: map[string]interface{}{
			"data": map[string]interface{}{
				"password": "would-overwrite",
				"username": "new-user",
			},
		},
	}
	ve := &VaultEndpoint{vaultObject: obj}
	ctx := newTestContext(client)

	err := ve.CreateOrMergeKV(ctx, true, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	stored, _ := store.get("secret/data/myapp")
	data, _ := stored["data"].(map[string]interface{})
	if data["password"] != "keep-this" {
		t.Errorf("existing key should be preserved, got %v", data["password"])
	}
	if data["username"] != "new-user" {
		t.Errorf("new key should be added, got %v", data["username"])
	}
}

func TestCreateOrMergeKV_KVv2_PreserveCreatesWhenPathNotFound(t *testing.T) {
	store := newFakeVaultStore()
	client, ts := newTestClient(t, store)
	defer ts.Close()

	obj := &mockVaultObject{
		path: "secret/data/myapp",
		payload: map[string]interface{}{
			"data": map[string]interface{}{"password": "first-value"},
		},
	}
	ve := &VaultEndpoint{vaultObject: obj}
	ctx := newTestContext(client)

	err := ve.CreateOrMergeKV(ctx, true, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	stored, ok := store.get("secret/data/myapp")
	if !ok {
		t.Fatal("expected secret to be created")
	}
	data, _ := stored["data"].(map[string]interface{})
	if data["password"] != "first-value" {
		t.Errorf("expected password=first-value, got %v", data["password"])
	}
}
