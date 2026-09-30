package geojson

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	geo "github.com/faustbrian/go-geo"
)

func TestDepthCountdownRejectsExhaustedAndNegativeLimits(t *testing.T) {
	t.Parallel()

	data := []byte(`{"type":"Point","coordinates":[0,0]}`)
	limits := geo.DefaultLimits()
	for _, remainingDepth := range []int{0, -1} {
		if _, err := unmarshalGeometry(data, geo.WGS84(), limits, remainingDepth); !errors.Is(err, geo.ErrTopology) {
			t.Fatalf("unmarshalGeometry(depth %d) error = %v", remainingDepth, err)
		}
	}
	limits.MaxCollectionDepth = -1
	if _, err := Unmarshal(data, geo.WGS84(), limits); !errors.Is(err, geo.ErrTopology) {
		t.Fatalf("Unmarshal(negative depth limit) error = %v", err)
	}
	deep := []byte(`{"type":"GeometryCollection","geometries":[{"type":"GeometryCollection","geometries":[]}]}`)
	if _, err := unmarshalGeometry(deep, geo.WGS84(), geo.DefaultLimits(), 1); err == nil ||
		!strings.Contains(err.Error(), "collection depth limit exceeded") {
		t.Fatalf("unmarshalGeometry(exhausted child depth) error = %v", err)
	}
}

func TestCollectionRejectsExcessChildrenBeforeDecodingThem(t *testing.T) {
	t.Parallel()

	data := []byte(`{"type":"GeometryCollection","geometries":[{"type":"Point","coordinates":[]},{"type":"Point","coordinates":[]}]}`)
	_, err := Unmarshal(data, geo.WGS84(), geo.Limits{MaxGeometries: 2})
	if !errors.Is(err, geo.ErrTopology) || !strings.Contains(err.Error(), "geometry limit exceeded") {
		t.Fatalf("Unmarshal(over geometry limit) error = %v, want geometry limit", err)
	}
}

func TestDiagnosticsDoNotEchoCallerControlledMetadata(t *testing.T) {
	t.Parallel()

	marker := "privatemarker"
	_, err := NewFeature(nil, map[string]json.RawMessage{marker: []byte("{")}, nil)
	if !errors.Is(err, geo.ErrEncoding) || !strings.Contains(err.Error(), "property is not valid JSON") {
		t.Fatalf("invalid property classification = %v", err)
	}
	if strings.Contains(err.Error(), marker) {
		t.Fatal("invalid property diagnostic retained caller input")
	}

	_, err = Unmarshal([]byte(`{"type":"privatemarker"}`), geo.WGS84(), geo.DefaultLimits())
	if !errors.Is(err, geo.ErrUnsupported) || !strings.Contains(err.Error(), "unsupported geometry type") {
		t.Fatalf("unsupported type classification = %v", err)
	}
	if strings.Contains(err.Error(), marker) {
		t.Fatal("unsupported type diagnostic retained caller input")
	}
}
