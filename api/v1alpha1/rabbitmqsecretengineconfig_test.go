package v1alpha1

import (
	"reflect"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestRabbitMQSecretEngineConfigGetPath(t *testing.T) {
	config := &RabbitMQSecretEngineConfig{
		Spec: RabbitMQSecretEngineConfigSpec{
			Path: "rabbitmq/engine",
		},
	}

	got := config.GetPath()
	want := "rabbitmq/engine/config/connection"
	if got != want {
		t.Errorf("GetPath() = %q, want %q", got, want)
	}
}

func TestRMQSEConfigRabbitMQToMap(t *testing.T) {
	cfg := RMQSEConfig{
		ConnectionURI:     "https://rmq.example:15672",
		VerifyConnection:  true,
		UsernameTemplate:  "{{.DisplayName}}",
		PasswordPolicy:    "default",
		retrievedUsername: "admin-user",
		retrievedPassword: "secret-pass",
	}

	result := cfg.rabbitMQToMap()

	keys := []string{
		"connection_uri", "verify_connection", "username", "password",
		"username_template", "password_policy",
	}
	if len(result) != len(keys) {
		t.Errorf("rabbitMQToMap() len = %d, want %d keys", len(result), len(keys))
	}
	for _, key := range keys {
		if _, ok := result[key]; !ok {
			t.Errorf("expected key %q in rabbitMQToMap() output", key)
		}
	}

	if result["connection_uri"] != "https://rmq.example:15672" {
		t.Errorf("connection_uri = %v", result["connection_uri"])
	}
	if result["verify_connection"] != true {
		t.Errorf("verify_connection = %v", result["verify_connection"])
	}
	if result["username"] != "admin-user" {
		t.Errorf("username = %v", result["username"])
	}
	if result["password"] != "secret-pass" {
		t.Errorf("password = %v", result["password"])
	}
	if result["username_template"] != "{{.DisplayName}}" {
		t.Errorf("username_template = %v", result["username_template"])
	}
	if result["password_policy"] != "default" {
		t.Errorf("password_policy = %v", result["password_policy"])
	}
}

func TestRMQSEConfigLeasesToMap(t *testing.T) {
	cfg := &RMQSEConfig{
		LeaseTTL:    3600,
		LeaseMaxTTL: 7200,
	}

	result := cfg.leasesToMap()

	if len(result) != 2 {
		t.Fatalf("leasesToMap() len = %d, want 2 keys", len(result))
	}
	ttl, ok := result["ttl"].(int)
	if !ok {
		t.Fatalf("ttl type = %T, want int", result["ttl"])
	}
	if ttl != 3600 {
		t.Errorf("ttl = %v, want 3600", ttl)
	}
	maxTTL, ok := result["max_ttl"].(int)
	if !ok {
		t.Fatalf("max_ttl type = %T, want int", result["max_ttl"])
	}
	if maxTTL != 7200 {
		t.Errorf("max_ttl = %v, want 7200", maxTTL)
	}
}

func TestRabbitMQSecretEngineConfigIsEquivalentMatching(t *testing.T) {
	rabbitMQ := &RabbitMQSecretEngineConfig{
		Spec: RabbitMQSecretEngineConfigSpec{
			Path: "rabbitmq/engine",
			RMQSEConfig: RMQSEConfig{
				LeaseTTL:    3600,
				LeaseMaxTTL: 7200,
			},
		},
	}

	payload := map[string]interface{}{
		"ttl":     3600,
		"max_ttl": 7200,
	}

	if !rabbitMQ.IsEquivalentToDesiredState(payload) {
		t.Error("expected true for lease payload matching leasesToMap()")
	}
}

func TestRabbitMQSecretEngineConfigIsEquivalentNonMatching(t *testing.T) {
	rabbitMQ := &RabbitMQSecretEngineConfig{
		Spec: RabbitMQSecretEngineConfigSpec{
			Path: "rabbitmq/engine",
			RMQSEConfig: RMQSEConfig{
				LeaseTTL:    3600,
				LeaseMaxTTL: 7200,
			},
		},
	}

	payload := map[string]interface{}{
		"ttl":     3600,
		"max_ttl": 9999,
	}

	if rabbitMQ.IsEquivalentToDesiredState(payload) {
		t.Error("expected false when max_ttl differs")
	}
}

func TestRabbitMQSecretEngineConfigIsEquivalentExtraFields(t *testing.T) {
	rabbitMQ := &RabbitMQSecretEngineConfig{
		Spec: RabbitMQSecretEngineConfigSpec{
			Path: "rabbitmq/engine",
			RMQSEConfig: RMQSEConfig{
				LeaseTTL:    3600,
				LeaseMaxTTL: 7200,
			},
		},
	}

	payload := map[string]interface{}{
		"ttl":       3600,
		"max_ttl":   7200,
		"extra_key": "x",
	}

	if rabbitMQ.IsEquivalentToDesiredState(payload) {
		t.Error("expected false when payload has extra keys (bare DeepEqual)")
	}
}

func TestRabbitMQSecretEngineConfigIsEquivalentRejectsConnectionPayload(t *testing.T) {
	rabbitMQ := &RabbitMQSecretEngineConfig{
		Spec: RabbitMQSecretEngineConfigSpec{
			Path: "rabbitmq/engine",
			RMQSEConfig: RMQSEConfig{
				ConnectionURI:     "https://rmq.example:15672",
				VerifyConnection:  true,
				PasswordPolicy:    "default",
				LeaseTTL:          3600,
				LeaseMaxTTL:       7200,
				retrievedUsername: "admin",
				retrievedPassword: "pass",
			},
		},
	}

	payload := rabbitMQ.Spec.rabbitMQToMap()

	if rabbitMQ.IsEquivalentToDesiredState(payload) {
		t.Error("expected false: IsEquivalentToDesiredState uses leasesToMap(), not rabbitMQToMap(); connection payload has different keys")
	}
}

func TestRabbitMQSecretEngineConfigIsDeletable(t *testing.T) {
	rabbitMQ := &RabbitMQSecretEngineConfig{}
	if rabbitMQ.IsDeletable() {
		t.Error("expected RabbitMQSecretEngineConfig not to be deletable")
	}
}

func TestRabbitMQSecretEngineConfigConditions(t *testing.T) {
	rabbitMQ := &RabbitMQSecretEngineConfig{}

	conditions := []metav1.Condition{
		{
			Type:   "ReconcileSuccessful",
			Status: metav1.ConditionTrue,
		},
	}

	rabbitMQ.SetConditions(conditions)
	got := rabbitMQ.GetConditions()

	if len(got) != 1 {
		t.Fatalf("expected 1 condition, got %d", len(got))
	}
	if got[0].Type != "ReconcileSuccessful" {
		t.Errorf("condition type = %v", got[0].Type)
	}
	if got[0].Status != metav1.ConditionTrue {
		t.Errorf("condition status = %v", got[0].Status)
	}
}

func TestRabbitMQSecretEngineConfigGetLeasePayload(t *testing.T) {
	rabbitMQ := &RabbitMQSecretEngineConfig{
		Spec: RabbitMQSecretEngineConfigSpec{
			Path: "rabbitmq/engine",
			RMQSEConfig: RMQSEConfig{
				LeaseTTL:    3600,
				LeaseMaxTTL: 7200,
			},
		},
	}

	got := rabbitMQ.GetLeasePayload()
	want := rabbitMQ.Spec.leasesToMap()

	if !reflect.DeepEqual(got, want) {
		t.Errorf("GetLeasePayload() = %#v, want %#v", got, want)
	}
}

func TestRabbitMQSecretEngineConfigGetLeasePath(t *testing.T) {
	rabbitMQ := &RabbitMQSecretEngineConfig{
		Spec: RabbitMQSecretEngineConfigSpec{
			Path: "rabbitmq/engine",
		},
	}

	got := rabbitMQ.GetLeasePath()
	want := "rabbitmq/engine/config/lease"
	if got != want {
		t.Errorf("GetLeasePath() = %q, want %q", got, want)
	}
}
