#!/usr/bin/env bash
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
temp_dir="$(mktemp -d)"

cleanup() {
  rm -rf "$temp_dir"
}
trap cleanup EXIT

fail() {
  printf 'FAIL: %s\n' "$1" >&2
  exit 1
}

assert_eq() {
  [[ "$1" == "$2" ]] || fail "expected [$2], got [$1]"
}

repo="$temp_dir/repo"
mkdir -p "$repo/scripts" "$temp_dir/bin"
git -C "$repo" init -q
git -C "$repo" config user.email test@example.invalid
git -C "$repo" config user.name test
cp "$root/scripts/build.sh" "$repo/scripts/build.sh"
chmod +x "$repo/scripts/build.sh"
printf 'FROM scratch\n' > "$repo/Dockerfile"
printf 'release\n' > "$repo/VERSION"
git -C "$repo" add .
git -C "$repo" commit -qm release
git -C "$repo" tag v1.2.0
printf 'local\n' > "$repo/VERSION"

cat > "$temp_dir/bin/docker" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
printf '%s|%s\n' "$3" "$(tr -d '\n' < "$4/VERSION")" >> "$DOCKER_LOG"
EOF
chmod +x "$temp_dir/bin/docker"

export PATH="$temp_dir/bin:$PATH"
export DOCKER_LOG="$temp_dir/docker.log"

(cd "$repo" && ./scripts/build.sh)
assert_eq "$(cat "$DOCKER_LOG")" "ai-code-tracker-server:latest|local"

: > "$DOCKER_LOG"
(cd "$repo" && ./scripts/build.sh v1.2.0)
assert_eq "$(cat "$DOCKER_LOG")" "ai-code-tracker-server:v1.2.0|release"

if (cd "$repo" && ./scripts/build.sh v9.9.9) > "$temp_dir/missing-tag.log" 2>&1; then
  fail 'missing Git tag unexpectedly built an image'
fi
grep -Fx 'Git tag not found: v9.9.9' "$temp_dir/missing-tag.log" > /dev/null || fail 'missing tag error was not reported'

printf 'PASS\n'
