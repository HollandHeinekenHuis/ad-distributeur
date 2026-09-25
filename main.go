package main

import (
	"context"
	"embed"
	"errors"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

//go:embed web
var webFS embed.FS

func main() {
	cfg := loadConfig()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	srv := &server{cfg: cfg, mailer: newMailer(cfg)}
	if cfg.DatabaseURL == "" {
		log.Println("DATABASE_URL not set: bookings are disabled")
	} else {
		store, err := newStore(ctx, cfg.DatabaseURL)
		if err != nil {
			log.Fatalf("database config: %v", err)
		}
		defer store.Close()
		srv.store = store
		go srv.migrateLoop(ctx)
	}

	static, err := fs.Sub(webFS, "web")
	if err != nil {
		log.Fatalf("static files: %v", err)
	}

	httpSrv := &http.Server{
		Addr:              "0.0.0.0:" + cfg.Port,
		Handler:           srv.routes(static),
		ReadHeaderTimeout: 10 * time.Second,
	}
	go func() {
		<-ctx.Done()
		shCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_ = httpSrv.Shutdown(shCtx)
	}()

	log.Printf("listening on :%s", cfg.Port)
	if err := httpSrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}
