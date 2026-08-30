# Specification decisions

This register records every observable standards interpretation. The machine
contract is `../specification/decisions.json`; conformance bindings are in
`../specification/conformance.json`. Unknown behavior remains unresolved rather
than becoming an undocumented default.

## GEO-GEOJSON-DEC-001: Longitude-latitude axis and CRS boundary

- **Status, owner, classification, scope:** resolved; go-geo maintainers; interoperability policy; normative.
- **Specification and authority:** RFC 7946 GeoJSON; RFC 7946; rfc7946-source; section 4 and 11.2; requirement strength not specified. Authoritative URL: https://www.rfc-editor.org/rfc/rfc7946.txt
- **Issue:** RFC 7946 fixes positions as longitude then latitude in WGS 84, while EPSG:4326 registry axis metadata lists latitude before longitude and generic CRS-aware APIs can otherwise invite an axis swap.
- **Credible interpretations:** Follow EPSG registry axis order in every representation. | Follow each wire format's axis contract and make the package coordinate order explicit. | Infer axis order from caller data.
- **Peer behavior:** PostGIS and Go geometry libraries expose both CRS-aware and GeoJSON-specific surfaces; their surrounding APIs do not define one package-wide automatic axis swap.
- **Selected behavior:** All package coordinate pairs are x then y and are named longitude then latitude; GeoJSON requires exact EPSG:4326 metadata, emits that order, and never transforms or swaps ordinates.
- **Rationale:** The GeoJSON wire contract is explicit, and silent axis inference would make identical numbers change meaning across codecs.
- **Security consequences:** Explicit order prevents location-policy bypasses caused by ambiguous or silently swapped coordinates.
- **Resource consequences:** CRS validation is constant-time and performs no registry or network lookup.
- **Compatibility consequences:** Callers with latitude-longitude tuples or projected coordinates must normalize and transform before construction.
- **Wire consequences:** GeoJSON positions are always longitude then latitude and contain no package-added CRS member.
- **Executable evidence:** TestPointUsesLongitudeLatitudeOrder, TestDecodeIsBoundedAndCRSExplicit, TestMarshalRejectsNilAndNonWGS84Geometry
- **Fixture evidence:** none.
- **Fuzz evidence:** FuzzDecode
- **Interoperability evidence:** none.
- **Differential evidence:** none.
- **Public APIs:** geo.NewCoordinate, geo.WGS84, geojson.Marshal, geojson.Unmarshal
- **Documentation:** docs/specification-decisions.md, docs/interoperability.md, docs/models-and-precision.md
- **Upstream status:** RFC 7946 deliberately uses the traditional GIS longitude-latitude order and notes the EPSG axis-order difference.
- **Reconsider when:** RFC 7946 is superseded or a versioned non-GeoJSON axis profile is introduced.

## GEO-GEOJSON-DEC-002: Two-dimensional geometry and Feature profile

- **Status, owner, classification, scope:** resolved; go-geo maintainers; optional behavior; application-policy.
- **Specification and authority:** RFC 7946 GeoJSON; RFC 7946; rfc7946-source; section 3.1, 3.2, 3.3, and 3.4; requirement strength MAY. Authoritative URL: https://www.rfc-editor.org/rfc/rfc7946.txt
- **Issue:** RFC 7946 permits optional third or additional position elements, bounding boxes, foreign members, and FeatureCollection, while the package model is deliberately two-dimensional and exposes geometry and individual Feature codecs only.
- **Credible interpretations:** Accept and discard unsupported dimensions and members. | Expand the model to every optional GeoJSON object. | Expose a strict two-dimensional geometry and Feature subset and reject unsupported coordinate dimensions.
- **Peer behavior:** Maintained GeoJSON implementations vary between permissive document preservation and typed two-dimensional models.
- **Selected behavior:** The package round-trips all seven geometry families plus individual Feature objects, preserves Feature IDs and raw properties, rejects positions that are not exactly two finite ordinates, and does not claim FeatureCollection or bbox support.
- **Rationale:** Rejecting unrepresentable dimensions avoids silent data loss while keeping the public immutable model explicit.
- **Security consequences:** Unsupported dimensions cannot be hidden from validation by being discarded.
- **Resource consequences:** Encoded bytes, point counts, geometry counts, and collection depth are bounded before retained allocation.
- **Compatibility consequences:** Documents using altitude, measures, FeatureCollection, or bbox require a separate document layer.
- **Wire consequences:** Emitted geometry and Feature objects contain only the represented RFC members and two-element positions.
- **Executable evidence:** TestEveryGeometryFamilyRoundTrips, TestFeaturePreservesRawPropertiesAndID, TestGeometryDecoderRejectsMalformedCoordinatesAndTopology
- **Fixture evidence:** none.
- **Fuzz evidence:** FuzzDecode
- **Interoperability evidence:** none.
- **Differential evidence:** none.
- **Public APIs:** geojson.Marshal, geojson.Unmarshal, geojson.NewFeature, geojson.MarshalFeature, geojson.UnmarshalFeature
- **Documentation:** docs/specification-decisions.md, docs/interoperability.md
- **Upstream status:** The omitted surfaces and additional position elements remain optional or separate object types in RFC 7946.
- **Reconsider when:** The public geometry model gains explicit additional dimensions or a versioned GeoJSON document API.

## GEO-OGC-DEC-001: Planar topology, winding, and boundary semantics

- **Status, owner, classification, scope:** resolved; go-geo maintainers; implementation-defined behavior; application-policy.
- **Specification and authority:** OGC Simple Features for SQL 1.1; OGC 99-049; ogc-sfs-1.1-source; section 2.1.5, 2.1.8, and the polygon examples; requirement strength not specified. Authoritative URL: https://docs.ogc.org/is/99-049/99-049.pdf
- **Issue:** Simple Features defines polygon interior, boundary, exterior, ring closure, and validity, but does not select this package's antimeridian unwrapping, winding-preservation, or exact floating-point boundary policy.
- **Credible interpretations:** Treat longitude-latitude edges as geodesics. | Apply planar topology directly to stored ordinates. | Unwrap antimeridian-crossing rings into one local planar domain while preserving input winding.
- **Peer behavior:** PostGIS agrees with the checked interior, exterior, shell-boundary, and hole-boundary vectors; planar peers can differ when geographic edges or tolerance rules are introduced.
- **Selected behavior:** Polygon validation and Locate use planar straight edges after deterministic antimeridian unwrapping, reject open, degenerate, self-intersecting, overlapping, or exterior holes, preserve either valid winding direction, and return Boundary for shell and hole edges.
- **Rationale:** This preserves the Simple Features three-state topology while making the non-geodesic longitude-latitude approximation observable.
- **Security consequences:** Invalid rings fail closed instead of producing unstable containment classifications.
- **Resource consequences:** Topology work is bounded by caller-selected point, ring, geometry, and depth limits.
- **Compatibility consequences:** Callers needing ellipsoidal polygon edges or tolerance-based boundary tests must use another explicitly selected engine.
- **Wire consequences:** Valid winding is preserved rather than normalized; topology rejection can prevent serialization of invalid input.
- **Executable evidence:** TestPolygonLocationMatchesOGCSimpleFeaturesVectors, TestPolygonRejectsDegenerateSelfIntersectingAndOutsideHoleTopology, TestPolygonAcceptsAndPreservesEitherWindingDirection
- **Fixture evidence:** postgis/testdata/interoperability.json
- **Fuzz evidence:** FuzzGeometryConstructors
- **Interoperability evidence:** postgis/testdata/interoperability.json
- **Differential evidence:** none.
- **Public APIs:** geo.NewPolygon, geo.Polygon.Locate
- **Documentation:** docs/specification-decisions.md, docs/models-and-precision.md, docs/verification.md
- **Upstream status:** The OGC topology model is stable; antimeridian and numerical policies remain package-owned.
- **Reconsider when:** A versioned geodesic polygon model or tolerance policy is added.

## GEO-OGC-DEC-002: Two-dimensional WKT and WKB profile

- **Status, owner, classification, scope:** resolved; go-geo maintainers; optional behavior; application-policy.
- **Specification and authority:** OGC Simple Feature Access Part 1: Common Architecture 1.2.1; OGC 06-103r4; ogc-sfa-1.2.1-source; section 6.1.2 and 6.1.3; requirement strength not specified. Authoritative URL: https://docs.ogc.org/is/06-103r4/06-103r4.pdf
- **Issue:** Simple Feature Access defines text and binary representations across dimensional variants and empty geometries, while the package model stores finite XY coordinates and cannot represent empty primitive Point, LineString, or Polygon values.
- **Credible interpretations:** Accept every dimension and discard unsupported ordinates. | Represent unsupported primitive empties with hidden sentinel coordinates. | Support exact two-dimensional representations and empty aggregate geometries while rejecting unrepresentable values.
- **Peer behavior:** go-geom v1.6.1 produces identical checked Point WKB bytes in both byte orders; broader peers support additional dimensions and primitive empty values.
- **Selected behavior:** WKT and WKB support the seven two-dimensional geometry families, caller-selected WKB byte order, caller-supplied CRS on decode, and empty MultiPoint, MultiLineString, MultiPolygon, and GeometryCollection; Z, M, ZM, non-finite coordinates, and empty primitive geometries are rejected.
- **Rationale:** Rejecting values the immutable model cannot preserve prevents dimensional or emptiness data loss.
- **Security consequences:** Unsupported flags, non-finite values, hostile counts, trailing input, and inconsistent nested types fail closed.
- **Resource consequences:** Encoded bytes, counts, and recursive geometry depth are bounded before allocation.
- **Compatibility consequences:** Consumers needing Z, M, ZM, or primitive empties require a different model or an explicitly versioned future extension.
- **Wire consequences:** Plain WKT and WKB never carry SRID metadata; WKB may be NDR or XDR as selected by the caller.
- **Executable evidence:** TestEveryGeometryFamilyAndEmptyCollectionRoundTrips, TestEveryGeometryFamilyRoundTripsInBothByteOrders, TestWKTRejectsMalformedDimensionsTopologyAndResourceExhaustion, TestDecoderRejectsHostileCountsTruncationAndTrailingBytes
- **Fixture evidence:** postgis/testdata/interoperability.json
- **Fuzz evidence:** FuzzDecode
- **Interoperability evidence:** postgis/testdata/interoperability.json
- **Differential evidence:** wkb/differential_test.go
- **Public APIs:** wkt.Marshal, wkt.Unmarshal, wkb.Marshal, wkb.Unmarshal
- **Documentation:** docs/specification-decisions.md, docs/interoperability.md, docs/models-and-precision.md
- **Upstream status:** OGC 06-103r4 remains the pinned representation authority; the supported subset is package policy.
- **Reconsider when:** The public geometry model gains explicit additional dimensions or primitive empty states.

## GEO-POSTGIS-DEC-001: EWKT SRID prefix and dimensional boundary

- **Status, owner, classification, scope:** resolved; go-geo maintainers; optional behavior; extension-specific.
- **Specification and authority:** PostGIS 3.6.4 EWKT and EWKB extensions; PostGIS 3.6.4; postgis-ewkt-source; section EWKT BNF; requirement strength not specified. Authoritative URL: https://raw.githubusercontent.com/postgis/postgis/3.6.4/doc/bnf-wkt.txt
- **Issue:** PostGIS EWKT extends WKT with SRID and dimensional forms, while plain WKT has no CRS and the package supports only finite two-dimensional geometry.
- **Credible interpretations:** Treat a missing SRID as zero. | Accept every PostGIS dimensional form. | Require one positive 32-bit SRID prefix and retain the package's strict two-dimensional profile.
- **Peer behavior:** PostGIS accepts EWKT with SRID metadata and additional dimensional forms; the package intentionally agrees only on its supported two-dimensional subset.
- **Selected behavior:** MarshalEWKT emits SRID=<positive integer>; before canonical WKT, UnmarshalEWKT requires that prefix, maps 4326 to WGS84 metadata, preserves other positive 32-bit SRIDs without transformation, and rejects Z, M, and ZM.
- **Rationale:** A mandatory SRID distinguishes the extended codec from plain WKT and prevents silently assigning unknown CRS metadata.
- **Security consequences:** Malformed, missing, zero, negative, or overflowing SRIDs fail closed.
- **Resource consequences:** The prefix and remaining WKT share the encoded-byte and structural limits.
- **Compatibility consequences:** Permissive EWKT without an SRID and dimensional EWKT are not accepted.
- **Wire consequences:** EWKT always contains exactly one top-level positive SRID prefix and canonical two-dimensional WKT.
- **Executable evidence:** TestPointWKTAndEWKTPreserveCoordinateOrderAndSRID, TestEWKTRequiresPositiveSRID, TestWKTRejectsInvalidMetadataAndEWKTPrefixes
- **Fixture evidence:** postgis/testdata/interoperability.json
- **Fuzz evidence:** FuzzDecode
- **Interoperability evidence:** postgis/testdata/interoperability.json
- **Differential evidence:** none.
- **Public APIs:** wkt.MarshalEWKT, wkt.UnmarshalEWKT
- **Documentation:** docs/specification-decisions.md, docs/interoperability.md
- **Upstream status:** The PostGIS extension supports a wider dimensional surface than this package.
- **Reconsider when:** PostGIS changes its EWKT grammar or the package gains explicit additional dimensions.

## GEO-POSTGIS-DEC-002: EWKB top-level SRID ownership

- **Status, owner, classification, scope:** resolved; go-geo maintainers; optional behavior; extension-specific.
- **Specification and authority:** PostGIS 3.6.4 EWKT and EWKB extensions; PostGIS 3.6.4; postgis-ewkb-source; section EWKB BNF; requirement strength not specified. Authoritative URL: https://raw.githubusercontent.com/postgis/postgis/3.6.4/doc/bnf-wkb.txt
- **Issue:** EWKB permits SRID flags on geometry records, creating choices about missing top-level SRIDs, repeated child SRIDs, and mismatched nested metadata.
- **Credible interpretations:** Accept a missing SRID as zero. | Let each nested geometry replace inherited CRS metadata. | Require one positive top-level SRID and require any repeated nested SRID to match.
- **Peer behavior:** go-geom v1.6.1 emits identical checked Point EWKB bytes, and PostGIS 3.5 and 3.6 agree with the complete live codec corpus.
- **Selected behavior:** MarshalEWKB writes a positive SRID on the top-level geometry, children inherit it, UnmarshalEWKB requires it, and any nested SRID must match; plain WKB and EWKB decoders reject each other's CRS contract.
- **Rationale:** One inherited CRS preserves collection invariants and keeps plain and extended binary APIs unambiguous.
- **Security consequences:** Conflicting nested metadata, dimensional flags, hostile counts, and trailing bytes fail closed.
- **Resource consequences:** Byte length, counts, depth, and allocation are bounded before retention.
- **Compatibility consequences:** EWKB streams with no top-level SRID or intentionally mixed child SRIDs are rejected.
- **Wire consequences:** The top-level type sets the PostGIS SRID flag and carries one positive 32-bit SRID in caller-selected byte order.
- **Executable evidence:** TestPointMatchesCanonicalLittleEndianWKBAndEWKB, TestWKBAndEWKBSRIDContractsAreDistinct, TestEWKBRejectsInvalidAndMismatchedSRIDs, TestPointEncodingMatchesGoGeom
- **Fixture evidence:** postgis/testdata/interoperability.json
- **Fuzz evidence:** FuzzDecode
- **Interoperability evidence:** postgis/testdata/interoperability.json
- **Differential evidence:** wkb/differential_test.go
- **Public APIs:** wkb.MarshalEWKB, wkb.UnmarshalEWKB, postgis.Value, postgis.Codec
- **Documentation:** docs/specification-decisions.md, docs/interoperability.md, docs/verification.md
- **Upstream status:** The checked PostGIS release lines and go-geom peer agree on the supported two-dimensional EWKB subset.
- **Reconsider when:** PostGIS changes EWKB SRID inheritance or a versioned mixed-CRS collection model is introduced.

## GEO-EPSG-DEC-001: WGS 84 model selection without CRS transformation

- **Status, owner, classification, scope:** resolved; go-geo maintainers; omission; application-policy.
- **Specification and authority:** EPSG Geodetic Parameter Dataset v13.102; EPSG Dataset v13.102; epsg-4326-source; section EPSG:4326; requirement strength not specified. Authoritative URL: https://epsg.org/crs/wkt/id/4326
- **Issue:** EPSG:4326 identifies the WGS 84 geographic CRS and ellipsoid but does not select whether a utility API should transform other CRSs, use ellipsoidal or spherical distance, or silently relabel ordinates.
- **Credible interpretations:** Transform every known SRID internally. | Treat every angular coordinate as WGS 84. | Require explicit EPSG:4326 and expose separately named spherical and ellipsoidal models without transformation.
- **Peer behavior:** PostGIS provides explicit geometry and geography casts and agrees with the checked WGS84 geography results; GeographicLib vectors agree with the ellipsoidal implementation.
- **Selected behavior:** geo.WGS84 returns explicit EPSG:4326 metadata; geodesy and geography helpers require it, never transform coordinates, WGS84Ellipsoid uses the WGS 84 ellipsoid, and MeanEarthSphere remains an explicitly approximate IUGG-radius model.
- **Rationale:** Model and transformation choices materially affect distance and must remain visible at the call site.
- **Security consequences:** Foreign or mislabeled CRS input fails instead of producing plausible but incorrect authorization distances.
- **Resource consequences:** No CRS database, grid file, or network lookup occurs; numerical work is bounded per operation.
- **Compatibility consequences:** Projected inputs must be transformed by an explicitly selected external component before geodesy or geography use.
- **Wire consequences:** SRIDs are preserved as metadata by extended codecs but no coordinate bytes are transformed.
- **Executable evidence:** TestWGS84MatchesGeographicLibAuthoritativeExamples, TestWGS84MatchesGeographicLibEdgeDistanceVectors, TestGeodesyRejectsCRSMismatchAndMarksCoincidentBearingsUndefined, TestPostGISIntegration
- **Fixture evidence:** postgis/testdata/interoperability.json
- **Fuzz evidence:** FuzzModels
- **Interoperability evidence:** postgis/testdata/interoperability.json
- **Differential evidence:** none.
- **Public APIs:** geo.WGS84, geodesy.WGS84Ellipsoid, geodesy.MeanEarthSphere, postgis.GeographyDWithin
- **Documentation:** docs/specification-decisions.md, docs/models-and-precision.md, docs/verification.md
- **Upstream status:** EPSG Dataset v13.102 retains EPSG:4326 as the WGS 84 ensemble geographic 2D CRS.
- **Reconsider when:** EPSG revises code 4326 materially or an explicit transformation subsystem is introduced.

## GEO-GEOHASH-DEC-001: Lowercase bounded geohash profile

- **Status, owner, classification, scope:** resolved; go-geo maintainers; interoperability policy; application-policy.
- **Specification and authority:** EPSG Geodetic Parameter Dataset v13.102; EPSG Dataset v13.102; epsg-4326-source; section EPSG:4326; requirement strength not specified. Authoritative URL: https://epsg.org/crs/wkt/id/4326
- **Issue:** Geohash has no normative standards body specification, and common implementations vary in accepted case, maximum precision, pole handling, and whether cells imply metric proximity.
- **Credible interpretations:** Accept implementation-specific spellings and arbitrary precision. | Treat geohash distance as geographic distance. | Publish a narrow package profile over explicit EPSG:4326 angular coordinates.
- **Peer behavior:** Common geohash implementations use the same lowercase base32 alphabet but differ in permissive input and operational limits; no peer behavior is treated as normative.
- **Selected behavior:** Encode and Decode use the lowercase geohash alphabet at precision 1 through 12 over explicit EPSG:4326 longitude-latitude coordinates; neighbors wrap the antimeridian and clamp poles, Cover is caller-bounded, and cells are indexing hints rather than distance proof.
- **Rationale:** A narrow explicit profile makes stored keys deterministic and prevents an informal algorithm from being presented as a geographic accuracy standard.
- **Security consequences:** Invalid case, alphabet, precision, and CRS fail closed; callers must verify proximity independently.
- **Resource consequences:** Precision is capped at 12 and Cover requires a positive maximum cell count.
- **Compatibility consequences:** Uppercase or longer geohashes accepted elsewhere must be normalized or rejected before use.
- **Wire consequences:** Hash strings are one to twelve lowercase ASCII characters from the canonical geohash alphabet.
- **Executable evidence:** TestEncodeMatchesCanonicalGeohashVector, TestGeohashRejectsInvalidPrecisionHashAndCRS, TestCoverIsBoundedAndHandlesAntimeridian
- **Fixture evidence:** none.
- **Fuzz evidence:** FuzzDecode
- **Interoperability evidence:** none.
- **Differential evidence:** none.
- **Public APIs:** geohash.Encode, geohash.Decode, geohash.Neighbors, geohash.Cover
- **Documentation:** docs/specification-decisions.md, docs/adoption.md, docs/models-and-precision.md
- **Upstream status:** No normative geohash authority or errata process exists; EPSG:4326 is pinned only for the coordinate reference boundary.
- **Reconsider when:** A normative geohash specification is published or a versioned alternative profile is introduced.
