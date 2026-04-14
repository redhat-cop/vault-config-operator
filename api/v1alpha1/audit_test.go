package v1alpha1

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestAuditGetPath(t *testing.T) {
	audit := &Audit{
		Spec: AuditSpec{
			Path: "file-audit",
		},
	}

	result := audit.GetPath()
	expected := "sys/audit/file-audit"
	if result != expected {
		t.Errorf("GetPath() = %v, expected %v", result, expected)
	}
}

func TestAuditGetPayload(t *testing.T) {
	audit := &Audit{
		Spec: AuditSpec{
			Type:        "file",
			Description: "File audit device",
			Local:       true,
			Options:     map[string]string{"file_path": "/var/log/vault_audit.log"},
		},
	}

	payload := audit.GetPayload()

	if len(payload) != 4 {
		t.Errorf("expected 4 keys in payload, got %d", len(payload))
	}
	if payload["type"] != "file" {
		t.Errorf("expected type 'file', got %v", payload["type"])
	}
	if payload["description"] != "File audit device" {
		t.Errorf("expected description 'File audit device', got %v", payload["description"])
	}
	if payload["local"] != true {
		t.Errorf("expected local true, got %v", payload["local"])
	}
	opts, ok := payload["options"].(map[string]string)
	if !ok {
		t.Fatal("expected options to be map[string]string")
	}
	if opts["file_path"] != "/var/log/vault_audit.log" {
		t.Errorf("expected file_path option, got %v", opts["file_path"])
	}
}

func TestAuditIsEquivalentMatching(t *testing.T) {
	audit := &Audit{
		Spec: AuditSpec{
			Type:        "file",
			Description: "File audit device",
			Local:       false,
			Options:     map[string]string{"file_path": "/var/log/vault_audit.log"},
		},
	}

	payload := map[string]interface{}{
		"type":        "file",
		"description": "File audit device",
		"local":       false,
		"options":     map[string]string{"file_path": "/var/log/vault_audit.log"},
	}
	if !audit.IsEquivalentToDesiredState(payload) {
		t.Error("expected matching payload to be equivalent")
	}
}

func TestAuditIsEquivalentTypeMismatch(t *testing.T) {
	audit := &Audit{
		Spec: AuditSpec{
			Type:        "file",
			Description: "File audit device",
			Local:       false,
			Options:     map[string]string{"file_path": "/var/log/vault_audit.log"},
		},
	}

	payload := map[string]interface{}{
		"type":        "socket",
		"description": "File audit device",
		"local":       false,
		"options":     map[string]string{"file_path": "/var/log/vault_audit.log"},
	}
	if audit.IsEquivalentToDesiredState(payload) {
		t.Error("expected different type to not be equivalent")
	}
}

func TestAuditIsEquivalentDescriptionMismatch(t *testing.T) {
	audit := &Audit{
		Spec: AuditSpec{
			Type:        "file",
			Description: "File audit device",
			Local:       false,
			Options:     map[string]string{"file_path": "/var/log/vault_audit.log"},
		},
	}

	payload := map[string]interface{}{
		"type":        "file",
		"description": "Different description",
		"local":       false,
		"options":     map[string]string{"file_path": "/var/log/vault_audit.log"},
	}
	if audit.IsEquivalentToDesiredState(payload) {
		t.Error("expected different description to not be equivalent")
	}
}

func TestAuditIsEquivalentLocalMismatch(t *testing.T) {
	audit := &Audit{
		Spec: AuditSpec{
			Type:        "file",
			Description: "File audit device",
			Local:       false,
			Options:     map[string]string{"file_path": "/var/log/vault_audit.log"},
		},
	}

	payload := map[string]interface{}{
		"type":        "file",
		"description": "File audit device",
		"local":       true,
		"options":     map[string]string{"file_path": "/var/log/vault_audit.log"},
	}
	if audit.IsEquivalentToDesiredState(payload) {
		t.Error("expected different local to not be equivalent")
	}
}

func TestAuditIsEquivalentOptionsMismatch(t *testing.T) {
	audit := &Audit{
		Spec: AuditSpec{
			Type:        "file",
			Description: "File audit device",
			Local:       false,
			Options:     map[string]string{"file_path": "/var/log/vault_audit.log"},
		},
	}

	payload := map[string]interface{}{
		"type":        "file",
		"description": "File audit device",
		"local":       false,
		"options":     map[string]string{"file_path": "/var/log/other.log"},
	}
	if audit.IsEquivalentToDesiredState(payload) {
		t.Error("expected different option value to not be equivalent")
	}
}

func TestAuditIsEquivalentOptionsLengthMismatch(t *testing.T) {
	audit := &Audit{
		Spec: AuditSpec{
			Type:        "file",
			Description: "File audit device",
			Local:       false,
			Options:     map[string]string{"file_path": "/var/log/vault_audit.log"},
		},
	}

	payload := map[string]interface{}{
		"type":        "file",
		"description": "File audit device",
		"local":       false,
		"options":     map[string]string{"file_path": "/var/log/vault_audit.log", "log_raw": "true"},
	}
	if audit.IsEquivalentToDesiredState(payload) {
		t.Error("expected different options length to not be equivalent")
	}
}

// Audit's type assertion to map[string]string fails when Vault returns
// map[string]interface{} for the options field, causing IsEquivalent to
// return false.
func TestAuditIsEquivalentOptionsWrongType(t *testing.T) {
	audit := &Audit{
		Spec: AuditSpec{
			Type:        "file",
			Description: "File audit device",
			Local:       false,
			Options:     map[string]string{"file_path": "/var/log/vault_audit.log"},
		},
	}

	payload := map[string]interface{}{
		"type":        "file",
		"description": "File audit device",
		"local":       false,
		"options":     map[string]interface{}{"file_path": "/var/log/vault_audit.log"},
	}
	if audit.IsEquivalentToDesiredState(payload) {
		t.Error("expected options as map[string]interface{} (not map[string]string) to fail type assertion and return false")
	}
}

// Audit checks only its 4 managed fields individually and ignores any
// extra top-level keys in the payload. This is different from most types
// that use reflect.DeepEqual on the full map.
func TestAuditIsEquivalentExtraFields(t *testing.T) {
	audit := &Audit{
		Spec: AuditSpec{
			Type:        "file",
			Description: "File audit device",
			Local:       false,
			Options:     map[string]string{"file_path": "/var/log/vault_audit.log"},
		},
	}

	payload := map[string]interface{}{
		"type":        "file",
		"description": "File audit device",
		"local":       false,
		"options":     map[string]string{"file_path": "/var/log/vault_audit.log"},
		"extra_field": "vault-returned-value",
		"path":        "file-audit/",
	}
	if !audit.IsEquivalentToDesiredState(payload) {
		t.Error("expected Audit to IGNORE extra top-level keys (field-by-field comparison, not DeepEqual)")
	}
}

func TestAuditIsDeletable(t *testing.T) {
	audit := &Audit{}
	if !audit.IsDeletable() {
		t.Error("expected Audit to be deletable")
	}
}

func TestAuditConditions(t *testing.T) {
	condition := metav1.Condition{
		Type:   "Ready",
		Status: metav1.ConditionTrue,
	}

	audit := &Audit{}
	audit.SetConditions([]metav1.Condition{condition})
	got := audit.GetConditions()
	if len(got) != 1 || got[0].Type != "Ready" {
		t.Error("expected Audit conditions to be set and retrieved")
	}
}
