package http

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSPAHandler_ServesIndexAsFallback(t *testing.T) {
	dir := t.TempDir()
	indexPath := filepath.Join(dir, "index.html")
	if err := os.WriteFile(indexPath, []byte("<html>spa</html>"), 0o644); err != nil {
		t.Fatalf("write index failed: %v", err)
	}

	h := NewSPAHandler(dir)
	req := httptest.NewRequest(http.MethodGet, "/dashboard/child", nil)
	rec := httptest.NewRecorder()

	h.Serve(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "spa") {
		t.Fatalf("unexpected body: %s", rec.Body.String())
	}
}

func TestSPAHandler_ServesStaticFile(t *testing.T) {
	dir := t.TempDir()
	indexPath := filepath.Join(dir, "index.html")
	assetPath := filepath.Join(dir, "assets", "app.js")
	if err := os.MkdirAll(filepath.Dir(assetPath), 0o755); err != nil {
		t.Fatalf("mkdir failed: %v", err)
	}
	if err := os.WriteFile(indexPath, []byte("<html>spa</html>"), 0o644); err != nil {
		t.Fatalf("write index failed: %v", err)
	}
	if err := os.WriteFile(assetPath, []byte("console.log('ok')"), 0o644); err != nil {
		t.Fatalf("write asset failed: %v", err)
	}

	h := NewSPAHandler(dir)
	req := httptest.NewRequest(http.MethodGet, "/assets/app.js", nil)
	rec := httptest.NewRecorder()

	h.Serve(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "console.log") {
		t.Fatalf("unexpected body: %s", rec.Body.String())
	}
}

func TestSPAHandler_ServesWebManifestWithManifestContentType(t *testing.T) {
	dir := t.TempDir()
	manifestPath := filepath.Join(dir, "manifest.webmanifest")
	if err := os.WriteFile(manifestPath, []byte(`{"name":"HomeDash"}`), 0o644); err != nil {
		t.Fatalf("write manifest failed: %v", err)
	}

	h := NewSPAHandler(dir)
	req := httptest.NewRequest(http.MethodGet, "/manifest.webmanifest", nil)
	rec := httptest.NewRecorder()

	h.Serve(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
	if got := rec.Header().Get("Content-Type"); got != "application/manifest+json" {
		t.Fatalf("expected content-type application/manifest+json, got %q", got)
	}
}

func TestSPAHandler_DoesNotHandleAPIPath(t *testing.T) {
	h := NewSPAHandler(t.TempDir())
	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	rec := httptest.NewRecorder()

	h.Serve(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"NOT_FOUND"`) {
		t.Fatalf("unexpected error body: %s", rec.Body.String())
	}
}

func TestSPAHandler_DoesNotHandleOtherAPIPrefixPath(t *testing.T) {
	h := NewSPAHandler(t.TempDir())
	req := httptest.NewRequest(http.MethodGet, "/api/v2/health", nil)
	rec := httptest.NewRecorder()

	h.Serve(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"NOT_FOUND"`) {
		t.Fatalf("unexpected error body: %s", rec.Body.String())
	}
}

func TestSPAHandler_Returns404ForMissingStaticAsset(t *testing.T) {
	dir := t.TempDir()
	indexPath := filepath.Join(dir, "index.html")
	if err := os.WriteFile(indexPath, []byte("<html>spa</html>"), 0o644); err != nil {
		t.Fatalf("write index failed: %v", err)
	}

	h := NewSPAHandler(dir)
	req := httptest.NewRequest(http.MethodGet, "/assets/missing.js", nil)
	rec := httptest.NewRecorder()

	h.Serve(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"NOT_FOUND"`) {
		t.Fatalf("unexpected error body: %s", rec.Body.String())
	}
}
