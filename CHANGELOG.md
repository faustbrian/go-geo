# Changelog

All notable changes are documented here. The project follows
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and uses semantic
versioning.

## Unreleased

### Changed

- Adopt the checksum-verified `go-library-tools` v1.4.0 CLI and immutable W14
  reusable workflow, including strict online specification validation, without
  changing the geospatial API or runtime behavior.

- Refresh the EPSG:4326 and release-history authority-body pins after reviewing
  mutable page presentation changes and confirming the v13.102 WKT definition
  and release status are unchanged.

- Adopt the checksum-verified `go-library-tools` v1.2.0 CLI and immutable
  shared workflow so local and hosted gates enforce specification governance
  while retaining package-owned source and evidence.
- Adopt the checksum-verified `go-library-tools` v1.3.0 CLI, schema-v2 cohesion
  metadata, and repository-local cohesion gate while retaining package-owned
  source and evidence.

### Documentation

- Link ecosystem and Domain utilities family guidance to the immutable v1.4.0
  documentation release.

- Replace archived monorepo and dated hardening terminology with package-owned
  documentation and numerical verification guidance.
- Publish the standards decision register, pinned authority monitoring,
  conformance bindings, maintained-peer evidence, and specification-aware
  change control. The current decision digests are recorded below.
- Link the module to the immutable v1.3.0 Golib ecosystem guidance and correct
  the documented minimum Go version to 1.26.6.

### Specification Decisions

<!-- Canonical digests are maintained with specification/decision-history.json. -->

- GEO-GEOJSON-DEC-001 sha256:5a7226c4cc2b27171488e0679ce794a0a7f755db4c21867f9be83ecd9404e0cf
- GEO-GEOJSON-DEC-002 sha256:cef5811efc3ebad7e9ca79ecfed36ca10073e002196c8c42c61687985bb61267
- GEO-OGC-DEC-001 sha256:04b721e5e1b367f53c5ca93b28b0c4e853943d435996acb027af1dc32825eee0
- GEO-OGC-DEC-002 sha256:6964003803fd82e62b195afed7beae6bf0b73a89fb3720f7c1b8f223f57406ea
- GEO-POSTGIS-DEC-001 sha256:b9f01928227d261d98696daf88dcd8f1b8a869134a4bed1163b9289cf6e40f18
- GEO-POSTGIS-DEC-002 sha256:5170a4305ec67948e5faeebf7690e98d935544ed97afa910fea8ef4f231ba3eb
- GEO-EPSG-DEC-001 sha256:3e420bc9960d2f42169f0bfda8ec50fd3bb752ae0be8d52c7760cd260b32b373
- GEO-GEOHASH-DEC-001 sha256:0e13343cad9cf9d001adc08f486dfb94bbdc311b2bdfad8936b97a68f27d67cc

[Decision register](docs/specification-decisions.md)

## 1.0.0 - 2026-08-25

### Changed

- Exclude intentional nested modules from root local-proxy archives so local,
  bootstrap, CI, and public module checksums describe the same source
  boundary.

- Track the pinned documentation-tool lockfile so clean CI checkouts install
  the exact validated cspell dependency.

- Reconcile standalone dependency checksums against deterministic current
  module archives so CI, local verification, and release consumers resolve
  identical content.

- Harden standalone documentation validation with deterministic spelling and
  link checks, package-specific documentation gates, and repository-local
  contributor guidance.

### Documentation

- Correct stale package, standalone, and authoritative-source links in public
  documentation.

### Documentation

- Link the package README to package-owned documentation.

### Changed

- Publish the module from its standalone `github.com/faustbrian/go-geo` identity while preserving its documented API and behavior.
- Execute API compatibility tooling against the isolated module graph so owned
  dependency source changes cannot conflict with release checksums.
- Require Go 1.25.12 or newer so consumers receive the standard library
  security fixes covered by the vulnerability gate.
- Use the repository-pinned current `apidiff` revision for the canonical API
  compatibility gate.

### Added

- Validated immutable scalar values, explicit CRS metadata, antimeridian-safe
  bounds, complete 2D geometry, and typed errors.
- Named mean-earth spherical and WGS84 ellipsoidal geodesy, bearings,
  destinations, envelopes, measurements, and in-memory nearest ranking.
- Bounded GeoJSON/Feature, WKT/EWKT, and WKB/EWKB codecs.
- Geohash indexing helpers, geom conversion, pgx/PostGIS codecs, and safe
  spatial SQL fragments.
- Authoritative, property, differential, hostile-input, fuzz, race, allocation,
  benchmark, and live PostGIS verification with exact statement coverage.
- Adoption, mathematical, interoperability, security, performance, migration,
  troubleshooting, contribution, compatibility, and release documentation.
- MIT licensing for open-source use, modification, and distribution.
- CI execution of every runnable example, not only package compilation.
- Reproducible API compatibility checks with a pinned `apidiff` release.
- A checked-in codec/PostGIS interoperability corpus, durable fuzz corpora,
  expanded GeographicLib edge vectors, allocation regression budgets, and a
  published numerical/dependency hardening matrix.
- A checked-in pre-release API baseline so compatibility checks remain
  substantive before the first release.

### Fixed

- Enforce exact geometry, codec, identifier, depth, size, and placeholder
  boundaries, including bounded WKT/WKB iteration for hostile input.
- Exact ellipsoidal antipodes and opposite poles now report undefined bearings
  instead of presenting one non-unique azimuth as meaningful.
- Bound and check pgx example connection shutdown, and keep equivalent GeoJSON
  and PostGIS validation paths clean under strict static analysis.
- Upgrade `golang.org/x/text` to the latest fixed release to remove
  `GO-2026-5970` from the pgx dependency path.
