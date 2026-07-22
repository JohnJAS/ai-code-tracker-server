# GHCR Release Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Publish version-tagged Linux AMD64 images to GHCR and attach an equivalent Docker archive to each tag-triggered GitHub Actions run.

**Architecture:** A `v*` tag workflow runs Go tests before one Buildx build writes two outputs: a registry image and a Docker archive. The repository name is converted to lowercase before it is used as the GHCR path.

**Tech Stack:** GitHub Actions, Docker Buildx, GHCR, Go 1.25, Docker archive exporter.

---

### Task 1: Define the release workflow contract

**Files:**
- Create: `scripts/release_workflow_test.sh`
- Test: `scripts/release_workflow_test.sh`

- [ ] **Step 1: Write the failing contract check**

Create an executable Bash script that reads `.github/workflows/release.yml`. For each required string, run `grep -F -- "$required" "$workflow" > /dev/null`; when any lookup fails, print `missing workflow content: <string>` to stderr and exit 1. The required strings are `tags: ["v*"]`, `packages: write`, `go test ./...`, `id: image_name`, `echo "name=ghcr.io/${GITHUB_REPOSITORY,,}" >> "$GITHUB_OUTPUT"`, `${{ steps.image_name.outputs.name }}:${{ github.ref_name }}`, `platforms: linux/amd64`, `type=registry`, `type=docker,dest=/tmp/ai-code-tracker-server-${{ github.ref_name }}-linux-amd64.tar`, and `actions/upload-artifact@v4`. Print `PASS` after all checks.

- [ ] **Step 2: Run the check to verify it fails**

Run: `"C:/Program Files/Git/bin/bash.exe" scripts/release_workflow_test.sh`

Expected: exit code 1 because the workflow does not exist.

- [ ] **Step 3: Commit the failing test**

Run `git add scripts/release_workflow_test.sh`, `git update-index --chmod=+x scripts/release_workflow_test.sh`, then `git commit -m "test: 定义镜像发布工作流约定"`.

### Task 2: Add the tag-triggered release workflow

**Files:**
- Create: `.github/workflows/release.yml`
- Modify: `scripts/release_workflow_test.sh`
- Test: `scripts/release_workflow_test.sh`

- [ ] **Step 1: Create the workflow**

Use `actions/checkout@v4`, `actions/setup-go@v5` with `go-version-file: go.mod`, then a `run: go test ./...` step. Configure `on.push.tags` as `["v*"]`, workflow permissions as `contents: read` and `packages: write`, and `runs-on: ubuntu-latest`. Use `docker/setup-buildx-action@v3`, then `docker/login-action@v3` with `registry: ghcr.io`, `username: ${{ github.actor }}`, and `password: ${{ secrets.GITHUB_TOKEN }}`.

Use a preceding `id: image_name` Bash step to write `ghcr.io/${GITHUB_REPOSITORY,,}` to `GITHUB_OUTPUT`. Configure `docker/build-push-action@v6` with `context: .`, `platforms: linux/amd64`, two tags based on `${{ steps.image_name.outputs.name }}`, and two outputs (`type=registry` and `type=docker,dest=/tmp/ai-code-tracker-server-${{ github.ref_name }}-linux-amd64.tar`). Finally, use `actions/upload-artifact@v4` to upload that archive as `ai-code-tracker-server-${{ github.ref_name }}-linux-amd64` with `if-no-files-found: error`.

- [ ] **Step 2: Verify the contract turns green**

Run: `"C:/Program Files/Git/bin/bash.exe" scripts/release_workflow_test.sh`

Expected: `PASS`.

- [ ] **Step 3: Lint the workflow**

Run: `go run github.com/rhysd/actionlint/cmd/actionlint@v1.7.7 .github/workflows/release.yml`

Expected: exit code 0 with no output.

- [ ] **Step 4: Commit the workflow**

Run `git add .github/workflows/release.yml scripts/release_workflow_test.sh`, then `git commit -m "ci: 按版本标签发布 GHCR 镜像"`.

### Task 3: Document release consumption

**Files:**
- Modify: `README.md:5-50`
- Test: `README.md`

- [ ] **Step 1: Add release examples**

Document `git tag v1.2.0` and `git push origin v1.2.0`; GHCR startup with `IMAGE_NAME=ghcr.io/johnjas/ai-code-tracker-server APP_VERSION=v1.2.0 docker compose up -d`; and archive startup with `docker load -i ai-code-tracker-server-v1.2.0-linux-amd64.tar` followed by the same Compose command. State that the archive is Linux AMD64, private packages require `docker login ghcr.io`, and each tag run publishes the exact version plus `latest`.

- [ ] **Step 2: Run local verification**

Run the contract script, Actionlint, `go test ./...`, and `git diff --check`; every command must exit 0.

- [ ] **Step 3: Commit documentation**

Run `git add README.md`, then `git commit -m "docs: 说明 GHCR 镜像和发布产物"`.
