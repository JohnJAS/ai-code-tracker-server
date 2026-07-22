# Versioned Docker Build Design

## Goal

Make local development builds and released builds unambiguous: local source is
published as `latest`, while a requested Git tag produces a Docker image with
the same tag. Docker Compose must run an explicit image tag without building
source code itself. MySQL data must be stored in a configurable host directory.

## Build Modes

`scripts/build.sh` has two modes:

- Without an argument, it builds the current local worktree as
  `ai-code-tracker-server:latest`. This mode is intended for local development
  and may include uncommitted changes.
- With a version argument such as `v1.2.0`, it verifies that the Git tag exists,
  creates a temporary detached worktree at that tag, and builds its contents as
  `ai-code-tracker-server:v1.2.0`. The current worktree and its uncommitted
  changes are not used.

The repository/image name is configurable through `IMAGE_NAME`, defaulting to
`ai-code-tracker-server`. The script removes the temporary worktree on exit,
including when Docker build fails.

## Compose Runtime

The server service declares only an `image` using
`${IMAGE_NAME:-ai-code-tracker-server}:${APP_VERSION:-latest}`. It has no
`build` section, so `docker compose up` never rebuilds local source. The
default runtime image is `latest`; setting `APP_VERSION=v1.2.0` starts the
corresponding versioned image.

MySQL uses a bind mount:

`$${MYSQL_DATA_DIR:-./data/mysql}:/var/lib/mysql`

The default persists database files below the repository. Deployments may set
`MYSQL_DATA_DIR` to an absolute host path. Data survives `docker compose down`;
it is deleted only when the host directory is explicitly removed.

## Documentation and Validation

The README documents both build modes, version-selected startup, and the MySQL
data directory lifecycle. Automated tests cover the build-script argument and
tag-validation behavior without invoking Docker. Compose configuration is
validated with `docker compose config`, using required database environment
variables.
