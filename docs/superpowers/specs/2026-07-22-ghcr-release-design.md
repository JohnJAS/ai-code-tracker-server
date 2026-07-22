# GHCR Release Design

## Goal

Publish deployable image artifacts whenever a version Git tag is pushed. Each
release produces a Linux AMD64 image in GitHub Container Registry and a
downloadable Docker archive in the GitHub Actions run.

## Trigger and Image Names

The release workflow runs only for pushed tags matching `v*`. For a tag such as
`v1.2.0`, the image name is:

`ghcr.io/JohnJAS/ai-code-tracker-server`

The workflow publishes both `:v1.2.0` and `:latest`. The local development
image name remains `ai-code-tracker-server:latest`; it is unrelated to the
GHCR namespace.

## Workflow

The GitHub Actions workflow reads repository contents and writes GitHub
Packages. It checks out the tagged revision, installs the module's Go version,
and runs `go test ./...` before creating artifacts. It authenticates to GHCR
with the workflow `GITHUB_TOKEN`.

Buildx creates one Linux AMD64 image build with two output exporters:

- Registry output pushes the two GHCR tags.
- Docker output writes `ai-code-tracker-server-<tag>-linux-amd64.tar`.

`actions/upload-artifact` uploads the tar archive for the workflow retention
period. A failed test, registry login, image build, push, or artifact upload
fails the release run without publishing a success result.

## Usage

Users release with `git tag v1.2.0` followed by `git push origin v1.2.0`.
Deployments either pull the GHCR image through Compose with
`IMAGE_NAME=ghcr.io/JohnJAS/ai-code-tracker-server` and `APP_VERSION=v1.2.0`,
or download the tar artifact, run `docker load -i`, and then use the local
image name in Compose.

The README documents the tag pattern, public/private package access
implications, registry deployment command, and archive-loading command.
