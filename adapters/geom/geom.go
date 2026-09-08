// Package geogeom converts between immutable go-geo geometry values and
// caller-owned github.com/twpayne/go-geom values through canonical EWKB.
//
// Conversion is synchronous, stateless, safe for concurrent calls when the
// caller does not concurrently mutate an input geom.T, and O(n) in geometry
// size and allocation. Only XY layouts with positive SRIDs are supported;
// coordinates are never transformed. Both directions return newly owned
// values and retain no mutable aliases to their inputs.
//
// FromGoGeom resolves geo.Limits, checks collection structure before outer
// layout and SRID, then checks descendant coordinates before canonical EWKB
// conversion. Collection depth is always capped at 32 even when a larger limit
// is requested. Nil children and cycles return *geo.EncodingError; point,
// geometry, and depth bounds return *geo.TopologyError; layout and SRID return
// *geo.UnsupportedError and *geo.CRSError. Marshal failures retain their cause,
// encoded-byte failures have no cause, and downstream decode errors are
// returned unchanged. These checks normalize the former GeometryCollection
// FlatCoords panic without recovering unrelated caller-induced panics.
//
// Migrate from adapter/gogeom either by importing this package with its
// declared geogeom identifier or by retaining an explicit gogeom import alias.
package geogeom

import (
	"encoding/binary"
	"reflect"

	"github.com/twpayne/go-geom"
	geomwkb "github.com/twpayne/go-geom/encoding/ewkb"

	geo "github.com/faustbrian/go-geo"
	"github.com/faustbrian/go-geo/wkb"
)

const maxSafeCollectionDepth = 32

// ToGoGeom converts an owned geo geometry through canonical little-endian
// EWKB. The returned geom.T is newly allocated and caller-owned.
func ToGoGeom(geometry geo.Geometry) (geom.T, error) {
	if geometry == nil {
		return nil, adapterError("cannot convert nil geometry", nil)
	}
	encoded, err := wkb.MarshalEWKB(geometry, binary.LittleEndian)
	if err != nil {
		return nil, err
	}
	converted, _ := geomwkb.Unmarshal(encoded)
	return converted, nil
}

// FromGoGeom converts a caller-owned geom value through canonical
// little-endian EWKB. The returned geo geometry is immutable and newly owned.
// Limits bound point and encoded-byte work; zero limits resolve to package
// defaults.
func FromGoGeom(value geom.T, limits geo.Limits) (geo.Geometry, error) {
	if nilGeometry(value) {
		return nil, adapterError("cannot convert nil geometry", nil)
	}
	if collection, ok := value.(*geom.GeometryCollection); ok {
		limits = geo.ResolveLimits(limits)
		if err := preflightCollection(collection, limits); err != nil {
			return nil, err
		}
		if err := validateLayoutAndSRID(collection); err != nil {
			return nil, err
		}
		if err := validateCollectionCoordinates(collection, limits); err != nil {
			return nil, err
		}
		return convert(value, limits)
	}
	if value.Layout() != geom.XY {
		return nil, &geo.UnsupportedError{
			Operation: "geom conversion",
			Reason:    "only the two-dimensional XY layout is supported",
		}
	}
	if value.SRID() <= 0 {
		return nil, &geo.CRSError{
			SRID:    int32(value.SRID()),
			Problem: "geom conversion requires a positive SRID",
		}
	}
	limits = geo.ResolveLimits(limits)
	coordinates := value.FlatCoords()
	if len(coordinates)%2 != 0 {
		return nil, adapterError("geom has malformed XY coordinates", nil)
	}
	if len(coordinates)/2 > limits.MaxPoints {
		return nil, &geo.TopologyError{
			Geometry: "geom",
			Problem:  "point limit exceeded",
		}
	}

	return convert(value, limits)
}

func convert(value geom.T, limits geo.Limits) (geo.Geometry, error) {
	encoded, err := geomwkb.Marshal(value, binary.LittleEndian)
	if err != nil {
		return nil, adapterError("cannot encode geom EWKB", err)
	}
	if int64(len(encoded)) > limits.MaxEncodedBytes {
		return nil, adapterError("encoded byte limit exceeded", nil)
	}
	return wkb.UnmarshalEWKB(encoded, limits)
}

func validateLayoutAndSRID(value geom.T) error {
	if value.Layout() != geom.XY {
		return &geo.UnsupportedError{
			Operation: "geom conversion",
			Reason:    "only the two-dimensional XY layout is supported",
		}
	}
	if value.SRID() <= 0 {
		return &geo.CRSError{
			SRID:    int32(value.SRID()),
			Problem: "geom conversion requires a positive SRID",
		}
	}
	return nil
}

type collectionFrame struct {
	collection *geom.GeometryCollection
	next       int
	depth      int
}

func preflightCollection(collection *geom.GeometryCollection, limits geo.Limits) error {
	effectiveDepth := min(limits.MaxCollectionDepth, maxSafeCollectionDepth)
	if 1 > limits.MaxGeometries {
		return topologyError("geometry limit exceeded")
	}
	if 1 > effectiveDepth {
		return topologyError("collection depth limit exceeded")
	}

	geometryCount := 1
	active := map[*geom.GeometryCollection]struct{}{collection: {}}
	stack := []collectionFrame{{collection: collection, depth: 1}}
	for len(stack) > 0 {
		index := len(stack) - 1
		frame := &stack[index]
		if frame.next == frame.collection.NumGeoms() {
			delete(active, frame.collection)
			stack = stack[:index]
			continue
		}

		child := frame.collection.Geom(frame.next)
		frame.next++
		if nilGeometry(child) {
			return adapterError("geom collection contains nil geometry", nil)
		}
		if geometryCount >= limits.MaxGeometries {
			return topologyError("geometry limit exceeded")
		}
		geometryCount++
		childDepth := frame.depth + 1
		if childDepth > effectiveDepth {
			return topologyError("collection depth limit exceeded")
		}
		childCollection, ok := child.(*geom.GeometryCollection)
		if !ok {
			continue
		}
		if _, cyclic := active[childCollection]; cyclic {
			return adapterError("geom collection contains a cycle", nil)
		}
		active[childCollection] = struct{}{}
		stack = append(stack, collectionFrame{collection: childCollection, depth: childDepth})
	}
	return nil
}

func validateCollectionCoordinates(collection *geom.GeometryCollection, limits geo.Limits) error {
	pointCount := 0
	stack := []collectionFrame{{collection: collection}}
	for len(stack) > 0 {
		index := len(stack) - 1
		frame := &stack[index]
		if frame.next == frame.collection.NumGeoms() {
			stack = stack[:index]
			continue
		}

		child := frame.collection.Geom(frame.next)
		frame.next++
		if childCollection, ok := child.(*geom.GeometryCollection); ok {
			stack = append(stack, collectionFrame{collection: childCollection})
			continue
		}
		coordinates := child.FlatCoords()
		if len(coordinates)%2 != 0 {
			return adapterError("geom has malformed XY coordinates", nil)
		}
		points := len(coordinates) / 2
		if points > limits.MaxPoints || pointCount > limits.MaxPoints-points {
			return topologyError("point limit exceeded")
		}
		pointCount += points
	}
	return nil
}

func topologyError(problem string) error {
	return &geo.TopologyError{Geometry: "geom", Problem: problem}
}

func nilGeometry(value geom.T) bool {
	if value == nil {
		return true
	}
	reflected := reflect.ValueOf(value)
	return reflected.Kind() == reflect.Pointer && reflected.IsNil()
}

func adapterError(problem string, cause error) error {
	return &geo.EncodingError{
		Format:  "geom adapter",
		Problem: problem,
		Cause:   cause,
	}
}
