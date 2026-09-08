# Canonical go-geom adapter

`github.com/faustbrian/go-geo/adapters/geom` (package `geogeom`) is the
canonical conversion boundary between immutable `geo.Geometry` values and
caller-owned `github.com/twpayne/go-geom` values.

```go
external := geom.NewPointFlat(geom.XY, []float64{24.9384, 60.1699}).SetSRID(4326)
owned, err := geogeom.FromGoGeom(external, geo.DefaultLimits())
if err != nil {
	return err
}
roundTrip, err := geogeom.ToGoGeom(owned)
```

Both functions convert through canonical little-endian EWKB. They are
synchronous, stateless, O(n), and safe for concurrent calls when the caller
does not mutate the same upstream value during a call. Only XY layouts with
positive SRIDs are accepted; coordinates are never transformed. Returned
values are newly owned and retain no mutable input aliases.

`FromGoGeom` resolves zero `geo.Limits`, rejects nil values, and preserves root
Geo error types, categories, fields, messages, and safe causes. Collections
receive a bounded iterative structural pass before recursive layout work, with
an effective depth ceiling of 32, followed by cumulative point validation
before marshal. Valid direct, nested, mixed-kind, and empty collections retain
order, geometry type, coordinates, and SRID.

See the [migration and compatibility guide](../../docs/adoption.md),
[interoperability contract](../../docs/interoperability.md),
[verification evidence](../../docs/verification.md), and
[API reference](https://pkg.go.dev/github.com/faustbrian/go-geo/adapters/geom).
The shared ecosystem conventions are documented in the
[versioned design language](https://github.com/faustbrian/go-library-tools/blob/v1.4.0/docs/ecosystem/design-language.md).
