# Contributing

Open an issue before proposing a new geometry family, CRS feature, numerical
model, or dependency. Keep business rules such as addresses, carriers,
geocoding, routing, and vendor APIs outside this module.

Every change must preserve explicit longitude/latitude order, CRS behavior,
immutability, typed errors, and resource bounds. Behavioral changes should begin
with a failing test. Numerical claims require authoritative or independent
evidence; codec changes require hostile cases and fuzzing.

Run before submitting:

```sh
gofmt -w <changed-go-files>
go vet ./...
test -z "$(go run golang.org/x/lint/golint@v0.0.0-20241112194109-818c5a804067 ./...)"
go test ./...
go test -race ./...
./scripts/check-coverage.sh
./scripts/fuzz-smoke.sh
go test ./... -run '^$' -bench . -benchtime=1x
```

PostGIS changes also require `POSTGIS_DSN=... go test ./postgis -run
TestPostGISIntegration -count=1`. Update public documentation and
`CHANGELOG.md`. Use conventional commits with a body that explains why.
