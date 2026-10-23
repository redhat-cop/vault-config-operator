package utils

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
)

func TestCleanPath(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"strips /v1/ prefix", "/v1/sys/mounts", "sys/mounts"},
		{"strips v1/ without leading slash", "v1/sys/mounts", "sys/mounts"},
		{"no v1 prefix untouched", "/sys/mounts", "sys/mounts"},
		{"strips trailing slash", "/v1/secret/data/foo/", "secret/data/foo"},
		{"normalizes double slashes", "/v1/auth//kubernetes//login", "auth/kubernetes/login"},
		{"empty string", "", ""},
		{"only v1", "/v1/", ""},
		{"spaces trimmed", "  /v1/sys/health  ", "sys/health"},
		{"just a path", "auth/kubernetes/role/my-role", "auth/kubernetes/role/my-role"},
		{"deep nesting preserved", "/v1/a/b/c/d/e", "a/b/c/d/e"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cleanPath(tt.input)
			if got != tt.want {
				t.Errorf("cleanPath(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestNormalizePath(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		// Empty
		{"empty", "", ""},

		// sys/ paths
		{"sys/policies/acl named", "sys/policies/acl/my-policy", "sys/policies/acl/*"},
		{"sys/policies/acl deep", "sys/policies/acl/a/b/c", "sys/policies/acl/*"},
		{"sys/policies/egp named", "sys/policies/egp/my-policy", "sys/policies/egp/*"},
		{"sys/policies/rgp named", "sys/policies/rgp/my-policy", "sys/policies/rgp/*"},
		{"sys/policy named", "sys/policy/default", "sys/policy/*"},
		{"sys/auth named", "sys/auth/kubernetes", "sys/auth/*"},
		{"sys/audit named", "sys/audit/file", "sys/audit/*"},
		{"sys/mounts named", "sys/mounts/secret", "sys/mounts/*"},
		{"sys/namespaces named", "sys/namespaces/tenant-a", "sys/namespaces/*"},
		{"sys/leases/lookup named", "sys/leases/lookup/auth/token/abc", "sys/leases/lookup/*"},
		{"sys/leases/renew named", "sys/leases/renew/auth/token/abc", "sys/leases/renew/*"},
		{"sys/leases/revoke named", "sys/leases/revoke/auth/token/abc", "sys/leases/revoke/*"},
		{"sys/wrapping/lookup named", "sys/wrapping/lookup/token-id", "sys/wrapping/lookup/*"},
		{"sys/health passthrough", "sys/health", "sys/health"},
		{"sys/seal passthrough", "sys/seal", "sys/seal"},
		{"sys/mounts no name", "sys/mounts", "sys/mounts"},

		// auth/ paths
		{"auth login", "auth/kubernetes/login", "auth/*/login"},
		{"auth role with name", "auth/kubernetes/role/my-role", "auth/*/role/*"},
		{"auth config", "auth/approle/config", "auth/*/config"},
		{"auth role-id with name", "auth/approle/role/my-role", "auth/*/role/*"},

		// Secret engine paths
		{"kv data", "secret/data/my-secret", "*/data/*"},
		{"kv metadata", "secret/metadata/my-secret", "*/metadata/*"},
		{"database roles", "database/roles/my-role", "*/roles/*"},
		{"database config", "database/config/my-db", "*/config/*"},
		{"pki issue", "pki/issue/my-role", "*/issue/*"},
		{"transit encrypt", "transit/encrypt/my-key", "*/encrypt/*"},
		{"single segment mount", "cubbyhole", "cubbyhole"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizePath(tt.input)
			if got != tt.want {
				t.Errorf("normalizePath(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestInstrumentedRoundTripper(t *testing.T) {
	reg := prometheus.NewRegistry()

	counter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "test_vault_api_requests_total",
			Help: "test counter",
		},
		[]string{"vault_address", "method", "path", "status_code"},
	)
	histogram := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "test_vault_api_request_duration_seconds",
			Help:    "test histogram",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"vault_address", "method", "path"},
	)
	reg.MustRegister(counter, histogram)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	rt := newInstrumentedRoundTripperWithRegistry(http.DefaultTransport, "https://vault.example.com:8200", counter, histogram)

	req, err := http.NewRequest("GET", srv.URL+"/v1/sys/health", nil)
	if err != nil {
		t.Fatalf("creating request: %v", err)
	}

	resp, err := rt.RoundTrip(req)
	if err != nil {
		t.Fatalf("round trip: %v", err)
	}
	_ = resp.Body.Close()

	// Verify counter was incremented
	families, err := reg.Gather()
	if err != nil {
		t.Fatalf("gather: %v", err)
	}

	foundCounter := false
	foundHistogram := false
	for _, family := range families {
		switch family.GetName() {
		case "test_vault_api_requests_total":
			foundCounter = true
			if len(family.GetMetric()) != 1 {
				t.Errorf("expected 1 counter metric, got %d", len(family.GetMetric()))
			}
			m := family.GetMetric()[0]
			if m.GetCounter().GetValue() != 1 {
				t.Errorf("counter value = %v, want 1", m.GetCounter().GetValue())
			}
			labels := labelMap(m.GetLabel())
			if labels["vault_address"] != "https://vault.example.com:8200" {
				t.Errorf("vault_address = %q, want %q", labels["vault_address"], "https://vault.example.com:8200")
			}
			if labels["method"] != "GET" {
				t.Errorf("method = %q, want %q", labels["method"], "GET")
			}
			if labels["status_code"] != "200" {
				t.Errorf("status_code = %q, want %q", labels["status_code"], "200")
			}
		case "test_vault_api_request_duration_seconds":
			foundHistogram = true
			if len(family.GetMetric()) != 1 {
				t.Errorf("expected 1 histogram metric, got %d", len(family.GetMetric()))
			}
			m := family.GetMetric()[0]
			if m.GetHistogram().GetSampleCount() != 1 {
				t.Errorf("histogram sample count = %v, want 1", m.GetHistogram().GetSampleCount())
			}
		}
	}
	if !foundCounter {
		t.Error("counter metric not found in registry")
	}
	if !foundHistogram {
		t.Error("histogram metric not found in registry")
	}
}

func TestInstrumentedRoundTripperTransportError(t *testing.T) {
	reg := prometheus.NewRegistry()

	counter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "test_err_vault_api_requests_total",
			Help: "test counter",
		},
		[]string{"vault_address", "method", "path", "status_code"},
	)
	histogram := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "test_err_vault_api_request_duration_seconds",
			Help:    "test histogram",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"vault_address", "method", "path"},
	)
	reg.MustRegister(counter, histogram)

	rt := newInstrumentedRoundTripperWithRegistry(http.DefaultTransport, "https://vault.example.com:8200", counter, histogram)

	req, err := http.NewRequest("GET", "http://127.0.0.1:1/v1/sys/health", nil)
	if err != nil {
		t.Fatalf("creating request: %v", err)
	}

	_, err = rt.RoundTrip(req) //nolint:bodyclose // transport error, no body
	if err == nil {
		t.Fatal("expected transport error")
	}

	families, err := reg.Gather()
	if err != nil {
		t.Fatalf("gather: %v", err)
	}

	for _, family := range families {
		if family.GetName() == "test_err_vault_api_requests_total" {
			m := family.GetMetric()[0]
			labels := labelMap(m.GetLabel())
			if labels["status_code"] != "0" {
				t.Errorf("status_code = %q, want %q for transport error", labels["status_code"], "0")
			}
			if m.GetCounter().GetValue() != 1 {
				t.Errorf("counter = %v, want 1", m.GetCounter().GetValue())
			}
			return
		}
	}
	t.Error("counter metric not found after transport error")
}

func TestInstrumentedRoundTripperDoubleWrap(t *testing.T) {
	rt := newInstrumentedRoundTripper(http.DefaultTransport, "https://vault:8200")
	rt2 := newInstrumentedRoundTripper(rt, "https://vault:8200")
	if rt != rt2 {
		t.Error("double wrapping should return the same round tripper")
	}
}

func labelMap(labels []*dto.LabelPair) map[string]string {
	m := make(map[string]string, len(labels))
	for _, lp := range labels {
		m[lp.GetName()] = lp.GetValue()
	}
	return m
}
