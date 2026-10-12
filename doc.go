// Package geo provides validated, immutable geospatial values and geometry.
//
// Coordinates are always constructed and serialized in longitude, latitude
// order. Every coordinate and geometry carries an explicit CRS; algorithms do
// not transform between CRSs. Constructors reject non-finite values and apply
// configurable resource limits to aggregate geometry. Custom CRS metadata does
// not relax longitude/latitude ranges or perform registry validation.
//
// Scalar zero values represent numeric zero; CRS and primitive geometry zero
// values are not validated constructions. Constructed geometry is immutable:
// slice/map accessors return copies, while inaccessible child storage may be
// shared. Calls are synchronous without cancellation or background resources.
//
// Polygon operations use planar longitude/latitude edges after unwrapping, not
// spherical or ellipsoidal geodesics. Use geodesy for explicitly selected
// distance models and inspect model-specific bearing validity flags.
package geo
