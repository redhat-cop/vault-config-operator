package v1alpha1

import (
	"context"
	"testing"

	vaultutils "github.com/redhat-cop/vault-config-operator/api/v1alpha1/utils"
)

func TestNamespaceWebhookValidateUpdate(t *testing.T) {
	tests := []struct {
		name    string
		oldSpec NamespaceSpec
		newSpec NamespaceSpec
		wantErr bool
	}{
		{"unchanged", NamespaceSpec{Path: "org"}, NamespaceSpec{Path: "org"}, false},
		{"path changed", NamespaceSpec{Path: "org"}, NamespaceSpec{Path: "org2"}, true},
		{"path removed", NamespaceSpec{Path: "org"}, NamespaceSpec{}, true},
		{"path added", NamespaceSpec{}, NamespaceSpec{Path: "org"}, true},
		{"target namespace changed", NamespaceSpec{Authentication: vaultutils.KubeAuthConfiguration{TargetNamespace: "tenant-a"}}, NamespaceSpec{Authentication: vaultutils.KubeAuthConfiguration{TargetNamespace: "tenant-b"}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("VAULT_NAMESPACE", "")
			_, err := (&Namespace{}).ValidateUpdate(context.Background(), &Namespace{Spec: tt.oldSpec}, &Namespace{Spec: tt.newSpec})
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateUpdate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
