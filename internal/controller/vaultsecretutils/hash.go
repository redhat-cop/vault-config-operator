package vaultsecretutils

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
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
