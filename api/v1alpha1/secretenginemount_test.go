package v1alpha1

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestSecretEngineMountGetPath(t *testing.T) {
	tests := []struct {
		name         string
		mount        *SecretEngineMount
		expectedPath string
	}{
		{
			name: "with path and name specified",
			mount: &SecretEngineMount{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-mount",
				},
				Spec: SecretEngineMountSpec{
					Path: "custom-path",
					Name: "custom-name",
				},
			},
			expectedPath: "sys/mounts/custom-path/custom-name",
		},
		{
			name: "with path but no name specified",
			mount: &SecretEngineMount{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-mount",
				},
				Spec: SecretEngineMountSpec{
					Path: "custom-path",
				},
			},
			expectedPath: "sys/mounts/custom-path/test-mount",
		},
		{
			name: "with name but no path specified",
			mount: &SecretEngineMount{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-mount",
				},
				Spec: SecretEngineMountSpec{
					Name: "custom-name",
				},
			},
			expectedPath: "sys/mounts/custom-name",
		},
		{
			name: "with neither path nor name specified",
			mount: &SecretEngineMount{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-mount",
				},
				Spec: SecretEngineMountSpec{},
			},
			expectedPath: "sys/mounts/test-mount",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.mount.GetPath()
			if result != tt.expectedPath {
				t.Errorf("GetPath() = %v, expected %v", result, tt.expectedPath)
			}
		})
	}
}
