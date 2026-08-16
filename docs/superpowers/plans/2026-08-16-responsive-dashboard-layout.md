# Responsive Dashboard Layout Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make the dashboard fill available desktop space and allocate more table width to repository and commit-message content.

**Architecture:** Keep the dashboard as its existing inline HTML document in `internal/httpapi/dashboard.go`. Change only CSS layout rules and add a `colgroup` to define stable compact widths for metadata columns while leaving the repository and message columns fluid. Lock the emitted layout contract with the existing handler test.

**Tech Stack:** Go standard library, `net/http/httptest`, inline HTML/CSS.

---

### Task 1: Lock the responsive layout contract

**Files:**
- Modify: `internal/httpapi/handler_test.go:25-58`
- Test: `internal/httpapi/handler_test.go:25-58`

- [ ] **Step 1: Add a failing dashboard markup test**

Add these strings to the `TestDashboardPage` required-content list:

```go
`main { margin: 0; padding: 32px clamp(16px, 2vw, 40px) 64px; }`,
`table { width: 100%; border-collapse: collapse; min-width: 920px; table-layout: fixed; background: #fff; }`,
`.date-column { width: 140px; }`,
`.repository-column { width: 22%; }`,
`.message-column { width: 30%; }`,
`<colgroup><col class="date-column"><col class="repository-column"><col class="author-column"><col class="tool-column"><col class="ai-lines-column"><col class="total-lines-column"><col class="commit-column"><col class="message-column"></colgroup>`,
```

- [ ] **Step 2: Run the focused test and verify it fails**

Run:

```powershell
go test ./internal/httpapi -run TestDashboardPage -count=1
```

Expected: `FAIL` because the dashboard does not yet contain the new fluid `main` rule, column classes, or `colgroup`.

- [ ] **Step 3: Implement the smallest dashboard markup and CSS change**

In `internal/httpapi/dashboard.go`, replace the current fixed-width `main` and table rules with:

```css
main { margin: 0; padding: 32px clamp(16px, 2vw, 40px) 64px; }
table { width: 100%; border-collapse: collapse; min-width: 920px; table-layout: fixed; background: #fff; }
.date-column { width: 140px; }
.repository-column { width: 22%; }
.author-column { width: 140px; }
.tool-column { width: 120px; }
.ai-lines-column, .total-lines-column { width: 84px; }
.commit-column { width: 130px; }
.message-column { width: 30%; }
```

Insert this `colgroup` immediately after the opening `<table>` tag:

```html
<colgroup><col class="date-column"><col class="repository-column"><col class="author-column"><col class="tool-column"><col class="ai-lines-column"><col class="total-lines-column"><col class="commit-column"><col class="message-column"></colgroup>
```

Keep `.table-wrap { overflow-x: auto; ... }` and the existing narrow-screen media queries unchanged.

- [ ] **Step 4: Run the focused test and verify it passes**

Run:

```powershell
go test ./internal/httpapi -run TestDashboardPage -count=1
```

Expected: `ok   ai-code-tracker-server/internal/httpapi`.

- [ ] **Step 5: Commit the focused implementation**

```powershell
git add -- internal/httpapi/dashboard.go internal/httpapi/handler_test.go
git commit -m "fix: make dashboard layout responsive"
```

### Task 2: Verify the complete service test suite

**Files:**
- Verify: `internal/httpapi/dashboard.go`
- Verify: `internal/httpapi/handler_test.go`

- [ ] **Step 1: Format the modified Go files**

Run:

```powershell
gofmt -w internal/httpapi/dashboard.go internal/httpapi/handler_test.go
```

- [ ] **Step 2: Run the complete test suite**

Run:

```powershell
go test ./...
```

Expected: all Go packages pass, with any database integration tests following their existing environment-driven skip behavior.

- [ ] **Step 3: Verify the final diff is scoped**

Run:

```powershell
git diff HEAD~1 --check
git status --short
```

Expected: no whitespace errors; only the committed dashboard layout and test changes are present, apart from untracked visual-companion artifacts under `.superpowers/`.
