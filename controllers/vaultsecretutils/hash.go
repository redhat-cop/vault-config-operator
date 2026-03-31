package vaultsecretutils

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func hashHashes(l [32]byte, r [32]byte) [32]byte {
	e := make([]byte, 0, len(l)+len(r))
	e = append(e, l[:]...)
	e = append(e, r[:]...)
	return sha256.Sum256(e)
}

func HashData(data map[string][]byte) string {

	keys := make([]string, 0, len(data))

	for k := range data {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	var rootSha [32]byte
	for _, k := range keys {
		keySha := sha256.Sum256([]byte(k))
		dataSha := sha256.Sum256(data[k])
		nodeSha := hashHashes(keySha, dataSha)

		rootSha = hashHashes(rootSha, nodeSha)
	}

	return hex.EncodeToString(rootSha[:])
}

// HashMeta returns a SHA-256 hash of the object's labels and annotations.
func HashMeta(meta metav1.ObjectMeta) string {
	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("%v", meta.Labels)))
	h.Write([]byte(fmt.Sprintf("%v", meta.Annotations)))
	return hex.EncodeToString(h.Sum(nil))
}

// GetResourceVersion returns a string combining the object's generation and a hash of its metadata.
// This is used to detect spec or metadata changes that should trigger an immediate resync.
func GetResourceVersion(meta metav1.ObjectMeta) string {
	return fmt.Sprintf("%d-%s", meta.GetGeneration(), HashMeta(meta))
}
