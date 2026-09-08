package gogeom_test

import (
	"errors"
	"math"
	"reflect"
	"sync"
	"testing"

	"github.com/twpayne/go-geom"

	geo "github.com/faustbrian/go-geo"
	"github.com/faustbrian/go-geo/adapter/gogeom"
	geogeom "github.com/faustbrian/go-geo/adapters/geom"
)

func TestLegacyAndSuccessorErrorsHaveExactParity(t *testing.T) {
	t.Parallel()

	fromCases := []struct {
		name   string
		value  geom.T
		limits geo.Limits
	}{
		{name: "literal nil", limits: geo.DefaultLimits()},
		{name: "typed nil", value: (*geom.Point)(nil), limits: geo.DefaultLimits()},
		{name: "layout", value: geom.NewPointFlat(geom.XYZ, []float64{1, 2, 3}).SetSRID(4326), limits: geo.DefaultLimits()},
		{name: "XYM layout", value: geom.NewPointFlat(geom.XYM, []float64{1, 2, 3}).SetSRID(4326), limits: geo.DefaultLimits()},
		{name: "XYZM layout", value: geom.NewPointFlat(geom.XYZM, []float64{1, 2, 3, 4}).SetSRID(4326), limits: geo.DefaultLimits()},
		{name: "negative SRID", value: geom.NewPointFlat(geom.XY, []float64{1, 2}).SetSRID(-1), limits: geo.DefaultLimits()},
		{name: "malformed coordinates", value: geom.NewLineStringFlat(geom.XY, []float64{1}).SetSRID(4326), limits: geo.DefaultLimits()},
		{name: "marshal failure", value: facadeFakeGeometry{}, limits: geo.DefaultLimits()},
		{name: "encoded limit", value: geom.NewPointFlat(geom.XY, []float64{1, 2}).SetSRID(4326), limits: geo.Limits{MaxPoints: 1, MaxRings: 1, MaxGeometries: 1, MaxCollectionDepth: 1, MaxEncodedBytes: 1}},
	}
	pointLimited := geom.NewLineStringFlat(geom.XY, []float64{0, 0, 1, 1}).SetSRID(4326)
	pointLimits := geo.DefaultLimits()
	pointLimits.MaxPoints = 1
	fromCases = append(fromCases, struct {
		name   string
		value  geom.T
		limits geo.Limits
	}{name: "point limit", value: pointLimited, limits: pointLimits})

	for _, test := range fromCases {
		t.Run("from/"+test.name, func(t *testing.T) {
			t.Parallel()
			_, oldErr := gogeom.FromGoGeom(test.value, test.limits)
			_, newErr := geogeom.FromGoGeom(test.value, test.limits)
			assertErrorsEquivalent(t, oldErr, newErr)
		})
	}

	_, oldErr := gogeom.ToGoGeom(nil)
	_, newErr := geogeom.ToGoGeom(nil)
	assertErrorsEquivalent(t, oldErr, newErr)
	var nilPoint *geo.Point
	_, oldErr = gogeom.ToGoGeom(nilPoint)
	_, newErr = geogeom.ToGoGeom(nilPoint)
	assertErrorsEquivalent(t, oldErr, newErr)
}

func TestLegacyAndSuccessorResultsHaveExactParity(t *testing.T) {
	t.Parallel()

	value := geom.NewLineStringFlat(geom.XY, []float64{0, 0, 1, 1}).SetSRID(4326)
	oldOwned, err := gogeom.FromGoGeom(value, geo.DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	newOwned, err := geogeom.FromGoGeom(value, geo.DefaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	if !geo.EqualGeometry(oldOwned, newOwned) {
		t.Fatal("old and new FromGoGeom results differ")
	}

	oldValue, err := gogeom.ToGoGeom(oldOwned)
	if err != nil {
		t.Fatal(err)
	}
	newValue, err := geogeom.ToGoGeom(oldOwned)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(oldValue, newValue) {
		t.Fatal("old and new ToGoGeom results differ")
	}
}

func TestLegacyAndSuccessorOwnNestedCollectionInputs(t *testing.T) {
	t.Parallel()

	point := geom.NewPointFlat(geom.XY, []float64{24, 60}).SetSRID(4326)
	nested := geom.NewGeometryCollection().MustPush(point).MustSetLayout(geom.XY).SetSRID(4326)
	outer := geom.NewGeometryCollection().MustPush(nested).MustSetLayout(geom.XY).SetSRID(4326)

	oldOwned, oldErr := gogeom.FromGoGeom(outer, geo.DefaultLimits())
	newOwned, newErr := geogeom.FromGoGeom(outer, geo.DefaultLimits())
	if oldErr != nil || newErr != nil {
		t.Fatalf("collection conversion errors = (%v, %v)", oldErr, newErr)
	}

	expectedPoint := mustPoint(t, 24, 60)
	expectedNested, err := geo.NewGeometryCollection([]geo.Geometry{expectedPoint}, geo.WGS84())
	if err != nil {
		t.Fatal(err)
	}
	expected, err := geo.NewGeometryCollection([]geo.Geometry{expectedNested}, geo.WGS84())
	if err != nil {
		t.Fatal(err)
	}

	point.FlatCoords()[0] = 1
	nested.Geoms()[0] = geom.NewPointFlat(geom.XY, []float64{2, 3}).SetSRID(4326)
	outer.Geoms()[0] = geom.NewPointFlat(geom.XY, []float64{4, 5}).SetSRID(4326)

	if !geo.EqualGeometry(oldOwned, expected) {
		t.Fatal("legacy result aliases nested collection input")
	}
	if !geo.EqualGeometry(newOwned, expected) {
		t.Fatal("successor result aliases nested collection input")
	}
	if !geo.EqualGeometry(oldOwned, newOwned) {
		t.Fatal("legacy and successor owned results differ")
	}
}

func TestLegacyAndSuccessorCollectionBranchesHaveExactParity(t *testing.T) {
	t.Parallel()

	success := geom.NewGeometryCollection().MustPush(
		geom.NewPointFlat(geom.XY, []float64{0, 0}).SetSRID(4326),
	).MustSetLayout(geom.XY).SetSRID(4326)
	oldOwned, oldErr := gogeom.FromGoGeom(success, geo.DefaultLimits())
	newOwned, newErr := geogeom.FromGoGeom(success, geo.DefaultLimits())
	if oldErr != nil || newErr != nil || !geo.EqualGeometry(oldOwned, newOwned) {
		t.Fatalf("collection results = (%v, %v, %v)", oldErr, newErr, geo.EqualGeometry(oldOwned, newOwned))
	}

	nilChild := geom.NewGeometryCollection().MustPush(
		geom.NewPointFlat(geom.XY, []float64{0, 0}).SetSRID(4326),
	).MustSetLayout(geom.XY).SetSRID(4326)
	nilChild.Geoms()[0] = (*geom.Point)(nil)
	cycle := geom.NewGeometryCollection().MustPush(
		geom.NewPointFlat(geom.XY, []float64{0, 0}).SetSRID(4326),
	).MustSetLayout(geom.XY).SetSRID(4326)
	cycle.Geoms()[0] = cycle
	indirectOuter := geom.NewGeometryCollection().MustPush(
		geom.NewPointFlat(geom.XY, []float64{0, 0}).SetSRID(4326),
	).MustSetLayout(geom.XY).SetSRID(4326)
	indirectInner := geom.NewGeometryCollection().MustPush(indirectOuter).MustSetLayout(geom.XY).SetSRID(4326)
	indirectOuter.Geoms()[0] = indirectInner
	shared := geom.NewGeometryCollection().MustPush(
		geom.NewPointFlat(geom.XY, []float64{0, 0}).SetSRID(4326),
	).MustSetLayout(geom.XY).SetSRID(4326)
	sharedOuter := geom.NewGeometryCollection().MustPush(shared, shared).MustSetLayout(geom.XY).SetSRID(4326)
	oldShared, oldSharedErr := gogeom.FromGoGeom(sharedOuter, geo.DefaultLimits())
	newShared, newSharedErr := geogeom.FromGoGeom(sharedOuter, geo.DefaultLimits())
	if oldSharedErr != nil || newSharedErr != nil || !geo.EqualGeometry(oldShared, newShared) {
		t.Fatalf("shared collection results = (%v, %v, %v)", oldSharedErr, newSharedErr, geo.EqualGeometry(oldShared, newShared))
	}
	pointLimited := geom.NewGeometryCollection().MustPush(
		geom.NewLineStringFlat(geom.XY, []float64{0, 0, 1, 1}).SetSRID(4326),
	).MustSetLayout(geom.XY).SetSRID(4326)
	pointLimits := geo.DefaultLimits()
	pointLimits.MaxPoints = 1
	geometryLimited := geom.NewGeometryCollection().MustPush(
		geom.NewPointFlat(geom.XY, []float64{0, 0}).SetSRID(4326),
	).MustSetLayout(geom.XY).SetSRID(4326)
	geometryLimits := geo.DefaultLimits()
	geometryLimits.MaxGeometries = 1
	depthLimited := facadeCollectionChain(33)
	depthLimits := geo.DefaultLimits()
	depthLimits.MaxGeometries = 33
	depthLimits.MaxCollectionDepth = 1_000_000
	structuralPrecedence := geom.NewGeometryCollection().MustPush(
		geom.NewPointFlat(geom.XY, []float64{0, 0}).SetSRID(4326),
	).SetSRID(0)
	structuralPrecedence.Geoms()[0] = (*geom.Point)(nil)
	malformed := geom.NewGeometryCollection().MustPush(
		geom.NewLineStringFlat(geom.XY, []float64{0}).SetSRID(4326),
	).MustSetLayout(geom.XY).SetSRID(4326)
	marshalFailure := geom.NewGeometryCollection().MustPush(facadeFakeGeometry{}).MustSetLayout(geom.XY).SetSRID(4326)
	encodedLimited := geom.NewGeometryCollection().MustPush(
		geom.NewPointFlat(geom.XY, []float64{0, 0}).SetSRID(4326),
	).MustSetLayout(geom.XY).SetSRID(4326)
	encodedLimits := geo.DefaultLimits()
	encodedLimits.MaxEncodedBytes = 1
	decodeFailure := geom.NewGeometryCollection().MustPush(
		geom.NewPointFlat(geom.XY, []float64{math.NaN(), 0}).SetSRID(4326),
	).MustSetLayout(geom.XY).SetSRID(4326)
	noLayout := geom.NewGeometryCollection().SetSRID(4326)
	missingSRID := geom.NewGeometryCollection().MustSetLayout(geom.XY)
	for _, test := range []struct {
		name   string
		value  *geom.GeometryCollection
		limits geo.Limits
	}{
		{name: "nil child", value: nilChild, limits: geo.DefaultLimits()},
		{name: "cycle", value: cycle, limits: geo.DefaultLimits()},
		{name: "indirect cycle", value: indirectOuter, limits: geo.DefaultLimits()},
		{name: "point limit", value: pointLimited, limits: pointLimits},
		{name: "geometry limit", value: geometryLimited, limits: geometryLimits},
		{name: "fixed depth cap", value: depthLimited, limits: depthLimits},
		{name: "structural precedence", value: structuralPrecedence, limits: geo.DefaultLimits()},
		{name: "malformed descendant", value: malformed, limits: geo.DefaultLimits()},
		{name: "marshal failure", value: marshalFailure, limits: geo.DefaultLimits()},
		{name: "encoded byte limit", value: encodedLimited, limits: encodedLimits},
		{name: "decode failure", value: decodeFailure, limits: geo.DefaultLimits()},
		{name: "layout", value: noLayout, limits: geo.DefaultLimits()},
		{name: "SRID", value: missingSRID, limits: geo.DefaultLimits()},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			_, oldErr := gogeom.FromGoGeom(test.value, test.limits)
			_, newErr := geogeom.FromGoGeom(test.value, test.limits)
			assertErrorsEquivalent(t, oldErr, newErr)
		})
	}
}

func TestLegacyAndSuccessorConvertEmptyAndNestedEmptyCollections(t *testing.T) {
	t.Parallel()

	empty := geom.NewGeometryCollection().MustSetLayout(geom.XY).SetSRID(4326)
	nestedEmpty := geom.NewGeometryCollection().MustSetLayout(geom.NoLayout).SetSRID(4326)
	nested := geom.NewGeometryCollection().MustPush(
		geom.NewLineStringFlat(geom.XY, []float64{0, 0, 1, 1}).SetSRID(4326),
		nestedEmpty,
	).SetSRID(4326)
	for name, value := range map[string]*geom.GeometryCollection{
		"explicit XY empty": empty,
		"nested empty":      nested,
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			legacyOwned, legacyErr := gogeom.FromGoGeom(value, geo.DefaultLimits())
			canonicalOwned, canonicalErr := geogeom.FromGoGeom(value, geo.DefaultLimits())
			if legacyErr != nil || canonicalErr != nil || !geo.EqualGeometry(legacyOwned, canonicalOwned) {
				t.Fatalf("results = (%#v, %v, %#v, %v)", legacyOwned, legacyErr, canonicalOwned, canonicalErr)
			}
		})
	}
}

func TestLegacyAndSuccessorReturnedCollectionsDoNotAliasSourceOrEachOther(t *testing.T) {
	t.Parallel()

	point := mustPoint(t, 24, 60)
	nested, err := geo.NewGeometryCollection([]geo.Geometry{point}, geo.WGS84())
	if err != nil {
		t.Fatal(err)
	}
	source, err := geo.NewGeometryCollection([]geo.Geometry{nested}, geo.WGS84())
	if err != nil {
		t.Fatal(err)
	}

	legacyValue, legacyErr := gogeom.ToGoGeom(source)
	canonicalValue, canonicalErr := geogeom.ToGoGeom(source)
	if legacyErr != nil || canonicalErr != nil {
		t.Fatalf("ToGoGeom() errors = (%v, %v)", legacyErr, canonicalErr)
	}
	legacyCollection := legacyValue.(*geom.GeometryCollection)
	canonicalCollection := canonicalValue.(*geom.GeometryCollection)
	legacyNested := legacyCollection.Geom(0).(*geom.GeometryCollection)
	legacyNested.Geom(0).FlatCoords()[0] = 1
	legacyNested.Geoms()[0] = geom.NewPointFlat(geom.XY, []float64{2, 3}).SetSRID(4326)
	legacyCollection.Geoms()[0] = geom.NewPointFlat(geom.XY, []float64{4, 5}).SetSRID(4326)

	canonicalNested := canonicalCollection.Geom(0).(*geom.GeometryCollection)
	canonicalPoint := canonicalNested.Geom(0).(*geom.Point)
	if canonicalPoint.FlatCoords()[0] != 24 || canonicalPoint.FlatCoords()[1] != 60 {
		t.Fatal("legacy result mutation changed independent canonical result")
	}
	sourceNested, ok := source.At(0)
	if !ok {
		t.Fatal("source nested collection is missing")
	}
	sourcePoint, ok := sourceNested.(geo.GeometryCollection).At(0)
	if !ok || !geo.EqualGeometry(sourcePoint, point) {
		t.Fatal("returned collection mutation changed immutable source")
	}

	canonicalPoint.FlatCoords()[0] = 6
	canonicalNested.Geoms()[0] = geom.NewPointFlat(geom.XY, []float64{7, 8}).SetSRID(4326)
	canonicalCollection.Geoms()[0] = geom.NewPointFlat(geom.XY, []float64{9, 10}).SetSRID(4326)
	sourceNested, ok = source.At(0)
	if !ok {
		t.Fatal("source nested collection is missing after canonical mutation")
	}
	sourcePoint, ok = sourceNested.(geo.GeometryCollection).At(0)
	if !ok || !geo.EqualGeometry(sourcePoint, point) {
		t.Fatal("canonical result mutation changed immutable source")
	}
}

func TestLegacyAndSuccessorPreserveCollectionFaultPrecedence(t *testing.T) {
	t.Parallel()

	point := geom.NewPointFlat(geom.XY, []float64{0, 0}).SetSRID(4326)
	nilChild := geom.NewGeometryCollection().MustPush(point).SetSRID(0)
	nilChild.Geoms()[0] = (*geom.Point)(nil)
	selfCycle := geom.NewGeometryCollection().MustPush(point).SetSRID(0)
	selfCycle.Geoms()[0] = selfCycle
	odd := geom.NewLineStringFlat(geom.XY, []float64{0}).SetSRID(4326)
	over := geom.NewLineStringFlat(geom.XY, []float64{0, 0, 1, 1}).SetSRID(4326)
	invalidLayout := geom.NewGeometryCollection().MustPush(facadeXYZOddGeometry{}).SetSRID(0)
	invalidSRID := geom.NewGeometryCollection().MustPush(odd, facadeFakeGeometry{}).MustSetLayout(geom.XY)
	semanticFaults := geom.NewGeometryCollection().MustPush(odd, over, facadeFakeGeometry{}).MustSetLayout(geom.XY).SetSRID(4326)
	pointBeforeMarshal := geom.NewGeometryCollection().MustPush(over, facadeFakeGeometry{}).MustSetLayout(geom.XY).SetSRID(4326)
	marshalBeforeBytes := geom.NewGeometryCollection().MustPush(facadeFakeGeometry{}).MustSetLayout(geom.XY).SetSRID(4326)
	bytesBeforeDecode := geom.NewGeometryCollection().MustPush(
		geom.NewPointFlat(geom.XY, []float64{math.NaN(), 0}).SetSRID(4326),
	).MustSetLayout(geom.XY).SetSRID(4326)

	defaultLimits := geo.DefaultLimits()
	nilLimits := defaultLimits
	nilLimits.MaxGeometries = 1
	geometryLimits := defaultLimits
	geometryLimits.MaxGeometries = 1
	geometryLimits.MaxCollectionDepth = 1
	depthLimits := defaultLimits
	depthLimits.MaxGeometries = 2
	depthLimits.MaxCollectionDepth = 1
	semanticLimits := defaultLimits
	semanticLimits.MaxPoints = 1
	semanticLimits.MaxEncodedBytes = 1
	byteLimits := defaultLimits
	byteLimits.MaxEncodedBytes = 1

	for _, test := range []struct {
		name     string
		value    *geom.GeometryCollection
		limits   geo.Limits
		category error
		problem  string
	}{
		{name: "nil before geometry layout and SRID", value: nilChild, limits: nilLimits, category: geo.ErrEncoding, problem: "geom collection contains nil geometry"},
		{name: "geometry before depth cycle layout and SRID", value: selfCycle, limits: geometryLimits, category: geo.ErrTopology, problem: "geometry limit exceeded"},
		{name: "depth before cycle layout and SRID", value: selfCycle, limits: depthLimits, category: geo.ErrTopology, problem: "collection depth limit exceeded"},
		{name: "cycle before layout and SRID", value: selfCycle, limits: defaultLimits, category: geo.ErrEncoding, problem: "geom collection contains a cycle"},
		{name: "layout before SRID semantic and marshal", value: invalidLayout, limits: semanticLimits, category: geo.ErrUnsupported, problem: "only the two-dimensional XY layout is supported"},
		{name: "SRID before semantic and marshal", value: invalidSRID, limits: semanticLimits, category: geo.ErrCRS, problem: "geom conversion requires a positive SRID"},
		{name: "malformed coordinates before point marshal and bytes", value: semanticFaults, limits: semanticLimits, category: geo.ErrEncoding, problem: "geom has malformed XY coordinates"},
		{name: "point limit before marshal and bytes", value: pointBeforeMarshal, limits: semanticLimits, category: geo.ErrTopology, problem: "point limit exceeded"},
		{name: "marshal before encoded byte limit", value: marshalBeforeBytes, limits: byteLimits, category: geo.ErrEncoding, problem: "cannot encode geom EWKB"},
		{name: "encoded byte limit before decode", value: bytesBeforeDecode, limits: byteLimits, category: geo.ErrEncoding, problem: "encoded byte limit exceeded"},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			_, legacyErr := gogeom.FromGoGeom(test.value, test.limits)
			_, canonicalErr := geogeom.FromGoGeom(test.value, test.limits)
			assertErrorsEquivalent(t, legacyErr, canonicalErr)
			assertCollectionFault(t, legacyErr, test.category, test.problem)
			assertCollectionFault(t, canonicalErr, test.category, test.problem)
		})
	}
}

func TestLegacyAndSuccessorCollectionParityUnderConcurrency(t *testing.T) {
	t.Parallel()

	collection := geom.NewGeometryCollection().MustPush(
		geom.NewPointFlat(geom.XY, []float64{24, 60}).SetSRID(4326),
		geom.NewLineStringFlat(geom.XY, []float64{0, 0, 1, 1}).SetSRID(4326),
	).MustSetLayout(geom.XY).SetSRID(4326)
	var wait sync.WaitGroup
	for range 16 {
		wait.Add(1)
		go func() {
			defer wait.Done()
			for range 50 {
				oldOwned, oldErr := gogeom.FromGoGeom(collection, geo.DefaultLimits())
				newOwned, newErr := geogeom.FromGoGeom(collection, geo.DefaultLimits())
				if oldErr != nil || newErr != nil || !geo.EqualGeometry(oldOwned, newOwned) {
					t.Errorf("concurrent results = (%v, %v, %v)", oldErr, newErr, geo.EqualGeometry(oldOwned, newOwned))
					return
				}
			}
		}()
	}
	wait.Wait()
}

func facadeCollectionChain(depth int) *geom.GeometryCollection {
	current := geom.NewGeometryCollection().MustSetLayout(geom.XY).SetSRID(4326)
	for range depth - 1 {
		current = geom.NewGeometryCollection().MustPush(current).MustSetLayout(geom.XY).SetSRID(4326)
	}
	return current
}

func assertErrorsEquivalent(t *testing.T, oldErr, newErr error) {
	t.Helper()

	if oldErr == nil || newErr == nil {
		t.Fatalf("errors = (%v, %v), want two errors", oldErr, newErr)
	}
	if reflect.TypeOf(oldErr) != reflect.TypeOf(newErr) || oldErr.Error() != newErr.Error() {
		t.Fatalf("errors differ: (%T %q, %T %q)", oldErr, oldErr, newErr, newErr)
	}
	assertErrorFieldsEquivalent(t, oldErr, newErr)
	oldCause, newCause := errors.Unwrap(oldErr), errors.Unwrap(newErr)
	if reflect.TypeOf(oldCause) != reflect.TypeOf(newCause) || errorText(oldCause) != errorText(newCause) {
		t.Fatalf("causes differ: (%T %v, %T %v)", oldCause, oldCause, newCause, newCause)
	}
	var encoding *geo.EncodingError
	if errors.As(oldErr, &encoding) && encoding.Problem == "cannot encode geom EWKB" && oldCause != newCause {
		t.Fatalf("cause identity differs: (%p, %p)", oldCause, newCause)
	}
}

func assertErrorFieldsEquivalent(t *testing.T, oldErr, newErr error) {
	t.Helper()
	switch oldTyped := oldErr.(type) {
	case *geo.EncodingError:
		newTyped, ok := newErr.(*geo.EncodingError)
		if !ok || oldTyped.Format != newTyped.Format || oldTyped.Problem != newTyped.Problem {
			t.Fatalf("encoding error fields differ: (%#v, %#v)", oldErr, newErr)
		}
	case *geo.TopologyError:
		newTyped, ok := newErr.(*geo.TopologyError)
		if !ok || *oldTyped != *newTyped {
			t.Fatalf("topology error fields differ: (%#v, %#v)", oldErr, newErr)
		}
	case *geo.UnsupportedError:
		newTyped, ok := newErr.(*geo.UnsupportedError)
		if !ok || *oldTyped != *newTyped {
			t.Fatalf("unsupported error fields differ: (%#v, %#v)", oldErr, newErr)
		}
	case *geo.CRSError:
		newTyped, ok := newErr.(*geo.CRSError)
		if !ok || *oldTyped != *newTyped {
			t.Fatalf("CRS error fields differ: (%#v, %#v)", oldErr, newErr)
		}
	default:
		if !reflect.DeepEqual(oldErr, newErr) {
			t.Fatalf("error fields differ: (%#v, %#v)", oldErr, newErr)
		}
	}
}

func assertCollectionFault(t *testing.T, err, category error, problem string) {
	t.Helper()
	if !errors.Is(err, category) {
		t.Fatalf("error = %v, want category %v", err, category)
	}
	switch category {
	case geo.ErrEncoding:
		var typed *geo.EncodingError
		if !errors.As(err, &typed) || typed.Format != "geom adapter" || typed.Problem != problem {
			t.Fatalf("error = %#v, want encoding problem %q", err, problem)
		}
		if problem == "cannot encode geom EWKB" && typed.Cause == nil {
			t.Fatal("marshal encoding error has no cause")
		}
		if problem != "cannot encode geom EWKB" && typed.Cause != nil {
			t.Fatalf("unexpected encoding cause = %v", typed.Cause)
		}
	case geo.ErrTopology:
		var typed *geo.TopologyError
		if !errors.As(err, &typed) || typed.Geometry != "geom" || typed.Problem != problem {
			t.Fatalf("error = %#v, want topology problem %q", err, problem)
		}
	case geo.ErrUnsupported:
		var typed *geo.UnsupportedError
		if !errors.As(err, &typed) || typed.Operation != "geom conversion" || typed.Reason != problem {
			t.Fatalf("error = %#v, want unsupported reason %q", err, problem)
		}
	case geo.ErrCRS:
		var typed *geo.CRSError
		if !errors.As(err, &typed) || typed.SRID != 0 || typed.Problem != problem {
			t.Fatalf("error = %#v, want CRS problem %q", err, problem)
		}
	default:
		t.Fatalf("unsupported expected category %v", category)
	}
}

func errorText(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

type facadeFakeGeometry struct{}

func (facadeFakeGeometry) Layout() geom.Layout   { return geom.XY }
func (facadeFakeGeometry) Stride() int           { return 2 }
func (facadeFakeGeometry) Bounds() *geom.Bounds  { return geom.NewBounds(geom.XY) }
func (facadeFakeGeometry) FlatCoords() []float64 { return []float64{0, 0} }
func (facadeFakeGeometry) Ends() []int           { return nil }
func (facadeFakeGeometry) Endss() [][]int        { return nil }
func (facadeFakeGeometry) SRID() int             { return 4326 }
func (facadeFakeGeometry) Empty() bool           { return false }

type facadeXYZOddGeometry struct{}

func (facadeXYZOddGeometry) Layout() geom.Layout   { return geom.XYZ }
func (facadeXYZOddGeometry) Stride() int           { return 3 }
func (facadeXYZOddGeometry) Bounds() *geom.Bounds  { return geom.NewBounds(geom.XYZ) }
func (facadeXYZOddGeometry) FlatCoords() []float64 { return []float64{0} }
func (facadeXYZOddGeometry) Ends() []int           { return nil }
func (facadeXYZOddGeometry) Endss() [][]int        { return nil }
func (facadeXYZOddGeometry) SRID() int             { return 0 }
func (facadeXYZOddGeometry) Empty() bool           { return false }
