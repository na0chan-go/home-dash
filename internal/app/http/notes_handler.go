package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	usenotes "github.com/na0chan-go/home-dash/internal/usecase/notes"
)

type NotesHandler struct {
	listUseCase    *usenotes.ListNotesUseCase
	addUseCase     *usenotes.AddNoteUseCase
	deleteUseCase  *usenotes.DeleteNoteUseCase
	setPinUseCase  *usenotes.SetPinUseCase
	setDoneUseCase *usenotes.SetDoneUseCase
}

type createNoteRequest struct {
	Kind   string `json:"kind"`
	Body   string `json:"body"`
	Pinned bool   `json:"pinned"`
	Done   bool   `json:"done"`
}

type setPinnedRequest struct {
	Pinned bool `json:"pinned"`
}

type setDoneRequest struct {
	Done bool `json:"done"`
}

type deleteNoteResponse struct {
	Deleted bool `json:"deleted"`
}

func NewNotesHandler(
	listUseCase *usenotes.ListNotesUseCase,
	addUseCase *usenotes.AddNoteUseCase,
	deleteUseCase *usenotes.DeleteNoteUseCase,
	setPinUseCase *usenotes.SetPinUseCase,
	setDoneUseCase *usenotes.SetDoneUseCase,
) *NotesHandler {
	return &NotesHandler{
		listUseCase:    listUseCase,
		addUseCase:     addUseCase,
		deleteUseCase:  deleteUseCase,
		setPinUseCase:  setPinUseCase,
		setDoneUseCase: setDoneUseCase,
	}
}

func (h *NotesHandler) HandleNotes(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.list(w, r)
	case http.MethodPost:
		h.create(w, r)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *NotesHandler) HandleByID(w http.ResponseWriter, r *http.Request) {
	id, action, ok := parseIDPath(r.URL.Path)
	if !ok {
		writeError(w, http.StatusNotFound, "not found")
		return
	}

	switch {
	case r.Method == http.MethodDelete && action == "":
		h.delete(w, r, id)
	case r.Method == http.MethodPatch && action == "pin":
		h.setPin(w, r, id)
	case r.Method == http.MethodPatch && action == "done":
		h.setDone(w, r, id)
	default:
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
	}
}

func (h *NotesHandler) list(w http.ResponseWriter, r *http.Request) {
	limit := 50
	if v := r.URL.Query().Get("limit"); v != "" {
		parsed, err := strconv.Atoi(v)
		if err != nil {
			writeError(w, http.StatusBadRequest, "limit must be an integer")
			return
		}
		limit = parsed
	}

	result, err := h.listUseCase.Execute(r.Context(), usenotes.ListNotesInput{
		Kind:  r.URL.Query().Get("kind"),
		Limit: limit,
	})
	if err != nil {
		h.handleUseCaseError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func (h *NotesHandler) create(w http.ResponseWriter, r *http.Request) {
	var req createNoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.addUseCase.Execute(r.Context(), usenotes.AddNoteInput{
		Kind:   req.Kind,
		Body:   req.Body,
		Pinned: req.Pinned,
		Done:   req.Done,
	})
	if err != nil {
		h.handleUseCaseError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, result)
}

func (h *NotesHandler) delete(w http.ResponseWriter, r *http.Request, id int64) {
	if err := h.deleteUseCase.Execute(r.Context(), id); err != nil {
		h.handleUseCaseError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, deleteNoteResponse{Deleted: true})
}

func (h *NotesHandler) setPin(w http.ResponseWriter, r *http.Request, id int64) {
	var req setPinnedRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.setPinUseCase.Execute(r.Context(), usenotes.SetPinInput{ID: id, Pinned: req.Pinned})
	if err != nil {
		h.handleUseCaseError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func (h *NotesHandler) setDone(w http.ResponseWriter, r *http.Request, id int64) {
	var req setDoneRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	result, err := h.setDoneUseCase.Execute(r.Context(), usenotes.SetDoneInput{ID: id, Done: req.Done})
	if err != nil {
		h.handleUseCaseError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, result)
}

func (h *NotesHandler) handleUseCaseError(w http.ResponseWriter, err error) {
	if errors.Is(err, usenotes.ErrNoteNotFound) {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	if usenotes.IsUserError(err) {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeError(w, http.StatusInternalServerError, "internal server error")
}

func parseIDPath(path string) (int64, string, bool) {
	trimmed := strings.TrimPrefix(path, "/api/v1/notes/")
	if trimmed == path || trimmed == "" {
		return 0, "", false
	}

	parts := strings.Split(trimmed, "/")
	if len(parts) != 1 && len(parts) != 2 {
		return 0, "", false
	}

	id, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || id < 1 {
		return 0, "", false
	}

	if len(parts) == 1 {
		return id, "", true
	}

	action := parts[1]
	if action != "pin" && action != "done" {
		return 0, "", false
	}
	return id, action, true
}
