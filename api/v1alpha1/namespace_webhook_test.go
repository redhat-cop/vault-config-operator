package v1alpha1

import (
	"context"
	"testing"
)

func TestNamespaceWebhookValidateUpdate(t *testing.T) {
	tests := []struct {
		name    string
		oldSpec NamespaceSpec
		newSpec NamespaceSpec
		wantErr bool
	}{
		{"unchanged", NamespaceSpec{Path: "org"}, NamespaceSpec{Path: "org"}, false},
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
