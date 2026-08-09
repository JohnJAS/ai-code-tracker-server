# Dashboard Query and Pagination Design

## Goal

Allow dashboard users to filter commit records by author keyword, repository
keyword, and committed date range. Display both the matching records in pages
and the five dashboard statistics calculated from the complete filtered result
set.

## Scope

- Add a paginated `GET /v1/records` endpoint.
- Support optional author and repository keyword filters.
- Support optional inclusive start and end dates.
- Return filtered dashboard statistics together with the selected result page.
- Add dashboard controls for filters, query submission, and previous/next
  pagination.
- Add focused HTTP unit tests and MySQL integration tests.

The change does not add filtering by commit message, AI commit status, sorting
controls, authentication, or record modification.

## API

`GET /v1/records` accepts these query parameters:

| Parameter | Required | Meaning |
| --- | --- | --- |
| `author` | No | Literal substring matched against the commit author. |
| `repository` | No | Literal substring matched against the normalized repository origin. |
| `start_date` | No | Inclusive lower date boundary in `YYYY-MM-DD` format. |
| `end_date` | No | Inclusive upper date boundary in `YYYY-MM-DD` format. |
| `page` | No | One-based page number; defaults to `1`. |
| `page_size` | No | Records per page; defaults to `20`. |

The server rejects malformed dates, an end date before a start date, and page
or page-size values less than one with `400 Bad Request`. Page size is capped
at `100` to prevent unbounded result requests. An empty keyword has the same
effect as omitting that filter.

Keywords are used as literal substring searches: `%`, `_`, and the SQL escape
character are escaped before they are used with parameterized `LIKE`
conditions. Filters are combined with logical AND.

Dates filter the `committed_at` timestamp. `start_date` includes records at or
after the start day. `end_date` includes the complete end day by applying an
exclusive boundary at the following day.

The response contains the existing five statistic fields, the requested
records, and pagination metadata:

```json
{
  "total_commits": 148,
  "ai_commits": 80,
  "ai_lines": 1240,
  "total_lines": 2160,
  "repositories": 12,
  "records": [],
  "total_records": 148,
  "page": 1,
  "page_size": 20,
  "total_pages": 8
}
```

`total_commits` and `total_records` both identify the number of matching
commit records. They are both returned so the dashboard statistics retain their
current field name while the pagination contract remains explicit.

Statistics always aggregate every record matching the active filters, not only
the records in the current page. `repositories` counts distinct repositories
within that same filtered result set. A request beyond the last page returns an
empty `records` array and intact statistics and pagination metadata. A query
with no matches returns `page: 1` and `total_pages: 0`.

The existing `GET /v1/dashboard` endpoint remains available with its current
unfiltered, recent-records response for compatibility. The dashboard web page
uses `GET /v1/records` for initial loading, filtering, and pagination.

## Storage

The store layer gains a query input type and a paginated result type so the
HTTP layer does not construct SQL. The MySQL implementation uses the same
filter predicate for:

1. aggregate statistics and the filtered distinct repository count;
2. matching record count;
3. ordered page retrieval.

Records are ordered by `committed_at DESC, commit_id DESC`, preserving the
current stable ordering. The page query uses parameterized `LIMIT` and
`OFFSET`. Existing indexes remain available for date-constrained queries.
Leading-wildcard keyword searches are not expected to use an index efficiently;
no migration is required for this initial feature.

## Dashboard Interaction

The records section includes:

- an author keyword input;
- a repository keyword input;
- start and end date inputs;
- a query button;
- a result count and current-page indicator;
- previous and next page buttons.

On initial load, the page requests the first 20 unfiltered records and their
statistics. Submitting filters resets the requested page to `1`. Previous and
next requests preserve all active filters. Buttons are disabled when there is
no preceding or following page. Empty matches render an explicit empty-state
table row and zero-valued statistics.

While a request is in progress, the controls prevent duplicate submissions.
If a request fails, the page reports that the data could not be loaded and
leaves the last successful statistics and table content visible.

## Testing

HTTP unit tests cover:

- default pagination;
- parsing valid filters and passing them to the store;
- invalid dates, inverted date ranges, and invalid pagination values;
- JSON response shape and pagination metadata.

MySQL integration tests seed records spanning multiple authors, repositories,
and days, then verify:

- author and repository literal substring matching;
- combined filters use intersection semantics;
- both date boundaries are inclusive at day granularity;
- filtered statistics aggregate across all matches rather than one page;
- distinct filtered repository count;
- stable ordering, total count, and empty out-of-range pages.

The existing dashboard and ingestion tests remain in place.
