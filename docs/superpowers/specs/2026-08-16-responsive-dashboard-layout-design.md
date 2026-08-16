# Responsive Dashboard Layout Design

## Goal

Make the dashboard use the available desktop window width so repository URLs and commit messages remain readable, while preserving usable behavior on narrow screens.

## Scope

- Remove the desktop maximum width from the page content area.
- Keep horizontal page padding so content does not touch the viewport edge.
- Assign table column widths by content priority:
  - Keep numeric metrics compact.
  - Give repository and commit-message columns the largest share of available width.
  - Keep date, author, AI tool, and abbreviated commit ID at stable readable widths.
- Keep the existing horizontal scroll behavior below the table's minimum readable width.
- Keep the existing filter, statistics, pagination, API requests, and record rendering behavior unchanged.

## Implementation

The dashboard stylesheet in `internal/httpapi/dashboard.go` will use a fluid `main` layout instead of the `1180px` cap. The table will use an explicit column layout so extra desktop space consistently benefits the repository and message columns rather than being distributed arbitrarily.

Existing responsive breakpoints will continue to stack filters and statistics on smaller screens. The table wrapper will retain `overflow-x: auto` and a minimum table width so narrow viewports can scroll instead of truncating every column.

## Testing

Extend `TestDashboardPage` to assert the emitted HTML contains the responsive layout rules and column-sizing hooks. Run the focused HTTP API test package, then the full Go test suite.

## Non-goals

- Changing record data, API shapes, filtering, pagination, or text copy.
- Adding client-side column resizing or persistence.
- Redesigning the dashboard visual language.
