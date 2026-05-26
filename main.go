// Command linkedin-post-executor runs the local HTTP server, background
// scheduler and serves the embedded UI as a single binary.
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

	"github.com/vitorandrade/linkedin-post-executor/internal/ai"
	"github.com/vitorandrade/linkedin-post-executor/internal/schedule"
	"github.com/vitorandrade/linkedin-post-executor/internal/server"
	"github.com/vitorandrade/linkedin-post-executor/internal/store"
	"github.com/vitorandrade/linkedin-post-executor/web"
)

func main() {
	addr := envOr("LPE_ADDR", ":8080")
	dbPath := envOr("LPE_DB", "data.db")

	st, err := store.Open(dbPath)
	if err != nil {
		log.Fatalf("store: %v", err)
	}
	defer st.Close()

	provider := ai.NewFromEnv()
	log.Printf("provedor de IA: %s", provider.Name())

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	scheduler := schedule.New(st)
	scheduler.Start(ctx)

	handler := server.New(server.Deps{
		Store: st,
		AI:    provider,
		UI:    web.DistFS(),
	})

	srv := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("LinkedIn Post Executor rodando em http://localhost%s", addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("encerrando...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown: %v", err)
	}
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
