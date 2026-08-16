package domain

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
)

const commitDateLayout = "2006-01-02 15:04:05"

var commitIDPattern = regexp.MustCompile(`^[0-9a-f]{40}([0-9a-f]{24})?$`)

type Batch struct {
	RepositoryURL string   `json:"repository_url"`
	Records       []Record `json:"records"`
}

type Record struct {
	Author     string `json:"author"`
	AiLines    int    `json:"ai_lines"`
	TotalLines int    `json:"total_lines"`
	IsAICommit bool   `json:"is_ai_commit"`
	AiTool     string `json:"ai_tool"`
	CommitID   string `json:"commit_id"`
	Date       string `json:"date"`
	Message    string `json:"message"`
}

func (b Batch) Validate() error {
	if strings.TrimSpace(b.RepositoryURL) == "" {
		return errors.New("repository_url is required")
	}
	if len(b.Records) == 0 {
		return errors.New("records is required")
	}
	for index, record := range b.Records {
		if err := record.Validate(); err != nil {
			return fmt.Errorf("records[%d]: %w", index, err)
		}
	}
	return nil
}

func (r Record) Validate() error {
	if strings.TrimSpace(r.Author) == "" {
		return errors.New("author is required")
	}
	if !commitIDPattern.MatchString(r.CommitID) {
		return errors.New("commit_id must be a 40- or 64-character lowercase hexadecimal SHA")
	}
	if r.AiLines < 0 || r.TotalLines < 0 {
		return errors.New("line counts must not be negative")
	}
	if r.AiLines > r.TotalLines {
		return errors.New("ai_lines must not exceed total_lines")
	}
	if _, err := time.Parse(commitDateLayout, r.Date); err != nil {
		return fmt.Errorf("date must use %s", commitDateLayout)
	}
	if strings.TrimSpace(r.Message) == "" {
		return errors.New("message is required")
	}
	return nil
}
