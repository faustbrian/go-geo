//lint:file-ignore SA1019 This parity fuzz target intentionally exercises the deprecated compatibility facade.

package geogeom_test

import (
	"encoding/binary"
	"errors"
	"math"
	"reflect"
	"testing"

	"github.com/twpayne/go-geom"

	geo "github.com/faustbrian/go-geo"
	legacy "github.com/faustbrian/go-geo/adapter/gogeom" //nolint:staticcheck // Required parity fuzzing exercises the deprecated facade.
	geogeom "github.com/faustbrian/go-geo/adapters/geom"
)

func FuzzFromGoGeom(f *testing.F) {
	f.Add([]byte{})
	f.Add([]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
	f.Add([]byte{1, 2, 3, 4, 5, 6, 7, 8})

	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 256 {
			data = data[:256]
		}
		selector := data
		flat := make([]float64, 0, len(data)/8)
		for len(data) >= 8 {
			flat = append(flat, math.Float64frombits(binary.LittleEndian.Uint64(data[:8])))
			data = data[8:]
		}
		line := geom.NewLineStringFlat(geom.XY, flat).SetSRID(4326)
		collection := fuzzCollection(selector, line)
		limits := fuzzAdapterLimits()
		legacyConverted, legacyErr := legacy.FromGoGeom(collection, limits) //nolint:staticcheck // Required parity fuzzing exercises the deprecated facade.
		converted, err := geogeom.FromGoGeom(collection, limits)
		assertFuzzParity(t, legacyConverted, converted, legacyErr, err)
		if err != nil {
			return
		}
		legacyValue, legacyToErr := legacy.ToGoGeom(legacyConverted) //nolint:staticcheck // Required parity fuzzing exercises the deprecated facade.
		value, toErr := geogeom.ToGoGeom(converted)
		assertFuzzErrors(t, legacyToErr, toErr)
		if toErr == nil && !reflect.DeepEqual(legacyValue, value) {
			t.Fatal("ToGoGeom results differ")
		}
	})
}

func fuzzCollection(data []byte, line *geom.LineString) *geom.GeometryCollection {
	mode := byte(0)
	if len(data) > 0 {
		mode = data[0] % 8
	}
	switch mode {
	case 0:
		return geom.NewGeometryCollection().MustPush(line).MustSetLayout(geom.XY).SetSRID(4326)
	case 1:
		return geom.NewGeometryCollection().MustSetLayout(geom.XY).SetSRID(4326)
	case 2:
		depth := 31
		if len(data) > 1 {
			depth += int(data[1] % 3)
		}
		return fuzzCollectionChain(depth)
	case 3:
		collection := geom.NewGeometryCollection().MustPush(line).MustSetLayout(geom.XY).SetSRID(4326)
		collection.Geoms()[0] = collection
		return collection
	case 4:
		collection := geom.NewGeometryCollection().MustPush(line).MustSetLayout(geom.XY).SetSRID(4326)
		collection.Geoms()[0] = (*geom.Point)(nil)
		return collection
	case 5:
		nested := geom.NewGeometryCollection().MustPush(line).MustSetLayout(geom.XY).SetSRID(4326)
		return geom.NewGeometryCollection().MustPush(nested, line).MustSetLayout(geom.XY).SetSRID(4326)
	case 6:
		collection := geom.NewGeometryCollection().MustPush(line).SetSRID(0)
		collection.Geoms()[0] = (*geom.Point)(nil)
		return collection
	default:
		odd := geom.NewLineStringFlat(geom.XY, []float64{0}).SetSRID(4326)
		return geom.NewGeometryCollection().MustPush(odd, line).MustSetLayout(geom.XY).SetSRID(4326)
	}
}

func assertFuzzParity(t *testing.T, legacyValue, value geo.Geometry, legacyErr, err error) {
	t.Helper()
	assertFuzzErrors(t, legacyErr, err)
	if err == nil && !geo.EqualGeometry(legacyValue, value) {
		t.Fatal("FromGoGeom results differ")
	}
}

func assertFuzzErrors(t *testing.T, legacyErr, err error) {
	t.Helper()
	if (legacyErr == nil) != (err == nil) {
		t.Fatalf("error presence differs: (%v, %v)", legacyErr, err)
	}
	if err == nil {
		return
	}
	if reflect.TypeOf(legacyErr) != reflect.TypeOf(err) || legacyErr.Error() != err.Error() ||
		reflect.TypeOf(errors.Unwrap(legacyErr)) != reflect.TypeOf(errors.Unwrap(err)) {
		t.Fatalf("errors differ: (%#v, %#v)", legacyErr, err)
	}
}

func fuzzCollectionChain(depth int) *geom.GeometryCollection {
	current := geom.NewGeometryCollection().MustSetLayout(geom.XY).SetSRID(4326)
	for range depth - 1 {
		current = geom.NewGeometryCollection().MustPush(current).MustSetLayout(geom.XY).SetSRID(4326)
	}
	return current
}

func fuzzAdapterLimits() geo.Limits {
	limits := geo.DefaultLimits()
	limits.MaxPoints = 32
	limits.MaxGeometries = 32
	limits.MaxCollectionDepth = 32
	limits.MaxEncodedBytes = 4096
	return limits
}
