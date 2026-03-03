package http

import (
	"net/http"

	"github.com/na0chan-go/home-dash/internal/usecase/health"
)

func NewRouter(healthUseCase *health.GetHealthUseCase) http.Handler {
	mux := http.NewServeMux()
	healthHandler := NewHealthHandler(healthUseCase)

	mux.HandleFunc("/api/v1/health", healthHandler.Get)
	return mux
}
