package vaultsecretutils

import (
	"strings"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	firstHash      = "fed1e523a1da70393b9b22364aa8984d549536f069c1eba46a85847c46e52dc1"
	emptyOrNilHash = "0000000000000000000000000000000000000000000000000000000000000000"
	k1             = "some data"
)

var k2 string

func TestMain(t *testing.T) {
	var b strings.Builder
	b.Grow(1048577)
	for i := 0; i < 1048577; i++ {
		b.WriteByte(0)
	}
	k2 = b.String()
}

func TestHash(t *testing.T) {

	m := make(map[string][]byte)
	m["k1"] = []byte(k1)
	m["k2"] = []byte(k2)

	hash := HashData(m)

	t.Logf("%v\n", hash)

	if hash != firstHash {
		t.Errorf("Unexpected Hash, got: %v, want: %v.", hash, firstHash)
	}

}

func TestUnorderedHash(t *testing.T) {
	m := make(map[string][]byte)
	m["k2"] = []byte(k2)
	m["k1"] = []byte(k1)
	hash := HashData(m)

	t.Logf("%v\n", hash)

	if hash != firstHash {
		t.Errorf("Unexpected Hash, got: %v, want: %v.", hash, firstHash)
	}
}

func TestSingleKeyHash(t *testing.T) {
	m := make(map[string][]byte)
	m["k1"] = []byte(k1)

	hash := HashData(m)

	t.Logf("%v\n", hash)

	k1DataHash := "3e8df15fa3fde92176fbdebdd649bf7058e9756d79932a176131aee4a3cc5745"

	if hash != k1DataHash {
		t.Errorf("Unexpected Hash, got: %v, want: %v.", hash, k1DataHash)
	}
}

func TestUnequalHash(t *testing.T) {
	m := make(map[string][]byte)

	m["k2"] = []byte(k2)
	m["k1"] = []byte("some other data")
	hash := HashData(m)

	t.Logf("%v\n", hash)

	if hash == firstHash {
		t.Errorf("Unexpected Hash, got: %v, do not want: %v.", hash, firstHash)
	}
}

func TestEmpty(t *testing.T) {
	m := make(map[string][]byte)

	hash := HashData(m)

	t.Logf("%v\n", hash)

	if hash != emptyOrNilHash {
		t.Errorf("Unexpected Hash, got: %v, want: %v.", hash, emptyOrNilHash)
	}
}

func TestNil(t *testing.T) {
	hash := HashData(nil)

	t.Logf("%v\n", hash)

	if hash != emptyOrNilHash {
		t.Errorf("Unexpected Hash, got: %v, want: %v.", hash, emptyOrNilHash)
	}
}

func TestHashMeta(t *testing.T) {
	meta := metav1.ObjectMeta{
		Labels:      map[string]string{"app": "test"},
		Annotations: map[string]string{"note": "value"},
	}

	hash := HashMeta(meta)

	if hash == "" {
		t.Error("HashMeta returned empty string")
	}

	// same input should produce same hash
	hash2 := HashMeta(meta)
	if hash != hash2 {
		t.Errorf("HashMeta not deterministic, got: %v and %v", hash, hash2)
	}
}

func TestHashMetaDifferentLabels(t *testing.T) {
	meta1 := metav1.ObjectMeta{
		Labels: map[string]string{"app": "test"},
	}
	meta2 := metav1.ObjectMeta{
		Labels: map[string]string{"app": "other"},
	}

	if HashMeta(meta1) == HashMeta(meta2) {
		t.Error("HashMeta should differ when labels differ")
	}
}

func TestHashMetaDifferentAnnotations(t *testing.T) {
	meta1 := metav1.ObjectMeta{
		Annotations: map[string]string{"key": "value1"},
	}
	meta2 := metav1.ObjectMeta{
		Annotations: map[string]string{"key": "value2"},
	}

	if HashMeta(meta1) == HashMeta(meta2) {
		t.Error("HashMeta should differ when annotations differ")
	}
}

func TestHashMetaNilMaps(t *testing.T) {
	meta := metav1.ObjectMeta{}

	hash := HashMeta(meta)

	if hash == "" {
		t.Error("HashMeta returned empty string for nil labels/annotations")
	}
}

func TestHashMetaEmptyMaps(t *testing.T) {
	meta1 := metav1.ObjectMeta{
		Labels:      map[string]string{},
		Annotations: map[string]string{},
	}
	meta2 := metav1.ObjectMeta{}

	// empty maps and nil maps should produce the same hash since fmt.Sprintf
	// renders both as "map[]"
	if HashMeta(meta1) != HashMeta(meta2) {
		t.Errorf("HashMeta should be equal for empty and nil maps, got: %v and %v", HashMeta(meta1), HashMeta(meta2))
	}
}

func TestGetResourceVersion(t *testing.T) {
	meta := metav1.ObjectMeta{
		Generation: 1,
		Labels:     map[string]string{"app": "test"},
	}

	rv := GetResourceVersion(meta)

	if rv == "" {
		t.Error("GetResourceVersion returned empty string")
	}

	// should start with the generation number
	if rv[:2] != "1-" {
		t.Errorf("GetResourceVersion should start with generation, got: %v", rv)
	}
}

func TestGetResourceVersionChangesOnGeneration(t *testing.T) {
	meta1 := metav1.ObjectMeta{
		Generation: 1,
		Labels:     map[string]string{"app": "test"},
	}
	meta2 := metav1.ObjectMeta{
		Generation: 2,
		Labels:     map[string]string{"app": "test"},
	}

	rv1 := GetResourceVersion(meta1)
	rv2 := GetResourceVersion(meta2)

	if rv1 == rv2 {
		t.Error("GetResourceVersion should differ when generation differs")
	}
}

func TestGetResourceVersionChangesOnMetadata(t *testing.T) {
	meta1 := metav1.ObjectMeta{
		Generation: 1,
		Labels:     map[string]string{"app": "test"},
	}
	meta2 := metav1.ObjectMeta{
		Generation: 1,
		Labels:     map[string]string{"app": "changed"},
	}

	rv1 := GetResourceVersion(meta1)
	rv2 := GetResourceVersion(meta2)

	if rv1 == rv2 {
		t.Error("GetResourceVersion should differ when labels differ")
	}
}

func TestGetResourceVersionStableWhenUnchanged(t *testing.T) {
	meta := metav1.ObjectMeta{
		Generation:  3,
		Labels:      map[string]string{"app": "test"},
		Annotations: map[string]string{"note": "value"},
	}

	rv1 := GetResourceVersion(meta)
	rv2 := GetResourceVersion(meta)

	if rv1 != rv2 {
		t.Errorf("GetResourceVersion should be stable, got: %v and %v", rv1, rv2)
	}
}
