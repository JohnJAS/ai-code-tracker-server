package httpapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"ai-code-tracker-server/internal/domain"
	"ai-code-tracker-server/internal/store"
)

func TestHealthz(t *testing.T) {
	response := httptest.NewRecorder()
	NewHandler(&fakeStore{}).ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/healthz", nil))

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if response.Body.String() != "{\"status\":\"ok\"}\n" {
		t.Fatalf("body = %q", response.Body.String())
	}
}

func TestDashboardPage(t *testing.T) {
	response := httptest.NewRecorder()
	NewHandler(&fakeStore{}).ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/", nil))

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if contentType := response.Header().Get("Content-Type"); !strings.HasPrefix(contentType, "text/html") {
		t.Fatalf("content type = %q, want HTML", contentType)
	}
	if !strings.Contains(response.Body.String(), "AI Code Tracker") {
		t.Fatalf("body does not contain dashboard title: %q", response.Body.String())
	}
}

func TestDashboardData(t *testing.T) {
	response := httptest.NewRecorder()
	NewHandler(&fakeStore{}).ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/v1/dashboard", nil))

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if response.Body.String() != "{\"total_commits\":0,\"ai_commits\":0,\"ai_lines\":0,\"total_lines\":0,\"repositories\":0,\"recent_records\":[]}\n" {
		t.Fatalf("body = %q", response.Body.String())
	}
}

func TestRecordsDataUsesDefaults(t *testing.T) {
	storage := &fakeStore{records: store.RecordPage{
		Records:  []store.DashboardRecord{},
		Page:     1,
		PageSize: 20,
	}}
	response := httptest.NewRecorder()

	NewHandler(storage).ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/v1/records", nil))

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if storage.query.Page != 1 || storage.query.PageSize != 20 {
		t.Fatalf("query = %#v", storage.query)
	}
}

func TestRecordsDataPassesFiltersToStore(t *testing.T) {
	storage := &fakeStore{records: store.RecordPage{
		Records:  []store.DashboardRecord{},
		Page:     2,
		PageSize: 10,
	}}
	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet,
		"/v1/records?author=dev&repository=acme&start_date=2026-08-01&end_date=2026-08-09&page=2&page_size=10", nil)

	NewHandler(storage).ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	wantStart := time.Date(2026, time.August, 1, 0, 0, 0, 0, time.UTC)
	wantEnd := time.Date(2026, time.August, 10, 0, 0, 0, 0, time.UTC)
	if storage.query.Author != "dev" || storage.query.Repository != "acme" ||
		!storage.query.Start.Equal(wantStart) || !storage.query.End.Equal(wantEnd) ||
		storage.query.Page != 2 || storage.query.PageSize != 10 {
		t.Fatalf("query = %#v", storage.query)
	}
}

func TestRecordsDataRejectsInvalidQuery(t *testing.T) {
	for _, target := range []string{
		"/v1/records?start_date=invalid",
		"/v1/records?start_date=2026-08-09&end_date=2026-08-08",
		"/v1/records?page=0",
		"/v1/records?page_size=101",
	} {
		t.Run(target, func(t *testing.T) {
			response := httptest.NewRecorder()

			NewHandler(&fakeStore{}).ServeHTTP(response, httptest.NewRequest(http.MethodGet, target, nil))

			if response.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
			}
		})
	}
}

func TestPostRecordsReturnsInsertResult(t *testing.T) {
	storage := &fakeStore{result: store.InsertResult{Inserted: 1, Duplicates: 1}}
	request := httptest.NewRequest(http.MethodPost, "/v1/records", strings.NewReader(validPayload()))
	response := httptest.NewRecorder()

	NewHandler(storage).ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusOK)
	}
	if response.Body.String() != "{\"received\":2,\"inserted\":1,\"duplicates\":1}\n" {
		t.Fatalf("body = %q", response.Body.String())
	}
	if storage.origin != "github.com/acme/demo" {
		t.Fatalf("origin = %q", storage.origin)
	}
}

func TestPostRecordsRejectsInvalidJSON(t *testing.T) {
	response := httptest.NewRecorder()
	NewHandler(&fakeStore{}).ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/v1/records", strings.NewReader("{")))

	if response.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", response.Code, http.StatusBadRequest)
	}
}

type fakeStore struct {
	result  store.InsertResult
	origin  string
	query   store.RecordQuery
	records store.RecordPage
}

func (s *fakeStore) UpsertBatch(_ context.Context, origin string, _ []domain.Record) (store.InsertResult, error) {
	s.origin = origin
	return s.result, nil
}

func (s *fakeStore) Dashboard(_ context.Context) (store.Dashboard, error) {
	return store.Dashboard{RecentRecords: []store.DashboardRecord{}}, nil
}

func (s *fakeStore) Records(_ context.Context, query store.RecordQuery) (store.RecordPage, error) {
	s.query = query
	return s.records, nil
}

func validPayload() string {
	return `{
  "repository_url": "git@github.com:acme/demo.git",
  "records": [
    {"author":"dev","ai_lines":1,"total_lines":2,"is_ai_commit":true,"commit_id":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","date":"2026-07-20 10:00:00","message":"feat: demo"},
    {"author":"dev","ai_lines":0,"total_lines":1,"is_ai_commit":false,"commit_id":"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb","date":"2026-07-20 10:01:00","message":"fix: demo"}
  ]
}`
}
