package v1alpha1

import (
	"testing"

	vaultutils "github.com/redhat-cop/vault-config-operator/api/v1alpha1/utils"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestAWSSecretEngineConfigGetPath(t *testing.T) {
	config := &AWSSecretEngineConfig{
		Spec: AWSSecretEngineConfigSpec{
			Path: "aws",
		},
	}
	expected := "aws/config/root"
	if got := config.GetPath(); got != expected {
		t.Errorf("GetPath() = %q, expected %q", got, expected)
	}
}

func TestAWSSecretEngineConfigGetPathWithDifferentMount(t *testing.T) {
	config := &AWSSecretEngineConfig{
		Spec: AWSSecretEngineConfigSpec{
			Path: "my-aws-engine",
		},
	}
	expected := "my-aws-engine/config/root"
	if got := config.GetPath(); got != expected {
		t.Errorf("GetPath() = %q, expected %q", got, expected)
	}
}

func TestAWSSecretEngineConfigIsDeletable(t *testing.T) {
	config := &AWSSecretEngineConfig{}
	if config.IsDeletable() {
		t.Error("expected AWSSecretEngineConfig not to be deletable")
	}
}

func TestAWSRootConfig_toMap(t *testing.T) {
	maxRetries := 3
	config := AWSRootConfig{
		Region:             "us-west-2",
		IAMEndpoint:        "https://iam.amazonaws.com",
		STSEndpoint:        "https://sts.us-west-2.amazonaws.com",
		MaxRetries:         &maxRetries,
		UsernameTemplate:   "{{ .DisplayName }}",
		retrievedAccessKey: "AKIAEXAMPLE",
		retrievedSecretKey: "wJalrXUtnFEMI/K7MDENG/bPxRfiCY",
	}

	result := config.toMap()

	expectedKeys := []string{"access_key", "secret_key", "region", "iam_endpoint", "sts_endpoint", "max_retries", "username_template"}
	if len(result) != len(expectedKeys) {
		t.Fatalf("expected %d keys in toMap() output, got %d: %v", len(expectedKeys), len(result), result)
	}
	for _, key := range expectedKeys {
		if _, ok := result[key]; !ok {
			t.Errorf("expected key %q in toMap() output", key)
		}
	}

	if ak, ok := result["access_key"].(string); !ok || ak != "AKIAEXAMPLE" {
		t.Errorf("access_key = %v (%T), expected string AKIAEXAMPLE", result["access_key"], result["access_key"])
	}
	if sk, ok := result["secret_key"].(string); !ok || sk != "wJalrXUtnFEMI/K7MDENG/bPxRfiCY" {
		t.Errorf("secret_key = %v (%T), expected secret key string", result["secret_key"], result["secret_key"])
	}
	if r, ok := result["region"].(string); !ok || r != "us-west-2" {
		t.Errorf("region = %v (%T), expected string us-west-2", result["region"], result["region"])
	}
	if mr, ok := result["max_retries"].(int); !ok || mr != 3 {
		t.Errorf("max_retries = %v (%T), expected int 3", result["max_retries"], result["max_retries"])
	}
}

func TestAWSRootConfig_toMap_MinimalFields(t *testing.T) {
	config := AWSRootConfig{
		retrievedAccessKey: "AKIAEXAMPLE",
		retrievedSecretKey: "secretkey",
	}

	result := config.toMap()

	requiredKeys := []string{"access_key", "secret_key", "region", "iam_endpoint", "sts_endpoint", "username_template", "max_retries"}
	if len(result) != len(requiredKeys) {
		t.Fatalf("expected %d keys for minimal config (all managed fields included), got %d: %v", len(requiredKeys), len(result), result)
	}
	if _, ok := result["access_key"]; !ok {
		t.Error("expected key 'access_key' in minimal toMap() output")
	}
	if _, ok := result["secret_key"]; !ok {
		t.Error("expected key 'secret_key' in minimal toMap() output")
	}
	if result["region"] != "" {
		t.Errorf("region should be empty string, got %v", result["region"])
	}
	if mr, ok := result["max_retries"].(int); !ok || mr != -1 {
		t.Errorf("max_retries should be -1 when nil, got %v", result["max_retries"])
	}
}

func TestAWSRootConfig_toMap_AccessKeyFromSpec(t *testing.T) {
	config := AWSRootConfig{
		AccessKey: "AKIAFROMSPEC",
	}

	result := config.toMap()

	if ak, ok := result["access_key"].(string); !ok || ak != "AKIAFROMSPEC" {
		t.Errorf("access_key = %v, expected AKIAFROMSPEC", result["access_key"])
	}
}

func TestAWSRootConfig_toMap_RetrievedAccessKeyOverridesSpec(t *testing.T) {
	config := AWSRootConfig{
		AccessKey:          "AKIAFROMSPEC",
		retrievedAccessKey: "AKIARETRIEVED",
		retrievedSecretKey: "secretkey",
	}

	result := config.toMap()

	if ak, ok := result["access_key"].(string); !ok || ak != "AKIARETRIEVED" {
		t.Errorf("access_key = %v, expected AKIARETRIEVED (retrieved should override spec)", result["access_key"])
	}
}

func TestAWSSecretEngineConfig_IsEquivalentToDesiredState_Match(t *testing.T) {
	maxRetries := -1
	config := &AWSSecretEngineConfig{
		Spec: AWSSecretEngineConfigSpec{
			Path: "aws",
			AWSRootConfig: AWSRootConfig{
				AccessKey:   "AKIAEXAMPLE",
				Region:      "us-west-2",
				IAMEndpoint: "https://iam.amazonaws.com",
				STSEndpoint: "https://sts.us-west-2.amazonaws.com",
				MaxRetries:  &maxRetries,
			},
		},
	}

	vaultPayload := map[string]any{
		"access_key":   "AKIAEXAMPLE",
		"region":       "us-west-2",
		"iam_endpoint": "https://iam.amazonaws.com",
		"sts_endpoint": "https://sts.us-west-2.amazonaws.com",
		"max_retries":  -1,
	}

	if !config.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected true when managed fields match (no credentials resolved)")
	}
}

func TestAWSSecretEngineConfig_IsEquivalentToDesiredState_Mismatch(t *testing.T) {
	config := &AWSSecretEngineConfig{
		Spec: AWSSecretEngineConfigSpec{
			Path: "aws",
			AWSRootConfig: AWSRootConfig{
				AccessKey: "AKIAEXAMPLE",
				Region:    "us-west-2",
			},
		},
	}

	vaultPayload := map[string]any{
		"access_key": "AKIAEXAMPLE",
		"region":     "eu-west-1",
	}

	if config.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected false when region differs")
	}
}

func TestAWSSecretEngineConfig_IsEquivalentToDesiredState_ExtraVaultFields(t *testing.T) {
	config := &AWSSecretEngineConfig{
		Spec: AWSSecretEngineConfigSpec{
			Path: "aws",
			AWSRootConfig: AWSRootConfig{
				AccessKey: "AKIAEXAMPLE",
				Region:    "us-west-2",
			},
		},
	}

	vaultPayload := map[string]any{
		"access_key":    "AKIAEXAMPLE",
		"region":        "us-west-2",
		"extra_field":   "from-vault",
		"another_field": 123,
	}

	if !config.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected true: extra keys not in desiredState are filtered from payload")
	}
}

func TestAWSSecretEngineConfig_IsEquivalentToDesiredState_SecretKeyInPayloadIgnored(t *testing.T) {
	config := &AWSSecretEngineConfig{
		Spec: AWSSecretEngineConfigSpec{
			Path: "aws",
			AWSRootConfig: AWSRootConfig{
				AccessKey: "AKIAEXAMPLE",
				Region:    "us-west-2",
			},
		},
	}

	vaultPayload := map[string]any{
		"access_key": "AKIAEXAMPLE",
		"region":     "us-west-2",
		"secret_key": "some-unexpected-key",
	}

	if !config.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected true: secret_key in Vault payload is filtered as extra field when no credentials are resolved")
	}
}

func TestAWSSecretEngineConfig_IsEquivalentToDesiredState_AccessKeyRotationDetected(t *testing.T) {
	config := &AWSSecretEngineConfig{
		Spec: AWSSecretEngineConfigSpec{
			Path: "aws",
			AWSRootConfig: AWSRootConfig{
				AccessKey: "AKIANEWKEY",
				Region:    "us-west-2",
			},
		},
	}

	vaultPayload := map[string]any{
		"access_key": "AKIAOLDKEY",
		"region":     "us-west-2",
	}

	if config.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected false: access_key rotation must be detected (AKIANEWKEY vs AKIAOLDKEY)")
	}
}

func TestAWSSecretEngineConfig_IsEquivalentToDesiredState_ClearedFieldDetected(t *testing.T) {
	config := &AWSSecretEngineConfig{
		Spec: AWSSecretEngineConfigSpec{
			Path: "aws",
			AWSRootConfig: AWSRootConfig{
				AccessKey: "AKIAEXAMPLE",
			},
		},
	}

	vaultPayload := map[string]any{
		"access_key":   "AKIAEXAMPLE",
		"region":       "us-west-2",
		"iam_endpoint": "https://iam.amazonaws.com",
	}

	if config.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected false: region and iam_endpoint were cleared from spec but still present in Vault")
	}
}

func TestAWSSecretEngineConfig_IsEquivalentToDesiredState_UnsetFieldsIgnored(t *testing.T) {
	config := &AWSSecretEngineConfig{
		Spec: AWSSecretEngineConfigSpec{
			Path: "aws",
			AWSRootConfig: AWSRootConfig{
				AccessKey: "AKIAEXAMPLE",
				Region:    "us-west-2",
			},
		},
	}

	vaultPayload := map[string]any{
		"access_key": "AKIAEXAMPLE",
		"region":     "us-west-2",
	}

	if !config.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected true: fields never set by user and absent from Vault should not cause false drift")
	}
}

func TestAWSSecretEngineConfigConditions(t *testing.T) {
	config := &AWSSecretEngineConfig{}

	conditions := []metav1.Condition{
		{
			Type:   "ReconcileSuccessful",
			Status: metav1.ConditionTrue,
		},
	}

	config.SetConditions(conditions)
	got := config.GetConditions()

	if len(got) != 1 {
		t.Fatalf("expected 1 condition, got %d", len(got))
	}
	if got[0].Type != "ReconcileSuccessful" {
		t.Errorf("expected condition type 'ReconcileSuccessful', got %v", got[0].Type)
	}
	if got[0].Status != metav1.ConditionTrue {
		t.Errorf("expected condition status True, got %v", got[0].Status)
	}
}

func TestAWSSecretEngineConfig_PrepareInternalValues_FromK8sSecret(t *testing.T) {
	ns := "ns-aws"
	sec := newK8sSecret(ns, "aws-creds", map[string][]byte{
		"username": []byte("AKIAEXAMPLE"),
		"password": []byte("wJalrXUtnFEMI"),
	})
	kube := newFakeKubeClient(sec)
	hc := newFakeVaultHandler()
	vc, ts := newFakeVaultClient(t, hc)
	defer ts.Close()
	ctx := pivContext(kube, vc)
	config := &AWSSecretEngineConfig{
		ObjectMeta: metav1.ObjectMeta{Namespace: ns},
		Spec: AWSSecretEngineConfigSpec{
			Path: "aws",
			RootCredentials: vaultutils.RootCredentialConfig{
				Secret:      &corev1.LocalObjectReference{Name: "aws-creds"},
				UsernameKey: "username",
				PasswordKey: "password",
			},
		},
	}
	if err := config.PrepareInternalValues(ctx, config); err != nil {
		t.Fatalf("PrepareInternalValues: %v", err)
	}
	if config.Spec.retrievedAccessKey != "AKIAEXAMPLE" {
		t.Errorf("retrievedAccessKey = %q, want %q", config.Spec.retrievedAccessKey, "AKIAEXAMPLE")
	}
	if config.Spec.retrievedSecretKey != "wJalrXUtnFEMI" {
		t.Errorf("retrievedSecretKey = %q, want %q", config.Spec.retrievedSecretKey, "wJalrXUtnFEMI")
	}
}

func TestAWSSecretEngineConfig_PrepareInternalValues_FromVaultSecret(t *testing.T) {
	ns := "ns-aws"
	vaultPath := "secret/data/aws-creds"
	handler := newFakeVaultHandler()
	handler.setGet(vaultPath, map[string]any{
		"username": "AKIAFROMVAULT",
		"password": "secret-from-vault",
	})
	kube := newFakeKubeClient()
	vc, ts := newFakeVaultClient(t, handler)
	defer ts.Close()
	ctx := pivContext(kube, vc)
	config := &AWSSecretEngineConfig{
		ObjectMeta: metav1.ObjectMeta{Namespace: ns},
		Spec: AWSSecretEngineConfigSpec{
			Path: "aws",
			RootCredentials: vaultutils.RootCredentialConfig{
				VaultSecret: &vaultutils.VaultSecretReference{Path: vaultPath},
				UsernameKey: "username",
				PasswordKey: "password",
			},
		},
	}
	if err := config.PrepareInternalValues(ctx, config); err != nil {
		t.Fatalf("PrepareInternalValues: %v", err)
	}
	if config.Spec.retrievedAccessKey != "AKIAFROMVAULT" {
		t.Errorf("retrievedAccessKey = %q, want %q", config.Spec.retrievedAccessKey, "AKIAFROMVAULT")
	}
	if config.Spec.retrievedSecretKey != "secret-from-vault" {
		t.Errorf("retrievedSecretKey = %q, want %q", config.Spec.retrievedSecretKey, "secret-from-vault")
	}
}

func TestAWSSecretEngineConfig_PrepareInternalValues_SpecAccessKeyOverride(t *testing.T) {
	ns := "ns-aws"
	sec := newK8sSecret(ns, "aws-creds", map[string][]byte{
		"username": []byte("AKIAFROMSECRET"),
		"password": []byte("secretkey"),
	})
	kube := newFakeKubeClient(sec)
	hc := newFakeVaultHandler()
	vc, ts := newFakeVaultClient(t, hc)
	defer ts.Close()
	ctx := pivContext(kube, vc)
	config := &AWSSecretEngineConfig{
		ObjectMeta: metav1.ObjectMeta{Namespace: ns},
		Spec: AWSSecretEngineConfigSpec{
			Path: "aws",
			AWSRootConfig: AWSRootConfig{
				AccessKey: "AKIAFROMSPEC",
			},
			RootCredentials: vaultutils.RootCredentialConfig{
				Secret:      &corev1.LocalObjectReference{Name: "aws-creds"},
				UsernameKey: "username",
				PasswordKey: "password",
			},
		},
	}
	if err := config.PrepareInternalValues(ctx, config); err != nil {
		t.Fatalf("PrepareInternalValues: %v", err)
	}
	if config.Spec.retrievedAccessKey != "AKIAFROMSPEC" {
		t.Errorf("retrievedAccessKey = %q, want %q (spec.accessKey should take precedence)", config.Spec.retrievedAccessKey, "AKIAFROMSPEC")
	}
}

func TestAWSSecretEngineConfig_IsValid_ExactlyOneCredentialSource(t *testing.T) {
	config := &AWSSecretEngineConfig{
		Spec: AWSSecretEngineConfigSpec{
			RootCredentials: vaultutils.RootCredentialConfig{
				Secret: &corev1.LocalObjectReference{Name: "test"},
			},
		},
	}
	ok, err := config.IsValid()
	if !ok || err != nil {
		t.Errorf("expected valid with exactly one credential source, got ok=%v, err=%v", ok, err)
	}
}

func TestAWSSecretEngineConfig_IsValid_NoCredentialSource(t *testing.T) {
	config := &AWSSecretEngineConfig{
		Spec: AWSSecretEngineConfigSpec{
			RootCredentials: vaultutils.RootCredentialConfig{},
		},
	}
	ok, err := config.IsValid()
	if ok || err == nil {
		t.Error("expected invalid with no credential source")
	}
}

func TestAWSSecretEngineConfig_IsValid_RandomSecretWithoutAccessKey(t *testing.T) {
	config := &AWSSecretEngineConfig{
		Spec: AWSSecretEngineConfigSpec{
			RootCredentials: vaultutils.RootCredentialConfig{
				RandomSecret: &corev1.LocalObjectReference{Name: "random"},
			},
		},
	}
	ok, err := config.IsValid()
	if ok || err == nil {
		t.Error("expected invalid when randomSecret is used without spec.accessKey")
	}
}

func TestAWSSecretEngineConfig_IsValid_RandomSecretWithAccessKey(t *testing.T) {
	config := &AWSSecretEngineConfig{
		Spec: AWSSecretEngineConfigSpec{
			AWSRootConfig: AWSRootConfig{
				AccessKey: "AKIAEXAMPLE",
			},
			RootCredentials: vaultutils.RootCredentialConfig{
				RandomSecret: &corev1.LocalObjectReference{Name: "random"},
			},
		},
	}
	ok, err := config.IsValid()
	if !ok || err != nil {
		t.Errorf("expected valid with randomSecret and spec.accessKey set, got ok=%v, err=%v", ok, err)
	}
}

func TestAWSRootConfig_toMap_MaxRetriesNilEmitsDefault(t *testing.T) {
	config := AWSRootConfig{
		Region:             "us-east-1",
		retrievedAccessKey: "AKIAEXAMPLE",
		retrievedSecretKey: "secretkey",
	}

	result := config.toMap()

	mr, ok := result["max_retries"]
	if !ok {
		t.Fatal("expected max_retries key to always be present in toMap() output")
	}
	if mrInt, ok := mr.(int); !ok || mrInt != -1 {
		t.Errorf("max_retries = %v (%T), expected int -1 when pointer is nil", mr, mr)
	}
}

func TestAWSSecretEngineConfig_IsEquivalentToDesiredState_ClearedMaxRetriesDetectsDrift(t *testing.T) {
	config := &AWSSecretEngineConfig{
		Spec: AWSSecretEngineConfigSpec{
			Path: "aws",
			AWSRootConfig: AWSRootConfig{
				AccessKey: "AKIAEXAMPLE",
				Region:    "us-west-2",
			},
		},
	}

	vaultPayload := map[string]any{
		"access_key":  "AKIAEXAMPLE",
		"region":      "us-west-2",
		"max_retries": 5,
	}

	if config.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected false: maxRetries cleared from spec (nil → -1 default) but Vault still has 5")
	}
}

func TestAWSSecretEngineConfig_IsEquivalentToDesiredState_MaxRetriesNilNoFalseDrift(t *testing.T) {
	config := &AWSSecretEngineConfig{
		Spec: AWSSecretEngineConfigSpec{
			Path: "aws",
			AWSRootConfig: AWSRootConfig{
				AccessKey: "AKIAEXAMPLE",
				Region:    "us-west-2",
			},
		},
	}

	vaultPayload := map[string]any{
		"access_key": "AKIAEXAMPLE",
		"region":     "us-west-2",
	}

	if !config.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected true: maxRetries nil (default -1) and absent from Vault should not cause false drift")
	}
}

func TestAWSSecretEngineConfig_IsEquivalentToDesiredState_SecretKeyExcludedFromComparison(t *testing.T) {
	config := &AWSSecretEngineConfig{
		Spec: AWSSecretEngineConfigSpec{
			Path: "aws",
			AWSRootConfig: AWSRootConfig{
				Region:             "us-west-2",
				retrievedAccessKey: "AKIAEXAMPLE",
				retrievedSecretKey: "secretkey",
			},
		},
	}

	vaultPayload := map[string]any{
		"access_key": "AKIAEXAMPLE",
		"region":     "us-west-2",
	}

	if !config.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected true: secret_key is write-only and must be excluded from drift comparison; credential changes are detected via Secret watch predicates")
	}
}

func TestAWSSecretEngineConfig_IsEquivalentToDesiredState_AccessKeyChangeDetectedWithCreds(t *testing.T) {
	config := &AWSSecretEngineConfig{
		Spec: AWSSecretEngineConfigSpec{
			Path: "aws",
			AWSRootConfig: AWSRootConfig{
				Region:             "us-west-2",
				retrievedAccessKey: "AKIANEWKEY",
				retrievedSecretKey: "secretkey",
			},
		},
	}

	vaultPayload := map[string]any{
		"access_key": "AKIAOLDKEY",
		"region":     "us-west-2",
	}

	if config.IsEquivalentToDesiredState(vaultPayload) {
		t.Error("expected false: access_key change must be detected even when credentials are resolved")
	}
}

func TestAWSSecretEngineConfig_ValidateUpdate_RejectsNameChange(t *testing.T) {
	r := &AWSSecretEngineConfig{}
	oldObj := &AWSSecretEngineConfig{
		Spec: AWSSecretEngineConfigSpec{
			Path: "aws",
			Name: "old-name",
			RootCredentials: vaultutils.RootCredentialConfig{
				Secret: &corev1.LocalObjectReference{Name: "cred"},
			},
		},
	}
	newObj := &AWSSecretEngineConfig{
		Spec: AWSSecretEngineConfigSpec{
			Path: "aws",
			Name: "new-name",
			RootCredentials: vaultutils.RootCredentialConfig{
				Secret: &corev1.LocalObjectReference{Name: "cred"},
			},
		},
	}
	_, err := r.ValidateUpdate(nil, oldObj, newObj)
	if err == nil {
		t.Error("expected error when spec.name is changed")
	}
}

func TestAWSSecretEngineConfig_ValidateUpdate_AllowsSameNameUpdate(t *testing.T) {
	r := &AWSSecretEngineConfig{}
	oldObj := &AWSSecretEngineConfig{
		Spec: AWSSecretEngineConfigSpec{
			Path: "aws",
			Name: "same-name",
			RootCredentials: vaultutils.RootCredentialConfig{
				Secret: &corev1.LocalObjectReference{Name: "cred"},
			},
		},
	}
	newObj := &AWSSecretEngineConfig{
		Spec: AWSSecretEngineConfigSpec{
			Path: "aws",
			Name: "same-name",
			AWSRootConfig: AWSRootConfig{
				Region: "eu-west-1",
			},
			RootCredentials: vaultutils.RootCredentialConfig{
				Secret: &corev1.LocalObjectReference{Name: "cred"},
			},
		},
	}
	_, err := r.ValidateUpdate(nil, oldObj, newObj)
	if err != nil {
		t.Errorf("expected no error when spec.name unchanged, got: %v", err)
	}
}

func TestAWSSecretEngineConfig_SetInternalCredentials_K8sSecretMissingPasswordKey(t *testing.T) {
	ns := "ns-aws"
	sec := newK8sSecret(ns, "aws-creds", map[string][]byte{
		"username": []byte("AKIAEXAMPLE"),
	})
	kube := newFakeKubeClient(sec)
	hc := newFakeVaultHandler()
	vc, ts := newFakeVaultClient(t, hc)
	defer ts.Close()
	ctx := pivContext(kube, vc)
	config := &AWSSecretEngineConfig{
		ObjectMeta: metav1.ObjectMeta{Namespace: ns},
		Spec: AWSSecretEngineConfigSpec{
			Path: "aws",
			RootCredentials: vaultutils.RootCredentialConfig{
				Secret:      &corev1.LocalObjectReference{Name: "aws-creds"},
				UsernameKey: "username",
				PasswordKey: "password",
			},
		},
	}
	err := config.PrepareInternalValues(ctx, config)
	if err == nil {
		t.Fatal("expected error when K8s Secret is missing password key")
	}
}

func TestAWSSecretEngineConfig_SetInternalCredentials_K8sSecretMissingUsernameKey(t *testing.T) {
	ns := "ns-aws"
	sec := newK8sSecret(ns, "aws-creds", map[string][]byte{
		"password": []byte("secretkey"),
	})
	kube := newFakeKubeClient(sec)
	hc := newFakeVaultHandler()
	vc, ts := newFakeVaultClient(t, hc)
	defer ts.Close()
	ctx := pivContext(kube, vc)
	config := &AWSSecretEngineConfig{
		ObjectMeta: metav1.ObjectMeta{Namespace: ns},
		Spec: AWSSecretEngineConfigSpec{
			Path: "aws",
			RootCredentials: vaultutils.RootCredentialConfig{
				Secret:      &corev1.LocalObjectReference{Name: "aws-creds"},
				UsernameKey: "username",
				PasswordKey: "password",
			},
		},
	}
	err := config.PrepareInternalValues(ctx, config)
	if err == nil {
		t.Fatal("expected error when K8s Secret is missing username key")
	}
}

func TestAWSSecretEngineConfig_SetInternalCredentials_VaultSecretMissingKey(t *testing.T) {
	ns := "ns-aws"
	vaultPath := "secret/data/aws-creds"
	handler := newFakeVaultHandler()
	handler.setGet(vaultPath, map[string]any{
		"username": "AKIAFROMVAULT",
	})
	kube := newFakeKubeClient()
	vc, ts := newFakeVaultClient(t, handler)
	defer ts.Close()
	ctx := pivContext(kube, vc)
	config := &AWSSecretEngineConfig{
		ObjectMeta: metav1.ObjectMeta{Namespace: ns},
		Spec: AWSSecretEngineConfigSpec{
			Path: "aws",
			RootCredentials: vaultutils.RootCredentialConfig{
				VaultSecret: &vaultutils.VaultSecretReference{Path: vaultPath},
				UsernameKey: "username",
				PasswordKey: "password",
			},
		},
	}
	err := config.PrepareInternalValues(ctx, config)
	if err == nil {
		t.Fatal("expected error when VaultSecret is missing password key")
	}
}
