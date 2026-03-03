package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/na0chan-go/home-dash/internal/app"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	a, err := app.New(ctx)
	if err != nil {
		log.Fatalf("failed to initialize app: %v", err)
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := a.Shutdown(shutdownCtx); err != nil {
			log.Printf("graceful shutdown failed: %v", err)
		}
	}()

	if err := a.Run(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("server exited with error: %v", err)
	}
}
