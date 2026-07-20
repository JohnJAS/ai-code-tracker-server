package domain

import (
	"strings"
	"testing"
)

func TestBatchValidateRejectsInvalidRecord(t *testing.T) {
	batch := Batch{
		RepositoryURL: "https://github.com/acme/demo.git",
		Records: []Record{{
			Author:     "dev",
			CommitID:   "bad",
			AiLines:    2,
			TotalLines: 1,
			Date:       "2026-07-20 10:00:00",
			Message:    "feat: demo",
		}},
	}

	if err := batch.Validate(); err == nil {
		t.Fatal("Validate() error = nil, want error")
	}
}

func TestBatchValidateAcceptsCSVRecord(t *testing.T) {
	batch := Batch{
		RepositoryURL: "git@github.com:acme/demo.git",
		Records: []Record{{
			Author:     "dev",
			CommitID:   strings.Repeat("a", 40),
			AiLines:    1,
			TotalLines: 3,
			Date:       "2026-07-20 10:00:00",
			Message:    "feat: demo",
		}},
	}

	if err := batch.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
}

func TestRecordValidateRejectsInvalidDate(t *testing.T) {
	record := Record{
		Author:     "dev",
		CommitID:   strings.Repeat("b", 40),
		AiLines:    1,
		TotalLines: 1,
		Date:       "tomorrow",
		Message:    "feat: demo",
	}

	if err := record.Validate(); err == nil {
		t.Fatal("Validate() error = nil, want error")
	}
}
