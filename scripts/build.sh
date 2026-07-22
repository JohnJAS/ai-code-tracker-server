#!/usr/bin/env bash
set -euo pipefail

image_name="${IMAGE_NAME:-ai-code-tracker-server}"
repo_root="$(git rev-parse --show-toplevel)"
version="${1:-latest}"
context="$repo_root"
temporary_worktree=""

cleanup() {
  if [[ -n "$temporary_worktree" ]]; then
    git -C "$repo_root" worktree remove --force "$temporary_worktree"
  fi
}
trap cleanup EXIT

if [[ "$version" != "latest" ]]; then
  git -C "$repo_root" rev-parse --verify --quiet "refs/tags/$version^{commit}" > /dev/null || {
    printf 'Git tag not found: %s\n' "$version" >&2
    exit 1
  }
  temporary_worktree="$(mktemp -d)"
  rmdir "$temporary_worktree"
  git -C "$repo_root" worktree add --quiet --detach "$temporary_worktree" "refs/tags/$version"
  context="$temporary_worktree"
fi

docker build -t "$image_name:$version" "$context"
