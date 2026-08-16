# Contributor Count Design

## Goal

Show the number of unique commit authors in the dashboard's top statistics.

## Scope

- Count unique non-empty `author` values from commit records.
- Apply the current author, repository, and date filters to the count.
- Add the count to the existing `GET /v1/records` response as `contributors`.
- Add a sixth top-level dashboard statistic labeled "Contributor count".

## Data Flow

`MySQLStore.Records` will calculate `COUNT(DISTINCT c.author)` using the same
filtered record source as the other summary metrics. The API returns that
value with the paginated records. The dashboard uses the returned value when
rendering its statistics.

## Error Handling

The count uses the existing records endpoint and its current load-error state.
No new endpoint or client-side error path is required.

## Testing

- Verify the records API serializes `contributors`.
- Verify the dashboard markup and rendering include the contributor statistic.
- Extend the MySQL integration test with repeated and distinct authors to
  prove the metric is de-duplicated and respects filters.
