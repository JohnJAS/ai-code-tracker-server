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
require 'contents: write'
require 'packages: write'
require 'go test ./...'
require 'id: image_name'
require 'echo "name=ghcr.io/${GITHUB_REPOSITORY,,}" >> "$GITHUB_OUTPUT"'
require '${{ steps.image_name.outputs.name }}:${{ github.ref_name }}'
require '${{ steps.image_name.outputs.name }}:latest'
require 'platforms: linux/amd64'
require 'type=registry'
require 'type=docker,dest=/tmp/ai-code-tracker-server-${{ github.ref_name }}-linux-amd64.tar'
require 'actions/upload-artifact@v4'
require 'softprops/action-gh-release@v2'
require 'generate_release_notes: true'
require 'files: /tmp/ai-code-tracker-server-${{ github.ref_name }}-linux-amd64.tar'

printf 'PASS\n'
