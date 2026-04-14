package v1alpha1

import (
	"testing"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestRandomSecretGetPath(t *testing.T) {
	tests := []struct {
		name         string
		rs           *RandomSecret
		expectedPath string
	}{
		{
			name: "with spec.name specified",
			rs: &RandomSecret{
				ObjectMeta: metav1.ObjectMeta{Name: "meta-name"},
				Spec: RandomSecretSpec{
					Path: "secret/data/myapp",
					Name: "spec-name",
				},
			},
			expectedPath: "secret/data/myapp/spec-name",
		},
		{
			name: "without spec.name falls back to metadata.name",
			rs: &RandomSecret{
				ObjectMeta: metav1.ObjectMeta{Name: "meta-name"},
				Spec: RandomSecretSpec{
					Path: "secret/data/myapp",
				},
			},
			expectedPath: "secret/data/myapp/meta-name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.rs.GetPath()
			if result != tt.expectedPath {
				t.Errorf("GetPath() = %v, expected %v", result, tt.expectedPath)
			}
		})
	}
}

func TestRandomSecretGetV1Payload(t *testing.T) {
	rs := &RandomSecret{
		Spec: RandomSecretSpec{
			SecretKey:        "password",
			calculatedSecret: "s3cr3t-value",
		},
	}

	payload := rs.getV1Payload()

	if payload["password"] != "s3cr3t-value" {
		t.Errorf("expected dynamic key 'password' = 's3cr3t-value', got %v", payload["password"])
	}
	if len(payload) != 1 {
		t.Errorf("expected 1 key in payload, got %d", len(payload))
	}
}

func TestRandomSecretGetV1PayloadWithRefreshPeriod(t *testing.T) {
	dur := 30 * time.Minute
	rs := &RandomSecret{
		Spec: RandomSecretSpec{
			SecretKey:        "password",
			calculatedSecret: "s3cr3t-value",
			RefreshPeriod:    &metav1.Duration{Duration: dur},
		},
	}

	payload := rs.getV1Payload()

	if payload["password"] != "s3cr3t-value" {
		t.Errorf("expected dynamic key 'password' = 's3cr3t-value', got %v", payload["password"])
	}
	if payload[ttlKey] != dur.String() {
		t.Errorf("expected ttl = %q, got %v", dur.String(), payload[ttlKey])
	}
	if len(payload) != 2 {
		t.Errorf("expected 2 keys in payload, got %d", len(payload))
	}
}

func TestRandomSecretGetV1PayloadNoRefreshPeriod(t *testing.T) {
	rs := &RandomSecret{
		Spec: RandomSecretSpec{
			SecretKey:        "password",
			calculatedSecret: "s3cr3t-value",
		},
	}

	payload := rs.getV1Payload()

	if _, ok := payload[ttlKey]; ok {
		t.Error("expected no ttl key when RefreshPeriod is nil")
	}
}

func TestRandomSecretGetV1PayloadZeroRefreshPeriod(t *testing.T) {
	rs := &RandomSecret{
		Spec: RandomSecretSpec{
			SecretKey:        "password",
			calculatedSecret: "s3cr3t-value",
			RefreshPeriod:    &metav1.Duration{Duration: 0},
		},
	}

	payload := rs.getV1Payload()

	if _, ok := payload[ttlKey]; ok {
		t.Error("expected no ttl key when RefreshPeriod duration is zero")
	}
}

func TestRandomSecretGetPayloadKVv2(t *testing.T) {
	rs := &RandomSecret{
		Spec: RandomSecretSpec{
			SecretKey:           "password",
			calculatedSecret:    "s3cr3t-value",
			IsKVSecretsEngineV2: true,
			Path:                "secret/data/myapp",
		},
	}

	payload := rs.GetPayload()

	data, ok := payload["data"].(map[string]interface{})
	if !ok {
		t.Fatal("expected KV v2 payload to have 'data' wrapper key")
	}
	if data["password"] != "s3cr3t-value" {
		t.Errorf("expected inner payload key 'password' = 's3cr3t-value', got %v", data["password"])
	}
	if len(payload) != 1 {
		t.Errorf("expected 1 top-level key ('data'), got %d", len(payload))
	}
}

func TestRandomSecretGetPayloadKVv1(t *testing.T) {
	rs := &RandomSecret{
		Spec: RandomSecretSpec{
			SecretKey:           "password",
			calculatedSecret:    "s3cr3t-value",
			IsKVSecretsEngineV2: false,
			Path:                "secret/myapp",
		},
	}

	payload := rs.GetPayload()

	if _, ok := payload["data"]; ok {
		t.Error("expected KV v1 payload to NOT have 'data' wrapper key")
	}
	if payload["password"] != "s3cr3t-value" {
		t.Errorf("expected key 'password' = 's3cr3t-value', got %v", payload["password"])
	}
}

func TestRandomSecretIsEquivalentAlwaysFalse(t *testing.T) {
	rs := &RandomSecret{
		Spec: RandomSecretSpec{
			SecretKey:        "password",
			calculatedSecret: "s3cr3t-value",
		},
	}

	if rs.IsEquivalentToDesiredState(map[string]interface{}{"password": "s3cr3t-value"}) {
		t.Error("expected IsEquivalentToDesiredState to always return false for RandomSecret")
	}

	if rs.IsEquivalentToDesiredState(map[string]interface{}{}) {
		t.Error("expected IsEquivalentToDesiredState to return false even for empty payload")
	}

	if rs.IsEquivalentToDesiredState(nil) {
		t.Error("expected IsEquivalentToDesiredState to return false even for nil payload")
	}
}

func TestRandomSecretIsDeletable(t *testing.T) {
	rs := &RandomSecret{}
	if !rs.IsDeletable() {
		t.Error("expected RandomSecret to be deletable")
	}
}

func TestRandomSecretConditions(t *testing.T) {
	condition := metav1.Condition{
		Type:   "Ready",
		Status: metav1.ConditionTrue,
	}

	rs := &RandomSecret{}
	rs.SetConditions([]metav1.Condition{condition})
	got := rs.GetConditions()
	if len(got) != 1 || got[0].Type != "Ready" {
		t.Error("expected RandomSecret conditions to be set and retrieved")
	}
}
