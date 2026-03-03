package app

import (
	"context"
	"database/sql"
	"net/http"
	"time"

	appconfig "github.com/na0chan-go/home-dash/internal/app/config"
	httpapi "github.com/na0chan-go/home-dash/internal/app/http"
	"github.com/na0chan-go/home-dash/internal/infra/config"
	"github.com/na0chan-go/home-dash/internal/infra/db"
	infranotes "github.com/na0chan-go/home-dash/internal/infra/notes"
	"github.com/na0chan-go/home-dash/internal/infra/system"
	"github.com/na0chan-go/home-dash/internal/usecase/health"
	usenotes "github.com/na0chan-go/home-dash/internal/usecase/notes"
)

type App struct {
	server *http.Server
	db     *sql.DB
}

func New(ctx context.Context) (*App, error) {
	cfg := appconfig.LoadFromEnv()

	sqliteDB, err := db.OpenSQLite(cfg.DBPath)
	if err != nil {
		return nil, err
	}

	if err := db.RunMigrations(ctx, sqliteDB); err != nil {
		_ = sqliteDB.Close()
		return nil, err
	}

	if _, err := config.LoadGarbageSchedule(cfg.GarbageSchedulePath); err != nil {
		_ = sqliteDB.Close()
		return nil, err
	}

	notesRepo := infranotes.NewSQLiteRepository(sqliteDB)
	listNotesUseCase := usenotes.NewListNotesUseCase(notesRepo)
	addNoteUseCase := usenotes.NewAddNoteUseCase(notesRepo)
	deleteNoteUseCase := usenotes.NewDeleteNoteUseCase(notesRepo)
	setPinUseCase := usenotes.NewSetPinUseCase(notesRepo)
	setDoneUseCase := usenotes.NewSetDoneUseCase(notesRepo)

	clock := system.NewClock()
	healthUseCase := health.NewGetHealthUseCase(clock)

	router := httpapi.NewRouter(
		healthUseCase,
		listNotesUseCase,
		addNoteUseCase,
		deleteNoteUseCase,
		setPinUseCase,
		setDoneUseCase,
	)
	server := &http.Server{
		Addr:              cfg.AppAddr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	return &App{server: server, db: sqliteDB}, nil
}

func (a *App) Run() error {
	return a.server.ListenAndServe()
}

func (a *App) Shutdown(ctx context.Context) error {
	if err := a.server.Shutdown(ctx); err != nil {
		_ = a.db.Close()
		return err
	}
	return a.db.Close()
}
