package geojson_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"

	geo "github.com/faustbrian/go-geo"
	"github.com/faustbrian/go-geo/geojson"
)

func TestFeatureIDRequiresOneCompleteScalar(t *testing.T) {
	for _, raw := range []string{`1 2`, `"first" "second"`, `0 garbage`, `"first" true`, `01`, `1[]`, `null`, `true`, `[]`, `{}`} {
		t.Run(raw, func(t *testing.T) {
			if _, err := geojson.NewFeature(nil, nil, json.RawMessage(raw)); !errors.Is(err, geo.ErrEncoding) {
				t.Fatalf("accepted incomplete Feature ID: %v", err)
			}
		})
	}
	for _, raw := range []string{`"id"`, `9007199254740993`, " \t-0\r\n ", `"escaped\u0069d"`} {
		feature, err := geojson.NewFeature(nil, nil, json.RawMessage(raw))
		if err != nil || string(feature.ID()) != raw {
			t.Fatalf("complete ID lost lexical identity: id=%q, err=%v", feature.ID(), err)
		}
		encoded, err := geojson.MarshalFeature(feature)
		if err != nil || !json.Valid(encoded) {
			t.Fatalf("complete ID cannot be published: %q, %v", encoded, err)
		}
	}
}

func TestMarshalFeaturePropagatesNestedGeometryError(t *testing.T) {
	foreign, err := geo.NewCRS(4326, "caller-label")
	if err != nil {
		t.Fatal(err)
	}
	lon, _ := geo.NewLongitude(1)
	lat, _ := geo.NewLatitude(2)
	coordinate, err := geo.NewCoordinate(lon, lat, foreign)
	if err != nil {
		t.Fatal(err)
	}
	point, err := geo.NewPoint(coordinate)
	if err != nil {
		t.Fatal(err)
	}
	feature, err := geojson.NewFeature(point, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := geojson.Marshal(point); !errors.Is(err, geo.ErrCRS) {
		t.Fatalf("nested-error control did not reach CRS refusal: %v", err)
	}
	encoded, err := geojson.MarshalFeature(feature)
	if !errors.Is(err, geo.ErrCRS) || len(encoded) != 0 {
		t.Fatalf("nested geometry failure became published null: bytes=%q, err=%v", encoded, err)
	}
	for _, geometry := range []geo.Geometry{nil, mustPoint(t, 1, 2)} {
		valid, err := geojson.NewFeature(geometry, nil, nil)
		if err != nil {
			t.Fatal(err)
		}
		encoded, err := geojson.MarshalFeature(valid)
		if err != nil || !json.Valid(encoded) {
			t.Fatalf("valid nullable/point Feature rejected: %q, %v", encoded, err)
		}
		if geometry == nil && !bytes.Contains(encoded, []byte(`"geometry":null`)) {
			t.Fatalf("actual absent geometry changed: %s", encoded)
		}
	}
}

func FuzzFeatureIDAdmission(f *testing.F) {
	for _, seed := range []string{`1 2`, `"first" true`, `01`, `null`, `true`, `[]`, `{}`, `9007199254740993`, `"id"`, " \t-0\n", ""} {
		f.Add([]byte(seed))
	}
	f.Fuzz(func(t *testing.T, raw []byte) {
		if len(raw) > 4096 {
			t.Skip()
		}
		feature, err := geojson.NewFeature(nil, nil, raw)
		if err != nil {
			return
		}
		if len(raw) > 0 && !json.Valid(raw) {
			t.Fatal("constructor accepted an incomplete JSON ID")
		}
		if len(raw) > 0 {
			decoder := json.NewDecoder(bytes.NewReader(raw))
			decoder.UseNumber()
			var value any
			if err := decoder.Decode(&value); err != nil {
				t.Fatal(err)
			}
			switch value.(type) {
			case string, json.Number:
			default:
				t.Fatal("constructor accepted a non-scalar Feature ID")
			}
		}
		if !bytes.Equal(feature.ID(), raw) {
			t.Fatal("constructor changed raw ID bytes")
		}
		encoded, err := geojson.MarshalFeature(feature)
		if err != nil || !json.Valid(encoded) {
			t.Fatalf("admitted ID cannot be published: %v", err)
		}
	})
}
