package v1alpha1

import (
	"context"
	"encoding/json"
	"testing"

	vaultutils "github.com/redhat-cop/vault-config-operator/api/v1alpha1/utils"
	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

func rabbitMQConfig(name string, path vaultutils.Path, namespace string, targetNamespace string) *RabbitMQSecretEngineConfig {
	return &RabbitMQSecretEngineConfig{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: "default"},
		Spec: RabbitMQSecretEngineConfigSpec{
			Path:           path,
			Authentication: vaultutils.KubeAuthConfiguration{Namespace: namespace, TargetNamespace: targetNamespace},
		},
	}
}

func rabbitMQRaw(t *testing.T, config *RabbitMQSecretEngineConfig) runtime.RawExtension {
	t.Helper()
	raw, err := json.Marshal(config)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return runtime.RawExtension{Raw: raw}
}

func TestRabbitMQSecretEngineConfigValidationCreate(t *testing.T) {
	tests := []struct {
		name           string
		vaultNamespace string
		existing       *RabbitMQSecretEngineConfig
		config         *RabbitMQSecretEngineConfig
		wantAllowed    bool
	}{
		{"same path at root", "", rabbitMQConfig("a", "rabbitmq", "", ""), rabbitMQConfig("b", "rabbitmq", "", ""), false},
		{"same path in the same namespace", "", rabbitMQConfig("a", "rabbitmq", "tenant-a", ""), rabbitMQConfig("b", "rabbitmq", "tenant-a", ""), false},
		{"same path in a different namespace", "", rabbitMQConfig("a", "rabbitmq", "tenant-a", ""), rabbitMQConfig("b", "rabbitmq", "tenant-b", ""), true},
		{"different path in the same namespace", "", rabbitMQConfig("a", "rabbitmq", "tenant-a", ""), rabbitMQConfig("b", "rabbitmq-2", "tenant-a", ""), true},
		{"root config and a config in a child namespace", "", rabbitMQConfig("a", "rabbitmq", "", "tenant-a"), rabbitMQConfig("b", "rabbitmq", "", ""), true},
		{"target namespace and namespace resolve to the same namespace", "", rabbitMQConfig("a", "rabbitmq", "org/tenant-a/", ""), rabbitMQConfig("b", "rabbitmq", "org", "tenant-a"), false},
		{"empty namespace resolves to VAULT_NAMESPACE", "admin", rabbitMQConfig("a", "rabbitmq", "admin", ""), rabbitMQConfig("b", "rabbitmq", "", ""), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("VAULT_NAMESPACE", tt.vaultNamespace)
			validator := &RabbitMQSecretEngineConfigValidation{Client: newFakeKubeClient(tt.existing)}
			resp := validator.Handle(context.Background(), admission.Request{AdmissionRequest: admissionv1.AdmissionRequest{
				Operation: admissionv1.Create,
				Object:    rabbitMQRaw(t, tt.config),
			}})
			if resp.Allowed != tt.wantAllowed {
				t.Errorf("allowed = %v, want %v (%v)", resp.Allowed, tt.wantAllowed, resp.Result)
			}
		})
	}
}

func TestRabbitMQSecretEngineConfigValidationUpdate(t *testing.T) {
	tests := []struct {
		name        string
		existing    *RabbitMQSecretEngineConfig
		old         *RabbitMQSecretEngineConfig
		new         *RabbitMQSecretEngineConfig
		wantAllowed bool
	}{
		{"path change", nil, rabbitMQConfig("b", "rabbitmq", "", ""), rabbitMQConfig("b", "rabbitmq-2", "", ""), false},
		{"target namespace change", nil, rabbitMQConfig("b", "rabbitmq", "", "tenant-a"), rabbitMQConfig("b", "rabbitmq", "", "tenant-b"), false},
		{"namespace change to a namespace with the same path", rabbitMQConfig("a", "rabbitmq", "tenant-a", ""), rabbitMQConfig("b", "rabbitmq", "tenant-b", ""), rabbitMQConfig("b", "rabbitmq", "tenant-a", ""), false},
		{"namespace change to a free namespace", rabbitMQConfig("a", "rabbitmq", "tenant-a", ""), rabbitMQConfig("b", "rabbitmq", "tenant-b", ""), rabbitMQConfig("b", "rabbitmq", "tenant-c", ""), true},
		{"namespace change does not conflict with the config itself", nil, rabbitMQConfig("b", "rabbitmq", "tenant-b", ""), rabbitMQConfig("b", "rabbitmq", "tenant-c", ""), true},
		{"no namespace change", rabbitMQConfig("a", "rabbitmq", "tenant-a", ""), rabbitMQConfig("b", "rabbitmq", "tenant-a", ""), rabbitMQConfig("b", "rabbitmq", "tenant-a", ""), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("VAULT_NAMESPACE", "")
			// The list in the webhook contains the stored version of the config itself
			objs := []*RabbitMQSecretEngineConfig{tt.old}
			if tt.existing != nil {
				objs = append(objs, tt.existing)
			}
			kubeClient := newFakeKubeClient()
			for _, obj := range objs {
				if err := kubeClient.Create(context.Background(), obj.DeepCopy()); err != nil {
					t.Fatalf("create: %v", err)
				}
			}
			validator := &RabbitMQSecretEngineConfigValidation{Client: kubeClient}
			resp := validator.Handle(context.Background(), admission.Request{AdmissionRequest: admissionv1.AdmissionRequest{
				Operation: admissionv1.Update,
				Object:    rabbitMQRaw(t, tt.new),
				OldObject: rabbitMQRaw(t, tt.old),
			}})
			if resp.Allowed != tt.wantAllowed {
				t.Errorf("allowed = %v, want %v (%v)", resp.Allowed, tt.wantAllowed, resp.Result)
			}
		})
	}
}
