# Security, limits, and performance

## Untrusted input

Never decode untrusted bytes with application-inappropriate limits. A zero
field in `geo.Limits` selects the conservative default:

- 1,000,000 points
- 10,000 rings
- 100,000 geometries
- collection depth 32
- 64 MiB encoded input

Reduce these values for request paths. Limits cover aggregate nested geometry,
not merely each child. Decoders reject truncated values, invalid lengths,
integer overflow, unsupported dimensions/types, malformed numbers, excessive
nesting, and trailing data. Constructors validate topology after applying
resource bounds.

SQL identifiers are validated and quoted; geometry and distance values remain
bound parameters. Do not concatenate `Fragment.Args()` into SQL.

## Complexity and allocation

Public algorithm comments state their complexity. Scalar validation, bounds,
spherical/ellipsoidal point geodesy, and radius envelopes are O(1). Line and
polygon measurements are O(points). `Nearest` is O(n log n) and allocates O(n).
Geohash encode/decode is O(precision); covers allocate O(returned cells) after
checking `maxCells`.

Codec work is O(encoded bytes plus geometry size). The WKB encoder pre-sizes its
buffer and has an allocation regression test: encoding a 100,000-point line is
bounded to three allocations. Benchmarks cover 10, 1,000, and 100,000-point
WKB, spherical versus ellipsoidal inverse calculations, and pgx binary codecs.

## Dependency isolation

- GeographicLib supplies audited WGS84 geodesics behind `geodesy.Model`.
- simplefeatures validates polygon topology behind immutable package geometry.
- go-geom is isolated to an adapter and differential tests.
- pgx is isolated to `postgis`.

These dependencies use redistributable licenses recorded by their modules.
Release review must inspect `go mod graph`, license metadata, `govulncheck`, and
benchmark/fuzz changes before upgrading them.

## Verification strategy

Statement coverage is exactly 100%, but coverage alone is not the confidence
claim. The suite also contains authoritative vectors, algebraic/property
invariants, go-geom and PostGIS differential tests, hostile size/depth/overflow
cases, fuzz targets for every decoder and aggregate constructor, the race
detector, and representative benchmarks.

Fuzz locally for longer than the CI smoke duration before modifying parsers:

```sh
go test ./wkb -run '^$' -fuzz '^FuzzDecode$' -fuzztime 1m
```
