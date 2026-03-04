package http

import (
	"encoding/json"
	"log"
	"net/http"
)

const (
	errorCodeValidation = "VALIDATION_ERROR"
	errorCodeNotFound   = "NOT_FOUND"
	errorCodeInternal   = "INTERNAL_ERROR"
	errorCodeConfig     = "CONFIG_ERROR"
)

type apiErrorDetail struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type apiErrorResponse struct {
	Error     apiErrorDetail `json:"error"`
	RequestID string         `json:"requestId"`
	Timestamp string         `json:"timestamp"`
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, r *http.Request, status int, code string, message string) {
	requestID := getOrCreateRequestID(w, r)
	writeJSON(w, status, apiErrorResponse{
		Error: apiErrorDetail{
			Code:    code,
			Message: message,
		},
		RequestID: requestID,
		Timestamp: tokyoNowISO(),
	})
}

func writeInternalError(w http.ResponseWriter, r *http.Request, code string, err error) {
	requestID := getOrCreateRequestID(w, r)
	log.Printf("level=error requestId=%s code=%s method=%s path=%s err=%v", requestID, code, r.Method, r.URL.Path, err)
	writeJSON(w, http.StatusInternalServerError, apiErrorResponse{
		Error: apiErrorDetail{
			Code:    code,
			Message: "internal server error",
		},
		RequestID: requestID,
		Timestamp: tokyoNowISO(),
	})
}
