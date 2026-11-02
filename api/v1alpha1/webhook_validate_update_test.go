package v1alpha1

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	admissionv1 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// Tests for residual Go-level webhook validation that was NOT migrated to VAPs.
// Path/name immutability and target-namespace-guard are now tested via VAPs
// in vap_test.go and targetnamespace_webhook_test.go.

type webhookUpdateTestCase struct {
	name         string
	validateFn   func() (admission.Warnings, error)
	expectErr    bool
	errSubstring string
}

func TestResidualWebhookValidation(t *testing.T) {
	cases := []webhookUpdateTestCase{
		{
			name: "AppRoleAuthEngineRole rejects localSecretIDs change",
			validateFn: func() (admission.Warnings, error) {
				r := &AppRoleAuthEngineRole{}
				return r.ValidateUpdate(context.Background(),
					&AppRoleAuthEngineRole{Spec: AppRoleAuthEngineRoleSpec{
						Path:        "same/path",
						AppRoleRole: AppRoleRole{LocalSecretIDs: false},
					}},
					&AppRoleAuthEngineRole{Spec: AppRoleAuthEngineRoleSpec{
						Path:        "same/path",
						AppRoleRole: AppRoleRole{LocalSecretIDs: true},
					}},
				)
			},
			expectErr:    true,
			errSubstring: "spec.localSecretIDs cannot be updated",
		},
		{
			name: "AppRoleAuthEngineRole allows mutable field update",
			validateFn: func() (admission.Warnings, error) {
				r := &AppRoleAuthEngineRole{}
				return r.ValidateUpdate(context.Background(),
					&AppRoleAuthEngineRole{Spec: AppRoleAuthEngineRoleSpec{
						Path: "same/path",
						Name: "same-name",
						AppRoleRole: AppRoleRole{
							BindSecretID:   true,
							LocalSecretIDs: false,
							TokenPolicies:  []string{"default"},
							TokenTTL:       "1h",
						},
					}},
					&AppRoleAuthEngineRole{Spec: AppRoleAuthEngineRoleSpec{
						Path: "same/path",
						Name: "same-name",
						AppRoleRole: AppRoleRole{
							BindSecretID:   true,
							LocalSecretIDs: false,
							TokenPolicies:  []string{"default", "new-policy"},
							TokenTTL:       "2h",
						},
					}},
				)
			},
			expectErr: false,
		},
		{
			name: "LDAPSecretEngineStaticRole rejects username change",
			validateFn: func() (admission.Warnings, error) {
				r := &LDAPSecretEngineStaticRole{}
				return r.ValidateUpdate(context.Background(),
					&LDAPSecretEngineStaticRole{Spec: LDAPSecretEngineStaticRoleSpec{
						Path:             "same/path",
						LDAPSEStaticRole: LDAPSEStaticRole{Username: "old-user"},
					}},
					&LDAPSecretEngineStaticRole{Spec: LDAPSecretEngineStaticRoleSpec{
						Path:             "same/path",
						LDAPSEStaticRole: LDAPSEStaticRole{Username: "new-user"},
					}},
				)
			},
			expectErr:    true,
			errSubstring: "spec.username cannot be updated",
		},
		{
			name: "LDAPSecretEngineStaticRole rejects dn change",
			validateFn: func() (admission.Warnings, error) {
				r := &LDAPSecretEngineStaticRole{}
				return r.ValidateUpdate(context.Background(),
					&LDAPSecretEngineStaticRole{Spec: LDAPSecretEngineStaticRoleSpec{
						Path:             "same/path",
						LDAPSEStaticRole: LDAPSEStaticRole{DN: "uid=old,dc=example,dc=com"},
					}},
					&LDAPSecretEngineStaticRole{Spec: LDAPSecretEngineStaticRoleSpec{
						Path:             "same/path",
						LDAPSEStaticRole: LDAPSEStaticRole{DN: "uid=new,dc=example,dc=com"},
					}},
				)
			},
			expectErr:    true,
			errSubstring: "spec.dn cannot be updated",
		},
		{
			name: "LDAPSecretEngineStaticRole allows mutable field update",
			validateFn: func() (admission.Warnings, error) {
				r := &LDAPSecretEngineStaticRole{}
				return r.ValidateUpdate(context.Background(),
					&LDAPSecretEngineStaticRole{ObjectMeta: metav1.ObjectMeta{Name: "role"}, Spec: LDAPSecretEngineStaticRoleSpec{
						Path:             "same/path",
						Name:             "same-name",
						LDAPSEStaticRole: LDAPSEStaticRole{Username: "same-user", DN: "uid=same,dc=example,dc=com", RotationPeriod: 86400},
					}},
					&LDAPSecretEngineStaticRole{ObjectMeta: metav1.ObjectMeta{Name: "role"}, Spec: LDAPSecretEngineStaticRoleSpec{
						Path:             "same/path",
						Name:             "same-name",
						LDAPSEStaticRole: LDAPSEStaticRole{Username: "same-user", DN: "uid=same,dc=example,dc=com", RotationPeriod: 43200},
					}},
				)
			},
			expectErr: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := tc.validateFn()

			if tc.expectErr {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tc.errSubstring)
				} else if !strings.Contains(err.Error(), tc.errSubstring) {
					t.Errorf("expected error containing %q, got: %v", tc.errSubstring, err)
				}
			} else {
				if err != nil {
					t.Errorf("expected no error, got: %v", err)
				}
			}
		})
	}
}

func TestRabbitMQSecretEngineConfigHandleAllowsNonPathUpdate(t *testing.T) {
	handler := &RabbitMQSecretEngineConfigValidation{Client: nil}

	t.Run("allows non-path update", func(t *testing.T) {
		newObj := &RabbitMQSecretEngineConfig{
			ObjectMeta: metav1.ObjectMeta{Name: "new"},
			Spec:       RabbitMQSecretEngineConfigSpec{Path: "same/path"},
		}
		oldObj := &RabbitMQSecretEngineConfig{
			ObjectMeta: metav1.ObjectMeta{Name: "old"},
			Spec:       RabbitMQSecretEngineConfigSpec{Path: "same/path"},
		}

		newRaw, err := json.Marshal(newObj)
		if err != nil {
			t.Fatalf("failed to marshal new object: %v", err)
		}
		oldRaw, err := json.Marshal(oldObj)
		if err != nil {
			t.Fatalf("failed to marshal old object: %v", err)
		}

		req := admission.Request{
			AdmissionRequest: admissionv1.AdmissionRequest{
				Operation: admissionv1.Update,
				Object:    runtime.RawExtension{Raw: newRaw},
				OldObject: runtime.RawExtension{Raw: oldRaw},
			},
		}

		resp := handler.Handle(context.Background(), req)
		if !resp.Allowed {
			t.Errorf("expected request to be allowed when spec.path unchanged, got rejected: %v", resp.Result)
		}
	})
}
