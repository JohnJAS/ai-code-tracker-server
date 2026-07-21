package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"ai-code-tracker-server/internal/domain"
	"ai-code-tracker-server/internal/repository"
	"ai-code-tracker-server/internal/store"
)

type handler struct {
	store store.Store
}

func NewHandler(storage store.Store) http.Handler {
	return handler{store: storage}
}

func (h handler) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	switch {
	case request.Method == http.MethodGet && request.URL.Path == "/":
		writeDashboard(writer)
	case request.Method == http.MethodGet && request.URL.Path == "/healthz":
		writeJSON(writer, http.StatusOK, map[string]string{"status": "ok"})
	case request.Method == http.MethodGet && request.URL.Path == "/v1/dashboard":
		h.getDashboard(writer, request)
	case request.Method == http.MethodPost && request.URL.Path == "/v1/records":
		h.postRecords(writer, request)
	default:
		writeJSON(writer, http.StatusNotFound, errorResponse{Error: "not found"})
	}
}

func (h handler) getDashboard(writer http.ResponseWriter, request *http.Request) {
	dashboard, err := h.store.Dashboard(request.Context())
	if err != nil {
		writeJSON(writer, http.StatusInternalServerError, errorResponse{Error: "could not load dashboard"})
		return
	}
	writeJSON(writer, http.StatusOK, dashboard)
}

func (h handler) postRecords(writer http.ResponseWriter, request *http.Request) {
	var batch domain.Batch
	decoder := json.NewDecoder(request.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&batch); err != nil {
		writeJSON(writer, http.StatusBadRequest, errorResponse{Error: "invalid JSON"})
		return
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		writeJSON(writer, http.StatusBadRequest, errorResponse{Error: "request must contain one JSON object"})
		return
	}
	if err := batch.Validate(); err != nil {
		writeJSON(writer, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}
	origin, err := repository.NormalizeOrigin(batch.RepositoryURL)
	if err != nil {
		writeJSON(writer, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return
	}
	result, err := h.store.UpsertBatch(request.Context(), origin, batch.Records)
	if err != nil {
		writeJSON(writer, http.StatusInternalServerError, errorResponse{Error: "could not store records"})
		return
	}
	writeJSON(writer, http.StatusOK, ingestResponse{
		Received:   len(batch.Records),
		Inserted:   result.Inserted,
		Duplicates: result.Duplicates,
	})
}

type ingestResponse struct {
	Received   int `json:"received"`
	Inserted   int `json:"inserted"`
	Duplicates int `json:"duplicates"`
}

type errorResponse struct {
	Error string `json:"error"`
}

func writeJSON(writer http.ResponseWriter, status int, value any) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(status)
	_ = json.NewEncoder(writer).Encode(value)
}
