package utils

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net/http"
	"testing"
	"time"

	vault "github.com/hashicorp/vault/api"
	corev1 "k8s.io/api/core/v1"
)

func TestApplyConnectionTLS_ClientCertAndCAFromSecret(t *testing.T) {
	material := newTestTLSMaterial(t)
	serverName := "vault.example.com"
	spec := &TLSConfig{
		TLSSecret:     &corev1.LocalObjectReference{Name: "vault-client-tls"},
		TLSServerName: &serverName,
		SkipVerify:    true,
	}
	config := newVaultConfig(t)

	if err := applyConnectionTLS(config, spec, map[string][]byte{
		"ca.crt":  material.caPEM,
		"tls.crt": material.certPEM,
		"tls.key": material.keyPEM,
	}); err != nil {
		t.Fatalf("applyConnectionTLS() error = %v", err)
	}

	transport := tlsTransport(t, config)
	if !transport.TLSClientConfig.InsecureSkipVerify {
		t.Fatal("InsecureSkipVerify = false, want true")
	}
	if transport.TLSClientConfig.ServerName != serverName {
		t.Fatalf("ServerName = %q, want %q", transport.TLSClientConfig.ServerName, serverName)
	}
	if transport.TLSClientConfig.GetClientCertificate == nil {
		t.Fatal("GetClientCertificate is nil")
	}
	got, err := transport.TLSClientConfig.GetClientCertificate(nil)
	if err != nil {
		t.Fatalf("GetClientCertificate() error = %v", err)
	}
	if got == nil || len(got.Certificate) == 0 || !bytes.Equal(got.Certificate[0], material.client.Raw) {
		t.Fatal("GetClientCertificate() did not return the certificate from tls.crt")
	}

	opts := x509.VerifyOptions{
		Roots:     transport.TLSClientConfig.RootCAs,
		DNSName:   serverName,
		KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}
	if _, err := material.server.Verify(opts); err != nil {
		t.Fatalf("server certificate was not trusted with ca.crt: %v", err)
	}
	if _, err := material.otherServer.Verify(opts); err == nil {
		t.Fatal("server certificate from an unrelated CA was trusted")
	}
}

func TestApplyConnectionTLS_CAOnlyLeavesClientCertificateUnset(t *testing.T) {
	material := newTestTLSMaterial(t)
	config := newVaultConfig(t)
	spec := &TLSConfig{TLSSecret: &corev1.LocalObjectReference{Name: "vault-ca"}}

	if err := applyConnectionTLS(config, spec, map[string][]byte{"ca.crt": material.caPEM}); err != nil {
		t.Fatalf("applyConnectionTLS() error = %v", err)
	}

	transport := tlsTransport(t, config)
	if transport.TLSClientConfig.GetClientCertificate != nil {
		t.Fatal("GetClientCertificate is set, want nil when tls.crt and tls.key are absent")
	}
	opts := x509.VerifyOptions{Roots: transport.TLSClientConfig.RootCAs, DNSName: "vault.example.com"}
	if _, err := material.server.Verify(opts); err != nil {
		t.Fatalf("server certificate was not trusted with ca.crt: %v", err)
	}
}

func TestApplyConnectionTLS_NoSecretKeepsDefaultClientCertificate(t *testing.T) {
	config := newVaultConfig(t)
	if err := applyConnectionTLS(config, &TLSConfig{}, nil); err != nil {
		t.Fatalf("applyConnectionTLS() error = %v", err)
	}
	transport := tlsTransport(t, config)
	if transport.TLSClientConfig.GetClientCertificate != nil {
		t.Fatal("GetClientCertificate is set, want nil")
	}
}

func TestApplyConnectionTLS_RejectsIncompleteOrInvalidPair(t *testing.T) {
	material := newTestTLSMaterial(t)
	tests := []struct {
		name string
		data map[string][]byte
	}{
		{name: "cert without key", data: map[string][]byte{"tls.crt": material.certPEM}},
		{name: "key without cert", data: map[string][]byte{"tls.key": material.keyPEM}},
		{name: "invalid pem", data: map[string][]byte{"tls.crt": []byte("not a certificate"), "tls.key": []byte("not a key")}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := applyConnectionTLS(newVaultConfig(t), &TLSConfig{
				TLSSecret: &corev1.LocalObjectReference{Name: "vault-client-tls"},
			}, tt.data)
			if err == nil {
				t.Fatal("applyConnectionTLS() error = nil, want invalid pair")
			}
			if !bytes.Contains([]byte(err.Error()), []byte(`secret "vault-client-tls"`)) {
				t.Fatalf("applyConnectionTLS() error = %q, want the secret name", err)
			}
		})
	}
}

func TestSetClientCertificateFromPEM_UnsupportedTransport(t *testing.T) {
	material := newTestTLSMaterial(t)

	t.Run("nil client", func(t *testing.T) {
		err := setClientCertificateFromPEM(&vault.Config{}, material.certPEM, material.keyPEM, "vault-client-tls")
		if err == nil {
			t.Fatal("setClientCertificateFromPEM() error = nil, want unsupported transport")
		}
	})

	t.Run("non-http transport", func(t *testing.T) {
		config := newVaultConfig(t)
		config.HttpClient.Transport = roundTripperFunc(func(*http.Request) (*http.Response, error) {
			return nil, nil
		})
		err := setClientCertificateFromPEM(config, material.certPEM, material.keyPEM, "vault-client-tls")
		if err == nil {
			t.Fatal("setClientCertificateFromPEM() error = nil, want unsupported transport")
		}
	})
}

type testTLSMaterial struct {
	caPEM       []byte
	certPEM     []byte
	keyPEM      []byte
	client      *x509.Certificate
	server      *x509.Certificate
	otherServer *x509.Certificate
}

func newTestTLSMaterial(t *testing.T) testTLSMaterial {
	t.Helper()
	notBefore := time.Now().Add(-time.Hour)
	notAfter := time.Now().Add(time.Hour)

	caKey := mustECDSAKey(t)
	ca := mustCertificate(t, &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "test-ca"},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
	}, nil, caKey, &caKey.PublicKey)

	clientKey := mustECDSAKey(t)
	client := mustCertificate(t, &x509.Certificate{
		SerialNumber: big.NewInt(2),
		Subject:      pkix.Name{CommonName: "vault-client"},
		NotBefore:    notBefore,
		NotAfter:     notAfter,
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}, ca, caKey, &clientKey.PublicKey)

	server := mustCertificate(t, &x509.Certificate{
		SerialNumber: big.NewInt(3),
		Subject:      pkix.Name{CommonName: "vault.example.com"},
		DNSNames:     []string{"vault.example.com"},
		NotBefore:    notBefore,
		NotAfter:     notAfter,
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}, ca, caKey, &mustECDSAKey(t).PublicKey)

	otherKey := mustECDSAKey(t)
	otherCA := mustCertificate(t, &x509.Certificate{
		SerialNumber:          big.NewInt(4),
		Subject:               pkix.Name{CommonName: "other-ca"},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature,
		BasicConstraintsValid: true,
	}, nil, otherKey, &otherKey.PublicKey)
	otherServer := mustCertificate(t, &x509.Certificate{
		SerialNumber: big.NewInt(5),
		Subject:      pkix.Name{CommonName: "vault.example.com"},
		DNSNames:     []string{"vault.example.com"},
		NotBefore:    notBefore,
		NotAfter:     notAfter,
		KeyUsage:     x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}, otherCA, otherKey, &mustECDSAKey(t).PublicKey)

	return testTLSMaterial{
		caPEM:       pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: ca.Raw}),
		certPEM:     pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: client.Raw}),
		keyPEM:      pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: mustMarshalECPrivateKey(t, clientKey)}),
		client:      client,
		server:      server,
		otherServer: otherServer,
	}
}

func mustECDSAKey(t *testing.T) *ecdsa.PrivateKey {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	return key
}

func mustMarshalECPrivateKey(t *testing.T, key *ecdsa.PrivateKey) []byte {
	t.Helper()
	der, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		t.Fatalf("marshal key: %v", err)
	}
	return der
}

func mustCertificate(t *testing.T, template, parent *x509.Certificate, signer *ecdsa.PrivateKey, public any) *x509.Certificate {
	t.Helper()
	if parent == nil {
		parent = template
	}
	der, err := x509.CreateCertificate(rand.Reader, template, parent, public, signer)
	if err != nil {
		t.Fatalf("create certificate: %v", err)
	}
	parsed, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatalf("parse certificate: %v", err)
	}
	return parsed
}

func newVaultConfig(t *testing.T) *vault.Config {
	t.Helper()
	for _, key := range []string{
		vault.EnvVaultCACert,
		vault.EnvVaultCACertBytes,
		vault.EnvVaultCAPath,
		vault.EnvVaultClientCert,
		vault.EnvVaultClientKey,
		vault.EnvVaultSkipVerify,
		vault.EnvVaultTLSServerName,
	} {
		t.Setenv(key, "")
	}
	config := vault.DefaultConfig()
	if config.Error != nil {
		t.Fatalf("DefaultConfig() error = %v", config.Error)
	}
	return config
}

func tlsTransport(t *testing.T, config *vault.Config) *http.Transport {
	t.Helper()
	transport, ok := config.HttpClient.Transport.(*http.Transport)
	if !ok || transport.TLSClientConfig == nil {
		t.Fatal("expected *http.Transport with a TLS config")
	}
	return transport
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}
