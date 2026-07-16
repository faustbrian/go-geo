# Dependency inventory

Direct dependencies are deliberately isolated behind package-owned contracts.
This inventory must be reviewed whenever `go.mod` changes.

| Module | Version | Purpose | License |
| --- | --- | --- | --- |
| `github.com/pymaxion/geographiclib-go/v2` | v2.1.2 | WGS84 Karney geodesics behind `geodesy.Model` | MIT |
| `github.com/peterstace/simplefeatures` | v0.59.0 | Polygon topology validation behind immutable geometry | MIT |
| `github.com/twpayne/go-geom` | v1.6.1 | Optional adapter and independent WKB differential | BSD-2-Clause |
| `github.com/jackc/pgx/v5` | v5.10.0 | Optional PostGIS wire codec and live integration | MIT |

The authoritative license text remains in each dependency module and its
source repository. `go.sum` pins downloaded content; CI runs module-integrity
and vulnerability checks. A dependency upgrade requires tests, fuzzing,
benchmarks, license review, and an entry in `CHANGELOG.md` when user-observable
behavior changes.
