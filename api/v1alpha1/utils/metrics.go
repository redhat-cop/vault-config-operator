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

package utils

import (
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
)

var (
	normalizePaths bool

	vaultAPIRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "vault_api_requests_total",
			Help: "Total number of Vault API requests.",
		},
		[]string{"vault_address", "method", "path", "status_code"},
	)

	vaultAPIRequestDurationSeconds = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "vault_api_request_duration_seconds",
			Help:    "Duration of Vault API requests in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"vault_address", "method", "path"},
	)
)

func init() {
	metrics.Registry.MustRegister(vaultAPIRequestsTotal, vaultAPIRequestDurationSeconds)
	normalizePaths = strings.EqualFold(os.Getenv("VAULT_METRICS_NORMALIZE_PATHS"), "true")
}

// instrumentedRoundTripper wraps an http.RoundTripper to record Prometheus
// metrics for every Vault HTTP request. It must be installed after TLS
// configuration is complete.
type instrumentedRoundTripper struct {
	inner        http.RoundTripper
	vaultAddress string
	counter      *prometheus.CounterVec
	histogram    *prometheus.HistogramVec
}

// newInstrumentedRoundTripper wraps rt with metrics instrumentation.
// If rt is already an *instrumentedRoundTripper, it is returned unchanged.
func newInstrumentedRoundTripper(rt http.RoundTripper, vaultAddress string) http.RoundTripper {
	if _, already := rt.(*instrumentedRoundTripper); already {
		return rt
	}
	return &instrumentedRoundTripper{
		inner:        rt,
		vaultAddress: vaultAddress,
		counter:      vaultAPIRequestsTotal,
		histogram:    vaultAPIRequestDurationSeconds,
	}
}

// newInstrumentedRoundTripperWithRegistry creates an instrumented round tripper
// backed by the given counter and histogram (for test isolation).
func newInstrumentedRoundTripperWithRegistry(rt http.RoundTripper, vaultAddress string, counter *prometheus.CounterVec, histogram *prometheus.HistogramVec) http.RoundTripper {
	if _, already := rt.(*instrumentedRoundTripper); already {
		return rt
	}
	return &instrumentedRoundTripper{
		inner:        rt,
		vaultAddress: vaultAddress,
		counter:      counter,
		histogram:    histogram,
	}
}

func (t *instrumentedRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	path := metricsPath(req.URL.Path)
	start := time.Now()

	resp, err := t.inner.RoundTrip(req)

	duration := time.Since(start).Seconds()
	statusCode := "0"
	if resp != nil {
		statusCode = fmt.Sprintf("%d", resp.StatusCode)
	}

	t.counter.WithLabelValues(t.vaultAddress, req.Method, path, statusCode).Inc()
	t.histogram.WithLabelValues(t.vaultAddress, req.Method, path).Observe(duration)

	return resp, err
}

// metricsPath applies cleanPath and, when normalization is enabled, normalizePath.
func metricsPath(raw string) string {
	cleaned := cleanPath(raw)
	if normalizePaths {
		return normalizePath(cleaned)
	}
	return cleaned
}

// cleanPath strips the /v1/ prefix (or v1/ if no leading slash) and
// normalizes repeated slashes and surrounding whitespace.
func cleanPath(raw string) string {
	p := strings.TrimSpace(raw)
	p = strings.TrimPrefix(p, "/")
	p = strings.TrimPrefix(p, "v1/")
	for strings.Contains(p, "//") {
		p = strings.ReplaceAll(p, "//", "/")
	}
	p = strings.TrimSuffix(p, "/")
	return p
}

// Compiled patterns for normalizePath — each captures a well-known prefix and
// replaces dynamic trailing segments with *.
var (
	// sys/ paths: sys/<subsystem>/<name> → sys/<subsystem>/*
	sysNamedRe = regexp.MustCompile(`^(sys/(?:policies/acl|policies/egp|policies/rgp|policy|auth|audit|mounts|namespaces|leases/lookup|leases/renew|leases/revoke|wrapping/lookup))/(.+)$`)

	// auth/<mount>/... paths: auth/<mount>/<verb>/<name> → auth/*/<verb>/*
	authRe = regexp.MustCompile(`^auth/([^/]+)/(.+)$`)

	// <mount>/... paths for secret engines: <mount>/<subpath> → */<subpath-normalized>
	secretEngineRe = regexp.MustCompile(`^([^/]+)/(.+)$`)
)

// normalizePath collapses dynamic path segments to * for metric cardinality control.
func normalizePath(path string) string {
	if path == "" {
		return path
	}

	// sys/ paths
	if strings.HasPrefix(path, "sys/") {
		if m := sysNamedRe.FindStringSubmatch(path); m != nil {
			return m[1] + "/*"
		}
		return path
	}

	// auth/ paths: auth/<mount>/<action>/<name> → auth/*/<action>/*
	if strings.HasPrefix(path, "auth/") {
		if m := authRe.FindStringSubmatch(path); m != nil {
			subpath := m[2]
			parts := strings.SplitN(subpath, "/", 2)
			if len(parts) == 2 {
				return "auth/*/" + parts[0] + "/*"
			}
			return "auth/*/" + parts[0]
		}
		return path
	}

	// Secret engine paths: <mount>/<subpath> → */<collapsed>
	if m := secretEngineRe.FindStringSubmatch(path); m != nil {
		subpath := m[2]
		parts := strings.Split(subpath, "/")
		normalized := make([]string, 0, len(parts))
		for i, p := range parts {
			if i%2 == 1 {
				normalized = append(normalized, "*")
			} else {
				normalized = append(normalized, p)
			}
		}
		return "*/" + strings.Join(normalized, "/")
	}

	return path
}

// wrapConfigTransport instruments the HTTP transport inside a vault.Config
// with Prometheus metrics. It is safe to call multiple times; the type-guard
// in newInstrumentedRoundTripper prevents double wrapping.
func wrapConfigTransport(httpClient *http.Client, vaultAddress string) {
	if httpClient == nil || httpClient.Transport == nil {
		return
	}
	httpClient.Transport = newInstrumentedRoundTripper(httpClient.Transport, vaultAddress)
}
