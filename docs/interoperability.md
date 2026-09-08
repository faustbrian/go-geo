# Codecs, pgx, and PostGIS

## GeoJSON

`geojson.Marshal` and `Unmarshal` support every package geometry.
`MarshalFeature` and `UnmarshalFeature` support a geometry, optional ID, and an
owned JSON property object. GeoJSON has no standard SRID field here: callers
must supply the CRS when decoding. Coordinate order remains longitude,
latitude.

## WKT, EWKT, WKB, and EWKB

Plain WKT/WKB omit CRS, so decoding requires a caller-supplied CRS. EWKT/EWKB
carry a positive SRID. Encoders emit canonical 2D representations and reject
nil or invalid geometry. WKB callers choose byte order. All decoders apply
`geo.Limits`, including `MaxEncodedBytes`, before unbounded work.

The live PostGIS corpus exercises every geometry family, supported empty
aggregate, GeoJSON/WKT/EWKT/WKB/EWKB representation, NDR and XDR byte order,
SRID metadata, and 2D rejection on PostGIS 16 / 3.5 and 18 / 3.6. See
[`verification.md`](verification.md#codec-and-postgis-interoperability-corpus).

## geom adapter

`adapters/geom` (package `geogeom`) is the canonical conversion boundary.
`adapter/gogeom` is a deprecated thin facade with the support interval and
migration forms documented in [`adoption.md`](adoption.md). Both paths convert
through canonical little-endian EWKB, accept only two-dimensional layouts with
positive SRIDs, preserve root error types and causes, and perform no CRS
transformation.

Inputs remain caller-owned. Do not mutate a `geom.T` concurrently with a
conversion. Returned `geom.T` values are newly owned by the caller, and
returned `geo.Geometry` values are immutable; neither direction retains a
mutable slice alias. Calls are synchronous, stateless, concurrency-safe under
that ownership rule, and O(n) in geometry size and allocation.

Collections receive an iterative structural check before go-geom's recursive
layout and EWKB operations. That pass rejects nil children and cycles, counts
the outer and nested geometries, and caps effective depth at 32 even when a
larger limit is supplied. Structural failures precede collection layout and
SRID checks. A second iterative pass then cumulatively bounds descendant
points before marshal. Valid direct, nested, mixed-kind, and empty XY
collections preserve child order, type, coordinates, and SRID without the
`GeometryCollection.FlatCoords` panic present in v1.0.0.

## database/sql

`postgis.Value` owns a geometry, implements `driver.Valuer` and `sql.Scanner`,
and represents SQL NULL when invalid. It accepts binary EWKB and canonical
hexadecimal EWKB. Initialize it with limits before scanning untrusted rows:

```go
value, err := postgis.NewValue(nil, limits)
err = value.Scan(source)
geometry, valid := value.Geometry()
```

## pgx registration

PostGIS OIDs are installation-specific. Query and register both spatial types
on every new connection that will use both helpers:

```go
var geometryOID, geographyOID uint32
err := conn.QueryRow(ctx,
    "SELECT 'geometry'::regtype::oid, 'geography'::regtype::oid",
).Scan(&geometryOID, &geographyOID)
postgis.Register(conn.TypeMap(), geometryOID, limits)
postgis.Register(conn.TypeMap(), geographyOID, limits)
```

Registration is connection/type-map local. Perform it in pool connection setup.

## Safe query fragments

`postgis.NewColumn` accepts one to three ordinary SQL identifier segments and
quotes each segment. `GeographyDWithin` and `Intersects` return fixed SQL plus
separate arguments. They never interpolate values. The caller owns the full
query and must pass the returned arguments in placeholder order.

Use `GeographyDWithin` for metre distances on WGS84 geography. Use `Intersects`
for the column's geometry coordinate system. These helpers are intentionally not
a query builder.

## Process or database?

Calculate in process when the candidate set is already small and loaded, when
the result is part of validation, or when deterministic serialization is the
goal. Query PostGIS when an index can reduce candidates, spatial data remains in
the database, or a join/aggregate would otherwise move large geometry sets into
the service.
