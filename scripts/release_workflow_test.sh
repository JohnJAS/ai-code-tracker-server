#!/usr/bin/env bash
set -euo pipefail

workflow=".github/workflows/release.yml"

require() {
  grep -F -- "$1" "$workflow" > /dev/null || {
    printf 'missing workflow content: %s\n' "$1" >&2
    exit 1
  }
}

require 'tags: ["v*"]'
require 'packages: write'
require 'go test ./...'
require 'ghcr.io/${{ github.repository }}'
require 'platforms: linux/amd64'
require 'type=registry'
require 'type=docker,dest=/tmp/ai-code-tracker-server-${{ github.ref_name }}-linux-amd64.tar'
require 'actions/upload-artifact@v4'

printf 'PASS\n'
