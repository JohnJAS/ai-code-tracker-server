# GitHub Release Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Create a GitHub Release with generated notes and a Linux AMD64 Docker archive whenever a `v*` tag successfully completes the release workflow.

**Architecture:** The existing workflow continues publishing GHCR images and the Actions artifact. A final Release action receives the same tar file from `/tmp`, so the workflow needs `contents: write` in addition to its existing package permission.

**Tech Stack:** GitHub Actions, softprops/action-gh-release v2, Docker Buildx, GHCR, Bash contract test.

---

### Task 1: Extend the workflow contract test

**Files:**
- Modify: `scripts/release_workflow_test.sh`
- Test: `scripts/release_workflow_test.sh`

- [ ] **Step 1: Write the failing release assertions**

Add these required strings to the script's `require` checks: `contents: write`, `softprops/action-gh-release@v2`, `generate_release_notes: true`, and `files: /tmp/ai-code-tracker-server-${{ github.ref_name }}-linux-amd64.tar`.

- [ ] **Step 2: Run the contract test to verify it fails**

Run: `"C:/Program Files/Git/bin/bash.exe" scripts/release_workflow_test.sh`

Expected: exit code 1 and a missing release-related string.

- [ ] **Step 3: Commit the failing test**

Run `git add scripts/release_workflow_test.sh`, then `git commit -m "test: 定义 GitHub Release 发布约定"`.

### Task 2: Create the GitHub Release and attach the archive

**Files:**
- Modify: `.github/workflows/release.yml:8-49`
- Test: `scripts/release_workflow_test.sh`

- [ ] **Step 1: Add workflow permission and Release step**

Set the workflow permissions to include both `contents: write` and `packages: write`. After `actions/upload-artifact@v4`, append this step:

```yaml
      - uses: softprops/action-gh-release@v2
        with:
          generate_release_notes: true
          files: /tmp/ai-code-tracker-server-${{ github.ref_name }}-linux-amd64.tar
```

- [ ] **Step 2: Verify the contract turns green**

Run: `"C:/Program Files/Git/bin/bash.exe" scripts/release_workflow_test.sh`

Expected: `PASS`.

- [ ] **Step 3: Lint the workflow**

Run: `go run github.com/rhysd/actionlint/cmd/actionlint@v1.7.7 .github/workflows/release.yml`

Expected: exit code 0 with no output.

- [ ] **Step 4: Commit the workflow**

Run `git add .github/workflows/release.yml scripts/release_workflow_test.sh`, then `git commit -m "ci: 自动创建 GitHub Release"`.

### Task 3: Document the permanent release asset

**Files:**
- Modify: `README.md:36-58`
- Test: `README.md`

- [ ] **Step 1: Update release documentation**

State that each new tag creates a GitHub Release with automatically generated notes and a permanent `ai-code-tracker-server-<tag>-linux-amd64.tar` asset. Direct archive consumers to download that Release asset, then run `docker load -i` as already documented. Preserve the note that existing tags do not gain a Release retroactively.

- [ ] **Step 2: Run all local checks**

Run the contract test, Actionlint, `go test ./...`, and `git diff --check`; each command must exit 0.

- [ ] **Step 3: Commit the documentation**

Run `git add README.md`, then `git commit -m "docs: 说明 GitHub Release 产物"`.
