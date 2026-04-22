package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"card-reading-system/internal/httpapi"
	"card-reading-system/internal/smartcard"
	"card-reading-system/internal/state"
)

func main() {
	logger := log.New(os.Stdout, "", log.LstdFlags)
	cardState := state.NewStore()
	service := smartcard.NewService(logger, cardState)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go service.Start(ctx)

	server := &http.Server{
		Addr:    ":4121",
		Handler: httpapi.NewHandler(cardState),
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), smartcard.DefaultPollInterval)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
	}()

	logger.Println("Local smartcard server running on http://localhost:4121")

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Fatalf("HTTP server error: %v", err)
	}
}
