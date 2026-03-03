package http

import (
	"net/http"

	usedashboard "github.com/na0chan-go/home-dash/internal/usecase/dashboard"
	usegarbage "github.com/na0chan-go/home-dash/internal/usecase/garbage"
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
	garbageTodayUseCase *usegarbage.GetTodayUseCase,
	garbageTomorrowUseCase *usegarbage.GetTomorrowUseCase,
	garbageSummaryUseCase *usegarbage.GetSummaryUseCase,
	dashboardUseCase *usedashboard.GetDashboardUseCase,
) http.Handler {
	mux := http.NewServeMux()
	healthHandler := NewHealthHandler(healthUseCase)
	notesHandler := NewNotesHandler(listNotesUseCase, addNoteUseCase, deleteNoteUseCase, setPinUseCase, setDoneUseCase)
	garbageHandler := NewGarbageHandler(garbageTodayUseCase, garbageTomorrowUseCase, garbageSummaryUseCase)
	dashboardHandler := NewDashboardHandler(dashboardUseCase)

	mux.HandleFunc("/api/v1/health", healthHandler.Get)
	mux.HandleFunc("/api/v1/notes", notesHandler.HandleNotes)
	mux.HandleFunc("/api/v1/notes/", notesHandler.HandleByID)
	mux.HandleFunc("/api/v1/garbage/today", garbageHandler.Today)
	mux.HandleFunc("/api/v1/garbage/tomorrow", garbageHandler.Tomorrow)
	mux.HandleFunc("/api/v1/garbage/summary", garbageHandler.Summary)
	mux.HandleFunc("/api/v1/dashboard", dashboardHandler.Get)
	return mux
}
