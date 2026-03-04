package http

import (
	"net/http"

	usebackup "github.com/na0chan-go/home-dash/internal/usecase/backup"
)

type AdminBackupHandler struct {
	useCase *usebackup.RunBackupUseCase
}

func NewAdminBackupHandler(useCase *usebackup.RunBackupUseCase) *AdminBackupHandler {
	return &AdminBackupHandler{useCase: useCase}
}

func (h *AdminBackupHandler) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, r, http.StatusBadRequest, errorCodeValidation, "method not allowed")
		return
	}

	result, err := h.useCase.Execute(r.Context())
	if err != nil {
		writeInternalError(w, r, errorCodeInternal, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}
