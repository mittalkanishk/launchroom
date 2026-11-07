// Command renderer renders explainer scenes for studio: scene source in, narrated MP4 and subtitles
// out. The same program runs in two keyless containers, one per engine: RENDERER_ENGINE=manim (the
// renderer container) or stage (the stage container, anime.js in Chromium). See internal/renderer.
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mittalkanishk/launch-room/internal/renderer"
	"github.com/mittalkanishk/launch-room/internal/run"
)

func main() {
	engine, err := pickEngine(env("RENDERER_ENGINE", "manim"))
	if err != nil {
		log.Fatalf("renderer: %v", err)
	}
	r := renderer.New(env("RENDERER_WORK", os.TempDir()), engine, run.NewExec(os.Stderr, os.Stderr), 15*time.Minute)
	server := &http.Server{
		Addr:              env("RENDERER_LISTEN", ":8091"),
		Handler:           r.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
	}

	// The server runs in this goroutine; it ends when Shutdown is called below on SIGTERM/SIGINT.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("renderer: %v", err)
		}
	}()
	log.Printf("renderer: listening on %s", server.Addr)
	<-ctx.Done()

	shutdown, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	server.Shutdown(shutdown)
}

// pickEngine is the one place the engine is chosen.
func pickEngine(name string) (renderer.Engine, error) {
	switch name {
	case "manim":
		return renderer.Manim{}, nil
	case "stage":
		return renderer.Stage{Script: env("STAGE_SCRIPT", "/opt/stage/runtime/render.mjs")}, nil
	}
	return nil, fmt.Errorf("unknown RENDERER_ENGINE %q; use manim or stage", name)
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
