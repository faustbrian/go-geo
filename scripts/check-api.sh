#!/usr/bin/env bash
set -euo pipefail

module="$(go list -m -f '{{.Path}}')"
root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
baseline="${1:-$root/api/baseline.txt}"
tool="golang.org/x/exp/cmd/apidiff@v0.0.0-20260709172345-9ea1abe57597"

if [[ ! -f "$baseline" ]]; then
    echo "API compatibility baseline not found: $baseline" >&2
    exit 1
fi

go run "$tool" -m "$baseline" "$module"
