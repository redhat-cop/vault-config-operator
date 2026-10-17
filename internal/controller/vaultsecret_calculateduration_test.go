package controller

import (
	"testing"
	"time"

	redhatcopv1alpha1 "github.com/redhat-cop/vault-config-operator/api/v1alpha1"
	"github.com/redhat-cop/vault-config-operator/internal/controller/vaultresourcecontroller"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func newTestVaultSecretReconciler() *VaultSecretReconciler {
	return &VaultSecretReconciler{
		ReconcilerBase: vaultresourcecontroller.ReconcilerBase{},
	}
}

func TestCalculateDuration_ExplicitRefreshPeriod(t *testing.T) {
	r := newTestVaultSecretReconciler()
	instance := &redhatcopv1alpha1.VaultSecret{
		Spec: redhatcopv1alpha1.VaultSecretSpec{
			RefreshPeriod:    &metav1.Duration{Duration: 5 * time.Minute},
			RefreshThreshold: 90,
		},
	}
	dur, ok := r.calculateDuration(instance)
	if !ok {
		t.Fatal("expected ok=true when RefreshPeriod is set")
	}
	if dur != 5*time.Minute {
		t.Fatalf("expected 5m, got %v", dur)
	}
}

func TestCalculateDuration_NoRefreshPeriodNoStatus(t *testing.T) {
	r := newTestVaultSecretReconciler()
	instance := &redhatcopv1alpha1.VaultSecret{
		Spec: redhatcopv1alpha1.VaultSecretSpec{
			RefreshThreshold: 90,
		},
	}
	_, ok := r.calculateDuration(instance)
	if ok {
		t.Fatal("expected ok=false when no RefreshPeriod and no status")
	}
}

func TestCalculateDuration_ZeroLeaseDuration_NoRefreshPeriod(t *testing.T) {
	r := newTestVaultSecretReconciler()
	instance := &redhatcopv1alpha1.VaultSecret{
		Spec: redhatcopv1alpha1.VaultSecretSpec{
			RefreshThreshold: 90,
		},
		Status: redhatcopv1alpha1.VaultSecretStatus{
			VaultSecretDefinitionsStatus: []redhatcopv1alpha1.VaultSecretDefinitionStatus{
				{Name: "kv-secret", LeaseDuration: 0},
			},
		},
	}
	_, ok := r.calculateDuration(instance)
	if ok {
		t.Fatal("expected ok=false for KV secrets with LeaseDuration=0 and no RefreshPeriod (would cause infinite loop)")
	}
}

func TestCalculateDuration_MixedLeaseDurations_ZeroIgnored(t *testing.T) {
	r := newTestVaultSecretReconciler()
	instance := &redhatcopv1alpha1.VaultSecret{
		Spec: redhatcopv1alpha1.VaultSecretSpec{
			RefreshThreshold: 90,
		},
		Status: redhatcopv1alpha1.VaultSecretStatus{
			VaultSecretDefinitionsStatus: []redhatcopv1alpha1.VaultSecretDefinitionStatus{
				{Name: "kv-secret", LeaseDuration: 0},
				{Name: "dynamic-secret", LeaseDuration: 3600},
			},
		},
	}
	dur, ok := r.calculateDuration(instance)
	if !ok {
		t.Fatal("expected ok=true when at least one positive LeaseDuration exists")
	}
	expected := time.Duration(float64(3600)*0.9) * time.Second
	if dur != expected {
		t.Fatalf("expected %v, got %v", expected, dur)
	}
}

func TestCalculateDuration_PositiveLeaseDuration(t *testing.T) {
	r := newTestVaultSecretReconciler()
	instance := &redhatcopv1alpha1.VaultSecret{
		Spec: redhatcopv1alpha1.VaultSecretSpec{
			RefreshThreshold: 90,
		},
		Status: redhatcopv1alpha1.VaultSecretStatus{
			VaultSecretDefinitionsStatus: []redhatcopv1alpha1.VaultSecretDefinitionStatus{
				{Name: "dynamic-secret-1", LeaseDuration: 7200},
				{Name: "dynamic-secret-2", LeaseDuration: 3600},
			},
		},
	}
	dur, ok := r.calculateDuration(instance)
	if !ok {
		t.Fatal("expected ok=true")
	}
	expected := time.Duration(float64(3600)*0.9) * time.Second
	if dur != expected {
		t.Fatalf("expected %v (smallest lease * threshold), got %v", expected, dur)
	}
}

func TestCalculateDuration_AllZeroLeaseDurations(t *testing.T) {
	r := newTestVaultSecretReconciler()
	instance := &redhatcopv1alpha1.VaultSecret{
		Spec: redhatcopv1alpha1.VaultSecretSpec{
			RefreshThreshold: 90,
		},
		Status: redhatcopv1alpha1.VaultSecretStatus{
			VaultSecretDefinitionsStatus: []redhatcopv1alpha1.VaultSecretDefinitionStatus{
				{Name: "kv-secret-1", LeaseDuration: 0},
				{Name: "kv-secret-2", LeaseDuration: 0},
			},
		},
	}
	_, ok := r.calculateDuration(instance)
	if ok {
		t.Fatal("expected ok=false when all LeaseDurations are 0")
	}
}
