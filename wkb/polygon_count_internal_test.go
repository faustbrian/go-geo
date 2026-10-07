package wkb

import (
	"encoding/binary"
	"errors"
	"math"
	"testing"

	geo "github.com/faustbrian/go-geo"
)

func TestPolygonRingCountPreservesWireRangeDiagnostic(t *testing.T) {
	for _, order := range []binary.ByteOrder{binary.LittleEndian, binary.BigEndian} {
		appendOrder := order.(binary.AppendByteOrder)
		for _, test := range []struct {
			holes int64
			rings uint32
		}{{0, 1}, {math.MaxUint32 - 1, math.MaxUint32}} {
			data, err := appendPolygonRingCount(nil, appendOrder, test.holes)
			if err != nil || len(data) != 4 || order.Uint32(data) != test.rings {
				t.Fatalf("valid hole count %d did not encode its exterior ring", test.holes)
			}
		}
		for _, holes := range []int64{math.MaxUint32, math.MaxUint32 + 1} {
			data, err := appendPolygonRingCount(nil, appendOrder, holes)
			var encoding *geo.EncodingError
			if data != nil || !errors.Is(err, geo.ErrTopology) || !errors.As(err, &encoding) || encoding.Problem != "polygon ring count exceeds WKB range" {
				t.Fatalf("invalid hole count %d lost its categorical polygon diagnostic", holes)
			}
		}
	}
}
