package http

import (
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type SPAHandler struct {
	distDir string
}

func NewSPAHandler(distDir string) *SPAHandler {
	return &SPAHandler{distDir: distDir}
}

func (h *SPAHandler) Serve(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, r, http.StatusNotFound, errorCodeNotFound, "not found")
		return
	}

	if strings.HasPrefix(r.URL.Path, "/api/") || r.URL.Path == "/api" {
		writeError(w, r, http.StatusNotFound, errorCodeNotFound, "not found")
		return
	}

	requestedPath := filepath.Clean(strings.TrimPrefix(r.URL.Path, "/"))
	if requestedPath == "." {
		requestedPath = ""
	}
	if strings.HasPrefix(requestedPath, "..") || filepath.IsAbs(requestedPath) {
		writeError(w, r, http.StatusNotFound, errorCodeNotFound, "not found")
		return
	}
	if requestedPath != "" {
		candidate := filepath.Join(h.distDir, requestedPath)
		if fi, err := os.Stat(candidate); err == nil && !fi.IsDir() {
			http.ServeFile(w, r, candidate)
			return
		}
		if isStaticAssetRequest(requestedPath) {
			writeError(w, r, http.StatusNotFound, errorCodeNotFound, "not found")
			return
		}
	}

	indexPath := filepath.Join(h.distDir, "index.html")
	if _, err := os.Stat(indexPath); err != nil {
		writeInternalError(w, r, errorCodeInternal, err)
		return
	}
	http.ServeFile(w, r, indexPath)
}

func isStaticAssetRequest(requestedPath string) bool {
	return path.Ext(requestedPath) != ""
}
