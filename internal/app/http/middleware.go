package http

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log"
	"net/http"
	"strings"
	"time"
)

type ctxKey string

const requestIDKey ctxKey = "request_id"

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func applyMiddlewares(next http.Handler, corsAllowOrigins []string, authToken string) http.Handler {
	h := authTokenMiddleware(next, authToken)
	h = corsMiddleware(h, corsAllowOrigins)
	h = requestIDAndAccessLogMiddleware(h)
	return h
}

func authTokenMiddleware(next http.Handler, authToken string) http.Handler {
	if authToken == "" {
		return next
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !isAPIv1Path(r.URL.Path) || r.Method == http.MethodOptions {
			next.ServeHTTP(w, r)
			return
		}

		if !isValidBearerToken(r.Header.Get("Authorization"), authToken) {
			writeError(w, r, http.StatusUnauthorized, errorCodeUnauthorized, "authentication required")
			return
		}

		next.ServeHTTP(w, r)
	})
}

func isAPIv1Path(path string) bool {
	return path == "/api/v1" || strings.HasPrefix(path, "/api/v1/")
}

func isValidBearerToken(authHeader string, expectedToken string) bool {
	parts := strings.Fields(authHeader)
	if len(parts) != 2 {
		return false
	}
	if !strings.EqualFold(parts[0], "Bearer") {
		return false
	}
	return parts[1] == expectedToken
}

func requestIDAndAccessLogMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := newRequestID()
		ctx := context.WithValue(r.Context(), requestIDKey, requestID)
		r = r.WithContext(ctx)
		w.Header().Set("X-Request-Id", requestID)

		rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		started := time.Now()
		next.ServeHTTP(rec, r)
		elapsed := time.Since(started)

		log.Printf("level=info requestId=%s method=%s path=%s status=%d durationMs=%d", requestID, r.Method, r.URL.Path, rec.status, elapsed.Milliseconds())
	})
}

func corsMiddleware(next http.Handler, allowOrigins []string) http.Handler {
	if len(allowOrigins) == 0 {
		return next
	}

	allowMap := make(map[string]struct{}, len(allowOrigins))
	for _, origin := range allowOrigins {
		allowMap[origin] = struct{}{}
	}

	allowMethods := "GET, POST, PATCH, DELETE, OPTIONS"
	allowHeaders := "Content-Type, Authorization, X-Requested-With, X-Request-Id"

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/api/v1/") {
			next.ServeHTTP(w, r)
			return
		}

		origin := r.Header.Get("Origin")
		if origin != "" {
			if _, ok := allowMap[origin]; ok {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Vary", "Origin")
				w.Header().Set("Access-Control-Allow-Methods", allowMethods)
				w.Header().Set("Access-Control-Allow-Headers", allowHeaders)
			}
		}

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func newRequestID() string {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		return time.Now().UTC().Format("20060102150405.000000000")
	}
	return hex.EncodeToString(buf)
}

func getOrCreateRequestID(w http.ResponseWriter, r *http.Request) string {
	if r != nil {
		if v, ok := r.Context().Value(requestIDKey).(string); ok && v != "" {
			return v
		}
	}
	if existing := w.Header().Get("X-Request-Id"); existing != "" {
		return existing
	}
	generated := newRequestID()
	w.Header().Set("X-Request-Id", generated)
	return generated
}

func tokyoNowISO() string {
	loc, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		loc = time.FixedZone("JST", 9*60*60)
	}
	return time.Now().In(loc).Format(time.RFC3339)
}
