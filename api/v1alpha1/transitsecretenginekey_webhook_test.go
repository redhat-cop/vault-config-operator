package v1alpha1

import (
	"context"
	"strings"
	"testing"
)

func TestTransitKeyWebhookValidateUpdate_ImmutableFields(t *testing.T) {
	cases := []struct {
		name         string
		oldSpec      TransitSecretEngineKeySpec
		newSpec      TransitSecretEngineKeySpec
		expectErr    bool
		errSubstring string
	}{
		{
			name:         "rejects spec.path change",
			oldSpec:      TransitSecretEngineKeySpec{Path: "transit", TransitKeyConfig: TransitKeyConfig{Type: "aes256-gcm96"}},
			newSpec:      TransitSecretEngineKeySpec{Path: "other-transit", TransitKeyConfig: TransitKeyConfig{Type: "aes256-gcm96"}},
			expectErr:    true,
			errSubstring: "spec.path cannot be updated",
		},
		{
			name:         "rejects spec.name change",
			oldSpec:      TransitSecretEngineKeySpec{Path: "transit", Name: "key-a", TransitKeyConfig: TransitKeyConfig{Type: "aes256-gcm96"}},
			newSpec:      TransitSecretEngineKeySpec{Path: "transit", Name: "key-b", TransitKeyConfig: TransitKeyConfig{Type: "aes256-gcm96"}},
			expectErr:    true,
			errSubstring: "spec.name cannot be updated",
		},
		{
			name:         "rejects spec.type change",
			oldSpec:      TransitSecretEngineKeySpec{Path: "transit", TransitKeyConfig: TransitKeyConfig{Type: "aes256-gcm96"}},
			newSpec:      TransitSecretEngineKeySpec{Path: "transit", TransitKeyConfig: TransitKeyConfig{Type: "ed25519"}},
			expectErr:    true,
			errSubstring: "spec.type cannot be updated after key creation",
		},
		{
			name:         "rejects spec.derived change",
			oldSpec:      TransitSecretEngineKeySpec{Path: "transit", TransitKeyConfig: TransitKeyConfig{Type: "aes256-gcm96", Derived: false}},
			newSpec:      TransitSecretEngineKeySpec{Path: "transit", TransitKeyConfig: TransitKeyConfig{Type: "aes256-gcm96", Derived: true}},
			expectErr:    true,
			errSubstring: "spec.derived cannot be updated after key creation",
		},
		{
			name:         "rejects spec.convergentEncryption change",
			oldSpec:      TransitSecretEngineKeySpec{Path: "transit", TransitKeyConfig: TransitKeyConfig{Type: "aes256-gcm96", Derived: true, ConvergentEncryption: false}},
			newSpec:      TransitSecretEngineKeySpec{Path: "transit", TransitKeyConfig: TransitKeyConfig{Type: "aes256-gcm96", Derived: true, ConvergentEncryption: true}},
			expectErr:    true,
			errSubstring: "spec.convergentEncryption cannot be updated after key creation",
		},
		{
			name:         "rejects spec.keySize change",
			oldSpec:      TransitSecretEngineKeySpec{Path: "transit", TransitKeyConfig: TransitKeyConfig{Type: "hmac", KeySize: 32}},
			newSpec:      TransitSecretEngineKeySpec{Path: "transit", TransitKeyConfig: TransitKeyConfig{Type: "hmac", KeySize: 64}},
			expectErr:    true,
			errSubstring: "spec.keySize cannot be updated after key creation",
		},
		{
			name:    "allows mutable config field update",
			oldSpec: TransitSecretEngineKeySpec{Path: "transit", TransitKeyConfig: TransitKeyConfig{Type: "aes256-gcm96", MinDecryptionVersion: 1}},
			newSpec: TransitSecretEngineKeySpec{Path: "transit", TransitKeyConfig: TransitKeyConfig{Type: "aes256-gcm96", MinDecryptionVersion: 3}},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := &TransitSecretEngineKey{}
			_, err := r.ValidateUpdate(context.Background(),
				&TransitSecretEngineKey{Spec: tc.oldSpec},
				&TransitSecretEngineKey{Spec: tc.newSpec},
			)
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

func TestTransitKeyWebhookValidateUpdate_OneWayFlags(t *testing.T) {
	cases := []struct {
		name         string
		oldSpec      TransitSecretEngineKeySpec
		newSpec      TransitSecretEngineKeySpec
		expectErr    bool
		errSubstring string
	}{
		{
			name:         "rejects exportable true->false",
			oldSpec:      TransitSecretEngineKeySpec{Path: "transit", TransitKeyConfig: TransitKeyConfig{Type: "aes256-gcm96", Exportable: true}},
			newSpec:      TransitSecretEngineKeySpec{Path: "transit", TransitKeyConfig: TransitKeyConfig{Type: "aes256-gcm96", Exportable: false}},
			expectErr:    true,
			errSubstring: "spec.exportable cannot be changed from true to false",
		},
		{
			name:    "allows exportable false->true",
			oldSpec: TransitSecretEngineKeySpec{Path: "transit", TransitKeyConfig: TransitKeyConfig{Type: "aes256-gcm96", Exportable: false}},
			newSpec: TransitSecretEngineKeySpec{Path: "transit", TransitKeyConfig: TransitKeyConfig{Type: "aes256-gcm96", Exportable: true}},
		},
		{
			name:    "allows exportable true->true (no change)",
			oldSpec: TransitSecretEngineKeySpec{Path: "transit", TransitKeyConfig: TransitKeyConfig{Type: "aes256-gcm96", Exportable: true}},
			newSpec: TransitSecretEngineKeySpec{Path: "transit", TransitKeyConfig: TransitKeyConfig{Type: "aes256-gcm96", Exportable: true}},
		},
		{
			name:         "rejects allowPlaintextBackup true->false",
			oldSpec:      TransitSecretEngineKeySpec{Path: "transit", TransitKeyConfig: TransitKeyConfig{Type: "aes256-gcm96", AllowPlaintextBackup: true}},
			newSpec:      TransitSecretEngineKeySpec{Path: "transit", TransitKeyConfig: TransitKeyConfig{Type: "aes256-gcm96", AllowPlaintextBackup: false}},
			expectErr:    true,
			errSubstring: "spec.allowPlaintextBackup cannot be changed from true to false",
		},
		{
			name:    "allows allowPlaintextBackup false->true",
			oldSpec: TransitSecretEngineKeySpec{Path: "transit", TransitKeyConfig: TransitKeyConfig{Type: "aes256-gcm96", AllowPlaintextBackup: false}},
			newSpec: TransitSecretEngineKeySpec{Path: "transit", TransitKeyConfig: TransitKeyConfig{Type: "aes256-gcm96", AllowPlaintextBackup: true}},
		},
		{
			name:    "allows allowPlaintextBackup true->true (no change)",
			oldSpec: TransitSecretEngineKeySpec{Path: "transit", TransitKeyConfig: TransitKeyConfig{Type: "aes256-gcm96", AllowPlaintextBackup: true}},
			newSpec: TransitSecretEngineKeySpec{Path: "transit", TransitKeyConfig: TransitKeyConfig{Type: "aes256-gcm96", AllowPlaintextBackup: true}},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := &TransitSecretEngineKey{}
			_, err := r.ValidateUpdate(context.Background(),
				&TransitSecretEngineKey{Spec: tc.oldSpec},
				&TransitSecretEngineKey{Spec: tc.newSpec},
			)
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

func TestTransitKeyWebhookValidateCreate_ConvergentEncryptionInvariant(t *testing.T) {
	cases := []struct {
		name         string
		spec         TransitSecretEngineKeySpec
		expectErr    bool
		errSubstring string
	}{
		{
			name:         "rejects convergentEncryption without derived",
			spec:         TransitSecretEngineKeySpec{Path: "transit", TransitKeyConfig: TransitKeyConfig{Type: "aes256-gcm96", ConvergentEncryption: true, Derived: false}},
			expectErr:    true,
			errSubstring: "spec.convergentEncryption requires spec.derived to be true",
		},
		{
			name: "allows convergentEncryption with derived",
			spec: TransitSecretEngineKeySpec{Path: "transit", TransitKeyConfig: TransitKeyConfig{Type: "aes256-gcm96", ConvergentEncryption: true, Derived: true}},
		},
		{
			name: "allows no convergentEncryption without derived",
			spec: TransitSecretEngineKeySpec{Path: "transit", TransitKeyConfig: TransitKeyConfig{Type: "aes256-gcm96", ConvergentEncryption: false, Derived: false}},
		},
		{
			name: "allows derived without convergentEncryption",
			spec: TransitSecretEngineKeySpec{Path: "transit", TransitKeyConfig: TransitKeyConfig{Type: "aes256-gcm96", ConvergentEncryption: false, Derived: true}},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := &TransitSecretEngineKey{}
			_, err := r.ValidateCreate(context.Background(),
				&TransitSecretEngineKey{Spec: tc.spec},
			)
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

func TestTransitKeyWebhookValidateCreate_KeySizeConstraints(t *testing.T) {
	cases := []struct {
		name         string
		spec         TransitSecretEngineKeySpec
		expectErr    bool
		errSubstring string
	}{
		{
			name:         "rejects keySize on non-hmac type",
			spec:         TransitSecretEngineKeySpec{Path: "transit", TransitKeyConfig: TransitKeyConfig{Type: "aes256-gcm96", KeySize: 32}},
			expectErr:    true,
			errSubstring: "spec.keySize is only applicable to hmac key type",
		},
		{
			name:         "rejects keySize < 32 on hmac type",
			spec:         TransitSecretEngineKeySpec{Path: "transit", TransitKeyConfig: TransitKeyConfig{Type: "hmac", KeySize: 16}},
			expectErr:    true,
			errSubstring: "spec.keySize must be at least 32 bytes",
		},
		{
			name: "allows keySize=0 on non-hmac type",
			spec: TransitSecretEngineKeySpec{Path: "transit", TransitKeyConfig: TransitKeyConfig{Type: "aes256-gcm96", KeySize: 0}},
		},
		{
			name: "allows keySize=32 on hmac type",
			spec: TransitSecretEngineKeySpec{Path: "transit", TransitKeyConfig: TransitKeyConfig{Type: "hmac", KeySize: 32}},
		},
		{
			name: "allows keySize=512 on hmac type",
			spec: TransitSecretEngineKeySpec{Path: "transit", TransitKeyConfig: TransitKeyConfig{Type: "hmac", KeySize: 512}},
		},
		{
			name: "allows keySize=0 on hmac type (uses Vault default)",
			spec: TransitSecretEngineKeySpec{Path: "transit", TransitKeyConfig: TransitKeyConfig{Type: "hmac", KeySize: 0}},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := &TransitSecretEngineKey{}
			_, err := r.ValidateCreate(context.Background(),
				&TransitSecretEngineKey{Spec: tc.spec},
			)
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

func TestTransitKeyWebhookValidateCreate_AutoRotatePeriodConstraints(t *testing.T) {
	cases := []struct {
		name         string
		spec         TransitSecretEngineKeySpec
		expectErr    bool
		errSubstring string
	}{
		{
			name:         "rejects autoRotatePeriod less than 1h",
			spec:         TransitSecretEngineKeySpec{Path: "transit", TransitKeyConfig: TransitKeyConfig{Type: "aes256-gcm96", AutoRotatePeriod: "30m"}},
			expectErr:    true,
			errSubstring: "spec.autoRotatePeriod must be at least",
		},
		{
			name:         "rejects invalid autoRotatePeriod format",
			spec:         TransitSecretEngineKeySpec{Path: "transit", TransitKeyConfig: TransitKeyConfig{Type: "aes256-gcm96", AutoRotatePeriod: "notaduration"}},
			expectErr:    true,
			errSubstring: "spec.autoRotatePeriod is not a valid duration",
		},
		{
			name: "allows autoRotatePeriod of exactly 1h",
			spec: TransitSecretEngineKeySpec{Path: "transit", TransitKeyConfig: TransitKeyConfig{Type: "aes256-gcm96", AutoRotatePeriod: "1h"}},
		},
		{
			name: "allows autoRotatePeriod of 24h",
			spec: TransitSecretEngineKeySpec{Path: "transit", TransitKeyConfig: TransitKeyConfig{Type: "aes256-gcm96", AutoRotatePeriod: "24h"}},
		},
		{
			name: "allows autoRotatePeriod of 3600s",
			spec: TransitSecretEngineKeySpec{Path: "transit", TransitKeyConfig: TransitKeyConfig{Type: "aes256-gcm96", AutoRotatePeriod: "3600s"}},
		},
		{
			name: "allows autoRotatePeriod of 0 (disabled)",
			spec: TransitSecretEngineKeySpec{Path: "transit", TransitKeyConfig: TransitKeyConfig{Type: "aes256-gcm96", AutoRotatePeriod: "0"}},
		},
		{
			name: "allows empty autoRotatePeriod (disabled)",
			spec: TransitSecretEngineKeySpec{Path: "transit", TransitKeyConfig: TransitKeyConfig{Type: "aes256-gcm96", AutoRotatePeriod: ""}},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r := &TransitSecretEngineKey{}
			_, err := r.ValidateCreate(context.Background(),
				&TransitSecretEngineKey{Spec: tc.spec},
			)
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
