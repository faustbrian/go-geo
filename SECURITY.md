# Security policy

Please report vulnerabilities privately through GitHub's security advisory
feature for this repository. Do not open a public issue until a fix and release
are available. Include affected versions, a minimal reproducer, impact, and any
suggested resource limit or mitigation.

Supported versions will be listed here after the first release. Until then,
security fixes target the latest commit on `main`.

The principal risk areas are hostile codec input, aggregate geometry exhaustion,
integer overflow, parser panics, SQL-fragment misuse, and dependency
vulnerabilities. CI runs bounded fuzz smoke tests, race tests, exact coverage,
and `govulncheck`; releases additionally require live PostGIS integration.
