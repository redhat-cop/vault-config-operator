package v1alpha1

import (
	"encoding/json"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestLeaseCountQuotaGetPath(t *testing.T) {
	tests := []struct {
		name         string
		quota        *LeaseCountQuota
		expectedPath string
	}{
		{
			name: "with spec.name specified",
			quota: &LeaseCountQuota{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-quota",
				},
				Spec: LeaseCountQuotaSpec{
					Name: "custom-name",
				},
			},
			expectedPath: "sys/quotas/lease-count/custom-name",
		},
		{
			name: "without spec.name uses metadata.name",
			quota: &LeaseCountQuota{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-quota",
				},
				Spec: LeaseCountQuotaSpec{},
			},
			expectedPath: "sys/quotas/lease-count/test-quota",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.quota.GetPath()
			if result != tt.expectedPath {
				t.Errorf("GetPath() = %v, expected %v", result, tt.expectedPath)
			}
		})
	}
}

func TestLeaseCountQuotaIsDeletable(t *testing.T) {
	q := &LeaseCountQuota{}
	if !q.IsDeletable() {
		t.Error("expected LeaseCountQuota to be deletable")
	}
}

func TestLeaseCountQuotaToMap(t *testing.T) {
	cfg := LeaseCountQuotaConfig{
		MaxLeases:   1000,
		QuotaPath:   "secret/",
		Role:        "admin",
		Inheritable: true,
	}
	result := cfg.toMap()

	if result["max_leases"] != json.Number("1000") {
		t.Errorf("expected max_leases=1000, got %v (%T)", result["max_leases"], result["max_leases"])
	}
	if result["path"] != "secret/" {
		t.Errorf("expected path=secret/, got %v", result["path"])
	}
	if result["role"] != "admin" {
		t.Errorf("expected role=admin, got %v", result["role"])
	}
	if result["inheritable"] != true {
		t.Errorf("expected inheritable=true, got %v", result["inheritable"])
	}
}

func TestLeaseCountQuotaToMapJsonNumber(t *testing.T) {
	cfg := LeaseCountQuotaConfig{
		MaxLeases: 500,
	}
	result := cfg.toMap()

	n, ok := result["max_leases"].(json.Number)
	if !ok {
		t.Fatalf("expected max_leases to be json.Number, got %T", result["max_leases"])
	}
	if n != json.Number("500") {
		t.Errorf("expected json.Number(500), got %v", n)
	}
}

func TestLeaseCountQuotaIsEquivalentToDesiredState(t *testing.T) {
	q := &LeaseCountQuota{
		Spec: LeaseCountQuotaSpec{
			LeaseCountQuotaConfig: LeaseCountQuotaConfig{
				MaxLeases: 1000,
			},
		},
	}

	matching := map[string]any{
		"max_leases":  json.Number("1000"),
		"path":        "",
		"role":        "",
		"inheritable": false,
	}
	if !q.IsEquivalentToDesiredState(matching) {
		t.Error("expected matching payload to be equivalent")
	}

	nonMatching := map[string]any{
		"max_leases":  json.Number("500"),
		"path":        "",
		"role":        "",
		"inheritable": false,
	}
	if q.IsEquivalentToDesiredState(nonMatching) {
		t.Error("expected non-matching payload to not be equivalent")
	}
}

func TestLeaseCountQuotaIsEquivalentExtraFields(t *testing.T) {
	q := &LeaseCountQuota{
		Spec: LeaseCountQuotaSpec{
			LeaseCountQuotaConfig: LeaseCountQuotaConfig{
				MaxLeases: 1000,
			},
		},
	}

	payloadWithExtra := map[string]any{
		"max_leases":  json.Number("1000"),
		"path":        "",
		"role":        "",
		"inheritable": false,
		"type":        "lease-count",
		"name":        "test",
	}
	if !q.IsEquivalentToDesiredState(payloadWithExtra) {
		t.Error("expected extra fields to be ignored")
	}
}

func TestLeaseCountQuotaConditions(t *testing.T) {
	condition := metav1.Condition{
		Type:   "Ready",
		Status: metav1.ConditionTrue,
	}

	q := &LeaseCountQuota{}
	q.SetConditions([]metav1.Condition{condition})
	if len(q.GetConditions()) != 1 || q.GetConditions()[0].Type != "Ready" {
		t.Error("expected conditions to be set and retrieved")
	}
}

func TestLeaseCountQuotaIsValid(t *testing.T) {
	valid := &LeaseCountQuota{
		Spec: LeaseCountQuotaSpec{
			LeaseCountQuotaConfig: LeaseCountQuotaConfig{MaxLeases: 1},
		},
	}
	ok, err := valid.IsValid()
	if !ok || err != nil {
		t.Errorf("expected valid quota, got ok=%v err=%v", ok, err)
	}

	invalid := &LeaseCountQuota{
		Spec: LeaseCountQuotaSpec{
			LeaseCountQuotaConfig: LeaseCountQuotaConfig{MaxLeases: 0},
		},
	}
	ok, err = invalid.IsValid()
	if ok || err == nil {
		t.Error("expected invalid quota with maxLeases=0")
	}
}
