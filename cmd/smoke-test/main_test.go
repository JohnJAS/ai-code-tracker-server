package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"
)

type receivedBatch struct {
	RepositoryURL string `json:"repository_url"`
	Records       []struct {
		Author     string `json:"author"`
		AiLines    int    `json:"ai_lines"`
		TotalLines int    `json:"total_lines"`
		IsAICommit bool   `json:"is_ai_commit"`
		CommitID   string `json:"commit_id"`
		Date       string `json:"date"`
		Message    string `json:"message"`
	} `json:"records"`
}

func TestSmokeCommandUploadsAndConfirmsDashboard(t *testing.T) {
	var mutex sync.Mutex
	var batch receivedBatch
	requests := make(map[string]int)
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		mutex.Lock()
		defer mutex.Unlock()
		requests[request.Method+" "+request.URL.Path]++

		switch request.Method + " " + request.URL.Path {
		case http.MethodPost + " /v1/records":
			if err := json.NewDecoder(request.Body).Decode(&batch); err != nil {
				t.Fatalf("decode ingest request: %v", err)
			}
			_ = json.NewEncoder(writer).Encode(map[string]int{"received": 1, "inserted": 1, "duplicates": 0})
		case http.MethodGet + " /v1/dashboard":
			if len(batch.Records) != 1 {
				t.Fatalf("records = %d, want 1", len(batch.Records))
			}
			record := batch.Records[0]
			_ = json.NewEncoder(writer).Encode(map[string]any{
				"recent_records": []map[string]any{{
					"repository_url": batch.RepositoryURL,
					"commit_id":      record.CommitID,
					"author":         record.Author,
					"ai_lines":       record.AiLines,
					"total_lines":    record.TotalLines,
					"is_ai_commit":   record.IsAICommit,
					"date":           record.Date,
					"message":        record.Message,
				}},
			})
		default:
			http.NotFound(writer, request)
		}
	}))
	defer server.Close()

	repoRoot, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatalf("resolve repository root: %v", err)
	}
	command := exec.Command("go", "run", "./cmd/smoke-test", "-url", server.URL+"/v1/records")
	command.Dir = repoRoot
	command.Env = os.Environ()
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("smoke command failed: %v\n%s", err, output)
	}
	if requests["POST /v1/records"] != 1 || requests["GET /v1/dashboard"] != 1 {
		t.Fatalf("requests = %#v", requests)
	}
	if len(batch.Records) != 1 || batch.Records[0].CommitID == "" {
		t.Fatalf("batch = %#v", batch)
	}
}
