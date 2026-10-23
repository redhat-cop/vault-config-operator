package v1alpha1

import (
	"encoding/json"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestRateLimitQuotaGetPath(t *testing.T) {
	tests := []struct {
		name         string
		quota        *RateLimitQuota
		expectedPath string
	}{
		{
			name: "with spec.name specified",
			quota: &RateLimitQuota{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-quota",
				},
				Spec: RateLimitQuotaSpec{
					Name: "custom-name",
				},
			},
			expectedPath: "sys/quotas/rate-limit/custom-name",
		},
		{
			name: "without spec.name uses metadata.name",
			quota: &RateLimitQuota{
				ObjectMeta: metav1.ObjectMeta{
					Name: "test-quota",
				},
				Spec: RateLimitQuotaSpec{},
			},
			expectedPath: "sys/quotas/rate-limit/test-quota",
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

func TestRateLimitQuotaIsDeletable(t *testing.T) {
	q := &RateLimitQuota{}
	if !q.IsDeletable() {
		t.Error("expected RateLimitQuota to be deletable")
	}
}

func TestRateLimitQuotaToMapDurationConversion(t *testing.T) {
	cfg := RateLimitQuotaConfig{
		Rate:          "100",
		Interval:      "1m",
		BlockInterval: "5m",
		QuotaPath:     "secret/",
		Role:          "admin",
	}
	result := cfg.toMap()

	if result["rate"] != json.Number("100") {
		t.Errorf("expected rate=100, got %v", result["rate"])
	}
	if result["interval"] != json.Number("60") {
		t.Errorf("expected interval=60 (1m in seconds), got %v", result["interval"])
	}
	if result["block_interval"] != json.Number("300") {
		t.Errorf("expected block_interval=300 (5m in seconds), got %v", result["block_interval"])
	}
	if result["path"] != "secret/" {
		t.Errorf("expected path=secret/, got %v", result["path"])
	}
	if result["role"] != "admin" {
		t.Errorf("expected role=admin, got %v", result["role"])
	}
}

func TestRateLimitQuotaToMapFloatRate(t *testing.T) {
	cfg := RateLimitQuotaConfig{
		Rate: "897.3",
	}
	result := cfg.toMap()

	if result["rate"] != json.Number("897.3") {
		t.Errorf("expected rate=897.3, got %v (%T)", result["rate"], result["rate"])
	}
}

func TestRateLimitQuotaToMapIntegerRate(t *testing.T) {
	cfg := RateLimitQuotaConfig{
		Rate: "100",
	}
	result := cfg.toMap()

	if result["rate"] != json.Number("100") {
		t.Errorf("expected rate=100 (not 100.0), got %v", result["rate"])
	}
}

func TestRateToJsonNumber(t *testing.T) {
	tests := []struct {
		name     string
		rate     string
		expected json.Number
	}{
		{"integer rate", "100", json.Number("100")},
		{"float rate", "897.3", json.Number("897.3")},
		{"zero rate", "0", json.Number("0")},
		{"large integer", "10000", json.Number("10000")},
		{"empty rate", "", json.Number("0")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rateToJsonNumber(tt.rate)
			if result != tt.expected {
				t.Errorf("rateToJsonNumber(%v) = %v, expected %v", tt.rate, result, tt.expected)
			}
		})
	}
}

func TestRateLimitQuotaIsEquivalentToDesiredState(t *testing.T) {
	q := &RateLimitQuota{
		Spec: RateLimitQuotaSpec{
			RateLimitQuotaConfig: RateLimitQuotaConfig{
				Rate:     "100",
				Interval: "1m",
			},
		},
	}

	matching := map[string]any{
		"rate":           json.Number("100"),
		"interval":       json.Number("60"),
		"block_interval": json.Number("0"),
		"path":           "",
		"role":           "",
		"inheritable":    false,
	}
	if !q.IsEquivalentToDesiredState(matching) {
		t.Error("expected matching payload to be equivalent")
	}

	nonMatching := map[string]any{
		"rate":           json.Number("200"),
		"interval":       json.Number("60"),
		"block_interval": json.Number("0"),
		"path":           "",
		"role":           "",
		"inheritable":    false,
	}
	if q.IsEquivalentToDesiredState(nonMatching) {
		t.Error("expected non-matching payload to not be equivalent")
	}
}

func TestRateLimitQuotaIsEquivalentExtraFields(t *testing.T) {
	q := &RateLimitQuota{
		Spec: RateLimitQuotaSpec{
			RateLimitQuotaConfig: RateLimitQuotaConfig{
				Rate: "100",
			},
		},
	}

	payloadWithExtra := map[string]any{
		"rate":           json.Number("100"),
		"interval":       json.Number("0"),
		"block_interval": json.Number("0"),
		"path":           "",
		"role":           "",
		"inheritable":    false,
		"type":           "rate-limit",
		"name":           "test",
	}
	if !q.IsEquivalentToDesiredState(payloadWithExtra) {
		t.Error("expected extra fields to be ignored")
	}
}

func TestRateLimitQuotaConditions(t *testing.T) {
	condition := metav1.Condition{
		Type:   "Ready",
		Status: metav1.ConditionTrue,
	}

	q := &RateLimitQuota{}
	q.SetConditions([]metav1.Condition{condition})
	if len(q.GetConditions()) != 1 || q.GetConditions()[0].Type != "Ready" {
		t.Error("expected conditions to be set and retrieved")
	}
}

func TestRateLimitQuotaIsValid(t *testing.T) {
	valid := &RateLimitQuota{
		Spec: RateLimitQuotaSpec{
			RateLimitQuotaConfig: RateLimitQuotaConfig{Rate: "100"},
		},
	}
	ok, err := valid.IsValid()
	if !ok || err != nil {
		t.Errorf("expected valid quota, got ok=%v err=%v", ok, err)
	}

	invalid := &RateLimitQuota{
		Spec: RateLimitQuotaSpec{
			RateLimitQuotaConfig: RateLimitQuotaConfig{Rate: "0"},
		},
	}
	ok, err = invalid.IsValid()
	if ok || err == nil {
		t.Error("expected invalid quota with rate=0")
	}

	invalidNaN := &RateLimitQuota{
		Spec: RateLimitQuotaSpec{
			RateLimitQuotaConfig: RateLimitQuotaConfig{Rate: "not-a-number"},
		},
	}
	ok, err = invalidNaN.IsValid()
	if ok || err == nil {
		t.Error("expected invalid quota with non-numeric rate")
	}
}
