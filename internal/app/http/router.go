package http

import (
	"net/http"

	"github.com/na0chan-go/home-dash/internal/usecase/health"
	usenotes "github.com/na0chan-go/home-dash/internal/usecase/notes"
)

func NewRouter(
	healthUseCase *health.GetHealthUseCase,
	listNotesUseCase *usenotes.ListNotesUseCase,
	addNoteUseCase *usenotes.AddNoteUseCase,
	deleteNoteUseCase *usenotes.DeleteNoteUseCase,
	setPinUseCase *usenotes.SetPinUseCase,
	setDoneUseCase *usenotes.SetDoneUseCase,
) http.Handler {
	mux := http.NewServeMux()
	healthHandler := NewHealthHandler(healthUseCase)
	notesHandler := NewNotesHandler(listNotesUseCase, addNoteUseCase, deleteNoteUseCase, setPinUseCase, setDoneUseCase)

	mux.HandleFunc("/api/v1/health", healthHandler.Get)
	mux.HandleFunc("/api/v1/notes", notesHandler.HandleNotes)
	mux.HandleFunc("/api/v1/notes/", notesHandler.HandleByID)
	return mux
}
