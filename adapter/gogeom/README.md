# go-geom compatibility adapter

`github.com/faustbrian/go-geo/adapter/gogeom` is the deprecated compatibility
path for conversion between immutable `geo.Geometry` values and caller-owned
`github.com/twpayne/go-geom` values. New code should use
[`adapters/geom`](../../adapters/geom), whose declared package identifier is
`geogeom`.

Existing call sites can migrate without renaming their qualifier:

```go
import gogeom "github.com/faustbrian/go-geo/adapters/geom"
```

The facade delegates without wrapping, so it has the successor's exact value,
typed-error, cause, limit, and ownership behavior. Conversion is synchronous,
stateless, and safe for concurrent calls when the caller does not mutate the
same upstream value during a call. Only XY layouts with positive SRIDs are
accepted, no CRS transformation occurs, and results retain no mutable aliases.
Collection structure is bounded before recursive upstream work; cumulative
points are checked after layout and SRID validation but before marshal.

The legacy path remains supported for the longer of 180 days after the
successor becomes publicly available and two subsequently published stable
minor releases. See the [migration and compatibility guide](../../docs/adoption.md),
[interoperability contract](../../docs/interoperability.md),
[verification evidence](../../docs/verification.md), and
[API reference](https://pkg.go.dev/github.com/faustbrian/go-geo/adapter/gogeom).
The shared ecosystem conventions are documented in the
[versioned design language](https://github.com/faustbrian/go-library-tools/blob/v1.4.0/docs/ecosystem/design-language.md).
