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
install the public `golib` release declared in `.golib.yaml` and verify
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

## Selected Monthly Assurance

[`assurance.yml`](.github/workflows/assurance.yml) runs on main at 08:17 UTC on
the 1st day of each month. It selects one of the 10 declared root-module
package/target pairs using the shared UTC-week selector, not full campaigns or
a guaranteed complete monthly rotation. Duplicate names in different packages
remain separate risks. Explicit main-only dispatch accepts a declared identity
in `package::^FuzzName$` form, for example `./wkb::^FuzzDecode$`.
A bare name is accepted only when unique; inputs are validated as declared
identities and are not forwarded as shell commands.

The selected risks include hostile geometry encodings, constructor/model inputs
and canonical/deprecated adapter parity. Decoder and PostGIS scan targets
exercise panic safety under configured limits while discarding decode results;
they do not prove rejection correctness or live PostGIS interoperability.
Adapter targets also assert value and error parity. Retain affected numerical
vectors, CRS/topology assertions, public consumers and explicit disposable
PostgreSQL/PostGIS qualification with its DSN.

The immutable shared runner bounds active fuzzing to 30 seconds, two workers
and two CPUs, with a five-minute combined native session, 2 GiB memory and
1 MiB output limits inside a ten-minute job. It installs the declared public
CLI without a source build, uses private disposable caches and keeps failures,
timeouts and exhausted budgets visible. Ordinary deterministic tests still run
seeds/corpus; ordinary CI and its Required contract are unchanged. Full/manual
operations and their original budgets remain in `.golib.yaml`.

Treat findings as owner-attributed diagnostics: promptly repair viable defects
with focused regressions; retain reviewed narrow equivalent/invalid dispositions
rather than weakening tests. Review cadence against useful findings and actual
cost; do not interpret a selected pass as full or release qualification.

## Adding A Module

Follow [repository structure policy](AGENTS.md#repository-structure). New modules
require an explicit purpose, ownership boundary, dependency review, package
catalog entry, full quality gates, documentation, changelog, license, security
policy, compatibility plan, and release dry-run.

## Source Documentation

Install the checksum-pinned public verifier declared in `.golib.yaml`.
`make source-docs-check` validates objective Go comment, marker, link and
generated-producer structure; `make check`, `make check-full` and `make ci`
require it before their existing checks. Shared CI runs the same strict
contract and includes its result in Required. This is structural enforcement,
not a substitute for source-truth review or behavioral verification.
