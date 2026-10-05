# Security model

Version 1 (2026-09-30). This model covers the root geometry module, its WKB,
WKT, GeoJSON, geohash, and geodesy packages, and the go-geom and PostGIS
adapters. The go-geo maintainers own package controls; integrating services own
request and database operation budgets.

## Trust boundaries and assets

Untrusted request bytes may become geometry, Feature properties, text or binary
spatial encodings, geohash bounds, or PostGIS rows. A service may use these to
make spatial decisions, generate SQL predicates, or allocate memory and CPU.
The protected assets are process availability, spatial result integrity, and
application data reachable through SQL. This module does not read credentials,
open connections, start goroutines, or issue queries. Callers own transport,
database, and request contexts.

## Controls

- `geo.Limits` bounds decoded bytes, points, rings, geometries, and collection
  depth. Callers should lower the defaults for untrusted request paths.
- WKB, WKT, and GeoJSON decoding validates structure and CRS; aggregate
  geometry construction checks topology and cumulative limits.
- PostGIS binary and hexadecimal EWKB scan paths enforce the configured
  decoded-byte ceiling before copying or converting wire values. A hexadecimal
  representation may contain twice as many source bytes, plus its prefix.
- The closed geometry model prevents external implementations from bypassing
  construction invariants. The go-geom adapter checks shape and collection
  structure before upstream marshaling.
- SQL column names are validated and quoted. Spatial values and distances are
  returned as separate bound arguments; callers must not interpolate them.
- Errors classify malformed input. Rendered `Error()` text for invalid Feature
  properties and unsupported GeoJSON/WKT geometry types omits the supplied
  names. Wrapped causes can retain input, such as an invalid WKT numeric token;
  applications must not retain, unwrap, or log untrusted error objects across
  a higher-trust boundary without their own redaction.

## Residual risks and review conditions

| Risk | Owner, rationale, mitigation, and review condition |
| --- | --- |
| Accepted bounded CPU cost | Integrating service. Valid large polygons can require expensive upstream topology work; finite package limits are not a wall-clock deadline. Set smaller request limits and operation deadlines from measured budgets. Review when traffic shape, limits, or topology dependency changes. |
| Accepted input buffering | Integrating service. pgx and database/sql may buffer a row before this package receives its bytes, and this package cannot preempt a running database query. Bound query results and use context deadlines at the database boundary. Review when scan transport or driver behavior changes. |
| Accepted wrapped-cause retention | Integrating service. `Unwrap` can expose caller input in a parser cause, including an invalid WKT numeric token. Treat the full error object as untrusted; redact before retention or cross-boundary logging. Review if error provenance or logging policy changes. |
| Accepted dependency risk | go-geo maintainers. GeographicLib, simplefeatures, go-geom, and pgx implement numerical or wire operations outside this module. Keep versions pinned, review advisories and conformance evidence, and rerun affected gates on upgrades. |

The PostGIS pre-limit allocation, GeoJSON collection pre-limit work, and
direct type and property-name echo in rendered diagnostics fixed in the
released v1.1.2 are not accepted residual risks. See the changelog and
focused regression tests.
