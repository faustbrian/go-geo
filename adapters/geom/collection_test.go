package geogeom_test

import (
	"errors"
	"testing"

	"github.com/twpayne/go-geom"

	geo "github.com/faustbrian/go-geo"
	geogeom "github.com/faustbrian/go-geo/adapters/geom"
)

func TestGeometryCollectionConvertsDirectNestedAndEmptyValues(t *testing.T) {
	t.Parallel()

	empty := geom.NewGeometryCollection().MustSetLayout(geom.XY).SetSRID(4326)
	nestedEmpty := geom.NewGeometryCollection().MustSetLayout(geom.NoLayout).SetSRID(4326)
	nested := geom.NewGeometryCollection().MustPush(
		geom.NewLineStringFlat(geom.XY, []float64{0, 0, 1, 1}).SetSRID(4326),
		nestedEmpty,
	).SetSRID(4326)
	direct := geom.NewGeometryCollection().MustPush(
		geom.NewPointFlat(geom.XY, []float64{24, 60}).SetSRID(4326),
		nested,
		geom.NewMultiPointFlat(geom.XY, nil).SetSRID(4326),
	).SetSRID(4326)

	for name, value := range map[string]*geom.GeometryCollection{
		"empty":  empty,
		"nested": direct,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			owned, err := geogeom.FromGoGeom(value, geo.DefaultLimits())
			if err != nil {
				t.Fatalf("FromGoGeom(): %v", err)
			}
			converted, ok := owned.(geo.GeometryCollection)
			if !ok {
				t.Fatalf("result type = %T, want geo.GeometryCollection", owned)
			}
			if converted.Len() != value.NumGeoms() {
				t.Fatalf("child count = %d, want %d", converted.Len(), value.NumGeoms())
			}
			roundTrip, err := geogeom.ToGoGeom(owned)
			if err != nil {
				t.Fatalf("ToGoGeom(): %v", err)
			}
			again, err := geogeom.FromGoGeom(roundTrip, geo.DefaultLimits())
			if err != nil || !geo.EqualGeometry(owned, again) {
				t.Fatalf("round trip = (%T, %v), want equal", again, err)
			}
		})
	}
}

func TestGeometryCollectionRejectsNilAndCyclesBeforeRecursiveMethods(t *testing.T) {
	t.Parallel()

	nilChild := geom.NewGeometryCollection().MustPush(
		geom.NewPointFlat(geom.XY, []float64{0, 0}).SetSRID(4326),
	).MustSetLayout(geom.XY).SetSRID(4326)
	nilChild.Geoms()[0] = (*geom.Point)(nil)
	assertCollectionEncodingError(t, nilChild, "geom collection contains nil geometry")

	selfCycle := geom.NewGeometryCollection().MustPush(
		geom.NewPointFlat(geom.XY, []float64{0, 0}).SetSRID(4326),
	).MustSetLayout(geom.XY).SetSRID(4326)
	selfCycle.Geoms()[0] = selfCycle
	assertCollectionEncodingError(t, selfCycle, "geom collection contains a cycle")

	outer := geom.NewGeometryCollection().MustPush(
		geom.NewPointFlat(geom.XY, []float64{0, 0}).SetSRID(4326),
	).MustSetLayout(geom.XY).SetSRID(4326)
	inner := geom.NewGeometryCollection().MustPush(outer).MustSetLayout(geom.XY).SetSRID(4326)
	outer.Geoms()[0] = inner
	assertCollectionEncodingError(t, outer, "geom collection contains a cycle")

	shared := geom.NewGeometryCollection().MustPush(
		geom.NewPointFlat(geom.XY, []float64{0, 0}).SetSRID(4326),
	).MustSetLayout(geom.XY).SetSRID(4326)
	acyclic := geom.NewGeometryCollection().MustPush(shared, shared).MustSetLayout(geom.XY).SetSRID(4326)
	if _, err := geogeom.FromGoGeom(acyclic, geo.DefaultLimits()); err != nil {
		t.Fatalf("shared acyclic child: %v", err)
	}
}

func TestGeometryCollectionAppliesGeometryDepthAndPointLimits(t *testing.T) {
	t.Parallel()

	direct := geom.NewGeometryCollection().MustPush(
		geom.NewPointFlat(geom.XY, []float64{0, 0}).SetSRID(4326),
		geom.NewLineStringFlat(geom.XY, []float64{0, 0, 1, 1}).SetSRID(4326),
	).MustSetLayout(geom.XY).SetSRID(4326)
	nested := geom.NewGeometryCollection().MustPush(
		geom.NewPointFlat(geom.XY, []float64{0, 0}).SetSRID(4326),
		geom.NewGeometryCollection().MustPush(
			geom.NewLineStringFlat(geom.XY, []float64{0, 0, 1, 1}).SetSRID(4326),
		).MustSetLayout(geom.XY).SetSRID(4326),
	).MustSetLayout(geom.XY).SetSRID(4326)

	for name, collection := range map[string]*geom.GeometryCollection{"direct": direct, "nested": nested} {
		for _, test := range []struct {
			name      string
			maxPoints int
			wantError bool
		}{
			{name: "one below", maxPoints: 4},
			{name: "exact", maxPoints: 3},
			{name: "one above", maxPoints: 2, wantError: true},
		} {
			t.Run(name+"/points/"+test.name, func(t *testing.T) {
				t.Parallel()
				limits := geo.DefaultLimits()
				limits.MaxPoints = test.maxPoints
				if test.wantError {
					assertCollectionTopologyError(t, collection, limits, "point limit exceeded")
					return
				}
				if _, err := geogeom.FromGoGeom(collection, limits); err != nil {
					t.Fatalf("FromGoGeom(): %v", err)
				}
			})
		}
	}

	for _, test := range []struct {
		name          string
		maxGeometries int
		wantError     bool
	}{
		{name: "one below", maxGeometries: 4},
		{name: "exact", maxGeometries: 3},
		{name: "one above", maxGeometries: 2, wantError: true},
	} {
		t.Run("geometries/"+test.name, func(t *testing.T) {
			t.Parallel()
			limits := geo.DefaultLimits()
			limits.MaxGeometries = test.maxGeometries
			if test.wantError {
				assertCollectionTopologyError(t, direct, limits, "geometry limit exceeded")
				return
			}
			if _, err := geogeom.FromGoGeom(direct, limits); err != nil {
				t.Fatalf("FromGoGeom(): %v", err)
			}
		})
	}

	exactSingleChildPoints := geom.NewGeometryCollection().MustPush(
		geom.NewLineStringFlat(geom.XY, []float64{0, 0, 1, 1, 2, 2}).SetSRID(4326),
	).MustSetLayout(geom.XY).SetSRID(4326)
	exactPointLimits := geo.DefaultLimits()
	exactPointLimits.MaxPoints = 3
	if _, err := geogeom.FromGoGeom(exactSingleChildPoints, exactPointLimits); err != nil {
		t.Fatalf("single child at exact point limit: %v", err)
	}

	exactRootDepth := geo.DefaultLimits()
	exactRootDepth.MaxCollectionDepth = 1
	if _, err := geogeom.FromGoGeom(
		geom.NewGeometryCollection().MustSetLayout(geom.XY).SetSRID(4326),
		exactRootDepth,
	); err != nil {
		t.Fatalf("root collection at exact depth limit: %v", err)
	}

	for _, depth := range []int{31, 32} {
		chain := collectionChain(depth)
		depthLimits := geo.DefaultLimits()
		depthLimits.MaxCollectionDepth = 32
		depthLimits.MaxGeometries = depth
		if _, err := geogeom.FromGoGeom(chain, depthLimits); err != nil {
			t.Fatalf("depth %d under fixed boundary: %v", depth, err)
		}
	}
	for _, requested := range []int{32, 1_000_000} {
		chain := collectionChain(33)
		depthLimits := geo.DefaultLimits()
		depthLimits.MaxCollectionDepth = requested
		depthLimits.MaxGeometries = 33
		assertCollectionTopologyError(t, chain, depthLimits, "collection depth limit exceeded")
	}

	zeroDefaults := geo.Limits{}
	if _, err := geogeom.FromGoGeom(nested, zeroDefaults); err != nil {
		t.Fatalf("zero geometry and depth limits did not resolve to defaults: %v", err)
	}
}

func TestGeometryCollectionStructuralWalkResumesAfterNestedChild(t *testing.T) {
	t.Parallel()

	nested := geom.NewGeometryCollection().MustSetLayout(geom.XY).SetSRID(4326)
	collection := geom.NewGeometryCollection().MustPush(
		nested,
		geom.NewPointFlat(geom.XY, []float64{0, 0}).SetSRID(4326),
	).MustSetLayout(geom.XY).SetSRID(4326)
	collection.Geoms()[1] = (*geom.Point)(nil)

	assertCollectionEncodingError(t, collection, "geom collection contains nil geometry")
}

func TestGeometryCollectionStructuralFailuresPrecedeLayoutAndSRID(t *testing.T) {
	t.Parallel()

	collection := geom.NewGeometryCollection().MustPush(
		geom.NewPointFlat(geom.XY, []float64{0, 0}).SetSRID(4326),
	).SetSRID(0)
	collection.Geoms()[0] = (*geom.Point)(nil)
	assertCollectionEncodingError(t, collection, "geom collection contains nil geometry")

	noLayout := geom.NewGeometryCollection().SetSRID(4326)
	_, err := geogeom.FromGoGeom(noLayout, geo.DefaultLimits())
	if !errors.Is(err, geo.ErrUnsupported) {
		t.Fatalf("NoLayout error = %v, want ErrUnsupported", err)
	}

	missingSRID := geom.NewGeometryCollection().MustSetLayout(geom.XY)
	_, err = geogeom.FromGoGeom(missingSRID, geo.DefaultLimits())
	if !errors.Is(err, geo.ErrCRS) {
		t.Fatalf("missing SRID error = %v, want ErrCRS", err)
	}

	limits := geo.DefaultLimits()
	limits.MaxGeometries = -1
	limits.MaxCollectionDepth = -1
	assertCollectionTopologyError(t, geom.NewGeometryCollection(), limits, "geometry limit exceeded")
	limits.MaxGeometries = 1
	assertCollectionTopologyError(t, geom.NewGeometryCollection(), limits, "collection depth limit exceeded")
}

func TestGeometryCollectionRejectsDeepNoLayoutBeforeRecursiveLayout(t *testing.T) {
	t.Parallel()

	current := geom.NewGeometryCollection().SetSRID(4326)
	for range 32 {
		current = geom.NewGeometryCollection().MustPush(current).SetSRID(4326)
	}
	limits := geo.DefaultLimits()
	limits.MaxGeometries = 33
	limits.MaxCollectionDepth = 32
	assertCollectionTopologyError(t, current, limits, "collection depth limit exceeded")
}

func TestGeometryCollectionPreservesChildOrderAndOwnsResults(t *testing.T) {
	t.Parallel()

	collection := geom.NewGeometryCollection().MustPush(
		geom.NewPointFlat(geom.XY, []float64{24, 60}).SetSRID(4326),
		geom.NewLineStringFlat(geom.XY, []float64{0, 0, 1, 1}).SetSRID(4326),
	).MustSetLayout(geom.XY).SetSRID(4326)
	owned, err := geogeom.FromGoGeom(collection, geo.DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	result := owned.(geo.GeometryCollection)
	firstGeometry, ok := result.At(0)
	if !ok {
		t.Fatal("child 0 is missing")
	}
	if _, ok := firstGeometry.(geo.Point); !ok {
		t.Fatalf("child 0 type = %T, want geo.Point", firstGeometry)
	}
	secondGeometry, ok := result.At(1)
	if !ok {
		t.Fatal("child 1 is missing")
	}
	if _, ok := secondGeometry.(geo.LineString); !ok {
		t.Fatalf("child 1 type = %T, want geo.LineString", secondGeometry)
	}

	collection.Geoms()[0] = geom.NewPointFlat(geom.XY, []float64{1, 1}).SetSRID(4326)
	first := firstGeometry.(geo.Point).Coordinate()
	if first.Longitude().Degrees() != 24 || first.Latitude().Degrees() != 60 {
		t.Fatal("result retained collection or child aliases")
	}
}

func TestGeometryCollectionMatchesIndependentNestedValue(t *testing.T) {
	t.Parallel()

	input := geom.NewGeometryCollection().MustPush(
		geom.NewPointFlat(geom.XY, []float64{24, 60}).SetSRID(4326),
		geom.NewGeometryCollection().MustPush(
			geom.NewLineStringFlat(geom.XY, []float64{0, 0, 1, 1}).SetSRID(4326),
			geom.NewGeometryCollection().MustSetLayout(geom.NoLayout).SetSRID(4326),
		).SetSRID(4326),
		geom.NewMultiPointFlat(geom.XY, nil).SetSRID(4326),
	).SetSRID(4326)

	first := mustCoordinate(t, 24, 60)
	lineFirst := mustCoordinate(t, 0, 0)
	lineSecond := mustCoordinate(t, 1, 1)
	point, err := geo.NewPoint(first)
	if err != nil {
		t.Fatal(err)
	}
	line, err := geo.NewLineString([]geo.Coordinate{lineFirst, lineSecond})
	if err != nil {
		t.Fatal(err)
	}
	empty, err := geo.NewGeometryCollection(nil, geo.WGS84())
	if err != nil {
		t.Fatal(err)
	}
	nested, err := geo.NewGeometryCollection([]geo.Geometry{line, empty}, geo.WGS84())
	if err != nil {
		t.Fatal(err)
	}
	emptyMultiPoint, err := geo.NewMultiPoint(nil, geo.WGS84())
	if err != nil {
		t.Fatal(err)
	}
	expected, err := geo.NewGeometryCollection([]geo.Geometry{point, nested, emptyMultiPoint}, geo.WGS84())
	if err != nil {
		t.Fatal(err)
	}

	converted, err := geogeom.FromGoGeom(input, geo.DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	if !geo.EqualGeometry(converted, expected) {
		t.Fatalf("converted = %#v, want %#v", converted, expected)
	}
}

func mustCoordinate(t *testing.T, longitude, latitude float64) geo.Coordinate {
	t.Helper()
	lon, err := geo.NewLongitude(longitude)
	if err != nil {
		t.Fatal(err)
	}
	lat, err := geo.NewLatitude(latitude)
	if err != nil {
		t.Fatal(err)
	}
	coordinate, err := geo.NewCoordinate(lon, lat, geo.WGS84())
	if err != nil {
		t.Fatal(err)
	}
	return coordinate
}

func TestGeometryCollectionSemanticWalkIsLeftToRight(t *testing.T) {
	t.Parallel()

	odd := geom.NewLineStringFlat(geom.XY, []float64{0}).SetSRID(4326)
	over := geom.NewLineStringFlat(geom.XY, []float64{0, 0, 1, 1}).SetSRID(4326)
	collection := geom.NewGeometryCollection().MustPush(odd, over).MustSetLayout(geom.XY).SetSRID(4326)
	limits := geo.DefaultLimits()
	limits.MaxPoints = 1
	_, err := geogeom.FromGoGeom(collection, limits)
	assertEncodingError(t, err, "geom has malformed XY coordinates", nil)
}

func TestGeometryCollectionSemanticWalkResumesAcrossNestedChildren(t *testing.T) {
	t.Parallel()

	nested := geom.NewGeometryCollection().MustPush(
		geom.NewPointFlat(geom.XY, []float64{0, 0}).SetSRID(4326),
	).MustSetLayout(geom.XY).SetSRID(4326)
	odd := geom.NewLineStringFlat(geom.XY, []float64{0}).SetSRID(4326)
	collection := geom.NewGeometryCollection().MustPush(nested, odd).MustSetLayout(geom.XY).SetSRID(4326)

	_, err := geogeom.FromGoGeom(collection, geo.DefaultLimits())
	assertEncodingError(t, err, "geom has malformed XY coordinates", nil)
}

func TestGeometryCollectionCumulativePointLimitPrecedesMarshal(t *testing.T) {
	t.Parallel()

	collection := geom.NewGeometryCollection().MustPush(
		geom.NewPointFlat(geom.XY, []float64{0, 0}).SetSRID(4326),
		geom.NewLineStringFlat(geom.XY, []float64{1, 1, 2, 2}).SetSRID(4326),
		geom.NewLineStringFlat(geom.XY, []float64{3, 3, 4, 4}).SetSRID(4326),
		fakeGeometry{},
	).MustSetLayout(geom.XY).SetSRID(4326)
	limits := geo.DefaultLimits()
	limits.MaxPoints = 4

	assertCollectionTopologyError(t, collection, limits, "point limit exceeded")
}

func collectionChain(depth int) *geom.GeometryCollection {
	current := geom.NewGeometryCollection().MustSetLayout(geom.XY).SetSRID(4326)
	for range depth - 1 {
		current = geom.NewGeometryCollection().MustPush(current).MustSetLayout(geom.XY).SetSRID(4326)
	}
	return current
}

func assertCollectionEncodingError(t *testing.T, value *geom.GeometryCollection, problem string) {
	t.Helper()
	_, err := geogeom.FromGoGeom(value, geo.DefaultLimits())
	assertEncodingError(t, err, problem, nil)
}

func assertCollectionTopologyError(t *testing.T, value *geom.GeometryCollection, limits geo.Limits, problem string) {
	t.Helper()
	_, err := geogeom.FromGoGeom(value, limits)
	var typed *geo.TopologyError
	if !errors.As(err, &typed) || typed.Geometry != "geom" || typed.Problem != problem {
		t.Fatalf("error = %#v, want geom topology error %q", err, problem)
	}
	if !errors.Is(err, geo.ErrTopology) {
		t.Fatalf("error = %v, want ErrTopology", err)
	}
}
