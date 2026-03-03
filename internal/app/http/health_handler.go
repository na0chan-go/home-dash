package http

import (
	"net/http"

	"github.com/na0chan-go/home-dash/internal/usecase/health"
)

type HealthHandler struct {
	useCase *health.GetHealthUseCase
}

func NewHealthHandler(useCase *health.GetHealthUseCase) *HealthHandler {
	return &HealthHandler{useCase: useCase}
}

func (h *HealthHandler) Get(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	result, err := h.useCase.Execute(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, result)
}
