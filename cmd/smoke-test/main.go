package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const defaultUploadURL = "http://127.0.0.1:8080/v1/records"

type batch struct {
	RepositoryURL string   `json:"repository_url"`
	Records       []record `json:"records"`
}

type record struct {
	Author     string `json:"author"`
	AiLines    int    `json:"ai_lines"`
	TotalLines int    `json:"total_lines"`
	IsAICommit bool   `json:"is_ai_commit"`
	CommitID   string `json:"commit_id"`
	Date       string `json:"date"`
	Message    string `json:"message"`
}

type dashboard struct {
	RecentRecords []record `json:"recent_records"`
}

func main() {
	uploadURL := flag.String("url", valueOrDefault(os.Getenv("AI_TRACKER_UPLOAD_URL"), defaultUploadURL), "record ingestion endpoint")
	flag.Parse()
	if err := run(context.Background(), *uploadURL, http.DefaultClient); err != nil {
		fmt.Fprintf(os.Stderr, "[ai-code-tracker] server smoke test failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("[ai-code-tracker] server smoke test passed")
}

func run(ctx context.Context, uploadURL string, client *http.Client) error {
	endpoint, err := url.Parse(uploadURL)
	if err != nil || endpoint.Scheme == "" || endpoint.Host == "" {
		return fmt.Errorf("invalid upload URL: %q", uploadURL)
	}
	commitID, err := randomSHA()
	if err != nil {
		return fmt.Errorf("generate commit ID: %w", err)
	}
	smokeRecord := record{
		Author:     "AI Tracker Server Smoke Test",
		AiLines:    1,
		TotalLines: 1,
		IsAICommit: true,
		CommitID:   commitID,
		Date:       time.Now().Format("2006-01-02 15:04:05"),
		Message:    "test: server smoke record",
	}
	payload := batch{
		RepositoryURL: "https://github.com/ai-code-tracker/server-smoke-" + commitID[:12] + ".git",
		Records:       []record{smokeRecord},
	}
	encoded, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("encode upload: %w", err)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint.String(), bytes.NewReader(encoded))
	if err != nil {
		return fmt.Errorf("create upload request: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("upload record: %w", err)
	}
	if err := requireSuccess(response); err != nil {
		return fmt.Errorf("upload record: %w", err)
	}

	endpoint.Path = "/v1/dashboard"
	endpoint.RawQuery = ""
	dashboardRequest, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint.String(), nil)
	if err != nil {
		return fmt.Errorf("create dashboard request: %w", err)
	}
	dashboardResponse, err := client.Do(dashboardRequest)
	if err != nil {
		return fmt.Errorf("load dashboard: %w", err)
	}
	if dashboardResponse.StatusCode < http.StatusOK || dashboardResponse.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("load dashboard: %s", dashboardResponse.Status)
	}
	defer dashboardResponse.Body.Close()
	var result dashboard
	if err := json.NewDecoder(dashboardResponse.Body).Decode(&result); err != nil {
		return fmt.Errorf("decode dashboard: %w", err)
	}
	if !containsCommit(result.RecentRecords, commitID) {
		return fmt.Errorf("dashboard did not include uploaded commit %s", commitID)
	}
	return nil
}

func randomSHA() (string, error) {
	bytes := make([]byte, 20)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func containsCommit(records []record, commitID string) bool {
	for _, record := range records {
		if record.CommitID == commitID {
			return true
		}
	}
	return false
}

func requireSuccess(response *http.Response) error {
	defer response.Body.Close()
	if response.StatusCode >= http.StatusOK && response.StatusCode < http.StatusMultipleChoices {
		return nil
	}
	body, _ := io.ReadAll(io.LimitReader(response.Body, 4*1024))
	return fmt.Errorf("%s: %s", response.Status, strings.TrimSpace(string(body)))
}

func valueOrDefault(value, fallback string) string {
	if value != "" {
		return value
	}
	return fallback
}
