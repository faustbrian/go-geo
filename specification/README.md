# Specification conformance matrix

The [specification decision register](../docs/specification-decisions.md)
records rationale and consequences. Machine-readable evidence bindings live in
`conformance.json`; authority monitoring and source pins live beside it.

| Source | Decision | Executable and external evidence | Disposition |
| --- | --- | --- | --- |
| RFC 7946 | GEO-GEOJSON-DEC-001 | GeoJSON coordinate, CRS, and fuzz checks | Resolved normative wire rule |
| RFC 7946 | GEO-GEOJSON-DEC-002 | Geometry-family, Feature, hostile-shape, and fuzz checks | Resolved package subset |
| OGC 99-049 | GEO-OGC-DEC-001 | OGC polygon vectors and PostGIS provider corpus | Resolved planar profile |
| OGC 06-103r4 | GEO-OGC-DEC-002 | WKT/WKB round trips, PostGIS corpus, and go-geom differential | Resolved two-dimensional profile |
| PostGIS 3.6.4 | GEO-POSTGIS-DEC-001 | EWKT SRID tests and PostGIS provider corpus | Resolved extension profile |
| PostGIS 3.6.4 | GEO-POSTGIS-DEC-002 | EWKB SRID tests, PostGIS corpus, and go-geom differential | Resolved extension profile |
| EPSG Dataset v13.102 | GEO-EPSG-DEC-001 | GeographicLib vectors and PostGIS geography corpus | Resolved model policy |
| EPSG Dataset v13.102 | GEO-GEOHASH-DEC-001 | Canonical vector, boundary, cover, and fuzz checks | Resolved informal algorithm profile |

The offline specification gate validates every binding and fails for missing,
contradictory, stale, unresolved, or untested decisions. The online form also
revalidates the pinned public authority bodies and change authorities. Live
PostGIS execution remains the provider boundary exercised by the repository's
integration lane; the static register does not claim that an offline check ran
that service.
