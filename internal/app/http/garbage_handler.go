package http

import (
	"net/http"
	"strings"

	usegarbage "github.com/na0chan-go/home-dash/internal/usecase/garbage"
)

type GarbageHandler struct {
	todayUseCase    *usegarbage.GetTodayUseCase
	tomorrowUseCase *usegarbage.GetTomorrowUseCase
	summaryUseCase  *usegarbage.GetSummaryUseCase
}

func NewGarbageHandler(
	todayUseCase *usegarbage.GetTodayUseCase,
	tomorrowUseCase *usegarbage.GetTomorrowUseCase,
	summaryUseCase *usegarbage.GetSummaryUseCase,
) *GarbageHandler {
	return &GarbageHandler{
		todayUseCase:    todayUseCase,
		tomorrowUseCase: tomorrowUseCase,
		summaryUseCase:  summaryUseCase,
	}
}

func (h *GarbageHandler) Today(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, r, http.StatusBadRequest, errorCodeValidation, "method not allowed")
		return
	}

	result, err := h.todayUseCase.Execute(r.Context())
	if err != nil {
		h.handleGarbageError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *GarbageHandler) Tomorrow(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, r, http.StatusBadRequest, errorCodeValidation, "method not allowed")
		return
	}

	result, err := h.tomorrowUseCase.Execute(r.Context())
	if err != nil {
		h.handleGarbageError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *GarbageHandler) Summary(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, r, http.StatusBadRequest, errorCodeValidation, "method not allowed")
		return
	}

	result, err := h.summaryUseCase.Execute(r.Context())
	if err != nil {
		h.handleGarbageError(w, r, err)
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (h *GarbageHandler) handleGarbageError(w http.ResponseWriter, r *http.Request, err error) {
	if isConfigError(err) {
		writeInternalError(w, r, errorCodeConfig, err)
		return
	}
	writeInternalError(w, r, errorCodeInternal, err)
}

func isConfigError(err error) bool {
	if err == nil {
		return false
	}
	return strings.Contains(err.Error(), "garbage schedule")
}
