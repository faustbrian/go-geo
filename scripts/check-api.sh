#!/usr/bin/env bash
set -euo pipefail

tag="${1:-$(git describe --tags --abbrev=0 2>/dev/null || true)}"
if [[ -z "$tag" ]]; then
    echo "API compatibility: no release tag exists; skipping baseline comparison"
    exit 0
fi

module="$(go list -m -f '{{.Path}}')"
go run golang.org/x/exp/cmd/apidiff@latest -m "${module}@${tag}" "$module"
