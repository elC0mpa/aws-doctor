#!/usr/bin/env bash
set -euo pipefail

summary="${1:-}"
type="${2:-chore}"

if [[ -z "$summary" ]]; then
  echo "Usage: record-change.sh \"summary\" [type]" >&2
  exit 1
fi

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
changelog="$root/CHANGELOG.md"

if [[ ! -f "$changelog" ]]; then
  cat > "$changelog" <<'EOF'
# Changelog

All notable changes to this project are recorded in this file.
This file follows the Keep a Changelog format.

## [Unreleased]

EOF
fi

entry="- $(date +%Y-%m-%d): $type - $summary"

tmp="$(mktemp)"
awk -v entry="$entry" '
  BEGIN { inserted=0; skip_blank=0 }
  /^## \[Unreleased\]$/ {
    print
    print ""
    print entry
    inserted=1
    skip_blank=1
    next
  }
  skip_blank {
    if ($0 ~ /^[[:space:]]*$/) { skip_blank=0; next }
    skip_blank=0
  }
  { print }
  END {
    if (inserted == 0) {
      exit 2
    }
  }
' "$changelog" > "$tmp"
mv "$tmp" "$changelog"
echo "Added entry to CHANGELOG.md"
