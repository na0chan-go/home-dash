package http

import (
	"net/http"

	usedashboard "github.com/na0chan-go/home-dash/internal/usecase/dashboard"
)

type DashboardHandler struct {
	useCase *usedashboard.GetDashboardUseCase
}

func NewDashboardHandler(useCase *usedashboard.GetDashboardUseCase) *DashboardHandler {
	return &DashboardHandler{useCase: useCase}
}

func (h *DashboardHandler) Get(w http.ResponseWriter, r *http.Request) {
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
