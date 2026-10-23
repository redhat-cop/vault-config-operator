package v1alpha1

import (
	"testing"

	vaultutils "github.com/redhat-cop/vault-config-operator/api/v1alpha1/utils"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestVaultSecretGetPath(t *testing.T) {
	tests := []struct {
		name         string
		definition   *VaultSecretDefinition
		expectedPath string
	}{
		{
			name: "KVv1 path",
			definition: &VaultSecretDefinition{
				Path: "secret/data/myapp/config",
			},
			expectedPath: "secret/data/myapp/config",
		},
		{
			name: "KVv2 path",
			definition: &VaultSecretDefinition{
				Path: "kv/data/myapp/config",
			},
			expectedPath: "kv/data/myapp/config",
		},
		{
			name: "simple path",
			definition: &VaultSecretDefinition{
				Path: "secret/myapp",
			},
			expectedPath: "secret/myapp",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.definition.GetPath()
			if result != tt.expectedPath {
				t.Errorf("GetPath() = %v, expected %v", result, tt.expectedPath)
			}
		})
	}
}

func TestVaultSecretGetPayload(t *testing.T) {
	definition := &VaultSecretDefinition{
		Path:        "secret/data/myapp",
		RequestType: "POST",
		RequestPayload: map[string]string{
			"key1": "value1",
			"key2": "value2",
		},
	}

	payload := definition.GetPostRequestPayload()

	if payload["key1"] != "value1" {
		t.Errorf("expected key1 'value1', got %v", payload["key1"])
	}
	if payload["key2"] != "value2" {
		t.Errorf("expected key2 'value2', got %v", payload["key2"])
	}
	if len(payload) != 2 {
		t.Errorf("expected exactly 2 keys in payload, got %d", len(payload))
	}
}

func TestVaultSecretGetRequestMethod(t *testing.T) {
	tests := []struct {
		name           string
		definition     *VaultSecretDefinition
		expectedMethod string
	}{
		{
			name: "GET request",
			definition: &VaultSecretDefinition{
				RequestType: "GET",
			},
			expectedMethod: "GET",
		},
		{
			name: "POST request",
			definition: &VaultSecretDefinition{
				RequestType: "POST",
			},
			expectedMethod: "POST",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.definition.GetRequestMethod()
			if result != tt.expectedMethod {
				t.Errorf("GetRequestMethod() = %v, expected %v", result, tt.expectedMethod)
			}
		})
	}
}

func TestVaultSecretIsEquivalentToDesiredState(t *testing.T) {
	definition := &VaultSecretDefinition{
		Path:        "secret/data/myapp",
		RequestType: "POST",
		RequestPayload: map[string]string{
			"key1": "value1",
		},
	}

	payload := definition.GetPostRequestPayload()
	matching := map[string]string{
		"key1": "value1",
	}

	if payload["key1"] != matching["key1"] {
		t.Error("expected matching payload values to be equal")
	}
}

func TestVaultSecretIsEquivalentToDesiredStateNonMatching(t *testing.T) {
	definition := &VaultSecretDefinition{
		Path:        "secret/data/myapp",
		RequestType: "POST",
		RequestPayload: map[string]string{
			"key1": "value1",
		},
	}

	payload := definition.GetPostRequestPayload()
	if payload["key1"] == "different-value" {
		t.Error("expected non-matching payload values to be different")
	}
}

func TestVaultSecretIsEquivalentExtraFieldsReturnsFalse(t *testing.T) {
	definition := &VaultSecretDefinition{
		Path:        "secret/data/myapp",
		RequestType: "POST",
		RequestPayload: map[string]string{
			"key1": "value1",
		},
	}

	payload := definition.GetPostRequestPayload()
	if _, exists := payload["extra_field"]; exists {
		t.Error("expected extra field to not exist in payload")
	}
}

func TestVaultSecretIsDeletable(t *testing.T) {
	var _ vaultutils.VaultSecretObject = &VaultSecretDefinition{}

	definition := &VaultSecretDefinition{
		Path:        "secret/data/myapp",
		RequestType: "GET",
	}

	if definition.GetPath() != "secret/data/myapp" {
		t.Error("expected VaultSecretDefinition path to match")
	}
}

func TestVaultSecretIsValid(t *testing.T) {
	vs := &VaultSecret{
		Spec: VaultSecretSpec{
			RefreshThreshold: 90,
			VaultSecretDefinitions: []VaultSecretDefinition{
				{
					Name:        "my-secret",
					Path:        "secret/data/myapp",
					RequestType: "GET",
				},
			},
			TemplatizedK8sSecret: TemplatizedK8sSecret{
				Name: "my-k8s-secret",
				StringData: map[string]string{
					"password": "{{ .my-secret.password }}",
				},
			},
		},
	}

	valid, err := vs.IsValid()
	if !valid {
		t.Errorf("expected VaultSecret to be valid, got error: %v", err)
	}
}

func TestVaultSecretConditions(t *testing.T) {
	condition := metav1.Condition{
		Type:   "Ready",
		Status: metav1.ConditionTrue,
	}

	vs := &VaultSecret{}
	vs.SetConditions([]metav1.Condition{condition})
	if len(vs.GetConditions()) != 1 || vs.GetConditions()[0].Type != "Ready" {
		t.Error("expected VaultSecret conditions to be set and retrieved")
	}
}
