# GitHub Release Design

## Goal

Make each future version tag produce a browsable GitHub Release in addition to
the existing GHCR images and Actions artifact. The Linux AMD64 Docker archive
must be downloadable from the release page as a permanent release asset.

## Trigger and Permissions

The existing `v*` tag trigger remains unchanged. The workflow keeps
`packages: write` for GHCR and adds `contents: write` so its `GITHUB_TOKEN`
can create a Release and upload assets.

## Release Flow

After Go tests, registry publishing, archive generation, and Actions artifact
upload have succeeded, the workflow uses a GitHub Release action to create a
release whose tag and title are both `github.ref_name`. It enables GitHub's
`generate_release_notes` option and attaches
`ai-code-tracker-server-<tag>-linux-amd64.tar` from `/tmp`.

If any preceding step fails, the release step is not run. If the Release action
fails, the workflow fails while the previously published GHCR image and
Actions artifact remain available. Existing tags do not retroactively create a
Release; the behavior applies to tags created after this workflow change.

## Documentation

The README describes the three release outputs: GHCR image tags, the workflow
artifact, and the GitHub Release asset. It directs users to Releases for the
stable archive download and keeps `docker load -i` instructions unchanged.
