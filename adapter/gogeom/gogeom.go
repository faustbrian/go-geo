// Package gogeom converts between geo and github.com/twpayne/go-geom.
//
// This compatibility facade delegates directly to adapters/geom and therefore
// has the same ownership, O(n) time and allocation, caller-managed concurrency,
// limit, and typed-error behavior. Conversions return newly owned values and
// retain no mutable input aliases. Only XY layouts with positive SRIDs are
// supported. Geometry collections are structurally checked before layout and
// SRID validation, are capped at depth 32, and return geo encoding, topology,
// unsupported, CRS, or downstream codec errors rather than the released
// GeometryCollection FlatCoords panic.
//
// Existing imports may migrate by replacing this package path with
// github.com/faustbrian/go-geo/adapters/geom while retaining an explicit
// gogeom alias, or by adopting that package's declared geogeom identifier.
//
// Deprecated: use github.com/faustbrian/go-geo/adapters/geom. This compatibility
// path remains supported for the longer of 180 days after the successor's
// public availability and two subsequently published stable minor releases.
package gogeom

import (
	"github.com/twpayne/go-geom"

	geo "github.com/faustbrian/go-geo"
	geogeom "github.com/faustbrian/go-geo/adapters/geom"
)

// ToGoGeom converts an owned geo geometry through canonical EWKB.
//
// Deprecated: use geogeom.ToGoGeom.
func ToGoGeom(geometry geo.Geometry) (geom.T, error) {
	return geogeom.ToGoGeom(geometry)
}

// FromGoGeom converts a two-dimensional geom value with a positive SRID.
//
// Deprecated: use geogeom.FromGoGeom.
func FromGoGeom(value geom.T, limits geo.Limits) (geo.Geometry, error) {
	return geogeom.FromGoGeom(value, limits)
}
