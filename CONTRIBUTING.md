# Contributing

## Before Editing

1. Read [`AGENTS.md`](AGENTS.md) and the affected package's documentation.
2. Run ordinary Go tests for the affected package before editing.
3. Identify owned dependencies and reverse dependants in `modules.json`.
4. Preserve unrelated work and generated/corpus provenance.

## Changes

Keep commits focused and conventional. Update every affected changelog with
the behavior and migration impact. Public API changes require compatibility
evidence and documentation. Specification behavior requires a decision record,
fixture coverage, and interoperability evidence.

New direct dependencies and dependency updates must follow the
[dependency governance policy](AGENTS.md#dependencies-and-supply-chain). Package-local
update bots are forbidden; the root policy owns every module and action update.

Specification-backed changes must follow the
[specification governance contract](AGENTS.md#design), update
the affected entries in the
[specification decision register](docs/specification-decisions.md), and
complete the Specification Decisions section of the pull request template. An
unresolved interpretation or stale source pin is release-blocking; peer
behavior cannot silently select policy.

Selected mutation gates must finish with zero surviving viable mutants.

Do not add package-local workflows, permanent replacements, machine-specific
paths, bypass flags, broad mutation exclusions, or aggregate quality metrics
that hide a failing package.

## Verification

Use Go 1.27.2 for development and ordinary CI; the public minimum remains
Go 1.27.0. An ordinary standalone checkout is sufficient for tests and a
focused active fuzz smoke, without sibling repositories or goal ledgers:

```bash
go test ./...
go test ./wkb -run '^$' -fuzz '^FuzzDecode$' -fuzztime=2s
```

Select the native fuzz target and package affected by your change. Ordinary
tests run deterministic fuzz seeds and saved corpora. For shared checks,
install the public `golib` v2.1.1 release declared in `.golib.yaml` and verify
its checksums:

```bash
make check
make ci
```

Both Make routes use bounded local checks. Base-free Make checks run
deterministic seeds/corpora; changed-source CI selects active fuzz smoke.
`make ci` also validates the specification register offline. The root module
declares no services or test tags. The untagged live PostGIS integration test
skips unless `POSTGIS_DSN` is set; ordinary results without it do not prove
live database interoperability.

Optional mutation analysis can select one affected package:

```bash
golib mutation --module . --package ./wkb
```

For PostGIS codec, SQL or numerical comparison changes, use a disposable
database from the supported PostgreSQL/PostGIS 16/3.5 and 18/3.6 matrix and
run the selected live assertion with its DSN:

```bash
POSTGIS_DSN='<disposable test database DSN>' go test ./postgis -run '^TestPostGISIntegration$' -count=1
```

Use `make check-full` for explicit online specification validation and full
assurance with the unchanged 10,000-iteration native fuzz budgets, numerical
conformance vectors, go-geom interoperability, API baseline and one-second
benchmarks. The canonical and deprecated geom adapters are packages within
this module, not sibling or nested modules. Live PostGIS qualification still
requires the explicit database environment; no full command manufactures it.
Coverage remains diagnostic as configured. Select directly affected public
consumers and relevant release qualification, including `golib release check`
and `golib release dry-run --all`, when their contract changes. Full campaigns
remain available for selected scheduled/release risks; ordinary scheduled CI
stays bounded and separately monitors specification sources online. Report
unavailable or failing checks without claiming release readiness.
CI/configuration-only changes do not require a library release.

## Adding A Module

Follow [repository structure policy](AGENTS.md#repository-structure). New modules
require an explicit purpose, ownership boundary, dependency review, package
catalog entry, full quality gates, documentation, changelog, license, security
policy, compatibility plan, and release dry-run.
