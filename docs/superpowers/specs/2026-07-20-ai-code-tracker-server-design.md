# AI Code Tracker Server Design

## Goal

Create a standalone Go service that centrally stores AI code-tracking records
produced by `git-code-tracker`. Developers' repositories upload only the
records for commits successfully pushed to a Git remote. This removes the need
to manually collect each repository's local CSV files.

## Scope

- A Go HTTP service backed by MySQL 8.
- A JSON ingestion API and a health endpoint.
- Idempotent storage keyed by normalized repository URL and commit ID.
- A `git-code-tracker` client extension that uploads after successful pushes,
  with durable local retry state.
- Docker deployment files, database migrations, documentation, and automated
  tests.

The first release does not provide authentication, a web UI, data deletion for
rewritten history, or centralized configuration management. It is intended for
deployment on a trusted internal network.

## Existing Data Contract

Each local CSV row represents one Git commit and contains:

| Field | Meaning |
| --- | --- |
| `author` | Git commit author name |
| `ai_lines` | Newly added lines attributed to AI edits |
| `total_lines` | Total newly added lines in the commit |
| `is_ai_commit` | Whether an AI tool initiated the commit command |
| `commit_id` | Full Git commit SHA |
| `date` | Commit time |
| `message` | Commit subject |

`ai_lines` and `is_ai_commit` are independent signals. The service preserves
both values.

## Architecture

```text
Developer repository
  post-push hook
    -> find pushed commits and matching local CSV rows
    -> retry local outbox entries
    -> POST /v1/records

Go service
  -> validate request
  -> normalize repository URL
  -> transactionally upsert MySQL records
```

The Go service uses the standard library HTTP server and `database/sql` with a
MySQL driver. SQL migrations are committed with the project and applied at
startup before accepting requests. The service is configured through
environment variables:

- `LISTEN_ADDR`, default `:8080`
- `MYSQL_DSN`, required

## API

`POST /v1/records` accepts:

```json
{
  "repository_url": "https://github.com/acme/example.git",
  "records": [
    {
      "author": "yooocen",
      "ai_lines": 70,
      "total_lines": 173,
      "is_ai_commit": true,
      "commit_id": "b6d9e478824ad43c4f67d54b9ded5d74317e3651",
      "date": "2026-05-14 09:21:30",
      "message": "feat: add slash commands"
    }
  ]
}
```

The request must contain a valid Git remote URL, at least one record, a
hexadecimal commit ID, non-negative line counts, and `ai_lines <= total_lines`.
The successful response reports the number of received records. Duplicate
records are successful requests, not errors.

`GET /healthz` returns a successful status only after the process is running;
it does not expose database credentials or tracking data.

## Database Model

`repositories` stores an internal ID, a unique normalized `origin` URL, and
creation time. `commit_records` stores the CSV fields, its repository foreign
key, receipt time, and a unique `(repository_id, commit_id)` index. Additional
indexes on repository/date and author/date support future reporting without
changing the ingestion contract.

Repository URLs are normalized before lookup by removing credentials, a
trailing slash, and a final `.git`, lowercasing the host, and storing the
canonical `host/path` value. Thus `git@host:team/repo.git` and
`https://host/team/repo.git` identify the same repository. The first release
records commits as append-only. A force-push does not delete a previously
received record.

## Client Integration

The existing tracker receives a new `uploadUrl` field in
`.ai-tracking/config.json`. The default is an empty string, meaning upload is
disabled until a deployment address such as
`http://tracker.internal:8080/v1/records` is configured.

The installed `post-push` hook executes after a successful push. For every
non-deletion ref from hook stdin, it compares the old remote SHA to the new
local SHA, identifies newly pushed commits, reads matching rows from all local
tracker CSV files, and sends one batch to `uploadUrl`.

If an upload fails, the records are persisted in
`.ai-tracking/upload-outbox.json`. A later successful `post-push` first retries
the outbox and then sends newly found records. Failure to upload never changes
the push result. The server's unique key makes retrying harmless.

If `uploadUrl` is empty, the client performs no network work and preserves all
current tracker behavior.

## Error Handling

- Invalid HTTP input receives a 400 response with a concise field error.
- MySQL failures receive a 500 response and are logged server-side without
  leaking DSNs or SQL details to the caller.
- The client treats non-2xx responses, timeouts, and malformed responses as a
  retryable upload failure.
- Local outbox failures are logged by the existing tracker error mechanism and
  do not block Git push.

## Testing

The Go project follows test-driven development. Unit tests cover request
validation, URL normalization, and HTTP response mapping. MySQL-backed
integration tests cover migrations, insertion, idempotency, and mixed batches
containing existing and new records. Docker Compose provides a reproducible
MySQL 8 environment for local verification.

The tracker change adds focused tests for pushed-ref commit selection, CSV
record selection, disabled configuration, outbox persistence, retry ordering,
and idempotent retry behavior.

## Deployment

The project includes a multi-stage Dockerfile and `compose.yaml` containing the
service and MySQL 8. Operators expose the service only on a trusted network in
this unauthenticated release and configure each participating repository's
`uploadUrl` to point at its `/v1/records` endpoint.
