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
	}), []string{}, "")

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
	}), []string{}, "")

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
	}), []string{"http://localhost:5173"}, "")

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
	}), []string{"http://localhost:5173"}, "")

	req := httptest.NewRequest(http.MethodOptions, "/api/v1/notes", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", rec.Code)
	}
	if rec.Header().Get("X-Request-Id") == "" {
		t.Fatal("expected X-Request-Id header for preflight request")
	}
}

func TestCORSPreflightReturnsNoContentWhenAuthTokenEnabled(t *testing.T) {
	h := applyMiddlewares(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}), []string{"http://localhost:5173"}, "secret-token")

	req := httptest.NewRequest(http.MethodOptions, "/api/v1/notes", nil)
	req.Header.Set("Origin", "http://localhost:5173")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", rec.Code)
	}
}

func TestAuthTokenDisabledAllowsAPIRequest(t *testing.T) {
	h := applyMiddlewares(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}), []string{}, "")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
}

func TestAuthTokenRejectsMissingBearerToken(t *testing.T) {
	h := applyMiddlewares(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}), []string{}, "secret-token")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/dashboard", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rec.Code)
	}

	var payload struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
		RequestID string `json:"requestId"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("failed to decode json: %v", err)
	}
	if payload.Error.Code != errorCodeUnauthorized {
		t.Fatalf("expected code %s, got %s", errorCodeUnauthorized, payload.Error.Code)
	}
	if payload.RequestID == "" {
		t.Fatal("requestId is empty")
	}
}

func TestAuthTokenRejectsInvalidToken(t *testing.T) {
	h := applyMiddlewares(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}), []string{}, "secret-token")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/notes", nil)
	req.Header.Set("Authorization", "Bearer wrong-token")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rec.Code)
	}
}

func TestAuthTokenAllowsValidBearerToken(t *testing.T) {
	h := applyMiddlewares(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}), []string{}, "secret-token")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/notes", nil)
	req.Header.Set("Authorization", "Bearer secret-token")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
}

func TestAuthTokenSkipsNonAPIPath(t *testing.T) {
	h := applyMiddlewares(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}), []string{}, "secret-token")

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
}

func TestAdminEndpointRequiresAuthTokenConfig(t *testing.T) {
	h := applyMiddlewares(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}), []string{}, "")

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/backup", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected status 403, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"FORBIDDEN"`) {
		t.Fatalf("unexpected response body: %s", rec.Body.String())
	}
}

func TestAdminEndpointRequiresBearerTokenWhenConfigured(t *testing.T) {
	h := applyMiddlewares(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}), []string{}, "secret-token")

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/backup", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", rec.Code)
	}
}

func TestAdminEndpointAllowsValidBearerTokenWhenConfigured(t *testing.T) {
	h := applyMiddlewares(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	}), []string{}, "secret-token")

	req := httptest.NewRequest(http.MethodPost, "/api/v1/admin/backup", nil)
	req.Header.Set("Authorization", "Bearer secret-token")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
}
