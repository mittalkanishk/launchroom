// Command studio is the video crew's back end: an HTTP service inside the stack that has explainer
// scenes rendered by the renderer, joins them, grades and mixes the cut, and measures it for QC.
// See internal/studio.
//
// Read in this order:
//
//	main.go                     the settings and the concrete types (newService)
//	../../internal/studio       the service: jobs, steps, API
//	then the packages newService names
package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mittalkanishk/launch-room/internal/media"
	"github.com/mittalkanishk/launch-room/internal/renderer"
	"github.com/mittalkanishk/launch-room/internal/run"
	"github.com/mittalkanishk/launch-room/internal/studio"
)

func main() {
	svc, err := newService()
	if err != nil {
		log.Fatalf("studio: %v", err)
	}
	server := &http.Server{
		Addr:              env("STUDIO_LISTEN", ":8090"),
		Handler:           svc.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
	}

	// The server runs in this goroutine; it ends when Shutdown is called below on SIGTERM/SIGINT.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("studio: %v", err)
		}
	}()
	log.Printf("studio: listening on %s", server.Addr)
	<-ctx.Done()

	shutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	server.Shutdown(shutdown)
}

// newService is the one place the concrete types are chosen: scenes render in their engine's keyless
// container (renderer for Manim, stage for anime.js; a long scene takes minutes), and
// post-production is ffmpeg here.
func newService() (*studio.Service, error) {
	dataDir := env("STUDIO_DATA", "/data")
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		return nil, err
	}
	slow := &http.Client{Timeout: 20 * time.Minute}
	renderers := map[string]studio.Renderer{
		"manim": renderer.NewClient(env("RENDERER_URL", "http://renderer:8091"), slow),
		"stage": renderer.NewClient(env("STAGE_URL", "http://stage:8092"), slow),
	}
	tools := media.NewTools(run.NewExec(os.Stderr, os.Stderr))
	return studio.NewService(dataDir, renderers, tools)
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
