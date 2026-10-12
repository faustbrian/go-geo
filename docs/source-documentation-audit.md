# Source documentation audit

## Contract and boundary

This Tier A pass owns source comments and directly contradicted current prose
at main `5bb065b1082ac76496db6ab8adc0abdf4c4cabae`. Preserve executable
syntax, literals, directives, dependencies, fixtures, historical decisions and
API data. The owner is go-geo; consumers are geometry callers, codec users,
measurement models, geom adapters, and SQL/pgx integrations. Acceptance is
useful source-derived contracts, not a new numerical or release qualification.

## Reproducible initial inventory

`git ls-files` and `golib source-docs inventory --json` identify one root module,
73 Go files (26 maintained program/library files and 47 test files), and 19
source directories: root; adapter/gogeom; adapters/geom; geodesy; geohash;
geojson; geotest; postgis; wkb; wkt; and nine example programs. All tracked
files, including fuzz callbacks, benchmark/assertion helpers, saved corpora,
and PostGIS fixtures, are in scope for classification. There are no tracked
vendor, generated Go, generator directives, platform suffix variants, build
constraints, or embedded executable templates. The two canonical-adapter lint
directives deliberately exercise the deprecated facade and remain unchanged.
The initial structural inventory reports 77 missing public member/method or
example package comments; zero findings alone will not prove semantics.

The configured API mode is `apidiff`, with `api/baseline.txt` and
`scripts/check-api.sh`; this is not a full-prose GoDoc snapshot. Preserve it.
The Make full check includes online authority, conformance, campaigns and
benchmarks; those unchanged-runtime boundaries are not this prose pass.

## Source-derived contract audit

- Scalars reject non-finite/range violations, normalize signed zero, and do
  not wrap arbitrary degrees. Geographic coordinate limits also apply to
  custom positive CRS metadata; no registry validation or reprojection occurs.
- Bounds canonicalize dateline endpoints for inclusive comparisons. Polygon
  topology unwraps against the first exterior longitude, whereas location
  unwraps each endpoint against the query. Both are planar, with exact floating
  boundary tests, not ellipsoidal edges or a hidden tolerance.
- Geometry constructors copy mutable inputs; immutable children may share
  inaccessible backing storage. Pointer canonicalization is not revalidation
  of zero-value geometry. Collections include their own node in count/depth.
- Codec byte ceilings bound input, not a byte-for-byte heap cap. GeoJSON first
  parses coordinate arrays before constructor aggregate checks; WKB preflights
  counts against remaining wire bytes, WKT grows bounded lists. Ring bounds are
  per polygon, not a global nested ring budget. Encoders have no caller limit.
- Models are caller-implementable synchronous callbacks, not a closed owned
  interface. Calls have no context or panic recovery; callers own budgets and
  synchronization. Nearest evaluates every candidate even for a zero limit.
  Sphere Name is a fixed label even for a custom radius. Radius envelopes use
  an outward floating margin, not exact-real arithmetic.
- Geohash midpoint equality selects upper halves; polar neighbors can repeat.
  Cover enumerates inclusive indexed rows/columns then sorts; work includes
  sorting, not just linear output traversal.
- Feature raw properties/ID bytes are copied. ID validation reads one JSON
  token; NewFeature does not enforce WGS84. Marshal checks it, but MarshalFeature
  ignores the nested failure and emits null. Scan receivers and
  pgx maps are caller-owned mutable state; this module opens no connection or
  goroutine. SQL fragments bind values; Args copies only its outer slice.
  GeographyDWithin checks numeric SRID 4326, not the CRS name.
- Test helpers return copied vectors, with deep copies for polygon cases.
  Tolerance rejects non-finite operands but float arithmetic can overflow;
  selected vectors are not complete OGC/GeographicLib conformance.

## Missing versus investigation

Missing comments will describe fields, errors, model callbacks and example
purpose. Current resource and CRS prose will be narrowed to the actual source;
any changed typed decision receives only an append-only native content digest.
The immutable authority monitoring pins and dated dependency/performance
reports will remain intact. Investigate only concrete existing edge claims:
geographically coincident but unequal coordinate encodings and Feature ID
trailing input. Runtime repair is outside this pass.

## Confirmed edge observations

On Go 1.27.2, darwin/arm64, equivalent pole encodings (0,90)/(90,90)
and dateline encodings (-180,0)/(180,0) returned zero ellipsoidal distance
with BearingsDefined true. The mean sphere returned approximately 5.52e-10 m
and 1.56e-9 m respectively, also flagged defined. A raw ID `1 2` was accepted
by NewFeature but rejected by MarshalFeature. A Point tagged SRID 4326 with
name caller-label was accepted by GeographyDWithin and rejected by geodesy.
MarshalFeature accepted that stored geometry but emitted null, ignoring the
nested geometry marshal error. These are existing source limits, not runtime
repairs or complete conformance observations.

## Smallest sufficient evidence

Use the retained structural/local-document/specification checks, pinned changed
content secret scan, explicit formatting, all-73-file syntax/significant-token/
literal/directive comparison, all tracked package GoDoc renders, and complete
ordinary diff inspection. A small disposable observation may settle the named
edge claims; meaningful standalone examples run without external services.
No database, race/leak, active fuzz, performance, online authority or release
campaign is justified by unchanged runtime source. Record executed versus
reused evidence and concrete limitations in the delivery report.

## Verification disposition

Go 1.27.2 formatting and all 73 tracked files preserve AST, significant tokens,
literals, implicit semicolon counts and directives exactly. All 371 comment
groups parse/render through go/doc/comment. All 19 native package renders were
inspected; after final prose correction only geohash and geojson were refreshed,
with the unchanged 17 renders explicitly reused. Ten fuzz and six benchmark
entrypoints were classified without campaigns. There are no native Example
functions; all nine standalone programs ran, with pgx producing only DSN-unset
setup guidance. No live database result is claimed.

The native typed decision serialization matched all eight baseline digests;
only three current descriptions append a digest, preserving every earlier entry.
Structural source, local documentation and offline specification checks pass.
The API baseline, dependency bytes, seven saved fuzz inputs, PostGIS fixture,
authority pins and dated qualification reports are preserved. Final ordinary
diff and changed-content/staged secret inspection precede the local commit;
remote main and CI delivery remain the coordinator's responsibility.
