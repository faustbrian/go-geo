package wkb_test

import (
	"testing"

	geo "github.com/faustbrian/go-geo"
	"github.com/faustbrian/go-geo/wkb"
)

// FuzzDecode invokes plain WKB and SRID-bearing EWKB under small input and
// geometry limits. Errors/results are discarded, not compared with an external
// wire oracle or checked for correct rejection.

func FuzzDecode(f *testing.F) {
	f.Add([]byte{1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
	f.Add([]byte{1, 1, 0, 0, 32, 230, 16, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
	f.Add([]byte{1, 7, 0, 0, 0, 255, 255, 255, 255})

	limits := geo.DefaultLimits()
	limits.MaxPoints = 64
	limits.MaxRings = 16
	limits.MaxGeometries = 32
	limits.MaxCollectionDepth = 8
	limits.MaxEncodedBytes = 4096
	f.Fuzz(func(_ *testing.T, data []byte) {
		_, _ = wkb.Unmarshal(data, geo.WGS84(), limits)
		_, _ = wkb.UnmarshalEWKB(data, limits)
	})
}
