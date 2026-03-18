package http

import (
	"net/http"

	usedashboard "github.com/na0chan-go/home-dash/internal/usecase/dashboard"
	usegarbage "github.com/na0chan-go/home-dash/internal/usecase/garbage"
	"github.com/na0chan-go/home-dash/internal/usecase/health"
	usenotes "github.com/na0chan-go/home-dash/internal/usecase/notes"
	usestatus "github.com/na0chan-go/home-dash/internal/usecase/status"
)

func NewRouter(
	healthUseCase *health.GetHealthUseCase,
	listNotesUseCase *usenotes.ListNotesUseCase,
	addNoteUseCase *usenotes.AddNoteUseCase,
	deleteNoteUseCase *usenotes.DeleteNoteUseCase,
	setPinUseCase *usenotes.SetPinUseCase,
	setAckUseCase *usenotes.SetAcknowledgedUseCase,
	setDoneUseCase *usenotes.SetDoneUseCase,
	garbageTodayUseCase *usegarbage.GetTodayUseCase,
	garbageTomorrowUseCase *usegarbage.GetTomorrowUseCase,
	garbageSummaryUseCase *usegarbage.GetSummaryUseCase,
	dashboardUseCase *usedashboard.GetDashboardUseCase,
	statusUseCase *usestatus.GetStatusUseCase,
	adminBackupHandler *AdminBackupHandler,
	spaHandler *SPAHandler,
	corsAllowOrigins []string,
	authToken string,
) http.Handler {
	mux := http.NewServeMux()
	healthHandler := NewHealthHandler(healthUseCase)
	notesHandler := NewNotesHandler(listNotesUseCase, addNoteUseCase, deleteNoteUseCase, setPinUseCase, setAckUseCase, setDoneUseCase)
	garbageHandler := NewGarbageHandler(garbageTodayUseCase, garbageTomorrowUseCase, garbageSummaryUseCase)
	dashboardHandler := NewDashboardHandler(dashboardUseCase)
	statusHandler := NewStatusHandler(statusUseCase)

	mux.HandleFunc("/api/v1/health", healthHandler.Get)
	mux.HandleFunc("/api/v1/notes", notesHandler.HandleNotes)
	mux.HandleFunc("/api/v1/notes/", notesHandler.HandleByID)
	mux.HandleFunc("/api/v1/garbage/today", garbageHandler.Today)
	mux.HandleFunc("/api/v1/garbage/tomorrow", garbageHandler.Tomorrow)
	mux.HandleFunc("/api/v1/garbage/summary", garbageHandler.Summary)
	mux.HandleFunc("/api/v1/dashboard", dashboardHandler.Get)
	mux.HandleFunc("/api/v1/status", statusHandler.Get)
	mux.HandleFunc("/api/v1/admin/backup", adminBackupHandler.Create)
	mux.HandleFunc("/api/v1", func(w http.ResponseWriter, r *http.Request) {
		writeError(w, r, http.StatusNotFound, errorCodeNotFound, "not found")
	})
	mux.HandleFunc("/api/v1/", func(w http.ResponseWriter, r *http.Request) {
		writeError(w, r, http.StatusNotFound, errorCodeNotFound, "not found")
	})

	mux.HandleFunc("/", spaHandler.Serve)

	return applyMiddlewares(mux, corsAllowOrigins, authToken)
}
