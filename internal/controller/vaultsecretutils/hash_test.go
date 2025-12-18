package vaultsecretutils

import (
	"strings"
	"testing"
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
