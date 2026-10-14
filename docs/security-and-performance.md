# Security, limits, and performance

[Documentation map](index.md) · [Repository README](../README.md)
[Next: Compatibility, migration, troubleshooting, and FAQ](adoption.md)

The [versioned security model](security.md) identifies trust boundaries,
package controls, residual-risk owners, and review conditions.

## Untrusted input

Never decode untrusted bytes with application-inappropriate limits. A zero
field in `geo.Limits` selects the conservative default:

- 1,000,000 points
- 10,000 rings
- 100,000 geometries
- collection depth 32
- 64 MiB encoded input

Reduce these values for request paths. Point/node totals are checked by
aggregate constructors; ring limits are per polygon, and binary nesting also
counts multi-geometry children. Zero fields select defaults; negative fields
are not normalized. Byte ceilings bound input, not total heap consumption.
WKB checks declared counts against limits and remaining bytes before
allocation. WKT checks lists while growing them. GeoJSON allocates decoded
arrays and raw child JSON before constructor aggregate checks. Constructors
validate topology after their structural/resource checks.

Codec decoders reject malformed numbers, excessive nesting and trailing
document data. NewFeature requires one complete string/number JSON ID and copies
its raw bytes without number coercion. MarshalFeature propagates geometry
encoding failures without publishing bytes; actually absent geometry stays null.
No encoder accepts a configurable output-byte ceiling.

These limits make work finite; they do not promise acceptable request latency.
Polygon validity uses an upstream robust planar topology implementation whose
worst-case CPU complexity is not a stable package contract. Choose substantially
smaller request-specific point and ring limits from measured budgets.

SQL identifiers are validated and quoted; geometry and distance values remain
bound parameters. Do not concatenate `Fragment.Args()` into SQL.

## Complexity and allocation

Public algorithm comments state their complexity. Scalar validation, bounds,
spherical/ellipsoidal point geodesy, and radius envelopes are O(1). Line and
polygon measurements are O(points). `Nearest` is O(n log n) and allocates O(n).
Geohash encode/decode is O(precision); covers allocate O(returned cells) after
checking `maxCells`, then sort in O(k log k) work for k returned cells.

Codec work is O(encoded bytes plus geometry size). The WKB encoder pre-sizes its
buffer and has an allocation regression test: encoding a 100,000-point line is
bounded to three allocations. Benchmarks cover 10, 1,000, and 100,000-point
WKB, 1,000-point polygon validation, geom conversion, spherical versus
ellipsoidal inverse calculations, and pgx binary codecs.

Executable allocation ceilings cover 1,000-point polygon validation,
1,000-point geom conversion, 100,000-point WKB encoding, and pgx binary
encode/scan. The dated measurement matrix is in
[`verification.md`](verification.md#complexity-allocation-and-benchmark-baseline).

## Dependency isolation

- GeographicLib supplies audited WGS84 geodesics behind `geodesy.Model`.
- simplefeatures validates polygon topology behind immutable package geometry.
- geom is isolated to an adapter and differential tests.
- pgx is isolated to `postgis`.

These dependencies use redistributable licenses recorded by their modules.
Release review must inspect `go mod graph`, license metadata, `govulncheck`, and
benchmark/fuzz changes before upgrading them.

The compiled closure has no cgo packages. `simplefeatures/geom` does import
`unsafe` for native-endian detection and zero-copy float/byte reinterpretation;
its optional GEOS/PROJ cgo packages are not linked. See
[`dependencies.md`](dependencies.md) for the dated advisory, version, license,
unsafe, cgo, and numerical-risk audit.

## Verification strategy

Statement coverage is diagnostic as configured; coverage alone is not the
confidence claim. The suite also contains authoritative vectors,
algebraic/property invariants, geom and PostGIS differential tests, hostile size/depth/overflow
cases, fuzz targets for every decoder, aggregate constructor, geodesic model,
and adapter boundary, the race detector, and representative benchmarks.

Run a focused two-second active smoke for the parser you change:

```sh
go test ./wkb -run '^$' -fuzz '^FuzzDecode$' -fuzztime=2s
```

Ordinary Go tests and base-free `make check` run deterministic seeds/corpora;
changed-source CI selects active smoke. Use `make check-full` for the existing
10,000-iteration campaigns and selected broader race/conformance assurance,
not as an unconditional parser-edit prerequisite.

Checked-in corpora live below each package's `testdata/fuzz` directory. Their
coverage and callback bounds are catalogued in
[`verification.md`](verification.md#fuzz-corpora-and-hostile-input-boundaries).
