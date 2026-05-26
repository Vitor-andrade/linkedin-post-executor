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
	"path/filepath"
	"syscall"
	"time"

	"github.com/Vitor-andrade/linkedin-post-executor/internal/ai"
	"github.com/Vitor-andrade/linkedin-post-executor/internal/linkedin"
	"github.com/Vitor-andrade/linkedin-post-executor/internal/schedule"
	"github.com/Vitor-andrade/linkedin-post-executor/internal/secret"
	"github.com/Vitor-andrade/linkedin-post-executor/internal/server"
	"github.com/Vitor-andrade/linkedin-post-executor/internal/store"
	"github.com/Vitor-andrade/linkedin-post-executor/web"
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
	log.Printf("AI provider: %s", provider.Name())

	cipher, err := secret.LoadOrCreate(envOr("LPE_KEY_FILE", defaultKeyPath()))
	if err != nil {
		log.Fatalf("secret: %v", err)
	}

	liCfg := linkedin.ConfigFromEnv()
	if liCfg.Configured() {
		log.Println("LinkedIn: credentials configured")
	} else {
		log.Println("LinkedIn: credentials not set (publishing disabled until configured)")
	}
	li := linkedin.NewService(liCfg, st, cipher)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	scheduler := schedule.New(st)
	scheduler.Start(ctx)

	handler := server.New(server.Deps{
		Store:    st,
		AI:       provider,
		LinkedIn: li,
		UI:       web.DistFS(),
	})

	srv := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("LinkedIn Post Executor running at http://localhost%s", addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutting down...")
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

// defaultKeyPath returns the path of the encryption key file, preferring the
// OS config directory and falling back to the working directory.
func defaultKeyPath() string {
	dir, err := os.UserConfigDir()
	if err != nil || dir == "" {
		return "lpe.key"
	}
	return filepath.Join(dir, "linkedin-post-executor", "key")
}
