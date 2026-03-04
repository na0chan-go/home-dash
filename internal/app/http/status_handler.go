package http

import (
	"log"
	"net/http"

	usestatus "github.com/na0chan-go/home-dash/internal/usecase/status"
)

type StatusHandler struct {
	useCase *usestatus.GetStatusUseCase
}

func NewStatusHandler(useCase *usestatus.GetStatusUseCase) *StatusHandler {
	return &StatusHandler{useCase: useCase}
}

func (h *StatusHandler) Get(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, r, http.StatusBadRequest, errorCodeValidation, "method not allowed")
		return
	}

	result, err := h.useCase.Execute(r.Context())
	if err != nil {
		writeInternalError(w, r, errorCodeInternal, err)
		return
	}

	requestID := getOrCreateRequestID(w, r)
	if !result.DB.OK {
		log.Printf("level=warn requestId=%s code=STATUS_DB_DEGRADED path=%s err=%s", requestID, r.URL.Path, result.DBError)
	}
	if !result.Config.GarbageScheduleLoaded {
		log.Printf("level=warn requestId=%s code=STATUS_CONFIG_DEGRADED path=%s err=%s", requestID, r.URL.Path, result.ConfigError)
	}
	if result.LastBackupError != "" {
		log.Printf("level=warn requestId=%s code=STATUS_BACKUP_CHECK_FAILED path=%s err=%s", requestID, r.URL.Path, result.LastBackupError)
	}

	writeJSON(w, http.StatusOK, result)
}
