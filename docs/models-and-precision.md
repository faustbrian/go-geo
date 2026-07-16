# Models, bounds, and precision

## Coordinates and CRS

Coordinates are `(longitude, latitude)` in every API and codec. WGS84 is
`EPSG:4326`. Custom positive SRIDs may be represented, but geodesy and geohash
operations require WGS84 and return `geo.ErrCRS` otherwise. The module never
transforms coordinates.

All scalar values use IEEE 754 `float64`. Constructors reject NaN, infinity,
and values outside their documented ranges. Signed zero is normalized.
Text/JSON marshaling uses the shortest decimal that round trips to the same
binary value; it does not imply survey-grade input accuracy.

## Distance and bearing models

`geodesy.WGS84Ellipsoid()` uses GeographicLib's Karney algorithms and the WGS84
parameters `a=6378137 m`, `f=1/298.257223563`. Use it for production earth
distance, initial/final bearing, and destination calculations.

`geodesy.MeanEarthSphere()` uses the IUGG mean radius `6,371,008.8 m`. It is a
great-circle approximation. Its error depends on latitude, direction, and
distance; do not represent it as an ellipsoidal answer. `geodesy.NewSphere`
supports an explicitly chosen positive radius.

Inverse operations are symmetric in distance, but initial and final bearings
depend on direction. Bearings are undefined for coincident points. Spherical
antipodes also have no unique bearing; inspect `BearingsDefined()`.

## Bounds and the antimeridian

`BoundingBox` order is west, south, east, north. Edges are inclusive. A west
longitude greater than east denotes an antimeridian-crossing box; it is not
normalized into a much larger non-crossing box. `Contains` and `Overlaps`
preserve this interpretation. A radius envelope that reaches either pole spans
all longitudes; a radius of at least half the sphere covers the world.

## Polygons

Rings must be closed, contain enough distinct points, share one CRS, and form a
valid simple polygon. Self-intersections, holes outside the exterior, nested or
overlapping holes, and degenerate rings fail with `geo.ErrTopology`. Input
winding is accepted when topology is valid and is preserved by the core model.

`Polygon.Locate` returns `Exterior`, `Interior`, or `Boundary` semantics through
the exported `Location` cases. Exterior and hole edges are boundaries. Points
inside holes are exterior to the polygon.

## Tolerances and evidence

Exact equality compares the represented values and CRS, not geographic
nearness. For numerical results use an application-selected tolerance or the
`geotest` helpers. WGS84 tests use published GeographicLib vectors. Polygon
location vectors use the OGC Simple Features 1.1 text example and its normative
interior/boundary/exterior model. Differential tests compare WKB with `go-geom`
and live integration compares codecs, predicates, spherical distance, and
spheroidal distance with PostGIS. The PostGIS spherical comparison allows 5 cm
for its different mean-radius rounding; ellipsoidal comparison allows 1 mm.

Sources:

- GeographicLib inverse example: <https://geographiclib.sourceforge.io/html/python/examples.html>
- OGC Simple Features for SQL 1.1, polygon rules and text example:
  <https://docs.ogc.org/is/99-049/99-049.pdf>
