package geohash_test

import (
	"testing"

	"github.com/faustbrian/go-geo/geohash"
)

// FuzzDecode exercises validation and neighbor construction for arbitrary hash
// spellings. The decoder rejects lengths above twelve; results are discarded
// rather than checked as an independent cell or rejection oracle.

func FuzzDecode(f *testing.F) {
	f.Add("ezs42")
	f.Add("u60g0b3")
	f.Add("")

	f.Fuzz(func(_ *testing.T, value string) {
		hash := geohash.Hash(value)
		_, _ = geohash.Decode(hash)
		_, _ = geohash.Neighbors(hash)
	})
}
