package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestErrorFormatIncludesRequestIDAndTokyoTimestamp(t *testing.T) {
	h := applyMiddlewares(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeError(w, r, http.StatusBadRequest, errorCodeValidation, "invalid request")
	}), []string{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Fatalf("unexpected content type: %s", ct)
	}
	if cc := rec.Header().Get("Cache-Control"); cc != "no-store" {
		t.Fatalf("unexpected cache-control: %s", cc)
	}

	var payload struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
		RequestID string `json:"requestId"`
		Timestamp string `json:"timestamp"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode json: %v", err)
	}

	if payload.Error.Code != errorCodeValidation {
		t.Fatalf("unexpected error code: %s", payload.Error.Code)
	}
	if payload.RequestID == "" {
		t.Fatal("requestId is empty")
	}
	if payload.RequestID != rec.Header().Get("X-Request-Id") {
		t.Fatalf("requestId mismatch: body=%s header=%s", payload.RequestID, rec.Header().Get("X-Request-Id"))
	}
	if !strings.HasSuffix(payload.Timestamp, "+09:00") {
		t.Fatalf("timestamp is not JST: %s", payload.Timestamp)
	}
	if _, err := time.Parse(time.RFC3339, payload.Timestamp); err != nil {
		t.Fatalf("timestamp is not RFC3339: %v", err)
	}
}

func TestCORSDisabledByDefault(t *testing.T) {
	h := applyMiddlewares(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}), []string{})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Fatalf("cors should be disabled, got origin=%s", rec.Header().Get("Access-Control-Allow-Origin"))
	}
}

func TestCORSEnabledForConfiguredOrigin(t *testing.T) {
	h := applyMiddlewares(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}), []string{"http://localhost:5173"})

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Header().Get("Access-Control-Allow-Origin") != "http://localhost:5173" {
		t.Fatalf("unexpected allow-origin: %s", rec.Header().Get("Access-Control-Allow-Origin"))
	}
	if methods := rec.Header().Get("Access-Control-Allow-Methods"); !strings.Contains(methods, "PATCH") || !strings.Contains(methods, "DELETE") {
		t.Fatalf("unexpected allow-methods: %s", methods)
	}
}

func TestCORSPreflightReturnsNoContent(t *testing.T) {
	h := applyMiddlewares(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}), []string{"http://localhost:5173"})

	req := httptest.NewRequest(http.MethodOptions, "/api/v1/notes", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", rec.Code)
	}
}
