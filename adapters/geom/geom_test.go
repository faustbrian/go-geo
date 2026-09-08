package geogeom_test

import (
	"encoding/binary"
	"errors"
	"math"
	"testing"

	"github.com/twpayne/go-geom"
	geomwkb "github.com/twpayne/go-geom/encoding/ewkb"

	geo "github.com/faustbrian/go-geo"
	geogeom "github.com/faustbrian/go-geo/adapters/geom"
)

func TestToGoGeomRejectsLiteralNilWithExactEncodingError(t *testing.T) {
	t.Parallel()

	converted, err := geogeom.ToGoGeom(nil)
	if converted != nil {
		t.Fatalf("ToGoGeom(nil) = %#v, want nil", converted)
	}
	assertEncodingError(t, err, "cannot convert nil geometry", nil)
}

func TestToGoGeomPreservesTypedNilTopologyError(t *testing.T) {
	t.Parallel()

	var point *geo.Point
	_, err := geogeom.ToGoGeom(point)
	if !errors.Is(err, geo.ErrTopology) {
		t.Fatalf("ToGoGeom(typed nil) error = %v, want ErrTopology", err)
	}
}

func TestFromGoGeomValidatesInReleasedOrder(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value geom.T
		check func(*testing.T, error)
	}{
		{
			name: "literal nil",
			check: func(t *testing.T, err error) {
				assertEncodingError(t, err, "cannot convert nil geometry", nil)
			},
		},
		{
			name:  "typed nil",
			value: (*geom.Point)(nil),
			check: func(t *testing.T, err error) {
				assertEncodingError(t, err, "cannot convert nil geometry", nil)
			},
		},
		{
			name: "layout precedes SRID and coordinates",
			value: geom.NewLineStringFlat(
				geom.XYZ,
				[]float64{1},
			),
			check: func(t *testing.T, err error) {
				var typed *geo.UnsupportedError
				if !errors.As(err, &typed) {
					t.Fatalf("error type = %T, want *geo.UnsupportedError", err)
				}
				if typed.Operation != "geom conversion" || typed.Reason != "only the two-dimensional XY layout is supported" {
					t.Fatalf("error = %#v", typed)
				}
				if !errors.Is(err, geo.ErrUnsupported) || err.Error() != "geo: unsupported geom conversion: only the two-dimensional XY layout is supported" {
					t.Fatalf("error = %v", err)
				}
			},
		},
		{
			name:  "SRID precedes coordinates",
			value: geom.NewLineStringFlat(geom.XY, []float64{1}),
			check: func(t *testing.T, err error) {
				var typed *geo.CRSError
				if !errors.As(err, &typed) {
					t.Fatalf("error type = %T, want *geo.CRSError", err)
				}
				if typed.SRID != 0 || typed.Problem != "geom conversion requires a positive SRID" {
					t.Fatalf("error = %#v", typed)
				}
				if !errors.Is(err, geo.ErrCRS) || err.Error() != "geo: invalid CRS SRID 0: geom conversion requires a positive SRID" {
					t.Fatalf("error = %v", err)
				}
			},
		},
		{
			name:  "coordinates precede point limit",
			value: geom.NewLineStringFlat(geom.XY, []float64{1}).SetSRID(4326),
			check: func(t *testing.T, err error) {
				assertEncodingError(t, err, "geom has malformed XY coordinates", nil)
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			limits := geo.DefaultLimits()
			limits.MaxPoints = 0
			_, err := geogeom.FromGoGeom(test.value, limits)
			test.check(t, err)
		})
	}
}

func TestFromGoGeomAppliesInclusiveLimits(t *testing.T) {
	t.Parallel()

	value := geom.NewLineStringFlat(geom.XY, []float64{0, 0, 1, 1, 2, 2}).SetSRID(4326)
	limits := geo.DefaultLimits()
	limits.MaxPoints = 2
	_, err := geogeom.FromGoGeom(value, limits)
	var topology *geo.TopologyError
	if !errors.As(err, &topology) || topology.Geometry != "geom" || topology.Problem != "point limit exceeded" {
		t.Fatalf("point-limit error = %#v", err)
	}
	if !errors.Is(err, geo.ErrTopology) {
		t.Fatalf("point-limit error = %v, want ErrTopology", err)
	}

	limits.MaxPoints = 3
	converted, err := geogeom.FromGoGeom(value, limits)
	if err != nil {
		t.Fatalf("exact point limit: %v", err)
	}
	upstream, err := geogeom.ToGoGeom(converted)
	if err != nil {
		t.Fatalf("ToGoGeom(): %v", err)
	}
	if upstream.SRID() != 4326 || upstream.Layout() != geom.XY {
		t.Fatalf("round trip = layout %v, SRID %d", upstream.Layout(), upstream.SRID())
	}
}

func TestFromGoGeomPreservesMarshalByteAndDecodeErrors(t *testing.T) {
	t.Parallel()

	_, err := geogeom.FromGoGeom(fakeGeometry{}, geo.DefaultLimits())
	assertWrappedEncodingError(t, err, "cannot encode geom EWKB")

	limits := geo.DefaultLimits()
	limits.MaxEncodedBytes = 1
	_, err = geogeom.FromGoGeom(
		geom.NewPointFlat(geom.XY, []float64{0, 0}).SetSRID(4326),
		limits,
	)
	assertEncodingError(t, err, "encoded byte limit exceeded", nil)

	_, err = geogeom.FromGoGeom(
		geom.NewPointFlat(geom.XY, []float64{math.NaN(), 0}).SetSRID(4326),
		geo.DefaultLimits(),
	)
	if !errors.Is(err, geo.ErrRange) {
		t.Fatalf("non-finite coordinate error = %v, want ErrRange", err)
	}
}

func TestConversionsDoNotRetainMutableAliases(t *testing.T) {
	t.Parallel()

	flat := []float64{24.9384, 60.1699}
	upstream := geom.NewPointFlat(geom.XY, flat).SetSRID(4326)
	owned, err := geogeom.FromGoGeom(upstream, geo.DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	if owned == nil {
		t.Fatal("FromGoGeom() returned nil geometry")
	}
	flat[0] = 0
	upstream.FlatCoords()[1] = 0
	point := owned.(geo.Point)
	if point.Coordinate().Longitude().Degrees() != 24.9384 || point.Coordinate().Latitude().Degrees() != 60.1699 {
		t.Fatal("FromGoGeom retained mutable upstream storage")
	}

	returned, err := geogeom.ToGoGeom(owned)
	if err != nil {
		t.Fatal(err)
	}
	returned.FlatCoords()[0] = 1
	if point.Coordinate().Longitude().Degrees() != 24.9384 {
		t.Fatal("ToGoGeom exposed mutable owned storage")
	}
}

func TestEverySupportedGeometryFamilyAndEmptyAggregateRoundTrips(t *testing.T) {
	t.Parallel()

	collection := geom.NewGeometryCollection().MustPush(
		geom.NewPointFlat(geom.XY, []float64{24, 60}),
		geom.NewLineStringFlat(geom.XY, []float64{24, 60, 25, 61}),
	).MustSetLayout(geom.XY).SetSRID(4326)
	values := map[string]geom.T{
		"point":                   geom.NewPointFlat(geom.XY, []float64{24, 60}).SetSRID(4326),
		"line string":             geom.NewLineStringFlat(geom.XY, []float64{24, 60, 25, 61}).SetSRID(4326),
		"polygon":                 geom.NewPolygonFlat(geom.XY, []float64{0, 0, 2, 0, 2, 2, 0, 0}, []int{8}).SetSRID(4326),
		"multi point":             geom.NewMultiPointFlat(geom.XY, []float64{24, 60, 25, 61}).SetSRID(4326),
		"multi line string":       geom.NewMultiLineStringFlat(geom.XY, []float64{24, 60, 25, 61}, []int{4}).SetSRID(4326),
		"multi polygon":           geom.NewMultiPolygonFlat(geom.XY, []float64{0, 0, 2, 0, 2, 2, 0, 0}, [][]int{{8}}).SetSRID(4326),
		"geometry collection":     collection,
		"empty multi point":       geom.NewMultiPointFlat(geom.XY, nil).SetSRID(4326),
		"empty multi line string": geom.NewMultiLineStringFlat(geom.XY, nil, nil).SetSRID(4326),
		"empty multi polygon":     geom.NewMultiPolygonFlat(geom.XY, nil, nil).SetSRID(4326),
		"empty collection":        geom.NewGeometryCollection().MustSetLayout(geom.XY).SetSRID(4326),
	}

	for name, value := range values {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			owned, err := geogeom.FromGoGeom(value, geo.DefaultLimits())
			if err != nil {
				t.Fatalf("FromGoGeom(): %v", err)
			}
			upstream, err := geogeom.ToGoGeom(owned)
			if err != nil {
				t.Fatalf("ToGoGeom(): %v", err)
			}
			originalEWKB, err := geomwkb.Marshal(value, binary.LittleEndian)
			if err != nil {
				t.Fatalf("marshal original upstream value: %v", err)
			}
			roundTripEWKB, err := geomwkb.Marshal(upstream, binary.LittleEndian)
			if err != nil {
				t.Fatalf("marshal round-trip upstream value: %v", err)
			}
			if string(roundTripEWKB) != string(originalEWKB) {
				t.Fatalf("round-trip EWKB = %x, want original upstream EWKB %x", roundTripEWKB, originalEWKB)
			}
			again, err := geogeom.FromGoGeom(upstream, geo.DefaultLimits())
			if err != nil {
				t.Fatalf("second FromGoGeom(): %v", err)
			}
			if !geo.EqualGeometry(owned, again) || upstream.SRID() != 4326 || upstream.Layout() != geom.XY {
				t.Fatal("round trip changed geometry, layout, or SRID")
			}
		})
	}
}

func TestNumericAdversariesPreserveRootValidation(t *testing.T) {
	t.Parallel()

	for name, coordinates := range map[string][]float64{
		"NaN":               {math.NaN(), 0},
		"positive infinity": {math.Inf(1), 0},
		"negative infinity": {math.Inf(-1), 0},
		"extrema":           {math.MaxFloat64, -math.MaxFloat64},
		"signed zero":       {math.Copysign(0, -1), 0},
		"east antimeridian": {180, 0},
		"west antimeridian": {-180, 0},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			_, err := geogeom.FromGoGeom(
				geom.NewPointFlat(geom.XY, coordinates).SetSRID(4326),
				geo.DefaultLimits(),
			)
			switch name {
			case "signed zero", "east antimeridian", "west antimeridian":
				if err != nil {
					t.Fatalf("FromGoGeom(): %v", err)
				}
			default:
				if !errors.Is(err, geo.ErrRange) {
					t.Fatalf("error = %v, want ErrRange", err)
				}
			}
		})
	}
}

func assertEncodingError(t *testing.T, err error, problem string, cause error) {
	t.Helper()

	var typed *geo.EncodingError
	if !errors.As(err, &typed) {
		t.Fatalf("error type = %T, want *geo.EncodingError", err)
	}
	if typed.Format != "geom adapter" || typed.Problem != problem || typed.Cause != cause {
		t.Fatalf("error = %#v", typed)
	}
	if !errors.Is(err, geo.ErrEncoding) || typed.Unwrap() != cause {
		t.Fatalf("error classification/cause = %v/%v", err, typed.Unwrap())
	}
	want := "geo: invalid geom adapter encoding: " + problem
	if err.Error() != want {
		t.Fatalf("error = %q, want %q", err, want)
	}
}

func assertWrappedEncodingError(t *testing.T, err error, problem string) {
	t.Helper()
	var typed *geo.EncodingError
	if !errors.As(err, &typed) || typed.Format != "geom adapter" || typed.Problem != problem || typed.Cause == nil {
		t.Fatalf("error = %#v, want wrapped encoding error %q", err, problem)
	}
	if errors.Unwrap(err) != typed.Cause || !errors.Is(err, geo.ErrEncoding) {
		t.Fatalf("cause/category = (%v, %v)", errors.Unwrap(err), err)
	}
}

type fakeGeometry struct{}

func (fakeGeometry) Layout() geom.Layout   { return geom.XY }
func (fakeGeometry) Stride() int           { return 2 }
func (fakeGeometry) Bounds() *geom.Bounds  { return geom.NewBounds(geom.XY) }
func (fakeGeometry) FlatCoords() []float64 { return []float64{0, 0} }
func (fakeGeometry) Ends() []int           { return nil }
func (fakeGeometry) Endss() [][]int        { return nil }
func (fakeGeometry) SRID() int             { return 4326 }
func (fakeGeometry) Empty() bool           { return false }
