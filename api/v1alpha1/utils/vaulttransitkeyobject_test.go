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

// mockTransitKeyObject implements VaultTransitKeyObject for testing.
type mockTransitKeyObject struct {
	path          string
	payload       map[string]any
	configPath    string
	configPayload map[string]any
	equivalent    bool
}

func (m *mockTransitKeyObject) GetPath() string                                  { return m.path }
func (m *mockTransitKeyObject) GetPayload() map[string]any                       { return m.payload }
func (m *mockTransitKeyObject) GetConfigPath() string                            { return m.configPath }
func (m *mockTransitKeyObject) GetConfigPayload() map[string]any                 { return m.configPayload }
func (m *mockTransitKeyObject) IsEquivalentToDesiredState(_ map[string]any) bool { return m.equivalent }
func (m *mockTransitKeyObject) IsInitialized() bool                              { return true }
func (m *mockTransitKeyObject) IsValid() (bool, error)                           { return true, nil }
func (m *mockTransitKeyObject) IsDeletable() bool                                { return true }
func (m *mockTransitKeyObject) PrepareInternalValues(_ context.Context, _ client.Object) error {
	return nil
}
func (m *mockTransitKeyObject) PrepareTLSConfig(_ context.Context, _ client.Object) error {
	return nil
}
func (m *mockTransitKeyObject) GetKubeAuthConfiguration() *KubeAuthConfiguration { return nil }
func (m *mockTransitKeyObject) GetVaultConnection() *VaultConnection             { return nil }

func TestTransitKeyCreateOrUpdate_CreatesKeyAndAppliesConfig(t *testing.T) {
	store := newFakeVaultStore()
	vaultClient, ts := newTestClient(t, store)
	defer ts.Close()

	obj := &mockTransitKeyObject{
		path:       "transit/keys/mykey",
		payload:    map[string]any{"type": "aes256-gcm96"},
		configPath: "transit/keys/mykey/config",
		configPayload: map[string]any{
			"min_decryption_version": 1,
			"min_encryption_version": 1,
			"deletion_allowed":       true,
		},
		equivalent: false,
	}
	ve := &VaultTransitKeyEndpoint{transitKeyObject: obj}
	ctx := newTestContext(vaultClient)

	err := ve.CreateOrUpdate(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	created, ok := store.get("transit/keys/mykey")
	if !ok {
		t.Fatal("expected key to be created at transit/keys/mykey")
	}
	if created["type"] != "aes256-gcm96" {
		t.Errorf("expected type=aes256-gcm96, got %v", created["type"])
	}

	config, ok := store.get("transit/keys/mykey/config")
	if !ok {
		t.Fatal("expected config to be written at transit/keys/mykey/config")
	}
	if config["deletion_allowed"] != true {
		t.Errorf("expected deletion_allowed=true, got %v", config["deletion_allowed"])
	}
	if config["min_decryption_version"] != float64(1) {
		t.Errorf("expected min_decryption_version=1, got %v", config["min_decryption_version"])
	}
	if config["min_encryption_version"] != float64(1) {
		t.Errorf("expected min_encryption_version=1, got %v", config["min_encryption_version"])
	}
}

func TestTransitKeyCreateOrUpdate_UpdatesConfigWhenDrifted(t *testing.T) {
	store := newFakeVaultStore()
	store.set("transit/keys/mykey", map[string]any{"type": "aes256-gcm96"})
	vaultClient, ts := newTestClient(t, store)
	defer ts.Close()

	obj := &mockTransitKeyObject{
		path:       "transit/keys/mykey",
		payload:    map[string]any{"type": "aes256-gcm96"},
		configPath: "transit/keys/mykey/config",
		configPayload: map[string]any{
			"min_decryption_version": 2,
			"deletion_allowed":       true,
		},
		equivalent: false,
	}
	ve := &VaultTransitKeyEndpoint{transitKeyObject: obj}
	ctx := newTestContext(vaultClient)

	err := ve.CreateOrUpdate(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	config, ok := store.get("transit/keys/mykey/config")
	if !ok {
		t.Fatal("expected config to be written at transit/keys/mykey/config")
	}
	if config["min_decryption_version"] != float64(2) {
		t.Errorf("expected min_decryption_version=2, got %v", config["min_decryption_version"])
	}
	if config["deletion_allowed"] != true {
		t.Errorf("expected deletion_allowed=true, got %v", config["deletion_allowed"])
	}
}

func TestTransitKeyCreateOrUpdate_NoOpWhenEquivalent(t *testing.T) {
	store := newFakeVaultStore()
	store.set("transit/keys/mykey", map[string]any{"type": "aes256-gcm96"})
	vaultClient, ts := newTestClient(t, store)
	defer ts.Close()

	obj := &mockTransitKeyObject{
		path:          "transit/keys/mykey",
		payload:       map[string]any{"type": "aes256-gcm96"},
		configPath:    "transit/keys/mykey/config",
		configPayload: map[string]any{"deletion_allowed": false},
		equivalent:    true,
	}
	ve := &VaultTransitKeyEndpoint{transitKeyObject: obj}
	ctx := newTestContext(vaultClient)

	err := ve.CreateOrUpdate(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, ok := store.get("transit/keys/mykey/config")
	if ok {
		t.Error("expected no config write when state is equivalent")
	}
}

func TestTransitKeyDeleteIfExists_DeletesKey(t *testing.T) {
	store := newFakeVaultStore()
	store.set("transit/keys/mykey", map[string]any{"type": "aes256-gcm96"})
	vaultClient, ts := newTestClient(t, store)
	defer ts.Close()

	obj := &mockTransitKeyObject{
		path:       "transit/keys/mykey",
		configPath: "transit/keys/mykey/config",
	}
	ve := &VaultTransitKeyEndpoint{transitKeyObject: obj}
	ctx := newTestContext(vaultClient)

	err := ve.DeleteIfExists(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTransitKeyCreateOrUpdate_ConfigFailureAfterCreate_RollsBackKey(t *testing.T) {
	var mu sync.Mutex
	data := map[string]map[string]any{}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path[len("/v1/"):]
		switch r.Method {
		case http.MethodGet:
			mu.Lock()
			v, ok := data[path]
			mu.Unlock()
			if !ok {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			resp := map[string]any{"data": v}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		case http.MethodPut, http.MethodPost:
			if path == "transit/keys/mykey/config" {
				w.WriteHeader(http.StatusInternalServerError)
				_ = json.NewEncoder(w).Encode(map[string]any{"errors": []string{"permission denied"}})
				return
			}
			var body map[string]any
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			mu.Lock()
			data[path] = body
			mu.Unlock()
			w.WriteHeader(http.StatusNoContent)
		case http.MethodDelete:
			mu.Lock()
			delete(data, path)
			mu.Unlock()
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	})

	ts := httptest.NewServer(handler)
	defer ts.Close()

	cfg := vault.DefaultConfig()
	cfg.Address = ts.URL
	vaultClient, err := vault.NewClient(cfg)
	if err != nil {
		t.Fatalf("failed to create vault client: %v", err)
	}

	obj := &mockTransitKeyObject{
		path:       "transit/keys/mykey",
		payload:    map[string]any{"type": "aes256-gcm96"},
		configPath: "transit/keys/mykey/config",
		configPayload: map[string]any{
			"min_decryption_version": 1,
			"deletion_allowed":       true,
		},
		equivalent: false,
	}
	ve := &VaultTransitKeyEndpoint{transitKeyObject: obj}
	ctx := newTestContext(vaultClient)

	err = ve.CreateOrUpdate(ctx)
	if err == nil {
		t.Fatal("expected error when config write fails after create, got nil")
	}

	mu.Lock()
	_, keyExists := data["transit/keys/mykey"]
	mu.Unlock()
	if keyExists {
		t.Fatal("expected key to be rolled back (deleted) after config failure, but it still exists")
	}
}
